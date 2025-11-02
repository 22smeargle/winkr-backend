package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AddPaymentMethodRequest represents a payment method addition request
type AddPaymentMethodRequest struct {
	UserID        uuid.UUID              `json:"user_id" validate:"required"`
	Type          string                 `json:"type" validate:"required,oneof=card bank_account sepa_debit"`
	CardNumber    string                 `json:"card_number,omitempty"`
	ExpiryMonth  int64                  `json:"expiry_month,omitempty"`
	ExpiryYear   int64                  `json:"expiry_year,omitempty"`
	CVC           string                 `json:"cvc,omitempty"`
	IsDefault     bool                   `json:"is_default"`
}

// AddPaymentMethodUseCase handles payment method addition
type AddPaymentMethodUseCase struct {
	userRepo         repositories.UserRepository
	paymentMethodRepo repositories.PaymentMethodRepository
	stripeService    *stripe.StripeService
}

// NewAddPaymentMethodUseCase creates a new AddPaymentMethodUseCase
func NewAddPaymentMethodUseCase(
	userRepo repositories.UserRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
	stripeService *stripe.StripeService,
) *AddPaymentMethodUseCase {
	return &AddPaymentMethodUseCase{
		userRepo:         userRepo,
		paymentMethodRepo: paymentMethodRepo,
		stripeService:    stripeService,
	}
}

// Execute adds a new payment method
func (uc *AddPaymentMethodUseCase) Execute(ctx context.Context, req AddPaymentMethodRequest) (*entities.PaymentMethod, error) {
	logger.Info("Adding payment method", map[string]interface{}{
		"user_id": req.UserID,
		"type":    req.Type,
	})

	// Validate user exists
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": req.UserID,
		})
		return nil, ErrInvalidUserID
	}

	// Get or create Stripe customer
	var stripeCustomerID string
	if user.StripeCustomerID != nil && *user.StripeCustomerID != "" {
		stripeCustomerID = *user.StripeCustomerID
	} else {
		// Create new Stripe customer
		customer, err := uc.stripeService.CreateCustomer(ctx, user.Email, user.FirstName+" "+user.LastName, map[string]string{
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

	// Prepare card details for Stripe
	cardDetails := make(map[string]interface{})
	if req.Type == "card" {
		cardDetails["number"] = req.CardNumber
		cardDetails["exp_month"] = req.ExpiryMonth
		cardDetails["exp_year"] = req.ExpiryYear
		cardDetails["cvc"] = req.CVC
	}

	// Create payment method in Stripe
	stripePaymentMethod, err := uc.stripeService.CreatePaymentMethod(ctx, req.Type, stripeCustomerID, cardDetails)
	if err != nil {
		logger.Error("Failed to create Stripe payment method", err, map[string]interface{}{
			"user_id":            req.UserID,
			"stripe_customer_id": stripeCustomerID,
			"type":              req.Type,
		})
		return nil, fmt.Errorf("failed to create Stripe payment method: %w", err)
	}

	// Create payment method entity
	paymentMethod := &entities.PaymentMethod{
		UserID:              req.UserID,
		StripePaymentMethodID: &stripePaymentMethod.ID,
		Type:                 req.Type,
		IsDefault:            req.IsDefault,
		IsVerified:           false, // Will be verified by Stripe webhook
		Metadata: map[string]string{
			"user_id":              req.UserID.String(),
			"stripe_customer_id":    stripeCustomerID,
			"stripe_payment_method_id": stripePaymentMethod.ID,
		},
	}

	// Set card details if available
	if stripePaymentMethod.Card != nil {
		paymentMethod.CardBrand = &stripePaymentMethod.Card.Brand
		paymentMethod.CardLast4 = &stripePaymentMethod.Card.Last4
		paymentMethod.CardExpiryMonth = &stripePaymentMethod.Card.ExpiryMonth
		paymentMethod.CardExpiryYear = &stripePaymentMethod.Card.ExpiryYear
		paymentMethod.CardFingerprint = &stripePaymentMethod.Card.Fingerprint
	}

	// If setting as default, unset other default payment methods
	if req.IsDefault {
		err = uc.paymentMethodRepo.UnsetDefault(ctx, req.UserID)
		if err != nil {
			logger.Error("Failed to unset default payment methods", err, map[string]interface{}{
				"user_id": req.UserID,
			})
			return nil, fmt.Errorf("failed to unset default payment methods: %w", err)
		}
	}

	// Save payment method to database
	err = uc.paymentMethodRepo.Create(ctx, paymentMethod)
	if err != nil {
		logger.Error("Failed to create payment method", err, map[string]interface{}{
			"user_id":        req.UserID,
			"payment_method_id": paymentMethod.ID,
		})
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	logger.Info("Payment method added successfully", map[string]interface{}{
		"user_id":          req.UserID,
		"payment_method_id": paymentMethod.ID,
		"type":             paymentMethod.Type,
		"is_default":       paymentMethod.IsDefault,
	})

	return paymentMethod, nil
}