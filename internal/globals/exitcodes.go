// Package exitcodes contains all well-defined exit codes that storage-system
// can return.
package globals

const (
	// ExitUsage - usage error like wrong cli syntax, wrong number of parameters.
	ExitUsage = 1
	// 2 is reserved because it is used by Go panic
	// 3 is reserved because it was used by earlier storage-system version as a generic
	// "mount" error.

	// ExitOrigin means that the ORIGIN does not exist, is not empty, or is not
	// a directory.
	ExitOrigin = 6
	// ExitMountPoint error means that the mountpoint is invalid (not empty etc).
	ExitMountPoint = 7
	// ExitInit is an error on filesystem init
	ExitInit = 8
	// ExitType is Filesystem type (1, 2)
	ExitType = 9

	ExitZip = 10

	// ExitOpenConf - the was an error opening the .conf file for reading
	ExitOpenConf = 20
	// ExitLoadConf is an error while loading .conf
	ExitLoadConf = 21
	// ExitSaveConf - could not save the .conf
	ExitSaveConf = 22
	// ExitChangeConf - the was an error changing .conf
	ExitChangeConf = 23

	// SigInt means we got SIGINT
	ExitSigInt = 30
	// ForkChild means forking the worker child failed
	ExitForkChild = 31

	// FuseNewServer - this exit code means that the call to fuse.NewServer failed.
	// This usually means that there was a problem executing fusermount, or
	// fusermount could not attach the mountpoint to the kernel.
	ExitFuseNewServer = 40

	// Other error - please inspect the message
	ExitOther = 50
)
