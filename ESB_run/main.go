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

	//es broker ip address
	esb := broker.NewEdgeServerBroker("EdgeServer", "127.0.0.1:20000", actController)
	go esb.StartEdgeServerBroker()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
