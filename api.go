package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/exitcodes"
)

func filesystemCreateAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	fmt.Printf("Create new Filesystem: %s\n", origindir)

	// create Directory if it's not exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		fmt.Printf("Create new directory: %s\n", origindir)
		os.Mkdir(origindir, 0755)
	} else {
		fmt.Printf("Directory %s is exist already!\n", origindir)
	}

	// TODO: initialize Filesystem, its Configuration
	// TODO: do something with Storage Database
	// TODO: do something else

	return nil
}

func filesystemDeleteAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	fmt.Printf("Delete existing Filesystem: %s\n", origindir)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		fmt.Printf("Directory %s is not exist!\n", origindir)
	} else {
		fmt.Printf("Delete existing directory: %s\n", origindir)
		os.Remove(origindir)
	}

	return nil
}

func filesystemMountAction(c *cli.Context) error {
	var err error

	if c.NArg() != 2 {
		fmt.Printf("Wrong number of arguments (have %d, want 2). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	if !c.GlobalBool("fg") && c.NArg() == 2 {
		ret := forkChild()
		os.Exit(ret)
	}

	args.fg = c.GlobalBool("fg")
	args.notifypid = c.GlobalInt("notifypid")

	// TODO: check permissions
	// Check origindir and mountpoint
	args.origindir, _ = filepath.Abs(c.Args()[0])
	err = checkDir(args.origindir)
	if err != nil {
		fmt.Printf("Invalid origindir: %v\n", err)
		os.Exit(exitcodes.OriginDir)
	}

	// TODO: check existing?
	args.mountpoint, err = filepath.Abs(c.Args()[1])
	if err != nil {
		fmt.Printf("Invalid mountpoint: %v\n", err)
		os.Exit(exitcodes.MountPoint)
	}

	fmt.Printf("Mount Filesystem %s into %s\n", args.origindir, args.mountpoint)

	// TODO: do something with Storage Database and/or Configuration
	// TODO: do something else

	// Do mounting with options
	ret := mount(&args)
	if ret != 0 {
		os.Exit(ret)
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return nil
}

func filesystemUnmountAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// TODO: check Directory (Filesystem)
	mountpoint, err := filepath.Abs(c.Args()[0])
	if err != nil {
		fmt.Printf("Invalid mountpoint: %v\n", err)
		os.Exit(exitcodes.MountPoint)
	}
	fmt.Printf("Unmount Filesystem %s\n", mountpoint)

	// TODO: do unmounting with options
	unmountPanic(mountpoint)

	// TODO: do something with Storage Database and/or Configuration
	// TODO: do something else

	return nil
}

// checkDir - check if "dir" exists and is a directory
func checkDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}
