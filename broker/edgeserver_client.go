// Edge server client - edge server is a client of edge node subscribing images
package broker

import (
        "github.com/arun-ravindran/Raven/api/edgenode"
        "github.com/arun-ravindran/Raven/api/edgeserver"
	"context"
	"fmt"
	"io"
	"log"
	"time"
)

type EdgeServerClient struct {
	Auth Authentication
}

func NewEdgeServerClient(login, password string) *EdgeServerClient {
	return &EdgeServerClient{Auth: Authentication{login: login, password: password}}
}

func (cc *EdgeServerClient) SubscribeImage(s *EdgeServerBroker, client edgenode.PubSubClient, impars *edgeserver.ImageStreamParameters) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			break
		}
		if err != nil {
			return fmt.Errorf("Edgeserver client: failed to receive image via stream from edge node %v", err)
		}

		// Store image and timestamp got from edgenode
		ts, _ := time.Parse(time.RFC850, im.GetTimestamp())
		s.store[impars.Camid].Append(im.GetImage(), ts)
		numImagesRecvd++
		log.Printf("EdgeServerClient: Number of images received %d, of size %d time %s", numImagesRecvd, len(im.Image), ts)

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
