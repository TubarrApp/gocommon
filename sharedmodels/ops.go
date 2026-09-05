// Package sharedmodels holds the operation models shared by Metarr and Tubarr.
//
// These are transport types: Tubarr parses them from its CLI, web and database
// layers and formats them back into Metarr's argv, which Metarr parses into its own
// typed dispatch models. JSON tags match Tubarr's stored representation.
package sharedmodels

// FilenameOps is a single filename operation applied by Metarr.
type FilenameOps struct {
	ChannelURL   string `json:"filename_op_channel_url"`
	OpType       string `json:"filename_op_type"`
	OpFindString string `json:"filename_op_find_string"`
	OpValue      string `json:"filename_op_value"`
	OpLoc        string `json:"filename_op_loc"`
	DateFormat   string `json:"filename_op_date_format"`
}

// MetaOps is a single metadata field operation applied by Metarr.
type MetaOps struct {
	ChannelURL   string `json:"meta_op_channel_url"`
	Field        string `json:"meta_op_field"`
	OpFindString string `json:"meta_op_find_string"`
	OpType       string `json:"meta_op_type"`
	OpValue      string `json:"meta_op_value"`
	OpLoc        string `json:"meta_op_loc"`
	DateFormat   string `json:"meta_op_date_format"`
}

// Filters select videos, or the operations applied to them, by metadata field.
type Filters struct {
	ChannelURL string `json:"filter_url_specific"`
	Field      string `json:"filter_field"`
	FilterType string `json:"filter_type"`
	Value      string `json:"filter_value"`
	MustAny    string `json:"filter_must_any"`
}

// FilteredMetaOps applies meta operations only when its filters match.
type FilteredMetaOps struct {
	Filters        []Filters
	MetaOps        []MetaOps
	FiltersMatched bool
}

// FilteredFilenameOps applies filename operations only when its filters match.
type FilteredFilenameOps struct {
	Filters        []Filters
	FilenameOps    []FilenameOps
	FiltersMatched bool
}
