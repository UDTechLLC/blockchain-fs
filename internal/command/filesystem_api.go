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
		// TEST: TestCreateUsage
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 1)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.ExitUsage)
	}

	origin := c.Args()[0]

	exitCode, err := ApiCreate(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func ApiCreate(origin string) (exitCode int, err error) {
	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		// TEST: TestCreateInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v.", err)
	}

	// TODO: create zip files
	if fstype == config.ZipFS {
		// TEST: TestCreateZipFS (like _archive.zip)
		return globals.ExitOrigin,
			fmt.Errorf("Creating zip files are not supported now")
	}
	if fstype == config.LZFS {
		originPath = globals.OriginDirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)
	}

	tlog.Debug.Printf("Creating new Filesystem %s on path %s...\n", origin, originPath)

	// create Directory if it's not exist
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", originPath)
		os.MkdirAll(originPath, 0755)
	} else {
		// TODO: what we should done when origin is exist already?
		return globals.ExitOrigin,
			fmt.Errorf("Directory %s is exist already!", originPath)
	}

	// do something with Filesystem configuration
	config.NewFilesystemConfig(origin, originPath, config.LoopbackFS).Save()

	// create LZFS archive
	if fstype == config.LZFS {
		targetFile := globals.OriginDirPath + origin
		err = util.ZipFile(originPath, targetFile)
		if err != nil {
			// TEST: TestCreateLZFS (like archive.zip)
			return globals.ExitZip,
				fmt.Errorf("LZFS file zipping failed: %v", err)
		}
		// remove temp directory
		os.RemoveAll(originPath)
	}

	// TODO: HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}
	err = config.CommonConfig.CreateFilesystem(origin, originPath, fstype)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	return 0, nil
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

	exitCode, err := ApiDelete(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func ApiDelete(origin string) (exitCode int, err error) {
	existOrigin, existMountpoint := config.CommonConfig.CheckFilesystem(origin)
	if !existOrigin {
		// TEST: TestDeleteNotExistingOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Did not find ORIGIN: %s in common config.", origin)
	}
	if existMountpoint {
		// TEST: TestDeleteAlreadyMounted
		return globals.ExitMountPoint,
			fmt.Errorf("This ORIGIN: %s is already mounted", origin)
	}

	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		// TEST: TestDeleteInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v", err)
	}

	// TODO: Delete zip files
	if fstype == config.ZipFS {
		// TEST: TestDeleteZipFS
		return globals.ExitOrigin,
			fmt.Errorf("Deleting zip files are not support now")
	}

	tlog.Debug.Printf("Delete existing Filesystem: %s", origin)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		// TODO: what we should done when origin is exist already?
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
	}
	err = config.CommonConfig.DeleteFilesystem(origin)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with deleting Filesystem from Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	return 0, nil
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

	existOrigin, existMountpoint := config.CommonConfig.CheckFilesystem(origin)
	if !existOrigin {
		// TEST: TestMountNotExistingOrigin
		err = fmt.Errorf("Did not find ORIGIN: %s in common config.", origin)
		return cli.NewExitError(err, globals.ExitOrigin)
	}
	if existMountpoint {
		// TEST: TestMountAlreadyMounted
		err = fmt.Errorf("This ORIGIN: %s is already mounted", origin)
		return cli.NewExitError(err, globals.ExitMountPoint)
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

	exitCode, err := ApiMount(origin, notifypid)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func ApiMount(origin string, notifypid int) (exitCode int, err error) {
	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		// TEST: TestMountInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v", err)
	}

	if fstype == config.LZFS {
		// unzip to temp directory - OriginDirPath + "temp/" + filename (. -> _)
		tempPath := globals.OriginDirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)

		err = util.UnzipFile(originPath, tempPath)
		if err != nil {
			// TEST: TestMountLZFSUnzip
			return globals.ExitZip,
				fmt.Errorf("LZFS file unzipping failed: %v", err)
		}

		originPath = tempPath
	}

	// TODO: check mountpoint
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
		//os.Exit(ret)
		return ret, nil
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return 0, nil
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

	exitCode, err := ApiUnmount(origin)
	if err != nil {
		//tlog.Warn.Println(err)
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func ApiUnmount(origin string) (exitCode int, err error) {
	existOrigin, existMountpoint := config.CommonConfig.CheckFilesystem(origin)
	if !existOrigin {
		// TEST: TestUnmountNotExistingOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Did not find ORIGIN: %s in common config.", origin)
	}
	if !existMountpoint {
		// TEST: TestUnmountNotMounted
		return globals.ExitMountPoint,
			fmt.Errorf("This ORIGIN: %s is not mounted yet", origin)
	}

	originPath := globals.OriginDirPath + origin
	fstype, err := checkOriginType(originPath)
	if err != nil {
		// TEST: TestUnmountInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v", err)
	}

	// TODO: check mountpoint
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

		err = util.ZipFile(tempPath, originPath)
		if err != nil {
			// TEST: TestUnmountLZFSZip
			return globals.ExitZip,
				fmt.Errorf("LZFS file zipping failed: %v", err)
		}

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
	}
	err = config.CommonConfig.UnmountFilesystem(mountpoint)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with unmounting Filesystem from Config: %v", err)
	} else {
		err = config.CommonConfig.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	return 0, nil
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
