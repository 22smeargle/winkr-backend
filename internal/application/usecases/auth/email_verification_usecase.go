package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// EmailVerificationUseCase handles email verification
type EmailVerificationUseCase struct {
	userRepo           repositories.UserRepository
	verificationService *services.VerificationService
}

// NewEmailVerificationUseCase creates a new EmailVerificationUseCase instance
func NewEmailVerificationUseCase(
	userRepo repositories.UserRepository,
	verificationService *services.VerificationService,
) *EmailVerificationUseCase {
	return &EmailVerificationUseCase{
		userRepo:           userRepo,
		verificationService: verificationService,
	}
}

// SendVerificationRequest represents send verification email request
type SendVerificationRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
}

// SendVerificationResponse represents send verification email response
type SendVerificationResponse struct {
	Message string `json:"message"`
}

// Execute handles sending verification email use case
func (uc *EmailVerificationUseCase) ExecuteSendVerification(ctx context.Context, req *SendVerificationRequest) (*SendVerificationResponse, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is already verified
	if user.IsVerified {
		return &SendVerificationResponse{
			Message: "Email is already verified",
		}, nil
	}

	// Send verification email
	err = uc.verificationService.SendEmailVerification(ctx, user.Email, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	response := &SendVerificationResponse{
		Message: "Verification email sent successfully",
	}

	return response, nil
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token     string `json:"token" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// VerifyEmailResponse represents email verification response
type VerifyEmailResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// Execute handles email verification use case
func (uc *EmailVerificationUseCase) ExecuteVerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	// Verify email code
	result, err := uc.verificationService.VerifyEmailCode(ctx, req.Token, req.IPAddress, req.UserAgent)
	if err != nil {
		return &VerifyEmailResponse{
			Message: "Invalid or expired verification token",
			Success: false,
		}, nil // Don't expose specific error
	}

	if !result.Success {
		return &VerifyEmailResponse{
			Message: "Email verification failed",
			Success: false,
		}, nil
	}

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, result.Identifier)
	if err != nil {
		return &VerifyEmailResponse{
			Message: "User not found",
			Success: false,
		}, nil // Don't expose specific error
	}

	// Update user verification status
	user.IsVerified = true
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to update user verification status")
	}

	response := &VerifyEmailResponse{
		Message: "Email verified successfully",
		Success: true,
	}

	return response, nil
}