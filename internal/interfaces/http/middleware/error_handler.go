package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ErrorHandlerConfig represents error handling configuration
type ErrorHandlerConfig struct {
	// EnableStackTrace enables stack trace logging
	EnableStackTrace bool `json:"enable_stack_trace"`
	
	// EnablePanicRecovery enables panic recovery
	EnablePanicRecovery bool `json:"enable_panic_recovery"`
	
	// LogErrors enables error logging
	LogErrors bool `json:"log_errors"`
	
	// LogRequests enables request logging on errors
	LogRequests bool `json:"log_requests"`
	
	// CustomHandlers allows custom error handling for specific error types
	CustomHandlers map[error]gin.HandlerFunc `json:"-"`
	
	// ErrorHandler is a custom error handler function
	ErrorHandler func(*gin.Context, error) `json:"-"`
	
	// SkipPaths is a list of paths to skip error handling
	SkipPaths []string `json:"skip_paths"`
	
	// DefaultErrorMessage is the default error message for internal errors
	DefaultErrorMessage string `json:"default_error_message"`
	
	// IncludeRequestID includes request ID in error responses
	IncludeRequestID bool `json:"include_request_id"`
	
	// RequestIDHeader is the header name for request ID
	RequestIDHeader string `json:"request_id_header"`
}

// DefaultErrorHandlerConfig returns a default error handling configuration
func DefaultErrorHandlerConfig() *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		EnableStackTrace:    true,
		EnablePanicRecovery: true,
		LogErrors:          true,
		LogRequests:        true,
		CustomHandlers:      make(map[error]gin.HandlerFunc),
		SkipPaths:         []string{},
		DefaultErrorMessage:  "Internal server error",
		IncludeRequestID:    true,
		RequestIDHeader:     "X-Request-ID",
	}
}

// ErrorHandler returns an error handling middleware with the given configuration
func ErrorHandler(config *ErrorHandlerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	return func(c *gin.Context) {
		// Skip error handling for specified paths
		if shouldSkipErrorHandling(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// Use defer to catch panics
		if config.EnablePanicRecovery {
			defer func() {
				if err := recover(); err != nil {
					handlePanic(c, err, config)
				}
			}()
		}

		// Process request
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			handleErrors(c, config)
		}
	}
}

// handlePanic handles panic recovery
func handlePanic(c *gin.Context, recovered interface{}, config *ErrorHandlerConfig) {
	// Log the panic
	if config.LogErrors {
		fields := logger.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
			"panic":  fmt.Sprintf("%v", recovered),
		}

		if config.IncludeRequestID {
			if requestID := c.GetHeader(config.RequestIDHeader); requestID != "" {
				fields["request_id"] = requestID
			}
		}

		if config.EnableStackTrace {
			fields["stack_trace"] = string(debug.Stack())
		}

		logger.WithFields(fields).Error("Panic recovered")
	}

	// Send error response
	errorResponse := utils.Response{
		Success: false,
		Error: &utils.ErrorInfo{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: config.DefaultErrorMessage,
		},
	}

	// Add request ID if enabled
	if config.IncludeRequestID {
		if requestID := c.GetHeader(config.RequestIDHeader); requestID != "" {
			errorResponse.Error.Details = fmt.Sprintf("Request ID: %s", requestID)
		}
	}

	c.JSON(http.StatusInternalServerError, errorResponse)
	c.Abort()
}

// handleErrors handles errors that occurred during request processing
func handleErrors(c *gin.Context, config *ErrorHandlerConfig) {
	// Get the last error (most recent)
	err := c.Errors.Last().Err

	// Check for custom error handlers
	if config.CustomHandlers != nil {
		if handler, exists := config.CustomHandlers[err]; exists {
			handler(c)
			return
		}
	}

	// Use custom error handler if provided
	if config.ErrorHandler != nil {
		config.ErrorHandler(c, err)
		return
	}

	// Default error handling
	handleDefaultError(c, err, config)
}

// handleDefaultError handles errors using default logic
func handleDefaultError(c *gin.Context, err error, config *ErrorHandlerConfig) {
	// Log the error
	if config.LogErrors {
		logError(c, err, config)
	}

	// Check if it's an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		utils.Error(c, appErr)
		return
	}

	// Handle specific error types
	switch {
	case strings.Contains(err.Error(), "validation"):
		utils.ValidationError(c, err.Error())
	case strings.Contains(err.Error(), "unauthorized"):
		utils.Unauthorized(c, err.Error())
	case strings.Contains(err.Error(), "forbidden"):
		utils.Forbidden(c, err.Error())
	case strings.Contains(err.Error(), "not found"):
		utils.NotFound(c, err.Error())
	case strings.Contains(err.Error(), "conflict"):
		utils.Conflict(c, err.Error())
	case strings.Contains(err.Error(), "rate limit"):
		utils.RateLimitExceeded(c, err.Error())
	default:
		// Default internal server error
		errorResponse := utils.Response{
			Success: false,
			Error: &utils.ErrorInfo{
				Code:    http.StatusText(http.StatusInternalServerError),
				Message: config.DefaultErrorMessage,
			},
		}

		// Add request ID if enabled
		if config.IncludeRequestID {
			if requestID := c.GetHeader(config.RequestIDHeader); requestID != "" {
				errorResponse.Error.Details = fmt.Sprintf("Request ID: %s", requestID)
			}
		}

		// Add error details in development
		if gin.Mode() == gin.DebugMode {
			if errorResponse.Error.Details == "" {
				errorResponse.Error.Details = err.Error()
			} else {
				errorResponse.Error.Details += fmt.Sprintf(" - %s", err.Error())
			}
		}

		c.JSON(http.StatusInternalServerError, errorResponse)
	}
}

// logError logs error with context
func logError(c *gin.Context, err error, config *ErrorHandlerConfig) {
	fields := logger.Fields{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"ip":     c.ClientIP(),
		"error":  err.Error(),
	}

	// Add request ID if enabled
	if config.IncludeRequestID {
		if requestID := c.GetHeader(config.RequestIDHeader); requestID != "" {
			fields["request_id"] = requestID
		}
	}

	// Add request details if enabled
	if config.LogRequests {
		fields["user_agent"] = c.Request.UserAgent()
		fields["referer"] = c.Request.Referer()
	}

	// Add user information if available
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}
	if email, exists := c.Get("email"); exists {
		fields["email"] = email
	}

	// Log based on error type
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.StatusCode() {
		case http.StatusBadRequest:
			logger.WithFields(fields).Warn("Bad request error")
		case http.StatusUnauthorized:
			logger.WithFields(fields).Warn("Unauthorized error")
		case http.StatusForbidden:
			logger.WithFields(fields).Warn("Forbidden error")
		case http.StatusNotFound:
			logger.WithFields(fields).Info("Not found error")
		case http.StatusConflict:
			logger.WithFields(fields).Warn("Conflict error")
		case http.StatusTooManyRequests:
			logger.WithFields(fields).Warn("Rate limit error")
		default:
			logger.WithFields(fields).Error("Application error")
		}
	} else {
		logger.WithFields(fields).Error("Unhandled error")
	}
}

// shouldSkipErrorHandling checks if error handling should be skipped for the path
func shouldSkipErrorHandling(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.EnablePanicRecovery = true
	config.LogErrors = true
	return ErrorHandler(config)
}

// CustomErrorHandler returns a middleware with custom error handling
func CustomErrorHandler(errorHandler func(*gin.Context, error)) gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.ErrorHandler = errorHandler
	return ErrorHandler(config)
}

// WithCustomErrorHandlers returns a middleware with custom error handlers for specific errors
func WithCustomErrorHandlers(customHandlers map[error]gin.HandlerFunc) gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.CustomHandlers = customHandlers
	return ErrorHandler(config)
}

// SilentErrorHandler returns a middleware that doesn't log errors
func SilentErrorHandler() gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.LogErrors = false
	config.EnableStackTrace = false
	return ErrorHandler(config)
}

// DevelopmentErrorHandler returns a middleware with detailed error information for development
func DevelopmentErrorHandler() gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.EnableStackTrace = true
	config.LogErrors = true
	config.LogRequests = true
	config.DefaultErrorMessage = "An error occurred"
	return ErrorHandler(config)
}

// ProductionErrorHandler returns a middleware suitable for production
func ProductionErrorHandler() gin.HandlerFunc {
	config := DefaultErrorHandlerConfig()
	config.EnableStackTrace = false
	config.LogErrors = true
	config.LogRequests = false
	config.DefaultErrorMessage = "Internal server error"
	return ErrorHandler(config)
}