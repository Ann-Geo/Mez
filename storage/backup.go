package storage

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"strconv"

	"github.com/Ann-Geo/Mez/storagepb"
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

		//fmt.Println("backup waiting")

		//wait for backupItem from memlog Append
		b = <-memlog.bchan
		//fmt.Println(b.pos)

		//create backupfile
		fname := fPath + strconv.FormatUint(b.pos, 10) + ".json"
		fmt.Println(fname)

		var _, err = os.Stat(fname)

		// create file if not exists
		if !os.IsNotExist(err) {
			_ = os.Remove(fname)
		}

		bFile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("cannot create backup file", err)
		}

		fmt.Println(len(b.tseg.ts))

		var i int
		for i = 0; i < len(b.tseg.ts); i++ {

			tproto, _ := ptypes.TimestampProto(b.tseg.ts[i])
			item := &storagepb.BFileItem{
				Img: append([]byte(nil), b.imseg.im[i]...), //b.imseg.im[i],
				Ts:  tproto,
			}

			//out, err := proto.Marshal(item)
			out, err := json.Marshal(item)
			//if i != 0 {
			//fmt.Fprintf(bFile, "mez")
			//}

			//fmt.Println(len(out))

			if err != nil {
				log.Fatalln("cannot serialize log item to bytes", err)
			}

			_, err = bFile.Write(out)
			if err != nil {
				log.Fatalln("cannot write log item to file", err)
			}

		}

		//crc calculation and persisting crc to a afile
		crcVal := crc32.ChecksumIEEE(b.imseg.im[i-1])

		//create crc backupfile
		cname := fPath + "crc" + strconv.FormatUint(b.pos, 10) + ".txt"
		fmt.Println(cname)

		_, err = os.Stat(cname)

		// create file if not exists
		if !os.IsNotExist(err) {
			_ = os.Remove(cname)
		}

		cFile, err := os.OpenFile(cname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("cannot create crc file", err)
		}

		fmt.Fprintf(cFile, "%d\n", crcVal)

		if err := bFile.Close(); err != nil {
			log.Fatalln("cannot close backup file", err)
		}

		//fmt.Println("backup done", time.Now())

	}
}
