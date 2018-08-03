package command

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/core/globals"
	core "bitbucket.org/udt/wizefs/core/primitives"
	"bitbucket.org/udt/wizefs/core/util"
)

// USECASE: wizefs create ORIGIN
func CmdCreateFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		// TEST: TestCreateUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	origin := c.Args()[0]
	exitCode, err := core.NewStorage().Create(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

// USECASE: wizefs delete ORIGIN
func CmdDeleteFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		// TEST: TestDeleteUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	origin := c.Args()[0]
	exitCode, err := core.NewStorage().Delete(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

// USECASE: wizefs mount ORIGIN
func CmdMountFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		// TEST: TestMountUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	// TODO: check permissions
	origin := c.Args()[0]

	// TEST: TestMountNotExistingOrigin, TestMountAlreadyMounted
	//exitCode, err := checkConfig(origin, false, true)
	//if err != nil {
	//	return cli.NewExitError(err, exitCode)
	//}

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	// TODO: think about ForkChild function
	fg := c.GlobalBool("fg")
	if !fg && c.NArg() == 1 {
		ret := util.ForkChild()
		os.Exit(ret)
	}

	notifypid := c.GlobalInt("notifypid")

	//exitCode, err = ApiMount(origin, notifypid)
	exitCode, err := core.NewStorage().Mount(origin, notifypid)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

// USECASE: wizefs unmount ORIGIN
func CmdUnmountFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		// TEST: TestUnmountUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	origin := c.Args()[0]

	//exitCode, err := ApiUnmount(origin)
	exitCode, err := core.NewStorage().Unmount(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}
