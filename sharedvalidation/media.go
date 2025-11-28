package sharedvalidation

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
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

// ValidateTranscodePreset validates the transcode preset string.
func ValidateTranscodePreset(q string) (string, error) {
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

// ValidateAccelTypeDeviceNode checks the entered driver directory is valid for the acceleration type (will NOT show as dir, do not use IsDir check).
func ValidateAccelTypeDeviceNode(g, nodePath string) (validNodePath string, err error) {
	if g == sharedconsts.AccelTypeAuto {
		return "", nil // No node path required.
	}

	// Check if node path is needed.
	if runtime.GOOS != "linux" {
		fmt.Fprintf(os.Stderr, "Non-linux systems do not need a device directory passed for HW acceleration.\n")
		return "", nil
	}

	// ---- LINUX SYSTEM ONLY ----

	// Ensure device node exists if required.
	if nodePath == "" {
		switch g {
		case sharedconsts.AccelTypeQSV,
			sharedconsts.AccelTypeVAAPI:
			return "", fmt.Errorf("acceleration type %q requires a device directory on Linux systems", g)
		default:
			return "", nil
		}
	}

	// Check device node.
	if _, err := os.Stat(nodePath); os.IsNotExist(err) {
		return "", fmt.Errorf("driver location %q does not appear to exist?", nodePath)
	}

	return nodePath, nil
}
