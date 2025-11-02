package verification

import (
	"context"
	"fmt"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetPendingVerificationsUseCase handles getting pending verifications for admin review
type GetPendingVerificationsUseCase struct {
	verificationService *services.VerificationWorkflowService
}

// NewGetPendingVerificationsUseCase creates a new use case
func NewGetPendingVerificationsUseCase(verificationService *services.VerificationWorkflowService) *GetPendingVerificationsUseCase {
	return &GetPendingVerificationsUseCase{
		verificationService: verificationService,
	}
}

// GetPendingVerificationsInput represents input for getting pending verifications
type GetPendingVerificationsInput struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}

// GetPendingVerificationsOutput represents output of getting pending verifications
type GetPendingVerificationsOutput struct {
	Verifications []*VerificationWithUser `json:"verifications"`
	TotalCount   int                    `json:"total_count"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
}

// VerificationWithUser represents verification with user information
type VerificationWithUser struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Type             string    `json:"type"`
	Status           string    `json:"status"`
	PhotoURL         string    `json:"photo_url"`
	AIScore          *float64  `json:"ai_score,omitempty"`
	RejectionReason  *string    `json:"rejection_reason,omitempty"`
	CreatedAt        string    `json:"created_at"`
	User             *UserInfo  `json:"user,omitempty"`
}

// UserInfo represents user information for verification display
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// Execute executes the get pending verifications use case
func (uc *GetPendingVerificationsUseCase) Execute(ctx context.Context, input GetPendingVerificationsInput) (*GetPendingVerificationsOutput, error) {
	logger.Info("Getting pending verifications", "limit", input.Limit, "offset", input.Offset)

	// Get pending verifications
	verifications, err := uc.verificationService.GetPendingVerifications(ctx, input.Limit, input.Offset)
	if err != nil {
		logger.Error("Failed to get pending verifications", err)
		return nil, errors.NewAppError(500, "Failed to get pending verifications", err.Error())
	}

	// Convert to output format
	verificationOutputs := make([]*VerificationWithUser, 0, len(verifications))
	for i, verification := range verifications {
		verificationOutput := &VerificationWithUser{
			ID:               verification.ID,
			UserID:           verification.UserID,
			Type:             string(verification.Type),
			Status:           verification.Status.String(),
			PhotoURL:         verification.PhotoURL,
			AIScore:          verification.AIScore,
			RejectionReason:  verification.RejectionReason,
			CreatedAt:        verification.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Add user information if available
		if verification.User != nil {
			verificationOutput.User = &UserInfo{
				ID:        verification.User.ID,
				FirstName: verification.User.FirstName,
				LastName:  verification.User.LastName,
				Email:     verification.User.Email,
			}
		}

		verificationOutputs[i] = verificationOutput
	}

	output := &GetPendingVerificationsOutput{
		Verifications: verificationOutputs,
		TotalCount:   len(verificationOutputs),
		Limit:        input.Limit,
		Offset:       input.Offset,
	}

	logger.Info("Pending verifications retrieved successfully", "count", len(verificationOutputs))
	return output, nil
}