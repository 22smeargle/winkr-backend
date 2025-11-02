package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetVerificationStatsUseCase handles retrieving verification statistics
type GetVerificationStatsUseCase struct {
	verificationRepo repositories.VerificationRepository
}

// NewGetVerificationStatsUseCase creates a new GetVerificationStatsUseCase
func NewGetVerificationStatsUseCase(verificationRepo repositories.VerificationRepository) *GetVerificationStatsUseCase {
	return &GetVerificationStatsUseCase{
		verificationRepo: verificationRepo,
	}
}

// GetVerificationStatsRequest represents a request to get verification statistics
type GetVerificationStatsRequest struct {
	AdminID          uuid.UUID  `json:"admin_id" validate:"required"`
	Period            string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	GroupBy           string     `json:"group_by" validate:"required,oneof=day week month"`
	VerificationType  string     `json:"verification_type" validate:"omitempty,oneof=selfie document all"`
}

// VerificationStatsResponse represents response from getting verification statistics
type VerificationStatsResponse struct {
	Overview    VerificationStatsOverview    `json:"overview"`
	Growth      VerificationGrowthStats     `json:"growth"`
	Performance VerificationPerformanceStats  `json:"performance"`
	GroupedData []VerificationStatsGroup    `json:"grouped_data"`
	Timestamp   time.Time                  `json:"timestamp"`
}

// VerificationStatsOverview represents high-level verification statistics
type VerificationStatsOverview struct {
	TotalVerifications     int64   `json:"total_verifications"`
	PendingVerifications   int64   `json:"pending_verifications"`
	ApprovedVerifications  int64   `json:"approved_verifications"`
	RejectedVerifications  int64   `json:"rejected_verifications"`
	VerificationRate      float64 `json:"verification_rate"`
	AvgProcessingTime     float64 `json:"avg_processing_time"` // in hours
}

// VerificationGrowthStats represents verification growth statistics
type VerificationGrowthStats struct {
	DailyGrowth   []GrowthData `json:"daily_growth"`
	WeeklyGrowth  []GrowthData `json:"weekly_growth"`
	MonthlyGrowth []GrowthData `json:"monthly_growth"`
	GrowthRate     float64      `json:"growth_rate"`
}

// VerificationPerformanceStats represents verification performance statistics
type VerificationPerformanceStats struct {
	ApprovalRate        float64 `json:"approval_rate"`
	RejectionRate        float64 `json:"rejection_rate"`
	AvgProcessingTime    float64 `json:"avg_processing_time"` // in hours
	ProcessingByType     map[string]ProcessingStats `json:"processing_by_type"`
	RejectionReasons     map[string]int64 `json:"rejection_reasons"`
}

// ProcessingStats represents processing statistics by verification type
type ProcessingStats struct {
	TotalCount      int64   `json:"total_count"`
	ApprovedCount   int64   `json:"approved_count"`
	RejectedCount   int64   `json:"rejected_count"`
	AvgProcessTime  float64 `json:"avg_process_time"` // in hours
}

// VerificationStatsGroup represents grouped verification statistics
type VerificationStatsGroup struct {
	Period       string                     `json:"period"`
	Date         string                     `json:"date"`
	Verifications VerificationStatsOverview `json:"verifications"`
}

// Execute retrieves verification statistics
func (uc *GetVerificationStatsUseCase) Execute(ctx context.Context, req GetVerificationStatsRequest) (*VerificationStatsResponse, error) {
	logger.Info("GetVerificationStats use case executed", "admin_id", req.AdminID, "period", req.Period, "group_by", req.GroupBy, "verification_type", req.VerificationType)

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange(req.Period, req.StartTime, req.EndTime)

	// Get overview statistics
	overview, err := uc.getOverviewStats(ctx, startTime, endTime, req.VerificationType)
	if err != nil {
		logger.Error("Failed to get overview statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get overview statistics: %w", err)
	}

	// Get growth statistics
	growth, err := uc.getGrowthStats(ctx, startTime, endTime, req.GroupBy, req.VerificationType)
	if err != nil {
		logger.Error("Failed to get growth statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get growth statistics: %w", err)
	}

	// Get performance statistics
	performance, err := uc.getPerformanceStats(ctx, startTime, endTime, req.VerificationType)
	if err != nil {
		logger.Error("Failed to get performance statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get performance statistics: %w", err)
	}

	// Get grouped data
	groupedData, err := uc.getGroupedData(ctx, startTime, endTime, req.GroupBy, req.VerificationType)
	if err != nil {
		logger.Error("Failed to get grouped data", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get grouped data: %w", err)
	}

	logger.Info("GetVerificationStats use case completed successfully", "admin_id", req.AdminID)
	return &VerificationStatsResponse{
		Overview:    *overview,
		Growth:      *growth,
		Performance:  *performance,
		GroupedData: groupedData,
		Timestamp:   time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetVerificationStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
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
func (uc *GetVerificationStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time, verificationType string) (*VerificationStatsOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &VerificationStatsOverview{
		TotalVerifications:     3000,
		PendingVerifications:   50,
		ApprovedVerifications:  2800,
		RejectedVerifications:  150,
		VerificationRate:      93.3,
		AvgProcessingTime:     2.5, // hours
	}, nil
}

func (uc *GetVerificationStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time, groupBy string, verificationType string) (*VerificationGrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &VerificationGrowthStats{
		DailyGrowth: []GrowthData{
			{Date: "2025-11-01", Value: 45, Growth: 12.5},
			{Date: "2025-11-02", Value: 48, Growth: 10.5},
			{Date: "2025-11-03", Value: 47, Growth: 8.9},
		},
		WeeklyGrowth: []GrowthData{
			{Date: "2025-10-30", Value: 310, Growth: 15.2},
			{Date: "2025-11-06", Value: 350, Growth: 18.5},
		},
		MonthlyGrowth: []GrowthData{
			{Date: "2025-09", Value: 1200, Growth: 14.5},
			{Date: "2025-10", Value: 1350, Growth: 16.2},
			{Date: "2025-11", Value: 1400, Growth: 18.5},
		},
		GrowthRate: 12.5,
	}, nil
}

func (uc *GetVerificationStatsUseCase) getPerformanceStats(ctx context.Context, startTime, endTime time.Time, verificationType string) (*VerificationPerformanceStats, error) {
	// Mock data - in real implementation, this would query the database
	return &VerificationPerformanceStats{
		ApprovalRate:     93.3,
		RejectionRate:     6.7,
		AvgProcessingTime:  2.5, // hours
		ProcessingByType: map[string]ProcessingStats{
			"selfie": ProcessingStats{
				TotalCount:     2000,
				ApprovedCount:   1900,
				RejectedCount:   100,
				AvgProcessTime:  2.0, // hours
			},
			"document": ProcessingStats{
				TotalCount:     1000,
				ApprovedCount:   900,
				RejectedCount:   100,
				AvgProcessTime:  3.0, // hours
			},
		},
		RejectionReasons: map[string]int64{
			"blurry_photo":     50,
			"wrong_document":   30,
			"face_mismatch":    25,
			"expired_document": 20,
			"other":            25,
		},
	}, nil
}

func (uc *GetVerificationStatsUseCase) getGroupedData(ctx context.Context, startTime, endTime time.Time, groupBy string, verificationType string) ([]VerificationStatsGroup, error) {
	// Mock data - in real implementation, this would query the database
	var groupedData []VerificationStatsGroup

	switch groupBy {
	case "day":
		groupedData = []VerificationStatsGroup{
			{Period: "day", Date: "2025-11-01", Verifications: VerificationStatsOverview{TotalVerifications: 45, ApprovedVerifications: 42, RejectedVerifications: 3}},
			{Period: "day", Date: "2025-11-02", Verifications: VerificationStatsOverview{TotalVerifications: 48, ApprovedVerifications: 45, RejectedVerifications: 3}},
			{Period: "day", Date: "2025-11-03", Verifications: VerificationStatsOverview{TotalVerifications: 47, ApprovedVerifications: 44, RejectedVerifications: 3}},
		}
	case "week":
		groupedData = []VerificationStatsGroup{
			{Period: "week", Date: "2025-10-30", Verifications: VerificationStatsOverview{TotalVerifications: 310, ApprovedVerifications: 290, RejectedVerifications: 20}},
			{Period: "week", Date: "2025-11-06", Verifications: VerificationStatsOverview{TotalVerifications: 350, ApprovedVerifications: 325, RejectedVerifications: 25}},
		}
	case "month":
		groupedData = []VerificationStatsGroup{
			{Period: "month", Date: "2025-09", Verifications: VerificationStatsOverview{TotalVerifications: 1200, ApprovedVerifications: 1120, RejectedVerifications: 80}},
			{Period: "month", Date: "2025-10", Verifications: VerificationStatsOverview{TotalVerifications: 1350, ApprovedVerifications: 1260, RejectedVerifications: 90}},
			{Period: "month", Date: "2025-11", Verifications: VerificationStatsOverview{TotalVerifications: 1400, ApprovedVerifications: 1310, RejectedVerifications: 90}},
		}
	}

	return groupedData, nil
}