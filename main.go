package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/leedark/storage-system/internal/exitcodes"
)

const (
	ProgramName    = "storage-system"
	ProgramVersion = "0.0"
)

func printVersion() {
	fmt.Printf("%s %s\n", ProgramName, ProgramVersion)
}

func main() {
	mxp := runtime.GOMAXPROCS(0)
	if mxp < 4 {
		// On a 2-core machine, setting maxprocs to 4 gives 10% better performance
		runtime.GOMAXPROCS(4)
	}

	var err error

	// Parse all command-line options (i.e. arguments starting with "-")
	// into "args". Path arguments are parsed below.
	args := parseCliOpts()

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	if !args.fg && flagSet.NArg() == 2 {
		ret := forkChild()
		os.Exit(ret)
	}

	// "-v"
	if args.version {
		printVersion()
		os.Exit(0)
	}

	// Every operation below requires CIPHERDIR. Exit if we don't have it.
	if flagSet.NArg() == 0 {
		if flagSet.NFlag() == 0 {
			// Naked call to "storage-system". Just print the help text.
			helpShort()
		} else {
			// The user has passed some flags, but CIPHERDIR is missing. State
			// what is wrong.
			fmt.Println("CIPHERDIR argument is missing")
		}
		os.Exit(exitcodes.Usage)
	}
	// Check that CIPHERDIR exists
	args.cipherdir, _ = filepath.Abs(flagSet.Arg(0))
	err = checkDir(args.cipherdir)
	if err != nil {
		fmt.Println("Invalid cipherdir: %v", err)
		os.Exit(exitcodes.CipherDir)
	}

	// Operation flags
	if args.info && args.init {
		fmt.Println("At most one of -info, -init is allowed")
		os.Exit(exitcodes.Usage)
	}
	// "-info"
	if args.info {
		if flagSet.NArg() > 1 {
			fmt.Println("Usage: %s -info CIPHERDIR", ProgramName)
			os.Exit(exitcodes.Usage)
		}
		info("configfile") // does not return
	}
	// "-init"
	if args.init {
		if flagSet.NArg() > 1 {
			fmt.Println("Usage: %s -init [OPTIONS] CIPHERDIR", ProgramName)
			os.Exit(exitcodes.Usage)
		}
		initDir(&args) // does not return
	}

	// Default operation: mount.
	if flagSet.NArg() != 2 {
		prettyArgs := prettyArgs()
		fmt.Println("Wrong number of arguments (have %d, want 2). You passed: %s",
			flagSet.NArg(), prettyArgs)
		fmt.Printf("Usage: %s [OPTIONS] CIPHERDIR MOUNTPOINT\n",
			ProgramName)
		os.Exit(exitcodes.Usage)
	}
	ret := doMount(&args)
	if ret != 0 {
		os.Exit(ret)
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
}

func info(filename string) {
	fmt.Println("info")
	os.Exit(0)
}

func initDir(args *argContainer) {
	fmt.Println("initDir")
	os.Exit(0)
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
