package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"

	"bitbucket.org/udt/wizefs/internal/core"
	"bitbucket.org/udt/wizefs/rest/controllers"
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

	router.HandleFunc("/", controllers.Home)

	// curl -X POST localhost:13000/buckets -d '{"data":{"origin":"REST1"}}'
	router.HandleFunc("/buckets", controllers.CreateBucket).Methods("POST")
	// curl -X DELETE localhost:13000/buckets/REST1
	router.HandleFunc("/buckets/{origin}", controllers.DeleteBucket).Methods("DELETE")
	// curl -X POST localhost:13000/buckets/REST1/mount
	router.HandleFunc("/buckets/{origin}/mount", controllers.MountBucket).Methods("POST")
	// curl -X POST localhost:13000/buckets/REST1/unmount
	router.HandleFunc("/buckets/{origin}/unmount", controllers.UnmountBucket).Methods("POST")

	// curl -F "filename=@/home/sergey/test.txt" -X POST localhost:13000/buckets/REST1/putfile
	router.HandleFunc("/buckets/{origin}/putfile", controllers.PutFile).Methods("POST")
	// curl -X POST localhost:13000/buckets/REST1/put -d '{"data":{"name":"...","content":"..."}}'
	router.HandleFunc("/buckets/{origin}/put", controllers.Put).Methods("POST")
	// curl -X GET localhost:13000/buckets/REST1/files/test.txt --output test.txt
	router.HandleFunc("/buckets/{origin}/files/{filename}", controllers.GetFile).Methods("GET")
	// curl -X DELETE localhost:13000/buckets/REST1/files/test.txt
	router.HandleFunc("/buckets/{origin}/files/{filename}", controllers.RemoveFile).Methods("DELETE")

	//corsHandler := cors.Default().Handler(router)
	c := cors.New(cors.Options{
		AllowedMethods:    []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"},
	})

	// Create a negroni instance
	n := negroni.Classic()
	n.Use(c)
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
			log.Printf("HTTP serve: %s", err)
		}
		//shutdown <- 1
	}()

	return nil
}

// Close closes the service.
func (s *Service) Close() {
	log.Println("rest closing")

	storage := core.NewStorage()
	for origin, bucket := range storage.MountedBuckets() {
		log.Printf("Unmounting Bucket: %s [%s]", origin, bucket.MountPoint)
		// Unmount a Bucket
		if exitCode, err := storage.Unmount(origin); err != nil {
			log.Printf("Error: %s Exit code: %d", err.Error(), exitCode)
		}
	}

	s.ln.Close()
	return
}
