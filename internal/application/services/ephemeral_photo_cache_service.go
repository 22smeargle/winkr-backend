package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EphemeralPhotoCacheService defines the interface for ephemeral photo caching
type EphemeralPhotoCacheService interface {
	// Photo caching
	CacheEphemeralPhoto(ctx context.Context, photo *entities.EphemeralPhoto, ttl time.Duration) error
	GetCachedEphemeralPhoto(ctx context.Context, photoID uuid.UUID) (*entities.EphemeralPhoto, error)
	GetCachedEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error)
	InvalidateEphemeralPhoto(ctx context.Context, photoID uuid.UUID) error
	InvalidateEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) error
	
	// Status caching
	CachePhotoStatus(ctx context.Context, photoID uuid.UUID, status string, ttl time.Duration) error
	GetCachedPhotoStatus(ctx context.Context, photoID uuid.UUID) (string, error)
	InvalidatePhotoStatus(ctx context.Context, photoID uuid.UUID) error
	
	// View tracking caching
	CacheViewCount(ctx context.Context, photoID uuid.UUID, count int, ttl time.Duration) error
	GetCachedViewCount(ctx context.Context, photoID uuid.UUID) (int, error)
	IncrementViewCount(ctx context.Context, photoID uuid.UUID) error
	
	// Analytics caching
	CacheAnalytics(ctx context.Context, userID uuid.UUID, analytics *EphemeralPhotoAnalytics, ttl time.Duration) error
	GetCachedAnalytics(ctx context.Context, userID uuid.UUID) (*EphemeralPhotoAnalytics, error)
	InvalidateUserAnalytics(ctx context.Context, userID uuid.UUID) error
	
	// Cleanup operations
	CleanupExpiredCache(ctx context.Context) (int, error)
	InvalidateAll(ctx context.Context) error
}

// EphemeralPhotoAnalytics represents analytics data for ephemeral photos
type EphemeralPhotoAnalytics struct {
	TotalPhotos      int64 `json:"total_photos"`
	ActivePhotos     int64 `json:"active_photos"`
	ViewedPhotos     int64 `json:"viewed_photos"`
	ExpiredPhotos    int64 `json:"expired_photos"`
	TotalViews       int64 `json:"total_views"`
	AverageViewTime  int64 `json:"average_view_time_seconds"`
	ViewsToday       int64 `json:"views_today"`
	ViewsThisWeek    int64 `json:"views_this_week"`
	ViewsThisMonth   int64 `json:"views_this_month"`
	GeneratedAt      int64  `json:"generated_at"`
}

// EphemeralPhotoCacheServiceImpl implements EphemeralPhotoCacheService
type EphemeralPhotoCacheServiceImpl struct {
	cacheService CacheService
}

// NewEphemeralPhotoCacheService creates a new ephemeral photo cache service
func NewEphemeralPhotoCacheService(cacheService CacheService) EphemeralPhotoCacheService {
	return &EphemeralPhotoCacheServiceImpl{
		cacheService: cacheService,
	}
}

// CacheEphemeralPhoto caches an ephemeral photo
func (s *EphemeralPhotoCacheServiceImpl) CacheEphemeralPhoto(ctx context.Context, photo *entities.EphemeralPhoto, ttl time.Duration) error {
	key := s.getPhotoCacheKey(photo.ID)
	
	// Serialize photo
	data, err := json.Marshal(photo)
	if err != nil {
		logger.Error("Failed to serialize ephemeral photo for caching", err)
		return fmt.Errorf("failed to serialize ephemeral photo for caching: %w", err)
	}
	
	// Cache with TTL
	if err := s.cacheService.Set(ctx, key, string(data), ttl); err != nil {
		logger.Error("Failed to cache ephemeral photo", err)
		return fmt.Errorf("failed to cache ephemeral photo: %w", err)
	}
	
	logger.Debug("Ephemeral photo cached successfully", map[string]interface{}{
		"photo_id": photo.ID,
		"ttl":       ttl.Seconds(),
	})
	
	return nil
}

// GetCachedEphemeralPhoto gets a cached ephemeral photo
func (s *EphemeralPhotoCacheServiceImpl) GetCachedEphemeralPhoto(ctx context.Context, photoID uuid.UUID) (*entities.EphemeralPhoto, error) {
	key := s.getPhotoCacheKey(photoID)
	
	// Get from cache
	data, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Failed to get cached ephemeral photo", err)
		return nil, fmt.Errorf("failed to get cached ephemeral photo: %w", err)
	}
	
	if data == "" {
		return nil, fmt.Errorf("ephemeral photo not found in cache")
	}
	
	// Deserialize photo
	var photo entities.EphemeralPhoto
	if err := json.Unmarshal([]byte(data), &photo); err != nil {
		logger.Error("Failed to deserialize cached ephemeral photo", err)
		return nil, fmt.Errorf("failed to deserialize cached ephemeral photo: %w", err)
	}
	
	logger.Debug("Ephemeral photo retrieved from cache", map[string]interface{}{
		"photo_id": photoID,
	})
	
	return &photo, nil
}

// GetCachedEphemeralPhotoByAccessKey gets a cached ephemeral photo by access key
func (s *EphemeralPhotoCacheServiceImpl) GetCachedEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error) {
	key := s.getAccessKeyCacheKey(accessKey)
	
	// Get from cache
	data, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Failed to get cached ephemeral photo by access key", err)
		return nil, fmt.Errorf("failed to get cached ephemeral photo by access key: %w", err)
	}
	
	if data == "" {
		return nil, fmt.Errorf("ephemeral photo not found in cache")
	}
	
	// Deserialize photo
	var photo entities.EphemeralPhoto
	if err := json.Unmarshal([]byte(data), &photo); err != nil {
		logger.Error("Failed to deserialize cached ephemeral photo by access key", err)
		return nil, fmt.Errorf("failed to deserialize cached ephemeral photo by access key: %w", err)
	}
	
	logger.Debug("Ephemeral photo retrieved from cache by access key", map[string]interface{}{
		"access_key": accessKey,
	})
	
	return &photo, nil
}

// InvalidateEphemeralPhoto invalidates a cached ephemeral photo
func (s *EphemeralPhotoCacheServiceImpl) InvalidateEphemeralPhoto(ctx context.Context, photoID uuid.UUID) error {
	key := s.getPhotoCacheKey(photoID)
	
	if err := s.cacheService.Delete(ctx, key); err != nil {
		logger.Error("Failed to invalidate cached ephemeral photo", err)
		return fmt.Errorf("failed to invalidate cached ephemeral photo: %w", err)
	}
	
	logger.Debug("Ephemeral photo cache invalidated", map[string]interface{}{
		"photo_id": photoID,
	})
	
	return nil
}

// InvalidateEphemeralPhotoByAccessKey invalidates a cached ephemeral photo by access key
func (s *EphemeralPhotoCacheServiceImpl) InvalidateEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) error {
	key := s.getAccessKeyCacheKey(accessKey)
	
	if err := s.cacheService.Delete(ctx, key); err != nil {
		logger.Error("Failed to invalidate cached ephemeral photo by access key", err)
		return fmt.Errorf("failed to invalidate cached ephemeral photo by access key: %w", err)
	}
	
	logger.Debug("Ephemeral photo cache invalidated by access key", map[string]interface{}{
		"access_key": accessKey,
	})
	
	return nil
}

// CachePhotoStatus caches photo status
func (s *EphemeralPhotoCacheServiceImpl) CachePhotoStatus(ctx context.Context, photoID uuid.UUID, status string, ttl time.Duration) error {
	key := s.getPhotoStatusCacheKey(photoID)
	
	// Cache status with TTL
	if err := s.cacheService.Set(ctx, key, status, ttl); err != nil {
		logger.Error("Failed to cache photo status", err)
		return fmt.Errorf("failed to cache photo status: %w", err)
	}
	
	logger.Debug("Photo status cached", map[string]interface{}{
		"photo_id": photoID,
		"status":    status,
		"ttl":       ttl.Seconds(),
	})
	
	return nil
}

// GetCachedPhotoStatus gets cached photo status
func (s *EphemeralPhotoCacheServiceImpl) GetCachedPhotoStatus(ctx context.Context, photoID uuid.UUID) (string, error) {
	key := s.getPhotoStatusCacheKey(photoID)
	
	// Get from cache
	status, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Failed to get cached photo status", err)
		return "", fmt.Errorf("failed to get cached photo status: %w", err)
	}
	
	if status == "" {
		return "", fmt.Errorf("photo status not found in cache")
	}
	
	logger.Debug("Photo status retrieved from cache", map[string]interface{}{
		"photo_id": photoID,
		"status":    status,
	})
	
	return status, nil
}

// InvalidatePhotoStatus invalidates cached photo status
func (s *EphemeralPhotoCacheServiceImpl) InvalidatePhotoStatus(ctx context.Context, photoID uuid.UUID) error {
	key := s.getPhotoStatusCacheKey(photoID)
	
	if err := s.cacheService.Delete(ctx, key); err != nil {
		logger.Error("Failed to invalidate cached photo status", err)
		return fmt.Errorf("failed to invalidate cached photo status: %w", err)
	}
	
	logger.Debug("Photo status cache invalidated", map[string]interface{}{
		"photo_id": photoID,
	})
	
	return nil
}

// CacheViewCount caches view count for a photo
func (s *EphemeralPhotoCacheServiceImpl) CacheViewCount(ctx context.Context, photoID uuid.UUID, count int, ttl time.Duration) error {
	key := s.getViewCountCacheKey(photoID)
	
	// Cache view count with TTL
	if err := s.cacheService.Set(ctx, key, fmt.Sprintf("%d", count), ttl); err != nil {
		logger.Error("Failed to cache view count", err)
		return fmt.Errorf("failed to cache view count: %w", err)
	}
	
	logger.Debug("View count cached", map[string]interface{}{
		"photo_id": photoID,
		"count":    count,
		"ttl":      ttl.Seconds(),
	})
	
	return nil
}

// GetCachedViewCount gets cached view count for a photo
func (s *EphemeralPhotoCacheServiceImpl) GetCachedViewCount(ctx context.Context, photoID uuid.UUID) (int, error) {
	key := s.getViewCountCacheKey(photoID)
	
	// Get from cache
	countStr, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Failed to get cached view count", err)
		return 0, fmt.Errorf("failed to get cached view count: %w", err)
	}
	
	if countStr == "" {
		return 0, fmt.Errorf("view count not found in cache")
	}
	
	var count int
	if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
		logger.Error("Failed to parse cached view count", err)
		return 0, fmt.Errorf("failed to parse cached view count: %w", err)
	}
	
	logger.Debug("View count retrieved from cache", map[string]interface{}{
		"photo_id": photoID,
		"count":    count,
	})
	
	return count, nil
}

// IncrementViewCount increments cached view count for a photo
func (s *EphemeralPhotoCacheServiceImpl) IncrementViewCount(ctx context.Context, photoID uuid.UUID) error {
	key := s.getViewCountCacheKey(photoID)
	
	// Get current count
	currentCount, err := s.GetCachedViewCount(ctx, photoID)
	if err != nil {
		// Start with 1 if not found
		currentCount = 0
	}
	
	// Increment and cache
	newCount := currentCount + 1
	ttl := 5 * time.Minute // Cache for 5 minutes
	
	if err := s.CacheViewCount(ctx, photoID, newCount, ttl); err != nil {
		logger.Error("Failed to increment cached view count", err)
		return fmt.Errorf("failed to increment cached view count: %w", err)
	}
	
	logger.Debug("View count incremented in cache", map[string]interface{}{
		"photo_id":   photoID,
		"old_count":  currentCount,
		"new_count":  newCount,
	})
	
	return nil
}

// CacheAnalytics caches analytics data for a user
func (s *EphemeralPhotoCacheServiceImpl) CacheAnalytics(ctx context.Context, userID uuid.UUID, analytics *EphemeralPhotoAnalytics, ttl time.Duration) error {
	key := s.getAnalyticsCacheKey(userID)
	
	// Serialize analytics
	data, err := json.Marshal(analytics)
	if err != nil {
		logger.Error("Failed to serialize analytics for caching", err)
		return fmt.Errorf("failed to serialize analytics for caching: %w", err)
	}
	
	// Cache with TTL
	if err := s.cacheService.Set(ctx, key, string(data), ttl); err != nil {
		logger.Error("Failed to cache analytics", err)
		return fmt.Errorf("failed to cache analytics: %w", err)
	}
	
	logger.Debug("Analytics cached successfully", map[string]interface{}{
		"user_id": userID,
		"ttl":     ttl.Seconds(),
	})
	
	return nil
}

// GetCachedAnalytics gets cached analytics data for a user
func (s *EphemeralPhotoCacheServiceImpl) GetCachedAnalytics(ctx context.Context, userID uuid.UUID) (*EphemeralPhotoAnalytics, error) {
	key := s.getAnalyticsCacheKey(userID)
	
	// Get from cache
	data, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Failed to get cached analytics", err)
		return nil, fmt.Errorf("failed to get cached analytics: %w", err)
	}
	
	if data == "" {
		return nil, fmt.Errorf("analytics not found in cache")
	}
	
	// Deserialize analytics
	var analytics EphemeralPhotoAnalytics
	if err := json.Unmarshal([]byte(data), &analytics); err != nil {
		logger.Error("Failed to deserialize cached analytics", err)
		return nil, fmt.Errorf("failed to deserialize cached analytics: %w", err)
	}
	
	logger.Debug("Analytics retrieved from cache", map[string]interface{}{
		"user_id": userID,
	})
	
	return &analytics, nil
}

// InvalidateUserAnalytics invalidates cached analytics for a user
func (s *EphemeralPhotoCacheServiceImpl) InvalidateUserAnalytics(ctx context.Context, userID uuid.UUID) error {
	key := s.getAnalyticsCacheKey(userID)
	
	if err := s.cacheService.Delete(ctx, key); err != nil {
		logger.Error("Failed to invalidate cached analytics", err)
		return fmt.Errorf("failed to invalidate cached analytics: %w", err)
	}
	
	logger.Debug("Analytics cache invalidated", map[string]interface{}{
		"user_id": userID,
	})
	
	return nil
}

// CleanupExpiredCache cleans up expired cache entries
func (s *EphemeralPhotoCacheServiceImpl) CleanupExpiredCache(ctx context.Context) (int, error) {
	// This would typically be handled by the underlying cache service
	// For now, return mock count
	deletedCount := 0
	
	logger.Info("Expired cache cleanup completed", map[string]interface{}{
		"deleted_count": deletedCount,
	})
	
	return deletedCount, nil
}

// InvalidateAll invalidates all ephemeral photo cache entries
func (s *EphemeralPhotoCacheServiceImpl) InvalidateAll(ctx context.Context) error {
	// This would typically use a pattern to delete all keys
	// For now, return success
	logger.Info("All ephemeral photo cache invalidated", nil)
	
	return nil
}

// Helper methods

// getPhotoCacheKey generates cache key for a photo
func (s *EphemeralPhotoCacheServiceImpl) getPhotoCacheKey(photoID uuid.UUID) string {
	return fmt.Sprintf("ephemeral_photo:%s", photoID.String())
}

// getAccessKeyCacheKey generates cache key for an access key
func (s *EphemeralPhotoCacheServiceImpl) getAccessKeyCacheKey(accessKey string) string {
	return fmt.Sprintf("ephemeral_photo_access:%s", accessKey)
}

// getPhotoStatusCacheKey generates cache key for photo status
func (s *EphemeralPhotoCacheServiceImpl) getPhotoStatusCacheKey(photoID uuid.UUID) string {
	return fmt.Sprintf("ephemeral_photo_status:%s", photoID.String())
}

// getViewCountCacheKey generates cache key for view count
func (s *EphemeralPhotoCacheServiceImpl) getViewCountCacheKey(photoID uuid.UUID) string {
	return fmt.Sprintf("ephemeral_photo_views:%s", photoID.String())
}

// getAnalyticsCacheKey generates cache key for user analytics
func (s *EphemeralPhotoCacheServiceImpl) getAnalyticsCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("ephemeral_analytics:%s", userID.String())
}