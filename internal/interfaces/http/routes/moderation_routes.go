package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// ModerationRoutes defines moderation routes
type ModerationRoutes struct {
	moderationHandler     *handlers.ModerationHandler
	adminModerationHandler *handlers.AdminModerationHandler
	securityConfig        middleware.SecurityConfig
	rateLimitConfig       middleware.RateLimiterConfig
	csrfConfig           middleware.CSRFConfig
}

// NewModerationRoutes creates a new ModerationRoutes instance
func NewModerationRoutes(
	moderationHandler *handlers.ModerationHandler,
	adminModerationHandler *handlers.AdminModerationHandler,
	securityConfig middleware.SecurityConfig,
	rateLimitConfig middleware.RateLimiterConfig,
	csrfConfig middleware.CSRFConfig,
) *ModerationRoutes {
	return &ModerationRoutes{
		moderationHandler:      moderationHandler,
		adminModerationHandler: adminModerationHandler,
		securityConfig:         securityConfig,
		rateLimitConfig:        rateLimitConfig,
		csrfConfig:            csrfConfig,
	}
}

// RegisterRoutes registers moderation routes
func (r *ModerationRoutes) RegisterRoutes(router *gin.RouterGroup, redisClient interface{}) {
	// Create security middleware components
	csrfMiddleware := middleware.NewCSRFMiddleware(r.csrfConfig)
	
	// Rate limiters for different endpoints
	reportRateLimiter := middleware.NewEnhancedRateLimiter(
		redisClient,
		middleware.RateLimiterConfig{
			RequestsPerMinute: 5,  // Limit to 5 reports per minute
			BurstSize:         10,
			KeyGenerator:      func(c *gin.Context) string { return "moderation:report:" + c.ClientIP() },
		},
		"moderation_report_rate_limit",
	)
	
	blockRateLimiter := middleware.NewEnhancedRateLimiter(
		redisClient,
		middleware.RateLimiterConfig{
			RequestsPerMinute: 10, // Limit to 10 blocks per minute
			BurstSize:         20,
			KeyGenerator:      func(c *gin.Context) string { return "moderation:block:" + c.ClientIP() },
		},
		"moderation_block_rate_limit",
	)
	
	generalRateLimiter := middleware.NewEnhancedRateLimiter(
		redisClient,
		r.rateLimitConfig,
		"moderation_general_rate_limit",
	)
	
	securityMiddleware := middleware.NewSecurityMiddleware(r.securityConfig, csrfMiddleware, generalRateLimiter)

	// User moderation routes
	moderation := router.Group("/moderation")
	
	// Apply security middleware to moderation routes
	securityMiddleware.ApplyToRouter(moderation)
	
	// Apply authentication middleware to all moderation routes
	moderation.Use(middleware.AuthMiddleware())
	
	{
		// Report content with rate limiting
		moderation.POST("/report", 
			middleware.RateLimitMiddleware(reportRateLimiter),
			r.moderationHandler.ReportContent,
		)
		
		// Block user with rate limiting
		moderation.POST("/block/:id", 
			middleware.RateLimitMiddleware(blockRateLimiter),
			r.moderationHandler.BlockUser,
		)
		
		// Unblock user with rate limiting
		moderation.DELETE("/block/:id", 
			middleware.RateLimitMiddleware(blockRateLimiter),
			r.moderationHandler.UnblockUser,
		)
		
		// Get blocked users
		moderation.GET("/me/blocked", r.moderationHandler.GetBlockedUsers)
		
		// Get user's reports
		moderation.GET("/me/reports", r.moderationHandler.GetMyReports)
		
		// Get report status
		moderation.GET("/reports/:id", r.moderationHandler.GetReportStatus)
		
		// Cancel report
		moderation.DELETE("/reports/:id", r.moderationHandler.CancelReport)
		
		// Check block status
		moderation.GET("/block/:id/status", r.moderationHandler.CheckBlockStatus)
		
		// Check mutual block
		moderation.GET("/block/:id/mutual", r.moderationHandler.CheckMutualBlock)
	}

	// Admin moderation routes
	admin := router.Group("/admin")
	
	// Apply security middleware to admin routes
	securityMiddleware.ApplyToRouter(admin)
	
	// Apply authentication middleware to all admin routes
	admin.Use(middleware.AuthMiddleware())
	
	// Apply admin authorization middleware
	admin.Use(middleware.AdminAuthMiddleware())
	
	{
		// Report management
		admin.GET("/reports", r.adminModerationHandler.GetReports)
		admin.GET("/reports/:id", r.adminModerationHandler.GetReportDetails)
		admin.POST("/reports/:id/review", r.adminModerationHandler.ReviewReport)
		
		// User management
		admin.POST("/users/:id/ban", r.adminModerationHandler.BanUser)
		admin.POST("/users/:id/suspend", r.adminModerationHandler.SuspendUser)
		admin.GET("/users/:id/bans", r.adminModerationHandler.GetBanHistory)
		admin.GET("/users/:id/appeals", r.adminModerationHandler.GetAppealHistory)
		
		// Appeal management
		admin.POST("/appeals/:id/review", r.adminModerationHandler.ReviewAppeal)
		
		// Moderation queue
		admin.GET("/moderation-queue", r.adminModerationHandler.GetModerationQueue)
		
		// Analytics
		admin.GET("/analytics", r.adminModerationHandler.GetModerationAnalytics)
	}
}

// GetRoutes returns all moderation routes for documentation
func (r *ModerationRoutes) GetRoutes() []RouteInfo {
	return []RouteInfo{
		// User moderation routes
		{
			Method:      "POST",
			Path:        "/api/v1/moderation/report",
			Description: "Report user or content",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/moderation/block/:id",
			Description: "Block a user",
		},
		{
			Method:      "DELETE",
			Path:        "/api/v1/moderation/block/:id",
			Description: "Unblock a user",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/moderation/me/blocked",
			Description: "List blocked users",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/moderation/me/reports",
			Description: "List user's reports",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/moderation/reports/:id",
			Description: "Get report details",
		},
		{
			Method:      "DELETE",
			Path:        "/api/v1/moderation/reports/:id",
			Description: "Cancel report",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/moderation/block/:id/status",
			Description: "Check block status",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/moderation/block/:id/mutual",
			Description: "Check mutual block",
		},
		// Admin moderation routes
		{
			Method:      "GET",
			Path:        "/api/v1/admin/reports",
			Description: "List all reports",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/reports/:id",
			Description: "Get report details",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/admin/reports/:id/review",
			Description: "Review and resolve report",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/admin/users/:id/ban",
			Description: "Ban user",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/admin/users/:id/suspend",
			Description: "Suspend user",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/users/:id/bans",
			Description: "Get user ban history",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/users/:id/appeals",
			Description: "Get user appeal history",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/admin/appeals/:id/review",
			Description: "Review appeal",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/moderation-queue",
			Description: "Get moderation queue",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/analytics",
			Description: "Get moderation analytics",
		},
	}
}