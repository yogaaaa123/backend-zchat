package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectContentType_JPEG(t *testing.T) {
	buff := []byte{0xFF, 0xD8, 0xFF, 0x00, 0x00, 0x00}
	result := detectContentType(buff)
	assert.Equal(t, "image/jpeg", result)
}

func TestDetectContentType_PNG(t *testing.T) {
	buff := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00}
	result := detectContentType(buff)
	assert.Equal(t, "image/png", result)
}

func TestDetectContentType_GIF(t *testing.T) {
	buff := []byte{0x47, 0x49, 0x46, 0x00, 0x00, 0x00}
	result := detectContentType(buff)
	assert.Equal(t, "image/gif", result)
}

func TestDetectContentType_WebP(t *testing.T) {
	buff := make([]byte, 12)
	buff[8] = 0x57 // W
	buff[9] = 0x45 // E
	buff[10] = 0x42 // B
	buff[11] = 0x50 // P
	result := detectContentType(buff)
	assert.Equal(t, "image/webp", result)
}

func TestDetectContentType_Unknown(t *testing.T) {
	buff := []byte{0x00, 0x00, 0x00, 0x00}
	result := detectContentType(buff)
	assert.Equal(t, "application/octet-stream", result)
}

func TestGetExtension_WithExt(t *testing.T) {
	result := getExtension("photo.jpg")
	assert.Equal(t, ".jpg", result)
}

func TestGetExtension_NoExt(t *testing.T) {
	result := getExtension("photo")
	assert.Equal(t, "", result)
}

func TestGetExtension_MultipleDots(t *testing.T) {
	result := getExtension("photo.min.jpg")
	assert.Equal(t, ".jpg", result)
}
