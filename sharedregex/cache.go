// Package sharedregex compiles and caches various regex expressions.
package sharedregex

import (
	"regexp"
	"sync"
)

// Regex strings for compilation.
const (
	ansiEscapeStr = `\x1b\[[0-9;]*m`
)

// Regex expressions, compiled once.
var (
	onceAnsiEscape sync.Once
)

// Compiled regex cache.
var (
	AnsiEscape *regexp.Regexp
)

// AnsiEscapeCompile compiles regex for ANSI escape codes.
func AnsiEscapeCompile() *regexp.Regexp {
	onceAnsiEscape.Do(func() {
		AnsiEscape = regexp.MustCompile(ansiEscapeStr)
	})
	return AnsiEscape
}
