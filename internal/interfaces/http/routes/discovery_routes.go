package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/matching"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// DiscoveryRoutes defines discovery and matching routes
type DiscoveryRoutes struct {
	handler *handlers.DiscoveryHandler
}

// NewDiscoveryRoutes creates new discovery routes
func NewDiscoveryRoutes(
	discoverUsersUseCase *matching.DiscoverUsersUseCase,
	likeUserUseCase *matching.LikeUserUseCase,
	dislikeUserUseCase *matching.DislikeUserUseCase,
	superLikeUserUseCase *matching.SuperLikeUserUseCase,
	getMatchesUseCase *matching.GetMatchesUseCase,
	getDiscoveryStatsUseCase *matching.GetDiscoveryStatsUseCase,
) *DiscoveryRoutes {
	handler := handlers.NewDiscoveryHandler(
		discoverUsersUseCase,
		likeUserUseCase,
		dislikeUserUseCase,
		superLikeUserUseCase,
		getMatchesUseCase,
		getDiscoveryStatsUseCase,
	)

	return &DiscoveryRoutes{
		handler: handler,
	}
}

// RegisterRoutes registers discovery routes with the router
func (r *DiscoveryRoutes) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, rateLimitMiddleware gin.HandlerFunc) {
	// Discovery group with authentication and rate limiting
	discoveryGroup := router.Group("/api/v1")
	discoveryGroup.Use(authMiddleware)                    // Require authentication
	discoveryGroup.Use(rateLimitMiddleware)                // Apply rate limiting

	// Discovery endpoints
	discoveryGroup.GET("/discover", r.handler.DiscoverUsers)
	discoveryGroup.POST("/like/:id", r.handler.LikeUser)
	discoveryGroup.POST("/dislike/:id", r.handler.DislikeUser)
	discoveryGroup.POST("/superlike/:id", r.handler.SuperLikeUser)
	discoveryGroup.GET("/matches", r.handler.GetMatches)
	discoveryGroup.GET("/discover/stats", r.handler.GetDiscoveryStats)
}

// GetRateLimitMiddleware returns rate limiting middleware for discovery endpoints
func GetRateLimitMiddleware() gin.HandlerFunc {
	// Return a middleware that applies discovery-specific rate limits
	// This would typically use the RateLimiter service
	return func(c *gin.Context) {
		// Apply discovery rate limits
		// This is a placeholder - actual implementation would use the rate limiter service
		c.Next()
	}
}

// GetAuthMiddleware returns authentication middleware for discovery endpoints
func GetAuthMiddleware() gin.HandlerFunc {
	// Return authentication middleware
	// This would typically validate JWT tokens and set user context
	return func(c *gin.Context) {
		// Extract user ID from JWT token
		// Set user_id in context for handlers to use
		c.Next()
	}
}