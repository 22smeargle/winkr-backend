package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// VerificationService handles email and SMS verification
type VerificationService struct {
	redisClient *redis.RedisClient
	emailService EmailService
	smsService   SMSService
}

// EmailService defines interface for sending emails
type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, code string) error
	SendPasswordResetEmail(ctx context.Context, to, token string) error
}

// SMSService defines interface for sending SMS
type SMSService interface {
	SendVerificationSMS(ctx context.Context, to, code string) error
}

// NewVerificationService creates a new verification service
func NewVerificationService(redisClient *redis.RedisClient, emailService EmailService, smsService SMSService) *VerificationService {
	return &VerificationService{
		redisClient: redisClient,
		emailService: emailService,
		smsService:   smsService,
	}
}

// GenerateVerificationCode generates a 6-digit verification code
func (vs *VerificationService) GenerateVerificationCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += n.String()
	}
	return code
}

// GenerateVerificationToken generates a secure verification token
func (vs *VerificationService) GenerateVerificationToken() string {
	return uuid.New().String()
}

// SendEmailVerification sends an email verification code
func (vs *VerificationService) SendEmailVerification(ctx context.Context, email, ipAddress, userAgent string) error {
	// Log request for security monitoring
	logger.Info("Email verification request", "email", email, "ip", ipAddress, "user_agent", userAgent)

	// Check rate limit
	rateLimitKey := fmt.Sprintf("email_verification_rate:%s", email)
	exists, err := vs.redisClient.Exists(ctx, rateLimitKey)
	if err != nil {
		logger.Error("Failed to check email verification rate limit", err)
		return errors.NewInternalError("Failed to check rate limit")
	}
	if exists {
		logger.Warn("Email verification rate limit exceeded", "email", email, "ip", ipAddress)
		return errors.NewAppError(429, "Too many verification requests", "Please wait before requesting another verification email")
	}

	// Generate verification code
	code := vs.GenerateVerificationCode()
	token := vs.GenerateVerificationToken()

	// Store verification data in Redis with 15-minute expiry
	verificationKey := fmt.Sprintf("email_verification:%s", token)
	verificationData := map[string]interface{}{
		"email":      email,
		"code":       code,
		"type":        "email",
		"ip_address":  ipAddress,
		"user_agent":   userAgent,
		"created_at":  time.Now().Unix(),
	}
	
	err = vs.redisClient.HMSet(ctx, verificationKey, verificationData)
	if err != nil {
		logger.Error("Failed to store email verification data", err)
		return errors.NewInternalError("Failed to store verification data")
	}

	// Set expiry for verification data
	err = vs.redisClient.Expire(ctx, verificationKey, 15*time.Minute)
	if err != nil {
		logger.Error("Failed to set expiry for email verification", err)
		// Continue anyway as the data is stored
	}

	// Set rate limit with 5-minute cooldown
	err = vs.redisClient.Set(ctx, rateLimitKey, "1", 5*time.Minute)
	if err != nil {
		logger.Error("Failed to set email verification rate limit", err)
		// Continue anyway as the verification is sent
	}

	// Send verification email
	err = vs.emailService.SendVerificationEmail(ctx, email, code)
	if err != nil {
		logger.Error("Failed to send verification email", err)
		return errors.NewExternalServiceError("email", "Failed to send verification email")
	}

	logger.Info("Email verification sent", "email", email, "token", token, "ip", ipAddress)
	return nil
}

// SendSMSVerification sends an SMS verification code
func (vs *VerificationService) SendSMSVerification(ctx context.Context, phoneNumber, ipAddress, userAgent string) error {
	// Log request for security monitoring
	logger.Info("SMS verification request", "phone", phoneNumber, "ip", ipAddress, "user_agent", userAgent)

	// Check rate limit
	rateLimitKey := fmt.Sprintf("sms_verification_rate:%s", phoneNumber)
	exists, err := vs.redisClient.Exists(ctx, rateLimitKey)
	if err != nil {
		logger.Error("Failed to check SMS verification rate limit", err)
		return errors.NewInternalError("Failed to check rate limit")
	}
	if exists {
		logger.Warn("SMS verification rate limit exceeded", "phone", phoneNumber, "ip", ipAddress)
		return errors.NewAppError(429, "Too many verification requests", "Please wait before requesting another verification SMS")
	}

	// Generate verification code
	code := vs.GenerateVerificationCode()
	token := vs.GenerateVerificationToken()

	// Store verification data in Redis with 15-minute expiry
	verificationKey := fmt.Sprintf("sms_verification:%s", token)
	verificationData := map[string]interface{}{
		"phone":      phoneNumber,
		"code":       code,
		"type":        "sms",
		"ip_address":  ipAddress,
		"user_agent":   userAgent,
		"created_at":  time.Now().Unix(),
	}
	
	err = vs.redisClient.HMSet(ctx, verificationKey, verificationData)
	if err != nil {
		logger.Error("Failed to store SMS verification data", err)
		return errors.NewInternalError("Failed to store verification data")
	}

	// Set expiry for verification data
	err = vs.redisClient.Expire(ctx, verificationKey, 15*time.Minute)
	if err != nil {
		logger.Error("Failed to set expiry for SMS verification", err)
		// Continue anyway as the data is stored
	}

	// Set rate limit with 5-minute cooldown
	err = vs.redisClient.Set(ctx, rateLimitKey, "1", 5*time.Minute)
	if err != nil {
		logger.Error("Failed to set SMS verification rate limit", err)
		// Continue anyway as the verification is sent
	}

	// Send verification SMS
	err = vs.smsService.SendVerificationSMS(ctx, phoneNumber, code)
	if err != nil {
		logger.Error("Failed to send verification SMS", err)
		return errors.NewExternalServiceError("sms", "Failed to send verification SMS")
	}

	logger.Info("SMS verification sent", "phone", phoneNumber, "token", token, "ip", ipAddress)
	return nil
}

// VerifyEmailCode verifies an email verification code
func (vs *VerificationService) VerifyEmailCode(ctx context.Context, token, ipAddress, userAgent string) (*VerificationResult, error) {
	verificationKey := fmt.Sprintf("email_verification:%s", token)
	
	// Get verification data
	data, err := vs.redisClient.HGetAll(ctx, verificationKey)
	if err != nil {
		logger.Error("Failed to get email verification data", err)
		return nil, errors.NewInternalError("Failed to verify code")
	}

	if len(data) == 0 {
		return nil, errors.NewAppError(400, "Invalid or expired verification token", "")
	}

	storedCode, ok := data["code"]
	if !ok {
		return nil, errors.NewInternalError("Invalid verification data")
	}

	email, _ := data["email"]

	// Verify code
	if storedCode != code {
		return nil, errors.NewAppError(400, "Invalid verification code", "")
	}

	// Delete verification data after successful verification
	err = vs.redisClient.Del(ctx, verificationKey)
	if err != nil {
		logger.Error("Failed to delete email verification data", err)
		// Continue anyway as verification is successful
	}

	return &VerificationResult{
		Identifier: email,
		Type:       "email",
		Success:    true,
	}, nil
}

// VerifySMSCode verifies an SMS verification code
func (vs *VerificationService) VerifySMSCode(ctx context.Context, token, code string) (*VerificationResult, error) {
	verificationKey := fmt.Sprintf("sms_verification:%s", token)
	
	// Get verification data
	data, err := vs.redisClient.HGetAll(ctx, verificationKey)
	if err != nil {
		logger.Error("Failed to get SMS verification data", err)
		return nil, errors.NewInternalError("Failed to verify code")
	}

	if len(data) == 0 {
		return nil, errors.NewAppError(400, "Invalid or expired verification token", "")
	}

	storedCode, ok := data["code"]
	if !ok {
		return nil, errors.NewInternalError("Invalid verification data")
	}

	phoneNumber, _ := data["phone"]

	// Verify code
	if storedCode != code {
		return nil, errors.NewAppError(400, "Invalid verification code", "")
	}

	// Delete verification data after successful verification
	err = vs.redisClient.Del(ctx, verificationKey)
	if err != nil {
		logger.Error("Failed to delete SMS verification data", err)
		// Continue anyway as verification is successful
	}

	return &VerificationResult{
		Identifier: phoneNumber,
		Type:       "sms",
		Success:    true,
	}, nil
}

// SendPasswordReset sends a password reset email
func (vs *VerificationService) SendPasswordReset(ctx context.Context, email, ipAddress, userAgent string) (string, error) {
	// Log request for security monitoring
	logger.Info("Password reset request", "email", email, "ip", ipAddress, "user_agent", userAgent)

	// Check rate limit
	rateLimitKey := fmt.Sprintf("password_reset_rate:%s", email)
	exists, err := vs.redisClient.Exists(ctx, rateLimitKey)
	if err != nil {
		logger.Error("Failed to check password reset rate limit", err)
		return "", errors.NewInternalError("Failed to check rate limit")
	}
	if exists {
		logger.Warn("Password reset rate limit exceeded", "email", email, "ip", ipAddress)
		return "", errors.NewAppError(429, "Too many password reset requests", "Please wait before requesting another password reset")
	}

	// Generate reset token
	token := vs.GenerateVerificationToken()

	// Store reset token in Redis with 1-hour expiry
	resetKey := fmt.Sprintf("password_reset:%s", token)
	resetData := map[string]interface{}{
		"email":      email,
		"type":        "password_reset",
		"ip_address":  ipAddress,
		"user_agent":   userAgent,
		"created_at":  time.Now().Unix(),
	}
	
	err = vs.redisClient.HMSet(ctx, resetKey, resetData)
	if err != nil {
		logger.Error("Failed to store password reset data", err)
		return "", errors.NewInternalError("Failed to store reset data")
	}

	// Set expiry for reset token
	err = vs.redisClient.Expire(ctx, resetKey, time.Hour)
	if err != nil {
		logger.Error("Failed to set expiry for password reset", err)
		// Continue anyway as the data is stored
	}

	// Set rate limit with 15-minute cooldown
	err = vs.redisClient.Set(ctx, rateLimitKey, "1", 15*time.Minute)
	if err != nil {
		logger.Error("Failed to set password reset rate limit", err)
		// Continue anyway as the reset is sent
	}

	// Send password reset email
	err = vs.emailService.SendPasswordResetEmail(ctx, email, token)
	if err != nil {
		logger.Error("Failed to send password reset email", err)
		return "", errors.NewExternalServiceError("email", "Failed to send password reset email")
	}

	logger.Info("Password reset sent", "email", email, "token", token, "ip", ipAddress)
	return token, nil
}

// VerifyPasswordReset verifies a password reset token
func (vs *VerificationService) VerifyPasswordReset(ctx context.Context, token, ipAddress, userAgent string) (*VerificationResult, error) {
	// Log verification attempt for security monitoring
	logger.Info("Password reset verification attempt", "token", token, "ip", ipAddress, "user_agent", userAgent)

	resetKey := fmt.Sprintf("password_reset:%s", token)
	
	// Get reset data
	data, err := vs.redisClient.HGetAll(ctx, resetKey)
	if err != nil {
		logger.Error("Failed to get password reset data", err)
		return nil, errors.NewInternalError("Failed to verify reset token")
	}

	if len(data) == 0 {
		logger.Warn("Password reset token not found or expired", "token", token, "ip", ipAddress)
		return nil, errors.NewAppError(400, "Invalid or expired reset token", "")
	}

	email, _ := data["email"]

	logger.Info("Password reset verification successful", "email", email, "ip", ipAddress)
	return &VerificationResult{
		Identifier: email,
		Type:       "password_reset",
		Success:    true,
	}, nil
}

// InvalidatePasswordReset invalidates a password reset token
func (vs *VerificationService) InvalidatePasswordReset(ctx context.Context, token string) error {
	resetKey := fmt.Sprintf("password_reset:%s", token)
	
	err := vs.redisClient.Del(ctx, resetKey)
	if err != nil {
		logger.Error("Failed to invalidate password reset token", err)
		return errors.NewInternalError("Failed to invalidate reset token")
	}

	return nil
}

// VerificationResult represents the result of a verification
type VerificationResult struct {
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Success    bool   `json:"success"`
}