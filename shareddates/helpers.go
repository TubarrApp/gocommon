package shareddates

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TubarrApp/gocommon/logging"
	"github.com/TubarrApp/gocommon/sharedenums"
)

// Month values.
var (
	monthNumSlice       = [12]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	monthNumStringSlice = [12]string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}
	monthWordFullSlice  = [12]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	monthWordShortSlice = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
)

// ParseDateComponents extracts and validates year, month, and day from the date string.
func ParseDateComponents(date string, dateFmt sharedenums.DateFormat, pl *logging.ProgramLogger) (year, month, day string, err error) {
	date = strings.ReplaceAll(date, "-", "")
	date = strings.TrimSpace(date)

	year, month, day, err = getYearMonthDay(date, dateFmt, pl)
	if err != nil {
		return "", "", "", err
	}

	return validateDateComponents(year, month, day)
}

// addDaySuffixes appends a numerical day with the appropriate suffix (e.g. '1st', '2nd', '3rd').
func addDaySuffixes(day string, pl *logging.ProgramLogger) string {
	num, err := strconv.Atoi(day)
	if err != nil {
		pl.E("Failed to convert date string to number")
		return day
	}

	// 11 and all 'teens.
	if num > 10 && num < 20 {
		return day + "th"
	}

	// All others.
	switch num % 10 {
	case 1:
		return day + "st"
	case 2:
		return day + "nd"
	case 3:
		return day + "rd"
	default:
		return day + "th"
	}
}

// joinNonEmptyWithHyphens joins non-empty strings from an array with hyphens.
func joinNonEmptyWithHyphens(parts [3]string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	return strings.Join(nonEmpty, "-")
}

// getYearMonthDay returns the year, month, and day digits from the date string.
func getYearMonthDay(d string, dateFmt sharedenums.DateFormat, pl *logging.ProgramLogger) (year, month, day string, err error) {
	d = strings.ReplaceAll(d, "-", "")
	d = strings.TrimSpace(d)

	// 'yyyymmdd' (with or without T00:00:00Z).
	if len(d) >= 8 {
		switch dateFmt {
		case sharedenums.DateDdMmYy, sharedenums.DateMmDdYy, sharedenums.DateYyDdMm, sharedenums.DateYyMmDd:
			year = d[2:4]
		default:
			year = d[:4]
		}
		month = d[4:6]
		day = d[6:8]

		return year, month, day, nil
	}

	// 'yymmdd'.
	if len(d) >= 6 {
		year = d[:2]
		month = d[2:4]
		day = d[4:6]

		return year, month, day, nil
	}

	// 'yyyy', 'mmdd', 'ddmm', etc.
	if len(d) == 4 { // Guess year, mmdd, or ddmm.
		return guessYearDayMonth(d, dateFmt, pl)
	}
	return "", "", "", fmt.Errorf("failed to parse year, month, and day from %q", d)
}

// validateDateComponents attempts to fix faulty date arrangements.
func validateDateComponents(year, month, day string) (y, m, d string, err error) {
	if isValidMonthNum(month) && isValidDayNum(day, month, year) {
		return year, month, day, nil
	}

	// Attempt swapping day and month.
	if isValidMonthNum(day) && isValidDayNum(month, day, year) {
		return year, day, month, nil
	}

	// Fail check:
	return "", "", "", fmt.Errorf("invalid date components: year=%s, month=%s, day=%s", year, month, day)
}

// isValidMonthNum checks if the month inputted is a valid month.
func isValidMonthNum(month string) bool {
	m, err := strconv.Atoi(month)
	if err != nil {
		return false
	}
	return m >= 1 && m <= 12
}

// isValidDayNum checks if the day inputted is a valid day.
func isValidDayNum(day, month, year string) bool {
	// Ensure numerical values.
	d, err := strconv.Atoi(day)
	if err != nil {
		return false
	}
	m, err := strconv.Atoi(month)
	if err != nil {
		return false
	}
	y, err := strconv.Atoi(year)
	if err != nil && year != "" {
		return false
	}

	// Valid day range.
	if d < 1 || d > 31 {
		return false
	}

	// Months with 30 days.
	if m == 4 || m == 6 || m == 9 || m == 11 {
		return d <= 30
	}

	// February.
	if m == 2 {
		if year == "" {
			return d <= 29
		}

		// Leap year check.
		isLeap := y%4 == 0 && (y%100 != 0 || y%400 == 0)
		if isLeap {
			return d <= 29
		}
		return d <= 28
	}

	// Valid day range wasn't February or a 30 day month.
	return true
}

// guessYearDayMonth tries to guess the meaning of the date string.
func guessYearDayMonth(d string, dateFmt sharedenums.DateFormat, pl *logging.ProgramLogger) (year, month, day string, err error) {
	// Get xxxx, xx__, and __xx.
	full, err := strconv.Atoi(d)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid date string %q threw error: %w", d, err)
	}

	i := full / 100 // 1999/100 = 19(.00) -> 19 (Go discards after decimal, no rounding)
	j := full % 100 // 1999/100 = 19 hundreds, 99 remainder -> 99

	currentYear := time.Now().Year()

	// Prefer a year when month and day would be impossible (whether it was 'mmdd' or 'ddmm').
	if (i == 19 || i == 20) && j > 12 && full <= currentYear {
		pl.I("Guessing date string %q as year", d)
		return guessedYear(d, dateFmt), "", "", nil
	}

	// Guess 'mmdd' first (US convention), then 'ddmm'.
	switch {
	case isValidMonthNum(d[:2]) && isValidDayNum(d[2:4], d[:2], ""):
		pl.I("Guessing date string %q as month-day", d)
		return "", d[:2], d[2:4], nil

	case isValidMonthNum(d[2:4]) && isValidDayNum(d[:2], d[2:4], ""):
		pl.I("Guessing date string %q as day-month", d)
		return "", d[2:4], d[:2], nil
	}

	// Last guess, ancient year (e.g. "Passage de Vénus", "View from the Window at Le Gras", ancient audio)
	// or possible near-future years.
	if full >= 1826 && full <= currentYear+5 {
		pl.I("Guessing date string %q as year", d)
		return guessedYear(d, dateFmt), "", "", nil
	}

	return "", "", "", fmt.Errorf("failed to guess year, month, and day from %q", d)
}

// guessedYear returns the year using the width required by the date format.
func guessedYear(d string, dateFmt sharedenums.DateFormat) string {
	switch dateFmt {
	case sharedenums.DateDdMmYy, sharedenums.DateMmDdYy, sharedenums.DateYyDdMm, sharedenums.DateYyMmDd:
		return d[2:4]
	default:
		return d[:4]
	}
}

// maxDaysInMonth returns the number of days in a given month. If year is entred zero, allows 29 for February.
func maxDaysInMonth(month, yearOrZero int) int {
	switch month {
	case 2: // February
		if yearOrZero == 0 || yearOrZero%400 == 0 || (yearOrZero%4 == 0 && yearOrZero%100 != 0) {
			return 29
		}
		return 28

	case 4, 6, 9, 11: // Apr, Jun, Sep, Nov
		return 30

	default: // All other months
		return 31
	}
}
