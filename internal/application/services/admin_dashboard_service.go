package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AdminDashboardService provides dashboard data for admin panel
type AdminDashboardService struct {
	adminService *AdminService
	cacheService *AdminCacheService
}

// NewAdminDashboardService creates a new AdminDashboardService
func NewAdminDashboardService(adminService *AdminService, cacheService *AdminCacheService) *AdminDashboardService {
	return &AdminDashboardService{
		adminService: adminService,
		cacheService: cacheService,
	}
}

// GetDashboardData retrieves comprehensive dashboard data
func (s *AdminDashboardService) GetDashboardData(ctx context.Context, adminID uuid.UUID) (*DashboardData, error) {
	logger.Info("AdminDashboardService.GetDashboardData called", "admin_id", adminID)

	// Try to get from cache first
	var dashboardData DashboardData
	err := s.cacheService.GetDashboard(ctx, adminID, &dashboardData)
	if err == nil {
		logger.Debug("Dashboard data retrieved from cache", "admin_id", adminID)
		return &dashboardData, nil
	}

	// Cache miss, fetch data
	dashboardData, err = s.fetchDashboardData(ctx, adminID)
	if err != nil {
		logger.Error("Failed to fetch dashboard data", err, "admin_id", adminID)
		return nil, fmt.Errorf("failed to fetch dashboard data: %w", err)
	}

	// Cache the data
	if err := s.cacheService.CacheDashboard(ctx, adminID, dashboardData); err != nil {
		logger.Warn("Failed to cache dashboard data", err, "admin_id", adminID)
	}

	logger.Info("Dashboard data retrieved successfully", "admin_id", adminID)
	return &dashboardData, nil
}

// fetchDashboardData fetches dashboard data from various sources
func (s *AdminDashboardService) fetchDashboardData(ctx context.Context, adminID uuid.UUID) (DashboardData, error) {
	// Fetch data concurrently using goroutines
	type result struct {
		overview    *OverviewStats
		userStats   *UserDashboardStats
		analytics   *AnalyticsData
		system      *SystemStatus
		content     *ContentModerationData
		alerts      []Alert
		activities  []RecentActivity
		err         error
	}

	results := make(chan result, 7)

	// Fetch overview stats
	go func() {
		overview, err := s.getOverviewStats(ctx, adminID)
		results <- result{overview: overview, err: err}
	}()

	// Fetch user stats
	go func() {
		userStats, err := s.getUserDashboardStats(ctx, adminID)
		results <- result{userStats: userStats, err: err}
	}()

	// Fetch analytics data
	go func() {
		analytics, err := s.getAnalyticsData(ctx, adminID)
		results <- result{analytics: analytics, err: err}
	}()

	// Fetch system status
	go func() {
		system, err := s.getSystemStatus(ctx, adminID)
		results <- result{system: system, err: err}
	}()

	// Fetch content moderation data
	go func() {
		content, err := s.getContentModerationData(ctx, adminID)
		results <- result{content: content, err: err}
	}()

	// Fetch alerts
	go func() {
		alerts, err := s.getAlerts(ctx, adminID)
		results <- result{alerts: alerts, err: err}
	}()

	// Fetch recent activities
	go func() {
		activities, err := s.getRecentActivities(ctx, adminID)
		results <- result{activities: activities, err: err}
	}()

	// Collect results
	var dashboardData DashboardData
	for i := 0; i < 7; i++ {
		res := <-results
		if res.err != nil {
			logger.Warn("Failed to fetch dashboard component", res.err)
			continue
		}

		if res.overview != nil {
			dashboardData.Overview = *res.overview
		}
		if res.userStats != nil {
			dashboardData.UserStats = *res.userStats
		}
		if res.analytics != nil {
			dashboardData.Analytics = *res.analytics
		}
		if res.system != nil {
			dashboardData.System = *res.system
		}
		if res.content != nil {
			dashboardData.Content = *res.content
		}
		if res.alerts != nil {
			dashboardData.Alerts = res.alerts
		}
		if res.activities != nil {
			dashboardData.RecentActivities = res.activities
		}
	}

	dashboardData.Timestamp = time.Now()
	return dashboardData, nil
}

// getOverviewStats fetches overview statistics
func (s *AdminDashboardService) getOverviewStats(ctx context.Context, adminID uuid.UUID) (*OverviewStats, error) {
	// Mock implementation - in real implementation, this would call admin service
	return &OverviewStats{
		TotalUsers:        50000,
		ActiveUsers:       12000,
		NewUsersToday:     150,
		TotalMatches:      250000,
		NewMatchesToday:   800,
		TotalMessages:     1000000,
		NewMessagesToday:  5000,
		TotalRevenue:      250000.50,
		RevenueToday:      1250.75,
		GrowthRate:        15.5,
		EngagementRate:    68.2,
	}, nil
}

// getUserDashboardStats fetches user-specific dashboard statistics
func (s *AdminDashboardService) getUserDashboardStats(ctx context.Context, adminID uuid.UUID) (*UserDashboardStats, error) {
	// Mock implementation - in real implementation, this would call admin service
	return &UserDashboardStats{
		RegistrationTrend: []TrendData{
			{Date: "2025-11-01", Value: 120},
			{Date: "2025-11-02", Value: 135},
			{Date: "2025-11-03", Value: 150},
			{Date: "2025-11-04", Value: 145},
			{Date: "2025-11-05", Value: 160},
			{Date: "2025-11-06", Value: 155},
			{Date: "2025-11-07", Value: 170},
		},
		ActiveUsersTrend: []TrendData{
			{Date: "2025-11-01", Value: 11000},
			{Date: "2025-11-02", Value: 11200},
			{Date: "2025-11-03", Value: 11500},
			{Date: "2025-11-04", Value: 11800},
			{Date: "2025-11-05", Value: 12000},
			{Date: "2025-11-06", Value: 11900},
			{Date: "2025-11-07", Value: 12100},
		},
		UserDemographics: UserDemographics{
			ByAge: map[string]int{
				"18-24": 15000,
				"25-34": 20000,
				"35-44": 10000,
				"45-54": 4000,
				"55+":   1000,
			},
			ByGender: map[string]int{
				"male":   25000,
				"female": 24000,
				"other":  1000,
			},
			ByLocation: map[string]int{
				"United States": 15000,
				"United Kingdom": 8000,
				"Canada": 5000,
				"Germany": 4000,
				"France": 3000,
				"Other": 15000,
			},
		},
	}, nil
}

// getAnalyticsData fetches analytics data for charts
func (s *AdminDashboardService) getAnalyticsData(ctx context.Context, adminID uuid.UUID) (*AnalyticsData, error) {
	// Mock implementation - in real implementation, this would call admin service
	return &AnalyticsData{
		MatchTrend: []TrendData{
			{Date: "2025-11-01", Value: 750},
			{Date: "2025-11-02", Value: 780},
			{Date: "2025-11-03", Value: 820},
			{Date: "2025-11-04", Value: 790},
			{Date: "2025-11-05", Value: 850},
			{Date: "2025-11-06", Value: 830},
			{Date: "2025-11-07", Value: 880},
		},
		MessageTrend: []TrendData{
			{Date: "2025-11-01", Value: 4500},
			{Date: "2025-11-02", Value: 4700},
			{Date: "2025-11-03", Value: 4900},
			{Date: "2025-11-04", Value: 4800},
			{Date: "2025-11-05", Value: 5100},
			{Date: "2025-11-06", Value: 5000},
			{Date: "2025-11-07", Value: 5200},
		},
		RevenueTrend: []TrendData{
			{Date: "2025-11-01", Value: 1100.50},
			{Date: "2025-11-02", Value: 1150.75},
			{Date: "2025-11-03", Value: 1200.25},
			{Date: "2025-11-04", Value: 1180.50},
			{Date: "2025-11-05", Value: 1250.75},
			{Date: "2025-11-06", Value: 1230.25},
			{Date: "2025-11-07", Value: 1280.50},
		},
		ConversionRates: ConversionRates{
			RegistrationToMatch: 25.5,
			MatchToMessage:       68.2,
			MessageToSubscription: 8.5,
		},
	}, nil
}

// getSystemStatus fetches system status information
func (s *AdminDashboardService) getSystemStatus(ctx context.Context, adminID uuid.UUID) (*SystemStatus, error) {
	// Mock implementation - in real implementation, this would call admin service
	return &SystemStatus{
		OverallStatus: "healthy",
		Services: []ServiceStatus{
			{Name: "API Server", Status: "healthy", ResponseTime: 45, Uptime: 99.9},
			{Name: "Database", Status: "healthy", ResponseTime: 12, Uptime: 99.8},
			{Name: "Redis Cache", Status: "healthy", ResponseTime: 3, Uptime: 99.9},
			{Name: "S3 Storage", Status: "healthy", ResponseTime: 85, Uptime: 99.7},
			{Name: "Email Service", Status: "degraded", ResponseTime: 250, Uptime: 98.5},
			{Name: "Payment Gateway", Status: "healthy", ResponseTime: 120, Uptime: 99.9},
		},
		ResourceUsage: ResourceUsage{
			CPU:    45.2,
			Memory: 68.5,
			Disk:   32.1,
			Network: 15.3,
		},
		PerformanceMetrics: PerformanceMetrics{
			RequestsPerSecond: 125.5,
			AverageResponseTime: 85.2,
			ErrorRate: 0.8,
			ActiveConnections: 450,
		},
	}, nil
}

// getContentModerationData fetches content moderation data
func (s *AdminDashboardService) getContentModerationData(ctx context.Context, adminID uuid.UUID) (*ContentModerationData, error) {
	// Mock implementation - in real implementation, this would call admin service
	return &ContentModerationData{
		PendingReports: 25,
		ResolvedToday:   15,
		TotalReports:    150,
		ReportTypes: map[string]int{
			"inappropriate_content": 80,
			"spam":                40,
			"harassment":          20,
			"fake_profile":        10,
		},
		RecentModeration: []ModerationAction{
			{
				ID:          uuid.New(),
				ContentType: "photo",
				Action:      "approved",
				Reason:      "No violation found",
				Timestamp:   time.Now().Add(-30 * time.Minute),
				AdminID:     adminID,
			},
			{
				ID:          uuid.New(),
				ContentType: "message",
				Action:      "deleted",
				Reason:      "Harassment",
				Timestamp:   time.Now().Add(-1 * time.Hour),
				AdminID:     adminID,
			},
			{
				ID:          uuid.New(),
				ContentType: "photo",
				Action:      "rejected",
				Reason:      "Inappropriate content",
				Timestamp:   time.Now().Add(-2 * time.Hour),
				AdminID:     adminID,
			},
		},
	}, nil
}

// getAlerts fetches system alerts
func (s *AdminDashboardService) getAlerts(ctx context.Context, adminID uuid.UUID) ([]Alert, error) {
	// Mock implementation - in real implementation, this would call admin service
	return []Alert{
		{
			ID:        uuid.New(),
			Level:     "warning",
			Title:     "High response time on Email Service",
			Message:   "Email service response time is above threshold",
			Service:   "Email Service",
			Timestamp: time.Now().Add(-30 * time.Minute),
			Acknowledged: false,
			Resolved:  false,
		},
		{
			ID:        uuid.New(),
			Level:     "info",
			Title:     "Scheduled maintenance completed",
			Message:   "Database maintenance has been completed successfully",
			Service:   "Database",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Acknowledged: true,
			Resolved:  true,
		},
		{
			ID:        uuid.New(),
			Level:     "error",
			Title:     "Payment gateway timeout",
			Message:   "Payment gateway is experiencing timeouts",
			Service:   "Payment Gateway",
			Timestamp: time.Now().Add(-4 * time.Hour),
			Acknowledged: true,
			Resolved:  false,
		},
	}, nil
}

// getRecentActivities fetches recent admin activities
func (s *AdminDashboardService) getRecentActivities(ctx context.Context, adminID uuid.UUID) ([]RecentActivity, error) {
	// Mock implementation - in real implementation, this would call admin service
	return []RecentActivity{
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			Action:    "User suspended",
			Target:    "user@example.com",
			Details:   "User suspended for 7 days due to policy violation",
			Timestamp: time.Now().Add(-15 * time.Minute),
			IPAddress: "192.168.1.100",
		},
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			Action:    "Content approved",
			Target:    "photo_12345",
			Details:   "Photo approved after review",
			Timestamp: time.Now().Add(-30 * time.Minute),
			IPAddress: "192.168.1.100",
		},
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			Action:    "System configuration updated",
			Target:    "email_settings",
			Details:   "Updated SMTP configuration",
			Timestamp: time.Now().Add(-1 * time.Hour),
			IPAddress: "192.168.1.100",
		},
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			Action:    "Report resolved",
			Target:    "report_67890",
			Details:   "Harassment report resolved with warning",
			Timestamp: time.Now().Add(-2 * time.Hour),
			IPAddress: "192.168.1.100",
		},
	}, nil
}

// Data structures for dashboard

// DashboardData represents the complete dashboard data
type DashboardData struct {
	Overview         OverviewStats          `json:"overview"`
	UserStats        UserDashboardStats     `json:"user_stats"`
	Analytics        AnalyticsData         `json:"analytics"`
	System           SystemStatus          `json:"system"`
	Content          ContentModerationData `json:"content"`
	Alerts           []Alert              `json:"alerts"`
	RecentActivities []RecentActivity      `json:"recent_activities"`
	Timestamp        time.Time            `json:"timestamp"`
}

// OverviewStats represents high-level overview statistics
type OverviewStats struct {
	TotalUsers       int64   `json:"total_users"`
	ActiveUsers      int64   `json:"active_users"`
	NewUsersToday    int64   `json:"new_users_today"`
	TotalMatches     int64   `json:"total_matches"`
	NewMatchesToday  int64   `json:"new_matches_today"`
	TotalMessages    int64   `json:"total_messages"`
	NewMessagesToday int64   `json:"new_messages_today"`
	TotalRevenue     float64 `json:"total_revenue"`
	RevenueToday    float64 `json:"revenue_today"`
	GrowthRate      float64 `json:"growth_rate"`
	EngagementRate  float64 `json:"engagement_rate"`
}

// UserDashboardStats represents user-specific statistics
type UserDashboardStats struct {
	RegistrationTrend []TrendData     `json:"registration_trend"`
	ActiveUsersTrend []TrendData     `json:"active_users_trend"`
	UserDemographics UserDemographics `json:"user_demographics"`
}

// UserDemographics represents user demographic data
type UserDemographics struct {
	ByAge     map[string]int `json:"by_age"`
	ByGender   map[string]int `json:"by_gender"`
	ByLocation map[string]int `json:"by_location"`
}

// AnalyticsData represents analytics data for charts
type AnalyticsData struct {
	MatchTrend      []TrendData    `json:"match_trend"`
	MessageTrend    []TrendData    `json:"message_trend"`
	RevenueTrend    []TrendData    `json:"revenue_trend"`
	ConversionRates ConversionRates `json:"conversion_rates"`
}

// ConversionRates represents conversion rate metrics
type ConversionRates struct {
	RegistrationToMatch float64 `json:"registration_to_match"`
	MatchToMessage     float64 `json:"match_to_message"`
	MessageToSubscription float64 `json:"message_to_subscription"`
}

// SystemStatus represents system status information
type SystemStatus struct {
	OverallStatus     string             `json:"overall_status"`
	Services          []ServiceStatus    `json:"services"`
	ResourceUsage     ResourceUsage      `json:"resource_usage"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
}

// ServiceStatus represents status of a service
type ServiceStatus struct {
	Name         string  `json:"name"`
	Status       string  `json:"status"`
	ResponseTime int     `json:"response_time"` // in milliseconds
	Uptime       float64 `json:"uptime"`       // percentage
}

// ResourceUsage represents resource usage information
type ResourceUsage struct {
	CPU    float64 `json:"cpu"`    // percentage
	Memory float64 `json:"memory"` // percentage
	Disk   float64 `json:"disk"`   // percentage
	Network float64 `json:"network"` // percentage
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	RequestsPerSecond  float64 `json:"requests_per_second"`
	AverageResponseTime float64 `json:"average_response_time"` // in milliseconds
	ErrorRate          float64 `json:"error_rate"`            // percentage
	ActiveConnections  int64   `json:"active_connections"`
}

// ContentModerationData represents content moderation information
type ContentModerationData struct {
	PendingReports    int                `json:"pending_reports"`
	ResolvedToday     int                `json:"resolved_today"`
	TotalReports      int                `json:"total_reports"`
	ReportTypes       map[string]int      `json:"report_types"`
	RecentModeration  []ModerationAction  `json:"recent_moderation"`
}

// ModerationAction represents a moderation action
type ModerationAction struct {
	ID          uuid.UUID `json:"id"`
	ContentType string    `json:"content_type"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	AdminID     uuid.UUID `json:"admin_id"`
}

// Alert represents a system alert
type Alert struct {
	ID           uuid.UUID `json:"id"`
	Level        string    `json:"level"` // info, warning, error, critical
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	Service      string    `json:"service"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
	Resolved     bool      `json:"resolved"`
}

// RecentActivity represents a recent admin activity
type RecentActivity struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"admin_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address"`
}

// TrendData represents data for trend charts
type TrendData struct {
	Date  string      `json:"date"`
	Value interface{} `json:"value"`
}