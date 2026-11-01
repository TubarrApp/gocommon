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

// NewFromCodec returns a 'FromCodec' model.
func NewFromCodec(codec string, compression sharedconsts.CompressionLevel) FromCodec {
	return FromCodec{
		Codec:       codec,
		Compression: compression,
	}
}
