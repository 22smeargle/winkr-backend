package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ModerationAnalyticsService handles moderation analytics and insights
type ModerationAnalyticsService struct {
	reportRepo         ReportRepository
	userRepo           UserRepository
	banRepo            BanRepository
	appealRepo         AppealRepository
	moderationActionRepo ModerationActionRepository
	contentAnalysisRepo ContentAnalysisRepository
	moderationQueueRepo ModerationQueueRepository
	cacheService       *ModerationCacheService
}

// NewModerationAnalyticsService creates a new moderation analytics service
func NewModerationAnalyticsService(
	reportRepo ReportRepository,
	userRepo UserRepository,
	banRepo BanRepository,
	appealRepo AppealRepository,
	moderationActionRepo ModerationActionRepository,
	contentAnalysisRepo ContentAnalysisRepository,
	moderationQueueRepo ModerationQueueRepository,
	cacheService *ModerationCacheService,
) *ModerationAnalyticsService {
	return &ModerationAnalyticsService{
		reportRepo:           reportRepo,
		userRepo:             userRepo,
		banRepo:              banRepo,
		appealRepo:           appealRepo,
		moderationActionRepo:  moderationActionRepo,
		contentAnalysisRepo:   contentAnalysisRepo,
		moderationQueueRepo:   moderationQueueRepo,
		cacheService:          cacheService,
	}
}

// ReportAnalytics represents report analytics data
type ReportAnalytics struct {
	TotalReports       int                    `json:"total_reports"`
	PendingReports     int                    `json:"pending_reports"`
	ResolvedReports    int                    `json:"resolved_reports"`
	DismissedReports   int                    `json:"dismissed_reports"`
	ReportsByReason    map[string]int          `json:"reports_by_reason"`
	ReportsByStatus    map[string]int          `json:"reports_by_status"`
	ReportsTrend      []TrendData            `json:"reports_trend"`
	AverageResolutionTime time.Duration         `json:"average_resolution_time"`
	TopReportedUsers  []UserReportStats      `json:"top_reported_users"`
}

// UserReportStats represents user report statistics
type UserReportStats struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Reports   int       `json:"reports"`
	Resolved  int       `json:"resolved"`
	Dismissed int       `json:"dismissed"`
}

// TrendData represents trend data over time
type TrendData struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// BanAnalytics represents ban analytics data
type BanAnalytics struct {
	TotalBans         int                    `json:"total_bans"`
	ActiveBans        int                    `json:"active_bans"`
	PermanentBans     int                    `json:"permanent_bans"`
	TemporaryBans     int                    `json:"temporary_bans"`
	BansByReason      map[string]int          `json:"bans_by_reason"`
	BansByDuration    map[string]int          `json:"bans_by_duration"`
	BansTrend         []TrendData            `json:"bans_trend"`
	AverageBanDuration time.Duration         `json:"average_ban_duration"`
	TopBannedUsers    []UserBanStats         `json:"top_banned_users"`
}

// UserBanStats represents user ban statistics
type UserBanStats struct {
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	Bans       int       `json:"bans"`
	Permanent  bool      `json:"permanent"`
	LastBan    time.Time `json:"last_ban"`
}

// AppealAnalytics represents appeal analytics data
type AppealAnalytics struct {
	TotalAppeals      int                    `json:"total_appeals"`
	PendingAppeals    int                    `json:"pending_appeals"`
	ApprovedAppeals   int                    `json:"approved_appeals"`
	RejectedAppeals   int                    `json:"rejected_appeals"`
	AppealSuccessRate float64                `json:"appeal_success_rate"`
	AppealsTrend     []TrendData            `json:"appeals_trend"`
	AverageReviewTime time.Duration         `json:"average_review_time"`
	TopAppealingUsers []UserAppealStats      `json:"top_appealing_users"`
}

// UserAppealStats represents user appeal statistics
type UserAppealStats struct {
	UserID          uuid.UUID `json:"user_id"`
	Username        string    `json:"username"`
	Appeals         int       `json:"appeals"`
	Approved        int       `json:"approved"`
	Rejected        int       `json:"rejected"`
	SuccessRate     float64   `json:"success_rate"`
}

// ContentAnalytics represents content analytics data
type ContentAnalytics struct {
	TotalAnalyses       int                    `json:"total_analyses"`
	SafeContent        int                    `json:"safe_content"`
	SuspiciousContent  int                    `json:"suspicious_content"`
	ViolatingContent   int                    `json:"violating_content"`
	AnalysesByType     map[string]int          `json:"analyses_by_type"`
	AnalysesByResult   map[string]int          `json:"analyses_by_result"`
	ContentTrend       []TrendData            `json:"content_trend"`
	AverageConfidence   float64                `json:"average_confidence"`
	TopViolatingTypes  []ContentTypeStats      `json:"top_violating_types"`
}

// ContentTypeStats represents content type statistics
type ContentTypeStats struct {
	ContentType string  `json:"content_type"`
	Count       int     `json:"count"`
	Violations  int     `json:"violations"`
	ViolationRate float64 `json:"violation_rate"`
}

// ModeratorAnalytics represents moderator performance analytics
type ModeratorAnalytics struct {
	ModeratorID        uuid.UUID               `json:"moderator_id"`
	ModeratorName      string                  `json:"moderator_name"`
	ReportsReviewed     int                     `json:"reports_reviewed"`
	ReportsResolved    int                     `json:"reports_resolved"`
	ReportsDismissed   int                     `json:"reports_dismissed"`
	AverageReviewTime   time.Duration           `json:"average_review_time"`
	AppealsReviewed    int                     `json:"appeals_reviewed"`
	AppealsApproved    int                     `json:"appeals_approved"`
	AppealsRejected    int                     `json:"appeals_rejected"`
	BansIssued        int                     `json:"bans_issued"`
	SuspensionsIssued  int                     `json:"suspensions_issued"`
	PerformanceScore    float64                 `json:"performance_score"`
	ActivityTrend      []TrendData             `json:"activity_trend"`
}

// OverallAnalytics represents overall moderation analytics
type OverallAnalytics struct {
	Period             string               `json:"period"`
	GeneratedAt        time.Time            `json:"generated_at"`
	Reports            ReportAnalytics       `json:"reports"`
	Bans               BanAnalytics         `json:"bans"`
	Appeals            AppealAnalytics      `json:"appeals"`
	Content            ContentAnalytics     `json:"content"`
	Moderators         []ModeratorAnalytics `json:"moderators"`
	UserReputation     map[string]int       `json:"user_reputation"`
	SystemHealth       SystemHealthStats    `json:"system_health"`
}

// SystemHealthStats represents system health statistics
type SystemHealthStats struct {
	QueueSize          int           `json:"queue_size"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	ProcessingRate     float64       `json:"processing_rate"`
	ErrorRate          float64       `json:"error_rate"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
}

// GetReportAnalytics generates report analytics for the specified period
func (mas *ModerationAnalyticsService) GetReportAnalytics(ctx context.Context, period string) (*ReportAnalytics, error) {
	logger.Info("Generating report analytics", "period", period)
	
	// Try to get from cache first
	if cached, err := mas.cacheService.GetReportStats(ctx); err == nil && cached != nil {
		logger.Debug("Report analytics retrieved from cache")
		return &ReportAnalytics{
			TotalReports:     getIntFromCache(cached, "total_reports"),
			PendingReports:   getIntFromCache(cached, "pending_reports"),
			ResolvedReports:  getIntFromCache(cached, "resolved_reports"),
			DismissedReports: getIntFromCache(cached, "dismissed_reports"),
		}, nil
	}
	
	// Calculate date range
	startDate, endDate := mas.getDateRange(period)
	
	// Get report statistics
	totalReports, err := mas.reportRepo.CountByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get total reports", err)
		return nil, fmt.Errorf("failed to get total reports: %w", err)
	}
	
	pendingReports, err := mas.reportRepo.CountByStatusAndDateRange(ctx, "pending", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get pending reports", err)
		return nil, fmt.Errorf("failed to get pending reports: %w", err)
	}
	
	resolvedReports, err := mas.reportRepo.CountByStatusAndDateRange(ctx, "resolved", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get resolved reports", err)
		return nil, fmt.Errorf("failed to get resolved reports: %w", err)
	}
	
	dismissedReports, err := mas.reportRepo.CountByStatusAndDateRange(ctx, "dismissed", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get dismissed reports", err)
		return nil, fmt.Errorf("failed to get dismissed reports: %w", err)
	}
	
	reportsByReason, err := mas.reportRepo.CountByReasonAndDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get reports by reason", err)
		return nil, fmt.Errorf("failed to get reports by reason: %w", err)
	}
	
	reportsByStatus, err := mas.reportRepo.CountByStatusAndDateRange(ctx, "", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get reports by status", err)
		return nil, fmt.Errorf("failed to get reports by status: %w", err)
	}
	
	reportsTrend, err := mas.getReportsTrend(ctx, period)
	if err != nil {
		logger.Error("Failed to get reports trend", err)
		return nil, fmt.Errorf("failed to get reports trend: %w", err)
	}
	
	averageResolutionTime, err := mas.getAverageResolutionTime(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get average resolution time", err)
		return nil, fmt.Errorf("failed to get average resolution time: %w", err)
	}
	
	topReportedUsers, err := mas.getTopReportedUsers(ctx, startDate, endDate, 10)
	if err != nil {
		logger.Error("Failed to get top reported users", err)
		return nil, fmt.Errorf("failed to get top reported users: %w", err)
	}
	
	analytics := &ReportAnalytics{
		TotalReports:        totalReports,
		PendingReports:      pendingReports,
		ResolvedReports:     resolvedReports,
		DismissedReports:    dismissedReports,
		ReportsByReason:     reportsByReason,
		ReportsByStatus:     reportsByStatus,
		ReportsTrend:       reportsTrend,
		AverageResolutionTime: averageResolutionTime,
		TopReportedUsers:    topReportedUsers,
	}
	
	// Cache the results
	cacheData := map[string]interface{}{
		"total_reports":     totalReports,
		"pending_reports":   pendingReports,
		"resolved_reports":  resolvedReports,
		"dismissed_reports": dismissedReports,
	}
	mas.cacheService.CacheReportStats(ctx, cacheData)
	
	logger.Info("Report analytics generated successfully", "period", period, "total_reports", totalReports)
	return analytics, nil
}

// GetBanAnalytics generates ban analytics for the specified period
func (mas *ModerationAnalyticsService) GetBanAnalytics(ctx context.Context, period string) (*BanAnalytics, error) {
	logger.Info("Generating ban analytics", "period", period)
	
	// Calculate date range
	startDate, endDate := mas.getDateRange(period)
	
	// Get ban statistics
	totalBans, err := mas.banRepo.CountByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get total bans", err)
		return nil, fmt.Errorf("failed to get total bans: %w", err)
	}
	
	activeBans, err := mas.banRepo.CountActiveByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get active bans", err)
		return nil, fmt.Errorf("failed to get active bans: %w", err)
	}
	
	permanentBans, err := mas.banRepo.CountPermanentByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get permanent bans", err)
		return nil, fmt.Errorf("failed to get permanent bans: %w", err)
	}
	
	temporaryBans, err := mas.banRepo.CountTemporaryByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get temporary bans", err)
		return nil, fmt.Errorf("failed to get temporary bans: %w", err)
	}
	
	bansByReason, err := mas.banRepo.CountByReasonAndDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get bans by reason", err)
		return nil, fmt.Errorf("failed to get bans by reason: %w", err)
	}
	
	bansByDuration, err := mas.banRepo.CountByDurationAndDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get bans by duration", err)
		return nil, fmt.Errorf("failed to get bans by duration: %w", err)
	}
	
	bansTrend, err := mas.getBansTrend(ctx, period)
	if err != nil {
		logger.Error("Failed to get bans trend", err)
		return nil, fmt.Errorf("failed to get bans trend: %w", err)
	}
	
	averageBanDuration, err := mas.getAverageBanDuration(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get average ban duration", err)
		return nil, fmt.Errorf("failed to get average ban duration: %w", err)
	}
	
	topBannedUsers, err := mas.getTopBannedUsers(ctx, startDate, endDate, 10)
	if err != nil {
		logger.Error("Failed to get top banned users", err)
		return nil, fmt.Errorf("failed to get top banned users: %w", err)
	}
	
	analytics := &BanAnalytics{
		TotalBans:         totalBans,
		ActiveBans:        activeBans,
		PermanentBans:     permanentBans,
		TemporaryBans:     temporaryBans,
		BansByReason:      bansByReason,
		BansByDuration:    bansByDuration,
		BansTrend:         bansTrend,
		AverageBanDuration: averageBanDuration,
		TopBannedUsers:    topBannedUsers,
	}
	
	logger.Info("Ban analytics generated successfully", "period", period, "total_bans", totalBans)
	return analytics, nil
}

// GetAppealAnalytics generates appeal analytics for the specified period
func (mas *ModerationAnalyticsService) GetAppealAnalytics(ctx context.Context, period string) (*AppealAnalytics, error) {
	logger.Info("Generating appeal analytics", "period", period)
	
	// Calculate date range
	startDate, endDate := mas.getDateRange(period)
	
	// Get appeal statistics
	totalAppeals, err := mas.appealRepo.CountByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get total appeals", err)
		return nil, fmt.Errorf("failed to get total appeals: %w", err)
	}
	
	pendingAppeals, err := mas.appealRepo.CountByStatusAndDateRange(ctx, "pending", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get pending appeals", err)
		return nil, fmt.Errorf("failed to get pending appeals: %w", err)
	}
	
	approvedAppeals, err := mas.appealRepo.CountByStatusAndDateRange(ctx, "approved", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get approved appeals", err)
		return nil, fmt.Errorf("failed to get approved appeals: %w", err)
	}
	
	rejectedAppeals, err := mas.appealRepo.CountByStatusAndDateRange(ctx, "rejected", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get rejected appeals", err)
		return nil, fmt.Errorf("failed to get rejected appeals: %w", err)
	}
	
	var appealSuccessRate float64
	if totalAppeals > 0 {
		appealSuccessRate = float64(approvedAppeals) / float64(totalAppeals) * 100
	}
	
	appealsTrend, err := mas.getAppealsTrend(ctx, period)
	if err != nil {
		logger.Error("Failed to get appeals trend", err)
		return nil, fmt.Errorf("failed to get appeals trend: %w", err)
	}
	
	averageReviewTime, err := mas.getAverageAppealReviewTime(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get average appeal review time", err)
		return nil, fmt.Errorf("failed to get average appeal review time: %w", err)
	}
	
	topAppealingUsers, err := mas.getTopAppealingUsers(ctx, startDate, endDate, 10)
	if err != nil {
		logger.Error("Failed to get top appealing users", err)
		return nil, fmt.Errorf("failed to get top appealing users: %w", err)
	}
	
	analytics := &AppealAnalytics{
		TotalAppeals:      totalAppeals,
		PendingAppeals:    pendingAppeals,
		ApprovedAppeals:   approvedAppeals,
		RejectedAppeals:   rejectedAppeals,
		AppealSuccessRate: appealSuccessRate,
		AppealsTrend:     appealsTrend,
		AverageReviewTime: averageReviewTime,
		TopAppealingUsers: topAppealingUsers,
	}
	
	logger.Info("Appeal analytics generated successfully", "period", period, "total_appeals", totalAppeals)
	return analytics, nil
}

// GetContentAnalytics generates content analytics for the specified period
func (mas *ModerationAnalyticsService) GetContentAnalytics(ctx context.Context, period string) (*ContentAnalytics, error) {
	logger.Info("Generating content analytics", "period", period)
	
	// Calculate date range
	startDate, endDate := mas.getDateRange(period)
	
	// Get content analysis statistics
	totalAnalyses, err := mas.contentAnalysisRepo.CountByDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get total analyses", err)
		return nil, fmt.Errorf("failed to get total analyses: %w", err)
	}
	
	safeContent, err := mas.contentAnalysisRepo.CountByResultAndDateRange(ctx, "safe", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get safe content count", err)
		return nil, fmt.Errorf("failed to get safe content count: %w", err)
	}
	
	suspiciousContent, err := mas.contentAnalysisRepo.CountByResultAndDateRange(ctx, "suspicious", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get suspicious content count", err)
		return nil, fmt.Errorf("failed to get suspicious content count: %w", err)
	}
	
	violatingContent, err := mas.contentAnalysisRepo.CountByResultAndDateRange(ctx, "violating", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get violating content count", err)
		return nil, fmt.Errorf("failed to get violating content count: %w", err)
	}
	
	analysesByType, err := mas.contentAnalysisRepo.CountByTypeAndDateRange(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get analyses by type", err)
		return nil, fmt.Errorf("failed to get analyses by type: %w", err)
	}
	
	analysesByResult, err := mas.contentAnalysisRepo.CountByResultAndDateRange(ctx, "", startDate, endDate)
	if err != nil {
		logger.Error("Failed to get analyses by result", err)
		return nil, fmt.Errorf("failed to get analyses by result: %w", err)
	}
	
	contentTrend, err := mas.getContentTrend(ctx, period)
	if err != nil {
		logger.Error("Failed to get content trend", err)
		return nil, fmt.Errorf("failed to get content trend: %w", err)
	}
	
	averageConfidence, err := mas.getAverageConfidence(ctx, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get average confidence", err)
		return nil, fmt.Errorf("failed to get average confidence: %w", err)
	}
	
	topViolatingTypes, err := mas.getTopViolatingTypes(ctx, startDate, endDate, 10)
	if err != nil {
		logger.Error("Failed to get top violating types", err)
		return nil, fmt.Errorf("failed to get top violating types: %w", err)
	}
	
	analytics := &ContentAnalytics{
		TotalAnalyses:      totalAnalyses,
		SafeContent:        safeContent,
		SuspiciousContent:  suspiciousContent,
		ViolatingContent:   violatingContent,
		AnalysesByType:     analysesByType,
		AnalysesByResult:   analysesByResult,
		ContentTrend:       contentTrend,
		AverageConfidence:   averageConfidence,
		TopViolatingTypes:  topViolatingTypes,
	}
	
	logger.Info("Content analytics generated successfully", "period", period, "total_analyses", totalAnalyses)
	return analytics, nil
}

// GetModeratorAnalytics generates moderator performance analytics for the specified period
func (mas *ModerationAnalyticsService) GetModeratorAnalytics(ctx context.Context, moderatorID uuid.UUID, period string) (*ModeratorAnalytics, error) {
	logger.Info("Generating moderator analytics", "moderator_id", moderatorID, "period", period)
	
	// Calculate date range
	startDate, endDate := mas.getDateRange(period)
	
	// Get moderator information
	moderator, err := mas.userRepo.GetByID(ctx, moderatorID)
	if err != nil {
		logger.Error("Failed to get moderator", err)
		return nil, fmt.Errorf("failed to get moderator: %w", err)
	}
	
	// Get moderator statistics
	reportsReviewed, err := mas.reportRepo.CountReviewedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get reports reviewed", err)
		return nil, fmt.Errorf("failed to get reports reviewed: %w", err)
	}
	
	reportsResolved, err := mas.reportRepo.CountResolvedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get reports resolved", err)
		return nil, fmt.Errorf("failed to get reports resolved: %w", err)
	}
	
	reportsDismissed, err := mas.reportRepo.CountDismissedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get reports dismissed", err)
		return nil, fmt.Errorf("failed to get reports dismissed: %w", err)
	}
	
	averageReviewTime, err := mas.getAverageModeratorReviewTime(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get average review time", err)
		return nil, fmt.Errorf("failed to get average review time: %w", err)
	}
	
	appealsReviewed, err := mas.appealRepo.CountReviewedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get appeals reviewed", err)
		return nil, fmt.Errorf("failed to get appeals reviewed: %w", err)
	}
	
	appealsApproved, err := mas.appealRepo.CountApprovedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get appeals approved", err)
		return nil, fmt.Errorf("failed to get appeals approved: %w", err)
	}
	
	appealsRejected, err := mas.appealRepo.CountRejectedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get appeals rejected", err)
		return nil, fmt.Errorf("failed to get appeals rejected: %w", err)
	}
	
	bansIssued, err := mas.banRepo.CountIssuedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get bans issued", err)
		return nil, fmt.Errorf("failed to get bans issued: %w", err)
	}
	
	suspensionsIssued, err := mas.banRepoCountSuspensionsIssuedByModeratorAndDateRange(ctx, moderatorID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get suspensions issued", err)
		return nil, fmt.Errorf("failed to get suspensions issued: %w", err)
	}
	
	performanceScore := mas.calculateModeratorPerformanceScore(
		reportsReviewed, reportsResolved, averageReviewTime,
		appealsReviewed, appealsApproved, bansIssued,
	)
	
	activityTrend, err := mas.getModeratorActivityTrend(ctx, moderatorID, period)
	if err != nil {
		logger.Error("Failed to get moderator activity trend", err)
		return nil, fmt.Errorf("failed to get moderator activity trend: %w", err)
	}
	
	analytics := &ModeratorAnalytics{
		ModeratorID:       moderatorID,
		ModeratorName:     moderator.FirstName + " " + moderator.LastName,
		ReportsReviewed:    reportsReviewed,
		ReportsResolved:    reportsResolved,
		ReportsDismissed:   reportsDismissed,
		AverageReviewTime:  averageReviewTime,
		AppealsReviewed:    appealsReviewed,
		AppealsApproved:    appealsApproved,
		AppealsRejected:    appealsRejected,
		BansIssued:        bansIssued,
		SuspensionsIssued:  suspensionsIssued,
		PerformanceScore:   performanceScore,
		ActivityTrend:      activityTrend,
	}
	
	logger.Info("Moderator analytics generated successfully", "moderator_id", moderatorID, "period", period)
	return analytics, nil
}

// GetOverallAnalytics generates comprehensive moderation analytics for the specified period
func (mas *ModerationAnalyticsService) GetOverallAnalytics(ctx context.Context, period string) (*OverallAnalytics, error) {
	logger.Info("Generating overall moderation analytics", "period", period)
	
	// Get individual analytics components
	reports, err := mas.GetReportAnalytics(ctx, period)
	if err != nil {
		logger.Error("Failed to get report analytics", err)
		return nil, fmt.Errorf("failed to get report analytics: %w", err)
	}
	
	bans, err := mas.GetBanAnalytics(ctx, period)
	if err != nil {
		logger.Error("Failed to get ban analytics", err)
		return nil, fmt.Errorf("failed to get ban analytics: %w", err)
	}
	
	appeals, err := mas.GetAppealAnalytics(ctx, period)
	if err != nil {
		logger.Error("Failed to get appeal analytics", err)
		return nil, fmt.Errorf("failed to get appeal analytics: %w", err)
	}
	
	content, err := mas.GetContentAnalytics(ctx, period)
	if err != nil {
		logger.Error("Failed to get content analytics", err)
		return nil, fmt.Errorf("failed to get content analytics: %w", err)
	}
	
	// Get moderator analytics
	moderators, err := mas.getAllModeratorAnalytics(ctx, period)
	if err != nil {
		logger.Error("Failed to get moderator analytics", err)
		return nil, fmt.Errorf("failed to get moderator analytics: %w", err)
	}
	
	// Get user reputation distribution
	userReputation, err := mas.getUserReputationDistribution(ctx)
	if err != nil {
		logger.Error("Failed to get user reputation distribution", err)
		return nil, fmt.Errorf("failed to get user reputation distribution: %w", err)
	}
	
	// Get system health statistics
	systemHealth, err := mas.getSystemHealthStats(ctx)
	if err != nil {
		logger.Error("Failed to get system health stats", err)
		return nil, fmt.Errorf("failed to get system health stats: %w", err)
	}
	
	analytics := &OverallAnalytics{
		Period:         period,
		GeneratedAt:     time.Now(),
		Reports:         *reports,
		Bans:            *bans,
		Appeals:         *appeals,
		Content:         *content,
		Moderators:      moderators,
		UserReputation:   userReputation,
		SystemHealth:     *systemHealth,
	}
	
	logger.Info("Overall moderation analytics generated successfully", "period", period)
	return analytics, nil
}

// Helper methods

func (mas *ModerationAnalyticsService) getDateRange(period string) (time.Time, time.Time) {
	now := time.Now()
	
	switch period {
	case "1d":
		return now.AddDate(0, 0, -1), now
	case "7d":
		return now.AddDate(0, 0, -7), now
	case "30d":
		return now.AddDate(0, 0, -30), now
	case "90d":
		return now.AddDate(0, 0, -90), now
	default:
		return now.AddDate(0, 0, -7), now // Default to 7 days
	}
}

func (mas *ModerationAnalyticsService) getReportsTrend(ctx context.Context, period string) ([]TrendData, error) {
	// Implementation would query reports grouped by day/week/month based on period
	// For now, return empty slice
	return []TrendData{}, nil
}

func (mas *ModerationAnalyticsService) getBansTrend(ctx context.Context, period string) ([]TrendData, error) {
	// Implementation would query bans grouped by day/week/month based on period
	// For now, return empty slice
	return []TrendData{}, nil
}

func (mas *ModerationAnalyticsService) getAppealsTrend(ctx context.Context, period string) ([]TrendData, error) {
	// Implementation would query appeals grouped by day/week/month based on period
	// For now, return empty slice
	return []TrendData{}, nil
}

func (mas *ModerationAnalyticsService) getContentTrend(ctx context.Context, period string) ([]TrendData, error) {
	// Implementation would query content analyses grouped by day/week/month based on period
	// For now, return empty slice
	return []TrendData{}, nil
}

func (mas *ModerationAnalyticsService) getModeratorActivityTrend(ctx context.Context, moderatorID uuid.UUID, period string) ([]TrendData, error) {
	// Implementation would query moderator actions grouped by day/week/month based on period
	// For now, return empty slice
	return []TrendData{}, nil
}

func (mas *ModerationAnalyticsService) getAverageResolutionTime(ctx context.Context, startDate, endDate time.Time) (time.Duration, error) {
	// Implementation would calculate average time from report creation to resolution
	// For now, return 0
	return 0, nil
}

func (mas *ModerationAnalyticsService) getAverageBanDuration(ctx context.Context, startDate, endDate time.Time) (time.Duration, error) {
	// Implementation would calculate average ban duration
	// For now, return 0
	return 0, nil
}

func (mas *ModerationAnalyticsService) getAverageAppealReviewTime(ctx context.Context, startDate, endDate time.Time) (time.Duration, error) {
	// Implementation would calculate average time from appeal creation to review
	// For now, return 0
	return 0, nil
}

func (mas *ModerationAnalyticsService) getAverageModeratorReviewTime(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (time.Duration, error) {
	// Implementation would calculate average time for moderator to review reports
	// For now, return 0
	return 0, nil
}

func (mas *ModerationAnalyticsService) getAverageConfidence(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	// Implementation would calculate average confidence score for content analyses
	// For now, return 0
	return 0, nil
}

func (mas *ModerationAnalyticsService) getTopReportedUsers(ctx context.Context, startDate, endDate time.Time, limit int) ([]UserReportStats, error) {
	// Implementation would get users with most reports
	// For now, return empty slice
	return []UserReportStats{}, nil
}

func (mas *ModerationAnalyticsService) getTopBannedUsers(ctx context.Context, startDate, endDate time.Time, limit int) ([]UserBanStats, error) {
	// Implementation would get users with most bans
	// For now, return empty slice
	return []UserBanStats{}, nil
}

func (mas *ModerationAnalyticsService) getTopAppealingUsers(ctx context.Context, startDate, endDate time.Time, limit int) ([]UserAppealStats, error) {
	// Implementation would get users with most appeals
	// For now, return empty slice
	return []UserAppealStats{}, nil
}

func (mas *ModerationAnalyticsService) getTopViolatingTypes(ctx context.Context, startDate, endDate time.Time, limit int) ([]ContentTypeStats, error) {
	// Implementation would get content types with most violations
	// For now, return empty slice
	return []ContentTypeStats{}, nil
}

func (mas *ModerationAnalyticsService) getAllModeratorAnalytics(ctx context.Context, period string) ([]ModeratorAnalytics, error) {
	// Implementation would get analytics for all moderators
	// For now, return empty slice
	return []ModeratorAnalytics{}, nil
}

func (mas *ModerationAnalyticsService) getUserReputationDistribution(ctx context.Context) (map[string]int, error) {
	// Implementation would get distribution of user reputation levels
	// For now, return empty map
	return map[string]int{}, nil
}

func (mas *ModerationAnalyticsService) getSystemHealthStats(ctx context.Context) (*SystemHealthStats, error) {
	// Implementation would get system health metrics
	// For now, return empty stats
	return &SystemHealthStats{}, nil
}

func (mas *ModerationAnalyticsService) calculateModeratorPerformanceScore(
	reportsReviewed, reportsResolved int,
	averageReviewTime time.Duration,
	appealsReviewed, appealsApproved, bansIssued int,
) float64 {
	// Implementation would calculate performance score based on various metrics
	// For now, return 0
	return 0
}

func getIntFromCache(cache map[string]interface{}, key string) int {
	if value, exists := cache[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return 0
}

// Repository interfaces (these would be defined elsewhere in the codebase)
type ReportRepository interface {
	CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountByStatusAndDateRange(ctx context.Context, status string, startDate, endDate time.Time) (int, error)
	CountByReasonAndDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)
	CountReviewedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
	CountResolvedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
	CountDismissedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type BanRepository interface {
	CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountActiveByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountPermanentByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountTemporaryByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountByReasonAndDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)
	CountByDurationAndDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)
	CountIssuedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
	banRepoCountSuspensionsIssuedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
}

type AppealRepository interface {
	CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountByStatusAndDateRange(ctx context.Context, status string, startDate, endDate time.Time) (int, error)
	CountReviewedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
	CountApprovedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
	CountRejectedByModeratorAndDateRange(ctx context.Context, moderatorID uuid.UUID, startDate, endDate time.Time) (int, error)
}

type ModerationActionRepository interface {
	// Methods for moderation actions
}

type ContentAnalysisRepository interface {
	CountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error)
	CountByResultAndDateRange(ctx context.Context, result string, startDate, endDate time.Time) (int, error)
	CountByTypeAndDateRange(ctx context.Context, startDate, endDate time.Time) (map[string]int, error)
}

type ModerationQueueRepository interface {
	// Methods for moderation queue
}