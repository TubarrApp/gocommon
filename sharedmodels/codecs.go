package sharedmodels

import "github.com/TubarrApp/gocommon/sharedconsts"

// FromCodec holds information about the 'from' codec.
// e.g. lc_h264 (h264 with low compression).
type FromCodec struct {
	Codec       string
	Compression sharedconsts.CompressionLevel
}

// NewFromCodec returns a 'FromCodec' model.
func NewFromCodec(codec string, compression sharedconsts.CompressionLevel) FromCodec {
	return FromCodec{
		Codec:       codec,
		Compression: compression,
	}
}
