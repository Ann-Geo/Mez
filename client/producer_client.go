// Producer client (camera & edgenode functionality)
package client

import (
	"github.com/arun-ravindran/Raven/api/edgenode"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type ProducerClient struct {
	Auth Authentication
}

func NewProducerClient(login, password string) *ProducerClient {
	return &ProducerClient{Auth: Authentication{login: login, password: password}}

}

func (pc *ProducerClient) PublishImage(client edgenode.PubSubClient) error {
	// Client side streaming
	var (
		writing        = true
		buf            []byte
		chunkSize      int = 1 << 20
		imageFilesPath     = "../test_images"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open a stream to gRPC server
	stream, err := client.Publish(ctx)
	if err != nil {
		return fmt.Errorf("%v.Publish(_) = _, %v", client, err)
	}
	defer stream.CloseSend()

	// Read image file names
	var files []string
	absImageFilePath, _ := filepath.Abs(imageFilesPath)
	err = filepath.Walk(absImageFilePath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}
	// Copy image file to buffer
	numSend := 0
	for i := 1; i < len(files); i++ {
		writing = true
		fd, err := os.Open(files[i])
		if err != nil {
			return err
		}
		defer fd.Close()
		buf = make([]byte, chunkSize)
		for writing {
			n, err := fd.Read(buf)
			if err != nil {
				if err == io.EOF {
					writing = false
					err = nil
					continue
				}
				return fmt.Errorf("errored while copying from file to buf")
			}
			time.Sleep(1 * time.Second)
			ts, _ := time.Parse(time.RFC850, time.Now().Format(time.RFC850))
			err = stream.Send(&edgenode.Image{
				Image:     buf[:n],                // Needed if n < chunksize
				Timestamp: ts.Format(time.RFC850), // Timestamp the image
			})
			numSend++
			log.Println("ProducerClient: Image size sent kB", n/1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to send chunk via stream")
			}
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Println("From TestBroker:PublishImage", reply)
	return nil
}
