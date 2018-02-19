package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/command"
	"bitbucket.org/udt/wizefs/internal/tlog"
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
		Usage:   "Create new Filesystem into directory",
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
		Usage:   "Delete existing Filesystem from directory",
		Action:  command.CmdDeleteFilesystem,
	},
	{
		Name:    "mount",
		Aliases: []string{"m"},
		Usage:   "Mount Filesystem into directory",
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
		Usage:   "Unmount Filesystem from directory",
		Action:  command.CmdUnmountFilesystem,
	},
	{
		Name:    "put",
		Aliases: []string{"p"},
		Usage:   "",
		Action:  command.CmdPutFile,
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
