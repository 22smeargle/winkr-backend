package verification

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ProcessVerificationResultUseCase handles processing verification results
type ProcessVerificationResultUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewProcessVerificationResultUseCase creates a new use case
func NewProcessVerificationResultUseCase(verificationService *services.VerificationWorkflowService) *ProcessVerificationResultUseCase {
	return &ProcessVerificationResultUseCase{
		verificationService: verificationService,
	}
}

// ProcessVerificationResultInput represents input for processing verification result
type ProcessVerificationResultInput struct {
	VerificationID uuid.UUID `json:"verification_id" validate:"required"`
	Approved        bool      `json:"approved" validate:"required"`
	Reason          string    `json:"reason,omitempty"`
	ReviewedBy       uuid.UUID `json:"reviewed_by" validate:"required"`
}

// ProcessVerificationResultOutput represents output of processing verification result
type ProcessVerificationResultOutput struct {
	VerificationID uuid.UUID `json:"verification_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
}

// Execute executes the process verification result use case
func (uc *ProcessVerificationResultUseCase) Execute(ctx context.Context, input ProcessVerificationResultInput) (*ProcessVerificationResultOutput, error) {
	logger.Info("Processing verification result", "verification_id", input.VerificationID, "approved", input.Approved, "reviewed_by", input.ReviewedBy)

	// Process verification result
	err := uc.verificationService.ProcessVerificationResult(ctx, input.VerificationID, input.Approved, input.Reason, input.ReviewedBy)
	if err != nil {
		logger.Error("Failed to process verification result", err, "verification_id", input.VerificationID)
		return nil, errors.NewAppError(400, "Failed to process verification result", err.Error())
	}

	status := "rejected"
	message := "Verification rejected"
	if input.Approved {
		status = "approved"
		message = "Verification approved"
	}

	output := &ProcessVerificationResultOutput{
		VerificationID: input.VerificationID,
		Status:        status,
		Message:       message,
	}

	logger.Info("Verification result processed successfully", "verification_id", input.VerificationID, "status", status)
	return output, nil
}