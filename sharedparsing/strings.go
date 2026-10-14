// Package sharedparsing holds the operation-string grammar shared by Metarr and Tubarr.
//
// An operation is separator-delimited, e.g. "title:set:cats", optionally behind a
// "channelURL|" prefix. The grammar has exactly one decoding pass: '|' delimiters are
// located without decoding ([SplitUnescaped]), and the final split on ':' decodes
// ([EscapedSplit]). Decoding only once is what allows the escape character itself to be
// escaped, so any value is representable.
//
// An operation's last part is everything after the previous ':', so it needs no
// escaping of ':'. A regex-replace pattern is wrapped in "[[ ]]" and taken verbatim.
package sharedparsing

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

const (
	escapeByte     = '\\'
	fieldSeparator = ':'
	urlSeparator   = '|'

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

// opensRawSpan reports whether seg, an operation scanned so far, ends at the pattern
// position of a regex-replace, where a raw span may begin.
func opensRawSpan(seg string) bool {
	const marker = sharedconsts.OpRegexReplace + string(fieldSeparator)
	if len(seg) < len(marker) || !strings.EqualFold(seg[len(seg)-len(marker):], marker) {
		return false
	}
	return len(seg) == len(marker) || seg[len(seg)-len(marker)-1] == fieldSeparator
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
	return strings.Join(EscapedSplit(s, fieldSeparator), string(fieldSeparator))
}

// escapeLast renders the last part of an operation, which needs no ':' escaped.
func escapeLast(s string) string {
	if !strings.ContainsAny(s, `\|`) {
		return s
	}

	var buf strings.Builder
	buf.Grow(len(s) + 8)
	for i := range len(s) {
		if s[i] == escapeByte || s[i] == urlSeparator {
			buf.WriteByte(escapeByte)
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}

// isMetaByte reports whether c carries meaning in the grammar and so must be escaped
// to appear literally in a value.
func isMetaByte(c byte) bool {
	return c == escapeByte || c == fieldSeparator || c == urlSeparator
}

// SplitUnescaped splits s on unescaped occurrences of desiredSeparator, leaving every
// escape sequence intact for a later [EscapedSplit].
//
// Use this for intermediate splits, such as stripping a channel URL or dividing filter
// rules from operations, so the escapes survive to the single decoding pass. A
// regex-replace pattern's "[[ ]]" span is copied whole, so it may hold the separator.
func SplitUnescaped(s string, desiredSeparator byte) []string {
	var parts []string
	var buf strings.Builder
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case escaped:
			buf.WriteByte(c)
			escaped = false
		case c == escapeByte:
			buf.WriteByte(c)
			escaped = true
		case c == rawOpen[0] && opensRawSpan(buf.String()):
			n := max(rawSpanEnd(s[i:]), 1) // Not a closed span, so copy just this byte.
			buf.WriteString(s[i : i+n])
			i += n - 1
		case c == desiredSeparator:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	return append(parts, buf.String())
}

// EscapedSplit splits s on unescaped occurrences of desiredSeparator and decodes the
// escape sequences, so it is the final pass over an operation string.
//
// A backslash before a grammar metacharacter ('\', ':' or '|') yields that character
// literally; before anything else it is retained verbatim, which leaves values such as
// "\d" untouched. Scanning is byte-wise so segments holding invalid UTF-8 survive.
func EscapedSplit(s string, desiredSeparator byte) []string {
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
// Every grammar metacharacter is escaped, not just the separator in use, because an
// operation is split on '|' before it is split on ':'.
func Escape(s string) string {
	if !strings.ContainsAny(s, `\:|`) {
		return s
	}

	var buf strings.Builder
	buf.Grow(len(s) + 8)
	for i := range len(s) {
		if isMetaByte(s[i]) {
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

// isChannelURL reports whether s is usable as a channel URL prefix.
//
// Both a scheme and a host are required. Testing only that s parses is not enough,
// since url.ParseRequestURI reads "title:set:cats" as the scheme "title" with the
// opaque body "set:cats", which would make any value holding a '|' look prefixed.
func isChannelURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// SplitOpURL separates an optional leading "channelURL|" prefix from an operation.
//
// rest keeps its escapes for the caller's decoding split. A channel URL cannot legally
// contain an unencoded '|' or '\', so it is returned as found.
//
// When the leading segment is not a channel URL there is no prefix, so op is returned
// whole and the '|' is left to the operation. That keeps unescaped values such as
// "title:set:Cats | Dogs" working. A non-nil err means the segment looked like an
// attempted URL but was unusable; rest is still valid, so callers should log and
// continue.
func SplitOpURL(op string) (chanURL, rest string, err error) {
	split := SplitUnescaped(op, urlSeparator)
	if len(split) < 2 {
		return "", op, nil
	}

	u := split[0]
	if !isChannelURL(u) {
		if strings.Contains(u, "://") {
			return "", op, fmt.Errorf("operation %q has an unusable channel URL in leading segment %q", op, u)
		}
		return "", op, nil
	}
	return u, strings.Join(split[1:], string(urlSeparator)), nil
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
