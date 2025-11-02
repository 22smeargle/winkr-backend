package photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MarkPhotoViewedRequest represents a request to mark a photo as viewed
type MarkPhotoViewedRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
	ViewerID uuid.UUID `json:"viewer_id" validate:"required"`
}

// MarkPhotoViewedResponse represents the response after marking a photo as viewed
type MarkPhotoViewedResponse struct {
	PhotoID  uuid.UUID `json:"photo_id"`
	ViewerID uuid.UUID `json:"viewer_id"`
	ViewedAt string    `json:"viewed_at"`
	Message   string    `json:"message"`
}

// MarkPhotoViewedUseCase handles marking photos as viewed logic
type MarkPhotoViewedUseCase struct {
	photoRepo repositories.PhotoRepository
}

// NewMarkPhotoViewedUseCase creates a new mark photo viewed use case
func NewMarkPhotoViewedUseCase(photoRepo repositories.PhotoRepository) *MarkPhotoViewedUseCase {
	return &MarkPhotoViewedUseCase{
		photoRepo: photoRepo,
	}
}

// Execute executes the mark photo viewed use case
func (uc *MarkPhotoViewedUseCase) Execute(ctx context.Context, req *MarkPhotoViewedRequest) (*MarkPhotoViewedResponse, error) {
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

	// Check if photo is verified
	if photo.VerificationStatus != "approved" {
		return nil, fmt.Errorf("photo must be approved to be viewed")
	}

	// Check if viewer is different from photo owner
	if req.ViewerID == req.UserID {
		return nil, fmt.Errorf("viewer cannot be the photo owner")
	}

	// Log the photo view (this would typically be stored in a separate analytics table)
	// For now, we'll just log it and return success
	uc.logPhotoView(ctx, req.PhotoID, req.UserID, req.ViewerID)

	logger.Info("Photo marked as viewed", map[string]interface{}{
		"photo_id":  req.PhotoID,
		"user_id":   req.UserID,
		"viewer_id": req.ViewerID,
	})

	return &MarkPhotoViewedResponse{
		PhotoID:  req.PhotoID,
		ViewerID: req.ViewerID,
		ViewedAt: photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Message:   "Photo marked as viewed successfully",
	}, nil
}

// validateRequest validates the mark photo viewed request
func (uc *MarkPhotoViewedUseCase) validateRequest(req *MarkPhotoViewedRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if req.PhotoID == uuid.Nil {
		return fmt.Errorf("photo ID is required")
	}

	if req.ViewerID == uuid.Nil {
		return fmt.Errorf("viewer ID is required")
	}

	return nil
}

// logPhotoView logs when a photo is viewed by another user
func (uc *MarkPhotoViewedUseCase) logPhotoView(ctx context.Context, photoID, userID, viewerID uuid.UUID) {
	// This would typically be stored in a separate analytics table
	// For now, we'll just log it
	logger.Info("Photo viewed", map[string]interface{}{
		"photo_id":  photoID,
		"user_id":   userID,
		"viewer_id": viewerID,
		"viewed_at": "now",
	})
}