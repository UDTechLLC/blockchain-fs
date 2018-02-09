package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leedark/storage-system/internal/exitcodes"
)

// argContainer stores the parsed CLI options and arguments
type argContainer struct {
	version, init, info, fg bool
	cipherdir, mountpoint   string
	notifypid               int
}

var flagSet *flag.FlagSet

// parseCliOpts - parse command line options (i.e. arguments that start with "-")
func parseCliOpts() (args argContainer) {
	var err error

	flagSet = flag.NewFlagSet(ProgramName, flag.ContinueOnError)
	flagSet.Usage = helpShort
	flagSet.BoolVar(&args.version, "version", false, "Print version and exit")
	flagSet.BoolVar(&args.fg, "fg", false, "Stay in the foreground")
	flagSet.BoolVar(&args.init, "init", false, "Initialize encrypted directory")
	flagSet.BoolVar(&args.info, "info", false, "Display information about CIPHERDIR")
	flagSet.IntVar(&args.notifypid, "notifypid", 0, "Send USR1 to the specified process after "+
		"successful mount - used internally for daemonization")

	// Actual parsing
	err = flagSet.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(exitcodes.Usage)
	}

	return args
}

// prettyArgs pretty-prints the command-line arguments.
func prettyArgs() string {
	pa := fmt.Sprintf("%q", os.Args[1:])
	// Get rid of "[" and "]"
	pa = pa[1 : len(pa)-1]
	return pa
}
