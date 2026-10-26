package sharedtemplates

import (
	"slices"
	"testing"
)

// TestValidTag tests tag names and directives against Metarr's tags.
func TestValidTag(t *testing.T) {
	for tag, want := range map[string]bool{
		"year":              true,
		"year:yy":           true,
		" Year : YYYY ":     true,
		"year:yyy":          false,
		"year:":             true, // Empty directive means the default.
		"month":             true,
		"month:number":      true,
		"month:abbreviated": true,
		"month:WORD":        true,
		"month:short":       false,
		"day":               true,
		"day:number":        false, // Day takes no directives.
		"video_title:yy":    false,
		"channel_name":      false, // Not a Metarr tag.
		"nope":              false,
	} {
		if got := ValidTag(tag, MetarrTemplateTags); got != want {
			t.Errorf("ValidTag(%q) = %v, want %v", tag, got, want)
		}
	}
}

// TestDescribeTags tests the supported tags list shown in errors.
func TestDescribeTags(t *testing.T) {
	got := DescribeTags(MetarrTemplateTags)
	for _, want := range []string{"day", "month[:abbreviated|number|word]", "year[:yy|yyyy]"} {
		if !slices.Contains(got, want) {
			t.Errorf("DescribeTags() = %v, missing %q", got, want)
		}
	}
	if !slices.IsSorted(got) {
		t.Errorf("DescribeTags() = %v, not sorted", got)
	}
}
