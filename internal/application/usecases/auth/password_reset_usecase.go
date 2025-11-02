package auth

import (
	"context"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// PasswordResetUseCase handles password reset
type PasswordResetUseCase struct {
	userRepo           repositories.UserRepository
	verificationService *services.VerificationService
	passwordHash       func(string) (string, error)
}

// NewPasswordResetUseCase creates a new PasswordResetUseCase instance
func NewPasswordResetUseCase(
	userRepo repositories.UserRepository,
	verificationService *services.VerificationService,
) *PasswordResetUseCase {
	return &PasswordResetUseCase{
		userRepo:           userRepo,
		verificationService: verificationService,
		passwordHash:       utils.HashPassword,
	}
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email     string `json:"email" validate:"required,email"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// PasswordResetResponse represents password reset response
type PasswordResetResponse struct {
	Message string `json:"message"`
}

// Execute handles password reset request use case
func (uc *PasswordResetUseCase) Execute(ctx context.Context, req *PasswordResetRequest) (*PasswordResetResponse, error) {
	// Check if user exists
	_, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return &PasswordResetResponse{
			Message: "If an account with that email exists, a password reset link has been sent",
		}, nil
	}

	// Send password reset email
	_, err = uc.verificationService.SendPasswordReset(ctx, req.Email, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	// Return response (always success to prevent email enumeration)
	response := &PasswordResetResponse{
		Message: "If an account with that email exists, a password reset link has been sent",
	}

	return response, nil
}

// ConfirmPasswordResetUseCase handles password reset confirmation
type ConfirmPasswordResetUseCase struct {
	userRepo           repositories.UserRepository
	verificationService *services.VerificationService
	passwordHash       func(string) (string, error)
}

// NewConfirmPasswordResetUseCase creates a new ConfirmPasswordResetUseCase instance
func NewConfirmPasswordResetUseCase(
	userRepo repositories.UserRepository,
	verificationService *services.VerificationService,
) *ConfirmPasswordResetUseCase {
	return &ConfirmPasswordResetUseCase{
		userRepo:           userRepo,
		verificationService: verificationService,
		passwordHash:       utils.HashPassword,
	}
}

// ConfirmPasswordResetRequest represents password reset confirmation request
type ConfirmPasswordResetRequest struct {
	Token     string `json:"token" validate:"required"`
	Password  string `json:"password" validate:"required,password"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ConfirmPasswordResetResponse represents password reset confirmation response
type ConfirmPasswordResetResponse struct {
	Message string `json:"message"`
}

// Execute handles password reset confirmation use case
func (uc *ConfirmPasswordResetUseCase) Execute(ctx context.Context, req *ConfirmPasswordResetRequest) (*ConfirmPasswordResetResponse, error) {
	// Verify reset token
	result, err := uc.verificationService.VerifyPasswordReset(ctx, req.Token, req.IPAddress, req.UserAgent)
	if err != nil {
		return &ConfirmPasswordResetResponse{
			Message: "Invalid or expired reset token",
		}, nil // Don't expose specific error
	}

	if !result.Success {
		return &ConfirmPasswordResetResponse{
			Message: "Password reset failed",
		}, nil
	}

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, result.Identifier)
	if err != nil {
		return &ConfirmPasswordResetResponse{
			Message: "User not found",
		}, nil // Don't expose specific error
	}

	// Hash new password
	hashedPassword, err := uc.passwordHash(req.Password)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to hash password")
	}

	// Update user password
	user.PasswordHash = hashedPassword
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to update password")
	}

	// Invalidate reset token
	err = uc.verificationService.InvalidatePasswordReset(ctx, req.Token)
	if err != nil {
		// Log error but don't fail password reset
		// logger.Error("Failed to invalidate reset token", err)
	}

	// Return response
	response := &ConfirmPasswordResetResponse{
		Message: "Password has been reset successfully",
	}

	return response, nil
}