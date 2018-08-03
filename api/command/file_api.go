package command

import (
	"fmt"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/core/globals"
	core "bitbucket.org/udt/wizefs/core/primitives"
)

// wizefs load FILE ORIGIN -> load FILE [ORIGIN]
// TODO: output result: stdout, JSON
// TODO: check permissions
func CmdPutFile(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		// TEST: TestPutUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 2)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	originalFile := c.Args()[0]
	origin := c.Args()[1]

	//exitCode, err := ApiPut(originalFile, origin, nil)
	var exitCode int
	bucket, ok := core.NewStorage().Bucket(origin)
	if ok {
		exitCode, err = bucket.PutFile(originalFile, nil)
	} else {
		err = fmt.Errorf("Bucket with ORIGIN: %s is not exist", origin)
		exitCode = globals.ExitOrigin
	}
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

// wizefs get FILE ORIGIN [DESTINATIONFILEPATH]
// TODO: output result: stdout, JSON + file content []byte, size
// TODO: check permissions
func CmdGetFile(c *cli.Context) (err error) {
	if c.NArg() < 2 {
		// TEST: TestGetUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 2 or 3)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	originalFile := c.Args()[0]
	origin := c.Args()[1]
	destinationFilePath := ""
	if c.NArg() == 3 {
		destinationFilePath = c.Args()[2]
	}

	// we don't need content, it's only for gRPC methods
	//_, exitCode, err := ApiGet(originalFile, origin, "", false)
	var exitCode int
	bucket, ok := core.NewStorage().Bucket(origin)
	if ok {
		_, exitCode, err = bucket.GetFile(originalFile, destinationFilePath, false)
	} else {
		err = fmt.Errorf("Bucket with ORIGIN: %s is not exist", origin)
		exitCode = globals.ExitOrigin
	}
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

// wizefs remove FILE ORIGIN
// TODO: check permissions
func CmdRemoveFile(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		// TEST: TestRemoveUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 2)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	originalFile := c.Args()[0]
	origin := c.Args()[1]

	//exitCode, err := ApiRemove(originalFile, origin)
	var exitCode int
	bucket, ok := core.NewStorage().Bucket(origin)
	if ok {
		exitCode, err = bucket.RemoveFile(originalFile)
	} else {
		err = fmt.Errorf("Bucket with ORIGIN: %s is not exist", origin)
		exitCode = globals.ExitOrigin
	}
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}
