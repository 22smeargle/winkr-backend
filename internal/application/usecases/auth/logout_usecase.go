package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// LogoutUseCase handles user logout
type LogoutUseCase struct {
	authService services.AuthService
	jwtUtils   *utils.JWTUtils
}

// NewLogoutUseCase creates a new LogoutUseCase instance
func NewLogoutUseCase(authService services.AuthService, jwtUtils *utils.JWTUtils) *LogoutUseCase {
	return &LogoutUseCase{
		authService: authService,
		jwtUtils:   jwtUtils,
	}
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	LogoutAll bool      `json:"logout_all,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// LogoutResponse represents logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// Execute handles the logout use case
func (uc *LogoutUseCase) Execute(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	// Validate token and extract session ID
	sessionID := ""
	if req.Token != "" {
		claims, err := uc.jwtUtils.ValidateToken(req.Token)
		if err != nil {
			return nil, errors.ErrInvalidToken
		}
		sessionID = claims.SessionID
	}

	// Perform logout
	if req.LogoutAll {
		// Logout from all devices
		err := uc.authService.LogoutFromAllDevices(ctx, req.UserID, req.IPAddress, req.UserAgent)
		if err != nil {
			return nil, err
		}
	} else {
		// Logout from current session
		err := uc.authService.Logout(ctx, req.UserID, sessionID, req.IPAddress, req.UserAgent)
		if err != nil {
			return nil, err
		}
	}

	// Blacklist the current token if provided
	if req.Token != "" {
		err := uc.jwtUtils.BlacklistToken(ctx, req.Token)
		if err != nil {
			// Log error but don't fail logout
			// logger.Error("Failed to blacklist token during logout", err)
		}
	}

	// Return response
	response := &LogoutResponse{
		Message: "Successfully logged out",
	}

	return response, nil
}