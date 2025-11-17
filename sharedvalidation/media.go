package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateTranscodeQuality validates a transcode quality value (0-51 for x264/x265).
// Returns the normalized quality string.
func ValidateTranscodeQuality(q string) (string, error) {
	if q == "" {
		return "", nil
	}

	// Normalize input.
	q = strings.ReplaceAll(q, " ", "")

	// Validate integer.
	qNum, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return "", fmt.Errorf("transcode quality should be numerical (0-51), got %q", q)
	}

	// Clamp to valid range.
	qNum = min(max(qNum, 0), 51)

	return strconv.FormatInt(qNum, 10), nil
}
