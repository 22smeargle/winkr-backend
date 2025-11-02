package photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SetPrimaryPhotoRequest represents the request to set a photo as primary
type SetPrimaryPhotoRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
}

// SetPrimaryPhotoResponse represents the response after setting primary photo
type SetPrimaryPhotoResponse struct {
	PhotoID    uuid.UUID `json:"photo_id"`
	IsPrimary   bool      `json:"is_primary"`
	UpdatedAt   string    `json:"updated_at"`
	Message     string    `json:"message"`
}

// SetPrimaryPhotoUseCase handles setting primary photo logic
type SetPrimaryPhotoUseCase struct {
	photoRepo repositories.PhotoRepository
}

// NewSetPrimaryPhotoUseCase creates a new set primary photo use case
func NewSetPrimaryPhotoUseCase(photoRepo repositories.PhotoRepository) *SetPrimaryPhotoUseCase {
	return &SetPrimaryPhotoUseCase{
		photoRepo: photoRepo,
	}
}

// Execute executes the set primary photo use case
func (uc *SetPrimaryPhotoUseCase) Execute(ctx context.Context, req *SetPrimaryPhotoRequest) (*SetPrimaryPhotoResponse, error) {
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
		return nil, fmt.Errorf("photo must be approved to be set as primary")
	}

	// Check if photo is already primary
	if photo.IsPrimary {
		return &SetPrimaryPhotoResponse{
			PhotoID:  req.PhotoID,
			IsPrimary: true,
			UpdatedAt: photo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Message:   "Photo is already set as primary",
		}, nil
	}

	// Set photo as primary
	if err := uc.photoRepo.SetPrimaryPhoto(ctx, req.UserID, req.PhotoID); err != nil {
		return nil, fmt.Errorf("failed to set primary photo: %w", err)
	}

	logger.Info("Primary photo set successfully", map[string]interface{}{
		"photo_id": req.PhotoID,
		"user_id":  req.UserID,
	})

	return &SetPrimaryPhotoResponse{
		PhotoID:  req.PhotoID,
		IsPrimary: true,
		UpdatedAt: photo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Message:   "Photo set as primary successfully",
	}, nil
}

// validateRequest validates the set primary photo request
func (uc *SetPrimaryPhotoUseCase) validateRequest(req *SetPrimaryPhotoRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if req.PhotoID == uuid.Nil {
		return fmt.Errorf("photo ID is required")
	}

	return nil
}