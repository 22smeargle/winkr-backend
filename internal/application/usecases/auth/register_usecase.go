package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// RegisterUseCase handles user registration
type RegisterUseCase struct {
	authService services.AuthService
}

// NewRegisterUseCase creates a new RegisterUseCase instance
func NewRegisterUseCase(authService services.AuthService) *RegisterUseCase {
	return &RegisterUseCase{
		authService: authService,
	}
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required,password"`
	FirstName    string   `json:"first_name" validate:"required,min=2,max=100"`
	LastName     string   `json:"last_name" validate:"required,min=2,max=100"`
	DateOfBirth  string   `json:"date_of_birth" validate:"required"`
	Gender       string   `json:"gender" validate:"required,oneof=male female other"`
	InterestedIn []string `json:"interested_in" validate:"required,min=1,dive,oneof=male female other"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
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

// Execute handles the user registration use case
func (uc *RegisterUseCase) Execute(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Convert to service request
	serviceReq := &services.RegisterRequest{
		Email:        req.Email,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		InterestedIn: req.InterestedIn,
	}

	// Call auth service
	authResp, err := uc.authService.Register(ctx, serviceReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	response := &RegisterResponse{
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