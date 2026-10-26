package sharedparsing_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

// TestParseDuration tests the duration formats scraped and metadata durations accept, and that others are errors.
func TestParseDuration(t *testing.T) {
	for in, want := range map[string]int{"00:47:35": 2855, "1:02:03": 3723, "62:03": 3723, "3723": 3723, "312.5": 312, "0": 0, "PT1H2M3S": 3723, "pt47m": 2820, "PT07H24M44S": 26684, "P1DT2H3M4S": 93784, "P2D": 172800} {
		if got, err := sharedparsing.ParseDuration(in); err != nil || got != want {
			t.Errorf("ParseDuration(%q) = %d, %v, want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"P1Y", "PT-1H", "P", "PT", "abc", "1:xx", "1:2:3:4", ""} {
		if got, err := sharedparsing.ParseDuration(in); err == nil {
			t.Errorf("ParseDuration(%q) = %d, want an error", in, got)
		}
	}
}
