package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// PhotoRepository defines the interface for photo data operations
type PhotoRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, photo *entities.Photo) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Photo, error)
	Update(ctx context.Context, photo *entities.Photo) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User photo operations
	GetUserPhotos(ctx context.Context, userID uuid.UUID, includeDeleted bool) ([]*entities.Photo, error)
	GetUserPrimaryPhoto(ctx context.Context, userID uuid.UUID) (*entities.Photo, error)
	GetUserPhotoCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Photo verification operations
	GetPendingVerificationPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error)
	GetPhotosByVerificationStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Photo, error)
	UpdateVerificationStatus(ctx context.Context, photoID uuid.UUID, status string, reason *string) error

	// Photo management operations
	SetPrimaryPhoto(ctx context.Context, userID, photoID uuid.UUID) error
	UnsetPrimaryPhoto(ctx context.Context, userID uuid.UUID) error
	SoftDeletePhoto(ctx context.Context, photoID uuid.UUID) error
	RestorePhoto(ctx context.Context, photoID uuid.UUID) error

	// File operations
	GetByFileKey(ctx context.Context, fileKey string) (*entities.Photo, error)
	UpdateFileURL(ctx context.Context, photoID uuid.UUID, fileURL string) error

	// Batch operations
	BatchCreate(ctx context.Context, photos []*entities.Photo) error
	BatchUpdate(ctx context.Context, photos []*entities.Photo) error
	BatchDelete(ctx context.Context, photoIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByFileKey(ctx context.Context, fileKey string) (bool, error)
	UserHasPhoto(ctx context.Context, userID, photoID uuid.UUID) (bool, error)

	// Admin operations
	GetAllPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error)
	GetRejectedPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error)
	GetApprovedPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error)

	// Analytics and statistics
	GetPhotoStats(ctx context.Context) (*PhotoStats, error)
	GetPhotosUploadedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetPhotosByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Photo, error)

	// Advanced queries
	GetPhotosForVerification(ctx context.Context, limit int) ([]*entities.Photo, error)
	GetUserVerifiedPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error)
	GetUserUnverifiedPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error)
}

// PhotoStats represents photo statistics
type PhotoStats struct {
	TotalPhotos      int64 `json:"total_photos"`
	PendingPhotos    int64 `json:"pending_photos"`
	ApprovedPhotos   int64 `json:"approved_photos"`
	RejectedPhotos   int64 `json:"rejected_photos"`
	PhotosToday      int64 `json:"photos_today"`
	PhotosThisWeek   int64 `json:"photos_this_week"`
	PhotosThisMonth  int64 `json:"photos_this_month"`
}