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
	return base[strings.Index(base, "."):]
}

// CheckDirOrZip
// - check if "dir" exists and is a directory
// - check if it's not directory but is a zip file
func CheckDirOrZip(dirOrZip string) (config.FSType, error) {
	// HACK
	fi, err := os.Stat(dirOrZip)
	if err != nil {
		return config.FSHack, err
	}

	if !fi.IsDir() {
		zips := map[string]int{
			".zip":     0,
			".tar":     1,
			".tar.gz":  2,
			".tar.bz2": 3,
		}

		ext := getExt(dirOrZip)

		_, ok := zips[ext]
		if ok {
			return config.FSZip, nil
		}

		return config.FSNone,
			fmt.Errorf("%s isn't a directory and isn't a zip file",
				dirOrZip)
	}
	return config.FSLoopback, nil
}
