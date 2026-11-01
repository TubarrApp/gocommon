package sharedparsing_test

import (
	"reflect"
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
)

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
				"h264": {Passes: 3, Args: []string{"-b:v", "3M"}},
				"av1":  {Passes: 2, Args: []string{"-b:v", "2M"}},
				"vp9":  {Quality: "31", Passes: 2, Args: []string{"-b:v", "0"}},
				"hevc": {Passes: 1, Args: []string{"-b:v", "4M"}},
			},
			false,
		},
		{
			"passes clamped",
			[]string{"mpeg2:-b:v 8M -pass 9", "vp8:-crf 10 -b:v 1M -pass 0"},
			map[string]sharedmodels.EncoderArgs{
				"mpeg2": {Passes: 3, Args: []string{"-b:v", "8M"}},
				"vp8":   {Quality: "10", Passes: 1, Args: []string{"-b:v", "1M"}},
			},
			false,
		},
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
		got, err := sharedparsing.ParseVideoEncoderArgs(tt.lines)
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
		{[]string{"-b:v", "3M", "-pass", "2"}, "-pass", true},
		{[]string{"-filter:a", "volume=2", "-filter_complex", "[0:v]null"}, "", false},
	}

	for _, tt := range tests {
		got, found := sharedparsing.EncoderOnlyOption(tt.args)
		if got != tt.want || found != tt.wantFound {
			t.Errorf("EncoderOnlyOption(%q) = %q, %v; want %q, %v", tt.args, got, found, tt.want, tt.wantFound)
		}
	}
}
