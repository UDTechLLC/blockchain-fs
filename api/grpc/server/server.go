package main

import (
	"flag"
	"fmt"
	"net"

	"google.golang.org/grpc"

	pb "bitbucket.org/udt/wizefs/api/grpc/wizefsservice"
	"bitbucket.org/udt/wizefs/core/tlog"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		tlog.Fatal.Printf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWizeFsServiceServer(grpcServer, pb.NewServer())
	grpcServer.Serve(lis)
}
