package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User specific operations
	GetByLocation(ctx context.Context, lat, lng float64, radiusKm int, limit, offset int) ([]*entities.User, error)
	GetPotentialMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetUsersByPreferences(ctx context.Context, userID uuid.UUID, preferences *entities.UserPreferences, limit, offset int) ([]*entities.User, error)
	UpdateLastActive(ctx context.Context, userID uuid.UUID) error
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)

	// User preferences operations
	GetPreferences(ctx context.Context, userID uuid.UUID) (*entities.UserPreferences, error)
	CreatePreferences(ctx context.Context, preferences *entities.UserPreferences) error
	UpdatePreferences(ctx context.Context, preferences *entities.UserPreferences) error

	// User statistics and analytics
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserStats, error)
	GetActiveUsersCount(ctx context.Context) (int64, error)
	GetUsersCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)

	// Admin operations
	GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
	GetBannedUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
	BanUser(ctx context.Context, userID uuid.UUID, reason string) error
	UnbanUser(ctx context.Context, userID uuid.UUID) error
	VerifyUser(ctx context.Context, userID uuid.UUID) error
	SetPremiumStatus(ctx context.Context, userID uuid.UUID, isPremium bool) error

	// Batch operations
	BatchCreate(ctx context.Context, users []*entities.User) error
	BatchUpdate(ctx context.Context, users []*entities.User) error
	BatchDelete(ctx context.Context, userIDs []uuid.UUID) error

	// Existence checks
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	// Advanced queries
	GetUsersWithPhotos(ctx context.Context, limit, offset int) ([]*entities.User, error)
	GetUsersWithoutPhotos(ctx context.Context, limit, offset int) ([]*entities.User, error)
	GetInactiveUsers(ctx context.Context, days int, limit, offset int) ([]*entities.User, error)
}

// UserStats represents user statistics
type UserStats struct {
	TotalSwipes      int64 `json:"total_swipes"`
	TotalMatches     int64 `json:"total_matches"`
	TotalMessages    int64 `json:"total_messages"`
	PhotosCount      int64 `json:"photos_count"`
	ProfileViews     int64 `json:"profile_views"`
	LastActiveDays   int   `json:"last_active_days"`
	AccountAgeDays   int   `json:"account_age_days"`
}