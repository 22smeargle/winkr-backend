package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/utils/response"
)

// MonitoringHandler handles monitoring endpoints
type MonitoringHandler struct {
	healthCheckService *services.HealthCheckService
	metricsService    *services.MetricsService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(
	healthCheckService *services.HealthCheckService,
	metricsService *services.MetricsService,
) *MonitoringHandler {
	return &MonitoringHandler{
		healthCheckService: healthCheckService,
		metricsService:    metricsService,
	}
}

// HandleHealth handles basic health check endpoint
// GET /health
func (h *MonitoringHandler) HandleHealth(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get basic health status
	health := h.healthCheckService.GetBasicHealth(ctx)
	
	// Set appropriate status code
	statusCode := http.StatusOK
	if health.Status == services.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == services.HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}
	
	// Add health check headers
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	c.JSON(statusCode, response.SuccessResponse(map[string]interface{}{
		"status":       health.Status,
		"timestamp":    health.Timestamp,
		"response_time": health.ResponseTime.Milliseconds(),
		"error":        health.Error,
	}))
}

// HandleDetailedHealth handles detailed health check endpoint
// GET /health/detailed
func (h *MonitoringHandler) HandleDetailedHealth(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get detailed health status
	systemHealth := h.healthCheckService.GetOverallHealth(ctx)
	
	// Set appropriate status code
	statusCode := http.StatusOK
	if systemHealth.Status == services.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if systemHealth.Status == services.HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}
	
	// Add health check headers
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	c.JSON(statusCode, response.SuccessResponse(systemHealth))
}

// HandleMetrics handles Prometheus metrics endpoint
// GET /metrics
func (h *MonitoringHandler) HandleMetrics(c *gin.Context) {
	// Get Prometheus format metrics
	metrics := h.metricsService.GetPrometheusMetrics()
	
	// Set appropriate headers for Prometheus
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	
	c.String(http.StatusOK, metrics)
}

// HandleMetricsJSON handles JSON metrics endpoint
// GET /metrics/json
func (h *MonitoringHandler) HandleMetricsJSON(c *gin.Context) {
	// Get all metrics
	httpMetrics := h.metricsService.GetHTTPMetrics()
	dbMetrics := h.metricsService.GetDatabaseMetrics()
	cacheMetrics := h.metricsService.GetCacheMetrics()
	businessMetrics := h.metricsService.GetBusinessMetrics()
	systemMetrics := h.metricsService.GetSystemMetrics()
	
	// Combine all metrics
	allMetrics := map[string]interface{}{
		"timestamp": time.Now(),
		"http":      httpMetrics,
		"database":  dbMetrics,
		"cache":     cacheMetrics,
		"business":  businessMetrics,
		"system":    systemMetrics,
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse(allMetrics))
}

// HandleMonitoringStatus handles monitoring dashboard data endpoint
// GET /monitoring/status
func (h *MonitoringHandler) HandleMonitoringStatus(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get query parameters
	includeDetails := c.DefaultQuery("details", "false") == "true"
	timeRange := c.DefaultQuery("range", "1h")
	
	// Parse time range
	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		duration = time.Hour // Default to 1 hour
	}
	
	// Get health status
	health := h.healthCheckService.GetOverallHealth(ctx)
	
	// Get metrics
	metrics := map[string]interface{}{
		"http":     h.metricsService.GetHTTPMetrics(),
		"database": h.metricsService.GetDatabaseMetrics(),
		"cache":    h.metricsService.GetCacheMetrics(),
	}
	
	// Include business and system metrics if requested
	if includeDetails {
		metrics["business"] = h.metricsService.GetBusinessMetrics()
		metrics["system"] = h.metricsService.GetSystemMetrics()
	}
	
	// Calculate uptime percentage (simplified)
	uptimePercentage := 100.0
	if health.Status == services.HealthStatusDegraded {
		uptimePercentage = 95.0
	} else if health.Status == services.HealthStatusUnhealthy {
		uptimePercentage = 0.0
	}
	
	// Prepare dashboard data
	dashboard := map[string]interface{}{
		"status": map[string]interface{}{
			"overall":           health.Status,
			"uptime_percentage":  uptimePercentage,
			"uptime":            health.Uptime.String(),
			"last_check":        health.Timestamp,
			"version":           health.Version,
		},
		"components": health.Components,
		"dependencies": health.Dependencies,
		"system_resources": health.System,
		"metrics": metrics,
		"time_range": timeRange,
		"generated_at": time.Now(),
	}
	
	// Add performance summary
	if includeDetails {
		dashboard["performance"] = map[string]interface{}{
			"average_response_time": health.Components["database"].ResponseTime.Milliseconds(),
			"error_rate":          h.metricsService.GetHTTPMetrics().ErrorRate,
			"requests_per_second": h.calculateRequestsPerSecond(duration),
		}
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse(dashboard))
}

// HandleComponentHealth handles health check for specific component
// GET /health/components/{component}
func (h *MonitoringHandler) HandleComponentHealth(c *gin.Context) {
	ctx := c.Request.Context()
	component := c.Param("component")
	
	var health services.HealthCheck
	var found bool
	
	// Check specific component
	switch component {
	case "database":
		if h.healthCheckService != nil {
			health = h.healthCheckService.CheckDatabaseHealth(ctx)
			found = true
		}
	case "redis":
		if h.healthCheckService != nil {
			health = h.healthCheckService.CheckRedisHealth(ctx)
			found = true
		}
	case "storage":
		if h.healthCheckService != nil {
			health = h.healthCheckService.CheckStorageHealth(ctx)
			found = true
		}
	case "stripe":
		if h.healthCheckService != nil {
			health = h.healthCheckService.CheckStripeHealth(ctx)
			found = true
		}
	}
	
	if !found {
		c.JSON(http.StatusNotFound, response.ErrorResponse("Component not found"))
		return
	}
	
	// Set appropriate status code
	statusCode := http.StatusOK
	if health.Status == services.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == services.HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}
	
	c.JSON(statusCode, response.SuccessResponse(health))
}

// HandleMetricsHistory handles historical metrics endpoint
// GET /metrics/history
func (h *MonitoringHandler) HandleMetricsHistory(c *gin.Context) {
	// Get query parameters
	metricType := c.DefaultQuery("type", "http")
	timeRange := c.DefaultQuery("range", "1h")
	limitStr := c.DefaultQuery("limit", "100")
	
	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	
	// Parse time range
	duration, err := time.ParseDuration(timeRange)
	if err != nil {
		duration = time.Hour // Default to 1 hour
	}
	
	// Get historical metrics (simplified - in production, you'd query from time-series storage)
	history := map[string]interface{}{
		"metric_type": metricType,
		"time_range":  timeRange,
		"limit":       limit,
		"data":        []map[string]interface{}{}, // Placeholder
		"generated_at": time.Now(),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse(history))
}

// HandleAlerts handles alerts endpoint
// GET /monitoring/alerts
func (h *MonitoringHandler) HandleAlerts(c *gin.Context) {
	// Get query parameters
	status := c.DefaultQuery("status", "active")
	severity := c.DefaultQuery("severity", "")
	limitStr := c.DefaultQuery("limit", "50")
	
	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	
	// Get alerts (simplified - in production, you'd query from alert storage)
	alerts := map[string]interface{}{
		"status":  status,
		"severity": severity,
		"limit":    limit,
		"alerts":   []map[string]interface{}{}, // Placeholder
		"generated_at": time.Now(),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse(alerts))
}

// HandleSystemInfo handles system information endpoint
// GET /monitoring/system
func (h *MonitoringHandler) HandleSystemInfo(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get system health
	health := h.healthCheckService.GetOverallHealth(ctx)
	
	// Get system metrics
	systemMetrics := h.metricsService.GetSystemMetrics()
	
	// Prepare system information
	systemInfo := map[string]interface{}{
		"status": health.Status,
		"uptime": health.Uptime.String(),
		"version": health.Version,
		"resources": map[string]interface{}{
			"cpu": map[string]interface{}{
				"usage":    systemMetrics.CPUUsage,
				"status":   health.System.CPU.Status,
				"threshold": health.System.CPU.Threshold,
			},
			"memory": map[string]interface{}{
				"usage":    systemMetrics.MemoryUsage,
				"total":    systemMetrics.MemoryTotal,
				"allocated": systemMetrics.MemoryAlloc,
				"status":   health.System.Memory.Status,
				"threshold": health.System.Memory.Threshold,
			},
			"disk": map[string]interface{}{
				"usage":    systemMetrics.DiskUsage,
				"status":   health.System.Disk.Status,
				"threshold": health.System.Disk.Threshold,
			},
			"goroutines": systemMetrics.Goroutines,
			"gc_cycles":  systemMetrics.GCCycles,
		},
		"network_io": systemMetrics.NetworkIO,
		"last_update": systemMetrics.LastUpdateTime,
		"generated_at": time.Now(),
	}
	
	c.JSON(http.StatusOK, response.SuccessResponse(systemInfo))
}

// Helper methods

func (h *MonitoringHandler) calculateRequestsPerSecond(duration time.Duration) float64 {
	httpMetrics := h.metricsService.GetHTTPMetrics()
	if duration.Seconds() <= 0 {
		return 0
	}
	return float64(httpMetrics.TotalRequests) / duration.Seconds()
}