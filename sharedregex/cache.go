// Package sharedregex compiles and caches various regex expressions.
package sharedregex

import (
	"regexp"
	"sync"
)

const (
	ansiEscapeStr   = `\x1b\[[0-9;]*m`
	extraSpacesStr  = `\s+`
	invalidCharsStr = `[<>:"/\\|?*\x00-\x1F]`
	specialCharsStr = `[^\w\s-]`
)

// Regex expressions, compiled once.
var (
	onceAnsiEscape   sync.Once
	onceExtraSpaces  sync.Once
	onceInvalidChars sync.Once
	onceSpecialChars sync.Once

	AnsiEscape   *regexp.Regexp
	ExtraSpaces  *regexp.Regexp
	InvalidChars *regexp.Regexp
	SpecialChars *regexp.Regexp
)

// AnsiEscapeCompile compiles regex for ANSI escape codes.
func AnsiEscapeCompile() *regexp.Regexp {
	onceAnsiEscape.Do(func() {
		AnsiEscape = regexp.MustCompile(ansiEscapeStr)
	})
	return AnsiEscape
}

// ExtraSpacesCompile compiles regex for extra spaces.
func ExtraSpacesCompile() *regexp.Regexp {
	onceExtraSpaces.Do(func() {
		ExtraSpaces = regexp.MustCompile(extraSpacesStr)
	})
	return ExtraSpaces
}

// InvalidCharsCompile compiles regex for invalid characters.
func InvalidCharsCompile() *regexp.Regexp {
	onceInvalidChars.Do(func() {
		InvalidChars = regexp.MustCompile(invalidCharsStr)
	})
	return InvalidChars
}

// SpecialCharsCompile compiles regex for special characters.
func SpecialCharsCompile() *regexp.Regexp {
	onceSpecialChars.Do(func() {
		SpecialChars = regexp.MustCompile(specialCharsStr)
	})
	return SpecialChars
}
