package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// EphemeralPhotoRepository defines the interface for ephemeral photo data operations
type EphemeralPhotoRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, photo *entities.EphemeralPhoto) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.EphemeralPhoto, error)
	GetByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error)
	Update(ctx context.Context, photo *entities.EphemeralPhoto) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User photo operations
	GetUserEphemeralPhotos(ctx context.Context, userID uuid.UUID, includeDeleted bool) ([]*entities.EphemeralPhoto, error)
	GetUserActiveEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error)
	GetUserEphemeralPhotoCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Status and lifecycle operations
	GetActiveEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetExpiredEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetViewedEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetExpiringSoonPhotos(ctx context.Context, within time.Duration, limit int) ([]*entities.EphemeralPhoto, error)

	// View tracking operations
	MarkAsViewed(ctx context.Context, photoID uuid.UUID) error
	IncrementViewCount(ctx context.Context, photoID uuid.UUID) error
	GetViewCount(ctx context.Context, photoID uuid.UUID) (int, error)

	// Expiration operations
	MarkAsExpired(ctx context.Context, photoID uuid.UUID) error
	BatchMarkExpired(ctx context.Context, photoIDs []uuid.UUID) error
	GetPhotosToExpire(ctx context.Context, before time.Time, limit int) ([]*entities.EphemeralPhoto, error)

	// File operations
	GetByFileKey(ctx context.Context, fileKey string) (*entities.EphemeralPhoto, error)
	UpdateFileURL(ctx context.Context, photoID uuid.UUID, fileURL string) error
	UpdateThumbnailURL(ctx context.Context, photoID uuid.UUID, thumbnailURL string) error

	// Batch operations
	BatchCreate(ctx context.Context, photos []*entities.EphemeralPhoto) error
	BatchUpdate(ctx context.Context, photos []*entities.EphemeralPhoto) error
	BatchDelete(ctx context.Context, photoIDs []uuid.UUID) error
	BatchSoftDelete(ctx context.Context, photoIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByAccessKey(ctx context.Context, accessKey string) (bool, error)
	ExistsByFileKey(ctx context.Context, fileKey string) (bool, error)
	UserHasPhoto(ctx context.Context, userID, photoID uuid.UUID) (bool, error)

	// Admin operations
	GetAllEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetEphemeralPhotosByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.EphemeralPhoto, error)

	// Analytics and statistics
	GetEphemeralPhotoStats(ctx context.Context) (*entities.EphemeralPhotoStats, error)
	GetPhotosUploadedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetPhotosViewedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetUserEphemeralPhotoStats(ctx context.Context, userID uuid.UUID) (*entities.EphemeralPhotoStats, error)

	// Advanced queries
	GetPhotosForCleanup(ctx context.Context, olderThan time.Time, limit int) ([]*entities.EphemeralPhoto, error)
	GetActivePhotosByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.EphemeralPhoto, error)
	GetExpiredPhotosByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.EphemeralPhoto, error)
}

// EphemeralPhotoViewRepository defines the interface for ephemeral photo view tracking
type EphemeralPhotoViewRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, view *entities.EphemeralPhotoView) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.EphemeralPhotoView, error)
	Update(ctx context.Context, view *entities.EphemeralPhotoView) error
	Delete(ctx context.Context, id uuid.UUID) error

	// View tracking operations
	GetViewsByPhoto(ctx context.Context, photoID uuid.UUID, limit, offset int) ([]*entities.EphemeralPhotoView, error)
	GetViewsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.EphemeralPhotoView, error)
	GetViewCountByPhoto(ctx context.Context, photoID uuid.UUID) (int, error)
	GetViewCountByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// Analytics operations
	GetViewsInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.EphemeralPhotoView, error)
	GetViewStats(ctx context.Context) (*ViewStats, error)
	GetUserViewStats(ctx context.Context, userID uuid.UUID) (*ViewStats, error)

	// Cleanup operations
	DeleteOldViews(ctx context.Context, olderThan time.Time) error
	MarkViewsAsExpired(ctx context.Context, photoID uuid.UUID) error

	// Batch operations
	BatchCreate(ctx context.Context, views []*entities.EphemeralPhotoView) error
	BatchDelete(ctx context.Context, viewIDs []uuid.UUID) error
}

// ViewStats represents view statistics
type ViewStats struct {
	TotalViews      int64 `json:"total_views"`
	UniqueViewers    int64 `json:"unique_viewers"`
	AverageViewTime  int64 `json:"average_view_time_seconds"`
	ViewsToday       int64 `json:"views_today"`
	ViewsThisWeek    int64 `json:"views_this_week"`
	ViewsThisMonth   int64 `json:"views_this_month"`
}