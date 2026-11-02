package sharedparsing

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// pathSeparatorReplacer replaces characters that would split a name into directories.
var pathSeparatorReplacer = strings.NewReplacer("/", "-", "\\", "-", "\x00", "")

// SanitizePathComponent makes s safe as a single file or directory name, cut to at most maxBytes. Control characters
// become spaces, as a name holding a line break splits paths read a line at a time.
func SanitizePathComponent(s string, maxBytes int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, pathSeparatorReplacer.Replace(s))
	s = strings.TrimSpace(s)
	if len(s) > maxBytes {
		cut := maxBytes
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = strings.TrimSpace(s[:cut])
	}

	// Checked last, as cutting can leave only dots.
	if s == "." || s == ".." {
		s = strings.ReplaceAll(s, ".", "-")
	}
	return s
}
