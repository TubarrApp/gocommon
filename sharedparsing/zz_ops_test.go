package sharedparsing_test

import (
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedparsing"
)

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
		"title:delete-date-tag:all:Ymd",
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
		"delete-date-tag:all:dm",
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
	if fops[0].OpFindString != "^a|b: " || fops[0].OpValue != "c" {
		t.Errorf("filename regex-replace = %+v", fops[0])
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
