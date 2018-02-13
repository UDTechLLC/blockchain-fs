// Package exitcodes contains all well-defined exit codes that storage-system
// can return.
package exitcodes

import (
	"fmt"
	"os"
)

const (
	// Usage - usage error like wrong cli syntax, wrong number of parameters.
	Usage = 1
	// 2 is reserved because it is used by Go panic
	// 3 is reserved because it was used by earlier storage-system version as a generic
	// "mount" error.

	// OriginDir means that the ORIGINDIR does not exist, is not empty, or is not
	// a directory.
	OriginDir = 6
	// MountPoint error means that the mountpoint is invalid (not empty etc).
	MountPoint = 7
	// Init is an error on filesystem init
	Init = 8
	// LoadConf is an error while loading gocryptfs.conf
	LoadConf = 9
	// OpenConf - the was an error opening the gocryptfs.conf file for reading
	OpenConf = 10
	// WriteConf - could not write the gocryptfs.conf
	WriteConf = 11

	// SigInt means we got SIGINT
	SigInt = 20
	// PanicLogNotEmpty means the panic log was not empty when we were unmounted
	PanicLogNotEmpty = 21
	// ForkChild means forking the worker child failed
	ForkChild = 22

	// FuseNewServer - this exit code means that the call to fuse.NewServer failed.
	// This usually means that there was a problem executing fusermount, or
	// fusermount could not attach the mountpoint to the kernel.
	FuseNewServer = 23

	// Other error - please inspect the message
	Other = 50
)

// Err wraps an error with an associated numeric exit code
type Err struct {
	error
	code int
}

// NewErr returns an error containing "msg" and the exit code "code".
func NewErr(msg string, code int) Err {
	return Err{
		error: fmt.Errorf(msg),
		code:  code,
	}
}

// Exit extracts the numeric exit code from "err" (if available) and exits the
// application.
func Exit(err error) {
	err2, ok := err.(Err)
	if !ok {
		os.Exit(Other)
	}
	os.Exit(err2.code)
}
