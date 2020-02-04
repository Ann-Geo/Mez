// Producer client (camera & edgenode functionality)
package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"vsc_workspace/Mez_upload/api/edgenode"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

type ProducerClient struct {
	Auth           Authentication
	ConnProdClient *grpc.ClientConn
	Cl             edgenode.PubSubClient
	Cancel         context.CancelFunc
	Ctx            context.Context
}

func NewProducerClient(login, password string) *ProducerClient {
	return &ProducerClient{Auth: Authentication{login: login, password: password}}

}

func (pc *ProducerClient) Connect(url, userAddress string) error {
	creds, err := credentials.NewClientTLSFromFile("../../cert/server.crt", "")
	if err != nil {
		return err
	}

	// Dial to Mez
	pc.ConnProdClient, err = grpc.Dial(url, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&pc.Auth))
	if err != nil {
		return err
	}

	pc.Cl = edgenode.NewPubSubClient(pc.ConnProdClient)

	pc.Ctx, pc.Cancel = context.WithTimeout(context.Background(), 100*time.Second)

	//Connect with Mez
	connReq := &edgenode.Url{
		Address: userAddress,
	}
	_, connErr := pc.Cl.Connect(pc.Ctx, connReq)
	if connErr != nil {
		return connErr
	}

	return nil
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
			ts, _ := time.Parse(customTimeformat, time.Now().Format(customTimeformat))
			err = stream.Send(&edgenode.Image{
				Image:     buf[:n],                     // Needed if n < chunksize
				Timestamp: ts.Format(customTimeformat), // Timestamp the image
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

func (pc *ProducerClient) PublishImageTest(client edgenode.PubSubClient, imageFilesPath string, numImagesInsert,
	frameRate uint64, imSizeParam string) (string, []string, []int, uint64) {

	var (
		imBuf []byte
	)
	tsPublished := make([]string, 0)
	imSizePublished := make([]int, 0)
	var numSend uint64

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration((numImagesInsert*frameRate)+2000)*time.Second)
	defer cancel()

	// Open a stream to gRPC server
	stream, err := client.Publish(ctx)
	if err != nil {

		return "error while invoking Publish", tsPublished, imSizePublished, numSend

	}
	defer stream.CloseSend()

	// Read image file names
	errMsg, files := walkAllFilesInDir(imageFilesPath)
	if errMsg != "file read success" {
		return "File read failed", tsPublished, imSizePublished, numSend

	}

	//slices to store timestamps and image sizes published

	var i uint64

	if imSizeParam == "S" {
		for i = 0; i < numImagesInsert; i++ {
			//read first file to buffer (conversion to bytes)
			imBuf, err = ioutil.ReadFile(files[0])
			if err != nil {
				return "Cannot read image file", tsPublished, imSizePublished, numSend

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			tsPublished = append(tsPublished, ts)
			numSend++
			if err != nil {
				if err == io.EOF {
					break
				}
				return "failed to send chunk via stream", tsPublished, imSizePublished, numSend
			}
			imSizePublished = append(imSizePublished, len(imBuf))

		}
	} else {

		for _, file := range files {
			if numSend == numImagesInsert {
				break
			}
			imBuf, err = ioutil.ReadFile(file)
			if err != nil {
				return "Cannot read image file", tsPublished, imSizePublished, numSend

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			tsPublished = append(tsPublished, ts)
			numSend++
			if err != nil {
				if err == io.EOF {
					break
				}
				return "failed to send chunk via stream", tsPublished, imSizePublished, numSend
			}
			imSizePublished = append(imSizePublished, len(imBuf))

		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return "stream CloseAndRecv() error", tsPublished, imSizePublished, numSend
	}

	return "publish success", tsPublished, imSizePublished, numSend
}

func (pc *ProducerClient) PublishImageTestConcurrent(client edgenode.PubSubClient, imageFilesPath string, numImagesInsert,
	frameRate uint64, imSizeParam string) {

	var (
		imBuf []byte
	)
	publishErrMsg := "publish success"
	tsPublished := make([]string, 0)
	imSizePublished := make([]int, 0)
	var numSend uint64

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration((numImagesInsert*frameRate)+2000)*time.Second)
	defer cancel()

	// Open a stream to gRPC server
	stream, err := client.Publish(ctx)
	if err != nil {

		//return "error while invoking Publish", tsPublished, imSizePublished, numSend
		publishErrMsg = "error while invoking Publish"

	}
	defer stream.CloseSend()

	// Read image file names
	errMsg, files := walkAllFilesInDir(imageFilesPath)
	if errMsg != "file read success" {
		//return "File read failed", tsPublished, imSizePublished, numSend
		publishErrMsg = "File read failed"

	}

	//slices to store timestamps and image sizes published

	var i uint64

	if imSizeParam == "S" {
		for i = 0; i < numImagesInsert; i++ {
			//read first file to buffer (conversion to bytes)
			imBuf, err = ioutil.ReadFile(files[0])
			if err != nil {
				//return "Cannot read image file", tsPublished, imSizePublished, numSend
				publishErrMsg = "Cannot read image file"

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			tsPublished = append(tsPublished, ts)
			numSend++
			if err != nil {
				if err == io.EOF {
					break
				}
				//return "failed to send chunk via stream", tsPublished, imSizePublished, numSend
				publishErrMsg = "failed to send chunk via stream"
			}
			imSizePublished = append(imSizePublished, len(imBuf))

		}
	} else {

		for _, file := range files {
			if numSend == numImagesInsert {
				break
			}
			imBuf, err = ioutil.ReadFile(file)
			if err != nil {
				//return "Cannot read image file", tsPublished, imSizePublished, numSend
				publishErrMsg = "Cannot read image file"

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			tsPublished = append(tsPublished, ts)
			numSend++
			if err != nil {
				if err == io.EOF {
					break
				}
				//return "failed to send chunk via stream", tsPublished, imSizePublished, numSend
				publishErrMsg = "failed to send chunk via stream"
			}
			imSizePublished = append(imSizePublished, len(imBuf))

		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		//return "stream CloseAndRecv() error", tsPublished, imSizePublished, numSend
		publishErrMsg = "stream CloseAndRecv() error"
	}

	//return "publish success", tsPublished, imSizePublished, numSend
	fmt.Println(publishErrMsg)

}

func (pc *ProducerClient) PublishImageTestESB(client edgenode.PubSubClient, imageFilesPath string, numImagesInsert,
	frameRate uint64, imSizeParam string) error {

	var (
		imBuf []byte
	)

	var numSend uint64

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration((numImagesInsert*frameRate)+2000)*time.Second)
	defer cancel()

	// Open a stream to gRPC server
	stream, err := client.Publish(ctx)

	if err != nil {

		return err

	}
	defer stream.CloseSend()

	// Read image file names
	errMsg, files := walkAllFilesInDir(imageFilesPath)
	if errMsg != "file read success" {
		err = errors.New("File read failed")
		return err

	}

	//slices to store timestamps and image sizes published

	var i uint64

	if imSizeParam == "S" {
		for i = 0; i < numImagesInsert; i++ {
			//read first file to buffer (conversion to bytes)
			imBuf, err = ioutil.ReadFile(files[0])
			if err != nil {
				return err

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			numSend++
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

		}
	} else {

		for _, file := range files {
			if numSend == numImagesInsert {
				break
			}
			imBuf, err = ioutil.ReadFile(file)
			if err != nil {
				return err

			}
			time.Sleep(time.Duration(frameRate) * time.Millisecond)
			ts := time.Now().Format(customTimeformat)
			err = stream.Send(&edgenode.Image{
				Image:     imBuf,
				Timestamp: ts,
			})

			numSend++
			fmt.Println(numSend, len(imBuf))
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

		}
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return err
	}

	return nil
}

//***************************Helper functions*****************************************

/*
To return list of absolute paths of files in a directory
This helper function is used to read the test images from input directory
Input - directory path
Output- error message, list of filepaths
*/
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
