package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"syscall"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "bitbucket.org/udt/wizefs/grpc/wizefsservice"
	api "bitbucket.org/udt/wizefs/internal/command"
	"bitbucket.org/udt/wizefs/internal/tlog"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type wizefsServer struct {
}

func (s *wizefsServer) Create(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	tlog.Info.Printf("Create method...")
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
	tlog.Info.Printf("Delete method...")
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
	tlog.Info.Printf("Mount method...")
	origin := request.GetOrigin()
	var message string = "OK"

	c := exec.Command("../mount/mount", origin)
	cerr := c.Start()
	if cerr != nil {
		message = fmt.Sprintf("starting command failed: %v", cerr)
	}
	tlog.Info.Printf("starting command...")
	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				message = fmt.Sprintf("wait returned an exit status: %d", waitstat.ExitStatus())
			}
		}
		message = fmt.Sprintf("wait returned an unknown error: %v", cerr)
	}
	tlog.Info.Printf("ending command...")

	return &pb.FilesystemResponse{
		Executed: true,
		Message:  message,
	}, nil
}

func (s *wizefsServer) Unmount(ctx context.Context, request *pb.FilesystemRequest) (response *pb.FilesystemResponse, err error) {
	tlog.Info.Printf("Unmount method...")
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

func (s *wizefsServer) Put(ctx context.Context, request *pb.PutRequest) (response *pb.PutResponse, err error) {
	tlog.Info.Printf("Put method...")
	filename := request.GetFilename()
	content := request.GetContent()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &pb.PutResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiPut(filename, origin, content); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
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
		tlog.Fatal.Printf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWizeFsServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
