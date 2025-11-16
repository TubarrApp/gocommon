// Package sharedconsts provides shared constants for terminal colors and other common values.
package sharedconsts

// ANSI color codes for terminal output.
const (
	ColorReset       = "\033[0m"
	ColorRed         = "\033[31m"
	ColorGreen       = "\033[32m"
	ColorYellow      = "\033[33m"
	ColorBlue        = "\033[34m"
	ColorPurple      = "\033[35m"
	ColorCyan        = "\033[36m"
	ColorDimCyan     = "\033[2;36m"
	ColorWhite       = "\033[37m"
	ColorBrightBlack = "\033[90m"
	ColorDimWhite    = "\033[2;37m"
)

// Log message prefixes with colors.
const (
	LogTagError   = ColorRed + "[ERROR] " + ColorReset
	LogTagSuccess = ColorGreen + "[SUCCESS] " + ColorReset
	LogTagDebug   = ColorYellow + "[DEBUG] " + ColorReset
	LogTagWarning = ColorYellow + "[WARNING] " + ColorReset
	LogTagInfo    = ColorBlue + "[INFO] " + ColorReset
)
