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
func ValidateDirectory(dir string, createIfNotFound bool, templateMap map[string]bool) (hasTemplating bool, fileInfo os.FileInfo, err error) {
	dir = filepath.Clean(dir)

	// Check template tags.
	if hasTemplating, errOrNil := checkTemplateTags(dir, templateMap); hasTemplating {
		return hasTemplating, nil, errOrNil // Do not stat templated directories, condition normal if tags are valid.
	}

	// Stat path.
	info, err := os.Stat(dir)
	if err == nil { // Err IS nil.
		if !info.IsDir() {
			return false, nil, fmt.Errorf("path %q is a file, not a directory", dir)
		}
		return false, info, nil
	}

	// Error other than non-existence.
	if !errors.Is(err, os.ErrNotExist) {
		return false, nil, fmt.Errorf("failed to stat directory %q: %w", dir, err)
	}

	// Does not exist, should not create.
	if !createIfNotFound {
		return false, nil, fmt.Errorf("directory %q does not exist", dir)
	}

	// Generate new directories.
	if err := os.MkdirAll(dir, sharedconsts.PermsGenericDir); err != nil {
		return false, nil, fmt.Errorf("directory %q does not exist and failed to create: %w", dir, err)
	}

	// Stat newly generated directory.
	info, err = os.Stat(dir)
	if err != nil {
		return false, nil, fmt.Errorf("failed to stat %q", dir)
	}
	return false, info, nil
}

// ValidateFile validates that the file exists, else creates it if desired.
func ValidateFile(path string, createIfNotFound bool, templateMap map[string]bool) (hasTemplating bool, fileInfo os.FileInfo, err error) {
	path = filepath.Clean(path)

	// Check template tags.
	if hasTemplating, errOrNil := checkTemplateTags(path, templateMap); hasTemplating {
		return hasTemplating, nil, errOrNil // Do not stat templated directories, condition normal if tags are valid.
	}

	// Stat path.
	info, err := os.Stat(path)
	if err == nil { // Err IS nil.
		if info.IsDir() {
			return false, nil, fmt.Errorf("path %q is a directory, not a file", path)
		}
		return false, info, nil
	}

	// Error other than non-existence.
	if !errors.Is(err, os.ErrNotExist) {
		return false, nil, fmt.Errorf("failed to stat file %q: %w", path, err)
	}

	// Does not exist, should not create.
	if !createIfNotFound {
		return false, nil, fmt.Errorf("file %q does not exist", path)
	}

	// Generate new file (must close after os.Create()).
	file, err := os.Create(path)
	if err != nil {
		return false, nil, fmt.Errorf("file %q does not exist and failed to create: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close file %q: %v", file.Name(), closeErr)
		}
	}()

	// Return info and nil/err.
	fileInfo, err = os.Stat(path)
	return false, fileInfo, err
}

// GetRenameFlag maps aliases from input if needed.
func GetRenameFlag(inFlag string) (outFlag string) {
	if inFlag == "" {
		return ""
	}

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
	case "skipped", "skipping", "skips", "none":
		return sharedconsts.RenameSkip
	}

	// No alias, send back input.
	return inFlag
}

// **** Private **********************************************************************************

// checkTemplateTags checks if the input string contains template elements.
func checkTemplateTags(s string, templateMap map[string]bool) (hasTemplating bool, err error) {
	if strings.Contains(s, "{{") && strings.Contains(s, "}}") {

		// Check all template tags for validity.
		allValid := checkAllTemplateTags(s, templateMap)
		if !allValid {
			tags := make([]string, 0, len(templateMap))
			for k := range templateMap {
				tags = append(tags, k)
			}
			return true, fmt.Errorf("directory contains unsupported template tags. Supported tags: %v", tags)
		} else {
			return true, nil
		}
	}
	return false, nil
}

// checkAllTemplateTags recursively checks template tags in string.
func checkAllTemplateTags(s string, templateMap map[string]bool) bool {
	for {
		// Start tag index.
		start := strings.Index(s, "{{")
		if start == -1 {
			return true
		}

		// End tag index.
		end := strings.Index(s[start:], "}}")
		if end == -1 {
			return false
		}

		// Check index safety.
		endAbs := start + end
		if start+end+2 > len(s) { // would slice past the string.
			return false
		}

		// Extract tag, compare against map.
		tag := s[start+2 : endAbs]
		if !templateMap[tag] {
			return false
		}

		s = s[endAbs+2:]
	}
}
