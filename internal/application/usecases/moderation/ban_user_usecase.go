package moderation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// BanUserRequest represents a request to ban a user
type BanUserRequest struct {
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	BannerID      uuid.UUID `json:"banner_id" validate:"required"`
	Reason        string    `json:"reason" validate:"required"`
	Duration      *string   `json:"duration,omitempty"` // "permanent", "7d", "30d", etc.
	Notes         *string   `json:"notes,omitempty"`
	Evidence      *string   `json:"evidence,omitempty"`
	NotifyUser    bool      `json:"notify_user,omitempty"`
}

// BanUserResponse represents the response after banning a user
type BanUserResponse struct {
	UserID     uuid.UUID `json:"user_id"`
	BanID      uuid.UUID `json:"ban_id"`
	Status      string    `json:"status"`
	Duration   *string   `json:"duration,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// BanRecord represents a user ban record
type BanRecord struct {
	ID          uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	BannerID     uuid.UUID  `json:"banner_id"`
	Reason       string     `json:"reason"`
	Duration     *string    `json:"duration,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	Evidence     *string    `json:"evidence,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AppealRequest represents an appeal request
type AppealRequest struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	OriginalBanID   uuid.UUID              `json:"original_ban_id"`
	Reason          string                 `json:"reason"`
	Description     string                 `json:"description"`
	Evidence        map[string]interface{} `json:"evidence,omitempty"`
	Status          string                 `json:"status"` // "pending", "reviewed", "approved", "rejected"
	ReviewedBy      *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time            `json:"reviewed_at,omitempty"`
	ReviewNotes     *string                `json:"review_notes,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// BanUserUseCase handles user banning functionality
type BanUserUseCase struct {
	userRepo           repositories.UserRepository
	adminUserRepo      repositories.AdminUserRepository
	banRepo           BanRepository
	appealRepo        AppealRepository
	validator          validator.Validator
	notificationService NotificationService
}

// BanRepository defines interface for ban operations
type BanRepository interface {
	Create(ctx context.Context, ban *BanRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*BanRecord, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, isActive bool) ([]*BanRecord, error)
	Update(ctx context.Context, ban *BanRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveBan(ctx context.Context, userID uuid.UUID) (*BanRecord, error)
	GetBanCount(ctx context.Context, userID uuid.UUID) (int64, error)
	ExistsActiveBan(ctx context.Context, userID uuid.UUID) (bool, error)
}

// AppealRepository defines interface for appeal operations
type AppealRepository interface {
	Create(ctx context.Context, appeal *AppealRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*AppealRequest, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AppealRequest, error)
	Update(ctx context.Context, appeal *AppealRequest) error
	GetPendingAppeals(ctx context.Context, limit, offset int) ([]*AppealRequest, error)
}

// NewBanUserUseCase creates a new BanUserUseCase
func NewBanUserUseCase(
	userRepo repositories.UserRepository,
	adminUserRepo repositories.AdminUserRepository,
	banRepo BanRepository,
	appealRepo AppealRepository,
	validator validator.Validator,
	notificationService NotificationService,
) *BanUserUseCase {
	return &BanUserUseCase{
		userRepo:           userRepo,
		adminUserRepo:      adminUserRepo,
		banRepo:           banRepo,
		appealRepo:        appealRepo,
		validator:          validator,
		notificationService: notificationService,
	}
}

// Execute executes the ban user use case
func (uc *BanUserUseCase) Execute(ctx context.Context, req BanUserRequest) (*BanUserResponse, error) {
	logger.Info("Executing BanUser use case", "user_id", req.UserID, "banner_id", req.BannerID, "reason", req.Reason)
	
	// Validate request
	if err := uc.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req)
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Additional business validation
	if err := uc.validateBusinessRules(ctx, req); err != nil {
		logger.Error("Business validation failed", err, "request", req)
		return nil, fmt.Errorf("business validation failed: %w", err)
	}
	
	// Verify users exist
	if err := uc.verifyUsers(ctx, req); err != nil {
		logger.Error("User verification failed", err, "request", req)
		return nil, fmt.Errorf("user verification failed: %w", err)
	}
	
	// Check if user is already banned
	hasActiveBan, err := uc.banRepo.ExistsActiveBan(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to check active ban", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to check active ban: %w", err)
	}
	if hasActiveBan {
		return nil, fmt.Errorf("user is already banned")
	}
	
	// Create ban record
	ban := &BanRecord{
		ID:       uuid.New(),
		UserID:   req.UserID,
		BannerID: req.BannerID,
		Reason:   req.Reason,
		Duration: req.Duration,
		Notes:    req.Notes,
		Evidence: req.Evidence,
		IsActive: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Calculate expiry time if duration is specified
	if req.Duration != nil {
		duration, err := uc.parseDuration(*req.Duration)
		if err == nil && duration > 0 {
			expiresAt := time.Now().Add(duration)
			ban.ExpiresAt = &expiresAt
		}
	}
	
	// Save ban to database
	if err := uc.banRepo.Create(ctx, ban); err != nil {
		logger.Error("Failed to create ban", err, "ban_id", ban.ID)
		return nil, fmt.Errorf("failed to create ban: %w", err)
	}
	
	// Apply ban to user
	if err := uc.applyBanToUser(ctx, req.UserID); err != nil {
		logger.Error("Failed to apply ban to user", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to apply ban to user: %w", err)
	}
	
	// Send notifications
	if err := uc.sendNotifications(ctx, req, ban); err != nil {
		logger.Error("Failed to send notifications", err, "ban_id", ban.ID)
		// Don't fail the operation, just log the error
	}
	
	response := &BanUserResponse{
		UserID:    req.UserID,
		BanID:     ban.ID,
		Status:     "active",
		Duration:   req.Duration,
		ExpiresAt:  ban.ExpiresAt,
		Message:    "User banned successfully",
		CreatedAt:  ban.CreatedAt,
	}
	
	logger.Info("BanUser use case executed successfully", "ban_id", ban.ID, "user_id", req.UserID)
	return response, nil
}

// validateBusinessRules validates business rules for banning
func (uc *BanUserUseCase) validateBusinessRules(ctx context.Context, req BanUserRequest) error {
	// Check if banner has permission to ban users
	banner, err := uc.adminUserRepo.GetByID(ctx, req.BannerID)
	if err != nil {
		return fmt.Errorf("banner not found: %w", err)
	}
	if !banner.CanBanUsers() {
		return fmt.Errorf("banner does not have permission to ban users")
	}
	
	// Cannot ban yourself
	if req.UserID == req.BannerID {
		return fmt.Errorf("cannot ban yourself")
	}
	
	// Check if banner is trying to ban a super admin
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// In a real implementation, you might check if user is a super admin
	// and prevent regular admins from banning them
	
	return nil
}

// verifyUsers verifies that both banner and banned users exist and are active
func (uc *BanUserUseCase) verifyUsers(ctx context.Context, req BanUserRequest) error {
	// Verify banner exists and is active
	banner, err := uc.adminUserRepo.GetByID(ctx, req.BannerID)
	if err != nil {
		return fmt.Errorf("banner not found: %w", err)
	}
	if !banner.IsActive {
		return fmt.Errorf("banner account is not active")
	}
	
	// Verify user exists
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Allow banning banned users (this might be for extending a ban)
	
	return nil
}

// applyBanToUser applies the ban to the user account
func (uc *BanUserUseCase) applyBanToUser(ctx context.Context, userID uuid.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	user.IsBanned = true
	user.IsActive = false
	
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// parseDuration parses duration string into time.Duration
func (uc *BanUserUseCase) parseDuration(durationStr string) (time.Duration, error) {
	switch durationStr {
	case "permanent":
		return 0, nil // 0 duration means permanent
	case "1h":
		return 1 * time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	case "90d":
		return 90 * 24 * time.Hour, nil
	case "1y":
		return 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration: %s", durationStr)
	}
}

// sendNotifications sends notifications related to the ban
func (uc *BanUserUseCase) sendNotifications(ctx context.Context, req BanUserRequest, ban *BanRecord) error {
	// Notify banned user if requested
	if req.NotifyUser {
		userData := map[string]interface{}{
			"ban_id":    ban.ID,
			"reason":     req.Reason,
			"duration":   req.Duration,
			"expires_at": ban.ExpiresAt,
			"notes":      req.Notes,
		}
		
		if err := uc.notificationService.SendNotification(ctx, req.UserID, "user_banned", userData); err != nil {
			return fmt.Errorf("failed to notify banned user: %w", err)
		}
	}
	
	// Notify other admins about the ban
	adminData := map[string]interface{}{
		"ban_id":    ban.ID,
		"user_id":   req.UserID,
		"banner_id": req.BannerID,
		"reason":     req.Reason,
		"duration":   req.Duration,
		"expires_at": ban.ExpiresAt,
	}
	
	if err := uc.notificationService.SendAdminNotification(ctx, uuid.Nil, "user_banned", adminData); err != nil {
		return fmt.Errorf("failed to notify admins about ban: %w", err)
	}
	
	return nil
}

// SubmitAppeal submits an appeal for a ban
func (uc *BanUserUseCase) SubmitAppeal(ctx context.Context, appeal *AppealRequest) error {
	logger.Info("Submitting ban appeal", "appeal_id", appeal.ID, "user_id", appeal.UserID, "ban_id", appeal.OriginalBanID)
	
	// Validate appeal
	if err := uc.validateAppeal(ctx, appeal); err != nil {
		return fmt.Errorf("appeal validation failed: %w", err)
	}
	
	// Check if user has an active ban
	activeBan, err := uc.banRepo.GetActiveBan(ctx, appeal.UserID)
	if err != nil {
		return fmt.Errorf("failed to get active ban: %w", err)
	}
	if activeBan == nil {
		return fmt.Errorf("user does not have an active ban to appeal")
	}
	
	// Create appeal record
	if err := uc.appealRepo.Create(ctx, appeal); err != nil {
		logger.Error("Failed to create appeal", err, "appeal_id", appeal.ID)
		return fmt.Errorf("failed to create appeal: %w", err)
	}
	
	// Send notifications
	if err := uc.sendAppealNotifications(ctx, appeal, activeBan); err != nil {
		logger.Error("Failed to send appeal notifications", err, "appeal_id", appeal.ID)
		// Don't fail the operation, just log the error
	}
	
	logger.Info("Ban appeal submitted successfully", "appeal_id", appeal.ID)
	return nil
}

// validateAppeal validates an appeal request
func (uc *BanUserUseCase) validateAppeal(ctx context.Context, appeal *AppealRequest) error {
	// Check if user exists
	_, err := uc.userRepo.GetByID(ctx, appeal.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Check if original ban exists
	_, err = uc.banRepo.GetByID(ctx, appeal.OriginalBanID)
	if err != nil {
		return fmt.Errorf("original ban not found: %w", err)
	}
	
	// Check if user already has a pending appeal
	pendingAppeals, err := uc.appealRepo.GetByUserID(ctx, appeal.UserID, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to check for pending appeals: %w", err)
	}
	
	for _, pendingAppeal := range pendingAppeals {
		if pendingAppeal.Status == "pending" {
			return fmt.Errorf("user already has a pending appeal")
		}
	}
	
	return nil
}

// sendAppealNotifications sends notifications related to the appeal
func (uc *BanUserUseCase) sendAppealNotifications(ctx context.Context, appeal *AppealRequest, ban *BanRecord) error {
	// Notify admins about the appeal
	adminData := map[string]interface{}{
		"appeal_id":      appeal.ID,
		"user_id":        appeal.UserID,
		"original_ban_id": appeal.OriginalBanID,
		"reason":         appeal.Reason,
		"description":    appeal.Description,
	}
	
	if err := uc.notificationService.SendAdminNotification(ctx, uuid.Nil, "ban_appeal", adminData); err != nil {
		return fmt.Errorf("failed to notify admins about appeal: %w", err)
	}
	
	return nil
}

// ReviewAppeal reviews an appeal request
func (uc *BanUserUseCase) ReviewAppeal(ctx context.Context, appealID, reviewerID uuid.UUID, approved bool, notes string) error {
	logger.Info("Reviewing ban appeal", "appeal_id", appealID, "reviewer_id", reviewerID, "approved", approved)
	
	// Get appeal
	appeal, err := uc.appealRepo.GetByID(ctx, appealID)
	if err != nil {
		logger.Error("Failed to get appeal", err, "appeal_id", appealID)
		return fmt.Errorf("failed to get appeal: %w", err)
	}
	
	if appeal.Status != "pending" {
		return fmt.Errorf("appeal is not pending review")
	}
	
	// Get reviewer
	reviewer, err := uc.adminUserRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	if !reviewer.CanBanUsers() {
		return fmt.Errorf("reviewer does not have permission to review appeals")
	}
	
	// Update appeal status
	now := time.Now()
	appeal.Status = "reviewed"
	appeal.ReviewedBy = &reviewerID
	appeal.ReviewedAt = &now
	appeal.ReviewNotes = &notes
	
	if approved {
		appeal.Status = "approved"
		// Lift the ban
		if err := uc.liftBan(ctx, appeal.OriginalBanID, reviewerID); err != nil {
			return fmt.Errorf("failed to lift ban: %w", err)
		}
	} else {
		appeal.Status = "rejected"
	}
	
	if err := uc.appealRepo.Update(ctx, appeal); err != nil {
		return fmt.Errorf("failed to update appeal: %w", err)
	}
	
	// Send notifications
	if err := uc.sendAppealDecisionNotifications(ctx, appeal, approved, notes, reviewer); err != nil {
		logger.Error("Failed to send appeal decision notifications", err, "appeal_id", appealID)
		// Don't fail the operation, just log the error
	}
	
	logger.Info("Ban appeal reviewed successfully", "appeal_id", appealID, "approved", approved)
	return nil
}

// liftBan lifts a ban
func (uc *BanUserUseCase) liftBan(ctx context.Context, banID, reviewerID uuid.UUID) error {
	// Get ban record
	ban, err := uc.banRepo.GetByID(ctx, banID)
	if err != nil {
		return fmt.Errorf("failed to get ban: %w", err)
	}
	
	if !ban.IsActive {
		return fmt.Errorf("ban is not active")
	}
	
	// Deactivate ban
	ban.IsActive = false
	ban.UpdatedAt = time.Now()
	
	if err := uc.banRepo.Update(ctx, ban); err != nil {
		return fmt.Errorf("failed to update ban: %w", err)
	}
	
	// Reactivate user
	if err := uc.reactivateUser(ctx, ban.UserID); err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}
	
	return nil
}

// reactivateUser reactivates a banned user
func (uc *BanUserUseCase) reactivateUser(ctx context.Context, userID uuid.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	user.IsBanned = false
	user.IsActive = true
	
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// sendAppealDecisionNotifications sends notifications about appeal decision
func (uc *BanUserUseCase) sendAppealDecisionNotifications(ctx context.Context, appeal *AppealRequest, approved bool, notes string, reviewer *entities.AdminUser) error {
	// Notify user about the decision
	userData := map[string]interface{}{
		"appeal_id":   appeal.ID,
		"approved":     approved,
		"notes":        notes,
		"reviewed_by":  reviewer.Email,
	}
	
	if err := uc.notificationService.SendNotification(ctx, appeal.UserID, "appeal_decision", userData); err != nil {
		return fmt.Errorf("failed to notify user about appeal decision: %w", err)
	}
	
	// Notify other admins about the decision
	adminData := map[string]interface{}{
		"appeal_id":   appeal.ID,
		"user_id":     appeal.UserID,
		"approved":     approved,
		"notes":        notes,
		"reviewed_by":  reviewer.Email,
	}
	
	if err := uc.notificationService.SendAdminNotification(ctx, uuid.Nil, "appeal_decision", adminData); err != nil {
		return fmt.Errorf("failed to notify admins about appeal decision: %w", err)
	}
	
	return nil
}

// GetUserBanHistory gets the ban history for a user
func (uc *BanUserUseCase) GetUserBanHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*BanRecord, error) {
	logger.Info("Getting user ban history", "user_id", userID, "limit", limit, "offset", offset)
	
	bans, err := uc.banRepo.GetByUserID(ctx, userID, false) // false to get all bans, not just active
	if err != nil {
		logger.Error("Failed to get user ban history", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user ban history: %w", err)
	}
	
	return bans, nil
}

// GetUserAppealHistory gets the appeal history for a user
func (uc *BanUserUseCase) GetUserAppealHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AppealRequest, error) {
	logger.Info("Getting user appeal history", "user_id", userID, "limit", limit, "offset", offset)
	
	appeals, err := uc.appealRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user appeal history", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user appeal history: %w", err)
	}
	
	return appeals, nil
}

// GetPendingAppeals gets all pending appeals for review
func (uc *BanUserUseCase) GetPendingAppeals(ctx context.Context, reviewerID uuid.UUID, limit, offset int) ([]*AppealRequest, error) {
	logger.Info("Getting pending appeals", "reviewer_id", reviewerID, "limit", limit, "offset", offset)
	
	// Check reviewer permissions
	reviewer, err := uc.adminUserRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	if !reviewer.CanBanUsers() {
		return nil, fmt.Errorf("reviewer does not have permission to review appeals")
	}
	
	appeals, err := uc.appealRepo.GetPendingAppeals(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get pending appeals", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to get pending appeals: %w", err)
	}
	
	return appeals, nil
}