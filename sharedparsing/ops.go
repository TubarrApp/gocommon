package sharedparsing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// ParseFilenameOps parses filename operation strings, e.g. "prefix:[COOL CATEGORY] "
// or "date-tag:prefix:ymd".
//
// Operation types and date tag locations are fixed vocabularies, so they are matched
// case-insensitively and stored lowercased. Values are left as written.
//
// Malformed entries are skipped and described in warnings rather than failing the
// batch; err is only returned when no entry was usable.
func ParseFilenameOps(filenameOps []string) (parsed []sharedmodels.FilenameOps, warnings []string, err error) {
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
			warnings = append(warnings, urlErr.Error())
		}

		split := EscapedSplit(opPart, fieldSeparator)
		if len(split) < 2 || len(split) > 3 {
			warnings = append(warnings, fmt.Sprintf("skipping invalid filename operation %q (expected 'prefix:[COOL CATEGORY] ' or 'date-tag:prefix:ymd')", op))
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
				warnings = append(warnings, fmt.Sprintf("skipping invalid filename operation type %q", split[0]))
				continue
			}
		}

		if _, ok := seen[key]; ok {
			warnings = append(warnings, fmt.Sprintf("skipping duplicate filename operation %q", opPart))
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
func ParseMetaOps(metaOps []string) (parsed []sharedmodels.MetaOps, warnings []string, err error) {
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
			warnings = append(warnings, urlErr.Error())
		}

		split := EscapedSplit(opPart, fieldSeparator)
		if len(split) < 3 || len(split) > 4 {
			warnings = append(warnings, fmt.Sprintf("skipping invalid meta operation %q (expected 'director:set:Spielberg' or 'title:date-tag:suffix:ymd')", op))
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
				warnings = append(warnings, fmt.Sprintf("skipping invalid four-part meta operation type %q", newOp.OpType))
				continue
			}
		}

		if _, ok := seen[key]; ok {
			warnings = append(warnings, fmt.Sprintf("skipping duplicate meta operation %q", opPart))
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

// duplicateWarnings describes entries dropped by [Deduplicate].
func duplicateWarnings(dupes []string) []string {
	if len(dupes) == 0 {
		return nil
	}

	warnings := make([]string, 0, len(dupes))
	for _, d := range dupes {
		warnings = append(warnings, fmt.Sprintf("removing duplicate entry %q", d))
	}
	return warnings
}
