package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// LoginUseCase handles user login
type LoginUseCase struct {
	authService services.AuthService
	jwtUtils   *utils.JWTUtils
}

// NewLoginUseCase creates a new LoginUseCase instance
func NewLoginUseCase(authService services.AuthService, jwtUtils *utils.JWTUtils) *LoginUseCase {
	return &LoginUseCase{
		authService: authService,
		jwtUtils:   jwtUtils,
	}
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email       string            `json:"email" validate:"required,email"`
	Password    string            `json:"password" validate:"required"`
	DeviceInfo  *utils.DeviceInfo `json:"device_info,omitempty"`
	IPAddress   string            `json:"ip_address,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User   *UserInfo   `json:"user"`
	Tokens *TokenPair   `json:"tokens"`
}

// UserInfo represents user information in response
type UserInfo struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	IsVerified    bool      `json:"is_verified"`
	IsPremium     bool      `json:"is_premium"`
	CreatedAt     string    `json:"created_at"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Execute handles the user login use case
func (uc *LoginUseCase) Execute(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Parse device info if not provided
	var deviceInfo *utils.DeviceInfo
	if req.DeviceInfo == nil && req.UserAgent != "" {
		deviceInfo = uc.jwtUtils.ParseDeviceInfo(req.UserAgent, req.IPAddress)
	} else if req.DeviceInfo != nil {
		deviceInfo = req.DeviceInfo
	} else {
		// Create minimal device info
		deviceInfo = &utils.DeviceInfo{
			UserAgent:   req.UserAgent,
			IPAddress:   req.IPAddress,
			Platform:    "unknown",
			Device:      "unknown",
			Browser:     "unknown",
		}
		deviceInfo.Fingerprint = uc.jwtUtils.GenerateDeviceFingerprint(deviceInfo)
	}

	// Convert to service request
	serviceReq := &services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Call auth service
	authResp, err := uc.authService.Login(ctx, serviceReq, deviceInfo, req.IPAddress)
	if err != nil {
		return nil, err
	}

	// Convert response
	response := &LoginResponse{
		User: &UserInfo{
			ID:        authResp.User.ID,
			Email:     authResp.User.Email,
			FirstName:  authResp.User.FirstName,
			LastName:   authResp.User.LastName,
			IsVerified: authResp.User.IsVerified,
			IsPremium:  authResp.User.IsPremium,
			CreatedAt:  authResp.User.CreatedAt,
		},
		Tokens: &TokenPair{
			AccessToken:  authResp.Tokens.AccessToken,
			RefreshToken: authResp.Tokens.RefreshToken,
			ExpiresIn:    authResp.Tokens.ExpiresIn,
		},
	}

	return response, nil
}