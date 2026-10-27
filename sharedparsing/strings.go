// Package sharedparsing holds the operation-string grammar shared by Metarr and Tubarr.
//
// An operation is separator-delimited, e.g. "title:set:cats". The grammar has exactly one
// decoding pass: the "||" dividing a filter from its operation is located without
// decoding ([SplitUnescaped]), and the final split on ':' decodes ([EscapedSplit]).
// Decoding only once is what allows the escape character itself to be escaped, so any
// value is representable.
//
// An operation's last part is everything after the previous ':', so it needs no
// escaping of ':'. A regex-replace pattern is wrapped in "[[ ]]" and taken verbatim.
package sharedparsing

import (
	"strings"
)

const (
	escapeByte     = '\\'
	fieldSeparator = ':'
	pipe           = '|' // Half of a FilterSeparator ('||').
	amp            = '&' // Half of a FilterJoiner ('&&').

	rawOpen  = "[["
	rawClose = "]]"
)

// rawSpanEnd returns the index just past the "]]" closing a raw span opening at the
// start of s, or -1. The span closes at the first "]]" followed by ':' or the end of s,
// so patterns may hold "]]" elsewhere, as in "[[:digit:]]".
func rawSpanEnd(s string) int {
	if !strings.HasPrefix(s, rawOpen) {
		return -1
	}
	for i := len(rawOpen); i+len(rawClose) <= len(s); i++ {
		if s[i:i+len(rawClose)] != rawClose {
			continue
		}
		if end := i + len(rawClose); end == len(s) || s[end] == fieldSeparator {
			return end
		}
	}
	return -1
}

// indexUnescaped returns the index of the first unescaped sep in s, or -1.
func indexUnescaped(s string, sep byte) int {
	escaped := false
	for i := range len(s) {
		switch {
		case escaped:
			escaped = false
		case s[i] == escapeByte:
			escaped = true
		case s[i] == sep:
			return i
		}
	}
	return -1
}

// lastIndexUnescaped returns the index of the last unescaped sep in s, or -1.
func lastIndexUnescaped(s string, sep byte) int {
	last := -1
	escaped := false
	for i := range len(s) {
		switch {
		case escaped:
			escaped = false
		case s[i] == escapeByte:
			escaped = true
		case s[i] == sep:
			last = i
		}
	}
	return last
}

// nextField decodes s up to its first unescaped ':' and returns the undecoded
// remainder after it. ok is false when s holds no separator.
func nextField(s string) (field, rest string, ok bool) {
	i := indexUnescaped(s, fieldSeparator)
	if i < 0 {
		return "", "", false
	}
	return unescape(s[:i]), s[i+1:], true
}

// unescape decodes every escape sequence in s, treating ':' as literal.
func unescape(s string) string {
	if !strings.ContainsRune(s, escapeByte) {
		return s
	}
	return strings.Join(SplitAndRemoveEscapes(s, fieldSeparator), string(fieldSeparator))
}

// escapeLast renders the last part of an operation, which needs only '\' escaped since
// it is never split.
func escapeLast(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

// isMetaByte reports whether c carries meaning in the grammar, so a backslash before it
// is an escape.
func isMetaByte(c byte) bool {
	return c == escapeByte || c == fieldSeparator || c == pipe || c == amp
}

// SplitAndKeepEscapes splits s on unescaped occurrences of desiredSeparator, leaving every
// escape sequence intact for a later [EscapedSplit].
//
// Use this for intermediate splits, such as dividing filter rules from operations, so
// the escapes survive to the single decoding pass.
func SplitAndKeepEscapes(s string, desiredSeparator byte) []string {
	var parts []string
	var buf strings.Builder
	escaped := false

	for i := range len(s) {
		c := s[i]
		switch {
		case escaped:
			buf.WriteByte(c)
			escaped = false
		case c == escapeByte:
			buf.WriteByte(c)
			escaped = true
		case c == desiredSeparator:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	return append(parts, buf.String())
}

// SplitAndRemoveEscapes splits s on unescaped occurrences of desiredSeparator and decodes the
// escape sequences, so it is the final pass over an operation string.
//
// A backslash before a grammar metacharacter ('\', ':', '|' or '&') yields that character
// literally; before anything else it is retained verbatim, which leaves values such as
// "\d" untouched. Scanning is byte-wise so segments holding invalid UTF-8 survive.
func SplitAndRemoveEscapes(s string, desiredSeparator byte) []string {
	var parts []string
	var buf strings.Builder
	escaped := false

	for i := range len(s) {
		c := s[i]
		switch {
		case escaped:
			if !isMetaByte(c) {
				buf.WriteByte(escapeByte) // Not a grammar character, so retain the backslash.
			}
			buf.WriteByte(c)
			escaped = false
		case c == escapeByte:
			escaped = true
		case c == desiredSeparator:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	if escaped {
		buf.WriteByte(escapeByte)
	}
	return append(parts, buf.String())
}

// Escape renders s so that a later [EscapedSplit] recovers it verbatim as one segment.
//
// Every grammar metacharacter is escaped, not just the separator in use, because a
// filtered entry is split at its "||" before its parts are split on ':'. A '&' is only
// escaped where it starts a [FilterJoiner].
func Escape(s string) string {
	if !strings.ContainsAny(s, `\:|`) && !strings.Contains(s, FilterJoiner) {
		return s
	}

	var buf strings.Builder
	buf.Grow(len(s) + 8)
	for i := range len(s) {
		if isMetaByte(s[i]) && (s[i] != amp || strings.HasPrefix(s[i:], FilterJoiner)) {
			buf.WriteByte(escapeByte)
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}

// JoinEscaped escapes each segment and joins them with desiredSeparator.
//
// Use this rather than concatenation when rebuilding an operation string, or values
// holding a separator will gain fields when re-parsed.
func JoinEscaped(parts []string, desiredSeparator byte) string {
	if len(parts) == 0 {
		return ""
	}

	escaped := make([]string, len(parts))
	for i, p := range parts {
		escaped[i] = Escape(p)
	}
	return strings.Join(escaped, string(desiredSeparator))
}

// Deduplicate removes duplicates from input, preserving first-appearance order and
// returning the removed entries for the caller to report.
func Deduplicate(input []string) (deduped, removed []string) {
	if len(input) == 0 {
		return input, nil
	}

	deduped = make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))

	for _, in := range input {
		if _, ok := seen[in]; ok {
			removed = append(removed, in)
			continue
		}
		seen[in] = struct{}{}
		deduped = append(deduped, in)
	}
	return deduped, removed
}
