package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli"

	"github.com/leedark/storage-system/internal/exitcodes"
)

// argContainer stores the parsed CLI options and arguments
type argContainer struct {
	fg                    bool
	origindir, mountpoint string
	notifypid             int
}

var args argContainer = argContainer{}

func main() {
	mxp := runtime.GOMAXPROCS(0)
	if mxp < 4 {
		// On a 2-core machine, setting maxprocs to 4 gives 10% better performance
		runtime.GOMAXPROCS(4)
	}

	app := cli.NewApp()
	app.Usage = "Internal API for Storage System"
	app.Version = "0.0.2"

	// Global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "fg",
			Usage: "Foreground...",
		},
		cli.IntFlag{
			Name:  "notifypid",
			Value: 0,
			Usage: "Notify PID...",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create new Filesystem into directory",
			BashComplete: func(c *cli.Context) {
				fmt.Printf("Bash create complete...\n")
			},
			Before: func(c *cli.Context) error {
				fmt.Printf("Before create...\n")
				return nil
			},
			After: func(c *cli.Context) error {
				fmt.Printf("After create...\n")
				return nil
			},
			Action: FilesystemCreateAction,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete existing Filesystem from directory",
			Action:  FilesystemDeleteAction,
		},
		{
			Name:    "mount",
			Aliases: []string{"m"},
			Usage:   "Mount Filesystem into directory",
			Before: func(c *cli.Context) error {
				fmt.Printf("Before mount...\n")
				return nil
			},
			After: func(c *cli.Context) error {
				fmt.Printf("After mount...\n")
				return nil
			},
			Action: FilesystemMountAction,
		},
		{
			Name:    "unmount",
			Aliases: []string{"u"},
			Usage:   "Unmount Filesystem from directory",
			Action:  FilesystemUnmountAction,
		},
	}

	app.Run(os.Args)
}

func FilesystemCreateAction(c *cli.Context) error {
	fmt.Printf("Create new Filesystem: %s\n", c.Args()[0])

	// TODO: create Directory?
	// TODO: initialize Filesystem, its Configuration
	// TODO: do something with Storage Database
	// TODO: do something else

	return nil
}

func FilesystemDeleteAction(c *cli.Context) error {
	fmt.Printf("Delete existing Filesystem: %s\n", c.Args()[0])
	return nil
}

func FilesystemMountAction(c *cli.Context) error {

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	if !c.GlobalBool("fg") {
		ret := forkChild()
		os.Exit(ret)
	}

	// TODO: check Directories
	if len(c.Args()) != 2 {
		fmt.Println("Wrong number of arguments (have %d, want 2). You passed: %s",
			len(c.Args()), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// Check origindir and mountpoint
	var err error

	args.notifypid = c.GlobalInt("notifypid")
	args.origindir, _ = filepath.Abs(c.Args()[0])
	err = checkDir(args.origindir)
	if err != nil {
		fmt.Println("Invalid origindir: %v", err)
		os.Exit(exitcodes.CipherDir)
	}

	args.mountpoint, err = filepath.Abs(c.Args()[1])
	if err != nil {
		fmt.Println("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}

	fmt.Printf("Mount Filesystem %s into %s\n", args.origindir, args.mountpoint)

	// TODO: do something with Storage Database and/or Configuration
	// TODO: do something else

	// TODO: do mounting with options
	ret := doMount(&args)
	if ret != 0 {
		os.Exit(ret)
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return nil
}

func FilesystemUnmountAction(c *cli.Context) error {
	fmt.Printf("Unmount Filesystem %s\n", c.Args()[0])

	// TODO: check Directory (Filesystem)
	// TODO: do unmounting with options
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
