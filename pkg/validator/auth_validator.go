package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// AuthValidator handles authentication-specific validation
type AuthValidator struct {
	validator *Validator
}

// NewAuthValidator creates a new auth validator
func NewAuthValidator() *AuthValidator {
	return &AuthValidator{
		validator: &Validator{},
	}
}

// RegistrationRequest represents registration validation request
type RegistrationRequest struct {
	Email        string   `validate:"required,email"`
	Password     string   `validate:"required,password"`
	FirstName    string   `validate:"required,min=2,max=100"`
	LastName     string   `validate:"required,min=2,max=100"`
	DateOfBirth  string   `validate:"required"`
	Gender       string   `validate:"required,oneof=male female other"`
	InterestedIn []string `validate:"required,min=1,dive,oneof=male female other"`
}

// LoginRequest represents login validation request
type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

// RefreshTokenRequest represents refresh token validation request
type RefreshTokenRequest struct {
	RefreshToken string `validate:"required,min=10"`
}

// LogoutRequest represents logout validation request
type LogoutRequest struct {
	Token     string `validate:"required,min=10"`
	LogoutAll bool   `validate:"omitempty"`
}

// PasswordResetRequest represents password reset validation request
type PasswordResetRequest struct {
	Email string `validate:"required,email"`
}

// ConfirmPasswordResetRequest represents password reset confirmation validation request
type ConfirmPasswordResetRequest struct {
	Token    string `validate:"required,min=10"`
	Password string `validate:"required,password"`
}

// VerifyEmailRequest represents email verification validation request
type VerifyEmailRequest struct {
	Token string `validate:"required,min=10"`
}

// ValidateRegistrationRequest validates registration request
func (av *AuthValidator) ValidateRegistrationRequest(req *RegistrationRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Email format validation with MX record check
	if err := av.validateEmailWithMX(req.Email); err != nil {
		return err
	}

	// Password strength validation
	if err := av.validatePasswordStrength(req.Password); err != nil {
		return err
	}

	// Date of birth validation (age restrictions)
	if err := av.validateDateOfBirth(req.DateOfBirth); err != nil {
		return err
	}

	// Username uniqueness validation (using email as username for now)
	if err := av.validateEmailUniqueness(req.Email); err != nil {
		return err
	}

	return nil
}

// ValidateLoginRequest validates login request
func (av *AuthValidator) ValidateLoginRequest(req *LoginRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Email format validation
	if err := av.validateEmailFormat(req.Email); err != nil {
		return err
	}

	return nil
}

// ValidateRefreshTokenRequest validates refresh token request
func (av *AuthValidator) ValidateRefreshTokenRequest(req *RefreshTokenRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Token format validation
	if err := av.validateTokenFormat(req.RefreshToken); err != nil {
		return err
	}

	return nil
}

// ValidateLogoutRequest validates logout request
func (av *AuthValidator) ValidateLogoutRequest(req *LogoutRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Token format validation
	if err := av.validateTokenFormat(req.Token); err != nil {
		return err
	}

	return nil
}

// ValidatePasswordResetRequest validates password reset request
func (av *AuthValidator) ValidatePasswordResetRequest(req *PasswordResetRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Email format validation
	if err := av.validateEmailFormat(req.Email); err != nil {
		return err
	}

	return nil
}

// ValidateConfirmPasswordResetRequest validates password reset confirmation request
func (av *AuthValidator) ValidateConfirmPasswordResetRequest(req *ConfirmPasswordResetRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Token format validation
	if err := av.validateTokenFormat(req.Token); err != nil {
		return err
	}

	// Password strength validation
	if err := av.validatePasswordStrength(req.Password); err != nil {
		return err
	}

	return nil
}

// ValidateVerifyEmailRequest validates email verification request
func (av *AuthValidator) ValidateVerifyEmailRequest(req *VerifyEmailRequest) error {
	// Basic validation
	if err := av.validator.Validate(req); err != nil {
		return err
	}

	// Token format validation
	if err := av.validateTokenFormat(req.Token); err != nil {
		return err
	}

	return nil
}

// validateEmailWithMX validates email format and checks MX records
func (av *AuthValidator) validateEmailWithMX(email string) error {
	// Basic format validation
	if err := av.validateEmailFormat(email); err != nil {
		return err
	}

	// Extract domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.NewValidationError("email", "Invalid email format")
	}
	domain := parts[1]

	// Check MX records
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		// Log error but don't fail validation as MX lookup might fail
		// due to network issues
		return nil
	}

	if len(mxRecords) == 0 {
		return errors.NewValidationError("email", "Email domain does not have valid MX records")
	}

	return nil
}

// validateEmailFormat validates email format
func (av *AuthValidator) validateEmailFormat(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.NewValidationError("email", "Invalid email format")
	}

	// Check for consecutive dots
	if strings.Contains(email, "..") {
		return errors.NewValidationError("email", "Email cannot contain consecutive dots")
	}

	// Check for leading/trailing dots
	if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return errors.NewValidationError("email", "Email cannot start or end with a dot")
	}

	return nil
}

// validatePasswordStrength validates password strength
func (av *AuthValidator) validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.NewValidationError("password", "Password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.NewValidationError("password", "Password must be less than 128 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.NewValidationError("password", "Password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.NewValidationError("password", "Password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return errors.NewValidationError("password", "Password must contain at least one number")
	}

	if !hasSpecial {
		return errors.NewValidationError("password", "Password must contain at least one special character")
	}

	// Check for common passwords
	if av.isCommonPassword(password) {
		return errors.NewValidationError("password", "Password is too common, please choose a more secure password")
	}

	return nil
}

// validateDateOfBirth validates date of birth and age restrictions
func (av *AuthValidator) validateDateOfBirth(dob string) error {
	// Parse date of birth
	date, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return errors.NewValidationError("date_of_birth", "Invalid date format, please use YYYY-MM-DD")
	}

	// Calculate age
	now := time.Now()
	age := now.Year() - date.Year()
	
	// Adjust age if birthday hasn't occurred this year yet
	if now.Month() < date.Month() || (now.Month() == date.Month() && now.Day() < date.Day()) {
		age--
	}

	// Check age restrictions (18-100)
	if age < 18 {
		return errors.NewValidationError("date_of_birth", "You must be at least 18 years old to register")
	}

	if age > 100 {
		return errors.NewValidationError("date_of_birth", "Please enter a valid date of birth")
	}

	return nil
}

// validateEmailUniqueness validates email uniqueness (placeholder)
func (av *AuthValidator) validateEmailUniqueness(email string) error {
	// TODO: Implement actual email uniqueness check against database
	// For now, just validate format
	return nil
}

// validateTokenFormat validates token format
func (av *AuthValidator) validateTokenFormat(token string) error {
	if len(token) < 10 {
		return errors.NewValidationError("token", "Invalid token format")
	}

	// Basic UUID format check
	uuidRegex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	if !uuidRegex.MatchString(token) {
		return errors.NewValidationError("token", "Invalid token format")
	}

	return nil
}

// isCommonPassword checks if password is in common passwords list
func (av *AuthValidator) isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "123456789", "password1",
		"abc123", "111111", "123123", "dragon", "master",
		"hello", "freedom", "whatever", "qazwsx", "trustno1",
		"123qwe", "1q2w3e4r", "zxcvbnm", "123abc", "password!",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	return false
}