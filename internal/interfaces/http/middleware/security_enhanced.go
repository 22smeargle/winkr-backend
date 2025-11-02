package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableCSRF            bool
	EnableRateLimit       bool
	EnableSecurityHeaders  bool
	EnableInputValidation bool
	MaxRequestSize        int64
	AllowedOrigins       []string
	AllowedMethods       []string
	AllowedHeaders       []string
	BlockedUserAgents   []string
	SuspiciousPatterns   []string
}

// SecurityMiddleware provides comprehensive security features
type SecurityMiddleware struct {
	config SecurityConfig
	csrfMiddleware *CSRFMiddleware
	rateLimiter    *EnhancedRateLimiter
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig, csrf *CSRFMiddleware, rateLimiter *EnhancedRateLimiter) *SecurityMiddleware {
	return &SecurityMiddleware{
		config:         config,
		csrfMiddleware: csrf,
		rateLimiter:    rateLimiter,
	}
}

// Middleware returns the security middleware function
func (sm *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply security features based on configuration

		// 1. Security headers
		if sm.config.EnableSecurityHeaders {
			sm.setSecurityHeaders(c)
		}

		// 2. Request size validation
		if sm.config.MaxRequestSize > 0 {
			sm.validateRequestSize(c)
		}

		// 3. User agent blocking
		if len(sm.config.BlockedUserAgents) > 0 {
			if sm.blockUserAgent(c) {
				utils.Error(c, errors.ErrForbidden)
				c.Abort()
				return
			}
		}

		// 4. Suspicious pattern detection
		if len(sm.config.SuspiciousPatterns) > 0 {
			if sm.detectSuspiciousPatterns(c) {
				utils.Error(c, errors.ErrForbidden)
				c.Abort()
				return
			}
		}

		// 5. Input validation
		if sm.config.EnableInputValidation {
			sm.validateInput(c)
		}

		// 6. CORS headers (if configured)
		if len(sm.config.AllowedOrigins) > 0 {
			sm.setCORSHeaders(c)
		}

		c.Next()
	}
}

// ApplyToRouter applies all security middleware to a router group
func (sm *SecurityMiddleware) ApplyToRouter(router *gin.RouterGroup) {
	// Apply rate limiting
	if sm.config.EnableRateLimit && sm.rateLimiter != nil {
		router.Use(sm.rateLimiter.RateLimitMiddleware())
		router.Use(sm.rateLimiter.AccountLockoutMiddleware())
	}

	// Apply CSRF protection
	if sm.config.EnableCSRF && sm.csrfMiddleware != nil {
		router.Use(sm.csrfMiddleware.Middleware())
	}

	// Apply general security middleware
	router.Use(sm.Middleware())
}

// setSecurityHeaders sets security-related HTTP headers
func (sm *SecurityMiddleware) setSecurityHeaders(c *gin.Context) {
	// Prevent clickjacking
	c.Header("X-Frame-Options", "DENY")

	// Prevent MIME type sniffing
	c.Header("X-Content-Type-Options", "nosniff")

	// Enable XSS protection
	c.Header("X-XSS-Protection", "1; mode=block")

	// Force HTTPS (if enabled)
	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// Content Security Policy
	csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';"
	c.Header("Content-Security-Policy", csp)

	// Referrer Policy
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions Policy
	permissionsPolicy := "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), fullscreen=(self), payment=(), accelerometer=(),"
	c.Header("Permissions-Policy", permissionsPolicy)
}

// validateRequestSize validates the request size
func (sm *SecurityMiddleware) validateRequestSize(c *gin.Context) {
	if c.Request.ContentLength > sm.config.MaxRequestSize {
		utils.ErrorWithDetails(c, http.StatusRequestEntityTooLarge, "Request too large", 
			"Maximum request size is "+string(rune(sm.config.MaxRequestSize))+" bytes")
		c.Abort()
		return
	}
}

// blockUserAgent blocks requests from suspicious user agents
func (sm *SecurityMiddleware) blockUserAgent(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return false
	}

	userAgent = strings.ToLower(userAgent)
	for _, blockedUA := range sm.config.BlockedUserAgents {
		if strings.Contains(userAgent, strings.ToLower(blockedUA)) {
			return true
		}
	}

	return false
}

// detectSuspiciousPatterns detects suspicious patterns in requests
func (sm *SecurityMiddleware) detectSuspiciousPatterns(c *gin.Context) bool {
	// Check URL path
	path := c.Request.URL.Path
	for _, pattern := range sm.config.SuspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, path)
		if matched {
			return true
		}
	}

	// Check query parameters
	query := c.Request.URL.RawQuery
	for _, pattern := range sm.config.SuspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	// Check common attack patterns in headers
	headers := []string{"User-Agent", "Referer", "X-Forwarded-For"}
	for _, header := range headers {
		value := c.GetHeader(header)
		for _, pattern := range sm.config.SuspiciousPatterns {
			matched, _ := regexp.MatchString(pattern, value)
			if matched {
				return true
			}
		}
	}

	return false
}

// validateInput validates input for common attack patterns
func (sm *SecurityMiddleware) validateInput(c *gin.Context) {
	// SQL Injection patterns
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(select\s+.*\s+from)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(update\s+.*\s+set)`,
		`(?i)(drop\s+table)`,
		`(?i)(create\s+table)`,
		`(?i)(alter\s+table)`,
		`(?i)(exec\s*\()`,
		`(?i)(execute\s*\()`,
		`(?i)(sp_executesql)`,
		`(?i)(xp_cmdshell)`,
		`(?i)(xp_regread)`,
		`(?i)(xp_regwrite)`,
	}

	// XSS patterns
	xssPatterns := []string{
		`(?i)(<script[^>]*>.*?</script>)`,
		`(?i)(javascript:)`,
		`(?i)(onload\s*=)`,
		`(?i)(onerror\s*=)`,
		`(?i)(onclick\s*=)`,
		`(?i)(onmouseover\s*=)`,
		`(?i)(<iframe[^>]*>)`,
		`(?i)(<object[^>]*>)`,
		`(?i)(<embed[^>]*>)`,
		`(?i)(eval\s*\()`,
		`(?i)(expression\s*\()`,
	}

	// Check all patterns against request data
	allPatterns := append(sqlPatterns, xssPatterns...)
	
	// Check URL and query
	url := c.Request.URL.String()
	for _, pattern := range allPatterns {
		matched, _ := regexp.MatchString(pattern, url)
		if matched {
			utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid input detected", "Suspicious pattern found in request")
			c.Abort()
			return
		}
	}

	// Check form data if present
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		c.Request.ParseForm()
		for key, values := range c.Request.PostForm {
			for _, value := range values {
				for _, pattern := range allPatterns {
					matched, _ := regexp.MatchString(pattern, value)
					if matched {
						utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid input detected", 
							"Suspicious pattern found in form field: "+key)
						c.Abort()
						return
					}
				}
			}
		}
	}
}

// setCORSHeaders sets CORS headers
func (sm *SecurityMiddleware) setCORSHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")
	
	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range sm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	if len(sm.config.AllowedMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
	}

	if len(sm.config.AllowedHeaders) > 0 {
		c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
	}

	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Max-Age", "86400") // 24 hours
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		EnableCSRF:            true,
		EnableRateLimit:       true,
		EnableSecurityHeaders:  true,
		EnableInputValidation: true,
		MaxRequestSize:        10 * 1024 * 1024, // 10MB
		AllowedOrigins:        []string{"*"},
		AllowedMethods:        []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:        []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		BlockedUserAgents:     []string{
			"sqlmap",
			"nikto",
			"dirbuster",
			"nmap",
			"masscan",
			"zap",
			"burp",
		},
		SuspiciousPatterns:     []string{
			`\.\./`,          // Directory traversal
			`(\.\.){2,}`,     // Multiple dots
			`(<|>|>|%3e|%3c)`, // HTML tags
			`(union|select|insert|update|delete|drop|create|alter|exec|execute)`, // SQL keywords
			`(script|javascript|vbscript|onload|onerror)`, // XSS patterns
			`(\.\./|\.\.\\)`, // Path traversal
			`(etc/passwd|etc/shadow)`, // System files
			`(cmd\.exe|powershell|bash|sh)`, // Command injection
		},
	}
}

// DevelopmentSecurityConfig returns a development-friendly security configuration
func DevelopmentSecurityConfig() SecurityConfig {
	config := DefaultSecurityConfig()
	config.EnableCSRF = false // Disable CSRF in development for easier testing
	config.EnableInputValidation = false // Disable strict validation in development
	config.MaxRequestSize = 50 * 1024 * 1024 // 50MB for development
	return config
}