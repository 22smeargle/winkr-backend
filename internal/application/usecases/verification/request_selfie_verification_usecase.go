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

// RequestSelfieVerificationUseCase handles selfie verification requests
type RequestSelfieVerificationUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewRequestSelfieVerificationUseCase creates a new use case
func NewRequestSelfieVerificationUseCase(verificationService *services.VerificationWorkflowService) *RequestSelfieVerificationUseCase {
	return &RequestSelfieVerificationUseCase{
		verificationService: verificationService,
	}
}

// RequestSelfieVerificationInput represents input for selfie verification request
type RequestSelfieVerificationInput struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	IPAddress string    `json:"ip_address" validate:"required"`
	UserAgent string    `json:"user_agent" validate:"required"`
}

// RequestSelfieVerificationOutput represents output of selfie verification request
type RequestSelfieVerificationOutput struct {
	VerificationID uuid.UUID `json:"verification_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	CanSubmit     bool      `json:"can_submit"`
}

// Execute executes the selfie verification request use case
func (uc *RequestSelfieVerificationUseCase) Execute(ctx context.Context, input RequestSelfieVerificationInput) (*RequestSelfieVerificationOutput, error) {
	logger.Info("Executing selfie verification request", "user_id", input.UserID)

	// Request selfie verification
	verification, err := uc.verificationService.RequestSelfieVerification(ctx, input.UserID, input.IPAddress, input.UserAgent)
	if err != nil {
		logger.Error("Failed to request selfie verification", err, "user_id", input.UserID)
		return nil, errors.NewAppError(400, "Failed to request selfie verification", err.Error())
	}

	output := &RequestSelfieVerificationOutput{
		VerificationID: verification.ID,
		Status:        verification.Status.String(),
		Message:       "Selfie verification requested successfully",
		CanSubmit:     true,
	}

	logger.Info("Selfie verification requested successfully", "user_id", input.UserID, "verification_id", verification.ID)
	return output, nil
}