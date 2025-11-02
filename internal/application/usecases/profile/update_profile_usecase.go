package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// UpdateProfileUseCase handles updating user profile
type UpdateProfileUseCase struct {
	userRepo     repositories.UserRepository
	photoRepo    repositories.PhotoRepository
	cacheService ProfileCacheService
	profileService ProfileService
}

// NewUpdateProfileUseCase creates a new UpdateProfileUseCase instance
func NewUpdateProfileUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	cacheService ProfileCacheService,
	profileService ProfileService,
) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		userRepo:      userRepo,
		photoRepo:     photoRepo,
		cacheService:  cacheService,
		profileService: profileService,
	}
}

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	UserID      uuid.UUID    `json:"user_id"`
	FirstName    *string      `json:"first_name"`
	LastName     *string      `json:"last_name"`
	Bio          *string      `json:"bio"`
	InterestedIn []string     `json:"interested_in"`
	Preferences  *Preferences `json:"preferences"`
}

// UpdateProfileResponse represents update profile response
type UpdateProfileResponse struct {
	ID             uuid.UUID    `json:"id"`
	Email          string       `json:"email"`
	FirstName      string       `json:"first_name"`
	LastName       string       `json:"last_name"`
	DateOfBirth    string       `json:"date_of_birth"`
	Gender         string       `json:"gender"`
	InterestedIn   []string     `json:"interested_in"`
	Bio            *string      `json:"bio"`
	Location       *Location    `json:"location"`
	IsVerified     bool         `json:"is_verified"`
	IsPremium      bool         `json:"is_premium"`
	Photos         []*Photo     `json:"photos"`
	Preferences    *Preferences `json:"preferences"`
	ProfileStats   *ProfileStats `json:"profile_stats"`
	CreatedAt      string       `json:"created_at"`
	UpdatedAt      string       `json:"updated_at"`
}

// Execute handles the update profile use case
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	// Get current user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is active and not banned
	if !user.IsActive || user.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Validate profile updates
	if err := uc.profileService.ValidateProfileUpdate(user, req); err != nil {
		return nil, err
	}

	// Update user fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if len(req.InterestedIn) > 0 {
		user.InterestedIn = req.InterestedIn
	}

	// Update user in database
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errors.WrapError(err, "Failed to update user profile")
	}

	// Update preferences if provided
	if req.Preferences != nil {
		preferences, err := uc.userRepo.GetPreferences(ctx, req.UserID)
		if err != nil {
			// Create preferences if they don't exist
			preferences = &repositories.UserPreferencesEntity{
				UserID: req.UserID,
			}
		}

		// Update preference fields
		preferences.AgeMin = req.Preferences.AgeMin
		preferences.AgeMax = req.Preferences.AgeMax
		preferences.MaxDistance = req.Preferences.MaxDistance
		preferences.ShowMe = req.Preferences.ShowMe

		// Save preferences
		if preferences.ID == uuid.Nil {
			if err := uc.userRepo.CreatePreferences(ctx, preferences); err != nil {
				return nil, errors.WrapError(err, "Failed to create user preferences")
			}
		} else {
			if err := uc.userRepo.UpdatePreferences(ctx, preferences); err != nil {
				return nil, errors.WrapError(err, "Failed to update user preferences")
			}
		}
	}

	// Invalidate cache
	if err := uc.cacheService.DeleteProfile(ctx, req.UserID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	// Get updated user data
	updatedUser, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get updated user")
	}

	// Get user photos
	photos, err := uc.photoRepo.GetUserPhotos(ctx, req.UserID, false)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user photos")
	}

	// Get updated preferences
	updatedPreferences, err := uc.userRepo.GetPreferences(ctx, req.UserID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get updated user preferences")
	}

	// Get user statistics
	stats, err := uc.userRepo.GetUserStats(ctx, req.UserID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user statistics")
	}

	// Build response
	response := &UpdateProfileResponse{
		ID:           updatedUser.ID,
		Email:        updatedUser.Email,
		FirstName:     updatedUser.FirstName,
		LastName:      updatedUser.LastName,
		DateOfBirth:   updatedUser.DateOfBirth.Format("2006-01-02"),
		Gender:        updatedUser.Gender,
		InterestedIn:  updatedUser.InterestedIn,
		Bio:           updatedUser.Bio,
		Location: &Location{
			Lat:      updatedUser.LocationLat,
			Lng:      updatedUser.LocationLng,
			City:     updatedUser.LocationCity,
			Country:  updatedUser.LocationCountry,
		},
		IsVerified:    updatedUser.IsVerified,
		IsPremium:     updatedUser.IsPremium,
		CreatedAt:     updatedUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Add photos to response
	for _, photo := range photos {
		responsePhoto := &Photo{
			ID:                photo.ID.String(),
			URL:               photo.FileURL,
			IsPrimary:         photo.IsPrimary,
			VerificationStatus: photo.VerificationStatus,
			VerificationReason: photo.VerificationReason,
			CreatedAt:         photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		response.Photos = append(response.Photos, responsePhoto)
	}

	// Add preferences to response
	if updatedPreferences != nil {
		response.Preferences = &Preferences{
			AgeMin:      updatedPreferences.AgeMin,
			AgeMax:      updatedPreferences.AgeMax,
			MaxDistance: updatedPreferences.MaxDistance,
			ShowMe:      updatedPreferences.ShowMe,
		}
	}

	// Add statistics to response
	response.ProfileStats = &ProfileStats{
		ProfileViews:    stats.ProfileViews,
		PhotosCount:     stats.PhotosCount,
		ProfileComplete: updatedUser.IsComplete(),
		LastActive:     formatLastActive(updatedUser.LastActive),
	}

	return response, nil
}