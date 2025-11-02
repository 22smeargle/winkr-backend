package ephemeral_photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// GetUserEphemeralPhotosRequest represents the request for getting user's ephemeral photos
type GetUserEphemeralPhotosRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	ActiveOnly bool      `json:"active_only"` // If true, only return active photos
	Limit      int       `json:"limit" validate:"min=1,max=50"`
	Offset     int       `json:"offset" validate:"min=0"`
}

// GetUserEphemeralPhotosResponse represents the response for getting user's ephemeral photos
type GetUserEphemeralPhotosResponse struct {
	Photos []EphemeralPhotoSummary `json:"photos"`
	Total  int                    `json:"total"`
	Stats  *UserPhotoStats         `json:"stats,omitempty"`
}

// EphemeralPhotoSummary represents a summary of an ephemeral photo
type EphemeralPhotoSummary struct {
	ID            uuid.UUID `json:"id"`
	FileURL       string    `json:"file_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	AccessKey     string    `json:"access_key"`
	IsViewed      bool      `json:"is_viewed"`
	IsExpired     bool      `json:"is_expired"`
	ViewCount     int       `json:"view_count"`
	MaxViews      int       `json:"max_views"`
	ExpiresAt     string    `json:"expires_at"`
	ViewStatus    string    `json:"view_status"`
	RemainingTime int       `json:"remaining_time_seconds"`
	CreatedAt     string    `json:"created_at"`
}

// UserPhotoStats represents user's ephemeral photo statistics
type UserPhotoStats struct {
	TotalPhotos     int64 `json:"total_photos"`
	ActivePhotos    int64 `json:"active_photos"`
	ViewedPhotos    int64 `json:"viewed_photos"`
	ExpiredPhotos   int64 `json:"expired_photos"`
	PhotosToday     int64 `json:"photos_today"`
	PhotosThisWeek  int64 `json:"photos_this_week"`
	PhotosThisMonth int64 `json:"photos_this_month"`
	TotalViews      int64 `json:"total_views"`
}

// GetUserEphemeralPhotosUseCase handles getting user's ephemeral photos
type GetUserEphemeralPhotosUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewGetUserEphemeralPhotosUseCase creates a new get user ephemeral photos use case
func NewGetUserEphemeralPhotosUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *GetUserEphemeralPhotosUseCase {
	return &GetUserEphemeralPhotosUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the get user ephemeral photos use case
func (uc *GetUserEphemeralPhotosUseCase) Execute(ctx context.Context, req *GetUserEphemeralPhotosRequest) (*GetUserEphemeralPhotosResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20
	}

	var photos []*services.EphemeralPhoto
	var err error

	// Get photos based on request
	if req.ActiveOnly {
		photos, err = uc.ephemeralPhotoService.GetUserActiveEphemeralPhotos(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user active ephemeral photos: %w", err)
		}
	} else {
		photos, err = uc.ephemeralPhotoService.GetUserEphemeralPhotos(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user ephemeral photos: %w", err)
		}
	}

	// Apply pagination
	total := len(photos)
	start := req.Offset
	end := start + req.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedPhotos := photos[start:end]

	// Convert to response format
	photoSummaries := make([]EphemeralPhotoSummary, len(paginatedPhotos))
	for i, photo := range paginatedPhotos {
		remainingTime := int(photo.GetRemainingTime().Seconds())
		if remainingTime < 0 {
			remainingTime = 0
		}

		photoSummaries[i] = EphemeralPhotoSummary{
			ID:            photo.ID,
			FileURL:       photo.FileURL,
			ThumbnailURL:  photo.ThumbnailURL,
			AccessKey:     photo.AccessKey,
			IsViewed:      photo.IsViewed,
			IsExpired:     photo.IsExpired,
			ViewCount:     photo.ViewCount,
			MaxViews:      photo.MaxViews,
			ExpiresAt:     photo.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			ViewStatus:    photo.GetViewStatus(),
			RemainingTime: remainingTime,
			CreatedAt:     photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Create response
	response := &GetUserEphemeralPhotosResponse{
		Photos: photoSummaries,
		Total:  total,
	}

	// Include stats if requested
	if req.Limit >= 20 && req.Offset == 0 {
		// Only include stats for first page to avoid performance issues
		stats, err := uc.ephemeralPhotoService.GetUserEphemeralPhotoStats(ctx, req.UserID)
		if err == nil {
			response.Stats = &UserPhotoStats{
				TotalPhotos:     stats.TotalPhotos,
				ActivePhotos:    stats.ActivePhotos,
				ViewedPhotos:    stats.ViewedPhotos,
				ExpiredPhotos:   stats.ExpiredPhotos,
				PhotosToday:     stats.PhotosToday,
				PhotosThisWeek:  stats.PhotosThisWeek,
				PhotosThisMonth: stats.PhotosThisMonth,
				TotalViews:      stats.TotalViews,
			}
		}
	}

	return response, nil
}