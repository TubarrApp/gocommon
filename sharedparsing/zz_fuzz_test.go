package sharedparsing_test

import (
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// FuzzMetaOpRoundTrip asserts that formatting an operation and parsing it back is
// lossless for arbitrary field values, which is what the Tubarr -> Metarr argv handoff
// relies on.
func FuzzMetaOpRoundTrip(f *testing.F) {
	f.Add("title", "set", "cats")
	f.Add("title", "set", "a|b:c")
	f.Add("description", "append", ` | piped : value `)
	f.Add("title", "set", `C:\Dogs`)
	f.Add("title", "set", "")
	f.Add("title", "set", `back\slash\`)
	f.Add("title", "set", "HELLO: THIS IS MY TEXT")
	f.Add("title", "keep-after", "| ")
	f.Add("title", "set", "https://a.com|b")

	f.Fuzz(func(t *testing.T, field, opType, opValue string) {
		// Restrict to the three-part form so the shape is well defined.
		if opType == "" || isMultiPartOpType(opType) {
			t.Skip()
		}
		// A field or operation type holding a separator would change the parsed shape.
		if strings.ContainsAny(field+opType, ":|") {
			t.Skip()
		}
		// Parsing validates, so only valid operations can round trip.
		want := sharedmodels.MetaOps{Field: field, OpType: opType, OpValue: opValue}
		if sharedvalidation.ValidateMetaOps([]sharedmodels.MetaOps{want}) != nil {
			t.Skip()
		}

		formatted := sharedparsing.FormatMetaOp(want, false)
		got, _, err := sharedparsing.ParseMetaOps([]string{formatted})
		if err != nil {
			t.Fatalf("re-parse of %q (from %+v) failed: %v", formatted, want, err)
		}
		if got[0] != want {
			t.Errorf("round trip changed the operation:\n formatted %q\n want %+v\n got  %+v", formatted, want, got[0])
		}
	})
}

// isMultiPartOpType reports whether opType uses the four-part operation form.
func isMultiPartOpType(opType string) bool {
	switch opType {
	case "date-tag", "delete-date-tag", "replace", "replace-suffix", "replace-prefix", "regex-replace":
		return true
	default:
		return false
	}
}

// FuzzRegexReplaceRoundTrip asserts that a regex-replace pattern survives formatting and
// re-parsing verbatim, including inside a filtered entry that is split on '|' first.
func FuzzRegexReplaceRoundTrip(f *testing.F) {
	f.Add(`^Kino Casino [^|]*\| `, "")
	f.Add(`(\d+):(\d+)`, "$2:$1")
	f.Add(`[[:digit:]]`, `a\|b`)
	f.Add(`\d{4}}`, "{{uploader}}")

	f.Fuzz(func(t *testing.T, pattern, replacement string) {
		// Parsing validates, so only valid patterns can round trip. This also skips "]]:",
		// which would close the span early.
		want := sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: pattern, OpValue: replacement}
		if sharedvalidation.ValidateMetaOps([]sharedmodels.MetaOps{want}) != nil {
			t.Skip()
		}
		formatted := sharedparsing.FormatMetaOp(want, false)

		got, _, err := sharedparsing.ParseMetaOps([]string{formatted})
		if err != nil {
			t.Fatalf("re-parse of %q failed: %v", formatted, err)
		}
		if got[0] != want {
			t.Errorf("round trip changed the operation:\n formatted %q\n want %+v\n got  %+v", formatted, want, got[0])
		}

		filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{"https://example.com|title:contains:x|" + formatted})
		if err != nil {
			t.Fatalf("filtered re-parse of %q failed: %v", formatted, err)
		}
		want.ChannelURL = "https://example.com"
		if len(filtered) != 1 || len(filtered[0].MetaOps) != 1 || filtered[0].MetaOps[0] != want {
			t.Errorf("filtered round trip changed the operation:\n formatted %q\n got %+v", formatted, filtered)
		}
	})
}
