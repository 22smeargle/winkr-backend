package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// SubscriptionRepository defines interface for subscription data operations
type SubscriptionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, subscription *entities.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error)
	Update(ctx context.Context, subscription *entities.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User subscription operations
	GetUserSubscription(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error)
	GetActiveUserSubscription(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error)
	GetUserSubscriptionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Subscription, error)

	// Subscription status operations
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Subscription, error)
	GetActiveSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)
	GetCanceledSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)
	GetPastDueSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)
	GetUnpaidSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)

	// Plan type operations
	GetByPlanType(ctx context.Context, planType string, limit, offset int) ([]*entities.Subscription, error)
	GetPremiumSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)
	GetPlatinumSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)

	// Stripe integration
	GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entities.Subscription, error)
	UpdateFromStripe(ctx context.Context, stripeSubscriptionID, planType, status string, currentPeriodStart, currentPeriodEnd interface{}, cancelAtPeriodEnd bool) error
	MarkForCancellation(ctx context.Context, stripeSubscriptionID string) error
	RemoveCancellation(ctx context.Context, stripeSubscriptionID string) error

	// Subscription lifecycle
	CancelSubscription(ctx context.Context, userID uuid.UUID, cancelAtPeriodEnd bool) error
	ReactivateSubscription(ctx context.Context, userID uuid.UUID) error
	UpgradeSubscription(ctx context.Context, userID uuid.UUID, newPlanType string) error
	DowngradeSubscription(ctx context.Context, userID uuid.UUID, newPlanType string) error

	// Period management
	UpdateSubscriptionPeriod(ctx context.Context, subscriptionID uuid.UUID, currentPeriodStart, currentPeriodEnd interface{}) error
	ExtendSubscriptionPeriod(ctx context.Context, subscriptionID uuid.UUID, days int) error
	SetCancelAtPeriodEnd(ctx context.Context, subscriptionID uuid.UUID, cancel bool) error

	// Batch operations
	BatchCreate(ctx context.Context, subscriptions []*entities.Subscription) error
	BatchUpdate(ctx context.Context, subscriptions []*entities.Subscription) error
	BatchCancel(ctx context.Context, subscriptionIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripeSubscriptionID string) (bool, error)
	UserHasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error)
	UserHasSubscription(ctx context.Context, userID uuid.UUID) (bool, error)

	// Analytics and statistics
	GetSubscriptionStats(ctx context.Context) (*SubscriptionStats, error)
	GetUserSubscriptionStats(ctx context.Context, userID uuid.UUID) (*UserSubscriptionStats, error)
	GetSubscriptionsCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetSubscriptionsByPlanStats(ctx context.Context, startDate, endDate interface{}) ([]*SubscriptionPlanStats, error)

	// Admin operations
	GetAllSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)
	GetSubscriptionsWithDetails(ctx context.Context, limit, offset int) ([]*SubscriptionWithDetails, error)
	GetExpiringSubscriptions(ctx context.Context, days int, limit int) ([]*entities.Subscription, error)
	GetExpiredSubscriptions(ctx context.Context, limit, offset int) ([]*entities.Subscription, error)

	// Advanced queries
	GetSubscriptionsExpiringSoon(ctx context.Context, days int) ([]*entities.Subscription, error)
	GetRecentlyCanceledSubscriptions(ctx context.Context, days int, limit int) ([]*entities.Subscription, error)
	GetSubscriptionRevenue(ctx context.Context, startDate, endDate interface{}) (*SubscriptionRevenue, error)
	GetChurnRate(ctx context.Context, startDate, endDate interface{}) (float64, error)
}

// SubscriptionStats represents overall subscription statistics
type SubscriptionStats struct {
	TotalSubscriptions      int64   `json:"total_subscriptions"`
	ActiveSubscriptions     int64   `json:"active_subscriptions"`
	CanceledSubscriptions   int64   `json:"canceled_subscriptions"`
	PastDueSubscriptions    int64   `json:"past_due_subscriptions"`
	UnpaidSubscriptions    int64   `json:"unpaid_subscriptions"`
	PremiumSubscriptions    int64   `json:"premium_subscriptions"`
	PlatinumSubscriptions  int64   `json:"platinum_subscriptions"`
	NewSubscriptionsToday   int64   `json:"new_subscriptions_today"`
	NewSubscriptionsThisWeek  int64   `json:"new_subscriptions_this_week"`
	NewSubscriptionsThisMonth int64   `json:"new_subscriptions_this_month"`
	MonthlyRecurringRevenue float64 `json:"monthly_recurring_revenue"`
	AverageSubscriptionValue float64 `json:"average_subscription_value"`
}

// UserSubscriptionStats represents subscription statistics for a user
type UserSubscriptionStats struct {
	TotalSubscriptions     int64   `json:"total_subscriptions"`
	CurrentSubscription    *entities.Subscription `json:"current_subscription,omitempty"`
	HasActiveSubscription bool     `json:"has_active_subscription"`
	TotalSpent           float64  `json:"total_spent"`
	SubscriptionDuration  int      `json:"subscription_duration_days"`
	LastSubscriptionDate  interface{} `json:"last_subscription_date"`
}

// SubscriptionPlanStats represents subscription statistics by plan
type SubscriptionPlanStats struct {
	PlanType    string  `json:"plan_type"`
	Count        int64   `json:"count"`
	Revenue      float64 `json:"revenue"`
	Percentage   float64 `json:"percentage"`
	ChurnRate    float64 `json:"churn_rate"`
}

// SubscriptionWithDetails represents a subscription with additional details
type SubscriptionWithDetails struct {
	*entities.Subscription
	User           *entities.User `json:"user"`
	DaysRemaining   int            `json:"days_remaining"`
	WillRenew      bool           `json:"will_renew"`
	IsExpiringSoon bool           `json:"is_expiring_soon"`
}

// SubscriptionRevenue represents subscription revenue data
type SubscriptionRevenue struct {
	TotalRevenue      float64 `json:"total_revenue"`
	PremiumRevenue    float64 `json:"premium_revenue"`
	PlatinumRevenue   float64 `json:"platinum_revenue"`
	NewRevenue        float64 `json:"new_revenue"`
	RecurringRevenue  float64 `json:"recurring_revenue"`
	RevenueGrowth     float64 `json:"revenue_growth_percentage"`
}