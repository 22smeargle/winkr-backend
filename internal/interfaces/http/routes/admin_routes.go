package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/auth"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// AdminRoutes configures admin routes
type AdminRoutes struct {
	adminUserHandler       *handlers.AdminUserHandler
	adminAnalyticsHandler  *handlers.AdminAnalyticsHandler
	adminSystemHandler     *handlers.AdminSystemHandler
	adminContentHandler    *handlers.AdminContentHandler
	adminAuthMiddleware    *middleware.AdminAuthMiddleware
	rateLimiterMiddleware  *middleware.RateLimiterMiddleware
	loggingMiddleware      *middleware.LoggingMiddleware
}

// NewAdminRoutes creates a new AdminRoutes
func NewAdminRoutes(
	adminService *services.AdminService,
	tokenManager auth.TokenManager,
	rateLimiterMiddleware *middleware.RateLimiterMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
) *AdminRoutes {
	adminAuthMiddleware := middleware.NewAdminAuthMiddleware(tokenManager)

	return &AdminRoutes{
		adminUserHandler:      handlers.NewAdminUserHandler(adminService),
		adminAnalyticsHandler:  handlers.NewAdminAnalyticsHandler(adminService),
		adminSystemHandler:     handlers.NewAdminSystemHandler(adminService),
		adminContentHandler:    handlers.NewAdminContentHandler(adminService),
		adminAuthMiddleware:    adminAuthMiddleware,
		rateLimiterMiddleware:  rateLimiterMiddleware,
		loggingMiddleware:      loggingMiddleware,
	}
}

// SetupAdminRoutes sets up admin routes
func (r *AdminRoutes) SetupAdminRoutes(router *gin.Engine) {
	// Admin group with authentication
	adminGroup := router.Group("/admin/v1")
	{
		// Apply authentication middleware to all admin routes
		adminGroup.Use(r.adminAuthMiddleware.RequireAdminAuth())
		adminGroup.Use(r.adminAuthMiddleware.AdminActivityLogger())
		adminGroup.Use(r.loggingMiddleware.LogRequest())

		// Apply rate limiting to admin routes
		adminGroup.Use(r.rateLimiterMiddleware.RateLimit(100, time.Minute)) // 100 requests per minute

		// User Management Routes
		usersGroup := adminGroup.Group("/users")
		{
			usersGroup.GET("", 
				r.adminAuthMiddleware.RequirePermission("users.read"),
				r.adminUserHandler.GetUsers,
			)
			usersGroup.GET("/:id", 
				r.adminAuthMiddleware.RequirePermission("users.read"),
				r.adminUserHandler.GetUserDetails,
			)
			usersGroup.PUT("/:id", 
				r.adminAuthMiddleware.RequirePermission("users.write"),
				r.adminUserHandler.UpdateUser,
			)
			usersGroup.DELETE("/:id", 
				r.adminAuthMiddleware.RequirePermission("users.delete"),
				r.adminUserHandler.DeleteUser,
			)
			usersGroup.POST("/:id/suspend", 
				r.adminAuthMiddleware.RequirePermission("users.write"),
				r.adminUserHandler.SuspendUser,
			)
			usersGroup.POST("/:id/unsuspend", 
				r.adminAuthMiddleware.RequirePermission("users.write"),
				r.adminUserHandler.UnsuspendUser,
			)
			usersGroup.POST("/:id/ban", 
				r.adminAuthMiddleware.RequirePermission("users.delete"),
				r.adminUserHandler.BanUser,
			)
			usersGroup.POST("/:id/unban", 
				r.adminAuthMiddleware.RequirePermission("users.delete"),
				r.adminUserHandler.UnbanUser,
			)
			usersGroup.GET("/:id/activity", 
				r.adminAuthMiddleware.RequirePermission("users.read"),
				r.adminUserHandler.GetUserActivity,
			)

			// Bulk operations
			usersGroup.POST("/bulk/update", 
				r.adminAuthMiddleware.RequirePermission("users.write"),
				r.adminUserHandler.BulkUpdateUsers,
			)
			usersGroup.POST("/bulk/delete", 
				r.adminAuthMiddleware.RequirePermission("users.delete"),
				r.adminUserHandler.BulkDeleteUsers,
			)
		}

		// Analytics Routes
		analyticsGroup := adminGroup.Group("/stats")
		{
			analyticsGroup.GET("", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetPlatformStats,
			)
			analyticsGroup.GET("/users", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetUserStats,
			)
			analyticsGroup.GET("/matches", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetMatchStats,
			)
			analyticsGroup.GET("/messages", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetMessageStats,
			)
			analyticsGroup.GET("/payments", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetPaymentStats,
			)
			analyticsGroup.GET("/verification", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetVerificationStats,
			)

			// Dashboard data
			analyticsGroup.GET("/dashboard", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetDashboardData,
			)
			analyticsGroup.GET("/export", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.ExportStats,
			)
		}

		// System Management Routes
		systemGroup := adminGroup.Group("/system")
		{
			systemGroup.GET("/health", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetSystemHealth,
			)
			systemGroup.GET("/metrics", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetSystemMetrics,
			)
			systemGroup.GET("/logs", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetSystemLogs,
			)
			systemGroup.POST("/maintenance", 
				r.adminAuthMiddleware.RequireRole("super_admin"),
				r.adminSystemHandler.ToggleMaintenanceMode,
			)
			systemGroup.GET("/config", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetSystemConfig,
			)
			systemGroup.PUT("/config", 
				r.adminAuthMiddleware.RequirePermission("system.write"),
				r.adminSystemHandler.UpdateSystemConfig,
			)
			systemGroup.POST("/restart", 
				r.adminAuthMiddleware.RequireRole("super_admin"),
				r.adminSystemHandler.RestartService,
			)
			systemGroup.GET("/cache/stats", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetCacheStats,
			)
			systemGroup.POST("/cache/clear", 
				r.adminAuthMiddleware.RequirePermission("system.write"),
				r.adminSystemHandler.ClearCache,
			)
		}

		// Content Management Routes
		contentGroup := adminGroup.Group("/content")
		{
			// Photo moderation
			photosGroup := contentGroup.Group("/photos")
			{
				photosGroup.GET("", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.GetReportedPhotos,
				)
				photosGroup.POST("/:id/approve", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.ApprovePhoto,
				)
				photosGroup.POST("/:id/reject", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.RejectPhoto,
				)
				photosGroup.POST("/bulk/moderate", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.BulkModeratePhotos,
				)
			}

			// Message moderation
			messagesGroup := contentGroup.Group("/messages")
			{
				messagesGroup.GET("", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.GetReportedMessages,
				)
				messagesGroup.POST("/:id/delete", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.DeleteMessage,
				)
				messagesGroup.POST("/bulk/moderate", 
					r.adminAuthMiddleware.RequirePermission("content.moderate"),
					r.adminContentHandler.BulkModerateMessages,
				)
			}

			// Content analytics
			contentGroup.GET("/analytics", 
				r.adminAuthMiddleware.RequirePermission("content.read"),
				r.adminContentHandler.GetContentAnalytics,
			)
			contentGroup.GET("/reports", 
				r.adminAuthMiddleware.RequirePermission("content.read"),
				r.adminContentHandler.GetReportedContent,
			)
		}

		// Admin Management Routes (Super Admin only)
		adminManagementGroup := adminGroup.Group("/management")
		{
			adminManagementGroup.Use(r.adminAuthMiddleware.RequireRole("super_admin"))
			
			adminManagementGroup.GET("/admins", 
				r.adminUserHandler.GetAdmins,
			)
			adminManagementGroup.POST("/admins", 
				r.adminUserHandler.CreateAdmin,
			)
			adminManagementGroup.PUT("/admins/:id", 
				r.adminUserHandler.UpdateAdmin,
			)
			adminManagementGroup.DELETE("/admins/:id", 
				r.adminUserHandler.DeleteAdmin,
			)
			adminManagementGroup.GET("/admins/:id/activity", 
				r.adminUserHandler.GetAdminActivity,
			)
		}

		// Alerts and Notifications
		alertsGroup := adminGroup.Group("/alerts")
		{
			alertsGroup.GET("", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetAlerts,
			)
			alertsGroup.POST("/:id/acknowledge", 
				r.adminAuthMiddleware.RequirePermission("system.write"),
				r.adminSystemHandler.AcknowledgeAlert,
			)
			alertsGroup.POST("/:id/resolve", 
				r.adminAuthMiddleware.RequirePermission("system.write"),
				r.adminSystemHandler.ResolveAlert,
			)
		}

		// Reports and Audits
		reportsGroup := adminGroup.Group("/reports")
		{
			reportsGroup.GET("/audit", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetAuditLogs,
			)
			reportsGroup.GET("/activity", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetActivityReport,
			)
			reportsGroup.GET("/performance", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GetPerformanceReport,
			)
			reportsGroup.POST("/generate", 
				r.adminAuthMiddleware.RequirePermission("analytics.read"),
				r.adminAnalyticsHandler.GenerateReport,
			)
		}

		// Settings and Configuration
		settingsGroup := adminGroup.Group("/settings")
		{
			settingsGroup.GET("", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetSettings,
			)
			settingsGroup.PUT("", 
				r.adminAuthMiddleware.RequirePermission("system.write"),
				r.adminSystemHandler.UpdateSettings,
			)
			settingsGroup.GET("/permissions", 
				r.adminAuthMiddleware.RequirePermission("system.read"),
				r.adminSystemHandler.GetPermissions,
			)
			settingsGroup.PUT("/permissions", 
				r.adminAuthMiddleware.RequireRole("super_admin"),
				r.adminSystemHandler.UpdatePermissions,
			)
		}
	}

	// Admin authentication routes (no auth required for login)
	authGroup := router.Group("/admin/v1/auth")
	{
		authGroup.POST("/login", 
			r.rateLimiterMiddleware.RateLimit(5, time.Minute), // 5 login attempts per minute
			r.adminUserHandler.AdminLogin,
		)
		authGroup.POST("/logout", 
			r.adminAuthMiddleware.RequireAdminAuth(),
			r.adminUserHandler.AdminLogout,
		)
		authGroup.POST("/refresh", 
			r.adminAuthMiddleware.OptionalAdminAuth(),
			r.adminUserHandler.RefreshToken,
		)
		authGroup.GET("/me", 
			r.adminAuthMiddleware.RequireAdminAuth(),
			r.adminUserHandler.GetCurrentAdmin,
		)
		authGroup.PUT("/password", 
			r.adminAuthMiddleware.RequireAdminAuth(),
			r.adminUserHandler.ChangePassword,
		)
	}

	// Health check (no auth required)
	router.GET("/admin/v1/health", r.adminSystemHandler.PublicHealthCheck)
}

// SetupAdminAPIRoutes sets up admin API routes with versioning
func (r *AdminRoutes) SetupAdminAPIRoutes(router *gin.Engine) {
	// API versioning
	v1 := router.Group("/api/admin/v1")
	{
		r.SetupAdminRoutes(router)
	}

	// Future versions can be added here
	// v2 := router.Group("/api/admin/v2")
	// {
	//     // Future admin routes
	// }
}