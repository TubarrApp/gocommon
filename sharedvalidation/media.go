package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateMaxFilesize validates a max filesize string (e.g., "100M", "2G", "500K").
// Returns normalized form with lowercase suffix (e.g., "100m", "2g").
func ValidateMaxFilesize(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	s := strings.ToLower(strings.TrimSpace(input))

	// Strip trailing 'b' if present (e.g., "100MB" -> "100m")
	s = strings.TrimSuffix(s, "b")

	// Handle K, M, G suffixes
	if len(s) > 0 {
		suffix := s[len(s)-1]
		if suffix == 'k' || suffix == 'm' || suffix == 'g' {
			n := s[:len(s)-1]
			if _, err := strconv.ParseFloat(n, 64); err != nil {
				return "", fmt.Errorf("invalid size number %q: %w", s, err)
			}
			return s, nil
		}
	}

	// Check raw integer is valid
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s, nil
	}

	return "", fmt.Errorf("invalid max filesize format %q (use format like '100M', '2G', '500K')", input)
}

// ValidateTranscodeQuality validates a transcode quality value (0-51 for x264/x265).
// Returns the normalized quality string.
func ValidateTranscodeQuality(q string) (string, error) {
	if q == "" {
		return "", nil
	}

	q = strings.TrimSpace(q)
	q = strings.ReplaceAll(q, " ", "")

	qNum, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return "", fmt.Errorf("transcode quality should be numerical (0-51), got %q", q)
	}

	// Clamp to valid range
	qNum = min(max(qNum, 0), 51)

	return strconv.FormatInt(qNum, 10), nil
}

// ValidateExtension validates and normalizes a file extension.
// Ensures it has a dot prefix and is not empty.
func ValidateExtension(ext string) string {
	ext = strings.TrimSpace(ext)

	// Handle empty or invalid cases
	if ext == "" || ext == "." {
		return ""
	}

	// Ensure proper dot prefix
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// Verify the extension is not just a lone dot
	if len(ext) <= 1 {
		return ""
	}

	return ext
}

// ValidateConcurrencyLimit validates a concurrency limit, returning at least 1.
func ValidateConcurrencyLimit(c int) int {
	return max(c, 1)
}

// ValidateMinFreeMem validates a minimum free memory string (e.g., "2G", "500M").
// Returns the normalized form or error.
func ValidateMinFreeMem(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	s := strings.ToUpper(strings.TrimSpace(input))

	// Must end with G or M
	if !strings.HasSuffix(s, "G") && !strings.HasSuffix(s, "M") {
		return "", fmt.Errorf("minimum free memory should end with G or M (e.g., '2G', '500M'), got %q", input)
	}

	// Validate the number portion
	numPart := s[:len(s)-1]
	if _, err := strconv.ParseFloat(numPart, 64); err != nil {
		return "", fmt.Errorf("invalid memory size number %q: %w", input, err)
	}

	return s, nil
}

// ValidateMaxCPU validates a max CPU percentage (0.0 to 100.0).
// Returns the clamped value.
func ValidateMaxCPU(maxCPU float64) float64 {
	return min(max(maxCPU, 100.0), 0.0)
}
