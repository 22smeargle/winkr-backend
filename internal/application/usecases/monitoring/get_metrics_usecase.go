package monitoring

import (
	"context"
	"time"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetMetricsRequest represents a request to get metrics
type GetMetricsRequest struct {
	Type      string `json:"type"`      // http, database, cache, business, system
	TimeRange string `json:"time_range"` // 1h, 24h, 7d, etc.
	Format    string `json:"format"`    // json, prometheus
	Limit     int    `json:"limit"`     // number of data points
}

// GetMetricsResponse represents a response for metrics
type GetMetricsResponse struct {
	Type      string                 `json:"type"`
	TimeRange string                 `json:"time_range"`
	Format    string                 `json:"format"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// GetMetricsUseCase handles getting metrics
type GetMetricsUseCase struct {
	metricsService *services.MetricsService
}

// NewGetMetricsUseCase creates a new GetMetricsUseCase
func NewGetMetricsUseCase(metricsService *services.MetricsService) *GetMetricsUseCase {
	return &GetMetricsUseCase{
		metricsService: metricsService,
	}
}

// Execute executes the get metrics use case
func (uc *GetMetricsUseCase) Execute(ctx context.Context, req GetMetricsRequest) (*GetMetricsResponse, error) {
	logger.Info("Getting metrics", map[string]interface{}{
		"type":       req.Type,
		"time_range": req.TimeRange,
		"format":     req.Format,
		"limit":      req.Limit,
	})

	// Validate request
	if req.Type == "" {
		req.Type = "all"
	}
	if req.Format == "" {
		req.Format = "json"
	}

	// Get metrics based on type
	var data map[string]interface{}
	
	switch req.Type {
	case "http":
		data = map[string]interface{}{
			"http": uc.metricsService.GetHTTPMetrics(),
		}
	case "database":
		data = map[string]interface{}{
			"database": uc.metricsService.GetDatabaseMetrics(),
		}
	case "cache":
		data = map[string]interface{}{
			"cache": uc.metricsService.GetCacheMetrics(),
		}
	case "business":
		data = map[string]interface{}{
			"business": uc.metricsService.GetBusinessMetrics(),
		}
	case "system":
		data = map[string]interface{}{
			"system": uc.metricsService.GetSystemMetrics(),
		}
	case "all":
		data = map[string]interface{}{
			"http":     uc.metricsService.GetHTTPMetrics(),
			"database": uc.metricsService.GetDatabaseMetrics(),
			"cache":    uc.metricsService.GetCacheMetrics(),
			"business": uc.metricsService.GetBusinessMetrics(),
			"system":   uc.metricsService.GetSystemMetrics(),
		}
	default:
		return nil, ErrInvalidMetricType
	}

	// Handle different formats
	if req.Format == "prometheus" {
		// For Prometheus format, we need to convert the data
		prometheusData := uc.metricsService.GetPrometheusMetrics()
		data = map[string]interface{}{
			"prometheus": prometheusData,
		}
	}

	return &GetMetricsResponse{
		Type:      req.Type,
		TimeRange: req.TimeRange,
		Format:    req.Format,
		Data:      data,
		Timestamp: time.Now(),
	}, nil
}