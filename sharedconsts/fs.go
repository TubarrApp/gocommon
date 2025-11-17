package sharedconsts

// Rename styles.
const (
	RenameSkip        = "skip"
	RenameSpaces      = "spaces"
	RenameUnderscores = "underscores"
	RenameFixesOnly   = "fixes-only"
)

var ValidRenameFlags = map[string]bool{
	RenameSkip:        true,
	RenameSpaces:      true,
	RenameUnderscores: true,
	RenameFixesOnly:   true,
}
