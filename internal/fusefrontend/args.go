package fusefrontend

// Args is a container for arguments that are passed from main() to fusefrontend
type Args struct {
	// Cipherdir is the backing storage directory (absolute path).
	// For reverse mode, Cipherdir actually contains *plaintext* files.
	Cipherdir string
}
