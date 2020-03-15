package storage

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Ann-Geo/Mez/storagepb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

type backupItem struct {
	pos   uint64
	imseg ImageLogSeg
	tseg  TimestampLogSeg
}

func (memlog *MemLog) Backup(fPath string) {
	memlog.bFlag = true
	var b backupItem
	for {

		//wait for backupItem from memlog Append
		b = <-memlog.bchan
		fmt.Println(b.pos)

		//create backupfile
		fname := fPath + strconv.FormatUint(b.pos, 10) + ".txt"
		bFile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("cannot create backup file", err)
		}

		for i := 0; i < len(b.tseg.ts); i++ {
			tproto, _ := ptypes.TimestampProto(b.tseg.ts[i])
			item := &storagepb.BFileItem{
				Img: b.imseg.im[i],
				Ts:  tproto,
			}

			out, err := proto.Marshal(item)
			if err != nil {
				log.Fatalln("cannot serialize log item to bytes", err)
			}

			if _, err := bFile.Write(out); err != nil {
				log.Fatalln("cannot write log item to file", err)
			}

		}

		if err := bFile.Close(); err != nil {
			log.Fatalln("cannot close backup file", err)
		}

		//fmt.Println("backup done", time.Now())

	}
}
