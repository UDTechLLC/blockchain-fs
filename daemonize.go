package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/leedark/storage-system/internal/exitcodes"
)

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

// forkChild - execute ourselves once again, this time with the "-fg" flag, and
// wait for SIGUSR1 or child exit.
// This is a workaround for the missing true fork function in Go.
func forkChild() int {
	name := os.Args[0]
	pid := os.Getpid()
	newArgs := []string{"-fg", fmt.Sprintf("-notifypid=%d", pid)}
	newArgs = append(newArgs, os.Args[1:]...)
	c := exec.Command(name, newArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	exitOnUsr1()
	err := c.Start()
	if err != nil {
		fmt.Printf("forkChild: starting %s failed: %v\n", name, err)
		return exitcodes.ForkChild
	}
	err = c.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if waitstat, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(waitstat.ExitStatus())
			}
		}
		fmt.Printf("forkChild: wait returned an unknown error: %v\n", err)
		return exitcodes.ForkChild
	}
	fmt.Printf("forkChild: starting %s with PID = %d\n", name, pid)
	// The child exited with 0 - let's do the same.
	return 0
}
