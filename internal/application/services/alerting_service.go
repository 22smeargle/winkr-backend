
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThreshold      AlertType = "threshold"
	AlertTypeAnomaly       AlertType = "anomaly"
	AlertTypeErrorRate      AlertType = "error_rate"
	AlertTypePerformance    AlertType = "performance"
	AlertTypeResource       AlertType = "resource"
	AlertTypeExternal       AlertType = "external"
)

// Alert represents an alert
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Status      AlertStatus            `json:"status"`
	Message     string                 `json:"message"`
	Description string                 `json:"description"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	TriggeredAt time.Time              `json:"triggered_at"`
	AckedAt    *time.Time            `json:"acked_at,omitempty"`
	ResolvedAt  *time.Time            `json:"resolved_at,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Source      string                 `json:"source"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Condition   string                 `json:"condition"`
	RuleID      string                 `json:"rule_id,omitempty"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Condition   string                 `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Source      string                 `json:"source"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
	LastTriggered *time.Time            `json:"last_triggered,omitempty"`
	TriggerCount int64                  `json:"trigger_count"`
}

// NotificationChannel represents a notification channel
type NotificationChannel struct {
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
}

// AlertingService provides alerting functionality
type AlertingService struct {
	config         *config.Config
	cacheService   *cache.CacheService
	mu             sync.RWMutex
	alerts         map[string]*Alert
	rules          map[string]*AlertRule
	channels       map[string]*NotificationChannel
	lastEvaluation  time.Time
}

// NewAlertingService creates a new alerting service
func NewAlertingService(
	cfg *config.Config,
	cacheService *cache.CacheService,
) *AlertingService {
	return &AlertingService{
		config:       cfg,
		cacheService: cacheService,
		alerts:      make(map[string]*Alert),
		rules:       make(map[string]*AlertRule),
		channels:    make(map[string]*NotificationChannel),
		lastEvaluation: time.Now(),
	}
}

// AddAlert adds a new alert
func (a *AlertingService) AddAlert(ctx context.Context, alert *Alert) error {
	if !a.config.Monitoring.Alerting.Enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Generate ID if not provided
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	// Set triggered time if not set
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now()
	}

	// Check if alert already exists
	if existing, exists := a.alerts[alert.ID]; exists {
		// Update existing alert
		existing.Status = alert.Status
		existing.Message = alert.Message
		existing.Annotations = alert.Annotations
		
		// Update resolved/acknowledged times
		if alert.Status == AlertStatusResolved && existing.ResolvedAt == nil {
			now := time.Now()
			existing.ResolvedAt = &now
			existing.Duration = now.Sub(existing.TriggeredAt)
		}
		if alert.Status == AlertStatusAcknowledged && existing.AckedAt == nil {
			now := time.Now()
			existing.AckedAt = &now
		}
		
		alert = existing
	} else {
		// Add new alert
		a.alerts[alert.ID] = alert
	}

	// Store alert in cache
	if err := a.storeAlert(ctx, alert); err != nil {
		logger.Error("Failed to store alert", err)
		return err
	}

	logger.Info("Alert added/updated", map[string]interface{}{
		"alert_id": alert.ID,
		"status":   alert.Status,
		"severity": alert.Severity,
	})

	return nil
}

// GetAlert gets an alert by ID
func (a *AlertingService) GetAlert(ctx context.Context, alertID string) (*Alert, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alert, exists := a.alerts[alertID]
	if !exists {
		return nil, ErrAlertNotFound
	}

	return alert, nil
}

// GetAlerts gets alerts with optional filtering
func (a *AlertingService) GetAlerts(ctx context.Context, status AlertStatus, severity AlertSeverity, limit int) ([]*Alert, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []*Alert
	count := 0

	for _, alert := range a.alerts {
		// Filter by status
		if status != "" && alert.Status != status {
			continue
		}

		// Filter by severity
		if severity != "" && alert.Severity != severity {
			continue
		}

		result = append(result, alert)
		count++

		// Apply limit
		if limit > 0 && count >= limit {
			break
		}
	}

	return result, nil
}

// AcknowledgeAlert acknowledges an alert
func (a *AlertingService) AcknowledgeAlert(ctx context.Context, alertID string, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	alert, exists := a.alerts[alertID]
	if !exists {
		return ErrAlertNotFound
	}

	if alert.Status != AlertStatusActive {
		return fmt.Errorf("alert cannot be acknowledged in status: %s", alert.Status)
	}

	now := time.Now()
	alert.Status = AlertStatusAcknowledged
	alert.AckedAt = &now

	// Add annotation
	if alert.Annotations == nil {
		alert.Annotations = make(map[string]string)
	}
	alert.Annotations["acknowledged_by"] = userID
	alert.Annotations["acknowledged_at"] = now.Format(time.RFC3339)

	// Store updated alert
	if err := a.storeAlert(ctx, alert); err != nil {
		logger.Error("Failed to store acknowledged alert", err)
		return err
	}

	logger.Info("Alert acknowledged", map[string]interface{}{
		"alert_id": alertID,
		"user_id":  userID,
	})

	return nil
}

// ResolveAlert resolves an alert
func (a *AlertingService) ResolveAlert(ctx context.Context, alertID string, userID string, resolution string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	alert, exists := a.alerts[alertID]
	if !exists {
		return ErrAlertNotFound
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.ResolvedAt = &now
	alert.Duration = now.Sub(alert.TriggeredAt)

	// Add annotations
	if alert.Annotations == nil {
		alert.Annotations = make(map[string]string)
	}
	alert.Annotations["resolved_by"] = userID
	alert.Annotations["resolved_at"] = now.Format(time.RFC3339)
	alert.Annotations["resolution"] = resolution

	// Store updated alert
	if err := a.storeAlert(ctx, alert); err != nil {
		logger.Error("Failed to store resolved alert", err)
		return err
	}

	logger.Info("Alert resolved", map[string]interface{}{
		"alert_id":   alertID,
		"user_id":    userID,
		"resolution":  resolution,
	})

	return nil
}

// AddAlertRule adds a new alert rule
func (a *AlertingService) AddAlertRule(ctx context.Context, rule *AlertRule) error {
	if !a.config.Monitoring.Alerting.Enabled {
		return ErrMonitoringDisabled
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if rule already exists
	if _, exists := a.rules[rule.ID]; exists {
		return ErrAlertAlreadyExists
	}

	// Set evaluated time if not set
	if rule.EvaluatedAt.IsZero() {
		rule.EvaluatedAt = time.Now()
	}

	// Store rule in cache
	if err := a.storeAlertRule(ctx, rule); err != nil {
		logger.Error("Failed to store alert rule", err)
		return err
	}

	a.rules[rule.ID] = rule

	logger.Info("Alert rule added", map[string]interface{}{
		"rule_id": rule.ID,
		"name":    rule.Name,
		"type":    rule.Type,
	})

	return nil
}

// GetAlertRules gets all alert rules
func (a *AlertingService) GetAlertRules(ctx context.Context) ([]*AlertRule, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []*AlertRule
	for _, rule := range a.rules {
		result = append(result, rule)
	}

	return result, nil
}

// EvaluateRules evaluates all alert rules
func (a *AlertingService) EvaluateRules(ctx context.Context, metrics map[string]interface{}) error {
	if !a.config.Monitoring.Alerting.Enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	
	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		// Check if it's time to evaluate (based on duration)
		if now.Sub(rule.LastTriggered.Time) < rule.Duration {
			continue
		}

		// Evaluate rule
		triggered, value := a.evaluateRule(rule, metrics)
		if triggered {
			// Create alert
			alert := &Alert{
				ID:          uuid.New().String(),
				Name:        rule.Name,
				Type:        rule.Type,
				Severity:    rule.Severity,
				Status:      AlertStatusActive,
				Message:     fmt.Sprintf("Alert '%s' triggered: %s", rule.Name, rule.Description),
				Description: rule.Description,
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
				TriggeredAt: now,
				Source:      rule.Source,
				Value:       value,
				Threshold:   rule.Threshold,
				Condition:   rule.Condition,
				RuleID:      rule.ID,
			}

			// Update rule trigger count and last triggered time
			rule.TriggerCount++
			rule.LastTriggered = &now

			// Store alert
			if err := a.AddAlert(ctx, alert); err != nil {
				logger.Error("Failed to store triggered alert", err)
			}

			// Store updated rule
			if err := a.storeAlertRule(ctx, rule); err != nil {
				logger.Error("Failed to store updated alert rule", err)
			}
		}
	}

	a.lastEvaluation = now
	return nil
}

// SendNotification sends a notification for an alert
func (a *AlertingService) SendNotification(ctx context.Context, alert *Alert) error {
	// Get notification channels for this alert type
	channels := a.getNotificationChannels(alert.Severity)
	
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		// Send notification based on channel type
		switch channel.Type {
		case "email":
			if err := a.sendEmailNotification(ctx, alert, channel.Config); err != nil {
				logger.Error("Failed to send email notification", err)
			}
		case "slack":
			if err := a.sendSlackNotification(ctx, alert, channel.Config); err != nil {
				logger.Error("Failed to send Slack notification", err)
			}
		case "webhook":
			if err := a.sendWebhookNotification(ctx, alert, channel.Config); err != nil {
				logger.Error("Failed to send webhook notification", err)
			}
		}
	}

	return nil
}

// CleanupOldAlerts removes old alerts based on retention policy
func (a *AlertingService) CleanupOldAlerts(ctx context.Context) error {
	if !a.config.Monitoring.Alerting.Enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().Add(-a.config.Monitoring.Storage.AlertRetention)
	
	for id, alert := range a.alerts {
		if alert.TriggeredAt.Before(cutoff) {
			// Remove from memory
			delete(a.alerts, id)
			
			// Remove from cache
			if err := a.deleteAlert(ctx, id); err != nil {
				logger.Error("Failed to delete old alert", err)
			}
		}
	}

	logger.Info("Cleaned up old alerts")
	return nil
}

// Helper methods

func (a *AlertingService) storeAlert(ctx context.Context, alert *Alert) error {
	key := fmt.Sprintf("alert:%s", alert.ID)
	return a.cacheService.Set(ctx, key, alert, a.config.Monitoring.Storage.AlertRetention)
}

func (a *AlertingService) deleteAlert(ctx context.Context, alertID string) error {
	key := fmt.Sprintf("alert:%s", alertID)
	return a.cacheService.Delete(ctx, key)
}

func (a *AlertingService) storeAlertRule(ctx context.Context, rule *AlertRule) error {
	key := fmt.Sprintf("alert_rule:%s", rule.ID)
	return a.cacheService.Set(ctx, key, rule, a.config.Monitoring.Storage.AlertRetention)
}

func (a *AlertingService) getNotificationChannels(severity AlertSeverity) []*NotificationChannel {
	var channels []*NotificationChannel
	
	for _, channel := range a.channels {
		// Check if channel handles this severity
		if a.channelHandlesSeverity(channel, severity) {
			channels = append(channels, channel)
		}
	}
	
	return channels
}

func (a *AlertingService) channelHandlesSeverity(channel *NotificationChannel, severity AlertSeverity) bool {
	// This is a simplified check - in production, you'd have more sophisticated logic
	return true
}

func (a *AlertingService) evaluateRule(rule *AlertRule, metrics map[string]interface{}) (bool, float64) {
	// This is a simplified rule evaluation
	// In production, you'd want more sophisticated rule evaluation logic
	
	// Get metric value
	value, exists := metrics[rule.Source]
	if !exists {
		return false, 0
	}
	
	var floatValue float64
	switch v := value.(type) {
	case float64:
		floatValue = v
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	default:
		return false, 0
	}
	
	// Evaluate condition based on rule type
	switch rule.Type {
	case AlertTypeThreshold:
		return a.evaluateThresholdRule(rule, floatValue)
	case AlertTypeErrorRate:
		return a.evaluateErrorRateRule(rule, metrics)
	case AlertTypePerformance:
		return a.evaluatePerformanceRule(rule, floatValue)
	case AlertTypeResource:
		return a.evaluateResourceRule(rule, floatValue)
	default:
		return false, 0
	}
}

func (a *AlertingService) evaluateThresholdRule(rule *AlertRule, value float64) (bool, float64) {
	switch rule.Condition {
	case "greater_than":
		return value > rule.Threshold, value
	case "less_than":
		return value < rule.Threshold, value
	case "equals":
		return value == rule.Threshold, value
	default:
		return false, 0
	}
}

func (a *AlertingService) evaluateErrorRateRule(rule *AlertRule, metrics map[string]interface{}) (bool, float64) {
	// Simplified error rate evaluation
	// In production, you'd calculate actual error rate from metrics
	errorRate := 5.0 // Placeholder
	return errorRate > rule.Threshold, errorRate
}

func (a *AlertingService) evaluatePerformanceRule(rule *AlertRule, value float64) (bool, float64) {
	// Simplified performance evaluation
	// In production, you'd compare against performance thresholds
	return value > rule.Threshold, value
}

func (a *AlertingService) evaluateResourceRule(rule *AlertRule, value float64) (bool, float64) {
	// Simplified resource evaluation
	// In production, you'd compare against resource thresholds
	return value > rule.Threshold, value
}

func (a *AlertingService) sendEmailNotification(ctx context.Context, alert *Alert, config map[string]interface{}) error {
	// This is a placeholder - in production, you'd implement actual email sending
	logger.Info("Sending email notification", map[string]interface{}{
		"alert_id": alert.ID,
		"to":       config["to"],
	})
	return nil
}

func (a *AlertingService) sendSlackNotification(ctx context.Context, alert *Alert, config map[string]interface{}) error {
	// This is a placeholder - in production, you'd implement actual Slack integration
	logger.Info("Sending Slack notification", map[string]interface{}{
		"alert_id": alert.ID,
		"webhook":  config["webhook"],
	})
	return nil
}

func (a *AlertingService) sendWebhookNotification(ctx context.Context, alert *Alert, config map[string]interface{}) error {
	// This is a placeholder - in production, you'd implement actual webhook sending
	logger.Info("Sending webhook notification", map[string]interface{}{
		"alert_id": alert.ID,
		"url":      config["url"],
	})
	return nil
}
		