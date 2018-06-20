package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/api/command"
	"bitbucket.org/udt/wizefs/core/tlog"
)

var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug output",
	},
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

var Commands = []cli.Command{
	{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Create new Bucket to Storage",
		Before: func(c *cli.Context) error {
			tlog.Debug.Printf("Before create...")
			return nil
		},
		After: func(c *cli.Context) error {
			tlog.Debug.Printf("After create...")
			return nil
		},
		Action: command.CmdCreateFilesystem,
	},
	{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete existing Bucket from Storage",
		Action:  command.CmdDeleteFilesystem,
	},
	{
		Name:    "mount",
		Aliases: []string{"m"},
		Usage:   "Mount Bucket",
		Before: func(c *cli.Context) error {
			tlog.Debug.Printf("Before mount...")
			return nil
		},
		After: func(c *cli.Context) error {
			tlog.Debug.Printf("After mount...")
			return nil
		},
		Action: command.CmdMountFilesystem,
	},
	{
		Name:    "unmount",
		Aliases: []string{"u"},
		Usage:   "Unmount Bucket",
		Action:  command.CmdUnmountFilesystem,
	},
	{
		Name:    "put",
		Aliases: []string{"p"},
		Usage:   "Put file to Bucket",
		Action:  command.CmdPutFile,
	},
	{
		Name:    "get",
		Aliases: []string{"g"},
		Usage:   "Get file from Bucket",
		Action:  command.CmdGetFile,
	},
	{
		Name:    "remove",
		Aliases: []string{"r"},
		Usage:   "Remove file from Bucket",
		Action:  command.CmdRemoveFile,
	},
}

// CommandNotFound implements action when subcommand not found
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}

// CommandBefore implements action before run command
func CommandBefore(c *cli.Context) error {
	if c.GlobalBool("debug") {
		tlog.Debug.Enabled = true
	}
	return nil
}
