package main

import (
	"os/exec"
)

const (
	projectPath = "/home/sergey/code/go/src/bitbucket.org/udt/wizefs/"
)

func RunCommand(arg ...string) (cerr error) {
	appPath := projectPath + "wizefs"
	c := exec.Command(appPath, arg...)
	//t.Logf("starting command %s...", command)
	cerr = c.Start()
	if cerr != nil {
		//t.Errorf("starting command failed: %v", cerr)
		return cerr
	}

	//t.Logf("waiting command %s...", command)
	cerr = c.Wait()

	//t.Logf("finishing command %s...", command)
	return cerr
}
