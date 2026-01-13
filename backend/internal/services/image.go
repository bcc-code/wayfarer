package services

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/buckket/go-blurhash"
	_ "golang.org/x/image/webp"
)

// ImageMetadata contains extracted metadata from an image
type ImageMetadata struct {
	Width    int
	Height   int
	Blurhash string
}

// blurhash component counts (4x4 as per user preference)
const (
	blurhashXComponents = 4
	blurhashYComponents = 4
)

// ProcessImage extracts dimensions and generates blurhash from image data.
// Returns nil if the data is not a valid image.
func ProcessImage(data []byte, mimeType string) (*ImageMetadata, error) {
	// Decode image to get dimensions
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Generate blurhash
	hash, err := blurhash.Encode(blurhashXComponents, blurhashYComponents, img)
	if err != nil {
		return nil, fmt.Errorf("failed to generate blurhash: %w", err)
	}

	return &ImageMetadata{
		Width:    width,
		Height:   height,
		Blurhash: hash,
	}, nil
}

// IsImageMimeType returns true if the MIME type is an image type
func IsImageMimeType(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}
