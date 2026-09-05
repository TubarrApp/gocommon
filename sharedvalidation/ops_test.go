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
		{"no field required", sharedmodels.FilenameOps{OpType: "append", OpValue: "(new)"}, false},
		{"unknown type", sharedmodels.FilenameOps{OpType: "explode"}, true},
		{"copy-to is meta only", sharedmodels.FilenameOps{OpType: "copy-to", OpValue: "x"}, true},
		{"date-tag ok", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "suffix", DateFormat: "dm"}, false},
		{"date-tag rejects all", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "all", DateFormat: "ymd"}, true},
		{"delete-date-tag allows all", sharedmodels.FilenameOps{OpType: "delete-date-tag", OpLoc: "all", DateFormat: "ymd"}, false},
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
		{"morethan numeric", sharedmodels.Filters{Field: "duration", FilterType: "morethan", Value: "3600", MustAny: "must"}, false},

		{"unknown type", sharedmodels.Filters{Field: "title", FilterType: "explodes", Value: "x", MustAny: "must"}, true},
		{"morethan non-numeric", sharedmodels.Filters{Field: "duration", FilterType: "morethan", Value: "soon", MustAny: "must"}, true},
		{"bad must/any", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "maybe"}, true},
		{"empty field", sharedmodels.Filters{FilterType: "contains", Value: "cat", MustAny: "must"}, true},
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
