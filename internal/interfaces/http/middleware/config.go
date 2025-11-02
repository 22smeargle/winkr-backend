package middleware

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// MiddlewareConfig represents all middleware configuration
type MiddlewareConfig struct {
	// CORS configuration
	CORS *CORSConfig `json:"cors"`
	
	// Logging configuration
	Logging *LoggingConfig `json:"logging"`
	
	// Rate limiting configuration
	RateLimit *RateLimitConfig `json:"rate_limit"`
	
	// Authentication configuration
	Auth *AuthConfig `json:"auth"`
	
	// Error handling configuration
	ErrorHandler *ErrorHandlerConfig `json:"error_handler"`
	
	// Validation configuration
	Validation *ValidationConfig `json:"validation"`
	
	// Security configuration
	Security *SecurityConfig `json:"security"`
}

// LoadMiddlewareConfig creates middleware configuration from application config
func LoadMiddlewareConfig(appConfig *config.Config, redisClient *redis.Client, jwtUtils *utils.JWTUtils) *MiddlewareConfig {
	return &MiddlewareConfig{
		CORS:        loadCORSConfig(appConfig),
		Logging:      loadLoggingConfig(appConfig),
		RateLimit:    loadRateLimitConfig(appConfig, redisClient),
		Auth:         loadAuthConfig(appConfig, jwtUtils),
		ErrorHandler:  loadErrorHandlerConfig(appConfig),
		Validation:    loadValidationConfig(appConfig),
		Security:     loadSecurityConfig(appConfig),
	}
}

// loadCORSConfig creates CORS configuration from app config
func loadCORSConfig(appConfig *config.Config) *CORSConfig {
	// In production, you might want to load these from config
	// For now, we'll use sensible defaults
	config := DefaultCORSConfig()
	
	// Adjust based on environment
	if appConfig.App.Env == "production" {
		// Restrictive CORS for production
		config.AllowedOrigins = []string{"https://yourdomain.com", "https://www.yourdomain.com"}
		config.AllowCredentials = true
	} else if appConfig.App.Env == "staging" {
		// Moderate CORS for staging
		config.AllowedOrigins = []string{"https://staging.yourdomain.com"}
		config.AllowCredentials = true
	}
	
	return config
}

// loadLoggingConfig creates logging configuration from app config
func loadLoggingConfig(appConfig *config.Config) *LoggingConfig {
	config := DefaultLoggingConfig()
	
	// Adjust based on environment
	if appConfig.App.Env == "production" {
		config.EnableColors = false
		config.LogRequestBody = false
		config.LogResponseBody = false
	} else {
		config.EnableColors = true
		config.LogRequestBody = true
		config.LogResponseBody = true
	}
	
	return config
}

// loadRateLimitConfig creates rate limiting configuration from app config
func loadRateLimitConfig(appConfig *config.Config, redisClient *redis.Client) *RateLimitConfig {
	config := DefaultRateLimitConfig(redisClient)
	
	// Override with config values if available
	if appConfig.RateLimit.RequestsPerMinute > 0 {
		config.DefaultRequestsPerMinute = appConfig.RateLimit.RequestsPerMinute
	}
	if appConfig.RateLimit.RequestsPerHour > 0 {
		config.DefaultRequestsPerHour = appConfig.RateLimit.RequestsPerHour
	}
	
	// Adjust based on environment
	if appConfig.App.Env == "development" {
		// More lenient rate limiting for development
		config.DefaultRequestsPerMinute = 1000
		config.DefaultRequestsPerHour = 10000
		config.SkipPaths = append(config.SkipPaths, "/debug", "/pprof")
	}
	
	return config
}

// loadAuthConfig creates authentication configuration from app config
func loadAuthConfig(appConfig *config.Config, jwtUtils *utils.JWTUtils) *AuthConfig {
	config := DefaultAuthConfig(jwtUtils)
	
	// Adjust based on environment
	if appConfig.App.Env == "development" {
		// More relaxed auth for development
		config.SkipPaths = append(config.SkipPaths, "/debug", "/test")
	}
	
	return config
}

// loadErrorHandlerConfig creates error handling configuration from app config
func loadErrorHandlerConfig(appConfig *config.Config) *ErrorHandlerConfig {
	config := DefaultErrorHandlerConfig()
	
	// Adjust based on environment
	if appConfig.App.Env == "production" {
		config.EnableStackTrace = false
		config.DefaultErrorMessage = "Internal server error"
	} else {
		config.EnableStackTrace = true
		config.DefaultErrorMessage = "An error occurred"
	}
	
	return config
}

// loadValidationConfig creates validation configuration from app config
func loadValidationConfig(appConfig *config.Config) *ValidationConfig {
	config := DefaultValidationConfig()
	
	// Adjust based on environment
	if appConfig.App.Env == "development" {
		config.MaxBodySize = 10 * 1024 * 1024 // 10MB for development
	} else {
		config.MaxBodySize = 1024 * 1024 // 1MB for production
	}
	
	return config
}

// loadSecurityConfig creates security configuration from app config
func loadSecurityConfig(appConfig *config.Config) *SecurityConfig {
	config := DefaultSecurityConfig()
	
	// Adjust based on environment
	if appConfig.App.Env == "production" {
		config.RequireHTTPS = true
		config.EnableHSTS = true
		config.MaxAge = 31536000 // 1 year
		config.Preload = true
		config.EnableIPFiltering = true
		config.EnableBotDetection = true
	} else if appConfig.App.Env == "staging" {
		config.RequireHTTPS = true
		config.EnableHSTS = true
		config.MaxAge = 3600 // 1 hour
		config.Preload = false
	} else {
		// Development
		config.RequireHTTPS = false
		config.EnableHSTS = false
		config.EnableIPFiltering = false
		config.EnableBotDetection = false
	}
	
	return config
}

// DevelopmentConfig returns middleware configuration for development environment
func DevelopmentConfig(redisClient *redis.Client, jwtUtils *utils.JWTUtils) *MiddlewareConfig {
	return &MiddlewareConfig{
		CORS:        DefaultCORSConfig(),
		Logging:      DefaultLoggingConfig(),
		RateLimit:    DefaultRateLimitConfig(redisClient),
		Auth:         DefaultAuthConfig(jwtUtils),
		ErrorHandler:  DefaultErrorHandlerConfig(),
		Validation:    DefaultValidationConfig(),
		Security:     DefaultSecurityConfig(),
	}
}

// ProductionConfig returns middleware configuration for production environment
func ProductionConfig(redisClient *redis.Client, jwtUtils *utils.JWTUtils) *MiddlewareConfig {
	corsConfig := DefaultCORSConfig()
	corsConfig.AllowedOrigins = []string{"https://yourdomain.com"}
	corsConfig.AllowCredentials = true
	
	loggingConfig := DefaultLoggingConfig()
	loggingConfig.EnableColors = false
	loggingConfig.LogRequestBody = false
	loggingConfig.LogResponseBody = false
	
	rateLimitConfig := DefaultRateLimitConfig(redisClient)
	rateLimitConfig.DefaultRequestsPerMinute = 60
	rateLimitConfig.DefaultRequestsPerHour = 1000
	
	authConfig := DefaultAuthConfig(jwtUtils)
	
	errorHandlerConfig := DefaultErrorHandlerConfig()
	errorHandlerConfig.EnableStackTrace = false
	errorHandlerConfig.DefaultErrorMessage = "Internal server error"
	
	validationConfig := DefaultValidationConfig()
	validationConfig.MaxBodySize = 1024 * 1024 // 1MB
	
	securityConfig := DefaultSecurityConfig()
	securityConfig.RequireHTTPS = true
	securityConfig.EnableHSTS = true
	securityConfig.MaxAge = 31536000 // 1 year
	securityConfig.Preload = true
	securityConfig.EnableIPFiltering = true
	securityConfig.EnableBotDetection = true
	
	return &MiddlewareConfig{
		CORS:        corsConfig,
		Logging:      loggingConfig,
		RateLimit:    rateLimitConfig,
		Auth:         authConfig,
		ErrorHandler:  errorHandlerConfig,
		Validation:    validationConfig,
		Security:     securityConfig,
	}
}

// TestingConfig returns middleware configuration for testing environment
func TestingConfig(redisClient *redis.Client, jwtUtils *utils.JWTUtils) *MiddlewareConfig {
	corsConfig := DefaultCORSConfig()
	corsConfig.AllowedOrigins = []string{"*"}
	
	loggingConfig := DefaultLoggingConfig()
	loggingConfig.EnableColors = false
	loggingConfig.LogRequestBody = false
	loggingConfig.LogResponseBody = false
	
	rateLimitConfig := DefaultRateLimitConfig(redisClient)
	rateLimitConfig.DefaultRequestsPerMinute = 10000 // Very high for testing
	rateLimitConfig.DefaultRequestsPerHour = 100000
	
	authConfig := DefaultAuthConfig(jwtUtils)
	authConfig.SkipPaths = append(authConfig.SkipPaths, "/test", "/mock")
	
	errorHandlerConfig := DefaultErrorHandlerConfig()
	errorHandlerConfig.EnableStackTrace = true
	errorHandlerConfig.DefaultErrorMessage = "Test error"
	
	validationConfig := DefaultValidationConfig()
	validationConfig.MaxBodySize = 50 * 1024 * 1024 // 50MB for testing
	
	securityConfig := DefaultSecurityConfig()
	securityConfig.RequireHTTPS = false
	securityConfig.EnableHSTS = false
	securityConfig.EnableIPFiltering = false
	securityConfig.EnableBotDetection = false
	
	return &MiddlewareConfig{
		CORS:        corsConfig,
		Logging:      loggingConfig,
		RateLimit:    rateLimitConfig,
		Auth:         authConfig,
		ErrorHandler:  errorHandlerConfig,
		Validation:    validationConfig,
		Security:     securityConfig,
	}
}

// CustomConfig allows creating custom middleware configuration
func CustomConfig(
	corsConfig *CORSConfig,
	loggingConfig *LoggingConfig,
	rateLimitConfig *RateLimitConfig,
	authConfig *AuthConfig,
	errorHandlerConfig *ErrorHandlerConfig,
	validationConfig *ValidationConfig,
	securityConfig *SecurityConfig,
) *MiddlewareConfig {
	return &MiddlewareConfig{
		CORS:        corsConfig,
		Logging:      loggingConfig,
		RateLimit:    rateLimitConfig,
		Auth:         authConfig,
		ErrorHandler:  errorHandlerConfig,
		Validation:    validationConfig,
		Security:     securityConfig,
	}
}