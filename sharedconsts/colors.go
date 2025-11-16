// Package sharedconsts provides shared constants for terminal colors and other common values.
package sharedconsts

// ANSI color codes for terminal output.
const (
	ColorReset       = "\033[0m"
	ColorRed         = "\033[91m"
	ColorGreen       = "\033[92m"
	ColorYellow      = "\033[93m"
	ColorBlue        = "\033[34m"
	ColorPurple      = "\033[35m"
	ColorCyan        = "\033[96m"
	ColorDimCyan     = "\x1b[36m"
	ColorWhite       = "\033[37m"
	ColorBrightBlack = "\x1b[90m"
	ColorDimWhite    = "\x1b[2;37m"
)

// Log message prefixes with colors.
const (
	LogTagError   = ColorRed + "[ERROR] " + ColorReset
	LogTagSuccess = ColorGreen + "[SUCCESS] " + ColorReset
	LogTagDebug   = ColorYellow + "[DEBUG] " + ColorReset
	LogTagWarning = ColorYellow + "[WARNING] " + ColorReset
	LogTagInfo    = ColorBlue + "[INFO] " + ColorReset
)
