package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RateLimiter handles rate limiting for discovery and matching operations
type RateLimiter interface {
	// Swipe rate limiting
	AllowSwipe(ctx context.Context, userID uuid.UUID) (bool, error)
	AllowSuperLike(ctx context.Context, userID uuid.UUID) (bool, error)
	GetSwipeCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error)
	GetSuperLikeCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error)

	// Discovery rate limiting
	AllowDiscovery(ctx context.Context, userID uuid.UUID) (bool, error)
	GetDiscoveryCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error)

	// Generic rate limiting
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Reset(ctx context.Context, key string) error
}

// RedisRateLimiter implements RateLimiter using Redis
type RedisRateLimiter struct {
	client RedisClient
	config RateLimitConfig
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Swipe limits
	SwipesPerHour    int           `json:"swipes_per_hour"`
	SwipesPerDay     int           `json:"swipes_per_day"`
	SuperLikesPerDay  int           `json:"super_likes_per_day"`

	// Discovery limits
	DiscoveryPerHour int           `json:"discovery_per_hour"`
	DiscoveryPerDay  int           `json:"discovery_per_day"`

	// Windows
	HourWindow time.Duration `json:"hour_window"`
	DayWindow  time.Duration `json:"day_window"`
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		SwipesPerHour:   100,
		SwipesPerDay:    1000,
		SuperLikesPerDay: 5,
		DiscoveryPerHour: 50,
		DiscoveryPerDay:  500,
		HourWindow:      time.Hour,
		DayWindow:       24 * time.Hour,
	}
}

// NewRedisRateLimiter creates a new RedisRateLimiter
func NewRedisRateLimiter(client RedisClient, config RateLimitConfig) *RedisRateLimiter {
	if config.SwipesPerHour == 0 {
		config = DefaultRateLimitConfig()
	}

	return &RedisRateLimiter{
		client: client,
		config: config,
	}
}

// AllowSwipe checks if user is allowed to swipe
func (r *RedisRateLimiter) AllowSwipe(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Check hourly limit
	hourlyKey := fmt.Sprintf("swipes:hour:%s", userID.String())
	hourlyCount, err := r.getCount(ctx, hourlyKey)
	if err != nil {
		return false, fmt.Errorf("failed to get hourly swipe count: %w", err)
	}

	if hourlyCount >= r.config.SwipesPerHour {
		return false, nil
	}

	// Check daily limit
	dailyKey := fmt.Sprintf("swipes:day:%s", userID.String())
	dailyCount, err := r.getCount(ctx, dailyKey)
	if err != nil {
		return false, fmt.Errorf("failed to get daily swipe count: %w", err)
	}

	if dailyCount >= r.config.SwipesPerDay {
		return false, nil
	}

	// Increment hourly counter
	err = r.incrementCount(ctx, hourlyKey, r.config.HourWindow)
	if err != nil {
		return false, fmt.Errorf("failed to increment hourly counter: %w", err)
	}

	// Increment daily counter
	err = r.incrementCount(ctx, dailyKey, r.config.DayWindow)
	if err != nil {
		return false, fmt.Errorf("failed to increment daily counter: %w", err)
	}

	return true, nil
}

// AllowSuperLike checks if user is allowed to super like
func (r *RedisRateLimiter) AllowSuperLike(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Check daily super like limit
	dailyKey := fmt.Sprintf("super_likes:day:%s", userID.String())
	dailyCount, err := r.getCount(ctx, dailyKey)
	if err != nil {
		return false, fmt.Errorf("failed to get daily super like count: %w", err)
	}

	if dailyCount >= r.config.SuperLikesPerDay {
		return false, nil
	}

	// Increment daily counter
	err = r.incrementCount(ctx, dailyKey, r.config.DayWindow)
	if err != nil {
		return false, fmt.Errorf("failed to increment super like counter: %w", err)
	}

	return true, nil
}

// GetSwipeCount gets swipe count for a user within a time window
func (r *RedisRateLimiter) GetSwipeCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error) {
	var key string
	switch window {
	case r.config.HourWindow:
		key = fmt.Sprintf("swipes:hour:%s", userID.String())
	case r.config.DayWindow:
		key = fmt.Sprintf("swipes:day:%s", userID.String())
	default:
		return 0, fmt.Errorf("unsupported time window: %v", window)
	}

	return r.getCount(ctx, key)
}

// GetSuperLikeCount gets super like count for a user within a time window
func (r *RedisRateLimiter) GetSuperLikeCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error) {
	if window != r.config.DayWindow {
		return 0, fmt.Errorf("super likes only tracked daily")
	}

	key := fmt.Sprintf("super_likes:day:%s", userID.String())
	return r.getCount(ctx, key)
}

// AllowDiscovery checks if user is allowed to make discovery requests
func (r *RedisRateLimiter) AllowDiscovery(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Check hourly limit
	hourlyKey := fmt.Sprintf("discovery:hour:%s", userID.String())
	hourlyCount, err := r.getCount(ctx, hourlyKey)
	if err != nil {
		return false, fmt.Errorf("failed to get hourly discovery count: %w", err)
	}

	if hourlyCount >= r.config.DiscoveryPerHour {
		return false, nil
	}

	// Check daily limit
	dailyKey := fmt.Sprintf("discovery:day:%s", userID.String())
	dailyCount, err := r.getCount(ctx, dailyKey)
	if err != nil {
		return false, fmt.Errorf("failed to get daily discovery count: %w", err)
	}

	if dailyCount >= r.config.DiscoveryPerDay {
		return false, nil
	}

	// Increment hourly counter
	err = r.incrementCount(ctx, hourlyKey, r.config.HourWindow)
	if err != nil {
		return false, fmt.Errorf("failed to increment hourly counter: %w", err)
	}

	// Increment daily counter
	err = r.incrementCount(ctx, dailyKey, r.config.DayWindow)
	if err != nil {
		return false, fmt.Errorf("failed to increment daily counter: %w", err)
	}

	return true, nil
}

// GetDiscoveryCount gets discovery count for a user within a time window
func (r *RedisRateLimiter) GetDiscoveryCount(ctx context.Context, userID uuid.UUID, window time.Duration) (int, error) {
	var key string
	switch window {
	case r.config.HourWindow:
		key = fmt.Sprintf("discovery:hour:%s", userID.String())
	case r.config.DayWindow:
		key = fmt.Sprintf("discovery:day:%s", userID.String())
	default:
		return 0, fmt.Errorf("unsupported time window: %v", window)
	}

	return r.getCount(ctx, key)
}

// Allow checks generic rate limit
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := r.getCount(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to get count: %w", err)
	}

	if count >= limit {
		return false, nil
	}

	err = r.incrementCount(ctx, key, window)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	return true, nil
}

// Reset resets a rate limit counter
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Delete(ctx, key)
}

// getCount gets current count for a key
func (r *RedisRateLimiter) getCount(ctx context.Context, key string) (int, error) {
	val, err := r.client.Get(ctx, key)
	if err != nil {
		return 0, nil // Key doesn't exist, count is 0
	}

	if count, ok := val.(int); ok {
		return count, nil
	}

	if count, ok := val.(int64); ok {
		return int(count), nil
	}

	if count, ok := val.(float64); ok {
		return int(count), nil
	}

	return 0, fmt.Errorf("invalid count type for key %s", key)
}

// incrementCount increments counter for a key with TTL
func (r *RedisRateLimiter) incrementCount(ctx context.Context, key string, ttl time.Duration) error {
	// Use Redis INCR with EXPIRE for atomic increment with TTL
	// This would typically be a Redis pipeline operation
	// For simplicity, we'll use a basic approach
	
	current, err := r.getCount(ctx, key)
	if err != nil {
		current = 0
	}

	newCount := current + 1
	return r.client.Set(ctx, key, newCount, ttl)
}

// RateLimitInfo contains rate limit information
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
	Window    time.Duration `json:"window"`
}

// GetRateLimitInfo gets current rate limit information
func (r *RedisRateLimiter) GetRateLimitInfo(ctx context.Context, userID uuid.UUID, operation string) (*RateLimitInfo, error) {
	var limit int
	var window time.Duration
	var key string

	switch operation {
	case "swipe":
		limit = r.config.SwipesPerHour
		window = r.config.HourWindow
		key = fmt.Sprintf("swipes:hour:%s", userID.String())
	case "super_like":
		limit = r.config.SuperLikesPerDay
		window = r.config.DayWindow
		key = fmt.Sprintf("super_likes:day:%s", userID.String())
	case "discovery":
		limit = r.config.DiscoveryPerHour
		window = r.config.HourWindow
		key = fmt.Sprintf("discovery:hour:%s", userID.String())
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}

	count, err := r.getCount(ctx, key)
	if err != nil {
		count = 0
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (simplified - would need to get TTL from Redis)
	resetTime := time.Now().Add(window)

	return &RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		ResetTime: resetTime,
		Window:    window,
	}, nil
}