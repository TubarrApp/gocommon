package sharedparsing

import "strings"

// NormalizeChannelURL renders a channel URL comparable, so two channel URLs name the same channel when they normalize
// the same.
//
// A trailing slash is ignored and comparison is case-insensitive, so
// "google.com/" and "GOOGLE.com" name the same channel. Nothing else is normalized:
// both sides are expected to be a channel URL as stored, so this is an equality test
// that forgives the ways the same URL gets typed rather than a containment test.
//
// Case is folded across the whole URL, not just the host. The host is case-insensitive
// by definition, and while a path can be case-sensitive, the only channel URLs where
// path case carries meaning are opaque IDs such as YouTube's "/channel/UC..." form.
// Two channels whose URLs differ only by path case is not a realistic collision,
// whereas a handle retyped in a different case is a realistic mistake.
func NormalizeChannelURL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/")
	s = strings.ToLower(s)
	return s
}
