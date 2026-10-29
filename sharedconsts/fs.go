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
	IUploadGIF  = "gif"
	IUploadJPG  = "jpg"
	IUploadJPEG = "jpeg"
	IUploadPNG  = "png"
)

// ImageUploadTypes holds a lookup map for accepted image types.
var ImageUploadTypes = map[string]bool{
	IUploadGIF:  true,
	IUploadJPG:  true,
	IUploadJPEG: true,
	IUploadPNG:  true,
}
