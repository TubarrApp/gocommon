package sharedconsts

import "testing"

// TestHasBlockedOptions tests blocked options are caught with a value or stream specifier, and normal args pass.
func TestHasBlockedOptions(t *testing.T) {
	tests := []struct {
		args      []string
		blockList map[string]struct{}
		want      string
	}{
		{[]string{"-passlogfile:v:0", "/tmp/x"}, BlockedFFmpegOptions, "-passlogfile"},
		{[]string{"-dump_attachment:t:0", "out.jpg"}, BlockedFFmpegOptions, "-dump_attachment"},
		{[]string{"-protocol_whitelist", "file,http"}, BlockedFFmpegOptions, "-protocol_whitelist"},
		{[]string{"--exec=echo:hi"}, BlockedYtdlpOptions, "--exec"},
		{[]string{"-progress", "/tmp/x"}, BlockedFFmpegOptions, "-progress"},
		{[]string{"-vn"}, BlockedFFmpegOptions, "-vn"},
		{[]string{"-f", "matroska"}, BlockedFFmpegOptions, "-f"},
		{[]string{"-filter_complex", "anullsrc"}, BlockedFFmpegOptions, "-filter_complex"},
		{[]string{"-lavfi", "anullsrc"}, BlockedFFmpegOptions, "-lavfi"},
		{[]string{"-fflags", "+bitexact", "-filter:a", "volume=2"}, BlockedFFmpegOptions, ""},
		{[]string{"--exec-before-download", "x"}, BlockedYtdlpOptions, "--exec-before-download"},
		{[]string{"--exec-b=x"}, BlockedYtdlpOptions, "--exec-b"},
		{[]string{"--config-location", "x"}, BlockedYtdlpOptions, "--config-location"},
		{[]string{"--netrc-c", "x"}, BlockedYtdlpOptions, "--netrc-c"},
		{[]string{"--plugin-d", "x"}, BlockedYtdlpOptions, "--plugin-d"},
		{[]string{"--alia", "a", "b"}, BlockedYtdlpOptions, "--alia"},
		{[]string{"--netrc", "--netrc-location", "x", "--", "https://example.com"}, BlockedYtdlpOptions, ""},
		{[]string{"-c:v", "libx264"}, BlockedFFmpegOptions, "-c"},
		{[]string{"-c:a", "aac"}, BlockedFFmpegOptions, "-c"},
		{[]string{"-c", "copy"}, BlockedFFmpegOptions, "-c"},
		{[]string{"-codec:v", "libx264"}, BlockedFFmpegOptions, "-codec"},
		{[]string{"-vcodec", "libx264"}, BlockedFFmpegOptions, "-vcodec"},
		{[]string{"-acodec", "aac"}, BlockedFFmpegOptions, "-acodec"},
		{[]string{"-b:v", "3M", "-map", "0:v", "-metadata:s:a:0", "language=eng", "-vf", "scale=1280:-2"}, BlockedFFmpegOptions, ""},
		{[]string{"-cpu-used", "8", "-coder", "0", "-color_primaries", "bt709"}, BlockedFFmpegOptions, ""},
		{[]string{"-f", "bestvideo[height<=720]+bestaudio", "-o", "%(title)s:%(id)s.%(ext)s", "https://example.com/watch?v=1"}, BlockedYtdlpOptions, ""},
	}
	for _, tt := range tests {
		blocked, got := HasBlockedOptions(tt.args, tt.blockList)
		if got != tt.want || blocked != (tt.want != "") {
			t.Errorf("HasBlockedOptions(%q) = %t, %q; want %q", tt.args, blocked, got, tt.want)
		}
	}
}
