package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MatchRepository defines interface for match and swipe data operations
type MatchRepository interface {
	// Match operations
	CreateMatch(ctx context.Context, match *entities.Match) error
	GetMatchByID(ctx context.Context, id uuid.UUID) (*entities.Match, error)
	GetMatchByUsers(ctx context.Context, user1ID, user2ID uuid.UUID) (*entities.Match, error)
	UpdateMatch(ctx context.Context, match *entities.Match) error
	DeleteMatch(ctx context.Context, id uuid.UUID) error

	// User match operations
	GetUserMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error)
	GetUserMatchesWithDetails(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*MatchWithDetails, error)
	GetActiveMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error)
	GetMatchCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// Swipe operations
	CreateSwipe(ctx context.Context, swipe *entities.Swipe) error
	GetSwipe(ctx context.Context, swiperID, swipedID uuid.UUID) (*entities.Swipe, error)
	UpdateSwipe(ctx context.Context, swipe *entities.Swipe) error
	DeleteSwipe(ctx context.Context, swiperID, swipedID uuid.UUID) error

	// User swipe operations
	GetUserSwipes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error)
	GetUserLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error)
	GetUserPasses(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error)
	GetSwipeCount(ctx context.Context, userID uuid.UUID) (int64, error)
	GetLikeCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// Swipe existence checks
	HasSwiped(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error)
	GetSwipeDirection(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) // returns isLike

	// Matching logic
	CheckForMatch(ctx context.Context, swiperID, swipedID uuid.UUID) (*entities.Match, bool, error)
	GetMutualLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetUsersWhoLikedUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)

	// Potential matches
	GetPotentialMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetPotentialMatchesExcluding(ctx context.Context, userID uuid.UUID, excludeUserIDs []uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetPotentialMatchesByPreferences(ctx context.Context, userID uuid.UUID, preferences *entities.UserPreferences, limit, offset int) ([]*entities.User, error)

	// Batch operations
	BatchCreateSwipes(ctx context.Context, swipes []*entities.Swipe) error
	BatchCreateMatches(ctx context.Context, matches []*entities.Match) error

	// Existence checks
	MatchExists(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, error)
	SwipeExists(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error)

	// Analytics and statistics
	GetMatchStats(ctx context.Context, userID uuid.UUID) (*MatchStats, error)
	GetSwipeStats(ctx context.Context, userID uuid.UUID) (*SwipeStats, error)
	GetMatchesCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error)
	GetSwipesCreatedInRange(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (int64, error)

	// Admin operations
	GetAllMatches(ctx context.Context, limit, offset int) ([]*entities.Match, error)
	GetAllSwipes(ctx context.Context, limit, offset int) ([]*entities.Swipe, error)
	GetMatchAnalytics(ctx context.Context, startDate, endDate interface{}) (*MatchAnalytics, error)

	// Advanced queries
	GetRecentMatches(ctx context.Context, userID uuid.UUID, days int, limit int) ([]*entities.Match, error)
	GetUnreadMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.Match, error)
	GetMatchesWithoutConversation(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.Match, error)
	GetSwipeHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*SwipeWithUser, error)
}

// MatchWithDetails represents a match with additional details
type MatchWithDetails struct {
	*entities.Match
	OtherUser      *entities.User  `json:"other_user"`
	LastMessage    *entities.Message `json:"last_message,omitempty"`
	UnreadCount    int              `json:"unread_count"`
	HasConversation bool             `json:"has_conversation"`
}

// SwipeWithUser represents a swipe with user details
type SwipeWithUser struct {
	*entities.Swipe
	SwipedUser *entities.User `json:"swiped_user"`
}

// MatchStats represents match statistics for a user
type MatchStats struct {
	TotalMatches     int64 `json:"total_matches"`
	ActiveMatches    int64 `json:"active_matches"`
	MatchesToday     int64 `json:"matches_today"`
	MatchesThisWeek  int64 `json:"matches_this_week"`
	MatchesThisMonth int64 `json:"matches_this_month"`
	AverageMatchAge  int   `json:"average_match_age_days"`
}

// SwipeStats represents swipe statistics for a user
type SwipeStats struct {
	TotalSwipes     int64 `json:"total_swipes"`
	TotalLikes       int64 `json:"total_likes"`
	TotalPasses      int64 `json:"total_passes"`
	SwipesToday     int64 `json:"swipes_today"`
	SwipesThisWeek  int64 `json:"swipes_this_week"`
	SwipesThisMonth int64 `json:"swipes_this_month"`
	LikeRate         float64 `json:"like_rate"`
}

// MatchAnalytics represents global match analytics
type MatchAnalytics struct {
	TotalMatches      int64   `json:"total_matches"`
	TotalSwipes       int64   `json:"total_swipes"`
	MatchRate         float64 `json:"match_rate"`
	AverageSwipesPerUser float64 `json:"average_swipes_per_user"`
	AverageMatchesPerUser float64 `json:"average_matches_per_user"`
	PeakMatchingHour  int     `json:"peak_matching_hour"`
	PeakMatchingDay   string  `json:"peak_matching_day"`
}