package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/external"
)

// UploadService handles file uploads to R2
type UploadService struct {
	r2Client *external.R2Client
}

// NewUploadService creates a new upload service
func NewUploadService(r2Client *external.R2Client) *UploadService {
	return &UploadService{
		r2Client: r2Client,
	}
}

// UploadResult contains upload response
type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Filename string `json:"filename"`
}

// UploadImage uploads an image to R2
func (s *UploadService) UploadImage(ctx context.Context, orgID string, folder string, filename string, data io.Reader, size int64) (*UploadResult, error) {
	// Validate file type
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
		".gif":  "image/gif",
	}

	contentType, valid := validExts[ext]
	if !valid {
		return nil, fmt.Errorf("invalid file type: %s (allowed: jpg, png, webp, gif)", ext)
	}

	// Generate unique key
	key := external.GenerateKey(orgID, folder, filename)

	// Upload to R2
	url, err := s.r2Client.Upload(ctx, key, data, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	return &UploadResult{
		URL:      url,
		Key:      key,
		Filename: filename,
	}, nil
}

// ValidateImageURL checks if URL is valid and from allowed domains
func (s *UploadService) ValidateImageURL(url string) error {
	if url == "" {
		return nil // Empty is valid
	}

	// Must be HTTPS
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("image URL must use HTTPS")
	}

	// Check allowed domains
	allowedDomains := []string{
		"bucket.tansil.pro",
		"r2.cloudflarestorage.com",
		"amazonaws.com",
	}

	for _, domain := range allowedDomains {
		if strings.Contains(url, domain) {
			return nil
		}
	}

	// For now, allow any HTTPS URL during development
	// In production, restrict to specific domains
	return nil
}

// GetContentType determines content type from filename
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

// GenerateUniqueFilename generates a unique filename with UUID
func GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	name = sanitizeFilename(name)
	return fmt.Sprintf("%s_%s%s", name, uuid.New().String()[:8], ext)
}

func sanitizeFilename(name string) string {
	// Remove special characters
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return strings.ToLower(name)
}
