package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// webhookEventRepositoryImpl implements WebhookEventRepository interface
type webhookEventRepositoryImpl struct {
	db *gorm.DB
}

// NewWebhookEventRepository creates a new webhook event repository
func NewWebhookEventRepository(db *gorm.DB) repositories.WebhookEventRepository {
	return &webhookEventRepositoryImpl{
		db: db,
	}
}

// Create creates a new webhook event
func (r *webhookEventRepositoryImpl) Create(ctx context.Context, webhookEvent *entities.WebhookEvent) error {
	if err := r.db.WithContext(ctx).Create(webhookEvent).Error; err != nil {
		return fmt.Errorf("failed to create webhook event: %w", err)
	}
	return nil
}

// GetByID retrieves a webhook event by ID
func (r *webhookEventRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.WebhookEvent, error) {
	var webhookEvent entities.WebhookEvent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&webhookEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrWebhookEventNotFound
		}
		return nil, fmt.Errorf("failed to get webhook event: %w", err)
	}
	return &webhookEvent, nil
}

// GetByStripeID retrieves a webhook event by Stripe event ID
func (r *webhookEventRepositoryImpl) GetByStripeID(ctx context.Context, stripeID string) (*entities.WebhookEvent, error) {
	var webhookEvent entities.WebhookEvent
	if err := r.db.WithContext(ctx).Where("stripe_event_id = ?", stripeID).First(&webhookEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrWebhookEventNotFound
		}
		return nil, fmt.Errorf("failed to get webhook event by Stripe ID: %w", err)
	}
	return &webhookEvent, nil
}

// Update updates a webhook event
func (r *webhookEventRepositoryImpl) Update(ctx context.Context, webhookEvent *entities.WebhookEvent) error {
	if err := r.db.WithContext(ctx).Save(webhookEvent).Error; err != nil {
		return fmt.Errorf("failed to update webhook event: %w", err)
	}
	return nil
}

// Delete deletes a webhook event
func (r *webhookEventRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.WebhookEvent{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete webhook event: %w", err)
	}
	return nil
}

// GetByType retrieves webhook events by type
func (r *webhookEventRepositoryImpl) GetByType(ctx context.Context, eventType string, limit, offset int) ([]*entities.WebhookEvent, error) {
	var webhookEvents []*entities.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("event_type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&webhookEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook events by type: %w", err)
	}
	return webhookEvents, nil
}

// GetByStatus retrieves webhook events by status
func (r *webhookEventRepositoryImpl) GetByStatus(ctx context.Context, status entities.WebhookEventStatus, limit, offset int) ([]*entities.WebhookEvent, error) {
	var webhookEvents []*entities.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&webhookEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook events by status: %w", err)
	}
	return webhookEvents, nil
}

// GetByDateRange retrieves webhook events within a date range
func (r *webhookEventRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.WebhookEvent, error) {
	var webhookEvents []*entities.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&webhookEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get webhook events by date range: %w", err)
	}
	return webhookEvents, nil
}

// GetFailedEvents retrieves failed webhook events
func (r *webhookEventRepositoryImpl) GetFailedEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error) {
	var webhookEvents []*entities.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("status = ?", entities.WebhookEventStatusFailed).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&webhookEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed webhook events: %w", err)
	}
	return webhookEvents, nil
}

// GetPendingEvents retrieves pending webhook events
func (r *webhookEventRepositoryImpl) GetPendingEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error) {
	var webhookEvents []*entities.WebhookEvent
	if err := r.db.WithContext(ctx).
		Where("status = ?", entities.WebhookEventStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&webhookEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending webhook events: %w", err)
	}
	return webhookEvents, nil
}

// GetCountByStatus retrieves count of webhook events by status
func (r *webhookEventRepositoryImpl) GetCountByStatus(ctx context.Context, status entities.WebhookEventStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get webhook event count by status: %w", err)
	}
	return count, nil
}

// GetCountByType retrieves count of webhook events by type
func (r *webhookEventRepositoryImpl) GetCountByType(ctx context.Context, eventType string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("event_type = ?", eventType).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get webhook event count by type: %w", err)
	}
	return count, nil
}

// GetAnalytics retrieves webhook event analytics
func (r *webhookEventRepositoryImpl) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*repositories.WebhookEventAnalytics, error) {
	var analytics repositories.WebhookEventAnalytics
	
	// Get total events
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get total webhook events: %w", err)
	}
	
	// Get processed events
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.WebhookEventStatusProcessed).
		Count(&analytics.ProcessedEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get processed webhook events: %w", err)
	}
	
	// Get failed events
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.WebhookEventStatusFailed).
		Count(&analytics.FailedEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed webhook events: %w", err)
	}
	
	// Get pending events
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.WebhookEventStatusPending).
		Count(&analytics.PendingEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending webhook events: %w", err)
	}
	
	return &analytics, nil
}

// GetEventTypeAnalytics retrieves webhook event type analytics
func (r *webhookEventRepositoryImpl) GetEventTypeAnalytics(ctx context.Context, startDate, endDate time.Time) ([]repositories.EventTypeAnalytics, error) {
	var eventTypeAnalytics []repositories.EventTypeAnalytics
	
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Select("event_type, COUNT(*) as count, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as processed, SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed", entities.WebhookEventStatusProcessed, entities.WebhookEventStatusFailed).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("event_type").
		Order("count DESC").
		Scan(&eventTypeAnalytics).Error; err != nil {
		return nil, fmt.Errorf("failed to get event type analytics: %w", err)
	}
	
	return eventTypeAnalytics, nil
}

// MarkAsProcessed marks a webhook event as processed
func (r *webhookEventRepositoryImpl) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     entities.WebhookEventStatusProcessed,
			"processed_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to mark webhook event as processed: %w", err)
	}
	return nil
}

// MarkAsFailed marks a webhook event as failed
func (r *webhookEventRepositoryImpl) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	if err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        entities.WebhookEventStatusFailed,
			"error_message": errorMessage,
			"processed_at":  time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to mark webhook event as failed: %w", err)
	}
	return nil
}

// CleanupOldEvents deletes old webhook events
func (r *webhookEventRepositoryImpl) CleanupOldEvents(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&entities.WebhookEvent{})
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old webhook events: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// BatchCreate creates multiple webhook events in a single transaction
func (r *webhookEventRepositoryImpl) BatchCreate(ctx context.Context, webhookEvents []*entities.WebhookEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, webhookEvent := range webhookEvents {
			if err := tx.Create(webhookEvent).Error; err != nil {
				return fmt.Errorf("failed to create webhook event: %w", err)
			}
		}
		return nil
	})
}

// BatchUpdate updates multiple webhook events in a single transaction
func (r *webhookEventRepositoryImpl) BatchUpdate(ctx context.Context, webhookEvents []*entities.WebhookEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, webhookEvent := range webhookEvents {
			if err := tx.Save(webhookEvent).Error; err != nil {
				return fmt.Errorf("failed to update webhook event: %w", err)
			}
		}
		return nil
	})
}