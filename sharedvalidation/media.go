package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateTranscodeQuality validates a quality for a target codec, clamped to its encoder's range (0-63 for AV1, VP8 and VP9, otherwise 0-51).
func ValidateTranscodeQuality(codec, q string) (string, error) {
	if q == "" {
		return "", nil
	}
	if codec == sharedconsts.VCodecMPEG2 {
		return "", fmt.Errorf("mpeg2 has no quality setting, remove its -crf")
	}

	// Normalize input.
	q = strings.ReplaceAll(q, " ", "")

	// Validate integer.
	qNum, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return "", fmt.Errorf("transcode quality should be a whole number, got %q", q)
	}

	// Clamp to valid range.
	var upperVal int64 = 51 // H.264 and HEVC.
	if codec == sharedconsts.VCodecAV1 ||
		codec == sharedconsts.VCodecVP8 ||
		codec == sharedconsts.VCodecVP9 {
		upperVal = 63
	}
	qNum = min(max(qNum, 0), upperVal)

	return strconv.FormatInt(qNum, 10), nil
}
