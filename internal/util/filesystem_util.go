package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"bitbucket.org/udt/wizefs/internal/config"
)

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// CheckDir - check if "dir" exists and is a directory
func CheckDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

func getExt(file string) string {
	base := filepath.Base(file)
	idx := strings.Index(base, ".")
	if idx == -1 {
		return ""
	}
	return base[idx:]
}

func isZipFS(file string) bool {
	base := filepath.Base(file)
	return base[0] == '_'
}

// CheckDirOrZip
// - check if "dir" exists and is a directory
// - check if it's not directory but is a zip file

// TEST: TestUtilCheckDirOrZip
func CheckDirOrZip(dirOrZip string) (config.FSType, error) {
	zips := map[string]int{
		".zip":     0,
		".tar":     1,
		".tar.gz":  2,
		".tar.bz2": 3,
	}

	ext := getExt(dirOrZip)

	_, ok := zips[ext]
	if ok {
		if isZipFS(dirOrZip) {
			return config.ZipFS, nil
		}
		return config.LZFS, nil
	}

	fi, err := os.Stat(dirOrZip)
	if err != nil {
		// HACK
		return config.HackFS, err
	}

	if !fi.IsDir() {
		return config.NoneFS,
			fmt.Errorf("%s isn't a directory and isn't a zip file",
				dirOrZip)
	}

	return config.LoopbackFS, nil
}
