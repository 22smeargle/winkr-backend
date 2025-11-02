package photo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetUploadURLRequest represents the request to get an upload URL
type GetUploadURLRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	FileName     string    `json:"file_name" validate:"required"`
	ContentType  string    `json:"content_type" validate:"required"`
	FileSize     int64     `json:"file_size" validate:"required"`
}

// GetUploadURLResponse represents the response with upload URL
type GetUploadURLResponse struct {
	UploadURL     string    `json:"upload_url"`
	FileKey       string    `json:"file_key"`
	ExpiresAt     string    `json:"expires_at"`
	MaxFileSize   int64     `json:"max_file_size"`
	AllowedTypes  []string  `json:"allowed_types"`
}

// GetUploadURLUseCase handles getting upload URL logic
type GetUploadURLUseCase struct {
	photoRepo      repositories.PhotoRepository
	storageService storage.StorageService
	maxFileSize    int64
	allowedTypes   []string
}

// NewGetUploadURLUseCase creates a new get upload URL use case
func NewGetUploadURLUseCase(
	photoRepo repositories.PhotoRepository,
	storageService storage.StorageService,
	maxFileSize int64,
	allowedTypes []string,
) *GetUploadURLUseCase {
	return &GetUploadURLUseCase{
		photoRepo:      photoRepo,
		storageService: storageService,
		maxFileSize:    maxFileSize,
		allowedTypes:   allowedTypes,
	}
}

// Execute executes the get upload URL use case
func (uc *GetUploadURLUseCase) Execute(ctx context.Context, req *GetUploadURLRequest) (*GetUploadURLResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check user photo limit
	currentPhotoCount, err := uc.photoRepo.GetUserPhotoCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user photo count: %w", err)
	}

	maxPhotosPerUser := 6
	if currentPhotoCount >= maxPhotosPerUser {
		return nil, fmt.Errorf("maximum photo limit reached (%d photos)", maxPhotosPerUser)
	}

	// Generate unique file key
	fileKey := uc.generateFileKey(req.UserID, req.FileName)

	// Generate presigned upload URL
	uploadURL, err := uc.storageService.GetUploadURL(ctx, fileKey, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(15 * time.Minute) // 15 minutes from now

	logger.Info("Upload URL generated", map[string]interface{}{
		"user_id":     req.UserID,
		"file_name":   req.FileName,
		"file_key":    fileKey,
		"expires_at":   expiresAt,
	})

	return &GetUploadURLResponse{
		UploadURL:    uploadURL,
		FileKey:      fileKey,
		ExpiresAt:    expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		MaxFileSize:  uc.maxFileSize,
		AllowedTypes: uc.allowedTypes,
	}, nil
}

// validateRequest validates the get upload URL request
func (uc *GetUploadURLUseCase) validateRequest(req *GetUploadURLRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
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

	// Check file size limit
	if req.FileSize > uc.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", req.FileSize, uc.maxFileSize)
	}

	// Check content type
	if !uc.isContentTypeAllowed(req.ContentType) {
		return fmt.Errorf("content type %s is not allowed", req.ContentType)
	}

	return nil
}

// generateFileKey generates a unique file key for the user
func (uc *GetUploadURLUseCase) generateFileKey(userID uuid.UUID, fileName string) string {
	timestamp := time.Now().Unix()
	fileID := uuid.New()
	
	// Extract file extension
	ext := ""
	if parts := strings.Split(fileName, "."); len(parts) > 1 {
		ext = parts[len(parts)-1]
	}
	
	return fmt.Sprintf("photos/%s/temp/%d_%s.%s", userID.String(), timestamp, fileID.String(), ext)
}

// isContentTypeAllowed checks if the content type is allowed
func (uc *GetUploadURLUseCase) isContentTypeAllowed(contentType string) bool {
	for _, allowedType := range uc.allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}