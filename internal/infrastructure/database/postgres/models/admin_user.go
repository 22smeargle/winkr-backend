package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminUser represents an admin user in the database
type AdminUser struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FirstName    string     `gorm:"not null" json:"first_name"`
	LastName     string     `gorm:"not null" json:"last_name"`
	Role         string     `gorm:"not null;default:'admin';check:role IN ('admin', 'moderator', 'super_admin')" json:"role"`
	IsActive     bool       `gorm:"default:true;index" json:"is_active"`
	LastLogin    *time.Time `gorm:"index" json:"last_login"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	// Relationships
	ReviewedVerifications []*Verification `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"-"`
	RevokedBadges        []*VerificationBadge `gorm:"foreignKey:RevokedBy;constraint:OnDelete:SET NULL" json:"-"`
}

// TableName returns the table name for AdminUser model
func (AdminUser) TableName() string {
	return "admin_users"
}

// BeforeCreate GORM hook
func (a *AdminUser) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// GetFullName returns the admin user's full name
func (a *AdminUser) GetFullName() string {
	return a.FirstName + " " + a.LastName
}

// IsSuperAdmin returns true if the admin user is a super admin
func (a *AdminUser) IsSuperAdmin() bool {
	return a.Role == "super_admin"
}

// IsModerator returns true if the admin user is a moderator
func (a *AdminUser) IsModerator() bool {
	return a.Role == "moderator"
}

// CanReviewVerifications returns true if the admin user can review verifications
func (a *AdminUser) CanReviewVerifications() bool {
	return a.IsActive && (a.Role == "admin" || a.Role == "super_admin" || a.Role == "moderator")
}

// CanManageAdmins returns true if the admin user can manage other admins
func (a *AdminUser) CanManageAdmins() bool {
	return a.IsActive && (a.Role == "super_admin")
}

// UpdateLastLogin updates the admin user's last login time
func (a *AdminUser) UpdateLastLogin() {
	now := time.Now()
	a.LastLogin = &now
}