package sharedparsing_test

import (
	"slices"
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

func TestParseMetaOps(t *testing.T) {
	in := []string{
		"director:set:Spielberg",
		"title:date-tag:suffix:ymd",
		"title:replace:old:new",
		"https://example.com|title:prefix:[CATS] ",
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
	if got[3].ChannelURL != "https://example.com" || got[3].OpValue != "[CATS] " {
		t.Errorf("url-scoped op = %+v", got[3])
	}
}

func TestParseMetaOpsSkipsInvalid(t *testing.T) {
	in := []string{
		"director:set:Spielberg",
		"too:few",
		"a:b:c:d:e",
		"title:not-an-op:x:y",
	}

	got, warnings, err := sharedparsing.ParseMetaOps(in)
	if err != nil {
		t.Fatalf("ParseMetaOps returned error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("parsed %d ops, want 1: %+v", len(got), got)
	}
	if len(warnings) != 3 {
		t.Errorf("got %d warnings, want 3: %v", len(warnings), warnings)
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
		"https://example.com|append:(new)",
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
	if got[3].ChannelURL != "https://example.com" {
		t.Errorf("url-scoped op = %+v", got[3])
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
		`description:append: \| piped \| value `,
		"title:set:md",
	}

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			first, _, err := sharedparsing.ParseMetaOps([]string{op})
			if err != nil {
				t.Fatalf("first parse of %q failed: %v", op, err)
			}

			formatted := sharedparsing.FormatMetaOp(first[0], false)

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
	}

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			first, _, err := sharedparsing.ParseFilenameOps([]string{op})
			if err != nil {
				t.Fatalf("first parse of %q failed: %v", op, err)
			}

			formatted := sharedparsing.FormatFilenameOp(first[0], false)

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

// TestMetaOpRoundTripWithURL covers the staged encode/decode, where the channel URL
// prefix must not disturb escapes belonging to the field separator.
func TestMetaOpRoundTripWithURL(t *testing.T) {
	op := sharedmodels.MetaOps{
		ChannelURL: "https://example.com",
		Field:      "title",
		OpType:     "set",
		OpValue:    `a|b:c`,
	}

	formatted := sharedparsing.FormatMetaOp(op, true)

	got, _, err := sharedparsing.ParseMetaOps([]string{formatted})
	if err != nil {
		t.Fatalf("re-parse of %q failed: %v", formatted, err)
	}
	if got[0] != op {
		t.Errorf("round trip changed the operation:\n formatted %q\n want %+v\n got  %+v", formatted, op, got[0])
	}
}

func TestFormatOpsFiltersByChannel(t *testing.T) {
	ops := []sharedmodels.MetaOps{
		{Field: "title", OpType: "set", OpValue: "global"},
		{ChannelURL: "https://a.com", Field: "title", OpType: "set", OpValue: "a"},
		{ChannelURL: "https://b.com", Field: "title", OpType: "set", OpValue: "b"},
	}

	got := sharedparsing.FormatMetaOps(ops, "https://a.com", false)
	want := []string{"title:set:global", "title:set:a"}
	if !slices.Equal(got, want) {
		t.Errorf("FormatMetaOps = %q, want %q", got, want)
	}

	if all := sharedparsing.FormatMetaOps(ops, "", false); len(all) != 3 {
		t.Errorf("empty chanURL should keep every op, got %q", all)
	}
}
