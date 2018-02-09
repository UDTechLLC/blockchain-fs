package main

import (
	"fmt"
)

const tUsage = "" +
	"Usage: " + ProgramName + " -init|-info [OPTIONS] CIPHERDIR\n" +
	"  or   " + ProgramName + " [OPTIONS] CIPHERDIR MOUNTPOINT\n"

// helpShort is what gets displayed when passed "-h" or on syntax error.
func helpShort() {
	printVersion()
	fmt.Printf("\n")
	fmt.Printf(tUsage)
	fmt.Printf(`
Common Options (use -hh to show all):
  -h, -help          This short help text
  -fg                Stay in the foreground
  -init              Initialize encrypted directory
  -info              Display information about encrypted directory
  -version           Print version information
  --                 Stop option parsing
`)
}
