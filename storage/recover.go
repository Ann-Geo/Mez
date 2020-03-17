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

	"github.com/Ann-Geo/Mez/storagepb"
	"github.com/golang/protobuf/ptypes"
)

func (memlog *MemLog) Recover(recoveryFile *os.File) {

	var recoveryPath string

	//Get the recoveryPath frrom recovery file
	scanner := bufio.NewScanner(recoveryFile)
	for scanner.Scan() {
		recoveryPath = scanner.Text()
		break
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

		/*f, _ := os.Open(file)
		defer f.Close()

		var offset int64 = 0
		var j uint64
		//for j = 0; j < SEGSIZE; j++ {
			//read length of a proto msg first
			//read that many bytes after that
			//set offset to new seek value
			//repeat this
		}*/

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
			fmt.Println(memlog.tsmemlog.tslog)
		}

		//array := bytes.Split(b, []byte("mez"))

		//fmt.Println(file)

		/*for _, m := range array {
			//fmt.Printf("%T", m)
			count := 0
			for _, k := range m {
				if k != 0 {
					count = count + 1
				}
			}
			//fmt.Println(count)
		}

		for j := 0; j < len(array); j++ {
			pb := &storagepb.BFileItem{}

			//err2 := proto.Unmarshal(append([]byte(nil), array[j]...), pb)

			err2 := json.Unmarshal(array[j], pb)
			if err2 != nil {
				log.Fatalln("Couldn't put the bytes into pb", err2)
			}

			im := pb.GetImg()
			ts := pb.GetTs()

			t, _ := ptypes.Timestamp(ts)
			memlog.Append(im, t)

		}*/

		/*
			pbSlice := make([]storagepb.BFileItem, 0)
			//err2 := json.Unmarshal(b, &pbSlice)

			//b1 := bytes.Trim([]byte(b), "{")
			//b2 := bytes.Trim([]byte(b1), ":")

			err2 := json.Unmarshal(b, &pbSlice)
			if err2 != nil {
				log.Fatalln("Cannot unmarshal bytes", err2)
			}

			for j := 0; j < len(pbSlice); j++ {
				im := pbSlice[j].GetImg()
				ts := pbSlice[j].GetTs()

				t, _ := ptypes.Timestamp(ts)
				memlog.Append(im, t)

			}*/

		/*pb := storagepb.BFileItem{}
		err2 := json.Unmarshal(append([]byte(nil), b...), &pb)
		if err2 != nil {
			log.Fatalln("Couldn't put the bytes into pb", err2)
		}

		//im := pb.GetImg()
		//ts := pb.GetTs()

		//t, _ := ptypes.Timestamp(ts)
		//memlog.Append(im, t)
		*/
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
