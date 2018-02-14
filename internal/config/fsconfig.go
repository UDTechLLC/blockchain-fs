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

	// mountpoint, origindir, type*
	//MountPoint string `json:"mountpoint"`
	OriginDir string `json:"origindir"`
	Type      uint16 `json:"type"`

	filename string
}

func NewFilesystemConfig(origindir string, itype uint16) *FilesystemConfig {

	originpath, err := filepath.Abs(origindir)
	if err != nil {
		originpath = origindir
	}

	return &FilesystemConfig{
		Creator: ProgramName + " ver. " + ProgramVersion,
		Version: FilesystemCurrentVersion,
		//MountPoint: mountpoint,
		OriginDir: originpath,
		Type:      itype,
		filename:  filepath.Join(origindir, FilesystemConfigFilename),
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
