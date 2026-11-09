package sharedparsing_test

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// ******************************** Channels ******************************************************************************************

// TestNormalizeChannelURL covers which differently typed channel URLs normalize to the same channel.
func TestNormalizeChannelURL(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"identical", "https://google.com", "https://google.com", true},
		{"trailing slash on one side", "https://google.com/", "https://google.com", true},
		{"trailing slash on the other", "https://google.com", "https://google.com/", true},
		{"trailing slash on both", "https://google.com/", "https://google.com/", true},
		{"case differs in host", "https://GOOGLE.com", "https://google.com", true},
		{"case differs in path", "https://youtube.com/@ChannelA", "https://youtube.com/@channela", true},
		{"surrounding whitespace", "  https://google.com/  ", "https://google.com", true},
		{"path with slash and case", "https://youtube.com/@ChannelA/", "https://youtube.com/@channela", true},

		// Everything else must still differ: this forgives typing, not identity.
		{"different channel on same host", "https://youtube.com/@ChannelA", "https://youtube.com/@ChannelB", false},
		{"different host", "https://google.com", "https://youtube.com", false},
		{"www is not stripped", "https://www.google.com", "https://google.com", false},
		{"scheme is not stripped", "http://google.com", "https://google.com", false},
		{"host only against a path", "https://youtube.com", "https://youtube.com/@ChannelA", false},
		{"interior slash is not trimmed", "https://youtube.com//@ChannelA", "https://youtube.com/@ChannelA", false},
		{"both empty", "", "", true},
		{"one empty", "", "https://google.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedparsing.NormalizeChannelURL(tt.a) == sharedparsing.NormalizeChannelURL(tt.b); got != tt.want {
				t.Errorf("NormalizeChannelURL(%q) == NormalizeChannelURL(%q) is %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestStripURLScheme tests only a leading scheme is dropped, keeping subdomains like www.
func TestStripURLScheme(t *testing.T) {
	for in, want := range map[string]string{
		"https://www.youtube.com/@cats": "www.youtube.com/@cats",
		"http://example.com":            "example.com",
		"HTTPS://Example.com/a":         "Example.com/a",
		"youtube.com/@cats":             "youtube.com/@cats",
		"svn+ssh://host/repo":           "host/repo",
		"a.b-c2://host":                 "host",
		"a b://c":                       "a b://c", // Not a scheme.
		"a_b://c":                       "a_b://c",
		"://c":                          "://c",
	} {
		if got := sharedparsing.StripURLScheme(in); got != want {
			t.Errorf("StripURLScheme(%q) = %q, want %q", in, got, want)
		}
	}
}

// ******************************** Durations *****************************************************************************************

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

		// Colon-separated durations with fractional seconds.
		"00:47:35.120":       2855,
		"1:02:03.5":          3723,
		"62:03.9":            3723,
		"01:23:45.678000000": 5025,

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

		// A fraction only on the seconds, and only of digits.
		"1:02:03.x",
		"1:02:03.5.5",
		"1:02:.5",
		"1.5:02:03",
		"1:02.5:03",

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

// ******************************** Filters *******************************************************************************************

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
		if slices.Contains(sharedparsing.SplitMultiField(field), "") || strings.ToLower(value) != value || (filterType != "contains" && filterType != "omits" && filterType != "equals") {
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

// TestParseFilterFieldList tests a filter's comma-separated field list is kept as written and written back the same,
// and that a list holding an empty field is rejected.
func TestParseFilterFieldList(t *testing.T) {
	got, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title,fulltitle:contains:lion||prefix:[LIONS] "})
	if err != nil || len(got) != 1 {
		t.Fatalf("ParseFilteredFilenameOps = %+v, %v", got, err)
	}
	if f := got[0].Filters[0]; f.Field != "title,fulltitle" || f.FilterType != "contains" || f.Value != "lion" {
		t.Errorf("filter = %+v", f)
	}
	if formatted := sharedparsing.FormatFilters(got[0].Filters); formatted != "title,fulltitle:contains:lion" {
		t.Errorf("FormatFilters = %q", formatted)
	}

	for _, bad := range []string{
		"title,:contains:lion||prefix:x",
		",title:contains:lion||prefix:x",
		"title,,fulltitle:contains:lion||prefix:x",
	} {
		if _, _, err := sharedparsing.ParseFilteredFilenameOps([]string{bad}); err == nil {
			t.Errorf("%q has an empty field, so should be rejected", bad)
		}
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

// ******************************** Fuzz **********************************************************************************************

// FuzzMetaOpRoundTrip asserts that formatting an operation and parsing it back is
// lossless for arbitrary field values, which is what the Tubarr -> Metarr argv handoff
// relies on.
func FuzzMetaOpRoundTrip(f *testing.F) {
	f.Add("title", "set", "cats")
	f.Add("title", "set", "a|b:c")
	f.Add("description", "suffix", ` | piped : value `)
	f.Add("title", "set", `C:\Dogs`)
	f.Add("title", "set", "")
	f.Add("title", "set", `back\slash\`)
	f.Add("title", "set", "HELLO: THIS IS MY TEXT")
	f.Add("title", "keep-after", "| ")

	f.Fuzz(func(t *testing.T, field, opType, opValue string) {
		// Restrict to the three-part form so the shape is well defined.
		if opType == "" || isMultiPartOpType(opType) {
			t.Skip()
		}
		// A field or operation type holding a separator would change the parsed shape, as would a field holding
		// a comma or surrounding spaces, since parsing splits field lists and trims each field.
		if strings.ContainsAny(field+opType, ":|,") || field != strings.TrimSpace(field) {
			t.Skip()
		}
		// Parsing validates, so only valid operations can round trip.
		want := sharedmodels.MetaOps{Field: field, OpType: opType, OpValue: opValue}
		if sharedvalidation.ValidateMetaOps([]sharedmodels.MetaOps{want}) != nil {
			t.Skip()
		}

		formatted := sharedparsing.FormatMetaOp(want)
		got, _, err := sharedparsing.ParseMetaOps([]string{formatted})
		if err != nil {
			t.Fatalf("re-parse of %q (from %+v) failed: %v", formatted, want, err)
		}
		if got[0] != want {
			t.Errorf("round trip changed the operation:\n formatted %q\n want %+v\n got  %+v", formatted, want, got[0])
		}

		// Joined with another operation in a filtered entry, it still reads back the same.
		const other = "artist:set:y"
		if formatted == other {
			return
		}
		filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:x||" + formatted + sharedconsts.FilterJoiner + other})
		if err != nil || len(filtered) != 1 || len(filtered[0].MetaOps) != 2 || filtered[0].MetaOps[0] != want {
			t.Errorf("joined round trip changed the operation:\n formatted %q\n want %+v\n got  %+v, %v", formatted, want, filtered, err)
		}
	})
}

// isMultiPartOpType reports whether opType takes anything other than one value after it.
func isMultiPartOpType(opType string) bool {
	switch opType {
	case "date-tag", "delete-date-tag", "replace", "replace-suffix", "replace-prefix", "regex-replace":
		return true
	default:
		return false
	}
}

// FuzzRegexReplaceRoundTrip asserts that a regex-replace pattern survives formatting and
// re-parsing verbatim, including inside a filtered entry that is split on '||' first.
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
		formatted := sharedparsing.FormatMetaOp(want)

		got, _, err := sharedparsing.ParseMetaOps([]string{formatted})
		if err != nil {
			t.Fatalf("re-parse of %q failed: %v", formatted, err)
		}
		if got[0] != want {
			t.Errorf("round trip changed the operation:\n formatted %q\n want %+v\n got  %+v", formatted, want, got[0])
		}

		filtered, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:x||" + formatted})
		if err != nil {
			t.Fatalf("filtered re-parse of %q failed: %v", formatted, err)
		}
		if len(filtered) != 1 || len(filtered[0].MetaOps) != 1 || filtered[0].MetaOps[0] != want {
			t.Errorf("filtered round trip changed the operation:\n formatted %q\n got %+v", formatted, filtered)
		}
	})
}

// ******************************** Ops ***********************************************************************************************

func TestParseMetaOps(t *testing.T) {
	in := []string{
		"director:set:Spielberg",
		"title:date-tag:suffix:ymd",
		"title:replace:old:new",
		"title:prefix:[CATS] ",
	}

	got, warnings, err := sharedparsing.ParseMetaOps(in)
	if err != nil {
		t.Fatalf("ParseMetaOps returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(got) != len(in) {
		t.Fatalf("parsed %d ops, want %d: %+v", len(got), len(in), got)
	}

	if got[0].Field != "director" || got[0].OpType != "set" || got[0].OpValue != "Spielberg" {
		t.Errorf("set op = %+v", got[0])
	}
	if got[1].OpLoc != "suffix" || got[1].DateFormat != "ymd" {
		t.Errorf("date-tag op = %+v", got[1])
	}
	if got[2].OpFindString != "old" || got[2].OpValue != "new" {
		t.Errorf("replace op = %+v", got[2])
	}
	if got[3].OpValue != "[CATS] " {
		t.Errorf("prefix op = %+v", got[3])
	}
}

// TestParseMetaOpsRejectsInvalid tests that any malformed entry fails the batch, naming every one.
func TestParseMetaOpsRejectsInvalid(t *testing.T) {
	in := []string{
		"director:set:Spielberg",
		"too:few",
		"title:date-tag:prefix",
		"title:regex-replace:^unwrapped:x",
	}

	got, _, err := sharedparsing.ParseMetaOps(in)
	if err == nil {
		t.Fatalf("expected an error, got %+v", got)
	}
	for _, entry := range in[1:] {
		if !strings.Contains(err.Error(), entry) {
			t.Errorf("error %q does not name %q", err, entry)
		}
	}
}

// TestChannelURLPrefixRejected tests that a leading channel URL is read as part of the operation, so it fails as
// invalid.
func TestChannelURLPrefixRejected(t *testing.T) {
	const u = "https://example.com|"
	if _, _, err := sharedparsing.ParseMetaOps([]string{u + "title:set:cats"}); err == nil {
		t.Error("ParseMetaOps should reject a channel URL prefix")
	}
	if _, _, err := sharedparsing.ParseFilenameOps([]string{u + "prefix:[A] "}); err == nil {
		t.Error("ParseFilenameOps should reject a channel URL prefix")
	}
	if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{u + "title:contains:cat||title:set:cats"}); err == nil {
		t.Error("ParseFilteredMetaOps should reject a channel URL prefix")
	}
	if _, _, err := sharedparsing.ParseFilteredFilenameOps([]string{u + "title:contains:cat||prefix:[A] "}); err == nil {
		t.Error("ParseFilteredFilenameOps should reject a channel URL prefix")
	}
}

// TestParseMetaOpsValidates tests that well-formed but invalid operations fail parsing.
func TestParseMetaOpsValidates(t *testing.T) {
	for _, op := range []string{"title:explode:x", "all-titles:copy-to:writer", "all-dates:set:2024/05/10", "title:regex-replace:[[(]]:x"} {
		if _, _, err := sharedparsing.ParseMetaOps([]string{op}); err == nil {
			t.Errorf("ParseMetaOps(%q) should fail validation", op)
		}
	}
	if _, _, err := sharedparsing.ParseMetaOps([]string{"all-titles:paste-from:release_date"}); err != nil {
		t.Errorf("all-titles:paste-from should be valid: %v", err)
	}
	if _, _, err := sharedparsing.ParseFilenameOps([]string{"explode:x"}); err == nil {
		t.Error("ParseFilenameOps should fail validation on an unknown type")
	}
	if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:explodes:cat|title:set:x"}); err == nil {
		t.Error("ParseFilteredMetaOps should fail validation on an unknown filter type")
	}
}

func TestParseMetaOpsAllInvalid(t *testing.T) {
	if _, _, err := sharedparsing.ParseMetaOps([]string{"too:few"}); err == nil {
		t.Error("expected an error when no operation is usable")
	}
}

func TestParseMetaOpsDeduplicates(t *testing.T) {
	in := []string{"director:set:Spielberg", "director:set:Spielberg"}

	got, warnings, err := sharedparsing.ParseMetaOps(in)
	if err != nil {
		t.Fatalf("ParseMetaOps returned error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("parsed %d ops, want 1", len(got))
	}
	if len(warnings) == 0 {
		t.Error("expected a warning describing the dropped duplicate")
	}
}

// TestParseMetaOpsMultiField tests a comma-separated field list gives one operation per trimmed field, leaving
// commas in the value alone, and that a duplicate warning names only the skipped field's operation.
func TestParseMetaOpsMultiField(t *testing.T) {
	got, warnings, err := sharedparsing.ParseMetaOps([]string{"actor, director ,artist:set:Example, Inc", "actor:set:Example, Inc"})
	if err != nil {
		t.Fatalf("ParseMetaOps: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("parsed %d ops, want 3: %+v", len(got), got)
	}
	for i, field := range []string{"actor", "director", "artist"} {
		if op := got[i]; op.Field != field || op.OpType != "set" || op.OpValue != "Example, Inc" {
			t.Errorf("op %d = %+v, want %s set to %q", i, op, field, "Example, Inc")
		}
	}
	if len(warnings) != 1 || warnings[0].Entry != "actor:set:Example, Inc" {
		t.Errorf("warnings = %v, want one for the repeated actor operation", warnings)
	}

	for _, bad := range []string{"title,:set:x", "title,,description:set:x"} {
		if _, _, err := sharedparsing.ParseMetaOps([]string{bad}); err == nil {
			t.Errorf("%q has an empty field, so should be rejected", bad)
		}
	}
}

func TestParseFilenameOps(t *testing.T) {
	in := []string{
		"prefix:[DOG VIDEOS]",
		"date-tag:prefix:ymd",
		"replace-suffix:_1:",
		"suffix:(new)",
	}

	got, warnings, err := sharedparsing.ParseFilenameOps(in)
	if err != nil {
		t.Fatalf("ParseFilenameOps returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(got) != len(in) {
		t.Fatalf("parsed %d ops, want %d: %+v", len(got), len(in), got)
	}

	if got[1].OpLoc != "prefix" || got[1].DateFormat != "ymd" {
		t.Errorf("date-tag op = %+v", got[1])
	}
	if got[2].OpFindString != "_1" || got[2].OpValue != "" {
		t.Errorf("replace-suffix op = %+v", got[2])
	}
	if got[3].OpType != "suffix" || got[3].OpValue != "(new)" {
		t.Errorf("suffix op = %+v", got[3])
	}
}

// TestMetaOpRoundTrip is the property the Tubarr -> Metarr handoff depends on: an
// operation must survive being parsed, formatted and re-parsed unchanged, including
// values holding the separators.
func TestMetaOpRoundTrip(t *testing.T) {
	ops := []string{
		"director:set:Spielberg",
		"title:date-tag:suffix:ymd",
		"title:delete-date-tag",
		"title,description:delete-date-tag",
		"title:replace:old:new",
		"title:replace-prefix:pre:post",
		`title:set:cats\: the sequel`,
		`title:replace:C\:\\Dogs:x`,
		`description:suffix: \| piped \| value `,
		"title:set:a|b:c",
		"title:set:md",
		"title:set:HELLO: THIS IS MY TEXT",
		"title:replace:old:new: text",
		`title:regex-replace:[[^Kino Casino [^|]*\| ]]:`,
		`title:regex-replace:[[(\d+):(\d+)]]:$2\:$1`,
		`title:regex-replace:[[[[:digit:]]]]:#`,
		"title:keep-after:| ",
	}

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			first, _, err := sharedparsing.ParseMetaOps([]string{op})
			if err != nil {
				t.Fatalf("first parse of %q failed: %v", op, err)
			}

			formatted := sharedparsing.FormatMetaOp(first[0])

			second, _, err := sharedparsing.ParseMetaOps([]string{formatted})
			if err != nil {
				t.Fatalf("re-parse of %q failed: %v", formatted, err)
			}
			if first[0] != second[0] {
				t.Errorf("round trip changed the operation:\n original  %q\n formatted %q\n first     %+v\n second    %+v", op, formatted, first[0], second[0])
			}
		})
	}
}

func TestFilenameOpRoundTrip(t *testing.T) {
	ops := []string{
		"prefix:[DOG VIDEOS]",
		"date-tag:prefix:ymd",
		"delete-date-tag",
		"replace-suffix:_1:",
		"replace:old:new",
		`prefix:[CATS\: THE SEQUEL] `,
		"prefix:[CATS: THE SEQUEL] ",
		`regex-replace:[[^Kino Casino [^|]*\| ]]:`,
		"keep-before: |",
	}

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			first, _, err := sharedparsing.ParseFilenameOps([]string{op})
			if err != nil {
				t.Fatalf("first parse of %q failed: %v", op, err)
			}

			formatted := sharedparsing.FormatFilenameOp(first[0])

			second, _, err := sharedparsing.ParseFilenameOps([]string{formatted})
			if err != nil {
				t.Fatalf("re-parse of %q failed: %v", formatted, err)
			}
			if first[0] != second[0] {
				t.Errorf("round trip changed the operation:\n original  %q\n formatted %q\n first     %+v\n second    %+v", op, formatted, first[0], second[0])
			}
		})
	}
}

// TestOpTypeCaseInsensitive covers the fixed vocabularies being matched regardless of
// case, which Metarr has always allowed and Tubarr previously rejected.
func TestOpTypeCaseInsensitive(t *testing.T) {
	got, warnings, err := sharedparsing.ParseMetaOps([]string{
		"title:SET:Cats",
		"title:DATE-TAG:PREFIX:ymd",
		"other:Replace:Old:New",
	})
	if err != nil {
		t.Fatalf("ParseMetaOps: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(got) != 3 {
		t.Fatalf("parsed %d ops, want 3: %+v", len(got), got)
	}

	if got[0].OpType != "set" || got[0].OpValue != "Cats" {
		t.Errorf("value case must be preserved, got %+v", got[0])
	}
	if got[1].OpType != "date-tag" || got[1].OpLoc != "prefix" {
		t.Errorf("date-tag not normalized: %+v", got[1])
	}
	if got[2].OpType != "replace" || got[2].OpFindString != "Old" || got[2].OpValue != "New" {
		t.Errorf("replace not normalized or values altered: %+v", got[2])
	}

	fops, _, err := sharedparsing.ParseFilenameOps([]string{"PREFIX:[CATS] ", "Date-Tag:SUFFIX:ymd"})
	if err != nil {
		t.Fatalf("ParseFilenameOps: %v", err)
	}
	if len(fops) != 2 {
		t.Fatalf("ParseFilenameOps gave %d operations, want 2", len(fops))
	}
	if fops[0].OpType != "prefix" || fops[0].OpValue != "[CATS] " {
		t.Errorf("filename prefix not normalized or value altered: %+v", fops[0])
	}
	if fops[1].OpType != "date-tag" || fops[1].OpLoc != "suffix" {
		t.Errorf("filename date-tag not normalized: %+v", fops[1])
	}
}

// TestDeleteDateTagTakesNothing covers delete-date-tag standing alone, in any case and with stray spaces or a trailing
// ':', and rejecting the old location and format.
func TestDeleteDateTagTakesNothing(t *testing.T) {
	for _, op := range []string{"title:delete-date-tag", "title:DELETE-DATE-TAG", "title:delete-date-tag ", "title:delete-date-tag:"} {
		got, _, err := sharedparsing.ParseMetaOps([]string{op})
		if err != nil || len(got) != 1 || got[0] != (sharedmodels.MetaOps{Field: "title", OpType: "delete-date-tag"}) {
			t.Errorf("ParseMetaOps(%q) = %+v, %v", op, got, err)
		}
	}
	for _, op := range []string{"delete-date-tag", "Delete-Date-Tag", "delete-date-tag "} {
		got, _, err := sharedparsing.ParseFilenameOps([]string{op})
		if err != nil || len(got) != 1 || got[0] != (sharedmodels.FilenameOps{OpType: "delete-date-tag"}) {
			t.Errorf("ParseFilenameOps(%q) = %+v, %v", op, got, err)
		}
	}

	for _, op := range []string{"title:delete-date-tag:all:ymd", "title:delete-date-tag:prefix"} {
		if _, _, err := sharedparsing.ParseMetaOps([]string{op}); err == nil {
			t.Errorf("ParseMetaOps(%q) accepted a location", op)
		}
	}
	if _, _, err := sharedparsing.ParseFilenameOps([]string{"delete-date-tag:all:ymd"}); err == nil {
		t.Error("ParseFilenameOps accepted a location")
	}

	// A stored model's stale location and format are dropped when it is written back.
	stale := sharedmodels.MetaOps{Field: "title", OpType: "delete-date-tag", OpLoc: "all", DateFormat: "ymd"}
	if got := sharedparsing.FormatMetaOp(stale); got != "title:delete-date-tag" {
		t.Errorf("FormatMetaOp(%+v) = %q", stale, got)
	}
	if got := sharedparsing.FormatFilenameOp(sharedmodels.FilenameOps{OpType: "delete-date-tag", OpLoc: "prefix", DateFormat: "ymd"}); got != "delete-date-tag" {
		t.Errorf("FormatFilenameOp = %q", got)
	}

	// Inside filtered entries, before and after another operation.
	meta, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:x||title:delete-date-tag&&title:prefix:[x] "})
	if err != nil || len(meta) != 1 || len(meta[0].MetaOps) != 2 || meta[0].MetaOps[0].OpType != "delete-date-tag" {
		t.Errorf("ParseFilteredMetaOps = %+v, %v", meta, err)
	}
	files, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:x||prefix:[x] && delete-date-tag"})
	if err != nil || len(files) != 1 || len(files[0].FilenameOps) != 2 || files[0].FilenameOps[1].OpType != "delete-date-tag" {
		t.Errorf("ParseFilteredFilenameOps = %+v, %v", files, err)
	}
}

// TestUnescapedPipeInValue covers values holding a bare '|', such as "Title | Channel".
func TestUnescapedPipeInValue(t *testing.T) {
	got, warnings, err := sharedparsing.ParseMetaOps([]string{"title:set:Cats | Dogs"})
	if err != nil {
		t.Fatalf("ParseMetaOps: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d ops, want 1", len(got))
	}
	if got[0].OpValue != "Cats | Dogs" {
		t.Errorf("OpValue = %q, want %q", got[0].OpValue, "Cats | Dogs")
	}
}

// TestDuplicatesOnlyWarn keeps duplicate input non-fatal, described only in warnings.
func TestDuplicatesOnlyWarn(t *testing.T) {
	got, warnings, err := sharedparsing.ParseMetaOps([]string{"title:set:cats", "title:set:cats"})
	if err != nil {
		t.Fatalf("ParseMetaOps: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("parsed %d ops, want 1", len(got))
	}
	if len(warnings) != 1 || warnings[0].Kind != sharedparsing.WarnDuplicate {
		t.Errorf("warnings = %v, want one duplicate", warnings)
	}
}

// TestLastPartNeedsNoEscaping covers the last part of an operation being the whole
// remainder, so ':' and template braces in it are literal.
func TestLastPartNeedsNoEscaping(t *testing.T) {
	tests := []struct {
		op        string
		wantFind  string
		wantValue string
	}{
		{"title:set:HELLO: THIS IS MY TEXT", "", "HELLO: THIS IS MY TEXT"},
		{"title:set:{{CAT VIDEOS}} This: Is Cool", "", "{{CAT VIDEOS}} This: Is Cool"},
		{`title:set:cats\: the sequel`, "", "cats: the sequel"},
		{"title:replace:old:new: text", "old", "new: text"},
		{"title:keep-after:| ", "", "| "},
		{"title:suffix: : x", "", " : x"},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, warnings, err := sharedparsing.ParseMetaOps([]string{tt.op})
			if err != nil || len(warnings) != 0 {
				t.Fatalf("ParseMetaOps(%q): err %v, warnings %v", tt.op, err, warnings)
			}
			if got[0].OpFindString != tt.wantFind || got[0].OpValue != tt.wantValue {
				t.Errorf("got find %q value %q, want %q %q", got[0].OpFindString, got[0].OpValue, tt.wantFind, tt.wantValue)
			}
		})
	}
}

// TestRegexReplacePattern covers the "[[ ]]" pattern being taken verbatim, with no
// escaping and with ':' and '|' inside it.
func TestRegexReplacePattern(t *testing.T) {
	tests := []struct {
		op          string
		wantPattern string
		wantValue   string
	}{
		{`title:regex-replace:[[^Kino Casino [^|]*\| ]]:`, `^Kino Casino [^|]*\| `, ""},
		{`title:regex-replace:[[^Kino Casino [^|]*\| ]]`, `^Kino Casino [^|]*\| `, ""},
		{`title:regex-replace:[[(\d+):(\d+)]]:$2:$1`, `(\d+):(\d+)`, "$2:$1"},
		{`title:regex-replace:[[[[:digit:]]+]]:#`, `[[:digit:]]+`, "#"},
		{`title:regex-replace:[[\d{4}}]]:x`, `\d{4}}`, "x"},
		{`title:REGEX-REPLACE:[[a|b]]:{{uploader}}`, "a|b", "{{uploader}}"},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			got, warnings, err := sharedparsing.ParseMetaOps([]string{tt.op})
			if err != nil || len(warnings) != 0 {
				t.Fatalf("ParseMetaOps(%q): err %v, warnings %v", tt.op, err, warnings)
			}
			if got[0].OpType != "regex-replace" || got[0].OpFindString != tt.wantPattern || got[0].OpValue != tt.wantValue {
				t.Errorf("got %+v, want pattern %q value %q", got[0], tt.wantPattern, tt.wantValue)
			}
		})
	}

	fops, _, err := sharedparsing.ParseFilenameOps([]string{"regex-replace:[[^a|b: ]]:c"})
	if err != nil {
		t.Fatalf("ParseFilenameOps: %v", err)
	}
	if len(fops) > 0 {
		if fops[0].OpFindString != "^a|b: " || fops[0].OpValue != "c" {
			t.Errorf("filename regex-replace = %+v", fops[0])
		}
	}

	if _, _, err := sharedparsing.ParseMetaOps([]string{"title:regex-replace:^unwrapped:x"}); err == nil {
		t.Error("an unwrapped pattern should be rejected")
	}
}

// TestFilteredOpTakesRemainder covers a filtered entry being one filter and one
// operation, divided at the first '||', so a lone '|' needs no escaping on either side.
func TestFilteredOpTakesRemainder(t *testing.T) {
	tests := []struct {
		entry     string
		wantType  string
		wantFind  string
		wantValue string
	}{
		{"title:contains:Kino||title:keep-after:| ", "keep-after", "", "| "},
		{"title:contains:Kino||title:set:Cats | Dogs | Birds", "set", "", "Cats | Dogs | Birds"},
		{"title:contains:Kino||title:regex-replace:[[^a|b ]]:x|y", "regex-replace", "^a|b ", "x|y"},
		{"title:contains:a | b: c||title:suffix:!", "suffix", "", "!"},
	}

	for _, tt := range tests {
		t.Run(tt.entry, func(t *testing.T) {
			got, warnings, err := sharedparsing.ParseFilteredMetaOps([]string{tt.entry})
			if err != nil || len(warnings) != 0 {
				t.Fatalf("ParseFilteredMetaOps: err %v, warnings %v", err, warnings)
			}
			if len(got) != 1 || len(got[0].Filters) != 1 || len(got[0].MetaOps) != 1 {
				t.Fatalf("got %+v, want one filter and one operation", got)
			}
			op := got[0].MetaOps[0]
			if op.OpType != tt.wantType || op.OpFindString != tt.wantFind || op.OpValue != tt.wantValue {
				t.Errorf("op = %+v", op)
			}
		})
	}

	// The filter value is the rest of the filter, so ':' and a lone '|' are taken as written.
	got, _, _ := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:Ep 5 | Show: Finale||title:suffix:!"})
	if len(got) != 1 || len(got[0].Filters) != 1 {
		t.Fatalf("ParseFilteredMetaOps = %+v, want one entry with one filter", got)
	}
	if got[0].Filters[0].Value != "ep 5 | show: finale" {
		t.Errorf("filter value = %+v", got[0].Filters[0])
	}
	// A '|' at the end of the value is escaped, so it isn't read as part of the divider.
	if got, _, _ = sharedparsing.ParseFilteredMetaOps([]string{`title:contains:a\|||title:suffix:!`}); len(got) != 1 || got[0].Filters[0].Value != "a|" {
		t.Errorf("escaped trailing '|' = %+v", got)
	}

	fgot, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:Kino||keep-after:| "})
	if err != nil {
		t.Fatalf("ParseFilteredFilenameOps: %v", err)
	}
	if len(fgot) != 1 || len(fgot[0].FilenameOps) != 1 {
		t.Fatalf("ParseFilteredFilenameOps = %+v, want one entry with one operation", fgot)
	}
	if op := fgot[0].FilenameOps[0]; op.OpType != "keep-after" || op.OpValue != "| " {
		t.Errorf("filename op = %+v", op)
	}
}

// ******************************** Paths *********************************************************************************************

// TestSanitizePathComponent tests that names can't split into directories or exceed the byte limit.
func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		name, in, want string
		max            int
	}{
		{"slash", "AC/DC Live", "AC-DC Live", 255},
		{"backslash", `a\b`, "a-b", 255},
		{"parent", "..", "--", 255},
		{"traversal", "../../etc", "..-..-etc", 255},
		{"nul", "a\x00b", "ab", 255},
		{"line breaks", "line one\nline two\r\n", "line one line two", 255},
		{"other controls", "a\tb\x7fc\u0085d", "a b c d", 255},
		{"trimmed", "  title  ", "title", 255},
		{"cut", "abcdef", "abc", 3},
		{"cut keeps runes whole", "aé", "a", 2},
		{"cut leaves parent", ".." + strings.Repeat(" ", 300) + "x", "--", 255},
		{"cut leaves current", "." + strings.Repeat(" ", 300) + "x", "-", 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sharedparsing.SanitizePathComponent(tt.in, tt.max); got != tt.want {
				t.Fatalf("SanitizePathComponent(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}

	long := sharedparsing.SanitizePathComponent(strings.Repeat("é", 200), sharedconsts.MaxNameBytes)
	if len(long) > sharedconsts.MaxNameBytes || !utf8.ValidString(long) {
		t.Fatalf("long name is %d bytes (valid UTF-8: %v), want at most %d valid bytes", len(long), utf8.ValidString(long), sharedconsts.MaxNameBytes)
	}
}

// ******************************** Strings *******************************************************************************************

func TestSplitAndRemoveEscapes(t *testing.T) {
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
				t.Errorf("SplitAndRemoveEscapes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStagedSplitPreservesInnerEscapes is the property that decides the escaping rule:
// dividing a filter from its operation must not consume escapes meant for the decoding pass.
func TestStagedSplitPreservesInnerEscapes(t *testing.T) {
	parts := sharedparsing.SplitAndKeepEscapes(`title:contains:a\:b||title:set:c\:d`, '|')
	if len(parts) != 3 || parts[1] != "" {
		t.Fatalf("SplitAndKeepEscapes = %q, want filter, empty, op", parts)
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
			t.Errorf("SplitAndRemoveEscapes(%q) = %q, want %q", want, got, in)
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

// TestShellSplit tests shell syntax stays plain text, as no shell runs these args, while quoting works as usual.
func TestShellSplit(t *testing.T) {
	tests := []struct {
		in      string
		want    []string
		wantErr bool
	}{
		{"--referer https://x/?a=1&b=2 --no-playlist", []string{"--referer", "https://x/?a=1&b=2", "--no-playlist"}, false},
		{"--match-filter duration>60 --no-playlist", []string{"--match-filter", "duration>60", "--no-playlist"}, false},
		{"-filter_complex [0:v]split[a][b];[a]hflip[c] -c:a aac", []string{"-filter_complex", "[0:v]split[a][b];[a]hflip[c]", "-c:a", "aac"}, false},
		{"-o %(title)s.%(ext)s --no-part", []string{"-o", "%(title)s.%(ext)s", "--no-part"}, false},
		{"-vf select=eq(n,0) -an", []string{"-vf", "select=eq(n,0)", "-an"}, false},
		{"2>file x", []string{"2>file", "x"}, false},
		{"`echo hi` x", []string{"`echo", "hi`", "x"}, false},
		{"ok ; --exec rm", []string{"ok", ";", "--exec", "rm"}, false},
		{`--title 'a;b' c\;d "e|f"`, []string{"--title", "a;b", "c;d", "e|f"}, false},
		{`--add-headers "Referer: a b"`, []string{"--add-headers", "Referer: a b"}, false},
		{`unterminated "quote`, nil, true},
		{`trailing\`, nil, true},
	}

	for _, tt := range tests {
		got, err := sharedparsing.ShellSplit(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("ShellSplit(%q) error = %v, want error %v", tt.in, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !slices.Equal(got, tt.want) {
			t.Errorf("ShellSplit(%q) = %q, want %q", tt.in, got, tt.want)
		}
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
