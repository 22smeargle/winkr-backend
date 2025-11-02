package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// CSRFConfig represents CSRF protection configuration
type CSRFConfig struct {
	Enabled          bool
	TokenLength      int
	CookieName       string
	HeaderName       string
	FormFieldName    string
	SecureCookie     bool
	SameSitePolicy   string
	ExcludedMethods  []string
	ExcludedPaths    []string
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config CSRFConfig
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(config CSRFConfig) *CSRFMiddleware {
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "csrf_token"
	}
	if config.SameSitePolicy == "" {
		config.SameSitePolicy = "Strict"
	}

	return &CSRFMiddleware{
		config: config,
	}
}

// GenerateToken generates a new CSRF token
func (csrf *CSRFMiddleware) GenerateToken() (string, error) {
	bytes := make([]byte, csrf.config.TokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Middleware returns the CSRF middleware function
func (csrf *CSRFMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF protection if disabled
		if !csrf.config.Enabled {
			c.Next()
			return
		}

		// Skip for excluded methods
		method := c.Request.Method
		for _, excludedMethod := range csrf.config.ExcludedMethods {
			if strings.ToUpper(method) == strings.ToUpper(excludedMethod) {
				c.Next()
				return
			}
		}

		// Skip for excluded paths
		path := c.Request.URL.Path
		for _, excludedPath := range csrf.config.ExcludedPaths {
			if strings.HasPrefix(path, excludedPath) {
				c.Next()
				return
			}
		}

		// For GET requests, generate and set CSRF token
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			token, err := csrf.GenerateToken()
			if err != nil {
				utils.Error(c, errors.NewInternalError("Failed to generate CSRF token"))
				c.Abort()
				return
			}

			// Set CSRF token in cookie
			c.SetCookie(csrf.config.CookieName, token, "/", "", true, csrf.config.SecureCookie, true, csrf.config.SameSitePolicy)
			
			// Also set token in context for easy access in templates
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For state-changing requests (POST, PUT, DELETE, PATCH), validate CSRF token
		if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
			if !csrf.validateToken(c) {
				utils.Error(c, errors.NewForbiddenError("CSRF token validation failed"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// validateToken validates the CSRF token for the current request
func (csrf *CSRFMiddleware) validateToken(c *gin.Context) bool {
	// Get token from cookie
	cookieToken, err := c.Cookie(csrf.config.CookieName)
	if err != nil {
		return false
	}

	// Get token from header or form
	var requestToken string
	
	// Try header first
	requestToken = c.GetHeader(csrf.config.HeaderName)
	
	// If not in header, try form field
	if requestToken == "" {
		requestToken = c.PostForm(csrf.config.FormFieldName)
	}

	// If still not found, try query parameter
	if requestToken == "" {
		requestToken = c.Query(csrf.config.FormFieldName)
	}

	// Compare tokens
	return requestToken != "" && requestToken == cookieToken
}

// GetToken returns the CSRF token from the context
func GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		return token.(string)
	}
	return ""
}

// SetCSRFHeaders sets CSRF-related headers for the response
func SetCSRFHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
}

// CSRFTokenResponse represents CSRF token response
type CSRFTokenResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Error   *errors.AppError `json:"error,omitempty"`
}

// GetCSRFTokenHandler returns a handler to get CSRF token
func GetCSRFTokenHandler(csrf *CSRFMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := csrf.GenerateToken()
		if err != nil {
			utils.Error(c, errors.NewInternalError("Failed to generate CSRF token"))
			return
		}

		response := &CSRFTokenResponse{
			Success: true,
			Token:   token,
		}

		utils.Success(c, response)
	}
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:         true,
		TokenLength:     32,
		CookieName:      "csrf_token",
		HeaderName:      "X-CSRF-Token",
		FormFieldName:   "csrf_token",
		SecureCookie:    true,
		SameSitePolicy:  "Strict",
		ExcludedMethods: []string{"GET", "HEAD", "OPTIONS"},
		ExcludedPaths:   []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/auth/password-reset",
			"/api/v1/auth/verify",
		},
	}
}

// DevelopmentCSRFConfig returns a development-friendly CSRF configuration
func DevelopmentCSRFConfig() CSRFConfig {
	config := DefaultCSRFConfig()
	config.SecureCookie = false
	config.SameSitePolicy = "Lax"
	return config
}