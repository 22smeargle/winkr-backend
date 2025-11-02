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

// SubmitDocumentVerificationUseCase handles document verification submissions
type SubmitDocumentVerificationUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewSubmitDocumentVerificationUseCase creates a new use case
func NewSubmitDocumentVerificationUseCase(verificationService *services.VerificationWorkflowService) *SubmitDocumentVerificationUseCase {
	return &SubmitDocumentVerificationUseCase{
		verificationService: verificationService,
	}
}

// SubmitDocumentVerificationInput represents input for document verification submission
type SubmitDocumentVerificationInput struct {
	VerificationID uuid.UUID `json:"verification_id" validate:"required"`
	PhotoKey       string    `json:"photo_key" validate:"required"`
	PhotoURL       string    `json:"photo_url" validate:"required"`
}

// SubmitDocumentVerificationOutput represents output of document verification submission
type SubmitDocumentVerificationOutput struct {
	VerificationID uuid.UUID `json:"verification_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	DocumentType   *string    `json:"document_type,omitempty"`
	AIScore        *float64  `json:"ai_score,omitempty"`
	RequiresReview bool      `json:"requires_review"`
}

// Execute executes document verification submission use case
func (uc *SubmitDocumentVerificationUseCase) Execute(ctx context.Context, input SubmitDocumentVerificationInput) (*SubmitDocumentVerificationOutput, error) {
	logger.Info("Executing document verification submission", "verification_id", input.VerificationID)

	// Submit document verification
	verification, err := uc.verificationService.SubmitDocumentVerification(ctx, input.VerificationID, input.PhotoKey, input.PhotoURL)
	if err != nil {
		logger.Error("Failed to submit document verification", err, "verification_id", input.VerificationID)
		return nil, errors.NewAppError(400, "Failed to submit document verification", err.Error())
	}

	output := &SubmitDocumentVerificationOutput{
		VerificationID: verification.ID,
		Status:        verification.Status.String(),
		Message:       "Document verification submitted successfully",
		DocumentType:   verification.DocumentType,
		AIScore:        verification.AIScore,
		RequiresReview: verification.Status.IsPending(),
	}

	logger.Info("Document verification submitted successfully", "verification_id", input.VerificationID, "status", verification.Status, "document_type", verification.DocumentType)
	return output, nil
}