package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// GetProfileUseCase handles getting user profile
type GetProfileUseCase struct {
	userRepo    repositories.UserRepository
	photoRepo   repositories.PhotoRepository
	cacheService ProfileCacheService
}

// NewGetProfileUseCase creates a new GetProfileUseCase instance
func NewGetProfileUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	cacheService ProfileCacheService,
) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo:     userRepo,
		photoRepo:    photoRepo,
		cacheService: cacheService,
	}
}

// GetProfileRequest represents get profile request
type GetProfileRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// GetProfileResponse represents get profile response
type GetProfileResponse struct {
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

// Location represents user location
type Location struct {
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
	City     *string  `json:"city"`
	Country  *string  `json:"country"`
}

// Photo represents user photo in profile response
type Photo struct {
	ID                string  `json:"id"`
	URL               string  `json:"url"`
	IsPrimary         bool    `json:"is_primary"`
	VerificationStatus string  `json:"verification_status"`
	VerificationReason *string `json:"verification_reason,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

// Preferences represents user preferences
type Preferences struct {
	AgeMin      int  `json:"age_min"`
	AgeMax      int  `json:"age_max"`
	MaxDistance int  `json:"max_distance"`
	ShowMe      bool `json:"show_me"`
}

// ProfileStats represents user profile statistics
type ProfileStats struct {
	ProfileViews    int64 `json:"profile_views"`
	PhotosCount     int64 `json:"photos_count"`
	ProfileComplete bool  `json:"profile_complete"`
	LastActive     string `json:"last_active"`
}

// Execute handles the get profile use case
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*GetProfileResponse, error) {
	// Try to get from cache first
	cachedProfile, err := uc.cacheService.GetProfile(ctx, userID)
	if err == nil && cachedProfile != nil {
		return cachedProfile, nil
	}

	// Get user from database
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is active and not banned
	if !user.IsActive || user.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Get user photos
	photos, err := uc.photoRepo.GetUserPhotos(ctx, userID, false)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user photos")
	}

	// Get user preferences
	preferences, err := uc.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user preferences")
	}

	// Get user statistics
	stats, err := uc.userRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user statistics")
	}

	// Build response
	response := &GetProfileResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		DateOfBirth:   user.DateOfBirth.Format("2006-01-02"),
		Gender:        user.Gender,
		InterestedIn:  user.InterestedIn,
		Bio:           user.Bio,
		Location: &Location{
			Lat:      user.LocationLat,
			Lng:      user.LocationLng,
			City:     user.LocationCity,
			Country:  user.LocationCountry,
		},
		IsVerified:    user.IsVerified,
		IsPremium:     user.IsPremium,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
	if preferences != nil {
		response.Preferences = &Preferences{
			AgeMin:      preferences.AgeMin,
			AgeMax:      preferences.AgeMax,
			MaxDistance: preferences.MaxDistance,
			ShowMe:      preferences.ShowMe,
		}
	}

	// Add statistics to response
	response.ProfileStats = &ProfileStats{
		ProfileViews:    stats.ProfileViews,
		PhotosCount:     stats.PhotosCount,
		ProfileComplete: user.IsComplete(),
		LastActive:     formatLastActive(user.LastActive),
	}

	// Cache the profile
	if err := uc.cacheService.SetProfile(ctx, userID, response, 15*time.Minute); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return response, nil
}

// formatLastActive formats the last active time
func formatLastActive(lastActive *time.Time) string {
	if lastActive == nil {
		return ""
	}
	return lastActive.Format("2006-01-02T15:04:05Z07:00")
}