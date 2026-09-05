package sharedparsing_test

import (
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

// FuzzMetaOpRoundTrip asserts that formatting an operation and parsing it back is
// lossless for arbitrary field values, which is what the Tubarr -> Metarr argv handoff
// relies on.
func FuzzMetaOpRoundTrip(f *testing.F) {
	f.Add("title", "set", "cats")
	f.Add("title", "set", "a|b:c")
	f.Add("description", "append", ` | piped : value `)
	f.Add("title", "set", `C:\Dogs`)
	f.Add("", "set", "")
	f.Add("title", "set", `back\slash\`)

	f.Fuzz(func(t *testing.T, field, opType, opValue string) {
		// Restrict to the three-part form so the shape is well defined.
		if opType == "" || isMultiPartOpType(opType) {
			t.Skip()
		}
		// A field or operation type holding a separator would change the parsed shape.
		if strings.ContainsAny(field+opType, ":|") {
			t.Skip()
		}

		want := sharedmodels.MetaOps{Field: field, OpType: opType, OpValue: opValue}

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
	case "date-tag", "delete-date-tag", "replace", "replace-suffix", "replace-prefix":
		return true
	default:
		return false
	}
}
