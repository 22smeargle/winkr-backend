package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SubscriptionRepositoryImpl implements SubscriptionRepository interface using GORM
type SubscriptionRepositoryImpl struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository instance
func NewSubscriptionRepository(db *gorm.DB) repositories.SubscriptionRepository {
	return &SubscriptionRepositoryImpl{db: db}
}

// Create creates a new subscription
func (r *SubscriptionRepositoryImpl) Create(ctx context.Context, subscription *entities.Subscription) error {
	modelSubscription := r.domainToModelSubscription(subscription)
	if err := r.db.WithContext(ctx).Create(modelSubscription).Error; err != nil {
		logger.Error("Failed to create subscription", err)
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	logger.Info("Subscription created successfully", map[string]interface{}{
		"subscription_id": subscription.ID,
		"user_id": subscription.UserID,
		"plan": subscription.Plan,
	})
	return nil
}

// GetByID retrieves a subscription by ID
func (r *SubscriptionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	var subscription models.Subscription
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		logger.Error("Failed to get subscription by ID", err)
		return nil, fmt.Errorf("failed to get subscription by ID: %w", err)
	}

	// Convert to domain entity
	domainSubscription := r.modelToDomainSubscription(&subscription)
	return domainSubscription, nil
}

// Update updates a subscription
func (r *SubscriptionRepositoryImpl) Update(ctx context.Context, subscription *entities.Subscription) error {
	modelSubscription := r.domainToModelSubscription(subscription)
	if err := r.db.WithContext(ctx).Save(modelSubscription).Error; err != nil {
		logger.Error("Failed to update subscription", err)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info("Subscription updated successfully", map[string]interface{}{
		"subscription_id": subscription.ID,
	})
	return nil
}

// Delete soft deletes a subscription
func (r *SubscriptionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete subscription", err)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	logger.Info("Subscription deleted successfully", map[string]interface{}{
		"subscription_id": id,
	})
	return nil
}

// GetUserSubscription retrieves subscription for a user
func (r *SubscriptionRepositoryImpl) GetUserSubscription(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	var subscription models.Subscription
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		logger.Error("Failed to get user subscription", err)
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	// Convert to domain entity
	domainSubscription := r.modelToDomainSubscription(&subscription)
	return domainSubscription, nil
}

// GetUserActiveSubscription retrieves active subscription for a user
func (r *SubscriptionRepositoryImpl) GetUserActiveSubscription(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	var subscription models.Subscription
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)", userID, true, time.Now()).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("active subscription not found")
		}
		logger.Error("Failed to get user active subscription", err)
		return nil, fmt.Errorf("failed to get user active subscription: %w", err)
	}

	// Convert to domain entity
	domainSubscription := r.modelToDomainSubscription(&subscription)
	return domainSubscription, nil
}

// UpdateSubscriptionStatus updates subscription status
func (r *SubscriptionRepositoryImpl) UpdateSubscriptionStatus(ctx context.Context, subscriptionID uuid.UUID, isActive bool) error {
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", subscriptionID).Update("is_active", isActive).Error; err != nil {
		logger.Error("Failed to update subscription status", err)
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	logger.Info("Subscription status updated", map[string]interface{}{
		"subscription_id": subscriptionID,
		"is_active": isActive,
	})
	return nil
}

// CancelSubscription cancels a subscription
func (r *SubscriptionRepositoryImpl) CancelSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	updates := map[string]interface{}{
		"is_canceled": true,
		"canceled_at": time.Now(),
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", subscriptionID).Updates(updates).Error; err != nil {
		logger.Error("Failed to cancel subscription", err)
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	logger.Info("Subscription canceled", map[string]interface{}{
		"subscription_id": subscriptionID,
	})
	return nil
}

// RenewSubscription renews a subscription
func (r *SubscriptionRepositoryImpl) RenewSubscription(ctx context.Context, subscriptionID uuid.UUID, newExpiryDate time.Time) error {
	updates := map[string]interface{}{
		"expires_at": newExpiryDate,
		"is_canceled": false,
		"canceled_at": nil,
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", subscriptionID).Updates(updates).Error; err != nil {
		logger.Error("Failed to renew subscription", err)
		return fmt.Errorf("failed to renew subscription: %w", err)
	}

	logger.Info("Subscription renewed", map[string]interface{}{
		"subscription_id": subscriptionID,
		"expires_at": newExpiryDate,
	})
	return nil
}

// UpdateStripeSubscriptionID updates Stripe subscription ID
func (r *SubscriptionRepositoryImpl) UpdateStripeSubscriptionID(ctx context.Context, subscriptionID uuid.UUID, stripeSubscriptionID string) error {
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", subscriptionID).Update("stripe_subscription_id", stripeSubscriptionID).Error; err != nil {
		logger.Error("Failed to update Stripe subscription ID", err)
		return fmt.Errorf("failed to update Stripe subscription ID: %w", err)
	}

	logger.Info("Stripe subscription ID updated", map[string]interface{}{
		"subscription_id": subscriptionID,
		"stripe_subscription_id": stripeSubscriptionID,
	})
	return nil
}

// UpdateStripeCustomerID updates Stripe customer ID
func (r *SubscriptionRepositoryImpl) UpdateStripeCustomerID(ctx context.Context, userID uuid.UUID, stripeCustomerID string) error {
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("user_id = ?", userID).Update("stripe_customer_id", stripeCustomerID).Error; err != nil {
		logger.Error("Failed to update Stripe customer ID", err)
		return fmt.Errorf("failed to update Stripe customer ID: %w", err)
	}

	logger.Info("Stripe customer ID updated", map[string]interface{}{
		"user_id": userID,
		"stripe_customer_id": stripeCustomerID,
	})
	return nil
}

// GetSubscriptionsByPlan retrieves subscriptions by plan
func (r *SubscriptionRepositoryImpl) GetSubscriptionsByPlan(ctx context.Context, plan string, limit, offset int) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("plan = ?", plan).Order("created_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get subscriptions by plan", err)
		return nil, fmt.Errorf("failed to get subscriptions by plan: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetActiveSubscriptions retrieves active subscriptions
func (r *SubscriptionRepositoryImpl) GetActiveSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, time.Now()).Order("created_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get active subscriptions", err)
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetExpiredSubscriptions retrieves expired subscriptions
func (r *SubscriptionRepositoryImpl) GetExpiredSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Order("expires_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get expired subscriptions", err)
		return nil, fmt.Errorf("failed to get expired subscriptions: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetCanceledSubscriptions retrieves canceled subscriptions
func (r *SubscriptionRepositoryImpl) GetCanceledSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("is_canceled = ?", true).Order("canceled_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get canceled subscriptions", err)
		return nil, fmt.Errorf("failed to get canceled subscriptions: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetSubscriptionsExpiringSoon retrieves subscriptions expiring soon
func (r *SubscriptionRepositoryImpl) GetSubscriptionsExpiringSoon(ctx context.Context, days int, limit, offset int) ([]*entities.Subscription, error) {
	expiryDate := time.Now().AddDate(0, 0, days)
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ? AND expires_at > ? AND is_canceled = ?", expiryDate, time.Now(), false).Order("expires_at ASC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get subscriptions expiring soon", err)
		return nil, fmt.Errorf("failed to get subscriptions expiring soon: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetSubscriptionCount retrieves subscription count
func (r *SubscriptionRepositoryImpl) GetSubscriptionCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Count(&count).Error; err != nil {
		logger.Error("Failed to get subscription count", err)
		return 0, fmt.Errorf("failed to get subscription count: %w", err)
	}

	return count, nil
}

// GetActiveSubscriptionCount retrieves active subscription count
func (r *SubscriptionRepositoryImpl) GetActiveSubscriptionCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, time.Now()).Count(&count).Error; err != nil {
		logger.Error("Failed to get active subscription count", err)
		return 0, fmt.Errorf("failed to get active subscription count: %w", err)
	}

	return count, nil
}

// GetSubscriptionCountByPlan retrieves subscription count by plan
func (r *SubscriptionRepositoryImpl) GetSubscriptionCountByPlan(ctx context.Context, plan string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("plan = ?", plan).Count(&count).Error; err != nil {
		logger.Error("Failed to get subscription count by plan", err)
		return 0, fmt.Errorf("failed to get subscription count by plan: %w", err)
	}

	return count, nil
}

// GetSubscriptionsCreatedInRange retrieves subscriptions created in date range
func (r *SubscriptionRepositoryImpl) GetSubscriptionsCreatedInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get subscriptions created in range", err)
		return nil, fmt.Errorf("failed to get subscriptions created in range: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// ExistsSubscription checks if subscription exists
func (r *SubscriptionRepositoryImpl) ExistsSubscription(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check subscription existence", err)
		return false, fmt.Errorf("failed to check subscription existence: %w", err)
	}

	return count > 0, nil
}

// ExistsUserSubscription checks if user has subscription
func (r *SubscriptionRepositoryImpl) ExistsUserSubscription(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to check user subscription existence", err)
		return false, fmt.Errorf("failed to check user subscription existence: %w", err)
	}

	return count > 0, nil
}

// GetAllSubscriptions retrieves all subscriptions with pagination
func (r *SubscriptionRepositoryImpl) GetAllSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error) {
	var subscriptions []models.Subscription
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to get all subscriptions", err)
		return nil, fmt.Errorf("failed to get all subscriptions: %w", err)
	}

	// Convert to domain entities
	domainSubscriptions := make([]*entities.Subscription, len(subscriptions))
	for i, subscription := range subscriptions {
		domainSubscriptions[i] = r.modelToDomainSubscription(&subscription)
	}

	return domainSubscriptions, nil
}

// GetSubscriptionStats retrieves subscription statistics
func (r *SubscriptionRepositoryImpl) GetSubscriptionStats(ctx context.Context) (*repositories.SubscriptionStats, error) {
	var stats repositories.SubscriptionStats
	
	// Get total subscriptions count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Count(&stats.TotalSubscriptions)
	
	// Get active subscriptions count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, time.Now()).Count(&stats.ActiveSubscriptions)
	
	// Get expired subscriptions count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Count(&stats.ExpiredSubscriptions)
	
	// Get canceled subscriptions count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("is_canceled = ?", true).Count(&stats.CanceledSubscriptions)
	
	// Get basic plan count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("plan = ?", "basic").Count(&stats.BasicSubscriptions)
	
	// Get premium plan count
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("plan = ?", "premium").Count(&stats.PremiumSubscriptions)
	
	// Get subscriptions created today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("DATE(created_at) = DATE(?)", today).Count(&stats.SubscriptionsToday)
	
	// Get subscriptions created this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("created_at >= ?", weekStart).Count(&stats.SubscriptionsThisWeek)
	
	// Get subscriptions created this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Subscription{}).Where("created_at >= ?", monthStart).Count(&stats.SubscriptionsThisMonth)
	
	return &stats, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainSubscription converts model Subscription to domain Subscription
func (r *SubscriptionRepositoryImpl) modelToDomainSubscription(model *models.Subscription) *entities.Subscription {
	return &entities.Subscription{
		ID:                    model.ID,
		UserID:                model.UserID,
		Plan:                  model.Plan,
		IsActive:              model.IsActive,
		ExpiresAt:             model.ExpiresAt,
		IsCanceled:            model.IsCanceled,
		CanceledAt:            model.CanceledAt,
		StripeSubscriptionID:  model.StripeSubscriptionID,
		StripeCustomerID:      model.StripeCustomerID,
		CreatedAt:             model.CreatedAt,
		UpdatedAt:             model.UpdatedAt,
	}
}

// domainToModelSubscription converts domain Subscription to model Subscription
func (r *SubscriptionRepositoryImpl) domainToModelSubscription(subscription *entities.Subscription) *models.Subscription {
	return &models.Subscription{
		ID:                    subscription.ID,
		UserID:                subscription.UserID,
		Plan:                  subscription.Plan,
		IsActive:              subscription.IsActive,
		ExpiresAt:             subscription.ExpiresAt,
		IsCanceled:            subscription.IsCanceled,
		CanceledAt:            subscription.CanceledAt,
		StripeSubscriptionID:  subscription.StripeSubscriptionID,
		StripeCustomerID:      subscription.StripeCustomerID,
		CreatedAt:             subscription.CreatedAt,
		UpdatedAt:             subscription.UpdatedAt,
	}
}