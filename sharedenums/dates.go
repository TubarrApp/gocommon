// Package sharedenums holds enumerated operation values shared by Metarr and Tubarr.
package sharedenums

import (
	"fmt"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// DateFormat holds the date format directive (e.g. yyyy-mm-dd).
type DateFormat int

// DateFormat definitions.
const (
	DateFmtSkip DateFormat = iota
	DateYyyyMmDd
	DateYyMmDd
	DateYyyyDdMm
	DateYyDdMm
	DateDdMmYyyy
	DateDdMmYy
	DateMmDdYyyy
	DateMmDdYy
	DateDdMm
	DateMmDd
)

// dateFormats maps the directive as written by users to its enum. Capital 'Y' means
// yyyy and lowercase 'y' means yy, so lookups are deliberately case-sensitive.
var dateFormats = map[string]DateFormat{
	"Ymd": DateYyyyMmDd,
	"ymd": DateYyMmDd,
	"Ydm": DateYyyyDdMm,
	"ydm": DateYyDdMm,
	"dmY": DateDdMmYyyy,
	"dmy": DateDdMmYy,
	"mdY": DateMmDdYyyy,
	"mdy": DateMmDdYy,
	"md":  DateMmDd,
	"dm":  DateDdMm,
}

// dateFormatStrings is the reverse of dateFormats, built once at init.
var dateFormatStrings = func() map[DateFormat]string {
	m := make(map[DateFormat]string, len(dateFormats))
	for s, e := range dateFormats {
		m[e] = s
	}
	return m
}()

// ValidDateFormats lists every accepted date format directive.
var ValidDateFormats = func() []string {
	s := make([]string, 0, len(dateFormats))
	for k := range dateFormats {
		s = append(s, k)
	}
	return s
}()

// ParseDateFormat returns the enum for a date format directive such as "ymd" or "dmY".
func ParseDateFormat(dateFmt string) (DateFormat, error) {
	if e, ok := dateFormats[dateFmt]; ok {
		return e, nil
	}
	return DateFmtSkip, fmt.Errorf("invalid date format %q, expected two or three of y/m/d (where 'Y' is yyyy and 'y' is yy), one of %v", dateFmt, ValidDateFormats)
}

// String returns the directive for d, or "" when d is [DateFmtSkip] or unknown.
func (d DateFormat) String() string {
	return dateFormatStrings[d]
}

// DateTagLocation determines where a date tag should be added in a string.
type DateTagLocation int

// DateTagLocation definitions.
const (
	DateTagLocPrefix DateTagLocation = iota
	DateTagLocSuffix
	DateTagLocAll
)

// ParseDateTagLocation returns the enum for a date tag location.
//
// allowAll permits "all", which suits delete-date-tag but not date-tag, since a tag
// can only be written to one place.
func ParseDateTagLocation(loc string, allowAll bool) (DateTagLocation, error) {
	switch loc {
	case sharedconsts.OpLocPrefix:
		return DateTagLocPrefix, nil
	case sharedconsts.OpLocSuffix:
		return DateTagLocSuffix, nil
	case sharedconsts.OpLocAll:
		if allowAll {
			return DateTagLocAll, nil
		}
		return DateTagLocPrefix, fmt.Errorf("date tag location %q is only valid for %s", loc, sharedconsts.OpDeleteDateTag)
	default:
		if allowAll {
			return DateTagLocPrefix, fmt.Errorf("invalid date tag location %q, expected %q, %q or %q", loc, sharedconsts.OpLocPrefix, sharedconsts.OpLocSuffix, sharedconsts.OpLocAll)
		}
		return DateTagLocPrefix, fmt.Errorf("invalid date tag location %q, expected %q or %q", loc, sharedconsts.OpLocPrefix, sharedconsts.OpLocSuffix)
	}
}

// String returns the location directive for l.
func (l DateTagLocation) String() string {
	switch l {
	case DateTagLocPrefix:
		return sharedconsts.OpLocPrefix
	case DateTagLocSuffix:
		return sharedconsts.OpLocSuffix
	case DateTagLocAll:
		return sharedconsts.OpLocAll
	default:
		return ""
	}
}
