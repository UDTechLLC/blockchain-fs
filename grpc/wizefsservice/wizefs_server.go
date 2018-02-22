package wizefsservice

import (
	"fmt"
	"os/exec"
	"syscall"

	"golang.org/x/net/context"

	api "bitbucket.org/udt/wizefs/internal/command"
	_ "bitbucket.org/udt/wizefs/internal/tlog"
)

type wizefsServer struct {
}

func (s *wizefsServer) Create(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	//tlog.Info.Printf("Create method...")
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiCreate(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Delete(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	//tlog.Info.Printf("Delete method...")
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiDelete(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Mount(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	//tlog.Info.Printf("Mount method...")
	origin := request.GetOrigin()
	var message string = "OK"
	response = &FilesystemResponse{
		Executed: true,
		Message:  message,
	}

	c := exec.Command("../mount/mount", origin)
	cerr := c.Start()
	if cerr != nil {
		message = fmt.Sprintf("starting command failed: %v", cerr)
	}
	//tlog.Info.Printf("starting command...")
	cerr = c.Wait()
	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				response.Executed = true
				response.Message = fmt.Sprintf("wait returned an exit status: %d", waitstat.ExitStatus())
			}
		} else {
			response.Executed = false
			response.Message = fmt.Sprintf("wait returned an unknown error: %v", cerr)
		}
	}
	//tlog.Info.Printf("ending command...")

	return
}

func (s *wizefsServer) Unmount(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	//tlog.Info.Printf("Unmount method...")
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiUnmount(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Put(ctx context.Context, request *PutRequest) (response *PutResponse, err error) {
	//tlog.Info.Printf("Put method...")
	filename := request.GetFilename()
	content := request.GetContent()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &PutResponse{
		Executed: true,
		Message:  "OK",
	}
	if err = api.ApiPut(filename, origin, content); err != nil {
		response.Executed = false
		response.Message = err.Error()
	}
	return
}

func (s *wizefsServer) Get(ctx context.Context, request *GetRequest) (response *GetResponse, err error) {
	//tlog.Info.Printf("Get method...")
	filename := request.GetFilename()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &GetResponse{
		Executed: true,
		Message:  "OK",
		Content:  nil,
	}
	if content, err := api.ApiGet(filename, origin, true); err != nil {
		response.Executed = false
		response.Message = err.Error()
	} else {
		response.Content = content
	}
	return
}

func NewServer() *wizefsServer {
	s := &wizefsServer{}
	return s
}
