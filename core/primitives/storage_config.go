package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/tlog"
)

const (
	StorageConfigVersion  = 1
	StorageConfigFilename = "wizedb.conf"
)

type FilesystemInfo struct {
	OriginPath    string         `json:"originpath"`
	Type          globals.FSType `json:"type"`
	MountpointKey string         `json:"mountpoint"`
}

type MountpointInfo struct {
	MountpointPath string `json:"mountpointpath"`
	OriginKey      string `json:"origin"`
}

type StorageConfig struct {
	// header

	// filesystems
	Filesystems map[string]FilesystemInfo `json:"created"`
	Mountpoints map[string]MountpointInfo `json:"mounted"`

	filename string
	mutex    sync.Mutex
}

// TEST: TestWizeConfigMake
func NewStorageConfig(path string) *StorageConfig {
	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		path = filepath.Dir(exe)
	}

	return &StorageConfig{
		Filesystems: make(map[string]FilesystemInfo),
		Mountpoints: make(map[string]MountpointInfo),
		filename:    filepath.Join(path, StorageConfigFilename),
	}
}

// TESTS: TestWizeConfig* (several tests)
func (wc *StorageConfig) CreateFilesystem(origin, originPath string, itype globals.FSType) error {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	wc.Load()

	_, ok := wc.Filesystems[origin]
	if ok {
		return fmt.Errorf("This filesystem is already added!")
	}

	wc.Filesystems[origin] = FilesystemInfo{
		OriginPath:    originPath,
		Type:          itype,
		MountpointKey: "",
	}

	tlog.Debug.Println("Add filesystem to the created map! ", wc)

	return nil
}

func (wc *StorageConfig) DeleteFilesystem(origin string) error {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	wc.Load()

	_, ok := wc.Filesystems[origin]
	if !ok {
		return fmt.Errorf("This filesystem is absent!")
	}

	delete(wc.Filesystems, origin)

	tlog.Debug.Println("Delete filesystem from the created map! ", wc)

	return nil
}

func (wc *StorageConfig) MountFilesystem(origin, mountpoint, mountpointpath string) error {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	// HACK2: for gRPC/Mount we don't need to Load() config, because it works via mount CLI app
	//wc.Load()

	_, ok := wc.Mountpoints[mountpoint]
	if ok {
		return fmt.Errorf("This filesystem is already mounted!")
	}

	wc.Mountpoints[mountpoint] = MountpointInfo{
		MountpointPath: mountpointpath,
		OriginKey:      origin,
	}

	fsi := wc.Filesystems[origin]
	wc.Filesystems[origin] = FilesystemInfo{
		OriginPath:    fsi.OriginPath,
		Type:          fsi.Type,
		MountpointKey: mountpoint,
	}

	tlog.Debug.Println("Add filesystem to the mounted map! ", wc)

	return nil
}

func (wc *StorageConfig) UnmountFilesystem(mountpoint string) error {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	wc.Load()

	mpi, ok := wc.Mountpoints[mountpoint]
	if !ok {
		return fmt.Errorf("This filesystem is not mounted!")
	}

	origin := mpi.OriginKey
	delete(wc.Mountpoints, mountpoint)

	fsi := wc.Filesystems[origin]
	wc.Filesystems[origin] = FilesystemInfo{
		OriginPath:    fsi.OriginPath,
		Type:          fsi.Type,
		MountpointKey: "",
	}

	tlog.Debug.Println("Delete filesystem from the mounted map! ", wc)

	return nil
}

func (wc *StorageConfig) Save() error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	tmp := wc.filename + ".tmp"
	// 0400 permissions: wizefs.conf should be kept secret and never be written to.
	//fd, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	// temporary solution
	fd, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	js, err := json.MarshalIndent(wc, "", "\t")
	if err != nil {
		return err
	}
	// For convenience for the user, add a newline at the end.
	js = append(js, '\n')
	_, err = fd.Write(js)
	if err != nil {
		return err
	}
	err = fd.Sync()
	if err != nil {
		return err
	}
	err = fd.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tmp, wc.filename)
	return err
}

func (wc *StorageConfig) Load() error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	// Read from disk
	js, err := ioutil.ReadFile(wc.filename)
	if err != nil {
		//tlog.Warn.Printf("Load config file: ReadFile: %v, %s\n", err, err.Error())
		return err
	}

	wc.clear()

	// Unmarshal
	err = json.Unmarshal(js, &wc)
	if err != nil {
		tlog.Warn.Printf("Failed to unmarshal config file")
		return err
	}

	return nil
}

func (wc *StorageConfig) CheckOriginGetMountpoint(origin string) (mountpointPath string, err error) {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	wc.Load()

	var ok bool
	var fsinfo FilesystemInfo
	fsinfo, ok = wc.Filesystems[origin]
	if !ok {
		tlog.Warn.Printf("Filesystem %s is not exist!", origin)
		return "", errors.New("Filesystem is not exist!")
	}

	if fsinfo.MountpointKey == "" {
		tlog.Warn.Printf("Filesystem %s is not mounted!", origin)
		return "", errors.New("Filesystem is not mounted!")
	}

	var mpinfo MountpointInfo
	mpinfo, ok = wc.Mountpoints[fsinfo.MountpointKey]
	if !ok {
		tlog.Warn.Printf("Mounted filesystem %s is not exist!", fsinfo.MountpointKey)
		return "", errors.New("Mounted filesystem is not exist!")
	}

	mountpointPath = mpinfo.MountpointPath
	// TODO: check mountpointPath?

	return mountpointPath, nil
}

func (wc *StorageConfig) Check(origin string, shouldFindOrigin, shouldMounted bool) (extCode int, err error) {
	wc.Load()

	existOrigin, existMountpoint := wc.checkFilesystem(origin)

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

// just for WizeFS UI app
func (wc *StorageConfig) GetInfoByOrigin(origin string) (fsinfo FilesystemInfo, mpinfo MountpointInfo, err error) {
	// HACK: this fixed problems with gRPC methods (and GUI?)
	wc.Load()

	fsinfo, ok := wc.Filesystems[origin]
	if !ok {
		tlog.Warn.Printf("Filesystem %s is not exist!", origin)
		return FilesystemInfo{}, MountpointInfo{}, errors.New("Filesystem is not exist!")
	}

	mpinfo, ok = wc.Mountpoints[fsinfo.MountpointKey]
	if !ok {
		tlog.Warn.Printf("Mounted filesystem %s is not exist!", fsinfo.MountpointKey)
		return fsinfo, MountpointInfo{}, errors.New("Mounted filesystem is not exist!")
	}

	return fsinfo, mpinfo, nil
}

func (wc *StorageConfig) clear() {
	// Just clear WizeConfig maps
	for k := range wc.Filesystems {
		delete(wc.Filesystems, k)
	}
	for k := range wc.Mountpoints {
		delete(wc.Mountpoints, k)
	}
}

func (wc *StorageConfig) checkFilesystem(origin string) (existOrigin bool, existMountpoint bool) {
	existMountpoint = false
	fsinfo, existOrigin := wc.Filesystems[origin]
	if existOrigin {
		if fsinfo.MountpointKey != "" {
			existMountpoint = true
		}
	}
	return
}
