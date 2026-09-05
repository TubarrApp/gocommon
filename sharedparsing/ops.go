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

		split := EscapedSplit(opPart, fieldSeparator)
		if len(split) < 2 || len(split) > 3 {
			warnings = append(warnings, invalidWarning(op, "expected 'prefix:[COOL CATEGORY] ' or 'date-tag:prefix:ymd'"))
			continue
		}

		var newOp sharedmodels.FilenameOps
		var key string

		switch len(split) {
		case 2: // e.g. 'prefix:[DOG VIDEOS]'
			newOp.OpType = strings.ToLower(split[0])
			newOp.OpValue = split[1]
			key = strings.Join([]string{newOp.OpType, newOp.OpValue}, ":")

		case 3: // e.g. 'replace-suffix:_1:' or 'date-tag:prefix:ymd'
			newOp.OpType = strings.ToLower(split[0])

			switch newOp.OpType {
			case sharedconsts.OpReplaceSuffix, sharedconsts.OpReplacePrefix, sharedconsts.OpReplace:
				newOp.OpFindString = split[1]
				newOp.OpValue = split[2]
				key = strings.Join([]string{newOp.OpType, newOp.OpFindString, newOp.OpValue}, ":")

			case sharedconsts.OpDateTag, sharedconsts.OpDeleteDateTag:
				newOp.OpLoc = strings.ToLower(split[1])
				newOp.DateFormat = split[2]
				key = newOp.OpType

			default:
				warnings = append(warnings, invalidWarning(op, fmt.Sprintf("unknown filename operation type %q", split[0])))
				continue
			}
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

		split := EscapedSplit(opPart, fieldSeparator)
		if len(split) < 3 || len(split) > 4 {
			warnings = append(warnings, invalidWarning(op, "expected 'director:set:Spielberg' or 'title:date-tag:suffix:ymd'"))
			continue
		}

		var newOp sharedmodels.MetaOps
		var key string

		switch len(split) {
		case 3: // e.g. 'director:set:Spielberg'
			newOp.Field = split[0]
			newOp.OpType = strings.ToLower(split[1])
			newOp.OpValue = split[2]
			key = strings.Join([]string{newOp.Field, newOp.OpType, newOp.OpValue}, ":")

		case 4: // e.g. 'title:date-tag:suffix:ymd' or 'title:replace:old:new'
			newOp.Field = split[0]
			newOp.OpType = strings.ToLower(split[1])

			switch newOp.OpType {
			case sharedconsts.OpDateTag, sharedconsts.OpDeleteDateTag:
				newOp.OpLoc = strings.ToLower(split[2])
				newOp.DateFormat = split[3]
				key = strings.Join([]string{newOp.Field, newOp.OpType}, ":")

			case sharedconsts.OpReplace, sharedconsts.OpReplaceSuffix, sharedconsts.OpReplacePrefix:
				newOp.OpFindString = split[2]
				newOp.OpValue = split[3]
				key = strings.Join([]string{newOp.Field, newOp.OpType, newOp.OpFindString, newOp.OpValue}, ":")

			default:
				warnings = append(warnings, invalidWarning(op, fmt.Sprintf("unknown four-part meta operation type %q", newOp.OpType)))
				continue
			}
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
	var fields []string
	switch f.OpType {
	case sharedconsts.OpDateTag, sharedconsts.OpDeleteDateTag:
		fields = []string{f.OpType, f.OpLoc, f.DateFormat}
	case sharedconsts.OpReplace, sharedconsts.OpReplaceSuffix, sharedconsts.OpReplacePrefix:
		fields = []string{f.OpType, f.OpFindString, f.OpValue}
	default:
		fields = []string{f.OpType, f.OpValue}
	}
	return formatOpFields(fields, f.ChannelURL, includeURL)
}

// FormatMetaOp renders a meta operation back to its string form, the inverse of
// [ParseMetaOps]. includeURL prepends the channel URL when one is set.
func FormatMetaOp(m sharedmodels.MetaOps, includeURL bool) string {
	var fields []string
	switch m.OpType {
	case sharedconsts.OpDateTag, sharedconsts.OpDeleteDateTag:
		fields = []string{m.Field, m.OpType, m.OpLoc, m.DateFormat}
	case sharedconsts.OpReplace, sharedconsts.OpReplaceSuffix, sharedconsts.OpReplacePrefix:
		fields = []string{m.Field, m.OpType, m.OpFindString, m.OpValue}
	default:
		fields = []string{m.Field, m.OpType, m.OpValue}
	}
	return formatOpFields(fields, m.ChannelURL, includeURL)
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

// formatOpFields escapes each field and joins them, optionally behind a "channelURL|"
// prefix. A channel URL cannot legally hold an unencoded metacharacter, so it is
// written as given.
func formatOpFields(fields []string, chanURL string, includeURL bool) string {
	op := JoinEscaped(fields, fieldSeparator)

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
