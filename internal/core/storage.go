package core

import (
	"fmt"
	"os"
	"strings"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/globals"
	"bitbucket.org/udt/wizefs/internal/tlog"
	"bitbucket.org/udt/wizefs/internal/util"
)

const (
	storageDirPath = "/.local/share/wize/fs/"
)

type Storage struct {
	DirPath string // globals.OriginDirPath
	//Config        *config.WizeConfig //
}

func NewStorage() *Storage {
	return &Storage{
		DirPath: util.UserHomeDir() + storageDirPath,
	}
}

func (s *Storage) Create(origin string) (exitCode int, err error) {
	//exitCode, err = checkConfig(origin, true, false)
	//if err != nil {
	//	return
	//}

	if origin == "" {
		// TEST: TestCreateInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: ['%s'].", origin)
	}

	originPath := s.DirPath + origin
	fstype, err := s.checkOriginType(originPath)
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
		originPath = s.DirPath +
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
		targetFile := s.DirPath + origin
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

func (s *Storage) Delete(origin string) (exitCode int, err error) {
	// TEST: TestDeleteNotExistingOrigin, TestDeleteAlreadyMounted
	exitCode, err = s.checkConfig(origin, false, true)
	if err != nil {
		return
	}

	originPath := s.DirPath + origin
	fstype, err := s.checkOriginType(originPath)
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
	mountpoint := s.getMountpoint(origin, fstype)
	mountpointPath := s.DirPath + mountpoint

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		//tlog.Warn.Printf("Directory %s is not exist!", mountpointPath)
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

func (s *Storage) Mount(origin string, notifypid int) (exitCode int, err error) {
	originPath := s.DirPath + origin
	fstype, err := s.checkOriginType(originPath)
	if err != nil {
		// TEST: TestMountInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v", err)
	}

	if fstype == config.LZFS {
		// unzip to temp directory - s.DirPath + "temp/" + filename (. -> _)
		tempPath := s.DirPath +
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
	mountpoint := s.getMountpoint(origin, fstype)
	mountpointPath := s.DirPath + mountpoint

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		tlog.Debug.Printf("Create new directory: %s", mountpointPath)
		os.MkdirAll(mountpointPath, 0755)
	} else {
		tlog.Warn.Printf("Directory %s is exist already!", mountpointPath)
	}

	tlog.Debug.Printf("Mount Filesystem %s into %s", originPath, mountpointPath)

	// Do mounting with options
	exitCode, err = util.DoMount(fstype, origin, originPath, mountpoint, mountpointPath, notifypid)
	if exitCode != 0 || err != nil {
		//os.Exit(exitCode)
		return exitCode, err
	}

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return 0, nil
}

func (s *Storage) Unmount(origin string) (exitCode int, err error) {
	// TEST: TestUnmountNotExistingOrigin, TestUnmountNotMounted
	exitCode, err = s.checkConfig(origin, false, false)
	if err != nil {
		return
	}

	originPath := s.DirPath + origin
	fstype, err := s.checkOriginType(originPath)
	if err != nil {
		// TEST: TestUnmountInvalidOrigin
		return globals.ExitOrigin,
			fmt.Errorf("Invalid origin: %v", err)
	}

	// TODO: check mountpoint
	// TODO: HACK - create/get mountpoint internally
	mountpoint := s.getMountpoint(origin, fstype)
	mountpointPath := s.DirPath + mountpoint

	tlog.Debug.Printf("Unmount Filesystem %s", mountpointPath)

	util.DoUnmount(mountpointPath)

	if fstype == config.LZFS {
		// zip temp directory
		tempPath := s.DirPath +
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

func (s Storage) checkOriginType(origin string) (fstype config.FSType, err error) {
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

func (s Storage) getMountpoint(origin string, fstype config.FSType) string {
	mountpoint := origin
	if fstype == config.ZipFS || fstype == config.LZFS {
		mountpoint = strings.Replace(mountpoint, ".", "_", -1)
	}
	mountpoint = "_mount" + mountpoint

	return mountpoint
}

func (s Storage) checkConfig(origin string, shouldFindOrigin, shouldMounted bool) (extCode int, err error) {
	if config.CommonConfig == nil {
		config.InitWizeConfig()
	}

	existOrigin, existMountpoint := config.CommonConfig.CheckFilesystem(origin)

	if shouldFindOrigin {
		if existOrigin {
			return globals.ExitOrigin,
				fmt.Errorf("ORIGIN: %s is already exist in common config.", origin)
		}
	} else {
		if !existOrigin {
			return globals.ExitOrigin,
				fmt.Errorf("Did not find ORIGIN: %s in common config.", origin)
		}
	}

	if shouldMounted {
		if existMountpoint {
			return globals.ExitMountPoint,
				fmt.Errorf("This ORIGIN: %s is already mounted", origin)
		}
	} else {
		if !existMountpoint {
			// TEST: TestUnmountNotMounted
			return globals.ExitMountPoint,
				fmt.Errorf("This ORIGIN: %s is not mounted yet", origin)
		}
	}

	return 0, nil
}
