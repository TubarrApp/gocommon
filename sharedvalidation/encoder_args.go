package sharedvalidation

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
)

// These are only values that can't work on any FFmpeg version. Anything that might is left to fail with FFmpeg's error.
var (
	// SVT-AV1's -preset is a number, and these are other encoders' names.
	badAV1Presets = []string{
		"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow", "placebo",
		"p1", "p2", "p3", "p4", "p5", "p6", "p7", "hp", "hq", "ll", "llhq",
	}

	// VP8 and VP9 profiles are numbers, and these are other codecs' names.
	badVPXProfiles = []string{"baseline", "high", "high10", "high422", "high444", "high444p", "main", "main12", "rext", "mv", "simple"}

	// encoderParamsCodecs maps options named after an encoder to its codec, as FFmpeg silently ignores them on others.
	encoderParamsCodecs = map[string]string{
		"-x264-params":   sharedconsts.VCodecH264,
		"-x265-params":   sharedconsts.VCodecHEVC,
		"-svtav1-params": sharedconsts.VCodecAV1,
	}

	// audioEncodingOptions only work on audio being encoded.
	audioEncodingOptions = []string{"-af", "-filter:a", "-ac", "-ar", "-b:a", "-q:a"}
)

// ValidateSoftwareEncoderArgs returns an error if a codec's encoder args can't work on its software encoder.
func ValidateSoftwareEncoderArgs(codec string, args sharedmodels.EncoderArgs) error {
	options := optionValues(args.Args)

	switch codec {
	case sharedconsts.VCodecAV1:
		if preset := options["-preset"]; slices.Contains(badAV1Presets, preset) {
			return fmt.Errorf("%s -preset is a number, e.g. -preset 8, not %q", codec, preset)
		}
	case sharedconsts.VCodecVP8, sharedconsts.VCodecVP9:
		if profile := options["-profile"]; slices.Contains(badVPXProfiles, profile) {
			return fmt.Errorf("%s -profile:v is a number, e.g. -profile:v 0, not %q", codec, profile)
		}
	}
	return nil
}

// ValidateEncoderArgsUsed returns an error if a codec's encoder args do nothing on any of its encoders: a "copy" line, or an option for another encoder.
func ValidateEncoderArgsUsed(codec string, args sharedmodels.EncoderArgs) error {
	if codec == sharedconsts.VCodecCopy {
		return fmt.Errorf("%q in video encoder args does nothing, as copied video isn't encoded, remove its line", codec)
	}
	options := optionValues(args.Args)
	for _, option := range slices.Sorted(maps.Keys(encoderParamsCodecs)) {
		if _, found := options[option]; found && encoderParamsCodecs[option] != codec {
			return fmt.Errorf("%s does nothing on %s, as it's only for %s", option, codec, encoderParamsCodecs[option])
		}
	}
	return nil
}

// ValidateEncoderArgsCanRun returns an error if a codec's encoder args can never run, as it's copied and nothing encodes to it.
func ValidateEncoderArgsCanRun(codec, plainCodec string, pairs map[sharedmodels.FromCodec]string, outputExt string) error {
	if plainCodec == codec || slices.Contains(slices.Collect(maps.Values(pairs)), codec) {
		return nil
	}

	// With no plain codec, a container that can't hold a stream could still choose this codec for it.
	ownPair := pairs[sharedmodels.NewFromCodec(codec, sharedconsts.CompressionBlank)]
	if plainCodec == sharedconsts.CodecCopy || (ownPair == sharedconsts.CodecCopy && (plainCodec != "" || keepsContainer(outputExt))) {
		return fmt.Errorf("%s encoder args do nothing, as %s is copied and nothing is encoded to it, remove the line", codec, codec)
	}
	return nil
}

// ValidateExtraArgsAudioOptions returns an error if extra FFmpeg args hold an option that only works on audio being encoded.
func ValidateExtraArgsAudioOptions(extraArgs []string) error {
	options := optionValues(extraArgs)
	for _, option := range audioEncodingOptions {
		if _, found := options[option]; found {
			return fmt.Errorf("%s in extra FFmpeg args only works on audio being encoded, so it goes in audio encoder args instead, e.g. aac:%s", option, option)
		}
	}
	return nil
}

// keepsContainer reports whether a file keeps its container, with no output extension set.
func keepsContainer(outputExt string) bool {
	return outputExt == "" || outputExt == sharedconsts.ExtSame
}

// optionValues returns each arg, without a video stream or first stream specifier, with the arg after it ("" if none).
func optionValues(args []string) map[string]string {
	options := make(map[string]string, len(args))
	for i, arg := range args {
		name := strings.TrimSuffix(strings.TrimSuffix(arg, ":0"), ":v")

		value := ""
		if i+1 < len(args) {
			value = args[i+1]
		}
		options[name] = value
	}
	return options
}
