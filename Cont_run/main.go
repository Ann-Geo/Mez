package main

import (
	"os"
	"os/signal"
	"vsc_workspace/Mez_upload/latcontroller"
)

func main() {

	cont := latcontroller.NewController("127.0.0.1:9002")
	go cont.StartLatencyController()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
