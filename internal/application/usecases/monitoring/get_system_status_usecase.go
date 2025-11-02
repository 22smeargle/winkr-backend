package monitoring

import (
	"context"
	"time"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetSystemStatusRequest represents a request to get system status
type GetSystemStatusRequest struct {
	IncludeResources bool   `json:"include_resources"`
	IncludeMetrics   bool   `json:"include_metrics"`
	IncludeHistory  bool   `json:"include_history"`
	TimeRange        string `json:"time_range"`        // 1h, 24h, 7d, etc.
}

// GetSystemStatusResponse represents the response for system status
type GetSystemStatusResponse struct {
	Status       string                 `json:"status"`
	Timestamp    time.Time               `json:"timestamp"`
	Uptime       time.Duration           `json:"uptime"`
	Version      string                 `json:"version"`
	Resources   *SystemResources        `json:"resources,omitempty"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	History     []SystemStatusHistory   `json:"history,omitempty"`
}

// SystemResources represents system resource information
type SystemResources struct {
	CPU    ResourceInfo `json:"cpu"`
	Memory ResourceInfo `json:"memory"`
	Disk   ResourceInfo `json:"disk"`
	Network NetworkInfo  `json:"network"`
}

// ResourceInfo represents resource information
type ResourceInfo struct {
	Usage    float64 `json:"usage"`
	Status    string  `json:"status"`
	Threshold float64 `json:"threshold"`
	Available float64 `json:"available"`
	Total     float64 `json:"total"`
}

// NetworkInfo represents network information
type NetworkInfo struct {
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
	Status    string `json:"status"`
}

// SystemStatusHistory represents historical system status
type SystemStatusHistory struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string     `json:"status"`
	CPU       float64    `json:"cpu"`
	Memory    float64    `json:"memory"`
	Disk       float64    `json:"disk"`
}

// GetSystemStatusUseCase handles getting system status
type GetSystemStatusUseCase struct {
	healthCheckService *services.HealthCheckService
	metricsService    *services.MetricsService
}

// NewGetSystemStatusUseCase creates a new GetSystemStatusUseCase
func NewGetSystemStatusUseCase(
	healthCheckService *services.HealthCheckService,
	metricsService *services.MetricsService,
) *GetSystemStatusUseCase {
	return &GetSystemStatusUseCase{
		healthCheckService: healthCheckService,
		metricsService:    metricsService,
	}
}

// Execute executes the get system status use case
func (uc *GetSystemStatusUseCase) Execute(ctx context.Context, req GetSystemStatusRequest) (*GetSystemStatusResponse, error) {
	logger.Info("Getting system status", map[string]interface{}{
		"include_resources": req.IncludeResources,
		"include_metrics":   req.IncludeMetrics,
		"include_history":  req.IncludeHistory,
		"time_range":        req.TimeRange,
	})

	// Get overall health
	systemHealth := uc.healthCheckService.GetOverallHealth(ctx)

	// Prepare response
	response := &GetSystemStatusResponse{
		Status:    string(systemHealth.Status),
		Timestamp: systemHealth.Timestamp,
		Uptime:    systemHealth.Uptime,
		Version:   systemHealth.Version,
	}

	// Include resources if requested
	if req.IncludeResources {
		systemMetrics := uc.metricsService.GetSystemMetrics()
		response.Resources = &SystemResources{
			CPU: ResourceInfo{
				Usage:    systemHealth.System.CPU.Usage,
				Status:    string(systemHealth.System.CPU.Status),
				Threshold: systemHealth.System.CPU.Threshold,
				Available: 100 - systemHealth.System.CPU.Usage,
				Total:     100,
			},
			Memory: ResourceInfo{
				Usage:    systemHealth.System.Memory.Usage,
				Status:    string(systemHealth.System.Memory.Status),
				Threshold: systemHealth.System.Memory.Threshold,
				Available: 100 - systemHealth.System.Memory.Usage,
				Total:     100,
			},
			Disk: ResourceInfo{
				Usage:    systemHealth.System.Disk.Usage,
				Status:    string(systemHealth.System.Disk.Status),
				Threshold: systemHealth.System.Disk.Threshold,
				Available: 100 - systemHealth.System.Disk.Usage,
				Total:     100,
			},
			Network: NetworkInfo{
				BytesSent: systemMetrics.NetworkIO.BytesSent,
				BytesRecv: systemMetrics.NetworkIO.BytesRecv,
				Status:    "healthy", // Simplified
			},
		}
	}

	// Include metrics if requested
	if req.IncludeMetrics {
		response.Metrics = map[string]interface{}{
			"http":     uc.metricsService.GetHTTPMetrics(),
			"database": uc.metricsService.GetDatabaseMetrics(),
			"cache":    uc.metricsService.GetCacheMetrics(),
			"business": uc.metricsService.GetBusinessMetrics(),
		}
	}

	// Include history if requested
	if req.IncludeHistory {
		// This is a simplified history - in production, you'd query from time-series storage
		response.History = []SystemStatusHistory{
			{
				Timestamp: time.Now().Add(-time.Hour),
				Status:    "healthy",
				CPU:       45.2,
				Memory:    62.8,
				Disk:       35.1,
			},
			{
				Timestamp: time.Now().Add(-2 * time.Hour),
				Status:    "healthy",
				CPU:       42.1,
				Memory:    60.3,
				Disk:       35.0,
			},
		}
	}

	return response, nil
}