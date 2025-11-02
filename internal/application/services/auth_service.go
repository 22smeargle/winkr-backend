package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// AuthService handles authentication operations
type AuthService interface {
	VerifyPassword(email, password string) error
	InvalidateAllUserSessions(ctx context.Context, userID uuid.UUID) error
	InvalidateUserSession(ctx context.Context, userID, sessionID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	IsSessionValid(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	DeviceInfo   *DeviceInfo `json:"device_info"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	LastActivity string     `json:"last_activity"`
	CreatedAt    string     `json:"created_at"`
	ExpiresAt    string     `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	Fingerprint string `json:"fingerprint"`
	Platform    string `json:"platform"`
	Device      string `json:"device"`
	Browser     string `json:"browser"`
}