package payment

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ProcessWebhookUseCase processes Stripe webhooks
type ProcessWebhookUseCase struct {
	webhookEventRepo repositories.WebhookEventRepository
	subscriptionRepo repositories.SubscriptionRepository
	paymentRepo     repositories.PaymentRepository
	paymentMethodRepo repositories.PaymentMethodRepository
	invoiceRepo     repositories.InvoiceRepository
	refundRepo      repositories.RefundRepository
	stripeService    *stripe.StripeService
}

// NewProcessWebhookUseCase creates a new ProcessWebhookUseCase
func NewProcessWebhookUseCase(
	webhookEventRepo repositories.WebhookEventRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	paymentRepo repositories.PaymentRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
	invoiceRepo repositories.InvoiceRepository,
	refundRepo repositories.RefundRepository,
	stripeService *stripe.StripeService,
) *ProcessWebhookUseCase {
	return &ProcessWebhookUseCase{
		webhookEventRepo: webhookEventRepo,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:     paymentRepo,
		paymentMethodRepo: paymentMethodRepo,
		invoiceRepo:     invoiceRepo,
		refundRepo:      refundRepo,
		stripeService:    stripeService,
	}
}

// Execute processes a Stripe webhook event
func (uc *ProcessWebhookUseCase) Execute(ctx context.Context, payload []byte, signatureHeader string) error {
	logger.Info("Processing webhook event", nil)

	// Verify webhook signature
	event, err := uc.stripeService.VerifyWebhook(ctx, payload, signatureHeader)
	if err != nil {
		logger.Error("Failed to verify webhook signature", err, nil)
		return ErrWebhookSignatureInvalid
	}

	// Check if event already processed
	existingEvent, err := uc.webhookEventRepo.GetByStripeEventID(ctx, event.ID)
	if err != nil {
		logger.Error("Failed to check existing webhook event", err, map[string]interface{}{
			"stripe_event_id": event.ID,
		})
		return fmt.Errorf("failed to check existing webhook event: %w", err)
	}

	if existingEvent != nil && existingEvent.IsProcessed() {
		logger.Info("Webhook event already processed", map[string]interface{}{
			"stripe_event_id": event.ID,
		})
		return ErrWebhookAlreadyProcessed
	}

	// Create webhook event record
	webhookEvent := &entities.WebhookEvent{
		StripeEventID: event.ID,
		Type:          event.Type,
		Processed:     false,
		RawData:       string(event.Data),
	}

	err = uc.webhookEventRepo.Create(ctx, webhookEvent)
	if err != nil {
		logger.Error("Failed to create webhook event record", err, map[string]interface{}{
			"stripe_event_id": event.ID,
			"type":            event.Type,
		})
		return fmt.Errorf("failed to create webhook event record: %w", err)
	}

	// Process event based on type
	processErr := uc.processEventByType(ctx, event)

	// Update webhook event record
	if processErr != nil {
		webhookEvent.SetProcessingError(processErr.Error())
		logger.Error("Failed to process webhook event", processErr, map[string]interface{}{
			"stripe_event_id": event.ID,
			"type":            event.Type,
		})
	} else {
		webhookEvent.SetProcessed()
		logger.Info("Webhook event processed successfully", map[string]interface{}{
			"stripe_event_id": event.ID,
			"type":            event.Type,
		})
	}

	err = uc.webhookEventRepo.Update(ctx, webhookEvent)
	if err != nil {
		logger.Error("Failed to update webhook event record", err, map[string]interface{}{
			"stripe_event_id": event.ID,
		})
		return fmt.Errorf("failed to update webhook event record: %w", err)
	}

	return processErr
}

// processEventByType processes webhook event based on its type
func (uc *ProcessWebhookUseCase) processEventByType(ctx context.Context, event *stripe.WebhookEvent) error {
	switch event.Type {
	case "customer.subscription.created":
		return uc.handleSubscriptionCreated(ctx, event.Data)
	case "customer.subscription.updated":
		return uc.handleSubscriptionUpdated(ctx, event.Data)
	case "customer.subscription.deleted":
		return uc.handleSubscriptionDeleted(ctx, event.Data)
	case "invoice.payment_succeeded":
		return uc.handleInvoicePaymentSucceeded(ctx, event.Data)
	case "invoice.payment_failed":
		return uc.handleInvoicePaymentFailed(ctx, event.Data)
	case "invoice.created":
		return uc.handleInvoiceCreated(ctx, event.Data)
	case "invoice.updated":
		return uc.handleInvoiceUpdated(ctx, event.Data)
	case "payment_intent.succeeded":
		return uc.handlePaymentIntentSucceeded(ctx, event.Data)
	case "payment_intent.payment_failed":
		return uc.handlePaymentIntentFailed(ctx, event.Data)
	case "payment_intent.canceled":
		return uc.handlePaymentIntentCanceled(ctx, event.Data)
	case "charge.succeeded":
		return uc.handleChargeSucceeded(ctx, event.Data)
	case "charge.failed":
		return uc.handleChargeFailed(ctx, event.Data)
	case "charge.dispute.created":
		return uc.handleChargeDisputeCreated(ctx, event.Data)
	default:
		logger.Info("Webhook event type not supported", map[string]interface{}{
			"type": event.Type,
		})
		return ErrWebhookEventNotSupported
	}
}

// handleSubscriptionCreated handles subscription.created webhook event
func (uc *ProcessWebhookUseCase) handleSubscriptionCreated(ctx context.Context, rawData json.RawMessage) error {
	var subscriptionData struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal(rawData, &subscriptionData)
	if err != nil {
		logger.Error("Failed to unmarshal subscription data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Get subscription from Stripe
	stripeSubscription, err := uc.stripeService.GetSubscription(ctx, subscriptionData.ID)
	if err != nil {
		logger.Error("Failed to get Stripe subscription", err, map[string]interface{}{
			"subscription_id": subscriptionData.ID,
		})
		return fmt.Errorf("failed to get Stripe subscription: %w", err)
	}

	// Update subscription in database
	subscription, err := uc.subscriptionRepo.GetByStripeSubscriptionID(ctx, subscriptionData.ID)
	if err != nil {
		logger.Error("Failed to get subscription by Stripe ID", err, map[string]interface{}{
			"stripe_subscription_id": subscriptionData.ID,
		})
		return fmt.Errorf("failed to get subscription by Stripe ID: %w", err)
	}

	if subscription == nil {
		logger.Info("Subscription not found in database, creating new record", map[string]interface{}{
			"stripe_subscription_id": subscriptionData.ID,
		})
		// Create new subscription record
		subscription = &entities.Subscription{
			StripeSubscriptionID:  &stripeSubscription.ID,
			PlanType:             uc.mapStripePlanToInternal(stripeSubscription.PlanID),
			Status:               stripeSubscription.Status,
			CurrentPeriodStart:    &stripeSubscription.CurrentPeriodStart,
			CurrentPeriodEnd:      &stripeSubscription.CurrentPeriodEnd,
			CancelAtPeriodEnd:    stripeSubscription.CancelAtPeriodEnd,
		}

		err = uc.subscriptionRepo.Create(ctx, subscription)
		if err != nil {
			logger.Error("Failed to create subscription", err, map[string]interface{}{
				"stripe_subscription_id": subscriptionData.ID,
			})
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	} else {
		// Update existing subscription
		subscription.UpdateFromStripe(
			stripeSubscription.ID,
			uc.mapStripePlanToInternal(stripeSubscription.PlanID),
			stripeSubscription.Status,
			stripeSubscription.CurrentPeriodStart,
			stripeSubscription.CurrentPeriodEnd,
			stripeSubscription.CancelAtPeriodEnd,
		)

		err = uc.subscriptionRepo.Update(ctx, subscription)
		if err != nil {
			logger.Error("Failed to update subscription", err, map[string]interface{}{
				"subscription_id":        subscription.ID,
				"stripe_subscription_id": subscriptionData.ID,
			})
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// handleSubscriptionUpdated handles subscription.updated webhook event
func (uc *ProcessWebhookUseCase) handleSubscriptionUpdated(ctx context.Context, rawData json.RawMessage) error {
	return uc.handleSubscriptionCreated(ctx, rawData) // Same logic as created
}

// handleSubscriptionDeleted handles subscription.deleted webhook event
func (uc *ProcessWebhookUseCase) handleSubscriptionDeleted(ctx context.Context, rawData json.RawMessage) error {
	var subscriptionData struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal(rawData, &subscriptionData)
	if err != nil {
		logger.Error("Failed to unmarshal subscription data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update subscription in database
	subscription, err := uc.subscriptionRepo.GetByStripeSubscriptionID(ctx, subscriptionData.ID)
	if err != nil {
		logger.Error("Failed to get subscription by Stripe ID", err, map[string]interface{}{
			"stripe_subscription_id": subscriptionData.ID,
		})
		return fmt.Errorf("failed to get subscription by Stripe ID: %w", err)
	}

	if subscription != nil {
		subscription.Cancel()
		err = uc.subscriptionRepo.Update(ctx, subscription)
		if err != nil {
			logger.Error("Failed to update subscription", err, map[string]interface{}{
				"subscription_id":        subscription.ID,
				"stripe_subscription_id": subscriptionData.ID,
			})
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// handleInvoicePaymentSucceeded handles invoice.payment_succeeded webhook event
func (uc *ProcessWebhookUseCase) handleInvoicePaymentSucceeded(ctx context.Context, rawData json.RawMessage) error {
	var invoiceData struct {
		ID           string `json:"id"`
		Subscription string `json:"subscription"`
		Payment      string `json:"payment"`
	}

	err := json.Unmarshal(rawData, &invoiceData)
	if err != nil {
		logger.Error("Failed to unmarshal invoice data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Get invoice from Stripe
	stripeInvoice, err := uc.stripeService.GetInvoice(ctx, invoiceData.ID)
	if err != nil {
		logger.Error("Failed to get Stripe invoice", err, map[string]interface{}{
			"invoice_id": invoiceData.ID,
		})
		return fmt.Errorf("failed to get Stripe invoice: %w", err)
	}

	// Update payment status
	if invoiceData.Payment != "" {
		payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, invoiceData.Payment)
		if err != nil {
			logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
				"payment_intent_id": invoiceData.Payment,
			})
			return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
		}

		if payment != nil {
			payment.SetSucceeded()
			err = uc.paymentRepo.Update(ctx, payment)
			if err != nil {
				logger.Error("Failed to update payment", err, map[string]interface{}{
					"payment_id": payment.ID,
				})
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}
	}

	return nil
}

// handleInvoicePaymentFailed handles invoice.payment_failed webhook event
func (uc *ProcessWebhookUseCase) handleInvoicePaymentFailed(ctx context.Context, rawData json.RawMessage) error {
	var invoiceData struct {
		ID      string `json:"id"`
		Payment string `json:"payment"`
	}

	err := json.Unmarshal(rawData, &invoiceData)
	if err != nil {
		logger.Error("Failed to unmarshal invoice data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment status
	if invoiceData.Payment != "" {
		payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, invoiceData.Payment)
		if err != nil {
			logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
				"payment_intent_id": invoiceData.Payment,
			})
			return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
		}

		if payment != nil {
			payment.SetFailed("Payment failed")
			err = uc.paymentRepo.Update(ctx, payment)
			if err != nil {
				logger.Error("Failed to update payment", err, map[string]interface{}{
					"payment_id": payment.ID,
				})
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}
	}

	return nil
}

// handleInvoiceCreated handles invoice.created webhook event
func (uc *ProcessWebhookUseCase) handleInvoiceCreated(ctx context.Context, rawData json.RawMessage) error {
	var invoiceData struct {
		ID           string `json:"id"`
		Subscription string `json:"subscription"`
	}

	err := json.Unmarshal(rawData, &invoiceData)
	if err != nil {
		logger.Error("Failed to unmarshal invoice data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Get invoice from Stripe
	stripeInvoice, err := uc.stripeService.GetInvoice(ctx, invoiceData.ID)
	if err != nil {
		logger.Error("Failed to get Stripe invoice", err, map[string]interface{}{
			"invoice_id": invoiceData.ID,
		})
		return fmt.Errorf("failed to get Stripe invoice: %w", err)
	}

	// Create or update invoice in database
	invoice, err := uc.invoiceRepo.GetByStripeInvoiceID(ctx, invoiceData.ID)
	if err != nil {
		logger.Error("Failed to get invoice by Stripe ID", err, map[string]interface{}{
			"stripe_invoice_id": invoiceData.ID,
		})
		return fmt.Errorf("failed to get invoice by Stripe ID: %w", err)
	}

	if invoice == nil {
		// Create new invoice record
		invoice = &entities.Invoice{
			StripeInvoiceID:  &stripeInvoice.ID,
			Status:           stripeInvoice.Status,
			Amount:           stripeInvoice.Amount,
			Currency:         stripeInvoice.Currency,
			DueDate:          stripeInvoice.DueDate,
			PaidAt:           stripeInvoice.PaidAt,
			HostedInvoiceURL: stripeInvoice.HostedInvoiceURL,
			InvoicePDFURL:    stripeInvoice.InvoicePDFURL,
		}

		err = uc.invoiceRepo.Create(ctx, invoice)
		if err != nil {
			logger.Error("Failed to create invoice", err, map[string]interface{}{
				"stripe_invoice_id": invoiceData.ID,
			})
			return fmt.Errorf("failed to create invoice: %w", err)
		}
	} else {
		// Update existing invoice
		invoice.Status = stripeInvoice.Status
		invoice.Amount = stripeInvoice.Amount
		invoice.Currency = stripeInvoice.Currency
		invoice.DueDate = stripeInvoice.DueDate
		invoice.PaidAt = stripeInvoice.PaidAt
		invoice.HostedInvoiceURL = stripeInvoice.HostedInvoiceURL
		invoice.InvoicePDFURL = stripeInvoice.InvoicePDFURL

		err = uc.invoiceRepo.Update(ctx, invoice)
		if err != nil {
			logger.Error("Failed to update invoice", err, map[string]interface{}{
				"invoice_id":        invoice.ID,
				"stripe_invoice_id": invoiceData.ID,
			})
			return fmt.Errorf("failed to update invoice: %w", err)
		}
	}

	return nil
}

// handleInvoiceUpdated handles invoice.updated webhook event
func (uc *ProcessWebhookUseCase) handleInvoiceUpdated(ctx context.Context, rawData json.RawMessage) error {
	return uc.handleInvoiceCreated(ctx, rawData) // Same logic as created
}

// handlePaymentIntentSucceeded handles payment_intent.succeeded webhook event
func (uc *ProcessWebhookUseCase) handlePaymentIntentSucceeded(ctx context.Context, rawData json.RawMessage) error {
	var paymentIntentData struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal(rawData, &paymentIntentData)
	if err != nil {
		logger.Error("Failed to unmarshal payment intent data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment status
	payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentData.ID)
	if err != nil {
		logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
			"payment_intent_id": paymentIntentData.ID,
		})
		return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
	}

	if payment != nil {
		payment.SetSucceeded()
		err = uc.paymentRepo.Update(ctx, payment)
		if err != nil {
			logger.Error("Failed to update payment", err, map[string]interface{}{
				"payment_id": payment.ID,
			})
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return nil
}

// handlePaymentIntentFailed handles payment_intent.payment_failed webhook event
func (uc *ProcessWebhookUseCase) handlePaymentIntentFailed(ctx context.Context, rawData json.RawMessage) error {
	var paymentIntentData struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal(rawData, &paymentIntentData)
	if err != nil {
		logger.Error("Failed to unmarshal payment intent data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment status
	payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentData.ID)
	if err != nil {
		logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
			"payment_intent_id": paymentIntentData.ID,
		})
		return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
	}

	if payment != nil {
		payment.SetFailed("Payment failed")
		err = uc.paymentRepo.Update(ctx, payment)
		if err != nil {
			logger.Error("Failed to update payment", err, map[string]interface{}{
				"payment_id": payment.ID,
			})
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return nil
}

// handlePaymentIntentCanceled handles payment_intent.canceled webhook event
func (uc *ProcessWebhookUseCase) handlePaymentIntentCanceled(ctx context.Context, rawData json.RawMessage) error {
	var paymentIntentData struct {
		ID string `json:"id"`
	}

	err := json.Unmarshal(rawData, &paymentIntentData)
	if err != nil {
		logger.Error("Failed to unmarshal payment intent data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment status
	payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, paymentIntentData.ID)
	if err != nil {
		logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
			"payment_intent_id": paymentIntentData.ID,
		})
		return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
	}

	if payment != nil {
		payment.SetCanceled()
		err = uc.paymentRepo.Update(ctx, payment)
		if err != nil {
			logger.Error("Failed to update payment", err, map[string]interface{}{
				"payment_id": payment.ID,
			})
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return nil
}

// handleChargeSucceeded handles charge.succeeded webhook event
func (uc *ProcessWebhookUseCase) handleChargeSucceeded(ctx context.Context, rawData json.RawMessage) error {
	var chargeData struct {
		ID           string `json:"id"`
		Payment      string `json:"payment_intent"`
		Amount       int64  `json:"amount"`
		Currency     string `json:"currency"`
	}

	err := json.Unmarshal(rawData, &chargeData)
	if err != nil {
		logger.Error("Failed to unmarshal charge data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment with charge ID
	if chargeData.Payment != "" {
		payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, chargeData.Payment)
		if err != nil {
			logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
				"payment_intent_id": chargeData.Payment,
			})
			return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
		}

		if payment != nil {
			stripeChargeID := chargeData.ID
			payment.StripeChargeID = &stripeChargeID
			err = uc.paymentRepo.Update(ctx, payment)
			if err != nil {
				logger.Error("Failed to update payment", err, map[string]interface{}{
					"payment_id": payment.ID,
				})
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}
	}

	return nil
}

// handleChargeFailed handles charge.failed webhook event
func (uc *ProcessWebhookUseCase) handleChargeFailed(ctx context.Context, rawData json.RawMessage) error {
	var chargeData struct {
		ID      string `json:"id"`
		Payment string `json:"payment_intent"`
	}

	err := json.Unmarshal(rawData, &chargeData)
	if err != nil {
		logger.Error("Failed to unmarshal charge data", err, nil)
		return ErrWebhookProcessingFailed
	}

	// Update payment status
	if chargeData.Payment != "" {
		payment, err := uc.paymentRepo.GetByStripePaymentIntentID(ctx, chargeData.Payment)
		if err != nil {
			logger.Error("Failed to get payment by Stripe ID", err, map[string]interface{}{
				"payment_intent_id": chargeData.Payment,
			})
			return fmt.Errorf("failed to get payment by Stripe ID: %w", err)
		}

		if payment != nil {
			payment.SetFailed("Charge failed")
			err = uc.paymentRepo.Update(ctx, payment)
			if err != nil {
				logger.Error("Failed to update payment", err, map[string]interface{}{
					"payment_id": payment.ID,
				})
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}
	}

	return nil
}

// handleChargeDisputeCreated handles charge.dispute.created webhook event
func (uc *ProcessWebhookUseCase) handleChargeDisputeCreated(ctx context.Context, rawData json.RawMessage) error {
	var disputeData struct {
		ID      string `json:"id"`
		Charge  string `json:"charge"`
		Amount  int64  `json:"amount"`
		Reason  string `json:"reason"`
	}

	err := json.Unmarshal(rawData, &disputeData)
	if err != nil {
		logger.Error("Failed to unmarshal dispute data", err, nil)
		return ErrWebhookProcessingFailed
	}

	logger.Warn("Charge dispute created", map[string]interface{}{
		"dispute_id": disputeData.ID,
		"charge_id":  disputeData.Charge,
		"amount":     disputeData.Amount,
		"reason":     disputeData.Reason,
	})

	return nil
}

// mapStripePlanToInternal maps Stripe plan ID to internal plan type
func (uc *ProcessWebhookUseCase) mapStripePlanToInternal(stripePlanID string) string {
	// Map Stripe price IDs to internal plan types
	switch stripePlanID {
	case "price_premium_monthly":
		return "premium"
	case "price_platinum_monthly":
		return "platinum"
	default:
		return "basic" // Default to basic for unknown plans
	}
}