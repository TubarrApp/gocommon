package sharedvalidation_test

import (
	"testing"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedparsing"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// ******************************** Encoder args **************************************************************************************

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

// ******************************** Ops ***********************************************************************************************

func TestValidateMetaOps(t *testing.T) {
	tests := []struct {
		name    string
		op      sharedmodels.MetaOps
		wantErr bool
	}{
		{"set", sharedmodels.MetaOps{Field: "title", OpType: "set", OpValue: "cats"}, false},
		{"paste-from", sharedmodels.MetaOps{Field: "title", OpType: "paste-from", OpValue: "x"}, false},
		{"unknown type", sharedmodels.MetaOps{Field: "title", OpType: "explode"}, true},
		{"empty field", sharedmodels.MetaOps{OpType: "set", OpValue: "cats"}, true},
		{"date-tag ok", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "ymd"}, false},
		{"date-tag two-char format", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "md"}, false},
		{"date-tag rejects all", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "all", DateFormat: "ymd"}, true},
		{"date-tag bad format", sharedmodels.MetaOps{Field: "title", OpType: "date-tag", OpLoc: "prefix", DateFormat: "nope"}, true},
		{"delete-date-tag", sharedmodels.MetaOps{Field: "title", OpType: "delete-date-tag"}, false},
		{"delete-date-tag ignores a stored location and format", sharedmodels.MetaOps{Field: "title", OpType: "delete-date-tag", OpLoc: "all", DateFormat: "ymd"}, false},
		{"regex-replace ok", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: `^\d+ \| `}, false},
		{"regex-replace bad pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: "("}, true},
		{"regex-replace empty pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace"}, true},
		{"regex-replace unrepresentable pattern", sharedmodels.MetaOps{Field: "title", OpType: "regex-replace", OpFindString: "[[:digit:]]:"}, true},
		{"replace ok", sharedmodels.MetaOps{Field: "title", OpType: "replace", OpFindString: "a", OpValue: "b"}, false},
		{"replace empty find", sharedmodels.MetaOps{Field: "title", OpType: "replace", OpValue: "b"}, true},
		{"replace-prefix empty find", sharedmodels.MetaOps{Field: "title", OpType: "replace-prefix", OpValue: "b"}, true},
		{"replace-suffix empty find", sharedmodels.MetaOps{Field: "title", OpType: "replace-suffix", OpValue: "b"}, true},
		{"keep-after ok", sharedmodels.MetaOps{Field: "title", OpType: "keep-after", OpValue: "| "}, false},
		{"keep-before empty", sharedmodels.MetaOps{Field: "title", OpType: "keep-before"}, true},
		{"all-titles set", sharedmodels.MetaOps{Field: "all-titles", OpType: "set", OpValue: "cats"}, false},
		{"all-titles suffix", sharedmodels.MetaOps{Field: "all-titles", OpType: "suffix", OpValue: "cats"}, false},
		{"all-titles paste-from", sharedmodels.MetaOps{Field: "all-titles", OpType: "paste-from", OpValue: "release_date"}, false},
		{"copy-to into all-titles", sharedmodels.MetaOps{Field: "release_date", OpType: "copy-to", OpValue: "all-titles"}, false},
		{"all-titles copy-to rejected", sharedmodels.MetaOps{Field: "all-titles", OpType: "copy-to", OpValue: "writer"}, true},
		{"paste-from all-titles rejected", sharedmodels.MetaOps{Field: "writer", OpType: "paste-from", OpValue: "all-titles"}, true},
		{"all-dates paste-from", sharedmodels.MetaOps{Field: "all-dates", OpType: "paste-from", OpValue: "upload_date"}, false},
		{"copy-to into all-dates", sharedmodels.MetaOps{Field: "upload_date", OpType: "copy-to", OpValue: "all-dates"}, false},
		{"all-dates suffix rejected", sharedmodels.MetaOps{Field: "all-dates", OpType: "suffix", OpValue: "x"}, true},
		{"all-dates date-tag rejected", sharedmodels.MetaOps{Field: "all-dates", OpType: "date-tag", OpLoc: "prefix", DateFormat: "ymd"}, true},
		{"all-dates ok", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-05-10"}, false},
		{"all-dates unhyphenated", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "20240510"}, true},
		{"all-dates with time", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-05-10T00:00:00Z"}, true},
		{"all-dates impossible day", sharedmodels.MetaOps{Field: "all-dates", OpType: "set", OpValue: "2024-02-30"}, true},
		{"release_date ok", sharedmodels.MetaOps{Field: "release_date", OpType: "set", OpValue: "2022-05-10"}, false},
		{"premiered single-digit month", sharedmodels.MetaOps{Field: "premiered", OpType: "set", OpValue: "2022-5-10"}, true},
		{"year ok", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "2024"}, false},
		{"year two digits", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "24"}, true},
		{"year full date", sharedmodels.MetaOps{Field: "year", OpType: "set", OpValue: "2024-05-10"}, true},
		{"creation_time ok", sharedmodels.MetaOps{Field: "creation_time", OpType: "set", OpValue: "2024-05-10T12:30:00Z"}, false},
		{"creation_time date only", sharedmodels.MetaOps{Field: "creation_time", OpType: "set", OpValue: "2024-05-10"}, true},
		{"date field suffix unchecked", sharedmodels.MetaOps{Field: "date", OpType: "suffix", OpValue: " x"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateMetaOps([]sharedmodels.MetaOps{tt.op})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetaOps(%+v) err = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateMetaOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilenameOps(t *testing.T) {
	tests := []struct {
		name    string
		op      sharedmodels.FilenameOps
		wantErr bool
	}{
		{"prefix", sharedmodels.FilenameOps{OpType: "prefix", OpValue: "[CATS] "}, false},
		{"no field required", sharedmodels.FilenameOps{OpType: "suffix", OpValue: "(new)"}, false},
		{"append removed", sharedmodels.FilenameOps{OpType: "append", OpValue: "(new)"}, true},
		{"unknown type", sharedmodels.FilenameOps{OpType: "explode"}, true},
		{"copy-to is meta only", sharedmodels.FilenameOps{OpType: "copy-to", OpValue: "x"}, true},
		{"date-tag ok", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "suffix", DateFormat: "dm"}, false},
		{"date-tag rejects all", sharedmodels.FilenameOps{OpType: "date-tag", OpLoc: "all", DateFormat: "ymd"}, true},
		{"delete-date-tag", sharedmodels.FilenameOps{OpType: "delete-date-tag"}, false},
		{"delete-date-tag ignores a stored location and format", sharedmodels.FilenameOps{OpType: "delete-date-tag", OpLoc: "prefix", DateFormat: "ymd"}, false},
		{"regex-replace ok", sharedmodels.FilenameOps{OpType: "regex-replace", OpFindString: "^a", OpValue: "b"}, false},
		{"regex-replace bad pattern", sharedmodels.FilenameOps{OpType: "regex-replace", OpFindString: "[a"}, true},
		{"replace ok", sharedmodels.FilenameOps{OpType: "replace", OpFindString: "a", OpValue: "b"}, false},
		{"replace empty find", sharedmodels.FilenameOps{OpType: "replace", OpValue: "b"}, true},
		{"replace-prefix empty find", sharedmodels.FilenameOps{OpType: "replace-prefix", OpValue: "b"}, true},
		{"replace-suffix empty find", sharedmodels.FilenameOps{OpType: "replace-suffix", OpValue: "b"}, true},
		{"keep-before ok", sharedmodels.FilenameOps{OpType: "keep-before", OpValue: " |"}, false},
		{"keep-after empty", sharedmodels.FilenameOps{OpType: "keep-after"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilenameOps([]sharedmodels.FilenameOps{tt.op})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilenameOps(%+v) err = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateFilenameOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilterOps(t *testing.T) {
	tests := []struct {
		name    string
		filter  sharedmodels.Filters
		wantErr bool
	}{
		{"contains", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}, false},
		{"omits any", sharedmodels.Filters{Field: "title", FilterType: "omits", Value: "dog", MustAny: "any"}, false},
		{"equals", sharedmodels.Filters{Field: "title", FilterType: "equals", Value: "cat", MustAny: "must"}, false},
		{"more-than numeric", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "3600", MustAny: "must"}, false},
		{"presence contains must", sharedmodels.Filters{Field: "date", FilterType: "contains", MustAny: "must"}, false},
		{"presence contains any", sharedmodels.Filters{Field: "date", FilterType: "contains", MustAny: "any"}, false},
		{"presence omits any", sharedmodels.Filters{Field: "date", FilterType: "omits", MustAny: "any"}, false},
		{"equals empty any", sharedmodels.Filters{Field: "title", FilterType: "equals", MustAny: "any"}, false},
		{"regex-contains", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: `(?i)^ep\.? ?\d+`, MustAny: "must"}, false},

		{"unknown type", sharedmodels.Filters{Field: "title", FilterType: "explodes", Value: "x", MustAny: "must"}, true},
		{"more-than non-numeric", sharedmodels.Filters{Field: "duration", FilterType: "more-than", Value: "soon", MustAny: "must"}, true},
		{"bad must/any", sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "maybe"}, true},
		{"empty field", sharedmodels.Filters{FilterType: "contains", Value: "cat", MustAny: "must"}, true},
		{"regex-contains empty pattern", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", MustAny: "must"}, true},
		{"regex-contains invalid pattern", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "(", MustAny: "must"}, true},
		{"regex-contains pattern with ]]:", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "[[:digit:]]:", MustAny: "must"}, true},
		{"regex-contains pattern with ||", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "a||b", MustAny: "must"}, true},
		{"regex-contains pattern with &&", sharedmodels.Filters{Field: "title", FilterType: "regex-contains", Value: "a&&b", MustAny: "must"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilterOps([]sharedmodels.Filters{tt.filter})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilterOps(%+v) err = %v, wantErr %v", tt.filter, err, tt.wantErr)
			}
		})
	}

	if err := sharedvalidation.ValidateFilterOps(nil); err != nil {
		t.Errorf("empty input should be valid, got: %v", err)
	}
}

func TestValidateFilteredMetaOps(t *testing.T) {
	goodFilter := sharedmodels.Filters{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}
	goodOp := sharedmodels.MetaOps{Field: "title", OpType: "prefix", OpValue: "[CAT] "}

	tests := []struct {
		name    string
		entry   sharedmodels.FilteredMetaOps
		wantErr bool
	}{
		{"valid", sharedmodels.FilteredMetaOps{Filters: []sharedmodels.Filters{goodFilter}, MetaOps: []sharedmodels.MetaOps{goodOp}}, false},
		{"no filters", sharedmodels.FilteredMetaOps{MetaOps: []sharedmodels.MetaOps{goodOp}}, true},
		{"no operations", sharedmodels.FilteredMetaOps{Filters: []sharedmodels.Filters{goodFilter}}, true},
		{
			name: "bad filter type",
			entry: sharedmodels.FilteredMetaOps{
				Filters: []sharedmodels.Filters{{Field: "title", FilterType: "explodes", Value: "x", MustAny: "must"}},
				MetaOps: []sharedmodels.MetaOps{goodOp},
			},
			wantErr: true,
		},
		{
			name: "bad operation type",
			entry: sharedmodels.FilteredMetaOps{
				Filters: []sharedmodels.Filters{goodFilter},
				MetaOps: []sharedmodels.MetaOps{{Field: "title", OpType: "explodes"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedvalidation.ValidateFilteredMetaOps([]sharedmodels.FilteredMetaOps{tt.entry})
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilteredFilenameOps(t *testing.T) {
	entry := sharedmodels.FilteredFilenameOps{
		Filters:     []sharedmodels.Filters{{Field: "title", FilterType: "contains", Value: "cat", MustAny: "must"}},
		FilenameOps: []sharedmodels.FilenameOps{{OpType: "prefix", OpValue: "[CAT] "}},
	}
	if err := sharedvalidation.ValidateFilteredFilenameOps([]sharedmodels.FilteredFilenameOps{entry}); err != nil {
		t.Errorf("valid entry rejected: %v", err)
	}

	entry.FilenameOps = nil
	if err := sharedvalidation.ValidateFilteredFilenameOps([]sharedmodels.FilteredFilenameOps{entry}); err == nil {
		t.Error("an entry with no filename operations should be rejected")
	}
}
