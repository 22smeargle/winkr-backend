package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
)

// VerificationRepository defines the interface for verification data operations
type VerificationRepository interface {
	// Verification operations
	CreateVerification(ctx context.Context, verification *entities.Verification) error
	GetVerificationByID(ctx context.Context, id uuid.UUID) (*entities.Verification, error)
	GetVerificationByUserAndType(ctx context.Context, userID uuid.UUID, vType entities.VerificationType) (*entities.Verification, error)
	GetPendingVerifications(ctx context.Context, limit, offset int) ([]*entities.Verification, error)
	GetVerificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Verification, error)
	UpdateVerification(ctx context.Context, verification *entities.Verification) error
	DeleteVerification(ctx context.Context, id uuid.UUID) error
	GetVerificationsForReview(ctx context.Context, status valueobjects.VerificationStatus, limit, offset int) ([]*entities.Verification, error)
	GetVerificationStats(ctx context.Context) (*VerificationStats, error)

	// Verification attempt operations
	CreateVerificationAttempt(ctx context.Context, attempt *entities.VerificationAttempt) error
	GetVerificationAttemptsByUser(ctx context.Context, userID uuid.UUID, vType entities.VerificationType, since time.Time) ([]*entities.VerificationAttempt, error)
	GetVerificationAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) ([]*entities.VerificationAttempt, error)
	DeleteExpiredAttempts(ctx context.Context, olderThan time.Time) error

	// Verification badge operations
	CreateVerificationBadge(ctx context.Context, badge *entities.VerificationBadge) error
	GetActiveBadgesByUser(ctx context.Context, userID uuid.UUID) ([]*entities.VerificationBadge, error)
	GetBadgeByUserAndType(ctx context.Context, userID uuid.UUID, badgeType string) (*entities.VerificationBadge, error)
	UpdateBadge(ctx context.Context, badge *entities.VerificationBadge) error
	RevokeBadge(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error
	DeleteExpiredBadges(ctx context.Context) error

	// User verification level operations
	GetUserVerificationLevel(ctx context.Context, userID uuid.UUID) (entities.VerificationLevel, error)
	UpdateUserVerificationLevel(ctx context.Context, userID uuid.UUID, level entities.VerificationLevel) error
	GetUsersByVerificationLevel(ctx context.Context, level entities.VerificationLevel, limit, offset int) ([]*uuid.UUID, error)

	// Admin operations
	GetAdminUserByEmail(ctx context.Context, email string) (*entities.AdminUser, error)
	GetAdminUserByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error)
	CreateAdminUser(ctx context.Context, admin *entities.AdminUser) error
	UpdateAdminUser(ctx context.Context, admin *entities.AdminUser) error
	UpdateAdminLastLogin(ctx context.Context, id uuid.UUID) error
	GetActiveAdminUsers(ctx context.Context) ([]*entities.AdminUser, error)
}

// VerificationStats represents verification statistics
type VerificationStats struct {
	TotalVerifications     int64 `json:"total_verifications"`
	PendingVerifications  int64 `json:"pending_verifications"`
	ApprovedVerifications int64 `json:"approved_verifications"`
	RejectedVerifications int64 `json:"rejected_verifications"`
	SelfieVerifications  int64 `json:"selfie_verifications"`
	DocumentVerifications int64 `json:"document_verifications"`
	VerifiedUsers        int64 `json:"verified_users"`
	VerificationRate     float64 `json:"verification_rate"` // Percentage of approved verifications
}

// VerificationFilter represents filters for verification queries
type VerificationFilter struct {
	UserID     *uuid.UUID                    `json:"user_id,omitempty"`
	Type       *entities.VerificationType      `json:"type,omitempty"`
	Status     *valueobjects.VerificationStatus `json:"status,omitempty"`
	ReviewedBy *uuid.UUID                    `json:"reviewed_by,omitempty"`
	DateFrom   *time.Time                    `json:"date_from,omitempty"`
	DateTo     *time.Time                    `json:"date_to,omitempty"`
	HasAIScore *bool                        `json:"has_ai_score,omitempty"`
	MinAIScore *float64                     `json:"min_ai_score,omitempty"`
	MaxAIScore *float64                     `json:"max_ai_score,omitempty"`
}

// VerificationAttemptFilter represents filters for verification attempt queries
type VerificationAttemptFilter struct {
	UserID    *uuid.UUID              `json:"user_id,omitempty"`
	Type      *entities.VerificationType `json:"type,omitempty"`
	Status    *string                  `json:"status,omitempty"`
	IPAddress *string                  `json:"ip_address,omitempty"`
	DateFrom  *time.Time               `json:"date_from,omitempty"`
	DateTo    *time.Time               `json:"date_to,omitempty"`
}

// VerificationBadgeFilter represents filters for verification badge queries
type VerificationBadgeFilter struct {
	UserID    *uuid.UUID              `json:"user_id,omitempty"`
	Level     *entities.VerificationLevel `json:"level,omitempty"`
	BadgeType *string                  `json:"badge_type,omitempty"`
	IsActive  *bool                    `json:"is_active,omitempty"`
	DateFrom  *time.Time               `json:"date_from,omitempty"`
	DateTo    *time.Time               `json:"date_to,omitempty"`
}

// AdminUserFilter represents filters for admin user queries
type AdminUserFilter struct {
	Email     *string    `json:"email,omitempty"`
	Role      *string    `json:"role,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	DateFrom  *time.Time `json:"date_from,omitempty"`
	DateTo    *time.Time `json:"date_to,omitempty"`
}