package sharedvalidation

import (
	"fmt"
	"maps"
	"slices"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedenums"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// sortedKeys returns the keys of m in a stable order, for use in error messages.
func sortedKeys(m map[string]struct{}) []string {
	return slices.Sorted(maps.Keys(m))
}

// validFilenameOpTypes are the operation types accepted for filename operations.
var validFilenameOpTypes = map[string]struct{}{
	sharedconsts.OpAppend:        {},
	sharedconsts.OpPrefix:        {},
	sharedconsts.OpReplacePrefix: {},
	sharedconsts.OpReplaceSuffix: {},
	sharedconsts.OpReplace:       {},
	sharedconsts.OpDateTag:       {},
	sharedconsts.OpDeleteDateTag: {},
	sharedconsts.OpSet:           {},
}

// validMetaOpTypes are the operation types accepted for meta operations.
var validMetaOpTypes = map[string]struct{}{
	sharedconsts.OpAppend:        {},
	sharedconsts.OpCopyTo:        {},
	sharedconsts.OpPasteFrom:     {},
	sharedconsts.OpPrefix:        {},
	sharedconsts.OpReplacePrefix: {},
	sharedconsts.OpReplaceSuffix: {},
	sharedconsts.OpReplace:       {},
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
		if op.Field == "" {
			return fmt.Errorf("meta operation at position %d has empty field", i)
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
