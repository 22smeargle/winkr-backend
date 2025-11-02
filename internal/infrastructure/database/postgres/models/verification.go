package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Verification represents a verification attempt in database
type Verification struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type             string     `gorm:"not null;check:type IN ('selfie', 'document');index" json:"type"`
	Status           string     `gorm:"not null;default:'pending';check:status IN ('pending', 'approved', 'rejected');index" json:"status"`
	PhotoURL         string     `gorm:"not null" json:"photo_url"`
	PhotoKey         string     `gorm:"not null;uniqueIndex" json:"photo_key"`
	DocumentType     *string    `gorm:"size:50" json:"document_type"` // 'id_card', 'passport', 'driver_license'
	DocumentData     *string    `gorm:"type:text" json:"document_data"` // JSON string
	AIScore          *float64   `gorm:"type:decimal(3,2)" json:"ai_score"` // 0.00-1.00
	AIDetails        *string    `gorm:"type:text" json:"ai_details"` // JSON string
	RejectionReason  *string    `gorm:"type:text" json:"rejection_reason"`
	ReviewedBy       *uuid.UUID `gorm:"type:uuid;index" json:"reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	ExpiresAt        *time.Time `gorm:"index" json:"expires_at"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User     *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Reviewer *AdminUser `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewer,omitempty"`
}

// TableName returns the table name for Verification model
func (Verification) TableName() string {
	return "verifications"
}

// BeforeCreate GORM hook
func (v *Verification) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
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
	return v.Status == "approved"
}

// IsRejected returns true if verification is rejected
func (v *Verification) IsRejected() bool {
	return v.Status == "rejected"
}

// IsPending returns true if verification is pending
func (v *Verification) IsPending() bool {
	return v.Status == "pending"
}

// CanBeReviewed returns true if verification can be reviewed
func (v *Verification) CanBeReviewed() bool {
	return v.Status == "pending" && !v.IsExpired()
}

// Approve marks verification as approved
func (v *Verification) Approve(reviewedBy uuid.UUID) {
	v.Status = "approved"
	v.ReviewedBy = &reviewedBy
	now := time.Now()
	v.ReviewedAt = &now
	v.RejectionReason = nil
}

// Reject marks verification as rejected with a reason
func (v *Verification) Reject(reason string, reviewedBy uuid.UUID) {
	v.Status = "rejected"
	v.RejectionReason = &reason
	v.ReviewedBy = &reviewedBy
	now := time.Now()
	v.ReviewedAt = &now
}

// SetAIScore sets AI confidence score and details
func (v *Verification) SetAIScore(score float64, details string) {
	v.AIScore = &score
	v.AIDetails = &details
}

// SetDocumentData sets extracted document data
func (v *Verification) SetDocumentData(documentType string, data string) {
	v.DocumentType = &documentType
	v.DocumentData = &data
}

// VerificationAttempt represents a verification attempt tracking in database
type VerificationAttempt struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string    `gorm:"not null;check:type IN ('selfie', 'document');index" json:"type"`
	IPAddress string    `gorm:"not null;index" json:"ip_address"`
	UserAgent string    `gorm:"not null" json:"user_agent"`
	Status    string    `gorm:"not null;index" json:"status"` // 'success', 'failure', 'rate_limited'
	Reason    *string   `gorm:"type:text" json:"reason"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for VerificationAttempt model
func (VerificationAttempt) TableName() string {
	return "verification_attempts"
}

// BeforeCreate GORM hook
func (va *VerificationAttempt) BeforeCreate(tx *gorm.DB) error {
	if va.ID == uuid.Nil {
		va.ID = uuid.New()
	}
	return nil
}

// VerificationBadge represents a verification badge in database
type VerificationBadge struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Level     int       `gorm:"not null;check:level IN (0, 1, 2);index" json:"level"`
	BadgeType string    `gorm:"not null;index" json:"badge_type"` // 'selfie_verified', 'document_verified'
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false;index" json:"is_revoked"`
	RevokedAt *time.Time `json:"revoked_at"`
	RevokedBy *uuid.UUID `gorm:"type:uuid;index" json:"revoked_by"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User    *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Revoker *AdminUser `gorm:"foreignKey:RevokedBy;constraint:OnDelete:SET NULL" json:"revoker,omitempty"`
}

// TableName returns the table name for VerificationBadge model
func (VerificationBadge) TableName() string {
	return "verification_badges"
}

// BeforeCreate GORM hook
func (vb *VerificationBadge) BeforeCreate(tx *gorm.DB) error {
	if vb.ID == uuid.Nil {
		vb.ID = uuid.New()
	}
	return nil
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

// Revoke revokes verification badge
func (vb *VerificationBadge) Revoke(revokedBy uuid.UUID) {
	vb.IsRevoked = true
	vb.RevokedBy = &revokedBy
	now := time.Now()
	vb.RevokedAt = &now
}
