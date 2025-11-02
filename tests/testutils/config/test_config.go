package config

import (
	"os"
	"time"

	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestEnvironment represents different test environments
type TestEnvironment string

const (
	UnitTestEnv     TestEnvironment = "unit"
	IntegrationEnv  TestEnvironment = "integration"
	E2ETestEnv      TestEnvironment = "e2e"
	LoadTestEnv     TestEnvironment = "load"
	SecurityTestEnv TestEnvironment = "security"
)

// TestConfig holds configuration for different test environments
type TestConfig struct {
	Environment TestEnvironment
	Database   config.DatabaseConfig
	Redis      config.RedisConfig
	JWT        config.JWTConfig
	Security   config.SecurityConfig
	Storage    config.StorageConfig
	EphemeralPhoto config.EphemeralPhotoConfig
	Chat       config.ChatConfig
}

// GetTestConfig returns configuration for the specified test environment
func GetTestConfig(env TestEnvironment) *TestConfig {
	switch env {
	case UnitTestEnv:
		return getUnitTestConfig()
	case IntegrationEnv:
		return getIntegrationTestConfig()
	case E2ETestEnv:
		return getE2ETestConfig()
	case LoadTestEnv:
		return getLoadTestConfig()
	case SecurityTestEnv:
		return getSecurityTestConfig()
	default:
		return getUnitTestConfig()
	}
}

// getUnitTestConfig returns configuration for unit tests
func getUnitTestConfig() *TestConfig {
	return &TestConfig{
		Environment: UnitTestEnv,
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test_user",
			Password:        "test_password",
			DBName:          "test_unit_db",
			SSLMode:         "disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime:  300,
			ConnMaxIdleTime:  60,
			Timezone:        "UTC",
			MigrationsPath:  "../../migrations",
		},
		Redis: config.RedisConfig{
			Host:               "localhost",
			Port:               6379,
			Password:           "",
			DB:                 2, // Separate DB for unit tests
			PoolSize:           5,
			MinIdleConns:       2,
			MaxRetries:         3,
			DialTimeout:        3 * time.Second,
			ReadTimeout:        2 * time.Second,
			WriteTimeout:       2 * time.Second,
			PoolTimeout:        3 * time.Second,
			IdleTimeout:        2 * time.Minute,
			IdleCheckFrequency: 30 * time.Second,
			ClusterEnabled:     false,
			ClusterAddresses:   []string{},
			MaxRedirects:       3,
			RouteByLatency:     false,
			RouteRandomly:      false,
		},
		JWT: config.JWTConfig{
			Secret:              "unit-test-secret-key-for-testing-only",
			AccessTokenExpiry:   15 * time.Minute,
			RefreshTokenExpiry:  24 * time.Hour,
			Issuer:              "unit-test-issuer",
			RefreshTokenRotation: true,
			MaxActiveSessions:   3,
		},
		Security: config.SecurityConfig{
			AccountLockoutEnabled:    true,
			MaxFailedAttempts:        3,
			AccountLockoutDuration:  5 * time.Minute,
			PasswordMinLength:        6,
			PasswordRequireUppercase: false,
			PasswordRequireLowercase: false,
			PasswordRequireNumbers:   false,
			PasswordRequireSymbols:   false,
			SessionTimeout:          1 * time.Hour,
			DeviceFingerprinting:    false,
			CSRFProtection:         false,
		},
		Storage: config.StorageConfig{
			Provider:       "mock",
			Region:         "us-east-1",
			AccessKeyID:    "unit-test-access-key",
			SecretAccessKey: "unit-test-secret-key",
			Bucket:         "unit-test-bucket",
			Endpoint:       "localhost:9000",
			UseSSL:         false,
			UploadExpiry:   5 * time.Minute,
			DownloadExpiry:  30 * time.Minute,
			MaxFileSize:    1 * 1024 * 1024, // 1MB
			AllowedTypes:   []string{"image/jpeg", "image/png"},
		},
		EphemeralPhoto: config.EphemeralPhotoConfig{
			MaxFileSize:        1 * 1024 * 1024, // 1MB
			AllowedTypes:      []string{"image/jpeg", "image/png"},
			MaxPhotosPerUser:  5,
			DefaultDuration:   10 * time.Second,
			MaxDuration:      60 * time.Second,
			ViewDuration:     10 * time.Second,
			AccessKeyLength:  16,
			EnableWatermark:  false,
			WatermarkText:    "Unit Test",
			PreventDownload:  false,
			StorageTier:      "hot",
			CleanupInterval:  1 * time.Minute,
			RetentionPeriod:  1 * time.Hour,
			CacheTTL:         30 * time.Second,
			ViewCacheTTL:     10 * time.Second,
			UploadRateLimit:  5,
			ViewRateLimit:    10,
			EnableAnalytics:  false,
			AnalyticsTTL:     1 * time.Hour,
			JobInterval:      30 * time.Second,
			JobBatchSize:     10,
			EnableJobRetry:   false,
			MaxJobRetries:    1,
		},
		Chat: config.ChatConfig{
			WebSocket: config.WebSocketConfig{
				Enabled:                false,
				Path:                   "/ws",
				AllowedOrigins:         []string{"*"},
				PingInterval:           30 * time.Second,
				PongWait:               60 * time.Second,
				WriteWait:              10 * time.Second,
				MaxMessageSize:         1024, // 1KB
				ReadBufferSize:         512,
				WriteBufferSize:        512,
				CompressionEnabled:     false,
				MaxConnectionsPerUser:  2,
				ConnectionTimeout:      30 * time.Second,
				ReconnectInterval:      5 * time.Second,
				HeartbeatInterval:      30 * time.Second,
			},
			Message: config.MessageConfig{
				MaxTextLength:          500,
				MaxMessageAge:          24 * time.Hour,
				MaxMessagesPerRequest:  10,
				AllowedMessageTypes:    []string{"text"},
				MaxPhotoSize:           1 * 1024 * 1024, // 1MB
				AllowedPhotoTypes:      []string{"image/jpeg"},
				PhotoExpiry:            24 * time.Hour,
				EphemeralPhotoDuration: 5 * time.Second,
				LocationAccuracy:       100.0,
				SystemMessagePrefix:    "[System]",
				EncryptionEnabled:      false,
				EncryptionKey:          "",
			},
			Security: config.ChatSecurityConfig{
				ContentFilteringEnabled: false,
				BannedWords:            []string{},
				BannedPatterns:         []string{},
				SpamDetectionEnabled:   false,
				SpamThreshold:          0.8,
				SpamWindow:             1 * time.Minute,
				MessageRateLimit:       10,
				MessageRateWindow:      1 * time.Minute,
				LinkPreviewEnabled:     false,
				AllowedLinkDomains:     []string{},
				BlockedLinkDomains:     []string{},
				PIIDetectionEnabled:     false,
				ReportThreshold:        3,
				AutoBanThreshold:       5,
			},
			Cache: config.ChatCacheConfig{
				UserOnlineStatusTTL:    30 * time.Second,
				TypingIndicatorTTL:     5 * time.Second,
				ConversationTTL:        10 * time.Minute,
				MessageTTL:             30 * time.Minute,
				UnreadCountTTL:         1 * time.Minute,
				LinkPreviewTTL:         1 * time.Hour,
				MaxCachedConversations: 10,
				MaxCachedMessages:      50,
				CleanupInterval:        5 * time.Minute,
				PubSubEnabled:          false,
			},
			RateLimit: config.ChatRateLimitConfig{
				MessagesPerMinute:         10,
				MessagesPerHour:           100,
				MessagesPerDay:            500,
				ConversationsPerDay:       10,
				PhotosPerDay:              5,
				ConnectionsPerMinute:      2,
				TypingIndicatorsPerMinute: 5,
			},
		},
	}
}

// getIntegrationTestConfig returns configuration for integration tests
func getIntegrationTestConfig() *TestConfig {
	cfg := getUnitTestConfig()
	cfg.Environment = IntegrationEnv
	cfg.Database.DBName = "test_integration_db"
	cfg.Redis.DB = 3
	cfg.JWT.Secret = "integration-test-secret-key-for-testing-only"
	cfg.JWT.Issuer = "integration-test-issuer"
	cfg.Storage.Bucket = "integration-test-bucket"
	cfg.EphemeralPhoto.WatermarkText = "Integration Test"
	cfg.Chat.WebSocket.Enabled = true
	cfg.Chat.Message.AllowedMessageTypes = []string{"text", "photo", "ephemeral_photo"}
	cfg.Chat.Security.ContentFilteringEnabled = true
	cfg.Chat.Security.SpamDetectionEnabled = true
	cfg.Chat.Cache.PubSubEnabled = true
	return cfg
}

// getE2ETestConfig returns configuration for end-to-end tests
func getE2ETestConfig() *TestConfig {
	cfg := getIntegrationTestConfig()
	cfg.Environment = E2ETestEnv
	cfg.Database.DBName = "test_e2e_db"
	cfg.Redis.DB = 4
	cfg.JWT.Secret = "e2e-test-secret-key-for-testing-only"
	cfg.JWT.Issuer = "e2e-test-issuer"
	cfg.Storage.Provider = "s3"
	cfg.Storage.Bucket = "e2e-test-bucket"
	cfg.EphemeralPhoto.EnableWatermark = true
	cfg.EphemeralPhoto.PreventDownload = true
	cfg.Chat.WebSocket.MaxConnectionsPerUser = 5
	cfg.Chat.Message.MaxTextLength = 2000
	cfg.Chat.Message.AllowedMessageTypes = []string{"text", "photo", "ephemeral_photo", "location", "system", "gift"}
	cfg.Chat.Security.PIIDetectionEnabled = true
	return cfg
}

// getLoadTestConfig returns configuration for load tests
func getLoadTestConfig() *TestConfig {
	cfg := getIntegrationTestConfig()
	cfg.Environment = LoadTestEnv
	cfg.Database.DBName = "test_load_db"
	cfg.Redis.DB = 5
	cfg.Database.MaxOpenConns = 50
	cfg.Database.MaxIdleConns = 20
	cfg.Redis.PoolSize = 50
	cfg.Redis.MinIdleConns = 20
	cfg.JWT.MaxActiveSessions = 10
	cfg.Storage.MaxFileSize = 10 * 1024 * 1024 // 10MB
	cfg.EphemeralPhoto.MaxPhotosPerUser = 50
	cfg.EphemeralPhoto.UploadRateLimit = 100
	cfg.EphemeralPhoto.ViewRateLimit = 500
	cfg.Chat.WebSocket.MaxConnectionsPerUser = 10
	cfg.Chat.Message.MaxMessagesPerRequest = 100
	cfg.Chat.Security.MessageRateLimit = 100
	cfg.Chat.RateLimit.MessagesPerMinute = 100
	cfg.Chat.RateLimit.MessagesPerHour = 5000
	cfg.Chat.RateLimit.MessagesPerDay = 50000
	return cfg
}

// getSecurityTestConfig returns configuration for security tests
func getSecurityTestConfig() *TestConfig {
	cfg := getIntegrationTestConfig()
	cfg.Environment = SecurityTestEnv
	cfg.Database.DBName = "test_security_db"
	cfg.Redis.DB = 6
	cfg.Security.AccountLockoutEnabled = true
	cfg.Security.MaxFailedAttempts = 5
	cfg.Security.AccountLockoutDuration = 15 * time.Minute
	cfg.Security.PasswordMinLength = 8
	cfg.Security.PasswordRequireUppercase = true
	cfg.Security.PasswordRequireLowercase = true
	cfg.Security.PasswordRequireNumbers = true
	cfg.Security.PasswordRequireSymbols = true
	cfg.Security.SessionTimeout = 24 * time.Hour
	cfg.Security.DeviceFingerprinting = true
	cfg.Security.CSRFProtection = true
	cfg.Chat.Security.ContentFilteringEnabled = true
	cfg.Chat.Security.SpamDetectionEnabled = true
	cfg.Chat.Security.PIIDetectionEnabled = true
	cfg.Chat.Security.ReportThreshold = 1
	cfg.Chat.Security.AutoBanThreshold = 3
	return cfg
}

// GetTestConfigFromEnv returns test configuration from environment variables
func GetTestConfigFromEnv() *TestConfig {
	env := TestEnvironment(getEnvOrDefault("TEST_ENV", string(UnitTestEnv)))
	return GetTestConfig(env)
}

// IsTestEnvironment checks if the current environment is a test environment
func IsTestEnvironment() bool {
	env := os.Getenv("APP_ENV")
	return env == "test" || env == "unit" || env == "integration" || env == "e2e" || env == "load" || env == "security"
}

// IsCIEnvironment checks if running in CI/CD environment
func IsCIEnvironment() bool {
	return os.Getenv("CI") == "true" || os.Getenv("CI") == "1"
}

// IsDockerAvailable checks if Docker is available
func IsDockerAvailable() bool {
	_, err := os.Stat("/var/run/docker.sock")
	return err == nil
}

// GetTestDatabaseConfig returns database configuration for tests
func GetTestDatabaseConfig(env TestEnvironment) config.DatabaseConfig {
	return GetTestConfig(env).Database
}

// GetTestRedisConfig returns Redis configuration for tests
func GetTestRedisConfig(env TestEnvironment) config.RedisConfig {
	return GetTestConfig(env).Redis
}

// GetTestJWTConfig returns JWT configuration for tests
func GetTestJWTConfig(env TestEnvironment) config.JWTConfig {
	return GetTestConfig(env).JWT
}

// GetTestSecurityConfig returns security configuration for tests
func GetTestSecurityConfig(env TestEnvironment) config.SecurityConfig {
	return GetTestConfig(env).Security
}

// GetTestStorageConfig returns storage configuration for tests
func GetTestStorageConfig(env TestEnvironment) config.StorageConfig {
	return GetTestConfig(env).Storage
}

// GetTestEphemeralPhotoConfig returns ephemeral photo configuration for tests
func GetTestEphemeralPhotoConfig(env TestEnvironment) config.EphemeralPhotoConfig {
	return GetTestConfig(env).EphemeralPhoto
}

// GetTestChatConfig returns chat configuration for tests
func GetTestChatConfig(env TestEnvironment) config.ChatConfig {
	return GetTestConfig(env).Chat
}

// OverrideTestConfigWithEnv overrides test configuration with environment variables
func OverrideTestConfigWithEnv(cfg *TestConfig) *TestConfig {
	if dbHost := os.Getenv("TEST_DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := os.Getenv("TEST_DB_PORT"); dbPort != "" {
		cfg.Database.Port = parseInt(dbPort, cfg.Database.Port)
	}
	if dbUser := os.Getenv("TEST_DB_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPassword := os.Getenv("TEST_DB_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	if dbName := os.Getenv("TEST_DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}
	
	if redisHost := os.Getenv("TEST_REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("TEST_REDIS_PORT"); redisPort != "" {
		cfg.Redis.Port = parseInt(redisPort, cfg.Redis.Port)
	}
	if redisPassword := os.Getenv("TEST_REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}
	if redisDB := os.Getenv("TEST_REDIS_DB"); redisDB != "" {
		cfg.Redis.DB = parseInt(redisDB, cfg.Redis.DB)
	}
	
	if jwtSecret := os.Getenv("TEST_JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}
	
	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	// Simple integer parsing - in real code you'd use strconv.Atoi
	var result int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}