package sharedparsing

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// ******************************** Encoder settings **********************************************************************************

// TestCheckEncoderSettings tests the settings are read and judged together, with the software checks only when software encoders run.
func TestCheckEncoderSettings(t *testing.T) {
	for _, c := range []struct {
		name string
		s    sharedmodels.EncoderSettings
		ok   bool
	}{
		{name: "nothing set", ok: true},
		{name: "working settings", s: sharedmodels.EncoderSettings{
			VideoArgs: []string{"av1:-preset 8", "h264:-crf 20"}, VideoCodecs: []string{"h264:av1"}, AudioArgs: []string{"aac:-b:a 128k -ac 2"}, AudioCodecs: []string{"aac"},
			ExtraArgs: "-movflags +faststart", OutputExt: ".mkv", Software: true,
		}, ok: true},

		// Lines that do nothing.
		{name: "copy line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"copy:-crf 20"}}},
		{name: "another encoder's option", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-x265-params log-level=error"}}},
		{name: "line for copied video", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-crf 20"}, VideoCodecs: []string{"copy"}}},
		{name: "audio copy line", s: sharedmodels.EncoderSettings{AudioArgs: []string{"copy:-b:a 128k"}}},
		{name: "line for copied audio", s: sharedmodels.EncoderSettings{AudioArgs: []string{"aac:-b:a 128k"}, AudioCodecs: []string{"copy"}}},
		{name: "line for audio its own pair copies", s: sharedmodels.EncoderSettings{AudioArgs: []string{"aac:-ac 2"}, AudioCodecs: []string{"aac:copy"}}},

		// Audio lines that can run.
		{name: "line for audio a pair encodes to", s: sharedmodels.EncoderSettings{AudioArgs: []string{"ac3:-b:a 448k"}, AudioCodecs: []string{"copy", "dts:ac3"}}, ok: true},
		{name: "line for audio already in its codec", s: sharedmodels.EncoderSettings{AudioArgs: []string{"aac:-ac 2"}}, ok: true},
		{name: "line when the container may choose its codec", s: sharedmodels.EncoderSettings{AudioArgs: []string{"aac:-af volume=0.5"}, AudioCodecs: []string{"aac:copy"}, OutputExt: ".mp4"}, ok: true},

		// Values software encoders can't take, only judged when one runs.
		{name: "AV1 preset name in software", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium"}, Software: true}},
		{name: "AV1 preset name otherwise", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium"}}, ok: true},
		{name: "VP9 profile name in software", s: sharedmodels.EncoderSettings{VideoArgs: []string{"vp9:-profile:v main"}, Software: true}},
		{name: "VP9 profile name otherwise", s: sharedmodels.EncoderSettings{VideoArgs: []string{"vp9:-profile:v main"}}, ok: true},

		// Later lines are still judged, whichever sorts first.
		{name: "a bad line after a good one", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset 8", "vp9:-profile:v main"}, Software: true}},

		// Resolution lines are judged with the plain line under them.
		{name: "a bad resolution line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-crf 30", "av1[4k]:-preset medium"}, Software: true}},
		{name: "another encoder's option in a resolution line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-crf 20", "h264[1080]:-x265-params log-level=error"}}},
		{name: "a resolution line with no plain line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1[1080]:-preset 8"}, Software: true}, ok: true},
		{name: "a bad plain line a resolution line overrides", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium", "av1[1080]:-preset 8"}, Software: true}},

		// Codec entries by resolution, for video only.
		{name: "video codec entries by resolution", s: sharedmodels.EncoderSettings{
			VideoCodecs: []string{"copy", "[720+]:av1", "lc_h264[1080-4k]:hevc"}, VideoArgs: []string{"av1:-crf 30", "hevc[2160]:-crf 24"},
		}, ok: true},
		{name: "an audio codec entry by resolution", s: sharedmodels.EncoderSettings{AudioCodecs: []string{"[720+]:aac"}}},

		// Audio encoding options in extra FFmpeg args, whether or not audio is encoded.
		{name: "filter in extra args", s: sharedmodels.EncoderSettings{ExtraArgs: "-af volume=0.5", AudioCodecs: []string{"aac"}}},
		{name: "bitrate in extra args", s: sharedmodels.EncoderSettings{ExtraArgs: "-b:a 128k", AudioCodecs: []string{"dts:ac3"}}},
		{name: "channels in extra args with no audio codec", s: sharedmodels.EncoderSettings{ExtraArgs: "-ac 1"}},

		// Settings that can't be read.
		{name: "a line with no codec", s: sharedmodels.EncoderSettings{VideoArgs: []string{"-crf 20"}}},
		{name: "an audio line with no codec", s: sharedmodels.EncoderSettings{AudioArgs: []string{"-b:a 128k"}}},
		{name: "a bad video codec", s: sharedmodels.EncoderSettings{VideoCodecs: []string{"bad"}}},
		{name: "a bad audio codec", s: sharedmodels.EncoderSettings{AudioCodecs: []string{"bad"}}},
		{name: "extra args with an open quote", s: sharedmodels.EncoderSettings{ExtraArgs: `-metadata "title`}},
	} {
		if err := CheckEncoderSettings(c.s); (err == nil) != c.ok {
			t.Errorf("%s: got %v, want ok %t", c.name, err, c.ok)
		}
	}
}

// ******************************** Encoders ******************************************************************************************

// TestParseCodecMappings tests splitting codec entries into a plain codec and per-source pairs, in any order.
func TestParseCodecMappings(t *testing.T) {
	low := sharedconsts.CompressionLow
	from := func(codec string, level sharedconsts.CompressionLevel, resolution int) sharedmodels.FromCodec {
		return sharedmodels.FromCodec{Codec: codec, Compression: level, Resolution: resolution}
	}
	tests := []struct {
		in        []string
		wantAll   string
		wantPairs map[sharedmodels.FromCodec]string
		ok        bool
	}{
		// Pass states.
		{[]string{"hevc", "vp9:h264"}, "hevc", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec("vp9", ""): "h264"}, true},
		{[]string{"vp9:h264", "hevc"}, "hevc", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec("vp9", ""): "h264"}, true},
		{[]string{"x264"}, "h264", map[sharedmodels.FromCodec]string{}, true},
		{[]string{"hevc", "libx265"}, "hevc", map[sharedmodels.FromCodec]string{}, true},
		{[]string{"vp9:h264", "vp09:x264"}, "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec("vp9", ""): "h264"}, true},
		{[]string{"", "av1:hevc"}, "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec("av1", ""): "hevc"}, true},
		{nil, "", map[sharedmodels.FromCodec]string{}, true},
		{[]string{"lc_h264:hevc"}, "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec("h264", sharedconsts.CompressionLow): "hevc"}, true},
		{[]string{"none"}, "", map[sharedmodels.FromCodec]string{}, true},

		// A resolution gives a pair for each tier it covers.
		{[]string{"lc_av1[720p-1080]:av1"}, "", map[sharedmodels.FromCodec]string{from("av1", low, 720): "av1", from("av1", low, 1080): "av1"}, true},
		{[]string{"copy", "[1440+]:av1"}, "copy", map[sharedmodels.FromCodec]string{
			from("", "", 1440): "av1", from("", "", 2160): "av1", from("", "", 4320): "av1",
		}, true},
		{[]string{"av1[4K]:h264", "av1:hevc"}, "", map[sharedmodels.FromCodec]string{from("av1", "", 2160): "h264", from("av1", "", 0): "hevc"}, true},
		{[]string{"av1[720-1080]:h264", "av1[1080]:h264"}, "", map[sharedmodels.FromCodec]string{from("av1", "", 720): "h264", from("av1", "", 1080): "h264"}, true},
		{[]string{"lc_av1[<720P]:copy", "lc_av1[720+]:av1"}, "", map[sharedmodels.FromCodec]string{
			from("av1", low, 144): "copy", from("av1", low, 240): "copy", from("av1", low, 360): "copy", from("av1", low, 480): "copy",
			from("av1", low, 720): "av1", from("av1", low, 1080): "av1", from("av1", low, 1440): "av1", from("av1", low, 2160): "av1", from("av1", low, 4320): "av1",
		}, true},

		// Fail states.
		{[]string{"av1[720+]"}, "", nil, false},
		{[]string{"[720+]"}, "", nil, false},
		{[]string{"lc_[720+]:hevc"}, "", nil, false},
		{[]string{"av1[720-1080]:h264", "av1[1080+]:hevc"}, "", nil, false},
		{[]string{"av1:hevc[720+]"}, "", nil, false},
		{[]string{"av1[1000]:h264"}, "", nil, false},
		{[]string{"av1[1080-720]:h264"}, "", nil, false},
		{[]string{"av1[720:h264"}, "", nil, false},
		{[]string{"lc_h264"}, "", nil, false},
		{[]string{"hevc", "h264"}, "", nil, false},
		{[]string{"vp9:h264", "libvpxvp9:hevc"}, "", nil, false},
		{[]string{":h264"}, "", nil, false},
		{[]string{"vp9:bad"}, "", nil, false},
		{[]string{"vp9:h264:av1"}, "", nil, false},
		{[]string{"bad"}, "", nil, false},
		{[]string{"zz_h264:hevc"}, "", nil, false},
	}

	for _, tt := range tests {
		all, pairs, err := ParseCodecMappings(tt.in, sharedvalidation.ValidateVideoCodec)
		if tt.ok != (err == nil) {
			t.Errorf("ParseCodecMappings(%q) error = %v, want ok %v", tt.in, err, tt.ok)
			continue
		}
		if all != tt.wantAll || !maps.Equal(pairs, tt.wantPairs) {
			t.Errorf("ParseCodecMappings(%q) = (%q, %v), want (%q, %v)", tt.in, all, pairs, tt.wantAll, tt.wantPairs)
		}
	}
}

// TestParseAudioCodecMappings tests a "_" splits off a known compression level, and is otherwise part of the codec's name.
func TestParseAudioCodecMappings(t *testing.T) {
	low := sharedconsts.CompressionLow
	tests := []struct {
		in        string
		wantAll   string
		wantPairs map[sharedmodels.FromCodec]string
	}{
		{"pcm_s16le", sharedconsts.ACodecPCM, map[sharedmodels.FromCodec]string{}},
		{"pcm_s16le:aac", "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec(sharedconsts.ACodecPCM, ""): sharedconsts.ACodecAAC}},
		{"aac:pcm_s16le", "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec(sharedconsts.ACodecAAC, ""): sharedconsts.ACodecPCM}},
		{"lc_pcm_s16le:aac", "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec(sharedconsts.ACodecPCM, low): sharedconsts.ACodecAAC}},
		{"lc_aac:opus", "", map[sharedmodels.FromCodec]string{sharedmodels.NewFromCodec(sharedconsts.ACodecAAC, low): sharedconsts.ACodecOpus}},
	}

	for _, tt := range tests {
		all, pairs, err := ParseAudioCodecMappings([]string{tt.in})
		if err != nil {
			t.Errorf("ParseAudioCodecMappings(%q) error = %v", tt.in, err)
			continue
		}
		if all != tt.wantAll || !maps.Equal(pairs, tt.wantPairs) {
			t.Errorf("ParseAudioCodecMappings(%q) = (%q, %v), want (%q, %v)", tt.in, all, pairs, tt.wantAll, tt.wantPairs)
		}
	}

	// A resolution is for video.
	for _, in := range []string{"aac[1080]:opus", "[720+]:aac"} {
		if _, _, err := ParseAudioCodecMappings([]string{in}); err == nil {
			t.Errorf("ParseAudioCodecMappings(%q) took a resolution", in)
		}
	}
}

// TestCodecFor tests the entry naming the most about a video wins: its codec first, then its compression level, then its resolution.
func TestCodecFor(t *testing.T) {
	plain, pairs, err := ParseCodecMappings([]string{
		"copy", "[720+]:av1", "h264:hevc", "h264[1080]:vp9", "lc_h264:mpeg2", "lc_h264[2160]:h264", "vp9:none",
	}, sharedvalidation.ValidateVideoCodec)
	if err != nil {
		t.Fatal(err)
	}
	video := func(codec string, level sharedconsts.CompressionLevel, resolution int) sharedmodels.FromCodec {
		return sharedmodels.FromCodec{Codec: codec, Compression: level, Resolution: resolution}
	}
	low := sharedconsts.CompressionLow
	high := sharedconsts.CompressionHigh

	for _, c := range []struct {
		name     string
		video    sharedmodels.FromCodec
		wantTo   string
		wantPair bool
	}{
		{"codec, level and resolution", video("h264", low, 2160), "h264", true},
		{"codec and level beat codec and resolution", video("h264", low, 1080), "mpeg2", true},
		{"codec and resolution", video("h264", high, 1080), "vp9", true},
		{"codec only", video("h264", high, 480), "hevc", true},
		{"codec beats a resolution entry", video("h264", "", 2160), "hevc", true},
		{"a pair to nothing still names the codec", video("vp9", "", 1080), "", true},
		{"resolution entry", video("av1", "", 1080), "av1", false},
		{"below every resolution entry", video("av1", "", 480), "copy", false},
		{"size unknown", video("mpeg4", "", 0), "copy", false},
	} {
		to, pair := CodecFor(c.video, plain, pairs)
		if to != c.wantTo || pair != c.wantPair {
			t.Errorf("%s: got %q, pair %t; want %q, pair %t", c.name, to, pair, c.wantTo, c.wantPair)
		}
	}
}

// TestResolutionTier tests a video is in the largest resolution tier its shorter side reaches.
func TestResolutionTier(t *testing.T) {
	for side, want := range map[int]int{0: 0, -1: 0, 64: 144, 144: 144, 239: 144, 240: 240, 800: 720, 1079: 720, 1080: 1080, 1200: 1080, 1439: 1080, 1440: 1440, 2160: 2160, 8640: 4320} {
		if got := ResolutionTier(side); got != want {
			t.Errorf("ResolutionTier(%d) = %d, want %d", side, got, want)
		}
	}
}

func TestParseCodecValueLines(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		wantByCodec map[codecLine]string
		wantErr     bool
	}{
		{"codec lines", []string{" av1: -crf 35 ", "", "hevc:-preset slow"}, map[codecLine]string{{"av1", 0}: "-crf 35", {"hevc", 0}: "-preset slow"}, false},
		{"alias normalized", []string{"x265:-crf 24"}, map[codecLine]string{{"hevc", 0}: "-crf 24"}, false},
		{"args keep inner colons", []string{"av1:-svtav1-params tune=0:film-grain=8"}, map[codecLine]string{{"av1", 0}: "-svtav1-params tune=0:film-grain=8"}, false},
		{
			"resolutions beside a plain line",
			[]string{"av1:-preset 6", "av1[1080p]:-crf 30", "AV1 [4K]:-crf 35", "x265[720i]:-crf 24"},
			map[codecLine]string{{"av1", 0}: "-preset 6", {"av1", 1080}: "-crf 30", {"av1", 2160}: "-crf 35", {"hevc", 720}: "-crf 24"},
			false,
		},
		{"resolution without a plain line", []string{"av1[1080]:-crf 30"}, map[codecLine]string{{"av1", 1080}: "-crf 30"}, false},
		{
			"a range and a resolution and up, each tier given the line",
			[]string{"av1[720-1080]:-crf 30", "hevc[2160+]:-crf 26"},
			map[codecLine]string{{"av1", 720}: "-crf 30", {"av1", 1080}: "-crf 30", {"hevc", 2160}: "-crf 26", {"hevc", 4320}: "-crf 26"},
			false,
		},
		{
			"below a resolution, next to it and up",
			[]string{"av1[<720]:-crf 28", "av1[720+]:-crf 32"},
			map[codecLine]string{
				{"av1", 144}: "-crf 28", {"av1", 240}: "-crf 28", {"av1", 360}: "-crf 28", {"av1", 480}: "-crf 28",
				{"av1", 720}: "-crf 32", {"av1", 1080}: "-crf 32", {"av1", 1440}: "-crf 32", {"av1", 2160}: "-crf 32", {"av1", 4320}: "-crf 32",
			},
			false,
		},
		{"no codec", []string{"-preset 6"}, nil, true},
		{"args with colon but no codec", []string{"-vf scale=1280:-2"}, nil, true},
		{"option with its own colon but no codec", []string{"-b:v 3M"}, nil, true},
		{"codec named twice", []string{"av1:-crf 30", "av1:-crf 35"}, nil, true},
		{"resolution named twice", []string{"av1[1080]:-crf 30", "av1[1080p]:-crf 35"}, nil, true},
		{"overlapping resolutions", []string{"av1[720-1080]:-crf 30", "av1[1080+]:-crf 35"}, nil, true},
		{"resolution not a tier", []string{"av1[1000]:-crf 30"}, nil, true},
		{"resolution not a number", []string{"av1[big]:-crf 30"}, nil, true},
		{"resolution of zero", []string{"av1[0]:-crf 30"}, nil, true},
		{"empty resolution", []string{"av1[]:-crf 30"}, nil, true},
		{"range the wrong way round", []string{"av1[1080-720]:-crf 30"}, nil, true},
		{"range and up at once", []string{"av1[720-1080+]:-crf 30"}, nil, true},
		{"below and up at once", []string{"av1[<720+]:-crf 30"}, nil, true},
		{"below and a range at once", []string{"av1[<720-1080]:-crf 30"}, nil, true},
		{"below the smallest", []string{"av1[<144]:-crf 30"}, nil, true},
		{"below a resolution that isn't a tier", []string{"av1[<1000]:-crf 30"}, nil, true},
		{"below a resolution overlapping it", []string{"av1[<720]:-crf 30", "av1[480]:-crf 28"}, nil, true},
		{"resolution not closed", []string{"av1[1080:-crf 30"}, nil, true},
		{"resolution with no codec", []string{"[1080]:-crf 30"}, nil, true},
		{"empty codec", []string{":-crf 35"}, nil, true},
		{"empty value", []string{"av1:"}, nil, true},
		{"empty value for a resolution", []string{"av1[1080]:"}, nil, true},
	}

	for _, tt := range tests {
		byCodec, err := parseCodecArgLines(tt.lines, sharedvalidation.ValidateVideoCodec)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: error = %v, want error %v", tt.name, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if !maps.Equal(byCodec, tt.wantByCodec) {
			t.Errorf("%s: got %v, want %v", tt.name, byCodec, tt.wantByCodec)
		}
	}
}

func TestParseVideoEncoderArgs(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    map[string]map[int]sharedmodels.EncoderArgs
		wantErr bool
	}{
		{
			"crf and pixel format taken out",
			[]string{"av1:-crf 35 -pix_fmt yuv420p10le -preset 6"},
			map[string]map[int]sharedmodels.EncoderArgs{"av1": {0: {Quality: "35", PixFmt: "yuv420p10le", Args: []string{"-preset", "6"}}}},
			false,
		},
		{
			"video stream specifiers",
			[]string{"hevc:-crf:v 24 -pix_fmt:v:0 YUV444P"},
			map[string]map[int]sharedmodels.EncoderArgs{"hevc": {0: {Quality: "24", PixFmt: "yuv444p"}}},
			false,
		},
		{
			"options for other streams pass through",
			[]string{"hevc:-pix_fmt:v:1 yuvj420p -x265-params crf=20"},
			map[string]map[int]sharedmodels.EncoderArgs{"hevc": {0: {Args: []string{"-pix_fmt:v:1", "yuvj420p", "-x265-params", "crf=20"}}}},
			false,
		},
		{
			"quoted args and video filter",
			[]string{`av1:-svtav1-params "tune=0:film-grain=8" -vf scale=1280:-2,fps=30`},
			map[string]map[int]sharedmodels.EncoderArgs{"av1": {0: {VideoFilter: "scale=1280:-2,fps=30", Args: []string{"-svtav1-params", "tune=0:film-grain=8"}}}},
			false,
		},
		{
			"filter:v taken out, audio filter passed through",
			[]string{"hevc:-filter:v:0 scale=1280:-2 -af volume=2"},
			map[string]map[int]sharedmodels.EncoderArgs{"hevc": {0: {VideoFilter: "scale=1280:-2", Args: []string{"-af", "volume=2"}}}},
			false,
		},
		{
			"crf clamped per codec",
			[]string{"av1:-crf 70", "h264:-crf 70"},
			map[string]map[int]sharedmodels.EncoderArgs{"av1": {0: {Quality: "63"}}, "h264": {0: {Quality: "51"}}},
			false,
		},
		{
			"passes with a bitrate",
			[]string{"h264:-b:v 3M -pass 3", "av1:-b:v:0 2M -pass 3", "vp9:-crf 31 -b:v 0 -pass 2", "hevc:-b:v 4M -pass 1"},
			map[string]map[int]sharedmodels.EncoderArgs{
				"h264": {0: {Passes: 3, Bitrate: "3000000"}},
				"av1":  {0: {Passes: 2, Bitrate: "2000000"}},
				"vp9":  {0: {Quality: "31", Passes: 2, Bitrate: "0"}},
				"hevc": {0: {Passes: 1, Bitrate: "4000000"}},
			},
			false,
		},
		{
			"quality with a bitrate where that's fine, or no target",
			[]string{"vp9:-crf 31 -b:v 2M", "av1:-crf 30 -b:v 0", "h264:-b:v 1k", "hevc:-b:v 2500000 -preset slow"},
			map[string]map[int]sharedmodels.EncoderArgs{
				"vp9":  {0: {Quality: "31", Bitrate: "2000000"}},
				"av1":  {0: {Quality: "30", Bitrate: "0"}},
				"h264": {0: {Bitrate: "1000"}},
				"hevc": {0: {Bitrate: "2500000", Args: []string{"-preset", "slow"}}},
			},
			false,
		},
		{
			"passes clamped",
			[]string{"mpeg2:-b:v 8M -pass 9", "vp8:-crf 10 -b:v 1M -pass 0"},
			map[string]map[int]sharedmodels.EncoderArgs{
				"mpeg2": {0: {Passes: 3, Bitrate: "8000000"}},
				"vp8":   {0: {Quality: "10", Passes: 1, Bitrate: "1000000"}},
			},
			false,
		},
		{
			"resolution lines laid over the plain line",
			[]string{
				"av1:-crf 30 -preset 6 -pix_fmt yuv420p10le -svtav1-params tune=0",
				"av1[1080p]:-preset:v 8 -g 240",
				"av1[4K]:-vf scale=-2:1080 -svtav1-params film-grain=8",
			},
			map[string]map[int]sharedmodels.EncoderArgs{"av1": {
				0:    {Quality: "30", PixFmt: "yuv420p10le", Args: []string{"-preset", "6", "-svtav1-params", "tune=0"}},
				1080: {Quality: "30", PixFmt: "yuv420p10le", Args: []string{"-svtav1-params", "tune=0", "-preset:v", "8", "-g", "240"}},
				2160: {Quality: "30", PixFmt: "yuv420p10le", VideoFilter: "scale=-2:1080", Args: []string{"-preset", "6", "-svtav1-params", "film-grain=8"}},
			}},
			false,
		},
		{
			"a range laid over the plain line for each tier",
			[]string{"hevc:-crf 24 -preset slow", "hevc[1440-4k]:-preset medium"},
			map[string]map[int]sharedmodels.EncoderArgs{"hevc": {
				0:    {Quality: "24", Args: []string{"-preset", "slow"}},
				1440: {Quality: "24", Args: []string{"-preset", "medium"}},
				2160: {Quality: "24", Args: []string{"-preset", "medium"}},
			}},
			false,
		},
		{
			"a resolution line's quality or bitrate replaces the plain line's either where the encoder takes one",
			[]string{"av1:-crf 30 -preset 6", "av1[4k]:-b:v 12M", "h264:-b:v 3M", "h264[720]:-crf 23"},
			map[string]map[int]sharedmodels.EncoderArgs{
				"av1":  {0: {Quality: "30", Args: []string{"-preset", "6"}}, 2160: {Bitrate: "12000000", Args: []string{"-preset", "6"}}},
				"h264": {0: {Bitrate: "3000000"}, 720: {Quality: "23"}},
			},
			false,
		},
		{
			"a resolution line's quality or bitrate is laid over separately where the encoder takes both",
			[]string{"vp9:-crf 31 -b:v 0", "vp9[1080]:-crf 33", "vp9[4k]:-b:v 8M"},
			map[string]map[int]sharedmodels.EncoderArgs{"vp9": {
				0:    {Quality: "31", Bitrate: "0"},
				1080: {Quality: "33", Bitrate: "0"},
				2160: {Quality: "31", Bitrate: "8000000"},
			}},
			false,
		},
		{
			"a resolution line's passes and negative values",
			[]string{"h264:-b:v 3M -pass 2 -bf -1", "h264[1080]:-pass 1 -bf 2"},
			map[string]map[int]sharedmodels.EncoderArgs{"h264": {
				0:    {Passes: 2, Bitrate: "3000000", Args: []string{"-bf", "-1"}},
				1080: {Passes: 1, Bitrate: "3000000", Args: []string{"-bf", "2"}},
			}},
			false,
		},
		{
			"resolution line without a plain line",
			[]string{"hevc[2160]:-crf 26"},
			map[string]map[int]sharedmodels.EncoderArgs{"hevc": {2160: {Quality: "26"}}},
			false,
		},
		{
			"shell characters stay in the args",
			[]string{"av1:-svtav1-params a=1;b=2 -crf 30"},
			map[string]map[int]sharedmodels.EncoderArgs{"av1": {0: {Quality: "30", Args: []string{"-svtav1-params", "a=1;b=2"}}}},
			false,
		},
		{"bitrate nan", []string{"av1:-b:v nan"}, nil, true},
		{"bitrate infinity", []string{"av1:-b:v inf"}, nil, true},
		{"bitrate infinity with unit", []string{"av1:-b:v InfM"}, nil, true},
		{"bitrate too large", []string{"av1:-b:v 1e30"}, nil, true},
		{"bitrate too large with unit", []string{"av1:-b:v 1e16G"}, nil, true},
		{"old bitrate per resolution", []string{"av1:-b:v 720=1M,1080=3M"}, nil, true},
		{"bitrate not a number", []string{"av1:-b:v fast"}, nil, true},
		{"bitrate twice", []string{"av1:-b:v 3M -b:v 4M"}, nil, true},
		{"zero bitrate with two passes on h264", []string{"h264:-b:v 0 -pass 2"}, nil, true},
		{"crf with two passes on h264", []string{"h264:-crf 23 -b:v 3M -pass 2"}, nil, true},
		{"crf with a bitrate on av1", []string{"av1:-crf 30 -b:v 3M"}, nil, true},
		{"crf with a bitrate on h264", []string{"h264:-crf 23 -b:v 3M"}, nil, true},
		{"crf with a bitrate on hevc", []string{"hevc:-crf 23 -b:v 3M"}, nil, true},
		{"crf with a bitrate in one resolution line", []string{"av1:-preset 6", "av1[4k]:-crf 30 -b:v 3M"}, nil, true},
		{"two passes from the plain line, no bitrate for a resolution", []string{"hevc:-pass 2 -b:v 3M", "hevc[1080]:-crf 24"}, nil, true},
		{"two passes from a resolution line, no bitrate on the plain line", []string{"av1:-preset 6", "av1[1080]:-pass 2"}, nil, true},
		{"overlapping resolution lines", []string{"av1[720-1080]:-preset 8", "av1[1080+]:-preset 10"}, nil, true},
		{"a resolution line with no codec", []string{"[720+]:-preset 8"}, nil, true},
		{"codec option", []string{"h264:-c:v libx265"}, nil, true},
		{"codec option in a resolution line", []string{"h264[1080]:-c:v libx265"}, nil, true},
		{"bitrate too low", []string{"vp8:-b:v 100"}, nil, true},
		{"two passes without a bitrate on hevc", []string{"hevc:-preset slow -pass 2"}, nil, true},
		{"audio bitrate isn't a target bitrate", []string{"av1:-b:a 128k -pass 2"}, nil, true},
		{"pass not a number", []string{"h264:-b:v 3M -pass two"}, nil, true},
		{"pass twice", []string{"h264:-b:v 3M -pass 2 -pass 3"}, nil, true},
		{"crf twice", []string{"av1:-crf 30 -crf 35"}, nil, true},
		{"two video filters", []string{"av1:-vf scale=1280:-2 -filter:v fps=30"}, nil, true},
		{"pixel format twice", []string{"av1:-pix_fmt yuv420p -pix_fmt yuv420p10le"}, nil, true},
		{"crf without value", []string{"av1:-preset 6 -crf"}, nil, true},
		{"crf not a number", []string{"av1:-crf high"}, nil, true},
		{"mpeg2 has no crf", []string{"mpeg2:-crf 20"}, nil, true},
		{"blocked option", []string{"av1:-report"}, nil, true},
		{"unbalanced quote", []string{`av1:-svtav1-params "tune=0`}, nil, true},
		{"no codec", []string{"-crf 35"}, nil, true},
	}

	for _, tt := range tests {
		got, err := ParseVideoEncoderArgs(tt.lines)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: error = %v, want error %v", tt.name, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

// TestNormalizeResolutions tests each resolution is saved in pixels, in args lines and codec entries, and the rest as typed.
func TestNormalizeResolutions(t *testing.T) {
	for in, want := range map[string]string{
		"av1:-crf 30":                "av1:-crf 30",
		"av1[1080p]:-crf 28":         "av1[1080]:-crf 28",
		"av1[1080P]:-crf 28":         "av1[1080]:-crf 28",
		"AV1 [ 720i ]:-crf 32":       "AV1[720]:-crf 32",
		"hevc[4K]:-vf scale=-2:1080": "hevc[2160]:-vf scale=-2:1080",
		"av1[720p - 4K]:-preset 8":   "av1[720-2160]:-preset 8",
		"lc_h264[1440P+]:hevc":       "lc_h264[1440+]:hevc",
		"lc_av1[< 720p]:copy":        "lc_av1[<720]:copy",
		"[8k]:av1":                   "[4320]:av1",
		"copy":                       "copy",
		"av1[bad]:-crf 1":            "av1[bad]:-crf 1",
	} {
		if got := NormalizeResolutions([]string{in}); !slices.Equal(got, []string{want}) {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

// TestParseAudioEncoderArgs tests -af, -ac and -ar are taken out as typed, and other options pass through.
func TestParseAudioEncoderArgs(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    map[string]sharedmodels.AudioEncoderArgs
		wantErr bool
	}{
		{
			"filter, channels and sample rate taken out",
			[]string{"aac:-b:a 128k -af loudnorm -ac 2 -ar 44.1k"},
			map[string]sharedmodels.AudioEncoderArgs{"aac": {Filter: "loudnorm", Channels: "2", SampleRate: "44.1k", Args: []string{"-b:a", "128k"}}},
			false,
		},
		{
			"audio stream specifiers",
			[]string{"opus:-filter:a:0 volume=0.5 -ac:a 6 -vbr on"},
			map[string]sharedmodels.AudioEncoderArgs{"opus": {Filter: "volume=0.5", Channels: "6", Args: []string{"-vbr", "on"}}},
			false,
		},
		{
			"options for other streams pass through",
			[]string{"mp3:-ac:a:1 1 -q:a 2"},
			map[string]sharedmodels.AudioEncoderArgs{"mp3": {Args: []string{"-ac:a:1", "1", "-q:a", "2"}}},
			false,
		},
		{
			"aliases normalized",
			[]string{"libopus:-b:a 96k", "libmp3lame:-q:a 2"},
			map[string]sharedmodels.AudioEncoderArgs{"opus": {Args: []string{"-b:a", "96k"}}, "mp3": {Args: []string{"-q:a", "2"}}},
			false,
		},
		{"copy line", []string{"copy:-b:a 128k"}, nil, true},
		{"video resolution", []string{"aac[1080]:-b:a 128k"}, nil, true},
		{"no codec", []string{"-b:a 128k"}, nil, true},
		{"video codec", []string{"hevc:-b:a 128k"}, nil, true},
		{"filter without value", []string{"aac:-b:a 128k -af"}, nil, true},
		{"two filters", []string{"aac:-af volume=2 -filter:a loudnorm"}, nil, true},
		{"channels twice", []string{"aac:-ac 2 -ac 1"}, nil, true},
		{"sample rate twice", []string{"aac:-ar 48000 -ar:a 44100"}, nil, true},
		{"codec option", []string{"aac:-c:a libopus"}, nil, true},
		{"blocked option", []string{"aac:-report"}, nil, true},
		{"unbalanced quote", []string{`aac:-af "volume=2`}, nil, true},
	}

	for _, tt := range tests {
		got, err := ParseAudioEncoderArgs(tt.lines)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: error = %v, want error %v", tt.name, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestEncoderOnlyOption(t *testing.T) {
	tests := []struct {
		args      []string
		want      string
		wantFound bool
	}{
		{[]string{"-movflags", "+faststart", "-map", "0:v"}, "", false},
		{[]string{"-movflags", "+faststart", "-crf", "30"}, "-crf", true},
		{[]string{"-pix_fmt:v:0", "yuv420p10le"}, "-pix_fmt:v:0", true},
		{[]string{"-vf", "scale=1280:-2"}, "-vf", true},
		{[]string{"-filter:v", "fps=30"}, "-filter:v", true},
		{[]string{"-b:a", "128k", "-pass", "2"}, "-pass", true},
		{[]string{"-b:v", "3M"}, "-b:v", true},
		{[]string{"-filter:a", "volume=2", "-filter_complex", "[0:v]null"}, "", false},
	}

	for _, tt := range tests {
		got, found := EncoderOnlyOption(tt.args)
		if got != tt.want || found != tt.wantFound {
			t.Errorf("EncoderOnlyOption(%q) = %q, %v; want %q, %v", tt.args, got, found, tt.want, tt.wantFound)
		}
	}
}
