package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetPaymentMethodsUseCase retrieves user's payment methods
type GetPaymentMethodsUseCase struct {
	userRepo         repositories.UserRepository
	paymentMethodRepo repositories.PaymentMethodRepository
}

// NewGetPaymentMethodsUseCase creates a new GetPaymentMethodsUseCase
func NewGetPaymentMethodsUseCase(
	userRepo repositories.UserRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
) *GetPaymentMethodsUseCase {
	return &GetPaymentMethodsUseCase{
		userRepo:         userRepo,
		paymentMethodRepo: paymentMethodRepo,
	}
}

// Execute retrieves user's payment methods
func (uc *GetPaymentMethodsUseCase) Execute(ctx context.Context, userID uuid.UUID, paymentMethodType string, limit, offset int) ([]*entities.PaymentMethod, error) {
	logger.Info("Getting user payment methods", map[string]interface{}{
		"user_id":           userID,
		"payment_method_type": paymentMethodType,
		"limit":             limit,
		"offset":            offset,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrInvalidUserID
	}

	var paymentMethods []*entities.PaymentMethod

	// Get payment methods by type if specified
	if paymentMethodType != "" {
		paymentMethods, err = uc.paymentMethodRepo.GetUserPaymentMethodsByType(ctx, userID, paymentMethodType, limit, offset)
		if err != nil {
			logger.Error("Failed to get user payment methods by type", err, map[string]interface{}{
				"user_id":           userID,
				"payment_method_type": paymentMethodType,
			})
			return nil, fmt.Errorf("failed to get user payment methods by type: %w", err)
		}
	} else {
		paymentMethods, err = uc.paymentMethodRepo.GetUserPaymentMethods(ctx, userID, limit, offset)
		if err != nil {
			logger.Error("Failed to get user payment methods", err, map[string]interface{}{
				"user_id": userID,
			})
			return nil, fmt.Errorf("failed to get user payment methods: %w", err)
		}
	}

	logger.Info("Retrieved user payment methods", map[string]interface{}{
		"user_id":    userID,
		"count":       len(paymentMethods),
	})

	return paymentMethods, nil
}

// GetDefaultPaymentMethodUseCase retrieves user's default payment method
type GetDefaultPaymentMethodUseCase struct {
	userRepo         repositories.UserRepository
	paymentMethodRepo repositories.PaymentMethodRepository
}

// NewGetDefaultPaymentMethodUseCase creates a new GetDefaultPaymentMethodUseCase
func NewGetDefaultPaymentMethodUseCase(
	userRepo repositories.UserRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
) *GetDefaultPaymentMethodUseCase {
	return &GetDefaultPaymentMethodUseCase{
		userRepo:         userRepo,
		paymentMethodRepo: paymentMethodRepo,
	}
}

// Execute retrieves user's default payment method
func (uc *GetDefaultPaymentMethodUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entities.PaymentMethod, error) {
	logger.Info("Getting user default payment method", map[string]interface{}{
		"user_id": userID,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrInvalidUserID
	}

	// Get user's default payment method
	paymentMethod, err := uc.paymentMethodRepo.GetUserDefaultPaymentMethod(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user default payment method", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, ErrUserHasNoDefaultPaymentMethod
	}

	logger.Info("Retrieved user default payment method", map[string]interface{}{
		"user_id":          userID,
		"payment_method_id": paymentMethod.ID,
		"type":             paymentMethod.Type,
	})

	return paymentMethod, nil
}

// DeletePaymentMethodUseCase deletes a user's payment method
type DeletePaymentMethodUseCase struct {
	userRepo         repositories.UserRepository
	paymentMethodRepo repositories.PaymentMethodRepository
}

// NewDeletePaymentMethodUseCase creates a new DeletePaymentMethodUseCase
func NewDeletePaymentMethodUseCase(
	userRepo repositories.UserRepository,
	paymentMethodRepo repositories.PaymentMethodRepository,
) *DeletePaymentMethodUseCase {
	return &DeletePaymentMethodUseCase{
		userRepo:         userRepo,
		paymentMethodRepo: paymentMethodRepo,
	}
}

// Execute deletes a user's payment method
func (uc *DeletePaymentMethodUseCase) Execute(ctx context.Context, userID, paymentMethodID uuid.UUID) error {
	logger.Info("Deleting payment method", map[string]interface{}{
		"user_id":          userID,
		"payment_method_id": paymentMethodID,
	})

	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user", err, map[string]interface{}{
			"user_id": userID,
		})
		return ErrInvalidUserID
	}

	// Get payment method
	paymentMethod, err := uc.paymentMethodRepo.GetByID(ctx, paymentMethodID)
	if err != nil {
		logger.Error("Failed to get payment method", err, map[string]interface{}{
			"user_id":          userID,
			"payment_method_id": paymentMethodID,
		})
		return ErrPaymentMethodNotFound
	}

	// Check if payment method belongs to user
	if paymentMethod.UserID != userID {
		logger.Error("Payment method does not belong to user", nil, map[string]interface{}{
			"user_id":          userID,
			"payment_method_id": paymentMethodID,
		})
		return ErrPaymentAccessDenied
	}

	// Check if payment method is default
	if paymentMethod.IsDefault {
		logger.Error("Cannot delete default payment method", nil, map[string]interface{}{
			"user_id":          userID,
			"payment_method_id": paymentMethodID,
		})
		return ErrDefaultPaymentMethodRequired
	}

	// Delete payment method
	err = uc.paymentMethodRepo.Delete(ctx, paymentMethodID)
	if err != nil {
		logger.Error("Failed to delete payment method", err, map[string]interface{}{
			"user_id":          userID,
			"payment_method_id": paymentMethodID,
		})
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	logger.Info("Payment method deleted successfully", map[string]interface{}{
		"user_id":          userID,
		"payment_method_id": paymentMethodID,
	})

	return nil
}