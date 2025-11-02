package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ManageSystemConfigUseCase handles system configuration management
type ManageSystemConfigUseCase struct {
	// In a real implementation, this would include configuration repository or service
}

// NewManageSystemConfigUseCase creates a new ManageSystemConfigUseCase
func NewManageSystemConfigUseCase() *ManageSystemConfigUseCase {
	return &ManageSystemConfigUseCase{}
}

// GetSystemConfigRequest represents a request to get system configuration
type GetSystemConfigRequest struct {
	AdminID uuid.UUID `json:"admin_id" validate:"required"`
	Section string    `json:"section" validate:"omitempty,oneof=general security email payment storage moderation"`
}

// UpdateSystemConfigRequest represents a request to update system configuration
type UpdateSystemConfigRequest struct {
	AdminID uuid.UUID              `json:"admin_id" validate:"required"`
	Section string                 `json:"section" validate:"required,oneof=general security email payment storage moderation"`
	Config  map[string]interface{} `json:"config" validate:"required"`
	Reason  string                 `json:"reason"`
}

// SystemConfigResponse represents response from getting system configuration
type SystemConfigResponse struct {
	Section    string                 `json:"section"`
	Config     map[string]interface{} `json:"config"`
	LastUpdated *time.Time            `json:"last_updated,omitempty"`
	UpdatedBy  *uuid.UUID             `json:"updated_by,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// UpdateSystemConfigResponse represents response from updating system configuration
type UpdateSystemConfigResponse struct {
	Section    string    `json:"section"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	UpdatedBy  uuid.UUID `json:"updated_by"`
	Timestamp  time.Time `json:"timestamp"`
}

// Execute retrieves system configuration
func (uc *ManageSystemConfigUseCase) Execute(ctx context.Context, req GetSystemConfigRequest) (*SystemConfigResponse, error) {
	logger.Info("ManageSystemConfig use case executed", "admin_id", req.AdminID, "section", req.Section)

	// Get configuration
	config, lastUpdated, updatedBy, err := uc.getSystemConfig(req.Section)
	if err != nil {
		logger.Error("Failed to get system configuration", err, "admin_id", req.AdminID, "section", req.Section)
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}

	logger.Info("ManageSystemConfig use case completed successfully", "admin_id", req.AdminID, "section", req.Section)
	return &SystemConfigResponse{
		Section:    req.Section,
		Config:     config,
		LastUpdated: lastUpdated,
		UpdatedBy:  updatedBy,
		Timestamp:  time.Now(),
	}, nil
}

// ExecuteUpdate updates system configuration
func (uc *ManageSystemConfigUseCase) ExecuteUpdate(ctx context.Context, req UpdateSystemConfigRequest) (*UpdateSystemConfigResponse, error) {
	logger.Info("ManageSystemConfig update use case executed", "admin_id", req.AdminID, "section", req.Section)

	// Validate configuration
	if err := uc.validateConfig(req.Section, req.Config); err != nil {
		logger.Error("Invalid configuration", err, "admin_id", req.AdminID, "section", req.Section)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Update configuration
	err := uc.updateSystemConfig(req.Section, req.Config, req.AdminID, req.Reason)
	if err != nil {
		logger.Error("Failed to update system configuration", err, "admin_id", req.AdminID, "section", req.Section)
		return nil, fmt.Errorf("failed to update system configuration: %w", err)
	}

	logger.Info("ManageSystemConfig update use case completed successfully", "admin_id", req.AdminID, "section", req.Section)
	return &UpdateSystemConfigResponse{
		Section:   req.Section,
		Status:    "success",
		Message:   "Configuration updated successfully",
		UpdatedBy: req.AdminID,
		Timestamp: time.Now(),
	}, nil
}

// getSystemConfig retrieves system configuration
func (uc *ManageSystemConfigUseCase) getSystemConfig(section string) (map[string]interface{}, *time.Time, *uuid.UUID, error) {
	// Mock data - in real implementation, this would query the configuration storage
	var config map[string]interface{}
	var lastUpdated *time.Time
	var updatedBy *uuid.UUID

	switch section {
	case "general":
		config = map[string]interface{}{
			"app_name":        "Winkr Dating App",
			"app_version":     "1.0.0",
			"environment":     "production",
			"debug_mode":      false,
			"maintenance_mode": false,
			"max_upload_size": 10485760, // 10MB
			"supported_languages": []string{"en", "es", "fr", "de", "it"},
			"default_language": "en",
		}
		lastUpdated = &[]time.Time{time.Now().Add(-24 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	case "security":
		config = map[string]interface{}{
			"password_min_length": 8,
			"password_require_uppercase": true,
			"password_require_lowercase": true,
			"password_require_numbers": true,
			"password_require_symbols": false,
			"session_timeout": 86400, // 24 hours
			"max_login_attempts": 5,
			"lockout_duration": 900, // 15 minutes
			"two_factor_auth": true,
			"jwt_expiry": 3600, // 1 hour
			"refresh_token_expiry": 604800, // 7 days
		}
		lastUpdated = &[]time.Time{time.Now().Add(-12 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	case "email":
		config = map[string]interface{}{
			"provider": "smtp",
			"smtp_host": "smtp.example.com",
			"smtp_port": 587,
			"smtp_username": "noreply@example.com",
			"smtp_use_tls": true,
			"from_email": "noreply@example.com",
			"from_name": "Winkr Dating",
			"verification_template": "verification",
			"password_reset_template": "password_reset",
			"welcome_template": "welcome",
		}
		lastUpdated = &[]time.Time{time.Now().Add(-48 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	case "payment":
		config = map[string]interface{}{
			"provider": "stripe",
			"stripe_publishable_key": "pk_test_...",
			"stripe_secret_key": "sk_test_...",
			"webhook_secret": "whsec_...",
			"currency": "USD",
			"supported_currencies": []string{"USD", "EUR", "GBP"},
			"subscription_plans": map[string]interface{}{
				"basic": map[string]interface{}{
					"name": "Basic",
					"price": 9.99,
					"duration": 30, // days
					"features": []string{"basic_swipes", "basic_chat"},
				},
				"premium": map[string]interface{}{
					"name": "Premium",
					"price": 19.99,
					"duration": 30, // days
					"features": []string{"unlimited_swipes", "unlimited_chat", "see_who_liked_you"},
				},
			},
		}
		lastUpdated = &[]time.Time{time.Now().Add(-72 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	case "storage":
		config = map[string]interface{}{
			"provider": "s3",
			"aws_region": "us-east-1",
			"s3_bucket": "winkr-storage",
			"cdn_url": "https://cdn.winkr.com",
			"photo_max_size": 5242880, // 5MB
			"photo_allowed_formats": []string{"jpg", "jpeg", "png"},
			"ephemeral_photo_ttl": 86400, // 24 hours
			"backup_enabled": true,
			"backup_frequency": "daily",
		}
		lastUpdated = &[]time.Time{time.Now().Add(-36 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	case "moderation":
		config = map[string]interface{}{
			"auto_moderation": true,
			"ai_moderation_enabled": true,
			"manual_review_threshold": 3, // number of reports
			"auto_ban_threshold": 10, // number of reports
			"banned_words": []string{"spam", "abuse", "harassment"},
			"photo_moderation": true,
			"message_moderation": true,
			"profile_moderation": true,
			"appeal_process": true,
		}
		lastUpdated = &[]time.Time{time.Now().Add(-6 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]

	default:
		// Return all sections if no specific section is requested
		config = map[string]interface{}{
			"general":    uc.getGeneralConfig(),
			"security":   uc.getSecurityConfig(),
			"email":      uc.getEmailConfig(),
			"payment":    uc.getPaymentConfig(),
			"storage":    uc.getStorageConfig(),
			"moderation": uc.getModerationConfig(),
		}
		lastUpdated = &[]time.Time{time.Now().Add(-1 * time.Hour)}[0]
		updatedBy = &[]uuid.UUID{uuid.New()}[0]
	}

	return config, lastUpdated, updatedBy, nil
}

// Helper methods for getting specific configurations
func (uc *ManageSystemConfigUseCase) getGeneralConfig() map[string]interface{} {
	return map[string]interface{}{
		"app_name":        "Winkr Dating App",
		"app_version":     "1.0.0",
		"environment":     "production",
		"debug_mode":      false,
		"maintenance_mode": false,
		"max_upload_size": 10485760, // 10MB
		"supported_languages": []string{"en", "es", "fr", "de", "it"},
		"default_language": "en",
	}
}

func (uc *ManageSystemConfigUseCase) getSecurityConfig() map[string]interface{} {
	return map[string]interface{}{
		"password_min_length": 8,
		"password_require_uppercase": true,
		"password_require_lowercase": true,
		"password_require_numbers": true,
		"password_require_symbols": false,
		"session_timeout": 86400, // 24 hours
		"max_login_attempts": 5,
		"lockout_duration": 900, // 15 minutes
		"two_factor_auth": true,
		"jwt_expiry": 3600, // 1 hour
		"refresh_token_expiry": 604800, // 7 days
	}
}

func (uc *ManageSystemConfigUseCase) getEmailConfig() map[string]interface{} {
	return map[string]interface{}{
		"provider": "smtp",
		"smtp_host": "smtp.example.com",
		"smtp_port": 587,
		"smtp_username": "noreply@example.com",
		"smtp_use_tls": true,
		"from_email": "noreply@example.com",
		"from_name": "Winkr Dating",
		"verification_template": "verification",
		"password_reset_template": "password_reset",
		"welcome_template": "welcome",
	}
}

func (uc *ManageSystemConfigUseCase) getPaymentConfig() map[string]interface{} {
	return map[string]interface{}{
		"provider": "stripe",
		"stripe_publishable_key": "pk_test_...",
		"stripe_secret_key": "sk_test_...",
		"webhook_secret": "whsec_...",
		"currency": "USD",
		"supported_currencies": []string{"USD", "EUR", "GBP"},
		"subscription_plans": map[string]interface{}{
			"basic": map[string]interface{}{
				"name": "Basic",
				"price": 9.99,
				"duration": 30, // days
				"features": []string{"basic_swipes", "basic_chat"},
			},
			"premium": map[string]interface{}{
				"name": "Premium",
				"price": 19.99,
				"duration": 30, // days
				"features": []string{"unlimited_swipes", "unlimited_chat", "see_who_liked_you"},
			},
		},
	}
}

func (uc *ManageSystemConfigUseCase) getStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"provider": "s3",
		"aws_region": "us-east-1",
		"s3_bucket": "winkr-storage",
		"cdn_url": "https://cdn.winkr.com",
		"photo_max_size": 5242880, // 5MB
		"photo_allowed_formats": []string{"jpg", "jpeg", "png"},
		"ephemeral_photo_ttl": 86400, // 24 hours
		"backup_enabled": true,
		"backup_frequency": "daily",
	}
}

func (uc *ManageSystemConfigUseCase) getModerationConfig() map[string]interface{} {
	return map[string]interface{}{
		"auto_moderation": true,
		"ai_moderation_enabled": true,
		"manual_review_threshold": 3, // number of reports
		"auto_ban_threshold": 10, // number of reports
		"banned_words": []string{"spam", "abuse", "harassment"},
		"photo_moderation": true,
		"message_moderation": true,
		"profile_moderation": true,
		"appeal_process": true,
	}
}

// validateConfig validates configuration values
func (uc *ManageSystemConfigUseCase) validateConfig(section string, config map[string]interface{}) error {
	// Mock validation - in real implementation, this would validate configuration values
	switch section {
	case "general":
		if appName, ok := config["app_name"].(string); ok && appName == "" {
			return fmt.Errorf("app_name cannot be empty")
		}
	case "security":
		if minLen, ok := config["password_min_length"].(int); ok && minLen < 6 {
			return fmt.Errorf("password_min_length must be at least 6")
		}
	case "email":
		if fromEmail, ok := config["from_email"].(string); ok && fromEmail == "" {
			return fmt.Errorf("from_email cannot be empty")
		}
	case "payment":
		if currency, ok := config["currency"].(string); ok && currency == "" {
			return fmt.Errorf("currency cannot be empty")
		}
	case "storage":
		if bucket, ok := config["s3_bucket"].(string); ok && bucket == "" {
			return fmt.Errorf("s3_bucket cannot be empty")
		}
	case "moderation":
		if threshold, ok := config["manual_review_threshold"].(int); ok && threshold < 1 {
			return fmt.Errorf("manual_review_threshold must be at least 1")
		}
	}

	return nil
}

// updateSystemConfig updates system configuration
func (uc *ManageSystemConfigUseCase) updateSystemConfig(section string, config map[string]interface{}, adminID uuid.UUID, reason string) error {
	// Mock implementation - in real implementation, this would update the configuration storage
	logger.Info("System configuration updated", "section", section, "admin_id", adminID, "reason", reason)
	return nil
}