package main

import (
	"os"
	"os/signal"
	"strings"

	"vsc_workspace/Mez_upload_woa/broker"
)

func main() {

	var actController string = "0"
	arguments := os.Args
	for i := 0; i < len(arguments); i++ {
		if strings.Compare(arguments[i], "-c") == 0 {
			if strings.Compare(arguments[i+1], "1") == 0 {
				actController = "1"
				break
			}
		}
	}

	//en broker address
	enb1 := broker.NewEdgeNodeBroker("cam1", "127.0.0.1:10000", actController)
	//es broker address
	go enb1.StartEdgeNodeBroker("127.0.0.1:20000", "client", "edge")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
