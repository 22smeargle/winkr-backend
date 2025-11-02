package middleware

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// SecurityConfig represents security configuration
type SecurityConfig struct {
	// Security headers
	EnableSecurityHeaders bool `json:"enable_security_headers"`
	
	// XSS Protection
	EnableXSSProtection bool `json:"enable_xss_protection"`
	
	// Content Type Options
	EnableContentTypeOptions bool `json:"enable_content_type_options"`
	
	// Frame Options
	EnableFrameOptions bool `json:"enable_frame_options"`
	FrameOptions string `json:"frame_options"` // DENY, SAMEORIGIN, ALLOW-FROM
	
	// HSTS
	EnableHSTS bool `json:"enable_hsts"`
	MaxAge     int  `json:"hsts_max_age"`
	IncludeSubDomains bool `json:"hsts_include_subdomains"`
	Preload          bool `json:"hsts_preload"`
	
	// Content Security Policy
	EnableCSP bool   `json:"enable_csp"`
	CSPPolicy string `json:"csp_policy"`
	
	// IP filtering
	EnableIPFiltering bool     `json:"enable_ip_filtering"`
	WhitelistedIPs    []string `json:"whitelisted_ips"`
	BlacklistedIPs    []string `json:"blacklisted_ips"`
	
	// Request size limits
	MaxRequestSize int64 `json:"max_request_size"`
	MaxHeaderSize  int   `json:"max_header_size"`
	
	// Rate limiting for suspicious IPs
	SuspiciousIPRateLimit int `json:"suspicious_ip_rate_limit"`
	SuspiciousIPWindow   time.Duration `json:"suspicious_ip_window"`
	
	// Bot detection
	EnableBotDetection bool `json:"enable_bot_detection"`
	BotUserAgents     []string `json:"bot_user_agents"`
	
	// Request validation
	EnableRequestValidation bool `json:"enable_request_validation"`
	AllowedMethods        []string `json:"allowed_methods"`
	AllowedContentTypes   []string `json:"allowed_content_types"`
	
	// SSL/TLS
	RequireHTTPS bool `json:"require_https"`
	
	// Skip paths
	SkipPaths []string `json:"skip_paths"`
	
	// Custom security headers
	CustomHeaders map[string]string `json:"custom_headers"`
	
	// Rate limiter for IP tracking
	ipTracker map[string]*ipInfo
	ipMutex   sync.RWMutex
}

// ipInfo tracks IP information for rate limiting
type ipInfo struct {
	requests int
	lastSeen time.Time
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableSecurityHeaders:   true,
		EnableXSSProtection:     true,
		EnableContentTypeOptions: true,
		EnableFrameOptions:       true,
		FrameOptions:           "DENY",
		EnableHSTS:             true,
		MaxAge:                31536000, // 1 year
		IncludeSubDomains:       true,
		Preload:                false,
		EnableCSP:              true,
		CSPPolicy:              "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
		EnableIPFiltering:       false,
		WhitelistedIPs:         []string{},
		BlacklistedIPs:         []string{},
		MaxRequestSize:         10 * 1024 * 1024, // 10MB
		MaxHeaderSize:          8192, // 8KB
		SuspiciousIPRateLimit:   100,
		SuspiciousIPWindow:      time.Minute,
		EnableBotDetection:      true,
		BotUserAgents:          []string{
			"bot",
			"crawler",
			"spider",
			"scraper",
			"cURL",
			"wget",
			"python-requests",
			"java",
			"apache-httpclient",
		},
		EnableRequestValidation: true,
		AllowedMethods:         []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedContentTypes:    []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"text/plain",
		},
		RequireHTTPS: false,
		SkipPaths:   []string{"/health", "/health/db", "/metrics"},
		CustomHeaders: map[string]string{},
		ipTracker:    make(map[string]*ipInfo),
	}
}

// Security returns a security middleware with the given configuration
func Security(config *SecurityConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Compile regex patterns for bot detection
	botPatterns := compileBotPatterns(config.BotUserAgents)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		clientIP := getClientIP(c)

		// Skip security for specified paths
		if shouldSkipSecurity(path, config.SkipPaths) {
			c.Next()
			return
		}

		// Check IP filtering
		if config.EnableIPFiltering {
			if !isIPAllowed(clientIP, config) {
				utils.Forbidden(c, "IP address not allowed")
				c.Abort()
				return
			}
		}

		// Check suspicious IP rate limiting
		if config.EnableIPFiltering && isSuspiciousIP(clientIP, config) {
			utils.RateLimitExceeded(c, "Too many requests from suspicious IP")
			c.Abort()
			return
		}

		// Check bot detection
		if config.EnableBotDetection && isBot(c, botPatterns) {
			utils.Forbidden(c, "Bot access not allowed")
			c.Abort()
			return
		}

		// Check HTTPS requirement
		if config.RequireHTTPS && c.Request.TLS == nil && !isLocalhost(c.Request) {
			httpsURL := "https://" + c.Request.Host + c.Request.URL.RequestURI()
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			c.Abort()
			return
		}

		// Validate request
		if config.EnableRequestValidation {
			if err := validateRequest(c, config); err != nil {
				utils.BadRequest(c, err.Error())
				c.Abort()
				return
			}
		}

		// Set security headers
		if config.EnableSecurityHeaders {
			setSecurityHeaders(c, config)
		}

		// Track IP for rate limiting
		if config.EnableIPFiltering {
			trackIP(clientIP, config)
		}

		c.Next()
	}
}

// getClientIP extracts the real client IP address
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// isIPAllowed checks if IP is allowed based on whitelist/blacklist
func isIPAllowed(ip string, config *SecurityConfig) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	// Check blacklist first
	for _, blacklistedIP := range config.BlacklistedIPs {
		if _, cidr, err := net.ParseCIDR(blacklistedIP); err == nil {
			if cidr.Contains(clientIP) {
				return false
			}
		} else if blacklistedIP == ip {
			return false
		}
	}

	// If whitelist is empty, allow all non-blacklisted IPs
	if len(config.WhitelistedIPs) == 0 {
		return true
	}

	// Check whitelist
	for _, whitelistedIP := range config.WhitelistedIPs {
		if _, cidr, err := net.ParseCIDR(whitelistedIP); err == nil {
			if cidr.Contains(clientIP) {
				return true
			}
		} else if whitelistedIP == ip {
			return true
		}
	}

	return false
}

// isSuspiciousIP checks if IP is making too many requests
func isSuspiciousIP(ip string, config *SecurityConfig) bool {
	config.ipMutex.RLock()
	defer config.ipMutex.RUnlock()

	info, exists := config.ipTracker[ip]
	if !exists {
		return false
	}

	// Check if IP exceeded rate limit
	if info.requests >= config.SuspiciousIPRateLimit {
		// Check if still within window
		if time.Since(info.lastSeen) < config.SuspiciousIPWindow {
			return true
		}
	}

	return false
}

// trackIP tracks IP requests for rate limiting
func trackIP(ip string, config *SecurityConfig) {
	config.ipMutex.Lock()
	defer config.ipMutex.Unlock()

	now := time.Now()
	info, exists := config.ipTracker[ip]
	if !exists {
		config.ipTracker[ip] = &ipInfo{
			requests: 1,
			lastSeen: now,
		}
		return
	}

	// Reset counter if window expired
	if now.Sub(info.lastSeen) > config.SuspiciousIPWindow {
		info.requests = 1
	} else {
		info.requests++
	}
	info.lastSeen = now

	// Clean up old entries
	cleanupIPTracker(config)
}

// cleanupIPTracker removes old entries from IP tracker
func cleanupIPTracker(config *SecurityConfig) {
	now := time.Now()
	for ip, info := range config.ipTracker {
		if now.Sub(info.lastSeen) > config.SuspiciousIPWindow*10 {
			delete(config.ipTracker, ip)
		}
	}
}

// isBot checks if the request is from a bot
func isBot(c *gin.Context, botPatterns []*regexp.Regexp) bool {
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		return true // Empty User-Agent is suspicious
	}

	userAgent = strings.ToLower(userAgent)
	for _, pattern := range botPatterns {
		if pattern.MatchString(userAgent) {
			return true
		}
	}

	return false
}

// compileBotPatterns compiles regex patterns for bot detection
func compileBotPatterns(botUserAgents []string) []*regexp.Regexp {
	var patterns []*regexp.Regexp
	for _, agent := range botUserAgents {
		pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(agent))
		if err == nil {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// validateRequest validates the request
func validateRequest(c *gin.Context, config *SecurityConfig) error {
	// Check method
	if len(config.AllowedMethods) > 0 {
		methodAllowed := false
		for _, method := range config.AllowedMethods {
			if c.Request.Method == method {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			return errors.NewValidationError("method", "Method not allowed")
		}
	}

	// Check content type
	if len(config.AllowedContentTypes) > 0 && c.Request.Method != "GET" && c.Request.Method != "DELETE" {
		contentType := c.GetHeader("Content-Type")
		if contentType != "" {
			contentTypeAllowed := false
			for _, allowedType := range config.AllowedContentTypes {
				if strings.Contains(contentType, allowedType) {
					contentTypeAllowed = true
					break
				}
			}
			if !contentTypeAllowed {
				return errors.NewValidationError("content_type", "Content type not allowed")
			}
		}
	}

	// Check request size
	if config.MaxRequestSize > 0 && c.Request.ContentLength > config.MaxRequestSize {
		return errors.NewValidationError("content_length", "Request too large")
	}

	// Check header size
	if config.MaxHeaderSize > 0 {
		headerSize := 0
		for key, values := range c.Request.Header {
			headerSize += len(key)
			for _, value := range values {
				headerSize += len(value)
			}
		}
		if headerSize > config.MaxHeaderSize {
			return errors.NewValidationError("headers", "Headers too large")
		}
	}

	return nil
}

// setSecurityHeaders sets security headers
func setSecurityHeaders(c *gin.Context, config *SecurityConfig) {
	// XSS Protection
	if config.EnableXSSProtection {
		c.Header("X-XSS-Protection", "1; mode=block")
	}

	// Content Type Options
	if config.EnableContentTypeOptions {
		c.Header("X-Content-Type-Options", "nosniff")
	}

	// Frame Options
	if config.EnableFrameOptions {
		c.Header("X-Frame-Options", config.FrameOptions)
	}

	// HSTS
	if config.EnableHSTS {
		hstsValue := "max-age=" + string(rune(config.MaxAge))
		if config.IncludeSubDomains {
			hstsValue += "; includeSubDomains"
		}
		if config.Preload {
			hstsValue += "; preload"
		}
		c.Header("Strict-Transport-Security", hstsValue)
	}

	// Content Security Policy
	if config.EnableCSP && config.CSPPolicy != "" {
		c.Header("Content-Security-Policy", config.CSPPolicy)
	}

	// Referrer Policy
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions Policy
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// Custom headers
	for key, value := range config.CustomHeaders {
		c.Header(key, value)
	}
}

// isLocalhost checks if the request is from localhost
func isLocalhost(r *http.Request) bool {
	host := r.Host
	return strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
}

// shouldSkipSecurity checks if security should be skipped for path
func shouldSkipSecurity(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// SecurityHeaders returns a middleware that sets security headers
func SecurityHeaders() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.EnableIPFiltering = false
	config.EnableBotDetection = false
	config.EnableRequestValidation = false
	return Security(config)
}

// IPWhitelist returns a middleware that only allows whitelisted IPs
func IPWhitelist(whitelistedIPs []string) gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.EnableIPFiltering = true
	config.WhitelistedIPs = whitelistedIPs
	config.BlacklistedIPs = []string{}
	return Security(config)
}

// IPBlacklist returns a middleware that blocks blacklisted IPs
func IPBlacklist(blacklistedIPs []string) gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.EnableIPFiltering = true
	config.WhitelistedIPs = []string{}
	config.BlacklistedIPs = blacklistedIPs
	return Security(config)
}

// BotProtection returns a middleware that blocks bots
func BotProtection() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.EnableBotDetection = true
	config.EnableIPFiltering = false
	config.EnableRequestValidation = false
	return Security(config)
}

// HTTPSOnly returns a middleware that enforces HTTPS
func HTTPSOnly() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.RequireHTTPS = true
	return Security(config)
}

// CustomSecurity returns a middleware with custom security configuration
func CustomSecurity(customConfig *SecurityConfig) gin.HandlerFunc {
	return Security(customConfig)
}