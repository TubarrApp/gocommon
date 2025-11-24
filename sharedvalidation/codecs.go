package sharedvalidation

import (
	"fmt"
	"runtime"
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
	c = strings.ToLower(strings.TrimSpace(c))
	c = strings.ReplaceAll(c, ".", "")
	c = strings.ReplaceAll(c, "-", "")
	c = strings.ReplaceAll(c, "_", "")

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.VideoCodecAlias[c]; ok {
		c = mapped
	}

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidVideoCodecs[c]; ok {
		return c, nil
	}

	return "", fmt.Errorf("video codec %q is not valid. Supported: %v", c, sharedconsts.ValidVideoCodecs)
}

// ValidateVideoCodecWithAccel validates a video codec string with GPU acceleration context.
//
// Returns an error if auto/none codec is used with specific GPU acceleration.
func ValidateVideoCodecWithAccel(c, accelType string) (validVideoCodec string, err error) {
	// Ensure valid codec.
	if c, err = ValidateVideoCodec(c); err != nil {
		return "", err
	}

	// Check if empty codec is allowed with the given acceleration type.
	if c == "" &&
		(accelType != "" && accelType != sharedconsts.AccelTypeAuto) {
		return "", fmt.Errorf("GPU acceleration type %q requires a codec (entered %q)", accelType, c)
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
	a = strings.ToLower(strings.TrimSpace(a))
	a = strings.ReplaceAll(a, ".", "")
	a = strings.ReplaceAll(a, "-", "")
	a = strings.ReplaceAll(a, "_", "")

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.AudioCodecAlias[a]; ok {
		a = mapped
	}

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidAudioCodecs[a]; ok {
		return a, nil
	}

	return "", fmt.Errorf("audio codec %q is not valid. Supported: %v", a, sharedconsts.ValidAudioCodecs)
}

// ValidateGPUAccelType validates a GPU acceleration type string.
func ValidateGPUAccelType(a string) (validAccelType string, err error) {
	// Normalize input.
	a = strings.ToLower(strings.TrimSpace(a))

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.AccelAlias[a]; ok {
		a = mapped
	}

	// Check against valid acceleration type map.
	if _, ok := sharedconsts.ValidGPUAccelTypes[a]; ok {
		return a, nil
	}

	// Return error on map check failure.
	return "", fmt.Errorf("%s GPU acceleration type %q is not valid. Supported: %v", sharedconsts.LogTagError, a, sharedconsts.ValidGPUAccelTypes)
}

// OSSupportsAccelType verified OS support for this acceleration type.
func OSSupportsAccelType(a string) bool {
	// Get OS.
	OS := runtime.GOOS

	// AMD.
	if a == sharedconsts.AccelTypeAMF {
		return OS == "windows" // Only supported on Windows.
	}
	return true
}
