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
	// TODO: create zip files
	if fstype == config.ZipFS {
		return cli.NewExitError(
			"Creating zip files are not supported now",
			globals.Origin)
	}
	if fstype == config.LZFS {
		originPath = globals.OriginDirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)
	}

	tlog.Debug.Printf("Create new Filesystem %s on path %s\n", origin, originPath)

	// create Directory if it's not exist
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", originPath)
		os.MkdirAll(originPath, 0755)
	} else {
		// TODO: to decide what is it: no error, error, error & Extt
		return nil
		//return fmt.Errorf("Directory %s is exist already!", originPath)
		//return cli.NewExitError(
		//	fmt.Sprintf("Directory %s is exist already!", originPath),
		//	globals.Origin)
	}

	// initialize Filesystem
	// do something with configuration
	config.NewFilesystemConfig(origin, originPath, config.LoopbackFS).Save()

	if fstype == config.LZFS {
		util.ZipFile(originPath, globals.OriginDirPath+origin)
		// remove temp directory
		os.RemoveAll(originPath)
	}

	// HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
	}

	err = config.CommonConfig.CreateFilesystem(origin, originPath, fstype)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Problem with adding Filesystem to Config: %v", err),
			globals.Origin)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("Problem with saving Config: %v", err),
				globals.Origin)
		}
	}

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
	// check if origin is located in the common config
	// if not then we just saying about this
	// TODO: hide this into WizeConfig struct method
	fsinfo, ok := config.CommonConfig.Filesystems[origin]
	if !ok {
		return cli.NewExitError(
			fmt.Sprintf("Did not find ORIGIN: %s in common config.", origin),
			globals.Origin)
	}
	if fsinfo.MountpointKey != "" {
		return cli.NewExitError(
			fmt.Sprintf("This ORIGIN: %s is already mounted under MOUNTPOINT %s", origin, fsinfo.MountpointKey),
			globals.Origin)
	}

	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}
	// TODO: Delete zip files
	if fstype == config.ZipFS {
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

	// TODO: HACK - get mountpoint internally
	mountpoint := getMountpoint(origin, fstype)
	mountpointPath := globals.OriginDirPath + mountpoint

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		tlog.Warn.Printf("Directory %s is not exist!", mountpointPath)
	} else {
		tlog.Debug.Printf("Delete existing directory: %s", mountpointPath)
		os.RemoveAll(mountpointPath)
	}

	// TODO: HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
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

	// TODO: check permissions
	//origin, _ := filepath.Abs(c.Args()[0])
	origin := c.Args()[0]

	// check if origin is located in the common config
	// if not then we just saying about this
	// TODO: hide this into WizeConfig struct method
	fsinfo, ok := config.CommonConfig.Filesystems[origin]
	if !ok {
		return cli.NewExitError(
			fmt.Sprintf("Did not find ORIGIN: %s in common config.", origin),
			globals.Origin)
	}
	if fsinfo.MountpointKey != "" {
		return cli.NewExitError(
			fmt.Sprintf("This ORIGIN: %s is already mounted under MOUNTPOINT %s", origin, fsinfo.MountpointKey),
			globals.Origin)
	}

	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	tlog.Warn.Printf("FS type: %d", fstype)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}
	//if fstype == config.FSZip {
	//	return cli.NewExitError(
	//		fmt.Sprintf("Zip files are not support now"),
	//		globals.Origin)
	//}
	//if fstype == config.LZFS {
	//	return cli.NewExitError(
	//		fmt.Sprintf("LZFS files are not support now"),
	//		globals.Origin)
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

	if fstype == config.LZFS {
		// unzip to temp directory - OriginDirPath + "temp/" + filename (. -> _)
		tempPath := globals.OriginDirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)

		err = util.UnzipFile(originPath, tempPath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("LZFS file unzipping failed: %v", err),
				globals.Origin)
		}

		originPath = tempPath
	}

	// FUTURE: check mountpoint
	//mountpoint := c.Args()[1]
	//mountpoint, err := filepath.Abs(c.Args()[1])
	//if err != nil {
	//	tlog.Warn.Printf("Invalid mountpoint: %v", err)
	//	os.Exit(globals.MountPoint)
	//}

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
	// check if origin is located in the common config
	// if not then we just saying about this
	// TODO: hide this into WizeConfig struct method
	fsinfo, ok := config.CommonConfig.Filesystems[origin]
	if !ok {
		return cli.NewExitError(
			fmt.Sprintf("Did not find ORIGIN: %s in common config.", origin),
			globals.Origin)
	}
	if fsinfo.MountpointKey == "" {
		return cli.NewExitError(
			fmt.Sprintf("This ORIGIN: %s is not mounted under yet", origin),
			globals.Origin)
	}

	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Invalid origin: %v", err),
			globals.Origin)
	}

	// FUTURE: check mountpoint
	//mountpoint := c.Args()[0]
	//mountpoint, err := filepath.Abs(c.Args()[0])
	//if err != nil {
	//	tlog.Warn.Printf("Invalid mountpoint: %v", err)
	//	os.Exit(globals.MountPoint)
	//}

	// TODO: HACK - create/get mountpoint internally
	mountpoint := getMountpoint(origin, fstype)
	mountpointPath := globals.OriginDirPath + mountpoint

	tlog.Debug.Printf("Unmount Filesystem %s", mountpointPath)

	util.DoUnmount(mountpointPath)

	if fstype == config.LZFS {
		// zip temp directory
		tempPath := globals.OriginDirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)

		os.Remove(originPath)

		util.ZipFile(tempPath, originPath)

		// remove temp directory
		os.RemoveAll(tempPath)
	}

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		tlog.Warn.Printf("Directory %s is not exist!", mountpointPath)
	} else {
		tlog.Debug.Printf("Delete existing directory: %s", mountpointPath)
		os.RemoveAll(mountpointPath)
	}

	// TODO: HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
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

	return nil
}

func checkOriginType(origin string) (fstype config.FSType, err error) {
	fstype, err = util.CheckDirOrZip(origin)
	if err != nil {
		// HACK: if fstype = config.HackFS
		if fstype == config.HackFS {
			return config.LoopbackFS, nil
		}
		return fstype, err
	}

	return fstype, nil
}

func getMountpoint(origin string, fstype config.FSType) string {
	mountpoint := origin
	if fstype == config.ZipFS || fstype == config.LZFS {
		mountpoint = strings.Replace(mountpoint, ".", "_", -1)
	}
	mountpoint = "_mount" + mountpoint

	return mountpoint
}
