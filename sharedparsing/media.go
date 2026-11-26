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

// ParseVideoEncoderArgs parses "codec:args" and "codec[resolution]:args" lines into encoder args per target codec, then
// per resolution tier, 0 for the plain "codec:" line. A resolution line comes laid over the plain line for each tier it covers.
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
		if parsed[line.codec][line.resolution], err = parseEncoderArgs(line.codec, args); err != nil {
			return nil, err
		}
	}

	// Lay each resolution line over its codec's plain line, then check every set of args a video can get.
	for _, codec := range slices.Sorted(maps.Keys(parsed)) {
		byResolution := parsed[codec]
		for _, resolution := range slices.Sorted(maps.Keys(byResolution)) {
			if resolution == 0 {
				if err := checkRateControl(codec, byResolution[0]); err != nil {
					return nil, err
				}
				continue
			}
			byResolution[resolution] = mergeResolutionArgs(codec, byResolution[0], byResolution[resolution])
			if err := checkRateControl(codec, byResolution[resolution]); err != nil {
				return nil, fmt.Errorf("for %s[%d], %w", codec, resolution, err)
			}
		}
	}
	return parsed, nil
}

// NormalizeResolutions returns codec entries or "codec:args" lines with each resolution as saved, e.g. "av1[720-2160]:"
// from "av1[720p-4K]:", and the rest as typed.
func NormalizeResolutions(lines []string) []string {
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		before, after, found := strings.Cut(line, ":")
		name, resolution, hasResolution := strings.Cut(before, "[")
		if _, saved, err := parseResolutions(resolution); found && hasResolution && err == nil {
			line = strings.TrimSpace(name) + "[" + saved + "]:" + after
		}
		normalized = append(normalized, line)
	}
	return normalized
}

// ResolutionTier returns the resolution tier a video's shorter side is in, e.g. 1080 for 1200, or 0 if its size is
// unknown. A video smaller than every tier is in the smallest.
func ResolutionTier(shorterSide int) int {
	if shorterSide <= 0 {
		return 0
	}
	tier := sharedconsts.VideoResolutions[0]
	for _, resolution := range sharedconsts.VideoResolutions {
		if resolution <= shorterSide {
			tier = resolution
		}
	}
	return tier
}

// ParseAudioEncoderArgs parses "codec:args" lines into encoder args per target audio codec.
func ParseAudioEncoderArgs(lines []string) (map[string]sharedmodels.AudioEncoderArgs, error) {
	byLine, err := parseCodecArgLines(lines, sharedvalidation.ValidateAudioCodec)
	if err != nil {
		return nil, err
	}

	// Parse each codec's args, refusing a "copy" line, as copied audio isn't encoded, and a resolution, as that's for video.
	parsed := make(map[string]sharedmodels.AudioEncoderArgs, len(byLine))
	for line, args := range byLine {
		if line.codec == sharedconsts.ACodecCopy {
			return nil, fmt.Errorf("%q in audio encoder args does nothing, as copied audio isn't encoded, remove its line", line.codec)
		}
		if line.resolution != 0 {
			return nil, fmt.Errorf("audio encoder args can't be set per video resolution, remove the [ ] from the %s line", line.codec)
		}
		if parsed[line.codec], err = parseAudioEncoderArgs(line.codec, args); err != nil {
			return nil, err
		}
	}
	return parsed, nil
}

// codecLine names a "codec:" or "codec[resolution]:" line by its codec and one resolution tier it covers, 0 for none.
type codecLine struct {
	codec      string
	resolution int
}

// parseCodecArgLines splits "codec:value" and "codec[resolution]:value" lines into values per target codec and
// resolution tier, a resolution covering several tiers giving each the line's value.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func parseCodecArgLines(lines []string, validate func(string) (string, error)) (byLine map[codecLine]string, err error) {
	byLine = make(map[codecLine]string, len(lines))
	namedBy := make(map[codecLine]string, len(lines)) // The "codec[resolution]" naming each, for errors.

	// Check lines.
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Every line names its codec first, so a line starting with an option has none.
		label, value, found := strings.Cut(line, ":")
		if !found || strings.HasPrefix(label, "-") {
			return nil, fmt.Errorf("line %q needs the codec it's for first, then ':'", line)
		}
		label = strings.TrimSpace(label)

		// Take off a resolution, e.g. 1080 from "av1[1080p]".
		codec, resolution, hasResolution := strings.Cut(label, "[")
		tiers := []int{0}
		saved := ""
		if hasResolution {
			if tiers, saved, err = parseResolutions(resolution); err != nil {
				return nil, fmt.Errorf("invalid resolution in line %q: %w", line, err)
			}
		}

		// Each codec and resolution tier named once.
		var name string
		if name, err = validate(codec); err != nil {
			return nil, err
		}
		if name == "" && hasResolution {
			return nil, fmt.Errorf("line %q needs the codec its args are for before the [ ], e.g. av1[%s]", line, saved)
		}
		if name == "" {
			return nil, fmt.Errorf("line %q has no codec before ':'", line)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("line %q has no value after ':'", line)
		}
		for _, tier := range tiers {
			key := codecLine{codec: name, resolution: tier}
			if earlier, named := namedBy[key]; named {
				if tier != 0 {
					return nil, fmt.Errorf("%s and %s both cover %dp, keep only one", earlier, label, tier)
				}
				return nil, fmt.Errorf("codec %q is named more than once, keep only one", name)
			}
			namedBy[key] = label
			byLine[key] = value
		}
	}
	return byLine, nil
}

// parseResolutions returns the tiers a resolution in [ ] covers and how it's saved, e.g. 720 to 2160 and "720-2160"
// from "720p-4K]". It's one tier, a range of them, or a tier and up, e.g. "1080]", "720-1080]" or "1440+]".
func parseResolutions(s string) (tiers []int, saved string, err error) {
	s, closed := strings.CutSuffix(strings.TrimSpace(s), "]")
	if !closed {
		return nil, "", fmt.Errorf("[%s has no closing ]", s)
	}
	s = strings.ReplaceAll(strings.ToLower(s), " ", "")

	// Read the lowest and highest tier.
	low, high, isRange := strings.Cut(s, "-")
	andUp := strings.Contains(s, "+")
	if andUp && isRange {
		return nil, "", fmt.Errorf("[%s] is a range and a resolution and up, use one, e.g. [720-1080] or [1440+]", s)
	}
	lowTier, err := parseResolutionTier(strings.TrimSuffix(low, "+"))
	if err != nil {
		return nil, "", err
	}
	highTier := lowTier
	switch {
	case andUp:
		highTier = sharedconsts.VideoResolutions[len(sharedconsts.VideoResolutions)-1]
		saved = strconv.Itoa(lowTier) + "+"
	case isRange:
		if highTier, err = parseResolutionTier(high); err != nil {
			return nil, "", err
		}
		if highTier < lowTier {
			return nil, "", fmt.Errorf("[%s] should have the lower resolution first, e.g. [720-1080]", s)
		}
		saved = strconv.Itoa(lowTier) + "-" + strconv.Itoa(highTier)
	default:
		saved = strconv.Itoa(lowTier)
	}

	// Collect the tiers covered.
	for _, tier := range sharedconsts.VideoResolutions {
		if tier >= lowTier && tier <= highTier {
			tiers = append(tiers, tier)
		}
	}
	return tiers, saved, nil
}

// parseResolutionTier returns a resolution tier in pixels, e.g. 1080 from "1080p" or 2160 from "4k".
func parseResolutionTier(s string) (int, error) {
	pixels, err := sharedvalidation.ValidateMaxResolution(s)
	tier, atoiErr := strconv.Atoi(pixels)
	if err != nil || atoiErr != nil || !slices.Contains(sharedconsts.VideoResolutions, tier) {
		return 0, fmt.Errorf("%q isn't a resolution, use 144, 240, 360, 480, 720, 1080, 1440, 2160 (4k) or 4320 (8k)", s)
	}
	return tier, nil
}

// ParseCodecMappings splits codec entries into one plain codec for everything else and "from:to" pairs, keyed by source
// codec, compression level and resolution tier. A resolution covering several tiers gives a pair for each, and an entry
// naming only a resolution, e.g. "[720+]:av1", has no source codec.
//
// Names are normalized by validate, and order doesn't matter: [CodecFor] picks the entry naming the most about a video.
//
// Pass in [sharedvalidation.ValidateVideoCodec] or [sharedvalidation.ValidateAudioCodec].
func ParseCodecMappings(entries []string, validate func(string) (string, error)) (allTo string, pairs map[sharedmodels.FromCodec]string, err error) {
	pairs = make(map[sharedmodels.FromCodec]string, len(entries))
	namedBy := make(map[sharedmodels.FromCodec]string, len(entries)) // The entry naming each pair, for errors.

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

		// Take off a resolution, e.g. 720p and up from "av1[720+]".
		from, resolution, hasResolution := strings.Cut(from, "[")
		var tiers []int
		var saved string
		if hasResolution {
			if tiers, saved, err = parseResolutions(resolution); err != nil {
				return "", nil, fmt.Errorf("invalid resolution in %q: %w", e, err)
			}
			if !isPair {
				return "", nil, fmt.Errorf("%q has a resolution but no codec to encode to, e.g. [%s]:%s", e, saved, cmp.Or(strings.TrimSpace(from), sharedconsts.VCodecAV1))
			}
		}

		// Grab 'FromCodec' model, which has no codec for an entry naming only a resolution.
		var fromCodec sharedmodels.FromCodec
		if strings.TrimSpace(from) != "" {
			if fromCodec, err = validateFromCodec(from, validate); err != nil {
				return "", nil, err
			}

			// Passed in validation function (video or audio codec name validator).
			if fromCodec.Codec, err = validate(fromCodec.Codec); err != nil {
				return "", nil, err
			}
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

		// Pair, which only leaves out the source codec when it names a resolution.
		if fromCodec.Codec == "" && !hasResolution {
			return "", nil, fmt.Errorf("codec pair %q has no source codec", e)
		}
		if fromCodec.Codec == "" && fromCodec.Compression != sharedconsts.CompressionBlank {
			return "", nil, fmt.Errorf("%q has a compression level but no source codec, e.g. lc_h264[%s]:%s", e, saved, to)
		}
		if strings.Contains(to, "[") {
			return "", nil, fmt.Errorf("%q has a resolution after ':', where it should go on the source, e.g. h264[720+]:hevc", e)
		}
		if to, err = validate(to); err != nil {
			return "", nil, err
		}

		// A resolution gives a pair for each tier it covers.
		froms := []sharedmodels.FromCodec{fromCodec}
		if hasResolution {
			froms = nil
			for _, tier := range tiers {
				fromCodec.Resolution = tier
				froms = append(froms, fromCodec)
			}
		}

		// Each source named once.
		for _, key := range froms {
			if prev, ok := pairs[key]; ok && prev != to {
				if key.Resolution != 0 {
					return "", nil, fmt.Errorf("%q and %q both cover %dp, keep only one", namedBy[key], e, key.Resolution)
				}
				return "", nil, fmt.Errorf("source codec %q is named more than once (%q and %q), keep only one", from, prev, to)
			}
			pairs[key] = to
			namedBy[key] = e
		}
	}
	return allTo, pairs, nil
}

// ParseAudioCodecMappings parses audio codec entries as [ParseCodecMappings] does, refusing a resolution, as that's for video.
func ParseAudioCodecMappings(entries []string) (allTo string, pairs map[sharedmodels.FromCodec]string, err error) {
	for _, e := range entries {
		if strings.Contains(e, "[") {
			return "", nil, fmt.Errorf("audio codec entries can't go by video resolution, remove the [ ] from %q", e)
		}
	}
	return ParseCodecMappings(entries, sharedvalidation.ValidateAudioCodec)
}

// CodecFor returns the codec entries' target for a video, and whether the entry names its codec, as a pair does. The
// entry naming the most about the video wins: its codec counting first, then its compression level, then its resolution.
func CodecFor(video sharedmodels.FromCodec, plain string, pairs map[sharedmodels.FromCodec]string) (to string, pair bool) {
	// Pairs naming the video's codec, most specific first.
	if video.Codec != "" {
		for _, from := range []sharedmodels.FromCodec{
			video,
			{Codec: video.Codec, Compression: video.Compression},
			{Codec: video.Codec, Resolution: video.Resolution},
			{Codec: video.Codec},
		} {
			if target, ok := pairs[from]; ok {
				return target, true
			}
		}
	}

	// Then an entry naming only the video's resolution, then the plain codec.
	if target, ok := pairs[sharedmodels.FromCodec{Resolution: video.Resolution}]; ok && video.Resolution != 0 {
		return target, false
	}
	return plain, false
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

// mergeResolutionArgs returns a resolution line's args laid over its codec's plain line, the resolution line's winning
// wherever both set an option.
func mergeResolutionArgs(codec string, plain, resolution sharedmodels.EncoderArgs) sharedmodels.EncoderArgs {
	merged := sharedmodels.EncoderArgs{
		Quality:     cmp.Or(resolution.Quality, plain.Quality),
		PixFmt:      cmp.Or(resolution.PixFmt, plain.PixFmt),
		VideoFilter: cmp.Or(resolution.VideoFilter, plain.VideoFilter),
		Passes:      cmp.Or(resolution.Passes, plain.Passes),
		Bitrate:     cmp.Or(resolution.Bitrate, plain.Bitrate),
		Args:        mergeArgs(plain.Args, resolution.Args),
	}

	// Where the encoder takes a quality or a target bitrate, a resolution line's one replaces the plain line's either.
	if _, either := sharedconsts.QualityOrBitrate[codec]; either && (resolution.Quality != "" || resolution.Bitrate != "") {
		merged.Quality = resolution.Quality
		merged.Bitrate = resolution.Bitrate
	}
	return merged
}

// mergeArgs returns the plain line's args, less any option the resolution line also sets, followed by the resolution line's.
func mergeArgs(plain, resolution []string) []string {
	// Note the options the resolution line sets.
	resolutionOptions := make(map[string]bool, len(resolution))
	for _, arg := range resolution {
		if isOption(arg) {
			resolutionOptions[videoOption(arg)] = true
		}
	}

	// Keep the plain line's other options, each with its value.
	var merged []string
	keep := true
	for _, arg := range plain {
		if isOption(arg) {
			keep = !resolutionOptions[videoOption(arg)]
		}
		if keep {
			merged = append(merged, arg)
		}
	}
	return append(merged, resolution...)
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
