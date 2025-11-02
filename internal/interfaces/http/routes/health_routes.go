package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
)

// HealthRoutes defines health check routes
type HealthRoutes struct {
	handler *handlers.HealthHandler
}

// NewHealthRoutes creates a new HealthRoutes instance
func NewHealthRoutes() *HealthRoutes {
	return &HealthRoutes{}
}

// RegisterRoutes registers health check routes
func (r *HealthRoutes) RegisterRoutes(router *gin.RouterGroup) {
	health := router.Group("/health")
	{
		// Basic health checks
		health.GET("/", r.HandleHealthRequest)
		health.GET("/overall", r.HandleHealthRequest)
		health.GET("/redis", r.HandleHealthRequest)
		health.GET("/cache", r.HandleHealthRequest)
		health.GET("/sessions", r.HandleHealthRequest)
		health.GET("/rate_limit", r.HandleHealthRequest)
		health.GET("/pubsub", r.HandleHealthRequest)
		
		// Kubernetes probes
		health.GET("/liveness", r.HandleLivenessProbe)
		health.GET("/readiness", r.HandleReadinessProbe)
		
		// Metrics endpoint
		health.GET("/metrics", r.HandleMetrics)
	}
}

// HandleHealthRequest handles basic health check requests
func (r *HealthRoutes) HandleHealthRequest(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"timestamp": "2025-01-01T00:00:00Z",
	})
}

// HandleLivenessProbe handles Kubernetes liveness probe
func (r *HealthRoutes) HandleLivenessProbe(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"probe": "liveness",
	})
}

// HandleReadinessProbe handles Kubernetes readiness probe
func (r *HealthRoutes) HandleReadinessProbe(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"probe": "readiness",
	})
}

// HandleMetrics handles metrics requests
func (r *HealthRoutes) HandleMetrics(c *gin.Context) {
	c.JSON(200, gin.H{
		"metrics": "placeholder",
	})
}