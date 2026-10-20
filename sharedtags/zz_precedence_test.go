package sharedtags

import "testing"

// TestFirstString tests that the first key with a non-empty string value wins.
func TestFirstString(t *testing.T) {
	m := map[string]any{JFulltitle: "", JTitle: "Short", JLongDesc: 5, JDescription: "Desc"}
	if got := FirstString(m, TitleKeys); got != "Short" {
		t.Errorf("title: got %q, want %q (blank fulltitle skipped)", got, "Short")
	}
	if got := FirstString(m, DescriptionKeys); got != "Desc" {
		t.Errorf("description: got %q, want %q (non-string longdescription skipped)", got, "Desc")
	}
	m[JFulltitle] = "Full"
	if got := FirstString(m, TitleKeys); got != "Full" {
		t.Errorf("title: got %q, want %q", got, "Full")
	}
	if got := FirstString(m, DateKeys); got != "" {
		t.Errorf("date: got %q from metadata without one, want none", got)
	}
	if got := FirstString(map[string]any{JSynopsis: "Syn", JSummary: "Sum"}, DescriptionKeys); got != "Sum" {
		t.Errorf("description fallback: got %q, want %q", got, "Sum")
	}
	if got := FirstString(map[string]any{JTitle: "Short", "full_title": "FT"}, TitleKeys); got != "FT" {
		t.Errorf("full_title over title: got %q, want %q", got, "FT")
	}
}
