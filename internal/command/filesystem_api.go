package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/globals"
	"bitbucket.org/udt/wizefs/internal/tlog"
	"bitbucket.org/udt/wizefs/internal/util"
)

// USECASE: wizefs create ORIGIN
func CmdCreateFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	origin := c.Args()[0]

	return ApiCreate(origin)
}

func ApiCreate(origin string) (err error) {
	originPath := globals.OriginDirPath + origin

	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}
	// TODO: Create zip files
	if fstype == config.FSZip {
		return cli.NewExitError(
			"Creating zip files are not support now",
			globals.Origin)
	}

	tlog.Debug.Printf("Create new Filesystem %s on path %s\n", origin, originPath)

	// create Directory if it's not exist
	// TODO: check permissions
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", originPath)
		os.MkdirAll(originPath, 0755)
	} else {
		tlog.Warn.Printf("Directory %s is exist already!", originPath)
		return nil
	}

	// TODO: initialize Filesystem
	// TODO: do something with configuration
	config.NewFilesystemConfig(origin, originPath, config.FSLoopback).Save()

	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}
	err = config.CommonConfig.CreateFilesystem(origin, originPath, fstype)
	if err != nil {
		tlog.Warn.Printf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			tlog.Warn.Printf("Problem with saving Config: %v", err)
		}
	}

	// TODO: do something else

	return nil
}

// USECASE: wizefs delete ORIGIN
func CmdDeleteFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	origin := c.Args()[0]

	return ApiDelete(origin)
}

func ApiDelete(origin string) (err error) {
	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}
	// TODO: Delete zip files
	if fstype == config.FSZip {
		return cli.NewExitError(
			"Deleting zip files are not support now",
			globals.Origin)
	}

	tlog.Debug.Printf("Delete existing Filesystem: %s", origin)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		tlog.Warn.Printf("Directory %s is not exist!", originPath)
	} else {
		tlog.Debug.Printf("Delete existing directory: %s", originPath)
		os.RemoveAll(originPath)
	}

	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}
	err = config.CommonConfig.DeleteFilesystem(origin)
	if err != nil {
		tlog.Warn.Printf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			tlog.Warn.Printf("Problem with saving Config: %v", err)
		}
	}

	return nil
}

// USECASE: wizefs mount ORIGIN
func CmdMountFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	// Fork a child into the background if "-fg" is not set AND we are mounting
	// a filesystem. The child will do all the work.
	// TODO: think about ForkChild function
	fg := c.GlobalBool("fg")
	if !fg && c.NArg() == 1 {
		ret := util.ForkChild()
		os.Exit(ret)
	}

	notifypid := c.GlobalInt("notifypid")

	// TODO: check permissions
	//origin, _ := filepath.Abs(c.Args()[0])
	origin := c.Args()[0]

	return ApiMount(origin, notifypid)
}

func ApiMount(origin string, notifypid int) (err error) {
	originPath := globals.OriginDirPath + origin

	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}
	//if fstype == config.FSZip {
	//	tlog.Warn.Printf("Zip files are not support now")
	//	os.Exit(globals.Origin)
	//}

	// TODO: check mountpoint
	//mountpoint, err := filepath.Abs(c.Args()[1])
	//if err != nil {
	//	tlog.Warn.Printf("Invalid mountpoint: %v", err)
	//	os.Exit(globals.MountPoint)
	//}

	//mountpoint := c.Args()[1]
	// TODO: HACK - create/get mountpoint internally
	mountpoint := getMountpoint(origin, fstype)
	mountpointPath := globals.OriginDirPath + mountpoint

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", mountpointPath)
		os.MkdirAll(mountpointPath, 0755)
	} else {
		tlog.Warn.Printf("Directory %s is exist already!", mountpointPath)
	}

	tlog.Debug.Printf("Mount Filesystem %s into %s", originPath, mountpointPath)

	// Do mounting with options
	ret := util.DoMount(fstype, origin, originPath, mountpoint, mountpointPath, notifypid)
	if ret != 0 {
		os.Exit(ret)
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return nil
}

// USECASE: wizefs unmount ORIGIN
func CmdUnmountFilesystem(c *cli.Context) (err error) {
	if c.NArg() != 1 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	origin := c.Args()[0]

	return ApiUnmount(origin)
}

func ApiUnmount(origin string) (err error) {
	originPath := globals.OriginDirPath + origin

	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}

	// TODO: check mountpoint
	//mountpoint, err := filepath.Abs(c.Args()[0])
	//if err != nil {
	//	tlog.Warn.Printf("Invalid mountpoint: %v", err)
	//	os.Exit(globals.MountPoint)
	//}
	//mountpoint := c.Args()[0]
	// TODO: HACK - create/get mountpoint internally
	mountpoint := getMountpoint(origin, fstype)
	mountpointPath := globals.OriginDirPath + mountpoint

	tlog.Debug.Printf("Unmount Filesystem %s", mountpointPath)

	util.DoUnmount(mountpointPath)

	// TODO: do something with configuration
	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}
	err = config.CommonConfig.UnmountFilesystem(mountpoint)
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

func checkOriginType(origin string) (fstype config.FSType, err error) {
	fstype, err = util.CheckDirOrZip(origin)
	if err != nil {
		// HACK: if fstype = config.FSHack
		if fstype == config.FSHack {
			return config.FSLoopback, nil
		}
		return fstype, err
	}

	tlog.Debug.Printf("Origin Type: %d", fstype)
	return fstype, nil
}

func getMountpoint(origin string, fstype config.FSType) string {
	mountpoint := origin
	if fstype == config.FSZip {
		mountpoint = strings.Replace(mountpoint, ".", "_", -1)
	}
	mountpoint = "_mount" + mountpoint

	tlog.Debug.Printf("Mountpoint: %s", mountpoint)
	return mountpoint
}
