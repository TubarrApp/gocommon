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
