package sharedparsing_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

// TestSanitizePathComponent tests that names can't split into directories or exceed the byte limit.
func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		name, in, want string
		max            int
	}{
		{"slash", "AC/DC Live", "AC-DC Live", 255},
		{"backslash", `a\b`, "a-b", 255},
		{"parent", "..", "--", 255},
		{"traversal", "../../etc", "..-..-etc", 255},
		{"nul", "a\x00b", "ab", 255},
		{"line breaks", "line one\nline two\r\n", "line one line two", 255},
		{"other controls", "a\tb\x7fc\u0085d", "a b c d", 255},
		{"trimmed", "  title  ", "title", 255},
		{"cut", "abcdef", "abc", 3},
		{"cut keeps runes whole", "aé", "a", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedparsing.SanitizePathComponent(tt.in, tt.max); got != tt.want {
				t.Fatalf("SanitizePathComponent(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}

	long := sharedparsing.SanitizePathComponent(strings.Repeat("é", 200), sharedconsts.MaxNameBytes)
	if len(long) > sharedconsts.MaxNameBytes || !utf8.ValidString(long) {
		t.Fatalf("long name is %d bytes (valid UTF-8: %v), want at most %d valid bytes", len(long), utf8.ValidString(long), sharedconsts.MaxNameBytes)
	}
}
