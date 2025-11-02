package photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// DeletePhotoRequest represents the request to delete a photo
type DeletePhotoRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
}

// DeletePhotoResponse represents the response after deleting a photo
type DeletePhotoResponse struct {
	PhotoID    uuid.UUID `json:"photo_id"`
	DeletedAt   string    `json:"deleted_at"`
	Message     string    `json:"message"`
}

// DeletePhotoUseCase handles photo deletion logic
type DeletePhotoUseCase struct {
	photoRepo      repositories.PhotoRepository
	storageService storage.StorageService
}

// NewDeletePhotoUseCase creates a new delete photo use case
func NewDeletePhotoUseCase(
	photoRepo repositories.PhotoRepository,
	storageService storage.StorageService,
) *DeletePhotoUseCase {
	return &DeletePhotoUseCase{
		photoRepo:      photoRepo,
		storageService: storageService,
	}
}

// Execute executes the delete photo use case
func (uc *DeletePhotoUseCase) Execute(ctx context.Context, req *DeletePhotoRequest) (*DeletePhotoResponse, error) {
	// Validate request
	if err := uc.validateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if photo exists and belongs to user
	photo, err := uc.photoRepo.GetByID(ctx, req.PhotoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	if photo.UserID != req.UserID {
		return nil, fmt.Errorf("photo does not belong to user")
	}

	// Check if photo can be deleted (primary photos cannot be deleted unless there are other photos)
	if photo.IsPrimary {
		userPhotos, err := uc.photoRepo.GetUserPhotos(ctx, req.UserID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get user photos: %w", err)
		}

		// Count non-deleted photos
		activePhotoCount := 0
		for _, p := range userPhotos {
			if !p.IsDeleted && p.ID != req.PhotoID {
				activePhotoCount++
			}
		}

		if activePhotoCount == 0 {
			return nil, fmt.Errorf("cannot delete primary photo - you must have at least one photo")
		}
	}

	// Get file key for storage deletion
	fileKey := photo.FileKey

	// Soft delete photo in database
	if err := uc.photoRepo.SoftDeletePhoto(ctx, req.PhotoID); err != nil {
		return nil, fmt.Errorf("failed to delete photo: %w", err)
	}

	// Delete file from storage (async operation - don't fail if storage deletion fails)
	go func() {
		if err := uc.storageService.DeleteFile(context.Background(), fileKey); err != nil {
			logger.Error("Failed to delete photo file from storage", err, map[string]interface{}{
				"photo_id": req.PhotoID,
				"file_key": fileKey,
			})
		} else {
			logger.Info("Photo file deleted from storage", map[string]interface{}{
				"photo_id": req.PhotoID,
				"file_key": fileKey,
			})
		}
	}()

	// If this was a primary photo, set another photo as primary if available
	if photo.IsPrimary {
		go func() {
			userPhotos, err := uc.photoRepo.GetUserPhotos(context.Background(), req.UserID, false)
			if err != nil {
				logger.Error("Failed to get user photos for primary reassignment", err)
				return
			}

			// Find first non-deleted photo to set as primary
			for _, p := range userPhotos {
				if !p.IsDeleted && p.ID != req.PhotoID {
					if err := uc.photoRepo.SetPrimaryPhoto(context.Background(), req.UserID, p.ID); err != nil {
						logger.Error("Failed to set new primary photo", err, map[string]interface{}{
							"user_id":   req.UserID,
							"photo_id":  p.ID,
						})
					} else {
						logger.Info("New primary photo set", map[string]interface{}{
							"user_id":   req.UserID,
							"photo_id":  p.ID,
						})
					}
					break
				}
			}
		}()
	}

	logger.Info("Photo deleted successfully", map[string]interface{}{
		"photo_id":   req.PhotoID,
		"user_id":    req.UserID,
		"was_primary": photo.IsPrimary,
	})

	return &DeletePhotoResponse{
		PhotoID:  req.PhotoID,
		DeletedAt: photo.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Message:   "Photo deleted successfully",
	}, nil
}

// validateRequest validates the delete request
func (uc *DeletePhotoUseCase) validateRequest(req *DeletePhotoRequest) error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if req.PhotoID == uuid.Nil {
		return fmt.Errorf("photo ID is required")
	}

	return nil
}