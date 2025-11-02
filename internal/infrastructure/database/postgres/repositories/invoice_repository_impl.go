package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// invoiceRepositoryImpl implements InvoiceRepository interface
type invoiceRepositoryImpl struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new invoice repository
func NewInvoiceRepository(db *gorm.DB) repositories.InvoiceRepository {
	return &invoiceRepositoryImpl{
		db: db,
	}
}

// Create creates a new invoice
func (r *invoiceRepositoryImpl) Create(ctx context.Context, invoice *entities.Invoice) error {
	if err := r.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

// GetByID retrieves an invoice by ID
func (r *invoiceRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetByStripeID retrieves an invoice by Stripe ID
func (r *invoiceRepositoryImpl) GetByStripeID(ctx context.Context, stripeID string) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.db.WithContext(ctx).Where("stripe_invoice_id = ?", stripeID).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice by Stripe ID: %w", err)
	}
	return &invoice, nil
}

// Update updates an invoice
func (r *invoiceRepositoryImpl) Update(ctx context.Context, invoice *entities.Invoice) error {
	if err := r.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

// Delete deletes an invoice
func (r *invoiceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Invoice{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

// GetByUserID retrieves invoices for a user
func (r *invoiceRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices for user: %w", err)
	}
	return invoices, nil
}

// GetBySubscriptionID retrieves invoices for a subscription
func (r *invoiceRepositoryImpl) GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices for subscription: %w", err)
	}
	return invoices, nil
}

// GetByStatus retrieves invoices by status
func (r *invoiceRepositoryImpl) GetByStatus(ctx context.Context, status entities.InvoiceStatus, limit, offset int) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by status: %w", err)
	}
	return invoices, nil
}

// GetByDateRange retrieves invoices within a date range
func (r *invoiceRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices by date range: %w", err)
	}
	return invoices, nil
}

// GetOverdue retrieves overdue invoices
func (r *invoiceRepositoryImpl) GetOverdue(ctx context.Context, limit, offset int) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("status = ? AND due_date < NOW()", entities.InvoiceStatusOpen).
		Order("due_date ASC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	return invoices, nil
}

// GetTotalAmountByDateRange retrieves total amount of invoices within a date range
func (r *invoiceRepositoryImpl) GetTotalAmountByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	var totalAmount float64
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []entities.InvoiceStatus{entities.InvoiceStatusPaid, entities.InvoiceStatusPaidOutOfBand}).
		Select("COALESCE(SUM(total), 0)").
		Scan(&totalAmount).Error; err != nil {
		return 0, fmt.Errorf("failed to get total invoice amount by date range: %w", err)
	}
	return totalAmount, nil
}

// GetCountByStatus retrieves count of invoices by status
func (r *invoiceRepositoryImpl) GetCountByStatus(ctx context.Context, status entities.InvoiceStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get invoice count by status: %w", err)
	}
	return count, nil
}

// GetAnalytics retrieves invoice analytics
func (r *invoiceRepositoryImpl) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*repositories.InvoiceAnalytics, error) {
	var analytics repositories.InvoiceAnalytics
	
	// Get total invoiced amount
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(total), 0)").
		Scan(&analytics.TotalInvoicedAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total invoiced amount: %w", err)
	}
	
	// Get total invoices
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get total invoices: %w", err)
	}
	
	// Get paid invoices
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []entities.InvoiceStatus{entities.InvoiceStatusPaid, entities.InvoiceStatusPaidOutOfBand}).
		Count(&analytics.PaidInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get paid invoices: %w", err)
	}
	
	// Get unpaid invoices
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []entities.InvoiceStatus{entities.InvoiceStatusOpen, entities.InvoiceStatusDraft}).
		Count(&analytics.UnpaidInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get unpaid invoices: %w", err)
	}
	
	// Get overdue invoices
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ? AND status = ? AND due_date < NOW()", startDate, endDate, entities.InvoiceStatusOpen).
		Count(&analytics.OverdueInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	
	// Get average invoice amount
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(AVG(total), 0)").
		Scan(&analytics.AverageInvoiceAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get average invoice amount: %w", err)
	}
	
	return &analytics, nil
}

// GetDailyRevenue retrieves daily revenue from invoices for a date range
func (r *invoiceRepositoryImpl) GetDailyRevenue(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyInvoiceRevenue, error) {
	var dailyRevenue []repositories.DailyInvoiceRevenue
	
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Select("DATE(created_at) as date, COALESCE(SUM(total), 0) as revenue, COUNT(*) as invoices").
		Where("created_at BETWEEN ? AND ? AND status IN ?", startDate, endDate, []entities.InvoiceStatus{entities.InvoiceStatusPaid, entities.InvoiceStatusPaidOutOfBand}).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyRevenue).Error; err != nil {
		return nil, fmt.Errorf("failed to get daily invoice revenue: %w", err)
	}
	
	return dailyRevenue, nil
}

// GetUnpaidInvoices retrieves unpaid invoices for a user
func (r *invoiceRepositoryImpl) GetUnpaidInvoices(ctx context.Context, userID uuid.UUID) ([]*entities.Invoice, error) {
	var invoices []*entities.Invoice
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN ?", userID, []entities.InvoiceStatus{entities.InvoiceStatusOpen, entities.InvoiceStatusDraft}).
		Order("due_date ASC").
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get unpaid invoices for user: %w", err)
	}
	return invoices, nil
}

// BatchCreate creates multiple invoices in a single transaction
func (r *invoiceRepositoryImpl) BatchCreate(ctx context.Context, invoices []*entities.Invoice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, invoice := range invoices {
			if err := tx.Create(invoice).Error; err != nil {
				return fmt.Errorf("failed to create invoice: %w", err)
			}
		}
		return nil
	})
}

// BatchUpdate updates multiple invoices in a single transaction
func (r *invoiceRepositoryImpl) BatchUpdate(ctx context.Context, invoices []*entities.Invoice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, invoice := range invoices {
			if err := tx.Save(invoice).Error; err != nil {
				return fmt.Errorf("failed to update invoice: %w", err)
			}
		}
		return nil
	})
}