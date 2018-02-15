package fusefrontend

import (
	"bitbucket.org/udt/wizefs/internal/config"
)

// Args is a container for arguments that are passed from main() to fusefrontend
type Args struct {
	// OriginDir is the backing storage directory (absolute path).
	// For reverse mode, OriginDir actually contains *plaintext* files.
	OriginDir string
	// Type of Origin:
	// 1 - directory (LoopbackFS)
	// 2 - zip file (ZipFS)
	Type config.FSType
}
