package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Photo represents a user's photo entity in database
type Photo struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	FileURL           string     `gorm:"not null" json:"file_url"`
	FileKey           string     `gorm:"not null;uniqueIndex" json:"file_key"`
	IsPrimary         bool       `gorm:"default:false;index" json:"is_primary"`
	VerificationStatus string     `gorm:"default:'pending';check:verification_status IN ('pending', 'approved', 'rejected');index" json:"verification_status"`
	VerificationReason *string    `gorm:"type:text" json:"verification_reason"`
	IsDeleted         bool       `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for the Photo model
func (Photo) TableName() string {
	return "photos"
}

// BeforeCreate GORM hook
func (p *Photo) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// IsVerified returns true if the photo is verified
func (p *Photo) IsVerified() bool {
	return p.VerificationStatus == "approved"
}

// IsRejected returns true if the photo is rejected
func (p *Photo) IsRejected() bool {
	return p.VerificationStatus == "rejected"
}

// IsPending returns true if the photo is pending verification
func (p *Photo) IsPending() bool {
	return p.VerificationStatus == "pending"
}

// CanBeDeleted returns true if the photo can be deleted
func (p *Photo) CanBeDeleted() bool {
	return !p.IsPrimary
}

// Approve marks the photo as approved
func (p *Photo) Approve() {
	p.VerificationStatus = "approved"
	p.VerificationReason = nil
}

// Reject marks the photo as rejected with a reason
func (p *Photo) Reject(reason string) {
	p.VerificationStatus = "rejected"
	p.VerificationReason = &reason
}

// SoftDelete marks the photo as deleted
func (p *Photo) SoftDelete() {
	p.IsDeleted = true
}

// Restore restores a soft-deleted photo
func (p *Photo) Restore() {
	p.IsDeleted = false
}