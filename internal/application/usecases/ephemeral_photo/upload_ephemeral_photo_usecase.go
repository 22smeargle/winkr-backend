package ephemeral_photo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// UploadEphemeralPhotoRequest represents the request for uploading an ephemeral photo
type UploadEphemeralPhotoRequest struct {
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	FileURL      string    `json:"file_url" validate:"required,url"`
	FileKey      string    `json:"file_key" validate:"required"`
	ThumbnailURL string    `json:"thumbnail_url" validate:"required,url"`
	ThumbnailKey string    `json:"thumbnail_key" validate:"required"`
	MaxViews     int       `json:"max_views" validate:"min=1,max=10"`
	ExpiresAfter int       `json:"expires_after_seconds" validate:"min=5,max=300"` // 5 seconds to 5 minutes
}

// UploadEphemeralPhotoResponse represents the response for uploading an ephemeral photo
type UploadEphemeralPhotoResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	FileURL       string    `json:"file_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	AccessKey     string    `json:"access_key"`
	MaxViews      int       `json:"max_views"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	ViewURL       string    `json:"view_url"` // URL for viewing the photo
}

// UploadEphemeralPhotoUseCase handles uploading ephemeral photos
type UploadEphemeralPhotoUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewUploadEphemeralPhotoUseCase creates a new upload ephemeral photo use case
func NewUploadEphemeralPhotoUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *UploadEphemeralPhotoUseCase {
	return &UploadEphemeralPhotoUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the upload ephemeral photo use case
func (uc *UploadEphemeralPhotoUseCase) Execute(ctx context.Context, req *UploadEphemeralPhotoRequest) (*UploadEphemeralPhotoResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert expires after seconds to duration
	expiresAfter := time.Duration(req.ExpiresAfter) * time.Second
	if expiresAfter == 0 {
		expiresAfter = 30 * time.Second // Default 30 seconds
	}

	// Upload ephemeral photo
	photo, err := uc.ephemeralPhotoService.UploadEphemeralPhoto(
		ctx,
		req.UserID,
		req.FileURL,
		req.FileKey,
		req.ThumbnailURL,
		req.ThumbnailKey,
		req.MaxViews,
		expiresAfter,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload ephemeral photo: %w", err)
	}

	// Create response
	response := &UploadEphemeralPhotoResponse{
		ID:            photo.ID,
		UserID:        photo.UserID,
		FileURL:       photo.FileURL,
		ThumbnailURL:  photo.ThumbnailURL,
		AccessKey:     photo.AccessKey,
		MaxViews:      photo.MaxViews,
		ExpiresAt:     photo.ExpiresAt,
		CreatedAt:     photo.CreatedAt,
		ViewURL:       fmt.Sprintf("/api/v1/ephemeral-photos/%s/view", photo.AccessKey),
	}

	return response, nil
}