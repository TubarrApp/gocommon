package sharedvalidation

import (
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

// TestFittingCodecs tests that the preferred codec a container can hold, and is usable, is picked.
func TestFittingCodecs(t *testing.T) {
	all := func(string) bool { return true }
	if got := FittingVideoCodec(".webm", all); got != sharedconsts.VCodecVP9 {
		t.Errorf("FittingVideoCodec(.webm) = %q, want vp9", got)
	}
	if got := FittingVideoCodec(".webm", func(c string) bool { return c != sharedconsts.VCodecVP9 }); got != sharedconsts.VCodecVP8 {
		t.Errorf("FittingVideoCodec(.webm) without vp9 = %q, want vp8", got)
	}
	if got := FittingAudioCodec(".webm", all); got != sharedconsts.ACodecOpus {
		t.Errorf("FittingAudioCodec(.webm) = %q, want opus", got)
	}
	if got := FittingVideoCodec(".mp4", all); got != sharedconsts.VCodecH264 {
		t.Errorf("FittingVideoCodec(.mp4) = %q, want h264", got)
	}
	if got := FittingVideoCodec(".ogm", all); got != "" {
		t.Errorf("FittingVideoCodec(.ogm) = %q, want none", got)
	}
}

// TestValidateFFmpegOutputExtUnwritable tests that containers ffmpeg can't write are refused as output.
func TestValidateFFmpegOutputExtUnwritable(t *testing.T) {
	for _, ext := range []string{"ogm", "rmvb"} {
		if _, err := ValidateFFmpegOutputExt(ext); err == nil {
			t.Errorf("ValidateFFmpegOutputExt(%q) accepted it", ext)
		}
	}
	if _, err := ValidateFFmpegOutputExt("webm"); err != nil {
		t.Errorf("ValidateFFmpegOutputExt(webm): %v", err)
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
