package sharedenums_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedenums"
)

func TestParseDateFormat(t *testing.T) {
	valid := map[string]sharedenums.DateFormat{
		"Ymd": sharedenums.DateYyyyMmDd,
		"ymd": sharedenums.DateYyMmDd,
		"Ydm": sharedenums.DateYyyyDdMm,
		"ydm": sharedenums.DateYyDdMm,
		"dmY": sharedenums.DateDdMmYyyy,
		"dmy": sharedenums.DateDdMmYy,
		"mdY": sharedenums.DateMmDdYyyy,
		"mdy": sharedenums.DateMmDdYy,
		"md":  sharedenums.DateMmDd,
		"dm":  sharedenums.DateDdMm,
	}

	for in, want := range valid {
		got, err := sharedenums.ParseDateFormat(in)
		if err != nil {
			t.Errorf("ParseDateFormat(%q) returned error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseDateFormat(%q) = %d, want %d", in, got, want)
		}
	}

	for _, in := range []string{"", "y", "d", "ymdY", "abc", "YMD", "xyz"} {
		if _, err := sharedenums.ParseDateFormat(in); err == nil {
			t.Errorf("ParseDateFormat(%q) should have returned an error", in)
		}
	}
}

// TestTwoCharDateFormatsAccepted guards the divergence that "md" and "dm" were
// unreachable in Tubarr's validator (its length guard rejected them) while Metarr
// accepted them, so a valid operation could not be saved.
func TestTwoCharDateFormatsAccepted(t *testing.T) {
	for _, in := range []string{"md", "dm"} {
		if _, err := sharedenums.ParseDateFormat(in); err != nil {
			t.Errorf("two-character date format %q must be accepted, got: %v", in, err)
		}
	}
}

func TestDateFormatStringRoundTrip(t *testing.T) {
	for _, in := range sharedenums.ValidDateFormats {
		e, err := sharedenums.ParseDateFormat(in)
		if err != nil {
			t.Fatalf("ParseDateFormat(%q) returned error: %v", in, err)
		}
		if got := e.String(); got != in {
			t.Errorf("round trip: %q -> %d -> %q", in, e, got)
		}
	}

	if got := sharedenums.DateFmtSkip.String(); got != "" {
		t.Errorf("DateFmtSkip.String() = %q, want empty", got)
	}
}

func TestParseDateTagLocation(t *testing.T) {
	tests := []struct {
		loc      string
		allowAll bool
		want     sharedenums.DateTagLocation
		wantErr  bool
	}{
		{sharedconsts.OpLocPrefix, false, sharedenums.DateTagLocPrefix, false},
		{sharedconsts.OpLocSuffix, false, sharedenums.DateTagLocSuffix, false},
		{sharedconsts.OpLocAll, false, sharedenums.DateTagLocPrefix, true},
		{sharedconsts.OpLocPrefix, true, sharedenums.DateTagLocPrefix, false},
		{sharedconsts.OpLocSuffix, true, sharedenums.DateTagLocSuffix, false},
		{sharedconsts.OpLocAll, true, sharedenums.DateTagLocAll, false},
		{"middle", true, sharedenums.DateTagLocPrefix, true},
		{"", false, sharedenums.DateTagLocPrefix, true},
	}

	for _, tt := range tests {
		got, err := sharedenums.ParseDateTagLocation(tt.loc, tt.allowAll)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDateTagLocation(%q, %v) err = %v, wantErr %v", tt.loc, tt.allowAll, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseDateTagLocation(%q, %v) = %d, want %d", tt.loc, tt.allowAll, got, tt.want)
		}
	}
}
