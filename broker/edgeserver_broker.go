package broker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Ann-Geo/Mez/api/edgenode"
	"github.com/Ann-Geo/Mez/api/edgeserver"
	"github.com/Ann-Geo/Mez/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type EdgeServerBroker struct {
	serverName      string
	ipaddr          string
	mutex           sync.Mutex
	store           map[string]storage.Store
	nodeInfoMap     map[string]string // key is cameraid
	numbSubscribers map[string]int    //key is cameraid
	stopSubcription map[string]bool   // key is appid + cameraid
	appMutex        sync.Mutex
	applicationPool map[string]string
	actController   string
}

func NewEdgeServerBroker(sname, ipaddr, actController string) *EdgeServerBroker {
	return &EdgeServerBroker{
		serverName:      sname,
		ipaddr:          ipaddr,
		mutex:           sync.Mutex{},
		store:           make(map[string]storage.Store),
		nodeInfoMap:     make(map[string]string),
		numbSubscribers: make(map[string]int),
		stopSubcription: make(map[string]bool),
		appMutex:        sync.Mutex{},
		applicationPool: make(map[string]string),
		actController:   actController,
	}
}

func (s *EdgeServerBroker) StartEdgeServerBroker() {
	log.Println("Starting edge server broker", s.serverName)
	lis, err := net.Listen("tcp", s.ipaddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile("../cert/server.crt", "../cert/server.key")
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds), grpc.UnaryInterceptor(s.UnaryInterceptor)}

	grpcServer := grpc.NewServer(opts...)
	// Attach client API to broker
	edgeserver.RegisterPubSubServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}

/************* Begin RPCs **************/
//Connect API returns id assigned by Mez
func (s *EdgeServerBroker) Connect(ctx context.Context, url *edgeserver.Url) (*edgeserver.Id, error) {

	//generates the id
	id := int32(rand.Intn(100-0) + 0)

	s.appMutex.Lock()
	s.applicationPool[url.GetAddress()] = string(id)
	s.appMutex.Unlock()

	resp := &edgeserver.Id{
		Id: string(id),
	}

	return resp, nil

}

// Called by Edge node
func (s *EdgeServerBroker) Register(ctx context.Context, nodeinfo *edgeserver.NodeInfo) (*edgeserver.Status, error) {
	//fmt.Println("invoked")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nodeInfoMap[nodeinfo.Camid] = nodeinfo.Ipaddr

	_, ok := s.nodeInfoMap[nodeinfo.Camid]

	if !ok {
		return &edgeserver.Status{
			Status: false,
		}, ErrRegistered

	}

	// Create storage for edge node at edgeserver
	s.store[nodeinfo.Camid] = storage.NewMemLog(storage.SEGSIZE, storage.LOGSIZE)

	//fmt.Println("registered")
	return &edgeserver.Status{
		Status: true,
	}, nil
}

// Called by Edge node
func (s *EdgeServerBroker) Unregister(ctx context.Context, nodeinfo *edgeserver.NodeInfo) (*edgeserver.Status, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.nodeInfoMap[nodeinfo.Camid]

	if !ok {
		return &edgeserver.Status{
			Status: false,
		}, ErrNotRegistered

	}

	if s.numbSubscribers[nodeinfo.Camid] == 0 { // No application is subscribing
		delete(s.nodeInfoMap, nodeinfo.Camid)

		// Delete storage for edge node at edge server
		delete(s.store, nodeinfo.Camid)
	}

	return &edgeserver.Status{
		Status: true,
	}, nil

}

// Called by consumer application
func (s *EdgeServerBroker) GetCameraInfo(ctx context.Context, campars *edgeserver.CameraParameters) (*edgeserver.CameraInfo, error) {
	// Camera parameters are empty in this implementation. Ideally, only edge nodes that satisfy the specified parameters will be returned
	s.mutex.Lock()
	defer s.mutex.Unlock()

	camlist := make([]string, 0)
	for k, _ := range s.nodeInfoMap {
		camlist = append(camlist, k)
	}
	return &edgeserver.CameraInfo{
		Camid: camlist,
	}, nil

}

// Called by consumer application
func (s *EdgeServerBroker) Subscribe(impars *edgeserver.ImageStreamParameters, stream edgeserver.PubSub_SubscribeServer) error {

	s.mutex.Lock()

	_, pres := s.nodeInfoMap[impars.Camid]
	if !pres {
		s.mutex.Unlock()
		return ErrNotRegistered
	}

	//fmt.Println(impars.Camid)

	// Inrement number of subscribers to camid
	s.numbSubscribers[impars.Camid]++
	//fmt.Println(s.numbSubscribers[impars.Camid])

	s.stopSubcription[impars.Appid+impars.Camid] = false

	s.mutex.Unlock()

	// Subscribe images from edge node - concurrent
	errchsub := make(chan error)
	defer close(errchsub)
	c := make(chan bool)

	//first check to see anyone subscribing to that camera

	//fmt.Println("num sub", s.numbSubscribers[impars.Camid])

	if s.numbSubscribers[impars.Camid] == 1 {
		//if no one then go subscribe
		//fmt.Println("Sub from EN")
		go s.subscribeFromEdgenode(impars, c) //,errchsub)

		//errsub := <-errchsub
		//fmt.Println(errsub)
		//if errsub != nil {
		//log.Println(errsub)
		//}
		<-c
	}

	//fmt.Println("start reading")
	// Serve image to consumer application

	tstart, _ := time.Parse(customTimeformat, impars.Start)
	tstop, _ := time.Parse(customTimeformat, impars.Stop)

	imts := make(chan storage.ImageTimestamp, 200*1024)
	errch := make(chan error)
	defer close(imts)
	defer close(errch)

	// Concurrent read of images from store got from edge node
	//fmt.Println("Read from ES log")

	go s.store[impars.Camid].Read(imts, tstart, tstop, errch)

	errc := <-errch
	if errc != nil {
		log.Printf("EdgeNodeBroker %s\n", errc)
	}

	var lastTs storage.Timestamp

	ok := true
	for ok {
		select {
		case image := <-imts:
			{
				//lastTs = image.Ts
				if err := stream.Send(&edgeserver.Image{
					Image:     image.Im,
					Timestamp: (image.Ts).Format(customTimeformat),
				}); err != nil {
					return fmt.Errorf("EdgeServerBroker %s\n", err)
				}
				//fmt.Println("Image sent to subscriber")
				ok = true
			}
		default:

			ok = false
		}
	}

	numIter := 0
	for lastTs.Before(tstop) { // More reading to be done; Poll
		if s.stopSubcription[impars.Appid+impars.Camid] { // From Unsubscribe API
			break
		}
		numIter++
		if numIter > maxPollTime {
			break
		}

		tstart = lastTs.Add(200 * time.Millisecond) // TODO: Hacky - Read from 1 second later timestamp
		//fmt.Println("here1111111111111111111")

		go s.store[impars.Camid].Read(imts, tstart, tstop, errch)
		errc := <-errch
		if errc == storage.ErrTimestampMissing {
			time.Sleep(1 * time.Microsecond) // TODO: Hacky - Sleep for a second
			//fmt.Println("here22222222222222222")
			continue
		}
		ok = true
		for ok {
			select {
			case image := <-imts:
				{
					lastTs = image.Ts
					if err := stream.Send(&edgeserver.Image{
						Image:     image.Im,
						Timestamp: (image.Ts).Format(customTimeformat),
					}); err != nil {
						return fmt.Errorf("EdgeServerBroker %s\n", err)
					}
					ok = true
				}
			default:
				ok = false
			}
		}

		time.Sleep(1 * time.Microsecond) // Sleep for a second
		//fmt.Println("here333333333333333333333")

	}

	//fmt.Println("returning from subscribe")

	return nil

}

func (s *EdgeServerBroker) Unsubscribe(ctx context.Context, appinfo *edgeserver.AppInfo) (*edgeserver.Status, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stopSubcription[appinfo.Appid+appinfo.Camid] = true
	s.numbSubscribers[appinfo.Camid]--
	return &edgeserver.Status{
		Status: true,
	}, nil
}

/**************Helpers **************/

// authenticateAgent check the client credentials
func (s *EdgeServerBroker) authenticateClient(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientLogin != "client" {
			return "", fmt.Errorf("unknown user %s", clientLogin)
		}
		if clientPassword != "edge" {
			return "", fmt.Errorf("bad password %s", clientPassword)
		}
		log.Printf("authenticated client: %s", clientLogin)
		return "42", nil
	}
	return "", fmt.Errorf("missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func (s *EdgeServerBroker) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	clientID, err := s.authenticateClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

/*
func (s *EdgeServerBroker) subscribeFromEdgenode(impars *edgeserver.ImageStreamParameters, errch chan<- error) {
	// TO DO: edge node username and password are hardcoded

	if s.actController == "0" {

		cons := NewEdgeServerClient("client", "edge") //username and password

		// Connect to edge nodebroker
		conn, err := grpc.Dial(s.nodeInfoMap[impars.Camid], grpc.WithInsecure())
		if err != nil {
			errch <- ErrConnect
			//return fmt.Errorf("Edgeserver: could did not connect to edgenode: %s", err)
		}
		defer conn.Close()
		cl := edgenode.NewPubSubClient(conn)

		err = cons.SubscribeImage(s, cl, impars)
		if err != nil {
			errch <- err
		}
		errch <- nil

	} else {
		cons := NewEdgeServerClientWithControl("client", "edge") //username and password
		//cons := NewEdgeServerClient("client", "edge") //username and password

		// Connect to edge nodebroker
		fmt.Println(s.nodeInfoMap[impars.Camid])
		conn, err := grpc.Dial(s.nodeInfoMap[impars.Camid], grpc.WithInsecure())
		if err != nil {
			fmt.Println("could not connect with Edge node broker")
			errch <- ErrConnect
			//return fmt.Errorf("Edgeserver: could did not connect to edgenode: %s", err)
		}
		defer conn.Close()
		cl := edgenode.NewPubSubClient(conn)

		err = cons.SubscribeImage(s, cl, impars)
		if err != nil {
			errch <- err
		}
		errch <- nil
	}

}
*/

func (s *EdgeServerBroker) subscribeFromEdgenode(impars *edgeserver.ImageStreamParameters, c chan<- bool) {
	// TO DO: edge node username and password are hardcoded

	if s.actController == "0" {

		cons := NewEdgeServerClient("client", "edge") //username and password

		creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
		if err != nil {
			log.Fatalf("could not load tls cert: %s", err)
		}

		// Connect to edge nodebroker
		conn, err := grpc.Dial(s.nodeInfoMap[impars.Camid], grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&cons.Auth))
		if err != nil {
			log.Fatalf("could not connect with edge node")
			//return fmt.Errorf("Edgeserver: could did not connect to edgenode: %s", err)
		}
		defer conn.Close()
		cl := edgenode.NewPubSubClient(conn)

		err = cons.SubscribeImage(s, cl, impars, c)
		if err != nil {
			log.Fatalf("error while subscribing edge node")
		}

	} else {
		cons := NewEdgeServerClientWithControl("client", "edge") //username and password
		//cons := NewEdgeServerClient("client", "edge") //username and password

		creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
		if err != nil {
			log.Fatalf("could not load tls cert: %s", err)
		}

		// Connect to edge nodebroker
		//fmt.Println(s.nodeInfoMap[impars.Camid])
		conn, err := grpc.Dial(s.nodeInfoMap[impars.Camid], grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&cons.Auth))
		if err != nil {
			//fmt.Println("could not connect with Edge node broker")
			log.Fatalf("could not connect with edge node")
			//return fmt.Errorf("Edgeserver: could did not connect to edgenode: %s", err)
		}
		defer conn.Close()
		cl := edgenode.NewPubSubClient(conn)

		err = cons.SubscribeImage(s, cl, impars, c)
		if err != nil {
			log.Fatalf("error while subscribing edge node")
		}

	}

}
