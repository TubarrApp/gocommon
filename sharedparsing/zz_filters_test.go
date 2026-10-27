package sharedparsing_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
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
	f.Add("title", "contains", "tom && jerry&", false)
	f.Add("a&&b&", "equals", "&&&", false)

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
		filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{formatted + sharedconsts.FilterSeparator + "title:set:x"})
		if err != nil || len(filtered) != 1 || filtered[0].Filters[0] != want {
			t.Fatalf("filtered round trip of %+v:\n formatted %q\n got %+v, %v", want, formatted, filtered, err)
		}

		// Before and after another filter in the same entry.
		other := sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "60", MustAny: "must"}
		for _, filters := range [][]sharedmodels.Filters{{want, other}, {other, want}} {
			entry := sharedparsing.FormatFilters(filters) + sharedconsts.FilterSeparator + "title:set:x"
			filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{entry})
			if err != nil || len(filtered) != 1 || !slices.Equal(filtered[0].Filters, filters) {
				t.Fatalf("joined round trip of %+v:\n formatted %q\n got %+v, %v", filters, entry, filtered, err)
			}
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

// TestParseFilteredMetaOpsMultiField tests a field list in the operation gives one operation per field under the
// same filters, while a comma in a filter value stays part of the value.
func TestParseFilteredMetaOpsMultiField(t *testing.T) {
	got, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:lizards, frogs||title,description:suffix: [lizards]"})
	if err != nil || len(got) != 1 {
		t.Fatalf("ParseFilteredMetaOps = %+v, %v", got, err)
	}

	want := []sharedmodels.Filters{{Field: "title", FilterType: "contains", Value: "lizards, frogs", MustAny: "must"}}
	if !slices.Equal(got[0].Filters, want) {
		t.Errorf("filters = %+v, want %+v", got[0].Filters, want)
	}

	ops := got[0].MetaOps
	if len(ops) != 2 || ops[0].Field != "title" || ops[1].Field != "description" {
		t.Fatalf("meta ops = %+v, want title then description", ops)
	}
	for _, op := range ops {
		if op.OpType != "suffix" || op.OpValue != " [lizards]" {
			t.Errorf("meta op = %+v, want suffix %q", op, " [lizards]")
		}
	}
}

// TestParseFilteredOpsMultiOp tests operations joined by "&&" all apply under the entry's filters, with spaces after
// a "&&" dropped and an escaped "\&&" kept in a value, and that formatted operations joined this way read back the same.
func TestParseFilteredOpsMultiOp(t *testing.T) {
	meta, _, err := sharedparsing.ParseFilteredMetaOps([]string{
		"title:contains:cats||title:suffix: [cats]&&description:prefix:[cat videos] ",
		`title:contains:tom||title,description:set:Tom \&& Jerry&& artist:set:Hanna`,
	})
	if err != nil || len(meta) != 2 {
		t.Fatalf("ParseFilteredMetaOps = %+v, %v", meta, err)
	}
	want := []sharedmodels.MetaOps{
		{Field: "title", OpType: "suffix", OpValue: " [cats]"},
		{Field: "description", OpType: "prefix", OpValue: "[cat videos] "},
	}
	if !slices.Equal(meta[0].MetaOps, want) {
		t.Errorf("meta ops = %+v, want %+v", meta[0].MetaOps, want)
	}
	want = []sharedmodels.MetaOps{
		{Field: "title", OpType: "set", OpValue: "Tom && Jerry"},
		{Field: "description", OpType: "set", OpValue: "Tom && Jerry"},
		{Field: "artist", OpType: "set", OpValue: "Hanna"},
	}
	if !slices.Equal(meta[1].MetaOps, want) {
		t.Errorf("meta ops = %+v, want %+v", meta[1].MetaOps, want)
	}

	filename, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:lions||set:{{fulltitle}}&&prefix:[Animals] &&suffix: - Animal Kingdom"})
	if err != nil || len(filename) != 1 {
		t.Fatalf("ParseFilteredFilenameOps = %+v, %v", filename, err)
	}
	wantFilename := []sharedmodels.FilenameOps{
		{OpType: "set", OpValue: "{{fulltitle}}"},
		{OpType: "prefix", OpValue: "[Animals] "},
		{OpType: "suffix", OpValue: " - Animal Kingdom"},
	}
	if !slices.Equal(filename[0].FilenameOps, wantFilename) {
		t.Errorf("filename ops = %+v, want %+v", filename[0].FilenameOps, wantFilename)
	}

	// A doubled or trailing '&' in a value is escaped, so formatted operations can be joined.
	ops := []sharedmodels.MetaOps{
		{Field: "title", OpType: "set", OpValue: "R&&B&"},
		{Field: "artist", OpType: "set", OpValue: `back\slash`},
	}
	formatted := sharedparsing.FormatMetaOps(ops)
	if formatted[0] != `title:set:R\&&B\&` {
		t.Errorf("FormatMetaOp = %q", formatted[0])
	}
	entry := "title:contains:x||" + strings.Join(formatted, sharedconsts.FilterJoiner)
	got, _, err := sharedparsing.ParseFilteredMetaOps([]string{entry})
	if err != nil || len(got) != 1 || !slices.Equal(got[0].MetaOps, ops) {
		t.Errorf("round trip of %q = %+v, %v", entry, got, err)
	}

	for _, bad := range []string{
		"title:contains:cats||title:set:a&&",                  // An empty operation.
		"title:contains:cats||&&title:set:a",                  // An empty operation.
		"title:contains:cats||title:regex-replace:[[a&&b]]:c", // A "&&" in a pattern.
	} {
		if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{bad}); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	if _, _, err := sharedparsing.ParseMetaOps([]string{"title:regex-replace:[[a&&b]]:c"}); err == nil {
		t.Error(`a regex-replace pattern holding "&&" was accepted`)
	}
}

// TestParseFilteredOpsSeveralFilters tests an entry's filters joined by "&&" are all read, and written back the same.
func TestParseFilteredOpsSeveralFilters(t *testing.T) {
	want := []sharedmodels.Filters{
		{Field: "title", FilterType: "contains", Value: "cats", MustAny: "must"},
		{Field: "duration", FilterType: "more-than", Value: "3600", MustAny: "must"},
	}

	meta, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:cats&&duration:more-than:3600||title:prefix:[LONG CAT VIDEO] "})
	if err != nil || len(meta) != 1 || !slices.Equal(meta[0].Filters, want) {
		t.Fatalf("ParseFilteredMetaOps = %+v, %v, want filters %+v", meta, err, want)
	}
	if op := meta[0].MetaOps[0]; op.Field != "title" || op.OpType != "prefix" || op.OpValue != "[LONG CAT VIDEO] " {
		t.Errorf("meta op = %+v", op)
	}

	filename, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:cats&&duration:more-than:3600||prefix:[LONG CAT VIDEO] "})
	if err != nil || len(filename) != 1 || !slices.Equal(filename[0].Filters, want) {
		t.Fatalf("ParseFilteredFilenameOps = %+v, %v, want filters %+v", filename, err, want)
	}

	if got := sharedparsing.FormatFilters(want); got != "title:contains:cats&&duration:more-than:3600" {
		t.Errorf("FormatFilters = %q", got)
	}

	// A "&&" or trailing '&' in a value is escaped, so it stays part of the value.
	escaped := []sharedmodels.Filters{
		{Field: "title", FilterType: "contains", Value: "tom && jerry&", MustAny: "must"},
		{Field: "title", FilterType: "omits", Value: "r&b", MustAny: "must"},
	}
	formatted := sharedparsing.FormatFilters(escaped)
	if formatted != `title:contains:tom \&& jerry\&&&title:omits:r&b` {
		t.Errorf("FormatFilters = %q", formatted)
	}
	got, _, err := sharedparsing.ParseFilteredMetaOps([]string{formatted + "||title:set:x"})
	if err != nil || len(got) != 1 || !slices.Equal(got[0].Filters, escaped) {
		t.Errorf("round trip of %q = %+v, %v", formatted, got, err)
	}

	for _, bad := range []string{"title:contains:cats&&||title:set:x", "&&title:contains:cats||title:set:x"} {
		if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{bad}); err == nil {
			t.Errorf("%q has an empty filter, so should be rejected", bad)
		}
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
