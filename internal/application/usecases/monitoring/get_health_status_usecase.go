package monitoring

import (
	"context"
	"time"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetHealthStatusRequest represents a request to get health status
type GetHealthStatusRequest struct {
	IncludeDetails bool `json:"include_details"`
	Components    []string `json:"components,omitempty"`
}

// GetHealthStatusResponse represents the response for health status
type GetHealthStatusResponse struct {
	Status       string                 `json:"status"`
	Timestamp    time.Time               `json:"timestamp"`
	Uptime       time.Duration           `json:"uptime"`
	Version      string                 `json:"version"`
	Components   map[string]interface{} `json:"components,omitempty"`
	System       map[string]interface{} `json:"system,omitempty"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
}

// GetHealthStatusUseCase handles getting health status
type GetHealthStatusUseCase struct {
	healthCheckService *services.HealthCheckService
}

// NewGetHealthStatusUseCase creates a new GetHealthStatusUseCase
func NewGetHealthStatusUseCase(healthCheckService *services.HealthCheckService) *GetHealthStatusUseCase {
	return &GetHealthStatusUseCase{
		healthCheckService: healthCheckService,
	}
}

// Execute executes the get health status use case
func (uc *GetHealthStatusUseCase) Execute(ctx context.Context, req GetHealthStatusRequest) (*GetHealthStatusResponse, error) {
	logger.Info("Getting health status", map[string]interface{}{
		"include_details": req.IncludeDetails,
		"components":     req.Components,
	})

	var healthData interface{}
	
	if req.IncludeDetails {
		// Get detailed health information
		healthData = uc.healthCheckService.GetOverallHealth(ctx)
	} else {
		// Get basic health information
		healthData = uc.healthCheckService.GetBasicHealth(ctx)
	}

	// Convert to response format
	switch health := healthData.(type) {
	case services.SystemHealth:
		return &GetHealthStatusResponse{
			Status:       string(health.Status),
			Timestamp:    health.Timestamp,
			Uptime:       health.Uptime,
			Version:      health.Version,
			Components:   convertComponents(health.Components),
			System:       convertSystemHealth(health.System),
			Dependencies: convertDependencies(health.Dependencies),
		}, nil
		
	case services.HealthCheck:
		return &GetHealthStatusResponse{
			Status:    string(health.Status),
			Timestamp: health.Timestamp,
		}, nil
		
	default:
		return nil, ErrInvalidHealthData
	}
}

// Helper functions

func convertComponents(components map[string]services.HealthCheck) map[string]interface{} {
	result := make(map[string]interface{})
	for name, component := range components {
		result[name] = map[string]interface{}{
			"status":        component.Status,
			"timestamp":     component.Timestamp,
			"response_time": component.ResponseTime.Milliseconds(),
			"error":         component.Error,
			"details":       component.Details,
		}
	}
	return result
}

func convertSystemHealth(system services.SystemResourceHealth) map[string]interface{} {
	return map[string]interface{}{
		"cpu": map[string]interface{}{
			"usage":    system.CPU.Usage,
			"status":    system.CPU.Status,
			"threshold": system.CPU.Threshold,
		},
		"memory": map[string]interface{}{
			"usage":    system.Memory.Usage,
			"status":    system.Memory.Status,
			"threshold": system.Memory.Threshold,
		},
		"disk": map[string]interface{}{
			"usage":    system.Disk.Usage,
			"status":    system.Disk.Status,
			"threshold": system.Disk.Threshold,
		},
	}
}

func convertDependencies(dependencies map[string]services.DependencyHealth) map[string]interface{} {
	result := make(map[string]interface{})
	for name, dep := range dependencies {
		result[name] = map[string]interface{}{
			"status":        dep.Status,
			"response_time": dep.ResponseTime.Milliseconds(),
			"error":         dep.Error,
		}
	}
	return result
}