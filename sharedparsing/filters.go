package sharedparsing

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// filterFormatWithMustAny describes the download filter form, where several filters
// combine and each states whether it must match or may match.
const filterFormatWithMustAny = "please enter filters in the format 'field:filter_type:value:must_or_any'.\n\n" +
	"'title:omits:frogs:must' ignores all videos with frogs in the metatitle.\n" +
	"'title:contains:cat:any','title:contains:dog:any' only includes videos with EITHER cat and dog in the title (use 'must' to require both).\n" +
	"'date:omits:must' omits videos only when the metafile contains a date field.\n" +
	"'duration:more-than:3600:must' only includes videos with a duration field more than 3600 (one hour in seconds).\n"

// filterFormatPlain describes the filtered-operation form, where every filter in an entry
// must match so must/any would mean nothing.
const filterFormatPlain = "please enter filters in the format 'field:filter_type:value', joining several with '&&' to require all of them.\n" +
	"('must'/'any' is not used here, as every filter must match).\n\n" +
	"'title:contains:cat' applies the operations to videos with cat in the metatitle.\n" +
	"'title:omits:frogs' skips the operations for videos with frogs in the metatitle.\n" +
	"'date:omits' applies the operations only when the metafile has no date field.\n" +
	"'duration:more-than:3600' applies the operations to videos with a duration field more than 3600 (one hour in seconds).\n" +
	"'title:contains:cat&&duration:more-than:3600' applies the operations to videos with cat in the metatitle that are over an hour long.\n"

// ParseFilterOps parses filter strings, e.g. "title:omits:frogs:must" or
// "title:contains:cat".
//
// When requireMustAny is false "must" is implied, as every filter has to match.
func ParseFilterOps(ops []string, requireMustAny bool) ([]sharedmodels.Filters, error) {
	if len(ops) == 0 {
		return nil, nil
	}
	ops, _ = Deduplicate(ops)

	formatErrorMsg := filterFormatPlain
	if requireMustAny {
		formatErrorMsg = filterFormatWithMustAny
	}

	filters := make([]sharedmodels.Filters, 0, len(ops))
	for _, op := range ops {
		rest := op

		// Must/any is always the last part when required.
		mustAny := sharedconsts.OpMust
		if requireMustAny {
			i := lastIndexUnescaped(rest, fieldSeparator)
			if i < 0 {
				return nil, fmt.Errorf("%s", formatErrorMsg)
			}
			mustAny = strings.ToLower(strings.TrimSpace(Unescape(rest[i+1:])))
			if mustAny != sharedconsts.OpMust && mustAny != sharedconsts.OpAny {
				return nil, fmt.Errorf("%s:\nInvalid must/any value %q", formatErrorMsg, mustAny)
			}
			rest = rest[:i]
		}

		field, afterField, ok := nextField(rest)
		if !ok {
			return nil, fmt.Errorf("%s", formatErrorMsg)
		}
		field = strings.ToLower(strings.TrimSpace(field))

		// The valueless form only acts on a field's presence, which only contains and
		// omits can make use of.
		filterType, rawValue, hasValue := nextField(afterField)
		if !hasValue {
			filterType = Unescape(afterField)
		}
		filterType = strings.ToLower(strings.TrimSpace(filterType))
		value := strings.ToLower(Unescape(rawValue))
		if !hasValue && filterType != sharedconsts.OpContains && filterType != sharedconsts.OpOmits {
			return nil, fmt.Errorf("%s", formatErrorMsg)
		}

		// Guard against common user mistake: Many filter ops don't use ':must' or ':any'
		if i := lastIndexUnescaped(rawValue, fieldSeparator); !requireMustAny && i >= 0 {
			// Grab final portion after last separator (e.g. ':must')
			cond := strings.ToLower(strings.TrimSpace(rawValue[i+1:]))
			if cond == sharedconsts.OpMust || cond == sharedconsts.OpAny {
				// Prints a line like: 'filter 'title:contains:Re: Zero:any' ends in 'any', which isn't used here. Use 'title:contains:Re: Zero', or 'title:contains:Re: Zero\:any' to match 'Re: Zero:any' literally'
				withoutCond := op[:len(op)-len(rawValue)+i]
				return nil, fmt.Errorf("filter '%s' ends in '%s', which isn't used here. Use '%s', or '%s' to match '%s' literally", op, cond, withoutCond, withoutCond+`\`+op[len(withoutCond):], Unescape(rawValue))
			}
		}

		// A regex-contains pattern is wrapped in "[[ ]]" and taken verbatim, keeping its case.
		if filterType == sharedconsts.OpRegexContains {
			end := rawSpanEnd(rawValue)
			if end != len(rawValue) {
				return nil, fmt.Errorf("filter '%s': %s expects a pattern wrapped in %s %s, e.g. 'title:%s:%s^cats? %s'", op, filterType, openSquareBrackets, closeSquareBrackets, filterType, openSquareBrackets, closeSquareBrackets)
			}
			value = rawValue[len(openSquareBrackets) : end-len(closeSquareBrackets)]
		}

		filters = append(filters, sharedmodels.Filters{
			Field:      field,
			FilterType: filterType,
			Value:      value,
			MustAny:    mustAny,
		})
	}
	return filters, nil
}

// FormatFilter renders a filter as [ParseFilterOps] reads it.
func FormatFilter(f sharedmodels.Filters, withMustAny bool) string {
	s := Escape(f.Field) + string(fieldSeparator) + f.FilterType
	switch {
	case f.FilterType == sharedconsts.OpRegexContains:
		s += string(fieldSeparator) + openSquareBrackets + f.Value + closeSquareBrackets
	case f.Value != "":
		v := escapeFilterValue(f.Value)
		if i := strings.LastIndexByte(v, fieldSeparator); !withMustAny && i >= 0 {
			if cond := strings.ToLower(strings.TrimSpace(v[i+1:])); cond == sharedconsts.OpMust || cond == sharedconsts.OpAny {
				v = v[:i] + string(escapeByte) + v[i:]
			}
		}
		s += string(fieldSeparator) + v
	case f.FilterType != sharedconsts.OpContains && f.FilterType != sharedconsts.OpOmits:
		s += string(fieldSeparator) // An empty value to compare against, rather than the presence-only form.
	}
	if withMustAny {
		s += string(fieldSeparator) + f.MustAny
	}
	return s
}

// FormatFilters renders a filtered entry's filters, the inverse of [SplitFilters] and [ParseFilterOps].
func FormatFilters(filters []sharedmodels.Filters) string {
	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		parts = append(parts, FormatFilter(f, false))
	}
	return strings.Join(parts, sharedconsts.FilterJoiner)
}

// escapeFilterValue escapes '\', and any '|' or '&' that could run into a [sharedcosnts.FilterSeparator] or [sharedconsts.FilterJoiner].
func escapeFilterValue(v string) string {
	var buf strings.Builder
	for i := range len(v) {
		// Escape every '\', and a '|' or '&' that's doubled or the last character.
		if v[i] == escapeByte ||
			((v[i] == pipe || v[i] == amp) && (i == len(v)-1 || v[i+1] == v[i])) {
			buf.WriteByte(escapeByte)
		}
		buf.WriteByte(v[i])
	}
	return buf.String()
}

// SplitMultiField divides a field list at each comma, e.g. "title, description" -> ["title", "description"].
func SplitMultiField(field string) []string {
	fields := strings.Split(field, sharedconsts.FieldSeparator)
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}

// SplitFiltered divides a filtered entry at its first unescaped [sharedconsts.FilterSeparator], keeping escapes.
func SplitFiltered(rest string) (filter, op string, ok bool) {
	escaped := false
	for i := range len(rest) {
		switch {
		case escaped:
			escaped = false
		case rest[i] == escapeByte:
			escaped = true
		case strings.HasPrefix(rest[i:], sharedconsts.FilterSeparator):
			return rest[:i], rest[i+len(sharedconsts.FilterSeparator):], true
		}
	}
	return "", "", false
}

// SplitFilters divides a filtered entry's filters at each unescaped [sharedconsts.FilterJoiner], leaving escapes for [ParseFilterOps].
func SplitFilters(s string) []string {
	var parts []string
	start := 0
	escaped := false
	for i := 0; i < len(s); i++ {
		switch {
		case escaped:
			escaped = false
		case s[i] == escapeByte:
			escaped = true
		case strings.HasPrefix(s[i:], sharedconsts.FilterJoiner):
			parts = append(parts, s[start:i])
			i += len(sharedconsts.FilterJoiner) - 1
			start = i + 1
		}
	}
	return append(parts, s[start:])
}

// ParseFilteredMetaOps parses filter-gated meta operations, one or more filters and one
// operation per entry, e.g. "title:contains:cat&&duration:more-than:600||director:set:Mr. Cat".
func ParseFilteredMetaOps(entries []string) (parsed []sharedmodels.FilteredMetaOps, warnings []Warning, err error) {
	if len(entries) == 0 {
		return nil, nil, nil
	}
	entries, dupes := Deduplicate(entries)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilteredMetaOps, 0, len(entries))

	// Iterate entries.
	for _, entry := range entries {
		filter, op, ok := SplitFiltered(entry)
		if !ok {
			return nil, warnings, fmt.Errorf("invalid format for filtered meta operation %q, use 'Filter Rule||Meta Operation'", entry)
		}

		filters, err := ParseFilterOps(SplitFilters(filter), false)
		if err != nil {
			return nil, warnings, err
		}

		// A field list in the operation gives one operation per field, see [ParseMetaOps].
		metaOps, opWarnings, err := ParseMetaOps([]string{op})
		warnings = append(warnings, opWarnings...)
		if err != nil {
			return nil, warnings, fmt.Errorf("filtered meta operation %q: %w", entry, err)
		}
		if len(filters) == 0 || len(metaOps) == 0 {
			continue
		}

		valid = append(valid, sharedmodels.FilteredMetaOps{Filters: filters, MetaOps: metaOps})
	}
	if err := sharedvalidation.ValidateFilteredMetaOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}

// ParseFilteredFilenameOps parses filter-gated filename operations, one or more filters and
// one operation per entry, e.g. "title:contains:cat&&duration:more-than:600||prefix:[CATS] ".
func ParseFilteredFilenameOps(entries []string) (parsed []sharedmodels.FilteredFilenameOps, warnings []Warning, err error) {
	if len(entries) == 0 {
		return nil, nil, nil
	}
	entries, dupes := Deduplicate(entries)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilteredFilenameOps, 0, len(entries))
	for _, entry := range entries {
		filter, op, ok := SplitFiltered(entry)
		if !ok {
			return nil, warnings, fmt.Errorf("invalid format for filtered filename operation %q, use 'Filter Rule||Filename Operation'", entry)
		}

		filters, err := ParseFilterOps(SplitFilters(filter), false)
		if err != nil {
			return nil, warnings, err
		}

		filenameOps, opWarnings, err := ParseFilenameOps([]string{op})
		warnings = append(warnings, opWarnings...)
		if err != nil {
			return nil, warnings, fmt.Errorf("filtered filename operation %q: %w", entry, err)
		}
		if len(filters) == 0 || len(filenameOps) == 0 {
			continue
		}

		valid = append(valid, sharedmodels.FilteredFilenameOps{Filters: filters, FilenameOps: filenameOps})
	}
	if err := sharedvalidation.ValidateFilteredFilenameOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}
