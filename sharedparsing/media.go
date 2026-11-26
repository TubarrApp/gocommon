package sharedparsing

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

const (
	crfOption        = "-crf"
	pixFmtOption     = "-pix_fmt"
	vfOption         = "-vf"
	afOption         = "-af"
	filterOption     = "-filter"
	passOption       = "-pass"
	bitrateOption    = "-b"
	channelsOption   = "-ac"
	sampleRateOption = "-ar"
)

// minBitrate is the lowest target bitrate taken, in bits per second, as encoders can write no frames at all below it.
const minBitrate = 1000

// EncoderOnlyOption reports the first option in args that Metarr only sets when encoding (-crf, -pix_fmt, -vf, -filter:v, -pass or -b:v).
func EncoderOnlyOption(args []string) (option string, found bool) {
	for _, arg := range args {
		if metarrSetsOption(videoOption(arg)) {
			return arg, true
		}
	}
	return "", false
}

// ParseVideoEncoderArgs parses "codec:args" and "codec[size]:args" lines into encoder args per target codec, then per
// smallest video size (shorter side) each is for, 0 for the plain "codec:" line. Size lines come laid over the plain line.
func ParseVideoEncoderArgs(lines []string) (map[string]map[int]sharedmodels.EncoderArgs, error) {
	byLine, err := parseCodecArgLines(lines, sharedvalidation.ValidateVideoCodec)
	if err != nil {
		return nil, err
	}

	// Parse each line's args.
	parsed := make(map[string]map[int]sharedmodels.EncoderArgs, len(byLine))
	for line, args := range byLine {
		if parsed[line.codec] == nil {
			parsed[line.codec] = make(map[int]sharedmodels.EncoderArgs)
		}
		if parsed[line.codec][line.size], err = parseEncoderArgs(line.codec, args); err != nil {
			return nil, err
		}
	}

	// Lay each size line over its codec's plain line, then check every set of args a video can get.
	for _, codec := range slices.Sorted(maps.Keys(parsed)) {
		sizes := parsed[codec]
		for _, size := range slices.Sorted(maps.Keys(sizes)) {
			if size == 0 {
				if err := checkRateControl(codec, sizes[0]); err != nil {
					return nil, err
				}
				continue
			}
			sizes[size] = mergeSizeArgs(codec, sizes[0], sizes[size])
			if err := checkRateControl(codec, sizes[size]); err != nil {
				return nil, fmt.Errorf("for %s[%d], %w", codec, size, err)
			}
		}
	}
	return parsed, nil
}

// NormalizeVideoArgSizes returns "codec[size]:args" lines with each size as its shorter side in pixels, e.g. "av1[1080]:" from "av1[1080p]:".
func NormalizeVideoArgSizes(lines []string) []string {
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		codec, args, found := strings.Cut(line, ":")
		name, size, sized := strings.Cut(codec, "[")
		if pixels, err := parseLineSize(size); found && sized && err == nil {
			line = fmt.Sprintf("%s[%d]:%s", strings.TrimSpace(name), pixels, args)
		}
		normalized = append(normalized, line)
	}
	return normalized
}

// ParseAudioEncoderArgs parses "codec:args" lines into encoder args per target audio codec.
func ParseAudioEncoderArgs(lines []string) (map[string]sharedmodels.AudioEncoderArgs, error) {
	byLine, err := parseCodecArgLines(lines, sharedvalidation.ValidateAudioCodec)
	if err != nil {
		return nil, err
	}

	// Parse each codec's args, refusing a "copy" line, as copied audio isn't encoded, and a size, as that's for video.
	parsed := make(map[string]sharedmodels.AudioEncoderArgs, len(byLine))
	for line, args := range byLine {
		if line.codec == sharedconsts.ACodecCopy {
			return nil, fmt.Errorf("%q in audio encoder args does nothing, as copied audio isn't encoded, remove its line", line.codec)
		}
		if line.size != 0 {
			return nil, fmt.Errorf("audio encoder args can't be set per video size, remove [%d] from the %s line", line.size, line.codec)
		}
		if parsed[line.codec], err = parseAudioEncoderArgs(line.codec, args); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}

// codecLine names a "codec:" or "codec[size]:" line, its size a video's shorter side in pixels, 0 for none.
type codecLine struct {
	codec string
	size  int
}

// parseCodecArgLines splits "codec:value" and "codec[size]:value" lines into values per target codec and size.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func parseCodecArgLines(lines []string, validate func(string) (string, error)) (byLine map[codecLine]string, err error) {
	byLine = make(map[codecLine]string, len(lines))

	// Check lines.
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Every line names its codec first, so a line starting with an option has none.
		codec, value, found := strings.Cut(line, ":")
		if !found || strings.HasPrefix(codec, "-") {
			return nil, fmt.Errorf("line %q needs the codec it's for first, then ':'", line)
		}

		// Take off a size, e.g. 1080 from "av1[1080p]".
		var key codecLine
		codec, size, sized := strings.Cut(codec, "[")
		if sized {
			if key.size, err = parseLineSize(size); err != nil {
				return nil, fmt.Errorf("line %q should name a size in [ ] like 720, 1080p or 4k, e.g. %s[1080]", line, strings.TrimSpace(codec))
			}
		}

		// Each codec and size named once.
		if key.codec, err = validate(codec); err != nil {
			return nil, err
		}
		if key.codec == "" {
			return nil, fmt.Errorf("line %q has no codec before ':'", line)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("line %q has no value after ':'", line)
		}
		if _, ok := byLine[key]; ok {
			if key.size != 0 {
				return nil, fmt.Errorf("%s[%d] is named more than once, keep only one", key.codec, key.size)
			}
			return nil, fmt.Errorf("codec %q is named more than once, keep only one", key.codec)
		}
		byLine[key] = value
	}
	return byLine, nil
}

// parseLineSize returns a line's size as a shorter side in pixels, e.g. 1080 from "1080p]" or 2160 from "4K]".
func parseLineSize(size string) (int, error) {
	size, closed := strings.CutSuffix(strings.TrimSpace(size), "]")
	if !closed {
		return 0, fmt.Errorf("size %q has no closing ]", size)
	}
	pixels, err := sharedvalidation.ValidateMaxResolution(size)
	if err != nil {
		return 0, err
	}
	if pixels == "" {
		return 0, fmt.Errorf("[ ] holds no size")
	}
	return strconv.Atoi(pixels)
}

// ParseCodecMappings splits codec entries into one plain codec for everything else and per-source "from:to" pairs.
//
// Names are normalized by validate, and order doesn't matter: pairs always override the plain codec.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func ParseCodecMappings(entries []string, validate func(string) (string, error)) (allTo string, pairs map[sharedmodels.FromCodec]string, err error) {
	pairs = make(map[sharedmodels.FromCodec]string, len(entries))

	for _, e := range entries {
		if e == "" {
			continue
		}

		// Split on ':' to separate 'from' and 'to' codec portion.
		from, to, isPair := strings.Cut(e, ":")
		if from == "" {
			if !isPair {
				return "", nil, fmt.Errorf("codec sent in empty (%q)", e)
			}
			return "", nil, fmt.Errorf("'from' codec sent in empty in string %s:%s", from, to)
		}

		// Grab 'FromCodec' model.
		fromCodec, err := validateFromCodec(from, validate)
		if err != nil {
			return "", nil, err
		}

		// Passed in validation function (video or audio codec name validator).
		if fromCodec.Codec, err = validate(fromCodec.Codec); err != nil {
			return "", nil, err
		}

		// Plain codec, only one allowed.
		if !isPair {
			if fromCodec.Compression != sharedconsts.CompressionBlank {
				return "", nil, fmt.Errorf("%q has a compression level but nothing to transcode to, e.g. %s:hevc", e, e)
			}
			if allTo != "" && fromCodec.Codec != "" && fromCodec.Codec != allTo {
				return "", nil, fmt.Errorf("more than one plain codec (%q and %q), keep only one", allTo, fromCodec.Codec)
			}
			if fromCodec.Codec != "" {
				allTo = fromCodec.Codec
			}
			continue
		}

		// Pair, each source named once.
		if fromCodec.Codec == "" {
			return "", nil, fmt.Errorf("codec pair %q has no source codec", e)
		}
		if to, err = validate(to); err != nil {
			return "", nil, err
		}
		if prev, ok := pairs[fromCodec]; ok && prev != to {
			return "", nil, fmt.Errorf("source codec %q is named more than once (%q and %q), keep only one", from, prev, to)
		}
		pairs[fromCodec] = to
	}
	return allTo, pairs, nil
}

// validateFromCodec returns a valid FromCodec model if valid.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func validateFromCodec(c string, validate func(string) (string, error)) (sharedmodels.FromCodec, error) {
	blankModel := sharedmodels.NewFromCodec(sharedconsts.CodecCopy, sharedconsts.CompressionBlank)

	// Check if codec string is empty.
	if c == "" {
		return blankModel, nil
	}

	// Get codec and compression level.
	validCodec, compressionLevel, err := parseFromCodec(c, validate)
	if err != nil {
		return blankModel, err
	}

	// Return validated FromCodec.
	return sharedmodels.NewFromCodec(validCodec, compressionLevel), nil
}

// parseFromCodec retrieves the valid compression level and codec.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func parseFromCodec(c string, validate func(string) (string, error)) (codec string, compressionLevel sharedconsts.CompressionLevel, err error) {
	// Normalize string.
	c = strings.ToLower(c)
	c = strings.TrimSpace(c)
	c = strings.ReplaceAll(c, ".", "")
	// Don't remove "_", is delimiter (e.g. 'mc_hevc')

	// Check compression level component. Any other "_" is part of the codec's name (e.g. 'pcm_s16le').
	compressionLevel = sharedconsts.CompressionBlank
	compression, rest, found := strings.Cut(c, "_")
	if level, isLevel := sharedconsts.ValidCompressionLevels[compression]; found && isLevel {
		c = rest // Extracted codec portion (e.g. 'h264' from 'lc_h264').
		compressionLevel = level
	}

	// Validate codec component.
	validCodec, err := validate(c)
	if err != nil {
		return "", compressionLevel, err
	}

	// Return retrieved codec and compression level.
	return validCodec, compressionLevel, nil
}

// splitEncoderArgs splits a codec's args shell style, refusing disallowed options.
func splitEncoderArgs(codec, args string) ([]string, error) {
	split, err := ShellSplit(args)
	if err != nil {
		return nil, fmt.Errorf("invalid %s encoder args %q: %w", codec, args, err)
	}

	// Refuse disallowed options.
	if !sharedconsts.InsecureArgumentsAllowed() {
		if blocked, option := sharedconsts.HasBlockedOptions(split, sharedconsts.BlockedFFmpegOptions); blocked {
			return nil, fmt.Errorf("option %q is not allowed in %s encoder args", option, codec)
		}
	}
	return split, nil
}

// parseEncoderArgs splits a codec's args, taking out -crf, -pix_fmt, -vf, -pass and -b:v for Metarr to convert per encoder.
func parseEncoderArgs(codec, args string) (sharedmodels.EncoderArgs, error) {
	var parsed sharedmodels.EncoderArgs
	split, err := splitEncoderArgs(codec, args)
	if err != nil {
		return parsed, err
	}

	// Take out the options converted per encoder.
	for i := 0; i < len(split); i++ {
		arg := split[i]
		option := videoOption(arg)

		// Pass other options through as typed.
		if !metarrSetsOption(option) {
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
			if parsed.Bitrate != "" {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("-b:v is set more than once in %s encoder args, keep only one", codec)
			}
			bitrate, bitrateErr := parseBitrate(value)
			if bitrateErr != nil {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("invalid -b:v in %s encoder args: %w", codec, bitrateErr)
			}
			parsed.Bitrate = strconv.FormatInt(bitrate, 10)
		default: // vfOption, filterOption
			if parsed.VideoFilter != "" {
				return sharedmodels.EncoderArgs{}, fmt.Errorf("more than one video filter in %s encoder args, join them with ',' in one -vf", codec)
			}
			parsed.VideoFilter = value
		}
	}
	return parsed, nil
}

// checkRateControl returns an error if a set of args holds a quality or target bitrate its codec's encoder can't use.
func checkRateControl(codec string, args sharedmodels.EncoderArgs) error {
	hasBitrate := args.Bitrate != "" && args.Bitrate != "0"

	// Some encoders need a target bitrate, not a quality, for more than one pass.
	if _, needsBitrate := sharedconsts.MultiPassNeedsBitrate[codec]; needsBitrate && args.Passes > 1 {
		if args.Quality != "" {
			return fmt.Errorf("%s can't use -crf with more than one pass, give a target bitrate with -b:v instead", codec)
		}
		if !hasBitrate {
			return fmt.Errorf("%s needs a target bitrate (-b:v) for more than one pass", codec)
		}
	}

	// Some encoders take a quality or a target bitrate, not both.
	if _, either := sharedconsts.QualityOrBitrate[codec]; either && args.Quality != "" && hasBitrate {
		return fmt.Errorf("%s can't use -crf with a target bitrate (-b:v), keep only one", codec)
	}
	return nil
}

// mergeSizeArgs returns a size line's args laid over its codec's plain line, the size line's winning wherever both set an option.
func mergeSizeArgs(codec string, plain, size sharedmodels.EncoderArgs) sharedmodels.EncoderArgs {
	merged := sharedmodels.EncoderArgs{
		Quality:     cmp.Or(size.Quality, plain.Quality),
		PixFmt:      cmp.Or(size.PixFmt, plain.PixFmt),
		VideoFilter: cmp.Or(size.VideoFilter, plain.VideoFilter),
		Passes:      cmp.Or(size.Passes, plain.Passes),
		Bitrate:     cmp.Or(size.Bitrate, plain.Bitrate),
		Args:        mergeArgs(plain.Args, size.Args),
	}

	// Where the encoder takes a quality or a target bitrate, a size line's one replaces the plain line's either.
	if _, either := sharedconsts.QualityOrBitrate[codec]; either && (size.Quality != "" || size.Bitrate != "") {
		merged.Quality = size.Quality
		merged.Bitrate = size.Bitrate
	}
	return merged
}

// mergeArgs returns the plain line's args, less any option the size line also sets, followed by the size line's.
func mergeArgs(plain, size []string) []string {
	// Note the options the size line sets.
	sizeOptions := make(map[string]bool, len(size))
	for _, arg := range size {
		if isOption(arg) {
			sizeOptions[videoOption(arg)] = true
		}
	}

	// Keep the plain line's other options, each with its value.
	var merged []string
	keep := true
	for _, arg := range plain {
		if isOption(arg) {
			keep = !sizeOptions[videoOption(arg)]
		}
		if keep {
			merged = append(merged, arg)
		}
	}
	return append(merged, size...)
}

// isOption reports whether an arg names an option rather than being a value, which may be a negative number like -1.
func isOption(arg string) bool {
	return len(arg) > 1 && arg[0] == '-' && (arg[1] < '0' || arg[1] > '9') && arg[1] != '.'
}

// parseBitrate reads a bitrate in bits per second, with k, M or G for thousands, millions or billions (e.g. 1500000 from "1.5M").
func parseBitrate(s string) (int64, error) {
	s = strings.TrimSpace(s)
	number := s
	multiplier := 1.0
	if s != "" {
		switch strings.ToLower(s[len(s)-1:]) {
		case "k":
			multiplier = 1e3 // e means number of zeros, not scientific notation, so 1e3 = 1000.
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
	if err != nil || math.IsNaN(n) || n < 0 || n*multiplier >= math.MaxInt64 {
		return 0, fmt.Errorf("bitrate should be a number with k, M or G, e.g. 3M or 1500k, got %q", s)
	}
	bitrate := int64(n * multiplier)

	// 0 is no target, which some encoders use with a quality.
	if bitrate > 0 && bitrate < minBitrate {
		return 0, fmt.Errorf("bitrate %q is too low to encode a video, use at least 1k, or 0 for no target", s)
	}
	return bitrate, nil
}

// parseAudioEncoderArgs splits a codec's args, taking out -af, -ac and -ar for Metarr to add itself.
func parseAudioEncoderArgs(codec, args string) (sharedmodels.AudioEncoderArgs, error) {
	var parsed sharedmodels.AudioEncoderArgs
	split, err := splitEncoderArgs(codec, args)
	if err != nil {
		return parsed, err
	}

	// Take out the options Metarr adds itself.
	for i := 0; i < len(split); i++ {
		arg := split[i]
		option := audioOption(arg)

		// Pass other options through as typed.
		if option != afOption && option != filterOption && option != channelsOption && option != sampleRateOption {
			parsed.Args = append(parsed.Args, arg)
			continue
		}
		if i+1 == len(split) {
			return sharedmodels.AudioEncoderArgs{}, fmt.Errorf("%s in %s encoder args has no value", arg, codec)
		}
		i++
		value := split[i]

		switch option {
		case channelsOption:
			if parsed.Channels != "" {
				return sharedmodels.AudioEncoderArgs{}, fmt.Errorf("-ac is set more than once in %s encoder args, keep only one", codec)
			}
			parsed.Channels = value
		case sampleRateOption:
			if parsed.SampleRate != "" {
				return sharedmodels.AudioEncoderArgs{}, fmt.Errorf("-ar is set more than once in %s encoder args, keep only one", codec)
			}
			parsed.SampleRate = value
		default: // afOption, filterOption
			if parsed.Filter != "" {
				return sharedmodels.AudioEncoderArgs{}, fmt.Errorf("more than one audio filter in %s encoder args, join them with ',' in one -af", codec)
			}
			parsed.Filter = value
		}
	}
	return parsed, nil
}

// metarrSetsOption reports whether Metarr takes this video option out of encoder args and adds it back in the form the running encoder needs.
func metarrSetsOption(option string) bool {
	switch option {
	case crfOption,
		pixFmtOption,
		vfOption,
		filterOption,
		passOption,
		bitrateOption:

		return true

	default:
		return false
	}
}

// videoOption returns an option's name without a video stream specifier, e.g. "-crf" from "-crf:v:0".
func videoOption(arg string) string {
	name, spec, _ := strings.Cut(arg, ":")
	if spec == "" || spec == "v" || spec == "v:0" {
		return name
	}
	return arg
}

// audioOption returns an option's name without an audio stream specifier, e.g. "-filter" from "-filter:a:0".
func audioOption(arg string) string {
	name, spec, _ := strings.Cut(arg, ":")
	if spec == "" || spec == "a" || spec == "a:0" {
		return name
	}
	return arg
}
