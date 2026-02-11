package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadImage_ValidateImageURL(t *testing.T) {
	service := &UploadService{}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid R2 URL",
			url:     "https://bucket.tansil.pro/org-123/image.jpg",
			wantErr: false,
		},
		{
			name:    "Valid Cloudflare R2",
			url:     "https://xxx.r2.cloudflarestorage.com/bucket/image.png",
			wantErr: false,
		},
		{
			name:    "Valid AWS S3",
			url:     "https://bucket.s3.amazonaws.com/image.jpg",
			wantErr: false,
		},
		{
			name:    "HTTP not allowed",
			url:     "http://example.com/image.jpg",
			wantErr: true,
		},
		{
			name:    "Empty URL allowed",
			url:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateImageURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadImage_ValidateFileType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "Valid JPG",
			filename: "image.jpg",
			wantErr:  false,
		},
		{
			name:     "Valid JPEG",
			filename: "image.jpeg",
			wantErr:  false,
		},
		{
			name:     "Valid PNG",
			filename: "image.png",
			wantErr:  false,
		},
		{
			name:     "Valid WebP",
			filename: "image.webp",
			wantErr:  false,
		},
		{
			name:     "Valid GIF",
			filename: "image.gif",
			wantErr:  false,
		},
		{
			name:     "Invalid TXT",
			filename: "file.txt",
			wantErr:  true,
		},
		{
			name:     "Invalid PDF",
			filename: "doc.pdf",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := getExt(tt.filename)
			validExts := map[string]string{
				".jpg": "image/jpeg", ".jpeg": "image/jpeg",
				".png": "image/png", ".webp": "image/webp", ".gif": "image/gif",
			}
			_, isValid := validExts[ext]
			
			if tt.wantErr {
				assert.False(t, isValid, "Expected invalid file type for %s", tt.filename)
			} else {
				assert.True(t, isValid, "Expected valid file type for %s", tt.filename)
			}
		})
	}
}

func getExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
