package main

import (
	"log"
	"os"
	"os/signal"
)

var httpAddr string = ":13000"

func main() {
	h := NewService(httpAddr)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	log.Println("rest started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("rest exiting")

	h.Close()
}
