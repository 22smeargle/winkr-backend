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

// ReviewReportRequest represents a request to review a report
type ReviewReportRequest struct {
	ReportID    uuid.UUID `json:"report_id" validate:"required"`
	ReviewerID   uuid.UUID `json:"reviewer_id" validate:"required"`
	Action       string    `json:"action" validate:"required,oneof=resolve dismiss escalate"`
	Reason       string    `json:"reason" validate:"required"`
	Notes        *string   `json:"notes,omitempty"`
	TakeAction   bool      `json:"take_action,omitempty"`
	ActionType   *string   `json:"action_type,omitempty"` // "ban", "suspend", "warn"
	ActionDuration *string   `json:"action_duration,omitempty"` // "permanent", "7d", "30d", etc.
}

// ReviewReportResponse represents the response after reviewing a report
type ReviewReportRequest struct {
	ReportID      uuid.UUID `json:"report_id"`
	Status        string    `json:"status"`
	ReviewedBy    uuid.UUID `json:"reviewed_by"`
	ReviewedAt    time.Time `json:"reviewed_at"`
	ActionTaken   bool      `json:"action_taken"`
	Message       string    `json:"message"`
}

// ReviewReportUseCase handles report reviewing functionality
type ReviewReportUseCase struct {
	reportRepo         repositories.ReportRepository
	userRepo           repositories.UserRepository
	adminUserRepo      repositories.AdminUserRepository
	validator          validator.Validator
	moderationService  ModerationService
	notificationService NotificationService
}

// AdminUserRepository defines interface for admin user operations
type AdminUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error)
	Update(ctx context.Context, adminUser *entities.AdminUser) error
}

// ModerationService defines interface for moderation operations
type ModerationService interface {
	ApplyModerationAction(ctx context.Context, userID uuid.UUID, action ModerationAction) error
	UpdateUserReputation(ctx context.Context, userID uuid.UUID, change float64) error
}

// NotificationService defines interface for notification operations
type NotificationService interface {
	SendNotification(ctx context.Context, userID uuid.UUID, notificationType string, data map[string]interface{}) error
	SendAdminNotification(ctx context.Context, adminID uuid.UUID, notificationType string, data map[string]interface{}) error
}

// ModerationAction represents a moderation action
type ModerationAction struct {
	Type        string                 `json:"type"`
	Reason      string                 `json:"reason"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	AppliedBy   uuid.UUID              `json:"applied_by"`
	AppliedAt   time.Time              `json:"applied_at"`
	ExpiresAt   *time.Time            `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewReviewReportUseCase creates a new ReviewReportUseCase
func NewReviewReportUseCase(
	reportRepo repositories.ReportRepository,
	userRepo repositories.UserRepository,
	adminUserRepo AdminUserRepository,
	validator validator.Validator,
	moderationService ModerationService,
	notificationService NotificationService,
) *ReviewReportUseCase {
	return &ReviewReportUseCase{
		reportRepo:         reportRepo,
		userRepo:           userRepo,
		adminUserRepo:      adminUserRepo,
		validator:          validator,
		moderationService:  moderationService,
		notificationService: notificationService,
	}
}

// Execute executes the review report use case
func (uc *ReviewReportUseCase) Execute(ctx context.Context, req ReviewReportRequest) (*ReviewReportResponse, error) {
	logger.Info("Executing ReviewReport use case", "report_id", req.ReportID, "reviewer_id", req.ReviewerID, "action", req.Action)
	
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
	
	// Get report
	report, err := uc.reportRepo.GetByID(ctx, req.ReportID)
	if err != nil {
		logger.Error("Failed to get report", err, "report_id", req.ReportID)
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	
	// Check if report can be reviewed
	if !report.CanBeReviewed() {
		return nil, fmt.Errorf("report cannot be reviewed")
	}
	
	// Get reviewer (admin user)
	reviewer, err := uc.adminUserRepo.GetByID(ctx, req.ReviewerID)
	if err != nil {
		logger.Error("Failed to get reviewer", err, "reviewer_id", req.ReviewerID)
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	// Check if reviewer has permission
	if !reviewer.CanManageReports() {
		return nil, fmt.Errorf("reviewer does not have permission to manage reports")
	}
	
	// Apply review action
	if err := uc.applyReviewAction(ctx, report, req, reviewer); err != nil {
		logger.Error("Failed to apply review action", err, "report_id", req.ReportID)
		return nil, fmt.Errorf("failed to apply review action: %w", err)
	}
	
	// Update reviewer's last activity
	reviewer.UpdateLastLogin()
	if err := uc.adminUserRepo.Update(ctx, reviewer); err != nil {
		logger.Error("Failed to update reviewer activity", err, "reviewer_id", req.ReviewerID)
		// Don't fail the operation, just log the error
	}
	
	// Send notifications
	if err := uc.sendNotifications(ctx, report, req, reviewer); err != nil {
		logger.Error("Failed to send notifications", err, "report_id", req.ReportID)
		// Don't fail the operation, just log the error
	}
	
	response := &ReviewReportResponse{
		ReportID:    req.ReportID,
		Status:      report.Status,
		ReviewedBy:   req.ReviewerID,
		ReviewedAt:   *report.ReviewedAt,
		ActionTaken:  req.TakeAction,
		Message:      fmt.Sprintf("Report %s successfully", req.Action),
	}
	
	logger.Info("ReviewReport use case executed successfully", "report_id", req.ReportID, "action", req.Action)
	return response, nil
}

// validateBusinessRules validates business rules for reviewing reports
func (uc *ReviewReportUseCase) validateBusinessRules(ctx context.Context, req ReviewReportRequest) error {
	// Check if reviewer is active
	reviewer, err := uc.adminUserRepo.GetByID(ctx, req.ReviewerID)
	if err != nil {
		return fmt.Errorf("reviewer not found: %w", err)
	}
	if !reviewer.IsActive {
		return fmt.Errorf("reviewer account is not active")
	}
	
	// Validate action-specific rules
	switch req.Action {
	case "resolve":
		if req.TakeAction && (req.ActionType == nil || *req.ActionType == "") {
			return fmt.Errorf("action_type is required when take_action is true")
		}
	case "dismiss":
		// Dismissal doesn't require additional validation
	case "escalate":
		// Escalation might require additional permissions
		if !reviewer.IsSuperAdmin() {
			return fmt.Errorf("only super admins can escalate reports")
		}
	default:
		return fmt.Errorf("unknown review action: %s", req.Action)
	}
	
	return nil
}

// applyReviewAction applies the review action to the report
func (uc *ReviewReportUseCase) applyReviewAction(ctx context.Context, report *entities.Report, req ReviewReportRequest, reviewer *entities.AdminUser) error {
	now := time.Now()
	
	switch req.Action {
	case "resolve":
		report.MarkAsResolved(req.ReviewerID)
		
		// Apply moderation action if requested
		if req.TakeAction && req.ActionType != nil {
			action := uc.buildModerationAction(req, reviewer.ID)
			if err := uc.moderationService.ApplyModerationAction(ctx, report.ReportedUserID, action); err != nil {
				return fmt.Errorf("failed to apply moderation action: %w", err)
			}
		}
		
	case "dismiss":
		report.MarkAsDismissed(req.ReviewerID)
		
		// Update reputation positively for dismissed reports
		if err := uc.moderationService.UpdateUserReputation(ctx, report.ReportedUserID, 5.0); err != nil {
			logger.Error("Failed to update user reputation", err, "user_id", report.ReportedUserID)
			// Don't fail the operation, just log the error
		}
		
	case "escalate":
		// Mark as reviewed but keep in pending for higher-level review
		report.MarkAsReviewed(req.ReviewerID)
		// Add escalation metadata
		if report.ReviewedAt != nil {
			// Store escalation reason in notes or metadata
			escalationNote := fmt.Sprintf("Escalated by %s: %s", reviewer.Email, req.Reason)
			if report.Description == nil {
				report.Description = &escalationNote
			} else {
				combinedNote := *report.Description + " | " + escalationNote
				report.Description = &combinedNote
			}
		}
		
	default:
		return fmt.Errorf("unknown review action: %s", req.Action)
	}
	
	// Update report in database
	if err := uc.reportRepo.Update(ctx, report); err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}
	
	return nil
}

// buildModerationAction builds a moderation action from the request
func (uc *ReviewReportUseCase) buildModerationAction(req ReviewReportRequest, reviewerID uuid.UUID) ModerationAction {
	action := ModerationAction{
		Type:      *req.ActionType,
		Reason:    req.Reason,
		AppliedBy: reviewerID,
		AppliedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// Set duration if provided
	if req.ActionDuration != nil {
		duration, err := uc.parseDuration(*req.ActionDuration)
		if err == nil {
			action.Duration = &duration
			action.ExpiresAt = uc.calculateExpiryTime(time.Now(), duration)
		}
		action.Metadata["duration_string"] = *req.ActionDuration
	}
	
	// Add review notes to metadata
	if req.Notes != nil {
		action.Metadata["review_notes"] = *req.Notes
	}
	
	return action
}

// parseDuration parses duration string into time.Duration
func (uc *ReviewReportUseCase) parseDuration(durationStr string) (time.Duration, error) {
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

// calculateExpiryTime calculates when an action expires
func (uc *ReviewReportUseCase) calculateExpiryTime(now time.Time, duration time.Duration) *time.Time {
	if duration == 0 {
		return nil // Permanent action
	}
	
	expiresAt := now.Add(duration)
	return &expiresAt
}

// sendNotifications sends notifications related to the review
func (uc *ReviewReportUseCase) sendNotifications(ctx context.Context, report *entities.Report, req ReviewReportRequest, reviewer *entities.AdminUser) error {
	// Notify reporter about the resolution
	reporterData := map[string]interface{}{
		"report_id":   report.ID,
		"status":       report.Status,
		"reviewed_by":  reviewer.Email,
		"reason":       req.Reason,
	}
	
	if err := uc.notificationService.SendNotification(ctx, report.ReporterID, "report_resolved", reporterData); err != nil {
		return fmt.Errorf("failed to notify reporter: %w", err)
	}
	
	// Notify reported user if action was taken
	if req.TakeAction && req.ActionType != nil {
		reportedData := map[string]interface{}{
			"report_id":   report.ID,
			"action_type":  *req.ActionType,
			"reason":       req.Reason,
			"reviewed_by":  reviewer.Email,
		}
		
		if err := uc.notificationService.SendNotification(ctx, report.ReportedUserID, "moderation_action", reportedData); err != nil {
			return fmt.Errorf("failed to notify reported user: %w", err)
		}
	}
	
	// Notify other admins if escalated
	if req.Action == "escalate" {
		escalationData := map[string]interface{}{
			"report_id":   report.ID,
			"escalated_by": reviewer.Email,
			"reason":       req.Reason,
		}
		
		if err := uc.notificationService.SendAdminNotification(ctx, uuid.Nil, "report_escalated", escalationData); err != nil {
			return fmt.Errorf("failed to notify admins about escalation: %w", err)
		}
	}
	
	return nil
}

// GetReportDetails gets detailed information about a report
func (uc *ReviewReportUseCase) GetReportDetails(ctx context.Context, reportID, reviewerID uuid.UUID) (*entities.Report, error) {
	logger.Info("Getting report details", "report_id", reportID, "reviewer_id", reviewerID)
	
	// Get report
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		logger.Error("Failed to get report", err, "report_id", reportID)
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	
	// Check if reviewer has permission to view this report
	reviewer, err := uc.adminUserRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	if !reviewer.CanManageReports() {
		return nil, fmt.Errorf("reviewer does not have permission to view reports")
	}
	
	return report, nil
}

// GetPendingReports gets all pending reports for review
func (uc *ReviewReportUseCase) GetPendingReports(ctx context.Context, reviewerID uuid.UUID, limit, offset int) ([]*entities.Report, error) {
	logger.Info("Getting pending reports", "reviewer_id", reviewerID, "limit", limit, "offset", offset)
	
	// Check reviewer permissions
	reviewer, err := uc.adminUserRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	if !reviewer.CanManageReports() {
		return nil, fmt.Errorf("reviewer does not have permission to view reports")
	}
	
	// Get pending reports
	reports, err := uc.reportRepo.GetPendingReports(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get pending reports", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to get pending reports: %w", err)
	}
	
	return reports, nil
}

// GetReportsByStatus gets reports by status
func (uc *ReviewReportUseCase) GetReportsByStatus(ctx context.Context, reviewerID uuid.UUID, status string, limit, offset int) ([]*entities.Report, error) {
	logger.Info("Getting reports by status", "reviewer_id", reviewerID, "status", status, "limit", limit, "offset", offset)
	
	// Check reviewer permissions
	reviewer, err := uc.adminUserRepo.GetByID(ctx, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}
	
	if !reviewer.CanManageReports() {
		return nil, fmt.Errorf("reviewer does not have permission to view reports")
	}
	
	// Get reports by status
	reports, err := uc.reportRepo.GetByStatus(ctx, status, limit, offset)
	if err != nil {
		logger.Error("Failed to get reports by status", err, "reviewer_id", reviewerID, "status", status)
		return nil, fmt.Errorf("failed to get reports by status: %w", err)
	}
	
	return reports, nil
}