package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AuthConfig represents authentication configuration
type AuthConfig struct {
	// JWT utilities instance
	JWTUtils *utils.JWTUtils
	
	// SkipPaths is a list of paths that don't require authentication
	SkipPaths []string `json:"skip_paths"`
	
	// SkipPathPrefixes is a list of path prefixes that don't require authentication
	SkipPathPrefixes []string `json:"skip_path_prefixes"`
	
	// OptionalPaths is a list of paths where authentication is optional
	OptionalPaths []string `json:"optional_paths"`
	
	// OptionalPathPrefixes is a list of path prefixes where authentication is optional
	OptionalPathPrefixes []string `json:"optional_path_prefixes"`
	
	// AdminPaths is a list of paths that require admin privileges
	AdminPaths []string `json:"admin_paths"`
	
	// AdminPathPrefixes is a list of path prefixes that require admin privileges
	AdminPathPrefixes []string `json:"admin_path_prefixes"`
	
	// TokenHeader is the header name for the authorization token
	TokenHeader string `json:"token_header"`
	
	// TokenScheme is the authorization scheme (e.g., "Bearer")
	TokenScheme string `json:"token_scheme"`
	
	// ContextKey is the key used to store user context
	ContextKey string `json:"context_key"`
	
	// EnableTokenRefresh enables automatic token refresh
	EnableTokenRefresh bool `json:"enable_token_refresh"`
	
	// RefreshTokenHeader is the header name for refresh token
	RefreshTokenHeader string `json:"refresh_token_header"`
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig(jwtUtils *utils.JWTUtils) *AuthConfig {
	return &AuthConfig{
		JWTUtils: jwtUtils,
		SkipPaths: []string{
			"/health",
			"/health/db",
			"/metrics",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/subscriptions/plans",
		},
		SkipPathPrefixes: []string{
			"/swagger",
			"/docs",
		},
		OptionalPaths: []string{},
		OptionalPathPrefixes: []string{
			"/api/v1/public",
		},
		AdminPaths: []string{
			"/api/v1/admin",
		},
		AdminPathPrefixes: []string{
			"/api/v1/admin/",
		},
		TokenHeader:        "Authorization",
		TokenScheme:        "Bearer",
		ContextKey:         "user",
		EnableTokenRefresh:  true,
		RefreshTokenHeader: "X-Refresh-Token",
	}
}

// Auth returns an authentication middleware with the given configuration
func Auth(config *AuthConfig) gin.HandlerFunc {
	if config == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip authentication for specified paths
		if shouldSkipAuth(path, config) {
			c.Next()
			return
		}

		// Check if authentication is optional
		isOptional := isOptionalAuth(path, config)

		// Extract token from header
		token, err := extractToken(c, config)
		if err != nil {
			if isOptional {
				c.Next()
				return
			}
			utils.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Validate token
		claims, err := config.JWTUtils.ValidateAccessToken(token)
		if err != nil {
			// Try to refresh token if enabled and refresh token is provided
			if config.EnableTokenRefresh {
				if newToken, refreshErr := tryRefreshToken(c, config); refreshErr == nil {
					// Set new token in response header
					c.Header("X-New-Access-Token", newToken)
					
					// Validate new token
					claims, err = config.JWTUtils.ValidateAccessToken(newToken)
					if err != nil {
						if isOptional {
							c.Next()
							return
						}
						utils.Unauthorized(c, "Invalid token")
						c.Abort()
						return
					}
				} else {
					if isOptional {
						c.Next()
						return
					}
					utils.Unauthorized(c, "Token expired and refresh failed")
					c.Abort()
					return
				}
			} else {
				if isOptional {
					c.Next()
					return
				}
				utils.Unauthorized(c, "Invalid or expired token")
				c.Abort()
				return
			}
		}

		// Check admin privileges if required
		if requiresAdmin(path, config) && !claims.IsAdmin {
			utils.Forbidden(c, "Admin privileges required")
			c.Abort()
			return
		}

		// Set user context
		setUserContext(c, claims, config)

		c.Next()
	}
}

// shouldSkipAuth checks if authentication should be skipped for the path
func shouldSkipAuth(path string, config *AuthConfig) bool {
	// Check exact paths
	for _, skipPath := range config.SkipPaths {
		if path == skipPath {
			return true
		}
	}

	// Check path prefixes
	for _, skipPrefix := range config.SkipPathPrefixes {
		if strings.HasPrefix(path, skipPrefix) {
			return true
		}
	}

	return false
}

// isOptionalAuth checks if authentication is optional for the path
func isOptionalAuth(path string, config *AuthConfig) bool {
	// Check exact paths
	for _, optionalPath := range config.OptionalPaths {
		if path == optionalPath {
			return true
		}
	}

	// Check path prefixes
	for _, optionalPrefix := range config.OptionalPathPrefixes {
		if strings.HasPrefix(path, optionalPrefix) {
			return true
		}
	}

	return false
}

// requiresAdmin checks if admin privileges are required for the path
func requiresAdmin(path string, config *AuthConfig) bool {
	// Check exact paths
	for _, adminPath := range config.AdminPaths {
		if path == adminPath {
			return true
		}
	}

	// Check path prefixes
	for _, adminPrefix := range config.AdminPathPrefixes {
		if strings.HasPrefix(path, adminPrefix) {
			return true
		}
	}

	return false
}

// extractToken extracts the JWT token from the request
func extractToken(c *gin.Context, config *AuthConfig) (string, error) {
	authHeader := c.GetHeader(config.TokenHeader)
	if authHeader == "" {
		return "", errors.ErrTokenMissing
	}

	// Check scheme
	scheme := config.TokenScheme + " "
	if !strings.HasPrefix(authHeader, scheme) {
		return "", errors.ErrInvalidToken
	}

	token := strings.TrimPrefix(authHeader, scheme)
	if token == "" {
		return "", errors.ErrInvalidToken
	}

	return token, nil
}

// tryRefreshToken attempts to refresh the access token using the refresh token
func tryRefreshToken(c *gin.Context, config *AuthConfig) (string, error) {
	refreshToken := c.GetHeader(config.RefreshTokenHeader)
	if refreshToken == "" {
		return "", errors.ErrTokenMissing
	}

	// Validate refresh token
	_, err := config.JWTUtils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	newToken, err := config.JWTUtils.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// setUserContext sets the user context in the Gin context
func setUserContext(c *gin.Context, claims *utils.Claims, config *AuthConfig) {
	// Set user claims
	c.Set(config.ContextKey, claims)
	
	// Set individual fields for easy access
	c.Set("user_id", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("is_admin", claims.IsAdmin)
	c.Set("token_type", claims.TokenType)
}

// RequireAuth returns a middleware that requires authentication
func RequireAuth(jwtUtils *utils.JWTUtils) gin.HandlerFunc {
	config := DefaultAuthConfig(jwtUtils)
	return Auth(config)
}

// OptionalAuth returns a middleware where authentication is optional
func OptionalAuth(jwtUtils *utils.JWTUtils) gin.HandlerFunc {
	config := DefaultAuthConfig(jwtUtils)
	config.OptionalPaths = []string{"*"} // All paths are optional
	return Auth(config)
}

// RequireAdmin returns a middleware that requires admin privileges
func RequireAdmin(jwtUtils *utils.JWTUtils) gin.HandlerFunc {
	config := DefaultAuthConfig(jwtUtils)
	config.AdminPaths = []string{"*"} // All paths require admin
	return Auth(config)
}

// CustomAuth returns a middleware with custom authentication logic
func CustomAuth(jwtUtils *utils.JWTUtils, skipPaths, adminPaths []string) gin.HandlerFunc {
	config := DefaultAuthConfig(jwtUtils)
	config.SkipPaths = skipPaths
	config.AdminPaths = adminPaths
	return Auth(config)
}

// GetUserFromContext retrieves user claims from the Gin context
func GetUserFromContext(c *gin.Context) (*utils.Claims, bool) {
	if user, exists := c.Get("user"); exists {
		if claims, ok := user.(*utils.Claims); ok {
			return claims, true
		}
	}
	return nil, false
}

// GetUserIDFromContext retrieves user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// GetEmailFromContext retrieves user email from the Gin context
func GetEmailFromContext(c *gin.Context) (string, bool) {
	if email, exists := c.Get("email"); exists {
		if e, ok := email.(string); ok {
			return e, true
		}
	}
	return "", false
}

// IsAdminFromContext checks if user is admin from the Gin context
func IsAdminFromContext(c *gin.Context) bool {
	if isAdmin, exists := c.Get("is_admin"); exists {
		if admin, ok := isAdmin.(bool); ok {
			return admin
		}
	}
	return false
}

// APIKeyAuth provides API key authentication for service-to-service communication
func APIKeyAuth(validKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			utils.Unauthorized(c, "API key is required")
			c.Abort()
			return
		}

		// Validate API key
		if serviceName, exists := validKeys[apiKey]; exists {
			c.Set("service_name", serviceName)
			c.Set("api_key_valid", true)
			c.Next()
			return
		}

		utils.Unauthorized(c, "Invalid API key")
		c.Abort()
	}
}

// BasicAuth provides basic HTTP authentication
func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// AuthMiddleware returns a middleware that requires authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (should be set by previous auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok && id != "" {
				c.Next()
				return
			}
		}
		
		utils.Unauthorized(c, "Authentication required")
		c.Abort()
	}
}

// AdminAuthMiddleware returns a middleware that requires admin privileges
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok && id != "" {
				// Check if user is admin
				if isAdmin, exists := c.Get("is_admin"); exists {
					if admin, ok := isAdmin.(bool); ok && admin {
						// Set admin_id in context for admin handlers
						c.Set("admin_id", id)
						c.Next()
						return
					}
				}
			}
		}
		
		utils.Forbidden(c, "Admin privileges required")
		c.Abort()
	}
}

// RateLimitMiddleware returns a middleware that applies rate limiting
func RateLimitMiddleware(rateLimiter interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would typically integrate with the rate limiter
		// For now, we'll just pass through
		// In a real implementation, you would check the rate limit here
		c.Next()
	}
}