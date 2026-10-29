package sharedfiles

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSanitizePNGFile performs a live test on a local test file.
func TestSanitizePNGFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(file), "testimages", "2.png")

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	got, err := SanitizeUploadedImage(data, "png")
	if err != nil {
		t.Fatalf("SanitizeUploadedImage() error = %v", err)
	}

	_, format, err := image.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
	if format != "png" {
		t.Fatalf("output format = %q, want %q", format, "png")
	}

	if err := os.WriteFile(src, got, 0600); err != nil {
		t.Fatalf("Failed to write sanitized image: %v", err)
	}

	t.Logf("Successfully sanitized PNG: %s", src)
}

// TestSanitizeJPGFile performs a live test on a local test file.
func TestSanitizeJPGFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(file), "testimages", "1.jpg")

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	got, err := SanitizeUploadedImage(data, "jpg")
	if err != nil {
		t.Fatalf("SanitizeUploadedImage() error = %v", err)
	}

	_, format, err := image.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
	if format != "jpeg" {
		t.Fatalf("output format = %q, want %q", format, "jpeg")
	}

	if err := os.WriteFile(src, got, 0600); err != nil {
		t.Fatalf("Failed to write sanitized image: %v", err)
	}

	t.Logf("Successfully sanitized JPG: %s", src)
}

// TestSanitizeGIFFile performs a live test on a local test file.
func TestSanitizeGIFFile(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(file), "testimages", "3.gif")

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	got, err := SanitizeUploadedImage(data, "gif")
	if err != nil {
		t.Fatalf("SanitizeUploadedImage() error = %v", err)
	}

	_, format, err := image.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
	if format != "gif" {
		t.Fatalf("output format = %q, want %q", format, "gif")
	}

	if err := os.WriteFile(src, got, 0600); err != nil {
		t.Fatalf("Failed to write sanitized image: %v", err)
	}

	t.Logf("Successfully sanitized GIF: %s", src)
}

// TestSanitizeImage checks that supported input formats can be sanitized while preserving their format.
func TestSanitizeImage(t *testing.T) {
	formats := []string{"jpeg", "png", "gif"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			data := encodeTestImage(t, format)

			got, err := SanitizeUploadedImage(data, format)
			if err != nil {
				t.Fatalf("SanitizeUploadedImage() error = %v", err)
			}

			_, gotFormat, err := image.Decode(bytes.NewReader(got))
			if err != nil {
				t.Fatalf("sanitized output is not a decodable image: %v", err)
			}
			if gotFormat != format {
				t.Errorf("output format = %q, want %q", gotFormat, format)
			}
		})
	}
}

// encodeTestImage encodes an image for test purposes.
func encodeTestImage(t *testing.T, format string) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{A: 255})

	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	default:
		t.Fatalf("unsupported test input format %q", format)
	}
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// mustDecodeBase64 returns the bytes represented by the base64 string sent in.
func mustDecodeBase64(t *testing.T, value string) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
