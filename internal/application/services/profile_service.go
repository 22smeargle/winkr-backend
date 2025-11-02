package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// ProfileService handles profile business logic
type ProfileService struct {
	userRepo    repositories.UserRepository
	photoRepo   repositories.PhotoRepository
	matchRepo   repositories.MatchRepository
	reportRepo  repositories.ReportRepository
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	matchRepo repositories.MatchRepository,
	reportRepo repositories.ReportRepository,
) *ProfileService {
	return &ProfileService{
		userRepo:   userRepo,
		photoRepo:  photoRepo,
		matchRepo:  matchRepo,
		reportRepo: reportRepo,
	}
}

// ValidateProfileUpdate validates profile update request
func (s *ProfileService) ValidateProfileUpdate(user *entities.User, req interface{}) error {
	// Type assertion to get the specific request type
	updateReq, ok := req.(*profile.UpdateProfileRequest)
	if !ok {
		return errors.ErrInvalidRequest
	}

	// Validate first name
	if updateReq.FirstName != nil {
		if err := s.validateFirstName(*updateReq.FirstName); err != nil {
			return err
		}
	}

	// Validate last name
	if updateReq.LastName != nil {
		if err := s.validateLastName(*updateReq.LastName); err != nil {
			return err
		}
	}

	// Validate bio
	if updateReq.Bio != nil {
		if err := s.validateBio(*updateReq.Bio); err != nil {
			return err
		}
	}

	// Validate interested in
	if len(updateReq.InterestedIn) > 0 {
		if err := s.validateInterestedIn(updateReq.InterestedIn); err != nil {
			return err
		}
	}

	// Validate preferences
	if updateReq.Preferences != nil {
		if err := s.validatePreferences(updateReq.Preferences); err != nil {
			return err
		}
	}

	return nil
}

// validateFirstName validates first name
func (s *ProfileService) validateFirstName(firstName string) error {
	firstName = strings.TrimSpace(firstName)
	
	if len(firstName) < 2 {
		return errors.ErrFirstNameTooShort
	}
	
	if len(firstName) > 100 {
		return errors.ErrFirstNameTooLong
	}
	
	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, firstName); !matched {
		return errors.ErrInvalidFirstName
	}
	
	return nil
}

// validateLastName validates last name
func (s *ProfileService) validateLastName(lastName string) error {
	lastName = strings.TrimSpace(lastName)
	
	if len(lastName) < 2 {
		return errors.ErrLastNameTooShort
	}
	
	if len(lastName) > 100 {
		return errors.ErrLastNameTooLong
	}
	
	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, lastName); !matched {
		return errors.ErrInvalidLastName
	}
	
	return nil
}

// validateBio validates bio
func (s *ProfileService) validateBio(bio string) error {
	bio = strings.TrimSpace(bio)
	
	if len(bio) > 500 {
		return errors.ErrBioTooLong
	}
	
	// Check for inappropriate content
	if s.containsInappropriateContent(bio) {
		return errors.ErrInappropriateBio
	}
	
	return nil
}

// validateInterestedIn validates interested in preferences
func (s *ProfileService) validateInterestedIn(interestedIn []string) error {
	if len(interestedIn) == 0 {
		return errors.ErrInterestedInRequired
	}
	
	validGenders := map[string]bool{
		"male":   true,
		"female": true,
		"other":  true,
	}
	
	for _, gender := range interestedIn {
		if !validGenders[strings.ToLower(gender)] {
			return errors.ErrInvalidGender
		}
	}
	
	return nil
}

// validatePreferences validates user preferences
func (s *ProfileService) validatePreferences(prefs *profile.Preferences) error {
	if prefs.AgeMin < 18 {
		return errors.ErrAgeMinTooYoung
	}
	
	if prefs.AgeMax > 100 {
		return errors.ErrAgeMaxTooOld
	}
	
	if prefs.AgeMin >= prefs.AgeMax {
		return errors.ErrInvalidAgeRange
	}
	
	if prefs.MaxDistance < 1 || prefs.MaxDistance > 500 {
		return errors.ErrInvalidMaxDistance
	}
	
	return nil
}

// containsInappropriateContent checks for inappropriate content in text
func (s *ProfileService) containsInappropriateContent(text string) bool {
	// List of inappropriate words (simplified for example)
	inappropriateWords := []string{
		"spam", "scam", "fraud", "illegal", "drugs", "violence",
		// Add more words as needed
	}
	
	textLower := strings.ToLower(text)
	for _, word := range inappropriateWords {
		if strings.Contains(textLower, word) {
			return true
		}
	}
	
	return false
}

// GetProfileCompletion calculates profile completion percentage
func (s *ProfileService) GetProfileCompletion(ctx context.Context, userID uuid.UUID) (*ProfileCompletion, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	
	completion := &ProfileCompletion{
		Percentage: 0,
		CompletedFields: make([]string, 0),
		MissingFields:   make([]string, 0),
	}
	
	totalFields := 7
	completedFields := 0
	
	// Check basic profile fields
	if user.FirstName != "" {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "first_name")
	} else {
		completion.MissingFields = append(completion.MissingFields, "first_name")
	}
	
	if user.LastName != "" {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "last_name")
	} else {
		completion.MissingFields = append(completion.MissingFields, "last_name")
	}
	
	if !user.DateOfBirth.IsZero() {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "date_of_birth")
	} else {
		completion.MissingFields = append(completion.MissingFields, "date_of_birth")
	}
	
	if user.Gender != "" {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "gender")
	} else {
		completion.MissingFields = append(completion.MissingFields, "gender")
	}
	
	if len(user.InterestedIn) > 0 {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "interested_in")
	} else {
		completion.MissingFields = append(completion.MissingFields, "interested_in")
	}
	
	if user.HasLocation() {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "location")
	} else {
		completion.MissingFields = append(completion.MissingFields, "location")
	}
	
	// Check photos
	photos, err := s.photoRepo.GetUserPhotos(ctx, userID, false)
	if err == nil && len(photos) > 0 {
		completedFields++
		completion.CompletedFields = append(completion.CompletedFields, "photos")
	} else {
		completion.MissingFields = append(completion.MissingFields, "photos")
	}
	
	completion.Percentage = int(float64(completedFields) / float64(totalFields) * 100)
	
	return completion, nil
}

// ProfileCompletion represents profile completion information
type ProfileCompletion struct {
	Percentage     int      `json:"percentage"`
	CompletedFields []string  `json:"completed_fields"`
	MissingFields   []string  `json:"missing_fields"`
}

// UpdateLastActive updates user's last active timestamp
func (s *ProfileService) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.UpdateLastActive(ctx, userID)
}

// IsProfileComplete checks if user profile is complete
func (s *ProfileService) IsProfileComplete(ctx context.Context, userID uuid.UUID) (bool, error) {
	completion, err := s.GetProfileCompletion(ctx, userID)
	if err != nil {
		return false, err
	}
	
	return completion.Percentage >= 80, nil
}

// GetProfileStats gets user profile statistics
func (s *ProfileService) GetProfileStats(ctx context.Context, userID uuid.UUID) (*ProfileStats, error) {
	stats, err := s.userRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	profileStats := &ProfileStats{
		ProfileViews:    stats.ProfileViews,
		PhotosCount:     stats.PhotosCount,
		TotalSwipes:     stats.TotalSwipes,
		TotalMatches:    stats.TotalMatches,
		TotalMessages:    stats.TotalMessages,
		LastActiveDays:   stats.LastActiveDays,
		AccountAgeDays:   stats.AccountAgeDays,
	}
	
	return profileStats, nil
}

// ProfileStats represents profile statistics
type ProfileStats struct {
	ProfileViews    int64 `json:"profile_views"`
	PhotosCount     int64 `json:"photos_count"`
	TotalSwipes     int64 `json:"total_swipes"`
	TotalMatches    int64 `json:"total_matches"`
	TotalMessages    int64 `json:"total_messages"`
	LastActiveDays   int   `json:"last_active_days"`
	AccountAgeDays   int   `json:"account_age_days"`
}

// ReportProfile reports a user profile
func (s *ProfileService) ReportProfile(ctx context.Context, reporterID, reportedUserID uuid.UUID, reason, description string) error {
	// Check if reporter has already reported this user
	exists, err := s.reportRepo.ReportExists(ctx, reporterID, reportedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check report existence")
	}
	
	if exists {
		return errors.ErrReportAlreadyExists
	}
	
	// Create report
	report := &entities.Report{
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

// BlockUser blocks a user
func (s *ProfileService) BlockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error {
	// Check if user is already blocked
	exists, err := s.userRepo.IsBlocked(ctx, blockerID, blockedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check block status")
	}
	
	if exists {
		return errors.ErrUserAlreadyBlocked
	}
	
	// Create block
	if err := s.userRepo.BlockUser(ctx, blockerID, blockedUserID); err != nil {
		return errors.WrapError(err, "Failed to block user")
	}
	
	return nil
}

// UnblockUser unblocks a user
func (s *ProfileService) UnblockUser(ctx context.Context, blockerID, blockedUserID uuid.UUID) error {
	// Check if user is blocked
	exists, err := s.userRepo.IsBlocked(ctx, blockerID, blockedUserID)
	if err != nil {
		return errors.WrapError(err, "Failed to check block status")
	}
	
	if !exists {
		return errors.ErrUserNotBlocked
	}
	
	// Remove block
	if err := s.userRepo.UnblockUser(ctx, blockerID, blockedUserID); err != nil {
		return errors.WrapError(err, "Failed to unblock user")
	}
	
	return nil
}

// GetBlockedUsers gets list of blocked users
func (s *ProfileService) GetBlockedUsers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.GetBlockedUsers(ctx, userID, limit, offset)
}