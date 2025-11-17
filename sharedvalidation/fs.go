// Package sharedvalidation provides common validation functions for filesystem operations.
package sharedvalidation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateDirectory validates that the directory exists, else creates it if desired.
func ValidateDirectory(dir string, createIfNotFound bool) (os.FileInfo, error) {
	// Check directory existence.
	dir = filepath.Clean(dir)

	// Stat path.
	info, err := os.Stat(dir)
	if err == nil { // Err IS nil.
		if !info.IsDir() {
			return nil, fmt.Errorf("path %q is a file, not a directory", dir)
		}
		return info, nil
	}

	// Error other than non-existence.
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to stat directory %q: %w", dir, err)
	}

	// Does not exist, should not create.
	if !createIfNotFound {
		return nil, fmt.Errorf("directory %q does not exist", dir)
	}

	// Generate new directories.
	if err := os.MkdirAll(dir, sharedconsts.PermsGenericDir); err != nil {
		return nil, fmt.Errorf("directory %q does not exist and failed to create: %w", dir, err)
	}

	// Stat newly generated directory.
	info, err = os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %q", dir)
	}
	return info, nil
}

// ValidateFile validates that the file exists, else creates it if desired.
func ValidateFile(path string, createIfNotFound bool) (os.FileInfo, error) {
	path = filepath.Clean(path)

	// Stat path.
	info, err := os.Stat(path)
	if err == nil { // Err IS nil.
		if info.IsDir() {
			return nil, fmt.Errorf("path %q is a directory, not a file", path)
		}
		return info, nil
	}

	// Error other than non-existence.
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to stat file %q: %w", path, err)
	}

	// Does not exist, should not create.
	if !createIfNotFound {
		return nil, fmt.Errorf("file %q does not exist", path)
	}

	// Generate new file (must close after os.Create()).
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("file %q does not exist and failed to create: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close file %q: %v", file.Name(), closeErr)
		}
	}()

	// Return info and nil/err.
	return os.Stat(path)
}

// GetRenameFlag maps aliases from input if needed.
func GetRenameFlag(inFlag string) (outFlag string) {
	// Normalize string.
	f := inFlag
	f = strings.ReplaceAll(f, " ", "")
	f = strings.ToLower(f)

	// Check map.
	if sharedconsts.ValidRenameFlags[f] {
		return f
	}

	// Find aliases.
	switch f {
	case "space", "spaced":
		return sharedconsts.RenameSpaces
	case "underscore", "underscored":
		return sharedconsts.RenameUnderscores
	case "fix", "fixed", "fixes", "fixesonly":
		return sharedconsts.RenameFixesOnly
	case "skipped", "none":
		return sharedconsts.RenameSkip
	}

	// No alias, send back input.
	return inFlag
}
