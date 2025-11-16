// Package sharedvalidation provides common validation functions for filesystem operations.
package sharedvalidation

import (
	"fmt"
	"os"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateDirectory validates that the directory exists, else creates it if desired.
func ValidateDirectory(dir string, createIfNotFound bool) (os.FileInfo, error) {
	// Check directory existence
	dirInfo, err := os.Stat(dir)
	switch {
	case err == nil: // If err IS nil
		if !dirInfo.IsDir() {
			return dirInfo, fmt.Errorf("path %q is a file, not a directory", dir)
		}
		return dirInfo, nil

	case os.IsNotExist(err):
		// path does not exist
		if createIfNotFound {
			if err := os.MkdirAll(dir, sharedconsts.PermsGenericDir); err != nil {
				return nil, fmt.Errorf("directory %q does not exist and failed to create: %w", dir, err)
			}
			if dirInfo, err = os.Stat(dir); err != nil { // re-stat to get correct FileInfo
				return dirInfo, fmt.Errorf("failed to stat %q", dir)
			}
			return dirInfo, nil
		}
		return nil, fmt.Errorf("directory %q does not exist", dir)

	default:
		// other error
		return nil, fmt.Errorf("failed to stat directory %q: %w", dir, err)
	}
}

// ValidateFile validates that the file exists, else creates it if desired.
func ValidateFile(f string, createIfNotFound bool) (os.FileInfo, error) {
	fileInfo, err := os.Stat(f)
	if err != nil {
		if os.IsNotExist(err) {
			if createIfNotFound {
				file, err := os.Create(f)
				if err != nil {
					return nil, fmt.Errorf("file %q does not exist and failed to create: %w", f, err)
				}
				file.Close()
				// Re-stat to get correct FileInfo
				if fileInfo, err = os.Stat(f); err != nil {
					return nil, fmt.Errorf("failed to stat %q after creation", f)
				}
				return fileInfo, nil
			}
			return nil, fmt.Errorf("file %q does not exist", f)
		}
		return nil, fmt.Errorf("failed to stat file %q: %w", f, err)
	}

	if fileInfo.IsDir() {
		return fileInfo, fmt.Errorf("path %q is a directory, not a file", f)
	}

	return fileInfo, nil
}
