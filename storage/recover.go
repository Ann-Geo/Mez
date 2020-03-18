package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

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
			return fileList[i].ModTime().Unix() < fileList[j].ModTime().Unix()
		})

		fmt.Println(fileList)

		for _, file := range fileList {

			b, e := ioutil.ReadFile(recoveryPath + file.Name())
			fmt.Println(recoveryPath + file.Name())
			if e != nil {
				panic(e)
			}

			pb := &storagepb.BFileItem{}

			dec := json.NewDecoder(bytes.NewReader(b))
			for {

				if err := dec.Decode(pb); err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}
				im := pb.GetImg()
				ts := pb.GetTs()

				t, _ := ptypes.Timestamp(ts)
				memlog.Append(im, t)
				//fmt.Println(memlog.tsmemlog.tslog)
			}

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
