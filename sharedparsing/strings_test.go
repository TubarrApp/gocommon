package sharedparsing_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

func TestEscapedSplit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"plain separator", `a:b`, []string{"a", "b"}},
		{"escaped separator", `a\:b`, []string{"a:b"}},
		{"escaped backslash", `a\\b`, []string{`a\b`}},
		{"escaped backslash then separator", `a\\:b`, []string{`a\`, "b"}},
		{"escaped pipe is decoded", `a\|b`, []string{`a|b`}},
		{"backslash before other char is kept", `a\db`, []string{`a\db`}},
		{"windows-style path", `C:\Dogs`, []string{"C", `\Dogs`}},
		{"trailing backslash", `a\`, []string{`a\`}},
		{"no separator", `abc`, []string{"abc"}},
		{"empty segments", `a::b`, []string{"a", "", "b"}},
		{"real op", `title:set:cats`, []string{"title", "set", "cats"}},
		{"value with colon", `title:set:cats\: the sequel`, []string{"title", "set", "cats: the sequel"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sharedparsing.EscapedSplit(tt.input, ':')
			if !slices.Equal(got, tt.want) {
				t.Errorf("EscapedSplit(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStagedSplitPreservesInnerEscapes is the property that decides the escaping rule:
// stripping a channel URL must not consume escapes meant for the decoding pass.
func TestStagedSplitPreservesInnerEscapes(t *testing.T) {
	const op = `https://example.com|title:set:a\:b`

	chanURL, rest, err := sharedparsing.SplitOpURL(op)
	if err != nil {
		t.Fatalf("SplitOpURL(%q) returned error: %v", op, err)
	}
	if chanURL != "https://example.com" {
		t.Errorf("chanURL = %q, want %q", chanURL, "https://example.com")
	}

	got := sharedparsing.EscapedSplit(rest, ':')
	if want := []string{"title", "set", "a:b"}; !slices.Equal(got, want) {
		t.Errorf("fields = %q, want %q (inner escape lost in the '|' pass)", got, want)
	}
}

// TestEscapeSplitRoundTrip is what makes the Tubarr -> Metarr handoff lossless: any
// segment must survive being escaped, joined and re-split.
func TestEscapeSplitRoundTrip(t *testing.T) {
	segments := [][]string{
		{"title", "set", "cats"},
		{"title", "set", "cats: the sequel"},
		{"title", "replace", `C:\Dogs`, "x"},
		{"title", "set", ""},
		{"description", "append", ` | piped | value `},
		{"title", "set", `a:b:c`},
		{"title", "replace", `\d`, "x"},
		{"title", "replace", `C:\Dogs\`, "x"},
		{"title", "set", `back\\slash`},
	}

	for _, want := range segments {
		t.Run(strings.Join(want, "_"), func(t *testing.T) {
			joined := sharedparsing.JoinEscaped(want, ':')
			got := sharedparsing.EscapedSplit(joined, ':')
			if !slices.Equal(got, want) {
				t.Errorf("round trip failed:\n segments %q\n joined   %q\n split    %q", want, joined, got)
			}
		})
	}
}

// TestBackslashHeavySegmentsRoundTrip covers the values that a separator-scoped escape
// rule could not represent: a backslash adjacent to a separator, or ending a segment.
func TestBackslashHeavySegmentsRoundTrip(t *testing.T) {
	cases := [][]string{
		{`a\`, "b"},
		{`00000\:`, "b"},
		{`a\`, `\b`},
		{`\`, `\\`},
		{`a\:b`, "c"},
		{`a\|b`, "c"},
	}

	for _, want := range cases {
		t.Run(strings.Join(want, "_"), func(t *testing.T) {
			joined := sharedparsing.JoinEscaped(want, ':')
			if got := sharedparsing.EscapedSplit(joined, ':'); !slices.Equal(got, want) {
				t.Errorf("round trip failed:\n segments %q\n joined   %q\n split    %q", want, joined, got)
			}
		})
	}
}

// TestNaiveJoinLosesSeparators covers why JoinEscaped exists: concatenating raw
// segments with the separator silently gains fields when re-split.
func TestNaiveJoinLosesSeparators(t *testing.T) {
	segments := []string{"title", "set", "cats: the sequel"}

	naive := strings.Join(segments, ":")
	if got := sharedparsing.EscapedSplit(naive, ':'); len(got) == len(segments) {
		t.Fatalf("expected naive join to corrupt the round trip, but got %q", got)
	}

	escaped := sharedparsing.JoinEscaped(segments, ':')
	if got := sharedparsing.EscapedSplit(escaped, ':'); !slices.Equal(got, segments) {
		t.Errorf("JoinEscaped round trip = %q, want %q", got, segments)
	}
}

func TestSplitOpURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantURL     string
		wantRest    string
		wantWarning bool
	}{
		{"no url", `title:set:cats`, "", `title:set:cats`, false},
		{"with url", `https://example.com|title:set:cats`, "https://example.com", `title:set:cats`, false},
		{"invalid url", `notaurl|title:set:cats`, "", `title:set:cats`, true},
		// Nothing is decoded here; every escape survives for the decoding split.
		{"escaped pipe is not a prefix", `title:set:a\|b`, "", `title:set:a\|b`, false},
		{"escaped pipe kept after url", `https://example.com|title:set:a\|b`, "https://example.com", `title:set:a\|b`, false},
		{"colon escape kept for decoding pass", `https://example.com|title:set:a\:b`, "https://example.com", `title:set:a\:b`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotRest, err := sharedparsing.SplitOpURL(tt.input)
			if gotURL != tt.wantURL {
				t.Errorf("chanURL = %q, want %q", gotURL, tt.wantURL)
			}
			if gotRest != tt.wantRest {
				t.Errorf("rest = %q, want %q", gotRest, tt.wantRest)
			}
			if (err != nil) != tt.wantWarning {
				t.Errorf("err = %v, want warning: %v", err, tt.wantWarning)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	in := []string{"a", "b", "a", "c", "b"}

	deduped, removed := sharedparsing.Deduplicate(in)
	if want := []string{"a", "b", "c"}; !slices.Equal(deduped, want) {
		t.Errorf("deduped = %q, want %q", deduped, want)
	}
	if want := []string{"a", "b"}; !slices.Equal(removed, want) {
		t.Errorf("removed = %q, want %q", removed, want)
	}
}
