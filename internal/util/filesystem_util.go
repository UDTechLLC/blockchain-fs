package util

import (
	"fmt"
	"os"
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
