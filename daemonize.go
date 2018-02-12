package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/leedark/storage-system/internal/exitcodes"
	"github.com/leedark/storage-system/internal/syscallcompat"
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
	fmt.Printf("forkChild: starting %s with PID = %d\n", name, pid)
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
	// The child exited with 0 - let's do the same.
	return 0
}

// redirectStdFds redirects stderr and stdout to syslog; stdin to /dev/null
func redirectStdFds() {
	// Create a pipe pair "pw" -> "pr" and start logger reading from "pr".
	// We do it ourselves instead of using StdinPipe() because we need access
	// to the fd numbers.
	pr, pw, err := os.Pipe()
	if err != nil {
		fmt.Printf("redirectStdFds: could not create pipe: %v\n", err)
		return
	}
	tag := fmt.Sprintf("storage-system-%d-logger", os.Getpid())
	cmd := exec.Command("/usr/bin/logger", "-t", tag)
	cmd.Stdin = pr
	err = cmd.Start()
	if err != nil {
		fmt.Printf("redirectStdFds: could not start logger: %v\n", err)
		return
	}
	// The logger now reads on "pr". We can close it.
	pr.Close()
	// Redirect stout and stderr to "pw".
	err = syscallcompat.Dup3(int(pw.Fd()), 1, 0)
	if err != nil {
		fmt.Printf("redirectStdFds: stdout dup error: %v\n", err)
	}
	syscallcompat.Dup3(int(pw.Fd()), 2, 0)
	if err != nil {
		fmt.Printf("redirectStdFds: stderr dup error: %v\n", err)
	}
	// Our stout and stderr point to "pw". We can close the extra copy.
	pw.Close()
	// Redirect stdin to /dev/null
	nullFd, err := os.Open("/dev/null")
	if err != nil {
		fmt.Printf("redirectStdFds: could not open /dev/null: %v\n", err)
		return
	}
	err = syscallcompat.Dup3(int(nullFd.Fd()), 0, 0)
	if err != nil {
		fmt.Printf("redirectStdFds: stdin dup error: %v\n", err)
	}
	nullFd.Close()
}

// Send signal USR1 to "pid" (usually our parent process). This notifies it
// that the mounting has completed successfully.
func sendUsr1(pid int) {
	p, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("sendUsr1: FindProcess: %v\n", err)
		return
	}
	err = p.Signal(syscall.SIGUSR1)
	if err != nil {
		fmt.Printf("sendUsr1: Signal: %v\n", err)
	}
}
