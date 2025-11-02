package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/monitoring"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// MockCacheService is a mock implementation of the cache service
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockCacheService) LPush(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCacheService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	args := m.Called(ctx, key, count, value)
	return args.Error(0)
}

func (m *MockCacheService) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	args := m.Called(ctx, key, score, member)
	return args.Error(0)
}

func (m *MockCacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]services.CacheServiceZResult, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]services.CacheServiceZResult), args.Error(1)
}

func (m *MockCacheService) ZRangeByScore(ctx context.Context, key string, min, max float64) ([]services.CacheServiceZResult, error) {
	args := m.Called(ctx, key, min, max)
	return args.Get(0).([]services.CacheServiceZResult), args.Error(1)
}

func (m *MockCacheService) ZRem(ctx context.Context, key string, member interface{}) error {
	args := m.Called(ctx, key, member)
	return args.Error(0)
}

func (m *MockCacheService) ZRemRangeByScore(ctx context.Context, key string, min, max float64) error {
	args := m.Called(ctx, key, min, max)
	return args.Error(0)
}

func (m *MockCacheService) ZCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

// TestHealthCheckEndpoint tests the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
		},
	}

	mockCache := &MockCacheService{}
	healthService := services.NewHealthCheckService(cfg, mockCache)
	getHealthUseCase := monitoring.NewGetHealthStatusUseCase(healthService)
	handler := handlers.NewMonitoringHandler(getHealthUseCase, nil, nil, nil, nil)

	// Test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetHealth(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "uptime")
}

// TestDetailedHealthCheckEndpoint tests the detailed health check endpoint
func TestDetailedHealthCheckEndpoint(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
		},
	}

	mockCache := &MockCacheService{}
	healthService := services.NewHealthCheckService(cfg, mockCache)
	getDetailedHealthUseCase := monitoring.NewGetDetailedHealthUseCase(healthService)
	handler := handlers.NewMonitoringHandler(nil, getDetailedHealthUseCase, nil, nil, nil)

	// Test request
	req := httptest.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetDetailedHealth(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Contains(t, response, "uptime")
	assert.Contains(t, response, "components")
	assert.Contains(t, response, "system")
}

// TestMetricsEndpoint tests the metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	metricsService := services.NewMetricsService(cfg, mockCache)
	getMetricsUseCase := monitoring.NewGetMetricsUseCase(metricsService)
	handler := handlers.NewMonitoringHandler(nil, nil, getMetricsUseCase, nil, nil)

	// Test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetMetrics(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check if it's Prometheus format
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")
}

// TestMetricsJSONEndpoint tests the metrics JSON endpoint
func TestMetricsJSONEndpoint(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	metricsService := services.NewMetricsService(cfg, mockCache)
	getMetricsUseCase := monitoring.NewGetMetricsUseCase(metricsService)
	handler := handlers.NewMonitoringHandler(nil, nil, getMetricsUseCase, nil, nil)

	// Test request
	req := httptest.NewRequest("GET", "/metrics/json", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetMetricsJSON(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "http")
	assert.Contains(t, response, "database")
	assert.Contains(t, response, "cache")
	assert.Contains(t, response, "business")
	assert.Contains(t, response, "system")
}

// TestMonitoringStatusEndpoint tests the monitoring status endpoint
func TestMonitoringStatusEndpoint(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	healthService := services.NewHealthCheckService(cfg, mockCache)
	metricsService := services.NewMetricsService(cfg, mockCache)
	getMonitoringDashboardUseCase := monitoring.NewGetMonitoringDashboardUseCase(healthService, metricsService)
	handler := handlers.NewMonitoringHandler(nil, nil, nil, nil, getMonitoringDashboardUseCase)

	// Test request
	req := httptest.NewRequest("GET", "/monitoring/status", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetMonitoringStatus(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "health")
	assert.Contains(t, response, "metrics")
	assert.Contains(t, response, "system")
	assert.Contains(t, response, "timestamp")
}

// TestHealthCheckService tests the health check service
func TestHealthCheckService(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
		},
	}

	mockCache := &MockCacheService{}
	healthService := services.NewHealthCheckService(cfg, mockCache)

	// Test database health check
	mockCache.On("Ping", mock.Anything).Return(nil)
	dbHealth := healthService.CheckDatabase(context.Background())
	assert.Equal(t, services.HealthStatusHealthy, dbHealth.Status)
	mockCache.AssertExpectations(t)

	// Test Redis health check
	mockCache = &MockCacheService{}
	mockCache.On("Ping", mock.Anything).Return(nil)
	healthService = services.NewHealthCheckService(cfg, mockCache)
	redisHealth := healthService.CheckRedis(context.Background())
	assert.Equal(t, services.HealthStatusHealthy, redisHealth.Status)
	mockCache.AssertExpectations(t)

	// Test overall health
	mockCache = &MockCacheService{}
	mockCache.On("Ping", mock.Anything).Return(nil).Twice()
	healthService = services.NewHealthCheckService(cfg, mockCache)
	overallHealth := healthService.GetOverallHealth(context.Background())
	assert.Equal(t, services.HealthStatusHealthy, overallHealth.Status)
	mockCache.AssertExpectations(t)
}

// TestMetricsService tests the metrics service
func TestMetricsService(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	metricsService := services.NewMetricsService(cfg, mockCache)

	// Test HTTP metrics recording
	metricsService.RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond)
	httpMetrics := metricsService.GetHTTPMetrics()
	assert.Equal(t, int64(1), httpMetrics.TotalRequests)
	assert.Equal(t, int64(1), httpMetrics.RequestsByStatus["200"])
	assert.Equal(t, int64(1), httpMetrics.RequestsByMethod["GET"])

	// Test database metrics recording
	metricsService.RecordDatabaseQuery("SELECT", 50*time.Millisecond, false)
	dbMetrics := metricsService.GetDatabaseMetrics()
	assert.Equal(t, int64(1), dbMetrics.TotalQueries)
	assert.Equal(t, int64(1), dbMetrics.QueriesByType["SELECT"])

	// Test cache metrics recording
	metricsService.RecordCacheOperation("get", true)
	cacheMetrics := metricsService.GetCacheMetrics()
	assert.Equal(t, int64(1), cacheMetrics.TotalOperations)
	assert.Equal(t, int64(1), cacheMetrics.Hits)
	assert.Equal(t, int64(1), cacheMetrics.OperationsByType["get"])

	// Test business metrics recording
	metricsService.RecordUserRegistration()
	businessMetrics := metricsService.GetBusinessMetrics()
	assert.Equal(t, int64(1), businessMetrics.TotalUsers)
	assert.Equal(t, int64(1), businessMetrics.RegistrationsToday)
}

// TestAlertingService tests the alerting service
func TestAlertingService(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Alerting: config.AlertingConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	alertingService := services.NewAlertingService(cfg, mockCache)

	// Test alert rule creation
	rule := &services.AlertRule{
		ID:          "test-rule",
		Name:        "Test Rule",
		Description: "Test alert rule",
		Metric:      "cpu_usage",
		Operator:    ">",
		Threshold:   80.0,
		Severity:    services.AlertSeverityWarning,
		Enabled:     true,
	}

	err := alertingService.CreateAlertRule(rule)
	assert.NoError(t, err)

	// Test alert rule evaluation
	metrics := map[string]interface{}{
		"cpu_usage": 85.0,
		"timestamp": time.Now(),
	}

	err = alertingService.EvaluateRules(context.Background(), metrics)
	assert.NoError(t, err)

	// Test alert retrieval
	alerts, err := alertingService.GetAlerts(context.Background(), services.AlertStatusActive, 10)
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, services.AlertStatusActive, alerts[0].Status)
	assert.Equal(t, services.AlertSeverityWarning, alerts[0].Severity)
}

// TestMonitoringMiddleware tests the monitoring middleware
func TestMonitoringMiddleware(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	metricsService := services.NewMetricsService(cfg, mockCache)
	middleware := services.NewMonitoringMiddleware(metricsService)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap the handler with middleware
	wrappedHandler := middleware.RequestTiming()(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Check if metrics were recorded
	httpMetrics := metricsService.GetHTTPMetrics()
	assert.Equal(t, int64(1), httpMetrics.TotalRequests)
	assert.Equal(t, int64(1), httpMetrics.RequestsByStatus["200"])
	assert.Equal(t, int64(1), httpMetrics.RequestsByMethod["GET"])
}

// TestMonitoringJobsService tests the monitoring jobs service
func TestMonitoringJobsService(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			BackgroundJobs: config.BackgroundJobsConfig{
				HealthCheckJobInterval:        time.Minute,
				MetricsAggregationInterval:    time.Minute * 5,
				LogCleanupInterval:            time.Hour,
				SystemMonitoringInterval:      time.Minute * 2,
				ExternalServiceCheckInterval:  time.Minute * 3,
			},
		},
	}

	mockCache := &MockCacheService{}
	healthService := services.NewHealthCheckService(cfg, mockCache)
	metricsService := services.NewMetricsService(cfg, mockCache)
	alertingService := services.NewAlertingService(cfg, mockCache)
	jobsService := services.NewMonitoringJobsService(cfg, mockCache, healthService, metricsService, alertingService)

	// Test starting jobs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := jobsService.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, jobsService.IsRunning())

	// Wait a short time for jobs to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping jobs
	err = jobsService.Stop()
	assert.NoError(t, err)
	assert.False(t, jobsService.IsRunning())
}

// TestMonitoringStorageService tests the monitoring storage service
func TestMonitoringStorageService(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Storage: config.StorageConfig{
				HealthRetentionPeriod:           time.Hour * 24,
				HTTPMetricsRetentionPeriod:      time.Hour * 24,
				DatabaseMetricsRetentionPeriod:  time.Hour * 24,
				CacheMetricsRetentionPeriod:     time.Hour * 24,
				BusinessMetricsRetentionPeriod:  time.Hour * 24,
				SystemMetricsRetentionPeriod:    time.Hour * 24,
				TimeSeriesRetentionPeriod:       time.Hour * 24 * 7,
				MaxTimeSeriesPoints:            1000,
			},
		},
	}

	mockCache := &MockCacheService{}
	storageService := services.NewMonitoringStorageService(cfg, mockCache)

	// Test storing health status
	health := &services.HealthStatus{
		Status:    services.HealthStatusHealthy,
		Timestamp: time.Now(),
		Uptime:    time.Hour,
	}

	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), health, time.Hour*24).Return(nil)
	mockCache.On("LPush", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	err := storageService.StoreHealthStatus(context.Background(), health)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)

	// Test storing metrics
	metrics := map[string]interface{}{
		"total_requests": 100,
		"error_rate":     0.05,
		"timestamp":      time.Now(),
	}

	mockCache = &MockCacheService{}
	mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), metrics, time.Hour*24).Return(nil)
	mockCache.On("LPush", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	err = storageService.StoreMetrics(context.Background(), "http", metrics)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)

	// Test storing time-series data
	mockCache = &MockCacheService{}
	mockCache.On("ZAdd", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("float64"), mock.AnythingOfType("float64")).Return(nil)
	mockCache.On("Expire", mock.Anything, mock.AnythingOfType("string"), time.Hour*24*7).Return(nil)
	mockCache.On("ZCard", mock.Anything, mock.AnythingOfType("string")).Return(int64(100), nil)

	err = storageService.StoreTimeSeriesData(context.Background(), "cpu_usage", 75.5, time.Now())
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

// TestMonitoringIntegration tests the full monitoring system integration
func TestMonitoringIntegration(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
			Alerting: config.AlertingConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	
	// Create services
	healthService := services.NewHealthCheckService(cfg, mockCache)
	metricsService := services.NewMetricsService(cfg, mockCache)
	alertingService := services.NewAlertingService(cfg, mockCache)
	
	// Create use cases
	getHealthUseCase := monitoring.NewGetHealthStatusUseCase(healthService)
	getDetailedHealthUseCase := monitoring.NewGetDetailedHealthUseCase(healthService)
	getMetricsUseCase := monitoring.NewGetMetricsUseCase(metricsService)
	getSystemStatusUseCase := monitoring.NewGetSystemStatusUseCase(healthService, metricsService)
	getMonitoringDashboardUseCase := monitoring.NewGetMonitoringDashboardUseCase(healthService, metricsService)
	
	// Create handler
	handler := handlers.NewMonitoringHandler(
		getHealthUseCase,
		getDetailedHealthUseCase,
		getMetricsUseCase,
		getSystemStatusUseCase,
		getMonitoringDashboardUseCase,
	)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.GetHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test detailed health endpoint
	req = httptest.NewRequest("GET", "/health/detailed", nil)
	w = httptest.NewRecorder()
	handler.GetDetailedHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test metrics endpoint
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	handler.GetMetrics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test metrics JSON endpoint
	req = httptest.NewRequest("GET", "/metrics/json", nil)
	w = httptest.NewRecorder()
	handler.GetMetricsJSON(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test monitoring status endpoint
	req = httptest.NewRequest("GET", "/monitoring/status", nil)
	w = httptest.NewRecorder()
	handler.GetMonitoringStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test system status endpoint
	req = httptest.NewRequest("GET", "/system/status", nil)
	w = httptest.NewRecorder()
	handler.GetSystemStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestMonitoringErrorHandling tests error handling in the monitoring system
func TestMonitoringErrorHandling(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			HealthCheck: config.HealthCheckConfig{
				Enabled: true,
				Timeout: 5 * time.Second,
			},
		},
	}

	// Test with failing cache
	mockCache := &MockCacheService{}
	mockCache.On("Ping", mock.Anything).Return(fmt.Errorf("cache connection failed"))
	
	healthService := services.NewHealthCheckService(cfg, mockCache)
	
	// Test database health check failure
	dbHealth := healthService.CheckDatabase(context.Background())
	assert.Equal(t, services.HealthStatusUnhealthy, dbHealth.Status)
	assert.NotEmpty(t, dbHealth.Error)
	
	// Test Redis health check failure
	redisHealth := healthService.CheckRedis(context.Background())
	assert.Equal(t, services.HealthStatusUnhealthy, redisHealth.Status)
	assert.NotEmpty(t, redisHealth.Error)
	
	// Test overall health with failures
	overallHealth := healthService.GetOverallHealth(context.Background())
	assert.Equal(t, services.HealthStatusDegraded, overallHealth.Status)
	
	mockCache.AssertExpectations(t)
}

// TestMonitoringPerformance tests the performance of the monitoring system
func TestMonitoringPerformance(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Monitoring: config.MonitoringConfig{
			Metrics: config.MetricsConfig{
				Enabled: true,
			},
		},
	}

	mockCache := &MockCacheService{}
	metricsService := services.NewMetricsService(cfg, mockCache)

	// Test recording a large number of metrics
	start := time.Now()
	for i := 0; i < 1000; i++ {
		metricsService.RecordHTTPRequest("GET", "/test", 200, time.Duration(i)*time.Microsecond)
	}
	duration := time.Since(start)

	// Assert that recording 1000 metrics takes less than 100ms
	assert.Less(t, duration, 100*time.Millisecond)

	// Test metrics retrieval
	start = time.Now()
	httpMetrics := metricsService.GetHTTPMetrics()
	duration = time.Since(start)

	// Assert that retrieving metrics takes less than 10ms
	assert.Less(t, duration, 10*time.Millisecond)
	assert.Equal(t, int64(1000), httpMetrics.TotalRequests)
}