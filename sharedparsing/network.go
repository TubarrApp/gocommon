package sharedparsing

import "strings"

// StripURLScheme drops a URL's leading "scheme://" (e.g. "https://youtube.com/@cats" → "youtube.com/@cats").
func StripURLScheme(u string) string {
	scheme, rest, ok := strings.Cut(u, "://")
	if !ok || !isURLScheme(scheme) {
		return u
	}
	return rest
}

// isURLScheme reports whether s can be a URL's scheme (the "https" in "https://"): letters, digits, "+", "-" and ".".
func isURLScheme(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		isLetter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != '+' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}
