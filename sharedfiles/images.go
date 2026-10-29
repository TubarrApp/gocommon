package sharedfiles

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// SanitizeUploadedImage decodes, validates, and re-encodes an image as GIF, JPEG or PNG.
func SanitizeUploadedImage(data []byte, formatExt string) ([]byte, error) {
	const (
		maxImageBytes  = 20 << 20 // 20 MiB.
		maxImagePixels = 40_000_000
	)

	// Check the output format.
	if !sharedconsts.ImageUploadTypes[formatExt] {
		return nil, fmt.Errorf("unsupported image format %q", formatExt)
	}

	// Check the image size.
	if len(data) > maxImageBytes {
		return nil, fmt.Errorf("image exceeds the maximum size of %d bytes", maxImageBytes)
	}

	// Check image validity and dimensions before decoding.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 ||
		int64(cfg.Width) > maxImagePixels/int64(cfg.Height) {
		return nil, fmt.Errorf("image has invalid or excessive dimensions")
	}

	// Decode the complete image.
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Check format match.
	if format == sharedconsts.IUploadJPEG && formatExt == sharedconsts.IUploadJPG {
		formatExt = sharedconsts.IUploadJPEG
	}
	if format != formatExt {
		return nil, fmt.Errorf("image content is %q, but filename indicates %q", format, formatExt)
	}

	// Re-encode the image into the selected format.
	var buf bytes.Buffer
	switch format {
	case sharedconsts.IUploadGIF:
		err = gif.Encode(&buf, img, nil)
	case sharedconsts.IUploadJPG, sharedconsts.IUploadJPEG:
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	case sharedconsts.IUploadPNG:
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported output image format %q", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to re-encode image: %w", err)
	}

	return buf.Bytes(), nil
}
