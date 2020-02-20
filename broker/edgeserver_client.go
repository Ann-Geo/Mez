// Edge server client - edge server is a client of edge node subscribing images
package broker

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"vsc_workspace/Mez_upload/api/edgenode"
	"vsc_workspace/Mez_upload/api/edgeserver"
)

type EdgeServerClient struct {
	Auth Authentication
}

type EdgeServerClientWithControl struct {
	Auth Authentication
}

func NewEdgeServerClient(login, password string) *EdgeServerClient {
	return &EdgeServerClient{Auth: Authentication{login: login, password: password}}
}

func NewEdgeServerClientWithControl(login, password string) *EdgeServerClientWithControl {
	return &EdgeServerClientWithControl{Auth: Authentication{login: login, password: password}}
}

func (cc *EdgeServerClient) SubscribeImage(s *EdgeServerBroker, client edgenode.PubSubClient, impars *edgeserver.ImageStreamParameters) error {

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	// Convert impars to edge node api image parameters
	imEdgeNodePars := &edgenode.ImageStreamParameters{Camid: impars.Camid, Latency: impars.Latency, Accuracy: impars.Accuracy,
		Start: impars.Start, Stop: impars.Stop}

	stream, err := client.Subscribe(ctx, imEdgeNodePars)
	if err != nil {
		return err
	}

	//numImagesRecvd := 0
	for {
		im, err := stream.Recv()
		if err == io.EOF {
			break
		}
		//break in channel, time out
		if err != nil {
			return fmt.Errorf("Edgeserver client: failed to receive image via stream from edge node %v", err)
		}

		// Store image and timestamp got from edgenode
		ts, _ := time.Parse(customTimeformat, im.GetTimestamp())
		s.store[impars.Camid].Append(im.GetImage(), ts)
		//numImagesRecvd++
		//log.Printf("EdgeServerClient: Number of images received %d, of size %d time %s", numImagesRecvd, len(im.Image), ts)

		s.mutex.Lock()

		if s.stopSubcription[impars.Appid+impars.Camid] { // If application unsubscribes, then decrease ref count
			s.numbSubscribers[impars.Camid]--
			if s.numbSubscribers[impars.Camid] == 0 { // no more apps subscribing to edge node
				s.mutex.Unlock()
				camidlist := []string{impars.Camid}
				camInfo := &edgenode.CameraInfo{Camid: camidlist}
				_, err := client.Unsubscribe(ctx, camInfo)
				if err != nil {
					return ErrUnsubscribe
				}
			} else {
				s.mutex.Unlock()
			}
		} else {
			s.mutex.Unlock()
		}

	}
	return nil
}

func (cc *EdgeServerClientWithControl) SubscribeImage(s *EdgeServerBroker, client edgenode.PubSubClient, impars *edgeserver.ImageStreamParameters) error {

	resultFile, err := os.Create("nw_lat_1st_sub.txt")
	if err != nil {
		log.Fatalf("Cannot create result file %v\n", err)
	}

	defer resultFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	// Convert impars to edge node api image parameters
	imEdgeNodePars := &edgenode.ImageStreamParameters{Camid: impars.Camid, Latency: impars.Latency, Accuracy: impars.Accuracy,
		Start: impars.Start, Stop: impars.Stop}

	stream, err := client.Subscribe(ctx, imEdgeNodePars)
	if err != nil {
		return err
	}

	numImagesRecvd := 0
	for {
		im, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("returning from edge server client")
			break
		}
		//break in channel, time out
		if err != nil {
			return fmt.Errorf("Edgeserver client: failed to receive image via stream from edge node %v", err)
		}

		//Sending image transfer latency to Edge node
		tsRcvd := time.Now()
		tsSendAndtsRecAndAccu := im.GetTimestamp()
		fmt.Println("tsSendAndtsRecAndAccu", tsSendAndtsRecAndAccu)
		tsSendAndtsRecAndAccuSlice := strings.Split(tsSendAndtsRecAndAccu, "and")
		tsSend, _ := time.Parse(customTimeformat, tsSendAndtsRecAndAccuSlice[0])
		imLatency := tsRcvd.Sub(tsSend)

		fmt.Println("latency=", imLatency)

		go sendMeasuredLatency(client, imLatency)

		imLen := len(im.GetImage())
		achAcc := tsSendAndtsRecAndAccuSlice[2]
		fmt.Fprintf(resultFile, "current time: %s, image_size: %d, latency: %s, accuracy: %s\n", tsRcvd, imLen, imLatency, achAcc)

		log.Printf("Response from Subscribe: current time: %s, image_size: %d, latency: %s, accuracy: %s\n", tsRcvd, imLen, imLatency, achAcc)

		// Store image and timestamp got from edgenode
		//ts, _ := time.Parse(customTimeformat, im.GetTimestamp())
		tsRec, _ := time.Parse(customTimeformat, tsSendAndtsRecAndAccuSlice[1])
		s.store[impars.Camid].Append(im.GetImage(), tsRec)
		numImagesRecvd++
		log.Printf("EdgeServerClient: Number of images received %d, of size %d time %s", numImagesRecvd, len(im.Image), tsRec)

		s.mutex.Lock()

		if s.stopSubcription[impars.Appid+impars.Camid] { // If application unsubscribes, then decrease ref count
			s.numbSubscribers[impars.Camid]--
			if s.numbSubscribers[impars.Camid] == 0 { // no more apps subscribing to edge node
				s.mutex.Unlock()
				camidlist := []string{impars.Camid}
				camInfo := &edgenode.CameraInfo{Camid: camidlist}
				_, err := client.Unsubscribe(ctx, camInfo)
				if err != nil {
					return ErrUnsubscribe
				}
			} else {
				s.mutex.Unlock()
			}
		} else {
			s.mutex.Unlock()
		}

	}

	return nil
}

/**********************************helper funcs***********************************/
func sendMeasuredLatency(client edgenode.PubSubClient, imLatency time.Duration) {

	imLatstr := imLatency.String()

	re := regexp.MustCompile("[0-9]+")
	integerSlice := re.FindAllString(imLatstr, -1)
	imLatstrFloatstr := integerSlice[0] + "." + integerSlice[1]
	var imLatstrFloat float64
	if s, err := strconv.ParseFloat(imLatstrFloatstr, 64); err == nil {
		imLatstrFloat = s
	}

	if strings.Contains(imLatstr, "µs") {
		imLatstrFloat = imLatstrFloat / 1000
	} else if strings.Contains(imLatstr, "ms") {
		imLatstrFloat = imLatstrFloat * 1
	} else {
		imLatstrFloat = imLatstrFloat * 1000
	}

	imLatstrFloatstrCorrectFmt := fmt.Sprintf("%f", imLatstrFloat)

	lat := &edgenode.LatencyMeasured{
		CurrentLat: imLatstrFloatstrCorrectFmt,
	}

	_, err := client.LatencyCalc(context.Background(), lat)
	if err != nil {
		log.Fatalf("error while calling LatencyCalc RPC: %v", err)
	}
	//log.Printf("Response from LatencyCalc: %v\n", status.GetStatus())
}
