package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Security     SecurityConfig     `mapstructure:"security"`
	Storage      StorageConfig      `mapstructure:"storage"`
	AWS          AWSConfig          `mapstructure:"aws"`
	Stripe       StripeConfig       `mapstructure:"stripe"`
	Email        EmailConfig        `mapstructure:"email"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Cache        CacheConfig        `mapstructure:"cache"`
	PubSub       PubSubConfig       `mapstructure:"pubsub"`
	Verification VerificationConfig `mapstructure:"verification"`
	Chat         ChatConfig         `mapstructure:"chat"`
	EphemeralPhoto EphemeralPhotoConfig `mapstructure:"ephemeral_photo"`
	Moderation  ModerationConfig  `mapstructure:"moderation"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
}

// AppConfig represents application configuration
type AppConfig struct {
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"db_name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
	Timezone        string `mapstructure:"timezone"`
	MigrationsPath  string `mapstructure:"migrations_path"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	Password           string        `mapstructure:"password"`
	DB                 int           `mapstructure:"db"`
	PoolSize           int           `mapstructure:"pool_size"`
	MinIdleConns       int           `mapstructure:"min_idle_conns"`
	MaxRetries         int           `mapstructure:"max_retries"`
	DialTimeout        time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	PoolTimeout        time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	IdleCheckFrequency time.Duration `mapstructure:"idle_check_frequency"`
	ClusterEnabled     bool          `mapstructure:"cluster_enabled"`
	ClusterAddresses   []string      `mapstructure:"cluster_addresses"`
	MaxRedirects       int           `mapstructure:"max_redirects"`
	RouteByLatency     bool          `mapstructure:"route_by_latency"`
	RouteRandomly      bool          `mapstructure:"route_randomly"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer             string `mapstructure:"issuer"`
	RefreshTokenRotation bool   `mapstructure:"refresh_token_rotation"`
	MaxActiveSessions   int    `mapstructure:"max_active_sessions"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	AccountLockoutEnabled    bool          `mapstructure:"account_lockout_enabled"`
	MaxFailedAttempts        int           `mapstructure:"max_failed_attempts"`
	AccountLockoutDuration  time.Duration `mapstructure:"account_lockout_duration"`
	PasswordMinLength        int           `mapstructure:"password_min_length"`
	PasswordRequireUppercase bool         `mapstructure:"password_require_uppercase"`
	PasswordRequireLowercase bool         `mapstructure:"password_require_lowercase"`
	PasswordRequireNumbers   bool         `mapstructure:"password_require_numbers"`
	PasswordRequireSymbols   bool         `mapstructure:"password_require_symbols"`
	SessionTimeout          time.Duration `mapstructure:"session_timeout"`
	DeviceFingerprinting    bool         `mapstructure:"device_fingerprinting"`
	CSRFProtection         bool         `mapstructure:"csrf_protection"`
}

// AWSConfig represents AWS configuration
type AWSConfig struct {
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	S3Bucket        string `mapstructure:"s3_bucket"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Provider       string        `mapstructure:"provider"`        // "s3" or "minio"
	Region         string        `mapstructure:"region"`
	AccessKeyID    string        `mapstructure:"access_key_id"`
	SecretAccessKey string        `mapstructure:"secret_access_key"`
	Bucket         string        `mapstructure:"bucket"`
	Endpoint       string        `mapstructure:"endpoint"`        // For MinIO
	UseSSL         bool          `mapstructure:"use_ssl"`
	UploadExpiry   time.Duration `mapstructure:"upload_expiry"`   // Signed URL expiry for uploads
	DownloadExpiry time.Duration `mapstructure:"download_expiry"` // Signed URL expiry for downloads
	MaxFileSize    int64         `mapstructure:"max_file_size"`  // Max file size in bytes
	AllowedTypes   []string      `mapstructure:"allowed_types"`  // Allowed file types
}

// StripeConfig represents Stripe configuration
type StripeConfig struct {
	// API Keys
	SecretKey      string `mapstructure:"secret_key"`
	PublishableKey string `mapstructure:"publishable_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	
	// Subscription Plans
	FreePlanID    string `mapstructure:"free_plan_id"`
	PremiumPlanID string `mapstructure:"premium_plan_id"`
	PlatinumPlanID string `mapstructure:"platinum_plan_id"`
	
	// Payment Settings
	DefaultCurrency string `mapstructure:"default_currency"`
	SuccessURL      string `mapstructure:"success_url"`
	CancelURL       string `mapstructure:"cancel_url"`
	
	// Webhook Settings
	WebhookEndpoint string `mapstructure:"webhook_endpoint"`
	
	// Security Settings
	EnableRadar      bool `mapstructure:"enable_radar"`
	FraudLevel       string `mapstructure:"fraud_level"`
	
	// Rate Limiting
	PaymentRateLimit int `mapstructure:"payment_rate_limit"`
	
	// Cache Settings
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
}

// EmailConfig represents email configuration
type EmailConfig struct {
	SendGridAPIKey string `mapstructure:"sendgrid_api_key"`
	FromEmail      string `mapstructure:"from_email"`
	FromName       string `mapstructure:"from_name"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	RequestsPerHour   int `mapstructure:"requests_per_hour"`
	
	// Discovery rate limits
	SwipesPerHour    int           `mapstructure:"swipes_per_hour"`
	SwipesPerDay     int           `mapstructure:"swipes_per_day"`
	SuperLikesPerDay  int           `mapstructure:"super_likes_per_day"`
	DiscoveryPerHour int           `mapstructure:"discovery_per_hour"`
	DiscoveryPerDay  int           `mapstructure:"discovery_per_day"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	UserProfileTTL           time.Duration `mapstructure:"user_profile_ttl"`
	PhotoMetadataTTL        time.Duration `mapstructure:"photo_metadata_ttl"`
	MatchRecommendationsTTL  time.Duration `mapstructure:"match_recommendations_ttl"`
	APIResponseTTL          time.Duration `mapstructure:"api_response_ttl"`
	GeoSpatialTTL           time.Duration `mapstructure:"geospatial_ttl"`
	OnlineStatusTTL         time.Duration `mapstructure:"online_status_ttl"`
	SessionTTL              time.Duration `mapstructure:"session_ttl"`
	MaxActiveSessions        int           `mapstructure:"max_active_sessions"`
	CleanupInterval          time.Duration `mapstructure:"cleanup_interval"`
	WarmupEnabled           bool          `mapstructure:"warmup_enabled"`
	WarmupConcurrency        int           `mapstructure:"warmup_concurrency"`
	WarmupBatchSize         int           `mapstructure:"warmup_batch_size"`
}

// PubSubConfig represents Pub/Sub configuration
type PubSubConfig struct {
	Enabled                 bool          `mapstructure:"enabled"`
	MaxConnectionsPerUser    int           `mapstructure:"max_connections_per_user"`
	ConnectionTimeout        time.Duration `mapstructure:"connection_timeout"`
	PingInterval           time.Duration `mapstructure:"ping_interval"`
	ReconnectInterval       time.Duration `mapstructure:"reconnect_interval"`
	MaxMessageSize         int           `mapstructure:"max_message_size"`
}

// VerificationConfig represents verification configuration
type VerificationConfig struct {
	// AI Service Configuration
	AIService AIServiceConfig `mapstructure:"ai_service"`
	
	// Verification Thresholds
	Thresholds VerificationThresholdsConfig `mapstructure:"thresholds"`
	
	// Verification Limits
	Limits VerificationLimitsConfig `mapstructure:"limits"`
	
	// Document Processing
	DocumentProcessing DocumentProcessingConfig `mapstructure:"document_processing"`
	
	// Security Settings
	Security VerificationSecurityConfig `mapstructure:"security"`
}

// AIServiceConfig represents AI service configuration
type AIServiceConfig struct {
	Provider           string `mapstructure:"provider"`            // "aws" or "mock"
	Region             string `mapstructure:"region"`
	AccessKeyID        string `mapstructure:"access_key_id"`
	SecretAccessKey    string `mapstructure:"secret_access_key"`
	FaceCollectionID    string `mapstructure:"face_collection_id"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold"` // Default: 0.85
	Enabled            bool   `mapstructure:"enabled"`
}

// VerificationThresholdsConfig represents verification thresholds
type VerificationThresholdsConfig struct {
	SelfieSimilarityThreshold    float64 `mapstructure:"selfie_similarity_threshold"`     // Default: 0.85
	DocumentConfidenceThreshold  float64 `mapstructure:"document_confidence_threshold"`   // Default: 0.80
	LivenessConfidenceThreshold  float64 `mapstructure:"liveness_confidence_threshold"`   // Default: 0.90
	NSFWThreshold               float64 `mapstructure:"nsfw_threshold"`                  // Default: 0.70
	ManualReviewThreshold        float64 `mapstructure:"manual_review_threshold"`        // Default: 0.75
}

// VerificationLimitsConfig represents verification limits
type VerificationLimitsConfig struct {
	MaxAttemptsPerDay      int           `mapstructure:"max_attempts_per_day"`       // Default: 3
	MaxAttemptsPerMonth    int           `mapstructure:"max_attempts_per_month"`     // Default: 10
	CooldownPeriod         time.Duration `mapstructure:"cooldown_period"`             // Default: 24h
	VerificationExpiry     time.Duration `mapstructure:"verification_expiry"`         // Default: 365d
	DocumentExpiry         time.Duration `mapstructure:"document_expiry"`             // Default: 730d
	MaxFileSize            int64         `mapstructure:"max_file_size"`              // Default: 10MB
	MaxSelfieFileSize      int64         `mapstructure:"max_selfie_file_size"`       // Default: 5MB
	MaxDocumentFileSize    int64         `mapstructure:"max_document_file_size"`     // Default: 10MB
}

// DocumentProcessingConfig represents document processing configuration
type DocumentProcessingConfig struct {
	OCRProvider           string `mapstructure:"ocr_provider"`              // "aws" or "mock"
	MinConfidence         float64 `mapstructure:"min_confidence"`            // Default: 0.80
	SupportedDocumentTypes []string `mapstructure:"supported_document_types"` // ["id_card", "passport", "driver_license"]
	ExtractionEnabled     bool   `mapstructure:"extraction_enabled"`
	ValidationEnabled     bool   `mapstructure:"validation_enabled"`
}

// VerificationSecurityConfig represents verification security configuration
type VerificationSecurityConfig struct {
	EncryptedStorage      bool `mapstructure:"encrypted_storage"`
	FraudDetectionEnabled bool `mapstructure:"fraud_detection_enabled"`
	IPTrackingEnabled     bool `mapstructure:"ip_tracking_enabled"`
	DeviceTrackingEnabled bool `mapstructure:"device_tracking_enabled"`
	RequireRecentPhoto    bool `mapstructure:"require_recent_photo"`     // Photo must be taken within last 24h
	MaxPhotoAge           time.Duration `mapstructure:"max_photo_age"`   // Default: 24h
}

// ChatConfig represents chat configuration
type ChatConfig struct {
	// WebSocket configuration
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	
	// Message configuration
	Message MessageConfig `mapstructure:"message"`
	
	// Security configuration
	Security ChatSecurityConfig `mapstructure:"security"`
	
	// Caching configuration
	Cache ChatCacheConfig `mapstructure:"cache"`
	
	// Rate limiting configuration
	RateLimit ChatRateLimitConfig `mapstructure:"rate_limit"`
}

// WebSocketConfig represents WebSocket configuration
type WebSocketConfig struct {
	Enabled                bool          `mapstructure:"enabled"`
	Path                   string        `mapstructure:"path"`
	AllowedOrigins         []string      `mapstructure:"allowed_origins"`
	PingInterval           time.Duration `mapstructure:"ping_interval"`
	PongWait               time.Duration `mapstructure:"pong_wait"`
	WriteWait              time.Duration `mapstructure:"write_wait"`
	MaxMessageSize         int64         `mapstructure:"max_message_size"`
	ReadBufferSize         int           `mapstructure:"read_buffer_size"`
	WriteBufferSize        int           `mapstructure:"write_buffer_size"`
	CompressionEnabled     bool          `mapstructure:"compression_enabled"`
	MaxConnectionsPerUser  int           `mapstructure:"max_connections_per_user"`
	ConnectionTimeout      time.Duration `mapstructure:"connection_timeout"`
	ReconnectInterval      time.Duration `mapstructure:"reconnect_interval"`
	HeartbeatInterval      time.Duration `mapstructure:"heartbeat_interval"`
}

// MessageConfig represents message configuration
type MessageConfig struct {
	// Message limits
	MaxTextLength          int           `mapstructure:"max_text_length"`
	MaxMessageAge          time.Duration `mapstructure:"max_message_age"`
	MaxMessagesPerRequest  int           `mapstructure:"max_messages_per_request"`
	
	// Message types
	AllowedMessageTypes    []string      `mapstructure:"allowed_message_types"`
	
	// Photo messages
	MaxPhotoSize           int64         `mapstructure:"max_photo_size"`
	AllowedPhotoTypes      []string      `mapstructure:"allowed_photo_types"`
	PhotoExpiry            time.Duration `mapstructure:"photo_expiry"`
	
	// Ephemeral photos
	EphemeralPhotoDuration time.Duration `mapstructure:"ephemeral_photo_duration"`
	
	// Location messages
	LocationAccuracy       float64       `mapstructure:"location_accuracy"`
	
	// System messages
	SystemMessagePrefix    string        `mapstructure:"system_message_prefix"`
	
	// Encryption
	EncryptionEnabled      bool          `mapstructure:"encryption_enabled"`
	EncryptionKey          string        `mapstructure:"encryption_key"`
}

// ChatSecurityConfig represents chat security configuration
type ChatSecurityConfig struct {
	// Content filtering
	ContentFilteringEnabled bool     `mapstructure:"content_filtering_enabled"`
	BannedWords            []string `mapstructure:"banned_words"`
	BannedPatterns         []string `mapstructure:"banned_patterns"`
	
	// Spam detection
	SpamDetectionEnabled   bool          `mapstructure:"spam_detection_enabled"`
	SpamThreshold          float64       `mapstructure:"spam_threshold"`
	SpamWindow             time.Duration `mapstructure:"spam_window"`
	
	// Rate limiting
	MessageRateLimit       int           `mapstructure:"message_rate_limit"`
	MessageRateWindow      time.Duration `mapstructure:"message_rate_window"`
	
	// Link security
	LinkPreviewEnabled     bool          `mapstructure:"link_preview_enabled"`
	AllowedLinkDomains     []string      `mapstructure:"allowed_link_domains"`
	BlockedLinkDomains     []string      `mapstructure:"blocked_link_domains"`
	
	// PII detection
	PIIDetectionEnabled     bool          `mapstructure:"pii_detection_enabled"`
	
	// Reporting
	ReportThreshold        int           `mapstructure:"report_threshold"`
	AutoBanThreshold       int           `mapstructure:"auto_ban_threshold"`
}

// ChatCacheConfig represents chat cache configuration
type ChatCacheConfig struct {
	// TTL settings
	UserOnlineStatusTTL    time.Duration `mapstructure:"user_online_status_ttl"`
	TypingIndicatorTTL     time.Duration `mapstructure:"typing_indicator_ttl"`
	ConversationTTL        time.Duration `mapstructure:"conversation_ttl"`
	MessageTTL             time.Duration `mapstructure:"message_ttl"`
	UnreadCountTTL         time.Duration `mapstructure:"unread_count_ttl"`
	LinkPreviewTTL         time.Duration `mapstructure:"link_preview_ttl"`
	
	// Cache sizes
	MaxCachedConversations int           `mapstructure:"max_cached_conversations"`
	MaxCachedMessages      int           `mapstructure:"max_cached_messages"`
	
	// Cleanup settings
	CleanupInterval        time.Duration `mapstructure:"cleanup_interval"`
	
	// Pub/Sub settings
	PubSubEnabled          bool          `mapstructure:"pub_sub_enabled"`
}

// ChatRateLimitConfig represents chat rate limiting configuration
type ChatRateLimitConfig struct {
	// Message rate limits
	MessagesPerMinute      int           `mapstructure:"messages_per_minute"`
	MessagesPerHour        int           `mapstructure:"messages_per_hour"`
	MessagesPerDay         int           `mapstructure:"messages_per_day"`
	
	// Conversation rate limits
	ConversationsPerDay    int           `mapstructure:"conversations_per_day"`
	
	// Photo rate limits
	PhotosPerDay           int           `mapstructure:"photos_per_day"`
	
	// Connection rate limits
	ConnectionsPerMinute   int           `mapstructure:"connections_per_minute"`
	
	// Typing indicator rate limits
	TypingIndicatorsPerMinute int       `mapstructure:"typing_indicators_per_minute"`
}

// EphemeralPhotoConfig represents ephemeral photo configuration
type EphemeralPhotoConfig struct {
	// Photo settings
	MaxFileSize        int64         `mapstructure:"max_file_size"`         // Max file size in bytes
	AllowedTypes      []string      `mapstructure:"allowed_types"`         // Allowed file types
	MaxPhotosPerUser  int           `mapstructure:"max_photos_per_user"`   // Max photos per user
	
	// Expiration settings
	DefaultDuration   time.Duration `mapstructure:"default_duration"`       // Default expiration time
	MaxDuration      time.Duration `mapstructure:"max_duration"`          // Maximum allowed duration
	ViewDuration     time.Duration `mapstructure:"view_duration"`          // Time after viewing before expiration
	
	// Security settings
	AccessKeyLength  int           `mapstructure:"access_key_length"`      // Length of access keys
	EnableWatermark  bool          `mapstructure:"enable_watermark"`       // Enable watermarking
	WatermarkText    string        `mapstructure:"watermark_text"`        // Watermark text
	PreventDownload bool          `mapstructure:"prevent_download"`      // Prevent downloads
	
	// Storage settings
	StorageTier      string        `mapstructure:"storage_tier"`          // Storage tier (hot/cold)
	CleanupInterval  time.Duration `mapstructure:"cleanup_interval"`      // Cleanup interval for expired photos
	RetentionPeriod  time.Duration `mapstructure:"retention_period"`      // How long to keep deleted photos
	
	// Caching settings
	CacheTTL        time.Duration `mapstructure:"cache_ttl"`             // Cache TTL for photo metadata
	ViewCacheTTL     time.Duration `mapstructure:"view_cache_ttl"`        // Cache TTL for view status
	
	// Rate limiting
	UploadRateLimit  int           `mapstructure:"upload_rate_limit"`     // Uploads per hour
	ViewRateLimit    int           `mapstructure:"view_rate_limit"`       // Views per hour
	
	// Analytics settings
	EnableAnalytics  bool          `mapstructure:"enable_analytics"`      // Enable analytics tracking
	AnalyticsTTL    time.Duration `mapstructure:"analytics_ttl"`         // Analytics data retention
	
	// Background job settings
	JobInterval     time.Duration `mapstructure:"job_interval"`          // Background job interval
	JobBatchSize    int           `mapstructure:"job_batch_size"`        // Batch size for cleanup jobs
	EnableJobRetry  bool          `mapstructure:"enable_job_retry"`      // Enable job retry on failure
	MaxJobRetries   int           `mapstructure:"max_job_retries"`       // Max job retry attempts
}

// ModerationConfig represents moderation configuration
type ModerationConfig struct {
	// AI moderation configuration
	AIModeration AIModerationConfig `mapstructure:"ai_moderation"`
	
	// Content analysis configuration
	ContentAnalysis ContentAnalysisConfig `mapstructure:"content_analysis"`
	
	// Moderation rules configuration
	Rules ModerationRulesConfig `mapstructure:"rules"`
	
	// Appeal process configuration
	Appeal AppealConfig `mapstructure:"appeal"`
	
	// Moderation queue configuration
	Queue ModerationQueueConfig `mapstructure:"queue"`
	
	// Analytics configuration
	Analytics ModerationAnalyticsConfig `mapstructure:"analytics"`
	
	// Rate limiting configuration
	RateLimit ModerationRateLimitConfig `mapstructure:"rate_limit"`
}

// AIModerationConfig represents AI moderation configuration
type AIModerationConfig struct {
	// AWS Rekognition settings
	Enabled           bool    `mapstructure:"enabled"`
	Region            string  `mapstructure:"region"`
	AccessKeyID       string  `mapstructure:"access_key_id"`
	SecretAccessKey   string  `mapstructure:"secret_access_key"`
	Bucket            string  `mapstructure:"bucket"`
	
	// Confidence thresholds
	ConfidenceThreshold float64 `mapstructure:"confidence_threshold"`
	NSFWThreshold      float64 `mapstructure:"nsfw_threshold"`
	ViolenceThreshold  float64 `mapstructure:"violence_threshold"`
	AdultThreshold     float64 `mapstructure:"adult_threshold"`
	
	// Processing settings
	BatchSize    int           `mapstructure:"batch_size"`
	MaxRetries   int           `mapstructure:"max_retries"`
	RetryDelay   time.Duration `mapstructure:"retry_delay"`
	Timeout      time.Duration `mapstructure:"timeout"`
	
	// Fallback settings
	EnableFallback     bool    `mapstructure:"enable_fallback"`
	FallbackThreshold float64 `mapstructure:"fallback_threshold"`
}

// ContentAnalysisConfig represents content analysis configuration
type ContentAnalysisConfig struct {
	// Text analysis settings
	ProfanityFilterEnabled bool     `mapstructure:"profanity_filter_enabled"`
	BannedWords           []string `mapstructure:"banned_words"`
	BannedPatterns        []string `mapstructure:"banned_patterns"`
	
	// PII detection settings
	PIIDetectionEnabled bool     `mapstructure:"pii_detection_enabled"`
	PIIPatterns        []string `mapstructure:"pii_patterns"`
	
	// Link analysis settings
	LinkAnalysisEnabled bool     `mapstructure:"link_analysis_enabled"`
	AllowedDomains      []string `mapstructure:"allowed_domains"`
	BlockedDomains      []string `mapstructure:"blocked_domains"`
	
	// Image analysis settings
	ImageAnalysisEnabled bool    `mapstructure:"image_analysis_enabled"`
	MinConfidence        float64 `mapstructure:"min_confidence"`
	
	// Video analysis settings
	VideoAnalysisEnabled bool    `mapstructure:"video_analysis_enabled"`
	MaxVideoDuration     int     `mapstructure:"max_video_duration"`
	
	// Processing settings
	BatchSize    int           `mapstructure:"batch_size"`
	MaxRetries   int           `mapstructure:"max_retries"`
	RetryDelay   time.Duration `mapstructure:"retry_delay"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// ModerationRulesConfig represents moderation rules configuration
type ModerationRulesConfig struct {
	// Automated actions
	AutoBanEnabled      bool    `mapstructure:"auto_ban_enabled"`
	AutoBanThreshold    int     `mapstructure:"auto_ban_threshold"`
	AutoSuspendEnabled  bool    `mapstructure:"auto_suspend_enabled"`
	AutoSuspendThreshold int    `mapstructure:"auto_suspend_threshold"`
	
	// Reputation system
	ReputationEnabled   bool    `mapstructure:"reputation_enabled"`
	InitialReputation   int     `mapstructure:"initial_reputation"`
	MinReputation       int     `mapstructure:"min_reputation"`
	MaxReputation       int     `mapstructure:"max_reputation"`
	
	// Report thresholds
	ReportThreshold     int     `mapstructure:"report_threshold"`
	SeverityThreshold   int     `mapstructure:"severity_threshold"`
	
	// Custom rules
	CustomRules         []CustomRule `mapstructure:"custom_rules"`
}

// CustomRule represents a custom moderation rule
type CustomRule struct {
	Name        string                 `mapstructure:"name"`
	Description string                 `mapstructure:"description"`
	Conditions  map[string]interface{} `mapstructure:"conditions"`
	Actions     []string               `mapstructure:"actions"`
	Enabled     bool                   `mapstructure:"enabled"`
}

// AppealConfig represents appeal process configuration
type AppealConfig struct {
	// Appeal settings
	Enabled            bool          `mapstructure:"enabled"`
	MaxAppealsPerUser  int           `mapstructure:"max_appeals_per_user"`
	AppealWindow       time.Duration `mapstructure:"appeal_window"`
	
	// Review settings
	AutoReviewEnabled  bool    `mapstructure:"auto_review_enabled"`
	AutoReviewThreshold float64 `mapstructure:"auto_review_threshold"`
	
	// Notification settings
	NotifyOnSubmit     bool `mapstructure:"notify_on_submit"`
	NotifyOnReview     bool `mapstructure:"notify_on_review"`
}

// ModerationQueueConfig represents moderation queue configuration
type ModerationQueueConfig struct {
	// Queue settings
	Enabled           bool          `mapstructure:"enabled"`
	MaxQueueSize      int           `mapstructure:"max_queue_size"`
	ProcessingInterval time.Duration `mapstructure:"processing_interval"`
	
	// Priority settings
	PriorityLevels    []string      `mapstructure:"priority_levels"`
	DefaultPriority   string        `mapstructure:"default_priority"`
	
	// Assignment settings
	AutoAssignEnabled bool          `mapstructure:"auto_assign_enabled"`
	AssignmentTimeout time.Duration `mapstructure:"assignment_timeout"`
}

// ModerationAnalyticsConfig represents moderation analytics configuration
type ModerationAnalyticsConfig struct {
	// Analytics settings
	Enabled           bool          `mapstructure:"enabled"`
	RetentionPeriod   time.Duration `mapstructure:"retention_period"`
	AggregationInterval time.Duration `mapstructure:"aggregation_interval"`
	
	// Metrics settings
	TrackReports      bool `mapstructure:"track_reports"`
	TrackBans         bool `mapstructure:"track_bans"`
	TrackAppeals      bool `mapstructure:"track_appeals"`
	TrackContent      bool `mapstructure:"track_content"`
	TrackModerators   bool `mapstructure:"track_moderators"`
	
	// Export settings
	ExportEnabled     bool     `mapstructure:"export_enabled"`
	ExportFormats     []string `mapstructure:"export_formats"`
	ExportInterval    time.Duration `mapstructure:"export_interval"`
}

// ModerationRateLimitConfig represents moderation rate limiting configuration
type ModerationRateLimitConfig struct {
	// Report rate limits
	ReportsPerMinute  int           `mapstructure:"reports_per_minute"`
	ReportsPerHour    int           `mapstructure:"reports_per_hour"`
	ReportsPerDay     int           `mapstructure:"reports_per_day"`
	
	// Block rate limits
	BlocksPerMinute   int           `mapstructure:"blocks_per_minute"`
	BlocksPerHour     int           `mapstructure:"blocks_per_hour"`
	BlocksPerDay      int           `mapstructure:"blocks_per_day"`
	
	// Appeal rate limits
	AppealsPerMinute  int           `mapstructure:"appeals_per_minute"`
	AppealsPerHour    int           `mapstructure:"appeals_per_hour"`
	AppealsPerDay     int           `mapstructure:"appeals_per_day"`
	
	// Admin rate limits
	AdminActionsPerMinute int       `mapstructure:"admin_actions_per_minute"`
	AdminActionsPerHour   int       `mapstructure:"admin_actions_per_hour"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	// Health check configuration
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	
	// Metrics configuration
	Metrics MetricsConfig `mapstructure:"metrics"`
	
	// Alerting configuration
	Alerting AlertingConfig `mapstructure:"alerting"`
	
	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`
	
	// Background jobs configuration
	BackgroundJobs BackgroundJobsConfig `mapstructure:"background_jobs"`
	
	// Storage configuration
	Storage MonitoringStorageConfig `mapstructure:"storage"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	// Check intervals
	Interval        time.Duration `mapstructure:"interval"`
	Timeout         time.Duration `mapstructure:"timeout"`
	
	// Component checks
	DatabaseEnabled bool `mapstructure:"database_enabled"`
	RedisEnabled    bool `mapstructure:"redis_enabled"`
	StorageEnabled  bool `mapstructure:"storage_enabled"`
	ExternalEnabled bool `mapstructure:"external_enabled"`
	
	// Thresholds
	DatabaseThreshold int `mapstructure:"database_threshold"`   // Max response time in ms
	RedisThreshold    int `mapstructure:"redis_threshold"`      // Max response time in ms
	StorageThreshold  int `mapstructure:"storage_threshold"`    // Max response time in ms
	
	// System resource checks
	CPUThreshold    float64 `mapstructure:"cpu_threshold"`     // CPU usage percentage
	MemoryThreshold  float64 `mapstructure:"memory_threshold"`  // Memory usage percentage
	DiskThreshold   float64 `mapstructure:"disk_threshold"`    // Disk usage percentage
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	// Collection settings
	Enabled         bool          `mapstructure:"enabled"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
	RetentionPeriod  time.Duration `mapstructure:"retention_period"`
	
	// HTTP metrics
	HTTPMetricsEnabled bool `mapstructure:"http_metrics_enabled"`
	
	// Database metrics
	DatabaseMetricsEnabled bool `mapstructure:"database_metrics_enabled"`
	
	// Cache metrics
	CacheMetricsEnabled bool `mapstructure:"cache_metrics_enabled"`
	
	// Business metrics
	BusinessMetricsEnabled bool `mapstructure:"business_metrics_enabled"`
	
	// System metrics
	SystemMetricsEnabled bool `mapstructure:"system_metrics_enabled"`
	
	// Prometheus settings
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Path        string `mapstructure:"path"`
	Port        int    `mapstructure:"port"`
	Namespace   string `mapstructure:"namespace"`
	Subsystem   string `mapstructure:"subsystem"`
}

// AlertingConfig represents alerting configuration
type AlertingConfig struct {
	// General settings
	Enabled bool `mapstructure:"enabled"`
	
	// Thresholds
	ErrorRateThreshold      float64 `mapstructure:"error_rate_threshold"`      // Error rate percentage
	ResponseTimeThreshold   int     `mapstructure:"response_time_threshold"`    // Response time in ms
	CPUUsageThreshold      float64 `mapstructure:"cpu_usage_threshold"`       // CPU usage percentage
	MemoryUsageThreshold   float64 `mapstructure:"memory_usage_threshold"`    // Memory usage percentage
	DiskUsageThreshold     float64 `mapstructure:"disk_usage_threshold"`      // Disk usage percentage
	
	// Notification settings
	NotificationChannels []NotificationChannel `mapstructure:"notification_channels"`
	
	// Alert rules
	Rules []AlertRule `mapstructure:"rules"`
}

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	Type     string                 `mapstructure:"type"`     // email, slack, webhook, etc.
	Enabled  bool                   `mapstructure:"enabled"`
	Config   map[string]interface{} `mapstructure:"config"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	Name        string                 `mapstructure:"name"`
	Description string                 `mapstructure:"description"`
	Enabled     bool                   `mapstructure:"enabled"`
	Condition   string                 `mapstructure:"condition"`
	Threshold   float64                `mapstructure:"threshold"`
	Duration    time.Duration          `mapstructure:"duration"`
	Severity    string                 `mapstructure:"severity"`    // critical, warning, info
	Labels      map[string]string      `mapstructure:"labels"`
	Annotations map[string]string      `mapstructure:"annotations"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	// Structured logging
	StructuredEnabled bool `mapstructure:"structured_enabled"`
	
	// Correlation IDs
	CorrelationIDEnabled bool `mapstructure:"correlation_id_enabled"`
	
	// Performance logging
	PerformanceLoggingEnabled bool `mapstructure:"performance_logging_enabled"`
	
	// Security event logging
	SecurityLoggingEnabled bool `mapstructure:"security_logging_enabled"`
	
	// Log aggregation
	AggregationEnabled bool `mapstructure:"aggregation_enabled"`
	
	// Log rotation and retention
	RotationEnabled  bool          `mapstructure:"rotation_enabled"`
	RetentionPeriod  time.Duration `mapstructure:"retention_period"`
	MaxFileSize      int64         `mapstructure:"max_file_size"`
	MaxBackups       int           `mapstructure:"max_backups"`
}

// BackgroundJobsConfig represents background jobs configuration
type BackgroundJobsConfig struct {
	// Health check jobs
	HealthCheckJobInterval time.Duration `mapstructure:"health_check_job_interval"`
	
	// Metrics aggregation jobs
	MetricsAggregationInterval time.Duration `mapstructure:"metrics_aggregation_interval"`
	
	// Log cleanup jobs
	LogCleanupInterval time.Duration `mapstructure:"log_cleanup_interval"`
	
	// System resource monitoring jobs
	SystemMonitoringInterval time.Duration `mapstructure:"system_monitoring_interval"`
	
	// External service health monitoring jobs
	ExternalServiceCheckInterval time.Duration `mapstructure:"external_service_check_interval"`
}

// MonitoringStorageConfig represents monitoring storage configuration
type MonitoringStorageConfig struct {
	// Metrics storage
	MetricsStorageType string        `mapstructure:"metrics_storage_type"` // redis, memory, file
	MetricsRetention   time.Duration `mapstructure:"metrics_retention"`
	
	// Alert state storage
	AlertStorageType string        `mapstructure:"alert_storage_type"` // redis, memory, file
	AlertRetention   time.Duration `mapstructure:"alert_retention"`
	
	// Historical data storage
	HistoricalDataEnabled bool          `mapstructure:"historical_data_enabled"`
	HistoricalRetention  time.Duration `mapstructure:"historical_retention"`
	
	// Backup settings
	BackupEnabled  bool          `mapstructure:"backup_enabled"`
	BackupInterval time.Duration `mapstructure:"backup_interval"`
	BackupLocation string        `mapstructure:"backup_location"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()

	// Read environment file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("app.host", "localhost")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "dating_user")
	viper.SetDefault("database.password", "dating_pass")
	viper.SetDefault("database.db_name", "dating_db")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 3600) // 1 hour in seconds
	viper.SetDefault("database.conn_max_idle_time", 300) // 5 minutes in seconds
	viper.SetDefault("database.timezone", "UTC")
	viper.SetDefault("database.migrations_path", "./migrations")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.pool_timeout", "4s")
	viper.SetDefault("redis.idle_timeout", "5m")
	viper.SetDefault("redis.idle_check_frequency", "1m")
	viper.SetDefault("redis.cluster_enabled", false)
	viper.SetDefault("redis.cluster_addresses", []string{})
	viper.SetDefault("redis.max_redirects", 3)
	viper.SetDefault("redis.route_by_latency", false)
	viper.SetDefault("redis.route_randomly", false)

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-super-secret-jwt-key")
	viper.SetDefault("jwt.access_token_expiry", "15m")
	viper.SetDefault("jwt.refresh_token_expiry", "168h")
	viper.SetDefault("jwt.issuer", "winkr-backend")
	viper.SetDefault("jwt.refresh_token_rotation", true)
	viper.SetDefault("jwt.max_active_sessions", 5)

	// Security defaults
	viper.SetDefault("security.account_lockout_enabled", true)
	viper.SetDefault("security.max_failed_attempts", 5)
	viper.SetDefault("security.account_lockout_duration", "15m")
	viper.SetDefault("security.password_min_length", 8)
	viper.SetDefault("security.password_require_uppercase", true)
	viper.SetDefault("security.password_require_lowercase", true)
	viper.SetDefault("security.password_require_numbers", true)
	viper.SetDefault("security.password_require_symbols", false)
	viper.SetDefault("security.session_timeout", "168h")
	viper.SetDefault("security.device_fingerprinting", true)
	viper.SetDefault("security.csrf_protection", true)

	// AWS defaults
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.s3_bucket", "dating-app-dev")

	// Storage defaults
	viper.SetDefault("storage.provider", "minio")
	viper.SetDefault("storage.region", "us-east-1")
	viper.SetDefault("storage.bucket", "dating-app-photos")
	viper.SetDefault("storage.endpoint", "localhost:9000")
	viper.SetDefault("storage.use_ssl", false)
	viper.SetDefault("storage.upload_expiry", "15m")
	viper.SetDefault("storage.download_expiry", "1h")
	viper.SetDefault("storage.max_file_size", 5242880) // 5MB in bytes
	viper.SetDefault("storage.allowed_types", []string{"image/jpeg", "image/png", "image/webp"})

	// Stripe defaults
	viper.SetDefault("stripe.secret_key", "")
	viper.SetDefault("stripe.publishable_key", "")
	viper.SetDefault("stripe.webhook_secret", "")
	viper.SetDefault("stripe.free_plan_id", "price_free")
	viper.SetDefault("stripe.premium_plan_id", "price_premium")
	viper.SetDefault("stripe.platinum_plan_id", "price_platinum")
	viper.SetDefault("stripe.default_currency", "usd")
	viper.SetDefault("stripe.success_url", "/payment/success")
	viper.SetDefault("stripe.cancel_url", "/payment/cancel")
	viper.SetDefault("stripe.webhook_endpoint", "/api/v1/payment/webhook")
	viper.SetDefault("stripe.enable_radar", true)
	viper.SetDefault("stripe.fraud_level", "normal")
	viper.SetDefault("stripe.payment_rate_limit", 10)
	viper.SetDefault("stripe.cache_ttl", "15m")

	// Rate limiting defaults
	viper.SetDefault("rate_limit.requests_per_minute", 1000)
	viper.SetDefault("rate_limit.requests_per_hour", 10000)
	
	// Discovery rate limits defaults
	viper.SetDefault("rate_limit.swipes_per_hour", 100)
	viper.SetDefault("rate_limit.swipes_per_day", 1000)
	viper.SetDefault("rate_limit.super_likes_per_day", 5)
	viper.SetDefault("rate_limit.discovery_per_hour", 50)
	viper.SetDefault("rate_limit.discovery_per_day", 500)

	// Cache defaults
	viper.SetDefault("cache.user_profile_ttl", "30m")
	viper.SetDefault("cache.photo_metadata_ttl", "15m")
	viper.SetDefault("cache.match_recommendations_ttl", "10m")
	viper.SetDefault("cache.api_response_ttl", "5m")
	viper.SetDefault("cache.geospatial_ttl", "60m")
	viper.SetDefault("cache.online_status_ttl", "2m")
	viper.SetDefault("cache.session_ttl", "24h")
	viper.SetDefault("cache.max_active_sessions", 5)
	viper.SetDefault("cache.cleanup_interval", "1h")
	viper.SetDefault("cache.warmup_enabled", false)
	viper.SetDefault("cache.warmup_concurrency", 5)
	viper.SetDefault("cache.warmup_batch_size", 100)

	// Pub/Sub defaults
	viper.SetDefault("pubsub.enabled", true)
	viper.SetDefault("pubsub.max_connections_per_user", 5)
	viper.SetDefault("pubsub.connection_timeout", "30s")
	viper.SetDefault("pubsub.ping_interval", "30s")
	viper.SetDefault("pubsub.reconnect_interval", "5s")
	viper.SetDefault("pubsub.max_message_size", 1024)

	// Verification defaults
	// AI Service defaults
	viper.SetDefault("verification.ai_service.provider", "aws")
	viper.SetDefault("verification.ai_service.region", "us-east-1")
	viper.SetDefault("verification.ai_service.face_collection_id", "dating-app-faces")
	viper.SetDefault("verification.ai_service.similarity_threshold", 0.85)
	viper.SetDefault("verification.ai_service.enabled", true)

	// Verification thresholds defaults
	viper.SetDefault("verification.thresholds.selfie_similarity_threshold", 0.85)
	viper.SetDefault("verification.thresholds.document_confidence_threshold", 0.80)
	viper.SetDefault("verification.thresholds.liveness_confidence_threshold", 0.90)
	viper.SetDefault("verification.thresholds.nsfw_threshold", 0.70)
	viper.SetDefault("verification.thresholds.manual_review_threshold", 0.75)

	// Verification limits defaults
	viper.SetDefault("verification.limits.max_attempts_per_day", 3)
	viper.SetDefault("verification.limits.max_attempts_per_month", 10)
	viper.SetDefault("verification.limits.cooldown_period", "24h")
	viper.SetDefault("verification.limits.verification_expiry", "8760h") // 365 days
	viper.SetDefault("verification.limits.document_expiry", "17520h")   // 730 days
	viper.SetDefault("verification.limits.max_file_size", 10485760)      // 10MB in bytes
	viper.SetDefault("verification.limits.max_selfie_file_size", 5242880) // 5MB in bytes
	viper.SetDefault("verification.limits.max_document_file_size", 10485760) // 10MB in bytes

	// Document processing defaults
	viper.SetDefault("verification.document_processing.ocr_provider", "aws")
	viper.SetDefault("verification.document_processing.min_confidence", 0.80)
	viper.SetDefault("verification.document_processing.supported_document_types", []string{"id_card", "passport", "driver_license"})
	viper.SetDefault("verification.document_processing.extraction_enabled", true)
	viper.SetDefault("verification.document_processing.validation_enabled", true)

	// Verification security defaults
	viper.SetDefault("verification.security.encrypted_storage", true)
	viper.SetDefault("verification.security.fraud_detection_enabled", true)
	viper.SetDefault("verification.security.ip_tracking_enabled", true)
	viper.SetDefault("verification.security.device_tracking_enabled", true)
	viper.SetDefault("verification.security.require_recent_photo", true)
	viper.SetDefault("verification.security.max_photo_age", "24h")

	// Chat defaults
	// WebSocket defaults
	viper.SetDefault("chat.websocket.enabled", true)
	viper.SetDefault("chat.websocket.path", "/ws")
	viper.SetDefault("chat.websocket.allowed_origins", []string{"*"})
	viper.SetDefault("chat.websocket.ping_interval", "30s")
	viper.SetDefault("chat.websocket.pong_wait", "60s")
	viper.SetDefault("chat.websocket.write_wait", "10s")
	viper.SetDefault("chat.websocket.max_message_size", 32768) // 32KB
	viper.SetDefault("chat.websocket.read_buffer_size", 1024)
	viper.SetDefault("chat.websocket.write_buffer_size", 1024)
	viper.SetDefault("chat.websocket.compression_enabled", true)
	viper.SetDefault("chat.websocket.max_connections_per_user", 5)
	viper.SetDefault("chat.websocket.connection_timeout", "30s")
	viper.SetDefault("chat.websocket.reconnect_interval", "5s")
	viper.SetDefault("chat.websocket.heartbeat_interval", "30s")

	// Message defaults
	viper.SetDefault("chat.message.max_text_length", 2000)
	viper.SetDefault("chat.message.max_message_age", "8760h") // 365 days
	viper.SetDefault("chat.message.max_messages_per_request", 50)
	viper.SetDefault("chat.message.allowed_message_types", []string{"text", "photo", "photo_ephemeral", "location", "system", "gift"})
	viper.SetDefault("chat.message.max_photo_size", 10485760) // 10MB
	viper.SetDefault("chat.message.allowed_photo_types", []string{"image/jpeg", "image/png", "image/webp"})
	viper.SetDefault("chat.message.photo_expiry", "8760h") // 365 days
	viper.SetDefault("chat.message.ephemeral_photo_duration", "10s")
	viper.SetDefault("chat.message.location_accuracy", 100.0) // 100 meters
	viper.SetDefault("chat.message.system_message_prefix", "[System]")
	viper.SetDefault("chat.message.encryption_enabled", false)
	viper.SetDefault("chat.message.encryption_key", "")

	// Chat security defaults
	viper.SetDefault("chat.security.content_filtering_enabled", true)
	viper.SetDefault("chat.security.banned_words", []string{})
	viper.SetDefault("chat.security.banned_patterns", []string{})
	viper.SetDefault("chat.security.spam_detection_enabled", true)
	viper.SetDefault("chat.security.spam_threshold", 0.8)
	viper.SetDefault("chat.security.spam_window", "1m")
	viper.SetDefault("chat.security.message_rate_limit", 30)
	viper.SetDefault("chat.security.message_rate_window", "1m")
	viper.SetDefault("chat.security.link_preview_enabled", true)
	viper.SetDefault("chat.security.allowed_link_domains", []string{})
	viper.SetDefault("chat.security.blocked_link_domains", []string{})
	viper.SetDefault("chat.security.pii_detection_enabled", true)
	viper.SetDefault("chat.security.report_threshold", 3)
	viper.SetDefault("chat.security.auto_ban_threshold", 10)

	// Chat cache defaults
	viper.SetDefault("chat.cache.user_online_status_ttl", "2m")
	viper.SetDefault("chat.cache.typing_indicator_ttl", "10s")
	viper.SetDefault("chat.cache.conversation_ttl", "30m")
	viper.SetDefault("chat.cache.message_ttl", "1h")
	viper.SetDefault("chat.cache.unread_count_ttl", "5m")
	viper.SetDefault("chat.cache.link_preview_ttl", "24h")
	viper.SetDefault("chat.cache.max_cached_conversations", 100)
	viper.SetDefault("chat.cache.max_cached_messages", 1000)
	viper.SetDefault("chat.cache.cleanup_interval", "1h")
	viper.SetDefault("chat.cache.pub_sub_enabled", true)

	// Chat rate limit defaults
	viper.SetDefault("chat.rate_limit.messages_per_minute", 30)
	viper.SetDefault("chat.rate_limit.messages_per_hour", 500)
	viper.SetDefault("chat.rate_limit.messages_per_day", 2000)
	viper.SetDefault("chat.rate_limit.conversations_per_day", 50)
	viper.SetDefault("chat.rate_limit.photos_per_day", 20)
	viper.SetDefault("chat.rate_limit.connections_per_minute", 10)
	viper.SetDefault("chat.rate_limit.typing_indicators_per_minute", 20)

	// Ephemeral photo defaults
	viper.SetDefault("ephemeral_photo.max_file_size", 5242880) // 5MB in bytes
	viper.SetDefault("ephemeral_photo.allowed_types", []string{"image/jpeg", "image/png", "image/webp"})
	viper.SetDefault("ephemeral_photo.max_photos_per_user", 10)
	viper.SetDefault("ephemeral_photo.default_duration", "30s")
	viper.SetDefault("ephemeral_photo.max_duration", "300s") // 5 minutes
	viper.SetDefault("ephemeral_photo.view_duration", "30s")
	viper.SetDefault("ephemeral_photo.access_key_length", 32)
	viper.SetDefault("ephemeral_photo.enable_watermark", true)
	viper.SetDefault("ephemeral_photo.watermark_text", "Ephemeral")
	viper.SetDefault("ephemeral_photo.prevent_download", true)
	viper.SetDefault("ephemeral_photo.storage_tier", "hot")
	viper.SetDefault("ephemeral_photo.cleanup_interval", "5m")
	viper.SetDefault("ephemeral_photo.retention_period", "24h")
	viper.SetDefault("ephemeral_photo.cache_ttl", "1m")
	viper.SetDefault("ephemeral_photo.view_cache_ttl", "30s")
	viper.SetDefault("ephemeral_photo.upload_rate_limit", 10)
	viper.SetDefault("ephemeral_photo.view_rate_limit", 50)
	viper.SetDefault("ephemeral_photo.enable_analytics", true)
	viper.SetDefault("ephemeral_photo.analytics_ttl", "168h") // 7 days
	viper.SetDefault("ephemeral_photo.job_interval", "1m")
	viper.SetDefault("ephemeral_photo.job_batch_size", 100)
	viper.SetDefault("ephemeral_photo.enable_job_retry", true)
	viper.SetDefault("ephemeral_photo.max_job_retries", 3)

	// Moderation defaults
	// AI moderation defaults
	viper.SetDefault("moderation.ai_moderation.enabled", true)
	viper.SetDefault("moderation.ai_moderation.region", "us-east-1")
	viper.SetDefault("moderation.ai_moderation.bucket", "dating-app-moderation")
	viper.SetDefault("moderation.ai_moderation.confidence_threshold", 0.75)
	viper.SetDefault("moderation.ai_moderation.nsfw_threshold", 0.70)
	viper.SetDefault("moderation.ai_moderation.violence_threshold", 0.80)
	viper.SetDefault("moderation.ai_moderation.adult_threshold", 0.75)
	viper.SetDefault("moderation.ai_moderation.batch_size", 10)
	viper.SetDefault("moderation.ai_moderation.max_retries", 3)
	viper.SetDefault("moderation.ai_moderation.retry_delay", "1s")
	viper.SetDefault("moderation.ai_moderation.timeout", "30s")
	viper.SetDefault("moderation.ai_moderation.enable_fallback", true)
	viper.SetDefault("moderation.ai_moderation.fallback_threshold", 0.60)

	// Content analysis defaults
	viper.SetDefault("moderation.content_analysis.profanity_filter_enabled", true)
	viper.SetDefault("moderation.content_analysis.banned_words", []string{})
	viper.SetDefault("moderation.content_analysis.banned_patterns", []string{})
	viper.SetDefault("moderation.content_analysis.pii_detection_enabled", true)
	viper.SetDefault("moderation.content_analysis.pii_patterns", []string{
		`\b\d{3}-\d{2}-\d{4}\b`, // SSN
		`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`, // Credit card
		`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, // Email
		`\b\d{3}-\d{3}-\d{4}\b`, // Phone number
	})
	viper.SetDefault("moderation.content_analysis.link_analysis_enabled", true)
	viper.SetDefault("moderation.content_analysis.allowed_domains", []string{})
	viper.SetDefault("moderation.content_analysis.blocked_domains", []string{})
	viper.SetDefault("moderation.content_analysis.image_analysis_enabled", true)
	viper.SetDefault("moderation.content_analysis.min_confidence", 0.70)
	viper.SetDefault("moderation.content_analysis.video_analysis_enabled", true)
	viper.SetDefault("moderation.content_analysis.max_video_duration", 300) // 5 minutes
	viper.SetDefault("moderation.content_analysis.batch_size", 5)
	viper.SetDefault("moderation.content_analysis.max_retries", 3)
	viper.SetDefault("moderation.content_analysis.retry_delay", "1s")
	viper.SetDefault("moderation.content_analysis.timeout", "30s")

	// Moderation rules defaults
	viper.SetDefault("moderation.rules.auto_ban_enabled", true)
	viper.SetDefault("moderation.rules.auto_ban_threshold", 10)
	viper.SetDefault("moderation.rules.auto_suspend_enabled", true)
	viper.SetDefault("moderation.rules.auto_suspend_threshold", 5)
	viper.SetDefault("moderation.rules.reputation_enabled", true)
	viper.SetDefault("moderation.rules.initial_reputation", 100)
	viper.SetDefault("moderation.rules.min_reputation", 0)
	viper.SetDefault("moderation.rules.max_reputation", 1000)
	viper.SetDefault("moderation.rules.report_threshold", 3)
	viper.SetDefault("moderation.rules.severity_threshold", 7)
	viper.SetDefault("moderation.rules.custom_rules", []CustomRule{})

	// Appeal process defaults
	viper.SetDefault("moderation.appeal.enabled", true)
	viper.SetDefault("moderation.appeal.max_appeals_per_user", 3)
	viper.SetDefault("moderation.appeal.appeal_window", "168h") // 7 days
	viper.SetDefault("moderation.appeal.auto_review_enabled", false)
	viper.SetDefault("moderation.appeal.auto_review_threshold", 0.90)
	viper.SetDefault("moderation.appeal.notify_on_submit", true)
	viper.SetDefault("moderation.appeal.notify_on_review", true)

	// Moderation queue defaults
	viper.SetDefault("moderation.queue.enabled", true)
	viper.SetDefault("moderation.queue.max_queue_size", 10000)
	viper.SetDefault("moderation.queue.processing_interval", "1m")
	viper.SetDefault("moderation.queue.priority_levels", []string{"low", "medium", "high", "critical"})
	viper.SetDefault("moderation.queue.default_priority", "medium")
	viper.SetDefault("moderation.queue.auto_assign_enabled", true)
	viper.SetDefault("moderation.queue.assignment_timeout", "30m")

	// Analytics defaults
	viper.SetDefault("moderation.analytics.enabled", true)
	viper.SetDefault("moderation.analytics.retention_period", "8760h") // 365 days
	viper.SetDefault("moderation.analytics.aggregation_interval", "1h")
	viper.SetDefault("moderation.analytics.track_reports", true)
	viper.SetDefault("moderation.analytics.track_bans", true)
	viper.SetDefault("moderation.analytics.track_appeals", true)
	viper.SetDefault("moderation.analytics.track_content", true)
	viper.SetDefault("moderation.analytics.track_moderators", true)
	viper.SetDefault("moderation.analytics.export_enabled", true)
	viper.SetDefault("moderation.analytics.export_formats", []string{"json", "csv"})
	viper.SetDefault("moderation.analytics.export_interval", "24h")

	// Rate limiting defaults
	viper.SetDefault("moderation.rate_limit.reports_per_minute", 5)
	viper.SetDefault("moderation.rate_limit.reports_per_hour", 50)
	viper.SetDefault("moderation.rate_limit.reports_per_day", 200)
	viper.SetDefault("moderation.rate_limit.blocks_per_minute", 10)
	viper.SetDefault("moderation.rate_limit.blocks_per_hour", 100)
	viper.SetDefault("moderation.rate_limit.blocks_per_day", 500)
	viper.SetDefault("moderation.rate_limit.appeals_per_minute", 2)
	viper.SetDefault("moderation.rate_limit.appeals_per_hour", 10)
	viper.SetDefault("moderation.rate_limit.appeals_per_day", 20)
	viper.SetDefault("moderation.rate_limit.admin_actions_per_minute", 20)
	viper.SetDefault("moderation.rate_limit.admin_actions_per_hour", 500)

	// Monitoring defaults
	// Health check defaults
	viper.SetDefault("monitoring.health_check.interval", "30s")
	viper.SetDefault("monitoring.health_check.timeout", "10s")
	viper.SetDefault("monitoring.health_check.database_enabled", true)
	viper.SetDefault("monitoring.health_check.redis_enabled", true)
	viper.SetDefault("monitoring.health_check.storage_enabled", true)
	viper.SetDefault("monitoring.health_check.external_enabled", true)
	viper.SetDefault("monitoring.health_check.database_threshold", 1000) // 1000ms
	viper.SetDefault("monitoring.health_check.redis_threshold", 500)     // 500ms
	viper.SetDefault("monitoring.health_check.storage_threshold", 2000)   // 2000ms
	viper.SetDefault("monitoring.health_check.cpu_threshold", 80.0)      // 80%
	viper.SetDefault("monitoring.health_check.memory_threshold", 85.0)    // 85%
	viper.SetDefault("monitoring.health_check.disk_threshold", 90.0)      // 90%

	// Metrics defaults
	viper.SetDefault("monitoring.metrics.enabled", true)
	viper.SetDefault("monitoring.metrics.collection_interval", "60s")
	viper.SetDefault("monitoring.metrics.retention_period", "168h") // 7 days
	viper.SetDefault("monitoring.metrics.http_metrics_enabled", true)
	viper.SetDefault("monitoring.metrics.database_metrics_enabled", true)
	viper.SetDefault("monitoring.metrics.cache_metrics_enabled", true)
	viper.SetDefault("monitoring.metrics.business_metrics_enabled", true)
	viper.SetDefault("monitoring.metrics.system_metrics_enabled", true)
	
	// Prometheus defaults
	viper.SetDefault("monitoring.metrics.prometheus.enabled", true)
	viper.SetDefault("monitoring.metrics.prometheus.path", "/metrics")
	viper.SetDefault("monitoring.metrics.prometheus.port", 9090)
	viper.SetDefault("monitoring.metrics.prometheus.namespace", "winkr")
	viper.SetDefault("monitoring.metrics.prometheus.subsystem", "backend")

	// Alerting defaults
	viper.SetDefault("monitoring.alerting.enabled", true)
	viper.SetDefault("monitoring.alerting.error_rate_threshold", 5.0)    // 5%
	viper.SetDefault("monitoring.alerting.response_time_threshold", 2000)  // 2000ms
	viper.SetDefault("monitoring.alerting.cpu_usage_threshold", 85.0)     // 85%
	viper.SetDefault("monitoring.alerting.memory_usage_threshold", 90.0)  // 90%
	viper.SetDefault("monitoring.alerting.disk_usage_threshold", 95.0)    // 95%
	viper.SetDefault("monitoring.alerting.notification_channels", []NotificationChannel{})
	viper.SetDefault("monitoring.alerting.rules", []AlertRule{})

	// Logging defaults
	viper.SetDefault("monitoring.logging.structured_enabled", true)
	viper.SetDefault("monitoring.logging.correlation_id_enabled", true)
	viper.SetDefault("monitoring.logging.performance_logging_enabled", true)
	viper.SetDefault("monitoring.logging.security_logging_enabled", true)
	viper.SetDefault("monitoring.logging.aggregation_enabled", true)
	viper.SetDefault("monitoring.logging.rotation_enabled", true)
	viper.SetDefault("monitoring.logging.retention_period", "168h") // 7 days
	viper.SetDefault("monitoring.logging.max_file_size", 104857600) // 100MB
	viper.SetDefault("monitoring.logging.max_backups", 10)

	// Background jobs defaults
	viper.SetDefault("monitoring.background_jobs.health_check_job_interval", "30s")
	viper.SetDefault("monitoring.background_jobs.metrics_aggregation_interval", "5m")
	viper.SetDefault("monitoring.background_jobs.log_cleanup_interval", "1h")
	viper.SetDefault("monitoring.background_jobs.system_monitoring_interval", "60s")
	viper.SetDefault("monitoring.background_jobs.external_service_check_interval", "60s")

	// Monitoring storage defaults
	viper.SetDefault("monitoring.storage.metrics_storage_type", "redis")
	viper.SetDefault("monitoring.storage.metrics_retention", "168h") // 7 days
	viper.SetDefault("monitoring.storage.alert_storage_type", "redis")
	viper.SetDefault("monitoring.storage.alert_retention", "168h") // 7 days
	viper.SetDefault("monitoring.storage.historical_data_enabled", true)
	viper.SetDefault("monitoring.storage.historical_retention", "2160h") // 90 days
	viper.SetDefault("monitoring.storage.backup_enabled", false)
	viper.SetDefault("monitoring.storage.backup_interval", "24h")
	viper.SetDefault("monitoring.storage.backup_location", "/backups/monitoring")
}