package sharedconsts

// MaxNameBytes is the longest file or directory name most filesystems allow.
const MaxNameBytes = 255

// Filename tags.
const (
	BackupTag = "_mbackup"
	TempTag   = "tmp_"
)

// Rename styles.
const (
	RenameFixesOnly   = "fixes-only"
	RenameSkip        = "skip"
	RenameSpaces      = "spaces"
	RenameUnderscores = "underscores"
)

// ValidRenameFlags are the accepted rename styles.
var ValidRenameFlags = map[string]struct{}{
	RenameFixesOnly:   {},
	RenameSkip:        {},
	RenameSpaces:      {},
	RenameUnderscores: {},
}

// RenameAlias maps alternative spellings to a rename style.
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

// Accepted uploaded image extensions.
const (
	ITypeGIF  = "gif"
	ITypeJPG  = "jpg"
	ITypeJPEG = "jpeg"
	ITypePNG  = "png"
)

// ImageTypes holds a lookup map for accepted image types.
var ImageUploadTypes = map[string]bool{
	ITypeGIF:  true,
	ITypeJPG:  true,
	ITypeJPEG: true,
	ITypePNG:  true,
}

// Maximums for uploaded images.
const (
	ImageUploadMaxBytes  = 10 << 20 // 10 MiB.
	ImageUploadMaxPixels = 40_000_000
)
