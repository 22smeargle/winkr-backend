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

// RequestDocumentVerificationUseCase handles document verification requests
type RequestDocumentVerificationUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewRequestDocumentVerificationUseCase creates a new use case
func NewRequestDocumentVerificationUseCase(verificationService *services.VerificationWorkflowService) *RequestDocumentVerificationUseCase {
	return &RequestDocumentVerificationUseCase{
		verificationService: verificationService,
	}
}

// RequestDocumentVerificationInput represents input for document verification request
type RequestDocumentVerificationInput struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	IPAddress string    `json:"ip_address" validate:"required"`
	UserAgent string    `json:"user_agent" validate:"required"`
}

// RequestDocumentVerificationOutput represents output of document verification request
type RequestDocumentVerificationOutput struct {
	VerificationID uuid.UUID `json:"verification_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	CanSubmit     bool      `json:"can_submit"`
}

// Execute executes the document verification request use case
func (uc *RequestDocumentVerificationUseCase) Execute(ctx context.Context, input RequestDocumentVerificationInput) (*RequestDocumentVerificationOutput, error) {
	logger.Info("Executing document verification request", "user_id", input.UserID)

	// Request document verification
	verification, err := uc.verificationService.RequestDocumentVerification(ctx, input.UserID, input.IPAddress, input.UserAgent)
	if err != nil {
		logger.Error("Failed to request document verification", err, "user_id", input.UserID)
		return nil, errors.NewAppError(400, "Failed to request document verification", err.Error())
	}

	output := &RequestDocumentVerificationOutput{
		VerificationID: verification.ID,
		Status:        verification.Status.String(),
		Message:       "Document verification requested successfully",
		CanSubmit:     true,
	}

	logger.Info("Document verification requested successfully", "user_id", input.UserID, "verification_id", verification.ID)
	return output, nil
}