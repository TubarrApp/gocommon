package sharedvalidation

import (
	"slices"
	"strings"
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// TestValidateCodecsForContainer tests that codecs a container can't hold are refused, naming where they fit.
func TestValidateCodecsForContainer(t *testing.T) {
	for _, c := range []struct {
		ext           string
		video, audio  []string
		wantErr, hint string
	}{
		{ext: "webm", video: []string{"vp9"}, audio: []string{"opus"}},
		{ext: ".MKV", video: []string{"h264", "hevc:h264"}, audio: []string{"aac"}},
		{ext: "mp4", video: []string{"copy"}, audio: []string{"copy", ""}},
		{ext: "xyz", video: []string{"h264"}}, // Containers missing from the tables aren't checked.
		{ext: "webm", video: []string{"h264"}, wantErr: `video codec "h264" can't go in a .webm file`, hint: ".mp4"},
		{ext: "webm", video: []string{"vp9:libx264"}, wantErr: `video codec "h264"`},
		{ext: "webm", audio: []string{"aac"}, wantErr: `audio codec "aac" can't go in a .webm file`, hint: ".mp4"},
		{ext: "mkv", audio: []string{"dts", "truehd"}},
		{ext: "webm", audio: []string{"truehd"}, wantErr: `audio codec "truehd" can't go in a .webm file`, hint: ".mkv"},
		{ext: "mp4", video: []string{"bogus"}, wantErr: `video codec "bogus" is not valid`},
	} {
		err := ValidateCodecsForContainer(c.ext, c.video, c.audio)
		switch {
		case c.wantErr == "" && err != nil:
			t.Errorf("%s %v %v: %v", c.ext, c.video, c.audio, err)
		case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr) || !strings.Contains(err.Error(), c.hint)):
			t.Errorf("%s %v %v: got %v, want %q naming %q", c.ext, c.video, c.audio, err, c.wantErr, c.hint)
		}
	}
}

// TestFittingCodecs tests that the codecs a container can hold are listed, most preferred first.
func TestFittingCodecs(t *testing.T) {
	for _, c := range []struct {
		name      string
		got, want []string
	}{
		{"webm video", FittingVideoCodecs(".webm"), []string{sharedconsts.VCodecVP9, sharedconsts.VCodecVP8, sharedconsts.VCodecAV1}},
		{"webm audio", FittingAudioCodecs(".webm"), []string{sharedconsts.ACodecOpus, sharedconsts.ACodecVorbis}},
		{"mp4 video", FittingVideoCodecs(".mp4"), []string{sharedconsts.VCodecH264, sharedconsts.VCodecHEVC, sharedconsts.VCodecVP9, sharedconsts.VCodecMPEG2, sharedconsts.VCodecAV1}},
		{"ogm video", FittingVideoCodecs(".ogm"), []string{sharedconsts.VCodecVP8}},
		{"rm video", FittingVideoCodecs(".rm"), nil},
	} {
		if !slices.Equal(c.got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestValidateFFmpegOutputExt tests that known containers are accepted as output, and unknown ones refused.
func TestValidateFFmpegOutputExt(t *testing.T) {
	for _, ext := range []string{"webm", "ogm", "ts", "m4v"} {
		if _, err := ValidateFFmpegOutputExt(ext); err != nil {
			t.Errorf("ValidateFFmpegOutputExt(%q): %v", ext, err)
		}
	}
	for _, ext := range []string{"xyz", "rm", "rmvb"} {
		if _, err := ValidateFFmpegOutputExt(ext); err == nil {
			t.Errorf("ValidateFFmpegOutputExt(%q) accepted it", ext)
		}
	}
}

// TestContainerTablesListValidCodecs tests that every codec in the tables is a valid one.
func TestContainerTablesListValidCodecs(t *testing.T) {
	for ext, codecs := range sharedconsts.ContainerCompatibleVideoCodecs {
		for _, c := range codecs {
			if _, ok := sharedconsts.ValidVideoCodecs[c]; !ok {
				t.Errorf("%s lists unknown video codec %q", ext, c)
			}
		}
	}
	for ext, codecs := range sharedconsts.ContainerCompatibleAudioCodecs {
		for _, c := range codecs {
			if _, ok := sharedconsts.ValidAudioCodecs[c]; !ok {
				t.Errorf("%s lists unknown audio codec %q", ext, c)
			}
		}
	}
}
