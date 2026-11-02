package sharedparsing

import (
	"maps"
	"reflect"
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

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
		wantByCodec map[string]string
		wantErr     bool
	}{
		{"codec lines", []string{" av1: -crf 35 ", "", "hevc:-preset slow"}, map[string]string{"av1": "-crf 35", "hevc": "-preset slow"}, false},
		{"alias normalized", []string{"x265:-crf 24"}, map[string]string{"hevc": "-crf 24"}, false},
		{"args keep inner colons", []string{"av1:-svtav1-params tune=0:film-grain=8"}, map[string]string{"av1": "-svtav1-params tune=0:film-grain=8"}, false},
		{"no codec", []string{"-preset 6"}, nil, true},
		{"args with colon but no codec", []string{"-vf scale=1280:-2"}, nil, true},
		{"codec named twice", []string{"av1:-crf 30", "av1:-crf 35"}, nil, true},
		{"empty codec", []string{":-crf 35"}, nil, true},
		{"empty value", []string{"av1:"}, nil, true},
	}

	for _, tt := range tests {
		byCodec, err := parseCodecArgLines(tt.lines)
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
		want    map[string]sharedmodels.EncoderArgs
		wantErr bool
	}{
		{
			"crf and pixel format taken out",
			[]string{"av1:-crf 35 -pix_fmt yuv420p10le -preset 6"},
			map[string]sharedmodels.EncoderArgs{"av1": {Quality: "35", PixFmt: "yuv420p10le", Args: []string{"-preset", "6"}}},
			false,
		},
		{
			"video stream specifiers",
			[]string{"hevc:-crf:v 24 -pix_fmt:v:0 YUV444P"},
			map[string]sharedmodels.EncoderArgs{"hevc": {Quality: "24", PixFmt: "yuv444p"}},
			false,
		},
		{
			"options for other streams pass through",
			[]string{"hevc:-pix_fmt:v:1 yuvj420p -x265-params crf=20"},
			map[string]sharedmodels.EncoderArgs{"hevc": {Args: []string{"-pix_fmt:v:1", "yuvj420p", "-x265-params", "crf=20"}}},
			false,
		},
		{
			"quoted args and video filter",
			[]string{`av1:-svtav1-params "tune=0:film-grain=8" -vf scale=1280:-2,fps=30`},
			map[string]sharedmodels.EncoderArgs{"av1": {VideoFilter: "scale=1280:-2,fps=30", Args: []string{"-svtav1-params", "tune=0:film-grain=8"}}},
			false,
		},
		{
			"filter:v taken out, audio filter passed through",
			[]string{"hevc:-filter:v:0 scale=1280:-2 -af volume=2"},
			map[string]sharedmodels.EncoderArgs{"hevc": {VideoFilter: "scale=1280:-2", Args: []string{"-af", "volume=2"}}},
			false,
		},
		{
			"crf clamped per codec",
			[]string{"av1:-crf 70", "h264:-crf 70"},
			map[string]sharedmodels.EncoderArgs{"av1": {Quality: "63"}, "h264": {Quality: "51"}},
			false,
		},
		{
			"passes with a bitrate",
			[]string{"h264:-b:v 3M -pass 3", "av1:-b:v 2M -pass 3", "vp9:-crf 31 -b:v 0 -pass 2", "hevc:-b:v 4M -pass 1"},
			map[string]sharedmodels.EncoderArgs{
				"h264": {Passes: 3, Bitrates: map[int]int64{0: 3_000_000}},
				"av1":  {Passes: 2, Bitrates: map[int]int64{0: 2_000_000}},
				"vp9":  {Quality: "31", Passes: 2, Bitrates: map[int]int64{0: 0}},
				"hevc": {Passes: 1, Bitrates: map[int]int64{0: 4_000_000}},
			},
			false,
		},
		{
			"passes clamped",
			[]string{"mpeg2:-b:v 8M -pass 9", "vp8:-crf 10 -b:v 1M -pass 0"},
			map[string]sharedmodels.EncoderArgs{
				"mpeg2": {Passes: 3, Bitrates: map[int]int64{0: 8_000_000}},
				"vp8":   {Quality: "10", Passes: 1, Bitrates: map[int]int64{0: 1_000_000}},
			},
			false,
		},
		{
			"bitrates by resolution",
			[]string{"av1:-b:v:0 720p=1.5M,1080=3000k,4K=10m -pass 2", "hevc:-b:v 2500000 -preset slow"},
			map[string]sharedmodels.EncoderArgs{
				"av1":  {Passes: 2, Bitrates: map[int]int64{720: 1_500_000, 1080: 3_000_000, 2160: 10_000_000}},
				"hevc": {Bitrates: map[int]int64{0: 2_500_000}, Args: []string{"-preset", "slow"}},
			},
			false,
		},
		{
			"shell characters stay in the args",
			[]string{"av1:-svtav1-params a=1;b=2 -crf 30"},
			map[string]sharedmodels.EncoderArgs{"av1": {Quality: "30", Args: []string{"-svtav1-params", "a=1;b=2"}}},
			false,
		},
		{"bitrate nan", []string{"av1:-b:v nan"}, nil, true},
		{"bitrate infinity", []string{"av1:-b:v inf"}, nil, true},
		{"bitrate infinity with unit", []string{"av1:-b:v InfM"}, nil, true},
		{"bitrate too large", []string{"av1:-b:v 1e30"}, nil, true},
		{"bitrate too large with unit", []string{"av1:-b:v 1e16G"}, nil, true},
		{"resolution named twice", []string{"av1:-b:v 1080=3M,1080p=4M"}, nil, true},
		{"entry without a resolution", []string{"av1:-b:v 1080=3M,2M"}, nil, true},
		{"bitrate not a number", []string{"av1:-b:v 1080=fast"}, nil, true},
		{"bitrate twice", []string{"av1:-b:v 3M -b:v 4M"}, nil, true},
		{"zero bitrate with two passes on h264", []string{"h264:-b:v 0 -pass 2"}, nil, true},
		{"crf with two passes on h264", []string{"h264:-crf 23 -b:v 3M -pass 2"}, nil, true},
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
