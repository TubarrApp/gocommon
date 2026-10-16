package sharedconsts

import "os"

// Permissions requested when creating files and directories.
//
// Shared modes leave the final permissions to the process umask (e.g. 0o755/0o644 with the usual 022).
// Private modes are fixed, so a loose umask can't expose them.
const (
	// Shared directories.
	PermsGenericDir os.FileMode = 0o777

	// Shared files.
	PermsLogFile os.FileMode = 0o666

	// Private directories and files (may contain sensitive data).
	PermsHomeProgDir os.FileMode = 0o700
	PermsPrivateFile os.FileMode = 0o600
)
