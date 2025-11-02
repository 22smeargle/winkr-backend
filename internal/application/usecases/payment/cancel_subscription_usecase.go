package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// CancelSubscriptionRequest represents a subscription cancellation request
type CancelSubscriptionRequest struct {
	UserID          uuid.UUID `json:"user_id" validate:"required"`
	CancelAtPeriodEnd bool      `json:"cancel_at_period_end"`
	Reason           string    `json:"reason,omitempty"`
}

// CancelSubscriptionUseCase handles subscription cancellation
type CancelSubscriptionUseCase struct {
	userRepo         repositories.UserRepository
	subscriptionRepo repositories.SubscriptionRepository
	stripeService    *stripe.StripeService
}

// NewCancelSubscriptionUseCase creates a new CancelSubscriptionUseCase
func NewCancelSubscriptionUseCase(
	userRepo repositories.UserRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	stripeService *stripe.StripeService,
) *CancelSubscriptionUseCase {
	return &CancelSubscriptionUseCase{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		stripeService:    stripeService,
	}
}

// Execute cancels a user's subscription
func (uc *CancelSubscriptionUseCase) Execute(ctx context.Context, req CancelSubscriptionRequest) error {
	logger.Info("Canceling subscription", map[string]interface{}{
		"user_id":            req.UserID,
		"cancel_at_period_end": req.CancelAtPeriodEnd,
		"reason":              req.Reason,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": req.UserID,
		})
		return ErrInvalidUserID
	}

	// Get user's subscription
	subscription, err := uc.subscriptionRepo.GetUserSubscription(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": req.UserID,
		})
		return ErrSubscriptionNotFound
	}

	// Check if subscription can be canceled
	if subscription.IsCanceled() {
		logger.Error("Subscription already canceled", nil, map[string]interface{}{
			"user_id":        req.UserID,
			"subscription_id": subscription.ID,
		})
		return ErrSubscriptionCanceled
	}

	if subscription.IsExpired() {
		logger.Error("Subscription already expired", nil, map[string]interface{}{
			"user_id":        req.UserID,
			"subscription_id": subscription.ID,
		})
		return ErrSubscriptionExpired
	}

	// Cancel subscription in Stripe
	if subscription.StripeSubscriptionID != nil {
		_, err := uc.stripeService.CancelSubscription(ctx, *subscription.StripeSubscriptionID, req.CancelAtPeriodEnd)
		if err != nil {
			logger.Error("Failed to cancel Stripe subscription", err, map[string]interface{}{
				"user_id":               req.UserID,
				"subscription_id":        subscription.ID,
				"stripe_subscription_id": *subscription.StripeSubscriptionID,
			})
			return fmt.Errorf("failed to cancel Stripe subscription: %w", err)
		}
	}

	// Update subscription in database
	if req.CancelAtPeriodEnd {
		subscription.SetCancelAtPeriodEnd(true)
	} else {
		subscription.Cancel()
	}

	err = uc.subscriptionRepo.Update(ctx, subscription)
	if err != nil {
		logger.Error("Failed to update subscription", err, map[string]interface{}{
			"user_id":        req.UserID,
			"subscription_id": subscription.ID,
		})
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	logger.Info("Subscription canceled successfully", map[string]interface{}{
		"user_id":               req.UserID,
		"subscription_id":        subscription.ID,
		"stripe_subscription_id": subscription.StripeSubscriptionID,
		"cancel_at_period_end":  req.CancelAtPeriodEnd,
		"reason":                req.Reason,
	})

	return nil
}