package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessImage_JPEG(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "test_image.jpg"))
	require.NoError(t, err)

	metadata, err := ProcessImage(data, "image/jpeg")
	require.NoError(t, err)

	assert.Equal(t, 100, metadata.Width)
	assert.Equal(t, 80, metadata.Height)
	assert.Equal(t, "UpQKhxo1|To1sUfQfQfQ]1fQA{fQsUfQfQfQ", metadata.Blurhash)
}

func TestProcessImage_PNG(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "test_image.png"))
	require.NoError(t, err)

	metadata, err := ProcessImage(data, "image/png")
	require.NoError(t, err)

	assert.Equal(t, 200, metadata.Width)
	assert.Equal(t, 150, metadata.Height)
	assert.Equal(t, "UsEqV5fj0SfjfjfQfQfQ4]fQ?BfQfjfQfQfQ", metadata.Blurhash)
}

func TestProcessImage_GIF(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "test_image.gif"))
	require.NoError(t, err)

	metadata, err := ProcessImage(data, "image/gif")
	require.NoError(t, err)

	assert.Equal(t, 50, metadata.Width)
	assert.Equal(t, 50, metadata.Height)
	assert.Equal(t, "UyJH^#fT09fTfTfQfQfQ09fQ-*fQfTfQfQfQ", metadata.Blurhash)
}

func TestProcessImage_WebP(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "test_image.webp"))
	require.NoError(t, err)

	metadata, err := ProcessImage(data, "image/webp")
	require.NoError(t, err)

	assert.Equal(t, 120, metadata.Width)
	assert.Equal(t, 90, metadata.Height)
	assert.Equal(t, "UbJ}=7ju0qju9%fQs+fQESfQ-OfQ$_fQR-fQ", metadata.Blurhash)
}

func TestProcessImage_InvalidData(t *testing.T) {
	invalidData := []byte("not an image")

	_, err := ProcessImage(invalidData, "image/jpeg")
	assert.Error(t, err)
}

func TestProcessImage_EmptyData(t *testing.T) {
	_, err := ProcessImage([]byte{}, "image/png")
	assert.Error(t, err)
}

func TestIsImageMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"video/mp4", false},
		{"application/pdf", false},
		{"text/plain", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.mimeType, func(t *testing.T) {
			result := IsImageMimeType(tc.mimeType)
			assert.Equal(t, tc.expected, result)
		})
	}
}
