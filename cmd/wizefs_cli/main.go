package main

import (
	"os"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	_ "bitbucket.org/udt/wizefs/internal/util"
)

func main() {
	config.InitWizeConfig()

	app := cli.NewApp()
	//app.Name = config.ProgramName
	app.Version = config.ProgramVersion
	app.Usage = "Internal API for Storage System"

	app.Flags = GlobalFlags
	app.Commands = Commands

	app.CommandNotFound = CommandNotFound
	app.Before = CommandBefore

	app.Run(os.Args)
}
