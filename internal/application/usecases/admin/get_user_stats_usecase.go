package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetUserStatsUseCase handles retrieving user statistics
type GetUserStatsUseCase struct {
	userRepo repositories.UserRepository
}

// NewGetUserStatsUseCase creates a new GetUserStatsUseCase
func NewGetUserStatsUseCase(userRepo repositories.UserRepository) *GetUserStatsUseCase {
	return &GetUserStatsUseCase{
		userRepo: userRepo,
	}
}

// GetUserStatsRequest represents a request to get user statistics
type GetUserStatsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Period     string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	GroupBy    string     `json:"group_by" validate:"required,oneof=day week month"`
}

// UserStatsResponse represents response from getting user statistics
type UserStatsResponse struct {
	Overview    UserStatsOverview    `json:"overview"`
	Growth      UserGrowthStats     `json:"growth"`
	Demographics UserDemographics     `json:"demographics"`
	Activity    UserActivityStats    `json:"activity"`
	GroupedData []UserStatsGroup    `json:"grouped_data"`
	Timestamp   time.Time           `json:"timestamp"`
}

// UserStatsOverview represents high-level user statistics
type UserStatsOverview struct {
	TotalUsers      int64 `json:"total_users"`
	ActiveUsers     int64 `json:"active_users"`
	NewUsers        int64 `json:"new_users"`
	VerifiedUsers   int64 `json:"verified_users"`
	PremiumUsers    int64 `json:"premium_users"`
	BannedUsers     int64 `json:"banned_users"`
	SuspendedUsers  int64 `json:"suspended_users"`
}

// UserGrowthStats represents user growth statistics
type UserGrowthStats struct {
	DailyGrowth   []GrowthData `json:"daily_growth"`
	WeeklyGrowth  []GrowthData `json:"weekly_growth"`
	MonthlyGrowth []GrowthData `json:"monthly_growth"`
	GrowthRate     float64      `json:"growth_rate"`
}

// UserDemographics represents user demographic statistics
type UserDemographics struct {
	AgeDistribution    map[string]int64 `json:"age_distribution"`
	GenderDistribution map[string]int64 `json:"gender_distribution"`
	LocationStats     []LocationStat  `json:"location_stats"`
	DeviceStats       map[string]int64 `json:"device_stats"`
}

// UserActivityStats represents user activity statistics
type UserActivityStats struct {
	DailyActiveUsers  int64   `json:"daily_active_users"`
	WeeklyActiveUsers int64   `json:"weekly_active_users"`
	MonthlyActiveUsers int64 `json:"monthly_active_users"`
	AvgSessionDuration float64 `json:"avg_session_duration"`
	LoginFrequency    []ActivityData `json:"login_frequency"`
}

// UserStatsGroup represents grouped user statistics
type UserStatsGroup struct {
	Period string `json:"period"`
	Date   string `json:"date"`
	Users   UserStatsOverview `json:"users"`
}

// GrowthData represents growth data point
type GrowthData struct {
	Date  string  `json:"date"`
	Value int64   `json:"value"`
	Growth float64 `json:"growth"`
}

// ActivityData represents activity data point
type ActivityData struct {
	Date  string  `json:"date"`
	Value int64   `json:"value"`
}

// Execute retrieves user statistics
func (uc *GetUserStatsUseCase) Execute(ctx context.Context, req GetUserStatsRequest) (*UserStatsResponse, error) {
	logger.Info("GetUserStats use case executed", "admin_id", req.AdminID, "period", req.Period, "group_by", req.GroupBy)

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange(req.Period, req.StartTime, req.EndTime)

	// Get overview statistics
	overview, err := uc.getOverviewStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get overview statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get overview statistics: %w", err)
	}

	// Get growth statistics
	growth, err := uc.getGrowthStats(ctx, startTime, endTime, req.GroupBy)
	if err != nil {
		logger.Error("Failed to get growth statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get growth statistics: %w", err)
	}

	// Get demographic statistics
	demographics, err := uc.getDemographics(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get demographic statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get demographic statistics: %w", err)
	}

	// Get activity statistics
	activity, err := uc.getActivityStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get activity statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get activity statistics: %w", err)
	}

	// Get grouped data
	groupedData, err := uc.getGroupedData(ctx, startTime, endTime, req.GroupBy)
	if err != nil {
		logger.Error("Failed to get grouped data", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get grouped data: %w", err)
	}

	logger.Info("GetUserStats use case completed successfully", "admin_id", req.AdminID)
	return &UserStatsResponse{
		Overview:    *overview,
		Growth:      *growth,
		Demographics: *demographics,
		Activity:    *activity,
		GroupedData: groupedData,
		Timestamp:   time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetUserStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
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
func (uc *GetUserStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time) (*UserStatsOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &UserStatsOverview{
		TotalUsers:     10000,
		ActiveUsers:    7500,
		NewUsers:        600,
		VerifiedUsers:   3000,
		PremiumUsers:    1500,
		BannedUsers:     50,
		SuspendedUsers:  25,
	}, nil
}

func (uc *GetUserStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time, groupBy string) (*UserGrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &UserGrowthStats{
		DailyGrowth: []GrowthData{
			{Date: "2025-11-01", Value: 9800, Growth: 2.1},
			{Date: "2025-11-02", Value: 9900, Growth: 1.0},
			{Date: "2025-11-03", Value: 10000, Growth: 1.0},
		},
		WeeklyGrowth: []GrowthData{
			{Date: "2025-10-30", Value: 9500, Growth: 3.2},
			{Date: "2025-11-06", Value: 10000, Growth: 5.3},
		},
		MonthlyGrowth: []GrowthData{
			{Date: "2025-09", Value: 9200, Growth: 4.5},
			{Date: "2025-10", Value: 9500, Growth: 3.3},
			{Date: "2025-11", Value: 10000, Growth: 5.3},
		},
		GrowthRate: 5.2,
	}, nil
}

func (uc *GetUserStatsUseCase) getDemographics(ctx context.Context, startTime, endTime time.Time) (*UserDemographics, error) {
	// Mock data - in real implementation, this would query the database
	return &UserDemographics{
		AgeDistribution: map[string]int64{
			"18-24": 2000,
			"25-34": 3500,
			"35-44": 2500,
			"45-54": 1500,
			"55+":   500,
		},
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
		DeviceStats: map[string]int64{
			"mobile": 6000,
			"desktop": 3500,
			"tablet": 500,
		},
	}, nil
}

func (uc *GetUserStatsUseCase) getActivityStats(ctx context.Context, startTime, endTime time.Time) (*UserActivityStats, error) {
	// Mock data - in real implementation, this would query the database
	return &UserActivityStats{
		DailyActiveUsers:  3000,
		WeeklyActiveUsers: 6000,
		MonthlyActiveUsers: 7500,
		AvgSessionDuration: 15.5, // minutes
		LoginFrequency: []ActivityData{
			{Date: "2025-11-01", Value: 2500},
			{Date: "2025-11-02", Value: 2600},
			{Date: "2025-11-03", Value: 2700},
		},
	}, nil
}

func (uc *GetUserStatsUseCase) getGroupedData(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]UserStatsGroup, error) {
	// Mock data - in real implementation, this would query the database
	var groupedData []UserStatsGroup

	switch groupBy {
	case "day":
		groupedData = []UserStatsGroup{
			{Period: "day", Date: "2025-11-01", Users: UserStatsOverview{TotalUsers: 9800, ActiveUsers: 7350, NewUsers: 25}},
			{Period: "day", Date: "2025-11-02", Users: UserStatsOverview{TotalUsers: 9900, ActiveUsers: 7425, NewUsers: 30}},
			{Period: "day", Date: "2025-11-03", Users: UserStatsOverview{TotalUsers: 10000, ActiveUsers: 7500, NewUsers: 35}},
		}
	case "week":
		groupedData = []UserStatsGroup{
			{Period: "week", Date: "2025-10-30", Users: UserStatsOverview{TotalUsers: 9500, ActiveUsers: 7125, NewUsers: 150}},
			{Period: "week", Date: "2025-11-06", Users: UserStatsOverview{TotalUsers: 10000, ActiveUsers: 7500, NewUsers: 200}},
		}
	case "month":
		groupedData = []UserStatsGroup{
			{Period: "month", Date: "2025-09", Users: UserStatsOverview{TotalUsers: 9200, ActiveUsers: 6900, NewUsers: 450}},
			{Period: "month", Date: "2025-10", Users: UserStatsOverview{TotalUsers: 9500, ActiveUsers: 7125, NewUsers: 500}},
			{Period: "month", Date: "2025-11", Users: UserStatsOverview{TotalUsers: 10000, ActiveUsers: 7500, NewUsers: 600}},
		}
	}

	return groupedData, nil
}