package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// UpdateUserUseCase handles updating user information
type UpdateUserUseCase struct {
	userRepo repositories.UserRepository
}

// NewUpdateUserUseCase creates a new UpdateUserUseCase
func NewUpdateUserUseCase(userRepo repositories.UserRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo: userRepo,
	}
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	AdminID    uuid.UUID  `json:"admin_id" validate:"required"`
	UserID     uuid.UUID  `json:"user_id" validate:"required"`
	FirstName   *string    `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName    *string    `json:"last_name" validate:"omitempty,min=1,max=100"`
	Bio         *string    `json:"bio" validate:"omitempty,max=500"`
	IsVerified  *bool      `json:"is_verified"`
	IsPremium   *bool      `json:"is_premium"`
	IsActive    *bool      `json:"is_active"`
	IsBanned    *bool      `json:"is_banned"`
	LocationLat *float64   `json:"location_lat" validate:"omitempty,min=-90,max=90"`
	LocationLng *float64   `json:"location_lng" validate:"omitempty,min=-180,max=180"`
	LocationCity *string    `json:"location_city" validate:"omitempty,max=100"`
	LocationCountry *string `json:"location_country" validate:"omitempty,max=100"`
	Reason      string     `json:"reason" validate:"required,min=5,max=500"`
}

// UpdateUserResponse represents the response from updating a user
type UpdateUserResponse struct {
	User      *entities.User `json:"user"`
	Updated   []string       `json:"updated_fields"`
	Timestamp time.Time      `json:"timestamp"`
}

// Execute updates user information
func (uc *UpdateUserUseCase) Execute(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
	logger.Info("UpdateUser use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "reason", req.Reason)

	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Track updated fields
	updatedFields := make([]string, 0)

	// Update fields if provided
	if req.FirstName != nil && *req.FirstName != user.FirstName {
		user.FirstName = *req.FirstName
		updatedFields = append(updatedFields, "first_name")
	}

	if req.LastName != nil && *req.LastName != user.LastName {
		user.LastName = *req.LastName
		updatedFields = append(updatedFields, "last_name")
	}

	if req.Bio != nil {
		if (req.Bio == nil && user.Bio != nil) || (req.Bio != nil && user.Bio == nil) || (req.Bio != nil && user.Bio != nil && *req.Bio != *user.Bio) {
			user.Bio = req.Bio
			updatedFields = append(updatedFields, "bio")
		}
	}

	if req.IsVerified != nil && *req.IsVerified != user.IsVerified {
		user.IsVerified = *req.IsVerified
		updatedFields = append(updatedFields, "is_verified")
	}

	if req.IsPremium != nil && *req.IsPremium != user.IsPremium {
		user.IsPremium = *req.IsPremium
		updatedFields = append(updatedFields, "is_premium")
	}

	if req.IsActive != nil && *req.IsActive != user.IsActive {
		user.IsActive = *req.IsActive
		updatedFields = append(updatedFields, "is_active")
	}

	if req.IsBanned != nil && *req.IsBanned != user.IsBanned {
		user.IsBanned = *req.IsBanned
		updatedFields = append(updatedFields, "is_banned")
	}

	if req.LocationLat != nil {
		if (req.LocationLat == nil && user.LocationLat != nil) || (req.LocationLat != nil && user.LocationLat == nil) || (req.LocationLat != nil && user.LocationLat != nil && *req.LocationLat != *user.LocationLat) {
			user.LocationLat = req.LocationLat
			updatedFields = append(updatedFields, "location_lat")
		}
	}

	if req.LocationLng != nil {
		if (req.LocationLng == nil && user.LocationLng != nil) || (req.LocationLng != nil && user.LocationLng == nil) || (req.LocationLng != nil && user.LocationLng != nil && *req.LocationLng != *user.LocationLng) {
			user.LocationLng = req.LocationLng
			updatedFields = append(updatedFields, "location_lng")
		}
	}

	if req.LocationCity != nil {
		if (req.LocationCity == nil && user.LocationCity != nil) || (req.LocationCity != nil && user.LocationCity == nil) || (req.LocationCity != nil && user.LocationCity != nil && *req.LocationCity != *user.LocationCity) {
			user.LocationCity = req.LocationCity
			updatedFields = append(updatedFields, "location_city")
		}
	}

	if req.LocationCountry != nil {
		if (req.LocationCountry == nil && user.LocationCountry != nil) || (req.LocationCountry != nil && user.LocationCountry == nil) || (req.LocationCountry != nil && user.LocationCountry != nil && *req.LocationCountry != *user.LocationCountry) {
			user.LocationCountry = req.LocationCountry
			updatedFields = append(updatedFields, "location_country")
		}
	}

	// If no fields were updated, return early
	if len(updatedFields) == 0 {
		return &UpdateUserResponse{
			User:      user,
			Updated:   updatedFields,
			Timestamp: time.Now(),
		}, nil
	}

	// Update the user in the repository
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		logger.Error("Failed to update user in repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the admin action
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "update_user", map[string]interface{}{
		"updated_fields": updatedFields,
		"reason":         req.Reason,
	})

	logger.Info("UpdateUser use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID, "updated_fields", updatedFields)
	return &UpdateUserResponse{
		User:      user,
		Updated:   updatedFields,
		Timestamp: time.Now(),
	}, nil
}

// logAdminAction logs an admin action for audit purposes
func (uc *UpdateUserUseCase) logAdminAction(ctx context.Context, adminID, userID uuid.UUID, action string, metadata map[string]interface{}) {
	// In a real implementation, this would log to an audit table or service
	logger.Info("Admin action logged", 
		"admin_id", adminID,
		"user_id", userID,
		"action", action,
		"metadata", metadata,
		"timestamp", time.Now(),
	)
}