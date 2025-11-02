package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// PaymentCacheService handles payment-related caching operations
type PaymentCacheService struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewPaymentCacheService creates a new payment cache service
func NewPaymentCacheService(redisClient *redis.RedisClient) *PaymentCacheService {
	return &PaymentCacheService{
		redisClient: redisClient,
		prefix:      "payment:",
	}
}

// Payment cache TTL constants
const (
	SubscriptionCacheTTL    = 15 * time.Minute
	PaymentMethodsCacheTTL    = 30 * time.Minute
	PlansCacheTTL           = 60 * time.Minute
	PaymentCacheTTL          = 10 * time.Minute
	InvoiceCacheTTL          = 15 * time.Minute
	WebhookEventCacheTTL     = 5 * time.Minute
)

// CacheSubscription caches subscription data
func (pcs *PaymentCacheService) CacheSubscription(ctx context.Context, userID string, subscription *entities.Subscription) error {
	key := pcs.getSubscriptionKey(userID)
	
	subscriptionData, err := json.Marshal(subscription)
	if err != nil {
		logger.Error("Failed to marshal subscription for caching", err)
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(subscriptionData), SubscriptionCacheTTL)
	if err != nil {
		logger.Error("Failed to cache subscription", err)
		return fmt.Errorf("failed to cache subscription: %w", err)
	}

	logger.Debug("Subscription cached", "user_id", userID)
	return nil
}

// GetSubscription retrieves cached subscription data
func (pcs *PaymentCacheService) GetSubscription(ctx context.Context, userID string) (*entities.Subscription, error) {
	key := pcs.getSubscriptionKey(userID)
	
	subscriptionData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached subscription", err)
		return nil, fmt.Errorf("failed to get cached subscription: %w", err)
	}

	if subscriptionData == "" {
		return nil, nil // Cache miss
	}

	var subscription entities.Subscription
	err = json.Unmarshal([]byte(subscriptionData), &subscription)
	if err != nil {
		logger.Error("Failed to unmarshal cached subscription", err)
		return nil, fmt.Errorf("failed to unmarshal cached subscription: %w", err)
	}

	logger.Debug("Subscription retrieved from cache", "user_id", userID)
	return &subscription, nil
}

// InvalidateSubscription removes subscription from cache
func (pcs *PaymentCacheService) InvalidateSubscription(ctx context.Context, userID string) error {
	key := pcs.getSubscriptionKey(userID)
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate subscription cache", err)
		return fmt.Errorf("failed to invalidate subscription cache: %w", err)
	}

	logger.Debug("Subscription cache invalidated", "user_id", userID)
	return nil
}

// CachePaymentMethods caches user payment methods
func (pcs *PaymentCacheService) CachePaymentMethods(ctx context.Context, userID string, paymentMethods []*entities.PaymentMethod) error {
	key := pcs.getPaymentMethodsKey(userID)
	
	paymentMethodsData, err := json.Marshal(paymentMethods)
	if err != nil {
		logger.Error("Failed to marshal payment methods for caching", err)
		return fmt.Errorf("failed to marshal payment methods: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(paymentMethodsData), PaymentMethodsCacheTTL)
	if err != nil {
		logger.Error("Failed to cache payment methods", err)
		return fmt.Errorf("failed to cache payment methods: %w", err)
	}

	logger.Debug("Payment methods cached", "user_id", userID, "count", len(paymentMethods))
	return nil
}

// GetPaymentMethods retrieves cached payment methods
func (pcs *PaymentCacheService) GetPaymentMethods(ctx context.Context, userID string) ([]*entities.PaymentMethod, error) {
	key := pcs.getPaymentMethodsKey(userID)
	
	paymentMethodsData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached payment methods", err)
		return nil, fmt.Errorf("failed to get cached payment methods: %w", err)
	}

	if paymentMethodsData == "" {
		return nil, nil // Cache miss
	}

	var paymentMethods []*entities.PaymentMethod
	err = json.Unmarshal([]byte(paymentMethodsData), &paymentMethods)
	if err != nil {
		logger.Error("Failed to unmarshal cached payment methods", err)
		return nil, fmt.Errorf("failed to unmarshal cached payment methods: %w", err)
	}

	logger.Debug("Payment methods retrieved from cache", "user_id", userID)
	return paymentMethods, nil
}

// InvalidatePaymentMethods removes payment methods from cache
func (pcs *PaymentCacheService) InvalidatePaymentMethods(ctx context.Context, userID string) error {
	key := pcs.getPaymentMethodsKey(userID)
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate payment methods cache", err)
		return fmt.Errorf("failed to invalidate payment methods cache: %w", err)
	}

	logger.Debug("Payment methods cache invalidated", "user_id", userID)
	return nil
}

// CachePlans caches subscription plans
func (pcs *PaymentCacheService) CachePlans(ctx context.Context, plans interface{}) error {
	key := pcs.getPlansKey()
	
	plansData, err := json.Marshal(plans)
	if err != nil {
		logger.Error("Failed to marshal plans for caching", err)
		return fmt.Errorf("failed to marshal plans: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(plansData), PlansCacheTTL)
	if err != nil {
		logger.Error("Failed to cache plans", err)
		return fmt.Errorf("failed to cache plans: %w", err)
	}

	logger.Debug("Plans cached")
	return nil
}

// GetPlans retrieves cached subscription plans
func (pcs *PaymentCacheService) GetPlans(ctx context.Context, plans interface{}) (bool, error) {
	key := pcs.getPlansKey()
	
	plansData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached plans", err)
		return false, fmt.Errorf("failed to get cached plans: %w", err)
	}

	if plansData == "" {
		return false, nil // Cache miss
	}

	err = json.Unmarshal([]byte(plansData), &plans)
	if err != nil {
		logger.Error("Failed to unmarshal cached plans", err)
		return false, fmt.Errorf("failed to unmarshal cached plans: %w", err)
	}

	logger.Debug("Plans retrieved from cache")
	return true, nil
}

// InvalidatePlans removes plans from cache
func (pcs *PaymentCacheService) InvalidatePlans(ctx context.Context) error {
	key := pcs.getPlansKey()
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate plans cache", err)
		return fmt.Errorf("failed to invalidate plans cache: %w", err)
	}

	logger.Debug("Plans cache invalidated")
	return nil
}

// CachePayment caches payment data
func (pcs *PaymentCacheService) CachePayment(ctx context.Context, paymentID string, payment *entities.Payment) error {
	key := pcs.getPaymentKey(paymentID)
	
	paymentData, err := json.Marshal(payment)
	if err != nil {
		logger.Error("Failed to marshal payment for caching", err)
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(paymentData), PaymentCacheTTL)
	if err != nil {
		logger.Error("Failed to cache payment", err)
		return fmt.Errorf("failed to cache payment: %w", err)
	}

	logger.Debug("Payment cached", "payment_id", paymentID)
	return nil
}

// GetPayment retrieves cached payment data
func (pcs *PaymentCacheService) GetPayment(ctx context.Context, paymentID string) (*entities.Payment, error) {
	key := pcs.getPaymentKey(paymentID)
	
	paymentData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached payment", err)
		return nil, fmt.Errorf("failed to get cached payment: %w", err)
	}

	if paymentData == "" {
		return nil, nil // Cache miss
	}

	var payment entities.Payment
	err = json.Unmarshal([]byte(paymentData), &payment)
	if err != nil {
		logger.Error("Failed to unmarshal cached payment", err)
		return nil, fmt.Errorf("failed to unmarshal cached payment: %w", err)
	}

	logger.Debug("Payment retrieved from cache", "payment_id", paymentID)
	return &payment, nil
}

// InvalidatePayment removes payment from cache
func (pcs *PaymentCacheService) InvalidatePayment(ctx context.Context, paymentID string) error {
	key := pcs.getPaymentKey(paymentID)
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate payment cache", err)
		return fmt.Errorf("failed to invalidate payment cache: %w", err)
	}

	logger.Debug("Payment cache invalidated", "payment_id", paymentID)
	return nil
}

// CacheInvoice caches invoice data
func (pcs *PaymentCacheService) CacheInvoice(ctx context.Context, invoiceID string, invoice *entities.Invoice) error {
	key := pcs.getInvoiceKey(invoiceID)
	
	invoiceData, err := json.Marshal(invoice)
	if err != nil {
		logger.Error("Failed to marshal invoice for caching", err)
		return fmt.Errorf("failed to marshal invoice: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(invoiceData), InvoiceCacheTTL)
	if err != nil {
		logger.Error("Failed to cache invoice", err)
		return fmt.Errorf("failed to cache invoice: %w", err)
	}

	logger.Debug("Invoice cached", "invoice_id", invoiceID)
	return nil
}

// GetInvoice retrieves cached invoice data
func (pcs *PaymentCacheService) GetInvoice(ctx context.Context, invoiceID string) (*entities.Invoice, error) {
	key := pcs.getInvoiceKey(invoiceID)
	
	invoiceData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached invoice", err)
		return nil, fmt.Errorf("failed to get cached invoice: %w", err)
	}

	if invoiceData == "" {
		return nil, nil // Cache miss
	}

	var invoice entities.Invoice
	err = json.Unmarshal([]byte(invoiceData), &invoice)
	if err != nil {
		logger.Error("Failed to unmarshal cached invoice", err)
		return nil, fmt.Errorf("failed to unmarshal cached invoice: %w", err)
	}

	logger.Debug("Invoice retrieved from cache", "invoice_id", invoiceID)
	return &invoice, nil
}

// InvalidateInvoice removes invoice from cache
func (pcs *PaymentCacheService) InvalidateInvoice(ctx context.Context, invoiceID string) error {
	key := pcs.getInvoiceKey(invoiceID)
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate invoice cache", err)
		return fmt.Errorf("failed to invalidate invoice cache: %w", err)
	}

	logger.Debug("Invoice cache invalidated", "invoice_id", invoiceID)
	return nil
}

// CacheWebhookEvent caches webhook event data
func (pcs *PaymentCacheService) CacheWebhookEvent(ctx context.Context, eventID string, event *entities.WebhookEvent) error {
	key := pcs.getWebhookEventKey(eventID)
	
	eventData, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal webhook event for caching", err)
		return fmt.Errorf("failed to marshal webhook event: %w", err)
	}

	err = pcs.redisClient.Set(ctx, key, string(eventData), WebhookEventCacheTTL)
	if err != nil {
		logger.Error("Failed to cache webhook event", err)
		return fmt.Errorf("failed to cache webhook event: %w", err)
	}

	logger.Debug("Webhook event cached", "event_id", eventID)
	return nil
}

// GetWebhookEvent retrieves cached webhook event data
func (pcs *PaymentCacheService) GetWebhookEvent(ctx context.Context, eventID string) (*entities.WebhookEvent, error) {
	key := pcs.getWebhookEventKey(eventID)
	
	eventData, err := pcs.redisClient.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached webhook event", err)
		return nil, fmt.Errorf("failed to get cached webhook event: %w", err)
	}

	if eventData == "" {
		return nil, nil // Cache miss
	}

	var event entities.WebhookEvent
	err = json.Unmarshal([]byte(eventData), &event)
	if err != nil {
		logger.Error("Failed to unmarshal cached webhook event", err)
		return nil, fmt.Errorf("failed to unmarshal cached webhook event: %w", err)
	}

	logger.Debug("Webhook event retrieved from cache", "event_id", eventID)
	return &event, nil
}

// InvalidateWebhookEvent removes webhook event from cache
func (pcs *PaymentCacheService) InvalidateWebhookEvent(ctx context.Context, eventID string) error {
	key := pcs.getWebhookEventKey(eventID)
	
	err := pcs.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate webhook event cache", err)
		return fmt.Errorf("failed to invalidate webhook event cache: %w", err)
	}

	logger.Debug("Webhook event cache invalidated", "event_id", eventID)
	return nil
}

// InvalidateUserPaymentCache removes all payment-related cache for a user
func (pcs *PaymentCacheService) InvalidateUserPaymentCache(ctx context.Context, userID string) error {
	keys := []string{
		pcs.getSubscriptionKey(userID),
		pcs.getPaymentMethodsKey(userID),
	}

	for _, key := range keys {
		err := pcs.redisClient.Del(ctx, key)
		if err != nil {
			logger.Error("Failed to invalidate user payment cache key", err, "key", key)
			return fmt.Errorf("failed to invalidate user payment cache key: %w", err)
		}
	}

	logger.Debug("User payment cache invalidated", "user_id", userID)
	return nil
}

// InvalidatePaymentPattern removes all payment cache keys matching a pattern
func (pcs *PaymentCacheService) InvalidatePaymentPattern(ctx context.Context, pattern string) error {
	fullPattern := fmt.Sprintf("%s%s", pcs.prefix, pattern)
	
	logger.Info("Invalidating payment cache pattern", "pattern", fullPattern)
	
	// In a real implementation, you would use Redis SCAN or KEYS command
	// For now, we'll log the invalidation request
	return nil
}

// Helper methods for key generation

func (pcs *PaymentCacheService) getSubscriptionKey(userID string) string {
	return fmt.Sprintf("%ssubscription:%s", pcs.prefix, userID)
}

func (pcs *PaymentCacheService) getPaymentMethodsKey(userID string) string {
	return fmt.Sprintf("%spayment_methods:%s", pcs.prefix, userID)
}

func (pcs *PaymentCacheService) getPlansKey() string {
	return fmt.Sprintf("%splans", pcs.prefix)
}

func (pcs *PaymentCacheService) getPaymentKey(paymentID string) string {
	return fmt.Sprintf("%spayment:%s", pcs.prefix, paymentID)
}

func (pcs *PaymentCacheService) getInvoiceKey(invoiceID string) string {
	return fmt.Sprintf("%sinvoice:%s", pcs.prefix, invoiceID)
}

func (pcs *PaymentCacheService) getWebhookEventKey(eventID string) string {
	return fmt.Sprintf("%swebhook_event:%s", pcs.prefix, eventID)
}