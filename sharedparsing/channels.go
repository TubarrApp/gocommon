package sharedparsing

import "strings"

// ChannelURLsEqual reports whether two channel URLs name the same channel.
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
func ChannelURLsEqual(a, b string) bool {
	return NormalizeChannelURL(a) == NormalizeChannelURL(b)
}

// AppliesToChannel reports whether something scoped to opURL applies to chanURL.
//
// An empty opURL is unscoped and applies everywhere, which is how an operation or
// filter with no channel URL is treated.
func AppliesToChannel(opURL, chanURL string) bool {
	return opURL == "" || ChannelURLsEqual(opURL, chanURL)
}

// NormalizeChannelURL renders a channel URL comparable, for use as a map key where a
// pairwise [ChannelURLsEqual] does not fit. See that function for the rules.
func NormalizeChannelURL(s string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(s), "/"))
}
