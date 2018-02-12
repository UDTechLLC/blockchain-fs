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

var args argContainer = argContainer{
	fg: false,
}

var test bool = false

func main() {

	fmt.Println("Start: ", os.Args)

	mxp := runtime.GOMAXPROCS(0)
	if mxp < 4 {
		// On a 2-core machine, setting maxprocs to 4 gives 10% better performance
		runtime.GOMAXPROCS(4)
	}

	app := cli.NewApp()

	fmt.Println("NewApp")

	app.Usage = "Internal API for Storage System"
	app.Version = "0.0.2"

	//	cli.BashCompletionFlag = cli.BoolFlag{
	//		Name:   "bg",
	//		Hidden: true,
	//	}
	//app.EnableBashCompletion = true

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
				fmt.Printf("Before command...\n")
				return nil
			},
			After: func(c *cli.Context) error {
				fmt.Printf("After command...\n")
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
			//			BashComplete: func(c *cli.Context) {
			//				fmt.Printf("Bash mount complete... %s and fg = %t and notifypid=%d\n", c.Args(), c.GlobalBool("fg"), c.GlobalInt("notifypid"))
			//				if !c.GlobalBool("fg") {
			//					ret := forkChild()
			//					os.Exit(ret)
			//				} else {
			//					FilesystemMountAction(c)
			//				}
			//			},
			Before: func(c *cli.Context) error {
				//fmt.Printf("Before command...\n")
				fmt.Printf("Before command... %s and fg = %t and notifypid=%d\n", c.Args(), c.GlobalBool("fg"), c.GlobalInt("notifypid"))
				return nil
			},
			After: func(c *cli.Context) error {
				fmt.Printf("After command... %s and fg = %t and notifypid=%d\n", c.Args(), c.GlobalBool("fg"), c.GlobalInt("notifypid"))
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

	fmt.Println("Before Run")

	app.Run(os.Args)

	fmt.Println("After Run")
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
	//	if !c.GlobalBool("fg") {
	//		ret := forkChild()
	//		os.Exit(ret)
	//	}

	fmt.Printf("Mount... %s and fg = %t and notifypid=%d\n", c.Args(), c.GlobalBool("fg"), c.GlobalInt("notifypid"))
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

	fmt.Println("Do mount")

	ret := doMount(&args)
	if ret != 0 {
		os.Exit(ret)
	}

	fmt.Println("After Do mount")

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
