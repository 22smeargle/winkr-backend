package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetSubscriptionUseCase retrieves a user's current subscription
type GetSubscriptionUseCase struct {
	userRepo         repositories.UserRepository
	subscriptionRepo repositories.SubscriptionRepository
}

// NewGetSubscriptionUseCase creates a new GetSubscriptionUseCase
func NewGetSubscriptionUseCase(
	userRepo repositories.UserRepository,
	subscriptionRepo repositories.SubscriptionRepository,
) *GetSubscriptionUseCase {
	return &GetSubscriptionUseCase{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute retrieves a user's current subscription
func (uc *GetSubscriptionUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	logger.Info("Getting user subscription", map[string]interface{}{
		"user_id": userID,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrInvalidUserID
	}

	// Get user's subscription
	subscription, err := uc.subscriptionRepo.GetUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrSubscriptionNotFound
	}

	logger.Info("Retrieved user subscription", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": subscription.ID,
		"plan_type":      subscription.PlanType,
		"status":         subscription.Status,
	})

	return subscription, nil
}

// GetActiveSubscriptionUseCase retrieves a user's active subscription
type GetActiveSubscriptionUseCase struct {
	userRepo         repositories.UserRepository
	subscriptionRepo repositories.SubscriptionRepository
}

// NewGetActiveSubscriptionUseCase creates a new GetActiveSubscriptionUseCase
func NewGetActiveSubscriptionUseCase(
	userRepo repositories.UserRepository,
	subscriptionRepo repositories.SubscriptionRepository,
) *GetActiveSubscriptionUseCase {
	return &GetActiveSubscriptionUseCase{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute retrieves a user's active subscription
func (uc *GetActiveSubscriptionUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	logger.Info("Getting user active subscription", map[string]interface{}{
		"user_id": userID,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrInvalidUserID
	}

	// Get user's active subscription
	subscription, err := uc.subscriptionRepo.GetActiveUserSubscription(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user active subscription", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrUserHasNoActiveSubscription
	}

	logger.Info("Retrieved user active subscription", map[string]interface{}{
		"user_id":        userID,
		"subscription_id": subscription.ID,
		"plan_type":      subscription.PlanType,
		"status":         subscription.Status,
	})

	return subscription, nil
}