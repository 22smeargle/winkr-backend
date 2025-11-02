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

// SubmitSelfieVerificationUseCase handles selfie verification submissions
type SubmitSelfieVerificationUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewSubmitSelfieVerificationUseCase creates a new use case
func NewSubmitSelfieVerificationUseCase(verificationService *services.VerificationWorkflowService) *SubmitSelfieVerificationUseCase {
	return &SubmitSelfieVerificationUseCase{
		verificationService: verificationService,
	}
}

// SubmitSelfieVerificationInput represents input for selfie verification submission
type SubmitSelfieVerificationInput struct {
	VerificationID uuid.UUID `json:"verification_id" validate:"required"`
	PhotoKey       string    `json:"photo_key" validate:"required"`
	PhotoURL       string    `json:"photo_url" validate:"required"`
}

// SubmitSelfieVerificationOutput represents output of selfie verification submission
type SubmitSelfieVerificationOutput struct {
	VerificationID uuid.UUID `json:"verification_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	AIScore        *float64  `json:"ai_score,omitempty"`
	RequiresReview bool      `json:"requires_review"`
}

// Execute executes the selfie verification submission use case
func (uc *SubmitSelfieVerificationUseCase) Execute(ctx context.Context, input SubmitSelfieVerificationInput) (*SubmitSelfieVerificationOutput, error) {
	logger.Info("Executing selfie verification submission", "verification_id", input.VerificationID)

	// Submit selfie verification
	verification, err := uc.verificationService.SubmitSelfieVerification(ctx, input.VerificationID, input.PhotoKey, input.PhotoURL)
	if err != nil {
		logger.Error("Failed to submit selfie verification", err, "verification_id", input.VerificationID)
		return nil, errors.NewAppError(400, "Failed to submit selfie verification", err.Error())
	}

	output := &SubmitSelfieVerificationOutput{
		VerificationID: verification.ID,
		Status:        verification.Status.String(),
		Message:       "Selfie verification submitted successfully",
		AIScore:        verification.AIScore,
		RequiresReview: verification.Status.IsPending(),
	}

	logger.Info("Selfie verification submitted successfully", "verification_id", input.VerificationID, "status", verification.Status, "ai_score", verification.AIScore)
	return output, nil
}