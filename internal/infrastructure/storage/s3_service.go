package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	// UploadFile uploads a file and returns the file key and URL
	UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error)
	
	// GetUploadURL generates a presigned URL for file upload
	GetUploadURL(ctx context.Context, key string, contentType string) (string, error)
	
	// GetDownloadURL generates a presigned URL for file download
	GetDownloadURL(ctx context.Context, key string) (string, error)
	
	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, key string) error
	
	// FileExists checks if a file exists
	FileExists(ctx context.Context, key string) (bool, error)
	
	// CopyFile copies a file to a new location
	CopyFile(ctx context.Context, sourceKey, destKey string) error
	
	// GetFileInfo gets file information
	GetFileInfo(ctx context.Context, key string) (*FileInfo, error)
}

// FileInfo represents file information
type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
}

// S3Storage implements StorageService using AWS S3 or MinIO
type S3Storage struct {
	client     *s3.Client
	config     *config.StorageConfig
	bucket     string
	isMinIO    bool
}

// NewS3Storage creates a new S3/MinIO storage service
func NewS3Storage(cfg *config.StorageConfig) (*S3Storage, error) {
	var awsConfig aws.Config
	var err error

	// Configure based on provider
	if strings.ToLower(cfg.Provider) == "minio" {
		// MinIO configuration
		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load MinIO config: %w", err)
		}

		// Custom endpoint for MinIO
		awsConfig.BaseEndpoint = aws.String(getEndpointURL(cfg.Endpoint, cfg.UseSSL))
	} else {
		// AWS S3 configuration
		awsConfig, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if strings.ToLower(cfg.Provider) == "minio" {
			// Disable path style addressing for MinIO
			o.UsePathStyle = true
		}
	})

	storage := &S3Storage{
		client:  client,
		config:  cfg,
		bucket:  cfg.Bucket,
		isMinIO: strings.ToLower(cfg.Provider) == "minio",
	}

	// Test connection by checking if bucket exists
	if err := storage.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to storage: %w", err)
	}

	logger.Info("Storage service initialized", map[string]interface{}{
		"provider": cfg.Provider,
		"bucket":   cfg.Bucket,
		"endpoint": cfg.Endpoint,
	})

	return storage, nil
}

// testConnection tests the connection to the storage service
func (s *S3Storage) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})

	if err != nil {
		// Try to create bucket if it doesn't exist
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return s.createBucket()
		}
		return fmt.Errorf("bucket access failed: %w", err)
	}

	return nil
}

// createBucket creates the bucket if it doesn't exist
func (s *S3Storage) createBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	input := &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	}

	// Set location constraint for AWS S3 (not needed for MinIO)
	if !s.isMinIO && s.config.Region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.config.Region),
		}
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	logger.Info("Bucket created successfully", map[string]interface{}{
		"bucket": s.bucket,
	})

	return nil
}

// UploadFile uploads a file and returns the file key and URL
func (s *S3Storage) UploadFile(ctx context.Context, file io.Reader, key string, contentType string) (string, error) {
	if key == "" {
		key = s.generateFileKey()
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})

	if err != nil {
		logger.Error("Failed to upload file", err)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	logger.Info("File uploaded successfully", map[string]interface{}{
		"key":  key,
		"bucket": s.bucket,
	})

	return key, nil
}

// GetUploadURL generates a presigned URL for file upload
func (s *S3Storage) GetUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	if key == "" {
		key = s.generateFileKey()
	}

	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	}, s.config.UploadExpiry)

	if err != nil {
		logger.Error("Failed to generate upload URL", err)
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	logger.Info("Upload URL generated", map[string]interface{}{
		"key": key,
	})

	return req.URL, nil
}

// GetDownloadURL generates a presigned URL for file download
func (s *S3Storage) GetDownloadURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s.config.DownloadExpiry)

	if err != nil {
		logger.Error("Failed to generate download URL", err)
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	logger.Info("Download URL generated", map[string]interface{}{
		"key": key,
	})

	return req.URL, nil
}

// DeleteFile deletes a file from storage
func (s *S3Storage) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("Failed to delete file", err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("File deleted successfully", map[string]interface{}{
		"key": key,
	})

	return nil
}

// FileExists checks if a file exists
func (s *S3Storage) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// CopyFile copies a file to a new location
func (s *S3Storage) CopyFile(ctx context.Context, sourceKey, destKey string) error {
	copySource := fmt.Sprintf("%s/%s", s.bucket, sourceKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
		ACL:        types.ObjectCannedACLPrivate,
	})

	if err != nil {
		logger.Error("Failed to copy file", err)
		return fmt.Errorf("failed to copy file: %w", err)
	}

	logger.Info("File copied successfully", map[string]interface{}{
		"source_key": sourceKey,
		"dest_key":   destKey,
	})

	return nil
}

// GetFileInfo gets file information
func (s *S3Storage) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("Failed to get file info", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	var size int64
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}

	var lastModified time.Time
	if resp.LastModified != nil {
		lastModified = *resp.LastModified
	}

	var contentType string
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	var etag string
	if resp.ETag != nil {
		etag = *resp.ETag
	}

	return &FileInfo{
		Key:          key,
		Size:         size,
		LastModified: lastModified,
		ContentType:  contentType,
		ETag:         etag,
	}, nil
}

// generateFileKey generates a unique file key
func (s *S3Storage) generateFileKey() string {
	id := uuid.New()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("photos/%d/%s", timestamp, id.String())
}

// getEndpointURL constructs the endpoint URL for MinIO
func getEndpointURL(endpoint string, useSSL bool) string {
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		endpoint = scheme + "://" + endpoint
	}

	// Validate URL
	if _, err := url.Parse(endpoint); err != nil {
		logger.Error("Invalid endpoint URL", err)
		return ""
	}

	return endpoint
}

// GetPublicURL returns the public URL for a file (if applicable)
func (s *S3Storage) GetPublicURL(key string) string {
	if s.isMinIO {
		return fmt.Sprintf("%s/%s/%s", getEndpointURL(s.config.Endpoint, s.config.UseSSL), s.bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.config.Region, key)
}

// IsAllowedContentType checks if the content type is allowed
func (s *S3Storage) IsAllowedContentType(contentType string) bool {
	for _, allowedType := range s.config.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

// GetMaxFileSize returns the maximum allowed file size
func (s *S3Storage) GetMaxFileSize() int64 {
	return s.config.MaxFileSize
}