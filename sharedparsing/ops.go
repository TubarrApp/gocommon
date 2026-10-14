package sharedparsing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// WarningKind classifies why parsing could not use an entry as written.
type WarningKind int

// WarningKind definitions.
const (
	// WarnInvalid means the entry was malformed or named an unknown operation type and
	// was skipped entirely.
	WarnInvalid WarningKind = iota
	// WarnDuplicate means the entry repeated one already accepted.
	WarnDuplicate
	// WarnChannelURL means a leading segment looked like a channel URL but was unusable.
	WarnChannelURL
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

// InvalidEntries returns the entries that were malformed, for callers that treat a bad
// operation as fatal rather than skipping it. Duplicates are not included, since
// dropping them changes nothing about the result.
func InvalidEntries(warnings []Warning) []string {
	var invalid []string
	for _, w := range warnings {
		if w.Kind == WarnInvalid {
			invalid = append(invalid, w.Entry)
		}
	}
	return invalid
}

// ParseFilenameOps parses filename operation strings, e.g. "prefix:[COOL CATEGORY] "
// or "date-tag:prefix:ymd".
//
// Operation types and date tag locations are fixed vocabularies, so they are matched
// case-insensitively and stored lowercased. Values are left as written.
//
// Malformed entries are skipped and described in warnings rather than failing the
// batch; err is only returned when no entry was usable.
func ParseFilenameOps(filenameOps []string) (parsed []sharedmodels.FilenameOps, warnings []Warning, err error) {
	if len(filenameOps) == 0 {
		return nil, nil, nil
	}

	filenameOps, dupes := Deduplicate(filenameOps)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.FilenameOps, 0, len(filenameOps))
	seen := make(map[string]struct{}, len(filenameOps))

	for _, op := range filenameOps {
		opURL, opPart, urlErr := SplitOpURL(op)
		if urlErr != nil {
			warnings = append(warnings, Warning{Kind: WarnChannelURL, Entry: op, Msg: urlErr.Error()})
		}

		opType, rest, ok := nextField(opPart)
		if !ok {
			warnings = append(warnings, invalidWarning(op, "expected 'prefix:[COOL CATEGORY] ' or 'date-tag:prefix:ymd'"))
			continue
		}

		newOp := sharedmodels.FilenameOps{OpType: strings.ToLower(opType)}
		args, argErr := parseOpArgs(newOp.OpType, rest)
		if argErr != nil {
			warnings = append(warnings, invalidWarning(op, argErr.Error()))
			continue
		}
		newOp.OpFindString, newOp.OpValue, newOp.OpLoc, newOp.DateFormat = args.find, args.value, args.loc, args.dateFormat

		key := FormatFilenameOp(newOp, false)
		if isDateTagOp(newOp.OpType) {
			key = newOp.OpType
		}

		if _, ok := seen[key]; ok {
			warnings = append(warnings, duplicateWarning(op))
			continue
		}
		seen[key] = struct{}{}

		newOp.ChannelURL = opURL
		valid = append(valid, newOp)
	}

	if len(valid) == 0 {
		return nil, warnings, errors.New("no valid filename operations")
	}
	return valid, warnings, nil
}

// ParseMetaOps parses meta operation strings, e.g. "director:set:Spielberg" or
// "title:date-tag:suffix:ymd".
//
// Operation types and date tag locations are fixed vocabularies, so they are matched
// case-insensitively and stored lowercased. Field names and values are left as written,
// since metadata keys are case-sensitive.
//
// Malformed entries are skipped and described in warnings rather than failing the
// batch; err is only returned when no entry was usable.
func ParseMetaOps(metaOps []string) (parsed []sharedmodels.MetaOps, warnings []Warning, err error) {
	if len(metaOps) == 0 {
		return nil, nil, nil
	}

	metaOps, dupes := Deduplicate(metaOps)
	warnings = duplicateWarnings(dupes)

	valid := make([]sharedmodels.MetaOps, 0, len(metaOps))
	seen := make(map[string]struct{}, len(metaOps))

	for _, op := range metaOps {
		opURL, opPart, urlErr := SplitOpURL(op)
		if urlErr != nil {
			warnings = append(warnings, Warning{Kind: WarnChannelURL, Entry: op, Msg: urlErr.Error()})
		}

		field, rest, ok := nextField(opPart)
		var opType string
		if ok {
			opType, rest, ok = nextField(rest)
		}
		if !ok {
			warnings = append(warnings, invalidWarning(op, "expected 'director:set:Spielberg' or 'title:date-tag:suffix:ymd'"))
			continue
		}

		newOp := sharedmodels.MetaOps{Field: field, OpType: strings.ToLower(opType)}
		args, argErr := parseOpArgs(newOp.OpType, rest)
		if argErr != nil {
			warnings = append(warnings, invalidWarning(op, argErr.Error()))
			continue
		}
		newOp.OpFindString, newOp.OpValue, newOp.OpLoc, newOp.DateFormat = args.find, args.value, args.loc, args.dateFormat

		key := FormatMetaOp(newOp, false)
		if isDateTagOp(newOp.OpType) {
			key = newOp.Field + string(fieldSeparator) + newOp.OpType
		}

		if _, ok := seen[key]; ok {
			warnings = append(warnings, duplicateWarning(op))
			continue
		}
		seen[key] = struct{}{}

		newOp.ChannelURL = opURL
		valid = append(valid, newOp)
	}

	if len(valid) == 0 {
		return nil, warnings, errors.New("no valid meta operations")
	}
	return valid, warnings, nil
}

// FormatFilenameOp renders a filename operation back to its string form, the inverse
// of [ParseFilenameOps]. includeURL prepends the channel URL when one is set.
func FormatFilenameOp(f sharedmodels.FilenameOps, includeURL bool) string {
	args := opArgs{find: f.OpFindString, value: f.OpValue, loc: f.OpLoc, dateFormat: f.DateFormat}
	return withURL(Escape(f.OpType)+string(fieldSeparator)+formatOpArgs(f.OpType, args), f.ChannelURL, includeURL)
}

// FormatMetaOp renders a meta operation back to its string form, the inverse of
// [ParseMetaOps]. includeURL prepends the channel URL when one is set.
func FormatMetaOp(m sharedmodels.MetaOps, includeURL bool) string {
	args := opArgs{find: m.OpFindString, value: m.OpValue, loc: m.OpLoc, dateFormat: m.DateFormat}
	lead := JoinEscaped([]string{m.Field, m.OpType}, fieldSeparator)
	return withURL(lead+string(fieldSeparator)+formatOpArgs(m.OpType, args), m.ChannelURL, includeURL)
}

// FormatFilenameOps renders each operation, skipping any that carry a channel URL
// other than chanURL. An empty chanURL keeps every operation.
func FormatFilenameOps(ops []sharedmodels.FilenameOps, chanURL string, includeURL bool) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		if !opAppliesTo(op.ChannelURL, chanURL) {
			continue
		}
		out = append(out, FormatFilenameOp(op, includeURL))
	}
	return out
}

// FormatMetaOps renders each operation, skipping any that carry a channel URL other
// than chanURL. An empty chanURL keeps every operation.
func FormatMetaOps(ops []sharedmodels.MetaOps, chanURL string, includeURL bool) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		if !opAppliesTo(op.ChannelURL, chanURL) {
			continue
		}
		out = append(out, FormatMetaOp(op, includeURL))
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
		a.loc, a.dateFormat = strings.ToLower(loc), unescape(format)

	case opType == sharedconsts.OpRegexReplace:
		end := rawSpanEnd(rest)
		if end < 0 {
			return a, fmt.Errorf("%s expects a pattern wrapped in %s %s, e.g. '%s:%s^foo %s:bar'", opType, rawOpen, rawClose, opType, rawOpen, rawClose)
		}
		a.find = rest[len(rawOpen) : end-len(rawClose)]
		if end < len(rest) {
			a.value = unescape(rest[end+1:])
		}

	case isFindReplaceOp(opType):
		find, value, ok := nextField(rest)
		if !ok {
			return a, fmt.Errorf("%s expects find text and a replacement, e.g. '%s:old:new'", opType, opType)
		}
		a.find, a.value = find, unescape(value)

	default:
		a.value = unescape(rest)
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
		return rawOpen + a.find + rawClose + string(fieldSeparator) + escapeLast(a.value)
	case isFindReplaceOp(opType):
		return Escape(a.find) + string(fieldSeparator) + escapeLast(a.value)
	default:
		return escapeLast(a.value)
	}
}

// withURL optionally puts op behind a "channelURL|" prefix. A channel URL cannot
// legally hold an unencoded metacharacter, so it is written as given.
func withURL(op, chanURL string, includeURL bool) string {
	if !includeURL || chanURL == "" {
		return op
	}
	return chanURL + string(urlSeparator) + op
}

// opAppliesTo reports whether an operation scoped to opURL applies to chanURL.
func opAppliesTo(opURL, chanURL string) bool {
	return opURL == "" || chanURL == "" || opURL == chanURL
}

// invalidWarning reports an entry that was skipped as malformed.
func invalidWarning(entry, reason string) Warning {
	return Warning{
		Kind:  WarnInvalid,
		Entry: entry,
		Msg:   fmt.Sprintf("skipping invalid operation %q: %s", entry, reason),
	}
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
