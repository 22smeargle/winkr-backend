package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/auth"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AdminAuthMiddleware handles admin authentication and authorization
type AdminAuthMiddleware struct {
	tokenManager auth.TokenManager
	// In a real implementation, this would include admin repository or service
}

// NewAdminAuthMiddleware creates a new AdminAuthMiddleware
func NewAdminAuthMiddleware(tokenManager auth.TokenManager) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		tokenManager: tokenManager,
	}
}

// AdminAuth represents authenticated admin user
type AdminAuth struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Permissions []string `json:"permissions"`
}

// RequireAdminAuth middleware that requires admin authentication
func (m *AdminAuthMiddleware) RequireAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "missing_auth_header",
			})
			c.Abort()
			return
		}

		// Extract token from Bearer format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "invalid_auth_format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token
		claims, err := m.tokenManager.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid token", "error", err, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "invalid_token",
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if !m.isAdmin(claims.UserID) {
			logger.Warn("Non-admin user attempted to access admin endpoint", "user_id", claims.UserID, "ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "admin_access_required",
			})
			c.Abort()
			return
		}

		// Get admin details
		admin, err := m.getAdminDetails(claims.UserID)
		if err != nil {
			logger.Error("Failed to get admin details", err, "user_id", claims.UserID, "ip", c.ClientIP())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to authenticate admin",
				"code":  "admin_auth_failed",
			})
			c.Abort()
			return
		}

		// Check if admin is active
		if !m.isAdminActive(admin.ID) {
			logger.Warn("Inactive admin attempted to access admin endpoint", "admin_id", admin.ID, "ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin account is inactive",
				"code":  "admin_inactive",
			})
			c.Abort()
			return
		}

		// Store admin in context
		c.Set("admin", admin)
		c.Set("admin_id", admin.ID)

		// Log admin activity
		m.logAdminActivity(admin.ID, c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()
	}
}

// RequirePermission middleware that requires specific permission
func (m *AdminAuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get admin from context
		adminInterface, exists := c.Get("admin")
		if !exists {
			logger.Warn("Admin not found in context", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "auth_required",
			})
			c.Abort()
			return
		}

		admin := adminInterface.(*AdminAuth)

		// Check if admin has required permission
		if !m.hasPermission(admin, permission) {
			logger.Warn("Admin lacks required permission", "admin_id", admin.ID, "permission", permission, "ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "insufficient_permissions",
				"required_permission": permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AdminAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get admin from context
		adminInterface, exists := c.Get("admin")
		if !exists {
			logger.Warn("Admin not found in context", "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "auth_required",
			})
			c.Abort()
			return
		}

		admin := adminInterface.(*AdminAuth)

		// Check if admin has required role
		if admin.Role != role && admin.Role != "super_admin" {
			logger.Warn("Admin lacks required role", "admin_id", admin.ID, "required_role", role, "admin_role", admin.Role, "ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient role privileges",
				"code":  "insufficient_role",
				"required_role": role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAdminAuth middleware that optionally authenticates admin
func (m *AdminAuthMiddleware) OptionalAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Extract token from Bearer format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := tokenParts[1]

		// Validate token
		claims, err := m.tokenManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Check if user is admin
		if !m.isAdmin(claims.UserID) {
			// Not admin, continue without authentication
			c.Next()
			return
		}

		// Get admin details
		admin, err := m.getAdminDetails(claims.UserID)
		if err != nil {
			// Failed to get admin details, continue without authentication
			c.Next()
			return
		}

		// Check if admin is active
		if !m.isAdminActive(admin.ID) {
			// Inactive admin, continue without authentication
			c.Next()
			return
		}

		// Store admin in context
		c.Set("admin", admin)
		c.Set("admin_id", admin.ID)

		// Log admin activity
		m.logAdminActivity(admin.ID, c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()
	}
}

// isAdmin checks if user is admin
func (m *AdminAuthMiddleware) isAdmin(userID uuid.UUID) bool {
	// Mock implementation - in real implementation, this would check the database
	// For demo purposes, we'll consider some specific UUIDs as admins
	adminIDs := []uuid.UUID{
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
	}

	for _, adminID := range adminIDs {
		if adminID == userID {
			return true
		}
	}

	return false
}

// getAdminDetails retrieves admin details
func (m *AdminAuthMiddleware) getAdminDetails(userID uuid.UUID) (*AdminAuth, error) {
	// Mock implementation - in real implementation, this would query the database
	admin := &AdminAuth{
		ID:       userID,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Permissions: []string{
			"users.read",
			"users.write",
			"users.delete",
			"analytics.read",
			"system.read",
			"system.write",
			"content.moderate",
		},
	}

	// Set super admin role for specific admin
	if userID == uuid.MustParse("123e4567-e89b-12d3-a456-426614174000") {
		admin.Role = "super_admin"
		admin.Permissions = append(admin.Permissions, "system.admin", "users.admin")
	}

	return admin, nil
}

// isAdminActive checks if admin is active
func (m *AdminAuthMiddleware) isAdminActive(adminID uuid.UUID) bool {
	// Mock implementation - in real implementation, this would check the database
	// For demo purposes, we'll consider all admins as active
	return true
}

// hasPermission checks if admin has specific permission
func (m *AdminAuthMiddleware) hasPermission(admin *AdminAuth, permission string) bool {
	// Super admin has all permissions
	if admin.Role == "super_admin" {
		return true
	}

	// Check specific permissions
	for _, p := range admin.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}

	return false
}

// logAdminActivity logs admin activity
func (m *AdminAuthMiddleware) logAdminActivity(adminID uuid.UUID, method, path, ip string) {
	logger.Info("Admin activity", 
		"admin_id", adminID,
		"method", method,
		"path", path,
		"ip", ip,
		"timestamp", time.Now(),
	)
}

// GetAdminFromContext helper function to get admin from gin context
func GetAdminFromContext(c *gin.Context) (*AdminAuth, bool) {
	adminInterface, exists := c.Get("admin")
	if !exists {
		return nil, false
	}

	admin, ok := adminInterface.(*AdminAuth)
	if !ok {
		return nil, false
	}

	return admin, true
}

// GetAdminIDFromContext helper function to get admin ID from gin context
func GetAdminIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	adminIDInterface, exists := c.Get("admin_id")
	if !exists {
		return uuid.Nil, false
	}

	adminID, ok := adminIDInterface.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return adminID, true
}

// AdminActivityLogger middleware that logs admin activities
func (m *AdminAuthMiddleware) AdminActivityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get admin from context
		adminInterface, exists := c.Get("admin")
		if !exists {
			c.Next()
			return
		}

		admin := adminInterface.(*AdminAuth)

		// Log request details
		startTime := time.Now()

		// Process request
		c.Next()

		// Log response details
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		logger.Info("Admin request completed",
			"admin_id", admin.ID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", statusCode,
			"duration", duration,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// AdminRateLimiter middleware for admin endpoints
func (m *AdminAuthMiddleware) AdminRateLimiter(requests int, window time.Duration) gin.HandlerFunc {
	// This would integrate with the existing rate limiter
	// For now, we'll return a simple pass-through middleware
	return func(c *gin.Context) {
		// In a real implementation, this would check rate limits based on admin ID
		c.Next()
	}
}