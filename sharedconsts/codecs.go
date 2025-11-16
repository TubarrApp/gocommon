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
	AccelTypeAuto   = "auto"
	AccelTypeAMF    = "amf"
	AccelTypeIntel  = "qsv"
	AccelTypeNvidia = "cuda"
	AccelTypeVAAPI  = "vaapi"
)

// ValidVideoCodecs maps valid video codec names to true.
var ValidVideoCodecs = map[string]bool{
	VCodecCopy:  true,
	VCodecAV1:   true,
	VCodecH264:  true,
	VCodecHEVC:  true,
	VCodecMPEG2: true,
	VCodecVP8:   true,
	VCodecVP9:   true,
}

// ValidAudioCodecs maps valid audio codec names to true.
var ValidAudioCodecs = map[string]bool{
	ACodecCopy:   true,
	ACodecAAC:    true,
	ACodecAC3:    true,
	ACodecALAC:   true,
	ACodecDTS:    true,
	ACodecEAC3:   true,
	ACodecFLAC:   true,
	ACodecMP2:    true,
	ACodecMP3:    true,
	ACodecOpus:   true,
	ACodecPCM:    true,
	ACodecTrueHD: true,
	ACodecVorbis: true,
	ACodecWAV:    true,
}

// ValidGPUAccelTypes maps valid GPU acceleration type names to true.
var ValidGPUAccelTypes = map[string]bool{
	AccelTypeAuto:   true,
	AccelTypeAMF:    true,
	AccelTypeIntel:  true,
	AccelTypeNvidia: true,
	AccelTypeVAAPI:  true,
}
