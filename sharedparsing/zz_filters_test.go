package sharedparsing_test

import (
	"testing"

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
	if _, err := sharedparsing.ParseFilterOps([]string{"duration:morethan"}, false); err == nil {
		t.Error("valueless morethan should be rejected")
	}
	if _, err := sharedparsing.ParseFilterOps([]string{"title:contains:cat:maybe"}, true); err == nil {
		t.Error("invalid must/any should be rejected")
	}
}

func TestParseFilteredMetaOps(t *testing.T) {
	got, warnings, err := sharedparsing.ParseFilteredMetaOps([]string{
		"title:contains:cat|title:prefix:[CAT VIDEOS] ",
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

// TestParseFilteredMetaOpsScopesChannel covers a leading channel URL applying to the
// whole entry rather than being mistaken for the filter.
func TestParseFilteredMetaOpsScopesChannel(t *testing.T) {
	got, _, err := sharedparsing.ParseFilteredMetaOps([]string{
		"https://example.com|title:contains:cat|title:prefix:[CAT] ",
	})
	if err != nil {
		t.Fatalf("ParseFilteredMetaOps: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("parsed %d entries, want 1", len(got))
	}
	if got[0].Filters[0].ChannelURL != "https://example.com" {
		t.Errorf("filter channel = %q", got[0].Filters[0].ChannelURL)
	}
	if got[0].MetaOps[0].ChannelURL != "https://example.com" {
		t.Errorf("op channel = %q", got[0].MetaOps[0].ChannelURL)
	}
	if got[0].Filters[0].Field != "title" || got[0].Filters[0].Value != "cat" {
		t.Errorf("filter = %+v", got[0].Filters[0])
	}
}

func TestParseFilteredMetaOpsRejectsMissingDivider(t *testing.T) {
	if _, _, err := sharedparsing.ParseFilteredMetaOps([]string{"title:contains:cat"}); err == nil {
		t.Error("an entry with no '|' divider should be rejected")
	}
}

func TestParseFilteredFilenameOps(t *testing.T) {
	got, _, err := sharedparsing.ParseFilteredFilenameOps([]string{"title:contains:cat|prefix:[CATS] "})
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
