package sharedparsing_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

// TestParseDuration tests the duration formats scraped and metadata durations accept, and that others are errors.
func TestParseDuration(t *testing.T) {
	for in, want := range map[string]int{
		// Integer and fractional seconds.
		"3723":   3723,
		"0":      0,
		"312.5":  312,
		"0.9":    0,
		"1.9":    1,
		"-1":     0,
		"-1.5":   0,
		" 3723 ": 3723,
		"+3723":  3723,
		"1e3":    1000,

		// Colon-separated durations.
		"00:47:35":      2855,
		"1:02:03":       3723,
		"62:03":         3723,
		"0:00":          0,
		"00:00:00":      0,
		"1:60":          120,
		"1:99":          159,
		"1:60:00":       7200,
		" 1 : 02 : 03 ": 3723,

		// ISO 8601 hours, minutes, and seconds.
		"PT1H2M3S":    3723,
		"pt47m":       2820,
		"PT07H24M44S": 26684,
		"PT1H":        3600,
		"PT1M":        60,
		"PT1S":        1,
		"PT90S":       90,
		"PT90M":       5400,
		"PT1H30M":     5400,
		"PT0S":        0,
		"PT0.5S":      0,
		"PT1.9S":      1,
		"PT1,9S":      1,

		// ISO 8601 days and time.
		"P2D":        172800,
		"P1DT2H3M4S": 93784,
		"P1DT1H":     90000,
		"P1DT30M":    88200,
		"P1DT0S":     86400,

		// ISO 8601 weeks.
		"P1W": 604800,
		"P2W": 1209600,
		"P0W": 0,
	} {
		if got, err := sharedparsing.ParseDuration(in); err != nil || got != want {
			t.Errorf("ParseDuration(%q) = %d, %v, want %d", in, got, err, want)
		}
	}

	for _, in := range []string{
		// Empty and unrecognised input.
		"",
		"abc",
		" ",
		"1:xx",
		"1:2:3:4",
		"1:",
		":1",
		"1::2",

		// Unsupported ISO 8601 years and months.
		"P",
		"PT",
		"P1Y",
		"P1M",
		"P1Y2D",
		"P1DT1Y",
		"P1MT1H",

		// Invalid ISO 8601 day and week syntax.
		"P-1D",
		"P1.5D",
		"PD",
		"P1DD",
		"P1DT2D",
		"P1W2D",
		"P1WT1H",
		"P1WW",

		// Invalid ISO 8601 time syntax.
		"PT-1H",
		"PT1H-2M",
		"PT1H2H",
		"PT1M2H",
		"PT1S2M",
		"PT1H2",
		"PT1H2M3",
		"PT1.5H30M",
		"PT1..5S",
		"PT.S",
		"PT1X",
		"PT1H2M3S4S",
		"PT1H2T3S",

		// Invalid numeric values.
		"NaN",
		"+Inf",
		"-Inf",
	} {
		if got, err := sharedparsing.ParseDuration(in); err == nil {
			t.Errorf("ParseDuration(%q) = %d, want an error", in, got)
		}
	}
}
