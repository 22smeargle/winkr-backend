package photo

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// UploadPhotoRequest represents the request to upload a photo
type UploadPhotoRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	File         io.Reader `json:"-" validate:"required"`
	FileName     string    `json:"file_name" validate:"required"`
	ContentType  string    `json:"content_type" validate:"required"`
	FileSize     int64     `json:"file_size" validate:"required"`
	IsPrimary    bool      `json:"is_primary"`
}

// UploadPhotoResponse represents the response after uploading a photo
type UploadPhotoResponse struct {
	PhotoID       uuid.UUID `json:"photo_id"`
	FileURL        string    `json:"file_url"`
	ThumbnailURL   string    `json:"thumbnail_url,omitempty"`
	IsPrimary      bool      `json:"is_primary"`
	VerificationStatus string    `json:"verification_status"`
	ProcessingTime  int64     `json:"processing_time_ms"`
}

// UploadPhotoUseCase handles photo upload logic
type UploadPhotoUseCase struct {
	photoRepo         repositories.PhotoRepository
	storageService    storage.StorageService
	imageProcessor    services.ImageProcessingService
	maxPhotosPerUser int
}

// NewUploadPhotoUseCase creates a new upload photo use case
func NewUploadPhotoUseCase(
	photoRepo repositories.PhotoRepository,
	storageService storage.StorageService,
	imageProcessor services.ImageProcessingService,
) *UploadPhotoUseCase {
	return &UploadPhotoUseCase{
		photoRepo:         photoRepo,
		storageService:    storageService,
		imageProcessor:    imageProcessor,
		maxPhotosPerUser: 6, // Maximum 6 photos per user
	}
}

// Execute executes the upload photo use case
func (uc *UploadPhotoUseCase) Execute(ctx context.Context, req *UploadPhotoRequest) (*UploadPhotoResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check user photo limit
	currentPhotoCount, err := uc.photoRepo.GetUserPhotoCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user photo count: %w", err)
	}

	if currentPhotoCount >= uc.maxPhotosPerUser {
		return nil, fmt.Errorf("maximum photo limit reached (%d photos)", uc.maxPhotosPerUser)
	}

	// Validate image
	validationResult, err := uc.imageProcessor.ValidateImage(ctx, req.File)
	if err != nil {
		return nil, fmt.Errorf("image validation failed: %w", err)
	}

	if !validationResult.IsValid {
		return nil, fmt.Errorf("image validation failed: %v", validationResult.Errors)
	}

	// Process image
	processOptions := &services.ProcessOptions{
		ResizeWidth:   1200,
		ResizeHeight:  1200,
		Quality:       85,
		GenerateThumb:  true,
		ThumbWidth:     300,
		ThumbHeight:    300,
		StripEXIF:      true,
		Optimize:       true,
	}

	processResult, err := uc.imageProcessor.ProcessImage(ctx, req.File, processOptions)
	if err != nil {
		return nil, fmt.Errorf("image processing failed: %w", err)
	}

	// Upload original image to storage
	originalKey := fmt.Sprintf("photos/%s/original/%s", req.UserID.String(), processResult.OriginalKey)
	originalURL, err := uc.storageService.UploadFile(ctx, req.File, originalKey, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload original image: %w", err)
	}

	// Upload processed image to storage
	processedKey := fmt.Sprintf("photos/%s/processed/%s", req.UserID.String(), processResult.ProcessedKey)
	processedURL, err := uc.storageService.UploadFile(
		ctx,
		bytes.NewReader(processResult.ProcessedSize), // This would need the actual processed data
		processedKey,
		req.ContentType,
	)
	if err != nil {
		// Clean up original image if processed upload fails
		_ = uc.storageService.DeleteFile(ctx, originalKey)
		return nil, fmt.Errorf("failed to upload processed image: %w", err)
	}

	// Upload thumbnail if generated
	var thumbnailURL string
	if processResult.ThumbnailKey != "" {
		thumbnailKey := fmt.Sprintf("photos/%s/thumbnails/%s", req.UserID.String(), processResult.ThumbnailKey)
		thumbnailURL, err = uc.storageService.UploadFile(
			ctx,
			bytes.NewReader(processResult.ThumbnailSize), // This would need the actual thumbnail data
			thumbnailKey,
			req.ContentType,
		)
		if err != nil {
			// Clean up other images if thumbnail upload fails
			_ = uc.storageService.DeleteFile(ctx, originalKey)
			_ = uc.storageService.DeleteFile(ctx, processedKey)
			return nil, fmt.Errorf("failed to upload thumbnail: %w", err)
		}
	}

	// Create photo entity
	photo := &entities.Photo{
		ID:                uuid.New(),
		UserID:            req.UserID,
		FileURL:           processedURL,
		FileKey:           processedKey,
		IsPrimary:         req.IsPrimary,
		VerificationStatus: "pending",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// If this is set as primary, unset other primary photos
	if req.IsPrimary {
		if err := uc.photoRepo.UnsetPrimaryPhoto(ctx, req.UserID); err != nil {
			// Clean up uploaded files if database operation fails
			_ = uc.storageService.DeleteFile(ctx, originalKey)
			_ = uc.storageService.DeleteFile(ctx, processedKey)
			if thumbnailURL != "" {
				_ = uc.storageService.DeleteFile(ctx, thumbnailKey)
			}
			return nil, fmt.Errorf("failed to unset primary photo: %w", err)
		}
	}

	// Save photo to database
	if err := uc.photoRepo.Create(ctx, photo); err != nil {
		// Clean up uploaded files if database operation fails
		_ = uc.storageService.DeleteFile(ctx, originalKey)
		_ = uc.storageService.DeleteFile(ctx, processedKey)
		if thumbnailURL != "" {
			_ = uc.storageService.DeleteFile(ctx, thumbnailKey)
		}
		return nil, fmt.Errorf("failed to save photo: %w", err)
	}

	processingTime := time.Since(startTime).Milliseconds()

	logger.Info("Photo uploaded successfully", map[string]interface{}{
		"photo_id":        photo.ID,
		"user_id":         req.UserID,
		"file_size":       req.FileSize,
		"processing_time":  processingTime,
		"is_primary":      req.IsPrimary,
	})

	return &UploadPhotoResponse{
		PhotoID:        photo.ID,
		FileURL:         processedURL,
		ThumbnailURL:    thumbnailURL,
		IsPrimary:       req.IsPrimary,
		VerificationStatus: photo.VerificationStatus,
		ProcessingTime:   processingTime,
	}, nil
}

// validateRequest validates the upload request
func (uc *UploadPhotoUseCase) validateRequest(req *UploadPhotoRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if req.File == nil {
		return fmt.Errorf("file is required")
	}

	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}

	if req.ContentType == "" {
		return fmt.Errorf("content type is required")
	}

	if req.FileSize <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	return nil
}