package sharedparsing

import (
	"cmp"
	"fmt"
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

// EncoderOnlyOption reports the first option in args that Metarr only sets when encoding (-crf, -pix_fmt, -vf, -filter:v, -pass or -b:v).
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

// parseEncoderArgs splits a codec's args, taking out -crf, -pix_fmt, -vf, -pass and -b:v for Metarr to convert per encoder.
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
		case bitrateOption:
			if parsed.Bitrates != nil {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-b:v is set more than once in %s encoder args, keep only one", codec)
			}
			if parsed.Bitrates, err = parseBitrates(value); err != nil {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("invalid -b:v in %s encoder args: %w", codec, err)
			}
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
		hasBitrate := false
		for _, bitrate := range parsed.Bitrates {
			hasBitrate = hasBitrate || bitrate > 0
		}
		if !hasBitrate {
			return sharedmodels.EncoderArgs{}, fmt.Errorf("%s needs a target bitrate (-b:v) for more than one pass", codec)
		}
	}
	return parsed, nil
}

// parseBitrates reads -b:v's value: one bitrate, or "resolution=bitrate" pairs joined by ',' (e.g. "720=1.5M,1080=3M").
func parseBitrates(value string) (map[int]int64, error) {
	if !strings.Contains(value, "=") {
		bitrate, err := parseBitrate(value)
		if err != nil {
			return nil, err
		}
		return map[int]int64{0: bitrate}, nil
	}

	// One bitrate per resolution.
	bitrates := make(map[int]int64)
	for pair := range strings.SplitSeq(value, ",") {
		res, rate, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			return nil, fmt.Errorf("%q should be resolution=bitrate, e.g. 1080=3M", pair)
		}
		resolution, err := sharedvalidation.ValidateMaxResolution(res)
		if err != nil || resolution == "" {
			return nil, fmt.Errorf("%q should start with a resolution like 720, 1080p or 4k", pair)
		}
		pixels, err := strconv.Atoi(resolution)
		if err != nil {
			return nil, err
		}
		if _, named := bitrates[pixels]; named {
			return nil, fmt.Errorf("%dp is named more than once, keep only one", pixels)
		}
		if bitrates[pixels], err = parseBitrate(rate); err != nil {
			return nil, err
		}
	}
	return bitrates, nil
}

// parseBitrate reads a bitrate in bits per second, with k, M or G for thousands, millions or billions (e.g. 1500000 from "1.5M").
func parseBitrate(s string) (int64, error) {
	s = strings.TrimSpace(s)
	number := s
	multiplier := 1.0
	if s != "" {
		switch strings.ToLower(s[len(s)-1:]) {
		case "k":
			multiplier = 1e3
		case "m":
			multiplier = 1e6
		case "g":
			multiplier = 1e9
		}
		if multiplier != 1 {
			number = s[:len(s)-1]
		}
	}
	n, err := strconv.ParseFloat(number, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("bitrate should be a number with k, M or G, e.g. 3M or 1500k, got %q", s)
	}
	return int64(n * multiplier), nil
}

// encoderOnly reports whether Metarr sets the option itself, converted for the encoder that runs.
func encoderOnly(option string) bool {
	switch option {
	case crfOption, pixFmtOption, vfOption, filterOption, passOption, bitrateOption:
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
