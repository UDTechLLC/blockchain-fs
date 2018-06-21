package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/tlog"
)

const (
	BucketConfigVersion  = 1
	BucketConfigFilename = "wizefs.conf"
)

type BucketConfig struct {
	// Creator is the WizeFS version string
	Creator string `json:"creator"`
	// Version is the version this filesystem uses
	Version    uint16         `json:"version"`
	Origin     string         `json:"origin"`
	OriginPath string         `json:"originpath"`
	Type       globals.FSType `json:"type"`

	filename string
	mutex    sync.Mutex
}

// TEST: TestBucketConfigMake
func NewBucketConfig(origin, originPath string, itype globals.FSType) *BucketConfig {
	return &BucketConfig{
		Creator:    globals.ProjectName + " ver. " + globals.ProjectVersion,
		Version:    BucketConfigVersion,
		Origin:     origin,
		OriginPath: originPath,
		Type:       itype,
		filename:   filepath.Join(originPath, BucketConfigFilename),
	}
}

// TEST: TestBucketConfigSave
func (c *BucketConfig) Save() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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
	tlog.Debugf("Save config file %s", c.filename)
	return err
}

// TEST: TestBucketConfigLoad
// TODO: load BucketConfig
func (c *BucketConfig) Load() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Read from disk
	js, err := ioutil.ReadFile(c.filename)
	if err != nil {
		tlog.Warnf("Load config file: ReadFile: %v, %s\n", err, err.Error())
		return err
	}

	// Unmarshal
	err = json.Unmarshal(js, &c)
	if err != nil {
		tlog.Warnf("Failed to unmarshal config file")
		return err
	}

	return nil
}
