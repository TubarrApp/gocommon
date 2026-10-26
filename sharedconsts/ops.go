package sharedconsts

// ValidFilterTypes are the comparisons accepted by a filter.
var ValidFilterTypes = map[string]struct{}{
	OpContains:      {},
	OpRegexContains: {},
	OpOmits:         {},
	OpEquals:        {},
	OpNotEquals:     {},
	OpMoreThan:      {},
	OpLessThan:      {},
}

// Operations for various features.
const (
	// Contains or omits.
	OpContains      = "contains"
	OpRegexContains = "regex-contains"
	OpOmits         = "omits"
	OpEquals        = "equals"
	OpNotEquals     = "not-equals"
	OpMoreThan      = "more-than"
	OpLessThan      = "less-than"

	// Musts and anys.
	OpMust = "must"
	OpAny  = "any"

	// Copy/paste.
	OpCopyTo    = "copy-to"
	OpPasteFrom = "paste-from"

	// Date tag operations.
	OpDateTag       = "date-tag"
	OpDeleteDateTag = "delete-date-tag"

	// Locations.
	OpLocAll    = "all"
	OpLocPrefix = "prefix"
	OpLocSuffix = "suffix"

	// Replacement operations.
	OpReplace       = "replace"
	OpReplaceSuffix = "replace-suffix"
	OpReplacePrefix = "replace-prefix"
	OpRegexReplace  = "regex-replace"

	// Keep the text on one side of a delimiter.
	OpKeepAfter  = "keep-after"
	OpKeepBefore = "keep-before"

	// Prefix and suffix.
	OpPrefix = "prefix"
	OpSuffix = "suffix"

	// Set.
	OpSet = "set"
)

// MetaOpAllFieldsOfType defines changes which replace all tags of
// the given type (e.g. credits, or date).
var MetaOpAllFieldsOfType = map[string]struct{}{
	MetaOpAllCredits:      {},
	MetaOpAllDates:        {},
	MetaOpAllDescriptions: {},
	MetaOpAllTitles:       {},
}

// Meta type operations.
const (
	MetaOpAllCredits      = "all-credits"
	MetaOpAllDates        = "all-dates"
	MetaOpAllDescriptions = "all-descriptions"
	MetaOpAllTitles       = "all-titles"
)
