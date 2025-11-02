package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// EphemeralPhotoRoutes defines routes for ephemeral photo endpoints
type EphemeralPhotoRoutes struct {
	handler *handlers.EphemeralPhotoHandler
}

// NewEphemeralPhotoRoutes creates new ephemeral photo routes
func NewEphemeralPhotoRoutes(handler *handlers.EphemeralPhotoHandler) *EphemeralPhotoRoutes {
	return &EphemeralPhotoRoutes{
		handler: handler,
	}
}

// RegisterRoutes registers ephemeral photo routes
func (r *EphemeralPhotoRoutes) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Group for ephemeral photo routes
	ephemeralPhotoGroup := router.Group("/api/v1/ephemeral-photos")
	{
		// Authenticated routes
		ephemeralPhotoGroup.Use(authMiddleware)
		
		// Upload ephemeral photo
		ephemeralPhotoGroup.POST("", r.handler.UploadEphemeralPhoto)
		
		// Get user's ephemeral photos
		ephemeralPhotoGroup.GET("", r.handler.GetUserEphemeralPhotos)
		
		// Delete ephemeral photo
		ephemeralPhotoGroup.DELETE("/:id", r.handler.DeleteEphemeralPhoto)
		
		// Get ephemeral photo status
		ephemeralPhotoGroup.GET("/:id/status", r.handler.GetEphemeralPhotoStatus)
		
		// Manually expire ephemeral photo
		ephemeralPhotoGroup.POST("/:id/expire", r.handler.ExpireEphemeralPhoto)
		
		// Get ephemeral photo analytics
		ephemeralPhotoGroup.GET("/analytics", r.handler.GetEphemeralPhotoAnalytics)
	}

	// Public routes (no authentication required)
	publicGroup := router.Group("/api/v1/ephemeral-photos")
	{
		// View ephemeral photo (public access via access key)
		publicGroup.GET("/:accessKey/view", r.handler.ViewEphemeralPhoto)
		
		// Track photo view (for analytics)
		publicGroup.POST("/track", r.handler.TrackPhotoView)
	}

	// Apply rate limiting to all ephemeral photo routes
	ephemeralPhotoGroup.Use(middleware.EphemeralPhotoRateLimit())
	publicGroup.Use(middleware.EphemeralPhotoRateLimit())
}

// GetRouteConfig returns configuration for ephemeral photo routes
func (r *EphemeralPhotoRoutes) GetRouteConfig() map[string]interface{} {
	return map[string]interface{}{
		"upload": map[string]interface{}{
			"path":        "/api/v1/ephemeral-photos",
			"method":      "POST",
			"auth":        true,
			"rate_limit":  map[string]int{"requests": 10, "window": 60}, // 10 uploads per minute
			"max_file_size": 5 * 1024 * 1024, // 5MB
			"allowed_types": []string{"image/jpeg", "image/png", "image/webp"},
		},
		"view": map[string]interface{}{
			"path":       "/api/v1/ephemeral-photos/:accessKey/view",
			"method":     "GET",
			"auth":       false,
			"rate_limit": map[string]int{"requests": 100, "window": 60}, // 100 views per minute
			"cache_control": "no-cache, no-store, must-revalidate",
		},
		"delete": map[string]interface{}{
			"path":       "/api/v1/ephemeral-photos/:id",
			"method":     "DELETE",
			"auth":       true,
			"rate_limit": map[string]int{"requests": 30, "window": 60}, // 30 deletes per minute
		},
		"status": map[string]interface{}{
			"path":       "/api/v1/ephemeral-photos/:id/status",
			"method":     "GET",
			"auth":       true,
			"rate_limit": map[string]int{"requests": 100, "window": 60}, // 100 status checks per minute
		},
		"expire": map[string]interface{}{
			"path":       "/api/v1/ephemeral-photos/:id/expire",
			"method":     "POST",
			"auth":       true,
			"rate_limit": map[string]int{"requests": 20, "window": 60}, // 20 expires per minute
		},
		"analytics": map[string]interface{}{
			"path":       "/api/v1/ephemeral-photos/analytics",
			"method":     "GET",
			"auth":       true,
			"rate_limit": map[string]int{"requests": 50, "window": 60}, // 50 analytics requests per minute
		},
	}
}

// GetSecurityHeaders returns security headers for ephemeral photo routes
func (r *EphemeralPhotoRoutes) GetSecurityHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy":  "default-src 'self'; script-src 'none'; object-src 'none'",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Permissions-Policy":       "geolocation=(), microphone=(), camera=()",
	}
}

// GetCacheHeaders returns cache headers for ephemeral photo routes
func (r *EphemeralPhotoRoutes) GetCacheHeaders() map[string]string {
	return map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}
}

// GetRateLimitConfig returns rate limit configuration for different operations
func (r *EphemeralPhotoRoutes) GetRateLimitConfig() map[string]map[string]int {
	return map[string]map[string]int{
		"upload": {
			"requests": 10,
			"window":   60, // 1 minute
		},
		"view": {
			"requests": 100,
			"window":   60, // 1 minute
		},
		"delete": {
			"requests": 30,
			"window":   60, // 1 minute
		},
		"status": {
			"requests": 100,
			"window":   60, // 1 minute
		},
		"expire": {
			"requests": 20,
			"window":   60, // 1 minute
		},
		"analytics": {
			"requests": 50,
			"window":   60, // 1 minute
		},
	}
}

// GetCORSConfig returns CORS configuration for ephemeral photo routes
func (r *EphemeralPhotoRoutes) GetCORSConfig() map[string]interface{} {
	return map[string]interface{}{
		"allowed_origins": []string{"https://yourdomain.com", "https://www.yourdomain.com"},
		"allowed_methods": []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		"allowed_headers": []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
		},
		"exposed_headers": []string{
			"Content-Length",
			"Content-Type",
			"X-Total-Count",
			"X-Rate-Limit-Limit",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},
		"max_age": 86400, // 24 hours
		"credentials": true,
	}
}

// GetValidationRules returns validation rules for ephemeral photo operations
func (r *EphemeralPhotoRoutes) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{
		"upload": map[string]interface{}{
			"max_views": map[string]interface{}{
				"min": 1,
				"max": 10,
			},
			"expires_after_seconds": map[string]interface{}{
				"min": 5,   // 5 seconds minimum
				"max": 300, // 5 minutes maximum
			},
			"file": map[string]interface{}{
				"max_size": 5 * 1024 * 1024, // 5MB
				"allowed_types": []string{
					"image/jpeg",
					"image/png", 
					"image/webp",
				},
				"min_dimensions": map[string]int{
					"width":  100,
					"height": 100,
				},
				"max_dimensions": map[string]int{
					"width":  4096,
					"height": 4096,
				},
			},
		},
	}
}

// GetDefaultExpirationTimes returns default expiration time options
func (r *EphemeralPhotoRoutes) GetDefaultExpirationTimes() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"seconds": 5,
			"label":   "5 seconds",
			"description": "Very short - for quick sharing",
		},
		{
			"seconds": 15,
			"label":   "15 seconds",
			"description": "Short - for quick preview",
		},
		{
			"seconds": 30,
			"label":   "30 seconds",
			"description": "Standard - default option",
		},
		{
			"seconds": 60,
			"label":   "1 minute",
			"description": "Extended - for longer viewing",
		},
		{
			"seconds": 300,
			"label":   "5 minutes",
			"description": "Maximum - for special cases",
		},
	}
}

// GetSecurityFeatures returns security features for ephemeral photos
func (r *EphemeralPhotoRoutes) GetSecurityFeatures() map[string]interface{} {
	return map[string]interface{}{
		"access_control": map[string]interface{}{
			"one_time_view":     true,
			"access_key_length":  32,
			"key_expiration":    "24 hours",
		},
		"view_protection": map[string]interface{}{
			"prevent_download":   true,
			"prevent_screenshot":  "client_side",
			"watermark":         true,
			"no_cache":          true,
		},
		"rate_limiting": map[string]interface{}{
			"per_user":    true,
			"per_ip":      true,
			"sliding_window": true,
		},
		"audit_trail": map[string]interface{}{
			"view_tracking":    true,
			"ip_logging":      true,
			"user_agent":      true,
			"duration_tracking": true,
		},
	}
}