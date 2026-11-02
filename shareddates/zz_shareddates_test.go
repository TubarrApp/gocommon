package shareddates

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/logging"
	"github.com/TubarrApp/gocommon/sharedenums"
)

var testPl *logging.ProgramLogger

// TestMain sets up the test environment for the app package.
func TestMain(m *testing.M) {
	testPl = &logging.ProgramLogger{
		Console: io.Discard,
	}
	m.Run()
}

// TestParseWordDate verifies parsing of natural language and standard word date strings.
func TestParseWordDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard ordinal word date",
			input:   "Jan 2nd, 2006",
			want:    "2006-01-02",
			wantErr: false,
		},
		{
			name:    "Invalid date string",
			input:   "not-a-date",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWordDate(tt.input) // Called directly since we are in the same package
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWordDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWordDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUnpaddedDates tests single digit months and days are padded, not shifted when the hyphens are removed.
func TestUnpaddedDates(t *testing.T) {
	for in, want := range map[string]string{
		"2025-1-2":            "2025-01-02",
		"2025-10-5":           "2025-10-05",
		"2025-1-15":           "2025-01-15",
		"2025-1-2T10:00:00Z":  "2025-01-02",
		"2025-01-02T10:00:00": "2025-01-02",
	} {
		y, m, d, err := ParseDateComponents(in, sharedenums.DateYyyyMmDd, testPl)
		if err != nil {
			t.Errorf("ParseDateComponents(%q): %v", in, err)
			continue
		}
		if got := FormatDateString(y, m, d, sharedenums.DateYyyyMmDd); got != want {
			t.Errorf("ParseDateComponents(%q) formats to %q, want %q", in, got, want)
		}
	}

	if got, made := YmdFromMeta("2025-1-2T10:00:00Z", testPl); !made || got != "2025-01-02T10:00:00Z" {
		t.Errorf("YmdFromMeta = %q, %v, want it padded", got, made)
	}
}

// TestParseNumDate tests days are written without a leading zero.
func TestParseNumDate(t *testing.T) {
	for in, want := range map[string]string{
		"2025-01-01": "January 1st, 2025",
		"2025-03-02": "March 2nd, 2025",
		"2025-04-13": "April 13th, 2025",
		"2025-11-21": "November 21st, 2025",
	} {
		if got, err := ParseNumDate(in, testPl); err != nil || got != want {
			t.Errorf("ParseNumDate(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

// TestDetectedFormatLog tests the detected date format is logged with its date filled in.
func TestDetectedFormatLog(t *testing.T) {
	var buf bytes.Buffer
	pl := &logging.ProgramLogger{Console: &buf}
	logging.SetLevel(1)
	defer logging.SetLevel(-1)

	for _, date := range []string{"20260110", "20262810"} {
		if _, _, _, err := ParseDateComponents(date, sharedenums.DateYyyyMmDd, pl); err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(buf.String(), "%s") {
		t.Errorf("log holds an unfilled placeholder: %q", buf.String())
	}
}

// TestFormatDateString verifies proper hyphen-separated assembly according to DateFormat enums.
func TestFormatDateString(t *testing.T) {
	tests := []struct {
		name    string
		year    string
		month   string
		day     string
		dateFmt sharedenums.DateFormat
		want    string
	}{
		{
			name:    "YYYY-MM-DD",
			year:    "2026",
			month:   "10",
			day:     "28",
			dateFmt: sharedenums.DateYyyyMmDd,
			want:    "2026-10-28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDateString(tt.year, tt.month, tt.day, tt.dateFmt)
			if got != tt.want {
				t.Errorf("FormatDateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDateComponents_Guesses verifies how the parser handles ambiguous inputs,
// historical years, and mmdd vs ddmm arrangements.
func TestParseDateComponents_Guesses(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		dateFmt   sharedenums.DateFormat
		wantYear  string
		wantMonth string
		wantDay   string
		wantErr   bool
	}{
		{
			name:      "Historical year guess (1889)",
			input:     "1889",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "1889",
			wantMonth: "",
			wantDay:   "",
			wantErr:   false,
		},
		{
			name:      "Recent year guess (2025)",
			input:     "2025",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "2025",
			wantMonth: "",
			wantDay:   "",
			wantErr:   false,
		},
		{
			name:      "Year guess in a two digit year format",
			input:     "2025",
			dateFmt:   sharedenums.DateYyMmDd,
			wantYear:  "25",
			wantMonth: "",
			wantDay:   "",
			wantErr:   false,
		},
		{
			name:      "Ambiguous 0101 (defaults to month-day US convention)",
			input:     "0101",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "",
			wantMonth: "01",
			wantDay:   "01",
			wantErr:   false,
		},
		{
			name:      "Ambiguous 1210 (December 10th)",
			input:     "1210",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "",
			wantMonth: "12",
			wantDay:   "10",
			wantErr:   false,
		},
		{
			name:      "Full 8-digit YYYYMMDD string",
			input:     "20261028",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "2026",
			wantMonth: "10",
			wantDay:   "28",
			wantErr:   false,
		},
		{
			name:      "Full 8-digit YYYYMMDD string",
			input:     "20262810",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "2026",
			wantMonth: "10",
			wantDay:   "28",
			wantErr:   false,
		},
		{
			name:      "Completely invalid date string",
			input:     "9999",
			dateFmt:   sharedenums.DateYyyyMmDd,
			wantYear:  "",
			wantMonth: "",
			wantDay:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, day, err := ParseDateComponents(tt.input, tt.dateFmt, testPl)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateComponents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if year != tt.wantYear || month != tt.wantMonth || day != tt.wantDay {
				t.Errorf("ParseDateComponents() got = (%q, %q, %q), want = (%q, %q, %q)",
					year, month, day, tt.wantYear, tt.wantMonth, tt.wantDay)
			}
		})
	}
}
