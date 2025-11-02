package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// ViewUserProfileUseCase handles viewing another user's profile
type ViewUserProfileUseCase struct {
	userRepo     repositories.UserRepository
	photoRepo    repositories.PhotoRepository
	matchRepo    repositories.MatchRepository
	cacheService ProfileCacheService
	privacyService ProfilePrivacyService
}

// NewViewUserProfileUseCase creates a new ViewUserProfileUseCase instance
func NewViewUserProfileUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	matchRepo repositories.MatchRepository,
	cacheService ProfileCacheService,
	privacyService ProfilePrivacyService,
) *ViewUserProfileUseCase {
	return &ViewUserProfileUseCase{
		userRepo:      userRepo,
		photoRepo:     photoRepo,
		matchRepo:     matchRepo,
		cacheService:  cacheService,
		privacyService: privacyService,
	}
}

// ViewUserProfileRequest represents view user profile request
type ViewUserProfileRequest struct {
	ViewerID    uuid.UUID `json:"viewer_id"`
	TargetUserID uuid.UUID `json:"target_user_id"`
}

// ViewUserProfileResponse represents view user profile response
type ViewUserProfileResponse struct {
	ID             uuid.UUID    `json:"id"`
	FirstName      string       `json:"first_name"`
	Age           int          `json:"age"`
	Bio            *string      `json:"bio"`
	Location       *Location    `json:"location"`
	IsVerified     bool         `json:"is_verified"`
	Photos         []*Photo     `json:"photos"`
	ProfileStats   *ProfileStats `json:"profile_stats"`
	IsMatch        bool         `json:"is_match"`
	CanMessage    bool         `json:"can_message"`
	CreatedAt      string       `json:"created_at"`
}

// Execute handles the view user profile use case
func (uc *ViewUserProfileUseCase) Execute(ctx context.Context, viewerID, targetUserID uuid.UUID) (*ViewUserProfileResponse, error) {
	// Check if viewer is trying to view their own profile
	if viewerID == targetUserID {
		return nil, errors.ErrCannotViewOwnProfile
	}

	// Try to get from cache first
	cacheKey := generateViewProfileCacheKey(viewerID, targetUserID)
	cachedProfile, err := uc.cacheService.GetViewProfile(ctx, cacheKey)
	if err == nil && cachedProfile != nil {
		return cachedProfile, nil
	}

	// Get target user from database
	targetUser, err := uc.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if target user is active and not banned
	if !targetUser.IsActive || targetUser.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Check privacy settings
	if !uc.privacyService.CanViewProfile(ctx, viewerID, targetUserID) {
		return nil, errors.ErrProfileNotVisible
	}

	// Get user photos (only verified photos for other users)
	photos, err := uc.photoRepo.GetUserPhotos(ctx, targetUserID, true)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user photos")
	}

	// Check if users have matched
	isMatch, err := uc.matchRepo.MatchExists(ctx, viewerID, targetUserID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to check match status")
	}

	// Check if viewer can message target user
	canMessage := isMatch // Only allow messaging if they've matched

	// Get user statistics (limited for privacy)
	stats, err := uc.userRepo.GetUserStats(ctx, targetUserID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user statistics")
	}

	// Apply privacy filters to location
	location := uc.privacyService.FilterLocation(ctx, viewerID, targetUserID, targetUser)

	// Build response
	response := &ViewUserProfileResponse{
		ID:           targetUser.ID,
		FirstName:     targetUser.FirstName,
		Age:          targetUser.GetAge(),
		Bio:          uc.privacyService.FilterBio(ctx, viewerID, targetUserID, targetUser.Bio),
		Location:      location,
		IsVerified:    targetUser.IsVerified,
		CreatedAt:     targetUser.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsMatch:       isMatch,
		CanMessage:    canMessage,
	}

	// Add photos to response (apply privacy filters)
	for _, photo := range photos {
		if uc.privacyService.CanViewPhoto(ctx, viewerID, targetUserID, photo) {
			responsePhoto := &Photo{
				ID:                photo.ID.String(),
				URL:               photo.FileURL,
				IsPrimary:         photo.IsPrimary,
				VerificationStatus: photo.VerificationStatus,
				CreatedAt:         photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			response.Photos = append(response.Photos, responsePhoto)
		}
	}

	// Add limited statistics to response
	response.ProfileStats = &ProfileStats{
		ProfileViews:    stats.ProfileViews,
		PhotosCount:     stats.PhotosCount,
		ProfileComplete: targetUser.IsComplete(),
		LastActive:     formatLastActive(targetUser.LastActive),
	}

	// Track profile view
	if err := uc.privacyService.TrackProfileView(ctx, viewerID, targetUserID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	// Cache the response with shorter TTL for privacy
	if err := uc.cacheService.SetViewProfile(ctx, cacheKey, response, 5*time.Minute); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return response, nil
}

// generateViewProfileCacheKey generates a cache key for view profile
func generateViewProfileCacheKey(viewerID, targetUserID uuid.UUID) string {
	return "view_profile:" + viewerID.String() + ":" + targetUserID.String()
}