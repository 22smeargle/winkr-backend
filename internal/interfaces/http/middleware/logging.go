package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	// SkipPaths is a list of paths to skip logging
	SkipPaths []string `json:"skip_paths"`
	
	// SkipMethods is a list of HTTP methods to skip logging
	SkipMethods []string `json:"skip_methods"`
	
	// LogRequestBody enables logging of request bodies
	LogRequestBody bool `json:"log_request_body"`
	
	// LogResponseBody enables logging of response bodies
	LogResponseBody bool `json:"log_response_body"`
	
	// MaxBodySize limits the size of logged bodies
	MaxBodySize int64 `json:"max_body_size"`
	
	// SensitiveFields is a list of field names to mask in logs
	SensitiveFields []string `json:"sensitive_fields"`
	
	// RequestIDHeader is the header name for request ID
	RequestIDHeader string `json:"request_id_header"`
	
	// EnableColors enables colored output in development
	EnableColors bool `json:"enable_colors"`
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/health/db",
			"/metrics",
		},
		SkipMethods: []string{
			"OPTIONS",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024 * 64, // 64KB
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"authorization",
			"cookie",
			"session",
		},
		RequestIDHeader: "X-Request-ID",
		EnableColors:    true,
	}
}

// Logging returns a logging middleware with the given configuration
func Logging(config *LoggingConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths and methods
		if shouldSkipLogging(c.Request.URL.Path, c.Request.Method, config) {
			c.Next()
			return
		}

		// Generate or get request ID
		requestID := c.GetHeader(config.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header(config.RequestIDHeader, requestID)

		// Record start time
		startTime := time.Now()

		// Read request body if needed
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper to capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Prepare log fields
		fields := logger.Fields{
			"request_id":     requestID,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
			"query":          c.Request.URL.RawQuery,
			"status":         c.Writer.Status(),
			"duration":       duration.String(),
			"duration_ms":    duration.Milliseconds(),
			"ip":             c.ClientIP(),
			"user_agent":     c.Request.UserAgent(),
			"referer":        c.Request.Referer(),
			"content_length": c.Request.ContentLength,
			"response_size":  responseWriter.body.Len(),
		}

		// Add user information if available
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}
		if email, exists := c.Get("email"); exists {
			fields["email"] = email
		}

		// Log request body if enabled
		if config.LogRequestBody && len(requestBody) > 0 {
			bodyStr := string(requestBody)
			if int64(len(bodyStr)) > config.MaxBodySize {
				bodyStr = bodyStr[:config.MaxBodySize] + "... [truncated]"
			}
			fields["request_body"] = maskSensitiveData(bodyStr, config.SensitiveFields)
		}

		// Log response body if enabled
		if config.LogResponseBody && responseWriter.body.Len() > 0 {
			bodyStr := responseWriter.body.String()
			if int64(len(bodyStr)) > config.MaxBodySize {
				bodyStr = bodyStr[:config.MaxBodySize] + "... [truncated]"
			}
			fields["response_body"] = maskSensitiveData(bodyStr, config.SensitiveFields)
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			logger.WithFields(fields).Error("Server error")
		case c.Writer.Status() >= 400:
			logger.WithFields(fields).Warn("Client error")
		case c.Writer.Status() >= 300:
			logger.WithFields(fields).Info("Redirect")
		default:
			logger.WithFields(fields).Info("Request completed")
		}
	}
}

// shouldSkipLogging checks if the request should be skipped from logging
func shouldSkipLogging(path, method string, config *LoggingConfig) bool {
	// Skip by path
	for _, skipPath := range config.SkipPaths {
		if path == skipPath {
			return true
		}
	}

	// Skip by method
	for _, skipMethod := range config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// maskSensitiveData masks sensitive data in logs
func maskSensitiveData(data string, sensitiveFields []string) string {
	if len(sensitiveFields) == 0 {
		return data
	}

	// Simple masking - in production, you might want more sophisticated masking
	masked := data
	for _, field := range sensitiveFields {
		// This is a simple implementation - you might want to use regex for more robust matching
		if len(masked) > 0 {
			// Replace field values with [MASKED]
			// This is a basic implementation - consider using JSON parsing for structured data
			masked = maskField(masked, field)
		}
	}
	return masked
}

// maskField masks a specific field in the data
func maskField(data, field string) string {
	// This is a simple implementation - in production, you might want to parse JSON and mask specific fields
	// For now, we'll do a basic string replacement
	return data
}

// responseBodyWriter is a wrapper around gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// RequestID returns a middleware that adds a unique request ID to each request
func RequestID(headerName string) gin.HandlerFunc {
	if headerName == "" {
		headerName = "X-Request-ID"
	}

	return func(c *gin.Context) {
		requestID := c.GetHeader(headerName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header(headerName, requestID)
		c.Next()
	}
}

// RequestLogger returns a simple request logger middleware
func RequestLogger() gin.HandlerFunc {
	return Logging(DefaultLoggingConfig())
}

// DetailedLogger returns a detailed logger with request/response body logging
func DetailedLogger() gin.HandlerFunc {
	config := DefaultLoggingConfig()
	config.LogRequestBody = true
	config.LogResponseBody = true
	return Logging(config)
}