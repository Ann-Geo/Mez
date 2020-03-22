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

		//Obtain filenames in the recovery path given
		//errMsg, fileList := walkFilesInDir(recoveryPath)
		fileList, err := ioutil.ReadDir(recoveryPath)
		if err != nil {
			log.Fatalln("File read failed", err)
		}

		sort.Slice(fileList, func(i, j int) bool {
			return fileList[i].ModTime().Unix() > fileList[j].ModTime().Unix()
		})

		for _, f := range fileList {
			fmt.Println(f.Name())
		}

		var i int = len(fileList) - 1
		for i > 0 {

			b, e := ioutil.ReadFile(recoveryPath + fileList[i].Name())
			fmt.Println(recoveryPath + fileList[i].Name())
			if e != nil {
				log.Fatalln("cannot read backup file", e)
			}

			pb := &storagepb.BFileItem{}

			dec := json.NewDecoder(bytes.NewReader(b))
			var im []byte
			var t time.Time
			var imByteSlice [][]byte
			var tsByteSlice []time.Time
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

			crcVal := crc32.ChecksumIEEE(im)
			fmt.Println("crc calculated", crcVal)
			crcFile, err := os.Open(recoveryPath + fileList[i-1].Name())
			if err != nil {
				log.Fatalln("CRC File reading error", err)
			}

			var crcRead uint32
			for {

				_, _ = fmt.Fscanf(crcFile, "%d\n", &crcRead)
				fmt.Println("crc Read", crcRead)

				/*if err != nil {

					if err == io.EOF {
						break // stop reading the file
					}
					log.Fatalln("cannot read crc", err)

				}*/
				break
			}

			if crcVal == crcRead {
				fmt.Println("yes")
				var j int
				for j < len(imByteSlice) {

					memlog.Append(imByteSlice[j], tsByteSlice[j])
					//fmt.Println(memlog.tsmemlog.tslog)
					j = j + 1
				}
			}

			i = i - 2

		}

	}

}

/*
func walkFilesInDir(dir string) (string, []string) {
	var errMsg string = "file read success"
	fileList := make([]string, 0)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errMsg = "Incorrect file path"
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() {
			fileList = append(fileList, path)

		}
		return nil
	})

	return errMsg, fileList
}*/
