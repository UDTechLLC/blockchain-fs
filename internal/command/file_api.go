package command

import (
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/exitcodes"
	"bitbucket.org/udt/wizefs/internal/tlog"
	//"bitbucket.org/udt/wizefs/internal/util"
)

// wizefs load FILE ORIGIN -> load FILE [ORIGIN]
// FILE?
// 1. existing file with full path?
// 2. MD5? SIZE? TYPE? other parameters
// Use Case:
// 0. wizefs list (JSON list)
// 1. wizefs create ORIGIN or wizefs find ORIGIN
// 2. wizefs mount ORIGIN {MOUNTPOINT}
// 3. wizefs load FILE ORIGIN
// Output? JSON? result
func CmdPutFile(c *cli.Context) {
	var err error
	var mountpointPath string

	if c.NArg() != 2 {
		tlog.Warn.Printf("Wrong number of arguments (have %d, want 1). You passed: %s",
			c.NArg(), c.Args())
		os.Exit(exitcodes.Usage)
	}

	// TODO: check MD5
	originalFile := c.Args()[0]
	// check PATH
	if !filepath.IsAbs(originalFile) {
		tlog.Warn.Println("FILE argument is not absolute path to file!")
		// HACK
		//os.Exit(exitcodes.Other)
		originalFile, _ = filepath.Abs(originalFile)
	}
	// check original file existing
	if _, err = os.Stat(originalFile); os.IsNotExist(err) {
		tlog.Warn.Println("Origin FILE does not exist!")
		os.Exit(exitcodes.Other)
	}
	originalFileBase := filepath.Base(originalFile)
	// TODO: check file, SIZE, TYPE, etc

	origin := c.Args()[1]
	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: auto-mount if (Filesystem is not mounted!)?
	mountpointPath, err = config.CommonConfig.CheckOriginGetMountpoint(origin)
	if err != nil {
		os.Exit(exitcodes.MountPoint)
	}

	// check destination file existing
	destinationFile := mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(destinationFile); os.IsExist(err) {
		tlog.Warn.Println("Destination FILE is exist!")
		os.Exit(exitcodes.Other)
	}

	// copy (replace?) file to mountpointPath
	err = copyFile(originalFile, destinationFile)
	if err != nil {
		tlog.Warn.Println("We have a problem with copy file!")
		os.Exit(exitcodes.Other)
	}
}

// wizefs get FILE ORIGIN
// Output? JSON? result + file data/buffer (binary data in JSON?)
func CmdGetFile(c *cli.Context) {

}

// wizefs search FILE
// Use Case:
// . wizefs search FILE
// Output? JSON? result + ??? file data/buffer (binary data in JSON?)
func CmdSearchFile(c *cli.Context) {

}

func copyFile(origFile, destFile string) error {
	// Open original file
	originalFile, err := os.Open(origFile)
	defer originalFile.Close()
	if err != nil {
		tlog.Warn.Println(err)
		return err
	}

	// Create new file
	newFile, err := os.Create(destFile)
	defer newFile.Close()
	if err != nil {
		tlog.Warn.Println(err)
		return err
	}

	// Copy the bytes to destination from source
	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {
		tlog.Warn.Println(err)
		return err
	}
	tlog.Debug.Printf("Copied %d bytes.", bytesWritten)

	// Commit the file contents
	// Flushes memory to disk
	err = newFile.Sync()
	if err != nil {
		tlog.Warn.Println(err)
		return err
	}

	return nil
}
