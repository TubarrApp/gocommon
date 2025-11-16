package sharedconsts

import "os"

// Recommended permissions for different types of files and directories.
const (
	// World Readable directories
	PermsGenericDir  os.FileMode = 0o755
	PermsHomeProgDir os.FileMode = 0o700

	// World Readable files
	PermsLogFile os.FileMode = 0o644

	// Private files (may contain sensitive data)
	PermsPrivateFile os.FileMode = 0o600
)
