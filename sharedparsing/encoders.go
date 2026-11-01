package sharedparsing

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

const (
	crfOption     = "-crf"
	pixFmtOption  = "-pix_fmt"
	vfOption      = "-vf"
	filterOption  = "-filter"
	passOption    = "-pass"
	bitrateOption = "-b"
)

// EncoderOnlyOption reports the first option in args that Metarr only sets when encoding (-crf, -pix_fmt, -vf, -filter:v or -pass).
func EncoderOnlyOption(args []string) (option string, found bool) {
	for _, arg := range args {
		if encoderOnly(videoOption(arg)) {
			return arg, true
		}
	}
	return "", false
}

// ParseVideoEncoderArgs parses "codec:args" lines into encoder args per target codec.
func ParseVideoEncoderArgs(lines []string) (map[string]sharedmodels.EncoderArgs, error) {
	byCodec, err := sharedvalidation.ParseCodecValueLines(lines)
	if err != nil {
		return nil, err
	}

	// Parse each codec's args.
	parsed := make(map[string]sharedmodels.EncoderArgs, len(byCodec))
	for codec, args := range byCodec {
		if parsed[codec], err = parseEncoderArgs(codec, args); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}

// parseEncoderArgs splits a codec's args, taking out -crf, -pix_fmt, -vf and -pass for Metarr to convert per encoder.
func parseEncoderArgs(codec, args string) (sharedmodels.EncoderArgs, error) {
	var parsed sharedmodels.EncoderArgs

	// Split shell style.
	split, err := ShellSplit(args)
	if err != nil {
		return parsed, fmt.Errorf("invalid %s encoder args %q: %w", codec, args, err)
	}

	// Refuse disallowed options.
	if blocked, option := sharedconsts.HasBlockedOptions(split, sharedconsts.BlockedFFmpegOptions); blocked {
		return parsed, fmt.Errorf("option %q is not allowed in %s encoder args", option, codec)
	}

	// Take out the options converted per encoder.
	for i := 0; i < len(split); i++ {
		arg := split[i]
		option := videoOption(arg)

		// Pass other options through as typed.
		if !encoderOnly(option) {
			parsed.Args = append(parsed.Args, arg)
			continue
		}
		if i+1 == len(split) {
			return sharedmodels.EncoderArgs{}, fmt.Errorf("%s in %s encoder args has no value", arg, codec)
		}
		i++
		value := split[i]

		switch option {
		case crfOption:
			if parsed.Quality != "" {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-crf is set more than once in %s encoder args, keep only one", codec)
			}
			if parsed.Quality, err = sharedvalidation.ValidateTranscodeQuality(codec, value); err != nil {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("invalid %s encoder args: %w", codec, err)
			}
		case pixFmtOption:
			if parsed.PixFmt != "" {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-pix_fmt is set more than once in %s encoder args, keep only one", codec)
			}
			parsed.PixFmt = strings.ToLower(value)
		case passOption:
			if parsed.Passes != 0 {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-pass is set more than once in %s encoder args, keep only one", codec)
			}
			passes, atoiErr := strconv.Atoi(value)
			if atoiErr != nil {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-pass in %s encoder args should be a number of passes, got %q", codec, value)
			}
			parsed.Passes = min(max(passes, 1), cmp.Or(sharedconsts.VideoCodecMaxPasses[codec], 1))
		default: // vfOption, filterOption
			if parsed.VideoFilter != "" {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("more than one video filter in %s encoder args, join them with ',' in one -vf", codec)
			}
			parsed.VideoFilter = value
		}
	}

	// Some encoders need a target bitrate, not a quality, for more than one pass.
	if _, needsBitrate := sharedconsts.MultiPassNeedsBitrate[codec]; needsBitrate && parsed.Passes > 1 {
		if parsed.Quality != "" {
			return sharedmodels.EncoderArgs{}, fmt.Errorf("%s can't use -crf with more than one pass, give a target bitrate with -b:v instead", codec)
		}
		if !slices.ContainsFunc(parsed.Args, func(arg string) bool { return videoOption(arg) == bitrateOption }) {
			return sharedmodels.EncoderArgs{}, fmt.Errorf("%s needs a target bitrate (-b:v) for more than one pass", codec)
		}
	}
	return parsed, nil
}

// encoderOnly reports whether Metarr sets the option itself, converted for the encoder that runs.
func encoderOnly(option string) bool {
	switch option {
	case crfOption, pixFmtOption, vfOption, filterOption, passOption:
		return true
	}
	return false
}

// videoOption returns an option's name without a video stream specifier, e.g. "-crf" from "-crf:v:0".
func videoOption(arg string) string {
	name, spec, _ := strings.Cut(arg, ":")
	if spec == "" || spec == "v" || spec == "v:0" {
		return name
	}
	return arg
}
