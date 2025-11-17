package sharedvalidation

import (
	"fmt"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// ValidateVideoCodec validates a video codec string and returns the normalized codec name.
// Handles common aliases and synonyms (e.g., "x264" -> "h264", "libx265" -> "hevc").
func ValidateVideoCodec(c string) (string, error) {
	// Normalize input.
	c = strings.ToLower(strings.TrimSpace(c))
	c = strings.ReplaceAll(c, ".", "")
	c = strings.ReplaceAll(c, "-", "")
	c = strings.ReplaceAll(c, "_", "")

	// Synonym and alias mapping.
	switch c {
	case "", "none", "auto", "automatic", "automated":
		return "", nil
	case "aom", "libaom", "libaomav1", "av01", "svtav1", "libsvtav1":
		c = sharedconsts.VCodecAV1
	case "x264", "avc", "h264avc", "mpeg4avc", "h264mpeg4", "libx264":
		c = sharedconsts.VCodecH264
	case "x265", "h265", "hevc265", "libx265", "hevc":
		c = sharedconsts.VCodecHEVC
	case "mpg2", "mpeg2video", "mpeg2v", "mpg", "mpeg", "mpeg2":
		c = sharedconsts.VCodecMPEG2
	case "libvpx", "vp08", "vpx", "vpx8":
		c = sharedconsts.VCodecVP8
	case "libvpxvp9", "libvpx9", "vpx9", "vp09", "vpxvp9":
		c = sharedconsts.VCodecVP9
	}

	// Check against valid codec map.
	if sharedconsts.ValidVideoCodecs[c] {
		return c, nil
	}

	return "", fmt.Errorf("video codec %q is not valid. Supported: %v", c, sharedconsts.ValidVideoCodecs)
}

// ValidateVideoCodecWithAccel validates a video codec string with GPU acceleration context.
// Returns an error if auto/none codec is used with specific GPU acceleration.
func ValidateVideoCodecWithAccel(c, accel string) (validCodec string, err error) {
	// Ensure valid codec.
	validated, err := ValidateVideoCodec(c)
	if err != nil {
		return "", err
	}

	// Check if empty codec is allowed with the given acceleration type.
	if validated == "" &&
		(accel != "" && accel != sharedconsts.AccelTypeAuto) {
		return "", fmt.Errorf("GPU acceleration %q requires a codec (entered %q)", accel, c)
	}

	return validated, nil
}

// ValidateAudioCodec validates an audio codec string and returns the normalized codec name.
// Handles common aliases and synonyms (e.g., "mp3" -> "mp3", "libmp3lame" -> "mp3").
func ValidateAudioCodec(a string) (string, error) {
	// Normalize input.
	a = strings.ToLower(strings.TrimSpace(a))
	a = strings.ReplaceAll(a, ".", "")
	a = strings.ReplaceAll(a, "-", "")
	a = strings.ReplaceAll(a, "_", "")

	// Synonym and alias mapping.
	switch a {
	case "", "none", "auto", "automatic", "automated":
		return "", nil
	case "aac", "aaclc", "m4a", "mp4a", "aaclowcomplexity":
		a = sharedconsts.ACodecAAC
	case "alac", "applelossless", "m4aalac":
		a = sharedconsts.ACodecALAC
	case "dca", "dts", "dtshd", "dtshdma", "dtsma", "dtsmahd", "dtscodec":
		a = sharedconsts.ACodecDTS
	case "ddplus", "dolbydigitalplus", "ac3e", "ec3", "eac3":
		a = sharedconsts.ACodecEAC3
	case "flac", "flaccodec", "fla", "losslessflac":
		a = sharedconsts.ACodecFLAC
	case "mp2", "mpa", "mpeg2audio", "mpeg2", "m2a", "mp2codec":
		a = sharedconsts.ACodecMP2
	case "mp3", "libmp3lame", "mpeg3", "mpeg3audio", "mpg3", "mp3codec":
		a = sharedconsts.ACodecMP3
	case "opus", "opuscodec", "oggopus", "webmopus":
		a = sharedconsts.ACodecOpus
	case "pcm", "wavpcm", "rawpcm", "pcm16", "pcms16le", "pcms24le", "pcmcodec":
		a = sharedconsts.ACodecPCM
	case "truehd", "dolbytruehd", "thd", "truehdcodec":
		a = sharedconsts.ACodecTrueHD
	case "vorbis", "oggvorbis", "webmvorbis", "vorbiscodec", "vorb":
		a = sharedconsts.ACodecVorbis
	case "wav", "wave", "waveform", "pcmwave", "wavcodec":
		a = sharedconsts.ACodecWAV
	}

	// Check against valid codec map.
	if sharedconsts.ValidAudioCodecs[a] {
		return a, nil
	}

	return "", fmt.Errorf("audio codec %q is not valid. Supported: %v", a, sharedconsts.ValidAudioCodecs)
}

// ValidateGPUAccelType validates a GPU acceleration type string.
func ValidateGPUAccelType(accel string) (string, error) {
	// Normalize input.
	accel = strings.ToLower(strings.TrimSpace(accel))

	// Synonym and alias mapping.
	switch accel {
	case "automatic", "automate", "automated":
		accel = sharedconsts.AccelTypeAuto
	case "radeon", "amd":
		accel = sharedconsts.AccelTypeAMF
	case "intel":
		accel = sharedconsts.AccelTypeIntel
	case "nvidia", "nvenc":
		accel = sharedconsts.AccelTypeNvidia
	}

	// Check against valid acceleration type map.
	if sharedconsts.ValidGPUAccelTypes[accel] {
		return accel, nil
	}

	// Return error on map check failure.
	return "", fmt.Errorf("%s GPU acceleration type %q is not valid. Supported: %v", sharedconsts.LogTagError, accel, sharedconsts.ValidGPUAccelTypes)
}
