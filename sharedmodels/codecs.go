package sharedmodels

import "github.com/TubarrApp/gocommon/sharedconsts"

// FromCodec holds information about the 'from' codec.
// e.g. lc_h264 (h264 with low compression).
type FromCodec struct {
	Codec       string
	Compression sharedconsts.CompressionLevel
}

// EncoderArgs holds a target codec's encoder args, with the ones Metarr converts for GPU encoders taken out.
type EncoderArgs struct {
	Quality     string        // From -crf.
	PixFmt      string        // From -pix_fmt.
	VideoFilter string        // From -vf or -filter:v.
	Passes      int           // From -pass, 0 for one pass.
	Bitrates    map[int]int64 // From -b:v, in bits per second. Keyed by the smallest resolution each is for, 0 for every size.
	Args        []string      // The rest, as typed.
}

// AudioEncoderArgs holds a target audio codec's encoder args, with -af, -ac and -ar taken out as typed for Metarr to add itself.
type AudioEncoderArgs struct {
	Filter     string   // From -af or -filter:a.
	Channels   string   // From -ac.
	SampleRate string   // From -ar.
	Args       []string // The rest, as typed.
}

// EncoderSettings holds the settings encoder args are judged with, as a video will get them.
type EncoderSettings struct {
	VideoArgs   []string // "codec:args" lines.
	AudioArgs   []string // "codec:args" lines for audio.
	VideoCodecs []string // Video codec entries.
	AudioCodecs []string // Audio codec entries.
	ExtraArgs   string   // Extra FFmpeg args.
	OutputExt   string   // Blank or "same" keeps the video's container.
	Software    bool     // Whether software encoders run, as GPU encoders take other values.
}

// NewFromCodec returns a 'FromCodec' model.
func NewFromCodec(codec string, compression sharedconsts.CompressionLevel) FromCodec {
	return FromCodec{
		Codec:       codec,
		Compression: compression,
	}
}
