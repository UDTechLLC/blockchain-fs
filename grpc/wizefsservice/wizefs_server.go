package wizefsservice

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	api "bitbucket.org/udt/wizefs/internal/command"
	_ "bitbucket.org/udt/wizefs/internal/tlog"
)

const (
	packagePath = "grpc"
	mountApp    = "cmd/wizefs_mount/wizefs_mount"
)

var (
	projectPath = getProjectPath()
)

func getProjectPath() string {
	_, testFilename, _, _ := runtime.Caller(0)
	idx := strings.Index(testFilename, packagePath)
	return testFilename[0:idx]
}

type wizefsServer struct {
}

func (s *wizefsServer) Create(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	origin := request.GetOrigin()
	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}

	if exitCode, err := api.ApiCreate(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
		response.Message += " ExitCode: " + strconv.Itoa(exitCode)
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
	if exitCode, err := api.ApiDelete(origin); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
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

	appPath := projectPath + mountApp
	c := exec.Command(appPath, origin)
	cerr := c.Start()
	if cerr != nil {
		message = fmt.Sprintf("starting command failed: %v", cerr)
	} else {
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
	}

	return
}

func (s *wizefsServer) Unmount(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	//tlog.Info.Printf("Unmount method...")
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if exitCode, err := api.ApiUnmount(origin); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
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
	if exitCode, err := api.ApiPut(filename, origin, content); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
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
	if content, exitCode, err := api.ApiGet(filename, origin, "", true); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	} else {
		response.Content = content
	}
	return
}

func NewServer() *wizefsServer {
	s := &wizefsServer{}
	return s
}
