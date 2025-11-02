package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/sub"
	"github.com/stripe/stripe-go/v76/webhook"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// StripeService handles all Stripe-related operations
type StripeService struct {
	client         *stripe.Client
	webhookSecret  string
	secretKey      string
	publishableKey string
}

// NewStripeService creates a new Stripe service instance
func NewStripeService(secretKey, publishableKey, webhookSecret string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{
		client:         stripe.NewClient(secretKey),
		webhookSecret:  webhookSecret,
		secretKey:      secretKey,
		publishableKey: publishableKey,
	}
}

// Customer represents a Stripe customer
type Customer struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Card         *PaymentMethodCard     `json:"card,omitempty"`
	CustomerID   string                 `json:"customer_id"`
	IsDefault    bool                   `json:"is_default"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]string      `json:"metadata"`
}

// PaymentMethodCard represents card details
type PaymentMethodCard struct {
	Brand         string `json:"brand"`
	Last4         string `json:"last4"`
	ExpiryMonth    int64  `json:"expiry_month"`
	ExpiryYear     int64  `json:"expiry_year"`
	Fingerprint    string `json:"fingerprint"`
	Funding        string `json:"funding"`
	Country        string `json:"country"`
	ThreeDSecure   string `json:"three_d_secure"`
}

// Subscription represents a Stripe subscription
type Subscription struct {
	ID                 string                 `json:"id"`
	CustomerID         string                 `json:"customer_id"`
	Status             string                 `json:"status"`
	PlanID             string                 `json:"plan_id"`
	PlanName           string                 `json:"plan_name"`
	Amount             int64                  `json:"amount"`
	Currency           string                 `json:"currency"`
	Interval           string                 `json:"interval"`
	CurrentPeriodStart time.Time              `json:"current_period_start"`
	CurrentPeriodEnd   time.Time              `json:"current_period_end"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	TrialStart         *time.Time             `json:"trial_start,omitempty"`
	TrialEnd           *time.Time             `json:"trial_end,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	Metadata           map[string]string      `json:"metadata"`
}

// PaymentIntent represents a payment intent
type PaymentIntent struct {
	ID              string                 `json:"id"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	Status          string                 `json:"status"`
	ClientSecret    string                 `json:"client_secret"`
	PaymentMethodID *string                `json:"payment_method_id,omitempty"`
	CustomerID      string                 `json:"customer_id"`
	ConfirmationMethod string               `json:"confirmation_method"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]string      `json:"metadata"`
}

// Invoice represents an invoice
type Invoice struct {
	ID              string                 `json:"id"`
	CustomerID      string                 `json:"customer_id"`
	SubscriptionID  *string                `json:"subscription_id,omitempty"`
	Status          string                 `json:"status"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	DueDate         *time.Time             `json:"due_date,omitempty"`
	PaidAt          *time.Time             `json:"paid_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]string      `json:"metadata"`
}

// Refund represents a refund
type Refund struct {
	ID             string                 `json:"id"`
	PaymentIntentID string                 `json:"payment_intent_id"`
	Amount         int64                  `json:"amount"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"`
	Reason         string                 `json:"reason"`
	CreatedAt      time.Time              `json:"created_at"`
	Metadata       map[string]string      `json:"metadata"`
}

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Created int64           `json:"created"`
}

// CreateCustomer creates a new customer in Stripe
func (s *StripeService) CreateCustomer(ctx context.Context, email, name string, metadata map[string]string) (*Customer, error) {
	params := &stripe.CustomerParams{
		Email:   stripe.String(email),
		Name:    stripe.String(name),
		Metadata: metadata,
	}

	cust, err := customer.New(params)
	if err != nil {
		logger.Error("Failed to create Stripe customer", err)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	customer := &Customer{
		ID:        cust.ID,
		Email:     cust.Email,
		Name:      cust.Name,
		Metadata:  cust.Metadata,
		CreatedAt: time.Unix(cust.Created, 0),
	}

	logger.Info("Stripe customer created", map[string]interface{}{
		"customer_id": customer.ID,
		"email":       customer.Email,
	})

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *StripeService) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		logger.Error("Failed to get Stripe customer", err)
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	customer := &Customer{
		ID:        cust.ID,
		Email:     cust.Email,
		Name:      cust.Name,
		Metadata:  cust.Metadata,
		CreatedAt: time.Unix(cust.Created, 0),
	}

	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *StripeService) UpdateCustomer(ctx context.Context, customerID, email, name string, metadata map[string]string) (*Customer, error) {
	params := &stripe.CustomerParams{}
	
	if email != "" {
		params.Email = stripe.String(email)
	}
	if name != "" {
		params.Name = stripe.String(name)
	}
	if metadata != nil {
		params.Metadata = metadata
	}

	cust, err := customer.Update(customerID, params)
	if err != nil {
		logger.Error("Failed to update Stripe customer", err)
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	customer := &Customer{
		ID:        cust.ID,
		Email:     cust.Email,
		Name:      cust.Name,
		Metadata:  cust.Metadata,
		CreatedAt: time.Unix(cust.Created, 0),
	}

	logger.Info("Stripe customer updated", map[string]interface{}{
		"customer_id": customer.ID,
	})

	return customer, nil
}

// DeleteCustomer deletes a customer
func (s *StripeService) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID)
	if err != nil {
		logger.Error("Failed to delete Stripe customer", err)
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	logger.Info("Stripe customer deleted", map[string]interface{}{
		"customer_id": customerID,
	})

	return nil
}

// CreatePaymentIntent creates a new payment intent
func (s *StripeService) CreatePaymentIntent(ctx context.Context, amount int64, currency, customerID, paymentMethodID string, metadata map[string]string) (*PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Customer: stripe.String(customerID),
		Metadata: metadata,
	}

	if paymentMethodID != "" {
		params.PaymentMethod = stripe.String(paymentMethodID)
		params.ConfirmationMethod = stripe.String(string(stripe.PaymentIntentConfirmationMethodManual))
	} else {
		params.ConfirmationMethod = stripe.String(string(stripe.PaymentIntentConfirmationMethodAutomatic))
		params.PaymentMethodTypes = stripe.StringSlice([]string{"card"})
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		logger.Error("Failed to create payment intent", err)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	paymentIntent := &PaymentIntent{
		ID:                 pi.ID,
		Amount:             pi.Amount,
		Currency:           string(pi.Currency),
		Status:             string(pi.Status),
		ClientSecret:       pi.ClientSecret,
		CustomerID:         pi.Customer.ID,
		ConfirmationMethod: string(pi.ConfirmationMethod),
		CreatedAt:          time.Unix(pi.Created, 0),
		Metadata:           pi.Metadata,
	}

	if pi.PaymentMethod != nil {
		paymentIntent.PaymentMethodID = &pi.PaymentMethod.ID
	}

	logger.Info("Payment intent created", map[string]interface{}{
		"payment_intent_id": paymentIntent.ID,
		"amount":           paymentIntent.Amount,
		"currency":         paymentIntent.Currency,
	})

	return paymentIntent, nil
}

// ConfirmPaymentIntent confirms a payment intent
func (s *StripeService) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntent, error) {
	pi, err := paymentintent.Confirm(paymentIntentID, nil)
	if err != nil {
		logger.Error("Failed to confirm payment intent", err)
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	paymentIntent := &PaymentIntent{
		ID:                 pi.ID,
		Amount:             pi.Amount,
		Currency:           string(pi.Currency),
		Status:             string(pi.Status),
		ClientSecret:       pi.ClientSecret,
		CustomerID:         pi.Customer.ID,
		ConfirmationMethod: string(pi.ConfirmationMethod),
		CreatedAt:          time.Unix(pi.Created, 0),
		Metadata:           pi.Metadata,
	}

	if pi.PaymentMethod != nil {
		paymentIntent.PaymentMethodID = &pi.PaymentMethod.ID
	}

	logger.Info("Payment intent confirmed", map[string]interface{}{
		"payment_intent_id": paymentIntent.ID,
		"status":           paymentIntent.Status,
	})

	return paymentIntent, nil
}

// CreatePaymentMethod creates a new payment method
func (s *StripeService) CreatePaymentMethod(ctx context.Context, paymentMethodType, customerID string, cardDetails map[string]interface{}) (*PaymentMethod, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(paymentMethodType),
	}

	if paymentMethodType == "card" {
		params.Card = &stripe.PaymentMethodCardParams{}
		if number, ok := cardDetails["number"].(string); ok {
			params.Card.Number = stripe.String(number)
		}
		if expMonth, ok := cardDetails["exp_month"].(int64); ok {
			params.Card.ExpMonth = stripe.Int64(expMonth)
		}
		if expYear, ok := cardDetails["exp_year"].(int64); ok {
			params.Card.ExpYear = stripe.Int64(expYear)
		}
		if cvc, ok := cardDetails["cvc"].(string); ok {
			params.Card.CVC = stripe.String(cvc)
		}
	}

	pm, err := paymentmethod.New(params)
	if err != nil {
		logger.Error("Failed to create payment method", err)
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	// Attach payment method to customer if customer ID is provided
	if customerID != "" {
		_, err = paymentmethod.Attach(pm.ID, &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(customerID),
		})
		if err != nil {
			logger.Error("Failed to attach payment method to customer", err)
			return nil, fmt.Errorf("failed to attach payment method: %w", err)
		}
	}

	paymentMethod := &PaymentMethod{
		ID:         pm.ID,
		Type:       string(pm.Type),
		CustomerID: customerID,
		CreatedAt:  time.Unix(pm.Created, 0),
		Metadata:   pm.Metadata,
	}

	if pm.Card != nil {
		paymentMethod.Card = &PaymentMethodCard{
			Brand:       string(pm.Card.Brand),
			Last4:       pm.Card.Last4,
			ExpiryMonth: pm.Card.ExpMonth,
			ExpiryYear:  pm.Card.ExpYear,
			Fingerprint: pm.Card.Fingerprint,
			Funding:     string(pm.Card.Funding),
			Country:     pm.Card.Country,
		}
	}

	logger.Info("Payment method created", map[string]interface{}{
		"payment_method_id": paymentMethod.ID,
		"type":             paymentMethod.Type,
		"customer_id":      customerID,
	})

	return paymentMethod, nil
}

// GetPaymentMethods retrieves payment methods for a customer
func (s *StripeService) GetPaymentMethods(ctx context.Context, customerID, paymentMethodType string) ([]*PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(paymentMethodType),
	}

	iter := paymentmethod.List(params)
	var paymentMethods []*PaymentMethod

	for iter.Next() {
		pm := iter.PaymentMethod()
		
		paymentMethod := &PaymentMethod{
			ID:         pm.ID,
			Type:       string(pm.Type),
			CustomerID: customerID,
			CreatedAt:  time.Unix(pm.Created, 0),
			Metadata:   pm.Metadata,
		}

		if pm.Card != nil {
			paymentMethod.Card = &PaymentMethodCard{
				Brand:       string(pm.Card.Brand),
				Last4:       pm.Card.Last4,
				ExpiryMonth: pm.Card.ExpMonth,
				ExpiryYear:  pm.Card.ExpYear,
				Fingerprint: pm.Card.Fingerprint,
				Funding:     string(pm.Card.Funding),
				Country:     pm.Card.Country,
			}
		}

		paymentMethods = append(paymentMethods, paymentMethod)
	}

	if err := iter.Err(); err != nil {
		logger.Error("Failed to get payment methods", err)
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}

	return paymentMethods, nil
}

// DeletePaymentMethod deletes a payment method
func (s *StripeService) DeletePaymentMethod(ctx context.Context, paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		logger.Error("Failed to delete payment method", err)
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	logger.Info("Payment method deleted", map[string]interface{}{
		"payment_method_id": paymentMethodID,
	})

	return nil
}

// CreateSubscription creates a new subscription
func (s *StripeService) CreateSubscription(ctx context.Context, customerID, priceID string, paymentMethodID string, trialPeriodDays int64, metadata map[string]string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Metadata: metadata,
	}

	if paymentMethodID != "" {
		params.DefaultPaymentMethod = stripe.String(paymentMethodID)
	}

	if trialPeriodDays > 0 {
		params.TrialPeriodDays = stripe.Int64(trialPeriodDays)
	}

	sub, err := sub.New(params)
	if err != nil {
		logger.Error("Failed to create subscription", err)
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	subscription := &Subscription{
		ID:                sub.ID,
		CustomerID:        sub.Customer.ID,
		Status:            string(sub.Status),
		CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CreatedAt:         time.Unix(sub.Created, 0),
		Metadata:          sub.Metadata,
	}

	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		subscription.PlanID = item.Price.ID
		subscription.PlanName = item.Price.Nickname
		subscription.Amount = item.Price.UnitAmount
		subscription.Currency = string(item.Price.Currency)
		subscription.Interval = string(item.Price.Recurring.Interval)
	}

	if sub.TrialStart != 0 {
		trialStart := time.Unix(sub.TrialStart, 0)
		subscription.TrialStart = &trialStart
	}

	if sub.TrialEnd != 0 {
		trialEnd := time.Unix(sub.TrialEnd, 0)
		subscription.TrialEnd = &trialEnd
	}

	logger.Info("Subscription created", map[string]interface{}{
		"subscription_id": subscription.ID,
		"customer_id":    subscription.CustomerID,
		"plan_id":        subscription.PlanID,
	})

	return subscription, nil
}

// GetSubscription retrieves a subscription by ID
func (s *StripeService) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	sub, err := sub.Get(subscriptionID, nil)
	if err != nil {
		logger.Error("Failed to get subscription", err)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	subscription := &Subscription{
		ID:                sub.ID,
		CustomerID:        sub.Customer.ID,
		Status:            string(sub.Status),
		CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CreatedAt:         time.Unix(sub.Created, 0),
		Metadata:          sub.Metadata,
	}

	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		subscription.PlanID = item.Price.ID
		subscription.PlanName = item.Price.Nickname
		subscription.Amount = item.Price.UnitAmount
		subscription.Currency = string(item.Price.Currency)
		subscription.Interval = string(item.Price.Recurring.Interval)
	}

	if sub.TrialStart != 0 {
		trialStart := time.Unix(sub.TrialStart, 0)
		subscription.TrialStart = &trialStart
	}

	if sub.TrialEnd != 0 {
		trialEnd := time.Unix(sub.TrialEnd, 0)
		subscription.TrialEnd = &trialEnd
	}

	return subscription, nil
}

// UpdateSubscription updates an existing subscription
func (s *StripeService) UpdateSubscription(ctx context.Context, subscriptionID, priceID string, prorationBehavior string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(subscriptionID),
				Price: stripe.String(priceID),
			},
		},
	}

	if prorationBehavior != "" {
		params.ProrationBehavior = stripe.String(prorationBehavior)
	}

	sub, err := sub.Update(subscriptionID, params)
	if err != nil {
		logger.Error("Failed to update subscription", err)
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	subscription := &Subscription{
		ID:                sub.ID,
		CustomerID:        sub.Customer.ID,
		Status:            string(sub.Status),
		CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CreatedAt:         time.Unix(sub.Created, 0),
		Metadata:          sub.Metadata,
	}

	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		subscription.PlanID = item.Price.ID
		subscription.PlanName = item.Price.Nickname
		subscription.Amount = item.Price.UnitAmount
		subscription.Currency = string(item.Price.Currency)
		subscription.Interval = string(item.Price.Recurring.Interval)
	}

	logger.Info("Subscription updated", map[string]interface{}{
		"subscription_id": subscription.ID,
		"plan_id":        subscription.PlanID,
	})

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (s *StripeService) CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) (*Subscription, error) {
	var sub *stripe.Subscription
	var err error

	if cancelAtPeriodEnd {
		// Cancel at period end
		sub, err = sub.Update(subscriptionID, &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		})
	} else {
		// Cancel immediately
		sub, err = sub.Cancel(subscriptionID, nil)
	}

	if err != nil {
		logger.Error("Failed to cancel subscription", err)
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	subscription := &Subscription{
		ID:                sub.ID,
		CustomerID:        sub.Customer.ID,
		Status:            string(sub.Status),
		CurrentPeriodStart: time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CreatedAt:         time.Unix(sub.Created, 0),
		Metadata:          sub.Metadata,
	}

	if len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		subscription.PlanID = item.Price.ID
		subscription.PlanName = item.Price.Nickname
		subscription.Amount = item.Price.UnitAmount
		subscription.Currency = string(item.Price.Currency)
		subscription.Interval = string(item.Price.Recurring.Interval)
	}

	logger.Info("Subscription canceled", map[string]interface{}{
		"subscription_id":   subscription.ID,
		"cancel_at_period_end": cancelAtPeriodEnd,
	})

	return subscription, nil
}

// CreateRefund creates a refund
func (s *StripeService) CreateRefund(ctx context.Context, paymentIntentID string, amount int64, reason string, metadata map[string]string) (*Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Metadata:     metadata,
	}

	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	if reason != "" {
		params.Reason = stripe.String(reason)
	}

	refund, err := refund.New(params)
	if err != nil {
		logger.Error("Failed to create refund", err)
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	refundObj := &Refund{
		ID:             refund.ID,
		PaymentIntentID: refund.PaymentIntent.ID,
		Amount:         refund.Amount,
		Currency:       string(refund.Currency),
		Status:         string(refund.Status),
		Reason:         string(refund.Reason),
		CreatedAt:      time.Unix(refund.Created, 0),
		Metadata:       refund.Metadata,
	}

	logger.Info("Refund created", map[string]interface{}{
		"refund_id":        refundObj.ID,
		"payment_intent_id": refundObj.PaymentIntentID,
		"amount":           refundObj.Amount,
	})

	return refundObj, nil
}

// GetInvoice retrieves an invoice by ID
func (s *StripeService) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		logger.Error("Failed to get invoice", err)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	invoice := &Invoice{
		ID:         inv.ID,
		CustomerID: inv.Customer.ID,
		Status:     string(inv.Status),
		Amount:     inv.AmountDue,
		Currency:   string(inv.Currency),
		CreatedAt:  time.Unix(inv.Created, 0),
		Metadata:   inv.Metadata,
	}

	if inv.Subscription != nil {
		subscriptionID := inv.Subscription.ID
		invoice.SubscriptionID = &subscriptionID
	}

	if inv.DueDate != 0 {
		dueDate := time.Unix(inv.DueDate, 0)
		invoice.DueDate = &dueDate
	}

	if inv.StatusTransitions != nil && inv.StatusTransitions.PaidAt != 0 {
		paidAt := time.Unix(inv.StatusTransitions.PaidAt, 0)
		invoice.PaidAt = &paidAt
	}

	return invoice, nil
}

// VerifyWebhook verifies and parses a webhook event
func (s *StripeService) VerifyWebhook(ctx context.Context, payload []byte, signatureHeader string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		logger.Error("Failed to verify webhook signature", err)
		return nil, fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	webhookEvent := &WebhookEvent{
		ID:      event.ID,
		Type:    event.Type,
		Data:    event.Data.Raw,
		Created: event.Created,
	}

	logger.Info("Webhook event verified", map[string]interface{}{
		"event_id": webhookEvent.ID,
		"type":     webhookEvent.Type,
	})

	return webhookEvent, nil
}

// CreateProduct creates a new product
func (s *StripeService) CreateProduct(ctx context.Context, name, description string, metadata map[string]string) (string, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Metadata:    metadata,
	}

	product, err := product.New(params)
	if err != nil {
		logger.Error("Failed to create product", err)
		return "", fmt.Errorf("failed to create product: %w", err)
	}

	logger.Info("Product created", map[string]interface{}{
		"product_id": product.ID,
		"name":       name,
	})

	return product.ID, nil
}

// CreatePrice creates a new price for a product
func (s *StripeService) CreatePrice(ctx context.Context, productID, nickname string, amount int64, currency, recurringInterval string, metadata map[string]string) (string, error) {
	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		Nickname:   stripe.String(nickname),
		UnitAmount: stripe.Int64(amount),
		Currency:   stripe.String(currency),
		Metadata:   metadata,
	}

	if recurringInterval != "" {
		params.Recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String(recurringInterval),
		}
	}

	price, err := price.New(params)
	if err != nil {
		logger.Error("Failed to create price", err)
		return "", fmt.Errorf("failed to create price: %w", err)
	}

	logger.Info("Price created", map[string]interface{}{
		"price_id": price.ID,
		"product_id": productID,
		"amount": amount,
		"currency": currency,
	})

	return price.ID, nil
}

// GetPublishableKey returns the publishable key
func (s *StripeService) GetPublishableKey() string {
	return s.publishableKey
}

// GetSecretKey returns the secret key
func (s *StripeService) GetSecretKey() string {
	return s.secretKey
}