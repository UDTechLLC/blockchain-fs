package globals

import (
	"os"
	"runtime"
)

const (
	ProjectName    = "WizeFS"
	ProjectVersion = "0.2.1"
)

type FSType int

const (
	HackFS FSType = iota - 1 // -1
	NoneFS
	LoopbackFS
	ZipFS
	LZFS
)

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
