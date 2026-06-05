package services

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/satria/obrolan-api/internal/config"
)

type UploadService struct {
	cloudinary *cloudinary.Cloudinary
}

func NewUploadService(cfg *config.Config) (*UploadService, error) {
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		return nil, err
	}

	return &UploadService{cloudinary: cld}, nil
}

var allowedMIMETypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
}

const maxFileSize = 5 * 1024 * 1024 // 5MB

func (s *UploadService) UploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file size
	if header.Size > maxFileSize {
		return "", errors.New("file too large: max 5MB")
	}

	// Validate MIME type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return "", errors.New("failed to read file header")
	}

	// Reset file pointer
	file.Seek(0, 0)

	mimeType := detectContentType(buff)
	if _, ok := allowedMIMETypes[mimeType]; !ok {
		return "", errors.New("invalid file type: allowed jpeg, png, gif, webp")
	}

	// Validate extension
	ext := strings.ToLower(getExtension(header.Filename))
	allowed := false
	for _, allowedExt := range allowedMIMETypes {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", errors.New("invalid file extension")
	}

	// Upload to Cloudinary
	ctx := context.Background()
	result, err := s.cloudinary.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:   "social-forum",
		ResourceType: "image",
	})
	if err != nil {
		return "", errors.New("failed to upload image: " + err.Error())
	}

	return result.SecureURL, nil
}

func (s *UploadService) DeleteImage(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	// Parse public ID from URL: https://res.cloudinary.com/.../social-forum/filename.jpg
	// public_id = "social-forum/filename"
	parts := strings.Split(imageURL, "/")
	if len(parts) < 2 {
		return nil
	}
	filename := parts[len(parts)-1]
	ext := filepath.Ext(filename)
	publicID := "social-forum/" + strings.TrimSuffix(filename, ext)

	ctx := context.Background()
	_, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		log.Printf("Warning: failed to delete image from Cloudinary: %v", err)
	}
	return nil
}

// ImageService interface — for thread delete (Cloudinary cleanup)
type ImageService interface {
	DeleteImage(imageURL string) error
}

// Uploader interface — for upload handler (supports real Cloudinary + no-op)
type Uploader interface {
	UploadImage(file multipart.File, header *multipart.FileHeader) (string, error)
	ImageService
}

// compile-time check
var _ Uploader = (*UploadService)(nil)
var _ Uploader = (*NoOpUploadService)(nil)

// NoOpUploadService is a no-op implementation for when Cloudinary is not configured
type NoOpUploadService struct{}

func NewNoOpUploadService() *NoOpUploadService {
	return &NoOpUploadService{}
}

var ErrUploadNotConfigured = errors.New("upload not configured: Cloudinary credentials missing")

func (s *NoOpUploadService) UploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	return "", ErrUploadNotConfigured
}

func (s *NoOpUploadService) DeleteImage(imageURL string) error {
	return nil
}

func detectContentType(buff []byte) string {
	// Simple MIME detection based on magic bytes
	if len(buff) < 4 {
		return "application/octet-stream"
	}

	// JPEG
	if buff[0] == 0xFF && buff[1] == 0xD8 && buff[2] == 0xFF {
		return "image/jpeg"
	}
	// PNG
	if buff[0] == 0x89 && buff[1] == 0x50 && buff[2] == 0x4E && buff[3] == 0x47 {
		return "image/png"
	}
	// GIF
	if buff[0] == 0x47 && buff[1] == 0x49 && buff[2] == 0x46 {
		return "image/gif"
	}
	// WebP
	if len(buff) > 11 && buff[8] == 0x57 && buff[9] == 0x45 && buff[10] == 0x42 && buff[11] == 0x50 {
		return "image/webp"
	}

	return "application/octet-stream"
}

func getExtension(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx == -1 {
		return ""
	}
	return filename[idx:]
}
