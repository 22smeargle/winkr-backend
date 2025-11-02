package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MonitoringMiddleware provides monitoring functionality for HTTP requests
type MonitoringMiddleware struct {
	metricsService *services.MetricsService
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(metricsService *services.MetricsService) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		metricsService: metricsService,
	}
}

// RequestTimingMiddleware tracks request timing
func (m *MonitoringMiddleware) RequestTimingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate correlation ID if not present
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
			c.Header("X-Correlation-ID", correlationID)
		}

		// Store correlation ID in context
		c.Set("correlation_id", correlationID)

		// Record start time
		start := time.Now()
		
		// Store start time in context
		c.Set("start_time", start)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Record HTTP metrics
		m.metricsService.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
		)

		// Add timing header
		c.Header("X-Response-Time", fmt.Sprintf("%d", duration.Milliseconds()))
	}
}

// ErrorTrackingMiddleware tracks errors and exceptions
func (m *MonitoringMiddleware) ErrorTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			// Get correlation ID from context
			correlationID, _ := c.Get("correlation_id").(string)
			
			// Log errors with correlation ID
			for _, err := range c.Errors {
				logger.Error("Request error", err, map[string]interface{}{
					"correlation_id": correlationID,
					"method":        c.Request.Method,
					"path":          c.FullPath(),
					"status":        c.Writer.Status(),
					"user_agent":     c.Request.UserAgent(),
					"remote_addr":   c.Request.RemoteAddr,
				})
			}
		}
	}
}

// RequestCountingMiddleware counts requests
func (m *MonitoringMiddleware) RequestCountingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Get correlation ID from context
		correlationID, _ := c.Get("correlation_id").(string)

		// Log request with correlation ID
		logger.Info("Request processed", map[string]interface{}{
			"correlation_id": correlationID,
			"method":        c.Request.Method,
			"path":          c.FullPath(),
			"status":        c.Writer.Status(),
			"user_agent":     c.Request.UserAgent(),
			"remote_addr":   c.Request.RemoteAddr(),
		})
	}
}

// PerformanceProfilingMiddleware profiles request performance
func (m *MonitoringMiddleware) PerformanceProfilingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if profiling is enabled
		if c.Query("profile") == "true" {
			// Get correlation ID
			correlationID, _ := c.Get("correlation_id").(string)
			
			// Start profiling
			start := time.Now()
			
			// Capture request details
			requestDetails := map[string]interface{}{
				"correlation_id": correlationID,
				"method":        c.Request.Method,
				"path":          c.FullPath(),
				"headers":       c.Request.Header,
				"user_agent":     c.Request.UserAgent(),
				"remote_addr":   c.Request.RemoteAddr(),
			}
			
			// Process request
			c.Next()
			
			// Calculate performance metrics
			duration := time.Since(start)
			
			// Log performance data
			logger.Info("Request profile", map[string]interface{}{
				"correlation_id":   correlationID,
				"request_details":  requestDetails,
				"duration_ms":     duration.Milliseconds(),
				"status":          c.Writer.Status(),
				"response_size":    c.Writer.Size(),
			})
		} else {
			// Normal processing
			c.Next()
		}
	}
}

// ResourceUsageTrackingMiddleware tracks resource usage
func (m *MonitoringMiddleware) ResourceUsageTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID
		correlationID, _ := c.Get("correlation_id").(string)
		
		// Capture initial resource state
		initialMemory := m.getCurrentMemoryUsage()
		
		// Process request
		c.Next()
		
		// Calculate resource usage
		finalMemory := m.getCurrentMemoryUsage()
		memoryDelta := finalMemory - initialMemory
		
		// Log resource usage
		logger.Debug("Resource usage", map[string]interface{}{
			"correlation_id":    correlationID,
			"method":           c.Request.Method,
			"path":             c.FullPath(),
			"memory_delta_kb":  memoryDelta / 1024,
			"final_memory_kb":  finalMemory / 1024,
		})
	}
}

// SecurityEventLoggingMiddleware logs security events
func (m *MonitoringMiddleware) SecurityEventLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID
		correlationID, _ := c.Get("correlation_id").(string)
		
		// Check for security events
		securityEvents := m.detectSecurityEvents(c)
		
		// Process request
		c.Next()
		
		// Log security events
		for _, event := range securityEvents {
			logger.Warn("Security event detected", map[string]interface{}{
				"correlation_id": correlationID,
				"event_type":     event.Type,
				"event_details":  event.Details,
				"method":         c.Request.Method,
				"path":           c.FullPath(),
				"remote_addr":    c.Request.RemoteAddr(),
				"user_agent":      c.Request.UserAgent(),
			})
		}
	}
}

// RequestBodyLoggingMiddleware logs request body for debugging
func (m *MonitoringMiddleware) RequestBodyLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID
		correlationID, _ := c.Get("correlation_id").(string)
		
		// Check if body logging is enabled
		if c.Query("log_body") == "true" {
			// Read and restore body
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			// Log request body (be careful with sensitive data)
			logger.Debug("Request body", map[string]interface{}{
				"correlation_id": correlationID,
				"method":        c.Request.Method,
				"path":          c.FullPath(),
				"content_type":  c.GetHeader("Content-Type"),
				"body_size":     len(bodyBytes),
				"body":          string(bodyBytes), // Only for debugging
			})
		}
		
		// Process request
		c.Next()
	}
}

// ResponseSizeTrackingMiddleware tracks response sizes
func (m *MonitoringMiddleware) ResponseSizeTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID
		correlationID, _ := c.Get("correlation_id").(string)
		
		// Wrap response writer to capture size
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			size:          0,
		}
		c.Writer = writer
		
		// Process request
		c.Next()
		
		// Log response size
		logger.Debug("Response size", map[string]interface{}{
			"correlation_id": correlationID,
			"method":        c.Request.Method,
			"path":          c.FullPath(),
			"status":        writer.Status(),
			"size_bytes":    writer.size,
		})
	}
}

// SlowQueryDetectionMiddleware detects slow database queries (placeholder)
func (m *MonitoringMiddleware) SlowQueryDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID
		correlationID, _ := c.Get("correlation_id").(string)
		
		// Process request
		c.Next()
		
		// Check for slow query header (would be set by database middleware)
		if slowQueryTime := c.GetHeader("X-Slow-Query-Time"); slowQueryTime != "" {
			logger.Warn("Slow query detected", map[string]interface{}{
				"correlation_id":    correlationID,
				"method":           c.Request.Method,
				"path":             c.FullPath(),
				"slow_query_time_ms": slowQueryTime,
			})
		}
	}
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	Type    string                 `json:"type"`
	Details map[string]interface{} `json:"details"`
}

// responseWriter wraps gin.ResponseWriter to capture response size
type responseWriter struct {
	gin.ResponseWriter
	size int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// Helper methods

func (m *MonitoringMiddleware) getCurrentMemoryUsage() uint64 {
	// This is a simplified memory usage calculation
	// In production, you'd want more sophisticated memory tracking
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func (m *MonitoringMiddleware) detectSecurityEvents(c *gin.Context) []SecurityEvent {
	var events []SecurityEvent
	
	// Check for suspicious patterns
	userAgent := c.Request.UserAgent()
	remoteAddr := c.Request.RemoteAddr()
	path := c.FullPath()
	method := c.Request.Method()
	
	// Detect potential SQL injection
	if strings.Contains(strings.ToLower(path), "select") ||
		strings.Contains(strings.ToLower(path), "union") ||
		strings.Contains(strings.ToLower(path), "drop") {
		events = append(events, SecurityEvent{
			Type: "potential_sql_injection",
			Details: map[string]interface{}{
				"path": path,
				"method": method,
			},
		})
	}
	
	// Detect suspicious user agents
	if strings.Contains(strings.ToLower(userAgent), "sqlmap") ||
		strings.Contains(strings.ToLower(userAgent), "nmap") ||
		strings.Contains(strings.ToLower(userAgent), "nikto") {
		events = append(events, SecurityEvent{
			Type: "suspicious_user_agent",
			Details: map[string]interface{}{
				"user_agent": userAgent,
				"remote_addr": remoteAddr,
			},
		})
	}
	
	// Detect unusual request methods
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		events = append(events, SecurityEvent{
			Type: "unusual_http_method",
			Details: map[string]interface{}{
				"method": method,
				"path": path,
			},
		})
	}
	
	// Detect path traversal attempts
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		events = append(events, SecurityEvent{
			Type: "path_traversal_attempt",
			Details: map[string]interface{}{
				"path": path,
				"remote_addr": remoteAddr,
			},
		})
	}
	
	return events
}

// CombinedMonitoringMiddleware combines all monitoring middleware
func (m *MonitoringMiddleware) CombinedMonitoringMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Generate correlation ID if not present
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
			c.Header("X-Correlation-ID", correlationID)
		}

		// Store correlation ID in context
		c.Set("correlation_id", correlationID)

		// Record start time
		start := time.Now()
		c.Set("start_time", start)

		// Capture initial state
		initialMemory := m.getCurrentMemoryUsage()
		
		// Detect security events
		securityEvents := m.detectSecurityEvents(c)
		
		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		finalMemory := m.getCurrentMemoryUsage()
		memoryDelta := finalMemory - initialMemory
		
		// Record HTTP metrics
		m.metricsService.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			duration,
		)

		// Log security events
		for _, event := range securityEvents {
			logger.Warn("Security event detected", map[string]interface{}{
				"correlation_id": correlationID,
				"event_type":     event.Type,
				"event_details":  event.Details,
				"method":         c.Request.Method,
				"path":           c.FullPath(),
				"remote_addr":    c.Request.RemoteAddr(),
				"user_agent":      c.Request.UserAgent(),
			})
		}

		// Log errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error", err, map[string]interface{}{
					"correlation_id": correlationID,
					"method":        c.Request.Method,
					"path":          c.FullPath(),
					"status":        c.Writer.Status(),
					"user_agent":     c.Request.UserAgent(),
					"remote_addr":   c.Request.RemoteAddr(),
				})
			}
		}

		// Add monitoring headers
		c.Header("X-Response-Time", strconv.FormatInt(duration.Milliseconds(), 10))
		c.Header("X-Memory-Delta", strconv.FormatInt(int(memoryDelta/1024), 10))
		
		// Log performance data
		logger.Debug("Request completed", map[string]interface{}{
			"correlation_id":   correlationID,
			"method":          c.Request.Method,
			"path":            c.FullPath(),
			"duration_ms":     duration.Milliseconds(),
			"status":          c.Writer.Status(),
			"memory_delta_kb":  memoryDelta / 1024,
			"final_memory_kb":  finalMemory / 1024,
		})
	})
}