package sharedvalidation

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateVideoCodec validates a video codec string and returns the normalized codec name.
//
// Handles common aliases and synonyms (e.g., "x264" -> "h264", "libx265" -> "hevc").
func ValidateVideoCodec(c string) (string, error) {
	if c == "" {
		return "", nil
	}

	// Normalize input.
	c = strings.TrimSpace(c)
	c = strings.ToLower(c)
	c = strings.ReplaceAll(c, ".", "")
	c = strings.ReplaceAll(c, "-", "")
	c = strings.ReplaceAll(c, "_", "")

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.VideoCodecAlias[c]; ok {
		c = mapped
	}

	// Blank, or an alias for no codec (e.g. "none").
	if c == "" {
		return "", nil
	}

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidVideoCodecs[c]; ok {
		return c, nil
	}

	return "", fmt.Errorf("video codec %q is not valid. Supported: %v", c, sortedKeys(sharedconsts.ValidVideoCodecs))
}

// ValidateVideoCodecWithAccel validates a video codec string with GPU acceleration context.
//
// Returns an error if a non-auto or blank acceleration type is provided without a valid codec.
// (GPU acceleration uses flags like nvenc_h264).
func ValidateVideoCodecWithAccel(c, accelType string) (validVideoCodec string, err error) {
	// Ensure valid codec.
	if c, err = ValidateVideoCodec(c); err != nil {
		return "", err
	}

	// Check if empty codec is allowed with the given acceleration type.
	if c == "" &&
		(accelType != "" && accelType != sharedconsts.AccelTypeAuto) {
		return "", fmt.Errorf("GPU acceleration type %q requires a codec", accelType)
	}

	return c, nil
}

// ValidateAudioCodec validates an audio codec string and returns the normalized codec name.
//
// Handles common aliases and synonyms (e.g., "mp3" -> "mp3", "libmp3lame" -> "mp3").
func ValidateAudioCodec(a string) (validAudioCodec string, err error) {
	if a == "" {
		return "", nil
	}

	// Normalize input.
	a = strings.TrimSpace(a)
	a = strings.ToLower(a)
	a = strings.ReplaceAll(a, ".", "")
	a = strings.ReplaceAll(a, "-", "")
	a = strings.ReplaceAll(a, "_", "")

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.AudioCodecAlias[a]; ok {
		a = mapped
	}

	// Blank, or an alias for no codec (e.g. "none").
	if a == "" {
		return "", nil
	}

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidAudioCodecs[a]; ok {
		return a, nil
	}

	return "", fmt.Errorf("audio codec %q is not valid. Supported: %v", a, sortedKeys(sharedconsts.ValidAudioCodecs))
}
