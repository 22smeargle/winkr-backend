package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription represents a user's subscription entity in database
type Subscription struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID               uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StripeSubscriptionID  *string    `gorm:"uniqueIndex" json:"stripe_subscription_id"`
	PlanType             string     `gorm:"not null;check:plan_type IN ('basic', 'premium', 'platinum');index" json:"plan_type"`
	Status               string     `gorm:"not null;check:status IN ('active', 'canceled', 'past_due', 'unpaid');index" json:"status"`
	CurrentPeriodStart    *time.Time `json:"current_period_start"`
	CurrentPeriodEnd      *time.Time `json:"current_period_end"`
	CancelAtPeriodEnd    bool       `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt            time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}

// BeforeCreate GORM hook
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsActive returns true if the subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == "active"
}

// IsCanceled returns true if the subscription is canceled
func (s *Subscription) IsCanceled() bool {
	return s.Status == "canceled"
}

// IsPastDue returns true if the subscription is past due
func (s *Subscription) IsPastDue() bool {
	return s.Status == "past_due"
}

// IsUnpaid returns true if the subscription is unpaid
func (s *Subscription) IsUnpaid() bool {
	return s.Status == "unpaid"
}

// IsBasic returns true if the subscription is basic plan
func (s *Subscription) IsBasic() bool {
	return s.PlanType == "basic"
}

// IsPremium returns true if the subscription is premium plan
func (s *Subscription) IsPremium() bool {
	return s.PlanType == "premium"
}

// IsPlatinum returns true if the subscription is platinum plan
func (s *Subscription) IsPlatinum() bool {
	return s.PlanType == "platinum"
}

// IsPaidPlan returns true if the subscription is a paid plan
func (s *Subscription) IsPaidPlan() bool {
	return s.IsPremium() || s.IsPlatinum()
}

// WillCancelAtPeriodEnd returns true if the subscription will cancel at period end
func (s *Subscription) WillCancelAtPeriodEnd() bool {
	return s.CancelAtPeriodEnd
}

// IsExpired returns true if the subscription period has ended
func (s *Subscription) IsExpired() bool {
	if s.CurrentPeriodEnd == nil {
		return false
	}
	return time.Now().After(*s.CurrentPeriodEnd)
}

// GetRemainingDays returns the number of remaining days in current period
func (s *Subscription) GetRemainingDays() int {
	if s.CurrentPeriodEnd == nil {
		return 0
	}
	remaining := time.Until(*s.CurrentPeriodEnd)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}

// Activate activates the subscription
func (s *Subscription) Activate() {
	s.Status = "active"
}

// Cancel cancels the subscription
func (s *Subscription) Cancel() {
	s.Status = "canceled"
}

// SetPastDue sets the subscription status to past due
func (s *Subscription) SetPastDue() {
	s.Status = "past_due"
}

// SetUnpaid sets the subscription status to unpaid
func (s *Subscription) SetUnpaid() {
	s.Status = "unpaid"
}

// UpdatePeriod updates the current period
func (s *Subscription) UpdatePeriod(start, end time.Time) {
	s.CurrentPeriodStart = &start
	s.CurrentPeriodEnd = &end
}

// SetCancelAtPeriodEnd sets the subscription to cancel at period end
func (s *Subscription) SetCancelAtPeriodEnd(cancel bool) {
	s.CancelAtPeriodEnd = cancel
}

// UpdateFromStripe updates subscription from Stripe data
func (s *Subscription) UpdateFromStripe(stripeSubID, planType, status string, currentPeriodStart, currentPeriodEnd time.Time, cancelAtPeriodEnd bool) {
	s.StripeSubscriptionID = &stripeSubID
	s.PlanType = planType
	s.Status = status
	s.UpdatePeriod(currentPeriodStart, currentPeriodEnd)
	s.SetCancelAtPeriodEnd(cancelAtPeriodEnd)
}

// SubscriptionPlan represents a subscription plan configuration
type SubscriptionPlan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Price        float64  `json:"price"`
	Currency     string   `json:"currency"`
	Interval     string   `json:"interval"`
	Features     []string `json:"features"`
	StripePriceID string   `json:"stripe_price_id"`
}

// GetAvailablePlans returns available subscription plans
func GetAvailablePlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID:           "premium",
			Name:         "Premium",
			Price:        9.99,
			Currency:      "USD",
			Interval:      "month",
			Features: []string{
				"Unlimited swipes",
				"See who likes you",
				"5 super likes per day",
				"1 boost per month",
			},
			StripePriceID: "price_premium_monthly",
		},
		{
			ID:           "platinum",
			Name:         "Platinum",
			Price:        19.99,
			Currency:      "USD",
			Interval:      "month",
			Features: []string{
				"All Premium features",
				"Unlimited super likes",
				"Weekly boosts",
				"Priority support",
			},
			StripePriceID: "price_platinum_monthly",
		},
	}
}

// GetPlanByID returns a subscription plan by ID
func GetPlanByID(planID string) (*SubscriptionPlan, bool) {
	plans := GetAvailablePlans()
	for _, plan := range plans {
		if plan.ID == planID {
			return &plan, true
		}
	}
	return nil, false
}

// IsValidPlanType checks if the plan type is valid
func IsValidPlanType(planType string) bool {
	validPlans := []string{"basic", "premium", "platinum"}
	for _, validPlan := range validPlans {
		if planType == validPlan {
			return true
		}
	}
	return false
}

// IsValidStatus checks if the subscription status is valid
func IsValidStatus(status string) bool {
	validStatuses := []string{"active", "canceled", "past_due", "unpaid"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}