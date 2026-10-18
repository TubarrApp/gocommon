package sharedfilters_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedfilters"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

func TestMatch(t *testing.T) {
	meta := map[string]any{
		"title":       "Cat Compilation",
		"description": "",
		"duration":    3600.0,
		"views":       "500",
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
		{"notequals hit", sharedmodels.Filters{Field: "title", FilterType: "notequals", Value: "dogs"}, true, false, false},

		// Presence-only form.
		{"presence contains hit", sharedmodels.Filters{Field: "title", FilterType: "contains"}, true, false, false},
		{"presence contains missing field", sharedmodels.Filters{Field: "absent", FilterType: "contains"}, false, false, false},
		{"presence omits missing field", sharedmodels.Filters{Field: "absent", FilterType: "omits"}, true, false, false},
		{"presence omits present field", sharedmodels.Filters{Field: "title", FilterType: "omits"}, false, false, false},
		{"presence needs a value", sharedmodels.Filters{Field: "title", FilterType: "equals"}, false, false, true},

		// Numeric.
		{"morethan hit", sharedmodels.Filters{Field: "duration", FilterType: "morethan", Value: "60"}, true, false, false},
		{"morethan miss", sharedmodels.Filters{Field: "duration", FilterType: "morethan", Value: "7200"}, false, false, false},
		{"lessthan hit", sharedmodels.Filters{Field: "views", FilterType: "lessthan", Value: "1000"}, true, false, false},
		{"numeric on missing field skips", sharedmodels.Filters{Field: "absent", FilterType: "morethan", Value: "1"}, false, true, false},
		{"numeric on empty field skips", sharedmodels.Filters{Field: "description", FilterType: "morethan", Value: "1"}, false, true, false},
		{"numeric on text field skips", sharedmodels.Filters{Field: "title", FilterType: "morethan", Value: "1"}, false, true, false},

		{"unknown type errors", sharedmodels.Filters{Field: "title", FilterType: "explodes", Value: "x"}, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, skipped, err := sharedfilters.Match(meta, tt.filter)
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
				{Field: "duration", FilterType: "morethan", Value: "60"},
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
			filters: []sharedmodels.Filters{{Field: "absent", FilterType: "morethan", Value: "1"}},
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
