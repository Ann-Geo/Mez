package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Ann-Geo/Mez/storagepb"
	"github.com/golang/protobuf/ptypes"
)

func (memlog *MemLog) Recover(recoveryFile *os.File, camid string) {

	var recoveryPath string

	//Get the recoveryPath frrom recovery file
	scanner := bufio.NewScanner(recoveryFile)
	for scanner.Scan() {

		if camid == "nil" {
			recoveryPath = scanner.Text()
		} else {
			rPath := scanner.Text()
			if strings.Contains(rPath, camid) {
				recoveryPath = rPath
			} else {
				continue
			}

		}

		//Obtain filenames in the recovery path
		fileList, err := ioutil.ReadDir(recoveryPath)
		if err != nil {
			log.Fatalln("File read failed", err)
		}

		//sort file names based on modified time
		//- to get the most recently written segments
		sort.Slice(fileList, func(i, j int) bool {
			return fileList[i].ModTime().Unix() > fileList[j].ModTime().Unix()
		})

		//starting from last file - most recently written
		var i int = len(fileList) - 1
		for i > 0 {

			//get data from the backup file
			/*b, e := ioutil.ReadFile(recoveryPath + fileList[i].Name())
			fmt.Println(recoveryPath + fileList[i].Name())
			if e != nil {
				log.Fatalln("cannot read backup file", e)
			}*/

			b, e := decryptFile(recoveryPath+fileList[i].Name(), "password")
			fmt.Println(recoveryPath + fileList[i].Name())
			if e != nil {
				log.Fatalln("cannot read and decrypt backup file", e)
			}

			pb := &storagepb.BFileItem{}

			//json decode data read to pb
			dec := json.NewDecoder(bytes.NewReader(b))
			var im []byte
			var t time.Time
			var imByteSlice [][]byte
			var tsByteSlice []time.Time

			//get images and timestamps to imByteSlice and tsByteSlice
			for {
				if err := dec.Decode(pb); err == io.EOF {
					break
				} else if err != nil {
					log.Fatalln("json data read error", err)
				}
				im = append([]byte(nil), pb.GetImg()...)
				ts := pb.GetTs()
				t, _ = ptypes.Timestamp(ts)
				imByteSlice = append(imByteSlice, im)
				tsByteSlice = append(tsByteSlice, t)
			}

			//comparing crc and discarding partially written segment files
			//calculate crc of last image read
			crcVal := crc32.ChecksumIEEE(im)
			fmt.Println("crc calculated", crcVal)

			//open crc file written during backup
			crcFile, err := os.Open(recoveryPath + fileList[i-1].Name())
			if err != nil {
				log.Fatalln("CRC File reading error", err)
			}

			var crcRead uint32

			//read the crc written in crc file during backup
			for {
				_, _ = fmt.Fscanf(crcFile, "%d\n", &crcRead)
				fmt.Println("crc Read", crcRead)
				break
			}

			//if crc calcultaed and crc written are same recover the
			//segment file to memlog
			if crcVal == crcRead {
				fmt.Println("yes")
				var j int
				for j < len(imByteSlice) {

					memlog.Append(imByteSlice[j], tsByteSlice[j])
					j = j + 1
				}
			}

			i = i - 2

		}

	}

}
