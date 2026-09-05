package sharedparsing_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

// TestChannelURLsEqual covers the ChannelURLsEqual function, which is used to determine if two channel URLs are equivalent for the purposes of scoping operations.
func TestChannelURLsEqual(t *testing.T) {
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
			if got := sharedparsing.ChannelURLsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("ChannelURLsEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestAppliesToChannel covers an unscoped operation applying everywhere.
func TestAppliesToChannel(t *testing.T) {
	if !sharedparsing.AppliesToChannel("", "https://google.com") {
		t.Error("an operation with no channel URL should apply to any channel")
	}
	if !sharedparsing.AppliesToChannel("https://google.com/", "https://GOOGLE.com") {
		t.Error("a scoped operation should apply to the same channel typed differently")
	}
	if sharedparsing.AppliesToChannel("https://google.com", "https://youtube.com") {
		t.Error("a scoped operation should not apply to another channel")
	}
}
