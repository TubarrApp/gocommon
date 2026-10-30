package sharedvalidation

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// validateFromCodec returns a valid FromCodec model if valid.
//
// Pass in [ValidateVideoCodec] or [ValidateAudioCodec].
func validateFromCodec(c string, validate func(string) (string, error)) (sharedmodels.FromCodec, error) {
	blankModel := sharedmodels.NewFromCodec(sharedconsts.CodecCopy, sharedconsts.CompressionBlank)

	// Check if codec string is empty.
	if c == "" {
		return blankModel, nil
	}

	// Get codec and compression level.
	validCodec, compressionLevel, err := parseFromCodec(c, validate)
	if err != nil {
		return blankModel, err
	}

	// Return validated FromCodec.
	return sharedmodels.NewFromCodec(validCodec, compressionLevel), nil
}

// parseFromCodec retrieves the valid compression level and codec.
//
// Pass in [ValidateVideoCodec] or [ValidateAudioCodec].
func parseFromCodec(c string, validate func(string) (string, error)) (codec string, compressionLevel sharedconsts.CompressionLevel, err error) {
	// Normalize string.
	c = strings.ToLower(c)
	c = strings.TrimSpace(c)
	c = strings.ReplaceAll(c, ".", "")
	// Don't remove "_", is delimiter (e.g. 'mc_hevc')

	// Check compression level component.
	compressionLevel = sharedconsts.CompressionBlank
	compression, codec, found := strings.Cut(c, "_")
	if found {
		c = codec // Extracted codec portion (e.g. 'h264' from 'lc_h264').

		// Get compression level.
		validCompressionLevel := false
		if compressionLevel, validCompressionLevel = sharedconsts.ValidCompressionLevels[compression]; !validCompressionLevel {
			return c, "", fmt.Errorf("compression level %s not valid in codec %s", compression, c)
		}
	}

	// Validate codec component.
	validCodec, err := validate(c)
	if err != nil {
		return "", compressionLevel, err
	}

	// Return retrieved codec and compression level.
	return validCodec, compressionLevel, nil
}

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

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidVideoCodecs[c]; ok {
		return c, nil
	}

	return "", fmt.Errorf("video codec %q is not valid. Supported: %v", c, sortedKeys(sharedconsts.ValidVideoCodecs))
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

	// Synonym and alias mapping.
	if mapped, ok := sharedconsts.AudioCodecAlias[a]; ok {
		a = mapped
	}

	// Check against valid codec map.
	if _, ok := sharedconsts.ValidAudioCodecs[a]; ok {
		return a, nil
	}

	return "", fmt.Errorf("audio codec %q is not valid. Supported: %v", a, sortedKeys(sharedconsts.ValidAudioCodecs))
}

// ParseCodecMappings splits codec entries into one plain codec for everything else and per-source "from:to" pairs.
//
// Names are normalized by validate, and order doesn't matter: pairs always override the plain codec.
//
// Pass in [ValidateVideoCodec] or [ValidateAudioCodec].
func ParseCodecMappings(entries []string, validate func(string) (string, error)) (allTo string, pairs map[sharedmodels.FromCodec]string, err error) {
	pairs = make(map[sharedmodels.FromCodec]string, len(entries))

	for _, e := range entries {
		if e == "" {
			continue
		}

		// Split on ':' to separate 'from' and 'to' codec portion.
		from, to, isPair := strings.Cut(e, ":")
		if from == "" {
			if !isPair {
				return "", nil, fmt.Errorf("codec sent in empty (%q)", e)
			}
			return "", nil, fmt.Errorf("'from' codec sent in empty in string %s:%s", from, to)
		}

		// Grab 'FromCodec' model.
		fromCodec, err := validateFromCodec(from, validate)
		if err != nil {
			return "", nil, err
		}

		// Passed in validation function (video or audio codec name validator).
		if fromCodec.Codec, err = validate(fromCodec.Codec); err != nil {
			return "", nil, err
		}

		// Plain codec, only one allowed.
		if !isPair {
			if allTo != "" && fromCodec.Codec != "" && fromCodec.Codec != allTo {
				return "", nil, fmt.Errorf("more than one plain codec (%q and %q), keep only one", allTo, fromCodec.Codec)
			}
			if fromCodec.Codec != "" {
				allTo = fromCodec.Codec
			}
			continue
		}

		// Pair, each source named once.
		if fromCodec.Codec == "" {
			return "", nil, fmt.Errorf("codec pair %q has no source codec", e)
		}
		if to, err = validate(to); err != nil {
			return "", nil, err
		}
		if prev, ok := pairs[fromCodec]; ok && prev != to {
			return "", nil, fmt.Errorf("source codec %q is named more than once (%q and %q), keep only one", from, prev, to)
		}
		pairs[fromCodec] = to
	}
	return allTo, pairs, nil
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
	return "", fmt.Errorf("GPU acceleration type %q is not valid. Supported: %v", a, sortedKeys(sharedconsts.ValidGPUAccelTypes))
}

// OSSupportsAccelType verified OS support for this acceleration type.
func OSSupportsAccelType(a string) bool {
	OS := runtime.GOOS

	// AMD.
	if a == sharedconsts.AccelTypeAMF {
		return OS == "windows" // Only supported on Windows.
	}
	// VAAPI.
	if a == sharedconsts.AccelTypeVAAPI {
		return OS == "linux" // Only supported on Linux.
	}
	// Mac and not videotoolbox.
	if OS == "darwin" {
		return a == sharedconsts.AccelTypeVideoToolbox
	}

	// Return true by default.
	return true
}

// ValidateAccelTypeDeviceNode checks the entered driver directory is valid for the acceleration type (will NOT show as dir, do not use IsDir check).
func ValidateAccelTypeDeviceNode(g, nodePath string) (validNodePath string, err error) {
	// Ensure device node exists if required.
	if nodePath == "" {
		switch g {
		case sharedconsts.AccelTypeVAAPI:
			return "", fmt.Errorf("acceleration type %q requires a device directory on Linux systems", g)
		default: // Sent in empty.
			return nodePath, nil
		}
	}

	// Check device node.
	if _, err := os.Stat(nodePath); os.IsNotExist(err) {
		return "", fmt.Errorf("driver node path %q does not appear to exist?", nodePath)
	}

	return nodePath, nil
}
