package sharedvalidation

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TubarrApp/gocommon/sharedconsts"
	"github.com/TubarrApp/gocommon/sharedtemplates"
)

// Codecs ------------------------------------------------------------------------------------

func TestValidateVideoCodec(t *testing.T) {
	tests := []struct {
		in     string
		expect string
		ok     bool
	}{
		{"x264", sharedconsts.VCodecH264, true},
		{"libx265", sharedconsts.VCodecHEVC, true},
		{"av01", sharedconsts.VCodecAV1, true},
		{"vp09", sharedconsts.VCodecVP9, true},
		{sharedconsts.VCodecH264, sharedconsts.VCodecH264, true},
		{"", "", true},
		{"invalid", "", false},
	}

	for _, tt := range tests {
		out, err := ValidateVideoCodec(tt.in)
		if tt.ok && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %q", tt.in)
		}
		if out != tt.expect {
			t.Errorf("expected %q got %q", tt.expect, out)
		}
	}
}

func FuzzValidateVideoCodec(f *testing.F) {
	f.Add("")
	for k := range sharedconsts.ValidVideoCodecs {
		f.Add(k)
	}
	for k := range sharedconsts.VideoCodecAlias {
		f.Add(k)
	}

	f.Fuzz(func(t *testing.T, a string) {
		fmt.Printf("Testing input %q", a)
		out, err := ValidateVideoCodec(a)
		if err != nil {
			if out != "" {
				t.Fatalf("Got non-empty output %q despite error: %v", out, err)
			}
		}

		if out == "" {
			return
		}

		if _, ok := sharedconsts.ValidVideoCodecs[out]; !ok {
			t.Fatalf("Input %q gave invalid output string %q", a, out)
		}
	})
}

func TestValidateOutputExt(t *testing.T) {
	tests := []struct {
		in     string
		expect string
		ok     bool
	}{
		{sharedconsts.ExtMP4, sharedconsts.ExtMP4, true},
		{"mp4", sharedconsts.ExtMP4, true},
		{sharedconsts.ExtOGV, sharedconsts.ExtOGV, true},
		{"rm", sharedconsts.ExtRM, true},
		{"    MP4   ", sharedconsts.ExtMP4, true},
		{"    JUNK   ", "", false},
	}

	for _, tt := range tests {
		out, err := ValidateFFmpegOutputExt(tt.in)
		if tt.ok && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %q", tt.in)
		}
		if out != tt.expect {
			t.Errorf("expected %q got %q", tt.expect, out)
		}
	}
}

func TestValidateVideoCodecWithAccel(t *testing.T) {
	tests := []struct {
		codec     string
		accelType string
		expect    string
		ok        bool
	}{
		// Pass states.
		{"x264", sharedconsts.AccelTypeAMF, sharedconsts.VCodecH264, true},
		{"libx265", sharedconsts.AccelTypeAuto, sharedconsts.VCodecHEVC, true},
		{sharedconsts.VCodecAV1, "", sharedconsts.VCodecAV1, true},
		{"vp09", "", sharedconsts.VCodecVP9, true},
		{"", "", "", true},

		// Fail states.
		{"invalid", "", "", false},
		{"", sharedconsts.AccelTypeCuda, "", false},
	}

	for _, tt := range tests {
		out, err := ValidateVideoCodecWithAccel(tt.codec, tt.accelType)

		if tt.ok && err != nil {
			t.Errorf("unexpected error for %#v: %v", tt, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %#v", tt)
		}
		if out != tt.expect {
			t.Errorf("expected %q got %q", tt.expect, out)
		}
	}
}

func TestValidateAudioCodec(t *testing.T) {
	tests := []struct {
		in     string
		expect string
		ok     bool
	}{
		// Pass states.
		{sharedconsts.ACodecMP3, sharedconsts.ACodecMP3, true},
		{"libmp3lame", sharedconsts.ACodecMP3, true},
		{"wave", sharedconsts.ACodecWAV, true},
		{"", "", true},

		// Fail states.
		{"badcodec", "", false},
	}

	for _, tt := range tests {
		out, err := ValidateAudioCodec(tt.in)
		if tt.ok && err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %q", tt.in)
		}
		if out != tt.expect {
			t.Errorf("expected %q got %q", tt.expect, out)
		}
	}
}

func FuzzValidateAudioCodec(f *testing.F) {
	f.Add("")
	for k := range sharedconsts.ValidAudioCodecs {
		f.Add(k)
	}
	for k := range sharedconsts.AudioCodecAlias {
		f.Add(k)
	}

	f.Fuzz(func(t *testing.T, a string) {
		fmt.Printf("Testing input %q", a)
		out, err := ValidateAudioCodec(a)
		if err != nil {
			if out != "" {
				t.Fatalf("Got non-empty output %q despite error: %v", out, err)
			}
		}

		if out == "" {
			return
		}

		if _, ok := sharedconsts.ValidAudioCodecs[out]; !ok {
			t.Fatalf("Input %q gave invalid output string %q", a, out)
		}
	})
}

// TestParseCodecMappings tests splitting codec entries into a plain codec and per-source pairs, in any order.
func TestParseCodecMappings(t *testing.T) {
	tests := []struct {
		in        []string
		wantAll   string
		wantPairs map[string]string
		ok        bool
	}{
		// Pass states.
		{[]string{"hevc", "vp9:h264"}, "hevc", map[string]string{"vp9": "h264"}, true},
		{[]string{"vp9:h264", "hevc"}, "hevc", map[string]string{"vp9": "h264"}, true},
		{[]string{"x264"}, "h264", map[string]string{}, true},
		{[]string{"hevc", "libx265"}, "hevc", map[string]string{}, true},
		{[]string{"vp9:h264", "vp09:x264"}, "", map[string]string{"vp9": "h264"}, true},
		{[]string{"", "av1:hevc"}, "", map[string]string{"av1": "hevc"}, true},
		{nil, "", map[string]string{}, true},

		// Fail states.
		{[]string{"hevc", "h264"}, "", nil, false},
		{[]string{"vp9:h264", "libvpxvp9:hevc"}, "", nil, false},
		{[]string{":h264"}, "", nil, false},
		{[]string{"vp9:bad"}, "", nil, false},
		{[]string{"vp9:h264:av1"}, "", nil, false},
		{[]string{"bad"}, "", nil, false},
	}

	for _, tt := range tests {
		all, pairs, err := ParseCodecMappings(tt.in, ValidateVideoCodec)
		if tt.ok != (err == nil) {
			t.Errorf("ParseCodecMappings(%q) error = %v, want ok %v", tt.in, err, tt.ok)
			continue
		}
		if all != tt.wantAll || !maps.Equal(pairs, tt.wantPairs) {
			t.Errorf("ParseCodecMappings(%q) = (%q, %v), want (%q, %v)", tt.in, all, pairs, tt.wantAll, tt.wantPairs)
		}
	}
}

func TestValidateGPUAccelType(t *testing.T) {
	out, err := ValidateGPUAccelType(sharedconsts.AccelTypeCuda)
	if err != nil || out != sharedconsts.AccelTypeCuda {
		t.Errorf("expected nvidia (got error? %v)", err)
	}

	out, err = ValidateGPUAccelType("nvenc")
	if err != nil || out != sharedconsts.AccelTypeCuda {
		t.Errorf("expected nvidia (got error? %v)", err)
	}

	out, err = ValidateGPUAccelType("vAaPi")
	if err != nil || out != sharedconsts.AccelTypeVAAPI {
		t.Errorf("expected vaapi (got error? %v)", err)
	}

	_, err = ValidateGPUAccelType("badtype")
	if err == nil {
		t.Errorf("expected failure for invalid accel")
	}

	_, err = ValidateGPUAccelType("")
	if err == nil {
		t.Errorf("expected failure for blank accel")
	}
}

func FuzzValidateGPUAccelType(f *testing.F) {
	f.Add("")
	for k := range sharedconsts.ValidGPUAccelTypes {
		f.Add(k)
	}
	for k := range sharedconsts.AccelAlias {
		f.Add(k)
	}

	f.Fuzz(func(t *testing.T, a string) {
		fmt.Printf("Testing input %q", a)
		out, err := ValidateGPUAccelType(a)
		if err != nil {
			if out != "" {
				t.Fatalf("Got non-empty output %q despite error: %v", out, err)
			}
		}

		if out == "" {
			return
		}

		if _, ok := sharedconsts.ValidGPUAccelTypes[out]; !ok {
			t.Fatalf("Input %q gave invalid output string %q", a, out)
		}
	})
}

// Filesystem --------------------------------------------------------------------------------

func TestValidateDirectory(t *testing.T) {
	now := time.Now()
	tmp := filepath.Join(os.TempDir(), "sv_test_dir"+now.String())
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	// Valid template directory.
	validTemplateTmp := filepath.Join(os.TempDir(), "{{channel_name}}", "{{year}}")
	if hasTemplating, _, err := ValidateDirectory(validTemplateTmp, false, sharedtemplates.AllTemplatesMap); !hasTemplating || err != nil {
		t.Errorf("expected template detection and pass")
	}

	// Tags with spaces and capitals, as the template parsers accept them.
	spacedTemplateTmp := filepath.Join(os.TempDir(), "{{ Channel_Name }}")
	if hasTemplating, _, err := ValidateDirectory(spacedTemplateTmp, false, sharedtemplates.AllTemplatesMap); !hasTemplating || err != nil {
		t.Errorf("expected spaced, capitalised tag to pass, got %v", err)
	}

	// Invalid template directory.
	invalidTemplateTmp := filepath.Join(os.TempDir(), "{{channel_name}}", "{{BOGUS}}")
	if hasTemplating, _, err := ValidateDirectory(invalidTemplateTmp, false, sharedtemplates.AllTemplatesMap); !hasTemplating || err == nil {
		t.Errorf("expected template detection and fail")
	}

	// Empty directory.
	if _, _, err := ValidateDirectory("", false, sharedtemplates.AllTemplatesMap); err == nil {
		t.Errorf("expected error for empty directory")
	}

	// Invalid directory.
	if _, _, err := ValidateDirectory(tmp, false, sharedtemplates.AllTemplatesMap); err == nil {
		t.Errorf("expected error")
	}

	// Creates directory.
	if _, _, err := ValidateDirectory(tmp, true, sharedtemplates.AllTemplatesMap); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Directory should now exist.
	if _, _, err := ValidateDirectory(tmp, false, sharedtemplates.AllTemplatesMap); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFile(t *testing.T) {
	now := time.Now()
	tmp := filepath.Join(os.TempDir(), "sv_test_file.txt"+now.String())
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	// Valid template directory.
	validTemplateTmp := filepath.Join(os.TempDir(), "{{channel_name}}", "{{year}}", "sv_test_file.txt"+now.String())
	if hasTemplating, _, err := ValidateDirectory(validTemplateTmp, false, sharedtemplates.AllTemplatesMap); !hasTemplating || err != nil {
		t.Errorf("expected pass")
	}

	// Invalid template directory.
	invalidTemplateTmp := filepath.Join(os.TempDir(), "{{channel_name}}", "{{BOGUS}}", "sv_test_file.txt"+now.String())
	if hasTemplating, _, err := ValidateDirectory(invalidTemplateTmp, false, sharedtemplates.AllTemplatesMap); !hasTemplating || err == nil {
		t.Errorf("expected fail")
	}

	// Non-existent path.
	if _, _, err := ValidateFile(tmp, false, sharedtemplates.AllTemplatesMap); err == nil {
		t.Errorf("expected error")
	}

	// Creates non-existent file.
	if _, _, err := ValidateFile(tmp, true, sharedtemplates.AllTemplatesMap); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// File should now exist.
	if _, _, err := ValidateFile(tmp, false, sharedtemplates.AllTemplatesMap); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetRenameFlag(t *testing.T) {
	tests := []struct {
		in     string
		expect string
	}{
		// Passthrough.
		{sharedconsts.RenameFixesOnly, sharedconsts.RenameFixesOnly},
		{sharedconsts.RenameSkip, sharedconsts.RenameSkip},

		// Aliases.
		{"fixed", sharedconsts.RenameFixesOnly},
		{"none", sharedconsts.RenameSkip},
		{"", ""},
		{"INVALID", ""},
	}

	for _, tt := range tests {
		flag := GetRenameFlag(tt.in)
		if flag != tt.expect {
			t.Fatalf("unexpected flag: sent %q, want %q, got: %q", tt.in, tt.expect, flag)
		}
	}
}

// Media -------------------------------------------------------------------------------------

func TestValidateTranscodeQuality(t *testing.T) {
	out, err := ValidateTranscodeQuality("25")
	if err != nil || out != "25" {
		t.Errorf("unexpected result")
	}

	out, err = ValidateTranscodeQuality("100")
	if err != nil || out != "51" {
		t.Errorf("clamp failed")
	}
}

func TestValidateConcurrencyLimit(t *testing.T) {
	if ValidateConcurrencyLimit(0) != 1 {
		t.Errorf("expected 1")
	}
	if ValidateConcurrencyLimit(5) != 5 {
		t.Errorf("expected 5")
	}
}

func TestValidateMinFreeMem(t *testing.T) {
	tests := []struct {
		in     string
		expect string
		ok     bool
	}{
		{"2G", "2G", true},
		{"500M", "500M", true},
		{"200K", "200K", true},
		{"2000", "2000", true},
		{"x1", "", false},
	}

	for _, tt := range tests {
		out, err := ValidateMinFreeMem(tt.in)
		if tt.ok && err != nil {
			t.Errorf("unexpected error for %q", tt.in)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected error for %q", tt.in)
		}
		if out != tt.expect {
			t.Errorf("expected %q got %q", tt.expect, out)
		}
	}
}

func TestValidateMaxCPU(t *testing.T) {
	// Zero values.
	if ValidateMaxCPU(0.0, false) != 101.0 {
		t.Errorf("expected 101")
	}
	if ValidateMaxCPU(0, true) != 0.0 {
		t.Errorf("expected passthrough")
	}

	// Non-zero values.
	if ValidateMaxCPU(100.0, false) != 101.0 {
		t.Errorf("expected 101")
	}
	if ValidateMaxCPU(10.0, false) != 10.0 {
		t.Errorf("expected passthrough")
	}
	if ValidateMaxCPU(2.0, false) != 5.0 {
		t.Errorf("expected clamp to 5")
	}
}
