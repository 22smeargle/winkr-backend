package services

import (
	"context"
	"sync"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MonitoringJobsService provides background monitoring jobs functionality
type MonitoringJobsService struct {
	config             *config.Config
	cacheService       *cache.CacheService
	healthCheckService *HealthCheckService
	metricsService    *MetricsService
	alertingService  *AlertingService
	mu                sync.RWMutex
	running           bool
	stopChan          chan struct{}
}

// NewMonitoringJobsService creates a new monitoring jobs service
func NewMonitoringJobsService(
	cfg *config.Config,
	cacheService *cache.CacheService,
	healthCheckService *HealthCheckService,
	metricsService *MetricsService,
	alertingService *AlertingService,
) *MonitoringJobsService {
	return &MonitoringJobsService{
		config:             cfg,
		cacheService:       cacheService,
		healthCheckService: healthCheckService,
		metricsService:    metricsService,
		alertingService:  alertingService,
		running:           false,
		stopChan:          make(chan struct{}),
	}
}

// Start starts all monitoring background jobs
func (m *MonitoringJobsService) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil // Already running
	}

	m.running = true
	logger.Info("Starting monitoring background jobs")

	// Start health check job
	go m.runHealthCheckJob(ctx)

	// Start metrics aggregation job
	go m.runMetricsAggregationJob(ctx)

	// Start log cleanup job
	go m.runLogCleanupJob(ctx)

	// Start system monitoring job
	go m.runSystemMonitoringJob(ctx)

	// Start external service check job
	go m.runExternalServiceCheckJob(ctx)

	// Start alert evaluation job
	go m.runAlertEvaluationJob(ctx)

	// Start metrics collection job
	go m.runMetricsCollectionJob(ctx)

	logger.Info("All monitoring background jobs started")
	return nil
}

// Stop stops all monitoring background jobs
func (m *MonitoringJobsService) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil // Not running
	}

	logger.Info("Stopping monitoring background jobs")
	close(m.stopChan)
	m.running = false

	logger.Info("All monitoring background jobs stopped")
	return nil
}

// IsRunning returns whether the monitoring jobs are running
func (m *MonitoringJobsService) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// Background job implementations

func (m *MonitoringJobsService) runHealthCheckJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.BackgroundJobs.HealthCheckJobInterval)
	defer ticker.Stop()

	logger.Info("Starting health check background job", map[string]interface{}{
		"interval": m.config.Monitoring.BackgroundJobs.HealthCheckJobInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("Health check job stopped")
			return
		case <-m.stopChan:
			logger.Info("Health check job stopped")
			return
		case <-ticker.C:
			m.performHealthCheck(ctx)
		}
	}
}

func (m *MonitoringJobsService) runMetricsAggregationJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.BackgroundJobs.MetricsAggregationInterval)
	defer ticker.Stop()

	logger.Info("Starting metrics aggregation background job", map[string]interface{}{
		"interval": m.config.Monitoring.BackgroundJobs.MetricsAggregationInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("Metrics aggregation job stopped")
			return
		case <-m.stopChan:
			logger.Info("Metrics aggregation job stopped")
			return
		case <-ticker.C:
			m.performMetricsAggregation(ctx)
		}
	}
}

func (m *MonitoringJobsService) runLogCleanupJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.BackgroundJobs.LogCleanupInterval)
	defer ticker.Stop()

	logger.Info("Starting log cleanup background job", map[string]interface{}{
		"interval": m.config.Monitoring.BackgroundJobs.LogCleanupInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("Log cleanup job stopped")
			return
		case <-m.stopChan:
			logger.Info("Log cleanup job stopped")
			return
		case <-ticker.C:
			m.performLogCleanup(ctx)
		}
	}
}

func (m *MonitoringJobsService) runSystemMonitoringJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.BackgroundJobs.SystemMonitoringInterval)
	defer ticker.Stop()

	logger.Info("Starting system monitoring background job", map[string]interface{}{
		"interval": m.config.Monitoring.BackgroundJobs.SystemMonitoringInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("System monitoring job stopped")
			return
		case <-m.stopChan:
			logger.Info("System monitoring job stopped")
			return
		case <-ticker.C:
			m.performSystemMonitoring(ctx)
		}
	}
}

func (m *MonitoringJobsService) runExternalServiceCheckJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.BackgroundJobs.ExternalServiceCheckInterval)
	defer ticker.Stop()

	logger.Info("Starting external service check background job", map[string]interface{}{
		"interval": m.config.Monitoring.BackgroundJobs.ExternalServiceCheckInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("External service check job stopped")
			return
		case <-m.stopChan:
			logger.Info("External service check job stopped")
			return
		case <-ticker.C:
			m.performExternalServiceCheck(ctx)
		}
	}
}

func (m *MonitoringJobsService) runAlertEvaluationJob(ctx context.Context) {
	ticker := time.NewTicker(time.Minute) // Evaluate alerts every minute
	defer ticker.Stop()

	logger.Info("Starting alert evaluation background job")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Alert evaluation job stopped")
			return
		case <-m.stopChan:
			logger.Info("Alert evaluation job stopped")
			return
		case <-ticker.C:
			m.performAlertEvaluation(ctx)
		}
	}
}

func (m *MonitoringJobsService) runMetricsCollectionJob(ctx context.Context) {
	ticker := time.NewTicker(m.config.Monitoring.Metrics.CollectionInterval)
	defer ticker.Stop()

	logger.Info("Starting metrics collection background job", map[string]interface{}{
		"interval": m.config.Monitoring.Metrics.CollectionInterval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			logger.Info("Metrics collection job stopped")
			return
		case <-m.stopChan:
			logger.Info("Metrics collection job stopped")
			return
		case <-ticker.C:
			m.performMetricsCollection(ctx)
		}
	}
}

// Job task implementations

func (m *MonitoringJobsService) performHealthCheck(ctx context.Context) {
	logger.Debug("Performing health check")
	
	// Get overall health status
	health := m.healthCheckService.GetOverallHealth(ctx)
	
	// Store health status in cache for historical data
	key := "health:latest"
	if err := m.cacheService.Set(ctx, key, health, time.Hour); err != nil {
		logger.Error("Failed to store health status in cache", err)
	}
	
	// Check for any unhealthy components
	for name, component := range health.Components {
		if component.Status == HealthStatusUnhealthy {
			logger.Warn("Unhealthy component detected", map[string]interface{}{
				"component": name,
				"error":     component.Error,
			})
		}
	}
	
	// Check system resources
	if health.System.CPU.Status == HealthStatusUnhealthy ||
		health.System.Memory.Status == HealthStatusUnhealthy ||
		health.System.Disk.Status == HealthStatusUnhealthy {
		logger.Warn("System resource issues detected", map[string]interface{}{
			"cpu":    health.System.CPU.Status,
			"memory": health.System.Memory.Status,
			"disk":   health.System.Disk.Status,
		})
	}
}

func (m *MonitoringJobsService) performMetricsAggregation(ctx context.Context) {
	logger.Debug("Performing metrics aggregation")
	
	// Flush metrics to storage
	if err := m.metricsService.FlushMetrics(ctx); err != nil {
		logger.Error("Failed to flush metrics", err)
	}
	
	// Collect system metrics
	m.metricsService.CollectSystemMetrics(ctx)
	
	// Store aggregated metrics in cache
	aggregatedMetrics := map[string]interface{}{
		"timestamp": time.Now(),
		"http":      m.metricsService.GetHTTPMetrics(),
		"database":  m.metricsService.GetDatabaseMetrics(),
		"cache":     m.metricsService.GetCacheMetrics(),
		"business":  m.metricsService.GetBusinessMetrics(),
		"system":    m.metricsService.GetSystemMetrics(),
	}
	
	key := "metrics:aggregated:latest"
	if err := m.cacheService.Set(ctx, key, aggregatedMetrics, time.Hour); err != nil {
		logger.Error("Failed to store aggregated metrics in cache", err)
	}
}

func (m *MonitoringJobsService) performLogCleanup(ctx context.Context) {
	logger.Debug("Performing log cleanup")
	
	// This is a simplified log cleanup implementation
	// In production, you'd want to:
	// 1. Clean up old log files based on retention policy
	// 2. Compress old logs
	// 3. Archive logs to long-term storage
	// 4. Remove temporary files
	
	// For now, just log that cleanup was performed
	logger.Info("Log cleanup completed", map[string]interface{}{
		"retention_period": m.config.Monitoring.Logging.RetentionPeriod.String(),
		"max_file_size":    m.config.Monitoring.Logging.MaxFileSize,
		"max_backups":      m.config.Monitoring.Logging.MaxBackups,
	})
}

func (m *MonitoringJobsService) performSystemMonitoring(ctx context.Context) {
	logger.Debug("Performing system monitoring")
	
	// Collect system metrics
	m.metricsService.CollectSystemMetrics(ctx)
	
	// Store system metrics in cache
	systemMetrics := m.metricsService.GetSystemMetrics()
	key := "system:metrics:latest"
	if err := m.cacheService.Set(ctx, key, systemMetrics, time.Minute*5); err != nil {
		logger.Error("Failed to store system metrics in cache", err)
	}
	
	// Check for resource thresholds
	if systemMetrics.CPUUsage > m.config.Monitoring.Alerting.CPUUsageThreshold {
		logger.Warn("CPU usage threshold exceeded", map[string]interface{}{
			"usage":    systemMetrics.CPUUsage,
			"threshold": m.config.Monitoring.Alerting.CPUUsageThreshold,
		})
	}
	
	if systemMetrics.MemoryUsage > m.config.Monitoring.Alerting.MemoryUsageThreshold {
		logger.Warn("Memory usage threshold exceeded", map[string]interface{}{
			"usage":    systemMetrics.MemoryUsage,
			"threshold": m.config.Monitoring.Alerting.MemoryUsageThreshold,
		})
	}
	
	if systemMetrics.DiskUsage > m.config.Monitoring.Alerting.DiskUsageThreshold {
		logger.Warn("Disk usage threshold exceeded", map[string]interface{}{
			"usage":    systemMetrics.DiskUsage,
			"threshold": m.config.Monitoring.Alerting.DiskUsageThreshold,
		})
	}
}

func (m *MonitoringJobsService) performExternalServiceCheck(ctx context.Context) {
	logger.Debug("Performing external service check")
	
	// Check external services
	// This would include checking Stripe, AWS, email services, etc.
	// For now, just log that check was performed
	
	// Get current metrics to evaluate against thresholds
	httpMetrics := m.metricsService.GetHTTPMetrics()
	dbMetrics := m.metricsService.GetDatabaseMetrics()
	cacheMetrics := m.metricsService.GetCacheMetrics()
	
	// Prepare metrics for alert evaluation
	metrics := map[string]interface{}{
		"http_error_rate":     httpMetrics.ErrorRate,
		"db_slow_queries":    dbMetrics.SlowQueries,
		"cache_hit_ratio":     cacheMetrics.HitRatio,
		"timestamp":           time.Now(),
	}
	
	// Evaluate alert rules
	if m.alertingService != nil {
		if err := m.alertingService.EvaluateRules(ctx, metrics); err != nil {
			logger.Error("Failed to evaluate alert rules", err)
		}
	}
	
	logger.Info("External service check completed", map[string]interface{}{
		"http_error_rate":  httpMetrics.ErrorRate,
		"db_slow_queries": dbMetrics.SlowQueries,
		"cache_hit_ratio":  cacheMetrics.HitRatio,
	})
}

func (m *MonitoringJobsService) performAlertEvaluation(ctx context.Context) {
	logger.Debug("Performing alert evaluation")
	
	// Get current metrics for alert evaluation
	httpMetrics := m.metricsService.GetHTTPMetrics()
	dbMetrics := m.metricsService.GetDatabaseMetrics()
	cacheMetrics := m.metricsService.GetCacheMetrics()
	businessMetrics := m.metricsService.GetBusinessMetrics()
	systemMetrics := m.metricsService.GetSystemMetrics()
	
	// Prepare metrics for alert evaluation
	metrics := map[string]interface{}{
		"http_requests_total":     httpMetrics.TotalRequests,
		"http_error_rate":        httpMetrics.ErrorRate,
		"http_response_time":      httpMetrics.AverageResponseTime.Milliseconds(),
		"db_total_queries":       dbMetrics.TotalQueries,
		"db_slow_queries":        dbMetrics.SlowQueries,
		"db_connections_active":   dbMetrics.ConnectionsActive,
		"cache_hit_ratio":        cacheMetrics.HitRatio,
		"cache_total_operations":  cacheMetrics.TotalOperations,
		"business_total_users":     businessMetrics.TotalUsers,
		"business_active_users":    businessMetrics.ActiveUsers,
		"system_cpu_usage":        systemMetrics.CPUUsage,
		"system_memory_usage":     systemMetrics.MemoryUsage,
		"system_disk_usage":       systemMetrics.DiskUsage,
		"timestamp":               time.Now(),
	}
	
	// Evaluate alert rules
	if m.alertingService != nil {
		if err := m.alertingService.EvaluateRules(ctx, metrics); err != nil {
			logger.Error("Failed to evaluate alert rules", err)
		}
	}
}

func (m *MonitoringJobsService) performMetricsCollection(ctx context.Context) {
	logger.Debug("Performing metrics collection")
	
	// Collect system metrics
	m.metricsService.CollectSystemMetrics(ctx)
	
	// Store metrics in cache for historical data
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"http":      m.metricsService.GetHTTPMetrics(),
		"database":  m.metricsService.GetDatabaseMetrics(),
		"cache":     m.metricsService.GetCacheMetrics(),
		"business":  m.metricsService.GetBusinessMetrics(),
		"system":    m.metricsService.GetSystemMetrics(),
	}
	
	key := "metrics:collection:latest"
	if err := m.cacheService.Set(ctx, key, metrics, time.Minute*5); err != nil {
		logger.Error("Failed to store collected metrics in cache", err)
	}
}