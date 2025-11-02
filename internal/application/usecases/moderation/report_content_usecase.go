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

// ReportContentRequest represents a request to report content
type ReportContentRequest struct {
	ReporterID     uuid.UUID `json:"reporter_id" validate:"required"`
	ReportedUserID uuid.UUID `json:"reported_user_id" validate:"required"`
	Reason         string    `json:"reason" validate:"required,oneof=inappropriate_behavior fake_profile spam harassment other"`
	Description     *string   `json:"description"`
	ContentType    string    `json:"content_type,omitempty"` // "message", "photo", "profile"
	ContentID      *string   `json:"content_id,omitempty"`
	Evidence       *string   `json:"evidence,omitempty"`
}

// ReportContentResponse represents the response after reporting content
type ReportContentResponse struct {
	ReportID    uuid.UUID `json:"report_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReportContentUseCase handles reporting content functionality
type ReportContentUseCase struct {
	reportRepo         repositories.ReportRepository
	userRepo           repositories.UserRepository
	validator          validator.Validator
	rateLimiter       RateLimiter
	maxReportsPerDay   int
	maxReportsPerUser  int
}

// NewReportContentUseCase creates a new ReportContentUseCase
func NewReportContentUseCase(
	reportRepo repositories.ReportRepository,
	userRepo repositories.UserRepository,
	validator validator.Validator,
	rateLimiter RateLimiter,
	maxReportsPerDay, maxReportsPerUser int,
) *ReportContentUseCase {
	return &ReportContentUseCase{
		reportRepo:        reportRepo,
		userRepo:          userRepo,
		validator:         validator,
		rateLimiter:       rateLimiter,
		maxReportsPerDay:  maxReportsPerDay,
		maxReportsPerUser: maxReportsPerUser,
	}
}

// Execute executes the report content use case
func (uc *ReportContentUseCase) Execute(ctx context.Context, req ReportContentRequest) (*ReportContentResponse, error) {
	logger.Info("Executing ReportContent use case", "reporter_id", req.ReporterID, "reported_user_id", req.ReportedUserID, "reason", req.Reason)
	
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
	
	// Check rate limits
	if err := uc.checkRateLimits(ctx, req.ReporterID); err != nil {
		logger.Error("Rate limit check failed", err, "reporter_id", req.ReporterID)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	
	// Verify users exist
	if err := uc.verifyUsers(ctx, req); err != nil {
		logger.Error("User verification failed", err, "request", req)
		return nil, fmt.Errorf("user verification failed: %w", err)
	}
	
	// Create report entity
	report := &entities.Report{
		ReporterID:     req.ReporterID,
		ReportedUserID: req.ReportedUserID,
		Reason:         req.Reason,
		Description:    req.Description,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	
	// Save report to database
	if err := uc.reportRepo.Create(ctx, report); err != nil {
		logger.Error("Failed to create report", err, "report_id", report.ID)
		return nil, fmt.Errorf("failed to create report: %w", err)
	}
	
	// Log additional evidence if provided
	if req.Evidence != nil && *req.Evidence != "" {
		if err := uc.logEvidence(ctx, report.ID, *req.Evidence); err != nil {
			logger.Error("Failed to log evidence", err, "report_id", report.ID)
			// Don't fail the operation, just log the error
		}
	}
	
	// Update reporter's statistics
	if err := uc.updateReporterStats(ctx, req.ReporterID); err != nil {
		logger.Error("Failed to update reporter stats", err, "reporter_id", req.ReporterID)
		// Don't fail the operation, just log the error
	}
	
	// Update reported user's statistics
	if err := uc.updateReportedUserStats(ctx, req.ReportedUserID); err != nil {
		logger.Error("Failed to update reported user stats", err, "reported_user_id", req.ReportedUserID)
		// Don't fail the operation, just log the error
	}
	
	response := &ReportContentResponse{
		ReportID:  report.ID,
		Status:    report.Status,
		Message:    "Report submitted successfully",
		CreatedAt:  report.CreatedAt,
	}
	
	logger.Info("ReportContent use case executed successfully", "report_id", report.ID, "reporter_id", req.ReporterID)
	return response, nil
}

// validateBusinessRules validates business rules for reporting
func (uc *ReportContentUseCase) validateBusinessRules(ctx context.Context, req ReportContentRequest) error {
	// Cannot report yourself
	if req.ReporterID == req.ReportedUserID {
		return fmt.Errorf("cannot report yourself")
	}
	
	// Check if users have an active conversation (optional business rule)
	if err := uc.checkUserRelationship(ctx, req); err != nil {
		return fmt.Errorf("user relationship check failed: %w", err)
	}
	
	// Check if there's already an active report
	hasActiveReport, err := uc.reportRepo.HasActiveReport(ctx, req.ReporterID, req.ReportedUserID)
	if err != nil {
		return fmt.Errorf("failed to check for active reports: %w", err)
	}
	if hasActiveReport {
		return fmt.Errorf("active report already exists for this user pair")
	}
	
	return nil
}

// checkRateLimits checks if user has exceeded rate limits
func (uc *ReportContentUseCase) checkRateLimits(ctx context.Context, userID uuid.UUID) error {
	// Check daily limit
	if err := uc.rateLimiter.CheckRateLimit(ctx, "reports_daily", userID.String(), uc.maxReportsPerDay, 24*time.Hour); err != nil {
		return fmt.Errorf("daily report limit exceeded: %w", err)
	}
	
	// Check per-user limit
	if err := uc.rateLimiter.CheckRateLimit(ctx, "reports_per_user", userID.String(), uc.maxReportsPerUser, 24*time.Hour); err != nil {
		return fmt.Errorf("per-user report limit exceeded: %w", err)
	}
	
	return nil
}

// verifyUsers verifies that both reporter and reported users exist and are active
func (uc *ReportContentUseCase) verifyUsers(ctx context.Context, req ReportContentRequest) error {
	// Verify reporter exists and is active
	reporter, err := uc.userRepo.GetByID(ctx, req.ReporterID)
	if err != nil {
		return fmt.Errorf("reporter not found: %w", err)
	}
	if !reporter.IsActive || reporter.IsBanned {
		return fmt.Errorf("reporter account is not active")
	}
	
	// Verify reported user exists
	reported, err := uc.userRepo.GetByID(ctx, req.ReportedUserID)
	if err != nil {
		return fmt.Errorf("reported user not found: %w", err)
	}
	
	// Allow reporting banned users (this is how we catch them)
	
	return nil
}

// checkUserRelationship checks if users have an active relationship (optional business rule)
func (uc *ReportContentUseCase) checkUserRelationship(ctx context.Context, req ReportContentRequest) error {
	// This is an optional business rule - some platforms allow reporting only if
	// users have interacted, others allow reporting anyone
	// For this implementation, we'll allow reporting anyone
	
	return nil
}

// logEvidence logs additional evidence for the report
func (uc *ReportContentUseCase) logEvidence(ctx context.Context, reportID uuid.UUID, evidence string) error {
	// In a real implementation, this would store evidence in a separate table
	// For now, we'll just log it
	logger.Info("Report evidence logged", "report_id", reportID, "evidence", evidence)
	return nil
}

// updateReporterStats updates the reporter's statistics
func (uc *ReportContentUseCase) updateReporterStats(ctx context.Context, reporterID uuid.UUID) error {
	// In a real implementation, this would update user statistics
	// For now, we'll just log it
	logger.Info("Reporter stats updated", "reporter_id", reporterID)
	return nil
}

// updateReportedUserStats updates the reported user's statistics
func (uc *ReportContentUseCase) updateReportedUserStats(ctx context.Context, reportedUserID uuid.UUID) error {
	// In a real implementation, this would update user statistics
	// For now, we'll just log it
	logger.Info("Reported user stats updated", "reported_user_id", reportedUserID)
	return nil
}

// GetReportStatus gets the status of a report
func (uc *ReportContentUseCase) GetReportStatus(ctx context.Context, reportID, userID uuid.UUID) (*entities.Report, error) {
	logger.Info("Getting report status", "report_id", reportID, "user_id", userID)
	
	// Get report
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		logger.Error("Failed to get report", err, "report_id", reportID)
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	
	// Check if user has permission to view this report
	if report.ReporterID != userID && report.ReportedUserID != userID {
		return nil, fmt.Errorf("unauthorized to view this report")
	}
	
	return report, nil
}

// GetUserReports gets all reports filed by a user
func (uc *ReportContentUseCase) GetUserReports(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Report, error) {
	logger.Info("Getting user reports", "user_id", userID, "limit", limit, "offset", offset)
	
	reports, err := uc.reportRepo.GetByReporter(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user reports", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user reports: %w", err)
	}
	
	return reports, nil
}

// GetReportsAgainstUser gets all reports filed against a user
func (uc *ReportContentUseCase) GetReportsAgainstUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Report, error) {
	logger.Info("Getting reports against user", "user_id", userID, "limit", limit, "offset", offset)
	
	reports, err := uc.reportRepo.GetByReportedUser(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get reports against user", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get reports against user: %w", err)
	}
	
	return reports, nil
}

// CancelReport cancels a pending report (only by the reporter)
func (uc *ReportContentUseCase) CancelReport(ctx context.Context, reportID, reporterID uuid.UUID) error {
	logger.Info("Cancelling report", "report_id", reportID, "reporter_id", reporterID)
	
	// Get report
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		logger.Error("Failed to get report", err, "report_id", reportID)
		return fmt.Errorf("failed to get report: %w", err)
	}
	
	// Check if user is the reporter
	if report.ReporterID != reporterID {
		return fmt.Errorf("only the reporter can cancel their own report")
	}
	
	// Check if report is still pending
	if !report.IsPending() {
		return fmt.Errorf("cannot cancel a report that is already being reviewed")
	}
	
	// Delete the report
	if err := uc.reportRepo.Delete(ctx, reportID); err != nil {
		logger.Error("Failed to delete report", err, "report_id", reportID)
		return fmt.Errorf("failed to delete report: %w", err)
	}
	
	logger.Info("Report cancelled successfully", "report_id", reportID)
	return nil
}