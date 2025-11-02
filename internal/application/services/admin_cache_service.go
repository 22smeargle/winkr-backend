package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AdminCacheService handles caching for admin operations
type AdminCacheService struct {
	cacheService cache.CacheService
}

// NewAdminCacheService creates a new AdminCacheService
func NewAdminCacheService(cacheService cache.CacheService) *AdminCacheService {
	return &AdminCacheService{
		cacheService: cacheService,
	}
}

// Cache keys
const (
	// User cache keys
	UserListKey         = "admin:users:list"
	UserDetailsKey      = "admin:users:details"
	UserStatsKey        = "admin:users:stats"
	
	// Analytics cache keys
	PlatformStatsKey    = "admin:stats:platform"
	UserStatsKey        = "admin:stats:users"
	MatchStatsKey       = "admin:stats:matches"
	MessageStatsKey     = "admin:stats:messages"
	PaymentStatsKey     = "admin:stats:payments"
	VerificationStatsKey = "admin:stats:verification"
	
	// System cache keys
	SystemHealthKey     = "admin:system:health"
	SystemMetricsKey    = "admin:system:metrics"
	SystemConfigKey     = "admin:system:config"
	SystemLogsKey       = "admin:system:logs"
	
	// Content cache keys
	ReportedContentKey  = "admin:content:reported"
	ContentAnalyticsKey = "admin:content:analytics"
	
	// Dashboard cache keys
	DashboardKey        = "admin:dashboard"
	AlertsKey          = "admin:alerts"
)

// Cache TTL durations
const (
	ShortTTL  = 5 * time.Minute   // For frequently changing data
	MediumTTL = 30 * time.Minute  // For moderately changing data
	LongTTL   = 2 * time.Hour      // For relatively stable data
	VeryLongTTL = 24 * time.Hour  // For very stable data
)

// User Caching Methods

// CacheUserList caches user list data
func (s *AdminCacheService) CacheUserList(ctx context.Context, key string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", UserListKey, key)
	return s.cacheWithTTL(ctx, cacheKey, data, ShortTTL)
}

// GetUserList retrieves cached user list data
func (s *AdminCacheService) GetUserList(ctx context.Context, key string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", UserListKey, key)
	return s.get(ctx, cacheKey, dest)
}

// InvalidateUserList invalidates user list cache
func (s *AdminCacheService) InvalidateUserList(ctx context.Context, key string) error {
	cacheKey := fmt.Sprintf("%s:%s", UserListKey, key)
	return s.cacheService.Delete(ctx, cacheKey)
}

// CacheUserDetails caches user details
func (s *AdminCacheService) CacheUserDetails(ctx context.Context, userID uuid.UUID, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", UserDetailsKey, userID.String())
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetUserDetails retrieves cached user details
func (s *AdminCacheService) GetUserDetails(ctx context.Context, userID uuid.UUID, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", UserDetailsKey, userID.String())
	return s.get(ctx, cacheKey, dest)
}

// InvalidateUserDetails invalidates user details cache
func (s *AdminCacheService) InvalidateUserDetails(ctx context.Context, userID uuid.UUID) error {
	cacheKey := fmt.Sprintf("%s:%s", UserDetailsKey, userID.String())
	return s.cacheService.Delete(ctx, cacheKey)
}

// Analytics Caching Methods

// CachePlatformStats caches platform statistics
func (s *AdminCacheService) CachePlatformStats(ctx context.Context, period string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", PlatformStatsKey, period)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetPlatformStats retrieves cached platform statistics
func (s *AdminCacheService) GetPlatformStats(ctx context.Context, period string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", PlatformStatsKey, period)
	return s.get(ctx, cacheKey, dest)
}

// CacheUserStats caches user statistics
func (s *AdminCacheService) CacheUserStats(ctx context.Context, period, groupBy string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", UserStatsKey, period, groupBy)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetUserStats retrieves cached user statistics
func (s *AdminCacheService) GetUserStats(ctx context.Context, period, groupBy string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", UserStatsKey, period, groupBy)
	return s.get(ctx, cacheKey, dest)
}

// CacheMatchStats caches match statistics
func (s *AdminCacheService) CacheMatchStats(ctx context.Context, period, groupBy string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", MatchStatsKey, period, groupBy)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetMatchStats retrieves cached match statistics
func (s *AdminCacheService) GetMatchStats(ctx context.Context, period, groupBy string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", MatchStatsKey, period, groupBy)
	return s.get(ctx, cacheKey, dest)
}

// CacheMessageStats caches message statistics
func (s *AdminCacheService) CacheMessageStats(ctx context.Context, period, groupBy string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", MessageStatsKey, period, groupBy)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetMessageStats retrieves cached message statistics
func (s *AdminCacheService) GetMessageStats(ctx context.Context, period, groupBy string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", MessageStatsKey, period, groupBy)
	return s.get(ctx, cacheKey, dest)
}

// CachePaymentStats caches payment statistics
func (s *AdminCacheService) CachePaymentStats(ctx context.Context, period, groupBy string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", PaymentStatsKey, period, groupBy)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetPaymentStats retrieves cached payment statistics
func (s *AdminCacheService) GetPaymentStats(ctx context.Context, period, groupBy string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", PaymentStatsKey, period, groupBy)
	return s.get(ctx, cacheKey, dest)
}

// CacheVerificationStats caches verification statistics
func (s *AdminCacheService) CacheVerificationStats(ctx context.Context, period, groupBy, verificationType string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", VerificationStatsKey, period, groupBy, verificationType)
	return s.cacheWithTTL(ctx, cacheKey, data, MediumTTL)
}

// GetVerificationStats retrieves cached verification statistics
func (s *AdminCacheService) GetVerificationStats(ctx context.Context, period, groupBy, verificationType string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s:%s", VerificationStatsKey, period, groupBy, verificationType)
	return s.get(ctx, cacheKey, dest)
}

// System Caching Methods

// CacheSystemHealth caches system health status
func (s *AdminCacheService) CacheSystemHealth(ctx context.Context, data interface{}) error {
	return s.cacheWithTTL(ctx, SystemHealthKey, data, ShortTTL)
}

// GetSystemHealth retrieves cached system health status
func (s *AdminCacheService) GetSystemHealth(ctx context.Context, dest interface{}) error {
	return s.get(ctx, SystemHealthKey, dest)
}

// CacheSystemMetrics caches system metrics
func (s *AdminCacheService) CacheSystemMetrics(ctx context.Context, data interface{}) error {
	return s.cacheWithTTL(ctx, SystemMetricsKey, data, ShortTTL)
}

// GetSystemMetrics retrieves cached system metrics
func (s *AdminCacheService) GetSystemMetrics(ctx context.Context, dest interface{}) error {
	return s.get(ctx, SystemMetricsKey, dest)
}

// CacheSystemConfig caches system configuration
func (s *AdminCacheService) CacheSystemConfig(ctx context.Context, section string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", SystemConfigKey, section)
	return s.cacheWithTTL(ctx, cacheKey, data, LongTTL)
}

// GetSystemConfig retrieves cached system configuration
func (s *AdminCacheService) GetSystemConfig(ctx context.Context, section string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", SystemConfigKey, section)
	return s.get(ctx, cacheKey, dest)
}

// InvalidateSystemConfig invalidates system configuration cache
func (s *AdminCacheService) InvalidateSystemConfig(ctx context.Context, section string) error {
	cacheKey := fmt.Sprintf("%s:%s", SystemConfigKey, section)
	return s.cacheService.Delete(ctx, cacheKey)
}

// Content Caching Methods

// CacheReportedContent caches reported content
func (s *AdminCacheService) CacheReportedContent(ctx context.Context, contentType, status string, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", ReportedContentKey, contentType, status)
	return s.cacheWithTTL(ctx, cacheKey, data, ShortTTL)
}

// GetReportedContent retrieves cached reported content
func (s *AdminCacheService) GetReportedContent(ctx context.Context, contentType, status string, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", ReportedContentKey, contentType, status)
	return s.get(ctx, cacheKey, dest)
}

// InvalidateReportedContent invalidates reported content cache
func (s *AdminCacheService) InvalidateReportedContent(ctx context.Context, contentType, status string) error {
	cacheKey := fmt.Sprintf("%s:%s:%s", ReportedContentKey, contentType, status)
	return s.cacheService.Delete(ctx, cacheKey)
}

// Dashboard Caching Methods

// CacheDashboard caches dashboard data
func (s *AdminCacheService) CacheDashboard(ctx context.Context, adminID uuid.UUID, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", DashboardKey, adminID.String())
	return s.cacheWithTTL(ctx, cacheKey, data, ShortTTL)
}

// GetDashboard retrieves cached dashboard data
func (s *AdminCacheService) GetDashboard(ctx context.Context, adminID uuid.UUID, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", DashboardKey, adminID.String())
	return s.get(ctx, cacheKey, dest)
}

// InvalidateDashboard invalidates dashboard cache
func (s *AdminCacheService) InvalidateDashboard(ctx context.Context, adminID uuid.UUID) error {
	cacheKey := fmt.Sprintf("%s:%s", DashboardKey, adminID.String())
	return s.cacheService.Delete(ctx, cacheKey)
}

// CacheAlerts caches admin alerts
func (s *AdminCacheService) CacheAlerts(ctx context.Context, adminID uuid.UUID, data interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", AlertsKey, adminID.String())
	return s.cacheWithTTL(ctx, cacheKey, data, ShortTTL)
}

// GetAlerts retrieves cached admin alerts
func (s *AdminCacheService) GetAlerts(ctx context.Context, adminID uuid.UUID, dest interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", AlertsKey, adminID.String())
	return s.get(ctx, cacheKey, dest)
}

// Bulk Cache Operations

// InvalidateAllUserCache invalidates all user-related cache
func (s *AdminCacheService) InvalidateAllUserCache(ctx context.Context) error {
	patterns := []string{
		fmt.Sprintf("%s:*", UserListKey),
		fmt.Sprintf("%s:*", UserDetailsKey),
		fmt.Sprintf("%s:*", UserStatsKey),
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate cache pattern", err, "pattern", pattern)
			return err
		}
	}

	return nil
}

// InvalidateAllAnalyticsCache invalidates all analytics cache
func (s *AdminCacheService) InvalidateAllAnalyticsCache(ctx context.Context) error {
	patterns := []string{
		fmt.Sprintf("%s:*", PlatformStatsKey),
		fmt.Sprintf("%s:*", UserStatsKey),
		fmt.Sprintf("%s:*", MatchStatsKey),
		fmt.Sprintf("%s:*", MessageStatsKey),
		fmt.Sprintf("%s:*", PaymentStatsKey),
		fmt.Sprintf("%s:*", VerificationStatsKey),
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate cache pattern", err, "pattern", pattern)
			return err
		}
	}

	return nil
}

// InvalidateAllSystemCache invalidates all system-related cache
func (s *AdminCacheService) InvalidateAllSystemCache(ctx context.Context) error {
	patterns := []string{
		fmt.Sprintf("%s:*", SystemHealthKey),
		fmt.Sprintf("%s:*", SystemMetricsKey),
		fmt.Sprintf("%s:*", SystemConfigKey),
		fmt.Sprintf("%s:*", SystemLogsKey),
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate cache pattern", err, "pattern", pattern)
			return err
		}
	}

	return nil
}

// InvalidateAllContentCache invalidates all content-related cache
func (s *AdminCacheService) InvalidateAllContentCache(ctx context.Context) error {
	patterns := []string{
		fmt.Sprintf("%s:*", ReportedContentKey),
		fmt.Sprintf("%s:*", ContentAnalyticsKey),
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate cache pattern", err, "pattern", pattern)
			return err
		}
	}

	return nil
}

// InvalidateAllAdminCache invalidates all admin cache
func (s *AdminCacheService) InvalidateAllAdminCache(ctx context.Context) error {
	patterns := []string{
		"admin:*",
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate cache pattern", err, "pattern", pattern)
			return err
		}
	}

	return nil
}

// Helper Methods

// cacheWithTTL caches data with specified TTL
func (s *AdminCacheService) cacheWithTTL(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal data for caching", err, "key", key)
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := s.cacheService.Set(ctx, key, string(jsonData), ttl); err != nil {
		logger.Error("Failed to cache data", err, "key", key)
		return fmt.Errorf("failed to cache data: %w", err)
	}

	logger.Debug("Data cached successfully", "key", key, "ttl", ttl)
	return nil
}

// get retrieves cached data
func (s *AdminCacheService) get(ctx context.Context, key string, dest interface{}) error {
	data, err := s.cacheService.Get(ctx, key)
	if err != nil {
		logger.Debug("Cache miss", "key", key, "error", err)
		return fmt.Errorf("cache miss: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		logger.Error("Failed to unmarshal cached data", err, "key", key)
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	logger.Debug("Cache hit", "key", key)
	return nil
}

// GetCacheStats returns cache statistics
func (s *AdminCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Mock implementation - in real implementation, this would get actual cache stats
	return map[string]interface{}{
		"total_keys": 1000,
		"memory_usage": "50MB",
		"hit_rate": 0.85,
		"miss_rate": 0.15,
		"evictions": 10,
		"connections": 5,
	}, nil
}

// WarmupCache preloads frequently accessed data into cache
func (s *AdminCacheService) WarmupCache(ctx context.Context) error {
	logger.Info("Starting admin cache warmup")

	// In a real implementation, this would preload frequently accessed data
	// For now, we'll just log the warmup attempt

	logger.Info("Admin cache warmup completed")
	return nil
}