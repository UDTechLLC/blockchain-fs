package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener
}

// New returns an uninitialized HTTP service.
func NewService(addr string) *Service {
	return &Service{
		addr: addr,
	}
}

// Start starts the service.
func (s *Service) Start() error {
	// Get the mux router object
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/", Home)
	router.HandleFunc("/bucket", Bucket).Methods("GET")

	// Create a negroni instance
	n := negroni.Classic()
	n.UseHandler(router)

	server := http.Server{
		Handler: n,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

// Close closes the service.
func (s *Service) Close() {
	s.ln.Close()
	return
}
