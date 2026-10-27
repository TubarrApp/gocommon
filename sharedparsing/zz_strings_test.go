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
		{"escaped ampersand is decoded", `a\&b`, []string{`a&b`}},
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
			got := sharedparsing.SplitAndRemoveEscapes(tt.input, ':')
			if !slices.Equal(got, tt.want) {
				t.Errorf("EscapedSplit(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStagedSplitPreservesInnerEscapes is the property that decides the escaping rule:
// dividing a filter from its operation must not consume escapes meant for the decoding pass.
func TestStagedSplitPreservesInnerEscapes(t *testing.T) {
	parts := sharedparsing.SplitAndKeepEscapes(`title:contains:a\:b||title:set:c\:d`, '|')
	if len(parts) != 3 || parts[1] != "" {
		t.Fatalf("SplitUnescaped = %q, want filter, empty, op", parts)
	}

	got := sharedparsing.SplitAndRemoveEscapes(parts[2], ':')
	if want := []string{"title", "set", "c:d"}; !slices.Equal(got, want) {
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
		{"description", "suffix", ` | piped | value `},
		{"title", "set", `a:b:c`},
		{"title", "replace", `\d`, "x"},
		{"title", "replace", `C:\Dogs\`, "x"},
		{"title", "set", `back\\slash`},
	}

	for _, want := range segments {
		t.Run(strings.Join(want, "_"), func(t *testing.T) {
			joined := sharedparsing.JoinEscaped(want, ':')
			got := sharedparsing.SplitAndRemoveEscapes(joined, ':')
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
			if got := sharedparsing.SplitAndRemoveEscapes(joined, ':'); !slices.Equal(got, want) {
				t.Errorf("round trip failed:\n segments %q\n joined   %q\n split    %q", want, joined, got)
			}
		})
	}
}

// TestEscapeAmpersands tests only a '&' starting "&&" is escaped, and that each decodes back.
func TestEscapeAmpersands(t *testing.T) {
	for in, want := range map[string]string{
		"Tom & Jerry": "Tom & Jerry",
		"R&&B":        `R\&&B`,
		"a&&&b":       `a\&\&&b`,
		"ends in &":   "ends in &",
	} {
		if got := sharedparsing.Escape(in); got != want {
			t.Errorf("Escape(%q) = %q, want %q", in, got, want)
		}
		if got := sharedparsing.SplitAndRemoveEscapes(want, ':'); len(got) != 1 || got[0] != in {
			t.Errorf("EscapedSplit(%q) = %q, want %q", want, got, in)
		}
	}
}

// TestNaiveJoinLosesSeparators covers why JoinEscaped exists: concatenating raw
// segments with the separator silently gains fields when re-split.
func TestNaiveJoinLosesSeparators(t *testing.T) {
	segments := []string{"title", "set", "cats: the sequel"}

	naive := strings.Join(segments, ":")
	if got := sharedparsing.SplitAndRemoveEscapes(naive, ':'); len(got) == len(segments) {
		t.Fatalf("expected naive join to corrupt the round trip, but got %q", got)
	}

	escaped := sharedparsing.JoinEscaped(segments, ':')
	if got := sharedparsing.SplitAndRemoveEscapes(escaped, ':'); !slices.Equal(got, segments) {
		t.Errorf("JoinEscaped round trip = %q, want %q", got, segments)
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
