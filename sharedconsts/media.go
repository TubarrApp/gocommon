package sharedconsts

// Simple string type for media forms, used in sharedvalidation/containers.go (etc.)
// to distinguish between video and audio.
type MediaType string

const (
	Video MediaType = "video"
	Audio MediaType = "audio"
)

// Compression thresholds.
const (
	BPPPFHighCompressionThreshold        = 0.04
	BPPPFMediumCompressionThreshold      = 0.08
	AudioBPSPCHighCompressionThreshold   = 64000
	AudioBPSPCMediumCompressionThreshold = 128000
)

// CompressionLevel defines the compression levels for a file.
type CompressionLevel string

const (
	CompressionBlank  CompressionLevel = ""
	CompressionLow    CompressionLevel = "lc"
	CompressionMedium CompressionLevel = "mc"
	CompressionHigh   CompressionLevel = "hc"
)

// ResolutionAlias maps resolution names to their shorter side in pixels.
var ResolutionAlias = map[string]string{
	"4k": "2160",
	"8k": "4320",
}

// ValidCompressionLevels maps valid compression levels.
var ValidCompressionLevels = map[string]CompressionLevel{
	string(CompressionBlank):  CompressionBlank,
	string(CompressionLow):    CompressionLow,
	string(CompressionMedium): CompressionMedium,
	string(CompressionHigh):   CompressionHigh,
}

// ValidVideoCodecs maps valid video codec names.
var ValidVideoCodecs = map[string]struct{}{
	VCodecCopy:  {},
	VCodecAV1:   {},
	VCodecH264:  {},
	VCodecHEVC:  {},
	VCodecMPEG2: {},
	VCodecVP8:   {},
	VCodecVP9:   {},
}

// VideoCodecMaxPasses maps video codecs to the most encoding passes their encoder uses.
var VideoCodecMaxPasses = map[string]int{
	VCodecAV1:   2,
	VCodecH264:  3,
	VCodecHEVC:  3,
	VCodecMPEG2: 3,
	VCodecVP8:   2,
	VCodecVP9:   2,
}

// MultiPassNeedsBitrate holds video codecs whose encoder needs a target bitrate, not a quality, for more than one pass.
var MultiPassNeedsBitrate = map[string]struct{}{
	VCodecAV1:   {},
	VCodecH264:  {},
	VCodecHEVC:  {},
	VCodecMPEG2: {},
}

// ValidAudioCodecs maps valid audio codec names.
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
