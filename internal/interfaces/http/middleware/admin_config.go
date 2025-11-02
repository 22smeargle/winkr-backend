package middleware

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AdminConfig represents admin-specific configuration
type AdminConfig struct {
	// Admin authentication configuration
	Auth *AdminAuthConfig `json:"auth"`
	
	// Admin rate limiting configuration
	RateLimit *AdminRateLimitConfig `json:"rate_limit"`
	
	// Admin permissions configuration
	Permissions *AdminPermissionsConfig `json:"permissions"`
	
	// Admin dashboard configuration
	Dashboard *AdminDashboardConfig `json:"dashboard"`
	
	// Admin security configuration
	Security *AdminSecurityConfig `json:"security"`
	
	// Admin audit configuration
	Audit *AdminAuditConfig `json:"audit"`
}

// AdminAuthConfig represents admin authentication configuration
type AdminAuthConfig struct {
	// JWT settings for admin tokens
	JWTSecret     string        `json:"jwt_secret"`
	JWTExpiry     time.Duration `json:"jwt_expiry"`
	RefreshExpiry  time.Duration `json:"refresh_expiry"`
	
	// Session settings
	SessionTimeout time.Duration `json:"session_timeout"`
	MaxSessions   int           `json:"max_sessions"`
	
	// Password requirements
	MinPasswordLength      int  `json:"min_password_length"`
	RequireUppercase      bool `json:"require_uppercase"`
	RequireLowercase      bool `json:"require_lowercase"`
	RequireNumbers        bool `json:"require_numbers"`
	RequireSymbols        bool `json:"require_symbols"`
	
	// Two-factor authentication
	Require2FA        bool   `json:"require_2fa"`
	2FAIssuer         string  `json:"2fa_issuer"`
	2FAWindow         int     `json:"2fa_window"`
	
	// Login attempts
	MaxLoginAttempts    int           `json:"max_login_attempts"`
	LockoutDuration    time.Duration `json:"lockout_duration"`
	
	// Trusted IPs
	TrustedIPs        []string `json:"trusted_ips"`
	RequireTrustedIP   bool     `json:"require_trusted_ip"`
}

// AdminRateLimitConfig represents admin rate limiting configuration
type AdminRateLimitConfig struct {
	// General admin rate limits
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	RequestsPerDay    int `json:"requests_per_day"`
	
	// Specific endpoint rate limits
	UserManagementRateLimit    int `json:"user_management_rate_limit"`
	ContentModerationRateLimit int `json:"content_moderation_rate_limit"`
	SystemManagementRateLimit int `json:"system_management_rate_limit"`
	
	// Burst settings
	BurstSize int `json:"burst_size"`
	
	// Whitelist for rate limiting
	WhitelistedIPs []string `json:"whitelisted_ips"`
}

// AdminPermissionsConfig represents admin permissions configuration
type AdminPermissionsConfig struct {
	// Default permissions for new admins
	DefaultPermissions []string `json:"default_permissions"`
	
	// Role definitions
	Roles map[string]*AdminRole `json:"roles"`
	
	// Permission hierarchy
	PermissionHierarchy map[string][]string `json:"permission_hierarchy"`
	
	// Super admin settings
	SuperAdminEmails []string `json:"super_admin_emails"`
	SuperAdminRole   string    `json:"super_admin_role"`
}

// AdminRole represents an admin role with permissions
type AdminRole struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsSystem    bool     `json:"is_system"`
}

// AdminDashboardConfig represents admin dashboard configuration
type AdminDashboardConfig struct {
	// Refresh settings
	AutoRefresh     bool          `json:"auto_refresh"`
	RefreshInterval  time.Duration `json:"refresh_interval"`
	
	// Data retention
	DataRetentionDays int `json:"data_retention_days"`
	
	// Chart settings
	DefaultChartPeriod string `json:"default_chart_period"`
	MaxDataPoints     int    `json:"max_data_points"`
	
	// Export settings
	ExportFormats     []string `json:"export_formats"`
	MaxExportRows    int       `json:"max_export_rows"`
	ExportTimeout     time.Duration `json:"export_timeout"`
	
	// Notification settings
	EnableNotifications bool     `json:"enable_notifications"`
	NotificationChannels []string `json:"notification_channels"`
}

// AdminSecurityConfig represents admin security configuration
type AdminSecurityConfig struct {
	// IP restrictions
	AllowedIPs    []string `json:"allowed_ips"`
	BlockedIPs     []string `json:"blocked_ips"`
	RequireIPWhitelist bool   `json:"require_ip_whitelist"`
	
	// Time restrictions
	AllowedHours    []int `json:"allowed_hours"` // 0-23
	AllowedDays     []int `json:"allowed_days"`  // 0-6 (Sunday-Saturday)
	RequireTimeRestrictions bool `json:"require_time_restrictions"`
	
	// Device restrictions
	RequireDeviceApproval bool `json:"require_device_approval"`
	MaxDevicesPerAdmin   int  `json:"max_devices_per_admin"`
	
	// Session security
	RequireHTTPS       bool `json:"require_https"`
	RequireSecureCookie bool `json:"require_secure_cookie"`
	
	// Audit settings
	LogAllActions      bool `json:"log_all_actions"`
	LogFailedAttempts  bool `json:"log_failed_attempts"`
	SensitiveDataMask bool `json:"sensitive_data_mask"`
}

// AdminAuditConfig represents admin audit configuration
type AdminAuditConfig struct {
	// Log retention
	LogRetentionDays int `json:"log_retention_days"`
	
	// Log levels
	LogLevel string `json:"log_level"`
	
	// Log destinations
	LogDestinations []string `json:"log_destinations"`
	
	// Sensitive operations
	SensitiveOperations []string `json:"sensitive_operations"`
	
	// Alert settings
	AlertOnSensitiveOps bool `json:"alert_on_sensitive_ops"`
	AlertRecipients     []string `json:"alert_recipients"`
}

// LoadAdminConfig creates admin configuration from application config
func LoadAdminConfig(appConfig *config.Config, redisClient *redis.Client, jwtUtils *utils.JWTUtils) *AdminConfig {
	return &AdminConfig{
		Auth:        loadAdminAuthConfig(appConfig, jwtUtils),
		RateLimit:    loadAdminRateLimitConfig(appConfig, redisClient),
		Permissions:  loadAdminPermissionsConfig(appConfig),
		Dashboard:    loadAdminDashboardConfig(appConfig),
		Security:     loadAdminSecurityConfig(appConfig),
		Audit:        loadAdminAuditConfig(appConfig),
	}
}

// loadAdminAuthConfig creates admin authentication configuration
func loadAdminAuthConfig(appConfig *config.Config, jwtUtils *utils.JWTUtils) *AdminAuthConfig {
	config := DefaultAdminAuthConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil {
		if appConfig.Admin.JWTSecret != "" {
			config.JWTSecret = appConfig.Admin.JWTSecret
		}
		if appConfig.Admin.JWTExpiry > 0 {
			config.JWTExpiry = appConfig.Admin.JWTExpiry
		}
		if appConfig.Admin.SessionTimeout > 0 {
			config.SessionTimeout = appConfig.Admin.SessionTimeout
		}
		if appConfig.Admin.MaxLoginAttempts > 0 {
			config.MaxLoginAttempts = appConfig.Admin.MaxLoginAttempts
		}
		if appConfig.Admin.LockoutDuration > 0 {
			config.LockoutDuration = appConfig.Admin.LockoutDuration
		}
	}
	
	// Adjust based on environment
	if appConfig.App.Env == "development" {
		config.Require2FA = false
		config.RequireTrustedIP = false
		config.MaxLoginAttempts = 100 // Very high for development
	} else if appConfig.App.Env == "production" {
		config.Require2FA = true
		config.RequireTrustedIP = true
		config.MaxLoginAttempts = 5
		config.LockoutDuration = 15 * time.Minute
	}
	
	return config
}

// loadAdminRateLimitConfig creates admin rate limiting configuration
func loadAdminRateLimitConfig(appConfig *config.Config, redisClient *redis.Client) *AdminRateLimitConfig {
	config := DefaultAdminRateLimitConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil {
		if appConfig.Admin.RequestsPerMinute > 0 {
			config.RequestsPerMinute = appConfig.Admin.RequestsPerMinute
		}
		if appConfig.Admin.RequestsPerHour > 0 {
			config.RequestsPerHour = appConfig.Admin.RequestsPerHour
		}
		if appConfig.Admin.RequestsPerDay > 0 {
			config.RequestsPerDay = appConfig.Admin.RequestsPerDay
		}
	}
	
	// Adjust based on environment
	if appConfig.App.Env == "development" {
		config.RequestsPerMinute = 1000
		config.RequestsPerHour = 10000
		config.RequestsPerDay = 100000
	} else if appConfig.App.Env == "production" {
		config.RequestsPerMinute = 100
		config.RequestsPerHour = 1000
		config.RequestsPerDay = 10000
	}
	
	return config
}

// loadAdminPermissionsConfig creates admin permissions configuration
func loadAdminPermissionsConfig(appConfig *config.Config) *AdminPermissionsConfig {
	config := DefaultAdminPermissionsConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil && len(appConfig.Admin.SuperAdminEmails) > 0 {
		config.SuperAdminEmails = appConfig.Admin.SuperAdminEmails
	}
	
	return config
}

// loadAdminDashboardConfig creates admin dashboard configuration
func loadAdminDashboardConfig(appConfig *config.Config) *AdminDashboardConfig {
	config := DefaultAdminDashboardConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil {
		if appConfig.Admin.DataRetentionDays > 0 {
			config.DataRetentionDays = appConfig.Admin.DataRetentionDays
		}
		if appConfig.Admin.MaxExportRows > 0 {
			config.MaxExportRows = appConfig.Admin.MaxExportRows
		}
	}
	
	return config
}

// loadAdminSecurityConfig creates admin security configuration
func loadAdminSecurityConfig(appConfig *config.Config) *AdminSecurityConfig {
	config := DefaultAdminSecurityConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil {
		if len(appConfig.Admin.AllowedIPs) > 0 {
			config.AllowedIPs = appConfig.Admin.AllowedIPs
		}
		if len(appConfig.Admin.BlockedIPs) > 0 {
			config.BlockedIPs = appConfig.Admin.BlockedIPs
		}
	}
	
	// Adjust based on environment
	if appConfig.App.Env == "production" {
		config.RequireHTTPS = true
		config.RequireSecureCookie = true
		config.RequireIPWhitelist = true
	} else {
		config.RequireHTTPS = false
		config.RequireSecureCookie = false
		config.RequireIPWhitelist = false
	}
	
	return config
}

// loadAdminAuditConfig creates admin audit configuration
func loadAdminAuditConfig(appConfig *config.Config) *AdminAuditConfig {
	config := DefaultAdminAuditConfig()
	
	// Override with config values if available
	if appConfig.Admin != nil {
		if appConfig.Admin.LogRetentionDays > 0 {
			config.LogRetentionDays = appConfig.Admin.LogRetentionDays
		}
		if appConfig.Admin.LogLevel != "" {
			config.LogLevel = appConfig.Admin.LogLevel
		}
	}
	
	return config
}

// Default configurations

// DefaultAdminAuthConfig returns default admin authentication configuration
func DefaultAdminAuthConfig() *AdminAuthConfig {
	return &AdminAuthConfig{
		JWTSecret:          "your-super-secret-jwt-key-change-in-production",
		JWTExpiry:          24 * time.Hour,
		RefreshExpiry:       7 * 24 * time.Hour,
		SessionTimeout:       8 * time.Hour,
		MaxSessions:         3,
		MinPasswordLength:    12,
		RequireUppercase:    true,
		RequireLowercase:    true,
		RequireNumbers:       true,
		RequireSymbols:       true,
		Require2FA:           true,
		2FAIssuer:           "Winkr Admin",
		2FAWindow:           3,
		MaxLoginAttempts:      5,
		LockoutDuration:      15 * time.Minute,
		TrustedIPs:          []string{},
		RequireTrustedIP:     false,
	}
}

// DefaultAdminRateLimitConfig returns default admin rate limiting configuration
func DefaultAdminRateLimitConfig() *AdminRateLimitConfig {
	return &AdminRateLimitConfig{
		RequestsPerMinute:          100,
		RequestsPerHour:            1000,
		RequestsPerDay:             10000,
		UserManagementRateLimit:     50,
		ContentModerationRateLimit:  200,
		SystemManagementRateLimit:   25,
		BurstSize:                  20,
		WhitelistedIPs:             []string{},
	}
}

// DefaultAdminPermissionsConfig returns default admin permissions configuration
func DefaultAdminPermissionsConfig() *AdminPermissionsConfig {
	return &AdminPermissionsConfig{
		DefaultPermissions: []string{
			"users.read",
			"analytics.read",
			"content.read",
		},
		Roles: map[string]*AdminRole{
			"admin": {
				Name:        "Admin",
				Description: "Standard administrator with limited permissions",
				Permissions: []string{
					"users.read",
					"users.write",
					"analytics.read",
					"content.read",
					"content.moderate",
					"system.read",
				},
				IsSystem: false,
			},
			"super_admin": {
				Name:        "Super Admin",
				Description: "Super administrator with full system access",
				Permissions: []string{
					"*", // All permissions
				},
				IsSystem: true,
			},
			"moderator": {
				Name:        "Moderator",
				Description: "Content moderator with limited permissions",
				Permissions: []string{
					"content.read",
					"content.moderate",
					"analytics.read",
				},
				IsSystem: false,
			},
			"analyst": {
				Name:        "Analyst",
				Description: "Data analyst with read-only permissions",
				Permissions: []string{
					"analytics.read",
					"users.read",
					"content.read",
					"system.read",
				},
				IsSystem: false,
			},
		},
		PermissionHierarchy: map[string][]string{
			"users.write":    {"users.read"},
			"users.delete":    {"users.write", "users.read"},
			"users.admin":     {"users.delete", "users.write", "users.read"},
			"content.moderate": {"content.read"},
			"content.admin":   {"content.moderate", "content.read"},
			"system.write":    {"system.read"},
			"system.admin":     {"system.write", "system.read"},
			"analytics.admin":  {"analytics.read"},
		},
		SuperAdminEmails: []string{},
		SuperAdminRole:   "super_admin",
	}
}

// DefaultAdminDashboardConfig returns default admin dashboard configuration
func DefaultAdminDashboardConfig() *AdminDashboardConfig {
	return &AdminDashboardConfig{
		AutoRefresh:      true,
		RefreshInterval:   5 * time.Minute,
		DataRetentionDays: 90,
		DefaultChartPeriod: "7d",
		MaxDataPoints:     1000,
		ExportFormats:     []string{"csv", "json", "xlsx"},
		MaxExportRows:     10000,
		ExportTimeout:     5 * time.Minute,
		EnableNotifications: true,
		NotificationChannels: []string{"email", "webhook"},
	}
}

// DefaultAdminSecurityConfig returns default admin security configuration
func DefaultAdminSecurityConfig() *AdminSecurityConfig {
	return &AdminSecurityConfig{
		AllowedIPs:              []string{},
		BlockedIPs:               []string{},
		RequireIPWhitelist:        false,
		AllowedHours:              []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}, // All hours
		AllowedDays:               []int{0, 1, 2, 3, 4, 5, 6}, // All days
		RequireTimeRestrictions:    false,
		RequireDeviceApproval:      false,
		MaxDevicesPerAdmin:        5,
		RequireHTTPS:              false,
		RequireSecureCookie:        false,
		LogAllActions:             true,
		LogFailedAttempts:         true,
		SensitiveDataMask:         true,
	}
}

// DefaultAdminAuditConfig returns default admin audit configuration
func DefaultAdminAuditConfig() *AdminAuditConfig {
	return &AdminAuditConfig{
		LogRetentionDays:     365,
		LogLevel:            "info",
		LogDestinations:      []string{"database", "file"},
		SensitiveOperations:  []string{
			"users.delete",
			"users.admin",
			"system.admin",
			"content.admin",
		},
		AlertOnSensitiveOps: true,
		AlertRecipients:     []string{"admin@example.com"},
	}
}