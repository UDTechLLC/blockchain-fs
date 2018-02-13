package config

type FilesystemInfo struct {
	// mountpoint, origindir, type*
	MountPoint string `json:"mountpoint"`
	OriginDir  string `json:"origindir"`
	Type       uint16 `json:"type"`
}

type WizeConfig struct {
	// header

	// filesystems
	Filesystems map[int]*FilesystemInfo

	filename string
}

// TODO: adding filesystem (FilesystemInfo) into Filesystems map
// TODO: save WizeConfig
