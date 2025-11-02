package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetUserDetailsUseCase handles retrieving detailed user information
type GetUserDetailsUseCase struct {
	userRepo    repositories.UserRepository
	photoRepo   repositories.PhotoRepository
	matchRepo   repositories.MatchRepository
	messageRepo repositories.MessageRepository
	reportRepo  repositories.ReportRepository
}

// NewGetUserDetailsUseCase creates a new GetUserDetailsUseCase
func NewGetUserDetailsUseCase(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	matchRepo repositories.MatchRepository,
	messageRepo repositories.MessageRepository,
	reportRepo repositories.ReportRepository,
) *GetUserDetailsUseCase {
	return &GetUserDetailsUseCase{
		userRepo:    userRepo,
		photoRepo:   photoRepo,
		matchRepo:   matchRepo,
		messageRepo: messageRepo,
		reportRepo:  reportRepo,
	}
}

// GetUserDetailsRequest represents request to get user details
type GetUserDetailsRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"`
}

// GetUserActivityRequest represents request to get user activity
type GetUserActivityRequest struct {
	AdminID      uuid.UUID  `json:"admin_id" validate:"required"`
	UserID       uuid.UUID  `json:"user_id" validate:"required"`
	ActivityType string     `json:"activity_type" validate:"omitempty,oneof=login swipe match message photo_update profile_update"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Limit         int        `json:"limit" validate:"min=1,max=100"`
	Offset        int        `json:"offset" validate:"min=0"`
}

// UserDetails represents detailed user information
type UserDetails struct {
	User         *entities.User          `json:"user"`
	Stats        *UserDetailedStats     `json:"stats"`
	Photos       []*entities.Photo       `json:"photos"`
	RecentMatches []*entities.Match      `json:"recent_matches"`
	RecentReports []*entities.Report     `json:"recent_reports"`
}

// UserDetailedStats represents detailed user statistics
type UserDetailedStats struct {
	TotalSwipes      int64     `json:"total_swipes"`
	TotalMatches     int64     `json:"total_matches"`
	TotalMessages    int64     `json:"total_messages"`
	PhotosCount      int64     `json:"photos_count"`
	ProfileViews     int64     `json:"profile_views"`
	LastActiveDays   int        `json:"last_active_days"`
	AccountAgeDays   int        `json:"account_age_days"`
	AverageResponseTime float64   `json:"average_response_time"`
	MessagesPerDay   float64    `json:"messages_per_day"`
	SwipesPerDay     float64    `json:"swipes_per_day"`
	MatchRate        float64    `json:"match_rate"`
}

// GetUserActivityResponse represents response from getting user activity
type GetUserActivityResponse struct {
	Activities []UserActivity `json:"activities"`
	Total      int64         `json:"total"`
}

// UserActivity represents a user activity event
type UserActivity struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Execute retrieves detailed user information
func (uc *GetUserDetailsUseCase) Execute(ctx context.Context, adminID, userID uuid.UUID) (*UserDetails, error) {
	logger.Info("GetUserDetails use case executed", "admin_id", adminID, "user_id", userID)

	// Get user details
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user from repository", err, "admin_id", adminID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get user statistics
	stats, err := uc.userRepo.GetUserStats(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user statistics", err, "admin_id", adminID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}

	// Convert to detailed stats
	detailedStats := uc.convertToDetailedStats(stats, user)

	// Get user photos
	photos, err := uc.photoRepo.GetUserPhotos(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user photos", err, "admin_id", adminID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user photos: %w", err)
	}

	// Get recent matches
	matches, err := uc.matchRepo.GetUserMatches(ctx, userID, 10, 0)
	if err != nil {
		logger.Error("Failed to get user matches", err, "admin_id", adminID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user matches: %w", err)
	}

	// Get recent reports
	reports, err := uc.reportRepo.GetReportsByUser(ctx, userID, 10, 0)
	if err != nil {
		logger.Error("Failed to get user reports", err, "admin_id", adminID, "user_id", userID)
		return nil, fmt.Errorf("failed to get user reports: %w", err)
	}

	logger.Info("GetUserDetails use case completed successfully", "admin_id", adminID, "user_id", userID)
	return &UserDetails{
		User:          user,
		Stats:         detailedStats,
		Photos:        photos,
		RecentMatches: matches,
		RecentReports: reports,
	}, nil
}

// GetActivity retrieves user activity logs
func (uc *GetUserDetailsUseCase) GetActivity(ctx context.Context, req GetUserActivityRequest) (*GetUserActivityResponse, error) {
	logger.Info("GetUserActivity use case executed", "admin_id", req.AdminID, "user_id", req.UserID)

	// This would typically query an activity log table or use an audit service
	// For now, we'll generate mock activity data
	activities := uc.generateMockActivity(req)

	// Calculate total (in a real implementation, this would come from the database)
	total := int64(len(activities))

	logger.Info("GetUserActivity use case completed successfully", "admin_id", req.AdminID, "user_id", req.UserID, "count", len(activities))
	return &GetUserActivityResponse{
		Activities: activities,
		Total:      total,
	}, nil
}

// convertToDetailedStats converts basic stats to detailed stats
func (uc *GetUserDetailsUseCase) convertToDetailedStats(stats *repositories.UserStats, user *entities.User) *UserDetailedStats {
	now := time.Now()
	
	// Calculate account age in days
	accountAgeDays := int(now.Sub(user.CreatedAt).Hours() / 24)
	
	// Calculate days since last active
	lastActiveDays := 0
	if user.LastActive != nil {
		lastActiveDays = int(now.Sub(*user.LastActive).Hours() / 24)
	}
	
	// Calculate rates
	var swipesPerDay, messagesPerDay, matchRate float64
	if accountAgeDays > 0 {
		swipesPerDay = float64(stats.TotalSwipes) / float64(accountAgeDays)
		messagesPerDay = float64(stats.TotalMessages) / float64(accountAgeDays)
	}
	if stats.TotalSwipes > 0 {
		matchRate = float64(stats.TotalMatches) / float64(stats.TotalSwipes) * 100
	}

	return &UserDetailedStats{
		TotalSwipes:        stats.TotalSwipes,
		TotalMatches:       stats.TotalMatches,
		TotalMessages:      stats.TotalMessages,
		PhotosCount:        stats.PhotosCount,
		ProfileViews:       stats.ProfileViews,
		LastActiveDays:     lastActiveDays,
		AccountAgeDays:     accountAgeDays,
		AverageResponseTime: 0, // Would be calculated from message timestamps
		MessagesPerDay:     messagesPerDay,
		SwipesPerDay:       swipesPerDay,
		MatchRate:          matchRate,
	}
}

// generateMockActivity generates mock activity data for demonstration
func (uc *GetUserDetailsUseCase) generateMockActivity(req GetUserActivityRequest) []UserActivity {
	// In a real implementation, this would query an activity log table
	// For now, we'll generate some mock data
	activities := make([]UserActivity, 0)
	
	// Generate some sample activities
	baseTime := time.Now().AddDate(0, 0, -30) // 30 days ago
	
	for i := 0; i < 50; i++ {
		activity := UserActivity{
			ID:        uuid.New(),
			UserID:    req.UserID,
			Type:      "login",
			Timestamp: baseTime.AddDate(0, 0, i),
			IPAddress: "192.168.1." + fmt.Sprintf("%d", 100+i%155),
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
			Metadata: map[string]interface{}{
				"device": "mobile",
			},
		}
		
		// Vary activity types
		switch i % 5 {
		case 0:
			activity.Type = "login"
			activity.Description = "User logged in"
		case 1:
			activity.Type = "swipe"
			activity.Description = "User swiped on a profile"
		case 2:
			activity.Type = "match"
			activity.Description = "User got a new match"
		case 3:
			activity.Type = "message"
			activity.Description = "User sent a message"
		case 4:
			activity.Type = "profile_update"
			activity.Description = "User updated profile"
		}
		
		activities = append(activities, activity)
	}
	
	// Apply filters
	if req.ActivityType != "" {
		filtered := make([]UserActivity, 0)
		for _, activity := range activities {
			if activity.Type == req.ActivityType {
				filtered = append(filtered, activity)
			}
		}
		activities = filtered
	}
	
	// Apply time range filters
	if req.StartTime != nil {
		filtered := make([]UserActivity, 0)
		for _, activity := range activities {
			if activity.Timestamp.After(*req.StartTime) {
				filtered = append(filtered, activity)
			}
		}
		activities = filtered
	}
	
	if req.EndTime != nil {
		filtered := make([]UserActivity, 0)
		for _, activity := range activities {
			if activity.Timestamp.Before(*req.EndTime) {
				filtered = append(filtered, activity)
			}
		}
		activities = filtered
	}
	
	// Apply pagination
	start := req.Offset
	end := start + req.Limit
	if end > len(activities) {
		end = len(activities)
	}
	if start >= len(activities) {
		return []UserActivity{}
	}
	
	return activities[start:end]
}