package ephemeral_photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// DeleteEphemeralPhotoRequest represents the request for deleting an ephemeral photo
type DeleteEphemeralPhotoRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
}

// DeleteEphemeralPhotoResponse represents the response for deleting an ephemeral photo
type DeleteEphemeralPhotoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteEphemeralPhotoUseCase handles deleting ephemeral photos
type DeleteEphemeralPhotoUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewDeleteEphemeralPhotoUseCase creates a new delete ephemeral photo use case
func NewDeleteEphemeralPhotoUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *DeleteEphemeralPhotoUseCase {
	return &DeleteEphemeralPhotoUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the delete ephemeral photo use case
func (uc *DeleteEphemeralPhotoUseCase) Execute(ctx context.Context, req *DeleteEphemeralPhotoRequest) (*DeleteEphemeralPhotoResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Delete the ephemeral photo
	err := uc.ephemeralPhotoService.DeleteEphemeralPhoto(ctx, req.UserID, req.PhotoID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete ephemeral photo: %w", err)
	}

	// Create response
	response := &DeleteEphemeralPhotoResponse{
		Success: true,
		Message: "Ephemeral photo deleted successfully",
	}

	return response, nil
}