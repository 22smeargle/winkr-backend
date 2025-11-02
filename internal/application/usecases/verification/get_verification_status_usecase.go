package verification

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetVerificationStatusUseCase handles getting verification status
type GetVerificationStatusUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewGetVerificationStatusUseCase creates a new use case
func NewGetVerificationStatusUseCase(verificationService *services.VerificationWorkflowService) *GetVerificationStatusUseCase {
	return &GetVerificationStatusUseCase{
		verificationService: verificationService,
	}
}

// GetVerificationStatusInput represents input for getting verification status
type GetVerificationStatusInput struct {
	UserID uuid.UUID              `json:"user_id" validate:"required"`
	Type   entities.VerificationType `json:"type" validate:"required,oneof=selfie document"`
}

// GetVerificationStatusOutput represents output of getting verification status
type GetVerificationStatusOutput struct {
	UserID          uuid.UUID                    `json:"user_id"`
	Type            entities.VerificationType      `json:"type"`
	Status          string                         `json:"status"`
	HasVerification bool                           `json:"has_verification"`
	CanRequest      bool                           `json:"can_request"`
	LastVerification *entities.Verification           `json:"last_verification,omitempty"`
	VerificationLevel entities.VerificationLevel       `json:"verification_level"`
}

// Execute executes the get verification status use case
func (uc *GetVerificationStatusUseCase) Execute(ctx context.Context, input GetVerificationStatusInput) (*GetVerificationStatusOutput, error) {
	logger.Info("Getting verification status", "user_id", input.UserID, "type", input.Type)

	// Get verification status
	result, err := uc.verificationService.GetVerificationStatus(ctx, input.UserID, input.Type)
	if err != nil {
		logger.Error("Failed to get verification status", err, "user_id", input.UserID, "type", input.Type)
		return nil, errors.NewAppError(400, "Failed to get verification status", err.Error())
	}

	output := &GetVerificationStatusOutput{
		UserID:          result.UserID,
		Type:            result.Type,
		Status:          result.Status.String(),
		HasVerification: result.HasVerification,
		CanRequest:      result.CanRequest,
		LastVerification: result.LastVerification,
		VerificationLevel: result.VerificationLevel,
	}

	logger.Info("Verification status retrieved successfully", "user_id", input.UserID, "type", input.Type, "status", result.Status, "level", result.VerificationLevel)
	return output, nil
}