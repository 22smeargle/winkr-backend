package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// ProfilePrivacyService handles profile privacy and security
type ProfilePrivacyService interface {
	CanViewProfile(ctx context.Context, viewerID, targetUserID uuid.UUID) bool
	FilterLocation(ctx context.Context, viewerID, targetUserID uuid.UUID, targetUser *repositories.UserEntity) *profile.Location
	FilterBio(ctx context.Context, viewerID, targetUserID uuid.UUID, bio *string) *string
	CanViewPhoto(ctx context.Context, viewerID, targetUserID uuid.UUID, photo *repositories.PhotoEntity) bool
	TrackProfileView(ctx context.Context, viewerID, targetUserID uuid.UUID) error
	IsUserBlocked(ctx context.Context, blockerID, blockedUserID uuid.UUID) (bool, error)
	BlockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repositories.UserEntity, error)
}

// RedisProfilePrivacyService implements ProfilePrivacyService
type RedisProfilePrivacyService struct {
	userRepo    repositories.UserRepository
	matchRepo   repositories.MatchRepository
	reportRepo  repositories.ReportRepository
	cacheService ProfileCacheService
}

// NewRedisProfilePrivacyService creates a new RedisProfilePrivacyService instance
func NewRedisProfilePrivacyService(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	reportRepo repositories.ReportRepository,
	cacheService ProfileCacheService,
) *RedisProfilePrivacyService {
	return &RedisProfilePrivacyService{
		userRepo:    userRepo,
		matchRepo:   matchRepo,
		reportRepo:  reportRepo,
		cacheService: cacheService,
	}
}

// CanViewProfile checks if a user can view another user's profile
func (s *RedisProfilePrivacyService) CanViewProfile(ctx context.Context, viewerID, targetUserID uuid.UUID) bool {
	// Check if viewer is blocked by target user
	isBlocked, err := s.IsUserBlocked(ctx, targetUserID, viewerID)
	if err != nil || isBlocked {
		return false
	}

	// Check if target user wants to be discovered
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return false
	}

	// Check if target user is active and not banned
	if !targetUser.IsActive || targetUser.IsBanned {
		return false
	}

	// Check user preferences
	preferences, err := s.userRepo.GetPreferences(ctx, targetUserID)
	if err == nil && !preferences.ShowMe {
		return false
	}

	// Check if users have matched (only matched users can see full profiles)
	hasMatched, err := s.matchRepo.MatchExists(ctx, viewerID, targetUserID)
	if err != nil {
		return false
	}

	// Allow viewing if they've matched or if target user is discoverable
	return hasMatched || (preferences != nil && preferences.ShowMe)
}

// FilterLocation filters location data based on privacy settings
func (s *RedisProfilePrivacyService) FilterLocation(ctx context.Context, viewerID, targetUserID uuid.UUID, targetUser *repositories.UserEntity) *profile.Location {
	// Check if users have matched
	hasMatched, err := s.matchRepo.MatchExists(ctx, viewerID, targetUserID)
	if err != nil {
		return nil
	}

	// If not matched, return approximate location or no location
	if !hasMatched {
		// Return city and country only, hide exact coordinates
		if targetUser.LocationCity != nil && targetUser.LocationCountry != nil {
			return &profile.Location{
				City:    targetUser.LocationCity,
				Country: targetUser.LocationCountry,
			}
		}
		return nil
	}

	// If matched, return full location
	if targetUser.HasLocation() {
		return &profile.Location{
			Lat:     targetUser.LocationLat,
			Lng:     targetUser.LocationLng,
			City:    targetUser.LocationCity,
			Country: targetUser.LocationCountry,
		}
	}

	return nil
}

// FilterBio filters bio based on privacy settings
func (s *RedisProfilePrivacyService) FilterBio(ctx context.Context, viewerID, targetUserID uuid.UUID, bio *string) *string {
	if bio == nil {
		return nil
	}

	// Check if users have matched
	hasMatched, err := s.matchRepo.MatchExists(ctx, viewerID, targetUserID)
	if err != nil {
		return nil
	}

	// If not matched, return truncated bio
	if !hasMatched {
		bioStr := *bio
		if len(bioStr) > 100 {
			truncated := bioStr[:97] + "..."
			return &truncated
		}
		return bio
	}

	// If matched, return full bio
	return bio
}

// CanViewPhoto checks if a user can view a photo
func (s *RedisProfilePrivacyService) CanViewPhoto(ctx context.Context, viewerID, targetUserID uuid.UUID, photo *repositories.PhotoEntity) bool {
	// Only verified photos can be viewed by non-matched users
	if photo.VerificationStatus != "approved" {
		// Check if users have matched
		hasMatched, err := s.matchRepo.MatchExists(ctx, viewerID, targetUserID)
		if err != nil {
			return false
		}
		return hasMatched
	}

	return true
}

// TrackProfileView tracks when a user views another user's profile
func (s *RedisProfilePrivacyService) TrackProfileView(ctx context.Context, viewerID, targetUserID uuid.UUID) error {
	// Don't track self views
	if viewerID == targetUserID {
		return nil
	}

	// Check if viewer is blocked
	isBlocked, err := s.IsUserBlocked(ctx, targetUserID, viewerID)
	if err != nil || isBlocked {
		return nil
	}

	// Track the view in cache for rate limiting
	cacheKey := "profile_view:" + viewerID.String() + ":" + targetUserID.String()
	
	// Check if already viewed recently
	exists, err := s.cacheService.Exists(ctx, cacheKey)
	if err != nil {
		return err
	}
	
	if exists {
		return nil // Already viewed recently
	}
	
	// Set cache entry with TTL (e.g., 1 hour)
	if err := s.cacheService.Set(ctx, cacheKey, "1", time.Hour); err != nil {
		return err
	}
	
	// In a real implementation, you would also:
	// 1. Update profile view count in database
	// 2. Send notification to target user (if they have notifications enabled)
	// 3. Log the view for analytics
	
	return nil
}

// IsUserBlocked checks if a user has blocked another user
func (s *RedisProfilePrivacyService) IsUserBlocked(ctx context.Context, blockerID, blockedUserID uuid.UUID) (bool, error) {
	return s.userRepo.IsBlocked(ctx, blockerID, blockedUserID)
}

// BlockUser blocks a user
func (s *RedisProfilePrivacyService) BlockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error {
	// Check if user is already blocked
	isBlocked, err := s.IsUserBlocked(ctx, blockerID, blockedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check block status")
	}
	
	if isBlocked {
		return errors.ErrUserAlreadyBlocked
	}
	
	// Create block
	if err := s.userRepo.BlockUser(ctx, blockerID, blockedUserID); err != nil {
		return errors.WrapError(err, "Failed to block user")
	}
	
	// Invalidate relevant caches
	if err := s.cacheService.InvalidateUserCache(ctx, blockedUserID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}
	
	return nil
}

// UnblockUser unblocks a user
func (s *RedisProfilePrivacyService) UnblockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error {
	// Check if user is blocked
	isBlocked, err := s.IsUserBlocked(ctx, blockerID, blockedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check block status")
	}
	
	if !isBlocked {
		return errors.ErrUserNotBlocked
	}
	
	// Remove block
	if err := s.userRepo.UnblockUser(ctx, blockerID, blockedUserID); err != nil {
		return errors.WrapError(err, "Failed to unblock user")
	}
	
	// Invalidate relevant caches
	if err := s.cacheService.InvalidateUserCache(ctx, blockedUserID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}
	
	return nil
}

// GetBlockedUsers gets list of blocked users
func (s *RedisProfilePrivacyService) GetBlockedUsers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repositories.UserEntity, error) {
	return s.userRepo.GetBlockedUsers(ctx, userID, limit, offset)
}

// ReportUser reports a user for inappropriate behavior
func (s *RedisProfilePrivacyService) ReportUser(ctx context.Context, reporterID, reportedUserID uuid.UUID, reason, description string) error {
	// Check if reporter has already reported this user
	exists, err := s.reportRepo.ReportExists(ctx, reporterID, reportedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check report existence")
	}
	
	if exists {
		return errors.ErrReportAlreadyExists
	}
	
	// Create report
	report := &repositories.ReportEntity{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		Reason:        reason,
		Description:    description,
		Status:        "pending",
	}
	
	if err := s.reportRepo.CreateReport(ctx, report); err != nil {
		return errors.WrapError(err, "Failed to create report")
	}
	
	return nil
}

// GetPrivacySettings gets user's privacy settings
func (s *RedisProfilePrivacyService) GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*PrivacySettings, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	
	preferences, err := s.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user preferences")
	}
	
	settings := &PrivacySettings{
		ShowProfile: true,
		ShowLocation: true,
		ShowAge:     true,
		ShowDistance: true,
	}
	
	if preferences != nil {
		settings.ShowProfile = preferences.ShowMe
	}
	
	// Additional privacy settings could be stored in a separate privacy_settings table
	// For now, we'll use basic preferences
	
	return settings, nil
}

// UpdatePrivacySettings updates user's privacy settings
func (s *RedisProfilePrivacyService) UpdatePrivacySettings(ctx context.Context, userID uuid.UUID, settings *PrivacySettings) error {
	preferences, err := s.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		// Create preferences if they don't exist
		preferences = &repositories.UserPreferencesEntity{
			UserID: userID,
		}
	}
	
	// Update privacy settings
	preferences.ShowMe = settings.ShowProfile
	
	// Save preferences
	if preferences.ID == uuid.Nil {
		if err := s.userRepo.CreatePreferences(ctx, preferences); err != nil {
			return errors.WrapError(err, "Failed to create user preferences")
		}
	} else {
		if err := s.userRepo.UpdatePreferences(ctx, preferences); err != nil {
			return errors.WrapError(err, "Failed to update user preferences")
		}
	}
	
	// Invalidate cache
	if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}
	
	return nil
}

// PrivacySettings represents user privacy settings
type PrivacySettings struct {
	ShowProfile  bool `json:"show_profile"`
	ShowLocation bool `json:"show_location"`
	ShowAge     bool `json:"show_age"`
	ShowDistance bool `json:"show_distance"`
}

// ProfileCacheService defines the interface for cache operations
type ProfileCacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
}