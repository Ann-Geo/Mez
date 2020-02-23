package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
	"vsc_workspace/Mez_upload_woa/api/edgenode"
	"vsc_workspace/Mez_upload_woa/client"

	"gocv.io/x/gocv"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"
var imageFilesPath string = "/home/research/pythonwork/SEM_5/Knobs-redo/optimize_controller/duke/medium/"
var frameRate uint64 = 200

func main() {

	//create a producer client with login and password
	producer := client.NewProducerClient("client", "edge")

	//Connect API, specify EN broker url and user address
	err := producer.Connect("127.0.0.1:10000", "127.0.0.1:9050")
	if err != nil {
		log.Fatalf("error while calling Connect")
	}

	defer producer.ConnProdClient.Close()
	defer producer.Cancel()

	// Open a stream to gRPC server
	stream, err := producer.Cl.Publish(producer.Ctx)
	if err != nil {

		log.Fatalf("error while calling publish")

	}
	defer stream.CloseSend()

	// Read image file names
	errMsg, files := walkAllFilesInDir(imageFilesPath)
	if errMsg != "file read success" {
		log.Fatalf("cannot read files")

	}

	//start publishing files
	for _, file := range files {
		fmt.Println(file)

		//imBuf, err := ioutil.ReadFile(file)
		//fmt.Printf("%T\n", imBuf)
		//if err != nil {
		//log.Fatalf("cannot read file")

		//}

		imBuf := gocv.IMRead(file, gocv.IMReadColor)
		buffer := imBuf.ToBytes()

		time.Sleep(time.Duration(frameRate) * time.Millisecond)
		ts := time.Now().Format(customTimeformat)
		err = stream.Send(&edgenode.Image{
			Image:     buffer,
			Timestamp: ts,
		})
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}

	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("cannot close stream")
	}

}

func walkAllFilesInDir(dir string) (string, []string) {
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
}
