package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// GetProfileUseCase handles getting user profile
type GetProfileUseCase struct {
	authService services.AuthService
}

// NewGetProfileUseCase creates a new GetProfileUseCase instance
func NewGetProfileUseCase(authService services.AuthService) *GetProfileUseCase {
	return &GetProfileUseCase{
		authService: authService,
	}
}

// GetProfileRequest represents get profile request
type GetProfileRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// GetProfileResponse represents get profile response
type GetProfileResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	DateOfBirth   string    `json:"date_of_birth"`
	Gender        string    `json:"gender"`
	InterestedIn  []string  `json:"interested_in"`
	Bio           *string   `json:"bio"`
	LocationLat   *float64  `json:"location_lat"`
	LocationLng   *float64  `json:"location_lng"`
	LocationCity  *string   `json:"location_city"`
	LocationCountry *string  `json:"location_country"`
	IsVerified    bool      `json:"is_verified"`
	IsPremium     bool      `json:"is_premium"`
	IsActive      bool      `json:"is_active"`
	LastActive    *string   `json:"last_active"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

// Execute handles the get profile use case
func (uc *GetProfileUseCase) Execute(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	// Convert user ID string to UUID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.NewValidationError("user_id", "Invalid user ID format")
	}

	// TODO: Get user from repository
	// For now, we'll return a placeholder response
	// user, err := uc.userRepo.GetByID(ctx, userID)
	// if err != nil {
	//     return nil, err
	// }

	// Convert to response
	response := &GetProfileResponse{
		ID:          userID,
		Email:       "user@example.com", // placeholder
		FirstName:    "John",           // placeholder
		LastName:     "Doe",            // placeholder
		IsVerified:   false,            // placeholder
		IsPremium:    false,            // placeholder
		IsActive:     true,             // placeholder
		CreatedAt:    "2025-01-01T00:00:00Z", // placeholder
		UpdatedAt:    "2025-01-01T00:00:00Z", // placeholder
	}

	return response, nil
}