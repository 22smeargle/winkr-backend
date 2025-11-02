package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ProfileRateLimiter handles rate limiting for profile endpoints
type ProfileRateLimiter struct {
	rateLimiter *cache.RateLimiter
}

// NewProfileRateLimiter creates a new ProfileRateLimiter instance
func NewProfileRateLimiter(redisClient cache.RedisClient) *ProfileRateLimiter {
	return &ProfileRateLimiter{
		rateLimiter: cache.NewRateLimiter(redisClient),
	}
}

// RateLimit applies rate limiting based on endpoint
func (r *ProfileRateLimiter) RateLimit(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// If no user ID, use IP-based limiting
			r.applyIPRateLimit(c, endpoint)
			return
		}

		// Apply user-based rate limiting
		r.applyUserRateLimit(c, endpoint, userID.(string))
	}
}

// applyIPRateLimit applies IP-based rate limiting
func (r *ProfileRateLimiter) applyIPRateLimit(c *gin.Context, endpoint string) {
	clientIP := c.ClientIP()
	key := "profile:" + endpoint + ":ip:" + clientIP
	
	// Get rate limit config for endpoint
	limit, window := r.getRateLimitConfig(endpoint)
	
	// Check rate limit
	allowed, resetTime, err := r.rateLimiter.AllowRequest(c.Request.Context(), key, limit, window)
	if err != nil {
		utils.Error(c, err)
		c.Abort()
		return
	}
	
	if !allowed {
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		utils.RateLimitExceeded(c, "Rate limit exceeded. Please try again later.")
		c.Abort()
		return
	}
	
	// Get current count
	current, err := r.rateLimiter.GetCurrentCount(c.Request.Context(), key)
	if err == nil {
		remaining := limit - current
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
	}
	
	c.Next()
}

// applyUserRateLimit applies user-based rate limiting
func (r *ProfileRateLimiter) applyUserRateLimit(c *gin.Context, endpoint, userID string) {
	key := "profile:" + endpoint + ":user:" + userID
	
	// Get rate limit config for endpoint
	limit, window := r.getRateLimitConfig(endpoint)
	
	// Check rate limit
	allowed, resetTime, err := r.rateLimiter.AllowRequest(c.Request.Context(), key, limit, window)
	if err != nil {
		utils.Error(c, err)
		c.Abort()
		return
	}
	
	if !allowed {
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", "0")
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		utils.RateLimitExceeded(c, "Rate limit exceeded. Please try again later.")
		c.Abort()
		return
	}
	
	// Get current count
	current, err := r.rateLimiter.GetCurrentCount(c.Request.Context(), key)
	if err == nil {
		remaining := limit - current
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
	}
	
	c.Next()
}

// getRateLimitConfig gets rate limit configuration for endpoint
func (r *ProfileRateLimiter) getRateLimitConfig(endpoint string) (int, time.Duration) {
	configs := map[string]struct {
		limit  int
		window time.Duration
	}{
		"get-profile":     {limit: 100, window: time.Hour},
		"update-profile":  {limit: 20, window: time.Hour},
		"view-profile":    {limit: 200, window: time.Hour},
		"update-location": {limit: 10, window: time.Hour},
		"get-matches":    {limit: 50, window: time.Hour},
		"delete-account":  {limit: 5, window: 24 * time.Hour},
	}
	
	if config, exists := configs[endpoint]; exists {
		return config.limit, config.window
	}
	
	// Default rate limit
	return 100, time.Hour
}

// GetUserRateLimitStatus gets current rate limit status for a user
func (r *ProfileRateLimiter) GetUserRateLimitStatus(ctx context.Context, userID, endpoint string) (*RateLimitStatus, error) {
	key := "profile:" + endpoint + ":user:" + userID
	limit, window := r.getRateLimitConfig(endpoint)
	
	current, err := r.rateLimiter.GetCurrentCount(ctx, key)
	if err != nil {
		return nil, err
	}
	
	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}
	
	resetTime, err := r.rateLimiter.GetResetTime(ctx, key)
	if err != nil {
		return nil, err
	}
	
	return &RateLimitStatus{
		Limit:     limit,
		Current:   current,
		Remaining:  remaining,
		ResetTime:  resetTime,
		Window:     window,
	}, nil
}

// RateLimitStatus represents rate limit status
type RateLimitStatus struct {
	Limit    int           `json:"limit"`
	Current  int           `json:"current"`
	Remaining int           `json:"remaining"`
	ResetTime int64         `json:"reset_time"`
	Window   time.Duration `json:"window"`
}

// ResetUserRateLimit resets rate limit for a user
func (r *ProfileRateLimiter) ResetUserRateLimit(ctx context.Context, userID, endpoint string) error {
	key := "profile:" + endpoint + ":user:" + userID
	return r.rateLimiter.Reset(ctx, key)
}

// ResetIPRateLimit resets rate limit for an IP
func (r *ProfileRateLimiter) ResetIPRateLimit(ctx context.Context, ip, endpoint string) error {
	key := "profile:" + endpoint + ":ip:" + ip
	return r.rateLimiter.Reset(ctx, key)
}

// GetRateLimitStats gets rate limit statistics
func (r *ProfileRateLimiter) GetRateLimitStats(ctx context.Context, endpoint string) (*RateLimitStats, error) {
	// This would typically get stats from Redis monitoring
	// For now, return placeholder data
	return &RateLimitStats{
		TotalRequests: 0,
		BlockedRequests: 0,
		AverageRequestsPerMinute: 0.0,
		PeakRequestsPerMinute: 0,
	}, nil
}

// RateLimitStats represents rate limit statistics
type RateLimitStats struct {
	TotalRequests              int64   `json:"total_requests"`
	BlockedRequests            int64   `json:"blocked_requests"`
	AverageRequestsPerMinute    float64 `json:"average_requests_per_minute"`
	PeakRequestsPerMinute     int     `json:"peak_requests_per_minute"`
}