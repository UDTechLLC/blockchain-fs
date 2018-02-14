package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
func CheckDirOrZip(dirOrZip string) (uint16, error) {
	fi, err := os.Stat(dirOrZip)
	if err != nil {
		return 0, err
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
			return 2, nil
		}

		return 0, errors.New(
			fmt.Sprintf("%s isn't a directory and isn't a zip file", dirOrZip))
	}
	return 1, nil
}
