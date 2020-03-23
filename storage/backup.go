package storage

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
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

		//create backupfile name
		fname := fPath + strconv.FormatUint(b.pos, 10) + ".json"
		fmt.Println(fname)

		var _, err = os.Stat(fname)

		//if file exists remove old file
		if !os.IsNotExist(err) {
			_ = os.Remove(fname)
		}

		//create and open backup file
		bFile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("cannot create backup file", err)
		}

		//fmt.Println(len(b.tseg.ts))

		var i int

		//get each image and timestamp
		//pack into item after marshalling
		//write json item to back up file

		items := &storagepb.BFileItem{}

		for i = 0; i < len(b.tseg.ts); i++ {

			tproto, _ := ptypes.TimestampProto(b.tseg.ts[i])
			item := &storagepb.BFileItem{
				Img: append([]byte(nil), b.imseg.im[i]...),
				Ts:  tproto,
			}

			proto.Merge(items, item)

		}

		out, err := json.Marshal(items)
		if err != nil {
			log.Fatalln("cannot serialize log item to bytes", err)
		}

		err = encryptFile(bFile, out, "password") //bFile.Write(out)
		if err != nil {
			log.Fatalln("cannot encrypt and write log item to file", err)
		}

		//crc calculation and persisting crc to a afile
		//calculate crc of last image written in a segment
		crcVal := crc32.ChecksumIEEE(b.imseg.im[i-1])

		//create crc backupfile for a segment
		cname := fPath + "crc" + strconv.FormatUint(b.pos, 10) + ".txt"
		fmt.Println(cname)

		_, err = os.Stat(cname)

		//if file already exists remove it
		if !os.IsNotExist(err) {
			_ = os.Remove(cname)
		}

		//create and open crc file
		cFile, err := os.OpenFile(cname, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("cannot create crc file", err)
		}

		fmt.Fprintf(cFile, "%d\n", crcVal)

		if err := bFile.Close(); err != nil {
			log.Fatalln("cannot close backup file", err)
		}

	}
}
