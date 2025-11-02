package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AuthRateLimiter provides rate limiting for authentication endpoints
type AuthRateLimiter struct {
	redisClient *redis.RedisClient
}

// NewAuthRateLimiter creates a new auth rate limiter
func NewAuthRateLimiter(redisClient *redis.RedisClient) *AuthRateLimiter {
	return &AuthRateLimiter{
		redisClient: redisClient,
	}
}

// RateLimitConfig defines rate limits for different endpoints
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour  int
	BlockDuration    time.Duration
}

// GetRateLimitConfig returns rate limit config for endpoint
func (arl *AuthRateLimiter) GetRateLimitConfig(endpoint string) RateLimitConfig {
	switch endpoint {
	case "register", "login", "refresh":
		return RateLimitConfig{
			RequestsPerMinute: 5,
			RequestsPerHour:  20,
			BlockDuration:    15 * time.Minute,
		}
	case "password-reset", "password-reset/confirm":
		return RateLimitConfig{
			RequestsPerMinute: 3,
			RequestsPerHour:  10,
			BlockDuration:    30 * time.Minute,
		}
	case "verify", "verify/send":
		return RateLimitConfig{
			RequestsPerMinute: 3,
			RequestsPerHour:  15,
			BlockDuration:    10 * time.Minute,
		}
	default:
		return RateLimitConfig{
			RequestsPerMinute: 10,
			RequestsPerHour:  100,
			BlockDuration:    5 * time.Minute,
		}
	}
}

// RateLimit returns a middleware for rate limiting
func (arl *AuthRateLimiter) RateLimit(endpoint string) gin.HandlerFunc {
	config := arl.GetRateLimitConfig(endpoint)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		
		// Create unique key for IP and endpoint
		key := fmt.Sprintf("rate_limit:%s:%s:%s", endpoint, clientIP, arl.hashUserAgent(userAgent))
		
		// Check current rate limits
		allowed, retryAfter, err := arl.checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "internal_error",
					"message": "Failed to check rate limit",
				},
			})
			c.Abort()
			return
		}
		
		if !allowed {
			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(retryAfter).Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(retryAfter.Seconds(), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "rate_limit_exceeded",
					"message": "Too many requests. Please try again later.",
					"details": fmt.Sprintf("Retry after %v", retryAfter),
				},
			})
			c.Abort()
			return
		}
		
		// Get current count for headers
		currentCount, _, _ := arl.getCurrentCount(c.Request.Context(), key)
		remaining := config.RequestsPerMinute - currentCount
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
		
		c.Next()
	}
}

// checkRateLimit checks if request is allowed
func (arl *AuthRateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (bool, time.Duration, error) {
	now := time.Now()
	
	// Check minute-based limit
	minuteKey := fmt.Sprintf("%s:minute", key)
	minuteCount, err := arl.redisClient.Get(ctx, minuteKey)
	if err != nil {
		// Key doesn't exist, start from 0
		minuteCount = "0"
	}
	
	var currentMinuteCount int
	if minuteCount != "" {
		_, err := fmt.Sscanf(minuteCount, "%d", &currentMinuteCount)
		if err != nil {
			currentMinuteCount = 0
		}
	}
	
	// Check if minute limit exceeded
	if currentMinuteCount >= config.RequestsPerMinute {
		// Check when the limit will reset
		resetTime, err := arl.redisClient.TTL(ctx, minuteKey)
		if err != nil {
			return false, config.BlockDuration, nil
		}
		return false, resetTime, nil
	}
	
	// Check hour-based limit
	hourKey := fmt.Sprintf("%s:hour", key)
	hourCount, err := arl.redisClient.Get(ctx, hourKey)
	if err != nil {
		// Key doesn't exist, start from 0
		hourCount = "0"
	}
	
	var currentHourCount int
	if hourCount != "" {
		_, err := fmt.Sscanf(hourCount, "%d", &currentHourCount)
		if err != nil {
			currentHourCount = 0
		}
	}
	
	// Check if hour limit exceeded
	if currentHourCount >= config.RequestsPerHour {
		// Check when the limit will reset
		resetTime, err := arl.redisClient.TTL(ctx, hourKey)
		if err != nil {
			return false, config.BlockDuration, nil
		}
		return false, resetTime, nil
	}
	
	return true, 0, nil
}

// incrementRateLimit increments the rate limit counters
func (arl *AuthRateLimiter) incrementRateLimit(ctx context.Context, key string, config RateLimitConfig) error {
	now := time.Now()
	
	// Increment minute counter
	minuteKey := fmt.Sprintf("%s:minute", key)
	err := arl.redisClient.Incr(ctx, minuteKey)
	if err != nil {
		return fmt.Errorf("failed to increment minute counter: %w", err)
	}
	
	// Set expiry for minute counter (1 minute)
	err = arl.redisClient.Expire(ctx, minuteKey, time.Minute)
	if err != nil {
		// Log error but continue
	}
	
	// Increment hour counter
	hourKey := fmt.Sprintf("%s:hour", key)
	err = arl.redisClient.Incr(ctx, hourKey)
	if err != nil {
		return fmt.Errorf("failed to increment hour counter: %w", err)
	}
	
	// Set expiry for hour counter (1 hour)
	err = arl.redisClient.Expire(ctx, hourKey, time.Hour)
	if err != nil {
		// Log error but continue
	}
	
	return nil
}

// getCurrentCount gets the current count for rate limit headers
func (arl *AuthRateLimiter) getCurrentCount(ctx context.Context, key string) (int, time.Duration, error) {
	minuteKey := fmt.Sprintf("%s:minute", key)
	count, err := arl.redisClient.Get(ctx, minuteKey)
	if err != nil {
		return 0, 0, nil
	}
	
	var currentCount int
	if count != "" {
		_, err := fmt.Sscanf(count, "%d", &currentCount)
		if err != nil {
			currentCount = 0
		}
	}
	
	// Get TTL for reset time
	ttl, err := arl.redisClient.TTL(ctx, minuteKey)
	if err != nil {
		return currentCount, 0, nil
	}
	
	return currentCount, ttl, nil
}

// hashUserAgent creates a hash of user agent for rate limiting
func (arl *AuthRateLimiter) hashUserAgent(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	
	// Simple hash - in production, you might want to use a proper hash function
	return utils.MD5Hash(userAgent)[:8]
}

// SuspiciousActivityDetector detects suspicious authentication patterns
type SuspiciousActivityDetector struct {
	redisClient *redis.RedisClient
}

// NewSuspiciousActivityDetector creates a new suspicious activity detector
func NewSuspiciousActivityDetector(redisClient *redis.RedisClient) *SuspiciousActivityDetector {
	return &SuspiciousActivityDetector{
		redisClient: redisClient,
	}
}

// CheckSuspiciousActivity checks for suspicious patterns
func (sad *SuspiciousActivityDetector) CheckSuspiciousActivity(c *gin.Context) error {
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	// Check for multiple IPs from same user agent
	if err := sad.checkMultipleIPs(userAgent, clientIP); err != nil {
		return err
	}
	
	// Check for rapid password attempts
	if c.Request.URL.Path == "/api/v1/auth/login" {
		if err := sad.checkRapidLoginAttempts(clientIP); err != nil {
			return err
		}
	}
	
	// Check for unusual user agent patterns
	if err := sad.checkUnusualUserAgent(userAgent); err != nil {
		return err
	}
	
	return nil
}

// checkMultipleIPs checks for multiple IPs from same user agent
func (sad *SuspiciousActivityDetector) checkMultipleIPs(userAgent, currentIP string) error {
	key := fmt.Sprintf("user_agent_ips:%s", utils.MD5Hash(userAgent))
	
	// Get existing IPs
	ips, err := sad.redisClient.SMembers(c.Request.Context(), key)
	if err != nil {
		return nil // Don't fail on error
	}
	
	// Add current IP
	err = sad.redisClient.SAdd(c.Request.Context(), key, currentIP)
	if err != nil {
		return nil
	}
	
	// Set expiry (1 hour)
	sad.redisClient.Expire(c.Request.Context(), key, time.Hour)
	
	// Check if too many IPs (more than 3)
	if len(ips) >= 3 {
		return errors.NewAppError(429, "Suspicious activity detected", "Multiple IP addresses detected")
	}
	
	return nil
}

// checkRapidLoginAttempts checks for rapid login attempts
func (sad *SuspiciousActivityDetector) checkRapidLoginAttempts(ip string) error {
	key := fmt.Sprintf("login_attempts:%s", ip)
	
	// Get current attempts
	attempts, err := sad.redisClient.Get(c.Request.Context(), key)
	if err != nil {
		// Key doesn't exist, start from 0
		attempts = "0"
	}
	
	var currentAttempts int
	if attempts != "" {
		_, err := fmt.Sscanf(attempts, "%d", &currentAttempts)
		if err != nil {
			currentAttempts = 0
		}
	}
	
	// Increment attempts
	currentAttempts++
	err = sad.redisClient.Set(c.Request.Context(), key, fmt.Sprintf("%d", currentAttempts), 5*time.Minute)
	if err != nil {
		return nil
	}
	
	// Check if too many attempts (more than 10 in 5 minutes)
	if currentAttempts > 10 {
		return errors.NewAppError(429, "Too many login attempts", "Please wait before trying again")
	}
	
	return nil
}

// checkUnusualUserAgent checks for unusual user agent patterns
func (sad *SuspiciousActivityDetector) checkUnusualUserAgent(userAgent string) error {
	if userAgent == "" {
		return errors.NewValidationError("user_agent", "User agent is required")
	}
	
	// Check for common bot patterns
	botPatterns := []string{
		"bot", "crawler", "spider", "scraper", "curl", "wget",
		"python", "java", "perl", "ruby", "php", "node",
	}
	
	lowerUserAgent := strings.ToLower(userAgent)
	for _, pattern := range botPatterns {
		if strings.Contains(lowerUserAgent, pattern) {
			return errors.NewValidationError("user_agent", "Automated requests are not allowed")
		}
	}
	
	return nil
}