package sharedparsing

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedmodels"
)

// TestCheckEncoderSettings tests the settings are read and judged together, with the software checks only when software encoders run.
func TestCheckEncoderSettings(t *testing.T) {
	for _, c := range []struct {
		name string
		s    sharedmodels.EncoderSettings
		ok   bool
	}{
		{name: "nothing set", ok: true},
		{name: "working settings", s: sharedmodels.EncoderSettings{
			VideoArgs: []string{"av1:-preset 8", "h264:-crf 20"}, VideoCodecs: []string{"h264:av1"}, AudioCodecs: []string{"aac"}, ExtraArgs: "-b:a 128k", OutputExt: ".mkv", Software: true,
		}, ok: true},

		// Lines that do nothing.
		{name: "copy line", s: sharedmodels.EncoderSettings{VideoArgs: []string{"copy:-crf 20"}}},
		{name: "another encoder's option", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-x265-params log-level=error"}}},
		{name: "line for copied video", s: sharedmodels.EncoderSettings{VideoArgs: []string{"h264:-crf 20"}, VideoCodecs: []string{"copy"}}},

		// Values software encoders can't take, only judged when one runs.
		{name: "AV1 preset name in software", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium"}, Software: true}},
		{name: "AV1 preset name otherwise", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset medium"}}, ok: true},
		{name: "VP9 profile name in software", s: sharedmodels.EncoderSettings{VideoArgs: []string{"vp9:-profile:v main"}, Software: true}},
		{name: "VP9 profile name otherwise", s: sharedmodels.EncoderSettings{VideoArgs: []string{"vp9:-profile:v main"}}, ok: true},

		// Later lines are still judged, whichever sorts first.
		{name: "a bad line after a good one", s: sharedmodels.EncoderSettings{VideoArgs: []string{"av1:-preset 8", "vp9:-profile:v main"}, Software: true}},

		// Audio only ever copied.
		{name: "filter with no audio codec", s: sharedmodels.EncoderSettings{ExtraArgs: "-af volume=0.5"}},
		{name: "option with no audio codec", s: sharedmodels.EncoderSettings{ExtraArgs: "-b:a 128k"}},
		{name: "option with plain copy", s: sharedmodels.EncoderSettings{ExtraArgs: "-ac 1", AudioCodecs: []string{"copy"}}},

		// Audio that can be encoded.
		{name: "filter with an audio codec", s: sharedmodels.EncoderSettings{ExtraArgs: "-af volume=0.5", AudioCodecs: []string{"aac"}}, ok: true},
		{name: "filter with a pair", s: sharedmodels.EncoderSettings{ExtraArgs: "-af volume=0.5", AudioCodecs: []string{"dts:ac3"}}, ok: true},
		{name: "filter when the container may force a codec", s: sharedmodels.EncoderSettings{ExtraArgs: "-af volume=0.5", OutputExt: ".mp4"}, ok: true},

		// Settings that can't be read.
		{name: "a line with no codec", s: sharedmodels.EncoderSettings{VideoArgs: []string{"-crf 20"}}},
		{name: "a bad video codec", s: sharedmodels.EncoderSettings{VideoCodecs: []string{"bad"}}},
		{name: "a bad audio codec", s: sharedmodels.EncoderSettings{AudioCodecs: []string{"bad"}}},
		{name: "extra args with an open quote", s: sharedmodels.EncoderSettings{ExtraArgs: `-metadata "title`}},
	} {
		if err := CheckEncoderSettings(c.s); (err == nil) != c.ok {
			t.Errorf("%s: got %v, want ok %t", c.name, err, c.ok)
		}
	}
}
