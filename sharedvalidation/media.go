package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateTranscodeQuality validates a transcode quality value (0-51 for x264/x265).
// Returns the normalized quality string.
func ValidateTranscodeQuality(q string) (string, int64, error) {
	if q == "" {
		return "", 0, nil
	}

	q = strings.TrimSpace(q)
	q = strings.ReplaceAll(q, " ", "")

	qNum, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("transcode quality should be numerical (0-51), got %q", q)
	}

	return strconv.FormatInt(qNum, 10), min(max(qNum, 0), 51), nil
}
