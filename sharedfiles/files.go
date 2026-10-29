// Package sharedfiles holds operations for safer handling of files.
package sharedfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// OpenRegularInRoot opens a regular file relative to the given path's parent directory.
func OpenRegularInRoot(p string) (*os.File, error) {
	rootDir, fileBase := getRootDirAndFile(p)

	// Open in root directory (disallow '..' type traversal).
	f, err := os.OpenInRoot(rootDir, fileBase)
	if err != nil {
		return nil, err
	}

	// Stat the file, ensure the file has no mode type bits.
	if err := statAndEnsureRegular(f); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return nil, errors.Join(
				err,
				fmt.Errorf("failed to close file at path %q: %w", p, closeErr),
			)
		}
		return nil, err
	}
	return f, nil
}

// EnsureMetadataExt checks that the file is a valid metadata extension.
func EnsureMetadataExt(s string) (isValid bool) {
	_, isValid = sharedconsts.AllMetaExtensions[filepath.Ext(s)]
	return isValid
}

// EnsureVideoExt checks that the file is a valid video extension.
func EnsureVideoExt(s string) (isValid bool) {
	_, isValid = sharedconsts.AllVidExtensions[filepath.Ext(s)]
	return isValid
}

const (
	maxImageBytes  = 20 << 20 // 20 MiB
	maxImagePixels = 40_000_000
)

// getRootDirAndFile returns the root directory and file from a given string.
func getRootDirAndFile(p string) (rootDir, fileBase string) {
	// '/etc/myapp/../file.txt' -> '/etc/file.txt'
	p = filepath.Clean(p)

	// '/etc/file.txt' -> '/etc'
	rootDir = filepath.Dir(p)
	// '/etc/file.txt' -> 'file.txt'
	fileBase = filepath.Base(p)
	return
}

// statAndEnsureRegular stats the file and checks if it is a regular file type.
func statAndEnsureRegular(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", f.Name())
	}
	return nil
}
