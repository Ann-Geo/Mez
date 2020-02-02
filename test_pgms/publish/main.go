package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
	"vsc_workspace/Mez_upload/api/edgenode"
	"vsc_workspace/Mez_upload/client"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"
var imageFilesPath string = "/home/research/pythonwork/SEM_5/openpose_obj_Det/simple/"
var frameRate uint64 = 200

func main() {

	producer := client.NewProducerClient("prodClient", "edge")

	creds, err := credentials.NewClientTLSFromFile("../../cert/server.crt", "")
	if err != nil {
		log.Fatalf("tls error")
	}

	// Connect test client with edge node broker
	connProdClient, err := grpc.Dial("127.0.0.1:10000", grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&producer.Auth))
	if err != nil {
		log.Fatalf("could not connect with EN broker")
	}
	defer connProdClient.Close()
	cl := edgenode.NewPubSubClient(connProdClient)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Open a stream to gRPC server
	stream, err := cl.Publish(ctx)
	if err != nil {

		log.Fatalf("error while calling publish")

	}
	defer stream.CloseSend()

	// Read image file names
	errMsg, files := walkAllFilesInDir(imageFilesPath)
	if errMsg != "file read success" {
		log.Fatalf("cannot read files")

	}

	for _, file := range files {
		fmt.Println("yes")

		imBuf, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("cannot read file")

		}
		time.Sleep(time.Duration(frameRate) * time.Millisecond)
		ts := time.Now().Format(customTimeformat)
		err = stream.Send(&edgenode.Image{
			Image:     imBuf,
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
