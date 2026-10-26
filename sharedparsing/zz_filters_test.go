package sharedparsing_test

import (
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

func TestParseFilterOps(t *testing.T) {
	withMustAny, err := sharedparsing.ParseFilterOps([]string{
		"title:omits:frogs:must",
		"title:contains:cat:any",
		"date:omits:must",
	}, true)
	if err != nil {
		t.Fatalf("ParseFilterOps with must/any: %v", err)
	}
	if len(withMustAny) != 3 {
		t.Fatalf("parsed %d filters, want 3", len(withMustAny))
	}
	if withMustAny[0].MustAny != "must" || withMustAny[1].MustAny != "any" {
		t.Errorf("must/any not read: %+v", withMustAny)
	}
	if withMustAny[2].Value != "" {
		t.Errorf("valueless form should have no value: %+v", withMustAny[2])
	}

	plain, err := sharedparsing.ParseFilterOps([]string{"title:contains:cat", "date:omits"}, false)
	if err != nil {
		t.Fatalf("ParseFilterOps plain: %v", err)
	}
	// Must is implied when each entry holds a single filter.
	for _, f := range plain {
		if f.MustAny != "must" {
			t.Errorf("expected implied must, got %+v", f)
		}
	}

	// A valueless form only makes sense for contains and omits.
	if _, err := sharedparsing.ParseFilterOps([]string{"duration:more-than"}, false); err == nil {
		t.Error("valueless more-than should be rejected")
	}
	if _, err := sharedparsing.ParseFilterOps([]string{"title:contains:cat:maybe"}, true); err == nil {
		t.Error("invalid must/any should be rejected")
	}
}

// TestParseFilterOpsValueIsRemainder tests a filter's value is everything after its type, before a must/any condition,
// so titles holding ':' or '|' need no escaping.
func TestParseFilterOpsValueIsRemainder(t *testing.T) {
	tests := []struct {
		in          string
		withMustAny bool
		wantValue   string
		wantMustAny string
	}{
		{"title:contains:Part 2: The Return", false, "part 2: the return", "must"},
		{"title:contains:Ep 5 | Show: Finale", false, "ep 5 | show: finale", "must"},
		{"title:contains:Re: Zero:any", true, "re: zero", "any"},
		{"title:contains:a:must:must", true, "a:must", "must"},
		{"title:equals:12:00:must", true, "12:00", "must"},
		{`title:contains:C:\Videos\`, false, `c:\videos\`, "must"},
	}
	for _, tt := range tests {
		got, err := sharedparsing.ParseFilterOps([]string{tt.in}, tt.withMustAny)
		if err != nil || len(got) != 1 {
			t.Errorf("ParseFilterOps(%q) = %+v, %v", tt.in, got, err)
			continue
		}
		if got[0].Field != "title" || got[0].Value != tt.wantValue || got[0].MustAny != tt.wantMustAny {
			t.Errorf("ParseFilterOps(%q) = %+v, want value %q and %s", tt.in, got[0], tt.wantValue, tt.wantMustAny)
		}
	}
	if _, err := sharedparsing.ParseFilterOps([]string{"title:contains:cat"}, true); err == nil {
		t.Error("a download filter with no must/any should be rejected")
	}
	for _, bad := range []string{"title:contains:cat:must", "title:omits:Re: Zero:ANY"} {
		if _, err := sharedparsing.ParseFilterOps([]string{bad}, false); err == nil {
			t.Errorf("%q has a must/any where none is used, so should be rejected", bad)
		}
	}
	for _, in := range []string{`title:contains:cat\:must`, "title:contains:any"} {
		if _, err := sharedparsing.ParseFilterOps([]string{in}, false); err != nil {
			t.Errorf("%q: %v", in, err)
		}
	}
}

// TestParseFilterOpsRegexContains tests a regex-contains pattern is read verbatim from inside "[[ ]]", keeping its case
// and escapes, and is written back the same way.
func TestParseFilterOpsRegexContains(t *testing.T) {
	tests := []struct {
		in          string
		withMustAny bool
		wantValue   string
	}{
		{`title:regex-contains:[[^Cat \d+: Part]]:must`, true, `^Cat \d+: Part`},
		{`title:regex-contains:[[cats|dogs]]`, false, `cats|dogs`},
		{`title:regex-contains:[[[A-Z]]]:any`, true, `[A-Z]`},
		{`title:regex-contains:[[(?i)ep\.? ?\d]]`, false, `(?i)ep\.? ?\d`},
	}
	for _, tt := range tests {
		got, err := sharedparsing.ParseFilterOps([]string{tt.in}, tt.withMustAny)
		if err != nil || len(got) != 1 || got[0].FilterType != "regex-contains" || got[0].Value != tt.wantValue {
			t.Errorf("ParseFilterOps(%q) = %+v, %v, want value %q", tt.in, got, err, tt.wantValue)
			continue
		}
		if formatted := sharedparsing.FormatFilter(got[0], tt.withMustAny); formatted != tt.in {
			t.Errorf("FormatFilter(%+v) = %q, want %q", got[0], formatted, tt.in)
		}
	}

	for _, bad := range []string{`title:regex-contains:cats`, `title:regex-contains:[[a]]:b]]`, `title:regex-contains:[[cats]]:any`} {
		if _, err := sharedparsing.ParseFilterOps([]string{bad}, false); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}

	filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{`title:regex-contains:[[^Cat|^Dog]]||genre:set:Pets`})
	if err != nil || len(filtered) != 1 || filtered[0].Filters[0].Value != "^Cat|^Dog" {
		t.Errorf("filtered regex-contains = %+v, %v", filtered, err)
	}
}

// TestFormatFilterEmptyEquals tests equals and not-equals with no value are written so they parse back, rather than as
// the presence-only form.
func TestFormatFilterEmptyEquals(t *testing.T) {
	for _, f := range []sharedmodels.Filters{
		{Field: "description", FilterType: "equals", MustAny: "must"},
		{Field: "description", FilterType: "not-equals", MustAny: "any"},
	} {
		for _, withMustAny := range []bool{true, false} {
			if !withMustAny && f.MustAny != "must" {
				continue
			}
			formatted := sharedparsing.FormatFilter(f, withMustAny)
			got, err := sharedparsing.ParseFilterOps([]string{formatted}, withMustAny)
			if err != nil || len(got) != 1 || got[0] != f {
				t.Errorf("round trip of %+v: formatted %q, got %+v, %v", f, formatted, got, err)
			}
		}
	}
}

// FuzzFilterRoundTrip tests a formatted filter parses back the same, alone and as a filtered entry's filter.
func FuzzFilterRoundTrip(f *testing.F) {
	f.Add("title", "contains", "Ep 5 | Show: Finale", false)
	f.Add("title", "omits", "a||b", true)
	f.Add("title", "contains", "ends in: must", false)
	f.Add("title", "contains", `a\:any`, false)
	f.Add("description", "contains", `ends in a pipe|`, false)
	f.Add("title", "contains", `back\slash\\ and \| escapes`, true)
	f.Add("date", "omits", "", true)
	f.Add("my:field|x", "equals", ":leading colon", false)

	f.Fuzz(func(t *testing.T, field, filterType, value string, withMustAny bool) {
		field = strings.ToLower(strings.TrimSpace(field))
		value = strings.ToLower(value)
		mustAny := "must"
		if withMustAny && len(value)%2 == 1 {
			mustAny = "any"
		}
		want := sharedmodels.Filters{Field: field, FilterType: filterType, Value: value, MustAny: mustAny}
		if field == "" || strings.ToLower(value) != value || (filterType != "contains" && filterType != "omits" && filterType != "equals") {
			t.Skip()
		}

		formatted := sharedparsing.FormatFilter(want, withMustAny)
		got, err := sharedparsing.ParseFilterOps([]string{formatted}, withMustAny)
		if err != nil || len(got) != 1 || got[0] != want {
			t.Fatalf("round trip of %+v:\n formatted %q\n got %+v, %v", want, formatted, got, err)
		}
		if withMustAny {
			return
		}
		filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{formatted + sharedparsing.FilterSeparator + "title:set:x"})
		if err != nil || len(filtered) != 1 || filtered[0].Filters[0] != want {
			t.Fatalf("filtered round trip of %+v:\n formatted %q\n got %+v, %v", want, formatted, filtered, err)
		}
	})
}

func TestParseFilteredMetaOps(t *testing.T) {
	got, warnings, err := sharedparsing.ParseFilteredMetaOps([]string{
		"title:contains:cat||title:prefix:[CAT VIDEOS] ",
	})
	if err != nil {
		t.Fatalf("ParseFilteredMetaOps: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d entries, want 1", len(got))
	}

	if f := got[0].Filters[0]; f.Field != "title" || f.FilterType != "contains" || f.Value != "cat" {
		t.Errorf("filter = %+v", f)
	}
	if op := got[0].MetaOps[0]; op.Field != "title" || op.OpType != "prefix" || op.OpValue != "[CAT VIDEOS] " {
		t.Errorf("meta op = %+v", op)
	}
}

func TestParseFilteredMetaOpsRejectsMissingDivider(t *testing.T) {
	for _, entry := range []string{"title:contains:cat", "title:contains:cat|title:set:x"} {
		if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{entry}); err == nil {
			t.Errorf("%q has no '||' divider, so should be rejected", entry)
		}
	}
}

func TestParseFilteredFilenameOps(t *testing.T) {
	got, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:cat||prefix:[CATS] "})
	if err != nil {
		t.Fatalf("ParseFilteredFilenameOps: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d entries, want 1", len(got))
	}
	if op := got[0].FilenameOps[0]; op.OpType != "prefix" || op.OpValue != "[CATS] " {
		t.Errorf("filename op = %+v", op)
	}
}
