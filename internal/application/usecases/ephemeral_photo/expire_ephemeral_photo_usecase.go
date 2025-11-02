package ephemeral_photo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// ExpireEphemeralPhotoRequest represents the request for expiring an ephemeral photo
type ExpireEphemeralPhotoRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	PhotoID uuid.UUID `json:"photo_id" validate:"required"`
}

// ExpireEphemeralPhotoResponse represents the response for expiring an ephemeral photo
type ExpireEphemeralPhotoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ExpireEphemeralPhotoUseCase handles expiring ephemeral photos
type ExpireEphemeralPhotoUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewExpireEphemeralPhotoUseCase creates a new expire ephemeral photo use case
func NewExpireEphemeralPhotoUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *ExpireEphemeralPhotoUseCase {
	return &ExpireEphemeralPhotoUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the expire ephemeral photo use case
func (uc *ExpireEphemeralPhotoUseCase) Execute(ctx context.Context, req *ExpireEphemeralPhotoRequest) (*ExpireEphemeralPhotoResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Expire the ephemeral photo
	err := uc.ephemeralPhotoService.ExpireEphemeralPhoto(ctx, req.UserID, req.PhotoID)
	if err != nil {
		return nil, fmt.Errorf("failed to expire ephemeral photo: %w", err)
	}

	// Create response
	response := &ExpireEphemeralPhotoResponse{
		Success: true,
		Message: "Ephemeral photo expired successfully",
	}

	return response, nil
}