package sharedparsing

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseDuration parses a duration in seconds, as "1:02:03", "62:03", "3723", or ISO 8601 "P1DT1H2M3S".
func ParseDuration(s string) (int, error) {
	s = strings.TrimSpace(s)

	// Define max and min integer values.
	maxInt := int(^uint(0) >> 1)
	floatIntLimit := math.Ldexp(1, strconv.IntSize-1)

	// Directly parsed without error.
	if secs, err := strconv.ParseFloat(s, 64); err == nil {

		// Reject NaN, infinity, and values outside the int range.
		if math.IsNaN(secs) || math.IsInf(secs, 0) || secs >= floatIntLimit || secs < -floatIntLimit {
			return 0, fmt.Errorf("duration %q is out of range", s)
		}
		// Return plain parsed seconds.
		return max(int(secs), 0), nil
	}

	// ISO 8601 Duration is prefixed with 'P' (e.g. 'P1DT1H2M3S').
	if rest, containsP := strings.CutPrefix(strings.ToUpper(s), "P"); containsP {
		if rest == "" {
			return 0, fmt.Errorf("duration %q has no days or time", s)
		}
		return parseISO8601(rest)
	}

	// Not ISO 8601 Duration, assume 'hh:mm:ss' format.
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("duration %q is not seconds, hh:mm:ss, mm:ss, or ISO 8601", s)
	}

	// Two or three part time (e.g. 'mm:ss', 'hh:mm:ss').
	total := 0
	for _, p := range parts {
		// Grab component number (i.e. 'hh', 'mm', or 'ss').
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid part %q in duration %q", p, s)
		}
		if total > (maxInt-n)/60 {
			return 0, fmt.Errorf("duration %q is out of range", s)
		}
		total = total*60 + n
	}
	return total, nil
}

// parseISO8601 parses ISO 8601 type time, such as "P1DT1H2M3S".
func parseISO8601(s string) (int, error) {
	// Separate the date and time components.
	date, clock, hasT := strings.Cut(s, "T")
	if strings.Contains(clock, "T") {
		return 0, fmt.Errorf("invalid time in duration %q", s)
	}

	// Weeks must be specified on their own (e.g. 'P1W').
	if weeks, after, foundW := strings.Cut(date, "W"); foundW {
		maxInt := int(^uint(0) >> 1)
		if after != "" || hasT || weeks == "" {
			return 0, fmt.Errorf("invalid weeks in duration %q", s)
		}
		n, err := strconv.Atoi(weeks)
		if err != nil || n < 0 || n > maxInt/(7*24*60*60) {
			return 0, fmt.Errorf("invalid weeks in duration %q", s)
		}
		return n * 7 * 24 * 60 * 60, nil
	}

	// Years and months cannot be converted to seconds without a reference date.
	if strings.ContainsAny(date, "YM") {
		return 0, fmt.Errorf("unsupported date component in duration %q", s)
	}

	// Days come before the 'T'.
	days := 0
	if date != "" {
		d, after, foundD := strings.Cut(date, "D")
		if !foundD || after != "" || d == "" {
			return 0, fmt.Errorf("invalid days in duration %q", s)
		}
		n, err := strconv.Atoi(d)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid days in duration %q", s)
		}
		days = n
	}

	// A duration must contain a date or time component.
	if date == "" && !hasT || hasT && clock == "" {
		return 0, fmt.Errorf("duration %q has no days or time", s)
	}

	// Parse the time components after 'T'.
	secs := 0
	if clock != "" {
		var err error
		secs, err = parseISO8601Time(clock)
		if err != nil {
			return 0, fmt.Errorf("invalid time in duration %q", s)
		}
	}

	// Add days in seconds, checking for overflow.
	maxInt := int(^uint(0) >> 1)
	if days > (maxInt-secs)/(24*60*60) {
		return 0, fmt.Errorf("duration %q is out of range", s)
	}
	return days*24*60*60 + secs, nil
}

// parseISO8601Time parses ISO 8601 time components, such as "1H2M3S".
func parseISO8601Time(s string) (int, error) {
	var total float64
	lastOrder := 4

	for s != "" {
		// Find the next unit.
		i := strings.IndexAny(s, "HMS")
		if i <= 0 {
			return 0, fmt.Errorf("invalid time component")
		}

		// Parse the number, allowing a decimal only on the final component.
		number := s[:i]
		dot := -1
		for j, r := range number {
			if r == '.' || r == ',' {
				if dot >= 0 || j == 0 || j == len(number)-1 {
					return 0, fmt.Errorf("invalid time number")
				}
				dot = j
			} else if r < '0' || r > '9' {
				return 0, fmt.Errorf("invalid time number")
			}
		}
		if dot >= 0 && i+1 < len(s) {
			return 0, fmt.Errorf("fraction must be on the final component")
		}

		n, err := strconv.ParseFloat(strings.ReplaceAll(number, ",", "."), 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, fmt.Errorf("invalid time number")
		}

		// Units must appear in descending order: hours, minutes, seconds.
		order, scale := 0, float64(1)
		switch s[i] {
		case 'H':
			order, scale = 3, 3600
		case 'M':
			order, scale = 2, 60
		case 'S':
			order, scale = 1, 1
		}
		if order >= lastOrder {
			return 0, fmt.Errorf("invalid time unit order")
		}

		total += n * scale

		// Ensure duration is in range.
		if math.IsInf(total, 0) || total >= math.Ldexp(1, strconv.IntSize-1) {
			return 0, fmt.Errorf("duration is out of range")
		}

		lastOrder = order
		s = s[i+1:]
	}

	return int(total), nil
}
