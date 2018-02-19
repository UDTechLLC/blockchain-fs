// Package exitcodes contains all well-defined exit codes that storage-system
// can return.
package exitcodes

const (
	// Usage - usage error like wrong cli syntax, wrong number of parameters.
	Usage = 1
	// 2 is reserved because it is used by Go panic
	// 3 is reserved because it was used by earlier storage-system version as a generic
	// "mount" error.

	// Origin means that the ORIGIN does not exist, is not empty, or is not
	// a directory.
	Origin = 6
	// MountPoint error means that the mountpoint is invalid (not empty etc).
	MountPoint = 7
	// Init is an error on filesystem init
	Init = 8
	// Type is Filesystem type (1, 2)
	Type = 9

	// LoadConf is an error while loading .conf
	LoadConf = 20
	// OpenConf - the was an error opening the .conf file for reading
	OpenConf = 21
	// WriteConf - could not write the .conf
	WriteConf = 22

	// SigInt means we got SIGINT
	SigInt = 30
	// PanicLogNotEmpty means the panic log was not empty when we were unmounted
	PanicLogNotEmpty = 31
	// ForkChild means forking the worker child failed
	ForkChild = 32

	// FuseNewServer - this exit code means that the call to fuse.NewServer failed.
	// This usually means that there was a problem executing fusermount, or
	// fusermount could not attach the mountpoint to the kernel.
	FuseNewServer = 33

	// Other error - please inspect the message
	Other = 50
)
