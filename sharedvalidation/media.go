package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateMaxResolution returns a maximum resolution as its shorter side in pixels, e.g. "1080" from "1080p" or "2160" from "4k".
func ValidateMaxResolution(r string) (string, error) {
	r = strings.ToLower(strings.TrimSpace(r))
	if r == "" {
		return "", nil
	}
	if pixels, ok := sharedconsts.ResolutionAlias[r]; ok {
		return pixels, nil
	}

	// A number of pixels, with or without a trailing 'p'.
	pixels, err := strconv.Atoi(strings.TrimSuffix(r, "p"))
	if err != nil || pixels <= 0 {
		return "", fmt.Errorf("max resolution should be a size like 1080, 1080p, 4k or 8k, got %q", r)
	}
	return strconv.Itoa(pixels), nil
}

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
