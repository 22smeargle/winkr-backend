package testutils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestConfig holds all test configuration
type TestConfig struct {
	Server    config.ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Storage   StorageConfig
	External  ExternalServicesConfig
	Security  SecurityConfig
}

// JWTConfig holds JWT configuration for tests
type JWTConfig struct {
	Secret         string
	ExpirationTime time.Duration
	RefreshTime    time.Duration
}

// StorageConfig holds storage configuration for tests
type StorageConfig struct {
	Type     string
	Endpoint string
	Region   string
	Bucket   string
	AccessKey string
	SecretKey string
}

// ExternalServicesConfig holds external services configuration for tests
type ExternalServicesConfig struct {
	Email    EmailConfig
	SMS      SMSConfig
	AI       AIConfig
	Payment  PaymentConfig
}

// EmailConfig holds email service configuration for tests
type EmailConfig struct {
	Provider string
	APIKey   string
	From     string
}

// SMSConfig holds SMS service configuration for tests
type SMSConfig struct {
	Provider string
	APIKey   string
	From     string
}

// AIConfig holds AI service configuration for tests
type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
}

// PaymentConfig holds payment service configuration for tests
type PaymentConfig struct {
	Provider   string
	SecretKey  string
	PublishKey string
	WebhookURL string
}

// SecurityConfig holds security configuration for tests
type SecurityConfig struct {
	RateLimiting RateLimitConfig
	CORS        CORSConfig
	CSRF        CSRFConfig
}

// RateLimitConfig holds rate limiting configuration for tests
type RateLimitConfig struct {
	Enabled    bool
	Requests   int
	Window     time.Duration
	BurstSize  int
}

// CORSConfig holds CORS configuration for tests
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// CSRFConfig holds CSRF configuration for tests
type CSRFConfig struct {
	Enabled    bool
	Secret     string
	CookieName string
	HeaderName string
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig() *TestConfig {
	return &TestConfig{
		Server: config.ServerConfig{
			Port: getEnv("TEST_SERVER_PORT", "8080"),
			Host: getEnv("TEST_SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("TEST_DB_HOST", "localhost"),
			Port:     getEnv("TEST_DB_PORT", "5432"),
			User:     getEnv("TEST_DB_USER", "test"),
			Password: getEnv("TEST_DB_PASSWORD", "test"),
			Name:     getEnv("TEST_DB_NAME", "winkr_test"),
			SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("TEST_REDIS_HOST", "localhost"),
			Port:     getEnv("TEST_REDIS_PORT", "6379"),
			Password: getEnv("TEST_REDIS_PASSWORD", ""),
			DB:       getEnvInt("TEST_REDIS_DB", 1),
		},
		JWT: JWTConfig{
			Secret:         getEnv("TEST_JWT_SECRET", "test-secret-key"),
			ExpirationTime: getEnvDuration("TEST_JWT_EXPIRATION", 24*time.Hour),
			RefreshTime:    getEnvDuration("TEST_JWT_REFRESH", 7*24*time.Hour),
		},
		Storage: StorageConfig{
			Type:      getEnv("TEST_STORAGE_TYPE", "local"),
			Endpoint:  getEnv("TEST_STORAGE_ENDPOINT", "http://localhost:9000"),
			Region:    getEnv("TEST_STORAGE_REGION", "us-east-1"),
			Bucket:    getEnv("TEST_STORAGE_BUCKET", "test-bucket"),
			AccessKey: getEnv("TEST_STORAGE_ACCESS_KEY", "test-access-key"),
			SecretKey: getEnv("TEST_STORAGE_SECRET_KEY", "test-secret-key"),
		},
		External: ExternalServicesConfig{
			Email: EmailConfig{
				Provider: getEnv("TEST_EMAIL_PROVIDER", "mock"),
				APIKey:   getEnv("TEST_EMAIL_API_KEY", "test-api-key"),
				From:     getEnv("TEST_EMAIL_FROM", "test@example.com"),
			},
			SMS: SMSConfig{
				Provider: getEnv("TEST_SMS_PROVIDER", "mock"),
				APIKey:   getEnv("TEST_SMS_API_KEY", "test-api-key"),
				From:     getEnv("TEST_SMS_FROM", "+1234567890"),
			},
			AI: AIConfig{
				Provider: getEnv("TEST_AI_PROVIDER", "mock"),
				APIKey:   getEnv("TEST_AI_API_KEY", "test-api-key"),
				Model:    getEnv("TEST_AI_MODEL", "test-model"),
			},
			Payment: PaymentConfig{
				Provider:   getEnv("TEST_PAYMENT_PROVIDER", "mock"),
				SecretKey:  getEnv("TEST_PAYMENT_SECRET_KEY", "test-secret-key"),
				PublishKey: getEnv("TEST_PAYMENT_PUBLISH_KEY", "test-publish-key"),
				WebhookURL: getEnv("TEST_PAYMENT_WEBHOOK_URL", "http://localhost:8080/webhook"),
			},
		},
		Security: SecurityConfig{
			RateLimiting: RateLimitConfig{
				Enabled:   getEnvBool("TEST_RATE_LIMIT_ENABLED", true),
				Requests:  getEnvInt("TEST_RATE_LIMIT_REQUESTS", 100),
				Window:    getEnvDuration("TEST_RATE_LIMIT_WINDOW", time.Minute),
				BurstSize: getEnvInt("TEST_RATE_LIMIT_BURST", 10),
			},
			CORS: CORSConfig{
				Enabled:          getEnvBool("TEST_CORS_ENABLED", true),
				AllowedOrigins:   getEnvSlice("TEST_CORS_ALLOWED_ORIGINS", []string{"*"}),
				AllowedMethods:   getEnvSlice("TEST_CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders:   getEnvSlice("TEST_CORS_ALLOWED_HEADERS", []string{"*"}),
				ExposedHeaders:   getEnvSlice("TEST_CORS_EXPOSED_HEADERS", []string{}),
				AllowCredentials: getEnvBool("TEST_CORS_ALLOW_CREDENTIALS", false),
				MaxAge:           getEnvDuration("TEST_CORS_MAX_AGE", time.Hour),
			},
			CSRF: CSRFConfig{
				Enabled:    getEnvBool("TEST_CSRF_ENABLED", false),
				Secret:     getEnv("TEST_CSRF_SECRET", "test-csrf-secret"),
				CookieName: getEnv("TEST_CSRF_COOKIE_NAME", "csrf_token"),
				HeaderName: getEnv("TEST_CSRF_HEADER_NAME", "X-CSRF-Token"),
			},
		},
	}
}

// GetDefaultTestConfig returns default test configuration
func GetDefaultTestConfig() *TestConfig {
	return &TestConfig{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Name:     "winkr_test",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1,
		},
		JWT: JWTConfig{
			Secret:         "test-secret-key",
			ExpirationTime: 24 * time.Hour,
			RefreshTime:    7 * 24 * time.Hour,
		},
		Storage: StorageConfig{
			Type:      "local",
			Endpoint:  "http://localhost:9000",
			Region:    "us-east-1",
			Bucket:    "test-bucket",
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
		},
		External: ExternalServicesConfig{
			Email: EmailConfig{
				Provider: "mock",
				APIKey:   "test-api-key",
				From:     "test@example.com",
			},
			SMS: SMSConfig{
				Provider: "mock",
				APIKey:   "test-api-key",
				From:     "+1234567890",
			},
			AI: AIConfig{
				Provider: "mock",
				APIKey:   "test-api-key",
				Model:    "test-model",
			},
			Payment: PaymentConfig{
				Provider:   "mock",
				SecretKey:  "test-secret-key",
				PublishKey: "test-publish-key",
				WebhookURL: "http://localhost:8080/webhook",
			},
		},
		Security: SecurityConfig{
			RateLimiting: RateLimitConfig{
				Enabled:   true,
				Requests:  100,
				Window:    time.Minute,
				BurstSize: 10,
			},
			CORS: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				ExposedHeaders:   []string{},
				AllowCredentials: false,
				MaxAge:           time.Hour,
			},
			CSRF: CSRFConfig{
				Enabled:    false,
				Secret:     "test-csrf-secret",
				CookieName: "csrf_token",
				HeaderName: "X-CSRF-Token",
			},
		},
	}
}

// ToAppConfig converts test config to application config
func (tc *TestConfig) ToAppConfig() *config.Config {
	return &config.Config{
		Server: tc.Server,
		Database: config.DatabaseConfig{
			Host:     tc.Database.Host,
			Port:     tc.Database.Port,
			User:     tc.Database.User,
			Password: tc.Database.Password,
			Name:     tc.Database.Name,
			SSLMode:  tc.Database.SSLMode,
		},
		Redis: config.RedisConfig{
			Host:     tc.Redis.Host,
			Port:     tc.Redis.Port,
			Password: tc.Redis.Password,
			DB:       tc.Redis.DB,
		},
		JWT: config.JWTConfig{
			Secret: tc.JWT.Secret,
		},
	}
}

// GetDatabaseConnectionString returns database connection string
func (tc *TestConfig) GetDatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		tc.Database.Host, tc.Database.Port, tc.Database.User, 
		tc.Database.Password, tc.Database.Name, tc.Database.SSLMode)
}

// GetRedisConnectionString returns Redis connection string
func (tc *TestConfig) GetRedisConnectionString() string {
	return fmt.Sprintf("%s:%s", tc.Redis.Host, tc.Redis.Port)
}

// IsMockProvider checks if an external service provider is set to mock
func (tc *TestConfig) IsMockProvider(service string) bool {
	switch service {
	case "email":
		return tc.External.Email.Provider == "mock"
	case "sms":
		return tc.External.SMS.Provider == "mock"
	case "ai":
		return tc.External.AI.Provider == "mock"
	case "payment":
		return tc.External.Payment.Provider == "mock"
	default:
		return false
	}
}

// IsLocalStorage checks if storage type is local
func (tc *TestConfig) IsLocalStorage() bool {
	return tc.Storage.Type == "local"
}

// IsRateLimitingEnabled checks if rate limiting is enabled
func (tc *TestConfig) IsRateLimitingEnabled() bool {
	return tc.Security.RateLimiting.Enabled
}

// IsCORSEnabled checks if CORS is enabled
func (tc *TestConfig) IsCORSEnabled() bool {
	return tc.Security.CORS.Enabled
}

// IsCSRFEnabled checks if CSRF is enabled
func (tc *TestConfig) IsCSRFEnabled() bool {
	return tc.Security.CSRF.Enabled
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In a real implementation, you might want more sophisticated parsing
		return []string{value}
	}
	return defaultValue
}

// TestEnvironmentType represents different test environment types
type TestEnvironmentType string

const (
	UnitTestEnv        TestEnvironmentType = "unit"
	IntegrationTestEnv TestEnvironmentType = "integration"
	E2ETestEnv         TestEnvironmentType = "e2e"
	LoadTestEnv        TestEnvironmentType = "load"
	SecurityTestEnv    TestEnvironmentType = "security"
)

// GetConfigForEnvironment returns configuration for a specific test environment type
func GetConfigForEnvironment(envType TestEnvironmentType) *TestConfig {
	config := GetDefaultTestConfig()
	
	switch envType {
	case UnitTestEnv:
		// Unit tests typically use mocks and minimal setup
		config.External.Email.Provider = "mock"
		config.External.SMS.Provider = "mock"
		config.External.AI.Provider = "mock"
		config.External.Payment.Provider = "mock"
		config.Storage.Type = "memory"
		config.Security.RateLimiting.Enabled = false
		config.Security.CSRF.Enabled = false
		
	case IntegrationTestEnv:
		// Integration tests use real services but in test mode
		config.Database.Name = "winkr_integration_test"
		config.Redis.DB = 2
		config.Storage.Type = "local"
		
	case E2ETestEnv:
		// E2E tests use production-like setup
		config.Database.Name = "winkr_e2e_test"
		config.Redis.DB = 3
		config.Storage.Type = "s3"
		
	case LoadTestEnv:
		// Load tests need optimized settings
		config.Database.Name = "winkr_load_test"
		config.Redis.DB = 4
		config.Security.RateLimiting.Enabled = false
		
	case SecurityTestEnv:
		// Security tests need all security features enabled
		config.Database.Name = "winkr_security_test"
		config.Redis.DB = 5
		config.Security.RateLimiting.Enabled = true
		config.Security.CSRF.Enabled = true
	}
	
	return config
}

// SaveTestConfig saves test configuration to environment variables
func SaveTestConfig(config *TestConfig) error {
	// Server
	os.Setenv("TEST_SERVER_PORT", config.Server.Port)
	os.Setenv("TEST_SERVER_HOST", config.Server.Host)
	
	// Database
	os.Setenv("TEST_DB_HOST", config.Database.Host)
	os.Setenv("TEST_DB_PORT", config.Database.Port)
	os.Setenv("TEST_DB_USER", config.Database.User)
	os.Setenv("TEST_DB_PASSWORD", config.Database.Password)
	os.Setenv("TEST_DB_NAME", config.Database.Name)
	os.Setenv("TEST_DB_SSLMODE", config.Database.SSLMode)
	
	// Redis
	os.Setenv("TEST_REDIS_HOST", config.Redis.Host)
	os.Setenv("TEST_REDIS_PORT", config.Redis.Port)
	os.Setenv("TEST_REDIS_PASSWORD", config.Redis.Password)
	os.Setenv("TEST_REDIS_DB", strconv.Itoa(config.Redis.DB))
	
	// JWT
	os.Setenv("TEST_JWT_SECRET", config.JWT.Secret)
	os.Setenv("TEST_JWT_EXPIRATION", config.JWT.ExpirationTime.String())
	os.Setenv("TEST_JWT_REFRESH", config.JWT.RefreshTime.String())
	
	// Storage
	os.Setenv("TEST_STORAGE_TYPE", config.Storage.Type)
	os.Setenv("TEST_STORAGE_ENDPOINT", config.Storage.Endpoint)
	os.Setenv("TEST_STORAGE_REGION", config.Storage.Region)
	os.Setenv("TEST_STORAGE_BUCKET", config.Storage.Bucket)
	os.Setenv("TEST_STORAGE_ACCESS_KEY", config.Storage.AccessKey)
	os.Setenv("TEST_STORAGE_SECRET_KEY", config.Storage.SecretKey)
	
	// External Services
	os.Setenv("TEST_EMAIL_PROVIDER", config.External.Email.Provider)
	os.Setenv("TEST_EMAIL_API_KEY", config.External.Email.APIKey)
	os.Setenv("TEST_EMAIL_FROM", config.External.Email.From)
	
	os.Setenv("TEST_SMS_PROVIDER", config.External.SMS.Provider)
	os.Setenv("TEST_SMS_API_KEY", config.External.SMS.APIKey)
	os.Setenv("TEST_SMS_FROM", config.External.SMS.From)
	
	os.Setenv("TEST_AI_PROVIDER", config.External.AI.Provider)
	os.Setenv("TEST_AI_API_KEY", config.External.AI.APIKey)
	os.Setenv("TEST_AI_MODEL", config.External.AI.Model)
	
	os.Setenv("TEST_PAYMENT_PROVIDER", config.External.Payment.Provider)
	os.Setenv("TEST_PAYMENT_SECRET_KEY", config.External.Payment.SecretKey)
	os.Setenv("TEST_PAYMENT_PUBLISH_KEY", config.External.Payment.PublishKey)
	os.Setenv("TEST_PAYMENT_WEBHOOK_URL", config.External.Payment.WebhookURL)
	
	// Security
	os.Setenv("TEST_RATE_LIMIT_ENABLED", strconv.FormatBool(config.Security.RateLimiting.Enabled))
	os.Setenv("TEST_RATE_LIMIT_REQUESTS", strconv.Itoa(config.Security.RateLimiting.Requests))
	os.Setenv("TEST_RATE_LIMIT_WINDOW", config.Security.RateLimiting.Window.String())
	os.Setenv("TEST_RATE_LIMIT_BURST", strconv.Itoa(config.Security.RateLimiting.BurstSize))
	
	os.Setenv("TEST_CORS_ENABLED", strconv.FormatBool(config.Security.CORS.Enabled))
	os.Setenv("TEST_CORS_ALLOW_CREDENTIALS", strconv.FormatBool(config.Security.CORS.AllowCredentials))
	os.Setenv("TEST_CORS_MAX_AGE", config.Security.CORS.MaxAge.String())
	
	os.Setenv("TEST_CSRF_ENABLED", strconv.FormatBool(config.Security.CSRF.Enabled))
	os.Setenv("TEST_CSRF_SECRET", config.Security.CSRF.Secret)
	os.Setenv("TEST_CSRF_COOKIE_NAME", config.Security.CSRF.CookieName)
	os.Setenv("TEST_CSRF_HEADER_NAME", config.Security.CSRF.HeaderName)
	
	return nil
}