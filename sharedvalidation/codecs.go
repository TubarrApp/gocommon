package sharedvalidation

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateVideoCodec validates a video codec string and returns the normalized codec name.
// Handles common aliases and synonyms (e.g., "x264" -> "h264", "libx265" -> "hevc").
func ValidateVideoCodec(c string) (string, error) {
	if c == "" {
		return "", nil
	}

	// Normalize input.
	c = strings.ToLower(strings.TrimSpace(c))
	c = strings.ReplaceAll(c, ".", "")
	c = strings.ReplaceAll(c, "-", "")
	c = strings.ReplaceAll(c, "_", "")

	// Synonym and alias mapping.
	if mapped, exists := sharedconsts.VideoCodecAlias[c]; exists {
		c = mapped
	}

	// Check against valid codec map.
	if sharedconsts.ValidVideoCodecs[c] {
		return c, nil
	}

	return "", fmt.Errorf("video codec %q is not valid. Supported: %v", c, sharedconsts.ValidVideoCodecs)
}

// ValidateVideoCodecWithAccel validates a video codec string with GPU acceleration context.
// Returns an error if auto/none codec is used with specific GPU acceleration.
func ValidateVideoCodecWithAccel(c, accel string) (validCodec string, err error) {
	// Ensure valid codec.
	validated, err := ValidateVideoCodec(c)
	if err != nil {
		return "", err
	}

	// Check if empty codec is allowed with the given acceleration type.
	if validated == "" &&
		(accel != "" && accel != sharedconsts.AccelTypeAuto) {
		return "", fmt.Errorf("GPU acceleration type %q requires a codec (entered %q)", accel, c)
	}

	return validated, nil
}

// ValidateAudioCodec validates an audio codec string and returns the normalized codec name.
// Handles common aliases and synonyms (e.g., "mp3" -> "mp3", "libmp3lame" -> "mp3").
func ValidateAudioCodec(a string) (string, error) {
	if a == "" {
		return "", nil
	}

	// Normalize input.
	a = strings.ToLower(strings.TrimSpace(a))
	a = strings.ReplaceAll(a, ".", "")
	a = strings.ReplaceAll(a, "-", "")
	a = strings.ReplaceAll(a, "_", "")

	// Synonym and alias mapping.
	if mapped, exists := sharedconsts.AudioCodecAlias[a]; exists {
		a = mapped
	}

	// Check against valid codec map.
	if sharedconsts.ValidAudioCodecs[a] {
		return a, nil
	}

	return "", fmt.Errorf("audio codec %q is not valid. Supported: %v", a, sharedconsts.ValidAudioCodecs)
}

// ValidateGPUAccelType validates a GPU acceleration type string.
func ValidateGPUAccelType(accel string) (string, error) {
	// Normalize input.
	accel = strings.ToLower(strings.TrimSpace(accel))

	// Synonym and alias mapping.
	if mapped, exists := sharedconsts.AccelAlias[accel]; exists {
		accel = mapped
	}

	// Check against valid acceleration type map.
	if sharedconsts.ValidGPUAccelTypes[accel] {
		return accel, nil
	}

	// Return error on map check failure.
	return "", fmt.Errorf("%s GPU acceleration type %q is not valid. Supported: %v", sharedconsts.LogTagError, accel, sharedconsts.ValidGPUAccelTypes)
}
