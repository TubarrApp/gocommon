package sharedconsts

// Video extensions.
const (
	Ext3GP  = ".3gp"
	Ext3G2  = ".3g2"
	ExtASF  = ".asf"
	ExtAVI  = ".avi"
	ExtF4V  = ".f4v"
	ExtFLV  = ".flv"
	ExtOGM  = ".ogm"
	ExtOGV  = ".ogv"
	ExtM4V  = ".m4v"
	ExtMKV  = ".mkv"
	ExtMOV  = ".mov"
	ExtMP4  = ".mp4"
	ExtMPEG = ".mpeg"
	ExtMPG  = ".mpg"
	ExtMTS  = ".mts"
	ExtRM   = ".rm"
	ExtRMVB = ".rmvb"
	ExtTS   = ".ts"
	ExtVOB  = ".vob"
	ExtWEBM = ".webm"
	ExtWMV  = ".wmv"
)

// Metafile extensions.
const (
	MExtJSON = ".json"
	MExtNFO  = ".nfo"
)

// AllVidExtensions is a list of video file extensions.
var AllVidExtensions = map[string]struct{}{
	Ext3GP:  {},
	Ext3G2:  {},
	ExtASF:  {},
	ExtAVI:  {},
	ExtF4V:  {},
	ExtFLV:  {},
	ExtOGM:  {},
	ExtOGV:  {},
	ExtM4V:  {},
	ExtMKV:  {},
	ExtMOV:  {},
	ExtMP4:  {},
	ExtMPEG: {},
	ExtMPG:  {},
	ExtMTS:  {},
	ExtRM:   {},
	ExtRMVB: {},
	ExtTS:   {},
	ExtVOB:  {},
	ExtWEBM: {},
	ExtWMV:  {},
}

// AllMetaExtensions contains the list of meta extensions.
var AllMetaExtensions = map[string]struct{}{
	MExtJSON: {},
	MExtNFO:  {},
}

// FilterByVidExtensions is a list of found video file extensions.
// Set true if user wants to work on files of this type.
var FilterByVidExtensions = map[string]bool{
	Ext3GP:  false,
	Ext3G2:  false,
	ExtASF:  false,
	ExtAVI:  false,
	ExtF4V:  false,
	ExtFLV:  false,
	ExtOGM:  false,
	ExtOGV:  false,
	ExtM4V:  false,
	ExtMKV:  false,
	ExtMOV:  false,
	ExtMP4:  false,
	ExtMPEG: false,
	ExtMPG:  false,
	ExtMTS:  false,
	ExtRM:   false,
	ExtRMVB: false,
	ExtTS:   false,
	ExtVOB:  false,
	ExtWEBM: false,
	ExtWMV:  false,
}

// FilterByMetaExtension is a list of found meta file extensions.
// Set true if user wants to work on files of this type.
var FilterByMetaExtension = map[string]bool{
	MExtJSON: false,
	MExtNFO:  false,
}
