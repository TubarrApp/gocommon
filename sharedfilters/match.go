// Package sharedfilters evaluates metadata filters for Metarr and Tubarr.
package sharedfilters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

// Match reports whether meta satisfies filter.
//
// A comma-separated field list, e.g. "title,fulltitle", matches when any of its fields does, except
// omits and not-equals, which need all of them to: "title,fulltitle:omits:cat" means neither field
// holds cat. A numeric comparison is skipped only when no listed field has anything to compare.
//
// Contains or omits filter with no value tests only whether the field is present.
//
// Equals and not-equals with no value compare the field against "" (a missing field counts
// as ""). A numeric comparison against a missing or empty field has nothing to compare, so
// it reports skipped rather than a result. A numeric field that is neither a number nor a duration
// doesn't match.
//
// Comparisons ignore case, except a regex-contains pattern, which matches the field as written
// unless the pattern itself sets "(?i)".
func Match(meta map[string]any, filter sharedmodels.Filters) (matched, skipped bool, metaText string, err error) {
	needsAll := filter.FilterType == sharedconsts.OpOmits || filter.FilterType == sharedconsts.OpNotEquals

	allSkipped := true
	for _, field := range sharedparsing.SplitMultiField(filter.Field) {
		// Get field value.
		raw, metaExists := meta[field]
		metaText = fmt.Sprint(raw)
		if !metaExists {
			metaText = ""
		}

		f := filter
		f.Field = field
		ok, fieldSkipped, err := matchField(metaText, metaExists, f)
		if err != nil {
			return false, false, metaText, err
		}
		if fieldSkipped {
			continue
		}
		allSkipped = false

		// One field settles it: a match for most types, a mismatch for omits and not-equals.
		if ok != needsAll {
			return ok, false, metaText, nil
		}
	}
	if allSkipped {
		return false, true, metaText, nil
	}
	return needsAll, false, metaText, nil
}

// matchField is [Match] for a filter on a single field.
func matchField(metaText string, metaExists bool, filter sharedmodels.Filters) (matched, skipped bool, err error) {
	value := strings.ToLower(metaText)
	want := strings.ToLower(filter.Value)

	if isNumeric(filter.FilterType) {
		if !metaExists || value == "" || value == "<nil>" {
			return false, true, nil
		}
		return matchNumeric(filter.FilterType, value, want)
	}

	// Presence-only form (e.g. 'title:must' means the title field must exist).
	if filter.Value == "" {
		switch filter.FilterType {
		case sharedconsts.OpContains:
			return metaExists, false, nil
		case sharedconsts.OpOmits:
			return !metaExists, false, nil
		}
	}

	// Filter has a value. Perform checks against value:
	switch filter.FilterType {
	case sharedconsts.OpContains:
		return strings.Contains(value, want), false, nil

	case sharedconsts.OpRegexContains:
		re, err := regexp.Compile(filter.Value)
		if err != nil {
			return false, false, fmt.Errorf("invalid pattern %q for field %q: %w", filter.Value, filter.Field, err)
		}
		return re.MatchString(metaText), false, nil

	case sharedconsts.OpOmits:
		return !strings.Contains(value, want), false, nil

	case sharedconsts.OpEquals:
		return value == want, false, nil

	case sharedconsts.OpNotEquals:
		return value != want, false, nil

	default:
		return false, false, fmt.Errorf("unknown filter type %q for field %q", filter.FilterType, filter.Field)
	}
}

// MatchAll reports whether meta satisfies every filter, which is the rule for a
// filtered operation: one entry's filters all have to hold for its operations to apply.
//
// Filters that report skipped are treated as not matching, since a filter the user
// wrote should not be assumed to pass.
func MatchAll(meta map[string]any, filters []sharedmodels.Filters) (matched bool, err error) {
	for _, f := range filters {
		ok, skipped, _, err := Match(meta, f)
		if err != nil {
			return false, err
		}
		if skipped || !ok {
			return false, nil
		}
	}
	return true, nil
}

// isNumeric reports whether the filter type compares numbers.
func isNumeric(filterType string) bool {
	return filterType == sharedconsts.OpMoreThan || filterType == sharedconsts.OpLessThan
}

// matchNumeric compares a field's number, or duration in seconds, with the filter's number.
func matchNumeric(filterType, value, want string) (matched, skipped bool, err error) {
	got, gotErr := strconv.ParseFloat(value, 64)
	if gotErr != nil {
		var secs int
		if secs, gotErr = sharedparsing.ParseDuration(value); gotErr == nil {
			got = float64(secs)
		}
	}
	target, targetErr := strconv.ParseFloat(want, 64)
	if gotErr != nil || targetErr != nil {
		return false, false, nil //nolint:nilerr // A value that isn't a number doesn't match, rather than failing the filter.
	}

	switch filterType {
	case sharedconsts.OpMoreThan:
		return got > target, false, nil
	case sharedconsts.OpLessThan:
		return got < target, false, nil
	default:
		return false, false, fmt.Errorf("filter type %q is not a numeric comparison", filterType)
	}
}
