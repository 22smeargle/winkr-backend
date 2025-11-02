package auth

import (
	"context"

	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	authService services.AuthService
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase instance
func NewRefreshTokenUseCase(authService services.AuthService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		authService: authService,
	}
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Execute handles the token refresh use case
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Convert to service request
	serviceResp, err := uc.authService.RefreshToken(ctx, req.RefreshToken, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	// Convert response
	response := &RefreshTokenResponse{
		AccessToken:  serviceResp.AccessToken,
		RefreshToken: serviceResp.RefreshToken,
		ExpiresIn:    serviceResp.ExpiresIn,
	}

	return response, nil
}