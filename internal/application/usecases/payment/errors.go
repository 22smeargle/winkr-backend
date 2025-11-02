package payment

import "errors"

// Payment use case errors
var (
	// General errors
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrPaymentMethodNotFound  = errors.New("payment method not found")
	ErrRefundNotFound       = errors.New("refund not found")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrWebhookEventNotFound  = errors.New("webhook event not found")
	
	// Plan errors
	ErrPlanNotFound         = errors.New("subscription plan not found")
	ErrInvalidPlanType      = errors.New("invalid subscription plan type")
	ErrPlanNotAvailable     = errors.New("subscription plan not available")
	
	// Payment errors
	ErrPaymentFailed        = errors.New("payment failed")
	ErrPaymentCanceled       = errors.New("payment canceled")
	ErrPaymentRefunded      = errors.New("payment already refunded")
	ErrPaymentCannotRefund  = errors.New("payment cannot be refunded")
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")
	ErrInvalidCurrency      = errors.New("invalid currency")
	
	// Payment method errors
	ErrInvalidPaymentMethodType = errors.New("invalid payment method type")
	ErrPaymentMethodExpired     = errors.New("payment method expired")
	ErrPaymentMethodNotVerified = errors.New("payment method not verified")
	ErrDefaultPaymentMethodRequired = errors.New("default payment method required")
	ErrPaymentMethodAlreadyDefault = errors.New("payment method already set as default")
	
	// Refund errors
	ErrRefundFailed        = errors.New("refund failed")
	ErrInvalidRefundAmount = errors.New("invalid refund amount")
	ErrRefundAmountExceedsPayment = errors.New("refund amount exceeds payment amount")
	ErrRefundAlreadyProcessed = errors.New("refund already processed")
	ErrInvalidRefundReason = errors.New("invalid refund reason")
	
	// Invoice errors
	ErrInvoiceAlreadyPaid   = errors.New("invoice already paid")
	ErrInvoiceNotPaid      = errors.New("invoice not paid")
	ErrInvoiceOverdue      = errors.New("invoice is overdue")
	ErrInvoiceVoid        = errors.New("invoice is void")
	ErrInvoiceUncollectible = errors.New("invoice is uncollectible")
	
	// Subscription errors
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrSubscriptionActive      = errors.New("subscription is already active")
	ErrSubscriptionInactive    = errors.New("subscription is not active")
	ErrSubscriptionCanceled    = errors.New("subscription is canceled")
	ErrSubscriptionExpired    = errors.New("subscription is expired")
	ErrSubscriptionPastDue    = errors.New("subscription is past due")
	ErrSubscriptionUnpaid     = errors.New("subscription is unpaid")
	ErrCannotCancelSubscription = errors.New("cannot cancel subscription")
	ErrCannotReactivateSubscription = errors.New("cannot reactivate subscription")
	ErrCannotUpgradeSubscription = errors.New("cannot upgrade subscription")
	ErrCannotDowngradeSubscription = errors.New("cannot downgrade subscription")
	
	// Webhook errors
	ErrWebhookSignatureInvalid = errors.New("webhook signature invalid")
	ErrWebhookEventNotSupported = errors.New("webhook event not supported")
	ErrWebhookProcessingFailed = errors.New("webhook processing failed")
	ErrWebhookAlreadyProcessed = errors.New("webhook event already processed")
	ErrWebhookMaxRetriesExceeded = errors.New("webhook max retries exceeded")
	
	// Stripe errors
	ErrStripeCustomerNotFound = errors.New("Stripe customer not found")
	ErrStripePaymentIntentNotFound = errors.New("Stripe payment intent not found")
	ErrStripePaymentMethodNotFound = errors.New("Stripe payment method not found")
	ErrStripeSubscriptionNotFound = errors.New("Stripe subscription not found")
	ErrStripeInvoiceNotFound = errors.New("Stripe invoice not found")
	ErrStripeRefundNotFound = errors.New("Stripe refund not found")
	ErrStripeAPIError = errors.New("Stripe API error")
	ErrStripeWebhookSecretInvalid = errors.New("Stripe webhook secret invalid")
	
	// Validation errors
	ErrInvalidUserID         = errors.New("invalid user ID")
	ErrInvalidPaymentID      = errors.New("invalid payment ID")
	ErrInvalidPaymentMethodID = errors.New("invalid payment method ID")
	ErrInvalidRefundID       = errors.New("invalid refund ID")
	ErrInvalidInvoiceID       = errors.New("invalid invoice ID")
	ErrInvalidSubscriptionID  = errors.New("invalid subscription ID")
	ErrInvalidWebhookEventID = errors.New("invalid webhook event ID")
	
	// Permission errors
	ErrPaymentAccessDenied   = errors.New("payment access denied")
	ErrRefundAccessDenied    = errors.New("refund access denied")
	ErrInvoiceAccessDenied   = errors.New("invoice access denied")
	ErrSubscriptionAccessDenied = errors.New("subscription access denied")
	
	// Business logic errors
	ErrUserHasActiveSubscription = errors.New("user already has an active subscription")
	ErrUserHasNoActiveSubscription = errors.New("user has no active subscription")
	ErrUserHasNoPaymentMethods = errors.New("user has no payment methods")
	ErrUserHasNoDefaultPaymentMethod = errors.New("user has no default payment method")
	ErrSubscriptionCannotBeDowngraded = errors.New("subscription cannot be downgraded")
	ErrSubscriptionCannotBeUpgraded = errors.New("subscription cannot be upgraded")
	ErrSubscriptionCannotBeCanceled = errors.New("subscription cannot be canceled")
	ErrSubscriptionCannotBeReactivated = errors.New("subscription cannot be reactivated")
	
	// Rate limiting errors
	ErrPaymentRateLimitExceeded = errors.New("payment rate limit exceeded")
	ErrRefundRateLimitExceeded = errors.New("refund rate limit exceeded")
	ErrWebhookRateLimitExceeded = errors.New("webhook rate limit exceeded")
	
	// System errors
	ErrPaymentServiceUnavailable = errors.New("payment service unavailable")
	ErrPaymentServiceTimeout = errors.New("payment service timeout")
	ErrPaymentServiceError = errors.New("payment service error")
	ErrDatabaseError = errors.New("database error")
	ErrCacheError = errors.New("cache error")
	ErrConfigurationError = errors.New("configuration error")
)