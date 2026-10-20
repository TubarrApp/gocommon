// Package sharedtemplates holds templating elements used across Tubarr and Metarr.
package sharedtemplates

import (
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxNameBytes is the longest file or directory name most filesystems allow.
const MaxNameBytes = 255

// pathSeparatorReplacer replaces characters that would split a name into directories.
var pathSeparatorReplacer = strings.NewReplacer("/", "-", "\\", "-", "\x00", "")

// StripURLScheme drops a URL's leading "scheme://" (e.g. "https://youtube.com/@cats" → "youtube.com/@cats").
func StripURLScheme(u string) string {
	scheme, rest, ok := strings.Cut(u, "://")
	if !ok || !isURLScheme(scheme) {
		return u
	}
	return rest
}

// isURLScheme reports whether s can be a URL's scheme (the "https" in "https://"): letters, digits, "+", "-" and ".".
func isURLScheme(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		isLetter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != '+' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

// SanitizePathComponent makes s safe as a single file or directory name, cut to at most maxBytes. Control characters
// become spaces, as a name holding a line break splits paths read a line at a time.
func SanitizePathComponent(s string, maxBytes int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, pathSeparatorReplacer.Replace(s))
	s = strings.TrimSpace(s)
	if s == "." || s == ".." {
		s = strings.ReplaceAll(s, ".", "-")
	}
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strings.TrimSpace(s[:cut])
}

// Template tags for channel elements.
const (
	ChannelID   = "channel_id"
	ChannelName = "channel_name"
	ChannelURL  = "channel_url"
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

// AllTemplatesMap contains the different template tags available across both Tubarr and Metarr.
var AllTemplatesMap = map[string]struct{}{
	ChannelID:     {},
	ChannelName:   {},
	ChannelURL:    {},
	MetVideoTitle: {},
	MetVideoURL:   {},
	MetDay:        {},
	MetMonth:      {},
	MetYear:       {},
	MetAuthor:     {},
	MetDirector:   {},
	MetDomain:     {},
}

// TubarrTemplateTags contains templating tags which are fillable by Tubarr channels.
var TubarrTemplateTags = map[string]struct{}{
	ChannelID:   {},
	ChannelName: {},
	ChannelURL:  {},
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
