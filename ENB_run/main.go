package main

import (
	"os"
	"os/signal"

	"vsc_workspace/Mez_upload/broker"
)

func main() {

	//en broker address
	enb1 := broker.NewEdgeNodeBroker("cam1", "127.0.0.1:10000")
	//es broker address
	go enb1.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
