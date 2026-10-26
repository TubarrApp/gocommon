package sharedvalidation

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// containerExt returns ext lowercased, with a leading dot.
func containerExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

// codecFits reports whether table lets codec go in a file of extension ext. Blank and "copy" codecs, and containers
// missing from table, are taken to fit.
func codecFits(table map[string][]string, ext, codec string) bool {
	codecs, known := table[containerExt(ext)]
	return !known || codec == "" || codec == sharedconsts.VCodecCopy || slices.Contains(codecs, codec)
}

// VideoCodecFits reports whether ffmpeg can write the (normalized) video codec into a file of extension ext.
func VideoCodecFits(ext, codec string) bool {
	return codecFits(sharedconsts.ContainerCompatibleVideoCodecs, ext, codec)
}

// AudioCodecFits reports whether ffmpeg can write the (normalized) audio codec into a file of extension ext.
func AudioCodecFits(ext, codec string) bool {
	return codecFits(sharedconsts.ContainerCompatibleAudioCodecs, ext, codec)
}

// FittingVideoCodecs returns the video codecs ext can hold, in sharedconsts.IncompatibleVideoCodecPreference order.
func FittingVideoCodecs(ext string) []string {
	var fits []string
	for _, c := range sharedconsts.IncompatibleVideoCodecPreference {
		if VideoCodecFits(ext, c) {
			fits = append(fits, c)
		}
	}
	return fits
}

// FittingAudioCodecs returns the audio codecs ext can hold, in sharedconsts.IncompatibleAudioCodecPreference order.
func FittingAudioCodecs(ext string) []string {
	var fits []string
	for _, c := range sharedconsts.IncompatibleAudioCodecPreference {
		if AudioCodecFits(ext, c) {
			fits = append(fits, c)
		}
	}
	return fits
}

// CodecMismatchError describes a codec that can't go in a container, naming the containers it fits.
func CodecMismatchError(kind, codec, ext string) error {
	table := sharedconsts.ContainerCompatibleVideoCodecs
	if kind == "audio" {
		table = sharedconsts.ContainerCompatibleAudioCodecs
	}
	var fits []string
	for _, e := range slices.Sorted(maps.Keys(table)) {
		if slices.Contains(table[e], codec) {
			fits = append(fits, e)
		}
	}
	if len(fits) == 0 {
		return fmt.Errorf("%s codec %q can't be written into any container (ffmpeg's encoder for it is experimental)", kind, codec)
	}
	return fmt.Errorf("%s codec %q can't go in a %s file. It fits: %s", kind, codec, containerExt(ext), strings.Join(fits, ", "))
}

// ValidateCodecsForContainer returns an error if a transcode codec can't go in a file of extension ext, naming the
// containers it fits. Codecs are given as for transcoding: "h264", or "vp9:h264" to transcode only VP9 sources to H.264.
func ValidateCodecsForContainer(ext string, videoCodecs, audioCodecs []string) error {
	for _, pair := range videoCodecs {
		c, err := ValidateVideoCodec(outputCodec(pair))
		if err != nil {
			return err
		}
		if !VideoCodecFits(ext, c) {
			return CodecMismatchError("video", c, ext)
		}
	}
	for _, pair := range audioCodecs {
		c, err := ValidateAudioCodec(outputCodec(pair))
		if err != nil {
			return err
		}
		if !AudioCodecFits(ext, c) {
			return CodecMismatchError("audio", c, ext)
		}
	}
	return nil
}

// outputCodec returns the codec a transcode entry outputs: the last of its ":"-separated parts.
func outputCodec(pair string) string {
	parts := strings.Split(pair, ":")
	return strings.TrimSpace(parts[len(parts)-1])
}
