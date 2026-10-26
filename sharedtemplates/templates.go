// Package sharedtemplates holds templating elements used across Tubarr and Metarr.
package sharedtemplates

import (
	"maps"
	"slices"
	"strings"
)

// init initializes the MoveOpTemplateTags map.
func init() {
	maps.Copy(MoveOpTemplateTags, AllTemplatesMap)
	MoveOpTemplateTags[OutputDirectory] = struct{}{}
}

// Template tags for channel elements.
const (
	ChannelID       = "channel_id"
	ChannelName     = "channel_name"
	ChannelURL      = "channel_url"
	ChannelURLName  = "channel_url_name"
	OutputDirectory = "output_directory"
)

// Template tags for individual video elements.
const (
	MetVideoTitle = "video_title"
	MetVideoURL   = "video_url"
)

// Template tags for date metadata.
const (
	MetDay   = "day"
	MetMonth = "month"
	MetYear  = "year"
)

// Directives choosing a date tag's format, written as {{tag:directive}}.
const (
	YearFull  = "yyyy" // 2026 (default).
	YearShort = "yy"   // 26.

	MonthNumber      = "number"      // 01 (default).
	MonthAbbreviated = "abbreviated" // Jan.
	MonthWord        = "word"        // January.
)

// TagDirectives holds the directives each tag accepts. Tags not listed accept none.
var TagDirectives = map[string]map[string]struct{}{
	MetYear:  {YearFull: {}, YearShort: {}},
	MetMonth: {MonthNumber: {}, MonthAbbreviated: {}, MonthWord: {}},
}

// ParseTag splits a template tag into its lowercased name and directive (e.g. "Month:Word" → "month", "word").
func ParseTag(tag string) (name, directive string) {
	name, directive, _ = strings.Cut(strings.ToLower(strings.TrimSpace(tag)), ":")
	return strings.TrimSpace(name), strings.TrimSpace(directive)
}

// ValidTag reports whether a tag's name is in templateMap and its directive, if any, is one the tag accepts.
func ValidTag(tag string, templateMap map[string]struct{}) bool {
	name, directive := ParseTag(tag)
	if _, ok := templateMap[name]; !ok {
		return false
	}
	if directive == "" {
		return true
	}
	_, ok := TagDirectives[name][directive]
	return ok
}

// DescribeTags lists the tags in templateMap in order, with any directives (e.g. "month[:abbreviated|number|word]").
func DescribeTags(templateMap map[string]struct{}) []string {
	out := make([]string, 0, len(templateMap))
	for _, name := range slices.Sorted(maps.Keys(templateMap)) {
		if d := TagDirectives[name]; len(d) > 0 {
			name += "[:" + strings.Join(slices.Sorted(maps.Keys(d)), "|") + "]"
		}
		out = append(out, name)
	}
	return out
}

// Template tags for cast metadata.
const (
	MetAuthor   = "author"
	MetDirector = "director"
)

// Template tags for URL metadata.
const (
	MetDomain = "domain"
)

// MoveOpTemplateTags contains the regular templates plus the special output_directory tag.
var MoveOpTemplateTags = make(map[string]struct{}, len(AllTemplatesMap)+1)

// AllTemplatesMap contains the different template tags available across both Tubarr and Metarr.
var AllTemplatesMap = map[string]struct{}{
	ChannelID:      {},
	ChannelName:    {},
	ChannelURL:     {},
	ChannelURLName: {},
	MetVideoTitle:  {},
	MetVideoURL:    {},
	MetDay:         {},
	MetMonth:       {},
	MetYear:        {},
	MetAuthor:      {},
	MetDirector:    {},
	MetDomain:      {},
}

// TubarrTemplateTags contains templating tags which are fillable by Tubarr channels.
var TubarrTemplateTags = map[string]struct{}{
	ChannelID:      {},
	ChannelName:    {},
	ChannelURL:     {},
	ChannelURLName: {},
}

// MetarrTemplateTags contains templating tags which are fillable by Metarr.
var MetarrTemplateTags = map[string]struct{}{
	MetVideoTitle: {},
	MetVideoURL:   {},
	MetYear:       {},
	MetMonth:      {},
	MetDay:        {},
	MetAuthor:     {},
	MetDirector:   {},
	MetDomain:     {},
}

// NoTemplateTags enforces a dis-allowance of template tags in a given filepath.
var NoTemplateTags = map[string]struct{}{}
