package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/services"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// GetSessionsUseCase handles getting user sessions
type GetSessionsUseCase struct {
	authService services.AuthService
}

// NewGetSessionsUseCase creates a new GetSessionsUseCase instance
func NewGetSessionsUseCase(authService services.AuthService) *GetSessionsUseCase {
	return &GetSessionsUseCase{
		authService: authService,
	}
}

// GetSessionsRequest represents get sessions request
type GetSessionsRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// GetSessionsResponse represents get sessions response
type GetSessionsResponse struct {
	Success  bool         `json:"success"`
	Data     []*SessionDTO `json:"data"`
	Message  string       `json:"message,omitempty"`
}

// SessionDTO represents a user session
type SessionDTO struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceType   string    `json:"device_type"`
	Platform     string    `json:"platform"`
	Browser      string    `json:"browser"`
	IsCurrent    bool      `json:"is_current"`
	LastActive   string    `json:"last_active"`
	CreatedAt    string    `json:"created_at"`
	ExpiresAt    string    `json:"expires_at"`
}

// Execute handles the get sessions use case
func (uc *GetSessionsUseCase) Execute(ctx context.Context, req *GetSessionsRequest) (*GetSessionsResponse, error) {
	// TODO: Get sessions from auth service
	// sessions, err := uc.authService.GetActiveSessions(ctx, req.UserID, req.IPAddress, req.UserAgent)
	// if err != nil {
	//     return nil, err
	// }

	// Convert to DTOs
	sessionDTOs := make([]*SessionDTO, 0) // TODO: Convert sessions to DTOs

	// Return response
	response := &GetSessionsResponse{
		Success: true,
		Data:    sessionDTOs,
	}

	return response, nil
}