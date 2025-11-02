package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ValidationConfig represents validation configuration
type ValidationConfig struct {
	// Validator instance
	Validator *validator.Validate
	
	// SkipPaths is a list of paths to skip validation
	SkipPaths []string `json:"skip_paths"`
	
	// SkipMethods is a list of HTTP methods to skip validation
	SkipMethods []string `json:"skip_methods"`
	
	// MaxBodySize limits the size of request body for validation
	MaxBodySize int64 `json:"max_body_size"`
	
	// RequiredHeaders is a map of required headers and their validation rules
	RequiredHeaders map[string]string `json:"required_headers"`
	
	// OptionalHeaders is a map of optional headers and their validation rules
	OptionalHeaders map[string]string `json:"optional_headers"`
	
	// CustomValidators allows custom validation functions
	CustomValidators map[string]validator.Func `json:"-"`
	
	// EnableBodyValidation enables request body validation
	EnableBodyValidation bool `json:"enable_body_validation"`
	
	// EnableQueryValidation enables query parameter validation
	EnableQueryValidation bool `json:"enable_query_validation"`
	
	// EnableHeaderValidation enables header validation
	EnableHeaderValidation bool `json:"enable_header_validation"`
	
	// StrictMode enables strict validation (fails on unknown fields)
	StrictMode bool `json:"strict_mode"`
}

// DefaultValidationConfig returns a default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	v := validator.New()
	
	// Register custom validators
	registerCustomValidators(v)
	
	return &ValidationConfig{
		Validator:             v,
		SkipPaths:            []string{"/health", "/health/db", "/metrics"},
		SkipMethods:          []string{"GET", "DELETE", "OPTIONS"},
		MaxBodySize:          1024 * 1024, // 1MB
		RequiredHeaders:       map[string]string{},
		OptionalHeaders:       map[string]string{},
		CustomValidators:      make(map[string]validator.Func),
		EnableBodyValidation:  true,
		EnableQueryValidation: true,
		EnableHeaderValidation: true,
		StrictMode:           false,
	}
}

// Validation returns a validation middleware with the given configuration
func Validation(config *ValidationConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Register custom validators
	for name, fn := range config.CustomValidators {
		config.Validator.RegisterValidation(name, fn)
	}

	return func(c *gin.Context) {
		// Skip validation for specified paths and methods
		if shouldSkipValidation(c.Request.URL.Path, c.Request.Method, config) {
			c.Next()
			return
		}

		// Validate headers
		if config.EnableHeaderValidation {
			if err := validateHeaders(c, config); err != nil {
				utils.ValidationError(c, err.Error())
				c.Abort()
				return
			}
		}

		// Validate query parameters
		if config.EnableQueryValidation {
			if err := validateQueryParams(c, config); err != nil {
				utils.ValidationError(c, err.Error())
				c.Abort()
				return
			}
		}

		// Validate request body
		if config.EnableBodyValidation && shouldValidateBody(c.Request.Method) {
			if err := validateBody(c, config); err != nil {
				utils.ValidationError(c, err.Error())
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// validateHeaders validates request headers
func validateHeaders(c *gin.Context, config *ValidationConfig) error {
	// Check required headers
	for header, rule := range config.RequiredHeaders {
		value := c.GetHeader(header)
		if value == "" {
			return fmt.Errorf("required header '%s' is missing", header)
		}

		if err := validateField(value, rule, config.Validator); err != nil {
			return fmt.Errorf("header '%s' validation failed: %w", header, err)
		}
	}

	// Check optional headers
	for header, rule := range config.OptionalHeaders {
		value := c.GetHeader(header)
		if value != "" {
			if err := validateField(value, rule, config.Validator); err != nil {
				return fmt.Errorf("header '%s' validation failed: %w", header, err)
			}
		}
	}

	return nil
}

// validateQueryParams validates query parameters
func validateQueryParams(c *gin.Context, config *ValidationConfig) error {
	// Get query parameters from route if available
	if routeParams := c.Params; len(routeParams) > 0 {
		for _, param := range routeParams {
			if err := config.Validator.Var(param.Value, "required"); err != nil {
				return fmt.Errorf("path parameter '%s' is required", param.Key)
			}
		}
	}

	// Validate query string parameters
	query := c.Request.URL.Query()
	for key, values := range query {
		for _, value := range values {
			if err := config.Validator.Var(value, "required"); err != nil {
				return fmt.Errorf("query parameter '%s' validation failed: %w", key, err)
			}
		}
	}

	return nil
}

// validateBody validates request body
func validateBody(c *gin.Context, config *ValidationConfig) error {
	// Check content type
	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "multipart/form-data") {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Check body size
	if int64(len(body)) > config.MaxBodySize {
		return fmt.Errorf("request body too large (max %d bytes)", config.MaxBodySize)
	}

	// Restore body for subsequent handlers
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Skip validation for empty body
	if len(body) == 0 {
		return nil
	}

	// Parse JSON for validation
	if strings.Contains(contentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			return fmt.Errorf("invalid JSON format: %w", err)
		}

		// Validate JSON structure
		if err := validateJSON(jsonData, config); err != nil {
			return err
		}
	}

	return nil
}

// validateField validates a single field using validator
func validateField(value, rule string, validator *validator.Validate) error {
	if rule == "" {
		return nil
	}

	// Split multiple validation rules
	rules := strings.Split(rule, ",")
	for _, r := range rules {
		r = strings.TrimSpace(r)
		if err := validator.Var(value, r); err != nil {
			return err
		}
	}

	return nil
}

// validateJSON validates JSON data structure
func validateJSON(data interface{}, config *ValidationConfig) error {
	// Convert to map for validation
	jsonMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid JSON structure")
	}

	// In strict mode, check for unknown fields
	if config.StrictMode {
		// This would require a schema definition to implement properly
		// For now, we'll just validate that the JSON is well-formed
	}

	// Validate each field
	for key, value := range jsonMap {
		// Convert value to string for validation
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case int, int8, int16, int32, int64:
			strValue = fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			strValue = fmt.Sprintf("%d", v)
		case float32, float64:
			strValue = fmt.Sprintf("%f", v)
		case bool:
			strValue = strconv.FormatBool(v)
		default:
			// For complex types, skip validation
			continue
		}

		// Basic validation
		if err := config.Validator.Var(strValue, "required"); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", key, err)
		}
	}

	return nil
}

// shouldSkipValidation checks if validation should be skipped
func shouldSkipValidation(path, method string, config *ValidationConfig) bool {
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

// shouldValidateBody checks if body should be validated for the method
func shouldValidateBody(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// registerCustomValidators registers custom validation functions
func registerCustomValidators(v *validator.Validate) {
	// Email validation with domain check
	v.RegisterValidation("email_domain", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		if email == "" {
			return true
		}
		
		// Basic email validation
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(email) {
			return false
		}
		
		// Domain validation (basic)
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return false
		}
		
		domain := parts[1]
		domainParts := strings.Split(domain, ".")
		return len(domainParts) >= 2
	})

	// Password strength validation
	v.RegisterValidation("password_strength", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if password == "" {
			return true
		}
		
		// At least 8 characters
		if len(password) < 8 {
			return false
		}
		
		// Contains uppercase
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		// Contains lowercase
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		// Contains number
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		// Contains special character
		hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
		
		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// Username validation
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if username == "" {
			return true
		}
		
		// 3-20 characters, alphanumeric and underscore only
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
		return usernameRegex.MatchString(username)
	})

	// Phone number validation
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if phone == "" {
			return true
		}
		
		// Basic phone validation (international format)
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		return phoneRegex.MatchString(phone)
	})

	// UUID validation
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		uuidStr := fl.Field().String()
		if uuidStr == "" {
			return true
		}
		
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		return uuidRegex.MatchString(strings.ToLower(uuidStr))
	})
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	v := validator.New()
	registerCustomValidators(v)
	return v.Struct(s)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	v := validator.New()
	registerCustomValidators(v)
	return v.Var(field, tag)
}

// BodyValidator returns a middleware that validates request body against a struct
func BodyValidator(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(model); err != nil {
			utils.ValidationError(c, err.Error())
			c.Abort()
			return
		}

		if err := ValidateStruct(model); err != nil {
			utils.ValidationError(c, err.Error())
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_body", model)
		c.Next()
	}
}

// QueryValidator returns a middleware that validates query parameters
func QueryValidator(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for param, rule := range rules {
			value := c.Query(param)
			if value != "" {
				if err := ValidateVar(value, rule); err != nil {
					utils.ValidationError(c, fmt.Sprintf("Query parameter '%s' validation failed: %s", param, err.Error()))
					c.Abort()
					return
				}
			} else if strings.Contains(rule, "required") {
				utils.ValidationError(c, fmt.Sprintf("Required query parameter '%s' is missing", param))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// HeaderValidator returns a middleware that validates headers
func HeaderValidator(required, optional map[string]string) gin.HandlerFunc {
	config := DefaultValidationConfig()
	config.RequiredHeaders = required
	config.OptionalHeaders = optional
	config.EnableBodyValidation = false
	config.EnableQueryValidation = false
	return Validation(config)
}

// CustomValidator returns a middleware with custom validation rules
func CustomValidator(customValidators map[string]validator.Func) gin.HandlerFunc {
	config := DefaultValidationConfig()
	config.CustomValidators = customValidators
	return Validation(config)
}