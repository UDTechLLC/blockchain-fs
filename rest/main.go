package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var httpAddr string = ":13000"

var (
	Signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP}
)

func main() {
	// http://www.bite-code.com/2015/07/22/implementing-graceful-shutdown-for-docker-containers-in-go/
	//shutdown := make(chan int)
	terminate := make(chan os.Signal, 1)

	h := NewService(httpAddr)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	//regDigest()

	log.Println("rest started successfully")

	signal.Notify(terminate, Signals...)
	//go func() {
	<-terminate
	log.Println("rest exiting")
	h.Close()
	//}()

	//<-shutdown
}
