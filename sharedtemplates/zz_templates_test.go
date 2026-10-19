package sharedtemplates

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestStripURLScheme tests only a leading scheme is dropped, keeping subdomains like www.
func TestStripURLScheme(t *testing.T) {
	for in, want := range map[string]string{
		"https://www.youtube.com/@cats": "www.youtube.com/@cats",
		"http://example.com":            "example.com",
		"HTTPS://Example.com/a":         "Example.com/a",
		"youtube.com/@cats":             "youtube.com/@cats",
		"a b://c":                       "a b://c", // Not a scheme.
		"://c":                          "://c",
	} {
		if got := StripURLScheme(in); got != want {
			t.Errorf("StripURLScheme(%q) = %q, want %q", in, got, want)
		}
	}
}

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
			if got := SanitizePathComponent(tt.in, tt.max); got != tt.want {
				t.Fatalf("SanitizePathComponent(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}

	long := SanitizePathComponent(strings.Repeat("é", 200), MaxNameBytes)
	if len(long) > MaxNameBytes || !utf8.ValidString(long) {
		t.Fatalf("long name is %d bytes (valid UTF-8: %v), want at most %d valid bytes", len(long), utf8.ValidString(long), MaxNameBytes)
	}
}

// TestValidTag tests tag names and directives against Metarr's tags.
func TestValidTag(t *testing.T) {
	for tag, want := range map[string]bool{
		"year":              true,
		"year:yy":           true,
		" Year : YYYY ":     true,
		"year:yyy":          false,
		"year:":             true, // Empty directive means the default.
		"month":             true,
		"month:number":      true,
		"month:abbreviated": true,
		"month:WORD":        true,
		"month:short":       false,
		"day":               true,
		"day:number":        false, // Day takes no directives.
		"video_title:yy":    false,
		"channel_name":      false, // Not a Metarr tag.
		"nope":              false,
	} {
		if got := ValidTag(tag, MetarrTemplateTags); got != want {
			t.Errorf("ValidTag(%q) = %v, want %v", tag, got, want)
		}
	}
}

// TestDescribeTags tests the supported tags list shown in errors.
func TestDescribeTags(t *testing.T) {
	got := DescribeTags(MetarrTemplateTags)
	for _, want := range []string{"day", "month[:abbreviated|number|word]", "year[:yy|yyyy]"} {
		if !slices.Contains(got, want) {
			t.Errorf("DescribeTags() = %v, missing %q", got, want)
		}
	}
	if !slices.IsSorted(got) {
		t.Errorf("DescribeTags() = %v, not sorted", got)
	}
}
