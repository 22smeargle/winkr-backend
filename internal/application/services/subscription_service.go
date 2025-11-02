package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SubscriptionService handles subscription business logic
type SubscriptionService struct {
	userRepo         repositories.UserRepository
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo     repositories.PaymentRepository
	paymentMethodRepo repositories.PaymentMethodRepository
	stripeService    *stripe.StripeService
}

// NewSubscriptionService creates a new SubscriptionService
func NewSubscriptionService(
	userRepo repositories.UserRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
	stripeService *stripe.StripeService,
) *SubscriptionService {
	return &SubscriptionService{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:     paymentRepo,
		paymentMethodRepo: paymentMethodRepo,
		stripeService:    stripeService,
	}
}

// UpgradeSubscription upgrades a user's subscription
func (s *SubscriptionService) UpgradeSubscription(ctx context.Context, userID uuid.UUID, newPlanID string) (*entities.Subscription, error) {
	logger.Info("Upgrading subscription", map[string]interface{}{
		"user_id":  userID,
		"new_plan": newPlanID,
	})

	// Get user's current subscription
	currentSubscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	if currentSubscription == nil {
		logger.Error("User has no subscription", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrUserHasNoActiveSubscription
	}

	// Check if new plan is valid
	newPlan, exists := entities.GetPlanByID(newPlanID)
	if !exists {
		logger.Error("Invalid plan ID", nil, map[string]interface{}{
			"plan_id": newPlanID,
		})
		return nil, ErrInvalidPlanType
	}

	// Check if upgrade is allowed
	if !s.canUpgrade(currentSubscription.PlanType, newPlanID) {
		logger.Error("Upgrade not allowed", nil, map[string]interface{}{
			"user_id":       userID,
			"current_plan": currentSubscription.PlanType,
			"new_plan":     newPlanID,
		})
		return nil, ErrCannotUpgradeSubscription
	}

	// Get Stripe customer
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		logger.Error("User has no Stripe customer ID", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrStripeCustomerNotFound
	}

	// Update subscription in Stripe
	stripeSubscription, err := s.stripeService.UpdateSubscription(ctx, *currentSubscription.StripeSubscriptionID, newPlan.StripePriceID, "create_prorations")
	if err != nil {
		logger.Error("Failed to update Stripe subscription", err, map[string]interface{}{
			"user_id":               userID,
			"stripe_subscription_id": *currentSubscription.StripeSubscriptionID,
			"new_plan":              newPlanID,
		})
		return nil, fmt.Errorf("failed to update Stripe subscription: %w", err)
	}

	// Update subscription in database
	currentSubscription.UpdateFromStripe(
		stripeSubscription.ID,
		newPlanID,
		stripeSubscription.Status,
		stripeSubscription.CurrentPeriodStart,
		stripeSubscription.CurrentPeriodEnd,
		stripeSubscription.CancelAtPeriodEnd,
	)

	err = s.subscriptionRepo.Update(ctx, currentSubscription)
	if err != nil {
		logger.Error("Failed to update subscription", err, map[string]interface{}{
			"user_id":        userID,
			"subscription_id": currentSubscription.ID,
		})
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info("Subscription upgraded successfully", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": currentSubscription.ID,
		"old_plan":       currentSubscription.PlanType,
		"new_plan":       newPlanID,
	})

	return currentSubscription, nil
}

// DowngradeSubscription downgrades a user's subscription
func (s *SubscriptionService) DowngradeSubscription(ctx context.Context, userID uuid.UUID, newPlanID string) (*entities.Subscription, error) {
	logger.Info("Downgrading subscription", map[string]interface{}{
		"user_id":  userID,
		"new_plan": newPlanID,
	})

	// Get user's current subscription
	currentSubscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	if currentSubscription == nil {
		logger.Error("User has no subscription", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrUserHasNoActiveSubscription
	}

	// Check if new plan is valid
	newPlan, exists := entities.GetPlanByID(newPlanID)
	if !exists {
		logger.Error("Invalid plan ID", nil, map[string]interface{}{
			"plan_id": newPlanID,
		})
		return nil, ErrInvalidPlanType
	}

	// Check if downgrade is allowed
	if !s.canDowngrade(currentSubscription.PlanType, newPlanID) {
		logger.Error("Downgrade not allowed", nil, map[string]interface{}{
			"user_id":       userID,
			"current_plan": currentSubscription.PlanType,
			"new_plan":     newPlanID,
		})
		return nil, ErrCannotDowngradeSubscription
	}

	// Get Stripe customer
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		logger.Error("User has no Stripe customer ID", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrStripeCustomerNotFound
	}

	// Update subscription in Stripe
	stripeSubscription, err := s.stripeService.UpdateSubscription(ctx, *currentSubscription.StripeSubscriptionID, newPlan.StripePriceID, "create_prorations")
	if err != nil {
		logger.Error("Failed to update Stripe subscription", err, map[string]interface{}{
			"user_id":               userID,
			"stripe_subscription_id": *currentSubscription.StripeSubscriptionID,
			"new_plan":              newPlanID,
		})
		return nil, fmt.Errorf("failed to update Stripe subscription: %w", err)
	}

	// Update subscription in database
	currentSubscription.UpdateFromStripe(
		stripeSubscription.ID,
		newPlanID,
		stripeSubscription.Status,
		stripeSubscription.CurrentPeriodStart,
		stripeSubscription.CurrentPeriodEnd,
		stripeSubscription.CancelAtPeriodEnd,
	)

	err = s.subscriptionRepo.Update(ctx, currentSubscription)
	if err != nil {
		logger.Error("Failed to update subscription", err, map[string]interface{}{
			"user_id":        userID,
			"subscription_id": currentSubscription.ID,
		})
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info("Subscription downgraded successfully", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": currentSubscription.ID,
		"old_plan":       currentSubscription.PlanType,
		"new_plan":       newPlanID,
	})

	return currentSubscription, nil
}

// ReactivateSubscription reactivates a user's subscription
func (s *SubscriptionService) ReactivateSubscription(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	logger.Info("Reactivating subscription", map[string]interface{}{
		"user_id": userID,
	})

	// Get user's subscription
	subscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}

	if subscription == nil {
		logger.Error("User has no subscription", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrUserHasNoActiveSubscription
	}

	// Check if subscription can be reactivated
	if !subscription.IsCanceled() && !subscription.IsExpired() {
		logger.Error("Subscription cannot be reactivated", nil, map[string]interface{}{
			"user_id":        userID,
			"subscription_id": subscription.ID,
			"status":          subscription.Status,
		})
		return nil, ErrCannotReactivateSubscription
	}

	// Get Stripe customer
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.StripeCustomerID == nil || *user.StripeCustomerID == "" {
		logger.Error("User has no Stripe customer ID", nil, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrStripeCustomerNotFound
	}

	// Reactivate subscription in Stripe
	if subscription.StripeSubscriptionID != nil {
		_, err := s.stripeService.UpdateSubscription(ctx, *subscription.StripeSubscriptionID, "", "none")
		if err != nil {
			logger.Error("Failed to reactivate Stripe subscription", err, map[string]interface{}{
				"user_id":               userID,
				"stripe_subscription_id": *subscription.StripeSubscriptionID,
			})
			return nil, fmt.Errorf("failed to reactivate Stripe subscription: %w", err)
		}
	}

	// Update subscription in database
	subscription.Activate()
	err = s.subscriptionRepo.Update(ctx, subscription)
	if err != nil {
		logger.Error("Failed to update subscription", err, map[string]interface{}{
			"user_id":        userID,
			"subscription_id": subscription.ID,
		})
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info("Subscription reactivated successfully", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": subscription.ID,
	})

	return subscription, nil
}

// ExtendSubscriptionPeriod extends a subscription's billing period
func (s *SubscriptionService) ExtendSubscriptionPeriod(ctx context.Context, userID uuid.UUID, days int) error {
	logger.Info("Extending subscription period", map[string]interface{}{
		"user_id": userID,
		"days":    days,
	})

	// Get user's subscription
	subscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return fmt.Errorf("failed to get user subscription: %w", err)
	}

	if subscription == nil {
		logger.Error("User has no subscription", nil, map[string]interface{}{
			"user_id": userID,
		})
		return ErrUserHasNoActiveSubscription
	}

	// Calculate new end date
	var newEndDate time.Time
	if subscription.CurrentPeriodEnd != nil {
		newEndDate = subscription.CurrentPeriodEnd.AddDate(0, 0, days)
	} else {
		newEndDate = time.Now().AddDate(0, 0, days)
	}

	// Update subscription in database
	err = s.subscriptionRepo.ExtendSubscriptionPeriod(ctx, subscription.ID, days)
	if err != nil {
		logger.Error("Failed to extend subscription period", err, map[string]interface{}{
			"user_id":        userID,
			"subscription_id": subscription.ID,
			"days":           days,
		})
		return fmt.Errorf("failed to extend subscription period: %w", err)
	}

	logger.Info("Subscription period extended successfully", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": subscription.ID,
		"days":           days,
		"new_end_date":   newEndDate,
	})

	return nil
}

// GetSubscriptionAnalytics returns subscription analytics
func (s *SubscriptionService) GetSubscriptionAnalytics(ctx context.Context, startDate, endDate interface{}) (*repositories.SubscriptionStats, error) {
	logger.Info("Getting subscription analytics", map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	})

	stats, err := s.subscriptionRepo.GetSubscriptionStats(ctx)
	if err != nil {
		logger.Error("Failed to get subscription stats", err, nil)
		return nil, fmt.Errorf("failed to get subscription stats: %w", err)
	}

	logger.Info("Retrieved subscription analytics", map[string]interface{}{
		"total_subscriptions": stats.TotalSubscriptions,
		"active_subscriptions":  stats.ActiveSubscriptions,
	})

	return stats, nil
}

// GetUserSubscriptionAnalytics returns subscription analytics for a user
func (s *SubscriptionService) GetUserSubscriptionAnalytics(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (*repositories.UserSubscriptionStats, error) {
	logger.Info("Getting user subscription analytics", map[string]interface{}{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	})

	stats, err := s.subscriptionRepo.GetUserSubscriptionStats(ctx, userID, startDate, endDate)
	if err != nil {
		logger.Error("Failed to get user subscription stats", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, fmt.Errorf("failed to get user subscription stats: %w", err)
	}

	logger.Info("Retrieved user subscription analytics", map[string]interface{}{
		"user_id":              userID,
		"total_subscriptions":   stats.TotalSubscriptions,
		"has_active_subscription": stats.HasActiveSubscription,
	})

	return stats, nil
}

// GetExpiringSubscriptions returns subscriptions expiring soon
func (s *SubscriptionService) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	logger.Info("Getting expiring subscriptions", map[string]interface{}{
		"days": days,
	})

	subscriptions, err := s.subscriptionRepo.GetSubscriptionsExpiringSoon(ctx, days)
	if err != nil {
		logger.Error("Failed to get expiring subscriptions", err, nil)
		return nil, fmt.Errorf("failed to get expiring subscriptions: %w", err)
	}

	logger.Info("Retrieved expiring subscriptions", map[string]interface{}{
		"count": len(subscriptions),
		"days":  days,
	})

	return subscriptions, nil
}

// GetExpiredSubscriptions returns expired subscriptions
func (s *SubscriptionService) GetExpiredSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error) {
	logger.Info("Getting expired subscriptions", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	subscriptions, err := s.subscriptionRepo.GetExpiredSubscriptions(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get expired subscriptions", err, nil)
		return nil, fmt.Errorf("failed to get expired subscriptions: %w", err)
	}

	logger.Info("Retrieved expired subscriptions", map[string]interface{}{
		"count": len(subscriptions),
	})

	return subscriptions, nil
}

// CalculateProration calculates proration amount for subscription change
func (s *SubscriptionService) CalculateProration(ctx context.Context, userID uuid.UUID, newPlanID string) (int64, error) {
	logger.Info("Calculating proration", map[string]interface{}{
		"user_id":  userID,
		"new_plan": newPlanID,
	})

	// Get user's current subscription
	currentSubscription, err := s.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return 0, fmt.Errorf("failed to get user subscription: %w", err)
	}

	if currentSubscription == nil {
		logger.Error("User has no subscription", nil, map[string]interface{}{
			"user_id": userID,
		})
		return 0, ErrUserHasNoActiveSubscription
	}

	// Get plan details
	currentPlan, exists := entities.GetPlanByID(currentSubscription.PlanType)
	if !exists {
		logger.Error("Current plan not found", nil, map[string]interface{}{
			"plan_id": currentSubscription.PlanType,
		})
		return 0, ErrPlanNotFound
	}

	newPlan, exists := entities.GetPlanByID(newPlanID)
	if !exists {
		logger.Error("New plan not found", nil, map[string]interface{}{
			"plan_id": newPlanID,
		})
		return 0, ErrPlanNotFound
	}

	// Calculate proration (simplified calculation)
	// In a real implementation, you would use Stripe's proration API
	currentPrice := int64(currentPlan.Price * 100) // Convert to cents
	newPrice := int64(newPlan.Price * 100)     // Convert to cents

	// Calculate remaining days in current period
	var remainingDays int
	if currentSubscription.CurrentPeriodEnd != nil {
		remaining := time.Until(*currentSubscription.CurrentPeriodEnd)
		if remaining > 0 {
			remainingDays = int(remaining.Hours() / 24)
		}
	}

	// Calculate daily rate and proration
	dailyRate := currentPrice / 30 // Assuming 30-day month
	prorationAmount := (newPrice - currentPrice) * remainingDays / 30

	logger.Info("Calculated proration", map[string]interface{}{
		"user_id":          userID,
		"current_price":    currentPrice,
		"new_price":        newPrice,
		"remaining_days":    remainingDays,
		"proration_amount": prorationAmount,
	})

	return prorationAmount, nil
}

// canUpgrade checks if upgrade is allowed
func (s *SubscriptionService) canUpgrade(currentPlan, newPlan string) bool {
	// Define upgrade hierarchy
	upgradeHierarchy := map[string][]string{
		"basic":     {"premium", "platinum"},
		"premium":   {"platinum"},
		"platinum":  {}, // No upgrade from platinum
	}

	allowedUpgrades, exists := upgradeHierarchy[currentPlan]
	if !exists {
		return false
	}

	for _, allowedPlan := range allowedUpgrades {
		if allowedPlan == newPlan {
			return true
		}
	}

	return false
}

// canDowngrade checks if downgrade is allowed
func (s *SubscriptionService) canDowngrade(currentPlan, newPlan string) bool {
	// Define downgrade hierarchy
	downgradeHierarchy := map[string][]string{
		"platinum":  {"premium", "basic"},
		"premium":   {"basic"},
		"basic":     {}, // No downgrade from basic
	}

	allowedDowngrades, exists := downgradeHierarchy[currentPlan]
	if !exists {
		return false
	}

	for _, allowedPlan := range allowedDowngrades {
		if allowedPlan == newPlan {
			return true
		}
	}

	return false
}