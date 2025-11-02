package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetPlatformStatsUseCase handles retrieving platform statistics
type GetPlatformStatsUseCase struct {
	userRepo    repositories.UserRepository
	matchRepo   repositories.MatchRepository
	messageRepo repositories.MessageRepository
	paymentRepo repositories.PaymentRepository
	reportRepo  repositories.ReportRepository
}

// NewGetPlatformStatsUseCase creates a new GetPlatformStatsUseCase
func NewGetPlatformStatsUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	messageRepo repositories.MessageRepository,
	paymentRepo repositories.PaymentRepository,
	reportRepo repositories.ReportRepository,
) *GetPlatformStatsUseCase {
	return &GetPlatformStatsUseCase{
		userRepo:    userRepo,
		matchRepo:   matchRepo,
		messageRepo: messageRepo,
		paymentRepo: paymentRepo,
		reportRepo:  reportRepo,
	}
}

// GetPlatformStatsRequest represents a request to get platform statistics
type GetPlatformStatsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Period     string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
}

// GetDashboardDataRequest represents a request to get dashboard data
type GetDashboardDataRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	Period  string    `json:"period" validate:"required,oneof=1d 7d 30d 90d"`
}

// ExportAnalyticsRequest represents a request to export analytics
type ExportAnalyticsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Type      string     `json:"type" validate:"required,oneof=users matches messages payments verification"`
	Format    string     `json:"format" validate:"required,oneof=csv xlsx json"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

// ExportAnalyticsResponse represents the response from exporting analytics
type ExportAnalyticsResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// PlatformStats represents platform-wide statistics
type PlatformStats struct {
	Overview    PlatformOverview    `json:"overview"`
	Users       UserStats          `json:"users"`
	Matches     MatchStats         `json:"matches"`
	Messages    MessageStats       `json:"messages"`
	Payments    PaymentStats       `json:"payments"`
	Reports     ReportStats        `json:"reports"`
	Growth      GrowthStats        `json:"growth"`
	Engagement  EngagementStats    `json:"engagement"`
	Timestamp   time.Time          `json:"timestamp"`
}

// PlatformOverview represents high-level platform overview
type PlatformOverview struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	NewUsersToday    int64 `json:"new_users_today"`
	NewUsersWeek     int64 `json:"new_users_week"`
	NewUsersMonth    int64 `json:"new_users_month"`
	TotalMatches     int64 `json:"total_matches"`
	MatchesToday     int64 `json:"matches_today"`
	MatchesWeek      int64 `json:"matches_week"`
	MatchesMonth     int64 `json:"matches_month"`
	TotalMessages    int64 `json:"total_messages"`
	MessagesToday    int64 `json:"messages_today"`
	MessagesWeek     int64 `json:"messages_week"`
	MessagesMonth    int64 `json:"messages_month"`
	TotalRevenue     float64 `json:"total_revenue"`
	RevenueToday     float64 `json:"revenue_today"`
	RevenueWeek      float64 `json:"revenue_week"`
	RevenueMonth     float64 `json:"revenue_month"`
	PendingReports   int64 `json:"pending_reports"`
	SystemHealth     string `json:"system_health"`
}

// UserStats represents user-related statistics
type UserStats struct {
	TotalUsers      int64   `json:"total_users"`
	ActiveUsers     int64   `json:"active_users"`
	VerifiedUsers   int64   `json:"verified_users"`
	PremiumUsers    int64   `json:"premium_users"`
	BannedUsers     int64   `json:"banned_users"`
	NewUsers        int64   `json:"new_users"`
	UserGrowth      float64 `json:"user_growth"`
	AvgAge          float64 `json:"avg_age"`
	GenderDistribution map[string]int64 `json:"gender_distribution"`
	LocationStats    []LocationStat `json:"location_stats"`
}

// MatchStats represents match-related statistics
type MatchStats struct {
	TotalMatches    int64   `json:"total_matches"`
	NewMatches      int64   `json:"new_matches"`
	MatchGrowth     float64 `json:"match_growth"`
	AvgMatchesPerUser float64 `json:"avg_matches_per_user"`
	MatchRate       float64 `json:"match_rate"`
	DailyMatches    []DailyStat `json:"daily_matches"`
}

// MessageStats represents message-related statistics
type MessageStats struct {
	TotalMessages     int64   `json:"total_messages"`
	NewMessages       int64   `json:"new_messages"`
	MessageGrowth     float64 `json:"message_growth"`
	AvgMessagesPerUser float64 `json:"avg_messages_per_user"`
	DailyMessages     []DailyStat `json:"daily_messages"`
}

// PaymentStats represents payment-related statistics
type PaymentStats struct {
	TotalRevenue     float64        `json:"total_revenue"`
	NewRevenue       float64        `json:"new_revenue"`
	RevenueGrowth    float64        `json:"revenue_growth"`
	TotalSubscriptions int64          `json:"total_subscriptions"`
	NewSubscriptions  int64           `json:"new_subscriptions"`
	SubscriptionGrowth float64         `json:"subscription_growth"`
	RevenueByPlan    map[string]float64 `json:"revenue_by_plan"`
	DailyRevenue     []DailyStat     `json:"daily_revenue"`
}

// ReportStats represents report-related statistics
type ReportStats struct {
	TotalReports     int64   `json:"total_reports"`
	PendingReports   int64   `json:"pending_reports"`
	ResolvedReports   int64   `json:"resolved_reports"`
	DismissedReports  int64   `json:"dismissed_reports"`
	ReportGrowth     float64 `json:"report_growth"`
	ReportsByType    map[string]int64 `json:"reports_by_type"`
	DailyReports     []DailyStat `json:"daily_reports"`
}

// GrowthStats represents growth statistics
type GrowthStats struct {
	UserGrowth      float64 `json:"user_growth"`
	MatchGrowth     float64 `json:"match_growth"`
	MessageGrowth    float64 `json:"message_growth"`
	RevenueGrowth    float64 `json:"revenue_growth"`
	SubscriptionGrowth float64 `json:"subscription_growth"`
	MonthlyGrowth   []MonthlyGrowth `json:"monthly_growth"`
}

// EngagementStats represents engagement statistics
type EngagementStats struct {
	DailyActiveUsers  int64   `json:"daily_active_users"`
	WeeklyActiveUsers int64   `json:"weekly_active_users"`
	MonthlyActiveUsers int64 `json:"monthly_active_users"`
	AvgSessionDuration float64 `json:"avg_session_duration"`
	AvgMessagesPerMatch float64 `json:"avg_messages_per_match"`
	MatchRate         float64 `json:"match_rate"`
	ResponseRate      float64 `json:"response_rate"`
}

// LocationStat represents location-based statistics
type LocationStat struct {
	Country    string  `json:"country"`
	City       string  `json:"city"`
	UserCount  int64   `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

// DailyStat represents daily statistics
type DailyStat struct {
	Date  string  `json:"date"`
	Value int64   `json:"value"`
}

// MonthlyGrowth represents monthly growth statistics
type MonthlyGrowth struct {
	Month      string  `json:"month"`
	UserGrowth float64 `json:"user_growth"`
	MatchGrowth float64 `json:"match_growth"`
	RevenueGrowth float64 `json:"revenue_growth"`
}

// Execute retrieves platform statistics
func (uc *GetPlatformStatsUseCase) Execute(ctx context.Context, req GetPlatformStatsRequest) (*PlatformStats, error) {
	logger.Info("GetPlatformStats use case executed", "admin_id", req.AdminID, "period", req.Period)

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange(req.Period, req.StartTime, req.EndTime)

	// Get overview statistics
	overview, err := uc.getOverviewStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get overview statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get overview statistics: %w", err)
	}

	// Get user statistics
	userStats, err := uc.getUserStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get user statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	// Get match statistics
	matchStats, err := uc.getMatchStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get match statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get match statistics: %w", err)
	}

	// Get message statistics
	messageStats, err := uc.getMessageStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get message statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get message statistics: %w", err)
	}

	// Get payment statistics
	paymentStats, err := uc.getPaymentStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get payment statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get payment statistics: %w", err)
	}

	// Get report statistics
	reportStats, err := uc.getReportStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get report statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get report statistics: %w", err)
	}

	// Get growth statistics
	growthStats, err := uc.getGrowthStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get growth statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get growth statistics: %w", err)
	}

	// Get engagement statistics
	engagementStats, err := uc.getEngagementStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get engagement statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get engagement statistics: %w", err)
	}

	logger.Info("GetPlatformStats use case completed successfully", "admin_id", req.AdminID)
	return &PlatformStats{
		Overview:   *overview,
		Users:      *userStats,
		Matches:    *matchStats,
		Messages:   *messageStats,
		Payments:   *paymentStats,
		Reports:    *reportStats,
		Growth:     *growthStats,
		Engagement: *engagementStats,
		Timestamp:  time.Now(),
	}, nil
}

// GetDashboardData retrieves dashboard data
func (uc *GetPlatformStatsUseCase) GetDashboardData(ctx context.Context, req GetDashboardDataRequest) (*PlatformStats, error) {
	logger.Info("GetDashboardData use case executed", "admin_id", req.AdminID, "period", req.Period)

	// Convert to platform stats request
	platformReq := GetPlatformStatsRequest{
		AdminID: req.AdminID,
		Period:  req.Period,
	}

	// Execute platform stats
	return uc.Execute(ctx, platformReq)
}

// ExportAnalytics exports analytics data
func (uc *GetPlatformStatsUseCase) ExportAnalytics(ctx context.Context, req ExportAnalyticsRequest) (*ExportAnalyticsResponse, error) {
	logger.Info("ExportAnalytics use case executed", "admin_id", req.AdminID, "type", req.Type, "format", req.Format)

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange("30d", req.StartTime, req.EndTime)

	// Get data based on type
	var data interface{}
	var filename string
	var contentType string

	switch req.Type {
	case "users":
		data, err := uc.getUserStats(ctx, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get user data: %w", err)
		}
		filename = "users_export"
	case "matches":
		data, err := uc.getMatchStats(ctx, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get match data: %w", err)
		}
		filename = "matches_export"
	case "messages":
		data, err := uc.getMessageStats(ctx, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get message data: %w", err)
		}
		filename = "messages_export"
	case "payments":
		data, err := uc.getPaymentStats(ctx, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment data: %w", err)
		}
		filename = "payments_export"
	case "verification":
		data, err := uc.getVerificationStats(ctx, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get verification data: %w", err)
		}
		filename = "verification_export"
	default:
		return nil, fmt.Errorf("invalid export type: %s", req.Type)
	}

	// Set content type based on format
	switch req.Format {
	case "csv":
		contentType = "text/csv"
		filename += ".csv"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename += ".xlsx"
	case "json":
		contentType = "application/json"
		filename += ".json"
	default:
		return nil, fmt.Errorf("invalid export format: %s", req.Format)
	}

	// Convert data to requested format (simplified for demo)
	exportData := []byte(fmt.Sprintf(`{"data": %v, "exported_at": "%s"}`, data, time.Now().Format(time.RFC3339)))

	logger.Info("ExportAnalytics use case completed successfully", "admin_id", req.AdminID, "type", req.Type, "format", req.Format)
	return &ExportAnalyticsResponse{
		Filename:    filename,
		ContentType: contentType,
		Data:        exportData,
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetPlatformStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
	now := time.Now()

	// If custom time range is provided, use it
	if startTime != nil && endTime != nil {
		return *startTime, *endTime
	}

	// Calculate based on period
	var start time.Time
	switch period {
	case "1d":
		start = now.AddDate(0, 0, -1)
	case "7d":
		start = now.AddDate(0, 0, -7)
	case "30d":
		start = now.AddDate(0, 0, -30)
	case "90d":
		start = now.AddDate(0, 0, -90)
	case "1y":
		start = now.AddDate(-1, 0, 0)
	default:
		start = now.AddDate(0, 0, -7) // Default to 7 days
	}

	return start, now
}

// Mock implementations for statistics (in real implementation, these would query the database)
func (uc *GetPlatformStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time) (*PlatformOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &PlatformOverview{
		TotalUsers:     10000,
		ActiveUsers:    7500,
		NewUsersToday:   25,
		NewUsersWeek:    150,
		NewUsersMonth:   600,
		TotalMatches:    50000,
		MatchesToday:    120,
		MatchesWeek:     800,
		MatchesMonth:    3200,
		TotalMessages:   250000,
		MessagesToday:   500,
		MessagesWeek:    3500,
		MessagesMonth:   14000,
		TotalRevenue:    50000.0,
		RevenueToday:    150.0,
		RevenueWeek:     1050.0,
		RevenueMonth:    4500.0,
		PendingReports:  15,
		SystemHealth:    "healthy",
	}, nil
}

func (uc *GetPlatformStatsUseCase) getUserStats(ctx context.Context, startTime, endTime time.Time) (*UserStats, error) {
	// Mock data - in real implementation, this would query the database
	return &UserStats{
		TotalUsers:   10000,
		ActiveUsers:  7500,
		VerifiedUsers: 3000,
		PremiumUsers:  1500,
		BannedUsers:   50,
		NewUsers:     600,
		UserGrowth:   5.2,
		AvgAge:       28.5,
		GenderDistribution: map[string]int64{
			"male":   5500,
			"female": 4300,
			"other":  200,
		},
		LocationStats: []LocationStat{
			{Country: "US", City: "New York", UserCount: 2000, Percentage: 20.0},
			{Country: "US", City: "Los Angeles", UserCount: 1500, Percentage: 15.0},
			{Country: "UK", City: "London", UserCount: 1200, Percentage: 12.0},
			{Country: "CA", City: "Toronto", UserCount: 800, Percentage: 8.0},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getMatchStats(ctx context.Context, startTime, endTime time.Time) (*MatchStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MatchStats{
		TotalMatches:      50000,
		NewMatches:        3200,
		MatchGrowth:       8.5,
		AvgMatchesPerUser: 5.0,
		MatchRate:         15.2,
		DailyMatches: []DailyStat{
			{Date: "2025-11-01", Value: 100},
			{Date: "2025-11-02", Value: 110},
			{Date: "2025-11-03", Value: 105},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getMessageStats(ctx context.Context, startTime, endTime time.Time) (*MessageStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MessageStats{
		TotalMessages:      250000,
		NewMessages:        14000,
		MessageGrowth:      12.3,
		AvgMessagesPerUser: 25.0,
		DailyMessages: []DailyStat{
			{Date: "2025-11-01", Value: 450},
			{Date: "2025-11-02", Value: 480},
			{Date: "2025-11-03", Value: 470},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getPaymentStats(ctx context.Context, startTime, endTime time.Time) (*PaymentStats, error) {
	// Mock data - in real implementation, this would query the database
	return &PaymentStats{
		TotalRevenue:     50000.0,
		NewRevenue:       4500.0,
		RevenueGrowth:    18.5,
		TotalSubscriptions: 1500,
		NewSubscriptions:  120,
		SubscriptionGrowth: 15.2,
		RevenueByPlan: map[string]float64{
			"basic":    5000.0,
			"premium":  35000.0,
			"platinum": 10000.0,
		},
		DailyRevenue: []DailyStat{
			{Date: "2025-11-01", Value: 150},
			{Date: "2025-11-02", Value: 160},
			{Date: "2025-11-03", Value: 155},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getReportStats(ctx context.Context, startTime, endTime time.Time) (*ReportStats, error) {
	// Mock data - in real implementation, this would query the database
	return &ReportStats{
		TotalReports:    500,
		PendingReports:  15,
		ResolvedReports:  450,
		DismissedReports: 35,
		ReportGrowth:     5.8,
		ReportsByType: map[string]int64{
			"inappropriate_behavior": 200,
			"fake_profile":          150,
			"spam":                 100,
			"harassment":           50,
		},
		DailyReports: []DailyStat{
			{Date: "2025-11-01", Value: 5},
			{Date: "2025-11-02", Value: 3},
			{Date: "2025-11-03", Value: 7},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getVerificationStats(ctx context.Context, startTime, endTime time.Time) (interface{}, error) {
	// Mock data - in real implementation, this would query the database
	return map[string]interface{}{
		"total_verifications": 3000,
		"pending_verifications": 50,
		"approved_verifications": 2800,
		"rejected_verifications": 150,
		"verification_rate": 93.3,
		"verification_by_type": map[string]int64{
			"selfie": 2000,
			"document": 1000,
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time) (*GrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &GrowthStats{
		UserGrowth:       5.2,
		MatchGrowth:      8.5,
		MessageGrowth:    12.3,
		RevenueGrowth:     18.5,
		SubscriptionGrowth: 15.2,
		MonthlyGrowth: []MonthlyGrowth{
			{Month: "2025-08", UserGrowth: 4.5, MatchGrowth: 7.2, RevenueGrowth: 15.0},
			{Month: "2025-09", UserGrowth: 5.0, MatchGrowth: 8.0, RevenueGrowth: 17.0},
			{Month: "2025-10", UserGrowth: 5.5, MatchGrowth: 9.0, RevenueGrowth: 19.0},
		},
	}, nil
}

func (uc *GetPlatformStatsUseCase) getEngagementStats(ctx context.Context, startTime, endTime time.Time) (*EngagementStats, error) {
	// Mock data - in real implementation, this would query the database
	return &EngagementStats{
		DailyActiveUsers:     3000,
		WeeklyActiveUsers:    6000,
		MonthlyActiveUsers:   7500,
		AvgSessionDuration:    15.5, // minutes
		AvgMessagesPerMatch:  8.2,
		MatchRate:           15.2, // percentage
		ResponseRate:        75.5, // percentage
	}, nil
}