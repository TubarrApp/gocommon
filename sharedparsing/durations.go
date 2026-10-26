package sharedparsing

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses a duration in seconds, as "1:02:03", "62:03", "3723", or ISO 8601 "P1DT1H2M3S".
func ParseDuration(s string) (int, error) {
	s = strings.TrimSpace(s)
	if secs, err := strconv.ParseFloat(s, 64); err == nil {
		return max(int(secs), 0), nil
	}
	if rest, ok := strings.CutPrefix(strings.ToUpper(s), "P"); ok {
		if rest == "" || rest == "T" {
			return 0, fmt.Errorf("duration %q has no days or time", s)
		}
		// Days come before the "T", and the time after it.
		days := 0
		if d, after, found := strings.Cut(rest, "D"); found {
			n, err := strconv.Atoi(d)
			if err != nil || n < 0 {
				return 0, fmt.Errorf("invalid days in duration %q", s)
			}
			days = n
			rest = after
		}
		secs := 0
		if rest = strings.TrimPrefix(rest, "T"); rest != "" {
			d, err := time.ParseDuration(strings.ToLower(rest))
			if err != nil || d < 0 {
				return 0, fmt.Errorf("invalid time in duration %q", s)
			}
			secs = int(d.Seconds())
		}
		return days*24*60*60 + secs, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("duration %q is not seconds, h:mm:ss, m:ss, or ISO 8601", s)
	}
	total := 0
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid part %q in duration %q", p, s)
		}
		total = total*60 + n
	}
	return total, nil
}
