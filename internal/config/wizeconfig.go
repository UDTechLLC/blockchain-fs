package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/udt/wizefs/internal/globals"
	"bitbucket.org/udt/wizefs/internal/tlog"
)

const (
	WizeCurrentVersion = 1
	WizeConfigFilename = "wizedb.conf"
)

type FilesystemInfo struct {
	OriginPath    string `json:"originpath"`
	Type          FSType `json:"type"`
	MountpointKey string `json:"mountpoint"`
}

type MountpointInfo struct {
	MountpointPath string `json:"mountpointpath"`
	OriginKey      string `json:"origin"`
}

type WizeConfig struct {
	// header

	// filesystems
	Filesystems map[string]FilesystemInfo `json:"created"`
	Mountpoints map[string]MountpointInfo `json:"mounted"`

	filename string
}

var CommonConfig *WizeConfig

func init() {
	//
}

func InitWizeConfig() {
	// create Directory if it's not exist
	if _, err := os.Stat(globals.OriginDirPath); os.IsNotExist(err) {
		tlog.Warn.Printf("Create ORIGIN DIR directory: %s", globals.OriginDirPath)
		os.MkdirAll(globals.OriginDirPath, 0755)
	}

	InitWizeConfigWithPath(globals.OriginDirPath)
}

func InitWizeConfigWithPath(path string) {
	CommonConfig = NewWizeConfig(path)
	err := CommonConfig.Load()
	if err != nil {
		CommonConfig.Save()
	}
}

func NewWizeConfig(path string) *WizeConfig {
	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			panic(err)
		}
		path = filepath.Dir(exe)
	}

	return &WizeConfig{
		Filesystems: make(map[string]FilesystemInfo),
		Mountpoints: make(map[string]MountpointInfo),
		filename:    filepath.Join(path, WizeConfigFilename),
	}
}

func (wc *WizeConfig) CreateFilesystem(origin, originPath string, itype FSType) error {
	// HACK: this fixed problems with gRPC methods
	wc.Load()

	_, ok := wc.Filesystems[origin]
	if ok {
		tlog.Warn.Println("This filesystem is already added!")
		return errors.New("This filesystem is already added!")
	}

	wc.Filesystems[origin] = FilesystemInfo{
		OriginPath:    originPath,
		Type:          itype,
		MountpointKey: "",
	}

	tlog.Debug.Println("Add filesystem to the created map! ", wc)

	return nil
}

func (wc *WizeConfig) DeleteFilesystem(origin string) error {
	// HACK: this fixed problems with gRPC methods
	wc.Load()

	_, ok := wc.Filesystems[origin]
	if ok {
		delete(wc.Filesystems, origin)
	}

	tlog.Debug.Println("Delete filesystem from the created map! ", wc)

	return nil
}

func (wc *WizeConfig) MountFilesystem(origin, mountpoint, mountpointpath string) error {
	// HACK: this fixed problems with gRPC methods
	// HACK2: for gRPC/Mount we don't need to Load() config, because it works via mount CLI app
	//wc.Load()

	_, ok := wc.Mountpoints[mountpoint]
	if ok {
		tlog.Warn.Println("This filesystem is already added!")
		return errors.New("This filesystem is already added!")
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

func (wc *WizeConfig) UnmountFilesystem(mountpoint string) error {
	// HACK: this fixed problems with gRPC methods
	wc.Load()

	mpi, ok := wc.Mountpoints[mountpoint]
	if ok {
		origin := mpi.OriginKey
		delete(wc.Mountpoints, mountpoint)

		fsi := wc.Filesystems[origin]
		wc.Filesystems[origin] = FilesystemInfo{
			OriginPath:    fsi.OriginPath,
			Type:          fsi.Type,
			MountpointKey: "",
		}
	}

	tlog.Debug.Println("Delete filesystem from the mounted map! ", wc)

	return nil
}

func (wc *WizeConfig) Save() error {
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

func (wc *WizeConfig) Load() error {
	// Read from disk
	js, err := ioutil.ReadFile(wc.filename)
	if err != nil {
		tlog.Warn.Printf("Load config file: ReadFile: %#v\n", err)
		return err
	}

	// Just clear wc maps
	for k := range wc.Filesystems {
		delete(wc.Filesystems, k)
	}
	for k := range wc.Mountpoints {
		delete(wc.Mountpoints, k)
	}

	// Unmarshal
	err = json.Unmarshal(js, &wc)
	if err != nil {
		tlog.Warn.Printf("Failed to unmarshal config file")
		return err
	}

	return nil
}

func (wc *WizeConfig) CheckOriginGetMountpoint(origin string) (mountpointPath string, err error) {
	// HACK: this fixed problems with gRPC methods
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

func (wc *WizeConfig) CheckFilesystem(origin string) (existOrigin bool, existMountpoint bool) {
	existMountpoint = false
	fsinfo, existOrigin := wc.Filesystems[origin]
	if existOrigin {
		if fsinfo.MountpointKey != "" {
			existMountpoint = true
		}
	}
	return
}

func (wc *WizeConfig) GetMountpointInfoByOrigin(origin string) (fsinfo FilesystemInfo, mpinfo MountpointInfo, err error) {
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
