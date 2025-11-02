package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetPaymentStatsUseCase handles retrieving payment statistics
type GetPaymentStatsUseCase struct {
	paymentRepo repositories.PaymentRepository
}

// NewGetPaymentStatsUseCase creates a new GetPaymentStatsUseCase
func NewGetPaymentStatsUseCase(paymentRepo repositories.PaymentRepository) *GetPaymentStatsUseCase {
	return &GetPaymentStatsUseCase{
		paymentRepo: paymentRepo,
	}
}

// GetPaymentStatsRequest represents a request to get payment statistics
type GetPaymentStatsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Period     string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	GroupBy    string     `json:"group_by" validate:"required,oneof=day week month"`
	Currency   string     `json:"currency" validate:"required,oneof=USD EUR GBP"`
}

// PaymentStatsResponse represents response from getting payment statistics
type PaymentStatsResponse struct {
	Overview    PaymentStatsOverview    `json:"overview"`
	Growth      PaymentGrowthStats     `json:"growth"`
	Conversion  PaymentConversionStats  `json:"conversion"`
	GroupedData []PaymentStatsGroup    `json:"grouped_data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PaymentStatsOverview represents high-level payment statistics
type PaymentStatsOverview struct {
	TotalRevenue        float64 `json:"total_revenue"`
	NewRevenue         float64 `json:"new_revenue"`
	TotalTransactions  int64   `json:"total_transactions"`
	NewTransactions   int64   `json:"new_transactions"`
	TotalSubscriptions int64   `json:"total_subscriptions"`
	NewSubscriptions  int64   `json:"new_subscriptions"`
	ChurnRate         float64 `json:"churn_rate"`
	AvgRevenuePerUser float64 `json:"avg_revenue_per_user"`
}

// PaymentGrowthStats represents payment growth statistics
type PaymentGrowthStats struct {
	DailyGrowth   []GrowthData `json:"daily_growth"`
	WeeklyGrowth  []GrowthData `json:"weekly_growth"`
	MonthlyGrowth []GrowthData `json:"monthly_growth"`
	GrowthRate     float64      `json:"growth_rate"`
}

// PaymentConversionStats represents payment conversion statistics
type PaymentConversionStats struct {
	TrialToPaidRate    float64 `json:"trial_to_paid_rate"`
	FreeToPremiumRate   float64 `json:"free_to_premium_rate"`
	PremiumToPlatinumRate float64 `json:"premium_to_platinum_rate"`
	ConversionRate      float64 `json:"conversion_rate"`
	AvgTimeToConvert   float64 `json:"avg_time_to_convert"` // in days
}

// PaymentStatsGroup represents grouped payment statistics
type PaymentStatsGroup struct {
	Period   string                `json:"period"`
	Date     string                `json:"date"`
	Payments PaymentStatsOverview `json:"payments"`
}

// Execute retrieves payment statistics
func (uc *GetPaymentStatsUseCase) Execute(ctx context.Context, req GetPaymentStatsRequest) (*PaymentStatsResponse, error) {
	logger.Info("GetPaymentStats use case executed", "admin_id", req.AdminID, "period", req.Period, "group_by", req.GroupBy, "currency", req.Currency)

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange(req.Period, req.StartTime, req.EndTime)

	// Get overview statistics
	overview, err := uc.getOverviewStats(ctx, startTime, endTime, req.Currency)
	if err != nil {
		logger.Error("Failed to get overview statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get overview statistics: %w", err)
	}

	// Get growth statistics
	growth, err := uc.getGrowthStats(ctx, startTime, endTime, req.GroupBy, req.Currency)
	if err != nil {
		logger.Error("Failed to get growth statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get growth statistics: %w", err)
	}

	// Get conversion statistics
	conversion, err := uc.getConversionStats(ctx, startTime, endTime, req.Currency)
	if err != nil {
		logger.Error("Failed to get conversion statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get conversion statistics: %w", err)
	}

	// Get grouped data
	groupedData, err := uc.getGroupedData(ctx, startTime, endTime, req.GroupBy, req.Currency)
	if err != nil {
		logger.Error("Failed to get grouped data", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get grouped data: %w", err)
	}

	logger.Info("GetPaymentStats use case completed successfully", "admin_id", req.AdminID)
	return &PaymentStatsResponse{
		Overview:    *overview,
		Growth:      *growth,
		Conversion:  *conversion,
		GroupedData: groupedData,
		Timestamp:   time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetPaymentStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
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
func (uc *GetPaymentStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time, currency string) (*PaymentStatsOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &PaymentStatsOverview{
		TotalRevenue:        50000.0,
		NewRevenue:         4500.0,
		TotalTransactions:  15000,
		NewTransactions:   1200,
		TotalSubscriptions: 1500,
		NewSubscriptions:  120,
		ChurnRate:         5.2,
		AvgRevenuePerUser: 5.0,
	}, nil
}

func (uc *GetPaymentStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time, groupBy string, currency string) (*PaymentGrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &PaymentGrowthStats{
		DailyGrowth: []GrowthData{
			{Date: "2025-11-01", Value: 150.0, Growth: 18.5},
			{Date: "2025-11-02", Value: 160.0, Growth: 15.2},
			{Date: "2025-11-03", Value: 155.0, Growth: 12.8},
		},
		WeeklyGrowth: []GrowthData{
			{Date: "2025-10-30", Value: 1050.0, Growth: 22.5},
			{Date: "2025-11-06", Value: 1200.0, Growth: 25.0},
		},
		MonthlyGrowth: []GrowthData{
			{Date: "2025-09", Value: 4500.0, Growth: 20.5},
			{Date: "2025-10", Value: 5000.0, Growth: 25.0},
			{Date: "2025-11", Value: 5500.0, Growth: 30.0},
		},
		GrowthRate: 18.5,
	}, nil
}

func (uc *GetPaymentStatsUseCase) getConversionStats(ctx context.Context, startTime, endTime time.Time, currency string) (*PaymentConversionStats, error) {
	// Mock data - in real implementation, this would query the database
	return &PaymentConversionStats{
		TrialToPaidRate:     25.5,
		FreeToPremiumRate:    15.2,
		PremiumToPlatinumRate: 8.5,
		ConversionRate:      18.5,
		AvgTimeToConvert:   7.5, // days
	}, nil
}

func (uc *GetPaymentStatsUseCase) getGroupedData(ctx context.Context, startTime, endTime time.Time, groupBy string, currency string) ([]PaymentStatsGroup, error) {
	// Mock data - in real implementation, this would query the database
	var groupedData []PaymentStatsGroup

	switch groupBy {
	case "day":
		groupedData = []PaymentStatsGroup{
			{Period: "day", Date: "2025-11-01", Payments: PaymentStatsOverview{TotalRevenue: 150.0, NewRevenue: 15.0, TotalSubscriptions: 5}},
			{Period: "day", Date: "2025-11-02", Payments: PaymentStatsOverview{TotalRevenue: 160.0, NewRevenue: 16.0, TotalSubscriptions: 6}},
			{Period: "day", Date: "2025-11-03", Payments: PaymentStatsOverview{TotalRevenue: 155.0, NewRevenue: 15.5, TotalSubscriptions: 4}},
		}
	case "week":
		groupedData = []PaymentStatsGroup{
			{Period: "week", Date: "2025-10-30", Payments: PaymentStatsOverview{TotalRevenue: 1050.0, NewRevenue: 105.0, TotalSubscriptions: 35}},
			{Period: "week", Date: "2025-11-06", Payments: PaymentStatsOverview{TotalRevenue: 1200.0, NewRevenue: 120.0, TotalSubscriptions: 40}},
		}
	case "month":
		groupedData = []PaymentStatsGroup{
			{Period: "month", Date: "2025-09", Payments: PaymentStatsOverview{TotalRevenue: 4500.0, NewRevenue: 450.0, TotalSubscriptions: 150}},
			{Period: "month", Date: "2025-10", Payments: PaymentStatsOverview{TotalRevenue: 5000.0, NewRevenue: 500.0, TotalSubscriptions: 170}},
			{Period: "month", Date: "2025-11", Payments: PaymentStatsOverview{TotalRevenue: 5500.0, NewRevenue: 550.0, TotalSubscriptions: 200}},
		}
	}

	return groupedData, nil
}