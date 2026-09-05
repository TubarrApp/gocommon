// Package sharedfilters evaluates metadata filters for Metarr and Tubarr.
package sharedfilters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// Match reports whether meta satisfies filter.
//
// A filter with no value tests only whether the field is present, which is why only
// contains and omits accept that form. A numeric comparison against a missing or
// non-numeric field has nothing to compare, so it reports skipped rather than a result.
func Match(meta map[string]any, filter sharedmodels.Filters) (matched, skipped bool, err error) {
	raw, exists := meta[filter.Field]
	value := strings.ToLower(fmt.Sprint(raw))
	if !exists {
		value = ""
	}
	want := strings.ToLower(filter.Value)

	if isNumeric(filter.FilterType) {
		if !exists || value == "" || value == "<nil>" {
			return false, true, nil
		}
		return matchNumeric(filter.FilterType, value, want)
	}

	// Presence-only form.
	if filter.Value == "" {
		switch filter.FilterType {
		case sharedconsts.OpContains:
			return exists, false, nil
		case sharedconsts.OpOmits:
			return !exists, false, nil
		default:
			return false, false, fmt.Errorf("filter type %q needs a value", filter.FilterType)
		}
	}

	switch filter.FilterType {
	case sharedconsts.OpContains:
		return strings.Contains(value, want), false, nil
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
		ok, skipped, err := Match(meta, f)
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

// matchNumeric compares two numeric strings, reporting skipped when either will not
// parse as a number.
func matchNumeric(filterType, value, want string) (matched, skipped bool, err error) {
	got, gotErr := strconv.ParseFloat(value, 64)
	target, targetErr := strconv.ParseFloat(want, 64)
	if gotErr != nil || targetErr != nil {
		return false, true, nil
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
