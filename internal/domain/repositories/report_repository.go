package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// ReportRepository defines interface for report data operations
type ReportRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, report *entities.Report) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error)
	Update(ctx context.Context, report *entities.Report) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Report filtering and search
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Report, error)
	GetByReporter(ctx context.Context, reporterID uuid.UUID, limit, offset int) ([]*entities.Report, error)
	GetByReportedUser(ctx context.Context, reportedUserID uuid.UUID, limit, offset int) ([]*entities.Report, error)
	GetByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.Report, error)
	GetByDateRange(ctx context.Context, startDate, endDate interface{}, limit, offset int) ([]*entities.Report, error)

	// Report management
	GetPendingReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)
	GetReviewedReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)
	GetResolvedReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)
	GetDismissedReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)

	// Report review operations
	MarkAsReviewed(ctx context.Context, reportID, reviewerID uuid.UUID) error
	MarkAsResolved(ctx context.Context, reportID, reviewerID uuid.UUID) error
	MarkAsDismissed(ctx context.Context, reportID, reviewerID uuid.UUID) error

	// User report statistics
	GetUserReportCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetUserReportsByStatus(ctx context.Context, userID uuid.UUID, status string) (int64, error)
	GetUserReportsAsReporter(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Report, error)
	GetUserReportsAsReported(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Report, error)

	// Batch operations
	BatchCreate(ctx context.Context, reports []*entities.Report) error
	BatchUpdate(ctx context.Context, reports []*entities.Report) error
	BatchMarkAsReviewed(ctx context.Context, reportIDs []uuid.UUID, reviewerID uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	UserCanReport(ctx context.Context, reporterID, reportedUserID uuid.UUID) (bool, error)
	HasActiveReport(ctx context.Context, reporterID, reportedUserID uuid.UUID) (bool, error)

	// Analytics and statistics
	GetReportStats(ctx context.Context) (*ReportStats, error)
	GetReportsByReasonStats(ctx context.Context, startDate, endDate interface{}) ([]*ReportReasonStats, error)
	GetReportsByDateStats(ctx context.Context, startDate, endDate interface{}) ([]*ReportDateStats, error)
	GetReportsCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)

	// Admin operations
	GetAllReports(ctx context.Context, limit, offset int) ([]*entities.Report, error)
	GetReportsWithDetails(ctx context.Context, limit, offset int) ([]*ReportWithDetails, error)
	GetRecentReports(ctx context.Context, hours int, limit int) ([]*entities.Report, error)

	// Advanced queries
	GetReportsByReviewer(ctx context.Context, reviewerID uuid.UUID, limit, offset int) ([]*entities.Report, error)
	GetOverdueReports(ctx context.Context, hours int, limit int) ([]*entities.Report, error)
	GetHighPriorityReports(ctx context.Context, limit int) ([]*entities.Report, error)
	SearchReports(ctx context.Context, query string, limit, offset int) ([]*entities.Report, error)
}

// ReportStats represents overall report statistics
type ReportStats struct {
	TotalReports      int64 `json:"total_reports"`
	PendingReports    int64 `json:"pending_reports"`
	ReviewedReports   int64 `json:"reviewed_reports"`
	ResolvedReports   int64 `json:"resolved_reports"`
	DismissedReports  int64 `json:"dismissed_reports"`
	ReportsToday     int64 `json:"reports_today"`
	ReportsThisWeek  int64 `json:"reports_this_week"`
	ReportsThisMonth int64 `json:"reports_this_month"`
	AverageResolutionTime float64 `json:"average_resolution_time_hours"`
}

// ReportReasonStats represents report statistics by reason
type ReportReasonStats struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ReportDateStats represents report statistics by date
type ReportDateStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ReportWithDetails represents a report with additional details
type ReportWithDetails struct {
	*entities.Report
	ReporterUser     *entities.User `json:"reporter_user"`
	ReportedUser    *entities.User `json:"reported_user"`
	ReviewerUser    *entities.AdminUser `json:"reviewer_user,omitempty"`
	ResolutionTime   *float64 `json:"resolution_time_hours,omitempty"`
}