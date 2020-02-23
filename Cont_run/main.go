package main

import (
	"os"
	"os/signal"
	"vsc_workspace/Mez_upload_woa/latcontroller"
)

func main() {

	cont := latcontroller.NewController("127.0.0.1:9002", "/home/research/goworkspace/src/vsc_workspace/")
	go cont.StartLatencyController()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

}
