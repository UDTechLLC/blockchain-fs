package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "bitbucket.org/udt/wizefs/grpc/wizefsservice"
	api "bitbucket.org/udt/wizefs/internal/command"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type wizefsServer struct {
}

func (s *wizefsServer) Create(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	origin := request.GetOrigin()

	response = &pb.FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiCreate(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Delete(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	origin := request.GetOrigin()

	response = &pb.FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiDelete(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Mount(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	return nil, nil
}

func (s *wizefsServer) Unmount(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	origin := request.GetOrigin()

	response = &pb.FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiUnmount(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Put(ctx context.Context, request *pb.PutRequest) (*pb.PutResponse, error) {
	return nil, nil
}

func (s *wizefsServer) Get(ctx context.Context, request *pb.GetRequest) (*pb.GetResponse, error) {
	return nil, nil
}

func newServer() *wizefsServer {
	s := &wizefsServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWizeFsServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
