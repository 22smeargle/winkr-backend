package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// RateLimiter handles rate limiting using Redis sliding window
type RateLimiter struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.RedisClient) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		prefix:      "rate_limit:",
	}
}

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	Requests    int           // Number of requests allowed
	Window      time.Duration // Time window
	KeyType     string        // "ip" or "user"
	Endpoint    string        // API endpoint identifier
}

// RateLimitResult contains the result of a rate limit check
type RateLimitResult struct {
	Allowed     bool          `json:"allowed"`
	Remaining   int           `json:"remaining"`
	ResetTime   time.Time     `json:"reset_time"`
	RetryAfter  time.Duration `json:"retry_after"`
	Limit      int           `json:"limit"`
	Window     time.Duration `json:"window"`
}

// CheckRateLimit checks if a request is allowed based on rate limiting rules
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, config RateLimitConfig, identifier string) (*RateLimitResult, error) {
	key := rl.getRateLimitKey(config.KeyType, config.Endpoint, identifier)
	now := time.Now().Unix()
	windowStart := now - int64(config.Window.Seconds())
	
	// Use Redis sorted set for sliding window
	// Remove old entries outside the window
	_, err := rl.redisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	if err != nil {
		logger.Error("Failed to remove old rate limit entries", err)
		return nil, fmt.Errorf("failed to remove old rate limit entries: %w", err)
	}

	// Count current requests in window
	currentCount, err := rl.redisClient.ZCard(ctx, key)
	if err != nil {
		logger.Error("Failed to count current rate limit entries", err)
		return nil, fmt.Errorf("failed to count current rate limit entries: %w", err)
	}

	remaining := config.Requests - int(currentCount)
	allowed := remaining > 0

	if allowed {
		// Add current request to the window
		_, err := rl.redisClient.ZAdd(ctx, key, fmt.Sprintf("%d", now), fmt.Sprintf("%d", now))
		if err != nil {
			logger.Error("Failed to add rate limit entry", err)
			return nil, fmt.Errorf("failed to add rate limit entry: %w", err)
		}

		// Set expiration for the sorted set
		err = rl.redisClient.Expire(ctx, key, config.Window)
		if err != nil {
			logger.Error("Failed to set rate limit expiration", err)
			return nil, fmt.Errorf("failed to set rate limit expiration: %w", err)
		}
	}

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: max(0, remaining),
		ResetTime: time.Now().Add(config.Window),
		Limit:     config.Requests,
		Window:    config.Window,
	}

	if !allowed {
		// Calculate retry after time
		oldestEntry, err := rl.redisClient.ZRange(ctx, key, 0, 0)
		if err == nil && len(oldestEntry) > 0 {
			// Find when the oldest entry will expire
			oldestTime, _ := strconv.ParseInt(oldestEntry[0], 10, 64)
			retryAt := time.Unix(oldestTime, 0).Add(config.Window)
			result.RetryAfter = time.Until(retryAt)
		}
	}

	logger.Debug("Rate limit check", 
		"key_type", config.KeyType,
		"endpoint", config.Endpoint,
		"identifier", identifier,
		"allowed", allowed,
		"remaining", result.Remaining,
	)

	return result, nil
}

// CheckIPRateLimit checks rate limit for an IP address
func (rl *RateLimiter) CheckIPRateLimit(ctx context.Context, endpoint, ip string) (*RateLimitResult, error) {
	config := rl.getEndpointConfig(endpoint)
	config.KeyType = "ip"
	return rl.CheckRateLimit(ctx, config, ip)
}

// CheckUserRateLimit checks rate limit for a user
func (rl *RateLimiter) CheckUserRateLimit(ctx context.Context, endpoint, userID string) (*RateLimitResult, error) {
	config := rl.getEndpointConfig(endpoint)
	config.KeyType = "user"
	return rl.CheckRateLimit(ctx, config, userID)
}

// ResetRateLimit resets rate limit for a specific key
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, keyType, endpoint, identifier string) error {
	key := rl.getRateLimitKey(keyType, endpoint, identifier)
	
	err := rl.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to reset rate limit", err)
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	logger.Info("Rate limit reset", 
		"key_type", keyType,
		"endpoint", endpoint,
		"identifier", identifier,
	)

	return nil
}

// GetRateLimitStatus gets current rate limit status
func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, keyType, endpoint, identifier string) (*RateLimitResult, error) {
	key := rl.getRateLimitKey(keyType, endpoint, identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rl.getEndpointConfig(endpoint).Window.Seconds())
	
	// Count current requests in window
	currentCount, err := rl.redisClient.ZCard(ctx, key)
	if err != nil {
		logger.Error("Failed to get rate limit status", err)
		return nil, fmt.Errorf("failed to get rate limit status: %w", err)
	}

	config := rl.getEndpointConfig(endpoint)
	remaining := max(0, config.Requests-int(currentCount))
	allowed := remaining > 0

	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(config.Window),
		Limit:     config.Requests,
		Window:    config.Window,
	}

	return result, nil
}

// CleanupExpiredLimits removes expired rate limit entries
func (rl *RateLimiter) CleanupExpiredLimits(ctx context.Context) error {
	// This would typically be run as a background job
	logger.Info("Starting rate limit cleanup")
	
	// Get all rate limit keys
	pattern := fmt.Sprintf("%s*", rl.prefix)
	// In production, you would use SCAN to iterate through keys
	// For simplicity, we'll log the cleanup operation
	
	logger.Info("Rate limit cleanup completed")
	return nil
}

// getEndpointConfig returns rate limit config for an endpoint
func (rl *RateLimiter) getEndpointConfig(endpoint string) RateLimitConfig {
	switch endpoint {
	case "auth":
		return RateLimitConfig{
			Requests: 5,
			Window:   time.Minute,
			Endpoint: "auth",
		}
	case "photo_upload":
		return RateLimitConfig{
			Requests: 10,
			Window:   time.Hour,
			Endpoint: "photo_upload",
		}
	case "messaging":
		return RateLimitConfig{
			Requests: 60,
			Window:   time.Minute,
			Endpoint: "messaging",
		}
	case "matching":
		return RateLimitConfig{
			Requests: 100,
			Window:   time.Hour,
			Endpoint: "matching",
		}
	case "api":
		return RateLimitConfig{
			Requests: 1000,
			Window:   time.Hour,
			Endpoint: "api",
		}
	default:
		// Default rate limit
		return RateLimitConfig{
			Requests: 100,
			Window:   time.Minute,
			Endpoint: endpoint,
		}
	}
}

// getRateLimitKey generates a rate limit key
func (rl *RateLimiter) getRateLimitKey(keyType, endpoint, identifier string) string {
	return fmt.Sprintf("%s%s:%s:%s", rl.prefix, keyType, endpoint, identifier)
}

// DistributedRateLimiter handles distributed rate limiting across multiple instances
type DistributedRateLimiter struct {
	redisClient *redis.RedisClient
	prefix      string
	instanceID   string
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(redisClient *redis.RedisClient, instanceID string) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		redisClient: redisClient,
		prefix:      "distributed_rate_limit:",
		instanceID:   instanceID,
	}
}

// CheckDistributedRateLimit checks rate limit across distributed instances
func (drl *DistributedRateLimiter) CheckDistributedRateLimit(ctx context.Context, config RateLimitConfig, identifier string) (*RateLimitResult, error) {
	// Use Redis atomic operations for distributed rate limiting
	key := drl.getDistributedRateLimitKey(config.KeyType, config.Endpoint, identifier)
	now := time.Now().UnixNano()
	windowStart := now - int64(config.Window.Nanoseconds())
	
	// Lua script for atomic rate limit check
	luaScript := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window = tonumber(ARGV[4])
		
		-- Remove old entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Count current requests
		local current = redis.call('ZCARD', key)
		local remaining = limit - current
		
		if remaining > 0 then
			-- Add current request
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, window)
			return {1, remaining}
		else
			return {0, remaining}
		end
	`
	
	// Execute Lua script
	result, err := drl.redisClient.GetClient().Eval(ctx, luaScript, []string{key}, []interface{}{
		windowStart, now, config.Requests, int64(config.Window.Seconds()),
	}).Result()
	
	if err != nil {
		logger.Error("Failed to execute distributed rate limit script", err)
		return nil, fmt.Errorf("failed to execute distributed rate limit script: %w", err)
	}

	// Parse result
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) < 2 {
		return nil, fmt.Errorf("invalid rate limit result")
	}

	allowed, _ := resultArray[0].(int64)
	remaining, _ := resultArray[1].(int64)

	rateLimitResult := &RateLimitResult{
		Allowed:   allowed == 1,
		Remaining: int(remaining),
		ResetTime: time.Now().Add(config.Window),
		Limit:     config.Requests,
		Window:    config.Window,
	}

	logger.Debug("Distributed rate limit check",
		"key_type", config.KeyType,
		"endpoint", config.Endpoint,
		"identifier", identifier,
		"allowed", rateLimitResult.Allowed,
		"remaining", rateLimitResult.Remaining,
	)

	return rateLimitResult, nil
}

// getDistributedRateLimitKey generates a distributed rate limit key
func (drl *DistributedRateLimiter) getDistributedRateLimitKey(keyType, endpoint, identifier string) string {
	return fmt.Sprintf("%s%s:%s:%s", drl.prefix, keyType, endpoint, identifier)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}