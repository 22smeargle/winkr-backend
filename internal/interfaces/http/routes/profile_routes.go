package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// ProfileRoutes defines profile routes
type ProfileRoutes struct {
	handler         *handlers.ProfileHandler
	securityConfig  middleware.SecurityConfig
	rateLimitConfig middleware.RateLimiterConfig
	csrfConfig     middleware.CSRFConfig
}

// NewProfileRoutes creates a new ProfileRoutes instance
func NewProfileRoutes(
	handler *handlers.ProfileHandler,
	securityConfig middleware.SecurityConfig,
	rateLimitConfig middleware.RateLimiterConfig,
	csrfConfig middleware.CSRFConfig,
) *ProfileRoutes {
	return &ProfileRoutes{
		handler:         handler,
		securityConfig:  securityConfig,
		rateLimitConfig: rateLimitConfig,
		csrfConfig:     csrfConfig,
	}
}

// RegisterRoutes registers profile routes
func (r *ProfileRoutes) RegisterRoutes(router *gin.RouterGroup, redisClient interface{}) {
	// Create security middleware components
	csrfMiddleware := middleware.NewCSRFMiddleware(r.csrfConfig)
	rateLimiter := middleware.NewEnhancedRateLimiter(
		redisClient,
		r.rateLimitConfig,
		"profile_rate_limit",
	)
	securityMiddleware := middleware.NewSecurityMiddleware(r.securityConfig, csrfMiddleware, rateLimiter)

	profile := router.Group("/profile")
	
	// Apply security middleware to profile routes
	securityMiddleware.ApplyToRouter(profile)
	
	// Apply authentication middleware to all profile routes
	profile.Use(middleware.AuthMiddleware())
	
	{
		// Current user profile routes
		profile.GET("/me", r.handler.GetProfile)
		profile.PUT("/me", r.handler.UpdateProfile)
		profile.PUT("/me/location", r.handler.UpdateLocation)
		profile.GET("/me/matches", r.handler.GetMatches)
		profile.DELETE("/me/account", r.handler.DeleteAccount)
		
		// Other user profile routes
		profile.GET("/users/:id", r.handler.ViewUserProfile)
	}
}

// GetRoutes returns all profile routes for documentation
func (r *ProfileRoutes) GetRoutes() []RouteInfo {
	return []RouteInfo{
		{
			Method: "GET",
			Path:   "/api/v1/profile/me",
			Description: "Get current user profile",
		},
		{
			Method: "PUT",
			Path:   "/api/v1/profile/me",
			Description: "Update current user profile",
		},
		{
			Method: "PUT",
			Path:   "/api/v1/profile/me/location",
			Description: "Update user location",
		},
		{
			Method: "GET",
			Path:   "/api/v1/profile/me/matches",
			Description: "Get user matches",
		},
		{
			Method: "DELETE",
			Path:   "/api/v1/profile/me/account",
			Description: "Delete user account",
		},
		{
			Method: "GET",
			Path:   "/api/v1/profile/users/{id}",
			Description: "View user profile",
		},
	}
}

// RouteInfo represents route information for documentation
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}