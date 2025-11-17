package sharedconsts

// Rename styles.
const (
	RenameFixesOnly   = "fixes-only"
	RenameSkip        = "skip"
	RenameSpaces      = "spaces"
	RenameUnderscores = "underscores"
)

var ValidRenameFlags = map[string]bool{
	RenameFixesOnly:   true,
	RenameSkip:        true,
	RenameSpaces:      true,
	RenameUnderscores: true,
}
