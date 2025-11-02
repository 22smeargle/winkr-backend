package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetMessageStatsUseCase handles retrieving message statistics
type GetMessageStatsUseCase struct {
	messageRepo repositories.MessageRepository
}

// NewGetMessageStatsUseCase creates a new GetMessageStatsUseCase
func NewGetMessageStatsUseCase(messageRepo repositories.MessageRepository) *GetMessageStatsUseCase {
	return &GetMessageStatsUseCase{
		messageRepo: messageRepo,
	}
}

// GetMessageStatsRequest represents a request to get message statistics
type GetMessageStatsRequest struct {
	AdminID   uuid.UUID  `json:"admin_id" validate:"required"`
	Period     string     `json:"period" validate:"required,oneof=1d 7d 30d 90d 1y"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	GroupBy    string     `json:"group_by" validate:"required,oneof=day week month"`
}

// MessageStatsResponse represents response from getting message statistics
type MessageStatsResponse struct {
	Overview    MessageStatsOverview    `json:"overview"`
	Growth      MessageGrowthStats     `json:"growth"`
	Engagement  MessageEngagementStats  `json:"engagement"`
	GroupedData []MessageStatsGroup    `json:"grouped_data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// MessageStatsOverview represents high-level message statistics
type MessageStatsOverview struct {
	TotalMessages      int64 `json:"total_messages"`
	NewMessages        int64 `json:"new_messages"`
	ActiveConversations int64 `json:"active_conversations"`
	AvgMessagesPerUser float64 `json:"avg_messages_per_user"`
	AvgMessageLength    float64 `json:"avg_message_length"`
	ResponseRate       float64 `json:"response_rate"`
}

// MessageGrowthStats represents message growth statistics
type MessageGrowthStats struct {
	DailyGrowth   []GrowthData `json:"daily_growth"`
	WeeklyGrowth  []GrowthData `json:"weekly_growth"`
	MonthlyGrowth []GrowthData `json:"monthly_growth"`
	GrowthRate     float64      `json:"growth_rate"`
}

// MessageEngagementStats represents message engagement statistics
type MessageEngagementStats struct {
	MessagesPerConversation float64 `json:"messages_per_conversation"`
	AvgResponseTime          float64 `json:"avg_response_time"` // in minutes
	ConversationDuration       float64 `json:"conversation_duration"` // in hours
	ActiveConversations       int64   `json:"active_conversations"`
	EngagementRate          float64 `json:"engagement_rate"`
}

// MessageStatsGroup represents grouped message statistics
type MessageStatsGroup struct {
	Period   string               `json:"period"`
	Date     string               `json:"date"`
	Messages MessageStatsOverview `json:"messages"`
}

// Execute retrieves message statistics
func (uc *GetMessageStatsUseCase) Execute(ctx context.Context, req GetMessageStatsRequest) (*MessageStatsResponse, error) {
	logger.Info("GetMessageStats use case executed", "admin_id", req.AdminID, "period", req.Period, "group_by", req.GroupBy)

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

	// Get engagement statistics
	engagement, err := uc.getEngagementStats(ctx, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get engagement statistics", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get engagement statistics: %w", err)
	}

	// Get grouped data
	groupedData, err := uc.getGroupedData(ctx, startTime, endTime, req.GroupBy)
	if err != nil {
		logger.Error("Failed to get grouped data", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get grouped data: %w", err)
	}

	logger.Info("GetMessageStats use case completed successfully", "admin_id", req.AdminID)
	return &MessageStatsResponse{
		Overview:    *overview,
		Growth:      *growth,
		Engagement:  *engagement,
		GroupedData: groupedData,
		Timestamp:   time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time based on period
func (uc *GetMessageStatsUseCase) calculateTimeRange(period string, startTime, endTime *time.Time) (time.Time, time.Time) {
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
func (uc *GetMessageStatsUseCase) getOverviewStats(ctx context.Context, startTime, endTime time.Time) (*MessageStatsOverview, error) {
	// Mock data - in real implementation, this would query the database
	return &MessageStatsOverview{
		TotalMessages:      250000,
		NewMessages:        14000,
		ActiveConversations: 3500,
		AvgMessagesPerUser: 25.0,
		AvgMessageLength:    45.5,
		ResponseRate:       75.5,
	}, nil
}

func (uc *GetMessageStatsUseCase) getGrowthStats(ctx context.Context, startTime, endTime time.Time, groupBy string) (*MessageGrowthStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MessageGrowthStats{
		DailyGrowth: []GrowthData{
			{Date: "2025-11-01", Value: 450, Growth: 12.3},
			{Date: "2025-11-02", Value: 480, Growth: 10.5},
			{Date: "2025-11-03", Value: 470, Growth: 8.9},
		},
		WeeklyGrowth: []GrowthData{
			{Date: "2025-10-30", Value: 3100, Growth: 15.2},
			{Date: "2025-11-06", Value: 3500, Growth: 18.5},
		},
		MonthlyGrowth: []GrowthData{
			{Date: "2025-09", Value: 12000, Growth: 14.5},
			{Date: "2025-10", Value: 13500, Growth: 16.2},
			{Date: "2025-11", Value: 14000, Growth: 18.5},
		},
		GrowthRate: 12.3,
	}, nil
}

func (uc *GetMessageStatsUseCase) getEngagementStats(ctx context.Context, startTime, endTime time.Time) (*MessageEngagementStats, error) {
	// Mock data - in real implementation, this would query the database
	return &MessageEngagementStats{
		MessagesPerConversation: 8.2,
		AvgResponseTime:          15.5, // minutes
		ConversationDuration:       2.5, // hours
		ActiveConversations:       3500,
		EngagementRate:          68.5,
	}, nil
}

func (uc *GetMessageStatsUseCase) getGroupedData(ctx context.Context, startTime, endTime time.Time, groupBy string) ([]MessageStatsGroup, error) {
	// Mock data - in real implementation, this would query the database
	var groupedData []MessageStatsGroup

	switch groupBy {
	case "day":
		groupedData = []MessageStatsGroup{
			{Period: "day", Date: "2025-11-01", Messages: MessageStatsOverview{TotalMessages: 450, NewMessages: 45, ResponseRate: 74.2}},
			{Period: "day", Date: "2025-11-02", Messages: MessageStatsOverview{TotalMessages: 480, NewMessages: 48, ResponseRate: 75.8}},
			{Period: "day", Date: "2025-11-03", Messages: MessageStatsOverview{TotalMessages: 470, NewMessages: 47, ResponseRate: 76.1}},
		}
	case "week":
		groupedData = []MessageStatsGroup{
			{Period: "week", Date: "2025-10-30", Messages: MessageStatsOverview{TotalMessages: 3100, NewMessages: 350, ResponseRate: 74.5}},
			{Period: "week", Date: "2025-11-06", Messages: MessageStatsOverview{TotalMessages: 3500, NewMessages: 400, ResponseRate: 75.2}},
		}
	case "month":
		groupedData = []MessageStatsGroup{
			{Period: "month", Date: "2025-09", Messages: MessageStatsOverview{TotalMessages: 12000, NewMessages: 1400, ResponseRate: 74.8}},
			{Period: "month", Date: "2025-10", Messages: MessageStatsOverview{TotalMessages: 13500, NewMessages: 1500, ResponseRate: 75.1}},
			{Period: "month", Date: "2025-11", Messages: MessageStatsOverview{TotalMessages: 14000, NewMessages: 1600, ResponseRate: 75.5}},
		}
	}

	return groupedData, nil
}