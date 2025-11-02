package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MatchFactory creates test match entities
type MatchFactory struct{}

// NewMatchFactory creates a new match factory
func NewMatchFactory() *MatchFactory {
	return &MatchFactory{}
}

// CreateMatch creates a test match with default values
func (f *MatchFactory) CreateMatch() *entities.Match {
	now := time.Now()
	matchID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()
	
	return &entities.Match{
		ID:         matchID,
		UserID1:    userID1,
		UserID2:    userID2,
		MatchedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateCustomMatch creates a test match with custom values
func (f *MatchFactory) CreateCustomMatch(opts ...MatchOption) *entities.Match {
	match := f.CreateMatch()
	
	for _, opt := range opts {
		opt(match)
	}
	
	return match
}

// MatchOption defines a function type for customizing match creation
type MatchOption func(*entities.Match)

// WithMatchID sets the match ID
func WithMatchID(id uuid.UUID) MatchOption {
	return func(m *entities.Match) {
		m.ID = id
	}
}

// WithUserIDs sets both user IDs
func WithUserIDs(userID1, userID2 uuid.UUID) MatchOption {
	return func(m *entities.Match) {
		m.UserID1 = userID1
		m.UserID2 = userID2
	}
}

// WithUserID1 sets the first user ID
func WithUserID1(userID1 uuid.UUID) MatchOption {
	return func(m *entities.Match) {
		m.UserID1 = userID1
	}
}

// WithUserID2 sets the second user ID
func WithUserID2(userID2 uuid.UUID) MatchOption {
	return func(m *entities.Match) {
		m.UserID2 = userID2
	}
}

// WithMatchedAt sets the match time
func WithMatchedAt(matchedAt time.Time) MatchOption {
	return func(m *entities.Match) {
		m.MatchedAt = matchedAt
	}
}

// WithMatchCreatedAt sets the creation time
func WithMatchCreatedAt(createdAt time.Time) MatchOption {
	return func(m *entities.Match) {
		m.CreatedAt = createdAt
	}
}

// WithMatchUpdatedAt sets the update time
func WithMatchUpdatedAt(updatedAt time.Time) MatchOption {
	return func(m *entities.Match) {
		m.UpdatedAt = updatedAt
	}
}

// CreateMultipleMatches creates multiple test matches
func (f *MatchFactory) CreateMultipleMatches(count int) []*entities.Match {
	matches := make([]*entities.Match, count)
	for i := 0; i < count; i++ {
		matches[i] = f.CreateMatch()
	}
	return matches
}

// CreateMultipleCustomMatches creates multiple test matches with custom options
func (f *MatchFactory) CreateMultipleCustomMatches(count int, opts ...MatchOption) []*entities.Match {
	matches := make([]*entities.Match, count)
	for i := 0; i < count; i++ {
		matches[i] = f.CreateCustomMatch(opts...)
	}
	return matches
}

// CreateMatchBetweenUsers creates a match between two specific users
func (f *MatchFactory) CreateMatchBetweenUsers(userID1, userID2 uuid.UUID) *entities.Match {
	return f.CreateCustomMatch(WithUserIDs(userID1, userID2))
}

// CreateMatchesForUser creates multiple matches for a specific user
func (f *MatchFactory) CreateMatchesForUser(userID uuid.UUID, count int) []*entities.Match {
	matches := make([]*entities.Match, count)
	for i := 0; i < count; i++ {
		match := f.CreateMatch()
		match.UserID1 = userID
		matches[i] = match
	}
	return matches
}