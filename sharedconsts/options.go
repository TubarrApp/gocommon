package sharedconsts

// BlockedYtdlpOptions contains YT-DLP options that users shouldn't use.
var BlockedYtdlpOptions = map[string]struct{}{
	"--alias":            {},
	"--config-locations": {},
	"--exec":             {},
	"--netrc-cmd":        {},
	"--plugin-dirs":      {},
}

// BlockedFFmpegOptions contains FFmpeg options that users shouldn't use.
var BlockedFFmpegOptions = map[string]struct{}{
	"-attach":             {},
	"-dump_attachment":    {},
	"-passlogfile":        {},
	"-print_graphs_file":  {},
	"-protocol-blacklist": {},
	"-protocol-whitelist": {},
	"-report":             {},
	"-sdp_file":           {},
	"-vstats_file":        {},
}
