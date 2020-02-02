package main

import (
	"os"
	"os/signal"

	"vsc_workspace/Mez_upload/broker"
)

func main() {

	//es broker ip address
	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000")
	go esb.StartEdgeServerBroker()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
