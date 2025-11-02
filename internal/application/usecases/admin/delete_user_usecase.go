package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// DeleteUserUseCase handles deleting user accounts with data cleanup
type DeleteUserUseCase struct {
	userRepo    repositories.UserRepository
	photoRepo   repositories.PhotoRepository
	messageRepo repositories.MessageRepository
	matchRepo   repositories.MatchRepository
	reportRepo  repositories.ReportRepository
}

// NewDeleteUserUseCase creates a new DeleteUserUseCase
func NewDeleteUserUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	messageRepo repositories.MessageRepository,
	matchRepo repositories.MatchRepository,
	reportRepo repositories.ReportRepository,
) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo:    userRepo,
		photoRepo:   photoRepo,
		messageRepo: messageRepo,
		matchRepo:   matchRepo,
		reportRepo:  reportRepo,
	}
}

// DeleteUserRequest represents a request to delete a user
type DeleteUserRequest struct {
	AdminID   uuid.UUID `json:"admin_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	HardDelete bool      `json:"hard_delete"`
	Reason     string    `json:"reason" validate:"required,min=5,max=500"`
}

// DeleteUserResponse represents the response from deleting a user
type DeleteUserResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Deleted     bool      `json:"deleted"`
	HardDeleted bool      `json:"hard_deleted"`
	DeletedAt   time.Time `json:"deleted_at"`
	Reason      string    `json:"reason"`
}

// Execute deletes a user account with proper data cleanup
func (uc *DeleteUserUseCase) Execute(ctx context.Context, req DeleteUserRequest) error {
	logger.Info("DeleteUser use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "hard_delete", req.HardDelete, "reason", req.Reason)

	// Get user to verify existence
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Log the deletion action before performing it
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "delete_user", map[string]interface{}{
		"hard_delete": req.HardDelete,
		"reason":      req.Reason,
	})

	if req.HardDelete {
		// Perform hard delete - permanently remove all user data
		err = uc.performHardDelete(ctx, req.UserID)
	} else {
		// Perform soft delete - mark user as deleted but keep data
		err = uc.performSoftDelete(ctx, req.UserID)
	}

	if err != nil {
		logger.Error("Failed to delete user", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("DeleteUser use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID, "hard_delete", req.HardDelete)
	return nil
}

// performSoftDelete performs a soft delete by marking user as inactive
func (uc *DeleteUserUseCase) performSoftDelete(ctx context.Context, userID uuid.UUID) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Mark user as inactive and banned
	user.IsActive = false
	user.IsBanned = true

	// Update user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// performHardDelete performs a hard delete by removing all user data
func (uc *DeleteUserUseCase) performHardDelete(ctx context.Context, userID uuid.UUID) error {
	// Delete user's photos
	err := uc.photoRepo.DeleteUserPhotos(ctx, userID)
	if err != nil {
		logger.Error("Failed to delete user photos", err, "user_id", userID)
		// Continue with other deletions even if photo deletion fails
	}

	// Delete user's messages
	err = uc.messageRepo.DeleteUserMessages(ctx, userID)
	if err != nil {
		logger.Error("Failed to delete user messages", err, "user_id", userID)
		// Continue with other deletions even if message deletion fails
	}

	// Delete user's matches (both sides)
	err = uc.matchRepo.DeleteUserMatches(ctx, userID)
	if err != nil {
		logger.Error("Failed to delete user matches", err, "user_id", userID)
		// Continue with other deletions even if match deletion fails
	}

	// Delete user's reports (both as reporter and reported)
	err = uc.reportRepo.DeleteUserReports(ctx, userID)
	if err != nil {
		logger.Error("Failed to delete user reports", err, "user_id", userID)
		// Continue with other deletions even if report deletion fails
	}

	// Finally, delete the user
	err = uc.userRepo.Delete(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// logAdminAction logs an admin action for audit purposes
func (uc *DeleteUserUseCase) logAdminAction(ctx context.Context, adminID, userID uuid.UUID, action string, metadata map[string]interface{}) {
	// In a real implementation, this would log to an audit table or service
	logger.Info("Admin action logged", 
		"admin_id", adminID,
		"user_id", userID,
		"action", action,
		"metadata", metadata,
		"timestamp", time.Now(),
	)
}