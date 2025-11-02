package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// DeleteAccountUseCase handles deleting user account
type DeleteAccountUseCase struct {
	userRepo       repositories.UserRepository
	photoRepo      repositories.PhotoRepository
	matchRepo      repositories.MatchRepository
	messageRepo    repositories.MessageRepository
	reportRepo     repositories.ReportRepository
	subscriptionRepo repositories.SubscriptionRepository
	cacheService   ProfileCacheService
	authService    AuthService
}

// NewDeleteAccountUseCase creates a new DeleteAccountUseCase instance
func NewDeleteAccountUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	matchRepo repositories.MatchRepository,
	messageRepo repositories.MessageRepository,
	reportRepo repositories.ReportRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	cacheService ProfileCacheService,
	authService AuthService,
) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{
		userRepo:       userRepo,
		photoRepo:      photoRepo,
		matchRepo:      matchRepo,
		messageRepo:    messageRepo,
		reportRepo:     reportRepo,
		subscriptionRepo: subscriptionRepo,
		cacheService:   cacheService,
		authService:    authService,
	}
}

// DeleteAccountRequest represents delete account request
type DeleteAccountRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Password  string    `json:"password"`
	Reason    string    `json:"reason"`
	Confirm   bool      `json:"confirm"`
}

// DeleteAccountResponse represents delete account response
type DeleteAccountResponse struct {
	Message string `json:"message"`
}

// Execute handles the delete account use case
func (uc *DeleteAccountUseCase) Execute(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	// Validate request
	if !req.Confirm {
		return nil, errors.ErrAccountDeletionNotConfirmed
	}

	// Get user to verify they exist and are active
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive || user.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Verify password if provided
	if req.Password != "" {
		if err := uc.authService.VerifyPassword(user.Email, req.Password); err != nil {
			return nil, errors.ErrInvalidPassword
		}
	}

	// Start account deletion process
	// This is a critical operation, so we'll use a transaction-like approach
	// In a real implementation, you'd use database transactions

	// 1. Cancel active subscriptions
	if err := uc.cancelSubscriptions(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to cancel subscriptions")
	}

	// 2. Delete user's photos
	if err := uc.deleteUserPhotos(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to delete user photos")
	}

	// 3. Delete user's messages
	if err := uc.deleteUserMessages(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to delete user messages")
	}

	// 4. Delete user's matches
	if err := uc.deleteUserMatches(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to delete user matches")
	}

	// 5. Delete reports made by user
	if err := uc.deleteUserReports(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to delete user reports")
	}

	// 6. Anonymize user data instead of hard delete for GDPR compliance
	if err := uc.anonymizeUserData(ctx, user, req.Reason); err != nil {
		return nil, errors.WrapError(err, "Failed to anonymize user data")
	}

	// 7. Invalidate all user sessions
	if err := uc.authService.InvalidateAllUserSessions(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to invalidate user sessions")
	}

	// 8. Clear all caches
	if err := uc.clearUserCaches(ctx, req.UserID); err != nil {
		return nil, errors.WrapError(err, "Failed to clear user caches")
	}

	return &DeleteAccountResponse{
		Message: "Account deleted successfully",
	}, nil
}

// cancelSubscriptions cancels all active subscriptions for the user
func (uc *DeleteAccountUseCase) cancelSubscriptions(ctx context.Context, userID uuid.UUID) error {
	// Get active subscriptions
	subscriptions, err := uc.subscriptionRepo.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return err
	}

	// Cancel each subscription
	for _, subscription := range subscriptions {
		if subscription.Status == "active" {
			if err := uc.subscriptionRepo.CancelSubscription(ctx, subscription.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// deleteUserPhotos deletes all photos associated with the user
func (uc *DeleteAccountUseCase) deleteUserPhotos(ctx context.Context, userID uuid.UUID) error {
	photos, err := uc.photoRepo.GetUserPhotos(ctx, userID, false)
	if err != nil {
		return err
	}

	for _, photo := range photos {
		if err := uc.photoRepo.DeletePhoto(ctx, photo.ID); err != nil {
			return err
		}
	}

	return nil
}

// deleteUserMessages deletes all messages sent by the user
func (uc *DeleteAccountUseCase) deleteUserMessages(ctx context.Context, userID uuid.UUID) error {
	// Get all conversations where user participated
	conversations, err := uc.messageRepo.GetUserConversations(ctx, userID)
	if err != nil {
		return err
	}

	// Delete messages in each conversation
	for _, conversation := range conversations {
		if err := uc.messageRepo.DeleteConversationMessages(ctx, conversation.ID, userID); err != nil {
			return err
		}
	}

	return nil
}

// deleteUserMatches deletes all matches associated with the user
func (uc *DeleteAccountUseCase) deleteUserMatches(ctx context.Context, userID uuid.UUID) error {
	matches, err := uc.matchRepo.GetUserMatches(ctx, userID, 0, 0)
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := uc.matchRepo.DeleteMatch(ctx, match.ID); err != nil {
			return err
		}
	}

	return nil
}

// deleteUserReports deletes all reports made by the user
func (uc *DeleteAccountUseCase) deleteUserReports(ctx context.Context, userID uuid.UUID) error {
	reports, err := uc.reportRepo.GetReportsByReporter(ctx, userID)
	if err != nil {
		return err
	}

	for _, report := range reports {
		if err := uc.reportRepo.DeleteReport(ctx, report.ID); err != nil {
			return err
		}
	}

	return nil
}

// anonymizeUserData anonymizes user data for GDPR compliance
func (uc *DeleteAccountUseCase) anonymizeUserData(ctx context.Context, user *repositories.UserEntity, reason string) error {
	// Generate anonymized data
	anonymizedEmail := "deleted_" + user.ID.String() + "@deleted.com"
	anonymizedFirstName := "Deleted"
	anonymizedLastName := "User"
	anonymizedBio := "This account has been deleted"

	// Update user with anonymized data
	user.Email = anonymizedEmail
	user.FirstName = anonymizedFirstName
	user.LastName = anonymizedLastName
	user.Bio = &anonymizedBio
	user.IsActive = false
	user.IsBanned = true // Mark as banned to prevent access
	user.LocationLat = nil
	user.LocationLng = nil
	user.LocationCity = nil
	user.LocationCountry = nil

	// Add deletion timestamp
	now := time.Now()
	user.DeletedAt = &now

	// Update user in database
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Log deletion reason for audit purposes
	// TODO: Add proper audit logging

	return nil
}

// clearUserCaches clears all caches associated with the user
func (uc *DeleteAccountUseCase) clearUserCaches(ctx context.Context, userID uuid.UUID) error {
	// Clear profile cache
	if err := uc.cacheService.DeleteProfile(ctx, userID); err != nil {
		return err
	}

	// Clear location cache
	if err := uc.cacheService.DeleteUserLocation(ctx, userID); err != nil {
		return err
	}

	// Clear matches cache
	if err := uc.cacheService.DeleteUserMatches(ctx, userID); err != nil {
		return err
	}

	return nil
}