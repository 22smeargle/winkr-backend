package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// paymentMethodRepositoryImpl implements PaymentMethodRepository interface
type paymentMethodRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentMethodRepository creates a new payment method repository
func NewPaymentMethodRepository(db *gorm.DB) repositories.PaymentMethodRepository {
	return &paymentMethodRepositoryImpl{
		db: db,
	}
}

// Create creates a new payment method
func (r *paymentMethodRepositoryImpl) Create(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Create(paymentMethod).Error; err != nil {
		return fmt.Errorf("failed to create payment method: %w", err)
	}
	return nil
}

// GetByID retrieves a payment method by ID
func (r *paymentMethodRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrPaymentMethodNotFound
		}
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}
	return &paymentMethod, nil
}

// GetByStripeID retrieves a payment method by Stripe ID
func (r *paymentMethodRepositoryImpl) GetByStripeID(ctx context.Context, stripeID string) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).Where("stripe_payment_method_id = ?", stripeID).First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrPaymentMethodNotFound
		}
		return nil, fmt.Errorf("failed to get payment method by Stripe ID: %w", err)
	}
	return &paymentMethod, nil
}

// Update updates a payment method
func (r *paymentMethodRepositoryImpl) Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	if err := r.db.WithContext(ctx).Save(paymentMethod).Error; err != nil {
		return fmt.Errorf("failed to update payment method: %w", err)
	}
	return nil
}

// Delete deletes a payment method
func (r *paymentMethodRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.PaymentMethod{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}
	return nil
}

// GetByUserID retrieves payment methods for a user
func (r *paymentMethodRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods for user: %w", err)
	}
	return paymentMethods, nil
}

// GetDefaultByUserID retrieves the default payment method for a user
func (r *paymentMethodRepositoryImpl) GetDefaultByUserID(ctx context.Context, userID uuid.UUID) (*entities.PaymentMethod, error) {
	var paymentMethod entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&paymentMethod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repositories.ErrDefaultPaymentMethodNotFound
		}
		return nil, fmt.Errorf("failed to get default payment method: %w", err)
	}
	return &paymentMethod, nil
}

// SetAsDefault sets a payment method as default for a user
func (r *paymentMethodRepositoryImpl) SetAsDefault(ctx context.Context, userID, paymentMethodID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset all existing default payment methods for the user
		if err := tx.Model(&entities.PaymentMethod{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset existing default payment methods: %w", err)
		}

		// Set the new default payment method
		if err := tx.Model(&entities.PaymentMethod{}).
			Where("id = ? AND user_id = ?", paymentMethodID, userID).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set new default payment method: %w", err)
		}

		return nil
	})
}

// GetByType retrieves payment methods by type
func (r *paymentMethodRepositoryImpl) GetByType(ctx context.Context, userID uuid.UUID, paymentMethodType entities.PaymentMethodType) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, paymentMethodType).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment methods by type: %w", err)
	}
	return paymentMethods, nil
}

// GetCountByUserID retrieves the count of payment methods for a user
func (r *paymentMethodRepositoryImpl) GetCountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.PaymentMethod{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get payment method count: %w", err)
	}
	return count, nil
}

// GetExpired retrieves expired payment methods
func (r *paymentMethodRepositoryImpl) GetExpired(ctx context.Context) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < NOW()").
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired payment methods: %w", err)
	}
	return paymentMethods, nil
}

// GetExpiringSoon retrieves payment methods that will expire within the specified duration
func (r *paymentMethodRepositoryImpl) GetExpiringSoon(ctx context.Context, duration string) ([]*entities.PaymentMethod, error) {
	var paymentMethods []*entities.PaymentMethod
	if err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at BETWEEN NOW() AND NOW() + INTERVAL ?", duration).
		Find(&paymentMethods).Error; err != nil {
		return nil, fmt.Errorf("failed to get expiring payment methods: %w", err)
	}
	return paymentMethods, nil
}

// BatchCreate creates multiple payment methods in a single transaction
func (r *paymentMethodRepositoryImpl) BatchCreate(ctx context.Context, paymentMethods []*entities.PaymentMethod) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, paymentMethod := range paymentMethods {
			if err := tx.Create(paymentMethod).Error; err != nil {
				return fmt.Errorf("failed to create payment method: %w", err)
			}
		}
		return nil
	})
}

// BatchDelete deletes multiple payment methods in a single transaction
func (r *paymentMethodRepositoryImpl) BatchDelete(ctx context.Context, ids []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entities.PaymentMethod{}, ids).Error; err != nil {
		return fmt.Errorf("failed to delete payment methods: %w", err)
	}
	return nil
}