package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// EphemeralPhotoRateLimitConfig defines rate limit configuration for different operations
type EphemeralPhotoRateLimitConfig struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
	Message  string         `json:"message"`
}

// EphemeralPhotoRateLimiter handles rate limiting for ephemeral photo operations
type EphemeralPhotoRateLimiter struct {
	uploadLimit   *RateLimiter
	viewLimit     *RateLimiter
	deleteLimit   *RateLimiter
	statusLimit   *RateLimiter
	expireLimit   *RateLimiter
	analyticsLimit *RateLimiter
}

// NewEphemeralPhotoRateLimiter creates a new ephemeral photo rate limiter
func NewEphemeralPhotoRateLimiter() *EphemeralPhotoRateLimiter {
	return &EphemeralPhotoRateLimiter{
		uploadLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 10, // 10 uploads per minute
			Window:   time.Minute,
			Message:  "Too many upload attempts. Please try again later.",
		}),
		viewLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 100, // 100 views per minute
			Window:   time.Minute,
			Message:  "Too many view attempts. Please try again later.",
		}),
		deleteLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 30, // 30 deletes per minute
			Window:   time.Minute,
			Message:  "Too many delete attempts. Please try again later.",
		}),
		statusLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 100, // 100 status checks per minute
			Window:   time.Minute,
			Message:  "Too many status check attempts. Please try again later.",
		}),
		expireLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 20, // 20 expires per minute
			Window:   time.Minute,
			Message:  "Too many expire attempts. Please try again later.",
		}),
		analyticsLimit: NewRateLimiter(&RateLimitConfig{
			Requests: 50, // 50 analytics requests per minute
			Window:   time.Minute,
			Message:  "Too many analytics requests. Please try again later.",
		}),
	}
}

// EphemeralPhotoRateLimit applies rate limiting based on the operation type
func EphemeralPhotoRateLimit() gin.HandlerFunc {
	limiter := NewEphemeralPhotoRateLimiter()
	
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		
		// Determine which rate limit to apply based on the endpoint
		var rateLimiter *RateLimiter
		var limitType string
		
		switch {
		case method == "POST" && path == "/api/v1/ephemeral-photos":
			rateLimiter = limiter.uploadLimit
			limitType = "upload"
		case method == "GET" && (path == "/api/v1/ephemeral-photos" || path == "/api/v1/ephemeral-photos/analytics"):
			rateLimiter = limiter.analyticsLimit
			limitType = "analytics"
		case method == "GET" && path == "/api/v1/ephemeral-photos/analytics":
			rateLimiter = limiter.analyticsLimit
			limitType = "analytics"
		case method == "DELETE" && path != "" && path != "/":
			rateLimiter = limiter.deleteLimit
			limitType = "delete"
		case method == "POST" && path != "" && path != "/":
			rateLimiter = limiter.expireLimit
			limitType = "expire"
		case method == "GET" && path != "" && path != "/":
			rateLimiter = limiter.statusLimit
			limitType = "status"
		default:
			// For public view endpoint, use view limit
			if method == "GET" && path != "" && path != "/" {
				rateLimiter = limiter.viewLimit
				limitType = "view"
			} else {
				// No rate limiting for other endpoints
				c.Next()
				return
			}
		}
		
		// Get identifier for rate limiting
		identifier := getRateLimitIdentifier(c, limitType)
		
		// Apply rate limit
		if !rateLimiter.Allow(identifier) {
			// Set rate limit headers
			c.Header("X-RateLimit-Limit", rateLimiter.GetConfig().Requests)
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", time.Now().Add(rateLimiter.GetConfig().Window).Unix())
			
			utils.ErrorResponse(c, http.StatusTooManyRequests, rateLimiter.GetConfig().Message)
			c.Abort()
			return
		}
		
		// Set rate limit headers for successful requests
		remaining := rateLimiter.GetRemaining(identifier)
		c.Header("X-RateLimit-Limit", rateLimiter.GetConfig().Requests)
		c.Header("X-RateLimit-Remaining", remaining)
		c.Header("X-RateLimit-Reset", time.Now().Add(rateLimiter.GetConfig().Window).Unix())
		
		c.Next()
	}
}

// getRateLimitIdentifier generates an identifier for rate limiting
func getRateLimitIdentifier(c *gin.Context, limitType string) string {
	// Try to get user ID first
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return "user:" + uid + ":" + limitType
		}
	}
	
	// Fall back to IP address
	ip := c.ClientIP()
	if ip != "" {
		return "ip:" + ip + ":" + limitType
	}
	
	// Final fallback
	return "unknown:" + limitType
}

// GetEphemeralPhotoRateLimits returns all rate limit configurations
func GetEphemeralPhotoRateLimits() map[string]EphemeralPhotoRateLimitConfig {
	limiter := NewEphemeralPhotoRateLimiter()
	
	return map[string]EphemeralPhotoRateLimitConfig{
		"upload": {
			Requests: limiter.uploadLimit.GetConfig().Requests,
			Window:   limiter.uploadLimit.GetConfig().Window,
			Message:  limiter.uploadLimit.GetConfig().Message,
		},
		"view": {
			Requests: limiter.viewLimit.GetConfig().Requests,
			Window:   limiter.viewLimit.GetConfig().Window,
			Message:  limiter.viewLimit.GetConfig().Message,
		},
		"delete": {
			Requests: limiter.deleteLimit.GetConfig().Requests,
			Window:   limiter.deleteLimit.GetConfig().Window,
			Message:  limiter.deleteLimit.GetConfig().Message,
		},
		"status": {
			Requests: limiter.statusLimit.GetConfig().Requests,
			Window:   limiter.statusLimit.GetConfig().Window,
			Message:  limiter.statusLimit.GetConfig().Message,
		},
		"expire": {
			Requests: limiter.expireLimit.GetConfig().Requests,
			Window:   limiter.expireLimit.GetConfig().Window,
			Message:  limiter.expireLimit.GetConfig().Message,
		},
		"analytics": {
			Requests: limiter.analyticsLimit.GetConfig().Requests,
			Window:   limiter.analyticsLimit.GetConfig().Window,
			Message:  limiter.analyticsLimit.GetConfig().Message,
		},
	}
}

// ApplyEphemeralPhotoSecurityHeaders applies security headers for ephemeral photo responses
func ApplyEphemeralPhotoSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent content type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent XSS
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none'")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// Cache control for sensitive content
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		
		c.Next()
	}
}

// ApplyEphemeralPhotoCORSHeaders applies CORS headers for ephemeral photo endpoints
func ApplyEphemeralPhotoCORSHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowedOrigins := []string{
			"https://yourdomain.com",
			"https://www.yourdomain.com",
			"http://localhost:3000", // For development
			"http://localhost:8080", // For development
		}
		
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		// Allowed methods
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		
		// Allowed headers
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Requested-With")
		
		// Exposed headers
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Total-Count, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
		
		// Max age
		c.Header("Access-Control-Max-Age", "86400") // 24 hours
		
		// Credentials
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}