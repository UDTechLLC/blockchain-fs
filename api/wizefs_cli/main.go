package main

import (
	"os"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/core/globals"
	_ "bitbucket.org/udt/wizefs/core/util"
)

func main() {
	app := cli.NewApp()
	app.Name = globals.ProjectName
	app.Version = globals.ProjectVersion
	app.Usage = "Command-line API for WizeFS Storage"

	app.Flags = GlobalFlags
	app.Commands = Commands

	app.CommandNotFound = CommandNotFound
	app.Before = CommandBefore

	app.Run(os.Args)
}
