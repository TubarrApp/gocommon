package sharedfilters_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedfilters"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

func TestMatch(t *testing.T) {
	meta := map[string]any{
		"title":           "Cat Compilation",
		"description":     "",
		"duration":        3600.0,
		"duration_string": "7:24:44",
		"length":          "PT07H24M44S",
		"views":           "500",
	}

	tests := []struct {
		name        string
		filter      sharedmodels.Filters
		wantMatched bool
		wantSkipped bool
		wantErr     bool
	}{
		{"contains hit", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat"}, true, false, false},
		{"contains miss", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "dog"}, false, false, false},
		{"contains is case insensitive", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "COMPILATION"}, true, false, false},
		{"omits hit", sharedmodels.Filters{Field: "title", FilterType: "omits", Value: "dog"}, true, false, false},
		{"omits miss", sharedmodels.Filters{Field: "title", FilterType: "omits", Value: "cat"}, false, false, false},
		{"equals hit", sharedmodels.Filters{Field: "title", FilterType: "equals", Value: "cat compilation"}, true, false, false},
		{"not-equals hit", sharedmodels.Filters{Field: "title", FilterType: "not-equals", Value: "dogs"}, true, false, false},

		// Presence-only form.
		{"presence contains hit", sharedmodels.Filters{Field: "title", FilterType: "contains"}, true, false, false},
		{"presence contains missing field", sharedmodels.Filters{Field: "absent", FilterType: "contains"}, false, false, false},
		{"presence omits missing field", sharedmodels.Filters{Field: "absent", FilterType: "omits"}, true, false, false},
		{"presence omits present field", sharedmodels.Filters{Field: "title", FilterType: "omits"}, false, false, false},

		// Equals and not-equals with no value compare against "".
		{"equals empty on empty field", sharedmodels.Filters{Field: "description", FilterType: "equals"}, true, false, false},
		{"equals empty on missing field", sharedmodels.Filters{Field: "absent", FilterType: "equals"}, true, false, false},
		{"equals empty on set field", sharedmodels.Filters{Field: "title", FilterType: "equals"}, false, false, false},
		{"not-equals empty on set field", sharedmodels.Filters{Field: "title", FilterType: "not-equals"}, true, false, false},
		{"not-equals empty on missing field", sharedmodels.Filters{Field: "absent", FilterType: "not-equals"}, false, false, false},

		// Numeric.
		{"more-than hit", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "60"}, true, false, false},
		{"more-than miss", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "7200"}, false, false, false},
		{"less-than hit", sharedmodels.Filters{Field: "views", FilterType: "less-than", Value: "1000"}, true, false, false},
		{"numeric on missing field skips", sharedmodels.Filters{Field: "absent", FilterType: "more-than", Value: "1"}, false, true, false},
		{"numeric on empty field skips", sharedmodels.Filters{Field: "description", FilterType: "more-than", Value: "1"}, false, true, false},
		{"numeric on text field doesn't match", sharedmodels.Filters{Field: "title", FilterType: "more-than", Value: "1"}, false, false, false},
		{"more-than duration string", sharedmodels.Filters{Field: "duration_string", FilterType: "more-than", Value: "3600"}, true, false, false},
		{"less-than duration string", sharedmodels.Filters{Field: "duration_string", FilterType: "less-than", Value: "3600"}, false, false, false},
		{"more-than iso duration", sharedmodels.Filters{Field: "length", FilterType: "more-than", Value: "26000"}, true, false, false},

		// Regex, which keeps case unless the pattern sets (?i).
		{"regex hit", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `^Cat \w+$`}, true, false, false},
		{"regex is case sensitive", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `^cat`}, false, false, false},
		{"regex (?i) ignores case", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `(?i)^cat`}, true, false, false},
		{"regex on missing field", sharedmodels.Filters{Field: "absent", FilterType: "regex-contains", Value: `x`}, false, false, false},
		{"invalid regex errors", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `(`}, false, false, true},

		{"unknown type errors", sharedmodels.Filters{Field: "title", FilterType: "explodes", Value: "x"}, false, false, true},

		// A field list matches when any field does, but omits and not-equals need all of them to.
		{"list contains in a later field", sharedmodels.Filters{Field: "description,title", FilterType: "contains", Value: "cat"}, true, false, false},
		{"list contains in no field", sharedmodels.Filters{Field: "title,description", FilterType: "contains", Value: "dog"}, false, false, false},
		{"list trims spaces", sharedmodels.Filters{Field: "absent, title", FilterType: "contains", Value: "cat"}, true, false, false},
		{"list equals any field", sharedmodels.Filters{Field: "absent,title", FilterType: "equals", Value: "cat compilation"}, true, false, false},
		{"list regex any field", sharedmodels.Filters{Field: "description,title", FilterType: "regex-contains", Value: `^Cat`}, true, false, false},
		{"list omits in every field", sharedmodels.Filters{Field: "title,description", FilterType: "omits", Value: "dog"}, true, false, false},
		{"list omits fails on one field", sharedmodels.Filters{Field: "title,description", FilterType: "omits", Value: "cat"}, false, false, false},
		{"list not-equals fails on one field", sharedmodels.Filters{Field: "title,description", FilterType: "not-equals"}, false, false, false},
		{"list presence contains any field", sharedmodels.Filters{Field: "absent,title", FilterType: "contains"}, true, false, false},
		{"list presence omits every field", sharedmodels.Filters{Field: "absent,title", FilterType: "omits"}, false, false, false},
		{"list numeric any field", sharedmodels.Filters{Field: "absent,duration", FilterType: "more-than", Value: "1000"}, true, false, false},
		{"list numeric miss", sharedmodels.Filters{Field: "absent,duration", FilterType: "more-than", Value: "9000"}, false, false, false},
		{"list numeric skips when no field has a value", sharedmodels.Filters{Field: "absent,description", FilterType: "more-than", Value: "5"}, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, skipped, _, err := sharedfilters.Match(meta, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if matched != tt.wantMatched || skipped != tt.wantSkipped {
				t.Errorf("matched=%v skipped=%v, want matched=%v skipped=%v", matched, skipped, tt.wantMatched, tt.wantSkipped)
			}
		})
	}
}

func TestMatchAll(t *testing.T) {
	meta := map[string]any{"title": "Cat Compilation", "duration": 3600.0}

	tests := []struct {
		name    string
		filters []sharedmodels.Filters
		want    bool
	}{
		{
			name: "all hit",
			filters: []sharedmodels.Filters{
				{Field: "title", FilterType: "contains", Value: "cat"},
				{Field: "duration", FilterType: "more-than", Value: "60"},
			},
			want: true,
		},
		{
			name: "one miss fails the set",
			filters: []sharedmodels.Filters{
				{Field: "title", FilterType: "contains", Value: "cat"},
				{Field: "title", FilterType: "contains", Value: "dog"},
			},
			want: false,
		},
		{
			name:    "a skipped filter does not pass",
			filters: []sharedmodels.Filters{{Field: "absent", FilterType: "more-than", Value: "1"}},
			want:    false,
		},
		{
			name:    "a non-numeric field does not pass",
			filters: []sharedmodels.Filters{{Field: "title", FilterType: "more-than", Value: "1"}},
			want:    false,
		},
		{name: "no filters match trivially", filters: nil, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sharedfilters.MatchAll(meta, tt.filters)
			if err != nil {
				t.Fatalf("MatchAll: %v", err)
			}
			if got != tt.want {
				t.Errorf("MatchAll = %v, want %v", got, tt.want)
			}
		})
	}
}
