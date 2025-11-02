package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Report represents a user report entity in database
type Report struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReporterID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"reporter_id"`
	ReportedUserID uuid.UUID  `gorm:"type:uuid;not null;index" json:"reported_user_id"`
	Reason         string     `gorm:"not null;check:reason IN ('inappropriate_behavior', 'fake_profile', 'spam', 'harassment', 'other');index" json:"reason"`
	Description    *string    `gorm:"type:text" json:"description"`
	Status         string     `gorm:"default:'pending';check:status IN ('pending', 'reviewed', 'resolved', 'dismissed');index" json:"status"`
	ReviewedBy     *uuid.UUID `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Reporter     *User      `gorm:"foreignKey:ReporterID;constraint:OnDelete:CASCADE" json:"reporter,omitempty"`
	ReportedUser *User      `gorm:"foreignKey:ReportedUserID;constraint:OnDelete:CASCADE" json:"reported_user,omitempty"`
	Reviewer     *AdminUser `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewer,omitempty"`
}

// TableName returns the table name for the Report model
func (Report) TableName() string {
	return "reports"
}

// BeforeCreate GORM hook
func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
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

// AdminUser represents an admin user entity in database
type AdminUser struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email      string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FirstName  string     `gorm:"not null" json:"first_name"`
	LastName   string     `gorm:"not null" json:"last_name"`
	Role       string     `gorm:"default:'admin';check:role IN ('admin', 'moderator', 'super_admin');index" json:"role"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastLogin  *time.Time `json:"last_login"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	ReviewedReports      []*Report          `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewed_reports,omitempty"`
	ReviewedVerifications []*Verification     `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"-"`
	RevokedBadges       []*VerificationBadge `gorm:"foreignKey:RevokedBy;constraint:OnDelete:SET NULL" json:"-"`
	
	// Moderation relationships
	BannedUsers         []*Ban             `gorm:"foreignKey:BannerID;constraint:OnDelete:CASCADE" json:"banned_users,omitempty"`
	ModerationActions   []*ModerationAction `gorm:"foreignKey:ModeratorID;constraint:OnDelete:CASCADE" json:"moderation_actions,omitempty"`
	ReviewedAppeals     []*Appeal          `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewed_appeals,omitempty"`
	AssignedQueueItems  []*ModerationQueue `gorm:"foreignKey:AssignedTo;constraint:OnDelete:SET NULL" json:"assigned_queue_items,omitempty"`
	EscalatedQueueItems []*ModerationQueue `gorm:"foreignKey:EscalatedTo;constraint:OnDelete:SET NULL" json:"escalated_queue_items,omitempty"`
	ReviewedAnalyses    []*ContentAnalysis `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewed_analyses,omitempty"`
}

// TableName returns the table name for the AdminUser model
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

// CanReviewVerifications returns true if admin can review verifications
func (a *AdminUser) CanReviewVerifications() bool {
	return a.IsActive && (a.IsModerator() || a.IsSuperAdmin())
}