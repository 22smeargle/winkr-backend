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

// paymentRepositoryImpl implements the PaymentRepository interface
type paymentRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) repositories.PaymentRepository {
	return &paymentRepositoryImpl{
		db: db,
	}
}

// Create creates a new payment
func (r *paymentRepositoryImpl) Create(ctx context.Context, payment *entities.Payment) error {
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetByID retrieves a payment by ID
func (r *paymentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error) {
	var payment entities.Payment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}

// GetByStripePaymentIntentID retrieves a payment by Stripe payment intent ID
func (r *paymentRepositoryImpl) GetByStripePaymentIntentID(ctx context.Context, paymentIntentID string) (*entities.Payment, error) {
	var payment entities.Payment
	if err := r.db.WithContext(ctx).Where("stripe_payment_intent_id = ?", paymentIntentID).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by payment intent ID: %w", err)
	}
	return &payment, nil
}

// Update updates a payment
func (r *paymentRepositoryImpl) Update(ctx context.Context, payment *entities.Payment) error {
	if err := r.db.WithContext(ctx).Save(payment).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}

// Delete deletes a payment
func (r *paymentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Payment{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}
	return nil
}

// GetByUserID retrieves payments for a user
func (r *paymentRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments for user: %w", err)
	}
	return payments, nil
}

// GetByStatus retrieves payments by status
func (r *paymentRepositoryImpl) GetByStatus(ctx context.Context, status entities.PaymentStatus, limit, offset int) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by status: %w", err)
	}
	return payments, nil
}

// GetByDateRange retrieves payments within a date range
func (r *paymentRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get payments by date range: %w", err)
	}
	return payments, nil
}

// GetTotalAmountByDateRange retrieves total amount of payments within a date range
func (r *paymentRepositoryImpl) GetTotalAmountByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	var totalAmount float64
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusSucceeded).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount).Error; err != nil {
		return 0, fmt.Errorf("failed to get total amount by date range: %w", err)
	}
	return totalAmount, nil
}

// GetCountByStatus retrieves count of payments by status
func (r *paymentRepositoryImpl) GetCountByStatus(ctx context.Context, status entities.PaymentStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get payment count by status: %w", err)
	}
	return count, nil
}

// GetAnalytics retrieves payment analytics
func (r *paymentRepositoryImpl) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*repositories.PaymentAnalytics, error) {
	var analytics repositories.PaymentAnalytics
	
	// Get total revenue
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusSucceeded).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&analytics.TotalRevenue).Error; err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}
	
	// Get total payments
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get total payments: %w", err)
	}
	
	// Get successful payments
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusSucceeded).
		Count(&analytics.SuccessfulPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful payments: %w", err)
	}
	
	// Get failed payments
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusFailed).
		Count(&analytics.FailedPayments).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed payments: %w", err)
	}
	
	// Get average payment amount
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusSucceeded).
		Select("COALESCE(AVG(amount), 0)").
		Scan(&analytics.AveragePaymentAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get average payment amount: %w", err)
	}
	
	return &analytics, nil
}

// GetDailyRevenue retrieves daily revenue for a date range
func (r *paymentRepositoryImpl) GetDailyRevenue(ctx context.Context, startDate, endDate time.Time) ([]repositories.DailyRevenue, error) {
	var dailyRevenue []repositories.DailyRevenue
	
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Select("DATE(created_at) as date, COALESCE(SUM(amount), 0) as revenue, COUNT(*) as payments").
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.PaymentStatusSucceeded).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyRevenue).Error; err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}
	
	return dailyRevenue, nil
}

// GetTopCustomers retrieves top customers by total amount spent
func (r *paymentRepositoryImpl) GetTopCustomers(ctx context.Context, limit int) ([]repositories.TopCustomer, error) {
	var topCustomers []repositories.TopCustomer
	
	if err := r.db.WithContext(ctx).
		Model(&entities.Payment{}).
		Select("user_id, COALESCE(SUM(amount), 0) as total_spent, COUNT(*) as payment_count").
		Where("status = ?", entities.PaymentStatusSucceeded).
		Group("user_id").
		Order("total_spent DESC").
		Limit(limit).
		Scan(&topCustomers).Error; err != nil {
		return nil, fmt.Errorf("failed to get top customers: %w", err)
	}
	
	return topCustomers, nil
}

// BatchCreate creates multiple payments in a single transaction
func (r *paymentRepositoryImpl) BatchCreate(ctx context.Context, payments []*entities.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, payment := range payments {
			if err := tx.Create(payment).Error; err != nil {
				return fmt.Errorf("failed to create payment: %w", err)
			}
		}
		return nil
	})
}

// BatchUpdate updates multiple payments in a single transaction
func (r *paymentRepositoryImpl) BatchUpdate(ctx context.Context, payments []*entities.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, payment := range payments {
			if err := tx.Save(payment).Error; err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}
		return nil
	})
}