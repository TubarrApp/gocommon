package sharedparsing

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// FilterSeparator divides a filtered entry's filter from its operation.
const FilterSeparator = "||"

// filterFormatWithMustAny describes the download filter form, where several filters
// combine and each states whether it must match or may match.
const filterFormatWithMustAny = "please enter filters in the format 'field:filter_type:value:must_or_any'.\n\n" +
	"'title:omits:frogs:must' ignores all videos with frogs in the metatitle.\n" +
	"'title:contains:cat:any','title:contains:dog:any' only includes videos with EITHER cat and dog in the title (use 'must' to require both).\n" +
	"'date:omits:must' omits videos only when the metafile contains a date field.\n" +
	"'duration:morethan:3600:must' only includes videos with a duration field more than 3600 (one hour in seconds).\n"

// filterFormatPlain describes the filtered-operation form, where each entry holds a
// single filter so must/any would mean nothing.
const filterFormatPlain = "please enter filters in the format 'field:filter_type:value'.\n" +
	"('must'/'any' is not used here, as each entry holds a single filter).\n\n" +
	"'title:contains:cat' applies the operations to videos with cat in the metatitle.\n" +
	"'title:omits:frogs' skips the operations for videos with frogs in the metatitle.\n" +
	"'date:omits' applies the operations only when the metafile has no date field.\n" +
	"'duration:morethan:3600' applies the operations to videos with a duration field more than 3600 (one hour in seconds).\n"

// ParseFilterOps parses filter strings, e.g. "title:omits:frogs:must" or
// "title:contains:cat".
//
// When requireMustAny is false each entry holds one filter, so "must" is implied.
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
		chanURL, rest, _ := SplitOpURL(op)

		// Must/any is always the last part when required.
		mustAny := sharedconsts.OpMust
		if requireMustAny {
			i := lastIndexUnescaped(rest, fieldSeparator)
			if i < 0 {
				return nil, fmt.Errorf("%s", formatErrorMsg)
			}
			mustAny = strings.ToLower(strings.TrimSpace(unescape(rest[i+1:])))
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
			filterType = unescape(afterField)
		}
		filterType = strings.ToLower(strings.TrimSpace(filterType))
		value := strings.ToLower(unescape(rawValue))
		if !hasValue && filterType != sharedconsts.OpContains && filterType != sharedconsts.OpOmits {
			return nil, fmt.Errorf("%s", formatErrorMsg)
		}
		if i := lastIndexUnescaped(rawValue, fieldSeparator); !requireMustAny && i >= 0 {
			if cond := strings.ToLower(strings.TrimSpace(rawValue[i+1:])); cond == sharedconsts.OpMust || cond == sharedconsts.OpAny {
				return nil, fmt.Errorf("filter %q ends in %q, which isn't used here (escape the ':' to keep it in the value)", op, cond)
			}
		}

		filters = append(filters, sharedmodels.Filters{
			Field:      field,
			FilterType: filterType,
			Value:      value,
			MustAny:    mustAny,
			ChannelURL: chanURL,
		})
	}
	return filters, nil
}

// FormatFilter renders a filter as [ParseFilterOps] reads it.
func FormatFilter(f sharedmodels.Filters, withMustAny bool) string {
	s := Escape(f.Field) + string(fieldSeparator) + f.FilterType
	if f.Value != "" {
		v := escapeFilterValue(f.Value)
		if i := strings.LastIndexByte(v, fieldSeparator); !withMustAny && i >= 0 {
			if cond := strings.ToLower(strings.TrimSpace(v[i+1:])); cond == sharedconsts.OpMust || cond == sharedconsts.OpAny {
				v = v[:i] + string(escapeByte) + v[i:]
			}
		}
		s += string(fieldSeparator) + v
	}
	if withMustAny {
		s += string(fieldSeparator) + f.MustAny
	}
	return withURL(s, f.ChannelURL, true)
}

// escapeFilterValue escapes '\', and any '|' that could run into a [FilterSeparator].
func escapeFilterValue(v string) string {
	var buf strings.Builder
	for i := range len(v) {
		if v[i] == escapeByte || (v[i] == urlSeparator && (i == len(v)-1 || v[i+1] == urlSeparator)) {
			buf.WriteByte(escapeByte)
		}
		buf.WriteByte(v[i])
	}
	return buf.String()
}

// splitFiltered divides a filtered entry at its first unescaped [FilterSeparator].
func splitFiltered(rest string) (filter, op string, ok bool) {
	escaped := false
	for i := range len(rest) {
		switch {
		case escaped:
			escaped = false
		case rest[i] == escapeByte:
			escaped = true
		case strings.HasPrefix(rest[i:], FilterSeparator):
			return rest[:i], rest[i+len(FilterSeparator):], true
		}
	}
	return "", "", false
}

// ParseFilteredMetaOps parses filter-gated meta operations, one filter and one operation
// per entry, e.g. "title:contains:cat||director:set:Mr. Cat".
//
// A leading channel URL scopes the whole entry.
func ParseFilteredMetaOps(entries []string) (parsed []sharedmodels.FilteredMetaOps, warnings []Warning, err error) {
	if len(entries) == 0 {
		return nil, nil, nil
	}
	entries, dupes := Deduplicate(entries)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilteredMetaOps, 0, len(entries))
	for _, entry := range entries {
		chanURL, rest, urlErr := SplitOpURL(entry)
		if urlErr != nil {
			warnings = append(warnings, Warning{Kind: WarnChannelURL, Entry: entry, Msg: urlErr.Error()})
		}

		filter, op, ok := splitFiltered(rest)
		if !ok {
			return nil, warnings, fmt.Errorf("invalid format for filtered meta operation %q, use 'Filter Rule||Meta Operation'", entry)
		}

		filters, err := ParseFilterOps([]string{filter}, false)
		if err != nil {
			return nil, warnings, err
		}

		metaOps, opWarnings, err := ParseMetaOps([]string{op})
		warnings = append(warnings, opWarnings...)
		if err != nil {
			return nil, warnings, fmt.Errorf("filtered meta operation %q: %w", entry, err)
		}
		if len(filters) == 0 || len(metaOps) == 0 {
			continue
		}

		scopeToChannel(chanURL, filters, metaOps)
		valid = append(valid, sharedmodels.FilteredMetaOps{Filters: filters, MetaOps: metaOps})
	}
	if err := sharedvalidation.ValidateFilteredMetaOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}

// ParseFilteredFilenameOps parses filter-gated filename operations, one filter and one
// operation per entry, e.g. "title:contains:cat||prefix:[CATS] ".
func ParseFilteredFilenameOps(entries []string) (parsed []sharedmodels.FilteredFilenameOps, warnings []Warning, err error) {
	if len(entries) == 0 {
		return nil, nil, nil
	}
	entries, dupes := Deduplicate(entries)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilteredFilenameOps, 0, len(entries))
	for _, entry := range entries {
		chanURL, rest, urlErr := SplitOpURL(entry)
		if urlErr != nil {
			warnings = append(warnings, Warning{Kind: WarnChannelURL, Entry: entry, Msg: urlErr.Error()})
		}

		filter, op, ok := splitFiltered(rest)
		if !ok {
			return nil, warnings, fmt.Errorf("invalid format for filtered filename operation %q, use 'Filter Rule||Filename Operation'", entry)
		}

		filters, err := ParseFilterOps([]string{filter}, false)
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

		for i := range filters {
			if filters[i].ChannelURL == "" {
				filters[i].ChannelURL = chanURL
			}
		}
		for i := range filenameOps {
			if filenameOps[i].ChannelURL == "" {
				filenameOps[i].ChannelURL = chanURL
			}
		}
		valid = append(valid, sharedmodels.FilteredFilenameOps{Filters: filters, FilenameOps: filenameOps})
	}
	if err := sharedvalidation.ValidateFilteredFilenameOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}

// scopeToChannel points an entry's filters and operations at chanURL, leaving any that
// already name their own alone.
func scopeToChannel(chanURL string, filters []sharedmodels.Filters, ops []sharedmodels.MetaOps) {
	if chanURL == "" {
		return
	}
	for i := range filters {
		if filters[i].ChannelURL == "" {
			filters[i].ChannelURL = chanURL
		}
	}
	for i := range ops {
		if ops[i].ChannelURL == "" {
			ops[i].ChannelURL = chanURL
		}
	}
}
