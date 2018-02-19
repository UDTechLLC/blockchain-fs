package globals

import (
	"os"
	"runtime"
)

// TODO: HACK - temporary solution is:
// to store all ORIGINs and MOUNTPOINTs in one place
var (
	OriginDirPath = userHomeDir() + "/code/test/"
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
