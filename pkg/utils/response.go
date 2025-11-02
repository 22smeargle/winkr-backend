package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// Success sends a success response
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a success response with a message
func SuccessWithMessage(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// SuccessMessage sends a success response with only a message
func SuccessMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode(), Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    http.StatusText(appErr.StatusCode()),
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	// Default error response
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: "Internal server error",
			Details: err.Error(),
		},
	})
}

// ErrorWithStatus sends an error response with custom status code
func ErrorWithStatus(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(statusCode),
			Message: message,
		},
	})
}

// ErrorWithDetails sends an error response with details
func ErrorWithDetails(c *gin.Context, statusCode int, message, details string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(statusCode),
			Message: message,
			Details: details,
		},
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: "Validation failed",
			Details: message,
		},
	})
}

// Unauthorized sends an unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusUnauthorized),
			Message: message,
		},
	})
}

// Forbidden sends a forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusForbidden),
			Message: message,
		},
	})
}

// NotFound sends a not found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusNotFound),
			Message: message,
		},
	})
}

// Conflict sends a conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusConflict),
			Message: message,
		},
	})
}

// RateLimitExceeded sends a rate limit exceeded response
func RateLimitExceeded(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusTooManyRequests),
			Message: message,
		},
	})
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, statusCode int, data interface{}, total, limit, offset int) {
	c.JSON(statusCode, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: PaginationInfo{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

// Created sends a created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// CreatedWithMessage sends a created response with a message
func CreatedWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// NoContent sends a no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest sends a bad request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusBadRequest),
			Message: message,
		},
	})
}

// InternalServerError sends an internal server error response
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusInternalServerError),
			Message: message,
		},
	})
}

// ServiceUnavailable sends a service unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusServiceUnavailable),
			Message: message,
		},
	})
}

// PaymentRequired sends a payment required response
func PaymentRequired(c *gin.Context, message string) {
	c.JSON(http.StatusPaymentRequired, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    http.StatusText(http.StatusPaymentRequired),
			Message: message,
		},
	})
}