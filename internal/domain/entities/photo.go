package entities

import (
	"time"

	"github.com/google/uuid"
)

// Photo represents a user's photo entity
type Photo struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	FileURL           string     `json:"file_url" gorm:"not null"`
	FileKey           string     `json:"file_key" gorm:"not null;uniqueIndex"`
	IsPrimary         bool       `json:"is_primary" gorm:"default:false"`
	VerificationStatus string     `json:"verification_status" gorm:"default:'pending';check:verification_status IN ('pending', 'approved', 'rejected')"`
	VerificationReason *string    `json:"verification_reason"`
	IsDeleted         bool       `json:"is_deleted" gorm:"default:false"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for the Photo entity
func (Photo) TableName() string {
	return "photos"
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