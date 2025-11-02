package ephemeral_photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// GetEphemeralPhotoStatusRequest represents the request for getting ephemeral photo status
type GetEphemeralPhotoStatusRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
}

// GetEphemeralPhotoStatusResponse represents the response for getting ephemeral photo status
type GetEphemeralPhotoStatusResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	FileURL       string    `json:"file_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	AccessKey     string    `json:"access_key"`
	IsViewed      bool      `json:"is_viewed"`
	IsExpired     bool      `json:"is_expired"`
	ViewCount     int       `json:"view_count"`
	MaxViews      int       `json:"max_views"`
	ExpiresAt     string    `json:"expires_at"`
	ViewedAt      *string   `json:"viewed_at,omitempty"`
	ExpiredAt     *string   `json:"expired_at,omitempty"`
	IsDeleted     bool      `json:"is_deleted"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	ViewStatus    string    `json:"view_status"`
	RemainingTime int       `json:"remaining_time_seconds"` // Time remaining until expiration
	// Analytics
	TotalViews    int64 `json:"total_views"`
	UniqueViewers int64 `json:"unique_viewers"`
	AverageTime   int64 `json:"average_view_time_seconds"`
}

// GetEphemeralPhotoStatusUseCase handles getting ephemeral photo status
type GetEphemeralPhotoStatusUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewGetEphemeralPhotoStatusUseCase creates a new get ephemeral photo status use case
func NewGetEphemeralPhotoStatusUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *GetEphemeralPhotoStatusUseCase {
	return &GetEphemeralPhotoStatusUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the get ephemeral photo status use case
func (uc *GetEphemeralPhotoStatusUseCase) Execute(ctx context.Context, req *GetEphemeralPhotoStatusRequest) (*GetEphemeralPhotoStatusResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get the ephemeral photo
	photo, err := uc.ephemeralPhotoService.GetEphemeralPhoto(ctx, req.PhotoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ephemeral photo: %w", err)
	}

	// Check if user owns the photo
	if photo.UserID != req.UserID {
		return nil, fmt.Errorf("user does not own this photo")
	}

	// Get view statistics
	viewStats, err := uc.ephemeralPhotoService.GetPhotoViewStats(ctx, req.PhotoID)
	if err != nil {
		// Log error but continue - view stats are not critical
		viewStats = &services.ViewStats{
			TotalViews:     0,
			UniqueViewers:   0,
			AverageViewTime: 0,
		}
	}

	// Calculate remaining time
	remainingTime := int(photo.GetRemainingTime().Seconds())
	if remainingTime < 0 {
		remainingTime = 0
	}

	// Format timestamps
	var viewedAt, expiredAt *string
	if photo.ViewedAt != nil {
		formatted := photo.ViewedAt.Format("2006-01-02T15:04:05Z07:00")
		viewedAt = &formatted
	}
	if photo.ExpiredAt != nil {
		formatted := photo.ExpiredAt.Format("2006-01-02T15:04:05Z07:00")
		expiredAt = &formatted
	}

	// Create response
	response := &GetEphemeralPhotoStatusResponse{
		ID:            photo.ID,
		UserID:        photo.UserID,
		FileURL:       photo.FileURL,
		ThumbnailURL:  photo.ThumbnailURL,
		AccessKey:     photo.AccessKey,
		IsViewed:      photo.IsViewed,
		IsExpired:     photo.IsExpired,
		ViewCount:     photo.ViewCount,
		MaxViews:      photo.MaxViews,
		ExpiresAt:     photo.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		ViewedAt:      viewedAt,
		ExpiredAt:     expiredAt,
		IsDeleted:     photo.IsDeleted,
		CreatedAt:     photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     photo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ViewStatus:    photo.GetViewStatus(),
		RemainingTime: remainingTime,
		TotalViews:    viewStats.TotalViews,
		UniqueViewers: viewStats.UniqueViewers,
		AverageTime:   viewStats.AverageViewTime,
	}

	return response, nil
}