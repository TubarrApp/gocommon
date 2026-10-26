package sharedvalidation

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedenums"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedtags"
)

// sortedKeys returns the keys of m in a stable order, for use in error messages.
func sortedKeys(m map[string]struct{}) []string {
	return slices.Sorted(maps.Keys(m))
}

// validFilenameOpTypes are the operation types accepted for filename operations.
var validFilenameOpTypes = map[string]struct{}{
	sharedconsts.OpPrefix:        {},
	sharedconsts.OpSuffix:        {},
	sharedconsts.OpReplacePrefix: {},
	sharedconsts.OpReplaceSuffix: {},
	sharedconsts.OpReplace:       {},
	sharedconsts.OpRegexReplace:  {},
	sharedconsts.OpKeepAfter:     {},
	sharedconsts.OpKeepBefore:    {},
	sharedconsts.OpDateTag:       {},
	sharedconsts.OpDeleteDateTag: {},
	sharedconsts.OpSet:           {},
}

// validMetaOpTypes are the operation types accepted for meta operations.
var validMetaOpTypes = map[string]struct{}{
	sharedconsts.OpCopyTo:        {},
	sharedconsts.OpPasteFrom:     {},
	sharedconsts.OpPrefix:        {},
	sharedconsts.OpSuffix:        {},
	sharedconsts.OpReplacePrefix: {},
	sharedconsts.OpReplaceSuffix: {},
	sharedconsts.OpReplace:       {},
	sharedconsts.OpRegexReplace:  {},
	sharedconsts.OpKeepAfter:     {},
	sharedconsts.OpKeepBefore:    {},
	sharedconsts.OpSet:           {},
	sharedconsts.OpDateTag:       {},
	sharedconsts.OpDeleteDateTag: {},
}

// ValidFilenameOpTypes lists the accepted filename operation types.
var ValidFilenameOpTypes = sortedKeys(validFilenameOpTypes)

// ValidMetaOpTypes lists the accepted meta operation types.
var ValidMetaOpTypes = sortedKeys(validMetaOpTypes)

// ValidateFilenameOps validates filename transformation operation models.
func ValidateFilenameOps(filenameOps []sharedmodels.FilenameOps) error {
	for i, op := range filenameOps {
		if _, ok := validFilenameOpTypes[op.OpType]; !ok {
			return fmt.Errorf("invalid filename operation type %q at position %d (valid: %v)", op.OpType, i, ValidFilenameOpTypes)
		}
		if err := validateDateTagOp(op.OpType, op.OpLoc, op.DateFormat, i); err != nil {
			return err
		}
		if err := validateTextOp(op.OpType, op.OpFindString, op.OpValue, i); err != nil {
			return err
		}
	}
	return nil
}

// ValidateMetaOps validates meta transformation operation models.
func ValidateMetaOps(metaOps []sharedmodels.MetaOps) error {
	for i, op := range metaOps {
		if _, ok := validMetaOpTypes[op.OpType]; !ok {
			return fmt.Errorf("invalid meta operation type %q at position %d (valid: %v)", op.OpType, i, ValidMetaOpTypes)
		}
		if err := validateDateTagOp(op.OpType, op.OpLoc, op.DateFormat, i); err != nil {
			return err
		}
		if err := validateTextOp(op.OpType, op.OpFindString, op.OpValue, i); err != nil {
			return err
		}
		if op.Field == "" {
			return fmt.Errorf("meta operation at position %d has empty field", i)
		}
		if err := validateGroupOp(op, i); err != nil {
			return err
		}
		if err := validateDateSet(op.OpType, op.Field, op.OpValue, i); err != nil {
			return err
		}
	}
	return nil
}

// validateGroupOp checks an operation involving an "all-x" group. A group cannot be copied or
// pasted from since its fields may differ, and all-dates only accepts set and paste-from.
func validateGroupOp(op sharedmodels.MetaOps, i int) error {
	_, fieldIsGroup := sharedconsts.MetaOpAllFieldsOfType[op.Field]
	_, valueIsGroup := sharedconsts.MetaOpAllFieldsOfType[op.OpValue]

	switch {
	case fieldIsGroup && op.OpType == sharedconsts.OpCopyTo:
		return fmt.Errorf("%s at position %d cannot copy from %q as its fields may differ, copy from a single field instead", op.OpType, i, op.Field)
	case valueIsGroup && op.OpType == sharedconsts.OpPasteFrom:
		return fmt.Errorf("%s at position %d cannot paste from %q as its fields may differ, paste from a single field instead", op.OpType, i, op.OpValue)
	case op.Field == sharedconsts.MetaOpAllDates && op.OpType != sharedconsts.OpSet && op.OpType != sharedconsts.OpPasteFrom:
		return fmt.Errorf("field %q at position %d only supports %q and %q, got %q", op.Field, i, sharedconsts.OpSet, sharedconsts.OpPasteFrom, op.OpType)
	}
	return nil
}

const yearOnly = "2006"

// dateFieldLayouts maps known date fields to the layout a set value on them must match.
var dateFieldLayouts = map[string]string{
	sharedconsts.MetaOpAllDates:     time.DateOnly,
	sharedtags.JDate:                time.DateOnly,
	sharedtags.JOriginallyAvailable: time.DateOnly,
	sharedtags.JReleaseDate:         time.DateOnly,
	sharedtags.JUploadDate:          time.DateOnly,
	sharedtags.NAired:               time.DateOnly,
	sharedtags.NPremiereDate:        time.DateOnly,
	sharedtags.NReleaseDate:         time.DateOnly,
	sharedtags.NReleased:            time.DateOnly,
	sharedtags.JYear:                yearOnly,
	sharedtags.JReleaseYear:         yearOnly,
	sharedtags.JCreationTime:        time.RFC3339,
}

// validateDateSet checks the value of a set on a known date field, and is a no-op otherwise.
func validateDateSet(opType, field, value string, i int) error {
	layout, isDate := dateFieldLayouts[field]
	if opType != sharedconsts.OpSet || !isDate {
		return nil
	}
	if _, err := time.Parse(layout, value); err != nil {
		return fmt.Errorf("set on %q at position %d: %q is not a valid date, must be formatted like %q (YYYY-MM-DD).", field, i, value, layout)
	}
	return nil
}

// validateTextOp checks the pattern of a regex-replace and the delimiter of a keep-before or keep-after,
// and is a no-op for every other operation type.
func validateTextOp(opType, find, value string, i int) error {
	switch opType {
	case sharedconsts.OpRegexReplace:
		if find == "" {
			return fmt.Errorf("%s at position %d has an empty pattern", opType, i)
		}
		if strings.Contains(find, "]]:") {
			return fmt.Errorf("%s at position %d: pattern cannot contain \"]]:\" (try [0-9] in place of [[:digit:]])", opType, i)
		}
		if _, err := regexp.Compile(find); err != nil {
			return fmt.Errorf("%s at position %d: invalid pattern: %w", opType, i, err)
		}
	case sharedconsts.OpKeepAfter, sharedconsts.OpKeepBefore:
		if value == "" {
			return fmt.Errorf("%s at position %d has no delimiter", opType, i)
		}
	}
	return nil
}

// validateDateTagOp checks the location and format of a date tag operation, and is a
// no-op for every other operation type. Only delete-date-tag accepts "all".
func validateDateTagOp(opType, opLoc, dateFormat string, i int) error {
	switch opType {
	case sharedconsts.OpDateTag, sharedconsts.OpDeleteDateTag:
	default:
		return nil
	}

	allowAll := opType == sharedconsts.OpDeleteDateTag
	if _, err := sharedenums.ParseDateTagLocation(opLoc, allowAll); err != nil {
		return fmt.Errorf("%s at position %d: %w", opType, i, err)
	}
	if _, err := sharedenums.ParseDateFormat(dateFormat); err != nil {
		return fmt.Errorf("%s at position %d: %w", opType, i, err)
	}
	return nil
}

// ValidFilterTypes lists the accepted filter types.
var ValidFilterTypes = sortedKeys(sharedconsts.ValidFilterTypes)

// ValidateFilterOps validates filter models.
func ValidateFilterOps(filters []sharedmodels.Filters) error {
	for i, filter := range filters {
		if _, ok := sharedconsts.ValidFilterTypes[filter.FilterType]; !ok {
			return fmt.Errorf("filter at position %d has invalid type %q (valid: %v)", i, filter.FilterType, ValidFilterTypes)
		}

		// A numeric comparison needs something numeric to compare against, and a regex a pattern that compiles and
		// can be written back inside "[[ ]]".
		switch filter.FilterType {
		case sharedconsts.OpMoreThan, sharedconsts.OpLessThan:
			if _, err := strconv.ParseFloat(filter.Value, 64); err != nil {
				return fmt.Errorf("filter at position %d has invalid value %q for type %q (must be a number)", i, filter.Value, filter.FilterType)
			}
		case sharedconsts.OpRegexContains:
			if filter.Value == "" {
				return fmt.Errorf("filter at position %d has an empty pattern for type %q", i, filter.FilterType)
			}
			if strings.Contains(filter.Value, "]]:") {
				return fmt.Errorf("filter at position %d: pattern cannot contain \"]]:\" (try [0-9] in place of [[:digit:]])", i)
			}
			if strings.Contains(filter.Value, "||") {
				return fmt.Errorf("filter at position %d: pattern cannot contain \"||\", which divides a filter from its operation", i)
			}
			if strings.Contains(filter.Value, "&&") {
				return fmt.Errorf("filter at position %d: pattern cannot contain \"&&\", which divides one filter from the next", i)
			}
			if _, err := regexp.Compile(filter.Value); err != nil {
				return fmt.Errorf("filter at position %d: invalid pattern %q: %w", i, filter.Value, err)
			}
		}

		if filter.MustAny != sharedconsts.OpMust && filter.MustAny != sharedconsts.OpAny {
			return fmt.Errorf("filter at position %d has invalid condition %q (must be %q or %q)", i, filter.MustAny, sharedconsts.OpMust, sharedconsts.OpAny)
		}
		if filter.Field == "" {
			return fmt.Errorf("filter at position %d has empty field", i)
		}
	}
	return nil
}

// ValidateFilteredMetaOps validates meta operations gated behind filters.
func ValidateFilteredMetaOps(filteredMetaOps []sharedmodels.FilteredMetaOps) error {
	for i, fmo := range filteredMetaOps {
		if err := ValidateFilterOps(fmo.Filters); err != nil {
			return fmt.Errorf("filtered meta operation at position %d has invalid filters: %w", i, err)
		}
		if err := ValidateMetaOps(fmo.MetaOps); err != nil {
			return fmt.Errorf("filtered meta operation at position %d has invalid meta operations: %w", i, err)
		}
		if len(fmo.Filters) == 0 {
			return fmt.Errorf("filtered meta operation at position %d has no filters", i)
		}
		if len(fmo.MetaOps) == 0 {
			return fmt.Errorf("filtered meta operation at position %d has no meta operations", i)
		}
	}
	return nil
}

// ValidateFilteredFilenameOps validates filename operations gated behind filters.
func ValidateFilteredFilenameOps(filteredFilenameOps []sharedmodels.FilteredFilenameOps) error {
	for i, ffo := range filteredFilenameOps {
		if err := ValidateFilterOps(ffo.Filters); err != nil {
			return fmt.Errorf("filtered filename operation at position %d has invalid filters: %w", i, err)
		}
		if err := ValidateFilenameOps(ffo.FilenameOps); err != nil {
			return fmt.Errorf("filtered filename operation at position %d has invalid filename operations: %w", i, err)
		}
		if len(ffo.Filters) == 0 {
			return fmt.Errorf("filtered filename operation at position %d has no filters", i)
		}
		if len(ffo.FilenameOps) == 0 {
			return fmt.Errorf("filtered filename operation at position %d has no filename operations", i)
		}
	}
	return nil
}
