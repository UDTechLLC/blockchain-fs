// Package fusefrontend interfaces directly with the go-fuse library.
package fusefrontend

// FUSE operations on paths

import (
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// FS implements the go-fuse virtual filesystem interface.
type FS struct {
	pathfs.FileSystem      // loopbackFileSystem, see go-fuse/fuse/pathfs/loopback.go
	args              Args // Stores configuration arguments
}

var _ pathfs.FileSystem = &FS{} // Verify that interface is implemented.

// NewFS returns a new encrypted FUSE overlay filesystem.
func NewFS(args Args) *FS {
	return &FS{
		FileSystem: pathfs.NewLoopbackFileSystem(args.Cipherdir),
		args:       args,
	}
}
