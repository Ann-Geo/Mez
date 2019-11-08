package main

import (
	"os"
	"os/signal"

	"github.com/arun-ravindran/Raven/broker"
)

func main() {

	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000")
	go esb.StartEdgeServerBroker()

	enb := broker.NewEdgeNodeBroker("cam1", "127.0.0.1:10000")
	go enb.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
