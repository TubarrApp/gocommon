package shareddates

import (
	"io"
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

// TestMaxDaysInMonth tests regular months, leap years, and non-leap years.
func TestMaxDaysInMonth(t *testing.T) {
	tests := []struct {
		name        string
		month       int
		yearOrZero  int
		wantMaxDays int
	}{
		{
			name:        "January (31 days)",
			month:       1,
			yearOrZero:  2026,
			wantMaxDays: 31,
		},
		{
			name:        "February in leap year",
			month:       2,
			yearOrZero:  2024,
			wantMaxDays: 29,
		},
		{
			name:        "February in non-leap year",
			month:       2,
			yearOrZero:  2026,
			wantMaxDays: 28,
		},
		{
			name:        "February with zero year (defaults to 29)",
			month:       2,
			yearOrZero:  0,
			wantMaxDays: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxDaysInMonth(tt.month, tt.yearOrZero)
			if got != tt.wantMaxDays {
				t.Errorf("maxDaysInMonth() = %v, want %v", got, tt.wantMaxDays)
			}
		})
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
