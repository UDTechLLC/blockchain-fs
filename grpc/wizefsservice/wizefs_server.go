package wizefsservice

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"bitbucket.org/udt/wizefs/internal/core"
	"bitbucket.org/udt/wizefs/internal/tlog"
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
	storage *core.Storage
}

func NewServer() *wizefsServer {
	s := &wizefsServer{
		storage: core.NewStorage(),
	}
	return s
}

func (s *wizefsServer) Create(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	origin := request.GetOrigin()
	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}

	if exitCode, err := s.storage.Create(origin); err != nil {
		response.Executed = false
		response.Message = err.Error()
		response.Message += " ExitCode: " + strconv.Itoa(exitCode)
	}
	return
}

func (s *wizefsServer) Delete(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if exitCode, err := s.storage.Delete(origin); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	}
	return
}

func (s *wizefsServer) Mount(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	origin := request.GetOrigin()
	var message string = "OK"
	response = &FilesystemResponse{
		Executed: true,
		Message:  message,
	}

	appPath := projectPath + mountApp
	tlog.Info.Println("appPath:", appPath)
	c := exec.Command(appPath, origin)
	cerr := c.Start()
	if cerr != nil {
		message = fmt.Sprintf("starting command failed: %v", cerr)
	} else {
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
	}

	return
}

func (s *wizefsServer) Unmount(ctx context.Context, request *FilesystemRequest) (response *FilesystemResponse, err error) {
	origin := request.GetOrigin()

	response = &FilesystemResponse{
		Executed: true,
		Message:  "OK",
	}
	if exitCode, err := s.storage.Unmount(origin); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	}
	return
}

func (s *wizefsServer) Put(ctx context.Context, request *PutRequest) (response *PutResponse, err error) {
	filename := request.GetFilename()
	content := request.GetContent()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &PutResponse{
		Executed: true,
		Message:  "OK",
	}
	bucket, ok := s.storage.Bucket(origin)
	if !ok {
		response.Executed = false
		response.Message = fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin)
		return
	}
	if exitCode, err := bucket.PutFile(filename, content); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	}
	return
}

func (s *wizefsServer) Get(ctx context.Context, request *GetRequest) (response *GetResponse, err error) {
	filename := request.GetFilename()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &GetResponse{
		Executed: true,
		Message:  "OK",
		Content:  nil,
	}
	bucket, ok := s.storage.Bucket(origin)
	if !ok {
		response.Executed = false
		response.Message = fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin)
		return
	}
	if content, exitCode, err := bucket.GetFile(filename, "", true); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	} else {
		response.Content = content
	}
	return
}

func (s *wizefsServer) Remove(ctx context.Context, request *RemoveRequest) (response *RemoveResponse, err error) {
	filename := request.GetFilename()
	origin := request.GetOrigin()

	// TODO: check all request's data

	response = &RemoveResponse{
		Executed: true,
		Message:  "OK",
	}
	bucket, ok := s.storage.Bucket(origin)
	if !ok {
		response.Executed = false
		response.Message = fmt.Sprintf("Bucket with ORIGIN: %s is not exist", origin)
		return
	}
	if exitCode, err := bucket.RemoveFile(filename); err != nil {
		response.Executed = false
		response.Message = fmt.Sprintf("Error: %s. Exit code: %d", err.Error(), exitCode)
	}
	return
}
