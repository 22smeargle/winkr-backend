package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EphemeralPhotoStorageService defines the interface for ephemeral photo storage operations
type EphemeralPhotoStorageService interface {
	// Upload operations
	UploadEphemeralPhoto(ctx context.Context, file io.Reader, filename string, userID uuid.UUID) (*EphemeralPhotoUploadResult, error)
	UploadEphemeralPhotoWithProcessing(ctx context.Context, file io.Reader, filename string, userID uuid.UUID, options *EphemeralPhotoProcessingOptions) (*EphemeralPhotoUploadResult, error)
	
	// Access operations
	GetEphemeralPhotoURL(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error)
	GetEphemeralPhotoThumbnailURL(ctx context.Context, thumbnailKey string, expiresIn time.Duration) (string, error)
	GetSignedURL(ctx context.Context, fileKey string, expiresIn time.Duration, download bool) (string, error)
	
	// Security operations
	GenerateWatermarkedURL(ctx context.Context, fileKey string, watermarkText string) (string, error)
	GenerateAccessKey(ctx context.Context) (string, error)
	ValidateAccessKey(ctx context.Context, accessKey string) (bool, error)
	
	// Cleanup operations
	DeleteEphemeralPhoto(ctx context.Context, fileKey string, thumbnailKey string) error
	BatchDeleteEphemeralPhotos(ctx context.Context, fileKeys []string, thumbnailKeys []string) error
	CleanupExpiredFiles(ctx context.Context, olderThan time.Duration) (int, error)
	
	// Analytics operations
	GetStorageUsage(ctx context.Context, userID uuid.UUID) (*EphemeralPhotoStorageUsage, error)
	GetFileMetadata(ctx context.Context, fileKey string) (*EphemeralPhotoMetadata, error)
}

// EphemeralPhotoUploadResult represents the result of uploading an ephemeral photo
type EphemeralPhotoUploadResult struct {
	FileKey       string    `json:"file_key"`
	FileURL       string    `json:"file_url"`
	FileSize      int64     `json:"file_size"`
	ThumbnailKey  string    `json:"thumbnail_key"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	ThumbnailSize int64     `json:"thumbnail_size"`
	ContentType   string    `json:"content_type"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	ProcessingTime int64     `json:"processing_time_ms"`
	AccessKey     string    `json:"access_key"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// EphemeralPhotoProcessingOptions represents processing options for ephemeral photos
type EphemeralPhotoProcessingOptions struct {
	ResizeWidth     int    `json:"resize_width"`
	ResizeHeight    int    `json:"resize_height"`
	Quality         int    `json:"quality"`
	GenerateThumb   bool   `json:"generate_thumb"`
	ThumbWidth      int    `json:"thumb_width"`
	ThumbHeight     int    `json:"thumb_height"`
	AddWatermark    bool   `json:"add_watermark"`
	WatermarkText   string `json:"watermark_text"`
	Optimize        bool   `json:"optimize"`
	StripEXIF       bool   `json:"strip_exif"`
}

// EphemeralPhotoStorageUsage represents storage usage statistics
type EphemeralPhotoStorageUsage struct {
	TotalFiles      int64 `json:"total_files"`
	TotalSize       int64 `json:"total_size"`
	ActiveFiles     int64 `json:"active_files"`
	ExpiredFiles    int64 `json:"expired_files"`
	ViewedFiles     int64 `json:"viewed_files"`
	StorageToday    int64 `json:"storage_today"`
	StorageThisWeek int64 `json:"storage_this_week"`
	StorageThisMonth int64 `json:"storage_this_month"`
}

// EphemeralPhotoMetadata represents metadata for an ephemeral photo file
type EphemeralPhotoMetadata struct {
	FileKey     string    `json:"file_key"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsExpired   bool      `json:"is_expired"`
	ViewCount   int       `json:"view_count"`
	LastAccess  *time.Time `json:"last_access,omitempty"`
}

// EphemeralPhotoStorageServiceImpl implements EphemeralPhotoStorageService
type EphemeralPhotoStorageServiceImpl struct {
	storageService StorageService
	imageService    ImageProcessingService
	config         *config.StorageConfig
}

// NewEphemeralPhotoStorageService creates a new ephemeral photo storage service
func NewEphemeralPhotoStorageService(
	storageService StorageService,
	imageService ImageProcessingService,
	config *config.StorageConfig,
) EphemeralPhotoStorageService {
	return &EphemeralPhotoStorageServiceImpl{
		storageService: storageService,
		imageService:    imageService,
		config:         config,
	}
}

// UploadEphemeralPhoto uploads an ephemeral photo
func (s *EphemeralPhotoStorageServiceImpl) UploadEphemeralPhoto(ctx context.Context, file io.Reader, filename string, userID uuid.UUID) (*EphemeralPhotoUploadResult, error) {
	startTime := time.Now()
	
	// Generate unique file keys
	fileKey := s.generateFileKey(userID, "original")
	thumbnailKey := s.generateFileKey(userID, "thumbnail")
	
	// Validate file
	validationResult, err := s.imageService.ValidateImage(ctx, file)
	if err != nil {
		logger.Error("Failed to validate ephemeral photo", err)
		return nil, fmt.Errorf("failed to validate ephemeral photo: %w", err)
	}
	
	if !validationResult.IsValid {
		return nil, fmt.Errorf("invalid image: %v", validationResult.Errors)
	}
	
	// Process image with default options
	options := &EphemeralPhotoProcessingOptions{
		ResizeWidth:   800,  // Max width for ephemeral photos
		ResizeHeight:  800,  // Max height for ephemeral photos
		Quality:       85,   // Good quality for web
		GenerateThumb:  true,
		ThumbWidth:    200,
		ThumbHeight:   200,
		AddWatermark:  true,  // Always watermark ephemeral photos
		WatermarkText: "EPHEMERAL",
		Optimize:      true,
		StripEXIF:     true,
	}
	
	return s.UploadEphemeralPhotoWithProcessing(ctx, file, filename, userID, options)
}

// UploadEphemeralPhotoWithProcessing uploads an ephemeral photo with custom processing
func (s *EphemeralPhotoStorageServiceImpl) UploadEphemeralPhotoWithProcessing(ctx context.Context, file io.Reader, filename string, userID uuid.UUID, options *EphemeralPhotoProcessingOptions) (*EphemeralPhotoUploadResult, error) {
	startTime := time.Now()
	
	// Generate unique file keys
	fileKey := s.generateFileKey(userID, "original")
	thumbnailKey := s.generateFileKey(userID, "thumbnail")
	
	// Validate file
	validationResult, err := s.imageService.ValidateImage(ctx, file)
	if err != nil {
		logger.Error("Failed to validate ephemeral photo", err)
		return nil, fmt.Errorf("failed to validate ephemeral photo: %w", err)
	}
	
	if !validationResult.IsValid {
		return nil, fmt.Errorf("invalid image: %v", validationResult.Errors)
	}
	
	// Process image
	processOptions := &ProcessOptions{
		ResizeWidth:   options.ResizeWidth,
		ResizeHeight:  options.ResizeHeight,
		Quality:       options.Quality,
		GenerateThumb:  options.GenerateThumb,
		ThumbWidth:     options.ThumbWidth,
		ThumbHeight:    options.ThumbHeight,
		AddWatermark:   options.AddWatermark,
		WatermarkText:  options.WatermarkText,
		Optimize:       options.Optimize,
		StripEXIF:      options.StripEXIF,
	}
	
	processResult, err := s.imageService.ProcessImage(ctx, file, processOptions)
	if err != nil {
		logger.Error("Failed to process ephemeral photo", err)
		return nil, fmt.Errorf("failed to process ephemeral photo: %w", err)
	}
	
	// Upload processed image
	fileURL, err := s.storageService.UploadFile(ctx, strings.NewReader(processResult.ProcessedKey), fileKey, "image/jpeg")
	if err != nil {
		logger.Error("Failed to upload processed ephemeral photo", err)
		return nil, fmt.Errorf("failed to upload processed ephemeral photo: %w", err)
	}
	
	// Upload thumbnail
	var thumbnailURL string
	var thumbnailSize int64
	if processResult.ThumbnailKey != "" {
		thumbnailURL, err = s.storageService.UploadFile(ctx, strings.NewReader(processResult.ThumbnailKey), thumbnailKey, "image/jpeg")
		if err != nil {
			logger.Error("Failed to upload thumbnail", err)
			return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
		}
		thumbnailSize = processResult.ThumbnailSize
	}
	
	// Generate access key
	accessKey, err := s.GenerateAccessKey(ctx)
	if err != nil {
		logger.Error("Failed to generate access key", err)
		return nil, fmt.Errorf("failed to generate access key: %w", err)
	}
	
	processingTime := time.Since(startTime).Milliseconds()
	
	result := &EphemeralPhotoUploadResult{
		FileKey:       fileKey,
		FileURL:       fileURL,
		FileSize:      processResult.ProcessedSize,
		ThumbnailKey:  thumbnailKey,
		ThumbnailURL:  thumbnailURL,
		ThumbnailSize: thumbnailSize,
		ContentType:   validationResult.Format,
		Width:         processResult.Width,
		Height:        processResult.Height,
		ProcessingTime: processingTime,
		AccessKey:     accessKey,
		ExpiresAt:     time.Now().Add(30 * time.Second), // Default 30 seconds
	}
	
	logger.Info("Ephemeral photo uploaded successfully", map[string]interface{}{
		"user_id":         userID,
		"file_key":        fileKey,
		"thumbnail_key":    thumbnailKey,
		"access_key":      accessKey,
		"processing_time":  processingTime,
	})
	
	return result, nil
}

// GetEphemeralPhotoURL gets a signed URL for an ephemeral photo
func (s *EphemeralPhotoStorageServiceImpl) GetEphemeralPhotoURL(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error) {
	// Generate short-lived signed URL for security
	url, err := s.GetSignedURL(ctx, fileKey, expiresIn, false)
	if err != nil {
		logger.Error("Failed to generate ephemeral photo URL", err)
		return "", fmt.Errorf("failed to generate ephemeral photo URL: %w", err)
	}
	
	return url, nil
}

// GetEphemeralPhotoThumbnailURL gets a signed URL for an ephemeral photo thumbnail
func (s *EphemeralPhotoStorageServiceImpl) GetEphemeralPhotoThumbnailURL(ctx context.Context, thumbnailKey string, expiresIn time.Duration) (string, error) {
	// Generate short-lived signed URL for thumbnail
	url, err := s.GetSignedURL(ctx, thumbnailKey, expiresIn, false)
	if err != nil {
		logger.Error("Failed to generate ephemeral photo thumbnail URL", err)
		return "", fmt.Errorf("failed to generate ephemeral photo thumbnail URL: %w", err)
	}
	
	return url, nil
}

// GetSignedURL generates a signed URL with expiration
func (s *EphemeralPhotoStorageServiceImpl) GetSignedURL(ctx context.Context, fileKey string, expiresIn time.Duration, download bool) (string, error) {
	// This would integrate with the actual storage service (S3, MinIO, etc.)
	// For now, return a mock URL
	baseURL := "https://cdn.example.com/ephemeral"
	
	if download {
		return fmt.Sprintf("%s/%s/download?expires=%d", baseURL, fileKey, time.Now().Add(expiresIn).Unix()), nil
	}
	
	return fmt.Sprintf("%s/%s?expires=%d", baseURL, fileKey, time.Now().Add(expiresIn).Unix()), nil
}

// GenerateWatermarkedURL generates a URL for watermarked version
func (s *EphemeralPhotoStorageServiceImpl) GenerateWatermarkedURL(ctx context.Context, fileKey string, watermarkText string) (string, error) {
	// Generate URL for watermarked version
	baseURL := "https://cdn.example.com/ephemeral"
	return fmt.Sprintf("%s/%s?watermark=true&text=%s", baseURL, fileKey, watermarkText), nil
}

// GenerateAccessKey generates a unique access key
func (s *EphemeralPhotoStorageServiceImpl) GenerateAccessKey(ctx context.Context) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("eph_%x", bytes), nil
}

// ValidateAccessKey validates an access key format
func (s *EphemeralPhotoStorageServiceImpl) ValidateAccessKey(ctx context.Context, accessKey string) (bool, error) {
	// Check if access key has correct format
	if len(accessKey) != 19 { // "eph_" + 16 hex chars
		return false, fmt.Errorf("invalid access key format")
	}
	
	if !strings.HasPrefix(accessKey, "eph_") {
		return false, fmt.Errorf("invalid access key prefix")
	}
	
	// Validate hex part
	hexPart := accessKey[3:]
	if len(hexPart) != 16 {
		return false, fmt.Errorf("invalid access key length")
	}
	
	// Additional validation can be added here
	return true, nil
}

// DeleteEphemeralPhoto deletes an ephemeral photo and its thumbnail
func (s *EphemeralPhotoStorageServiceImpl) DeleteEphemeralPhoto(ctx context.Context, fileKey string, thumbnailKey string) error {
	// Delete main file
	if err := s.storageService.DeleteFile(ctx, fileKey); err != nil {
		logger.Error("Failed to delete ephemeral photo file", err)
		return fmt.Errorf("failed to delete ephemeral photo file: %w", err)
	}
	
	// Delete thumbnail
	if thumbnailKey != "" {
		if err := s.storageService.DeleteFile(ctx, thumbnailKey); err != nil {
			logger.Error("Failed to delete ephemeral photo thumbnail", err)
			return fmt.Errorf("failed to delete ephemeral photo thumbnail: %w", err)
		}
	}
	
	logger.Info("Ephemeral photo deleted successfully", map[string]interface{}{
		"file_key":     fileKey,
		"thumbnail_key": thumbnailKey,
	})
	
	return nil
}

// BatchDeleteEphemeralPhotos deletes multiple ephemeral photos
func (s *EphemeralPhotoStorageServiceImpl) BatchDeleteEphemeralPhotos(ctx context.Context, fileKeys []string, thumbnailKeys []string) error {
	// Delete main files
	for _, fileKey := range fileKeys {
		if err := s.storageService.DeleteFile(ctx, fileKey); err != nil {
			logger.Error("Failed to delete ephemeral photo file in batch", err)
			return fmt.Errorf("failed to delete ephemeral photo file in batch: %w", err)
		}
	}
	
	// Delete thumbnails
	for _, thumbnailKey := range thumbnailKeys {
		if thumbnailKey != "" {
			if err := s.storageService.DeleteFile(ctx, thumbnailKey); err != nil {
				logger.Error("Failed to delete ephemeral photo thumbnail in batch", err)
				return fmt.Errorf("failed to delete ephemeral photo thumbnail in batch: %w", err)
			}
		}
	}
	
	logger.Info("Ephemeral photos deleted successfully in batch", map[string]interface{}{
		"file_count":     len(fileKeys),
		"thumbnail_count": len(thumbnailKeys),
	})
	
	return nil
}

// CleanupExpiredFiles cleans up expired files
func (s *EphemeralPhotoStorageServiceImpl) CleanupExpiredFiles(ctx context.Context, olderThan time.Duration) (int, error) {
	// This would typically query for expired files and delete them
	// For now, return mock count
	deletedCount := 0
	
	logger.Info("Expired files cleanup completed", map[string]interface{}{
		"older_than":    olderThan,
		"deleted_count": deletedCount,
	})
	
	return deletedCount, nil
}

// GetStorageUsage gets storage usage statistics
func (s *EphemeralPhotoStorageServiceImpl) GetStorageUsage(ctx context.Context, userID uuid.UUID) (*EphemeralPhotoStorageUsage, error) {
	// This would typically query storage usage from database
	// For now, return mock data
	usage := &EphemeralPhotoStorageUsage{
		TotalFiles:      0,
		TotalSize:       0,
		ActiveFiles:     0,
		ExpiredFiles:    0,
		ViewedFiles:     0,
		StorageToday:    0,
		StorageThisWeek: 0,
		StorageThisMonth: 0,
	}
	
	return usage, nil
}

// GetFileMetadata gets metadata for a file
func (s *EphemeralPhotoStorageServiceImpl) GetFileMetadata(ctx context.Context, fileKey string) (*EphemeralPhotoMetadata, error) {
	// This would typically get metadata from storage service
	// For now, return mock data
	metadata := &EphemeralPhotoMetadata{
		FileKey:     fileKey,
		FileSize:    0,
		ContentType: "image/jpeg",
		Width:       800,
		Height:      600,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(30 * time.Second),
		IsExpired:   false,
		ViewCount:   0,
	}
	
	return metadata, nil
}

// Helper methods

// generateFileKey generates a unique file key
func (s *EphemeralPhotoStorageServiceImpl) generateFileKey(userID uuid.UUID, fileType string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ephemeral/%s/%s/%d", userID.String(), fileType, timestamp)
}