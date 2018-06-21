package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/tlog"
	"bitbucket.org/udt/wizefs/core/util"
)

const (
	storageDirPath = "/.local/share/wize/fs/"
)

type StorageApi interface {
	Create(origin string) (exitCode int, err error)
	Delete(origin string) (exitCode int, err error)
	Mount(origin string, notifypid int) (exitCode int, err error)
	Unmount(origin string) (exitCode int, err error)
}

type Storage struct {
	DirPath string
	Config  *StorageConfig
	buckets map[string]*Bucket
}

func NewStorage() *Storage {
	storage := &Storage{
		buckets: make(map[string]*Bucket),
	}

	storage.DirPath = storage.userHomeDir() + storageDirPath
	storage.Config = NewStorageConfig(storage.DirPath)
	err := storage.Config.Load()
	if err != nil {
		storage.Config.Save()
	}

	// Now we just read WizeConfig and set Storate info and buckets
	for origin, fsinfo := range storage.Config.Filesystems {
		storage.buckets[origin] = NewBucket(storage, origin, fsinfo.OriginPath, fsinfo.Type)
		if fsinfo.MountpointKey != "" {
			storage.buckets[origin].MountPoint = fsinfo.MountpointKey
			storage.buckets[origin].mounted = true
		}
	}
	//for origin, fsinfo := range s.config.Mountpoints {
	//}

	return storage
}

func (s *Storage) String() string {
	return fmt.Sprintf("Path: %s, Buckets count: %d", s.DirPath, len(s.buckets))
}

func (s *Storage) Bucket(origin string) (*Bucket, bool) {
	bucket, ok := s.buckets[origin]
	return bucket, ok
}

func (s *Storage) MountedBuckets() map[string]*Bucket {
	buckets := make(map[string]*Bucket)
	for origin, bucket := range s.buckets {
		if bucket.mounted {
			buckets[origin] = bucket
		}
	}
	return buckets
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
	if fstype == globals.ZipFS {
		// TEST: TestCreateZipFS (like _archive.zip)
		return globals.ExitOrigin,
			fmt.Errorf("Creating zip files are not supported now")
	}
	if fstype == globals.LZFS {
		originPath = s.DirPath +
			"temp/" + strings.Replace(origin, ".", "_", -1)
	}

	tlog.Debugf("Creating new Filesystem %s on path %s...\n", origin, originPath)

	// create Directory if it's not exist
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		tlog.Debugf("Create new directory: %s", originPath)
		os.MkdirAll(originPath, 0755)
	} else {
		// TODO: what we should done when origin is exist already?
		return globals.ExitOrigin,
			fmt.Errorf("Directory %s is exist already!", originPath)
	}

	// create LZFS archive
	if fstype == globals.LZFS {
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
	if s.Config == nil {
		tlog.Info("CommonConfig == nil")
		//config.InitWizeConfig()
	}
	err = s.Config.CreateFilesystem(origin, originPath, fstype)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with adding Filesystem to Config: %v", err)
	} else {
		err = s.Config.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	// Adding to Buckets
	s.buckets[origin] = NewBucket(s, origin, originPath, fstype)

	return 0, nil
}

func (s *Storage) Delete(origin string) (exitCode int, err error) {
	// TEST: TestDeleteNotExistingOrigin, TestDeleteAlreadyMounted
	exitCode, err = s.Config.Check(origin, false, true)
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
	if fstype == globals.ZipFS {
		// TEST: TestDeleteZipFS
		return globals.ExitOrigin,
			fmt.Errorf("Deleting zip files are not support now")
	}

	tlog.Debugf("Delete existing Filesystem: %s", origin)

	// delete Directory if it's exist
	// TODO: check permissions
	if _, err := os.Stat(originPath); os.IsNotExist(err) {
		// TODO: what we should done when origin is exist already?
		tlog.Warnf("Directory %s is not exist!", originPath)
	} else {
		tlog.Debugf("Delete existing directory: %s", originPath)
		os.RemoveAll(originPath)
	}

	// TODO: HACK - get mountpoint internally
	mountpoint := s.getMountpoint(origin, fstype)
	mountpointPath := s.DirPath + mountpoint

	if _, err := os.Stat(mountpointPath); os.IsNotExist(err) {
		//tlog.Warn.Printf("Directory %s is not exist!", mountpointPath)
	} else {
		tlog.Debugf("Delete existing directory: %s", mountpointPath)
		os.RemoveAll(mountpointPath)
	}

	// TODO: HACK for gRPC methods
	if s.Config == nil {
		tlog.Info("CommonConfig == nil")
		//config.InitWizeConfig()
	}
	err = s.Config.DeleteFilesystem(origin)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with deleting Filesystem from Config: %v", err)
	} else {
		err = s.Config.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	// Removing from Buckets
	_, ok := s.buckets[origin]
	if ok {
		delete(s.buckets, origin)
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

	if fstype == globals.LZFS {
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
		tlog.Debugf("Create new directory: %s", mountpointPath)
		os.MkdirAll(mountpointPath, 0755)
	} else {
		tlog.Warnf("Directory %s is exist already!", mountpointPath)
	}

	tlog.Debugf("Mount Filesystem %s into %s", originPath, mountpointPath)

	// Do mounting with options
	exitCode, err = s.doMount(fstype, origin, originPath, mountpoint, mountpointPath, notifypid)
	if exitCode != 0 || err != nil {
		//os.Exit(exitCode)
		return exitCode, err
	}

	// FIXME: Mounting the Bucket
	s.buckets[origin].mounted = true
	s.buckets[origin].MountPoint = mountpoint

	// Don't call os.Exit on success to give deferred functions a chance to
	// run
	return 0, nil
}

func (s *Storage) Unmount(origin string) (exitCode int, err error) {
	// TEST: TestUnmountNotExistingOrigin, TestUnmountNotMounted
	exitCode, err = s.Config.Check(origin, false, false)
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

	tlog.Debugf("Unmount Filesystem %s", mountpointPath)

	err = s.doUnmount(mountpointPath)
	if err != nil {
		// TEST: TestUnmount
		return globals.ExitMountPoint,
			fmt.Errorf("doUnmount failed: %v", err)
	}

	if fstype == globals.LZFS {
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
		tlog.Warnf("Directory %s is not exist!", mountpointPath)
	} else {
		tlog.Debugf("Delete existing directory: %s", mountpointPath)
		os.RemoveAll(mountpointPath)
	}

	// TODO: HACK for gRPC methods
	if s.Config == nil {
		tlog.Info("CommonConfig == nil")
		//config.InitWizeConfig()
	}
	err = s.Config.UnmountFilesystem(mountpoint)
	if err != nil {
		return globals.ExitChangeConf,
			fmt.Errorf("Problem with unmounting Filesystem from Config: %v", err)
	} else {
		err = s.Config.Save()
		if err != nil {
			return globals.ExitSaveConf,
				fmt.Errorf("Problem with saving Config: %v", err)
		}
	}

	// Unmounting the Bucket
	s.buckets[origin].mounted = false
	s.buckets[origin].MountPoint = ""

	return 0, nil
}

func (s Storage) checkOriginType(origin string) (fstype globals.FSType, err error) {
	fstype, err = s.checkDirOrZip(origin)
	if err != nil {
		// HACK: if fstype = globals.HackFS
		if fstype == globals.HackFS {
			return globals.LoopbackFS, nil
		}
		return fstype, err
	}

	return fstype, nil
}

func (s Storage) getMountpoint(origin string, fstype globals.FSType) string {
	mountpoint := origin
	if fstype == globals.ZipFS || fstype == globals.LZFS {
		mountpoint = strings.Replace(mountpoint, ".", "_", -1)
	}
	mountpoint = "_mount" + mountpoint

	return mountpoint
}

func (s Storage) userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// TEST: TestUtilCheckDirOrZip
func (s Storage) checkDirOrZip(dirOrZip string) (globals.FSType, error) {
	// check on zip/tar archive
	zips := map[string]int{
		".zip":     0,
		".tar":     1,
		".tar.gz":  2,
		".tar.bz2": 3,
	}

	getExt := func(file string) string {
		base := filepath.Base(file)
		idx := strings.Index(base, ".")
		if idx == -1 {
			return ""
		}
		return base[idx:]
	}
	ext := getExt(dirOrZip)
	_, ok := zips[ext]
	if ok {
		isZipFS := func(file string) bool {
			base := filepath.Base(file)
			return base[0] == '_'
		}
		if isZipFS(dirOrZip) {
			return globals.ZipFS, nil
		}
		return globals.LZFS, nil
	}

	// check for existing another extension
	if ext != "" {
		return globals.NoneFS,
			fmt.Errorf("%s isn't a directory and isn't a zip file",
				dirOrZip)
	}

	fi, err := os.Stat(dirOrZip)
	if err != nil {
		// HACK: if directory doesn't exist yet
		return globals.HackFS, err
	}

	if !fi.IsDir() {
		return globals.NoneFS,
			fmt.Errorf("%s isn't a directory and isn't a zip file",
				dirOrZip)
	}

	return globals.LoopbackFS, nil
}
