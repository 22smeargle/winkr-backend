package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
)

// MonitoringRoutes defines monitoring routes
type MonitoringRoutes struct {
	handler *handlers.MonitoringHandler
}

// NewMonitoringRoutes creates a new MonitoringRoutes instance
func NewMonitoringRoutes(handler *handlers.MonitoringHandler) *MonitoringRoutes {
	return &MonitoringRoutes{
		handler: handler,
	}
}

// RegisterRoutes registers monitoring routes
func (r *MonitoringRoutes) RegisterRoutes(router *gin.RouterGroup) {
	monitoring := router.Group("/monitoring")
	{
		// Health check endpoints
		monitoring.GET("/health", r.handler.HandleHealth)
		monitoring.GET("/health/detailed", r.handler.HandleDetailedHealth)
		monitoring.GET("/health/components/:component", r.handler.HandleComponentHealth)
		
		// Metrics endpoints
		monitoring.GET("/metrics", r.handler.HandleMetrics)
		monitoring.GET("/metrics/json", r.handler.HandleMetricsJSON)
		monitoring.GET("/metrics/history", r.handler.HandleMetricsHistory)
		
		// Monitoring dashboard endpoints
		monitoring.GET("/status", r.handler.HandleMonitoringStatus)
		monitoring.GET("/alerts", r.handler.HandleAlerts)
		monitoring.GET("/system", r.handler.HandleSystemInfo)
	}
	
	// Also register health endpoints at root level for compatibility
	health := router.Group("/health")
	{
		health.GET("/", r.handler.HandleHealth)
		health.GET("/detailed", r.handler.HandleDetailedHealth)
		health.GET("/components/:component", r.handler.HandleComponentHealth)
	}
	
	// Register metrics endpoint at root level for Prometheus compatibility
	router.GET("/metrics", r.handler.HandleMetrics)
}