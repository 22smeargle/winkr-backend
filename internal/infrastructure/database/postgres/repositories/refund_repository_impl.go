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

// refundRepositoryImpl implements RefundRepository interface
type refundRepositoryImpl struct {
	db *gorm.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *gorm.DB) repositories.RefundRepository {
	return &refundRepositoryImpl{
		db: db,
	}
}

// Create creates a new refund
func (r *refundRepositoryImpl) Create(ctx context.Context, refund *entities.Refund) error {
	if err := r.db.WithContext(ctx).Create(refund).Error; err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}
	return nil
}

// GetByID retrieves a refund by ID
func (r *refundRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Refund, error) {
	var refund entities.Refund
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}
	return &refund, nil
}

// GetByStripeID retrieves a refund by Stripe ID
func (r *refundRepositoryImpl) GetByStripeID(ctx context.Context, stripeID string) (*entities.Refund, error) {
	var refund entities.Refund
	if err := r.db.WithContext(ctx).Where("stripe_refund_id = ?", stripeID).First(&refund).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to get refund by Stripe ID: %w", err)
	}
	return &refund, nil
}

// Update updates a refund
func (r *refundRepositoryImpl) Update(ctx context.Context, refund *entities.Refund) error {
	if err := r.db.WithContext(ctx).Save(refund).Error; err != nil {
		return fmt.Errorf("failed to update refund: %w", err)
	}
	return nil
}

// Delete deletes a refund
func (r *refundRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Refund{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete refund: %w", err)
	}
	return nil
}

// GetByPaymentID retrieves refunds for a payment
func (r *refundRepositoryImpl) GetByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entities.Refund, error) {
	var refunds []*entities.Refund
	if err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get refunds for payment: %w", err)
	}
	return refunds, nil
}

// GetByUserID retrieves refunds for a user
func (r *refundRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Refund, error) {
	var refunds []*entities.Refund
	if err := r.db.WithContext(ctx).
		Joins("JOIN payments ON refunds.payment_id = payments.id").
		Where("payments.user_id = ?", userID).
		Order("refunds.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get refunds for user: %w", err)
	}
	return refunds, nil
}

// GetByStatus retrieves refunds by status
func (r *refundRepositoryImpl) GetByStatus(ctx context.Context, status entities.RefundStatus, limit, offset int) ([]*entities.Refund, error) {
	var refunds []*entities.Refund
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get refunds by status: %w", err)
	}
	return refunds, nil
}

// GetByDateRange retrieves refunds within a date range
func (r *refundRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entities.Refund, error) {
	var refunds []*entities.Refund
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get refunds by date range: %w", err)
	}
	return refunds, nil
}

// GetTotalAmountByDateRange retrieves total amount of refunds within a date range
func (r *refundRepositoryImpl) GetTotalAmountByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	var totalAmount float64
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusSucceeded).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalAmount).Error; err != nil {
		return 0, fmt.Errorf("failed to get total refund amount by date range: %w", err)
	}
	return totalAmount, nil
}

// GetCountByStatus retrieves count of refunds by status
func (r *refundRepositoryImpl) GetCountByStatus(ctx context.Context, status entities.RefundStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get refund count by status: %w", err)
	}
	return count, nil
}

// GetAnalytics retrieves refund analytics
func (r *refundRepositoryImpl) GetAnalytics(ctx context.Context, startDate, endDate time.Time) (*repositories.RefundAnalytics, error) {
	var analytics repositories.RefundAnalytics
	
	// Get total refunded amount
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusSucceeded).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&analytics.TotalRefundedAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total refunded amount: %w", err)
	}
	
	// Get total refunds
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&analytics.TotalRefunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get total refunds: %w", err)
	}
	
	// Get successful refunds
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusSucceeded).
		Count(&analytics.SuccessfulRefunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful refunds: %w", err)
	}
	
	// Get failed refunds
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusFailed).
		Count(&analytics.FailedRefunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed refunds: %w", err)
	}
	
	// Get pending refunds
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusPending).
		Count(&analytics.PendingRefunds).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending refunds: %w", err)
	}
	
	// Get average refund amount
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusSucceeded).
		Select("COALESCE(AVG(amount), 0)").
		Scan(&analytics.AverageRefundAmount).Error; err != nil {
		return nil, fmt.Errorf("failed to get average refund amount: %w", err)
	}
	
	return &analytics, nil
}

// GetRefundReasons retrieves refund reasons analytics
func (r *refundRepositoryImpl) GetRefundReasons(ctx context.Context, startDate, endDate time.Time) ([]repositories.RefundReason, error) {
	var refundReasons []repositories.RefundReason
	
	if err := r.db.WithContext(ctx).
		Model(&entities.Refund{}).
		Select("reason, COUNT(*) as count, COALESCE(SUM(amount), 0) as total_amount").
		Where("created_at BETWEEN ? AND ? AND status = ?", startDate, endDate, entities.RefundStatusSucceeded).
		Group("reason").
		Order("count DESC").
		Scan(&refundReasons).Error; err != nil {
		return nil, fmt.Errorf("failed to get refund reasons: %w", err)
	}
	
	return refundReasons, nil
}

// BatchCreate creates multiple refunds in a single transaction
func (r *refundRepositoryImpl) BatchCreate(ctx context.Context, refunds []*entities.Refund) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, refund := range refunds {
			if err := tx.Create(refund).Error; err != nil {
				return fmt.Errorf("failed to create refund: %w", err)
			}
		}
		return nil
	})
}

// BatchUpdate updates multiple refunds in a single transaction
func (r *refundRepositoryImpl) BatchUpdate(ctx context.Context, refunds []*entities.Refund) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, refund := range refunds {
			if err := tx.Save(refund).Error; err != nil {
				return fmt.Errorf("failed to update refund: %w", err)
			}
		}
		return nil
	})
}