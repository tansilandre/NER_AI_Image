package external

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client wraps Cloudflare R2 operations
type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// R2Config holds R2 configuration
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

// NewR2Client creates a new R2 client
func NewR2Client(cfg R2Config) (*R2Client, error) {
	// R2 uses S3-compatible API
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Client{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}, nil
}

// Upload uploads data to R2 and returns the public URL
func (r *R2Client) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return public URL
	if r.publicURL != "" {
		return fmt.Sprintf("%s/%s", r.publicURL, key), nil
	}
	return key, nil
}

// UploadBytes uploads byte data to R2
func (r *R2Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return r.Upload(ctx, key, io.NopCloser(io.Reader(nil)), contentType)
}

// Download downloads data from R2
func (r *R2Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from R2: %w", err)
	}
	return result.Body, nil
}

// Delete deletes an object from R2
func (r *R2Client) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	return nil
}

// GenerateKey generates a unique key for storage
func GenerateKey(orgID string, folder string, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%s/%d_%s", orgID, folder, timestamp, filename)
}
