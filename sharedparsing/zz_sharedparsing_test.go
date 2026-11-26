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

		// Size lines are judged with the plain line under them.
		{name: "a bad size line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-crf 30", "av1[4k]:-preset medium"}, Software: true}},
		{name: "another encoder's option in a size line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-crf 20", "h264[1080]:-x265-params log-level=error"}}},
		{name: "a size line with no plain line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1[1080]:-preset 8"}, Software: true}, ok: true},
		{name: "a bad plain line a size line overrides", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium", "av1[1080]:-preset 8"}, Software: true}},

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

		// Fail states.
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
		all, pairs, err := ParseCodecMappings([]string{tt.in}, sharedvalidation.ValidateAudioCodec)
		if err != nil {
			t.Errorf("ParseCodecMappings(%q) error = %v", tt.in, err)
			continue
		}
		if all != tt.wantAll || !maps.Equal(pairs, tt.wantPairs) {
			t.Errorf("ParseCodecMappings(%q) = (%q, %v), want (%q, %v)", tt.in, all, pairs, tt.wantAll, tt.wantPairs)
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
			"sizes beside a plain line",
			[]string{"av1:-preset 6", "av1[1080p]:-crf 30", "AV1 [4K]:-crf 35", "x265[720i]:-crf 24"},
			map[codecLine]string{{"av1", 0}: "-preset 6", {"av1", 1080}: "-crf 30", {"av1", 2160}: "-crf 35", {"hevc", 720}: "-crf 24"},
			false,
		},
		{"size without a plain line", []string{"av1[1080]:-crf 30"}, map[codecLine]string{{"av1", 1080}: "-crf 30"}, false},
		{"no codec", []string{"-preset 6"}, nil, true},
		{"args with colon but no codec", []string{"-vf scale=1280:-2"}, nil, true},
		{"option with its own colon but no codec", []string{"-b:v 3M"}, nil, true},
		{"codec named twice", []string{"av1:-crf 30", "av1:-crf 35"}, nil, true},
		{"size named twice", []string{"av1[1080]:-crf 30", "av1[1080p]:-crf 35"}, nil, true},
		{"size not a number", []string{"av1[big]:-crf 30"}, nil, true},
		{"size of zero", []string{"av1[0]:-crf 30"}, nil, true},
		{"empty size", []string{"av1[]:-crf 30"}, nil, true},
		{"size not closed", []string{"av1[1080:-crf 30"}, nil, true},
		{"size with no codec", []string{"[1080]:-crf 30"}, nil, true},
		{"empty codec", []string{":-crf 35"}, nil, true},
		{"empty value", []string{"av1:"}, nil, true},
		{"empty value for a size", []string{"av1[1080]:"}, nil, true},
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
			"size lines laid over the plain line",
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
			"a size line's quality or bitrate replaces the plain line's either where the encoder takes one",
			[]string{"av1:-crf 30 -preset 6", "av1[4k]:-b:v 12M", "h264:-b:v 3M", "h264[720]:-crf 23"},
			map[string]map[int]sharedmodels.EncoderArgs{
				"av1":  {0: {Quality: "30", Args: []string{"-preset", "6"}}, 2160: {Bitrate: "12000000", Args: []string{"-preset", "6"}}},
				"h264": {0: {Bitrate: "3000000"}, 720: {Quality: "23"}},
			},
			false,
		},
		{
			"a size line's quality or bitrate is laid over separately where the encoder takes both",
			[]string{"vp9:-crf 31 -b:v 0", "vp9[1080]:-crf 33", "vp9[4k]:-b:v 8M"},
			map[string]map[int]sharedmodels.EncoderArgs{"vp9": {
				0:    {Quality: "31", Bitrate: "0"},
				1080: {Quality: "33", Bitrate: "0"},
				2160: {Quality: "31", Bitrate: "8000000"},
			}},
			false,
		},
		{
			"a size line's passes and negative values",
			[]string{"h264:-b:v 3M -pass 2 -bf -1", "h264[1080]:-pass 1 -bf 2"},
			map[string]map[int]sharedmodels.EncoderArgs{"h264": {
				0:    {Passes: 2, Bitrate: "3000000", Args: []string{"-bf", "-1"}},
				1080: {Passes: 1, Bitrate: "3000000", Args: []string{"-bf", "2"}},
			}},
			false,
		},
		{
			"size line without a plain line",
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
		{"crf with a bitrate in one size line", []string{"av1:-preset 6", "av1[4k]:-crf 30 -b:v 3M"}, nil, true},
		{"two passes from the plain line, no bitrate for a size", []string{"hevc:-pass 2 -b:v 3M", "hevc[1080]:-crf 24"}, nil, true},
		{"two passes from a size line, no bitrate on the plain line", []string{"av1:-preset 6", "av1[1080]:-pass 2"}, nil, true},
		{"codec option", []string{"h264:-c:v libx265"}, nil, true},
		{"codec option in a size line", []string{"h264[1080]:-c:v libx265"}, nil, true},
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

// TestNormalizeVideoArgSizes tests each size is saved as its shorter side in pixels, and everything else as typed.
func TestNormalizeVideoArgSizes(t *testing.T) {
	in := []string{"av1:-crf 30", "av1[1080p]:-crf 28", "av1[1080P]:-crf 28", "AV1 [ 720i ]:-crf 32", "hevc[4K]:-vf scale=-2:1080", "av1[2160]:-crf 35", "av1[bad]:-crf 1"}
	want := []string{"av1:-crf 30", "av1[1080]:-crf 28", "av1[1080]:-crf 28", "AV1[720]:-crf 32", "hevc[2160]:-vf scale=-2:1080", "av1[2160]:-crf 35", "av1[bad]:-crf 1"}
	if got := NormalizeVideoArgSizes(in); !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
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
		{"video size", []string{"aac[1080]:-b:a 128k"}, nil, true},
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
