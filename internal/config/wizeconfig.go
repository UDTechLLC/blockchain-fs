package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/udt/wizefs/internal/tlog"
)

const (
	WizeCurrentVersion = 1
	WizeConfigFilename = "wizedb.conf"
)

type FilesystemInfo struct {
	// mountpoint, origindir, type*
	//OriginDir  string `json:"origindir"`
	MountPoint string `json:"mountpoint"`
	Type       uint16 `json:"type"`
}

type WizeConfig struct {
	// header

	// filesystems
	Filesystems map[string]FilesystemInfo `json:"filesystems"`

	filename string
}

var CommonConfig *WizeConfig

func init() {
	CommonConfig = NewWizeConfig()
	err := CommonConfig.Load()
	if err != nil {
		CommonConfig.Save()
	}
}

func NewWizeConfig() *WizeConfig {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(exe)

	return &WizeConfig{
		Filesystems: make(map[string]FilesystemInfo),
		filename:    filepath.Join(exePath, WizeConfigFilename),
	}
}

func (wc *WizeConfig) AddFilesystem(origindir, mountpoint string, itype uint16) error {
	_, ok := wc.Filesystems[origindir]
	if ok {
		tlog.Warn.Println("This filesystem is already added!")
		return errors.New("This filesystem is already added!")
	}

	wc.Filesystems[origindir] = FilesystemInfo{
		//OriginDir: origindir,
		MountPoint: mountpoint,
		Type:       itype,
	}

	tlog.Debug.Println("Add filesystem to the List! ", wc)

	return nil
}

func (wc *WizeConfig) DeleteFilesystem(mountpoint string) error {
	// Find origindir by mountpoint
	var origindir string
	for key, value := range wc.Filesystems {
		if value.MountPoint == mountpoint {
			origindir = key
			break
		}
	}
	if origindir == "" {
		return errors.New("OriginDir was not find in the common configuraion!")
	}

	_, ok := wc.Filesystems[origindir]
	if ok {
		delete(wc.Filesystems, origindir)
	}

	tlog.Debug.Println("Delete filesystem from the List! ", wc)

	return nil
}

func (wc *WizeConfig) Save() error {
	tmp := wc.filename + ".tmp"
	// 0400 permissions: wizefs.conf should be kept secret and never be written to.
	fd, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
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

	// Unmarshal
	err = json.Unmarshal(js, &wc)
	if err != nil {
		tlog.Warn.Printf("Failed to unmarshal config file")
		return err
	}

	return nil
}
