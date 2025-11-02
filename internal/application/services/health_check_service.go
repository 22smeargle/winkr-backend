package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Status          HealthStatus          `json:"status"`
	Timestamp       time.Time            `json:"timestamp"`
	ResponseTime    time.Duration        `json:"response_time"`
	Error           string               `json:"error,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

// SystemHealth represents overall system health
type SystemHealth struct {
	Status       HealthStatus                   `json:"status"`
	Timestamp     time.Time                     `json:"timestamp"`
	Uptime        time.Duration                 `json:"uptime"`
	Version       string                        `json:"version"`
	Components    map[string]HealthCheck         `json:"components"`
	System        SystemResourceHealth           `json:"system"`
	Dependencies  map[string]DependencyHealth    `json:"dependencies"`
}

// SystemResourceHealth represents system resource health
type SystemResourceHealth struct {
	CPU    ResourceHealth `json:"cpu"`
	Memory ResourceHealth `json:"memory"`
	Disk   ResourceHealth `json:"disk"`
}

// ResourceHealth represents resource health
type ResourceHealth struct {
	Usage    float64 `json:"usage"`
	Status    HealthStatus `json:"status"`
	Threshold float64 `json:"threshold"`
}

// DependencyHealth represents external dependency health
type DependencyHealth struct {
	Status       HealthStatus `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string       `json:"error,omitempty"`
}

// HealthCheckService provides health checking functionality
type HealthCheckService struct {
	config         *config.Config
	db             *postgres.Database
	redisClient    *cache.CacheService
	storageService  storage.StorageService
	stripeService  *stripe.StripeService
	startTime      time.Time
	mu             sync.RWMutex
	lastChecks     map[string]HealthCheck
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(
	cfg *config.Config,
	db *postgres.Database,
	redisClient *cache.CacheService,
	storageService storage.StorageService,
	stripeService *stripe.StripeService,
) *HealthCheckService {
	return &HealthCheckService{
		config:        cfg,
		db:            db,
		redisClient:   redisClient,
		storageService: storageService,
		stripeService: stripeService,
		startTime:     time.Now(),
		lastChecks:    make(map[string]HealthCheck),
	}
}

// CheckDatabaseHealth checks database connectivity and performance
func (h *HealthCheckService) CheckDatabaseHealth(ctx context.Context) HealthCheck {
	start := time.Now()
	
	// Check database connectivity
	err := h.db.Health()
	responseTime := time.Since(start)
	
	if err != nil {
		logger.Error("Database health check failed", err)
		return HealthCheck{
			Status:       HealthStatusUnhealthy,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}
	
	// Get database stats
	stats := h.db.GetStats()
	
	// Check if response time is within threshold
	threshold := time.Duration(h.config.Monitoring.HealthCheck.DatabaseThreshold) * time.Millisecond
	status := HealthStatusHealthy
	if responseTime > threshold {
		status = HealthStatusDegraded
	}
	
	return HealthCheck{
		Status:       status,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"stats":     stats,
			"threshold":  threshold.Milliseconds(),
		},
	}
}

// CheckRedisHealth checks Redis connectivity and performance
func (h *HealthCheckService) CheckRedisHealth(ctx context.Context) HealthCheck {
	start := time.Now()
	
	// Test Redis connectivity
	testKey := "health_check_test"
	testValue := "test_value_" + strconv.FormatInt(time.Now().Unix(), 10)
	
	err := h.redisClient.Set(ctx, testKey, testValue, time.Minute)
	responseTime := time.Since(start)
	
	if err != nil {
		logger.Error("Redis health check failed", err)
		return HealthCheck{
			Status:       HealthStatusUnhealthy,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}
	
	// Test Redis get operation
	var retrievedValue string
	err = h.redisClient.Get(ctx, testKey, &retrievedValue)
	if err != nil || retrievedValue != testValue {
		return HealthCheck{
			Status:       HealthStatusUnhealthy,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        "Redis get/set operation failed",
		}
	}
	
	// Clean up test key
	h.redisClient.Delete(ctx, testKey)
	
	// Check if response time is within threshold
	threshold := time.Duration(h.config.Monitoring.HealthCheck.RedisThreshold) * time.Millisecond
	status := HealthStatusHealthy
	if responseTime > threshold {
		status = HealthStatusDegraded
	}
	
	return HealthCheck{
		Status:       status,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"threshold": threshold.Milliseconds(),
		},
	}
}

// CheckStorageHealth checks storage service connectivity and performance
func (h *HealthCheckService) CheckStorageHealth(ctx context.Context) HealthCheck {
	start := time.Now()
	
	// Test storage connectivity
	testKey := "health_check_test_" + strconv.FormatInt(time.Now().Unix(), 10)
	testContent := "health check test content"
	
	// Test upload
	_, err := h.storageService.UploadFile(ctx, nil, testKey, "text/plain")
	responseTime := time.Since(start)
	
	if err != nil {
		logger.Error("Storage health check failed", err)
		return HealthCheck{
			Status:       HealthStatusUnhealthy,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}
	
	// Test delete
	err = h.storageService.DeleteFile(ctx, testKey)
	if err != nil {
		return HealthCheck{
			Status:       HealthStatusDegraded,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        "Storage delete operation failed: " + err.Error(),
		}
	}
	
	// Check if response time is within threshold
	threshold := time.Duration(h.config.Monitoring.HealthCheck.StorageThreshold) * time.Millisecond
	status := HealthStatusHealthy
	if responseTime > threshold {
		status = HealthStatusDegraded
	}
	
	return HealthCheck{
		Status:       status,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"threshold": threshold.Milliseconds(),
		},
	}
}

// CheckStripeHealth checks Stripe service connectivity
func (h *HealthCheckService) CheckStripeHealth(ctx context.Context) HealthCheck {
	if !h.config.Monitoring.HealthCheck.ExternalEnabled {
		return HealthCheck{
			Status:    HealthStatusHealthy,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"message": "External service checks disabled",
			},
		}
	}
	
	start := time.Now()
	
	// Test Stripe API connectivity
	// This is a simple health check - in production, you might want to check account balance or make a minimal API call
	_, err := h.stripeService.GetAccountInfo(ctx)
	responseTime := time.Since(start)
	
	if err != nil {
		logger.Error("Stripe health check failed", err)
		return HealthCheck{
			Status:       HealthStatusDegraded,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			Error:        err.Error(),
		}
	}
	
	return HealthCheck{
		Status:       HealthStatusHealthy,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
	}
}

// CheckSystemResources checks system resource usage
func (h *HealthCheckService) CheckSystemResources(ctx context.Context) SystemResourceHealth {
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate CPU usage (simplified - in production, you might want more sophisticated CPU monitoring)
	cpuUsage := h.getCPUUsage()
	
	// Calculate memory usage percentage
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100
	
	// Get disk usage (simplified - in production, you might want more sophisticated disk monitoring)
	diskUsage := h.getDiskUsage()
	
	// Determine health status based on thresholds
	cpuStatus := HealthStatusHealthy
	if cpuUsage > h.config.Monitoring.HealthCheck.CPUThreshold {
		cpuStatus = HealthStatusDegraded
		if cpuUsage > 95 {
			cpuStatus = HealthStatusUnhealthy
		}
	}
	
	memoryStatus := HealthStatusHealthy
	if memoryUsage > h.config.Monitoring.HealthCheck.MemoryThreshold {
		memoryStatus = HealthStatusDegraded
		if memoryUsage > 95 {
			memoryStatus = HealthStatusUnhealthy
		}
	}
	
	diskStatus := HealthStatusHealthy
	if diskUsage > h.config.Monitoring.HealthCheck.DiskThreshold {
		diskStatus = HealthStatusDegraded
		if diskUsage > 98 {
			diskStatus = HealthStatusUnhealthy
		}
	}
	
	return SystemResourceHealth{
		CPU: ResourceHealth{
			Usage:    cpuUsage,
			Status:    cpuStatus,
			Threshold: h.config.Monitoring.HealthCheck.CPUThreshold,
		},
		Memory: ResourceHealth{
			Usage:    memoryUsage,
			Status:    memoryStatus,
			Threshold: h.config.Monitoring.HealthCheck.MemoryThreshold,
		},
		Disk: ResourceHealth{
			Usage:    diskUsage,
			Status:    diskStatus,
			Threshold: h.config.Monitoring.HealthCheck.DiskThreshold,
		},
	}
}

// GetOverallHealth returns overall system health
func (h *HealthCheckService) GetOverallHealth(ctx context.Context) SystemHealth {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	components := make(map[string]HealthCheck)
	dependencies := make(map[string]DependencyHealth)
	
	// Check all enabled components
	if h.config.Monitoring.HealthCheck.DatabaseEnabled {
		components["database"] = h.CheckDatabaseHealth(ctx)
	}
	
	if h.config.Monitoring.HealthCheck.RedisEnabled {
		components["redis"] = h.CheckRedisHealth(ctx)
	}
	
	if h.config.Monitoring.HealthCheck.StorageEnabled {
		components["storage"] = h.CheckStorageHealth(ctx)
	}
	
	// Check external dependencies
	if h.config.Monitoring.HealthCheck.ExternalEnabled {
		stripeHealth := h.CheckStripeHealth(ctx)
		dependencies["stripe"] = DependencyHealth{
			Status:       stripeHealth.Status,
			ResponseTime: stripeHealth.ResponseTime,
			Error:        stripeHealth.Error,
		}
	}
	
	// Check system resources
	systemHealth := h.CheckSystemResources(ctx)
	
	// Determine overall status
	overallStatus := HealthStatusHealthy
	for _, component := range components {
		if component.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if component.Status == HealthStatusDegraded {
			overallStatus = HealthStatusDegraded
		}
	}
	
	// Check system resources
	if systemHealth.CPU.Status == HealthStatusUnhealthy ||
		systemHealth.Memory.Status == HealthStatusUnhealthy ||
		systemHealth.Disk.Status == HealthStatusUnhealthy {
		overallStatus = HealthStatusUnhealthy
	} else if systemHealth.CPU.Status == HealthStatusDegraded ||
		systemHealth.Memory.Status == HealthStatusDegraded ||
		systemHealth.Disk.Status == HealthStatusDegraded {
		overallStatus = HealthStatusDegraded
	}
	
	// Check dependencies
	for _, dep := range dependencies {
		if dep.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusDegraded // External dependencies don't make system unhealthy, just degraded
		}
	}
	
	// Cache the results
	h.lastChecks = components
	
	return SystemHealth{
		Status:      overallStatus,
		Timestamp:    time.Now(),
		Uptime:       time.Since(h.startTime),
		Version:      "1.0.0", // This should come from build info
		Components:   components,
		System:       systemHealth,
		Dependencies: dependencies,
	}
}

// GetBasicHealth returns basic health status for load balancers
func (h *HealthCheckService) GetBasicHealth(ctx context.Context) HealthCheck {
	// Quick health check for load balancers
	start := time.Now()
	
	// Check database connectivity
	if h.config.Monitoring.HealthCheck.DatabaseEnabled {
		if err := h.db.Health(); err != nil {
			return HealthCheck{
				Status:       HealthStatusUnhealthy,
				Timestamp:    time.Now(),
				ResponseTime: time.Since(start),
				Error:        "Database connectivity failed",
			}
		}
	}
	
	// Check Redis connectivity
	if h.config.Monitoring.HealthCheck.RedisEnabled {
		testKey := "basic_health_check"
		err := h.redisClient.Set(ctx, testKey, "ok", time.Second*10)
		if err != nil {
			return HealthCheck{
				Status:       HealthStatusUnhealthy,
				Timestamp:    time.Now(),
				ResponseTime: time.Since(start),
				Error:        "Redis connectivity failed",
			}
		}
		h.redisClient.Delete(ctx, testKey)
	}
	
	return HealthCheck{
		Status:       HealthStatusHealthy,
		Timestamp:    time.Now(),
		ResponseTime: time.Since(start),
	}
}

// GetLastChecks returns the last cached health check results
func (h *HealthCheckService) GetLastChecks() map[string]HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	result := make(map[string]HealthCheck)
	for k, v := range h.lastChecks {
		result[k] = v
	}
	return result
}

// Helper methods

func (h *HealthCheckService) getCPUUsage() float64 {
	// This is a simplified CPU usage calculation
	// In production, you might want to use more sophisticated methods
	// like reading from /proc/stat on Linux or using system libraries
	return float64(runtime.NumGoroutine()) / 1000.0 * 100 // Simplified calculation
}

func (h *HealthCheckService) getDiskUsage() float64 {
	// This is a simplified disk usage calculation
	// In production, you would want to read actual disk usage
	// from the filesystem
	return 50.0 // Placeholder - 50% disk usage
}