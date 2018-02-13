package fusefrontend

// Args is a container for arguments that are passed from main() to fusefrontend
type Args struct {
	// Origindir is the backing storage directory (absolute path).
	// For reverse mode, Origindir actually contains *plaintext* files.
	Origindir string
}
