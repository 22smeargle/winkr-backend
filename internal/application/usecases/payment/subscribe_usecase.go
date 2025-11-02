package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SubscribeRequest represents a subscription request
type SubscribeRequest struct {
	UserID          uuid.UUID `json:"user_id" validate:"required"`
	PlanID           string    `json:"plan_id" validate:"required"`
	PaymentMethodID   string    `json:"payment_method_id" validate:"required"`
	TrialPeriodDays   int64     `json:"trial_period_days,omitempty"`
}

// SubscribeResponse represents a subscription response
type SubscribeResponse struct {
	SubscriptionID    string `json:"subscription_id"`
	ClientSecret      string `json:"client_secret"`
	Status           string `json:"status"`
	CurrentPeriodStart int64  `json:"current_period_start"`
	CurrentPeriodEnd   int64  `json:"current_period_end"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
}

// SubscribeUseCase handles subscription creation
type SubscribeUseCase struct {
	userRepo         repositories.UserRepository
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo     repositories.PaymentRepository
	paymentMethodRepo repositories.PaymentMethodRepository
	stripeService    *stripe.StripeService
}

// NewSubscribeUseCase creates a new SubscribeUseCase
func NewSubscribeUseCase(
	userRepo repositories.UserRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
	stripeService *stripe.StripeService,
) *SubscribeUseCase {
	return &SubscribeUseCase{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:     paymentRepo,
		paymentMethodRepo: paymentMethodRepo,
		stripeService:    stripeService,
	}
}

// Execute creates a new subscription
func (uc *SubscribeUseCase) Execute(ctx context.Context, req SubscribeRequest) (*SubscribeResponse, error) {
	logger.Info("Creating subscription", map[string]interface{}{
		"user_id":    req.UserID,
		"plan_id":     req.PlanID,
		"payment_method_id": req.PaymentMethodID,
	})

	// Validate plan exists
	plan, exists := entities.GetPlanByID(req.PlanID)
	if !exists {
		logger.Error("Subscription plan not found", nil, map[string]interface{}{
			"plan_id": req.PlanID,
		})
		return nil, ErrPlanNotFound
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": req.UserID,
		})
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user already has an active subscription
	hasActiveSub, err := uc.subscriptionRepo.UserHasActiveSubscription(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to check user active subscription", err, map[string]interface{}{
			"user_id": req.UserID,
		})
		return nil, fmt.Errorf("failed to check user active subscription: %w", err)
	}

	if hasActiveSub {
		logger.Error("User already has active subscription", nil, map[string]interface{}{
			"user_id": req.UserID,
		})
		return nil, ErrUserHasActiveSubscription
	}

	// Get or create Stripe customer
	var stripeCustomerID string
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		stripeCustomerID = *user.StripeCustomerID
	} else {
		// Create new Stripe customer
		customer, err := uc.stripeService.CreateCustomer(ctx, user.Email, fmt.Sprintf("%s %s", user.FirstName, user.LastName), map[string]string{
			"user_id": req.UserID.String(),
		})
		if err != nil {
			logger.Error("Failed to create Stripe customer", err, map[string]interface{}{
				"user_id": req.UserID,
				"email":   user.Email,
			})
			return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
		}
		stripeCustomerID = customer.ID

		// Update user with Stripe customer ID
		user.StripeCustomerID = &stripeCustomerID
		err = uc.userRepo.Update(ctx, user)
		if err != nil {
			logger.Error("Failed to update user with Stripe customer ID", err, map[string]interface{}{
				"user_id":            req.UserID,
				"stripe_customer_id": stripeCustomerID,
			})
			return nil, fmt.Errorf("failed to update user with Stripe customer ID: %w", err)
		}
	}

	// Create subscription in Stripe
	stripeSubscription, err := uc.stripeService.CreateSubscription(ctx, stripeCustomerID, plan.StripePriceID, req.PaymentMethodID, req.TrialPeriodDays, map[string]string{
		"user_id": req.UserID.String(),
		"plan_id": req.PlanID,
	})
	if err != nil {
		logger.Error("Failed to create Stripe subscription", err, map[string]interface{}{
			"user_id":            req.UserID,
			"stripe_customer_id": stripeCustomerID,
			"plan_id":           req.PlanID,
			"payment_method_id":  req.PaymentMethodID,
		})
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}

	// Create subscription entity
	subscription := &entities.Subscription{
		UserID:               req.UserID,
		StripeSubscriptionID:  &stripeSubscription.ID,
		PlanType:             req.PlanID,
		Status:               stripeSubscription.Status,
		CurrentPeriodStart:    &stripeSubscription.CurrentPeriodStart,
		CurrentPeriodEnd:      &stripeSubscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:    stripeSubscription.CancelAtPeriodEnd,
	}

	// Save subscription to database
	err = uc.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		logger.Error("Failed to create subscription", err, map[string]interface{}{
			"user_id":          req.UserID,
			"subscription_id":   subscription.ID,
			"stripe_subscription_id": stripeSubscription.ID,
		})
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Create payment record for initial subscription payment
	payment := &entities.Payment{
		UserID:               req.UserID,
		SubscriptionID:       &subscription.ID,
		StripePaymentIntentID: nil, // Will be set when payment is confirmed
		Amount:               plan.Price * 100, // Convert to cents
		Currency:             plan.Currency,
		Status:               "pending",
		Description:          fmt.Sprintf("Subscription to %s plan", plan.Name),
		Metadata: map[string]string{
			"user_id":        req.UserID.String(),
			"subscription_id": subscription.ID.String(),
			"plan_id":        req.PlanID,
			"stripe_subscription_id": stripeSubscription.ID,
		},
	}

	err = uc.paymentRepo.Create(ctx, payment)
	if err != nil {
		logger.Error("Failed to create payment record", err, map[string]interface{}{
			"user_id":        req.UserID,
			"subscription_id": subscription.ID,
		})
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	logger.Info("Subscription created successfully", map[string]interface{}{
		"user_id":          req.UserID,
		"subscription_id":   subscription.ID,
		"stripe_subscription_id": stripeSubscription.ID,
		"plan_id":          req.PlanID,
	})

	response := &SubscribeResponse{
		SubscriptionID:    stripeSubscription.ID,
		ClientSecret:      "", // Will be set when payment intent is created
		Status:           stripeSubscription.Status,
		CurrentPeriodStart: stripeSubscription.CurrentPeriodStart.Unix(),
		CurrentPeriodEnd:   stripeSubscription.CurrentPeriodEnd.Unix(),
		CancelAtPeriodEnd: stripeSubscription.CancelAtPeriodEnd,
	}

	return response, nil
}