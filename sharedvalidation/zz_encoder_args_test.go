package sharedvalidation_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// parseLine returns a "codec:args" line's codec and args.
func parseLine(t *testing.T, codec, line string) sharedmodels.EncoderArgs {
	t.Helper()
	parsed, err := sharedparsing.ParseVideoEncoderArgs([]string{line})
	if err != nil {
		t.Fatalf("%q: %v", line, err)
	}
	return parsed[codec]
}

// TestValidateSoftwareEncoderArgs tests only a word where a software encoder takes a number is refused, as anything else might work on some FFmpeg version.
func TestValidateSoftwareEncoderArgs(t *testing.T) {
	for _, c := range []struct {
		codec string
		line  string
		ok    bool
	}{
		// A word where the encoder takes a number.
		{codec: "av1", line: "av1:-preset medium"},
		{codec: "av1", line: "av1:-preset ultrafast"},
		{codec: "av1", line: "av1:-preset p5"},
		{codec: "vp9", line: "vp9:-profile:v main"},
		{codec: "vp9", line: "vp9:-profile main"},
		{codec: "vp9", line: "vp9:-profile:v:0 main"},
		{codec: "vp8", line: "vp8:-profile:v high"},

		// Numbers, and what those encoders may take.
		{codec: "av1", line: "av1:-preset 8", ok: true},
		{codec: "av1", line: "av1:-preset -1", ok: true},
		{codec: "av1", line: "av1:-preset 99", ok: true},
		{codec: "vp9", line: "vp9:-profile:v 2", ok: true},
		{codec: "vp9", line: "vp9:-profile:v 9", ok: true},

		// The same words on encoders that do take them.
		{codec: "h264", line: "h264:-preset medium -profile:v main", ok: true},
		{codec: "hevc", line: "hevc:-preset medium -profile:v main", ok: true},

		// Never measured.
		{codec: "av1", line: "av1:-preset somefuturepreset", ok: true},
		{codec: "vp9", line: "vp9:-profile:v somefutureprofile", ok: true},

		// Fail on FFmpeg 9 with these builds, but that could change.
		{codec: "av1", line: "av1:-profile:v high", ok: true},
		{codec: "h264", line: "h264:-preset p5", ok: true},
		{codec: "vp9", line: "vp9:-tune film", ok: true},
		{codec: "vp8", line: "vp8:-pix_fmt yuv420p10le", ok: true},
	} {
		if err := sharedvalidation.ValidateSoftwareEncoderArgs(c.codec, parseLine(t, c.codec, c.line)); (err == nil) != c.ok {
			t.Errorf("%q: got %v, want ok %t", c.line, err, c.ok)
		}
	}
}

// TestValidateEncoderArgsUsed tests a "copy" line and options for another encoder are refused, as they do nothing on any encoder.
func TestValidateEncoderArgsUsed(t *testing.T) {
	for _, c := range []struct {
		codec string
		line  string
		ok    bool
	}{
		{codec: "copy", line: "copy:-crf 20"},
		{codec: "copy", line: "copy:-pix_fmt yuv420p10le"},
		{codec: "vp9", line: "vp9:-x265-params log-level=error"},
		{codec: "h264", line: "h264:-svtav1-params film-grain=8"},
		{codec: "hevc", line: "hevc:-x264-params keyint=12"},
		{codec: "av1", line: "av1:-x264-params keyint=12"},
		{codec: "h264", line: "h264:-x264-params keyint=12", ok: true},
		{codec: "hevc", line: "hevc:-x265-params log-level=error", ok: true},
		{codec: "av1", line: "av1:-svtav1-params film-grain=8", ok: true},
		{codec: "h264", line: "h264:-crf 20 -row-mt 1", ok: true},
	} {
		if err := sharedvalidation.ValidateEncoderArgsUsed(c.codec, parseLine(t, c.codec, c.line)); (err == nil) != c.ok {
			t.Errorf("%q: got %v, want ok %t", c.line, err, c.ok)
		}
	}
}

// TestValidateEncoderArgsCanRun tests a codec's line is refused when its videos are copied and nothing encodes to it.
func TestValidateEncoderArgsCanRun(t *testing.T) {
	for _, c := range []struct {
		name  string
		video []string
		ext   string
		ok    bool
	}{
		{name: "plain copy", video: []string{"copy"}},
		{name: "plain copy with a container set", video: []string{"copy"}, ext: ".mp4"},
		{name: "plain copy with another source encoded elsewhere", video: []string{"copy", "vp9:hevc"}},
		{name: "a pair to copy", video: []string{"h264:copy"}},
		{name: "a pair to copy with the same container", video: []string{"h264:copy"}, ext: sharedconsts.ExtSame},
		{name: "a pair to copy beside another plain codec", video: []string{"hevc", "h264:copy"}},
		{name: "a pair to copy beside another plain codec with a container set", video: []string{"hevc", "h264:copy"}, ext: ".mp4"},

		{name: "plain copy with a pair encoding to it", video: []string{"copy", "vp9:h264"}, ok: true},
		{name: "plain copy with its own pair", video: []string{"copy", "h264:h264"}, ok: true},
		{name: "a pair to copy with a pair encoding to it", video: []string{"h264:copy", "vp9:h264"}, ok: true},
		{name: "a pair to copy when a container may choose this codec", video: []string{"h264:copy"}, ext: ".mp4", ok: true},
		{name: "a pair to copy beside its own plain codec, which others encode to", video: []string{"h264", "h264:copy"}, ok: true},
		{name: "a level's pair to copy, as other levels can still encode", video: []string{"lc_h264:copy"}, ok: true},
		{name: "plain own codec", video: []string{"h264"}, ok: true},
		{name: "a pair for another source", video: []string{"vp9:hevc"}, ok: true},
		{name: "nothing named", ok: true},
	} {
		plain, pairs, err := sharedparsing.ParseCodecMappings(c.video, sharedvalidation.ValidateVideoCodec)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if err := sharedvalidation.ValidateEncoderArgsCanRun("h264", plain, pairs, c.ext); (err == nil) != c.ok {
			t.Errorf("%s: got %v, want ok %t", c.name, err, c.ok)
		}
	}
}

// TestAudioAlwaysCopied tests when no setting could encode audio.
func TestAudioAlwaysCopied(t *testing.T) {
	for _, c := range []struct {
		name  string
		audio []string
		ext   string
		want  bool
	}{
		{name: "no codec", want: true},
		{name: "no codec with the same container", ext: sharedconsts.ExtSame, want: true},
		{name: "plain copy", audio: []string{"copy"}, want: true},
		{name: "plain copy with a container set", audio: []string{"copy"}, ext: ".mp4", want: true},
		{name: "no codec when the container may force one", ext: ".mp4"},
		{name: "a plain codec", audio: []string{"aac"}},
		{name: "a pair, which depends on the video", audio: []string{"dts:ac3"}},
	} {
		plain, pairs, err := sharedparsing.ParseCodecMappings(c.audio, sharedvalidation.ValidateAudioCodec)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := sharedvalidation.AudioAlwaysCopied(plain, pairs, c.ext); got != c.want {
			t.Errorf("%s: got %t, want %t", c.name, got, c.want)
		}
	}
}

// TestCopiedAudioOptions tests audio filters and audio encoding options are told apart, as only filters stop FFmpeg.
func TestCopiedAudioOptions(t *testing.T) {
	for _, c := range []struct {
		name       string
		extra      []string
		wantFilter bool
		wantOption bool
	}{
		{name: "-af", extra: []string{"-af", "volume=0.5"}, wantFilter: true},
		{name: "-filter:a", extra: []string{"-filter:a", "volume=0.5"}, wantFilter: true},
		{name: "first stream's filter", extra: []string{"-filter:a:0", "volume=0.5"}, wantFilter: true},
		{name: "-ac", extra: []string{"-ac", "1"}, wantOption: true},
		{name: "-ar", extra: []string{"-ar", "44100"}, wantOption: true},
		{name: "-b:a", extra: []string{"-b:a", "128k"}, wantOption: true},
		{name: "first stream's bitrate", extra: []string{"-b:a:0", "128k"}, wantOption: true},
		{name: "-q:a", extra: []string{"-q:a", "2"}, wantOption: true},
		{name: "settings that don't touch audio", extra: []string{"-movflags", "+faststart", "-threads", "1"}},
		{name: "a value that looks like an option's name", extra: []string{"-metadata", "title=-af"}},
		{name: "nothing"},
	} {
		if err := sharedvalidation.ValidateCopiedAudioFilter(c.extra); (err != nil) != c.wantFilter {
			t.Errorf("%s filter: got %v, want %t", c.name, err, c.wantFilter)
		}
		if err := sharedvalidation.ValidateCopiedAudioOptions(c.extra); (err != nil) != c.wantOption {
			t.Errorf("%s option: got %v, want %t", c.name, err, c.wantOption)
		}
	}
}
