package sharedparsing

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// filterFormatWithMustAny describes the download filter form, where several filters
// combine and each states whether it must match or may match.
const filterFormatWithMustAny = "please enter filters in the format 'field:filter_type:value:must_or_any'.\n\n" +
	"'title:omits:frogs:must' ignores all videos with frogs in the metatitle.\n" +
	"'title:contains:cat:any','title:contains:dog:any' only includes videos with EITHER cat and dog in the title (use 'must' to require both).\n" +
	"'date:omits:must' omits videos only when the metafile contains a date field." +
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

	// The shortest form omits the value and the longest includes it; a must/any
	// condition adds one part to each.
	minParts, maxParts := 2, 3
	formatErrorMsg := filterFormatPlain
	if requireMustAny {
		minParts, maxParts = 3, 4
		formatErrorMsg = filterFormatWithMustAny
	}

	filters := make([]sharedmodels.Filters, 0, len(ops))
	for _, op := range ops {
		chanURL, rest, _ := SplitOpURL(op)
		split := EscapedSplit(rest, fieldSeparator)

		if len(split) < minParts || len(split) > maxParts {
			return nil, fmt.Errorf("%s", formatErrorMsg)
		}

		// Declared per entry so no part carries over from the last filter.
		field := strings.ToLower(strings.TrimSpace(split[0]))
		filterType := strings.ToLower(strings.TrimSpace(split[1]))
		mustAny := sharedconsts.OpMust
		var value string

		// Must/any is always the last part when required.
		if requireMustAny {
			mustAny = strings.ToLower(strings.TrimSpace(split[len(split)-1]))
			if mustAny != sharedconsts.OpMust && mustAny != sharedconsts.OpAny {
				return nil, fmt.Errorf("%s:\nInvalid must/any value %q", formatErrorMsg, mustAny)
			}
		}

		// A value is only present in the longest form, and is always the third part.
		if len(split) == maxParts {
			value = strings.ToLower(split[2])
		}

		// The valueless form only acts on a field's presence, which only contains and
		// omits can make use of.
		if len(split) == minParts &&
			filterType != sharedconsts.OpContains && filterType != sharedconsts.OpOmits {
			return nil, fmt.Errorf("%s", formatErrorMsg)
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

// ParseFilteredMetaOps parses filter-gated meta operations, e.g.
// "title:contains:cat|director:set:Mr. Cat".
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

		split := SplitUnescaped(rest, urlSeparator)
		if len(split) < 2 {
			return nil, warnings, fmt.Errorf("invalid format for filtered meta operation %q, use 'Filter Rules|Meta Operations'", entry)
		}

		filters, err := ParseFilterOps(split[:1], false)
		if err != nil {
			return nil, warnings, err
		}

		metaOps, opWarnings, err := ParseMetaOps(split[1:])
		warnings = append(warnings, opWarnings...)
		if err != nil {
			return nil, warnings, err
		}
		if len(filters) == 0 || len(metaOps) == 0 {
			continue
		}

		scopeToChannel(chanURL, filters, metaOps)
		valid = append(valid, sharedmodels.FilteredMetaOps{Filters: filters, MetaOps: metaOps})
	}
	return valid, warnings, nil
}

// ParseFilteredFilenameOps parses filter-gated filename operations, e.g.
// "title:contains:cat|prefix:[CATS] ".
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

		split := SplitUnescaped(rest, urlSeparator)
		if len(split) < 2 {
			return nil, warnings, fmt.Errorf("invalid format for filtered filename operation %q, use 'Filter Rules|Filename Operations'", entry)
		}

		filters, err := ParseFilterOps(split[:1], false)
		if err != nil {
			return nil, warnings, err
		}

		filenameOps, opWarnings, err := ParseFilenameOps(split[1:])
		warnings = append(warnings, opWarnings...)
		if err != nil {
			return nil, warnings, err
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
