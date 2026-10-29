package sharedfiles

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TubarrApp/gocommon/sharedconsts"
)

// SanitizeUploadedImage decodes, validates, and re-encodes an image as GIF, JPEG or PNG.
func SanitizeUploadedImage(path string) (retErr error) {
	const (
		maxImageBytes  = 20 << 20 // 20 MiB.
		maxImagePixels = 40_000_000
	)

	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

	// Check the output format.
	if !sharedconsts.ImageUploadTypes[format] {
		return fmt.Errorf("unsupported output image format %q", format)
	}

	// Open the source image.
	f, err := OpenRegularInRoot(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the image, limiting its size.
	data, err := io.ReadAll(io.LimitReader(f, maxImageBytes+1))
	if err != nil {
		return fmt.Errorf("failed to read image %q: %w", path, err)
	}
	if len(data) > maxImageBytes {
		return fmt.Errorf("image %q exceeds the maximum size of %d bytes", path, maxImageBytes)
	}

	// Check image validity and dimensions before decoding.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("invalid image %q: %w", path, err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 ||
		int64(cfg.Width) > maxImagePixels/int64(cfg.Height) {
		return fmt.Errorf("image %q has invalid or excessive dimensions", path)
	}

	// Decode the complete image.
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode image %q: %w", path, err)
	}

	// Create a temporary file in the destination directory.
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary image for %q: %w", path, err)
	}

	tmpName := tmp.Name()
	defer func() {
		if tmp != nil {
			if err := tmp.Close(); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed to close temporary image: %w", err))
			}
		}
		if tmpName != "" {
			if err := os.Remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
				retErr = errors.Join(retErr, fmt.Errorf("failed to remove temporary image: %w", err))
			}
		}
	}()

	// Re-encode the image into the selected format.
	switch format {
	case sharedconsts.IUploadGIF:
		err = gif.Encode(tmp, img, nil)
	case sharedconsts.IUploadJPG, sharedconsts.IUploadJPEG:
		err = jpeg.Encode(tmp, img, &jpeg.Options{Quality: 90})
	case sharedconsts.IUploadPNG:
		err = png.Encode(tmp, img)
	}
	if err != nil {
		return fmt.Errorf("failed to re-encode image %q: %w", path, err)
	}

	// Flush encoded data to disk.
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary image: %w", err)
	}

	// Close the temporary file before replacing the destination.
	err = tmp.Close()
	tmp = nil
	if err != nil {
		return fmt.Errorf("failed to close temporary image: %w", err)
	}

	// Replace the destination with the validated image.
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to replace image at path %q: %w", path, err)
	}
	tmpName = "" // Defer func will clean if left non-zero.

	return nil
}
