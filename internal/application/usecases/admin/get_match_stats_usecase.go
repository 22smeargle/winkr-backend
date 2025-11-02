package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetMatchStatsUseCase handles retrieving match statistics
type GetMatchStatsUseCase struct {
	matchRepo repositories.MatchRepository
}

// NewGetMatchStatsUseCase creates a new GetMatchStatsUseCase
func NewGetMatchStatsUseCase(matchRepo repositories.MatchRepository) *GetMatchStatsUseCase {
	return &GetMatchStatsUseCase{
		matchRepo: matchRepo,
	}
}

// GetMatchStatsRequest represents a request to get match statistics
type GetMatchStatsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Period     string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	GroupBy    string     `json:"group_by" validate:"required,oneof=day week month"`
}

// MatchStatsResponse represents response from getting match statistics
type MatchStatsResponse struct {
	Overview    MatchStatsOverview    `json:"overview"`
	Growth      MatchGrowthStats     `json:"growth"`
	Conversion  MatchConversionStats  `json:"conversion"`
	GroupedData []MatchStatsGroup    `json:"grouped_data"`
	Timestamp   time.Time           `json:"timestamp"`
}

// MatchStatsOverview represents high-level match statistics
type MatchStatsOverview struct {
	TotalMatches      int64 `json:"total_matches"`
	NewMatches        int64 `json:"new_matches"`
	ActiveMatches     int64 `json:"active_matches"`
	CompletedMatches  int64 `json:"completed_matches"`
	AvgMatchesPerUser float64 `json:"avg_matches_per_user"`
	MatchRate         float64 `json:"match_rate"`
}

// MatchGrowthStats represents match growth statistics
type MatchGrowthStats struct {
	DailyGrowth   []GrowthData `json:"daily_growth"`
	WeeklyGrowth  []GrowthData `json:"weekly_growth"`
	MonthlyGrowth []GrowthData `json:"monthly_growth"`
	GrowthRate     float64      `json:"growth_rate"`
}

// MatchConversionStats represents match conversion statistics
type MatchConversionStats struct {
	SwipeToMatchRate   float64 `json:"swipe_to_match_rate"`
	MatchToMessageRate float64 `json:"match_to_message_rate"`
	MatchToDateRate    float64 `json:"match_to_date_rate"`
	AvgTimeToMatch     float64 `json:"avg_time_to_match"` // in hours
	AvgTimeToMessage   float64 `json:"avg_time_to_message"` // in hours
}

// MatchStatsGroup represents grouped match statistics
type MatchStatsGroup struct {
	Period string           `json:"period"`
	Date   string           `json:"date"`
	Matches MatchStatsOverview `json:"matches"`
}

// Execute retrieves match statistics
func (uc *GetMatchStatsUseCase) Execute(ctx context.Context, req GetMatchStatsRequest) (*MatchStatsResponse, error) {
	logger.Info("GetMatchStats use case executed", "admin_id", req.AdminID, "period", req.Period, "group_by", req.GroupBy)

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

	// Get conversion statistics
	conversion, err := uc.getConversionStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get conversion statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get conversion statistics: %w", err)
	}

	// Get grouped data
	groupedData, err := uc.getGroupedData(ctx, startTime, endTime, req.GroupBy)
	if err != nil {
		logger.Error("Failed to get grouped data", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get grouped data: %w", err)
	}

	logger.Info("GetMatchStats use case completed successfully", "admin_id", req.AdminID)
	return &MatchStatsResponse{
		Overview:    *overview,
		Growth:      *growth,
		Conversion:  *conversion,
		GroupedData: groupedData,
		Timestamp:   time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetMatchStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
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
func (uc *GetMatchStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time) (*MatchStatsOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &MatchStatsOverview{
		TotalMatches:      50000,
		NewMatches:        3200,
		ActiveMatches:     45000,
		CompletedMatches:  42000,
		AvgMatchesPerUser: 5.0,
		MatchRate:         15.2,
	}, nil
}

func (uc *GetMatchStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time, groupBy string) (*MatchGrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MatchGrowthStats{
		DailyGrowth: []GrowthData{
			{Date: "2025-11-01", Value: 4800, Growth: 3.2},
			{Date: "2025-11-02", Value: 4900, Growth: 2.1},
			{Date: "2025-11-03", Value: 5000, Growth: 2.0},
		},
		WeeklyGrowth: []GrowthData{
			{Date: "2025-10-30", Value: 4700, Growth: 4.5},
			{Date: "2025-11-06", Value: 5000, Growth: 6.4},
		},
		MonthlyGrowth: []GrowthData{
			{Date: "2025-09", Value: 4500, Growth: 5.5},
			{Date: "2025-10", Value: 4700, Growth: 4.4},
			{Date: "2025-11", Value: 5000, Growth: 6.4},
		},
		GrowthRate: 8.5,
	}, nil
}

func (uc *GetMatchStatsUseCase) getConversionStats(ctx context.Context, startTime, endTime time.Time) (*MatchConversionStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MatchConversionStats{
		SwipeToMatchRate:   15.2,
		MatchToMessageRate: 75.5,
		MatchToDateRate:    25.3,
		AvgTimeToMatch:     2.5, // hours
		AvgTimeToMessage:   4.2, // hours
	}, nil
}

func (uc *GetMatchStatsUseCase) getGroupedData(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]MatchStatsGroup, error) {
	// Mock data - in real implementation, this would query the database
	var groupedData []MatchStatsGroup

	switch groupBy {
	case "day":
		groupedData = []MatchStatsGroup{
			{Period: "day", Date: "2025-11-01", Matches: MatchStatsOverview{TotalMatches: 4800, NewMatches: 150, MatchRate: 14.8}},
			{Period: "day", Date: "2025-11-02", Matches: MatchStatsOverview{TotalMatches: 4900, NewMatches: 160, MatchRate: 15.0}},
			{Period: "day", Date: "2025-11-03", Matches: MatchStatsOverview{TotalMatches: 5000, NewMatches: 170, MatchRate: 15.5}},
		}
	case "week":
		groupedData = []MatchStatsGroup{
			{Period: "week", Date: "2025-10-30", Matches: MatchStatsOverview{TotalMatches: 4700, NewMatches: 800, MatchRate: 14.5}},
			{Period: "week", Date: "2025-11-06", Matches: MatchStatsOverview{TotalMatches: 5000, NewMatches: 900, MatchRate: 16.0}},
		}
	case "month":
		groupedData = []MatchStatsGroup{
			{Period: "month", Date: "2025-09", Matches: MatchStatsOverview{TotalMatches: 4500, NewMatches: 2400, MatchRate: 14.0}},
			{Period: "month", Date: "2025-10", Matches: MatchStatsOverview{TotalMatches: 4700, NewMatches: 2600, MatchRate: 14.5}},
			{Period: "month", Date: "2025-11", Matches: MatchStatsOverview{TotalMatches: 5000, NewMatches: 3200, MatchRate: 16.0}},
		}
	}

	return groupedData, nil
}