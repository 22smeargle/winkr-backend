package services

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a collected metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Timestamp   time.Time             `json:"timestamp"`
	Help        string                `json:"help,omitempty"`
}

// HTTPMetrics represents HTTP request metrics
type HTTPMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	RequestsByMethod   map[string]int64 `json:"requests_by_method"`
	RequestsByStatus   map[string]int64 `json:"requests_by_status"`
	RequestsByPath     map[string]int64 `json:"requests_by_path"`
	ResponseTimes      []time.Duration `json:"response_times"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	ErrorRate          float64        `json:"error_rate"`
	LastRequestTime    time.Time      `json:"last_request_time"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	TotalQueries        int64         `json:"total_queries"`
	QueriesByType      map[string]int64 `json:"queries_by_type"`
	AverageQueryTime   time.Duration   `json:"average_query_time"`
	SlowQueries        int64          `json:"slow_queries"`
	ConnectionsActive   int64          `json:"connections_active"`
	ConnectionsIdle     int64          `json:"connections_idle"`
	ConnectionsTotal    int64          `json:"connections_total"`
	LastQueryTime      time.Time      `json:"last_query_time"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	TotalOperations     int64         `json:"total_operations"`
	Hits               int64         `json:"hits"`
	Misses             int64         `json:"misses"`
	HitRatio           float64        `json:"hit_ratio"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	OperationsByType   map[string]int64 `json:"operations_by_type"`
	LastOperationTime  time.Time      `json:"last_operation_time"`
}

// BusinessMetrics represents business metrics
type BusinessMetrics struct {
	TotalUsers         int64     `json:"total_users"`
	ActiveUsers       int64     `json:"active_users"`
	NewRegistrations   int64     `json:"new_registrations"`
	TotalMatches       int64     `json:"total_matches"`
	TotalMessages      int64     `json:"total_messages"`
	TotalSwipes       int64     `json:"total_swipes"`
	PremiumSubscriptions int64    `json:"premium_subscriptions"`
	PhotoUploads      int64     `json:"photo_uploads"`
	VerificationRequests int64    `json:"verification_requests"`
	LastActivityTime  time.Time `json:"last_activity_time"`
}

// SystemMetrics represents system resource metrics
type SystemMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	MemoryTotal    uint64   `json:"memory_total"`
	MemoryAlloc    uint64   `json:"memory_alloc"`
	Goroutines     int      `json:"goroutines"`
	GCCycles       uint32   `json:"gc_cycles"`
	DiskUsage      float64 `json:"disk_usage"`
	NetworkIO      NetworkIOMetrics `json:"network_io"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// NetworkIOMetrics represents network I/O metrics
type NetworkIOMetrics struct {
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

// MetricsService provides metrics collection functionality
type MetricsService struct {
	config         *config.Config
	db             *postgres.Database
	redisClient    *cache.CacheService
	mu             sync.RWMutex
	httpMetrics    HTTPMetrics
	dbMetrics      DatabaseMetrics
	cacheMetrics   CacheMetrics
	businessMetrics BusinessMetrics
	systemMetrics  SystemMetrics
	metricsBuffer  []Metric
	lastFlush      time.Time
}

// NewMetricsService creates a new metrics service
func NewMetricsService(
	cfg *config.Config,
	db *postgres.Database,
	redisClient *cache.CacheService,
) *MetricsService {
	return &MetricsService{
		config:        cfg,
		db:            db,
		redisClient:   redisClient,
		httpMetrics:    HTTPMetrics{
			RequestsByMethod: make(map[string]int64),
			RequestsByStatus: make(map[string]int64),
			RequestsByPath:   make(map[string]int64),
			ResponseTimes:    make([]time.Duration, 0, 1000), // Keep last 1000 response times
		},
		dbMetrics: DatabaseMetrics{
			QueriesByType: make(map[string]int64),
		},
		cacheMetrics: CacheMetrics{
			OperationsByType: make(map[string]int64),
		},
		businessMetrics: BusinessMetrics{},
		systemMetrics:  SystemMetrics{},
		metricsBuffer:  make([]Metric, 0, 1000),
		lastFlush:     time.Now(),
	}
}

// RecordHTTPRequest records an HTTP request metric
func (m *MetricsService) RecordHTTPRequest(method, path string, statusCode int, responseTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.httpMetrics.TotalRequests++
	m.httpMetrics.RequestsByMethod[method]++
	m.httpMetrics.RequestsByStatus[strconv.Itoa(statusCode)]++
	m.httpMetrics.RequestsByPath[path]++
	m.httpMetrics.LastRequestTime = time.Now()

	// Keep only last 1000 response times for average calculation
	if len(m.httpMetrics.ResponseTimes) >= 1000 {
		m.httpMetrics.ResponseTimes = m.httpMetrics.ResponseTimes[1:]
	}
	m.httpMetrics.ResponseTimes = append(m.httpMetrics.ResponseTimes, responseTime)

	// Calculate average response time
	var total time.Duration
	for _, rt := range m.httpMetrics.ResponseTimes {
		total += rt
	}
	m.httpMetrics.AverageResponseTime = total / time.Duration(len(m.httpMetrics.ResponseTimes))

	// Calculate error rate (4xx and 5xx status codes)
	errorCount := int64(0)
	for status, count := range m.httpMetrics.RequestsByStatus {
		if status[0] == '4' || status[0] == '5' {
			errorCount += count
		}
	}
	if m.httpMetrics.TotalRequests > 0 {
		m.httpMetrics.ErrorRate = float64(errorCount) / float64(m.httpMetrics.TotalRequests) * 100
	}

	// Add to metrics buffer for Prometheus
	m.addMetric(Metric{
		Name:  "http_requests_total",
		Type:  MetricTypeCounter,
		Value: 1,
		Labels: map[string]string{
			"method": method,
			"path":   path,
			"status": strconv.Itoa(statusCode),
		},
		Timestamp: time.Now(),
		Help:     "Total number of HTTP requests",
	})

	m.addMetric(Metric{
		Name:  "http_request_duration_seconds",
		Type:  MetricTypeHistogram,
		Value: responseTime.Seconds(),
		Labels: map[string]string{
			"method": method,
			"path":   path,
		},
		Timestamp: time.Now(),
		Help:     "HTTP request duration in seconds",
	})
}

// RecordDatabaseQuery records a database query metric
func (m *MetricsService) RecordDatabaseQuery(queryType string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dbMetrics.TotalQueries++
	m.dbMetrics.QueriesByType[queryType]++
	m.dbMetrics.LastQueryTime = time.Now()

	// Calculate average query time
	// This is a simplified calculation - in production, you might want a more sophisticated approach
	if duration > time.Second {
		m.dbMetrics.SlowQueries++
	}

	// Add to metrics buffer
	m.addMetric(Metric{
		Name:  "database_queries_total",
		Type:  MetricTypeCounter,
		Value: 1,
		Labels: map[string]string{
			"type": queryType,
		},
		Timestamp: time.Now(),
		Help:     "Total number of database queries",
	})

	m.addMetric(Metric{
		Name:  "database_query_duration_seconds",
		Type:  MetricTypeHistogram,
		Value: duration.Seconds(),
		Labels: map[string]string{
			"type": queryType,
		},
		Timestamp: time.Now(),
		Help:     "Database query duration in seconds",
	})
}

// RecordCacheOperation records a cache operation metric
func (m *MetricsService) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheMetrics.TotalOperations++
	m.cacheMetrics.OperationsByType[operation]++
	m.cacheMetrics.LastOperationTime = time.Now()

	if hit {
		m.cacheMetrics.Hits++
	} else {
		m.cacheMetrics.Misses++
	}

	// Calculate hit ratio
	if m.cacheMetrics.TotalOperations > 0 {
		m.cacheMetrics.HitRatio = float64(m.cacheMetrics.Hits) / float64(m.cacheMetrics.TotalOperations) * 100
	}

	// Add to metrics buffer
	m.addMetric(Metric{
		Name:  "cache_operations_total",
		Type:  MetricTypeCounter,
		Value: 1,
		Labels: map[string]string{
			"operation": operation,
			"result":    "hit",
		},
		Timestamp: time.Now(),
		Help:     "Total number of cache operations",
	})

	m.addMetric(Metric{
		Name:  "cache_hit_ratio",
		Type:  MetricTypeGauge,
		Value: m.cacheMetrics.HitRatio,
		Labels: map[string]string{
			"operation": operation,
		},
		Timestamp: time.Now(),
		Help:     "Cache hit ratio percentage",
	})
}

// RecordBusinessMetric records a business metric
func (m *MetricsService) RecordBusinessMetric(metric string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch metric {
	case "total_users":
		m.businessMetrics.TotalUsers = int64(value)
	case "active_users":
		m.businessMetrics.ActiveUsers = int64(value)
	case "new_registrations":
		m.businessMetrics.NewRegistrations++
	case "total_matches":
		m.businessMetrics.TotalMatches = int64(value)
	case "total_messages":
		m.businessMetrics.TotalMessages = int64(value)
	case "total_swipes":
		m.businessMetrics.TotalSwipes = int64(value)
	case "premium_subscriptions":
		m.businessMetrics.PremiumSubscriptions = int64(value)
	case "photo_uploads":
		m.businessMetrics.PhotoUploads++
	case "verification_requests":
		m.businessMetrics.VerificationRequests++
	}

	m.businessMetrics.LastActivityTime = time.Now()

	// Add to metrics buffer
	m.addMetric(Metric{
		Name:      "business_metrics",
		Type:      MetricTypeGauge,
		Value:     value,
		Labels:    map[string]string{"metric": metric},
		Timestamp: time.Now(),
		Help:      "Business metrics",
	})
}

// CollectSystemMetrics collects system resource metrics
func (m *MetricsService) CollectSystemMetrics(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.systemMetrics.MemoryUsage = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	m.systemMetrics.MemoryTotal = memStats.Sys
	m.systemMetrics.MemoryAlloc = memStats.Alloc
	m.systemMetrics.Goroutines = runtime.NumGoroutine()
	m.systemMetrics.GCCycles = memStats.NumGC

	// Get CPU usage (simplified)
	m.systemMetrics.CPUUsage = m.getCPUUsage()

	// Get disk usage (simplified)
	m.systemMetrics.DiskUsage = m.getDiskUsage()

	// Get network I/O (simplified)
	m.systemMetrics.NetworkIO = m.getNetworkIO()

	m.systemMetrics.LastUpdateTime = time.Now()

	// Add to metrics buffer
	m.addMetric(Metric{
		Name:      "system_cpu_usage",
		Type:      MetricTypeGauge,
		Value:     m.systemMetrics.CPUUsage,
		Timestamp: time.Now(),
		Help:      "System CPU usage percentage",
	})

	m.addMetric(Metric{
		Name:      "system_memory_usage",
		Type:      MetricTypeGauge,
		Value:     m.systemMetrics.MemoryUsage,
		Timestamp: time.Now(),
		Help:      "System memory usage percentage",
	})

	m.addMetric(Metric{
		Name:      "system_goroutines",
		Type:      MetricTypeGauge,
		Value:     float64(m.systemMetrics.Goroutines),
		Timestamp: time.Now(),
		Help:      "Number of goroutines",
	})
}

// GetHTTPMetrics returns current HTTP metrics
func (m *MetricsService) GetHTTPMetrics() HTTPMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.httpMetrics
}

// GetDatabaseMetrics returns current database metrics
func (m *MetricsService) GetDatabaseMetrics() DatabaseMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Update connection stats from database
	if m.db != nil {
		stats := m.db.GetStats()
		if openConns, ok := stats["open_connections"].(int64); ok {
			m.dbMetrics.ConnectionsActive = openConns
		}
		if idle, ok := stats["idle"].(int64); ok {
			m.dbMetrics.ConnectionsIdle = idle
		}
	}
	
	return m.dbMetrics
}

// GetCacheMetrics returns current cache metrics
func (m *MetricsService) GetCacheMetrics() CacheMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cacheMetrics
}

// GetBusinessMetrics returns current business metrics
func (m *MetricsService) GetBusinessMetrics() BusinessMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.businessMetrics
}

// GetSystemMetrics returns current system metrics
func (m *MetricsService) GetSystemMetrics() SystemMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.systemMetrics
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (m *MetricsService) GetPrometheusMetrics() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var output string
	
	// Group metrics by name and labels
	metricGroups := make(map[string]map[string][]Metric)
	for _, metric := range m.metricsBuffer {
		if _, exists := metricGroups[metric.Name]; !exists {
			metricGroups[metric.Name] = make(map[string][]Metric)
		}
		
		labelKey := ""
		for k, v := range metric.Labels {
			labelKey += fmt.Sprintf("%s=\"%s\",", k, v)
		}
		if len(labelKey) > 0 {
			labelKey = "{" + labelKey[:len(labelKey)-1] + "}"
		}
		
		metricGroups[metric.Name][labelKey] = append(metricGroups[metric.Name][labelKey], metric)
	}
	
	// Generate Prometheus format
	for name, labelGroups := range metricGroups {
		if len(labelGroups) > 0 {
			// Add help and type
			output += fmt.Sprintf("# HELP %s %s\n", name, labelGroups[labelGroups[0][0]].Help)
			output += fmt.Sprintf("# TYPE %s %s\n", name, labelGroups[labelGroups[0][0]].Type)
			
			for labels, metrics := range labelGroups {
				// Use the latest value for gauges, sum for counters
				var value float64
				if len(metrics) > 0 {
					if metrics[0].Type == MetricTypeCounter {
						for _, m := range metrics {
							value += m.Value
						}
					} else {
						value = metrics[len(metrics)-1].Value
					}
				}
				
				output += fmt.Sprintf("%s%s %f\n", name, labels, value)
			}
		}
	}
	
	return output
}

// FlushMetrics flushes metrics to storage
func (m *MetricsService) FlushMetrics(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Monitoring.Metrics.Enabled {
		return nil
	}

	// Check if it's time to flush
	if time.Since(m.lastFlush) < m.config.Monitoring.Metrics.CollectionInterval {
		return nil
	}

	// Store metrics in Redis for time-series data
	if m.redisClient != nil {
		for _, metric := range m.metricsBuffer {
			key := fmt.Sprintf("metrics:%s", metric.Name)
			value := map[string]interface{}{
				"value":     metric.Value,
				"labels":    metric.Labels,
				"timestamp": metric.Timestamp.Unix(),
			}
			
			// Store with TTL based on retention period
			ttl := m.config.Monitoring.Metrics.RetentionPeriod
			if err := m.redisClient.Set(ctx, key, value, ttl); err != nil {
				logger.Error("Failed to store metric in Redis", err)
			}
		}
	}

	// Clear buffer
	m.metricsBuffer = m.metricsBuffer[:0]
	m.lastFlush = time.Now()

	return nil
}

// CleanupOldMetrics removes old metrics data
func (m *MetricsService) CleanupOldMetrics(ctx context.Context) error {
	if !m.config.Monitoring.Metrics.Enabled {
		return nil
	}

	// This is a simplified cleanup - in production, you might want more sophisticated cleanup
	// based on the retention policy and storage backend
	logger.Info("Cleaning up old metrics data")
	
	return nil
}

// Helper methods

func (m *MetricsService) addMetric(metric Metric) {
	m.metricsBuffer = append(m.metricsBuffer, metric)
	
	// Keep buffer size manageable
	if len(m.metricsBuffer) >= 1000 {
		m.metricsBuffer = m.metricsBuffer[500:] // Keep last 500
	}
}

func (m *MetricsService) getCPUUsage() float64 {
	// This is a simplified CPU usage calculation
	// In production, you might want to use more sophisticated methods
	return float64(runtime.NumGoroutine()) / 1000.0 * 100
}

func (m *MetricsService) getDiskUsage() float64 {
	// This is a simplified disk usage calculation
	// In production, you would want to read actual disk usage
	return 50.0 // Placeholder - 50% disk usage
}

func (m *MetricsService) getNetworkIO() NetworkIOMetrics {
	// This is a simplified network I/O calculation
	// In production, you would want to read actual network stats
	return NetworkIOMetrics{
		BytesSent: 1024 * 1024, // 1MB
		BytesRecv: 2 * 1024 * 1024, // 2MB
	}
}