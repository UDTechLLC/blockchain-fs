package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"bitbucket.org/udt/wizefs/internal/config"
	"bitbucket.org/udt/wizefs/internal/globals"
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
// TODO: Output? JSON? result
// TODO: check permissions
func CmdPutFile(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 2)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	originalFile := c.Args()[0]
	origin := c.Args()[1]

	return ApiPut(originalFile, origin, nil)
}

func ApiPut(originalFile, origin string, content []byte) (err error) {
	var mountpointPath string
	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: add more information about errors
	// TODO: auto-mount if (Filesystem is not mounted!)?

	// TODO: HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
	}

	mountpointPath, err = config.CommonConfig.CheckOriginGetMountpoint(origin)
	if err != nil {
		return cli.NewExitError(
			"Did not find MOUNTPOINT in common config.",
			globals.MountPoint)
	}

	// TODO: check MD5?

	if content == nil {
		// check PATH
		if !filepath.IsAbs(originalFile) {
			// TODO: HACK - for temporary testing
			//return cli.NewExitError(
			//	"FILE argument is not absolute path to file.",
			//	globals.Other)

			tlog.Debug.Println("HACK: FILE argument is not absolute path to file.")
			originalFile, _ = filepath.Abs(originalFile)
		}

		// check original file existing
		if _, err = os.Stat(originalFile); os.IsNotExist(err) {
			return cli.NewExitError(
				fmt.Sprintf("Original FILE (%s) does not exist.", originalFile),
				globals.Other)
		}
	}
	originalFileBase := filepath.Base(originalFile)
	// TODO: check file, SIZE, TYPE, etc

	// check destination file existing
	destinationFile := mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(destinationFile); err == nil {
		return cli.NewExitError(
			fmt.Sprintf("Destination FILE (%s) is exist.", destinationFile),
			globals.Other)
	}

	// copy (replace?) file to mountpointPath
	_, err = copyFile(originalFile, destinationFile, content)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("We have a problem with copy file: %v", err),
			globals.Other)
	}

	return nil
}

// wizefs get FILE ORIGIN
// TODO: Output? JSON? result + file data/buffer (binary data in JSON?)
// TODO: check permissions
func CmdGetFile(c *cli.Context) (err error) {
	if c.NArg() != 2 {
		return cli.NewExitError(
			fmt.Sprintf("Wrong number of arguments (have %d, want 2)."+
				" You passed: %s.", c.NArg(), c.Args()),
			globals.Usage)
	}

	originalFile := c.Args()[0]
	origin := c.Args()[1]

	_, err = ApiGet(originalFile, origin, false)
	return err
}

func ApiGet(originalFile, origin string, getContentOnly bool) (content []byte, err error) {
	var mountpointPath string

	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: add more information about errors
	// TODO: auto-mount if (Filesystem is not mounted!)?

	// TODO: HACK for gRPC methods
	if config.CommonConfig == nil {
		config.InitWizeConfig()
		//} else {
		//	config.CommonConfig.Load()
	}

	mountpointPath, err = config.CommonConfig.CheckOriginGetMountpoint(origin)
	if err != nil {
		return nil, cli.NewExitError(
			"Did not find MOUNTPOINT in common config.",
			globals.MountPoint)
	}

	// TODO: check MD5?

	// FIXME: check PATH
	if filepath.IsAbs(originalFile) {
		return nil, cli.NewExitError(
			fmt.Sprintf("FILE argument (%s) is absolute path to file.", originalFile),
			globals.Other)
	}

	// FIXME: get Base?
	originalFileBase := filepath.Base(originalFile)

	// check original file existing
	originalFile = mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(originalFile); os.IsNotExist(err) {
		return nil, cli.NewExitError(
			fmt.Sprintf("Original FILE (%s) does not exist.", originalFile),
			globals.Other)
	}
	// TODO: check file, SIZE, TYPE, etc

	// check destination file existing
	var destinationFile string = ""
	if !getContentOnly {
		// TODO: HACK - we just copy file into application directory
		destinationFile, _ := filepath.Abs(originalFileBase)
		if _, err = os.Stat(destinationFile); os.IsExist(err) {
			return nil, cli.NewExitError(
				fmt.Sprintf("Destination FILE (%s) is exist.", destinationFile),
				globals.Other)
		}
	}

	// copy (replace?) file to mountpointPath
	content, err = copyFile(originalFile, destinationFile, nil)
	if err != nil {
		return nil, cli.NewExitError(
			fmt.Sprintf("We have a problem with copy file: %v", err),
			globals.Other)
	}

	return
}

// wizefs search FILE
// Use Case:
// . wizefs search FILE
// Output? JSON? result + ??? file data/buffer (binary data in JSON?)
func CmdSearchFile(c *cli.Context) {

}

// TODO: optimize coping
// TODO: add replace
func copyFile(origFile, destFile string, origContent []byte) (destContent []byte, err error) {
	var oldFile, newFile *os.File
	var bytes64 int64
	var bytesWritten int

	destContent = nil

	// TODO: to refactor this code
	switch {
	case origContent == nil && destFile != "":
		// Open original file
		oldFile, err = os.Open(origFile)
		defer oldFile.Close()
		if err != nil {
			tlog.Warn.Printf("Open: %v", err)
			return nil, err
		}
		// Create new file
		newFile, err = os.Create(destFile)
		defer newFile.Close()
		if err != nil {
			tlog.Warn.Printf("Create: %v", err)
			return nil, err
		}

		bytes64, err = io.Copy(newFile, oldFile)
		bytesWritten = int(bytes64)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return nil, err
		}

		// Commit the file contents
		// Flushes memory to disk
		err = newFile.Sync()
		if err != nil {
			tlog.Warn.Printf("Sync: %v", err)
			return nil, err
		}

	case origContent != nil && destFile != "":
		// Create new file
		newFile, err = os.Create(destFile)
		defer newFile.Close()
		if err != nil {
			tlog.Warn.Printf("Create: %v", err)
			return nil, err
		}

		bytesWritten, err = newFile.Write(origContent)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return nil, err
		}

		// Commit the file contents
		// Flushes memory to disk
		err = newFile.Sync()
		if err != nil {
			tlog.Warn.Printf("Sync: %v", err)
			return nil, err
		}

	case destFile == "":
		// Open original file
		oldFile, err = os.Open(origFile)
		defer oldFile.Close()
		if err != nil {
			tlog.Warn.Printf("Open: %v", err)
			return nil, err
		}

		destContent, err = ioutil.ReadAll(oldFile)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return nil, err
		}
		bytesWritten = len(destContent)

	default:
	}

	tlog.Debug.Printf("Copied %d bytes.", bytesWritten)

	return
}
