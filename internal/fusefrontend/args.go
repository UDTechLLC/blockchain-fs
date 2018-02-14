package fusefrontend

// Args is a container for arguments that are passed from main() to fusefrontend
type Args struct {
	// Origin is the backing storage directory (absolute path).
	// For reverse mode, Origin actually contains *plaintext* files.
	Origin string
	// Type of Origin:
	// 1 - directory (LoopbackFS)
	// 2 - zip file (ZipFS)
	Type uint16
}
