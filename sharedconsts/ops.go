package sharedconsts

// Operations for various features.
const (
	// Contains or omits.
	OpContains  = "contains"
	OpOmits     = "omits"
	OpEquals    = "equals"
	OpNotEquals = "notequals"
	OpMoreThan  = "morethan"
	OpLessThan  = "lessthan"

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

	// Append and suffix.
	OpAppend = "append"
	OpPrefix = "prefix"

	// Set.
	OpSet = "set"
)
