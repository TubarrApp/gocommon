package sharedvalidation_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

func TestValidateMetaOps(t *testing.T) {
	tests := []struct {
		name    string
		op      sharedmodels.MetaOps
		wantErr bool
	}{
		{"set", sharedmodels.MetaOps{Field: "title", OpType: "set", OpValue: "cats"}, false},
		{"paste-from", sharedmodels.MetaOps{Field: "title", OpType: "paste-from", OpValue: "x"}, false},
		{"unknown type", sharedmodels.MetaOps{Field: "title", OpType: "explode"}, true},
		{"empty field", sharedmodels.MetaOps{OpType: "set", OpValue: "cats"}, true},
		{"date-tag ok", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "ymd"}, false},
		{"date-tag two-char format", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "md"}, false},
		{"date-tag rejects all", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "all", DateFormat: "ymd"}, true},
		{"date-tag bad format", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "nope"}, true},
		{"delete-date-tag allows all", sharedmodels.MetaOps{Field: "title", OpType: "delete-date-tag", OpLoc: "all", DateFormat: "ymd"}, false},
		{"regex-replace ok", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: `^\d+ \| `}, false},
		{"regex-replace bad pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: "("}, true},
		{"regex-replace empty pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace"}, true},
		{"regex-replace unrepresentable pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: "[[:digit:]]:"}, true},
		{"keep-after ok", sharedmodels.MetaOps{Field: "title", OpType: "keep-after", OpValue: "| "}, false},
		{"keep-before empty", sharedmodels.MetaOps{Field: "title", OpType: "keep-before"}, true},
		{"all-titles set", sharedmodels.MetaOps{Field: "all-titles", OpType: "set", OpValue: "cats"}, false},
		{"all-titles suffix", sharedmodels.MetaOps{Field: "all-titles", OpType: "suffix", OpValue: "cats"}, false},
		{"all-titles paste-from", sharedmodels.MetaOps{Field: "all-titles", OpType: "paste-from", OpValue: "release_date"}, false},
		{"copy-to into all-titles", sharedmodels.MetaOps{Field: "release_date", OpType: "copy-to", OpValue: "all-titles"}, false},
		{"all-titles copy-to rejected", sharedmodels.MetaOps{Field: "all-titles", OpType: "copy-to", OpValue: "writer"}, true},
		{"paste-from all-titles rejected", sharedmodels.MetaOps{Field: "writer", OpType: "paste-from", OpValue: "all-titles"}, true},
		{"all-dates paste-from", sharedmodels.MetaOps{Field: "all-dates", OpType: "paste-from", OpValue: "upload_date"}, false},
		{"copy-to into all-dates", sharedmodels.MetaOps{Field: "upload_date", OpType: "copy-to", OpValue: "all-dates"}, false},
		{"all-dates suffix rejected", sharedmodels.MetaOps{Field: "all-dates", OpType: "suffix", OpValue: "x"}, true},
		{"all-dates date-tag rejected", sharedmodels.MetaOps{Field: "all-dates", OpType: "date-tag", OpLoc: "prefix", DateFormat: "ymd"}, true},
		{"all-dates ok", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-05-10"}, false},
		{"all-dates unhyphenated", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "20240510"}, true},
		{"all-dates with time", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-05-10T00:00:00Z"}, true},
		{"all-dates impossible day", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-02-30"}, true},
		{"release_date ok", sharedmodels.MetaOps{Field: "release_date", OpType: "set", OpValue: "2022-05-10"}, false},
		{"premiered single-digit month", sharedmodels.MetaOps{Field: "premiered", OpType: "set", OpValue: "2022-5-10"}, true},
		{"year ok", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "2024"}, false},
		{"year two digits", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "24"}, true},
		{"year full date", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "2024-05-10"}, true},
		{"creation_time ok", sharedmodels.MetaOps{Field: "creation_time", OpType: "set", OpValue: "2024-05-10T12:30:00Z"}, false},
		{"creation_time date only", sharedmodels.MetaOps{Field: "creation_time", OpType: "set", OpValue: "2024-05-10"}, true},
		{"date field suffix unchecked", sharedmodels.MetaOps{Field: "date", OpType: "suffix", OpValue: " x"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateMetaOps([]sharedmodels.MetaOps{tt.op})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetaOps(%+v) err = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateMetaOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilenameOps(t *testing.T) {
	tests := []struct {
		name    string
		op      sharedmodels.FilenameOps
		wantErr bool
	}{
		{"prefix", sharedmodels.FilenameOps{OpType: "prefix", OpValue: "[CATS] "}, false},
		{"no field required", sharedmodels.FilenameOps{OpType: "suffix", OpValue: "(new)"}, false},
		{"append removed", sharedmodels.FilenameOps{OpType: "append", OpValue: "(new)"}, true},
		{"unknown type", sharedmodels.FilenameOps{OpType: "explode"}, true},
		{"copy-to is meta only", sharedmodels.FilenameOps{OpType: "copy-to", OpValue: "x"}, true},
		{"date-tag ok", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "suffix", DateFormat: "dm"}, false},
		{"date-tag rejects all", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "all", DateFormat: "ymd"}, true},
		{"delete-date-tag allows all", sharedmodels.FilenameOps{OpType: "delete-date-tag", OpLoc: "all", DateFormat: "ymd"}, false},
		{"regex-replace ok", sharedmodels.FilenameOps{OpType: "regex-replace", OpFindString: "^a", OpValue: "b"}, false},
		{"regex-replace bad pattern", sharedmodels.FilenameOps{OpType: "regex-replace", OpFindString: "[a"}, true},
		{"keep-before ok", sharedmodels.FilenameOps{OpType: "keep-before", OpValue: " |"}, false},
		{"keep-after empty", sharedmodels.FilenameOps{OpType: "keep-after"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilenameOps([]sharedmodels.FilenameOps{tt.op})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilenameOps(%+v) err = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateFilenameOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilterOps(t *testing.T) {
	tests := []struct {
		name    string
		filter  sharedmodels.Filters
		wantErr bool
	}{
		{"contains", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}, false},
		{"omits any", sharedmodels.Filters{Field: "title", FilterType: "omits", Value: "dog", MustAny: "any"}, false},
		{"equals", sharedmodels.Filters{Field: "title", FilterType: "equals", Value: "cat", MustAny: "must"}, false},
		{"more-than numeric", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "3600", MustAny: "must"}, false},
		{"presence contains must", sharedmodels.Filters{Field: "date", FilterType: "contains", MustAny: "must"}, false},
		{"presence contains any", sharedmodels.Filters{Field: "date", FilterType: "contains", MustAny: "any"}, false},
		{"presence omits any", sharedmodels.Filters{Field: "date", FilterType: "omits", MustAny: "any"}, false},
		{"equals empty any", sharedmodels.Filters{Field: "title", FilterType: "equals", MustAny: "any"}, false},
		{"regex-contains", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `(?i)^ep\.? ?\d+`, MustAny: "must"}, false},

		{"unknown type", sharedmodels.Filters{Field: "title", FilterType: "explodes", Value: "x", MustAny: "must"}, true},
		{"more-than non-numeric", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "soon", MustAny: "must"}, true},
		{"bad must/any", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "maybe"}, true},
		{"empty field", sharedmodels.Filters{FilterType: "contains", Value: "cat", MustAny: "must"}, true},
		{"regex-contains empty pattern", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", MustAny: "must"}, true},
		{"regex-contains invalid pattern", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "(", MustAny: "must"}, true},
		{"regex-contains pattern with ]]:", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "[[:digit:]]:", MustAny: "must"}, true},
		{"regex-contains pattern with ||", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "a||b", MustAny: "must"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilterOps([]sharedmodels.Filters{tt.filter})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilterOps(%+v) err = %v, wantErr %v", tt.filter, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateFilterOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilteredMetaOps(t *testing.T) {
	goodFilter := sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}
	goodOp := sharedmodels.MetaOps{Field: "title", OpType: "prefix", OpValue: "[CAT] "}

	tests := []struct {
		name    string
		entry   sharedmodels.FilteredMetaOps
		wantErr bool
	}{
		{"valid", sharedmodels.FilteredMetaOps{Filters: []sharedmodels.Filters{goodFilter}, MetaOps: []sharedmodels.MetaOps{goodOp}}, false},
		{"no filters", sharedmodels.FilteredMetaOps{MetaOps: []sharedmodels.MetaOps{goodOp}}, true},
		{"no operations", sharedmodels.FilteredMetaOps{Filters: []sharedmodels.Filters{goodFilter}}, true},
		{
			name: "bad filter type",
			entry: sharedmodels.FilteredMetaOps{
				Filters: []sharedmodels.Filters{{Field: "title", FilterType: "explodes", Value: "x", MustAny: "must"}},
				MetaOps: []sharedmodels.MetaOps{goodOp},
			},
			wantErr: true,
		},
		{
			name: "bad operation type",
			entry: sharedmodels.FilteredMetaOps{
				Filters: []sharedmodels.Filters{goodFilter},
				MetaOps: []sharedmodels.MetaOps{{Field: "title", OpType: "explodes"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilteredMetaOps([]sharedmodels.FilteredMetaOps{tt.entry})
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilteredFilenameOps(t *testing.T) {
	entry := sharedmodels.FilteredFilenameOps{
		Filters:     []sharedmodels.Filters{{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}},
		FilenameOps: []sharedmodels.FilenameOps{{OpType: "prefix", OpValue: "[CAT] "}},
	}
	if err := sharedvalidation.ValidateFilteredFilenameOps([]sharedmodels.FilteredFilenameOps{entry}); err != nil {
		t.Errorf("valid entry rejected: %v", err)
	}

	entry.FilenameOps = nil
	if err := sharedvalidation.ValidateFilteredFilenameOps([]sharedmodels.FilteredFilenameOps{entry}); err == nil {
		t.Error("an entry with no filename operations should be rejected")
	}
}
