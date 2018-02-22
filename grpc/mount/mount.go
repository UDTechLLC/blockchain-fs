package main

import (
	"flag"
	"os"
	"runtime"

	api "bitbucket.org/udt/wizefs/internal/command"
	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/tlog"
	"bitbucket.org/udt/wizefs/internal/util"
)

// argContainer stores the parsed CLI options and arguments
type argContainer struct {
	fg        bool
	notifypid int
}

var flagSet *flag.FlagSet

// parseCliOpts - parse command line options (i.e. arguments that start with "-")
func parseCliOpts() (args argContainer) {
	var err error

	flagSet = flag.NewFlagSet("mount", flag.ContinueOnError)
	//flagSet.Usage = helpShort
	flagSet.BoolVar(&args.fg, "fg", false, "Stay in the foreground")
	flagSet.IntVar(&args.notifypid, "notifypid", 0, "Send USR1 to the specified process after "+
		"successful mount - used internally for daemonization")

	// Actual parsing
	err = flagSet.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(1)
	}

	return args
}

func main() {
	mxp := runtime.GOMAXPROCS(0)
	if mxp < 4 {
		// On a 2-core machine, setting maxprocs to 4 gives 10% better performance
		runtime.GOMAXPROCS(4)
	}

	tlog.Debug.Enabled = true
	tlog.Info.Println("Before mount")

	// Parse all command-line options (i.e. arguments starting with "-")
	// into "args". Path arguments are parsed below.
	args := parseCliOpts()

	origin := flagSet.Arg(0)

	// HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}

	// Check that ORIGIN exists
	fsinfo, ok := config.CommonConfig.Filesystems[origin]
	if !ok {
		tlog.Warn.Printf("Did not find ORIGIN: %s in common config.", origin)
	}
	if fsinfo.MountpointKey != "" {
		tlog.Warn.Printf("This ORIGIN: %s is already mounted under MOUNTPOINT %s", origin, fsinfo.MountpointKey)
	}

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	if !args.fg && flagSet.NArg() == 1 {
		ret := util.ForkChild()
		os.Exit(ret)
	}

	// Every operation below requires ORIGIN. Exit if we don't have it.
	if flagSet.NArg() == 0 {
		if flagSet.NFlag() == 0 {
			// Naked call to "mount". Just print the help text.
			//helpShort()
		} else {
			// The user has passed some flags, but ORIGIN is missing. State
			// what is wrong.
			tlog.Info.Println("ORIGIN argument is missing")
		}
		os.Exit(1)
	}

	err := api.ApiMount(origin, args.notifypid)
	if err != nil {
		tlog.Warn.Println("Error with ApiMount: %v", err)
	}
}
