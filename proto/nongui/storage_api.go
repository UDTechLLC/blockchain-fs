package nongui

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	packagePath = "proto"
)

var (
	projectPath = getProjectPath()
)

func getProjectPath() string {
	_, testFilename, _, _ := runtime.Caller(0)
	idx := strings.Index(testFilename, packagePath)
	return testFilename[0:idx]
}

func RunCommand(arg ...string) (cerr error) {
	appPath := projectPath + "wizefs"

	var outbuf, errbuf bytes.Buffer
	c := exec.Command(appPath, arg...)
	c.Stdout = &outbuf
	c.Stderr = &errbuf

	//t.Logf("starting command %s...", command)
	cerr = c.Start()
	if cerr != nil {
		//t.Errorf("starting command failed: %v", cerr)
		return cerr
	}

	//t.Logf("waiting command %s...", command)
	cerr = c.Wait()

	if cerr != nil {
		if exiterr, ok := cerr.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("Exit status: %d. [%s]",
					waitstat.ExitStatus(), errbuf.String()[:errbuf.Len()-1])
			}
		} else {
			return fmt.Errorf("Unknown error: %v. [%s]",
				cerr, errbuf.String()[:errbuf.Len()-1])
		}
	}

	//t.Logf("finishing command %s...", command)
	return cerr
}

func CreateStorage(origin string) error {
	return RunCommand("create", origin)
}

func MountStorage(origin string) (cerr error) {
	cerr = RunCommand("mount", origin)
	if cerr != nil {
		return cerr
	}

	// TODO: we must wait until mount finishes its actions
	// TODO: check ORIGIN? every 100 milliseconds
	time.Sleep(500 * time.Millisecond)
	return nil
}

func UnmountStorage(origin string) error {
	return RunCommand("unmount", origin)
}

func PutFile(filename, origin string) error {
	return RunCommand("put", filename, origin)
}

func GetFile(source, origin, destination string) error {
	return RunCommand("xget", source, origin, destination)
}

func RemoveFile(filename, origin string) error {
	return RunCommand("remove", filename, origin)
}
