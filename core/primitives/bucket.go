package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/udt/wizefs/core/globals"
	"bitbucket.org/udt/wizefs/core/tlog"
)

type BucketApi interface {
	PutFile(originalFile string, content []byte) (exitCode int, err error)
	GetFile(originalFile, destinationFilePath string, getContentOnly bool) (content []byte, exitCode int, err error)
	RemoveFile(originalFile string) (exitCode int, err error)
}

type Bucket struct {
	storage    *Storage
	Origin     string
	MountPoint string
	Config     *BucketConfig
	mounted    bool
}

func NewBucket(s *Storage, origin, originPath string, fstype globals.FSType) *Bucket {
	bucket := &Bucket{
		storage:    s,
		Origin:     origin,
		MountPoint: "",
		mounted:    false,
	}

	bucket.Config = NewBucketConfig(origin, originPath, fstype)
	err := bucket.Config.Load()
	if err != nil {
		bucket.Config.Save()
	}

	return bucket
}

func (b *Bucket) IsMounted() bool {
	return b.mounted
}

func (b *Bucket) SetMounted(value bool) {
	b.mounted = value
}

func (b *Bucket) PutFile(originalFile string, content []byte) (exitCode int, err error) {
	// TEST: TestPutNotExistingOrigin, TestPutNotMounted
	exitCode, err = b.storage.Config.Check(b.Origin, false, false)
	if err != nil {
		return
	}

	var mountpointPath string
	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: auto-mount if (Filesystem is not mounted!)?

	// TODO: HACK for gRPC methods
	if b.storage.Config == nil {
		tlog.Info.Println("CommonConfig == nil")
		//config.InitWizeConfig()
	}

	mountpointPath, err = b.storage.Config.CheckOriginGetMountpoint(b.Origin)
	if err != nil {
		// TEST: TestPutFailedMountpointPath
		return globals.ExitMountPoint,
			fmt.Errorf("Did not find MOUNTPOINT in common config.")
	}

	// content is used for gRPC methods
	if content == nil {
		// check PATH
		// TEST: TestPutFullFilename, TestPutShortFilename
		if !filepath.IsAbs(originalFile) {
			//return cli.NewExitError(
			//	"FILE argument is not absolute path to file.",
			//	globals.Other)

			tlog.Debug.Println("HACK: FILE argument is not absolute path to file.")

			// HACK: for temporary testing
			originalFile, _ = filepath.Abs(originalFile)
		}

		// check original file existing
		if _, err = os.Stat(originalFile); os.IsNotExist(err) {
			// TEST: TestPutNotExistingFile
			return globals.ExitFile,
				fmt.Errorf("Original FILE (%s) does not exist.", originalFile)
		}
	}
	originalFileBase := filepath.Base(originalFile)

	// check destination file existing
	destinationFile := mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(destinationFile); err == nil {
		// TEST: TestPutExistingDestinationFile
		return globals.ExitFile,
			fmt.Errorf("Destination FILE (%s) is exist.", destinationFile)
	}

	// copy (replace?) file to mountpointPath
	_, err = b.copyFile(originalFile, destinationFile, content)
	if err != nil {
		// TEST: TestPutFailedCopyFile
		return globals.ExitFile,
			fmt.Errorf("We have a problem with copy file: %v", err)
	}

	return 0, nil
}

func (b *Bucket) GetFile(originalFile, destinationFilePath string, getContentOnly bool) (content []byte, exitCode int, err error) {
	// TEST: TestGetNotExistingOrigin, TestGettNotMounted
	exitCode, err = b.storage.Config.Check(b.Origin, false, false)
	if err != nil {
		return
	}

	var mountpointPath string
	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: auto-mount if (Filesystem is not mounted!)?

	// TODO: HACK for gRPC methods
	if b.storage.Config == nil {
		tlog.Info.Println("CommonConfig == nil")
		//config.InitWizeConfig()
	}

	mountpointPath, err = b.storage.Config.CheckOriginGetMountpoint(b.Origin)
	if err != nil {
		// TEST: TestGetFailedMountpointPath
		return nil, globals.ExitMountPoint,
			fmt.Errorf("Did not find MOUNTPOINT in common config.")
	}

	// FIXME: check PATH
	if filepath.IsAbs(originalFile) {
		// TEST: TestGetFullFilename, TestGetShortFilename
		return nil, globals.ExitFile,
			fmt.Errorf("FILE argument (%s) is absolute path to file.", originalFile)
	}

	// FIXME: get Base?
	originalFileBase := filepath.Base(originalFile)

	// check original file existing
	originalFile = mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(originalFile); os.IsNotExist(err) {
		// TEST: TestGetNotExistingFile
		return nil, globals.ExitFile,
			fmt.Errorf("Original FILE (%s) does not exist.", originalFile)
	}

	// check destination file existing
	var destinationFile string = ""
	if !getContentOnly {
		if destinationFilePath != "" {
			destinationFile = destinationFilePath
		} else {
			// TODO: HACK - we just copy file into application directory
			destinationFile, _ = filepath.Abs(originalFileBase)
		}
		if _, err = os.Stat(destinationFile); os.IsExist(err) {
			// TEST: TestGetExistingDestinationFile
			return nil, globals.ExitFile,
				fmt.Errorf("Destination FILE (%s) is exist.", destinationFile)
		}
	}

	// copy (replace?) file to mountpointPath
	content, err = b.copyFile(originalFile, destinationFile, nil)
	if err != nil {
		// TEST: TestGetFailedCopyFile
		return nil, globals.ExitFile,
			fmt.Errorf("We have a problem with copy file: %v", err)
	}

	return content, 0, nil
}

func (b *Bucket) RemoveFile(originalFile string) (exitCode int, err error) {
	// TEST: TestRemoveNotExistingOrigin, TestRemoveNotMounted
	exitCode, err = b.storage.Config.Check(b.Origin, false, false)
	if err != nil {
		return
	}

	var mountpointPath string
	//originPath := OriginDir + origin

	// check origin via config file (database) and get mountpoint if it exists
	// TODO: auto-mount if (Filesystem is not mounted!)?

	// TODO: HACK for gRPC methods
	if b.storage.Config == nil {
		tlog.Info.Println("CommonConfig == nil")
		//config.InitWizeConfig()
	}

	mountpointPath, err = b.storage.Config.CheckOriginGetMountpoint(b.Origin)
	if err != nil {
		// TEST: TestRemoveFailedMountpointPath
		return globals.ExitMountPoint,
			fmt.Errorf("Did not find MOUNTPOINT in common config.")
	}

	// FIXME: check PATH
	if filepath.IsAbs(originalFile) {
		// TEST: TestRemoveFullFilename, TestRemoveShortFilename
		return globals.ExitFile,
			fmt.Errorf("FILE argument (%s) is absolute path to file.", originalFile)
	}

	// FIXME: get Base?
	originalFileBase := filepath.Base(originalFile)

	// check original file existing
	originalFile = mountpointPath + "/" + originalFileBase
	if _, err = os.Stat(originalFile); os.IsNotExist(err) {
		// TEST: TestRemoveNotExistingFile
		return globals.ExitFile,
			fmt.Errorf("Original FILE (%s) does not exist.", originalFile)
	}

	// remove file from mountpointPath
	err = os.Remove(originalFile)
	if err != nil {
		// TEST: TestRemoveFailedRemoveFile
		return globals.ExitFile,
			fmt.Errorf("We have a problem with removing file: %v", err)
	}

	return 0, nil
}

// TODO: add replace
// TEST: TestCopyFile (several tests)
func (b Bucket) copyFile(origFile, destFile string, origContent []byte) (destContent []byte, err error) {
	var oldFile, newFile *os.File
	var bytes64 int64
	var bytesWritten int

	destContent = nil

	// CLI & gRPC/Get
	if origContent == nil {
		// Open original file
		oldFile, err = os.Open(origFile)
		defer oldFile.Close()
		if err != nil {
			tlog.Warn.Printf("Open: %v", err)
			return
		}
	}

	// CLI & gRPC/Put
	if destFile != "" {
		// Create new file
		newFile, err = os.Create(destFile)
		defer newFile.Close()
		if err != nil {
			tlog.Warn.Printf("Create: %v", err)
			return
		}
	}

	switch {
	// CLI
	case origContent == nil && destFile != "":
		bytes64, err = io.Copy(newFile, oldFile)
		bytesWritten = int(bytes64)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return
		}

	// gRPC/Put
	case origContent != nil && destFile != "":
		bytesWritten, err = newFile.Write(origContent)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return
		}

	// gRPC/Get
	case destFile == "":
		destContent, err = ioutil.ReadAll(oldFile)
		if err != nil {
			tlog.Warn.Printf("Copy: %v", err)
			return
		}
		bytesWritten = len(destContent)

	default:
	}

	// CLI & gRPC/Put
	if destFile != "" {
		// Commit the file contents
		// Flushes memory to disk
		err = newFile.Sync()
		if err != nil {
			tlog.Warn.Printf("Sync: %v", err)
			return
		}
	}

	tlog.Debug.Printf("Copied %d bytes.", bytesWritten)

	return
}
