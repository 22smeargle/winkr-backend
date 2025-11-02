package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// RateLimiterConfig represents rate limiter configuration
type RateLimiterConfig struct {
	// General rate limits
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int

	// Endpoint-specific rate limits
	AuthRequestsPerMinute int
	AuthRequestsPerHour   int
	PasswordResetPerHour  int
	EmailVerifyPerHour    int

	// Account lockout settings
	MaxFailedAttempts    int
	LockoutDuration     time.Duration
}

// EnhancedRateLimiter provides enhanced rate limiting with Redis backend
type EnhancedRateLimiter struct {
	redisClient *redis.RedisClient
	config      RateLimiterConfig
	prefix      string
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(redisClient *redis.RedisClient, config RateLimiterConfig, prefix string) *EnhancedRateLimiter {
	if prefix == "" {
		prefix = "rate_limit"
	}
	return &EnhancedRateLimiter{
		redisClient: redisClient,
		config:      config,
		prefix:      prefix,
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func (rl *EnhancedRateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		path := c.Request.URL.Path

		// Check different rate limits based on endpoint
		if strings.HasPrefix(path, "/api/v1/auth/login") || strings.HasPrefix(path, "/api/v1/auth/register") {
			if !rl.checkAuthRateLimit(c, clientIP, userAgent) {
				return
			}
		} else if strings.HasPrefix(path, "/api/v1/auth/password-reset") {
			if !rl.checkPasswordResetRateLimit(c, clientIP) {
				return
			}
		} else if strings.HasPrefix(path, "/api/v1/auth/verify") {
			if !rl.checkEmailVerifyRateLimit(c, clientIP) {
				return
			}
		} else {
			if !rl.checkGeneralRateLimit(c, clientIP, userAgent) {
				return
			}
		}

		c.Next()
	}
}

// AccountLockoutMiddleware creates an account lockout middleware
func (rl *EnhancedRateLimiter) AccountLockoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1/auth/login") {
			c.Next()
			return
		}

		// For login attempts, check account lockout
		if strings.HasPrefix(path, "/api/v1/auth/login") {
			var loginReq struct {
				Email string `json:"email"`
			}
			c.ShouldBindJSON(&loginReq)

			if loginReq.Email != "" {
				if rl.isAccountLocked(c, loginReq.Email) {
					utils.Error(c, errors.ErrAccountLocked)
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// checkGeneralRateLimit checks general rate limits
func (rl *EnhancedRateLimiter) checkGeneralRateLimit(c *gin.Context, clientIP, userAgent string) bool {
	ctx := c.Request.Context()
	now := time.Now()

	// Check per-minute limit
	minuteKey := fmt.Sprintf("%s:minute:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15:04"))
	minuteCount, err := rl.redisClient.Incr(ctx, minuteKey)
	if err != nil {
		// Log error but don't block request
		return true
	}

	if minuteCount == 1 {
		rl.redisClient.Expire(ctx, minuteKey, time.Minute)
	}

	if minuteCount > int64(rl.config.RequestsPerMinute) {
		rl.sendRateLimitResponse(c, "Too many requests per minute")
		return false
	}

	// Check per-hour limit
	hourKey := fmt.Sprintf("%s:hour:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15"))
	hourCount, err := rl.redisClient.Incr(ctx, hourKey)
	if err != nil {
		return true
	}

	if hourCount == 1 {
		rl.redisClient.Expire(ctx, hourKey, time.Hour)
	}

	if hourCount > int64(rl.config.RequestsPerHour) {
		rl.sendRateLimitResponse(c, "Too many requests per hour")
		return false
	}

	// Check per-day limit
	dayKey := fmt.Sprintf("%s:day:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02"))
	dayCount, err := rl.redisClient.Incr(ctx, dayKey)
	if err != nil {
		return true
	}

	if dayCount == 1 {
		rl.redisClient.Expire(ctx, dayKey, 24*time.Hour)
	}

	if dayCount > int64(rl.config.RequestsPerDay) {
		rl.sendRateLimitResponse(c, "Too many requests per day")
		return false
	}

	return true
}

// checkAuthRateLimit checks authentication-specific rate limits
func (rl *EnhancedRateLimiter) checkAuthRateLimit(c *gin.Context, clientIP, userAgent string) bool {
	ctx := c.Request.Context()
	now := time.Now()

	// Check per-minute limit for auth endpoints
	minuteKey := fmt.Sprintf("%s:auth:minute:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15:04"))
	minuteCount, err := rl.redisClient.Incr(ctx, minuteKey)
	if err != nil {
		return true
	}

	if minuteCount == 1 {
		rl.redisClient.Expire(ctx, minuteKey, time.Minute)
	}

	if minuteCount > int64(rl.config.AuthRequestsPerMinute) {
		rl.sendRateLimitResponse(c, "Too many authentication attempts per minute")
		return false
	}

	// Check per-hour limit for auth endpoints
	hourKey := fmt.Sprintf("%s:auth:hour:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15"))
	hourCount, err := rl.redisClient.Incr(ctx, hourKey)
	if err != nil {
		return true
	}

	if hourCount == 1 {
		rl.redisClient.Expire(ctx, hourKey, time.Hour)
	}

	if hourCount > int64(rl.config.AuthRequestsPerHour) {
		rl.sendRateLimitResponse(c, "Too many authentication attempts per hour")
		return false
	}

	return true
}

// checkPasswordResetRateLimit checks password reset rate limits
func (rl *EnhancedRateLimiter) checkPasswordResetRateLimit(c *gin.Context, clientIP string) bool {
	ctx := c.Request.Context()
	now := time.Now()

	// Check per-hour limit for password reset
	hourKey := fmt.Sprintf("%s:password_reset:hour:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15"))
	hourCount, err := rl.redisClient.Incr(ctx, hourKey)
	if err != nil {
		return true
	}

	if hourCount == 1 {
		rl.redisClient.Expire(ctx, hourKey, time.Hour)
	}

	if hourCount > int64(rl.config.PasswordResetPerHour) {
		rl.sendRateLimitResponse(c, "Too many password reset attempts per hour")
		return false
	}

	return true
}

// checkEmailVerifyRateLimit checks email verification rate limits
func (rl *EnhancedRateLimiter) checkEmailVerifyRateLimit(c *gin.Context, clientIP string) bool {
	ctx := c.Request.Context()
	now := time.Now()

	// Check per-hour limit for email verification
	hourKey := fmt.Sprintf("%s:email_verify:hour:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15"))
	hourCount, err := rl.redisClient.Incr(ctx, hourKey)
	if err != nil {
		return true
	}

	if hourCount == 1 {
		rl.redisClient.Expire(ctx, hourKey, time.Hour)
	}

	if hourCount > int64(rl.config.EmailVerifyPerHour) {
		rl.sendRateLimitResponse(c, "Too many email verification attempts per hour")
		return false
	}

	return true
}

// isAccountLocked checks if an account is locked
func (rl *EnhancedRateLimiter) isAccountLocked(c *gin.Context, email string) bool {
	ctx := c.Request.Context()
	lockKey := fmt.Sprintf("%s:account_lock:%s", rl.prefix, email)

	locked, err := rl.redisClient.Exists(ctx, lockKey)
	if err != nil {
		return false
	}

	return locked
}

// LockAccount locks an account for a specified duration
func (rl *EnhancedRateLimiter) LockAccount(ctx context.Context, email string) error {
	lockKey := fmt.Sprintf("%s:account_lock:%s", rl.prefix, email)
	return rl.redisClient.Set(ctx, lockKey, "1", rl.config.LockoutDuration)
}

// UnlockAccount unlocks an account
func (rl *EnhancedRateLimiter) UnlockAccount(ctx context.Context, email string) error {
	lockKey := fmt.Sprintf("%s:account_lock:%s", rl.prefix, email)
	return rl.redisClient.Del(ctx, lockKey)
}

// IncrementFailedAttempts increments failed login attempts for an email
func (rl *EnhancedRateLimiter) IncrementFailedAttempts(ctx context.Context, email string) error {
	attemptsKey := fmt.Sprintf("%s:failed_attempts:%s", rl.prefix, email)
	
	attempts, err := rl.redisClient.Incr(ctx, attemptsKey)
	if err != nil {
		return err
	}

	// Set expiry on first attempt
	if attempts == 1 {
		rl.redisClient.Expire(ctx, attemptsKey, rl.config.LockoutDuration)
	}

	// Lock account if max attempts reached
	if attempts >= int64(rl.config.MaxFailedAttempts) {
		return rl.LockAccount(ctx, email)
	}

	return nil
}

// ResetFailedAttempts resets failed login attempts for an email
func (rl *EnhancedRateLimiter) ResetFailedAttempts(ctx context.Context, email string) error {
	attemptsKey := fmt.Sprintf("%s:failed_attempts:%s", rl.prefix, email)
	return rl.redisClient.Del(ctx, attemptsKey)
}

// GetFailedAttempts gets the number of failed attempts for an email
func (rl *EnhancedRateLimiter) GetFailedAttempts(ctx context.Context, email string) (int64, error) {
	attemptsKey := fmt.Sprintf("%s:failed_attempts:%s", rl.prefix, email)
	attemptsStr, err := rl.redisClient.Get(ctx, attemptsKey)
	if err != nil {
		return 0, nil // No attempts yet
	}

	attempts, err := strconv.ParseInt(attemptsStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return attempts, nil
}

// sendRateLimitResponse sends a rate limit response
func (rl *EnhancedRateLimiter) sendRateLimitResponse(c *gin.Context, message string) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMinute))
	c.Header("X-RateLimit-Remaining", "0")
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
	
	utils.ErrorWithDetails(c, http.StatusTooManyRequests, "Rate limit exceeded", message)
	c.Abort()
}

// GetRateLimitHeaders returns rate limit headers for the current request
func (rl *EnhancedRateLimiter) GetRateLimitHeaders(c *gin.Context) map[string]string {
	clientIP := c.ClientIP()
	ctx := c.Request.Context()
	now := time.Now()

	minuteKey := fmt.Sprintf("%s:minute:%s:%s", rl.prefix, clientIP, now.Format("2006-01-02-15:04"))
	
	minuteCount, err := rl.redisClient.Get(ctx, minuteKey)
	if err != nil {
		minuteCount = "0"
	}

	headers := map[string]string{
		"X-RateLimit-Limit":     strconv.Itoa(rl.config.RequestsPerMinute),
		"X-RateLimit-Remaining":  strconv.Itoa(rl.config.RequestsPerMinute - int(minuteCount)),
		"X-RateLimit-Reset":     strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10),
	}

	return headers
}