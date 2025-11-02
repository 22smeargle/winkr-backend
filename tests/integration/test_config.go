package integration

import (
	"time"

	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestConfig returns configuration for integration tests
func TestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:  "test",
			Port:  8080,
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test_user",
			Password:        "test_password",
			DBName:          "test_db",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 300,
			Timezone:        "UTC",
			MigrationsPath:  "../../migrations",
		},
		Redis: config.RedisConfig{
			Host:               "localhost",
			Port:               6379,
			Password:           "",
			DB:                 1, // Use separate DB for tests
			PoolSize:           10,
			MinIdleConns:       5,
			MaxRetries:         3,
			DialTimeout:        5 * time.Second,
			ReadTimeout:        3 * time.Second,
			WriteTimeout:       3 * time.Second,
			PoolTimeout:        4 * time.Second,
			IdleTimeout:        5 * time.Minute,
			IdleCheckFrequency: 1 * time.Minute,
			ClusterEnabled:     false,
			ClusterAddresses:   []string{},
			MaxRedirects:       3,
			RouteByLatency:     false,
			RouteRandomly:      false,
		},
		JWT: config.JWTConfig{
			Secret:              "test-secret-key-for-integration-tests",
			AccessTokenExpiry:   15 * time.Minute,
			RefreshTokenExpiry:  24 * time.Hour,
			Issuer:              "test-issuer",
			RefreshTokenRotation: true,
			MaxActiveSessions:   5,
		},
		Security: config.SecurityConfig{
			AccountLockoutEnabled:    true,
			MaxFailedAttempts:        5,
			AccountLockoutDuration:  15 * time.Minute,
			PasswordMinLength:        8,
			PasswordRequireUppercase: true,
			PasswordRequireLowercase: true,
			PasswordRequireNumbers:   true,
			PasswordRequireSymbols:   false,
			SessionTimeout:          24 * time.Hour,
			DeviceFingerprinting:    true,
			CSRFProtection:         true,
		},
		Storage: config.StorageConfig{
			Provider:       "mock", // Use mock storage for tests
			Region:         "us-east-1",
			AccessKeyID:    "test-access-key",
			SecretAccessKey: "test-secret-key",
			Bucket:         "test-bucket",
			Endpoint:       "localhost:9000",
			UseSSL:         false,
			UploadExpiry:   15 * time.Minute,
			DownloadExpiry: 1 * time.Hour,
			MaxFileSize:    5 * 1024 * 1024, // 5MB
			AllowedTypes:   []string{"image/jpeg", "image/png", "image/webp"},
		},
		EphemeralPhoto: config.EphemeralPhotoConfig{
			MaxFileSize:        5 * 1024 * 1024, // 5MB
			AllowedTypes:      []string{"image/jpeg", "image/png", "image/webp"},
			MaxPhotosPerUser:  10,
			DefaultDuration:   30 * time.Second,
			MaxDuration:      300 * time.Second, // 5 minutes
			ViewDuration:     30 * time.Second,
			AccessKeyLength:  32,
			EnableWatermark:  true,
			WatermarkText:    "Test",
			PreventDownload:  true,
			StorageTier:      "hot",
			CleanupInterval:  5 * time.Minute,
			RetentionPeriod:  24 * time.Hour,
			CacheTTL:         1 * time.Minute,
			ViewCacheTTL:     30 * time.Second,
			UploadRateLimit:  10,
			ViewRateLimit:    50,
			EnableAnalytics:  true,
			AnalyticsTTL:     168 * time.Hour, // 7 days
			JobInterval:      1 * time.Minute,
			JobBatchSize:     100,
			EnableJobRetry:   true,
			MaxJobRetries:    3,
		},
		Chat: config.ChatConfig{
			WebSocket: config.WebSocketConfig{
				Enabled:                true,
				Path:                   "/ws",
				AllowedOrigins:         []string{"*"},
				PingInterval:           30 * time.Second,
				PongWait:               60 * time.Second,
				WriteWait:              10 * time.Second,
				MaxMessageSize:         32768, // 32KB
				ReadBufferSize:         1024,
				WriteBufferSize:        1024,
				CompressionEnabled:     true,
				MaxConnectionsPerUser:  5,
				ConnectionTimeout:      30 * time.Second,
				ReconnectInterval:      5 * time.Second,
				HeartbeatInterval:      30 * time.Second,
			},
			Message: config.MessageConfig{
				MaxTextLength:          2000,
				MaxMessageAge:          8760 * time.Hour, // 365 days
				MaxMessagesPerRequest:  50,
				AllowedMessageTypes:    []string{"text", "photo", "ephemeral_photo", "location", "system", "gift"},
				MaxPhotoSize:           10 * 1024 * 1024, // 10MB
				AllowedPhotoTypes:      []string{"image/jpeg", "image/png", "image/webp"},
				PhotoExpiry:            8760 * time.Hour, // 365 days
				EphemeralPhotoDuration: 10 * time.Second,
				LocationAccuracy:       100.0, // 100 meters
				SystemMessagePrefix:    "[System]",
				EncryptionEnabled:      false,
				EncryptionKey:          "",
			},
			Security: config.ChatSecurityConfig{
				ContentFilteringEnabled: true,
				BannedWords:            []string{},
				BannedPatterns:         []string{},
				SpamDetectionEnabled:   true,
				SpamThreshold:          0.8,
				SpamWindow:             1 * time.Minute,
				MessageRateLimit:       30,
				MessageRateWindow:      1 * time.Minute,
				LinkPreviewEnabled:     true,
				AllowedLinkDomains:     []string{},
				BlockedLinkDomains:     []string{},
				PIIDetectionEnabled:     true,
				ReportThreshold:        3,
				AutoBanThreshold:       10,
			},
			Cache: config.ChatCacheConfig{
				UserOnlineStatusTTL:    2 * time.Minute,
				TypingIndicatorTTL:     10 * time.Second,
				ConversationTTL:        30 * time.Minute,
				MessageTTL:             1 * time.Hour,
				UnreadCountTTL:         5 * time.Minute,
				LinkPreviewTTL:         24 * time.Hour,
				MaxCachedConversations: 100,
				MaxCachedMessages:      1000,
				CleanupInterval:        1 * time.Hour,
				PubSubEnabled:          true,
			},
			RateLimit: config.ChatRateLimitConfig{
				MessagesPerMinute:         30,
				MessagesPerHour:           500,
				MessagesPerDay:            2000,
				ConversationsPerDay:       50,
				PhotosPerDay:              20,
				ConnectionsPerMinute:      10,
				TypingIndicatorsPerMinute: 20,
			},
		},
	}
}

// TestDatabaseConfig returns database configuration for integration tests
func TestDatabaseConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "test_user",
		Password:        "test_password",
		DBName:          "test_db",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 300,
		Timezone:        "UTC",
		MigrationsPath:  "../../migrations",
	}
}

// TestRedisConfig returns Redis configuration for integration tests
func TestRedisConfig() config.RedisConfig {
	return config.RedisConfig{
		Host:               "localhost",
		Port:               6379,
		Password:           "",
		DB:                 1, // Use separate DB for tests
		PoolSize:           10,
		MinIdleConns:       5,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: 1 * time.Minute,
		ClusterEnabled:     false,
		ClusterAddresses:   []string{},
		MaxRedirects:       3,
		RouteByLatency:     false,
		RouteRandomly:      false,
	}
}