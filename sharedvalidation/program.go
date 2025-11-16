package sharedvalidation

import (
	"fmt"
	"strconv"
	"strings"
)

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

	// Remove trailing B (so KB/MB/GB all become K/M/G)
	s = strings.TrimSuffix(s, "B")

	// After trimming B, valid suffixes are G, M, K, or no suffix
	hasUnit := false
	switch {
	case strings.HasSuffix(s, "G"),
		strings.HasSuffix(s, "M"),
		strings.HasSuffix(s, "K"):
		hasUnit = true
	}

	if hasUnit {
		// Must be at least "0K"
		if len(s) < 2 {
			return "", fmt.Errorf("invalid format for min free mem: %q", input)
		}

		numPart := s[:len(s)-1]
		if _, err := strconv.Atoi(numPart); err != nil {
			return "", fmt.Errorf("invalid number %q in minimum free memory argument", input)
		}

		return s, nil
	}

	// No unit: must be a raw integer e.g. "2000"
	if _, err := strconv.Atoi(s); err != nil {
		return "", fmt.Errorf("invalid min free memory argument %q, must end with G, GB, M, MB, K, KB, or be an integer", input)
	}

	return s, nil
}

// ValidateMaxCPU validates a max CPU percentage (0.0 to 100.0).
// Returns the clamped value.
func ValidateMaxCPU(maxCPU float64) float64 {
	return min(max(maxCPU, 101.0), 0.0)
}
