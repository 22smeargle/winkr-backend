package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EphemeralPhoto represents a user's ephemeral photo entity in database
type EphemeralPhoto struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	FileURL         string     `gorm:"not null" json:"file_url"`
	FileKey         string     `gorm:"not null;uniqueIndex" json:"file_key"`
	ThumbnailURL    string     `gorm:"not null" json:"thumbnail_url"`
	ThumbnailKey    string     `gorm:"not null" json:"thumbnail_key"`
	AccessKey       string     `gorm:"not null;uniqueIndex;index" json:"access_key"`
	IsViewed        bool       `gorm:"default:false;index" json:"is_viewed"`
	IsExpired       bool       `gorm:"default:false;index" json:"is_expired"`
	ViewCount       int        `gorm:"default:0" json:"view_count"`
	MaxViews        int        `gorm:"default:1" json:"max_views"`
	ExpiresAt       time.Time  `gorm:"not null;index" json:"expires_at"`
	ViewedAt        *time.Time `gorm:"type:timestamp" json:"viewed_at"`
	ExpiredAt       *time.Time `gorm:"type:timestamp" json:"expired_at"`
	IsDeleted       bool       `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for the EphemeralPhoto model
func (EphemeralPhoto) TableName() string {
	return "ephemeral_photos"
}

// BeforeCreate GORM hook
func (e *EphemeralPhoto) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// IsAccessible returns true if the photo can be accessed
func (e *EphemeralPhoto) IsAccessible() bool {
	return !e.IsDeleted && !e.IsExpired && !e.IsViewed && time.Now().Before(e.ExpiresAt)
}

// CanBeViewed returns true if the photo can be viewed
func (e *EphemeralPhoto) CanBeViewed() bool {
	return e.IsAccessible() && e.ViewCount < e.MaxViews
}

// MarkAsViewed marks the photo as viewed
func (e *EphemeralPhoto) MarkAsViewed() {
	now := time.Now()
	e.IsViewed = true
	e.ViewedAt = &now
	e.ViewCount++
}

// MarkAsExpired marks the photo as expired
func (e *EphemeralPhoto) MarkAsExpired() {
	now := time.Now()
	e.IsExpired = true
	e.ExpiredAt = &now
}

// SoftDelete marks the photo as deleted
func (e *EphemeralPhoto) SoftDelete() {
	e.IsDeleted = true
}

// Restore restores a soft-deleted photo
func (e *EphemeralPhoto) Restore() {
	e.IsDeleted = false
}

// GetRemainingTime returns the remaining time before expiration
func (e *EphemeralPhoto) GetRemainingTime() time.Duration {
	return time.Until(e.ExpiresAt)
}

// IsExpiredByTime returns true if the photo is expired by time
func (e *EphemeralPhoto) IsExpiredByTime() bool {
	return time.Now().After(e.ExpiresAt)
}

// GetViewStatus returns the current view status
func (e *EphemeralPhoto) GetViewStatus() string {
	if e.IsDeleted {
		return "deleted"
	}
	if e.IsExpired {
		return "expired"
	}
	if e.IsViewed {
		return "viewed"
	}
	if e.IsExpiredByTime() {
		return "time_expired"
	}
	return "available"
}

// EphemeralPhotoView represents a view record for analytics
type EphemeralPhotoView struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PhotoID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"photo_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ViewerID   *uuid.UUID `gorm:"type:uuid;index" json:"viewer_id"` // Optional: who viewed it
	IPAddress  string     `gorm:"not null" json:"ip_address"`
	UserAgent  string     `gorm:"type:text" json:"user_agent"`
	ViewedAt   time.Time  `gorm:"autoCreateTime;index" json:"viewed_at"`
	Duration   int        `gorm:"default:0" json:"duration_seconds"` // How long they viewed it
	IsExpired  bool       `gorm:"default:false" json:"is_expired"`

	// Relationships
	Photo *EphemeralPhoto `gorm:"foreignKey:PhotoID;constraint:OnDelete:CASCADE" json:"photo,omitempty"`
	User  *User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for the EphemeralPhotoView model
func (EphemeralPhotoView) TableName() string {
	return "ephemeral_photo_views"
}

// BeforeCreate GORM hook
func (e *EphemeralPhotoView) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}