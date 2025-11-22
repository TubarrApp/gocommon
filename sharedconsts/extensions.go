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
var AllVidExtensions = map[string]bool{
	Ext3GP:  true,
	Ext3G2:  true,
	ExtASF:  true,
	ExtAVI:  true,
	ExtF4V:  true,
	ExtFLV:  true,
	ExtOGM:  true,
	ExtOGV:  true,
	ExtM4V:  true,
	ExtMKV:  true,
	ExtMOV:  true,
	ExtMP4:  true,
	ExtMPEG: true,
	ExtMPG:  true,
	ExtMTS:  true,
	ExtRM:   true,
	ExtRMVB: true,
	ExtTS:   true,
	ExtVOB:  true,
	ExtWEBM: true,
	ExtWMV:  true,
}

// AllMetaExtensions contains the list of meta extensions.
var AllMetaExtensions = map[string]bool{
	MExtJSON: true,
	MExtNFO:  true,
}
