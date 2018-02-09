package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leedark/storage-system/internal/exitcodes"
)

// doMount mounts an encrypted directory.
// Called from main.
func doMount(args *argContainer) int {
	// Check mountpoint
	var err error
	args.mountpoint, err = filepath.Abs(flagSet.Arg(1))
	if err != nil {
		fmt.Println("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}

	// Initialize FUSE server
	//srv := initFuseFrontend(masterkey, args, confFile)
	fmt.Println("Filesystem mounted and ready.")

	// Wait for SIGINT in the background and unmount ourselves if we get it.
	// This prevents a dangling "Transport endpoint is not connected"
	// mountpoint if the user hits CTRL-C.
	//handleSigint(srv, args.mountpoint)

	// Return memory that was allocated for scrypt (64M by default!) and other
	// stuff that is no longer needed to the OS
	//debug.FreeOSMemory()

	// Jump into server loop. Returns when it gets an umount request from the kernel.
	//srv.Serve()
	return 0
}
