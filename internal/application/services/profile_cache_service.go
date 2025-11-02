package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/profile"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// ProfileCacheService handles profile caching
type ProfileCacheService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*profile.GetProfileResponse, error)
	SetProfile(ctx context.Context, userID uuid.UUID, profile *profile.GetProfileResponse, ttl time.Duration) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	
	GetViewProfile(ctx context.Context, cacheKey string) (*profile.ViewUserProfileResponse, error)
	SetViewProfile(ctx context.Context, cacheKey string, profile *profile.ViewUserProfileResponse, ttl time.Duration) error
	DeleteViewProfile(ctx context.Context, cacheKey string) error
	
	GetMatches(ctx context.Context, cacheKey string) (*profile.GetMatchesResponse, error)
	SetMatches(ctx context.Context, cacheKey string, matches *profile.GetMatchesResponse, ttl time.Duration) error
	DeleteMatches(ctx context.Context, cacheKey string) error
	
	UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error
	DeleteUserLocation(ctx context.Context, userID uuid.UUID) error
	DeleteUserMatches(ctx context.Context, userID uuid.UUID) error
	
	GetUserStats(ctx context.Context, userID uuid.UUID) (*profile.ProfileStats, error)
	SetUserStats(ctx context.Context, userID uuid.UUID, stats *profile.ProfileStats, ttl time.Duration) error
	DeleteUserStats(ctx context.Context, userID uuid.UUID) error
	
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error
	WarmProfileCache(ctx context.Context, userID uuid.UUID) error
}

// RedisProfileCacheService implements ProfileCacheService using Redis
type RedisProfileCacheService struct {
	redisClient RedisClient
	keyPrefix   string
}

// NewRedisProfileCacheService creates a new RedisProfileCacheService instance
func NewRedisProfileCacheService(redisClient RedisClient) *RedisProfileCacheService {
	return &RedisProfileCacheService{
		redisClient: redisClient,
		keyPrefix:   "profile:",
	}
}

// GetProfile gets user profile from cache
func (s *RedisProfileCacheService) GetProfile(ctx context.Context, userID uuid.UUID) (*profile.GetProfileResponse, error) {
	key := s.generateProfileKey(userID)
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	if data == "" {
		return nil, errors.ErrCacheMiss
	}
	
	var profile profile.GetProfileResponse
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, errors.WrapError(err, "Failed to unmarshal profile from cache")
	}
	
	return &profile, nil
}

// SetProfile sets user profile in cache
func (s *RedisProfileCacheService) SetProfile(ctx context.Context, userID uuid.UUID, profile *profile.GetProfileResponse, ttl time.Duration) error {
	key := s.generateProfileKey(userID)
	data, err := json.Marshal(profile)
	if err != nil {
		return errors.WrapError(err, "Failed to marshal profile for cache")
	}
	
	return s.redisClient.Set(ctx, key, string(data), ttl)
}

// DeleteProfile deletes user profile from cache
func (s *RedisProfileCacheService) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	key := s.generateProfileKey(userID)
	return s.redisClient.Delete(ctx, key)
}

// GetViewProfile gets view profile from cache
func (s *RedisProfileCacheService) GetViewProfile(ctx context.Context, cacheKey string) (*profile.ViewUserProfileResponse, error) {
	key := s.keyPrefix + "view:" + cacheKey
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	if data == "" {
		return nil, errors.ErrCacheMiss
	}
	
	var profile profile.ViewUserProfileResponse
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, errors.WrapError(err, "Failed to unmarshal view profile from cache")
	}
	
	return &profile, nil
}

// SetViewProfile sets view profile in cache
func (s *RedisProfileCacheService) SetViewProfile(ctx context.Context, cacheKey string, profile *profile.ViewUserProfileResponse, ttl time.Duration) error {
	key := s.keyPrefix + "view:" + cacheKey
	data, err := json.Marshal(profile)
	if err != nil {
		return errors.WrapError(err, "Failed to marshal view profile for cache")
	}
	
	return s.redisClient.Set(ctx, key, string(data), ttl)
}

// DeleteViewProfile deletes view profile from cache
func (s *RedisProfileCacheService) DeleteViewProfile(ctx context.Context, cacheKey string) error {
	key := s.keyPrefix + "view:" + cacheKey
	return s.redisClient.Delete(ctx, key)
}

// GetMatches gets matches from cache
func (s *RedisProfileCacheService) GetMatches(ctx context.Context, cacheKey string) (*profile.GetMatchesResponse, error) {
	key := s.keyPrefix + "matches:" + cacheKey
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	if data == "" {
		return nil, errors.ErrCacheMiss
	}
	
	var matches profile.GetMatchesResponse
	if err := json.Unmarshal([]byte(data), &matches); err != nil {
		return nil, errors.WrapError(err, "Failed to unmarshal matches from cache")
	}
	
	return &matches, nil
}

// SetMatches sets matches in cache
func (s *RedisProfileCacheService) SetMatches(ctx context.Context, cacheKey string, matches *profile.GetMatchesResponse, ttl time.Duration) error {
	key := s.keyPrefix + "matches:" + cacheKey
	data, err := json.Marshal(matches)
	if err != nil {
		return errors.WrapError(err, "Failed to marshal matches for cache")
	}
	
	return s.redisClient.Set(ctx, key, string(data), ttl)
}

// DeleteMatches deletes matches from cache
func (s *RedisProfileCacheService) DeleteMatches(ctx context.Context, cacheKey string) error {
	key := s.keyPrefix + "matches:" + cacheKey
	return s.redisClient.Delete(ctx, key)
}

// UpdateUserLocation updates user location in cache
func (s *RedisProfileCacheService) UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	key := s.keyPrefix + "location:" + userID.String()
	locationData := map[string]float64{
		"lat": lat,
		"lng": lng,
	}
	
	data, err := json.Marshal(locationData)
	if err != nil {
		return errors.WrapError(err, "Failed to marshal location for cache")
	}
	
	return s.redisClient.Set(ctx, key, string(data), 24*time.Hour)
}

// DeleteUserLocation deletes user location from cache
func (s *RedisProfileCacheService) DeleteUserLocation(ctx context.Context, userID uuid.UUID) error {
	key := s.keyPrefix + "location:" + userID.String()
	return s.redisClient.Delete(ctx, key)
}

// DeleteUserMatches deletes all matches cache for a user
func (s *RedisProfileCacheService) DeleteUserMatches(ctx context.Context, userID uuid.UUID) error {
	pattern := s.keyPrefix + "matches:" + userID.String() + ":*"
	return s.redisClient.DeletePattern(ctx, pattern)
}

// GetUserStats gets user stats from cache
func (s *RedisProfileCacheService) GetUserStats(ctx context.Context, userID uuid.UUID) (*profile.ProfileStats, error) {
	key := s.keyPrefix + "stats:" + userID.String()
	data, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	if data == "" {
		return nil, errors.ErrCacheMiss
	}
	
	var stats profile.ProfileStats
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		return nil, errors.WrapError(err, "Failed to unmarshal stats from cache")
	}
	
	return &stats, nil
}

// SetUserStats sets user stats in cache
func (s *RedisProfileCacheService) SetUserStats(ctx context.Context, userID uuid.UUID, stats *profile.ProfileStats, ttl time.Duration) error {
	key := s.keyPrefix + "stats:" + userID.String()
	data, err := json.Marshal(stats)
	if err != nil {
		return errors.WrapError(err, "Failed to marshal stats for cache")
	}
	
	return s.redisClient.Set(ctx, key, string(data), ttl)
}

// DeleteUserStats deletes user stats from cache
func (s *RedisProfileCacheService) DeleteUserStats(ctx context.Context, userID uuid.UUID) error {
	key := s.keyPrefix + "stats:" + userID.String()
	return s.redisClient.Delete(ctx, key)
}

// InvalidateUserCache invalidates all cache entries for a user
func (s *RedisProfileCacheService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	userIDStr := userID.String()
	
	// Delete all user-related cache entries
	keys := []string{
		s.generateProfileKey(userID),
		s.keyPrefix + "location:" + userIDStr,
		s.keyPrefix + "stats:" + userIDStr,
	}
	
	// Delete matches with pattern
	if err := s.DeleteUserMatches(ctx, userID); err != nil {
		return err
	}
	
	// Delete individual keys
	for _, key := range keys {
		if err := s.redisClient.Delete(ctx, key); err != nil {
			return err
		}
	}
	
	return nil
}

// WarmProfileCache warms up profile cache for a user
func (s *RedisProfileCacheService) WarmProfileCache(ctx context.Context, userID uuid.UUID) error {
	// This would typically be called by a background job
	// to pre-warm caches for active users
	// Implementation depends on specific requirements
	
	// For now, just ensure the profile is cached
	// In a real implementation, you might:
	// 1. Get the user's profile from database
	// 2. Cache it with appropriate TTL
	// 3. Cache related data like stats, recent matches, etc.
	
	return nil
}

// generateProfileKey generates a profile cache key
func (s *RedisProfileCacheService) generateProfileKey(userID uuid.UUID) string {
	return s.keyPrefix + "user:" + userID.String()
}

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitCount  int64 `json:"hit_count"`
	MissCount int64 `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
}

// GetCacheStats gets cache statistics
func (s *RedisProfileCacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	// This would typically get stats from Redis monitoring
	// For now, return placeholder data
	return &CacheStats{
		HitCount:  0,
		MissCount: 0,
		HitRate:   0.0,
	}, nil
}

// ClearExpiredCache clears expired cache entries
func (s *RedisProfileCacheService) ClearExpiredCache(ctx context.Context) error {
	// This would typically be run as a background job
	// Redis handles TTL automatically, so this might not be needed
	// But could be used for cleanup or maintenance
	
	return nil
}