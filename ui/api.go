package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

const (
	packagePath = "ui"
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
