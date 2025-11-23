package sharedconsts

// Rename styles.
const (
	RenameFixesOnly   = "fixes-only"
	RenameSkip        = "skip"
	RenameSpaces      = "spaces"
	RenameUnderscores = "underscores"
)

var ValidRenameFlags = map[string]struct{}{
	RenameFixesOnly:   {},
	RenameSkip:        {},
	RenameSpaces:      {},
	RenameUnderscores: {},
}

var RenameAlias = map[string]string{
	// Skip.
	"none":     RenameSkip,
	"skipped":  RenameSkip,
	"skipping": RenameSkip,
	"skips":    RenameSkip,

	// Fixes only.
	"fix":       RenameFixesOnly,
	"fixed":     RenameFixesOnly,
	"fixes":     RenameFixesOnly,
	"fixesonly": RenameFixesOnly,

	// Spaces.
	"space":   RenameSpaces,
	"spaced":  RenameSpaces,
	"spacing": RenameSpaces,

	"underscore":   RenameUnderscores,
	"underscored":  RenameUnderscores,
	"underscoring": RenameUnderscores,
}
