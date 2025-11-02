package monitoring

import (
	"context"
	"time"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetDetailedHealthRequest represents a request to get detailed health status
type GetDetailedHealthRequest struct {
	Components    []string `json:"components,omitempty"`
	IncludeSystem bool     `json:"include_system"`
	IncludeDeps   bool     `json:"include_dependencies"`
}

// GetDetailedHealthResponse represents the response for detailed health status
type GetDetailedHealthResponse struct {
	Status       string                         `json:"status"`
	Timestamp    time.Time                       `json:"timestamp"`
	Uptime       time.Duration                   `json:"uptime"`
	Version      string                         `json:"version"`
	Components   map[string]ComponentHealth      `json:"components"`
	System        *SystemResourceHealth           `json:"system,omitempty"`
	Dependencies map[string]DependencyHealth      `json:"dependencies,omitempty"`
}

// ComponentHealth represents detailed component health
type ComponentHealth struct {
	Status       string                 `json:"status"`
	Timestamp    time.Time               `json:"timestamp"`
	ResponseTime time.Duration           `json:"response_time"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// DependencyHealth represents external dependency health
type DependencyHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// SystemResourceHealth represents system resource health
type SystemResourceHealth struct {
	CPU    ResourceHealth `json:"cpu"`
	Memory ResourceHealth `json:"memory"`
	Disk   ResourceHealth `json:"disk"`
}

// ResourceHealth represents resource health
type ResourceHealth struct {
	Usage    float64 `json:"usage"`
	Status    string  `json:"status"`
	Threshold float64 `json:"threshold"`
}

// GetDetailedHealthUseCase handles getting detailed health status
type GetDetailedHealthUseCase struct {
	healthCheckService *services.HealthCheckService
}

// NewGetDetailedHealthUseCase creates a new GetDetailedHealthUseCase
func NewGetDetailedHealthUseCase(healthCheckService *services.HealthCheckService) *GetDetailedHealthUseCase {
	return &GetDetailedHealthUseCase{
		healthCheckService: healthCheckService,
	}
}

// Execute executes the get detailed health use case
func (uc *GetDetailedHealthUseCase) Execute(ctx context.Context, req GetDetailedHealthRequest) (*GetDetailedHealthResponse, error) {
	logger.Info("Getting detailed health status", map[string]interface{}{
		"components":      req.Components,
		"include_system":  req.IncludeSystem,
		"include_deps":    req.IncludeDeps,
	})

	// Get overall health
	systemHealth := uc.healthCheckService.GetOverallHealth(ctx)

	// Filter components if specified
	components := make(map[string]ComponentHealth)
	if len(req.Components) > 0 {
		for _, componentName := range req.Components {
			if component, exists := systemHealth.Components[componentName]; exists {
				components[componentName] = ComponentHealth{
					Status:       string(component.Status),
					Timestamp:    component.Timestamp,
					ResponseTime: component.ResponseTime,
					Error:        component.Error,
					Details:      component.Details,
				}
			}
		}
	} else {
		// Include all components
		for name, component := range systemHealth.Components {
			components[name] = ComponentHealth{
				Status:       string(component.Status),
				Timestamp:    component.Timestamp,
				ResponseTime: component.ResponseTime,
				Error:        component.Error,
				Details:      component.Details,
			}
		}
	}

	// Prepare response
	response := &GetDetailedHealthResponse{
		Status:     string(systemHealth.Status),
		Timestamp:  systemHealth.Timestamp,
		Uptime:     systemHealth.Uptime,
		Version:    systemHealth.Version,
		Components: components,
	}

	// Include system resources if requested
	if req.IncludeSystem {
		response.System = &SystemResourceHealth{
			CPU: ResourceHealth{
				Usage:    systemHealth.System.CPU.Usage,
				Status:    string(systemHealth.System.CPU.Status),
				Threshold: systemHealth.System.CPU.Threshold,
			},
			Memory: ResourceHealth{
				Usage:    systemHealth.System.Memory.Usage,
				Status:    string(systemHealth.System.Memory.Status),
				Threshold: systemHealth.System.Memory.Threshold,
			},
			Disk: ResourceHealth{
				Usage:    systemHealth.System.Disk.Usage,
				Status:    string(systemHealth.System.Disk.Status),
				Threshold: systemHealth.System.Disk.Threshold,
			},
		}
	}

	// Include dependencies if requested
	if req.IncludeDeps {
		dependencies := make(map[string]DependencyHealth)
		for name, dep := range systemHealth.Dependencies {
			dependencies[name] = DependencyHealth{
				Status:       string(dep.Status),
				ResponseTime: dep.ResponseTime,
				Error:        dep.Error,
			}
		}
		response.Dependencies = dependencies
	}

	return response, nil
}