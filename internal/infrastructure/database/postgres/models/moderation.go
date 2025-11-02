package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ban represents a user ban record in the database
type Ban struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	BannerID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"banner_id"`
	Reason      string     `gorm:"not null" json:"reason"`
	ActionType  string     `gorm:"not null;check:action_type IN ('ban', 'suspend');index" json:"action_type"`
	Duration    *string    `gorm:"size:20" json:"duration"` // e.g., "7d", "30d", "permanent"
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at"`
	IsActive    bool       `gorm:"default:true;index" json:"is_active"`
	Notes       *string    `gorm:"type:text" json:"notes"`
	Evidence    *string    `gorm:"type:text" json:"evidence"`
	IsPermanent bool       `gorm:"default:false" json:"is_permanent"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User   *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Banner *AdminUser `gorm:"foreignKey:BannerID;constraint:OnDelete:CASCADE" json:"banner,omitempty"`
	Appeal *Appeal    `gorm:"foreignKey:BanID;constraint:OnDelete:CASCADE" json:"appeal,omitempty"`
}

// TableName returns the table name for the Ban model
func (Ban) TableName() string {
	return "bans"
}

// BeforeCreate GORM hook
func (b *Ban) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// IsActiveBan returns true if the ban is currently active
func (b *Ban) IsActiveBan() bool {
	if !b.IsActive {
		return false
	}
	
	if b.IsPermanent {
		return true
	}
	
	if b.ExpiresAt == nil {
		return true
	}
	
	return time.Now().Before(*b.ExpiresAt)
}

// IsExpired returns true if the ban has expired
func (b *Ban) IsExpired() bool {
	if b.IsPermanent {
		return false
	}
	
	if b.ExpiresAt == nil {
		return false
	}
	
	return time.Now().After(*b.ExpiresAt)
}

// Deactivate deactivates the ban
func (b *Ban) Deactivate() {
	b.IsActive = false
}

// Activate activates the ban
func (b *Ban) Activate() {
	b.IsActive = true
}

// SetExpiration sets the expiration time based on duration string
func (b *Ban) SetExpiration(duration string) error {
	if duration == "permanent" {
		b.IsPermanent = true
		b.ExpiresAt = nil
		return nil
	}
	
	d, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	
	b.IsPermanent = false
	expiresAt := time.Now().Add(d)
	b.ExpiresAt = &expiresAt
	
	return nil
}

// Appeal represents a ban appeal record in the database
type Appeal struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BanID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"ban_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Reason      string     `gorm:"not null;type:text" json:"reason"`
	Status      string     `gorm:"default:'pending';check:status IN ('pending', 'approved', 'rejected');index" json:"status"`
	ReviewedBy  *uuid.UUID `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	ReviewNotes *string    `gorm:"type:text" json:"review_notes"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Ban      *Ban       `gorm:"foreignKey:BanID;constraint:OnDelete:CASCADE" json:"ban,omitempty"`
	User     *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Reviewer *AdminUser `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewer,omitempty"`
}

// TableName returns the table name for the Appeal model
func (Appeal) TableName() string {
	return "appeals"
}

// BeforeCreate GORM hook
func (a *Appeal) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsPending returns true if the appeal is pending review
func (a *Appeal) IsPending() bool {
	return a.Status == "pending"
}

// IsApproved returns true if the appeal has been approved
func (a *Appeal) IsApproved() bool {
	return a.Status == "approved"
}

// IsRejected returns true if the appeal has been rejected
func (a *Appeal) IsRejected() bool {
	return a.Status == "rejected"
}

// CanBeReviewed returns true if the appeal can be reviewed
func (a *Appeal) CanBeReviewed() bool {
	return a.IsPending()
}

// Approve approves the appeal
func (a *Appeal) Approve(reviewedBy uuid.UUID, notes *string) {
	now := time.Now()
	a.Status = "approved"
	a.ReviewedBy = &reviewedBy
	a.ReviewedAt = &now
	a.ReviewNotes = notes
}

// Reject rejects the appeal
func (a *Appeal) Reject(reviewedBy uuid.UUID, notes *string) {
	now := time.Now()
	a.Status = "rejected"
	a.ReviewedBy = &reviewedBy
	a.ReviewedAt = &now
	a.ReviewNotes = notes
}

// ModerationAction represents a moderation action record in the database
type ModerationAction struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ModeratorID uuid.UUID  `gorm:"type:uuid;not null;index" json:"moderator_id"`
	ActionType  string     `gorm:"not null;check:action_type IN ('warning', 'content_removal', 'temporary_ban', 'permanent_ban', 'suspension');index" json:"action_type"`
	Reason      string     `gorm:"not null" json:"reason"`
	ContentID   *uuid.UUID `gorm:"type:uuid" json:"content_id"`
	ContentType *string    `gorm:"size:50" json:"content_type"`
	Severity    int        `gorm:"default:1;check:severity >= 1 AND severity <= 5" json:"severity"`
	Notes       *string    `gorm:"type:text" json:"notes"`
	Evidence    *string    `gorm:"type:text" json:"evidence"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	User      *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Moderator *AdminUser `gorm:"foreignKey:ModeratorID;constraint:OnDelete:CASCADE" json:"moderator,omitempty"`
}

// TableName returns the table name for the ModerationAction model
func (ModerationAction) TableName() string {
	return "moderation_actions"
}

// BeforeCreate GORM hook
func (m *ModerationAction) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// UserReputation represents a user's reputation score in the database
type UserReputation struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Score             int        `gorm:"default:100;index" json:"score"`
	ReportsReceived   int        `gorm:"default:0" json:"reports_received"`
	ReportsResolved   int        `gorm:"default:0" json:"reports_resolved"`
	ContentRemoved    int        `gorm:"default:0" json:"content_removed"`
	WarningsReceived  int        `gorm:"default:0" json:"warnings_received"`
	BansReceived      int        `gorm:"default:0" json:"bans_received"`
	AppealsSubmitted  int        `gorm:"default:0" json:"appeals_submitted"`
	AppealsApproved   int        `gorm:"default:0" json:"appeals_approved"`
	LastUpdated       time.Time  `gorm:"autoUpdateTime" json:"last_updated"`
	LastScoreChange   *time.Time `json:"last_score_change"`
	LastActionDate    *time.Time `json:"last_action_date"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for the UserReputation model
func (UserReputation) TableName() string {
	return "user_reputations"
}

// BeforeCreate GORM hook
func (u *UserReputation) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UpdateScore updates the reputation score
func (u *UserReputation) UpdateScore(change int, reason string) {
	u.Score += change
	now := time.Now()
	u.LastScoreChange = &now
	u.LastActionDate = &now
}

// IncrementReportsReceived increments the reports received count
func (u *UserReputation) IncrementReportsReceived() {
	u.ReportsReceived++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementReportsResolved increments the reports resolved count
func (u *UserReputation) IncrementReportsResolved() {
	u.ReportsResolved++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementContentRemoved increments the content removed count
func (u *UserReputation) IncrementContentRemoved() {
	u.ContentRemoved++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementWarningsReceived increments the warnings received count
func (u *UserReputation) IncrementWarningsReceived() {
	u.WarningsReceived++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementBansReceived increments the bans received count
func (u *UserReputation) IncrementBansReceived() {
	u.BansReceived++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementAppealsSubmitted increments the appeals submitted count
func (u *UserReputation) IncrementAppealsSubmitted() {
	u.AppealsSubmitted++
	now := time.Now()
	u.LastActionDate = &now
}

// IncrementAppealsApproved increments the appeals approved count
func (u *UserReputation) IncrementAppealsApproved() {
	u.AppealsApproved++
	now := time.Now()
	u.LastActionDate = &now
}

// GetReputationLevel returns the reputation level based on score
func (u *UserReputation) GetReputationLevel() string {
	if u.Score >= 90 {
		return "excellent"
	} else if u.Score >= 75 {
		return "good"
	} else if u.Score >= 50 {
		return "average"
	} else if u.Score >= 25 {
		return "poor"
	} else {
		return "terrible"
	}
}

// ModerationQueue represents an item in the moderation queue
type ModerationQueue struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReportID     *uuid.UUID `gorm:"type:uuid;index" json:"report_id"`
	AppealID     *uuid.UUID `gorm:"type:uuid;index" json:"appeal_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ContentType  string     `gorm:"not null;check:content_type IN ('report', 'appeal', 'content');index" json:"content_type"`
	ContentID    *uuid.UUID `gorm:"type:uuid" json:"content_id"`
	Priority     int        `gorm:"default:3;check:priority >= 1 AND priority <= 5;index" json:"priority"`
	Status       string     `gorm:"default:'pending';check:status IN ('pending', 'in_progress', 'completed', 'escalated');index" json:"status"`
	AssignedTo   *uuid.UUID `gorm:"type:uuid" json:"assigned_to"`
	AssignedAt   *time.Time `json:"assigned_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	EscalatedAt  *time.Time `json:"escalated_at"`
	EscalatedTo  *uuid.UUID `gorm:"type:uuid" json:"escalated_to"`
	Notes        *string    `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Report    *Report    `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE" json:"report,omitempty"`
	Appeal    *Appeal    `gorm:"foreignKey:AppealID;constraint:OnDelete:CASCADE" json:"appeal,omitempty"`
	User      *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Assigned  *AdminUser `gorm:"foreignKey:AssignedTo;constraint:OnDelete:SET NULL" json:"assigned,omitempty"`
	Escalated *AdminUser `gorm:"foreignKey:EscalatedTo;constraint:OnDelete:SET NULL" json:"escalated,omitempty"`
}

// TableName returns the table name for the ModerationQueue model
func (ModerationQueue) TableName() string {
	return "moderation_queue"
}

// BeforeCreate GORM hook
func (m *ModerationQueue) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsPending returns true if the queue item is pending
func (m *ModerationQueue) IsPending() bool {
	return m.Status == "pending"
}

// IsInProgress returns true if the queue item is in progress
func (m *ModerationQueue) IsInProgress() bool {
	return m.Status == "in_progress"
}

// IsCompleted returns true if the queue item is completed
func (m *ModerationQueue) IsCompleted() bool {
	return m.Status == "completed"
}

// IsEscalated returns true if the queue item is escalated
func (m *ModerationQueue) IsEscalated() bool {
	return m.Status == "escalated"
}

// CanBeAssigned returns true if the queue item can be assigned
func (m *ModerationQueue) CanBeAssigned() bool {
	return m.IsPending()
}

// Assign assigns the queue item to a moderator
func (m *ModerationQueue) Assign(moderatorID uuid.UUID) {
	now := time.Now()
	m.Status = "in_progress"
	m.AssignedTo = &moderatorID
	m.AssignedAt = &now
	m.StartedAt = &now
}

// Complete marks the queue item as completed
func (m *ModerationQueue) Complete() {
	now := time.Now()
	m.Status = "completed"
	m.CompletedAt = &now
}

// Escalate escalates the queue item to a higher level moderator
func (m *ModerationQueue) Escalate(escalatedTo uuid.UUID) {
	now := time.Now()
	m.Status = "escalated"
	m.EscalatedAt = &now
	m.EscalatedTo = &escalatedTo
}

// Unassign unassigns the queue item
func (m *ModerationQueue) Unassign() {
	m.Status = "pending"
	m.AssignedTo = nil
	m.AssignedAt = nil
	m.StartedAt = nil
}

// Block represents a user block record in the database
type Block struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BlockerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"blocker_id"`
	BlockedID uuid.UUID  `gorm:"type:uuid;not null;index" json:"blocked_id"`
	Reason    *string    `gorm:"type:text" json:"reason"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Blocker *User `gorm:"foreignKey:BlockerID;constraint:OnDelete:CASCADE" json:"blocker,omitempty"`
	Blocked *User `gorm:"foreignKey:BlockedID;constraint:OnDelete:CASCADE" json:"blocked,omitempty"`
}

// TableName returns the table name for the Block model
func (Block) TableName() string {
	return "blocks"
}

// BeforeCreate GORM hook
func (b *Block) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ContentAnalysis represents a content analysis record in the database
type ContentAnalysis struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ContentType     string     `gorm:"not null;check:content_type IN ('text', 'image', 'video', 'profile');index" json:"content_type"`
	ContentID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"content_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	AnalysisType    string     `gorm:"not null;check:analysis_type IN ('profanity', 'nsfw', 'violence', 'hate_speech', 'pii', 'link_safety');index" json:"analysis_type"`
	Score           float64    `gorm:"not null" json:"score"`
	Confidence      float64    `gorm:"not null" json:"confidence"`
	Result          string     `gorm:"not null;check:result IN ('safe', 'suspicious', 'violating');index" json:"result"`
	Details         *string    `gorm:"type:jsonb" json:"details"`
	ProcessedAt     time.Time  `gorm:"autoCreateTime" json:"processed_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	ReviewedBy      *uuid.UUID `gorm:"type:uuid" json:"reviewed_by"`
	ReviewResult    *string    `gorm:"check:review_result IN ('confirmed', 'false_positive', 'needs_review')" json:"review_result"`

	// Relationships
	User     *User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Reviewer *AdminUser `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL" json:"reviewer,omitempty"`
}

// TableName returns the table name for the ContentAnalysis model
func (ContentAnalysis) TableName() string {
	return "content_analyses"
}

// BeforeCreate GORM hook
func (c *ContentAnalysis) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsSafe returns true if the content is safe
func (c *ContentAnalysis) IsSafe() bool {
	return c.Result == "safe"
}

// IsSuspicious returns true if the content is suspicious
func (c *ContentAnalysis) IsSuspicious() bool {
	return c.Result == "suspicious"
}

// IsViolating returns true if the content is violating
func (c *ContentAnalysis) IsViolating() bool {
	return c.Result == "violating"
}

// NeedsReview returns true if the content needs review
func (c *ContentAnalysis) NeedsReview() bool {
	return c.IsSuspicious() || c.IsViolating()
}

// Confirm confirms the analysis result
func (c *ContentAnalysis) Confirm(reviewedBy uuid.UUID) {
	now := time.Now()
	c.ReviewedAt = &now
	c.ReviewedBy = &reviewedBy
	result := "confirmed"
	c.ReviewResult = &result
}

// MarkAsFalsePositive marks the analysis as false positive
func (c *ContentAnalysis) MarkAsFalsePositive(reviewedBy uuid.UUID) {
	now := time.Now()
	c.ReviewedAt = &now
	c.ReviewedBy = &reviewedBy
	result := "false_positive"
	c.ReviewResult = &result
}

// MarkAsNeedsReview marks the analysis as needing review
func (c *ContentAnalysis) MarkAsNeedsReview(reviewedBy uuid.UUID) {
	now := time.Now()
	c.ReviewedAt = &now
	c.ReviewedBy = &reviewedBy
	result := "needs_review"
	c.ReviewResult = &result
}