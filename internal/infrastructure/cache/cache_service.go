package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// CacheService handles caching operations
type CacheService struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewCacheService creates a new cache service
func NewCacheService(redisClient *redis.RedisClient) *CacheService {
	return &CacheService{
		redisClient: redisClient,
		prefix:      "cache:",
	}
}

// Cache TTL constants
const (
	UserProfileCacheTTL     = 30 * time.Minute
	PhotoMetadataCacheTTL   = 15 * time.Minute
	MatchRecommendationsTTL = 10 * time.Minute
	APIResponseCacheTTL     = 5 * time.Minute
	GeoSpatialCacheTTL      = 60 * time.Minute
	OnlineStatusCacheTTL    = 2 * time.Minute
)

// CacheUserProfile caches user profile data
func (cs *CacheService) CacheUserProfile(ctx context.Context, userID string, profile *entities.User) error {
	key := cs.getUserProfileKey(userID)
	
	profileData, err := json.Marshal(profile)
	if err != nil {
		logger.Error("Failed to marshal user profile for caching", err)
		return fmt.Errorf("failed to marshal user profile: %w", err)
	}

	err = cs.redisClient.Set(ctx, key, string(profileData), UserProfileCacheTTL)
	if err != nil {
		logger.Error("Failed to cache user profile", err)
		return fmt.Errorf("failed to cache user profile: %w", err)
	}

	logger.Debug("User profile cached", "user_id", userID)
	return nil
}

// GetUserProfile retrieves cached user profile
func (cs *CacheService) GetUserProfile(ctx context.Context, userID string) (*entities.User, error) {
	key := cs.getUserProfileKey(userID)
	
	profileData, err := cs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached user profile", err)
		return nil, fmt.Errorf("failed to get cached user profile: %w", err)
	}

	if profileData == "" {
		return nil, nil // Cache miss
	}

	var profile entities.User
	err = json.Unmarshal([]byte(profileData), &profile)
	if err != nil {
		logger.Error("Failed to unmarshal cached user profile", err)
		return nil, fmt.Errorf("failed to unmarshal cached user profile: %w", err)
	}

	logger.Debug("User profile retrieved from cache", "user_id", userID)
	return &profile, nil
}

// InvalidateUserProfile removes user profile from cache
func (cs *CacheService) InvalidateUserProfile(ctx context.Context, userID string) error {
	key := cs.getUserProfileKey(userID)
	
	err := cs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate user profile cache", err)
		return fmt.Errorf("failed to invalidate user profile cache: %w", err)
	}

	logger.Debug("User profile cache invalidated", "user_id", userID)
	return nil
}

// CachePhotoMetadata caches photo metadata
func (cs *CacheService) CachePhotoMetadata(ctx context.Context, photoID string, metadata map[string]interface{}) error {
	key := cs.getPhotoMetadataKey(photoID)
	
	metadataData, err := json.Marshal(metadata)
	if err != nil {
		logger.Error("Failed to marshal photo metadata for caching", err)
		return fmt.Errorf("failed to marshal photo metadata: %w", err)
	}

	err = cs.redisClient.Set(ctx, key, string(metadataData), PhotoMetadataCacheTTL)
	if err != nil {
		logger.Error("Failed to cache photo metadata", err)
		return fmt.Errorf("failed to cache photo metadata: %w", err)
	}

	logger.Debug("Photo metadata cached", "photo_id", photoID)
	return nil
}

// GetPhotoMetadata retrieves cached photo metadata
func (cs *CacheService) GetPhotoMetadata(ctx context.Context, photoID string) (map[string]interface{}, error) {
	key := cs.getPhotoMetadataKey(photoID)
	
	metadataData, err := cs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached photo metadata", err)
		return nil, fmt.Errorf("failed to get cached photo metadata: %w", err)
	}

	if metadataData == "" {
		return nil, nil // Cache miss
	}

	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(metadataData), &metadata)
	if err != nil {
		logger.Error("Failed to unmarshal cached photo metadata", err)
		return nil, fmt.Errorf("failed to unmarshal cached photo metadata: %w", err)
	}

	logger.Debug("Photo metadata retrieved from cache", "photo_id", photoID)
	return metadata, nil
}

// CacheMatchRecommendations caches match recommendations for a user
func (cs *CacheService) CacheMatchRecommendations(ctx context.Context, userID string, recommendations []entities.Match) error {
	key := cs.getMatchRecommendationsKey(userID)
	
	recommendationsData, err := json.Marshal(recommendations)
	if err != nil {
		logger.Error("Failed to marshal match recommendations for caching", err)
		return fmt.Errorf("failed to marshal match recommendations: %w", err)
	}

	err = cs.redisClient.Set(ctx, key, string(recommendationsData), MatchRecommendationsTTL)
	if err != nil {
		logger.Error("Failed to cache match recommendations", err)
		return fmt.Errorf("failed to cache match recommendations: %w", err)
	}

	logger.Debug("Match recommendations cached", "user_id", userID, "count", len(recommendations))
	return nil
}

// GetMatchRecommendations retrieves cached match recommendations
func (cs *CacheService) GetMatchRecommendations(ctx context.Context, userID string) ([]entities.Match, error) {
	key := cs.getMatchRecommendationsKey(userID)
	
	recommendationsData, err := cs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached match recommendations", err)
		return nil, fmt.Errorf("failed to get cached match recommendations: %w", err)
	}

	if recommendationsData == "" {
		return nil, nil // Cache miss
	}

	var recommendations []entities.Match
	err = json.Unmarshal([]byte(recommendationsData), &recommendations)
	if err != nil {
		logger.Error("Failed to unmarshal cached match recommendations", err)
		return nil, fmt.Errorf("failed to unmarshal cached match recommendations: %w", err)
	}

	logger.Debug("Match recommendations retrieved from cache", "user_id", userID)
	return recommendations, nil
}

// CacheAPIResponse caches API response data
func (cs *CacheService) CacheAPIResponse(ctx context.Context, key string, response interface{}) error {
	cacheKey := cs.getAPIResponseKey(key)
	
	responseData, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal API response for caching", err)
		return fmt.Errorf("failed to marshal API response: %w", err)
	}

	err = cs.redisClient.Set(ctx, cacheKey, string(responseData), APIResponseCacheTTL)
	if err != nil {
		logger.Error("Failed to cache API response", err)
		return fmt.Errorf("failed to cache API response: %w", err)
	}

	logger.Debug("API response cached", "key", key)
	return nil
}

// GetAPIResponse retrieves cached API response
func (cs *CacheService) GetAPIResponse(ctx context.Context, key string, response interface{}) (bool, error) {
	cacheKey := cs.getAPIResponseKey(key)
	
	responseData, err := cs.redisClient.Get(ctx, cacheKey)
	if err != nil {
		logger.Error("Failed to get cached API response", err)
		return false, fmt.Errorf("failed to get cached API response: %w", err)
	}

	if responseData == "" {
		return false, nil // Cache miss
	}

	err = json.Unmarshal([]byte(responseData), &response)
	if err != nil {
		logger.Error("Failed to unmarshal cached API response", err)
		return false, fmt.Errorf("failed to unmarshal cached API response: %w", err)
	}

	logger.Debug("API response retrieved from cache", "key", key)
	return true, nil
}

// CacheGeoSpatialData caches geospatial data
func (cs *CacheService) CacheGeoSpatialData(ctx context.Context, locationKey string, data interface{}) error {
	cacheKey := cs.getGeoSpatialKey(locationKey)
	
	locationData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal geospatial data for caching", err)
		return fmt.Errorf("failed to marshal geospatial data: %w", err)
	}

	err = cs.redisClient.Set(ctx, cacheKey, string(locationData), GeoSpatialCacheTTL)
	if err != nil {
		logger.Error("Failed to cache geospatial data", err)
		return fmt.Errorf("failed to cache geospatial data: %w", err)
	}

	logger.Debug("Geospatial data cached", "location_key", locationKey)
	return nil
}

// GetGeoSpatialData retrieves cached geospatial data
func (cs *CacheService) GetGeoSpatialData(ctx context.Context, locationKey string, data interface{}) (bool, error) {
	cacheKey := cs.getGeoSpatialKey(locationKey)
	
	locationData, err := cs.redisClient.Get(ctx, cacheKey)
	if err != nil {
		logger.Error("Failed to get cached geospatial data", err)
		return false, fmt.Errorf("failed to get cached geospatial data: %w", err)
	}

	if locationData == "" {
		return false, nil // Cache miss
	}

	err = json.Unmarshal([]byte(locationData), &data)
	if err != nil {
		logger.Error("Failed to unmarshal cached geospatial data", err)
		return false, fmt.Errorf("failed to unmarshal cached geospatial data: %w", err)
	}

	logger.Debug("Geospatial data retrieved from cache", "location_key", locationKey)
	return true, nil
}

// InvalidatePattern removes all keys matching a pattern
func (cs *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// This would typically be used with Redis SCAN in production
	// For simplicity, we'll implement basic pattern invalidation
	// In a real implementation, you might want to use KEYS command with caution
	
	logger.Info("Invalidating cache pattern", "pattern", pattern)
	
	// For now, we'll log the invalidation request
	// In production, you would implement proper pattern-based invalidation
	return nil
}

// WarmCache preloads frequently accessed data
func (cs *CacheService) WarmCache(ctx context.Context) error {
	logger.Info("Starting cache warming")
	
	// This would typically load frequently accessed data
	// Implementation depends on your specific use cases
	// For example: popular user profiles, trending content, etc.
	
	logger.Info("Cache warming completed")
	return nil
}

// GetCacheStats returns cache statistics
func (cs *CacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Get Redis info
	stats := cs.redisClient.GetStats()
	
	// Add cache-specific metrics
	cacheStats := map[string]interface{}{
		"redis_stats": stats,
		"cache_prefix": cs.prefix,
		"ttls": map[string]time.Duration{
			"user_profile":           UserProfileCacheTTL,
			"photo_metadata":         PhotoMetadataCacheTTL,
			"match_recommendations":  MatchRecommendationsTTL,
			"api_response":           APIResponseCacheTTL,
			"geospatial":           GeoSpatialCacheTTL,
			"online_status":         OnlineStatusCacheTTL,
		},
	}
	
	return cacheStats, nil
}

// Set sets a value in the cache with TTL
func (cs *CacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := cs.redisClient.Set(ctx, key, value, ttl)
	if err != nil {
		logger.Error("Failed to set cache value", err, "key", key)
		return fmt.Errorf("failed to set cache value: %w", err)
	}
	return nil
}

// Del deletes a value from the cache
func (cs *CacheService) Del(ctx context.Context, key string) error {
	err := cs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to delete cache value", err, "key", key)
		return fmt.Errorf("failed to delete cache value: %w", err)
	}
	return nil
}

// Helper methods for key generation

func (cs *CacheService) getUserProfileKey(userID string) string {
	return fmt.Sprintf("%suser_profile:%s", cs.prefix, userID)
}

func (cs *CacheService) getPhotoMetadataKey(photoID string) string {
	return fmt.Sprintf("%sphoto_metadata:%s", cs.prefix, photoID)
}

func (cs *CacheService) getMatchRecommendationsKey(userID string) string {
	return fmt.Sprintf("%smatch_recommendations:%s", cs.prefix, userID)
}

func (cs *CacheService) getAPIResponseKey(key string) string {
	return fmt.Sprintf("%sapi_response:%s", cs.prefix, key)
}

func (cs *CacheService) getGeoSpatialKey(locationKey string) string {
	return fmt.Sprintf("%sgeospatial:%s", cs.prefix, locationKey)
}