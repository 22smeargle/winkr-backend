package handlers

import (
	"context"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/utils/response"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	redisClient   *redis.RedisClient
	sessionMgr    *cache.SessionManager
	cacheService  *cache.CacheService
	rateLimiter   *cache.RateLimiter
	pubSubService  *cache.PubSubService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	redisClient *redis.RedisClient,
	sessionMgr *cache.SessionManager,
	cacheService *cache.CacheService,
	rateLimiter *cache.RateLimiter,
	pubSubService *cache.PubSubService,
) *HealthHandler {
	return &HealthHandler{
		redisClient:   redisClient,
		sessionMgr:    sessionMgr,
		cacheService:  cacheService,
		rateLimiter:   rateLimiter,
		pubSubService:  pubSubService,
	}
}

// CheckRedisHealth checks Redis health
func (h *HealthHandler) CheckRedisHealth() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Basic ping test
	err := h.redisClient.Ping()
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
			"timestamp": time.Now(),
		}
	}

	// Get Redis stats
	stats := h.redisClient.GetStats()
	
	// Check connection pool stats
	poolStats := map[string]interface{}{}
	if poolData, ok := stats["pool_stats"]; ok {
		poolStats = poolData.(map[string]interface{})
	}

	// Get Redis metrics
	metrics := h.redisClient.GetMetrics()
	
	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":   time.Now(),
		"redis": map[string]interface{}{
			"ping":           "ok",
			"stats":          stats,
			"pool_stats":     poolStats,
			"metrics": map[string]interface{}{
				"connections_created":  metrics.ConnectionsCreated,
				"connections_closed":   metrics.ConnectionsClosed,
				"connection_errors":   metrics.ConnectionErrors,
				"commands_executed":   metrics.CommandsExecuted,
				"command_errors":      metrics.CommandErrors,
				"last_connection":     metrics.LastConnectionTime,
				"last_error":          metrics.LastErrorTime,
			},
		},
	}
}

// CheckCacheHealth checks cache service health
func (h *HealthHandler) CheckCacheHealth() map[string]interface{} {
	ctx := context.Background()
	
	// Test cache operations
	testKey := "health_check_test"
	testValue := "test_value"
	
	// Test set/get
	err := h.cacheService.CacheAPIResponse(ctx, testKey, testValue)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "cache set operation failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Test get
	found, err := h.cacheService.GetAPIResponse(ctx, testKey, &struct{}{})
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "cache get operation failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	if !found {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "cache get operation returned not found",
			"timestamp": time.Now(),
		}
	}
	
	// Test invalidation
	err = h.cacheService.InvalidateUserProfile(ctx, "test_user")
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "cache invalidation failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Get cache stats
	cacheStats, err := h.cacheService.GetCacheStats(ctx)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "failed to get cache stats: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":   time.Now(),
		"cache": map[string]interface{}{
			"set_test":      "ok",
			"get_test":      "ok",
			"invalidate_test": "ok",
			"stats":         cacheStats,
		},
	}
}

// CheckSessionHealth checks session manager health
func (h *HealthHandler) CheckSessionHealth() map[string]interface{} {
	ctx := context.Background()
	
	// Test session creation
	testUserID := "health_test_user"
	testToken := "test_token"
	testIP := "127.0.0.1"
	testUserAgent := "Health-Check-Agent/1.0"
	
	deviceInfo := cache.DeviceInfo{
		DeviceID:   "health_check_device",
		DeviceType: "server",
		OS:         "linux",
		Browser:    "health_check",
		Location:   "server",
	}
	
	session, err := h.sessionMgr.CreateSession(ctx, testUserID, testToken, "refresh_token", testIP, testUserAgent, deviceInfo)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "session creation failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Test session retrieval
	retrievedSession, err := h.sessionMgr.GetSession(ctx, session.ID)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "session retrieval failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	if retrievedSession.UserID != testUserID {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "session data mismatch",
			"timestamp": time.Now(),
		}
	}
	
	// Test session deletion
	err = h.sessionMgr.DeleteSession(ctx, session.ID)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "session deletion failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Test online status
	isOnline, err := h.sessionMgr.IsUserOnline(ctx, testUserID)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "online status check failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	onlineUsers, err := h.sessionMgr.GetOnlineUsers(ctx)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "get online users failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	return map[string]interface{}{
		"status":       "healthy",
		"timestamp":     time.Now(),
		"sessions": map[string]interface{}{
			"create_test":      "ok",
			"retrieve_test":    "ok",
			"delete_test":      "ok",
			"online_status":    isOnline,
			"online_users":     len(onlineUsers),
		},
	}
}

// CheckRateLimitHealth checks rate limiter health
func (h *HealthHandler) CheckRateLimitHealth() map[string]interface{} {
	ctx := context.Background()
	
	// Test rate limiting
	testIP := "127.0.0.1"
	testEndpoint := "test_endpoint"
	
	config := cache.RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: testEndpoint,
		KeyType:  "ip",
	}
	
	result, err := h.rateLimiter.CheckRateLimit(ctx, config, testIP)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "rate limit check failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Test distributed rate limiting
	distributedConfig := cache.RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: testEndpoint,
		KeyType:  "ip",
	}
	
	distributedResult, err := h.rateLimiter.CheckDistributedRateLimit(ctx, distributedConfig, testIP)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "distributed rate limit check failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":   time.Now(),
		"rate_limiting": map[string]interface{}{
			"standard_test": map[string]interface{}{
				"allowed":    result.Allowed,
				"remaining":  result.Remaining,
				"limit":      result.Limit,
				"window":     result.Window.String(),
			},
			"distributed_test": map[string]interface{}{
				"allowed":    distributedResult.Allowed,
				"remaining":  distributedResult.Remaining,
				"limit":      distributedResult.Limit,
				"window":     distributedResult.Window.String(),
			},
		},
	}
}

// CheckPubSubHealth checks Pub/Sub service health
func (h *HealthHandler) CheckPubSubHealth() map[string]interface{} {
	ctx := context.Background()
	
	// Test Pub/Sub subscription
	msgChan, err := h.pubSubService.SubscribeToNotifications(ctx, "health_test_user")
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "Pub/Sub subscription failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Test Pub/Sub publishing
	testMessage := cache.Message{
		Type:      cache.MessageTypeNotification,
		Channel:   "test_channel",
		Data:      map[string]interface{}{"test": "health_check"},
		Timestamp: time.Now(),
	}
	
	err = h.pubSubService.PublishNotification(ctx, "health_test_user", "test", "Health Check Test", testMessage.Data)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "Pub/Sub publish failed: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Get Pub/Sub stats
	stats, err := h.pubSubService.GetActiveSubscriptions(ctx)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"error":    "failed to get Pub/Sub stats: " + err.Error(),
			"timestamp": time.Now(),
		}
	}
	
	// Close subscription
	close(msgChan)
	
	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":   time.Now(),
		"pubsub": map[string]interface{}{
			"subscribe_test":    "ok",
			"publish_test":     "ok",
			"stats":           stats,
		},
	}
}

// GetOverallHealth returns overall system health
func (h *HealthHandler) GetOverallHealth() map[string]interface{} {
	redisHealth := h.CheckRedisHealth()
	cacheHealth := h.CheckCacheHealth()
	sessionHealth := h.CheckSessionHealth()
	rateLimitHealth := h.CheckRateLimitHealth()
	pubSubHealth := h.CheckPubSubHealth()
	
	// Determine overall status
	overallStatus := "healthy"
	if redisHealth["status"] != "healthy" ||
		cacheHealth["status"] != "healthy" ||
		sessionHealth["status"] != "healthy" ||
		rateLimitHealth["status"] != "healthy" ||
		pubSubHealth["status"] != "healthy" {
		overallStatus = "degraded"
	}
	
	return map[string]interface{}{
		"status":           overallStatus,
		"timestamp":         time.Now(),
		"uptime":           time.Since(time.Now()).String(), // This would be tracked in a real implementation
		"version":          "1.0.0",
		"components": map[string]interface{}{
			"redis":    redisHealth,
			"cache":    cacheHealth,
			"sessions": sessionHealth,
			"rate_limit": rateLimitHealth,
			"pubsub":    pubSubHealth,
		},
	}
}

// HandleHealthRequest handles health check requests
func (h *HealthHandler) HandleHealthRequest(c *gin.Context) {
	checkType := c.Query("check")
	if checkType == "" {
		checkType = "overall"
	}
	
	var health map[string]interface{}
	var statusCode int
	
	switch checkType {
	case "redis":
		health = h.CheckRedisHealth()
		statusCode = http.StatusOK
	case "cache":
		health = h.CheckCacheHealth()
		statusCode = http.StatusOK
	case "sessions":
		health = h.CheckSessionHealth()
		statusCode = http.StatusOK
	case "rate_limit":
		health = h.CheckRateLimitHealth()
		statusCode = http.StatusOK
	case "pubsub":
		health = h.CheckPubSubHealth()
		statusCode = http.StatusOK
	case "overall":
		health = h.GetOverallHealth()
		if health["status"] == "healthy" {
			statusCode = http.StatusOK
		} else if health["status"] == "degraded" {
			statusCode = http.StatusServiceUnavailable // 503
		} else {
			statusCode = http.StatusInternalServerError // 500
		}
	default:
		health = map[string]interface{}{
			"status":   "error",
			"error":    "invalid check type: " + checkType,
			"timestamp": time.Now(),
		}
		statusCode = http.StatusBadRequest
	}
	
	// Add health check headers
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Health-Check-Cache", "false")
	
	// Set appropriate status code based on health
	if statusCode >= 500 {
		c.JSON(statusCode, health)
	} else {
		c.JSON(statusCode, health)
	}
}

// HandleLivenessProbe handles Kubernetes liveness probe
func (h *HealthHandler) HandleLivenessProbe(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// HandleReadinessProbe handles Kubernetes readiness probe
func (h *HealthHandler) HandleReadinessProbe(c *gin.Context) {
	// Check if all components are ready
	health := h.GetOverallHealth()
	
	if health["status"] == "healthy" {
		c.String(http.StatusOK, "OK")
	} else {
		c.String(http.StatusServiceUnavailable, "Not Ready")
	}
}

// HandleMetrics returns detailed metrics
func (h *HealthHandler) HandleMetrics(c *gin.Context) {
	ctx := context.Background()
	
	// Get detailed metrics from all components
	redisStats := h.redisClient.GetStats()
	cacheStats, _ := h.cacheService.GetCacheStats(ctx)
	redisMetrics := h.redisClient.GetMetrics()
	
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()).String(),
		"redis": map[string]interface{}{
			"stats":  redisStats,
			"metrics": redisMetrics,
		},
		"cache": cacheStats,
		"system": map[string]interface{}{
			"goroutines": strconv.Itoa(runtime.NumGoroutine()),
			"memory": map[string]interface{}{
				"alloc":      runtime.MemStats().Alloc,
				"total_alloc": runtime.MemStats().TotalAlloc,
				"sys":        runtime.MemStats().Sys,
				"heap_alloc": runtime.MemStats().HeapAlloc,
			},
			"gc": map[string]interface{}{
				"num_gc":      runtime.NumGC(),
				"last_gc":     time.Now(),
			},
		},
	}
	
	c.JSON(http.StatusOK, metrics)
}