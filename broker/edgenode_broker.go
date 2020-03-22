package broker

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Ann-Geo/Mez/api/controller"
	"github.com/Ann-Geo/Mez/api/edgenode"
	"github.com/Ann-Geo/Mez/api/edgeserver"
	"github.com/Ann-Geo/Mez/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var customTimeformat string = "Monday, 02-Jan-06 15:04:05.00000 MST"

var curLatLock sync.Mutex
var currentLat string

type EdgeNodeBroker struct {
	serverName      string
	ipaddr          string
	mutex           sync.Mutex
	store           map[string]storage.Store
	stopSubcription bool
	applicationPool map[string]string
	actController   string
	storePath       string
	brokerRestart   string
	cLat            chan string
	c1              chan bool
	recoveryAddr    string
	recoveryFile    *os.File
}

func NewEdgeNodeBroker(sname, ipaddr, actController, storePath, brokerRestart string) *EdgeNodeBroker {

	return &EdgeNodeBroker{
		serverName:      sname,
		ipaddr:          ipaddr,
		mutex:           sync.Mutex{},
		store:           make(map[string]storage.Store),
		stopSubcription: false,
		applicationPool: make(map[string]string),
		actController:   actController,
		storePath:       storePath,
		brokerRestart:   brokerRestart,
		cLat:            make(chan string),
		c1:              make(chan bool),
		recoveryAddr:    "../broker/enb.txt",
	}
}

func (s *EdgeNodeBroker) StartEdgeNodeBroker(edgeServerIpaddr, login, password string) error {

	log.Println("Starting edge node broker", s.serverName)
	lis, err := net.Listen("tcp", s.ipaddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile("../cert/server.crt", "../cert/server.key")
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	}

	// Create default log storage
	s.store[s.serverName] = storage.NewMemLog(storage.SEGSIZE, storage.LOGSIZE)

	//recovery process
	if s.brokerRestart == "1" {
		recoveryFile, err := os.Open(s.recoveryAddr)
		if err != nil {
			log.Fatalln("cannot open recovery file", err)
		}
		s.recoveryFile = recoveryFile

		defer s.recoveryFile.Close()

		recoveryFileInfo, err := s.recoveryFile.Stat()
		if err != nil {
			log.Fatalln("cannot Stat recovery file", err)
		}
		if recoveryFileInfo.Size() != 0 {

			s.store[s.serverName].Recover(s.recoveryFile, "nil")

		}
	}

	//start the back up process in the background if p flag is enabled
	if s.storePath != "../../def_store/" {
		//copy path to recoveryFile
		err := os.Remove(s.recoveryAddr)
		if err != nil {
			log.Fatalf("cannot delete recovery file: %s\n", err)
		}
		recoveryFile, err := os.Create(s.recoveryAddr)
		if err != nil {
			log.Fatalf("Cannot create recovery file %v\n", err)
		}

		s.recoveryFile = recoveryFile

		defer s.recoveryFile.Close()
		fmt.Fprintf(s.recoveryFile, "%s\n", s.storePath)
		go s.store[s.serverName].Backup(s.storePath)
	}

	// Register with edge server
	err = s.regsterWithEdgeServer(edgeServerIpaddr, login, password)
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker %s\n", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds), grpc.UnaryInterceptor(s.UnaryInterceptor)}

	grpcServer := grpc.NewServer(opts...)
	// Attach client API to broker
	edgenode.RegisterPubSubServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}
	return nil
}

/************* Begin RPCs **************/
//Connect API returns id assigned by Mez
func (s *EdgeNodeBroker) Connect(ctx context.Context, url *edgenode.Url) (*edgenode.Id, error) {

	//generates the id
	id := int32(rand.Intn(100-0) + 0)

	s.applicationPool[url.GetAddress()] = string(id)

	resp := &edgenode.Id{
		Id: string(id),
	}

	return resp, nil

}

// Publish only supported by Edge node brokers
func (s *EdgeNodeBroker) Publish(stream edgenode.PubSub_PublishServer) error {

	// GRPC - Client side streaming
	numImagesRecvd := 0
	for {
		im, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&edgenode.Status{
				Status: true,
			})
		}
		if err != nil {
			return fmt.Errorf("EdgeNodeBroker %s\n", err)
		}

		// Store image and timestamp
		ts, _ := time.Parse(customTimeformat, im.GetTimestamp())

		s.store[s.serverName].Append(im.GetImage(), ts)

		numImagesRecvd++
		fmt.Println(numImagesRecvd)

	}
}

type enbWithCntlrClient struct {
	numPublished uint64
	initialLat   string
	subResChan   chan controller.CustomImage
	conn         *grpc.ClientConn
	cl           controller.LatencyControllerClient
	store        map[string]storage.Store
}

func newEnbWithCntlrClient(ipaddrCont string) *enbWithCntlrClient {
	//new client to python server
	conn, err := grpc.Dial(ipaddrCont, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Could not connect %v\n", err)
	}

	cl := controller.NewLatencyControllerClient(conn)
	return &enbWithCntlrClient{
		numPublished: 0,
		initialLat:   "1",
		subResChan:   make(chan controller.CustomImage),
		conn:         conn,
		cl:           cl,
		store:        make(map[string]storage.Store),
	}
}

// Subscribe supported by both Edge server and Edge node brokers. However their implementation is different
// Edge node subscription interacts with the controller to produce a stream that satisfies the requested latency accuracy
func (s *EdgeNodeBroker) Subscribe(imPars *edgenode.ImageStreamParameters, stream edgenode.PubSub_SubscribeServer) error {

	// GRPC - Server side streaming

	if s.serverName != imPars.Camid {
		return ErrWrongNode
	}
	if s.stopSubcription {
		s.stopSubcription = false
	}

	tstart, _ := time.Parse(customTimeformat, imPars.GetStart())
	tstop, _ := time.Parse(customTimeformat, imPars.GetStop())

	imts := make(chan storage.ImageTimestamp, 200*1024)
	errch := make(chan error)
	defer close(imts)
	defer close(errch)

	fmt.Println("here1")

	fmt.Println("here2")
	// Concurrent read of images from store got from producer
	go s.store[s.serverName].Read(imts, tstart, tstop, errch)
	imCount := 0

	errc := <-errch
	if errc != nil {
		log.Printf("EdgeNodeBroker %s\n", errc)
	}

	var lastTs storage.Timestamp

	if s.actController == "1" {

		curLatLock.Lock()
		currentLat = "1"
		lat := "1"
		curLatLock.Unlock()

		//connect the ENB client with controller server
		enbC := newEnbWithCntlrClient("0.0.0.0:9002")
		//create a new storage
		enbC.store["modified"] = storage.NewMemLog(storage.SEGSIZE, storage.LOGSIZE)

		//Get targets to send to control
		conTargets := &controller.Targets{
			TargetLat: imPars.GetLatency(),
			TargetAcc: imPars.GetAccuracy(),
		}
		//call set taget rpc of controller
		_, err := enbC.cl.SetTarget(context.Background(), conTargets)
		if err != nil {
			return err
		}

		//invoke the Control RPC
		conStream, err := enbC.cl.Control(context.Background())
		if err != nil {
			return err

		}

		waitc := make(chan struct{})
		//receiving modified images from the controller
		go func() {
			for {
				res, err := conStream.Recv()

				if err == io.EOF {
					break
				}

				if err != nil {
					numIter := 5
					for n := 0; n < numIter; n++ {
						time.Sleep(20 * time.Millisecond)
						if err == nil {
							break
						}
					}
					if err != nil {
						log.Fatalln("connection timeout: could not reach controller", err)
					}
				}

				//save the images to the log
				ts := time.Now()
				img := res.GetImage()
				tsRecord, _ := time.Parse(customTimeformat, strings.Split(res.GetAcheivedAcc(), "and")[0])
				enbC.store["modified"].Append(img, tsRecord)

				//send it to the ES broker
				modIm := &edgenode.Image{
					Image:     img,
					Timestamp: ts.Format(customTimeformat) + "and" + res.GetAcheivedAcc(),
				}

				stream.Send(modIm)

			}

			close(waitc)
		}()

		//images read from log are send to the controller
		ok := true
		for ok {
			select {
			case image := <-imts:
				{
					lastTs = image.Ts

					if len(image.Im) != 0 {
						if imCount != 0 {
							lat = <-s.cLat
						}

						req := &controller.OriginalImage{
							Image: image.Im,

							CurrentLat: (image.Ts).Format(customTimeformat) + "and" + lat,
						}

						conStream.Send(req)
						ok = true
						imCount = imCount + 1
					}

				}
			default:
				ok = false
			}
		}

		numIter := 0
		for lastTs.Before(tstop) { // More reading to be done; Poll for maxPollTime
			if s.stopSubcription {
				break
			}
			numIter++

			if numIter > maxPollTime {
				break
			}

			tstart = lastTs.Add(200 * time.Millisecond)

			go s.store[s.serverName].Read(imts, tstart, tstop, errch)

			errc := <-errch
			if errc == storage.ErrTimestampMissing {
				time.Sleep(2 * time.Millisecond)

				continue
			}

			ok = true
			for ok {
				select {
				case image := <-imts:
					{
						lastTs = image.Ts

						req := &controller.OriginalImage{
							Image: image.Im,

							CurrentLat: (image.Ts).Format(customTimeformat) + "and" + lat,
						}

						conStream.Send(req)
						ok = true
						imCount = imCount + 1
					}
				default:
					ok = false
				}
			}
			time.Sleep(2 * time.Millisecond)

		}

		imCount = 0

		conStream.CloseSend()
		<-waitc

	} else {

		ok := true
		for ok {
			select {
			case image := <-imts:
				{
					lastTs = image.Ts
					if err := stream.Send(&edgenode.Image{
						Image:     image.Im,
						Timestamp: (image.Ts).Format(customTimeformat),
					}); err != nil {
						return fmt.Errorf("EdgeNodeBroker %s\n", err)
					}
					fmt.Println("sending 1.1.1", (image.Ts).Format(customTimeformat), time.Now())
					ok = true
				}
			default:
				ok = false
			}
		}

		numIter := 0
		for lastTs.Before(tstop) { // More reading to be done; Poll for maxPollTime

			if s.stopSubcription {
				break
			}
			numIter++

			if numIter > maxPollTime {
				break
			}

			tstart = lastTs.Add(200 * time.Millisecond)

			go s.store[s.serverName].Read(imts, tstart, tstop, errch)

			errc := <-errch

			if errc == storage.ErrTimestampMissing {
				time.Sleep(1 * time.Microsecond)
				continue
			}

			ok = true
			for ok {
				select {
				case image := <-imts:
					{
						lastTs = image.Ts

						if err := stream.Send(&edgenode.Image{
							Image:     image.Im,
							Timestamp: (image.Ts).Format(customTimeformat),
						}); err != nil {
							return fmt.Errorf("EdgeNodeBroker %s\n", err)
						}

						ok = true
					}
				default:
					ok = false
				}
			}
			time.Sleep(1 * time.Microsecond)
		}

	}

	return nil
}

func (s *EdgeNodeBroker) Unsubscribe(ctx context.Context, caminfo *edgenode.CameraInfo) (*edgenode.Status, error) {

	s.stopSubcription = true

	return &edgenode.Status{
		Status: true,
	}, nil
}

func (s *EdgeNodeBroker) LatencyCalc(ctx context.Context, lat *edgenode.LatencyMeasured) (*edgenode.Status, error) {

	s.cLat <- lat.GetCurrentLat()

	currentLat = lat.GetCurrentLat()

	status := &edgenode.Status{
		Status: true,
	}
	return status, nil
}

/************* End RPCs **************/

/**************Helpers **************/

// authenticateAgent check the client credentials
func (s *EdgeNodeBroker) authenticateClient(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientLogin != "client" {
			return "", fmt.Errorf("EdgeNodeBroker: unknown user %s\n", clientLogin)
		}
		if clientPassword != "edge" {
			return "", fmt.Errorf("EdgeNodeBroker: bad password %s\n", clientPassword)
		}
		return "42", nil
	}
	return "", fmt.Errorf("EdgeNodeBroker: missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func (s *EdgeNodeBroker) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	clientID, err := s.authenticateClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func (s *EdgeNodeBroker) regsterWithEdgeServer(edgeServerIpaddr, login, password string) error {
	en := NewEdgeNodeClient(login, password) //username and password

	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker: could not load tls cert: %s\n", err)
	}

	// Connect to edge server
	conn, err := grpc.Dial(edgeServerIpaddr, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&en.Auth))
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker did not connect: %s\n", err)
	}
	defer conn.Close()
	cl := edgeserver.NewPubSubClient(conn)

	err = en.Register(s.ipaddr, s.serverName, cl)
	if err != nil {
		return fmt.Errorf("EdgeNodeBroker %s\n", err)
	}
	return nil

}
