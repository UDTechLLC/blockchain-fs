package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/urfave/cli"
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
	app.Name = "storage-system"
	app.Usage = "Internal API for Storage System"
	app.Version = "0.0.3"

	// Global flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "fg, f",
			Usage: "Stay in the foreground",
		},
		cli.IntFlag{
			Name:  "notifypid",
			Value: 0,
			Usage: "Send USR1 to the specified process after " +
				"successful mount - used internally for daemonization",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Create new Filesystem into directory",
			Before: func(c *cli.Context) error {
				fmt.Printf("Before create...\n")
				return nil
			},
			After: func(c *cli.Context) error {
				fmt.Printf("After create...\n")
				return nil
			},
			Action: filesystemCreateAction,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "Delete existing Filesystem from directory",
			Action:  filesystemDeleteAction,
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
			Action: filesystemMountAction,
		},
		{
			Name:    "unmount",
			Aliases: []string{"u"},
			Usage:   "Unmount Filesystem from directory",
			Action:  filesystemUnmountAction,
		},
	}

	app.Run(os.Args)
}
