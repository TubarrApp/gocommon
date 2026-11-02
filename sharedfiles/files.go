// Package sharedfiles holds operations for safer handling of files.
package sharedfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// OpenRegularInRoot opens a regular file relative to the given path's parent directory.
func OpenRegularInRoot(p string) (*os.File, error) {
	// Refuse a FIFO before opening, as opening one blocks until something writes to it.
	if info, err := os.Stat(p); err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", p)
	}

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
