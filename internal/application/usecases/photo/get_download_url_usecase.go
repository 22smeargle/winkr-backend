package photo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetDownloadURLRequest represents the request to get a download URL
type GetDownloadURLRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
	ViewerID uuid.UUID `json:"viewer_id,omitempty"` // Optional: who is viewing the photo
}

// GetDownloadURLResponse represents the response with download URL
type GetDownloadURLResponse struct {
	DownloadURL string    `json:"download_url"`
	ExpiresAt   string    `json:"expires_at"`
	PhotoInfo   *PhotoInfo `json:"photo_info"`
}

// PhotoInfo represents basic photo information for download
type PhotoInfo struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	VerificationStatus string    `json:"verification_status"`
	IsPrimary         bool      `json:"is_primary"`
	CreatedAt         string    `json:"created_at"`
}

// GetDownloadURLUseCase handles getting download URL logic
type GetDownloadURLUseCase struct {
	photoRepo      repositories.PhotoRepository
	storageService storage.StorageService
}

// NewGetDownloadURLUseCase creates a new get download URL use case
func NewGetDownloadURLUseCase(
	photoRepo repositories.PhotoRepository,
	storageService storage.StorageService,
) *GetDownloadURLUseCase {
	return &GetDownloadURLUseCase{
		photoRepo:      photoRepo,
		storageService: storageService,
	}
}

// Execute executes the get download URL use case
func (uc *GetDownloadURLUseCase) Execute(ctx context.Context, req *GetDownloadURLRequest) (*GetDownloadURLResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get photo from database
	photo, err := uc.photoRepo.GetByID(ctx, req.PhotoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	// Check if photo belongs to user
	if photo.UserID != req.UserID {
		return nil, fmt.Errorf("photo does not belong to user")
	}

	// Check if photo is deleted
	if photo.IsDeleted {
		return nil, fmt.Errorf("photo has been deleted")
	}

	// Check access permissions based on verification status
	if !uc.canAccessPhoto(photo, req.UserID, req.ViewerID) {
		return nil, fmt.Errorf("access denied to photo")
	}

	// Generate presigned download URL
	downloadURL, err := uc.storageService.GetDownloadURL(ctx, photo.FileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(1 * time.Hour) // 1 hour from now

	// Prepare photo info
	photoInfo := &PhotoInfo{
		ID:                photo.ID,
		UserID:            photo.UserID,
		VerificationStatus: photo.VerificationStatus,
		IsPrimary:         photo.IsPrimary,
		CreatedAt:         photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Log access for analytics (async)
	go func() {
		if req.ViewerID != uuid.Nil && req.ViewerID != req.UserID {
			// This is someone else viewing the user's photo
			uc.logPhotoView(context.Background(), req.PhotoID, req.UserID, req.ViewerID)
		}
	}()

	logger.Info("Download URL generated", map[string]interface{}{
		"photo_id":   req.PhotoID,
		"user_id":    req.UserID,
		"viewer_id":   req.ViewerID,
		"expires_at":  expiresAt,
	})

	return &GetDownloadURLResponse{
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		PhotoInfo:   photoInfo,
	}, nil
}

// validateRequest validates the get download URL request
func (uc *GetDownloadURLUseCase) validateRequest(req *GetDownloadURLRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if req.PhotoID == uuid.Nil {
		return fmt.Errorf("photo ID is required")
	}

	return nil
}

// canAccessPhoto checks if the user can access the photo
func (uc *GetDownloadURLUseCase) canAccessPhoto(photo *entities.Photo, userID, viewerID uuid.UUID) bool {
	// User can always access their own photos
	if userID == photo.UserID {
		return true
	}

	// Only approved photos can be accessed by other users
	if photo.VerificationStatus != "approved" {
		return false
	}

	// If no viewer ID specified, deny access
	if viewerID == uuid.Nil {
		return false
	}

	// Viewer can access approved photos
	return true
}

// logPhotoView logs when a photo is viewed by another user
func (uc *GetDownloadURLUseCase) logPhotoView(ctx context.Context, photoID, userID, viewerID uuid.UUID) {
	// This would typically be stored in a separate analytics table
	// For now, we'll just log it
	logger.Info("Photo viewed", map[string]interface{}{
		"photo_id":  photoID,
		"user_id":   userID,
		"viewer_id": viewerID,
		"viewed_at": time.Now(),
	})
}