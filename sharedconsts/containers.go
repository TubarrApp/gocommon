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

// The codecs ffmpeg can write into each video container, found by encoding a test clip into every container with the
// encoder Metarr uses for each codec and reading it back with ffprobe (Metarr's TestContainerCodecsWithFFmpeg, ffmpeg
// 9.0.2; DTS and TrueHD with -strict experimental, as Metarr runs them).
// .ogm files are written by the .ogv muxer, so they share its lists. .rm and .rmvb list nothing, which refuses them as
// output: ffmpeg's RealMedia muxer only holds RealVideo, and even a copied stream comes out undecodable.

// ContainerCompatibleVideoCodecs lists the video codecs each container can hold.
var ContainerCompatibleVideoCodecs = map[string][]string{
	Ext3GP:  {VCodecH264},
	Ext3G2:  {VCodecH264},
	ExtASF:  {VCodecAV1, VCodecH264, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtAVI:  {VCodecAV1, VCodecH264, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtF4V:  {VCodecH264},
	ExtFLV:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecVP9},
	ExtOGM:  {VCodecVP8},
	ExtOGV:  {VCodecVP8},
	ExtM4V:  {VCodecH264},
	ExtMKV:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtMOV:  {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtMP4:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP9},
	ExtMPEG: {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtMPG:  {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtMTS:  {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtRM:   {},
	ExtRMVB: {},
	ExtTS:   {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtVOB:  {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtWEBM: {VCodecAV1, VCodecVP8, VCodecVP9},
	ExtWMV:  {VCodecAV1, VCodecH264, VCodecMPEG2, VCodecVP8, VCodecVP9},
}

// ContainerCompatibleAudioCodecs lists the audio codecs each container can hold.
var ContainerCompatibleAudioCodecs = map[string][]string{
	Ext3GP:  {ACodecAAC},
	Ext3G2:  {ACodecAAC},
	ExtASF:  {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecPCM, ACodecVorbis, ACodecWAV},
	ExtAVI:  {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecPCM, ACodecVorbis, ACodecWAV},
	ExtF4V:  {ACodecAAC, ACodecMP3},
	ExtFLV:  {ACodecAAC, ACodecAC3, ACodecEAC3, ACodecFLAC, ACodecMP3, ACodecOpus},
	ExtOGM:  {ACodecFLAC, ACodecOpus, ACodecVorbis},
	ExtOGV:  {ACodecFLAC, ACodecOpus, ACodecVorbis},
	ExtM4V:  {ACodecAAC, ACodecAC3, ACodecALAC},
	ExtMKV:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecVorbis, ACodecWAV},
	ExtMOV:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecMP2, ACodecMP3, ACodecPCM, ACodecWAV},
	ExtMP4:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecWAV},
	ExtMPEG: {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtMPG:  {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtMTS:  {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecMP2, ACodecMP3, ACodecOpus},
	ExtRM:   {},
	ExtRMVB: {},
	ExtTS:   {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecMP2, ACodecMP3, ACodecOpus},
	ExtVOB:  {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtWEBM: {ACodecOpus, ACodecVorbis},
	ExtWMV:  {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecPCM, ACodecVorbis, ACodecWAV},
}

// IncompatibleVideoCodecPreference is the order a video codec is picked in when a video's own doesn't fit its output container.
var IncompatibleVideoCodecPreference = []string{VCodecH264, VCodecHEVC, VCodecVP9, VCodecVP8, VCodecMPEG2, VCodecAV1}

// IncompatibleAudioCodecPreference is the order an audio codec is picked in when a video's own doesn't fit its output container.
var IncompatibleAudioCodecPreference = []string{ACodecAAC, ACodecOpus, ACodecMP3, ACodecAC3, ACodecVorbis, ACodecFLAC, ACodecEAC3,
	ACodecMP2, ACodecALAC, ACodecPCM}
