package main

import (
	"os"

	"github.com/urfave/cli"

	_ "bitbucket.org/udt/wizefs/internal/util"
)

// Name is program name
const Name string = "wizefs"

// Version is program version
var Version string = "0.0.4"

func main() {
	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Usage = "Internal API for Storage System"

	app.Flags = GlobalFlags
	app.Commands = Commands

	app.CommandNotFound = CommandNotFound
	app.Before = CommandBefore

	app.Run(os.Args)
}
