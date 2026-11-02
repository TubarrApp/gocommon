package sharedvalidation

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

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

	// Mac has nothing else, though "auto" can still pick VideoToolbox.
	if OS == "darwin" {
		return a == sharedconsts.AccelTypeAuto || a == sharedconsts.AccelTypeVideoToolbox
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

// ValidateConcurrencyLimit validates a concurrency limit, returning at least 1.
func ValidateConcurrencyLimit(c int) int {
	return max(c, 1)
}

// ValidateMinFreeMem verifies that a minimum free memory string is valid.
// Accepts formats like: "2G", "2GB", "500M", "500MB", "200K", "200KB", "2000" (plain int).
func ValidateMinFreeMem(input string) (num string, err error) {
	if input == "" {
		return "", nil
	}

	s := strings.ToUpper(strings.TrimSpace(input))

	// Remove trailing B (so KB/MB/GB all become K/M/G).
	s = strings.TrimSuffix(s, "B")

	// After trimming B, valid suffixes are G, M, K, or no suffix.
	hasUnit := false
	switch {
	case strings.HasSuffix(s, "G"),
		strings.HasSuffix(s, "M"),
		strings.HasSuffix(s, "K"):
		hasUnit = true
	}

	if hasUnit {
		// Must be at least "0K".
		if len(s) < 2 {
			return "", fmt.Errorf("invalid format for min free mem: %q", input)
		}

		numPart := s[:len(s)-1]
		if _, err := strconv.ParseUint(numPart, 10, 64); err != nil {
			return "", fmt.Errorf("invalid number %q in minimum free memory argument", input)
		}

		return s, nil
	}

	// No unit: must be a raw integer e.g. "2000".
	if _, err := strconv.ParseUint(s, 10, 64); err != nil {
		return "", fmt.Errorf("invalid min free memory argument %q, must end with G, GB, M, MB, K, KB, or be an integer", input)
	}

	return s, nil
}

// ValidateMaxCPU validates a max CPU percentage (0.0 to 100.0).
// Returns the clamped value.
func ValidateMaxCPU(maxCPU float64, allowZero bool) float64 {
	// Handle zero value.
	if maxCPU == 0.0 {
		if allowZero {
			return 0.0
		}
		return 101.0
	}

	// Handle 100.0 case (set to 101 to bypass CPU limitation).
	if maxCPU == 100.0 {
		return 101.0
	}

	// Clamp non-zero value between accepted bounds.
	return min(max(maxCPU, 5.0), 101.0)
}

// ValidateFFmpegOutputExt validates that the output extension is correct.
func ValidateFFmpegOutputExt(o string) (string, error) {
	o = strings.TrimSpace(o)
	o = strings.ToLower(o)
	if !strings.HasPrefix(o, ".") {
		o = "." + o
	}

	// Check valid video extension.
	if _, ok := sharedconsts.AllVidExtensions[o]; !ok {
		return "", fmt.Errorf("output filetype %q is not supported", o)
	}

	// Some containers ffmpeg can read, but not write.
	if v, ok := sharedconsts.ContainerCompatibleVideoCodecs[o]; ok && len(v) == 0 && len(sharedconsts.ContainerCompatibleAudioCodecs[o]) == 0 {
		return "", fmt.Errorf("output filetype %q is not supported, as ffmpeg can't write it", o)
	}
	return o, nil
}
