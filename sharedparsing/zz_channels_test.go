package sharedparsing_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

// TestNormalizeChannelURL covers which differently typed channel URLs normalize to the same channel.
func TestNormalizeChannelURL(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"identical", "https://google.com", "https://google.com", true},
		{"trailing slash on one side", "https://google.com/", "https://google.com", true},
		{"trailing slash on the other", "https://google.com", "https://google.com/", true},
		{"trailing slash on both", "https://google.com/", "https://google.com/", true},
		{"case differs in host", "https://GOOGLE.com", "https://google.com", true},
		{"case differs in path", "https://youtube.com/@ChannelA", "https://youtube.com/@channela", true},
		{"surrounding whitespace", "  https://google.com/  ", "https://google.com", true},
		{"path with slash and case", "https://youtube.com/@ChannelA/", "https://youtube.com/@channela", true},

		// Everything else must still differ: this forgives typing, not identity.
		{"different channel on same host", "https://youtube.com/@ChannelA", "https://youtube.com/@ChannelB", false},
		{"different host", "https://google.com", "https://youtube.com", false},
		{"www is not stripped", "https://www.google.com", "https://google.com", false},
		{"scheme is not stripped", "http://google.com", "https://google.com", false},
		{"host only against a path", "https://youtube.com", "https://youtube.com/@ChannelA", false},
		{"interior slash is not trimmed", "https://youtube.com//@ChannelA", "https://youtube.com/@ChannelA", false},
		{"both empty", "", "", true},
		{"one empty", "", "https://google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedparsing.NormalizeChannelURL(tt.a) == sharedparsing.NormalizeChannelURL(tt.b); got != tt.want {
				t.Errorf("NormalizeChannelURL(%q) == NormalizeChannelURL(%q) is %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestStripURLScheme tests only a leading scheme is dropped, keeping subdomains like www.
func TestStripURLScheme(t *testing.T) {
	for in, want := range map[string]string{
		"https://www.youtube.com/@cats": "www.youtube.com/@cats",
		"http://example.com":            "example.com",
		"HTTPS://Example.com/a":         "Example.com/a",
		"youtube.com/@cats":             "youtube.com/@cats",
		"svn+ssh://host/repo":           "host/repo",
		"a.b-c2://host":                 "host",
		"a b://c":                       "a b://c", // Not a scheme.
		"a_b://c":                       "a_b://c",
		"://c":                          "://c",
	} {
		if got := sharedparsing.StripURLScheme(in); got != want {
			t.Errorf("StripURLScheme(%q) = %q, want %q", in, got, want)
		}
	}
}
