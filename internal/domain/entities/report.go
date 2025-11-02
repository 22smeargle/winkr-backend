package entities

import (
	"time"

	"github.com/google/uuid"
)

// Report represents a user report entity
type Report struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReporterID     uuid.UUID  `json:"reporter_id" gorm:"type:uuid;not null;index"`
	ReportedUserID uuid.UUID  `json:"reported_user_id" gorm:"type:uuid;not null;index"`
	Reason         string     `json:"reason" gorm:"not null;check:reason IN ('inappropriate_behavior', 'fake_profile', 'spam', 'harassment', 'other')"`
	Description    *string    `json:"description"`
	Status         string     `json:"status" gorm:"default:'pending';check:status IN ('pending', 'reviewed', 'resolved', 'dismissed')"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by" gorm:"type:uuid"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Reporter     *User      `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	ReportedUser *User      `json:"reported_user,omitempty" gorm:"foreignKey:ReportedUserID"`
	Reviewer     *AdminUser `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// TableName returns the table name for the Report entity
func (Report) TableName() string {
	return "reports"
}

// AdminUser represents an admin user entity
type AdminUser struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email      string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string   `json:"-" gorm:"not null"`
	FirstName  string     `json:"first_name" gorm:"not null"`
	LastName   string     `json:"last_name" gorm:"not null"`
	Role       string     `json:"role" gorm:"default:'admin';check:role IN ('admin', 'moderator', 'super_admin')"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	LastLogin  *time.Time `json:"last_login"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	ReviewedReports []*Report `json:"reviewed_reports,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// TableName returns the table name for the AdminUser entity
func (AdminUser) TableName() string {
	return "admin_users"
}

// IsPending returns true if the report is pending review
func (r *Report) IsPending() bool {
	return r.Status == "pending"
}

// IsReviewed returns true if the report has been reviewed
func (r *Report) IsReviewed() bool {
	return r.Status == "reviewed"
}

// IsResolved returns true if the report has been resolved
func (r *Report) IsResolved() bool {
	return r.Status == "resolved"
}

// IsDismissed returns true if the report has been dismissed
func (r *Report) IsDismissed() bool {
	return r.Status == "dismissed"
}

// CanBeReviewed returns true if the report can be reviewed
func (r *Report) CanBeReviewed() bool {
	return r.IsPending()
}

// MarkAsReviewed marks the report as reviewed by an admin
func (r *Report) MarkAsReviewed(reviewedBy uuid.UUID) {
	now := time.Now()
	r.Status = "reviewed"
	r.ReviewedBy = &reviewedBy
	r.ReviewedAt = &now
}

// MarkAsResolved marks the report as resolved by an admin
func (r *Report) MarkAsResolved(reviewedBy uuid.UUID) {
	now := time.Now()
	r.Status = "resolved"
	r.ReviewedBy = &reviewedBy
	r.ReviewedAt = &now
}

// MarkAsDismissed marks the report as dismissed by an admin
func (r *Report) MarkAsDismissed(reviewedBy uuid.UUID) {
	now := time.Now()
	r.Status = "dismissed"
	r.ReviewedBy = &reviewedBy
	r.ReviewedAt = &now
}

// IsValidReason checks if the report reason is valid
func (r *Report) IsValidReason() bool {
	validReasons := []string{
		"inappropriate_behavior",
		"fake_profile",
		"spam",
		"harassment",
		"other",
	}
	
	for _, validReason := range validReasons {
		if r.Reason == validReason {
			return true
		}
	}
	return false
}

// IsAdmin returns true if the admin user has admin role
func (a *AdminUser) IsAdmin() bool {
	return a.Role == "admin"
}

// IsModerator returns true if the admin user has moderator role
func (a *AdminUser) IsModerator() bool {
	return a.Role == "moderator"
}

// IsSuperAdmin returns true if the admin user has super admin role
func (a *AdminUser) IsSuperAdmin() bool {
	return a.Role == "super_admin"
}

// CanManageReports returns true if the admin can manage reports
func (a *AdminUser) CanManageReports() bool {
	return a.IsActive && (a.IsAdmin() || a.IsModerator() || a.IsSuperAdmin())
}

// CanBanUsers returns true if the admin can ban users
func (a *AdminUser) CanBanUsers() bool {
	return a.IsActive && (a.IsAdmin() || a.IsSuperAdmin())
}

// CanManageAdmins returns true if the admin can manage other admins
func (a *AdminUser) CanManageAdmins() bool {
	return a.IsActive && a.IsSuperAdmin()
}

// UpdateLastLogin updates the last login time
func (a *AdminUser) UpdateLastLogin() {
	now := time.Now()
	a.LastLogin = &now
}

// Deactivate deactivates the admin user
func (a *AdminUser) Deactivate() {
	a.IsActive = false
}

// Activate activates the admin user
func (a *AdminUser) Activate() {
	a.IsActive = true
}