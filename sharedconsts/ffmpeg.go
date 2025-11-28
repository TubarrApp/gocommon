package sharedconsts

// Video codec constants.
const (
	VCodecCopy  = "copy"
	VCodecAV1   = "av1"
	VCodecH264  = "h264"
	VCodecHEVC  = "hevc"
	VCodecMPEG2 = "mpeg2"
	VCodecVP8   = "vp8"
	VCodecVP9   = "vp9"
)

// Audio codec constants.
const (
	ACodecCopy   = "copy"
	ACodecAAC    = "aac"
	ACodecAC3    = "ac3"
	ACodecALAC   = "alac"
	ACodecDTS    = "dts"
	ACodecEAC3   = "eac3"
	ACodecFLAC   = "flac"
	ACodecMP2    = "mp2"
	ACodecMP3    = "mp3"
	ACodecOpus   = "opus"
	ACodecPCM    = "pcm"
	ACodecTrueHD = "truehd"
	ACodecVorbis = "vorbis"
	ACodecWAV    = "wav"
)

// GPU acceleration type constants.
const (
	AccelTypeAuto         = "auto"
	AccelTypeAMF          = "amf"
	AccelTypeCuda         = "cuda"
	AccelTypeQSV          = "qsv"
	AccelTypeVAAPI        = "vaapi"
	AccelTypeVideoToolbox = "videotoolbox"
)

// ValidVideoCodecs maps valid video codec names to true.
var ValidVideoCodecs = map[string]struct{}{
	VCodecCopy:  {},
	VCodecAV1:   {},
	VCodecH264:  {},
	VCodecHEVC:  {},
	VCodecMPEG2: {},
	VCodecVP8:   {},
	VCodecVP9:   {},
}

// ValidAudioCodecs maps valid audio codec names to true.
var ValidAudioCodecs = map[string]struct{}{
	ACodecCopy:   {},
	ACodecAAC:    {},
	ACodecAC3:    {},
	ACodecALAC:   {},
	ACodecDTS:    {},
	ACodecEAC3:   {},
	ACodecFLAC:   {},
	ACodecMP2:    {},
	ACodecMP3:    {},
	ACodecOpus:   {},
	ACodecPCM:    {},
	ACodecTrueHD: {},
	ACodecVorbis: {},
	ACodecWAV:    {},
}

// ValidGPUAccelTypes maps valid GPU acceleration type names to true.
var ValidGPUAccelTypes = map[string]struct{}{
	AccelTypeAuto:         {},
	AccelTypeAMF:          {},
	AccelTypeCuda:         {},
	AccelTypeQSV:          {},
	AccelTypeVAAPI:        {},
	AccelTypeVideoToolbox: {},
}

// ** Aliases ***********************************************************************************************

// ** Video **************************************************

// VideoCodecAlias maps video codec aliases to the valid string.
var VideoCodecAlias = map[string]string{
	// None.
	"":          "",
	"auto":      "",
	"automate":  "",
	"automated": "",
	"automatic": "",
	"none":      "",

	// AV1.
	"aom":       VCodecAV1,
	"av01":      VCodecAV1,
	"libaom":    VCodecAV1,
	"libaomav1": VCodecAV1,
	"libsvtav1": VCodecAV1,
	"svtav1":    VCodecAV1,

	// HEVC.
	"h265":     VCodecHEVC,
	"h265hevc": VCodecHEVC,
	"hevc265":  VCodecHEVC,
	"hvc1":     VCodecHEVC,
	"libx265":  VCodecHEVC,
	"x265":     VCodecHEVC,

	// MPEG2.
	"mpeg":       VCodecMPEG2,
	"mpeg2":      VCodecMPEG2,
	"mpeg2v":     VCodecMPEG2,
	"mpeg2video": VCodecMPEG2,
	"mpg":        VCodecMPEG2,
	"mpg2":       VCodecMPEG2,

	// VP8.
	"libvpx": VCodecVP8,
	"vp08":   VCodecVP8,
	"vpx":    VCodecVP8,
	"vpx8":   VCodecVP8,

	// VP9.
	"libvpx9":   VCodecVP9,
	"libvpxvp9": VCodecVP9,
	"vp09":      VCodecVP9,
	"vpx9":      VCodecVP9,
	"vpxvp9":    VCodecVP9,

	// h264.
	"avc":       VCodecH264,
	"avc1":      VCodecH264,
	"h264avc":   VCodecH264,
	"h264mpeg4": VCodecH264,
	"libx264":   VCodecH264,
	"mpeg4avc":  VCodecH264,
	"x264":      VCodecH264,
	"x264rgb":   VCodecH264,
}

// AudioCodecAlias maps audio codec aliases to the valid string.
var AudioCodecAlias = map[string]string{
	// None.
	"":          "",
	"auto":      "",
	"automate":  "",
	"automated": "",
	"automatic": "",
	"none":      "",

	// AAC
	"aac":              ACodecAAC,
	"aaclc":            ACodecAAC,
	"aaclowcomplexity": ACodecAAC,
	"aacplus":          ACodecAAC,
	"aache":            ACodecAAC,
	"heaac":            ACodecAAC,
	"m4a":              ACodecAAC,
	"mp4a":             ACodecAAC,

	// ALAC
	"alac":          ACodecALAC,
	"applelossless": ACodecALAC,
	"m4aalac":       ACodecALAC,

	// DTS
	"dca":      ACodecDTS,
	"dts":      ACodecDTS,
	"dtscodec": ACodecDTS,
	"dtsma":    ACodecDTS,
	"dtsmahd":  ACodecDTS,
	"dtshd":    ACodecDTS,
	"dtshdma":  ACodecDTS,

	// EAC3
	"ac3e":             ACodecEAC3,
	"ddp":              ACodecEAC3,
	"ddplus":           ACodecEAC3,
	"dolbydigitalplus": ACodecEAC3,
	"eac3":             ACodecEAC3,
	"ec3":              ACodecEAC3,

	// FLAC
	"fla":          ACodecFLAC,
	"flac":         ACodecFLAC,
	"flaccodec":    ACodecFLAC,
	"losslessflac": ACodecFLAC,

	// MP2
	"m2a":        ACodecMP2,
	"mp2":        ACodecMP2,
	"mp2codec":   ACodecMP2,
	"mpa":        ACodecMP2,
	"mpeg2":      ACodecMP2,
	"mpeg2audio": ACodecMP2,

	// MP3
	"libmp3lame": ACodecMP3,
	"mpeg3":      ACodecMP3,
	"mpeg3audio": ACodecMP3,
	"mp3":        ACodecMP3,
	"mp3codec":   ACodecMP3,
	"mpg3":       ACodecMP3,

	// Opus
	"oggopus":   ACodecOpus,
	"opus":      ACodecOpus,
	"opuscodec": ACodecOpus,
	"webmopus":  ACodecOpus,

	// PCM
	"pcm":      ACodecPCM,
	"pcm16":    ACodecPCM,
	"pcmcodec": ACodecPCM,
	"pcms16le": ACodecPCM,
	"pcms24le": ACodecPCM,
	"rawpcm":   ACodecPCM,
	"wavpcm":   ACodecPCM,

	// TrueHD
	"dolbytruehd": ACodecTrueHD,
	"thd":         ACodecTrueHD,
	"truehd":      ACodecTrueHD,
	"truehdcodec": ACodecTrueHD,

	// Vorbis
	"oggvorbis":   ACodecVorbis,
	"vorb":        ACodecVorbis,
	"vorbis":      ACodecVorbis,
	"vorbiscodec": ACodecVorbis,
	"webmvorbis":  ACodecVorbis,

	// WAV
	"pcmwave":  ACodecWAV,
	"wav":      ACodecWAV,
	"wavcodec": ACodecWAV,
	"wave":     ACodecWAV,
	"waveform": ACodecWAV,
}

// AccelAlias maps GPU acceleration codec aliases to the valid string.
var AccelAlias = map[string]string{
	// Auto
	"auto":      AccelTypeAuto,
	"automate":  AccelTypeAuto,
	"automated": AccelTypeAuto,
	"automatic": AccelTypeAuto,

	// AMD
	"amd":    AccelTypeAMF,
	"amdgpu": AccelTypeAMF,
	"radeon": AccelTypeAMF,
	"vcn":    AccelTypeAMF,

	// Intel
	"intel":     AccelTypeQSV,
	"qs":        AccelTypeQSV,
	"quicksync": AccelTypeQSV,

	// Nvidia
	"cuvid":  AccelTypeCuda,
	"nvdec":  AccelTypeCuda,
	"nvcuda": AccelTypeCuda,
	"nvidia": AccelTypeCuda,
	"nvenc":  AccelTypeCuda,
}
