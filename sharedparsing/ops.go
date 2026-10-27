package sharedparsing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// WarningKind classifies why parsing could not use an entry as written.
type WarningKind int

// WarningKind definitions.
const (
	// WarnDuplicate means the entry repeated one already accepted.
	WarnDuplicate WarningKind = iota
)

// Warning describes an entry parsing could not use as written.
type Warning struct {
	Kind  WarningKind
	Entry string
	Msg   string
}

// String renders the warning for logging.
func (w Warning) String() string {
	return w.Msg
}

// ParseFilenameOps parses and validates filename operation strings, e.g.
// "prefix:[COOL CATEGORY] " or "date-tag:prefix:ymd".
//
// Operation types and date tag locations are fixed vocabularies, so they are matched
// case-insensitively and stored lowercased. Values are left as written.
//
// Any malformed or invalid entry fails the batch, with err listing every malformed one.
// Duplicates are only described in warnings.
func ParseFilenameOps(filenameOps []string) (parsed []sharedmodels.FilenameOps, warnings []Warning, err error) {
	if len(filenameOps) == 0 {
		return nil, nil, nil
	}

	filenameOps, dupes := Deduplicate(filenameOps)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilenameOps, 0, len(filenameOps))
	seen := make(map[string]struct{}, len(filenameOps))
	var invalid []string

	for _, op := range filenameOps {
		opType, rest, ok := nextField(op)
		if !ok {
			invalid = append(invalid, invalidEntry(op, "expected 'prefix:[COOL CATEGORY] ' or 'date-tag:prefix:ymd'"))
			continue
		}

		newOp := sharedmodels.FilenameOps{OpType: strings.ToLower(opType)}
		args, argErr := parseOpArgs(newOp.OpType, rest)
		if argErr != nil {
			invalid = append(invalid, invalidEntry(op, argErr.Error()))
			continue
		}
		newOp.OpFindString, newOp.OpValue, newOp.OpLoc, newOp.DateFormat = args.find, args.value, args.loc, args.dateFormat

		key := FormatFilenameOp(newOp)
		if isDateTagOp(newOp.OpType) {
			key = newOp.OpType
		}

		if _, ok := seen[key]; ok {
			warnings = append(warnings, duplicateWarning(op))
			continue
		}
		seen[key] = struct{}{}

		valid = append(valid, newOp)
	}

	if len(invalid) > 0 {
		return nil, warnings, invalidError(invalid)
	}
	if len(valid) == 0 {
		return nil, warnings, errors.New("no valid filename operations")
	}
	if err := sharedvalidation.ValidateFilenameOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}

// ParseMetaOps parses and validates meta operation strings, e.g. "director:set:Spielberg"
// or "title:date-tag:suffix:ymd".
//
// Operation types and date tag locations are fixed vocabularies, so they are matched
// case-insensitively and stored lowercased. Field names and values keep their case,
// since metadata keys are case-sensitive. A comma-separated field list, e.g.
// "actor,director:set:Spielberg", gives one operation per field.
//
// Any malformed or invalid entry fails the batch, with err listing every malformed one.
// Duplicates are only described in warnings.
func ParseMetaOps(metaOps []string) (parsed []sharedmodels.MetaOps, warnings []Warning, err error) {
	if len(metaOps) == 0 {
		return nil, nil, nil
	}

	metaOps, dupes := Deduplicate(metaOps)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.MetaOps, 0, len(metaOps))
	seen := make(map[string]struct{}, len(metaOps))
	var invalid []string

	for _, op := range metaOps {
		field, rest, ok := nextField(op)
		var opType string
		if ok {
			opType, rest, ok = nextField(rest)
		}
		if !ok {
			invalid = append(invalid, invalidEntry(op, "expected 'director:set:Spielberg' or 'title:date-tag:suffix:ymd'"))
			continue
		}

		opType = strings.ToLower(opType)
		args, argErr := parseOpArgs(opType, rest)
		if argErr != nil {
			invalid = append(invalid, invalidEntry(op, argErr.Error()))
			continue
		}

		// One operation per listed field, e.g. "title,description:suffix: [x]".
		for _, f := range SplitMultiField(field) {
			newOp := sharedmodels.MetaOps{
				Field:        f,
				OpType:       opType,
				OpFindString: args.find,
				OpValue:      args.value,
				OpLoc:        args.loc,
				DateFormat:   args.dateFormat,
			}

			key := FormatMetaOp(newOp)
			if isDateTagOp(newOp.OpType) {
				key = newOp.Field + string(fieldSeparator) + newOp.OpType
			}

			// Name the skipped operation itself, as the rest of a field list may still apply.
			if _, ok := seen[key]; ok {
				warnings = append(warnings, duplicateWarning(FormatMetaOp(newOp)))
				continue
			}
			seen[key] = struct{}{}

			valid = append(valid, newOp)
		}
	}

	if len(invalid) > 0 {
		return nil, warnings, invalidError(invalid)
	}
	if len(valid) == 0 {
		return nil, warnings, errors.New("no valid meta operations")
	}
	if err := sharedvalidation.ValidateMetaOps(valid); err != nil {
		return nil, warnings, err
	}
	return valid, warnings, nil
}

// FormatFilenameOp renders a filename operation back to its string form, the inverse
// of [ParseFilenameOps].
func FormatFilenameOp(f sharedmodels.FilenameOps) string {
	args := opArgs{find: f.OpFindString, value: f.OpValue, loc: f.OpLoc, dateFormat: f.DateFormat}
	return Escape(f.OpType) + string(fieldSeparator) + formatOpArgs(f.OpType, args)
}

// FormatMetaOp renders a meta operation back to its string form, the inverse of
// [ParseMetaOps].
func FormatMetaOp(m sharedmodels.MetaOps) string {
	args := opArgs{find: m.OpFindString, value: m.OpValue, loc: m.OpLoc, dateFormat: m.DateFormat}
	lead := JoinEscaped([]string{m.Field, m.OpType}, fieldSeparator)
	return lead + string(fieldSeparator) + formatOpArgs(m.OpType, args)
}

// FormatFilenameOps renders each operation.
func FormatFilenameOps(ops []sharedmodels.FilenameOps) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, FormatFilenameOp(op))
	}
	return out
}

// FormatMetaOps renders each operation.
func FormatMetaOps(ops []sharedmodels.MetaOps) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, FormatMetaOp(op))
	}
	return out
}

// opArgs are the parts of an operation after its field and type.
type opArgs struct {
	find, value, loc, dateFormat string
}

// isDateTagOp reports whether opType takes a location and date format.
func isDateTagOp(opType string) bool {
	return opType == sharedconsts.OpDateTag || opType == sharedconsts.OpDeleteDateTag
}

// isFindReplaceOp reports whether opType takes find text and a replacement.
func isFindReplaceOp(opType string) bool {
	switch opType {
	case sharedconsts.OpReplace, sharedconsts.OpReplacePrefix, sharedconsts.OpReplaceSuffix:
		return true
	}
	return false
}

// parseOpArgs reads the parts of an operation after its type. The last part is the
// whole remainder, so it may hold ':' unescaped.
func parseOpArgs(opType, rest string) (a opArgs, err error) {
	switch {
	case isDateTagOp(opType):
		loc, format, ok := nextField(rest)
		if !ok {
			return a, fmt.Errorf("%s expects a location and date format, e.g. '%s:prefix:ymd'", opType, opType)
		}
		a.loc, a.dateFormat = strings.ToLower(loc), Unescape(format)

	case opType == sharedconsts.OpRegexReplace:
		end := rawSpanEnd(rest)
		if end < 0 {
			return a, fmt.Errorf("%s expects a pattern wrapped in %s %s, e.g. '%s:%s^foo %s:bar'", opType, openSquareBrackets, closeSquareBrackets, opType, openSquareBrackets, closeSquareBrackets)
		}
		a.find = rest[len(openSquareBrackets) : end-len(closeSquareBrackets)]
		if end < len(rest) {
			a.value = Unescape(rest[end+1:])
		}

	case isFindReplaceOp(opType):
		find, value, ok := nextField(rest)
		if !ok {
			return a, fmt.Errorf("%s expects find text and a replacement, e.g. '%s:old:new'", opType, opType)
		}
		a.find, a.value = find, Unescape(value)

	default:
		a.value = Unescape(rest)
	}
	return a, nil
}

// formatOpArgs renders the parts of an operation after its type, the inverse of
// [parseOpArgs].
func formatOpArgs(opType string, a opArgs) string {
	switch {
	case isDateTagOp(opType):
		return Escape(a.loc) + string(fieldSeparator) + escapeLast(a.dateFormat)
	case opType == sharedconsts.OpRegexReplace:
		return openSquareBrackets + a.find + closeSquareBrackets + string(fieldSeparator) + escapeLast(a.value)
	case isFindReplaceOp(opType):
		return Escape(a.find) + string(fieldSeparator) + escapeLast(a.value)
	default:
		return escapeLast(a.value)
	}
}

// invalidEntry describes a malformed entry.
func invalidEntry(entry, reason string) string {
	return fmt.Sprintf("%q: %s", entry, reason)
}

// invalidError reports every malformed entry at once, so all can be fixed in one pass.
func invalidError(invalid []string) error {
	return fmt.Errorf("invalid operations: %s", strings.Join(invalid, "; "))
}

// duplicateWarning reports an entry that repeated one already accepted.
func duplicateWarning(entry string) Warning {
	return Warning{
		Kind:  WarnDuplicate,
		Entry: entry,
		Msg:   fmt.Sprintf("skipping duplicate operation %q", entry),
	}
}

// duplicateWarnings describes entries dropped by [Deduplicate].
func duplicateWarnings(dupes []string) []Warning {
	if len(dupes) == 0 {
		return nil
	}

	warnings := make([]Warning, 0, len(dupes))
	for _, d := range dupes {
		warnings = append(warnings, duplicateWarning(d))
	}
	return warnings
}
