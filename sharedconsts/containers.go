package sharedconsts

// The codecs ffmpeg can write into each video container, found by encoding a test clip into every container with the
// encoder Metarr uses for each codec (ffmpeg 2026-10; DTS and TrueHD with -strict experimental, as Metarr runs them).
// ffmpeg can't write .ogm or .rmvb files at all.

// ContainerCompatibleVideoCodecs lists the video codecs each container can hold.
var ContainerCompatibleVideoCodecs = map[string][]string{
	Ext3GP:  {VCodecH264},
	Ext3G2:  {VCodecH264},
	ExtASF:  {VCodecAV1, VCodecH264, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtAVI:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtF4V:  {VCodecH264},
	ExtFLV:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecVP9},
	ExtOGM:  {},
	ExtOGV:  {VCodecVP8},
	ExtM4V:  {VCodecH264},
	ExtMKV:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtMOV:  {VCodecH264, VCodecHEVC, VCodecMPEG2},
	ExtMP4:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP9},
	ExtMPEG: {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtMPG:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtMTS:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtRM:   {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtRMVB: {},
	ExtTS:   {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
	ExtVOB:  {VCodecAV1, VCodecH264, VCodecHEVC, VCodecMPEG2, VCodecVP8, VCodecVP9},
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
	ExtOGM:  {},
	ExtOGV:  {ACodecFLAC, ACodecOpus, ACodecVorbis},
	ExtM4V:  {ACodecAAC, ACodecAC3, ACodecALAC},
	ExtMKV:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecVorbis, ACodecWAV},
	ExtMOV:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecMP2, ACodecMP3, ACodecPCM, ACodecWAV},
	ExtMP4:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecWAV},
	ExtMPEG: {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtMPG:  {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtMTS:  {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecVorbis, ACodecWAV},
	ExtRM:   {ACodecAAC, ACodecAC3},
	ExtRMVB: {},
	ExtTS:   {ACodecAAC, ACodecAC3, ACodecALAC, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecOpus, ACodecPCM, ACodecTrueHD, ACodecVorbis, ACodecWAV},
	ExtVOB:  {ACodecAC3, ACodecDTS, ACodecMP2, ACodecMP3},
	ExtWEBM: {ACodecOpus, ACodecVorbis},
	ExtWMV:  {ACodecAAC, ACodecAC3, ACodecDTS, ACodecEAC3, ACodecFLAC, ACodecMP2, ACodecMP3, ACodecPCM, ACodecVorbis, ACodecWAV},
}

// IncompatibleVideoCodecPreference is the order a video codec is picked in when a video's own doesn't fit its output container.
var IncompatibleVideoCodecPreference = []string{VCodecH264, VCodecHEVC, VCodecVP9, VCodecVP8, VCodecMPEG2, VCodecAV1}

// IncompatibleAudioCodecPreference is the order an audio codec is picked in when a video's own doesn't fit its output container.
var IncompatibleAudioCodecPreference = []string{ACodecAAC, ACodecOpus, ACodecMP3, ACodecAC3, ACodecVorbis, ACodecFLAC, ACodecEAC3,
	ACodecMP2, ACodecALAC, ACodecPCM}
