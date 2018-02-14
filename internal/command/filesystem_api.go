package command

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/exitcodes"
	"bitbucket.org/udt/wizefs/internal/tlog"
	"bitbucket.org/udt/wizefs/internal/util"
)

func CmdCreateFilesystem(c *cli.Context) error {
	if c.NArg() != 1 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	tlog.Debug.Printf("Create new Filesystem: %s\n", origindir)

	// create Directory if it's not exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", origindir)
		os.Mkdir(origindir, 0755)
	} else {
		tlog.Warn.Printf("Directory %s is exist already!", origindir)
		return nil
	}

	// TODO: initialize Filesystem
	// TODO: do something with configuration
	config.NewFilesystemConfig(origindir, 1).Save()

	// TODO: do something else

	return nil
}

func CmdDeleteFilesystem(c *cli.Context) error {
	if c.NArg() != 1 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	origindir := c.Args()[0]
	tlog.Debug.Printf("Delete existing Filesystem: %s", origindir)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(origindir); os.IsNotExist(err) {
		tlog.Warn.Printf("Directory %s is not exist!", origindir)
	} else {
		tlog.Debug.Printf("Delete existing directory: %s", origindir)
		os.RemoveAll(origindir)
	}

	return nil
}

func CmdMountFilesystem(c *cli.Context) error {
	var err error

	if c.NArg() != 2 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 2). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	fg := c.GlobalBool("fg")
	if !fg && c.NArg() == 2 {
		ret := util.ForkChild()
		os.Exit(ret)
	}

	notifypid := c.GlobalInt("notifypid")

	// TODO: check permissions
	// Check origindir and mountpoint
	origindir, _ := filepath.Abs(c.Args()[0])
	err = util.CheckDir(origindir)
	if err != nil {
		tlog.Warn.Printf("Invalid origindir: %v", err)
		os.Exit(exitcodes.OriginDir)
	}

	// TODO: check existing?
	mountpoint, err := filepath.Abs(c.Args()[1])
	if err != nil {
		tlog.Warn.Printf("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}

	tlog.Debug.Printf("Mount Filesystem %s into %s", origindir, mountpoint)

	// Do mounting with options
	ret := util.DoMount(origindir, mountpoint, notifypid)
	if ret != 0 {
		os.Exit(ret)
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return nil
}

func CmdUnmountFilesystem(c *cli.Context) error {
	var err error

	if c.NArg() != 1 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// TODO: check Directory (Filesystem)
	mountpoint, err := filepath.Abs(c.Args()[0])
	if err != nil {
		tlog.Warn.Printf("Invalid mountpoint: %v", err)
		os.Exit(exitcodes.MountPoint)
	}
	tlog.Debug.Printf("Unmount Filesystem %s", mountpoint)

	// TODO: do unmounting with options
	util.DoUnmount(mountpoint)

	// TODO: do something with configuration
	err = config.CommonConfig.DeleteFilesystem(mountpoint)
	if err != nil {
		tlog.Warn.Printf("Problem with deleteing Filesystem from Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			tlog.Warn.Printf("Problem with saving Config: %v", err)
		}
	}

	// TODO: do something else

	return nil
}
