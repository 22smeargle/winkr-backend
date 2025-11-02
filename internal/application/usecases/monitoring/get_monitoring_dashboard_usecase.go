package monitoring

import (
	"context"
	"time"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetMonitoringDashboardRequest represents a request to get monitoring dashboard data
type GetMonitoringDashboardRequest struct {
	TimeRange     string `json:"time_range"`      // 1h, 24h, 7d, etc.
	IncludeDetails bool   `json:"include_details"` // Include detailed metrics
	Refresh       bool   `json:"refresh"`        // Force refresh of data
}

// GetMonitoringDashboardResponse represents a response for monitoring dashboard
type GetMonitoringDashboardResponse struct {
	Status       DashboardStatus       `json:"status"`
	Health       HealthDashboard       `json:"health"`
	Performance  PerformanceDashboard  `json:"performance"`
	Business     BusinessDashboard      `json:"business"`
	System       SystemDashboard       `json:"system"`
	Alerts       []AlertInfo          `json:"alerts"`
	TimeRange    string                `json:"time_range"`
	GeneratedAt  time.Time             `json:"generated_at"`
}

// DashboardStatus represents overall dashboard status
type DashboardStatus struct {
	Overall         string  `json:"overall"`
	Uptime          string  `json:"uptime"`
	LastCheck       string  `json:"last_check"`
	ActiveAlerts    int     `json:"active_alerts"`
	CriticalAlerts  int     `json:"critical_alerts"`
	WarningAlerts   int     `json:"warning_alerts"`
}

// HealthDashboard represents health information for dashboard
type HealthDashboard struct {
	Overall     string                    `json:"overall"`
	Components  map[string]ComponentStatus  `json:"components"`
	Dependencies map[string]DependencyStatus  `json:"dependencies"`
}

// ComponentStatus represents component status for dashboard
type ComponentStatus struct {
	Status       string  `json:"status"`
	ResponseTime int     `json:"response_time"`
	LastCheck    string  `json:"last_check"`
	Issues       []string `json:"issues,omitempty"`
}

// DependencyStatus represents dependency status for dashboard
type DependencyStatus struct {
	Status       string `json:"status"`
	ResponseTime int    `json:"response_time"`
	LastCheck    string `json:"last_check"`
}

// PerformanceDashboard represents performance metrics for dashboard
type PerformanceDashboard struct {
	HTTP       HTTPPerformance       `json:"http"`
	Database   DatabasePerformance   `json:"database"`
	Cache      CachePerformance      `json:"cache"`
	ResponseTime ResponseTimeMetrics  `json:"response_time"`
	ErrorRate   ErrorRateMetrics     `json:"error_rate"`
	Throughput  ThroughputMetrics    `json:"throughput"`
}

// HTTPPerformance represents HTTP performance metrics
type HTTPPerformance struct {
	TotalRequests    int64   `json:"total_requests"`
	RequestsPerSec   float64  `json:"requests_per_sec"`
	AverageResponse  int      `json:"average_response_ms"`
	P95Response     int      `json:"p95_response_ms"`
	ErrorRate        float64  `json:"error_rate_percent"`
	TopEndpoints     []EndpointStats `json:"top_endpoints"`
}

// DatabasePerformance represents database performance metrics
type DatabasePerformance struct {
	TotalQueries     int64   `json:"total_queries"`
	QueriesPerSec    float64  `json:"queries_per_sec"`
	AverageQueryTime int      `json:"average_query_ms"`
	SlowQueries      int64    `json:"slow_queries"`
	ConnectionsActive int      `json:"connections_active"`
	ConnectionsIdle   int      `json:"connections_idle"`
}

// CachePerformance represents cache performance metrics
type CachePerformance struct {
	TotalOperations  int64   `json:"total_operations"`
	OperationsPerSec float64  `json:"operations_per_sec"`
	HitRate         float64  `json:"hit_rate_percent"`
	AverageTime      int      `json:"average_time_ms"`
	MissRate        float64  `json:"miss_rate_percent"`
}

// ResponseTimeMetrics represents response time metrics
type ResponseTimeMetrics struct {
	Average int `json:"average_ms"`
	P50     int `json:"p50_ms"`
	P95     int `json:"p95_ms"`
	P99     int `json:"p99_ms"`
}

// ErrorRateMetrics represents error rate metrics
type ErrorRateMetrics struct {
	Overall   float64            `json:"overall_percent"`
	ByStatus  map[string]float64 `json:"by_status"`
	ByPath    map[string]float64 `json:"by_path"`
}

// ThroughputMetrics represents throughput metrics
type ThroughputMetrics struct {
	RequestsPerSec float64 `json:"requests_per_sec"`
	QueriesPerSec  float64 `json:"queries_per_sec"`
	OperationsPerSec float64 `json:"operations_per_sec"`
}

// EndpointStats represents endpoint statistics
type EndpointStats struct {
	Path        string  `json:"path"`
	Requests    int64   `json:"requests"`
	AvgResponse int      `json:"avg_response_ms"`
	ErrorRate   float64  `json:"error_rate_percent"`
}

// BusinessDashboard represents business metrics for dashboard
type BusinessDashboard struct {
	Users        UserMetrics        `json:"users"`
	Engagement  EngagementMetrics  `json:"engagement"`
	Revenue      RevenueMetrics      `json:"revenue"`
	Growth       GrowthMetrics       `json:"growth"`
}

// UserMetrics represents user metrics
type UserMetrics struct {
	Total          int64 `json:"total"`
	Active         int64 `json:"active"`
	NewToday       int64 `json:"new_today"`
	NewThisWeek    int64 `json:"new_this_week"`
	NewThisMonth   int64 `json:"new_this_month"`
	PremiumUsers   int64 `json:"premium_users"`
	VerifiedUsers  int64 `json:"verified_users"`
}

// EngagementMetrics represents engagement metrics
type EngagementMetrics struct {
	TotalMatches    int64 `json:"total_matches"`
	TotalMessages   int64 `json:"total_messages"`
	TotalSwipes     int64 `json:"total_swipes"`
	SwipesPerUser   float64 `json:"swipes_per_user"`
	MessagesPerUser  float64 `json:"messages_per_user"`
	MatchRate        float64 `json:"match_rate_percent"`
}

// RevenueMetrics represents revenue metrics
type RevenueMetrics struct {
	TotalRevenue      float64 `json:"total_revenue"`
	MRR               float64 `json:"monthly_recurring_revenue"`
	ARPU              float64 `json:"average_revenue_per_user"`
	ConversionRate    float64 `json:"conversion_rate_percent"`
	ChurnRate         float64 `json:"churn_rate_percent"`
}

// GrowthMetrics represents growth metrics
type GrowthMetrics struct {
	UserGrowth       float64 `json:"user_growth_percent"`
	RevenueGrowth     float64 `json:"revenue_growth_percent"`
	EngagementGrowth  float64 `json:"engagement_growth_percent"`
	WeeklyGrowth      []float64 `json:"weekly_growth"`
}

// SystemDashboard represents system metrics for dashboard
type SystemDashboard struct {
	Resources   ResourceDashboard   `json:"resources"`
	Performance SystemPerformance   `json:"performance"`
	Uptime      UptimeMetrics       `json:"uptime"`
}

// ResourceDashboard represents resource metrics for dashboard
type ResourceDashboard struct {
	CPU    ResourceUsage `json:"cpu"`
	Memory ResourceUsage `json:"memory"`
	Disk   ResourceUsage `json:"disk"`
	Network NetworkUsage `json:"network"`
}

// ResourceUsage represents resource usage
type ResourceUsage struct {
	Current   float64 `json:"current"`
	Threshold float64 `json:"threshold"`
	Status    string  `json:"status"`
	Trend     string  `json:"trend"`
}

// NetworkUsage represents network usage
type NetworkUsage struct {
	BytesSent    uint64 `json:"bytes_sent"`
	BytesRecv    uint64 `json:"bytes_recv"`
	BandwidthUsed float64 `json:"bandwidth_used_percent"`
}

// SystemPerformance represents system performance metrics
type SystemPerformance struct {
	Goroutines int    `json:"goroutines"`
	GCCycles   uint32  `json:"gc_cycles"`
	LoadAverage float64 `json:"load_average"`
}

// UptimeMetrics represents uptime metrics
type UptimeMetrics struct {
	Current    string  `json:"current"`
	Percentage float64 `json:"percentage"`
	LastDowntime string `json:"last_downtime"`
	TotalDowntime string `json:"total_downtime"`
}

// AlertInfo represents alert information for dashboard
type AlertInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
	AckedAt    *time.Time `json:"acked_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// GetMonitoringDashboardUseCase handles getting monitoring dashboard data
type GetMonitoringDashboardUseCase struct {
	healthCheckService *services.HealthCheckService
	metricsService    *services.MetricsService
}

// NewGetMonitoringDashboardUseCase creates a new GetMonitoringDashboardUseCase
func NewGetMonitoringDashboardUseCase(
	healthCheckService *services.HealthCheckService,
	metricsService *services.MetricsService,
) *GetMonitoringDashboardUseCase {
	return &GetMonitoringDashboardUseCase{
		healthCheckService: healthCheckService,
		metricsService:    metricsService,
	}
}

// Execute executes the get monitoring dashboard use case
func (uc *GetMonitoringDashboardUseCase) Execute(ctx context.Context, req GetMonitoringDashboardRequest) (*GetMonitoringDashboardResponse, error) {
	logger.Info("Getting monitoring dashboard", map[string]interface{}{
		"time_range":      req.TimeRange,
		"include_details": req.IncludeDetails,
		"refresh":         req.Refresh,
	})

	// Get health status
	systemHealth := uc.healthCheckService.GetOverallHealth(ctx)
	
	// Get metrics
	httpMetrics := uc.metricsService.GetHTTPMetrics()
	dbMetrics := uc.metricsService.GetDatabaseMetrics()
	cacheMetrics := uc.metricsService.GetCacheMetrics()
	businessMetrics := uc.metricsService.GetBusinessMetrics()
	systemMetrics := uc.metricsService.GetSystemMetrics()

	// Prepare dashboard data
	dashboard := &GetMonitoringDashboardResponse{
		Status: DashboardStatus{
			Overall:        string(systemHealth.Status),
			Uptime:         systemHealth.Uptime.String(),
			LastCheck:      systemHealth.Timestamp.Format(time.RFC3339),
			ActiveAlerts:   0, // Placeholder - would come from alerting service
			CriticalAlerts: 0, // Placeholder
			WarningAlerts:  0, // Placeholder
		},
		Health: HealthDashboard{
			Overall: string(systemHealth.Status),
			Components: uc.convertComponentsForDashboard(systemHealth.Components),
			Dependencies: uc.convertDependenciesForDashboard(systemHealth.Dependencies),
		},
		Performance: uc.buildPerformanceDashboard(httpMetrics, dbMetrics, cacheMetrics),
		Business: uc.buildBusinessDashboard(businessMetrics),
		System: uc.buildSystemDashboard(systemMetrics, systemHealth),
		Alerts: []AlertInfo{}, // Placeholder - would come from alerting service
		TimeRange: req.TimeRange,
		GeneratedAt: time.Now(),
	}

	return dashboard, nil
}

// Helper methods

func (uc *GetMonitoringDashboardUseCase) convertComponentsForDashboard(components map[string]services.HealthCheck) map[string]ComponentStatus {
	result := make(map[string]ComponentStatus)
	for name, component := range components {
		status := ComponentStatus{
			Status:       string(component.Status),
			ResponseTime: int(component.ResponseTime.Milliseconds()),
			LastCheck:    component.Timestamp.Format(time.RFC3339),
		}
		
		// Add issues if component is not healthy
		if component.Status != services.HealthStatusHealthy {
			status.Issues = []string{component.Error}
		}
		
		result[name] = status
	}
	return result
}

func (uc *GetMonitoringDashboardUseCase) convertDependenciesForDashboard(dependencies map[string]services.DependencyHealth) map[string]DependencyStatus {
	result := make(map[string]DependencyStatus)
	for name, dep := range dependencies {
		result[name] = DependencyStatus{
			Status:       string(dep.Status),
			ResponseTime: int(dep.ResponseTime.Milliseconds()),
			LastCheck:    time.Now().Format(time.RFC3339), // Simplified
		}
	}
	return result
}

func (uc *GetMonitoringDashboardUseCase) buildPerformanceDashboard(httpMetrics services.HTTPMetrics, dbMetrics services.DatabaseMetrics, cacheMetrics services.CacheMetrics) PerformanceDashboard {
	return PerformanceDashboard{
		HTTP: HTTPPerformance{
			TotalRequests:   httpMetrics.TotalRequests,
			RequestsPerSec:  float64(httpMetrics.TotalRequests) / 3600, // Simplified
			AverageResponse: int(httpMetrics.AverageResponseTime.Milliseconds()),
			P95Response:     int(httpMetrics.AverageResponseTime.Milliseconds()), // Simplified
			ErrorRate:       httpMetrics.ErrorRate,
			TopEndpoints:    []EndpointStats{}, // Placeholder
		},
		Database: DatabasePerformance{
			TotalQueries:     dbMetrics.TotalQueries,
			QueriesPerSec:    float64(dbMetrics.TotalQueries) / 3600, // Simplified
			AverageQueryTime: int(dbMetrics.AverageQueryTime.Milliseconds()),
			SlowQueries:      dbMetrics.SlowQueries,
			ConnectionsActive: int(dbMetrics.ConnectionsActive),
			ConnectionsIdle:   int(dbMetrics.ConnectionsIdle),
		},
		Cache: CachePerformance{
			TotalOperations:  cacheMetrics.TotalOperations,
			OperationsPerSec: float64(cacheMetrics.TotalOperations) / 3600, // Simplified
			HitRate:         cacheMetrics.HitRatio,
			AverageTime:      int(cacheMetrics.AverageResponseTime.Milliseconds()),
			MissRate:        100 - cacheMetrics.HitRatio,
		},
		ResponseTime: ResponseTimeMetrics{
			Average: int(httpMetrics.AverageResponseTime.Milliseconds()),
			P50:     int(httpMetrics.AverageResponseTime.Milliseconds()), // Simplified
			P95:     int(httpMetrics.AverageResponseTime.Milliseconds()), // Simplified
			P99:     int(httpMetrics.AverageResponseTime.Milliseconds()), // Simplified
		},
		ErrorRate: ErrorRateMetrics{
			Overall:  httpMetrics.ErrorRate,
			ByStatus: map[string]float64{}, // Placeholder
			ByPath:   map[string]float64{}, // Placeholder
		},
		Throughput: ThroughputMetrics{
			RequestsPerSec:   float64(httpMetrics.TotalRequests) / 3600,
			QueriesPerSec:    float64(dbMetrics.TotalQueries) / 3600,
			OperationsPerSec: float64(cacheMetrics.TotalOperations) / 3600,
		},
	}
}

func (uc *GetMonitoringDashboardUseCase) buildBusinessDashboard(businessMetrics services.BusinessMetrics) BusinessDashboard {
	return BusinessDashboard{
		Users: UserMetrics{
			Total:         businessMetrics.TotalUsers,
			Active:        businessMetrics.ActiveUsers,
			NewToday:      0, // Placeholder
			NewThisWeek:   0, // Placeholder
			NewThisMonth:  0, // Placeholder
			PremiumUsers:  businessMetrics.PremiumSubscriptions,
			VerifiedUsers: 0, // Placeholder
		},
		Engagement: EngagementMetrics{
			TotalMatches:    businessMetrics.TotalMatches,
			TotalMessages:   businessMetrics.TotalMessages,
			TotalSwipes:     businessMetrics.TotalSwipes,
			SwipesPerUser:   0, // Placeholder
			MessagesPerUser:  0, // Placeholder
			MatchRate:        0, // Placeholder
		},
		Revenue: RevenueMetrics{
			TotalRevenue:   0, // Placeholder
			MRR:            0, // Placeholder
			ARPU:           0, // Placeholder
			ConversionRate: 0, // Placeholder
			ChurnRate:      0, // Placeholder
		},
		Growth: GrowthMetrics{
			UserGrowth:       0, // Placeholder
			RevenueGrowth:     0, // Placeholder
			EngagementGrowth:  0, // Placeholder
			WeeklyGrowth:      []float64{}, // Placeholder
		},
	}
}

func (uc *GetMonitoringDashboardUseCase) buildSystemDashboard(systemMetrics services.SystemMetrics, systemHealth services.SystemHealth) SystemDashboard {
	return SystemDashboard{
		Resources: ResourceDashboard{
			CPU: ResourceUsage{
				Current:   systemHealth.System.CPU.Usage,
				Threshold: systemHealth.System.CPU.Threshold,
				Status:    string(systemHealth.System.CPU.Status),
				Trend:     "stable", // Placeholder
			},
			Memory: ResourceUsage{
				Current:   systemHealth.System.Memory.Usage,
				Threshold: systemHealth.System.Memory.Threshold,
				Status:    string(systemHealth.System.Memory.Status),
				Trend:     "stable", // Placeholder
			},
			Disk: ResourceUsage{
				Current:   systemHealth.System.Disk.Usage,
				Threshold: systemHealth.System.Disk.Threshold,
				Status:    string(systemHealth.System.Disk.Status),
				Trend:     "stable", // Placeholder
			},
			Network: NetworkUsage{
				BytesSent:    systemMetrics.NetworkIO.BytesSent,
				BytesRecv:    systemMetrics.NetworkIO.BytesRecv,
				BandwidthUsed: 0, // Placeholder
			},
		},
		Performance: SystemPerformance{
			Goroutines:  systemMetrics.Goroutines,
			GCCycles:    systemMetrics.GCCycles,
			LoadAverage: 0, // Placeholder
		},
		Uptime: UptimeMetrics{
			Current:      "healthy", // Simplified
			Percentage:   99.9,     // Placeholder
			LastDowntime: "",       // Placeholder
			TotalDowntime: "",       // Placeholder
		},
	}
}