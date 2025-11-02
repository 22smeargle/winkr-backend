package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payment represents a payment transaction entity in database
type Payment struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	SubscriptionID   *uuid.UUID `gorm:"type:uuid;index" json:"subscription_id"`
	StripePaymentIntentID *string `gorm:"uniqueIndex" json:"stripe_payment_intent_id"`
	Amount           int64      `gorm:"not null" json:"amount"`
	Currency         string     `gorm:"not null;default:'USD'" json:"currency"`
	Status           string     `gorm:"not null;check:status IN ('pending', 'processing', 'succeeded', 'failed', 'canceled', 'refunded')" json:"status"`
	PaymentMethodID  *string    `json:"payment_method_id"`
	StripeChargeID   *string    `json:"stripe_charge_id"`
	RefundID         *string    `json:"refund_id"`
	FailureReason    *string    `json:"failure_reason"`
	Description      *string    `json:"description"`
	Metadata         string     `gorm:"type:json" json:"metadata"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User         *User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:SET NULL" json:"subscription,omitempty"`
}

// TableName returns the table name for Payment model
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate GORM hook
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsPending returns true if the payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == "pending"
}

// IsProcessing returns true if the payment is processing
func (p *Payment) IsProcessing() bool {
	return p.Status == "processing"
}

// IsSucceeded returns true if the payment succeeded
func (p *Payment) IsSucceeded() bool {
	return p.Status == "succeeded"
}

// IsFailed returns true if the payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == "failed"
}

// IsCanceled returns true if the payment was canceled
func (p *Payment) IsCanceled() bool {
	return p.Status == "canceled"
}

// IsRefunded returns true if the payment was refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == "refunded"
}

// CanBeRefunded returns true if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.IsSucceeded() && p.RefundID == nil
}

// SetPending sets the payment status to pending
func (p *Payment) SetPending() {
	p.Status = "pending"
}

// SetProcessing sets the payment status to processing
func (p *Payment) SetProcessing() {
	p.Status = "processing"
}

// SetSucceeded sets the payment status to succeeded
func (p *Payment) SetSucceeded() {
	p.Status = "succeeded"
}

// SetFailed sets the payment status to failed
func (p *Payment) SetFailed(reason string) {
	p.Status = "failed"
	p.FailureReason = &reason
}

// SetCanceled sets the payment status to canceled
func (p *Payment) SetCanceled() {
	p.Status = "canceled"
}

// SetRefunded sets the payment status to refunded
func (p *Payment) SetRefunded(refundID string) {
	p.Status = "refunded"
	p.RefundID = &refundID
}

// PaymentMethod represents a payment method entity in database
type PaymentMethod struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StripePaymentMethodID *string `gorm:"uniqueIndex" json:"stripe_payment_method_id"`
	Type            string     `gorm:"not null;check:type IN ('card', 'bank_account', 'sepa_debit')" json:"type"`
	IsDefault       bool       `gorm:"default:false" json:"is_default"`
	CardBrand       *string    `json:"card_brand"`
	CardLast4       *string    `json:"card_last4"`
	CardExpiryMonth *int64     `json:"card_expiry_month"`
	CardExpiryYear  *int64     `json:"card_expiry_year"`
	CardFingerprint *string    `json:"card_fingerprint"`
	BankName        *string    `json:"bank_name"`
	BankLast4       *string    `json:"bank_last4"`
	IsVerified      bool       `gorm:"default:false" json:"is_verified"`
	Metadata        string     `gorm:"type:json" json:"metadata"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for PaymentMethod model
func (PaymentMethod) TableName() string {
	return "payment_methods"
}

// BeforeCreate GORM hook
func (pm *PaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	return nil
}

// IsCard returns true if the payment method is a card
func (pm *PaymentMethod) IsCard() bool {
	return pm.Type == "card"
}

// IsBankAccount returns true if the payment method is a bank account
func (pm *PaymentMethod) IsBankAccount() bool {
	return pm.Type == "bank_account"
}

// IsSepaDebit returns true if the payment method is a SEPA debit
func (pm *PaymentMethod) IsSepaDebit() bool {
	return pm.Type == "sepa_debit"
}

// SetDefault sets the payment method as default
func (pm *PaymentMethod) SetDefault() {
	pm.IsDefault = true
}

// UnsetDefault unsets the payment method as default
func (pm *PaymentMethod) UnsetDefault() {
	pm.IsDefault = false
}

// Verify marks the payment method as verified
func (pm *PaymentMethod) Verify() {
	pm.IsVerified = true
}

// Unverify marks the payment method as unverified
func (pm *PaymentMethod) Unverify() {
	pm.IsVerified = false
}

// Refund represents a refund entity in database
type Refund struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"payment_id"`
	StripeRefundID   *string    `gorm:"uniqueIndex" json:"stripe_refund_id"`
	Amount           int64      `gorm:"not null" json:"amount"`
	Currency         string     `gorm:"not null;default:'USD'" json:"currency"`
	Status           string     `gorm:"not null;check:status IN ('pending', 'succeeded', 'failed', 'canceled')" json:"status"`
	Reason           *string    `gorm:"check:reason IN ('duplicate', 'fraudulent', 'requested_by_customer')" json:"reason"`
	ReceiptNumber    *string    `json:"receipt_number"`
	Description      *string    `json:"description"`
	Metadata         string     `gorm:"type:json" json:"metadata"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Payment *Payment `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
}

// TableName returns the table name for Refund model
func (Refund) TableName() string {
	return "refunds"
}

// BeforeCreate GORM hook
func (r *Refund) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// IsPending returns true if the refund is pending
func (r *Refund) IsPending() bool {
	return r.Status == "pending"
}

// IsSucceeded returns true if the refund succeeded
func (r *Refund) IsSucceeded() bool {
	return r.Status == "succeeded"
}

// IsFailed returns true if the refund failed
func (r *Refund) IsFailed() bool {
	return r.Status == "failed"
}

// IsCanceled returns true if the refund was canceled
func (r *Refund) IsCanceled() bool {
	return r.Status == "canceled"
}

// SetPending sets the refund status to pending
func (r *Refund) SetPending() {
	r.Status = "pending"
}

// SetSucceeded sets the refund status to succeeded
func (r *Refund) SetSucceeded() {
	r.Status = "succeeded"
}

// SetFailed sets the refund status to failed
func (r *Refund) SetFailed() {
	r.Status = "failed"
}

// SetCanceled sets the refund status to canceled
func (r *Refund) SetCanceled() {
	r.Status = "canceled"
}

// Invoice represents an invoice entity in database
type Invoice struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	SubscriptionID   *uuid.UUID `gorm:"type:uuid;index" json:"subscription_id"`
	StripeInvoiceID  *string    `gorm:"uniqueIndex" json:"stripe_invoice_id"`
	Number           *string    `json:"number"`
	Status           string     `gorm:"not null;check:status IN ('draft', 'open', 'paid', 'void', 'uncollectible')" json:"status"`
	Amount           int64      `gorm:"not null" json:"amount"`
	Currency         string     `gorm:"not null;default:'USD'" json:"currency"`
	DueDate          *time.Time `json:"due_date"`
	PaidAt           *time.Time `json:"paid_at"`
	HostedInvoiceURL *string    `json:"hosted_invoice_url"`
	InvoicePDFURL    *string    `json:"invoice_pdf_url"`
	Description      *string    `json:"description"`
	Metadata         string     `gorm:"type:json" json:"metadata"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User         *User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:SET NULL" json:"subscription,omitempty"`
}

// TableName returns the table name for Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// BeforeCreate GORM hook
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// IsDraft returns true if the invoice is a draft
func (i *Invoice) IsDraft() bool {
	return i.Status == "draft"
}

// IsOpen returns true if the invoice is open
func (i *Invoice) IsOpen() bool {
	return i.Status == "open"
}

// IsPaid returns true if the invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == "paid"
}

// IsVoid returns true if the invoice is void
func (i *Invoice) IsVoid() bool {
	return i.Status == "void"
}

// IsUncollectible returns true if the invoice is uncollectible
func (i *Invoice) IsUncollectible() bool {
	return i.Status == "uncollectible"
}

// IsOverdue returns true if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.DueDate == nil || i.IsPaid() || i.IsVoid() || i.IsUncollectible() {
		return false
	}
	return time.Now().After(*i.DueDate)
}

// SetDraft sets the invoice status to draft
func (i *Invoice) SetDraft() {
	i.Status = "draft"
}

// SetOpen sets the invoice status to open
func (i *Invoice) SetOpen() {
	i.Status = "open"
}

// SetPaid sets the invoice status to paid
func (i *Invoice) SetPaid() {
	i.Status = "paid"
	now := time.Now()
	i.PaidAt = &now
}

// SetVoid sets the invoice status to void
func (i *Invoice) SetVoid() {
	i.Status = "void"
}

// SetUncollectible sets the invoice status to uncollectible
func (i *Invoice) SetUncollectible() {
	i.Status = "uncollectible"
}

// WebhookEvent represents a webhook event entity in database
type WebhookEvent struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StripeEventID   string     `gorm:"uniqueIndex;not null" json:"stripe_event_id"`
	Type            string     `gorm:"not null;index" json:"type"`
	Processed       bool       `gorm:"default:false" json:"processed"`
	ProcessingError *string    `json:"processing_error"`
	RawData         string     `gorm:"type:text" json:"raw_data"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName returns the table name for WebhookEvent model
func (WebhookEvent) TableName() string {
	return "webhook_events"
}

// BeforeCreate GORM hook
func (we *WebhookEvent) BeforeCreate(tx *gorm.DB) error {
	if we.ID == uuid.Nil {
		we.ID = uuid.New()
	}
	return nil
}

// IsProcessed returns true if the webhook event has been processed
func (we *WebhookEvent) IsProcessed() bool {
	return we.Processed
}

// SetProcessed marks the webhook event as processed
func (we *WebhookEvent) SetProcessed() {
	we.Processed = true
	we.ProcessingError = nil
}

// SetProcessingError sets the processing error
func (we *WebhookEvent) SetProcessingError(error string) {
	we.Processed = false
	we.ProcessingError = &error
}

// IsValidPaymentStatus checks if the payment status is valid
func IsValidPaymentStatus(status string) bool {
	validStatuses := []string{"pending", "processing", "succeeded", "failed", "canceled", "refunded"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPaymentMethodType checks if the payment method type is valid
func IsValidPaymentMethodType(paymentMethodType string) bool {
	validTypes := []string{"card", "bank_account", "sepa_debit"}
	for _, validType := range validTypes {
		if paymentMethodType == validType {
			return true
		}
	}
	return false
}

// IsValidRefundStatus checks if the refund status is valid
func IsValidRefundStatus(status string) bool {
	validStatuses := []string{"pending", "succeeded", "failed", "canceled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidInvoiceStatus checks if the invoice status is valid
func IsValidInvoiceStatus(status string) bool {
	validStatuses := []string{"draft", "open", "paid", "void", "uncollectible"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}