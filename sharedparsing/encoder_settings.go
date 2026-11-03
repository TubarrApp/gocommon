package sharedparsing

import (
	"maps"
	"slices"

	"github.com/TubarrApp/gocommon/sharedmodels"
	"github.com/TubarrApp/gocommon/sharedvalidation"
)

// CheckEncoderSettings returns an error if encoder settings, as a video will get them, hold something that can't work on any FFmpeg version.
func CheckEncoderSettings(s sharedmodels.EncoderSettings) error {
	args, err := ParseVideoEncoderArgs(s.VideoArgs)
	if err != nil {
		return err
	}
	plainVideo, videoPairs, err := ParseCodecMappings(s.VideoCodecs, sharedvalidation.ValidateVideoCodec)
	if err != nil {
		return err
	}

	// Check each codec's line, in a set order so the same settings always give the same error.
	for _, codec := range slices.Sorted(maps.Keys(args)) {
		if err := sharedvalidation.ValidateEncoderArgsUsed(codec, args[codec]); err != nil {
			return err
		}
		if s.Software {
			if err := sharedvalidation.ValidateSoftwareEncoderArgs(codec, args[codec]); err != nil {
				return err
			}
		}
		if err := sharedvalidation.ValidateEncoderArgsCanRun(codec, plainVideo, videoPairs, s.OutputExt); err != nil {
			return err
		}
	}

	// Check extra FFmpeg args on audio that's only ever copied.
	plainAudio, audioPairs, err := ParseCodecMappings(s.AudioCodecs, sharedvalidation.ValidateAudioCodec)
	if err != nil {
		return err
	}
	if !sharedvalidation.AudioAlwaysCopied(plainAudio, audioPairs, s.OutputExt) {
		return nil
	}
	extra, err := ShellSplit(s.ExtraArgs)
	if err != nil {
		return err
	}
	if err := sharedvalidation.ValidateCopiedAudioFilter(extra); err != nil {
		return err
	}
	return sharedvalidation.ValidateCopiedAudioOptions(extra)
}
