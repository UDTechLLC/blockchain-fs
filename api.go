package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/exitcodes"
	"bitbucket.org/udt/wizefs/internal/tlog"
)

var configfile *config.FilesystemConfig

func filesystemCreateAction(c *cli.Context) error {
	if c.NArg() != 1 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	tlog.Debug.Printf("Create new Filesystem: %s\n", origindir)

	// create Directory if it's not exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", origindir)
		os.Mkdir(origindir, 0755)
	} else {
		tlog.Warn.Printf("Directory %s is exist already!", origindir)
		return nil
	}

	// TODO: initialize Filesystem, its Configuration
	configfile = config.NewFilesystemConfig("wizefs", origindir, 1)
	configfile.Save()

	// TODO: do something with Storage Database
	// TODO: do something else

	return nil
}

func filesystemDeleteAction(c *cli.Context) error {
	if c.NArg() != 1 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	tlog.Debug.Printf("Delete existing Filesystem: %s", origindir)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		tlog.Warn.Printf("Directory %s is not exist!", origindir)
	} else {
		tlog.Debug.Printf("Delete existing directory: %s", origindir)
		os.Remove(origindir)
	}

	return nil
}

func filesystemMountAction(c *cli.Context) error {
	var err error

	if c.NArg() != 2 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 2). You passed: %s",
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
		tlog.Warn.Printf("Invalid origindir: %v", err)
		os.Exit(exitcodes.OriginDir)
	}

	// TODO: check existing?
	args.mountpoint, err = filepath.Abs(c.Args()[1])
	if err != nil {
		tlog.Warn.Printf("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}

	tlog.Debug.Printf("Mount Filesystem %s into %s", args.origindir, args.mountpoint)

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
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// TODO: check Directory (Filesystem)
	mountpoint, err := filepath.Abs(c.Args()[0])
	if err != nil {
		tlog.Warn.Printf("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}
	tlog.Debug.Printf("Unmount Filesystem %s", mountpoint)

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
