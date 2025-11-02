package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users map[uuid.UUID]*entities.User
	
	// Method call tracking
	CalledCreate              bool
	CalledGetByID             bool
	CalledGetByEmail          bool
	CalledUpdate              bool
	CalledDelete              bool
	CalledGetByLocation       bool
	CalledGetByPreferences    bool
	CalledUpdateLocation       bool
	CalledUpdateLastActive     bool
	CalledGetActiveUsers      bool
	CalledGetUsersByIDs       bool
	CalledSearchUsers         bool
	CalledCountUsers          bool
	CalledGetUserStats        bool
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*entities.User),
	}
}

// Create creates a new user
func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	m.CalledCreate = true
	m.users[user.ID] = user
	return nil
}

// GetByID gets a user by ID
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	m.CalledGetByID = true
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, repositories.ErrUserNotFound
}

// GetByEmail gets a user by email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	m.CalledGetByEmail = true
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

// Update updates a user
func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	m.CalledUpdate = true
	m.users[user.ID] = user
	return nil
}

// Delete deletes a user
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.CalledDelete = true
	delete(m.users, id)
	return nil
}

// GetByLocation gets users by location
func (m *MockUserRepository) GetByLocation(ctx context.Context, lat, lng float64, radiusKm float64, limit int) ([]*entities.User, error) {
	m.CalledGetByLocation = true
	var nearby []*entities.User
	for _, user := range m.users {
		if user.HasLocation() && user.IsActive && !user.IsBanned {
			nearby = append(nearby, user)
		}
	}
	if limit > 0 && len(nearby) > limit {
		nearby = nearby[:limit]
	}
	return nearby, nil
}

// GetByPreferences gets users by preferences
func (m *MockUserRepository) GetByPreferences(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.User, error) {
	m.CalledGetByPreferences = true
	var matched []*entities.User
	for _, user := range m.users {
		if user.ID != userID && user.IsActive && !user.IsBanned {
			matched = append(matched, user)
		}
	}
	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

// UpdateLocation updates a user's location
func (m *MockUserRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64, city, country string) error {
	m.CalledUpdateLocation = true
	if user, exists := m.users[userID]; exists {
		user.LocationLat = &lat
		user.LocationLng = &lng
		user.LocationCity = &city
		user.LocationCountry = &country
		m.users[userID] = user
	}
	return nil
}

// UpdateLastActive updates a user's last active time
func (m *MockUserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	m.CalledUpdateLastActive = true
	if user, exists := m.users[userID]; exists {
		now := time.Now()
		user.LastActive = &now
		m.users[userID] = user
	}
	return nil
}

// GetActiveUsers gets active users
func (m *MockUserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	m.CalledGetActiveUsers = true
	var active []*entities.User
	for _, user := range m.users {
		if user.IsActive && !user.IsBanned {
			active = append(active, user)
		}
	}
	if offset > 0 {
		if offset >= len(active) {
			return []*entities.User{}, nil
		}
		active = active[offset:]
	}
	if limit > 0 && len(active) > limit {
		active = active[:limit]
	}
	return active, nil
}

// GetUsersByIDs gets users by their IDs
func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*entities.User, error) {
	m.CalledGetUsersByIDs = true
	var users []*entities.User
	for _, id := range userIDs {
		if user, exists := m.users[id]; exists {
			users = append(users, user)
		}
	}
	return users, nil
}

// SearchUsers searches users by criteria
func (m *MockUserRepository) SearchUsers(ctx context.Context, query repositories.UserSearchQuery) ([]*entities.User, error) {
	m.CalledSearchUsers = true
	var results []*entities.User
	for _, user := range m.users {
		if user.IsActive && !user.IsBanned {
			results = append(results, user)
		}
	}
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}
	return results, nil
}

// CountUsers counts users by criteria
func (m *MockUserRepository) CountUsers(ctx context.Context, query repositories.UserSearchQuery) (int64, error) {
	m.CalledCountUsers = true
	var count int64
	for _, user := range m.users {
		if user.IsActive && !user.IsBanned {
			count++
		}
	}
	return count, nil
}

// GetUserStats gets user statistics
func (m *MockUserRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*repositories.UserStats, error) {
	m.CalledGetUserStats = true
	stats := &repositories.UserStats{
		TotalUsers:    int64(len(m.users)),
		ActiveUsers:    0,
		NewUsersToday:  0,
		NewUsersThisWeek: 0,
		NewUsersThisMonth: 0,
	}
	
	for _, user := range m.users {
		if user.IsActive && !user.IsBanned {
			stats.ActiveUsers++
		}
	}
	
	return stats, nil
}

// Reset resets the mock repository state
func (m *MockUserRepository) Reset() {
	m.users = make(map[uuid.UUID]*entities.User)
	
	m.CalledCreate = false
	m.CalledGetByID = false
	m.CalledGetByEmail = false
	m.CalledUpdate = false
	m.CalledDelete = false
	m.CalledGetByLocation = false
	m.CalledGetByPreferences = false
	m.CalledUpdateLocation = false
	m.CalledUpdateLastActive = false
	m.CalledGetActiveUsers = false
	m.CalledGetUsersByIDs = false
	m.CalledSearchUsers = false
	m.CalledCountUsers = false
	m.CalledGetUserStats = false
}

// AddUser adds a user to the mock repository
func (m *MockUserRepository) AddUser(user *entities.User) {
	m.users[user.ID] = user
}

// GetUserCount returns the number of users in the mock repository
func (m *MockUserRepository) GetUserCount() int {
	return len(m.users)
}