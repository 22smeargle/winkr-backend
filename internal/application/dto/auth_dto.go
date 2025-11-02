package dto

import (
	"github.com/go-playground/validator/v10"
)

// AuthDTOs contain all authentication related data transfer objects

// RegisterRequestDTO represents user registration request DTO
type RegisterRequestDTO struct {
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required,password"`
	FirstName    string   `json:"first_name" validate:"required,min=2,max=100"`
	LastName     string   `json:"last_name" validate:"required,min=2,max=100"`
	DateOfBirth  string   `json:"date_of_birth" validate:"required"`
	Gender       string   `json:"gender" validate:"required,oneof=male female other"`
	InterestedIn []string `json:"interested_in" validate:"required,min=1,dive,oneof=male female other"`
}

// LoginRequestDTO represents user login request DTO
type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequestDTO represents token refresh request DTO
type RefreshTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequestDTO represents password change request DTO
type ChangePasswordRequestDTO struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,password"`
}

// ResetPasswordRequestDTO represents password reset request DTO
type ResetPasswordRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmPasswordResetRequestDTO represents password reset confirmation DTO
type ConfirmPasswordResetRequestDTO struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

// AuthResponseDTO represents authentication response DTO
type AuthResponseDTO struct {
	Success bool          `json:"success"`
	Data    *AuthDataDTO  `json:"data,omitempty"`
	Message string         `json:"message,omitempty"`
	Error   *ErrorDTO     `json:"error,omitempty"`
}

// AuthDataDTO represents authentication data DTO
type AuthDataDTO struct {
	User   *UserDTO    `json:"user"`
	Tokens *TokensDTO  `json:"tokens"`
}

// TokensDTO represents token pair DTO
type TokensDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenResponseDTO represents token response DTO
type TokenResponseDTO struct {
	Success bool           `json:"success"`
	Data    *TokenDataDTO  `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
	Error   *ErrorDTO      `json:"error,omitempty"`
}

// TokenDataDTO represents token data DTO
type TokenDataDTO struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// LogoutResponseDTO represents logout response DTO
type LogoutResponseDTO struct {
	Success bool     `json:"success"`
	Message string    `json:"message"`
	Error   *ErrorDTO `json:"error,omitempty"`
}

// UserDTO represents user DTO in auth responses
type UserDTO struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	IsVerified   bool   `json:"is_verified"`
	IsPremium    bool   `json:"is_premium"`
	CreatedAt    string `json:"created_at"`
}

// ErrorDTO represents error DTO
type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Validate validates the DTO using the validator
func (dto *RegisterRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *LoginRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *RefreshTokenRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *ChangePasswordRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *ResetPasswordRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *ConfirmPasswordResetRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

// NewAuthResponseDTO creates a new auth response DTO
func NewAuthResponseDTO(user *UserDTO, tokens *TokensDTO) *AuthResponseDTO {
	return &AuthResponseDTO{
		Success: true,
		Data: &AuthDataDTO{
			User:   user,
			Tokens: tokens,
		},
	}
}

// NewTokenResponseDTO creates a new token response DTO
func NewTokenResponseDTO(accessToken string, expiresIn int64) *TokenResponseDTO {
	return &TokenResponseDTO{
		Success: true,
		Data: &TokenDataDTO{
			AccessToken: accessToken,
			ExpiresIn:   expiresIn,
		},
	}
}

// NewLogoutResponseDTO creates a new logout response DTO
func NewLogoutResponseDTO(message string) *LogoutResponseDTO {
	return &LogoutResponseDTO{
		Success: true,
		Message: message,
	}
}

// NewErrorResponseDTO creates a new error response DTO
func NewErrorResponseDTO(code, message, details string) *AuthResponseDTO {
	return &AuthResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// LogoutRequestDTO represents logout request DTO
type LogoutRequestDTO struct {
	LogoutAll bool `json:"logout_all,omitempty"`
}

// MessageResponseDTO represents a simple message response DTO
type MessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   *ErrorDTO `json:"error,omitempty"`
}

// VerifyEmailRequestDTO represents email verification request DTO
type VerifyEmailRequestDTO struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmailResponseDTO represents email verification response DTO
type VerifyEmailResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   *ErrorDTO `json:"error,omitempty"`
}

// SessionDTO represents session DTO
type SessionDTO struct {
	ID           string `json:"id"`
	DeviceInfo   *DeviceInfoDTO `json:"device_info"`
	IPAddress    string `json:"ip_address"`
	LastActivity string `json:"last_activity"`
	CreatedAt    string `json:"created_at"`
	ExpiresAt    string `json:"expires_at"`
	IsActive     bool   `json:"is_active"`
}

// DeviceInfoDTO represents device information DTO
type DeviceInfoDTO struct {
	Fingerprint string `json:"fingerprint"`
	Platform    string `json:"platform"`
	Device      string `json:"device"`
	Browser     string `json:"browser"`
}

// SessionsResponseDTO represents sessions response DTO
type SessionsResponseDTO struct {
	Success bool          `json:"success"`
	Data    []*SessionDTO `json:"data,omitempty"`
	Error   *ErrorDTO     `json:"error,omitempty"`
}

// UserProfileResponseDTO represents user profile response DTO
type UserProfileResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// Validate validates the DTO using the validator
func (dto *LogoutRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *VerifyEmailRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

// NewMessageResponseDTO creates a new message response DTO
func NewMessageResponseDTO(message string) *MessageResponseDTO {
	return &MessageResponseDTO{
		Success: true,
		Message: message,
	}
}

// NewVerifyEmailResponseDTO creates a new verify email response DTO
func NewVerifyEmailResponseDTO(message string, success bool) *VerifyEmailResponseDTO {
	return &VerifyEmailResponseDTO{
		Success: success,
		Message: message,
	}
}

// NewSessionsResponseDTO creates a new sessions response DTO
func NewSessionsResponseDTO(sessions []*SessionDTO) *SessionsResponseDTO {
	return &SessionsResponseDTO{
		Success: true,
		Data:    sessions,
	}
}

// NewUserProfileResponseDTO creates a new user profile response DTO
func NewUserProfileResponseDTO(data interface{}) *UserProfileResponseDTO {
	return &UserProfileResponseDTO{
		Success: true,
		Data:    data,
	}
}

// NewTokenErrorResponseDTO creates a new token error response DTO
func NewTokenErrorResponseDTO(code, message, details string) *TokenResponseDTO {
	return &TokenResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewLogoutErrorResponseDTO creates a new logout error response DTO
func NewLogoutErrorResponseDTO(code, message, details string) *LogoutResponseDTO {
	return &LogoutResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}