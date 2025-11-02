package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
)

// VerificationType represents the type of verification
type VerificationType string

const (
	VerificationTypeSelfie   VerificationType = "selfie"
	VerificationTypeDocument VerificationType = "document"
)

// VerificationLevel represents the verification level of a user
type VerificationLevel int

const (
	VerificationLevelNone     VerificationLevel = 0
	VerificationLevelSelfie  VerificationLevel = 1
	VerificationLevelDocument VerificationLevel = 2
)

// Verification represents a verification attempt
type Verification struct {
	ID               uuid.UUID                    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID                    `json:"user_id" gorm:"type:uuid;not null;index"`
	Type             VerificationType             `json:"type" gorm:"not null;check:type IN ('selfie', 'document')"`
	Status           valueobjects.VerificationStatus `json:"status" gorm:"not null;default:'pending';check:status IN ('pending', 'approved', 'rejected')"`
	PhotoURL         string                       `json:"photo_url" gorm:"not null"`
	PhotoKey         string                       `json:"photo_key" gorm:"not null;uniqueIndex"`
	DocumentType     *string                      `json:"document_type"` // For document verification: 'id_card', 'passport', 'driver_license'
	DocumentData     *string                      `json:"document_data"` // JSON string with extracted document data
	AIScore          *float64                     `json:"ai_score"` // AI confidence score (0-1)
	AIDetails        *string                      `json:"ai_details"` // JSON string with AI analysis details
	RejectionReason  *string                      `json:"rejection_reason"`
	ReviewedBy       *uuid.UUID                   `json:"reviewed_by"` // Admin ID if manually reviewed
	ReviewedAt       *time.Time                   `json:"reviewed_at"`
	ExpiresAt        *time.Time                   `json:"expires_at"`
	CreatedAt        time.Time                    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time                    `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User     *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Reviewer *AdminUser `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// TableName returns the table name for Verification entity
func (Verification) TableName() string {
	return "verifications"
}

// IsExpired returns true if verification is expired
func (v *Verification) IsExpired() bool {
	if v.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*v.ExpiresAt)
}

// IsApproved returns true if verification is approved
func (v *Verification) IsApproved() bool {
	return v.Status.IsApproved()
}

// IsRejected returns true if verification is rejected
func (v *Verification) IsRejected() bool {
	return v.Status.IsRejected()
}

// IsPending returns true if verification is pending
func (v *Verification) IsPending() bool {
	return v.Status.IsPending()
}

// CanBeReviewed returns true if verification can be reviewed
func (v *Verification) CanBeReviewed() bool {
	return v.Status.IsPending() && !v.IsExpired()
}

// Approve marks verification as approved
func (v *Verification) Approve(reviewedBy uuid.UUID) {
	v.Status = valueobjects.VerificationStatusApproved
	v.ReviewedBy = &reviewedBy
	now := time.Now()
	v.ReviewedAt = &now
	v.RejectionReason = nil
}

// Reject marks verification as rejected with a reason
func (v *Verification) Reject(reason string, reviewedBy uuid.UUID) {
	v.Status = valueobjects.VerificationStatusRejected
	v.RejectionReason = &reason
	v.ReviewedBy = &reviewedBy
	now := time.Now()
	v.ReviewedAt = &now
}

// SetAIScore sets the AI confidence score and details
func (v *Verification) SetAIScore(score float64, details string) {
	v.AIScore = &score
	v.AIDetails = &details
}

// SetDocumentData sets the extracted document data
func (v *Verification) SetDocumentData(documentType string, data string) {
	v.DocumentType = &documentType
	v.DocumentData = &data
}

// VerificationAttempt represents a verification attempt tracking
type VerificationAttempt struct {
	ID         uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Type       VerificationType  `json:"type" gorm:"not null;check:type IN ('selfie', 'document')"`
	IPAddress  string           `json:"ip_address" gorm:"not null"`
	UserAgent  string           `json:"user_agent" gorm:"not null"`
	Status     string           `json:"status" gorm:"not null"` // 'success', 'failure', 'rate_limited'
	Reason     *string          `json:"reason"` // Failure reason
	CreatedAt  time.Time        `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for VerificationAttempt entity
func (VerificationAttempt) TableName() string {
	return "verification_attempts"
}

// VerificationBadge represents a verification badge awarded to users
type VerificationBadge struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Level       VerificationLevel `json:"level" gorm:"not null;check:level IN (0, 1, 2)"`
	BadgeType   string           `json:"badge_type" gorm:"not null"` // 'selfie_verified', 'document_verified'
	ExpiresAt   *time.Time       `json:"expires_at"`
	IsRevoked   bool             `json:"is_revoked" gorm:"default:false"`
	RevokedAt   *time.Time       `json:"revoked_at"`
	RevokedBy   *uuid.UUID       `json:"revoked_by"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User     *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Revoker  *AdminUser `json:"revoker,omitempty" gorm:"foreignKey:RevokedBy"`
}

// TableName returns the table name for VerificationBadge entity
func (VerificationBadge) TableName() string {
	return "verification_badges"
}

// IsExpired returns true if badge is expired
func (vb *VerificationBadge) IsExpired() bool {
	if vb.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*vb.ExpiresAt)
}

// IsActive returns true if badge is active (not expired and not revoked)
func (vb *VerificationBadge) IsActive() bool {
	return !vb.IsRevoked && !vb.IsExpired()
}

// Revoke revokes the verification badge
func (vb *VerificationBadge) Revoke(revokedBy uuid.UUID) {
	vb.IsRevoked = true
	vb.RevokedBy = &revokedBy
	now := time.Now()
	vb.RevokedAt = &now
}

// AdminUser represents an admin user for verification review
type AdminUser struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	FirstName    string    `json:"first_name" gorm:"not null"`
	LastName     string    `json:"last_name" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'admin';check:role IN ('admin', 'moderator', 'super_admin')"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for AdminUser entity
func (AdminUser) TableName() string {
	return "admin_users"
}

// IsModerator returns true if admin is moderator or higher
func (au *AdminUser) IsModerator() bool {
	return au.Role == "moderator" || au.Role == "super_admin"
}

// IsSuperAdmin returns true if admin is super admin
func (au *AdminUser) IsSuperAdmin() bool {
	return au.Role == "super_admin"
}

// CanReviewVerifications returns true if admin can review verifications
func (au *AdminUser) CanReviewVerifications() bool {
	return au.IsActive && (au.IsModerator() || au.IsSuperAdmin())
}

// GetAllVerificationTypes returns all valid verification types
func GetAllVerificationTypes() []VerificationType {
	return []VerificationType{
		VerificationTypeSelfie,
		VerificationTypeDocument,
	}
}

// GetAllVerificationLevels returns all valid verification levels
func GetAllVerificationLevels() []VerificationLevel {
	return []VerificationLevel{
		VerificationLevelNone,
		VerificationLevelSelfie,
		VerificationLevelDocument,
	}
}

// GetVerificationLevelDisplayName returns the display name for verification level
func GetVerificationLevelDisplayName(level VerificationLevel) string {
	switch level {
	case VerificationLevelNone:
		return "Not Verified"
	case VerificationLevelSelfie:
		return "Selfie Verified"
	case VerificationLevelDocument:
		return "Document Verified"
	default:
		return "Unknown"
	}
}

// GetVerificationTypeDisplayName returns the display name for verification type
func GetVerificationTypeDisplayName(vType VerificationType) string {
	switch vType {
	case VerificationTypeSelfie:
		return "Selfie Verification"
	case VerificationTypeDocument:
		return "Document Verification"
	default:
		return "Unknown"
	}
}