package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// CacheService handles caching operations for discovery and matching
type CacheService interface {
	// Discovery caching
	GetDiscoveryUsers(ctx context.Context, key string) (*dto.DiscoverUsersResponse, error)
	SetDiscoveryUsers(ctx context.Context, key string, response *dto.DiscoverUsersResponse, ttl time.Duration) error
	InvalidateUserDiscoveryCache(ctx context.Context, userID uuid.UUID) error

	// Matches caching
	GetMatches(ctx context.Context, key string) (*dto.GetMatchesResponse, error)
	SetMatches(ctx context.Context, key string, response *dto.GetMatchesResponse, ttl time.Duration) error
	GetMatch(ctx context.Context, key string) (*entities.Match, error)
	SetMatch(ctx context.Context, key string, match *entities.Match, ttl time.Duration) error
	GetUserMatches(ctx context.Context, key string) (*UserMatchesCache, error)
	SetUserMatches(ctx context.Context, key string, cache *UserMatchesCache, ttl time.Duration) error

	// Potential matches caching
	GetPotentialMatches(ctx context.Context, key string) (*PotentialMatchesCache, error)
	SetPotentialMatches(ctx context.Context, key string, cache *PotentialMatchesCache, ttl time.Duration) error

	// Swipe caching
	GetSwipedUsers(ctx context.Context, key string) ([]uuid.UUID, error)
	SetSwipedUsers(ctx context.Context, key string, userIDs []uuid.UUID, ttl time.Duration) error
	GetSwipeStats(ctx context.Context, key string) (*SwipeStatsCache, error)
	SetSwipeStats(ctx context.Context, key string, stats *SwipeStatsCache, ttl time.Duration) error

	// Discovery stats caching
	GetDiscoveryStats(ctx context.Context, key string) (*dto.GetDiscoveryStatsResponse, error)
	SetDiscoveryStats(ctx context.Context, key string, response *dto.GetDiscoveryStatsResponse, ttl time.Duration) error

	// Generic cache operations
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// RedisCacheService implements CacheService using Redis
type RedisCacheService struct {
	client RedisClient
}

// NewRedisCacheService creates a new RedisCacheService
func NewRedisCacheService(client RedisClient) *RedisCacheService {
	return &RedisCacheService{
		client: client,
	}
}

// GetDiscoveryUsers gets discovery users from cache
func (r *RedisCacheService) GetDiscoveryUsers(ctx context.Context, key string) (*dto.DiscoverUsersResponse, error) {
	var response dto.DiscoverUsersResponse
	err := r.client.GetJSON(ctx, key, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetDiscoveryUsers sets discovery users in cache
func (r *RedisCacheService) SetDiscoveryUsers(ctx context.Context, key string, response *dto.DiscoverUsersResponse, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, response, ttl)
}

// InvalidateUserDiscoveryCache invalidates all discovery cache keys for a user
func (r *RedisCacheService) InvalidateUserDiscoveryCache(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("discovery:%s:*", userID.String())
	return r.client.DeletePattern(ctx, pattern)
}

// GetMatches gets matches from cache
func (r *RedisCacheService) GetMatches(ctx context.Context, key string) (*dto.GetMatchesResponse, error) {
	var response dto.GetMatchesResponse
	err := r.client.GetJSON(ctx, key, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetMatches sets matches in cache
func (r *RedisCacheService) SetMatches(ctx context.Context, key string, response *dto.GetMatchesResponse, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, response, ttl)
}

// GetMatch gets a single match from cache
func (r *RedisCacheService) GetMatch(ctx context.Context, key string) (*entities.Match, error) {
	var match entities.Match
	err := r.client.GetJSON(ctx, key, &match)
	if err != nil {
		return nil, err
	}
	return &match, nil
}

// SetMatch sets a single match in cache
func (r *RedisCacheService) SetMatch(ctx context.Context, key string, match *entities.Match, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, match, ttl)
}

// GetUserMatches gets user matches from cache
func (r *RedisCacheService) GetUserMatches(ctx context.Context, key string) (*UserMatchesCache, error) {
	var cache UserMatchesCache
	err := r.client.GetJSON(ctx, key, &cache)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// SetUserMatches sets user matches in cache
func (r *RedisCacheService) SetUserMatches(ctx context.Context, key string, cache *UserMatchesCache, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, cache, ttl)
}

// GetPotentialMatches gets potential matches from cache
func (r *RedisCacheService) GetPotentialMatches(ctx context.Context, key string) (*PotentialMatchesCache, error) {
	var cache PotentialMatchesCache
	err := r.client.GetJSON(ctx, key, &cache)
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// SetPotentialMatches sets potential matches in cache
func (r *RedisCacheService) SetPotentialMatches(ctx context.Context, key string, cache *PotentialMatchesCache, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, cache, ttl)
}

// GetSwipedUsers gets swiped users from cache
func (r *RedisCacheService) GetSwipedUsers(ctx context.Context, key string) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.client.GetJSON(ctx, key, &userIDs)
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}

// SetSwipedUsers sets swiped users in cache
func (r *RedisCacheService) SetSwipedUsers(ctx context.Context, key string, userIDs []uuid.UUID, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, userIDs, ttl)
}

// GetSwipeStats gets swipe stats from cache
func (r *RedisCacheService) GetSwipeStats(ctx context.Context, key string) (*SwipeStatsCache, error) {
	var stats SwipeStatsCache
	err := r.client.GetJSON(ctx, key, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// SetSwipeStats sets swipe stats in cache
func (r *RedisCacheService) SetSwipeStats(ctx context.Context, key string, stats *SwipeStatsCache, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, stats, ttl)
}

// GetDiscoveryStats gets discovery stats from cache
func (r *RedisCacheService) GetDiscoveryStats(ctx context.Context, key string) (*dto.GetDiscoveryStatsResponse, error) {
	var response dto.GetDiscoveryStatsResponse
	err := r.client.GetJSON(ctx, key, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetDiscoveryStats sets discovery stats in cache
func (r *RedisCacheService) SetDiscoveryStats(ctx context.Context, key string, response *dto.GetDiscoveryStatsResponse, ttl time.Duration) error {
	return r.client.SetJSON(ctx, key, response, ttl)
}

// Get gets generic value from cache
func (r *RedisCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	return r.client.Get(ctx, key)
}

// Set sets generic value in cache
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl)
}

// Delete deletes a key from cache
func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	return r.client.Delete(ctx, key)
}

// DeletePattern deletes keys matching a pattern
func (r *RedisCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return r.client.DeletePattern(ctx, pattern)
}

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
}

// SwipeStatsCache represents cached swipe statistics
type SwipeStatsCache struct {
	TotalSwipes     int64 `json:"total_swipes"`
	TotalLikes      int64 `json:"total_likes"`
	TotalPasses     int64 `json:"total_passes"`
	SwipesToday     int64 `json:"swipes_today"`
	SwipesThisWeek int64 `json:"swipes_this_week"`
	SwipesThisMonth int64 `json:"swipes_this_month"`
	LikeRate        float64 `json:"like_rate"`
	CachedAt        time.Time `json:"cached_at"`
}

// PotentialMatchesCache represents cached potential matches
type PotentialMatchesCache struct {
	Users []*entities.User `json:"users"`
	Total int64            `json:"total"`
	CachedAt time.Time       `json:"cached_at"`
}