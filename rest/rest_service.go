package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	http.Handle("/", s)

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

// ServeHTTP allows Service to serve HTTP requests.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/bucket") {
		s.handleBucketRequest(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) handleBucketRequest(w http.ResponseWriter, r *http.Request) {
	//w.WriteHeader(http.StatusOK)
	k := "hello"
	v := "rest"

	b, err := json.Marshal(map[string]string{k: v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(b))

	return
}
