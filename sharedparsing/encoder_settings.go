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

	// Check each codec's lines, in a set order so the same settings always give the same error.
	for _, codec := range slices.Sorted(maps.Keys(args)) {
		for _, resolution := range slices.Sorted(maps.Keys(args[codec])) {
			if err := sharedvalidation.ValidateEncoderArgsUsed(codec, args[codec][resolution]); err != nil {
				return err
			}
			if s.Software {
				if err := sharedvalidation.ValidateSoftwareEncoderArgs(codec, args[codec][resolution]); err != nil {
					return err
				}
			}
		}
		if err := sharedvalidation.ValidateEncoderArgsCanRun(codec, plainVideo, videoPairs, s.OutputExt); err != nil {
			return err
		}
	}

	// Check each audio codec's line likewise.
	audioArgs, err := ParseAudioEncoderArgs(s.AudioArgs)
	if err != nil {
		return err
	}
	plainAudio, audioPairs, err := ParseAudioCodecMappings(s.AudioCodecs)
	if err != nil {
		return err
	}
	for _, codec := range slices.Sorted(maps.Keys(audioArgs)) {
		if err := sharedvalidation.ValidateEncoderArgsCanRun(codec, plainAudio, audioPairs, s.OutputExt); err != nil {
			return err
		}
	}

	// Check extra FFmpeg args hold no options for encoding audio.
	extra, err := ShellSplit(s.ExtraArgs)
	if err != nil {
		return err
	}
	return sharedvalidation.ValidateExtraArgsAudioOptions(extra)
}
