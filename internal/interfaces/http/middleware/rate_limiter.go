package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	// Redis client for distributed rate limiting
	RedisClient *redis.Client
	
	// Default limits
	DefaultRequestsPerMinute int `json:"default_requests_per_minute"`
	DefaultRequestsPerHour   int `json:"default_requests_per_hour"`
	
	// Endpoint-specific limits (endpoint -> requests per minute)
	EndpointLimits map[string]int `json:"endpoint_limits"`
	
	// User type limits (user_type -> requests per minute)
	UserTypeLimits map[string]int `json:"user_type_limits"`
	
	// IP-based limits
	IPBasedLimiting bool `json:"ip_based_limiting"`
	IPLimitPerMinute int `json:"ip_limit_per_minute"`
	
	// Sliding window configuration
	WindowDuration time.Duration `json:"window_duration"`
	
	// Headers to include in response
	IncludeHeaders bool `json:"include_headers"`
	
	// Skip paths
	SkipPaths []string `json:"skip_paths"`
	
	// Key prefix for Redis
	KeyPrefix string `json:"key_prefix"`
}

// DefaultRateLimitConfig returns a default rate limiting configuration
func DefaultRateLimitConfig(redisClient *redis.Client) *RateLimitConfig {
	return &RateLimitConfig{
		RedisClient:              redisClient,
		DefaultRequestsPerMinute: 60,
		DefaultRequestsPerHour:   1000,
		EndpointLimits: map[string]int{
			"/api/v1/auth/login":   5,
			"/api/v1/auth/register": 3,
			"/api/v1/photos":        10,
			"/api/v1/messages":      30,
			"/api/v1/matches/swipe": 100,
		},
		UserTypeLimits: map[string]int{
			"free":     60,
			"premium":  200,
			"platinum": 500,
		},
		IPBasedLimiting:    true,
		IPLimitPerMinute:   100,
		WindowDuration:      time.Minute,
		IncludeHeaders:     true,
		SkipPaths:         []string{"/health", "/health/db", "/metrics"},
		KeyPrefix:         "rate_limit:",
	}
}

// RateLimiter returns a rate limiting middleware with the given configuration
func RateLimiter(config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Skip rate limiting for specified paths
		if shouldSkipRateLimit(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Determine the rate limit for this request
		limit := getRateLimit(path, method, c, config)

		// Generate rate limit keys
		userKey := ""
		if userID, exists := c.Get("user_id"); exists {
			userKey = fmt.Sprintf("%suser:%s:%s", config.KeyPrefix, userID, path)
		}
		
		ipKey := fmt.Sprintf("%sip:%s:%s", config.KeyPrefix, clientIP, path)

		// Check user-based rate limit
		if userKey != "" {
			allowed, remaining, resetTime, err := checkRateLimit(ctx, config.RedisClient, userKey, limit, config.WindowDuration)
			if err != nil {
				utils.Error(c, errors.NewExternalServiceError("redis", err.Error()))
				c.Abort()
				return
			}

			if !allowed {
				handleRateLimitExceeded(c, remaining, resetTime, config)
				c.Abort()
				return
			}

			if config.IncludeHeaders {
				c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
				c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
				c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
			}
		}

		// Check IP-based rate limit if enabled
		if config.IPBasedLimiting {
			ipLimit := config.IPLimitPerMinute
			// Use endpoint-specific IP limit if available
			if endpointLimit, exists := config.EndpointLimits[path]; exists {
				ipLimit = endpointLimit * 2 // Allow more requests per IP than per user
			}

			allowed, _, _, err := checkRateLimit(ctx, config.RedisClient, ipKey, ipLimit, config.WindowDuration)
			if err != nil {
				utils.Error(c, errors.NewExternalServiceError("redis", err.Error()))
				c.Abort()
				return
			}

			if !allowed {
				handleRateLimitExceeded(c, 0, time.Now().Add(config.WindowDuration).Unix(), config)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// checkRateLimit checks if the request is allowed using sliding window algorithm
func checkRateLimit(ctx context.Context, redisClient *redis.Client, key string, limit int, window time.Duration) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-window).Unix()
	
	// Use Redis pipeline for atomic operations
	pipe := redisClient.Pipeline()
	
	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	
	// Count current requests
	countCmd := pipe.ZCard(ctx, key)
	
	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.Unix()),
		Member: now.UnixNano(),
	})
	
	// Set expiration
	pipe.Expire(ctx, key, window)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}
	
	currentCount := countCmd.Val()
	remaining := limit - int(currentCount) - 1
	resetTime := now.Add(window).Unix()
	
	// Allow request if under limit
	if currentCount < int64(limit) {
		return true, remaining, resetTime, nil
	}
	
	return false, 0, resetTime, nil
}

// getRateLimit determines the appropriate rate limit for the request
func getRateLimit(path, method string, c *gin.Context, config *RateLimitConfig) int {
	// Check endpoint-specific limits first
	if limit, exists := config.EndpointLimits[path]; exists {
		return limit
	}

	// Check user type limits
	if userType, exists := c.Get("user_type"); exists {
		if userTypeName, ok := userType.(string); ok {
			if limit, exists := config.UserTypeLimits[userTypeName]; exists {
				return limit
			}
		}
	}

	// Check if user is premium
	if isPremium, exists := c.Get("is_premium"); exists {
		if premium, ok := isPremium.(bool); ok && premium {
			return config.UserTypeLimits["premium"]
		}
	}

	// Return default limit
	return config.DefaultRequestsPerMinute
}

// shouldSkipRateLimit checks if rate limiting should be skipped for the path
func shouldSkipRateLimit(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// handleRateLimitExceeded handles rate limit exceeded responses
func handleRateLimitExceeded(c *gin.Context, remaining int, resetTime int64, config *RateLimitConfig) {
	if config.IncludeHeaders {
		c.Header("X-RateLimit-Limit", "0")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		c.Header("Retry-After", strconv.FormatInt(resetTime-time.Now().Unix(), 10))
	}

	utils.RateLimitExceeded(c, "Rate limit exceeded. Please try again later.")
}

// SlidingWindowRateLimiter implements a more sophisticated sliding window rate limiter
func SlidingWindowRateLimiter(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	config := &RateLimitConfig{
		RedisClient:     redisClient,
		WindowDuration:  window,
		IncludeHeaders:  true,
		KeyPrefix:       "sliding_window:",
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("%s%s:%s", config.KeyPrefix, c.ClientIP(), c.Request.URL.Path)

		allowed, remaining, resetTime, err := checkRateLimit(ctx, redisClient, key, requests, window)
		if err != nil {
			utils.Error(c, errors.NewExternalServiceError("redis", err.Error()))
			c.Abort()
			return
		}

		if !allowed {
			handleRateLimitExceeded(c, remaining, resetTime, config)
			c.Abort()
			return
		}

		if config.IncludeHeaders {
			c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		}

		c.Next()
	}
}

// TokenBucketRateLimiter implements token bucket rate limiting
func TokenBucketRateLimiter(redisClient *redis.Client, capacity int, refillRate int, refillInterval time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("token_bucket:%s:%s", c.ClientIP(), c.Request.URL.Path)

		// Lua script for atomic token bucket operations
		luaScript := `
			local key = KEYS[1]
			local capacity = tonumber(ARGV[1])
			local tokens = tonumber(ARGV[2])
			local interval = tonumber(ARGV[3])
			local now = tonumber(ARGV[4])
			
			local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
			local current_tokens = tonumber(bucket[1]) or capacity
			local last_refill = tonumber(bucket[2]) or now
			
			-- Calculate tokens to add based on time elapsed
			local elapsed = now - last_refill
			local tokens_to_add = math.floor(elapsed / interval * tokens)
			current_tokens = math.min(capacity, current_tokens + tokens_to_add)
			
			-- Check if request can be processed
			if current_tokens >= 1 then
				current_tokens = current_tokens - 1
				redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', now)
				redis.call('EXPIRE', key, interval * 2)
				return {1, current_tokens}
			else
				redis.call('HMSET', key, 'tokens', current_tokens, 'last_refill', now)
				redis.call('EXPIRE', key, interval * 2)
				return {0, current_tokens}
			end
		`

		now := time.Now().Unix()
		result, err := redisClient.Eval(ctx, luaScript, []string{key}, capacity, refillRate, refillInterval.Seconds(), now).Result()
		if err != nil {
			utils.Error(c, errors.NewExternalServiceError("redis", err.Error()))
			c.Abort()
			return
		}

		resultSlice := result.([]interface{})
		allowed := resultSlice[0].(int64) == 1
		remaining := int(resultSlice[1].(int64))

		if !allowed {
			utils.RateLimitExceeded(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(capacity))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Next()
	}
}

// FixedWindowRateLimiter implements simple fixed window rate limiting
func FixedWindowRateLimiter(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		now := time.Now()
		windowStart := now.Truncate(window)
		key := fmt.Sprintf("fixed_window:%s:%s:%d", c.ClientIP(), c.Request.URL.Path, windowStart.Unix())

		// Increment counter
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			utils.Error(c, errors.NewExternalServiceError("redis", err.Error()))
			c.Abort()
			return
		}

		// Set expiration on first request
		if count == 1 {
			redisClient.Expire(ctx, key, window)
		}

		remaining := requests - int(count)
		resetTime := windowStart.Add(window).Unix()

		if count > int64(requests) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
			c.Header("Retry-After", strconv.FormatInt(resetTime-now.Unix(), 10))
			
			utils.RateLimitExceeded(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		c.Next()
	}
}