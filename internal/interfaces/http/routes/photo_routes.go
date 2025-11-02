package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// PhotoRoutes defines photo-related routes
type PhotoRoutes struct {
	photoHandler *handlers.PhotoHandler
	rateLimiter  *middleware.RateLimiter
}

// NewPhotoRoutes creates a new photo routes instance
func NewPhotoRoutes(
	photoHandler *handlers.PhotoHandler,
	rateLimiter *middleware.RateLimiter,
) *PhotoRoutes {
	return &PhotoRoutes{
		photoHandler: photoHandler,
		rateLimiter:  rateLimiter,
	}
}

// RegisterRoutes registers photo routes
func (r *PhotoRoutes) RegisterRoutes(router *gin.RouterGroup, redisClient *redis.Client) {
	// Apply rate limiting for photo operations
	photoGroup := router.Group("/photos")
	photoGroup.Use(r.rateLimiter.LimitByUser("photo_operations", 10, 60)) // 10 operations per minute per user

	// Upload photo route
	photoGroup.POST("", r.photoHandler.UploadPhoto)

	// Delete photo route
	photoGroup.DELETE("/:photo_id", r.photoHandler.DeletePhoto)

	// Set primary photo route
	photoGroup.PUT("/:photo_id/set-primary", r.photoHandler.SetPrimaryPhoto)

	// Media routes (for upload/download URLs)
	mediaGroup := router.Group("/media")
	mediaGroup.Use(r.rateLimiter.LimitByUser("media_operations", 20, 60)) // 20 operations per minute per user

	// Get upload URL route
	mediaGroup.GET("/request-upload", r.photoHandler.GetUploadURL)

	// Get download URL route
	mediaGroup.GET("/:photo_id/url", r.photoHandler.GetDownloadURL)

	// Mark photo as viewed route
	mediaGroup.POST("/:photo_id/view", r.photoHandler.MarkPhotoViewed)
}

// GetPhotoRoutes returns all photo route definitions for documentation
func (r *PhotoRoutes) GetPhotoRoutes() []RouteDefinition {
	return []RouteDefinition{
		{
			Method:      "POST",
			Path:         "/me/photos",
			Description:  "Upload a new photo",
			AuthRequired: true,
			RateLimit:   "10 requests per minute per user",
		},
		{
			Method:      "DELETE",
			Path:         "/me/photos/{photo_id}",
			Description:  "Delete a photo",
			AuthRequired: true,
			RateLimit:   "10 requests per minute per user",
		},
		{
			Method:      "PUT",
			Path:         "/me/photos/{photo_id}/set-primary",
			Description:  "Set photo as primary",
			AuthRequired: true,
			RateLimit:   "10 requests per minute per user",
		},
		{
			Method:      "GET",
			Path:         "/media/request-upload",
			Description:  "Get upload URL for photo",
			AuthRequired: true,
			RateLimit:   "20 requests per minute per user",
		},
		{
			Method:      "GET",
			Path:         "/media/{photo_id}/url",
			Description:  "Get download URL for photo",
			AuthRequired: true,
			RateLimit:   "20 requests per minute per user",
		},
		{
			Method:      "POST",
			Path:         "/media/{photo_id}/view",
			Description:  "Mark photo as viewed",
			AuthRequired: true,
			RateLimit:   "20 requests per minute per user",
		},
	}
}

// RouteDefinition represents a route definition for documentation
type RouteDefinition struct {
	Method      string `json:"method"`
	Path         string `json:"path"`
	Description  string `json:"description"`
	AuthRequired bool   `json:"auth_required"`
	RateLimit   string `json:"rate_limit"`
}