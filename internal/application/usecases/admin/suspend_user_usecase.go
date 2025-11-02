package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SuspendUserUseCase handles suspending and banning users
type SuspendUserUseCase struct {
	userRepo repositories.UserRepository
}

// NewSuspendUserUseCase creates a new SuspendUserUseCase
func NewSuspendUserUseCase(userRepo repositories.UserRepository) *SuspendUserUseCase {
	return &SuspendUserUseCase{
		userRepo: userRepo,
	}
}

// SuspendUserRequest represents a request to suspend a user
type SuspendUserRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Duration string    `json:"duration" validate:"required,oneof=1d 3d 7d 14d 30d permanent"`
	Reason  string    `json:"reason" validate:"required,min=5,max=500"`
}

// UnsuspendUserRequest represents a request to unsuspend a user
type UnsuspendUserRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Reason  string    `json:"reason" validate:"required,min=5,max=500"`
}

// BanUserRequest represents a request to ban a user
type BanUserRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Reason  string    `json:"reason" validate:"required,min=5,max=500"`
}

// UnbanUserRequest represents a request to unban a user
type UnbanUserRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Reason  string    `json:"reason" validate:"required,min=5,max=500"`
}

// SuspendUserResponse represents the response from suspending a user
type SuspendUserResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Suspended   bool      `json:"suspended"`
	Duration    string    `json:"duration"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Reason      string    `json:"reason"`
	SuspendedAt time.Time `json:"suspended_at"`
}

// UnsuspendUserResponse represents the response from unsuspending a user
type UnsuspendUserResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Unsuspended  bool      `json:"unsuspended"`
	Reason       string    `json:"reason"`
	UnsuspendedAt time.Time `json:"unsuspended_at"`
}

// BanUserResponse represents the response from banning a user
type BanUserResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Banned     bool      `json:"banned"`
	Reason     string    `json:"reason"`
	BannedAt   time.Time `json:"banned_at"`
}

// UnbanUserResponse represents the response from unbanning a user
type UnbanUserResponse struct {
	UserID     uuid.UUID `json:"user_id"`
	Unbanned   bool      `json:"unbanned"`
	Reason      string    `json:"reason"`
	UnbannedAt time.Time `json:"unbanned_at"`
}

// Execute suspends a user account
func (uc *SuspendUserUseCase) Execute(ctx context.Context, req SuspendUserRequest) (*SuspendUserResponse, error) {
	logger.Info("SuspendUser use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "duration", req.Duration, "reason", req.Reason)

	// Get user to verify existence
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Calculate suspension expiry time
	var expiresAt *time.Time
	if req.Duration != "permanent" {
		duration, err := time.ParseDuration(req.Duration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		expiry := time.Now().Add(duration)
		expiresAt = &expiry
	}

	// Mark user as suspended (inactive but not banned)
	user.IsActive = false
	user.IsBanned = false

	// Update user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		logger.Error("Failed to update user", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the suspension action
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "suspend_user", map[string]interface{}{
		"duration":   req.Duration,
		"expires_at": expiresAt,
		"reason":     req.Reason,
	})

	logger.Info("SuspendUser use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID, "duration", req.Duration)
	return &SuspendUserResponse{
		UserID:      req.UserID,
		Suspended:   true,
		Duration:    req.Duration,
		ExpiresAt:   expiresAt,
		Reason:      req.Reason,
		SuspendedAt: time.Now(),
	}, nil
}

// Unsuspend unsuspends a user account
func (uc *SuspendUserUseCase) Unsuspend(ctx context.Context, req UnsuspendUserRequest) (*UnsuspendUserResponse, error) {
	logger.Info("Unsuspend use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "reason", req.Reason)

	// Get user to verify existence
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark user as active
	user.IsActive = true

	// Update user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		logger.Error("Failed to update user", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the unsuspension action
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "unsuspend_user", map[string]interface{}{
		"reason": req.Reason,
	})

	logger.Info("Unsuspend use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID)
	return &UnsuspendUserResponse{
		UserID:       req.UserID,
		Unsuspended:  true,
		Reason:       req.Reason,
		UnsuspendedAt: time.Now(),
	}, nil
}

// Ban bans a user account permanently
func (uc *SuspendUserUseCase) Ban(ctx context.Context, req BanUserRequest) (*BanUserResponse, error) {
	logger.Info("Ban use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "reason", req.Reason)

	// Get user to verify existence
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark user as banned
	user.IsActive = false
	user.IsBanned = true

	// Update user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		logger.Error("Failed to update user", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the ban action
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "ban_user", map[string]interface{}{
		"reason": req.Reason,
	})

	logger.Info("Ban use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID)
	return &BanUserResponse{
		UserID:  req.UserID,
		Banned:   true,
		Reason:   req.Reason,
		BannedAt: time.Now(),
	}, nil
}

// Unban unbans a user account
func (uc *SuspendUserUseCase) Unban(ctx context.Context, req UnbanUserRequest) (*UnbanUserResponse, error) {
	logger.Info("Unban use case executed", "admin_id", req.AdminID, "user_id", req.UserID, "reason", req.Reason)

	// Get user to verify existence
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Mark user as active and not banned
	user.IsActive = true
	user.IsBanned = false

	// Update user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		logger.Error("Failed to update user", err, "admin_id", req.AdminID, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the unban action
	uc.logAdminAction(ctx, req.AdminID, req.UserID, "unban_user", map[string]interface{}{
		"reason": req.Reason,
	})

	logger.Info("Unban use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID)
	return &UnbanUserResponse{
		UserID:     req.UserID,
		Unbanned:   true,
		Reason:      req.Reason,
		UnbannedAt: time.Now(),
	}, nil
}

// logAdminAction logs an admin action for audit purposes
func (uc *SuspendUserUseCase) logAdminAction(ctx context.Context, adminID, userID uuid.UUID, action string, metadata map[string]interface{}) {
	// In a real implementation, this would log to an audit table or service
	logger.Info("Admin action logged", 
		"admin_id", adminID,
		"user_id", userID,
		"action", action,
		"metadata", metadata,
		"timestamp", time.Now(),
	)
}