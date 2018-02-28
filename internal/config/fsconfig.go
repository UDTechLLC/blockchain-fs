package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"bitbucket.org/udt/wizefs/internal/tlog"
)

const (
	ProgramName              = "wizefs"
	ProgramVersion           = "0.0.7"
	FilesystemCurrentVersion = 1
	FilesystemConfigFilename = "wizefs.conf"
)

type FSType int

const (
	HackFS FSType = iota - 1 // -1
	NoneFS
	LoopbackFS
	ZipFS
	LZFS
)

type FilesystemConfig struct {
	// Creator is the WizeFS version string
	Creator string `json:"creator"`
	// Version is the version this filesystem uses
	Version    uint16 `json:"version"`
	Origin     string `json:"origin"`
	OriginPath string `json:"originpath"`
	Type       FSType `json:"type"`

	filename string
}

// TEST: TestFSConfigMake
func NewFilesystemConfig(origin, originPath string, itype FSType) *FilesystemConfig {
	return &FilesystemConfig{
		Creator:    ProgramName + " ver. " + ProgramVersion,
		Version:    FilesystemCurrentVersion,
		Origin:     origin,
		OriginPath: originPath,
		Type:       itype,
		filename:   filepath.Join(originPath, FilesystemConfigFilename),
	}
}

// TEST: TestFSConfigSave
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
	tlog.Debug.Printf("Save config file %s", c.filename)
	return err
}

// TEST: TestFSConfigLoad
// TODO: load FilesystemConfig
