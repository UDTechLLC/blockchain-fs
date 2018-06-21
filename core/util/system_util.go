package util

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/tlog"
)

func init() {
	mxp := runtime.GOMAXPROCS(0)
	if mxp < 4 {
		// On a 2-core machine, setting maxprocs to 4 gives 10% better performance
		runtime.GOMAXPROCS(4)
	}
}

// forkChild - execute ourselves once again, this time with the "-fg" flag, and
// wait for SIGUSR1 or child exit.
// This is a workaround for the missing true fork function in Go.
func ForkChild() int {
	name := os.Args[0]
	pid := os.Getpid()
	newArgs := []string{"--fg", fmt.Sprintf("--notifypid=%d", pid)}
	newArgs = append(newArgs, os.Args[1:]...)

	c := exec.Command(name, newArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	exitOnUsr1()
	err := c.Start()

	if err != nil {
		tlog.Warnf("forkChild: starting %s failed: %v", name, err)
		return globals.ExitForkChild
	}

	tlog.Debugf("forkChild: starting %s with PID = %d", name, pid)

	err = c.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(waitstat.ExitStatus())
			}
		}
		tlog.Warnf("forkChild: wait returned an unknown error: %v", err)
		return globals.ExitForkChild
	}

	// The child exited with 0 - let's do the same.
	return 0
}

// The child sends us USR1 if the mount was successful. Exit with error code
// 0 if we get it.
func exitOnUsr1() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		<-c
		os.Exit(0)
	}()
}
