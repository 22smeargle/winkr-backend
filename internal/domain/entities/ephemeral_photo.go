package entities

import (
	"time"

	"github.com/google/uuid"
)

// EphemeralPhoto represents an ephemeral photo entity
type EphemeralPhoto struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	FileURL         string     `json:"file_url" gorm:"not null"`
	FileKey         string     `json:"file_key" gorm:"not null;uniqueIndex"`
	ThumbnailURL    string     `json:"thumbnail_url" gorm:"not null"`
	ThumbnailKey    string     `json:"thumbnail_key" gorm:"not null"`
	AccessKey       string     `json:"access_key" gorm:"not null;uniqueIndex"`
	IsViewed        bool       `json:"is_viewed" gorm:"default:false;index"`
	IsExpired       bool       `json:"is_expired" gorm:"default:false;index"`
	ViewCount       int        `json:"view_count" gorm:"default:0"`
	MaxViews        int        `json:"max_views" gorm:"default:1"`
	ExpiresAt       time.Time  `json:"expires_at" gorm:"not null;index"`
	ViewedAt        *time.Time `json:"viewed_at"`
	ExpiredAt       *time.Time `json:"expired_at"`
	IsDeleted       bool       `json:"is_deleted" gorm:"default:false;index"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for the EphemeralPhoto entity
func (EphemeralPhoto) TableName() string {
	return "ephemeral_photos"
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

// EphemeralPhotoStats represents ephemeral photo statistics
type EphemeralPhotoStats struct {
	TotalPhotos      int64 `json:"total_photos"`
	ActivePhotos     int64 `json:"active_photos"`
	ViewedPhotos     int64 `json:"viewed_photos"`
	ExpiredPhotos    int64 `json:"expired_photos"`
	DeletedPhotos    int64 `json:"deleted_photos"`
	PhotosToday      int64 `json:"photos_today"`
	PhotosThisWeek   int64 `json:"photos_this_week"`
	PhotosThisMonth  int64 `json:"photos_this_month"`
	AverageViewTime  int64 `json:"average_view_time_seconds"`
	TotalViews       int64 `json:"total_views"`
}

// EphemeralPhotoView represents a view record for analytics
type EphemeralPhotoView struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PhotoID    uuid.UUID  `json:"photo_id" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	ViewerID   *uuid.UUID `json:"viewer_id" gorm:"type:uuid;index"` // Optional: who viewed it
	IPAddress  string     `json:"ip_address" gorm:"not null"`
	UserAgent  string     `json:"user_agent"`
	ViewedAt   time.Time  `json:"viewed_at" gorm:"autoCreateTime"`
	Duration   int        `json:"duration_seconds"` // How long they viewed it
	IsExpired  bool       `json:"is_expired" gorm:"default:false"`
}

// TableName returns the table name for the EphemeralPhotoView entity
func (EphemeralPhotoView) TableName() string {
	return "ephemeral_photo_views"
}