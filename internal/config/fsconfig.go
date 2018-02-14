package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ProgramName              = "wizefs"
	ProgramVersion           = "0.0.4"
	FilesystemCurrentVersion = 1
	FilesystemConfigFilename = "wizefs.conf"
)

type FilesystemConfig struct {
	// Creator is the WizeFS version string
	Creator string `json:"creator"`
	// Version is the version this filesystem uses
	Version uint16 `json:"version"`

	// mountpoint, origin, type*
	//MountPoint string `json:"mountpoint"`
	Origin string `json:"origin"`
	Type   uint16 `json:"type"`

	filename string
}

func NewFilesystemConfig(origin string, itype uint16) *FilesystemConfig {

	originpath, err := filepath.Abs(origin)
	if err != nil {
		originpath = origin
	}

	return &FilesystemConfig{
		Creator: ProgramName + " ver. " + ProgramVersion,
		Version: FilesystemCurrentVersion,
		//MountPoint: mountpoint,
		Origin:   originpath,
		Type:     itype,
		filename: filepath.Join(origin, FilesystemConfigFilename),
	}
}

func (c *FilesystemConfig) Save() error {
	tmp := c.filename + ".tmp"
	// 0400 permissions: wizefs.conf should be kept secret and never be written to.
	fd, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return err
	}
	js, err := json.MarshalIndent(c, "", "\t")
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
	err = os.Rename(tmp, c.filename)
	return err
}

// TODO: load FilesystemConfig
