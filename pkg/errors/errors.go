package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// StatusCode returns the HTTP status code
func (e *AppError) StatusCode() int {
	return e.Code
}

// NewAppError creates a new application error
func NewAppError(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined application errors
var (
	// Validation errors
	ErrValidationFailed = NewAppError(http.StatusBadRequest, "Validation failed", "")
	ErrInvalidInput     = NewAppError(http.StatusBadRequest, "Invalid input", "")
	ErrRequiredField     = NewAppError(http.StatusBadRequest, "Required field is missing", "")

	// Authentication errors
	ErrUnauthorized      = NewAppError(http.StatusUnauthorized, "Unauthorized", "")
	ErrInvalidCredentials = NewAppError(http.StatusUnauthorized, "Invalid credentials", "")
	ErrTokenExpired      = NewAppError(http.StatusUnauthorized, "Token expired", "")
	ErrInvalidToken      = NewAppError(http.StatusUnauthorized, "Invalid token", "")
	ErrTokenMissing      = NewAppError(http.StatusUnauthorized, "Authorization token is required", "")
	ErrAccountLocked    = NewAppError(http.StatusLocked, "Account is locked", "")
	ErrTooManyFailedAttempts = NewAppError(http.StatusTooManyRequests, "Too many failed attempts", "")
	ErrEmailNotVerified = NewAppError(http.StatusForbidden, "Email is not verified", "")
	ErrInvalidDevice    = NewAppError(http.StatusUnauthorized, "Invalid device", "")
	ErrSessionExpired   = NewAppError(http.StatusUnauthorized, "Session expired", "")
	ErrInvalidPasswordReset = NewAppError(http.StatusBadRequest, "Invalid or expired password reset token", "")
	ErrVerificationCodeInvalid = NewAppError(http.StatusBadRequest, "Invalid verification code", "")
	ErrVerificationCodeExpired = NewAppError(http.StatusBadRequest, "Verification code expired", "")

	// Authorization errors
	ErrForbidden         = NewAppError(http.StatusForbidden, "Forbidden", "")
	ErrInsufficientPerms = NewAppError(http.StatusForbidden, "Insufficient permissions", "")

	// Not found errors
	ErrNotFound          = NewAppError(http.StatusNotFound, "Resource not found", "")
	ErrUserNotFound      = NewAppError(http.StatusNotFound, "User not found", "")
	ErrPhotoNotFound     = NewAppError(http.StatusNotFound, "Photo not found", "")
	ErrMatchNotFound     = NewAppError(http.StatusNotFound, "Match not found", "")
	ErrMessageNotFound   = NewAppError(http.StatusNotFound, "Message not found", "")
	ErrConversationNotFound = NewAppError(http.StatusNotFound, "Conversation not found", "")

	// Conflict errors
	ErrConflict          = NewAppError(http.StatusConflict, "Resource conflict", "")
	ErrEmailExists       = NewAppError(http.StatusConflict, "Email already exists", "")
	ErrUserExists        = NewAppError(http.StatusConflict, "User already exists", "")
	ErrAlreadyMatched    = NewAppError(http.StatusConflict, "Users already matched", "")
	ErrAlreadySwiped     = NewAppError(http.StatusConflict, "Already swiped this user", "")

	// Rate limiting errors
	ErrRateLimitExceeded = NewAppError(http.StatusTooManyRequests, "Rate limit exceeded", "")
	ErrTooManyRequests   = NewAppError(http.StatusTooManyRequests, "Too many requests", "")

	// Business logic errors
	ErrBusinessLogic     = NewAppError(http.StatusUnprocessableEntity, "Business logic error", "")
	ErrInvalidOperation  = NewAppError(http.StatusUnprocessableEntity, "Invalid operation", "")
	ErrAccountBanned     = NewAppError(http.StatusForbidden, "Account is banned", "")
	ErrAccountInactive   = NewAppError(http.StatusForbidden, "Account is inactive", "")
	ErrPhotoLimitExceeded = NewAppError(http.StatusUnprocessableEntity, "Photo limit exceeded", "")
	ErrPrimaryPhotoRequired = NewAppError(http.StatusUnprocessableEntity, "Primary photo is required", "")
	ErrCannotDeletePrimaryPhoto = NewAppError(http.StatusUnprocessableEntity, "Cannot delete primary photo", "")

	// File upload errors
	ErrFileUpload        = NewAppError(http.StatusBadRequest, "File upload failed", "")
	ErrInvalidFileType   = NewAppError(http.StatusBadRequest, "Invalid file type", "")
	ErrFileSizeExceeded  = NewAppError(http.StatusBadRequest, "File size exceeded", "")
	ErrPhotoVerification = NewAppError(http.StatusBadRequest, "Photo verification failed", "")

	// Payment errors
	ErrPayment           = NewAppError(http.StatusPaymentRequired, "Payment required", "")
	ErrSubscription      = NewAppError(http.StatusPaymentRequired, "Subscription required", "")
	ErrPaymentFailed     = NewAppError(http.StatusPaymentRequired, "Payment failed", "")
	ErrInvalidPlan       = NewAppError(http.StatusBadRequest, "Invalid subscription plan", "")

	// External service errors
	ErrExternalService  = NewAppError(http.StatusBadGateway, "External service error", "")
	ErrAWSService       = NewAppError(http.StatusBadGateway, "AWS service error", "")
	ErrStripeService    = NewAppError(http.StatusBadGateway, "Stripe service error", "")
	ErrEmailService     = NewAppError(http.StatusBadGateway, "Email service error", "")

	// Database errors
	ErrDatabase          = NewAppError(http.StatusInternalServerError, "Database error", "")
	ErrConnectionFailed  = NewAppError(http.StatusInternalServerError, "Database connection failed", "")
	ErrQueryFailed       = NewAppError(http.StatusInternalServerError, "Database query failed", "")

	// Internal server errors
	ErrInternalServer    = NewAppError(http.StatusInternalServerError, "Internal server error", "")
	ErrServiceUnavailable = NewAppError(http.StatusServiceUnavailable, "Service unavailable", "")
	ErrTimeout           = NewAppError(http.StatusRequestTimeout, "Request timeout", "")
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternalServer
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) *AppError {
	if IsAppError(err) {
		appErr := GetAppError(err)
		return NewAppError(appErr.Code, message, appErr.Error())
	}
	return NewAppError(http.StatusInternalServerError, message, err.Error())
}

// NewValidationError creates a validation error with field details
func NewValidationError(field, message string) *AppError {
	return NewAppError(http.StatusBadRequest, "Validation failed", fmt.Sprintf("%s: %s", field, message))
}

// NewNotFoundError creates a not found error for a specific resource
func NewNotFoundError(resource string) *AppError {
	return NewAppError(http.StatusNotFound, fmt.Sprintf("%s not found", resource), "")
}

// NewConflictError creates a conflict error with a specific message
func NewConflictError(message string) *AppError {
	return NewAppError(http.StatusConflict, message, "")
}

// NewUnauthorizedError creates an unauthorized error with a specific message
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, message, "")
}

// NewForbiddenError creates a forbidden error with a specific message
func NewForbiddenError(message string) *AppError {
	return NewAppError(http.StatusForbidden, message, "")
}

// NewInternalError creates an internal server error with a specific message
func NewInternalError(message string) *AppError {
	return NewAppError(http.StatusInternalServerError, message, "")
}

// NewExternalServiceError creates an external service error with a specific message
func NewExternalServiceError(service, message string) *AppError {
	return NewAppError(http.StatusBadGateway, fmt.Sprintf("%s service error", service), message)
}

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication  ErrorType = "authentication"
	ErrorTypeAuthorization   ErrorType = "authorization"
	ErrorTypeNotFound        ErrorType = "not_found"
	ErrorTypeConflict        ErrorType = "conflict"
	ErrorTypeRateLimit       ErrorType = "rate_limit"
	ErrorTypeBusinessLogic   ErrorType = "business_logic"
	ErrorTypeFileUpload      ErrorType = "file_upload"
	ErrorTypePayment         ErrorType = "payment"
	ErrorTypeExternalService ErrorType = "external_service"
	ErrorTypeDatabase        ErrorType = "database"
	ErrorTypeInternal        ErrorType = "internal"
)

// GetErrorType returns the type of error
func GetErrorType(err error) ErrorType {
	if !IsAppError(err) {
		return ErrorTypeInternal
	}

	appErr := GetAppError(err)
	switch appErr.Code {
	case http.StatusBadRequest:
		return ErrorTypeValidation
	case http.StatusUnauthorized:
		return ErrorTypeAuthentication
	case http.StatusForbidden:
		return ErrorTypeAuthorization
	case http.StatusNotFound:
		return ErrorTypeNotFound
	case http.StatusConflict:
		return ErrorTypeConflict
	case http.StatusTooManyRequests:
		return ErrorTypeRateLimit
	case http.StatusUnprocessableEntity:
		return ErrorTypeBusinessLogic
	case http.StatusPaymentRequired:
		return ErrorTypePayment
	case http.StatusBadGateway:
		return ErrorTypeExternalService
	case http.StatusInternalServerError:
		return ErrorTypeInternal
	case http.StatusServiceUnavailable:
		return ErrorTypeExternalService
	case http.StatusRequestTimeout:
		return ErrorTypeInternal
	default:
		return ErrorTypeInternal
	}
}