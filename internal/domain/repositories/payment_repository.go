package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// PaymentRepository defines interface for payment data operations
type PaymentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, payment *entities.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User payment operations
	GetUserPayments(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payment, error)
	GetUserPaymentByStripeID(ctx context.Context, userID uuid.UUID, stripePaymentIntentID string) (*entities.Payment, error)
	GetUserPaymentsByStatus(ctx context.Context, userID uuid.UUID, status string, limit, offset int) ([]*entities.Payment, error)

	// Payment status operations
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Payment, error)
	GetPendingPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
	GetProcessingPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
	GetSucceededPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
	GetFailedPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
	GetRefundedPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)

	// Stripe integration
	GetByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entities.Payment, error)
	UpdateFromStripe(ctx context.Context, stripePaymentIntentID, status string, stripeChargeID *string) error
	MarkAsRefunded(ctx context.Context, paymentID uuid.UUID, refundID string) error

	// Subscription payment operations
	GetSubscriptionPayments(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*entities.Payment, error)
	GetLatestSubscriptionPayment(ctx context.Context, subscriptionID uuid.UUID) (*entities.Payment, error)

	// Payment analytics
	GetPaymentStats(ctx context.Context, startDate, endDate interface{}) (*PaymentStats, error)
	GetUserPaymentStats(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (*UserPaymentStats, error)
	GetPaymentsCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetRevenueByPeriod(ctx context.Context, startDate, endDate interface{}) (*RevenueStats, error)

	// Batch operations
	BatchCreate(ctx context.Context, payments []*entities.Payment) error
	BatchUpdate(ctx context.Context, payments []*entities.Payment) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripePaymentIntentID string) (bool, error)
	UserHasPayments(ctx context.Context, userID uuid.UUID) (bool, error)

	// Admin operations
	GetAllPayments(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
	GetPaymentsWithDetails(ctx context.Context, limit, offset int) ([]*PaymentWithDetails, error)
	GetFailedPaymentsByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.Payment, error)
}

// PaymentMethodRepository defines interface for payment method data operations
type PaymentMethodRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, paymentMethod *entities.PaymentMethod) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.PaymentMethod, error)
	Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User payment method operations
	GetUserPaymentMethods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PaymentMethod, error)
	GetUserPaymentMethodsByType(ctx context.Context, userID uuid.UUID, paymentMethodType string, limit, offset int) ([]*entities.PaymentMethod, error)
	GetUserDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*entities.PaymentMethod, error)

	// Stripe integration
	GetByStripePaymentMethodID(ctx context.Context, stripePaymentMethodID string) (*entities.PaymentMethod, error)
	UpdateFromStripe(ctx context.Context, stripePaymentMethodID string, cardBrand, cardLast4 *string, cardExpiryMonth, cardExpiryYear *int64) error

	// Payment method management
	SetAsDefault(ctx context.Context, userID uuid.UUID, paymentMethodID uuid.UUID) error
	UnsetDefault(ctx context.Context, userID uuid.UUID) error
	VerifyPaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error
	UnverifyPaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error

	// Payment method status operations
	GetVerifiedPaymentMethods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PaymentMethod, error)
	GetUnverifiedPaymentMethods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.PaymentMethod, error)
	GetDefaultPaymentMethods(ctx context.Context, limit, offset int) ([]*entities.PaymentMethod, error)

	// Batch operations
	BatchCreate(ctx context.Context, paymentMethods []*entities.PaymentMethod) error
	BatchUpdate(ctx context.Context, paymentMethods []*entities.PaymentMethod) error
	BatchDelete(ctx context.Context, paymentMethodIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripePaymentMethodID string) (bool, error)
	UserHasPaymentMethods(ctx context.Context, userID uuid.UUID) (bool, error)
	UserHasDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (bool, error)

	// Admin operations
	GetAllPaymentMethods(ctx context.Context, limit, offset int) ([]*entities.PaymentMethod, error)
	GetPaymentMethodsWithDetails(ctx context.Context, limit, offset int) ([]*PaymentMethodWithDetails, error)
	GetPaymentMethodsByType(ctx context.Context, paymentMethodType string, limit, offset int) ([]*entities.PaymentMethod, error)
}

// RefundRepository defines interface for refund data operations
type RefundRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, refund *entities.Refund) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Refund, error)
	Update(ctx context.Context, refund *entities.Refund) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Payment refund operations
	GetPaymentRefunds(ctx context.Context, paymentID uuid.UUID, limit, offset int) ([]*entities.Refund, error)
	GetLatestPaymentRefund(ctx context.Context, paymentID uuid.UUID) (*entities.Refund, error)

	// Refund status operations
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Refund, error)
	GetPendingRefunds(ctx context.Context, limit, offset int) ([]*entities.Refund, error)
	GetSucceededRefunds(ctx context.Context, limit, offset int) ([]*entities.Refund, error)
	GetFailedRefunds(ctx context.Context, limit, offset int) ([]*entities.Refund, error)

	// Stripe integration
	GetByStripeRefundID(ctx context.Context, stripeRefundID string) (*entities.Refund, error)
	UpdateFromStripe(ctx context.Context, stripeRefundID, status string, receiptNumber *string) error

	// Refund analytics
	GetRefundStats(ctx context.Context, startDate, endDate interface{}) (*RefundStats, error)
	GetUserRefundStats(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (*UserRefundStats, error)
	GetRefundsCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetRefundAmountByPeriod(ctx context.Context, startDate, endDate interface{}) (int64, error)

	// Batch operations
	BatchCreate(ctx context.Context, refunds []*entities.Refund) error
	BatchUpdate(ctx context.Context, refunds []*entities.Refund) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripeRefundID string) (bool, error)
	PaymentHasRefunds(ctx context.Context, paymentID uuid.UUID) (bool, error)

	// Admin operations
	GetAllRefunds(ctx context.Context, limit, offset int) ([]*entities.Refund, error)
	GetRefundsWithDetails(ctx context.Context, limit, offset int) ([]*RefundWithDetails, error)
	GetRefundsByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.Refund, error)
}

// InvoiceRepository defines interface for invoice data operations
type InvoiceRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, invoice *entities.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error)
	Update(ctx context.Context, invoice *entities.Invoice) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User invoice operations
	GetUserInvoices(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Invoice, error)
	GetUserInvoiceByStripeID(ctx context.Context, userID uuid.UUID, stripeInvoiceID string) (*entities.Invoice, error)

	// Subscription invoice operations
	GetSubscriptionInvoices(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*entities.Invoice, error)
	GetLatestSubscriptionInvoice(ctx context.Context, subscriptionID uuid.UUID) (*entities.Invoice, error)

	// Invoice status operations
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Invoice, error)
	GetDraftInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetOpenInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetPaidInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetVoidInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetUncollectibleInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetOverdueInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)

	// Stripe integration
	GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*entities.Invoice, error)
	UpdateFromStripe(ctx context.Context, stripeInvoiceID, status string, hostedInvoiceURL, invoicePDFURL *string) error
	MarkAsPaid(ctx context.Context, invoiceID uuid.UUID) error

	// Invoice analytics
	GetInvoiceStats(ctx context.Context, startDate, endDate interface{}) (*InvoiceStats, error)
	GetUserInvoiceStats(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (*UserInvoiceStats, error)
	GetInvoicesCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetInvoiceAmountByPeriod(ctx context.Context, startDate, endDate interface{}) (int64, error)

	// Batch operations
	BatchCreate(ctx context.Context, invoices []*entities.Invoice) error
	BatchUpdate(ctx context.Context, invoices []*entities.Invoice) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripeInvoiceID string) (bool, error)
	UserHasInvoices(ctx context.Context, userID uuid.UUID) (bool, error)
	SubscriptionHasInvoices(ctx context.Context, subscriptionID uuid.UUID) (bool, error)

	// Admin operations
	GetAllInvoices(ctx context.Context, limit, offset int) ([]*entities.Invoice, error)
	GetInvoicesWithDetails(ctx context.Context, limit, offset int) ([]*InvoiceWithDetails, error)
}

// WebhookEventRepository defines interface for webhook event data operations
type WebhookEventRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, webhookEvent *entities.WebhookEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.WebhookEvent, error)
	Update(ctx context.Context, webhookEvent *entities.WebhookEvent) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Webhook event operations
	GetByStripeEventID(ctx context.Context, stripeEventID string) (*entities.WebhookEvent, error)
	GetByType(ctx context.Context, eventType string, limit, offset int) ([]*entities.WebhookEvent, error)
	GetProcessedEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error)
	GetUnprocessedEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error)
	GetFailedEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error)

	// Webhook event management
	MarkAsProcessed(ctx context.Context, eventID uuid.UUID) error
	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error
	RetryFailedEvents(ctx context.Context, limit int) ([]*entities.WebhookEvent, error)

	// Batch operations
	BatchCreate(ctx context.Context, webhookEvents []*entities.WebhookEvent) error
	BatchUpdate(ctx context.Context, webhookEvents []*entities.WebhookEvent) error
	BatchDelete(ctx context.Context, eventIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByStripeID(ctx context.Context, stripeEventID string) (bool, error)
	HasUnprocessedEvents(ctx context.Context) (bool, error)

	// Admin operations
	GetAllEvents(ctx context.Context, limit, offset int) ([]*entities.WebhookEvent, error)
	GetEventsWithDetails(ctx context.Context, limit, offset int) ([]*WebhookEventWithDetails, error)
	GetEventsByDateRange(ctx context.Context, startDate, endDate interface{}, limit, offset int) ([]*entities.WebhookEvent, error)
}

// Stats and detail types

// PaymentStats represents payment statistics
type PaymentStats struct {
	TotalPayments      int64   `json:"total_payments"`
	PendingPayments    int64   `json:"pending_payments"`
	ProcessingPayments int64   `json:"processing_payments"`
	SucceededPayments  int64   `json:"succeeded_payments"`
	FailedPayments    int64   `json:"failed_payments"`
	RefundedPayments  int64   `json:"refunded_payments"`
	TotalAmount       int64   `json:"total_amount"`
	RefundedAmount    int64   `json:"refunded_amount"`
	NetAmount         int64   `json:"net_amount"`
	AverageAmount     float64 `json:"average_amount"`
	PaymentsToday    int64   `json:"payments_today"`
	PaymentsThisWeek  int64   `json:"payments_this_week"`
	PaymentsThisMonth int64   `json:"payments_this_month"`
}

// UserPaymentStats represents payment statistics for a user
type UserPaymentStats struct {
	TotalPayments     int64   `json:"total_payments"`
	SuccessfulPayments int64   `json:"successful_payments"`
	FailedPayments    int64   `json:"failed_payments"`
	RefundedPayments  int64   `json:"refunded_payments"`
	TotalAmount       int64   `json:"total_amount"`
	RefundedAmount    int64   `json:"refunded_amount"`
	NetAmount         int64   `json:"net_amount"`
	AverageAmount     float64 `json:"average_amount"`
	LastPaymentDate   interface{} `json:"last_payment_date"`
}

// RevenueStats represents revenue statistics
type RevenueStats struct {
	TotalRevenue      int64   `json:"total_revenue"`
	GrossRevenue      int64   `json:"gross_revenue"`
	RefundedAmount    int64   `json:"refunded_amount"`
	NetRevenue       int64   `json:"net_revenue"`
	RevenueGrowth    float64 `json:"revenue_growth_percentage"`
	RevenueToday      int64   `json:"revenue_today"`
	RevenueThisWeek  int64   `json:"revenue_this_week"`
	RevenueThisMonth int64   `json:"revenue_this_month"`
}

// PaymentWithDetails represents a payment with additional details
type PaymentWithDetails struct {
	*entities.Payment
	User           *entities.User         `json:"user"`
	Subscription    *entities.Subscription `json:"subscription,omitempty"`
	PaymentMethod  *entities.PaymentMethod `json:"payment_method,omitempty"`
	Refund         *entities.Refund        `json:"refund,omitempty"`
	DaysSinceCreated int                  `json:"days_since_created"`
}

// PaymentMethodWithDetails represents a payment method with additional details
type PaymentMethodWithDetails struct {
	*entities.PaymentMethod
	User            *entities.User `json:"user"`
	PaymentsCount   int64         `json:"payments_count"`
	LastUsedAt      interface{}    `json:"last_used_at"`
	IsExpired       bool          `json:"is_expired"`
}

// RefundStats represents refund statistics
type RefundStats struct {
	TotalRefunds     int64   `json:"total_refunds"`
	PendingRefunds   int64   `json:"pending_refunds"`
	SucceededRefunds int64   `json:"succeeded_refunds"`
	FailedRefunds    int64   `json:"failed_refunds"`
	TotalAmount      int64   `json:"total_amount"`
	AverageAmount    float64 `json:"average_amount"`
	RefundsToday     int64   `json:"refunds_today"`
	RefundsThisWeek int64   `json:"refunds_this_week"`
	RefundsThisMonth int64   `json:"refunds_this_month"`
}

// UserRefundStats represents refund statistics for a user
type UserRefundStats struct {
	TotalRefunds     int64   `json:"total_refunds"`
	SucceededRefunds int64   `json:"succeeded_refunds"`
	FailedRefunds    int64   `json:"failed_refunds"`
	TotalAmount      int64   `json:"total_amount"`
	AverageAmount    float64 `json:"average_amount"`
	LastRefundDate   interface{} `json:"last_refund_date"`
}

// RefundWithDetails represents a refund with additional details
type RefundWithDetails struct {
	*entities.Refund
	Payment *entities.Payment `json:"payment"`
	User     *entities.User    `json:"user"`
}

// InvoiceStats represents invoice statistics
type InvoiceStats struct {
	TotalInvoices      int64   `json:"total_invoices"`
	DraftInvoices      int64   `json:"draft_invoices"`
	OpenInvoices       int64   `json:"open_invoices"`
	PaidInvoices       int64   `json:"paid_invoices"`
	VoidInvoices      int64   `json:"void_invoices"`
	UncollectibleInvoices int64 `json:"uncollectible_invoices"`
	OverdueInvoices   int64   `json:"overdue_invoices"`
	TotalAmount        int64   `json:"total_amount"`
	PaidAmount        int64   `json:"paid_amount"`
	OutstandingAmount  int64   `json:"outstanding_amount"`
	InvoicesToday     int64   `json:"invoices_today"`
	InvoicesThisWeek  int64   `json:"invoices_this_week"`
	InvoicesThisMonth int64   `json:"invoices_this_month"`
}

// UserInvoiceStats represents invoice statistics for a user
type UserInvoiceStats struct {
	TotalInvoices     int64   `json:"total_invoices"`
	PaidInvoices      int64   `json:"paid_invoices"`
	OverdueInvoices   int64   `json:"overdue_invoices"`
	TotalAmount       int64   `json:"total_amount"`
	PaidAmount        int64   `json:"paid_amount"`
	OutstandingAmount  int64   `json:"outstanding_amount"`
	LastInvoiceDate   interface{} `json:"last_invoice_date"`
}

// InvoiceWithDetails represents an invoice with additional details
type InvoiceWithDetails struct {
	*entities.Invoice
	User         *entities.User         `json:"user"`
	Subscription  *entities.Subscription `json:"subscription,omitempty"`
	Payments     []*entities.Payment   `json:"payments,omitempty"`
	DaysOverdue  int                  `json:"days_overdue"`
}

// WebhookEventWithDetails represents a webhook event with additional details
type WebhookEventWithDetails struct {
	*entities.WebhookEvent
	ProcessedAt interface{} `json:"processed_at,omitempty"`
	RetryCount  int         `json:"retry_count"`
}