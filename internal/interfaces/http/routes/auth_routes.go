package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// AuthRoutes defines authentication routes
type AuthRoutes struct {
	handler         *handlers.AuthHandler
	securityConfig  middleware.SecurityConfig
	rateLimitConfig middleware.RateLimiterConfig
	csrfConfig     middleware.CSRFConfig
}

// NewAuthRoutes creates a new AuthRoutes instance
func NewAuthRoutes(
	handler *handlers.AuthHandler,
	securityConfig middleware.SecurityConfig,
	rateLimitConfig middleware.RateLimiterConfig,
	csrfConfig middleware.CSRFConfig,
) *AuthRoutes {
	return &AuthRoutes{
		handler:         handler,
		securityConfig:  securityConfig,
		rateLimitConfig: rateLimitConfig,
		csrfConfig:     csrfConfig,
	}
}

// RegisterRoutes registers authentication routes
func (r *AuthRoutes) RegisterRoutes(router *gin.RouterGroup, redisClient interface{}) {
	// Create security middleware components
	csrfMiddleware := middleware.NewCSRFMiddleware(r.csrfConfig)
	rateLimiter := middleware.NewEnhancedRateLimiter(
		redisClient,
		r.rateLimitConfig,
		"auth_rate_limit",
	)
	securityMiddleware := middleware.NewSecurityMiddleware(r.securityConfig, csrfMiddleware, rateLimiter)

	auth := router.Group("/auth")
	
	// Apply security middleware to auth routes
	securityMiddleware.ApplyToRouter(auth)
	
	{
		// Public routes
		auth.POST("/register", r.handler.Register)
		auth.POST("/login", r.handler.Login)
		auth.POST("/refresh", r.handler.RefreshToken)
		auth.POST("/password-reset", r.handler.PasswordReset)
		auth.POST("/password-reset/confirm", r.handler.ConfirmPasswordReset)
		auth.POST("/verify", r.handler.VerifyEmail)
		
		// CSRF token endpoint (public)
		auth.GET("/csrf-token", middleware.GetCSRFTokenHandler(csrfMiddleware))
		
		// Protected routes (require authentication)
		protected := auth.Group("")
		// TODO: Add actual auth middleware
		// protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/logout", r.handler.Logout)
			protected.GET("/profile", r.handler.GetProfile)
			protected.POST("/verify/send", r.handler.SendEmailVerification)
			protected.GET("/sessions", r.handler.GetSessions)
		}
	}
}

// GetRoutes returns all authentication routes for documentation
func (r *AuthRoutes) GetRoutes() []RouteInfo {
	return []RouteInfo{
		{
			Method: "POST",
			Path:   "/api/v1/auth/register",
			Description: "Register a new user",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/login",
			Description: "Authenticate user",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/refresh",
			Description: "Refresh access token",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/logout",
			Description: "Logout user",
		},
		{
			Method: "GET",
			Path:   "/api/v1/auth/profile",
			Description: "Get user profile",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/password-reset",
			Description: "Request password reset",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/password-reset/confirm",
			Description: "Confirm password reset",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/verify",
			Description: "Verify email address",
		},
		{
			Method: "POST",
			Path:   "/api/v1/auth/verify/send",
			Description: "Send email verification",
		},
		{
			Method: "GET",
			Path:   "/api/v1/auth/sessions",
			Description: "Get user sessions",
		},
	}
}

// RouteInfo represents route information for documentation
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}