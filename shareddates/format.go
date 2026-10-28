// Package shareddates performs operations related to parsing and handling dates.
package shareddates

import (
	"fmt"
	"slices"
	"strings"

	"github.com/TubarrApp/gocommon/logging"
	"github.com/TubarrApp/gocommon/sharedenums"
	"github.com/araddon/dateparse"
)

// ParseWordDate parses and formats the inputted word date (e.g. Jan 2nd, 2006).
func ParseWordDate(dateString string) (string, error) {
	t, err := dateparse.ParseAny(dateString)
	if err != nil {
		return "", fmt.Errorf("unable to parse date %q to word date: %w", dateString, err)
	}
	return t.Format("2006-01-02"), nil
}

// ParseNumDate parses and formats the inputted numerical date string.
func ParseNumDate(dateNum string, pl *logging.ProgramLogger) (string, error) {
	t, err := dateparse.ParseAny(dateNum)
	if err != nil {
		return "", fmt.Errorf("unable to parse date %q to word date: %w", dateNum, err)
	}

	// Get day, month, year.
	date := t.Format("01022006") // t.Format("01022006") always returns length of 8.
	day := date[2:4]
	month := date[:2]
	year := date[4:]

	// Get 'January 1st' type entries.
	indexOfMonth := slices.Index(monthNumStringSlice[:], month) // Find index in slice (e.g. the index where '02' appears).
	month = monthWordFullSlice[indexOfMonth]                    // Grab value at the same index in the alternate month format slice.
	day = addDaySuffixes(day, pl)

	// Write date string (e.g. 'January 1st, 2001').
	dateStr := fmt.Sprintf("%s %s, %s", month, day, year)
	pl.S("Made string %q from date input %q", dateStr, dateNum)

	return dateStr, nil
}

// YmdFromMeta converts inputted date strings into the user's defined format.
func YmdFromMeta(date string, pl *logging.ProgramLogger) (t string, madeDate bool) {
	if date == "" {
		pl.D(1, "Date sent in empty (%q), not formatting date element.", date)
		return date, false
	}

	// Get date and T00:00:00Z components.
	beforeT, afterT, containsT := strings.Cut(date, "T")
	if containsT {
		t = "T" + afterT
	}

	// Remove hyphens ('yy-mm-dd' -> 'yymmdd').
	beforeT = strings.ReplaceAll(beforeT, "-", "")

	// Return formatted date, or original if malformed.
	switch len(beforeT) {
	case 8:
		formatted := beforeT[:4] + "-" + beforeT[4:6] + "-" + beforeT[6:] + t
		pl.D(3, "Made YYYY-MM-DD date %s", formatted)
		return formatted, true
	case 6:
		formatted := beforeT[:2] + "-" + beforeT[2:4] + "-" + beforeT[4:] + t
		pl.D(3, "Made YY-MM-DD date %s", formatted)
		return formatted, true
	default:
		pl.D(3, "Returning malformed date element (%s) without formatting.", date)
		return date, false
	}
}

// FormatDateString formats the date as a hyphenated string.
func FormatDateString(year, month, day string, dateFmt sharedenums.DateFormat) string {
	var parts []string

	switch dateFmt {
	case sharedenums.DateYyyyMmDd, sharedenums.DateYyMmDd:
		parts = []string{year, month, day}
	case sharedenums.DateYyyyDdMm, sharedenums.DateYyDdMm:
		parts = []string{year, day, month}
	case sharedenums.DateDdMmYyyy, sharedenums.DateDdMmYy:
		parts = []string{day, month, year}
	case sharedenums.DateMmDdYyyy, sharedenums.DateMmDdYy:
		parts = []string{month, day, year}
	case sharedenums.DateMmDd:
		parts = []string{month, day}
	case sharedenums.DateDdMm:
		parts = []string{day, month}
	}
	return joinNonEmptyWithHyphens(parts[:])
}
