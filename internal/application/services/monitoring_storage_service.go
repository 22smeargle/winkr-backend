package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MonitoringStorageService provides storage for monitoring data
type MonitoringStorageService struct {
	config       *config.Config
	cacheService *cache.CacheService
}

// NewMonitoringStorageService creates a new monitoring storage service
func NewMonitoringStorageService(
	cfg *config.Config,
	cacheService *cache.CacheService,
) *MonitoringStorageService {
	return &MonitoringStorageService{
		config:       cfg,
		cacheService: cacheService,
	}
}

// StoreHealthStatus stores health status data
func (m *MonitoringStorageService) StoreHealthStatus(ctx context.Context, health *HealthStatus) error {
	key := fmt.Sprintf("health:history:%d", time.Now().Unix())
	
	// Store with TTL based on retention policy
	ttl := m.config.Monitoring.Storage.HealthRetentionPeriod
	
	if err := m.cacheService.Set(ctx, key, health, ttl); err != nil {
		logger.Error("Failed to store health status", err)
		return fmt.Errorf("failed to store health status: %w", err)
	}
	
	// Update index for time-based queries
	if err := m.updateHealthIndex(ctx, key); err != nil {
		logger.Error("Failed to update health index", err)
		// Don't fail the operation, just log the error
	}
	
	return nil
}

// GetHealthHistory retrieves health status history
func (m *MonitoringStorageService) GetHealthHistory(ctx context.Context, from, to time.Time) ([]*HealthStatus, error) {
	// Get health status keys from index
	keys, err := m.getHealthKeys(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get health keys: %w", err)
	}
	
	// Retrieve health status data
	var healthHistory []*HealthStatus
	for _, key := range keys {
		var health HealthStatus
		if err := m.cacheService.Get(ctx, key, &health); err != nil {
			logger.Warn("Failed to retrieve health status", map[string]interface{}{
				"key": key,
				"err": err,
			})
			continue
		}
		healthHistory = append(healthHistory, &health)
	}
	
	return healthHistory, nil
}

// StoreMetrics stores metrics data
func (m *MonitoringStorageService) StoreMetrics(ctx context.Context, metricsType string, metrics interface{}) error {
	timestamp := time.Now().Unix()
	key := fmt.Sprintf("metrics:%s:%d", metricsType, timestamp)
	
	// Store with TTL based on retention policy
	var ttl time.Duration
	switch metricsType {
	case "http":
		ttl = m.config.Monitoring.Storage.HTTPMetricsRetentionPeriod
	case "database":
		ttl = m.config.Monitoring.Storage.DatabaseMetricsRetentionPeriod
	case "cache":
		ttl = m.config.Monitoring.Storage.CacheMetricsRetentionPeriod
	case "business":
		ttl = m.config.Monitoring.Storage.BusinessMetricsRetentionPeriod
	case "system":
		ttl = m.config.Monitoring.Storage.SystemMetricsRetentionPeriod
	default:
		ttl = time.Hour * 24 // Default to 24 hours
	}
	
	if err := m.cacheService.Set(ctx, key, metrics, ttl); err != nil {
		logger.Error("Failed to store metrics", err)
		return fmt.Errorf("failed to store metrics: %w", err)
	}
	
	// Update index for time-based queries
	if err := m.updateMetricsIndex(ctx, metricsType, key); err != nil {
		logger.Error("Failed to update metrics index", err)
		// Don't fail the operation, just log the error
	}
	
	return nil
}

// GetMetricsHistory retrieves metrics history
func (m *MonitoringStorageService) GetMetricsHistory(ctx context.Context, metricsType string, from, to time.Time) ([]map[string]interface{}, error) {
	// Get metrics keys from index
	keys, err := m.getMetricsKeys(ctx, metricsType, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics keys: %w", err)
	}
	
	// Retrieve metrics data
	var metricsHistory []map[string]interface{}
	for _, key := range keys {
		var metrics map[string]interface{}
		if err := m.cacheService.Get(ctx, key, &metrics); err != nil {
			logger.Warn("Failed to retrieve metrics", map[string]interface{}{
				"key": key,
				"err": err,
			})
			continue
		}
		metricsHistory = append(metricsHistory, metrics)
	}
	
	return metricsHistory, nil
}

// StoreAlert stores alert data
func (m *MonitoringStorageService) StoreAlert(ctx context.Context, alert *Alert) error {
	key := fmt.Sprintf("alerts:%s", alert.ID)
	
	// Store with longer TTL for alerts
	ttl := time.Hour * 24 * 30 // 30 days
	
	if err := m.cacheService.Set(ctx, key, alert, ttl); err != nil {
		logger.Error("Failed to store alert", err)
		return fmt.Errorf("failed to store alert: %w", err)
	}
	
	// Update alert index
	if err := m.updateAlertIndex(ctx, alert); err != nil {
		logger.Error("Failed to update alert index", err)
		// Don't fail the operation, just log the error
	}
	
	return nil
}

// GetAlerts retrieves alerts
func (m *MonitoringStorageService) GetAlerts(ctx context.Context, status string, limit int) ([]*Alert, error) {
	// Get alert keys from index
	keys, err := m.getAlertKeys(ctx, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert keys: %w", err)
	}
	
	// Retrieve alert data
	var alerts []*Alert
	for _, key := range keys {
		var alert Alert
		if err := m.cacheService.Get(ctx, key, &alert); err != nil {
			logger.Warn("Failed to retrieve alert", map[string]interface{}{
				"key": key,
				"err": err,
			})
			continue
		}
		alerts = append(alerts, &alert)
	}
	
	return alerts, nil
}

// StoreTimeSeriesData stores time-series data
func (m *MonitoringStorageService) StoreTimeSeriesData(ctx context.Context, metric string, value float64, timestamp time.Time) error {
	// Use Redis sorted sets for time-series data
	key := fmt.Sprintf("timeseries:%s", metric)
	score := float64(timestamp.Unix())
	
	// Store in sorted set
	if err := m.cacheService.ZAdd(ctx, key, score, value); err != nil {
		logger.Error("Failed to store time-series data", err)
		return fmt.Errorf("failed to store time-series data: %w", err)
	}
	
	// Set TTL for the sorted set
	ttl := m.config.Monitoring.Storage.TimeSeriesRetentionPeriod
	if err := m.cacheService.Expire(ctx, key, ttl); err != nil {
		logger.Warn("Failed to set TTL for time-series data", err)
	}
	
	// Clean up old data points
	if err := m.cleanupTimeSeriesData(ctx, key); err != nil {
		logger.Warn("Failed to cleanup time-series data", err)
	}
	
	return nil
}

// GetTimeSeriesData retrieves time-series data
func (m *MonitoringStorageService) GetTimeSeriesData(ctx context.Context, metric string, from, to time.Time) ([]TimeSeriesPoint, error) {
	key := fmt.Sprintf("timeseries:%s", metric)
	min := float64(from.Unix())
	max := float64(to.Unix())
	
	// Get data from sorted set
	results, err := m.cacheService.ZRangeByScore(ctx, key, min, max)
	if err != nil {
		return nil, fmt.Errorf("failed to get time-series data: %w", err)
	}
	
	// Convert to time-series points
	var points []TimeSeriesPoint
	for _, result := range results {
		timestamp := time.Unix(int64(result.Score), 0)
		value := result.Member.(float64)
		points = append(points, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	
	return points, nil
}

// AggregateTimeSeriesData aggregates time-series data
func (m *MonitoringStorageService) AggregateTimeSeriesData(ctx context.Context, metric string, from, to time.Time, interval time.Duration) ([]TimeSeriesPoint, error) {
	// Get raw time-series data
	points, err := m.GetTimeSeriesData(ctx, metric, from, to)
	if err != nil {
		return nil, err
	}
	
	// Aggregate by interval
	aggregated := make(map[int64][]float64)
	for _, point := range points {
		bucket := point.Timestamp.Unix() / int64(interval.Seconds())
		aggregated[bucket] = append(aggregated[bucket], point.Value)
	}
	
	// Calculate averages for each bucket
	var result []TimeSeriesPoint
	for bucket, values := range aggregated {
		timestamp := time.Unix(bucket*int64(interval.Seconds()), 0)
		
		// Calculate average
		var sum float64
		for _, v := range values {
			sum += v
		}
		avg := sum / float64(len(values))
		
		result = append(result, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     avg,
		})
	}
	
	return result, nil
}

// CleanupOldData removes old monitoring data
func (m *MonitoringStorageService) CleanupOldData(ctx context.Context) error {
	logger.Info("Starting cleanup of old monitoring data")
	
	// Cleanup old health status data
	if err := m.cleanupOldHealthData(ctx); err != nil {
		logger.Error("Failed to cleanup old health data", err)
	}
	
	// Cleanup old metrics data
	if err := m.cleanupOldMetricsData(ctx); err != nil {
		logger.Error("Failed to cleanup old metrics data", err)
	}
	
	// Cleanup old alerts
	if err := m.cleanupOldAlerts(ctx); err != nil {
		logger.Error("Failed to cleanup old alerts", err)
	}
	
	// Cleanup old time-series data
	if err := m.cleanupOldTimeSeriesData(ctx); err != nil {
		logger.Error("Failed to cleanup old time-series data", err)
	}
	
	logger.Info("Completed cleanup of old monitoring data")
	return nil
}

// Helper methods

func (m *MonitoringStorageService) updateHealthIndex(ctx context.Context, key string) error {
	indexKey := "health:index"
	return m.cacheService.LPush(ctx, indexKey, key)
}

func (m *MonitoringStorageService) getHealthKeys(ctx context.Context, from, to time.Time) ([]string, error) {
	indexKey := "health:index"
	
	// Get all keys from index
	keys, err := m.cacheService.LRange(ctx, indexKey, 0, -1)
	if err != nil {
		return nil, err
	}
	
	// Filter by timestamp
	var filteredKeys []string
	for _, key := range keys {
		// Extract timestamp from key
		timestampStr := key[len("health:history:"):]
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}
		
		keyTime := time.Unix(timestamp, 0)
		if keyTime.After(from) && keyTime.Before(to) {
			filteredKeys = append(filteredKeys, key)
		}
	}
	
	return filteredKeys, nil
}

func (m *MonitoringStorageService) updateMetricsIndex(ctx context.Context, metricsType, key string) error {
	indexKey := fmt.Sprintf("metrics:index:%s", metricsType)
	return m.cacheService.LPush(ctx, indexKey, key)
}

func (m *MonitoringStorageService) getMetricsKeys(ctx context.Context, metricsType string, from, to time.Time) ([]string, error) {
	indexKey := fmt.Sprintf("metrics:index:%s", metricsType)
	
	// Get all keys from index
	keys, err := m.cacheService.LRange(ctx, indexKey, 0, -1)
	if err != nil {
		return nil, err
	}
	
	// Filter by timestamp
	var filteredKeys []string
	for _, key := range keys {
		// Extract timestamp from key
		parts := fmt.Sprintf("%s", key)
		timestampStr := parts[len(fmt.Sprintf("metrics:%s:", metricsType)):]
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}
		
		keyTime := time.Unix(timestamp, 0)
		if keyTime.After(from) && keyTime.Before(to) {
			filteredKeys = append(filteredKeys, key)
		}
	}
	
	return filteredKeys, nil
}

func (m *MonitoringStorageService) updateAlertIndex(ctx context.Context, alert *Alert) error {
	// Update status index
	statusIndexKey := fmt.Sprintf("alerts:index:status:%s", alert.Status)
	if err := m.cacheService.LPush(ctx, statusIndexKey, alert.ID); err != nil {
		return err
	}
	
	// Update severity index
	severityIndexKey := fmt.Sprintf("alerts:index:severity:%s", alert.Severity)
	if err := m.cacheService.LPush(ctx, severityIndexKey, alert.ID); err != nil {
		return err
	}
	
	// Update time index
	timeIndexKey := "alerts:index:time"
	if err := m.cacheService.LPush(ctx, timeIndexKey, alert.ID); err != nil {
		return err
	}
	
	return nil
}

func (m *MonitoringStorageService) getAlertKeys(ctx context.Context, status string, limit int) ([]string, error) {
	indexKey := fmt.Sprintf("alerts:index:status:%s", status)
	
	// Get keys from index with limit
	keys, err := m.cacheService.LRange(ctx, indexKey, 0, int64(limit-1))
	if err != nil {
		return nil, err
	}
	
	// Convert to alert keys
	var alertKeys []string
	for _, id := range keys {
		alertKeys = append(alertKeys, fmt.Sprintf("alerts:%s", id))
	}
	
	return alertKeys, nil
}

func (m *MonitoringStorageService) cleanupTimeSeriesData(ctx context.Context, key string) error {
	// Get the current number of data points
	count, err := m.cacheService.ZCard(ctx, key)
	if err != nil {
		return err
	}
	
	// If we have more than the maximum number of points, remove the oldest
	maxPoints := m.config.Monitoring.Storage.MaxTimeSeriesPoints
	if count > maxPoints {
		// Get the oldest points to remove
		toRemove := count - maxPoints
		results, err := m.cacheService.ZRange(ctx, key, 0, int64(toRemove-1))
		if err != nil {
			return err
		}
		
		// Remove the oldest points
		for _, result := range results {
			if err := m.cacheService.ZRem(ctx, key, result.Member); err != nil {
				logger.Warn("Failed to remove old time-series data point", err)
			}
		}
	}
	
	return nil
}

func (m *MonitoringStorageService) cleanupOldHealthData(ctx context.Context) error {
	// Get health index keys
	cutoff := time.Now().Add(-m.config.Monitoring.Storage.HealthRetentionPeriod)
	
	// This is a simplified cleanup - in production you'd want to use
	// more efficient methods like Redis SCAN with pattern matching
	keys, err := m.getHealthKeys(ctx, time.Time{}, cutoff)
	if err != nil {
		return err
	}
	
	// Delete old health data
	for _, key := range keys {
		if err := m.cacheService.Del(ctx, key); err != nil {
			logger.Warn("Failed to delete old health data", map[string]interface{}{
				"key": key,
				"err": err,
			})
		}
	}
	
	return nil
}

func (m *MonitoringStorageService) cleanupOldMetricsData(ctx context.Context) error {
	metricsTypes := []string{"http", "database", "cache", "business", "system"}
	
	for _, metricsType := range metricsTypes {
		var retentionPeriod time.Duration
		switch metricsType {
		case "http":
			retentionPeriod = m.config.Monitoring.Storage.HTTPMetricsRetentionPeriod
		case "database":
			retentionPeriod = m.config.Monitoring.Storage.DatabaseMetricsRetentionPeriod
		case "cache":
			retentionPeriod = m.config.Monitoring.Storage.CacheMetricsRetentionPeriod
		case "business":
			retentionPeriod = m.config.Monitoring.Storage.BusinessMetricsRetentionPeriod
		case "system":
			retentionPeriod = m.config.Monitoring.Storage.SystemMetricsRetentionPeriod
		}
		
		cutoff := time.Now().Add(-retentionPeriod)
		keys, err := m.getMetricsKeys(ctx, metricsType, time.Time{}, cutoff)
		if err != nil {
			logger.Warn("Failed to get old metrics keys", map[string]interface{}{
				"type": metricsType,
				"err":  err,
			})
			continue
		}
		
		// Delete old metrics data
		for _, key := range keys {
			if err := m.cacheService.Del(ctx, key); err != nil {
				logger.Warn("Failed to delete old metrics data", map[string]interface{}{
					"key": key,
					"err": err,
				})
			}
		}
	}
	
	return nil
}

func (m *MonitoringStorageService) cleanupOldAlerts(ctx context.Context) error {
	// Get alert index
	indexKey := "alerts:index:time"
	
	// Get all alert IDs
	ids, err := m.cacheService.LRange(ctx, indexKey, 0, -1)
	if err != nil {
		return err
	}
	
	// Check each alert
	cutoff := time.Now().Add(-time.Hour * 24 * 30) // 30 days
	for _, id := range ids {
		key := fmt.Sprintf("alerts:%s", id)
		
		var alert Alert
		if err := m.cacheService.Get(ctx, key, &alert); err != nil {
			// Alert doesn't exist or is corrupted, remove from index
			m.cacheService.LRem(ctx, indexKey, 0, id)
			continue
		}
		
		// If alert is old and resolved, delete it
		if alert.CreatedAt.Before(cutoff) && alert.Status == AlertStatusResolved {
			if err := m.cacheService.Del(ctx, key); err != nil {
				logger.Warn("Failed to delete old resolved alert", map[string]interface{}{
					"id":  id,
					"err": err,
				})
			}
			
			// Remove from indexes
			statusIndexKey := fmt.Sprintf("alerts:index:status:%s", alert.Status)
			m.cacheService.LRem(ctx, statusIndexKey, 0, id)
			
			severityIndexKey := fmt.Sprintf("alerts:index:severity:%s", alert.Severity)
			m.cacheService.LRem(ctx, severityIndexKey, 0, id)
			
			m.cacheService.LRem(ctx, indexKey, 0, id)
		}
	}
	
	return nil
}

func (m *MonitoringStorageService) cleanupOldTimeSeriesData(ctx context.Context) error {
	// Get all time-series keys
	// This is a simplified approach - in production you'd want to use
	// Redis SCAN with pattern matching to find all timeseries:* keys
	
	metricsTypes := []string{
		"http_requests_total",
		"http_response_time",
		"db_query_time",
		"cache_hit_ratio",
		"cpu_usage",
		"memory_usage",
		"disk_usage",
	}
	
	for _, metric := range metricsTypes {
		key := fmt.Sprintf("timeseries:%s", metric)
		
		// Remove old data points
		cutoff := time.Now().Add(-m.config.Monitoring.Storage.TimeSeriesRetentionPeriod)
		minScore := 0.0
		maxScore := float64(cutoff.Unix())
		
		if err := m.cacheService.ZRemRangeByScore(ctx, key, minScore, maxScore); err != nil {
			logger.Warn("Failed to remove old time-series data", map[string]interface{}{
				"metric": metric,
				"err":    err,
			})
		}
	}
	
	return nil
}

// TimeSeriesPoint represents a single time-series data point
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// CacheServiceZMember represents a member in a sorted set
type CacheServiceZMember struct {
	Score  float64       `json:"score"`
	Member interface{}   `json:"member"`
}

// CacheServiceZResult represents a result from a sorted set operation
type CacheServiceZResult struct {
	Score  float64 `json:"score"`
	Member string  `json:"member"`
}

// Extend cache service to support sorted sets (these would be implemented in the actual cache service)
func (c *cache.CacheService) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	// This would be implemented in the actual cache service
	// For now, just return nil as a placeholder
	return nil
}

func (c *cache.CacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]CacheServiceZResult, error) {
	// This would be implemented in the actual cache service
	// For now, just return nil as a placeholder
	return nil, nil
}

func (c *cache.CacheService) ZRangeByScore(ctx context.Context, key string, min, max float64) ([]CacheServiceZResult, error) {
	// This would be implemented in the actual cache service
	// For now, just return nil as a placeholder
	return nil, nil
}

func (c *cache.CacheService) ZRem(ctx context.Context, key string, member interface{}) error {
	// This would be implemented in the actual cache service
	// For now, just return nil as a placeholder
	return nil
}

func (c *cache.CacheService) ZRemRangeByScore(ctx context.Context, key string, min, max float64) error {
	// This would be implemented in the actual cache service
	// For now, just return nil as a placeholder
	return nil
}

func (c *cache.CacheService) ZCard(ctx context.Context, key string) (int64, error) {
	// This would be implemented in the actual cache service
	// For now, just return 0 as a placeholder
	return 0, nil
}