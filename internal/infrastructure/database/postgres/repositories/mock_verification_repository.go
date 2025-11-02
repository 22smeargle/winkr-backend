package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
)

// MockVerificationRepository is a mock implementation of VerificationRepository for testing
type MockVerificationRepository struct {
	verifications        map[uuid.UUID]*entities.Verification
	attempts            map[uuid.UUID][]*entities.VerificationAttempt
	badges              map[uuid.UUID][]*entities.VerificationBadge
	adminUsers          map[uuid.UUID]*entities.AdminUser
	users               map[uuid.UUID]*entities.User
	
	// Method call tracking
	CalledCreateVerification              bool
	CalledGetVerificationByID            bool
	CalledGetVerificationByUserAndType  bool
	CalledGetPendingVerifications       bool
	CalledGetVerificationsByUser        bool
	CalledUpdateVerification            bool
	CalledDeleteVerification            bool
	CalledGetVerificationsForReview    bool
	CalledGetVerificationStats          bool
	CalledCreateVerificationAttempt     bool
	CalledGetVerificationAttemptsByUser bool
	CalledGetVerificationAttemptsByIP   bool
	CalledDeleteExpiredAttempts         bool
	CalledCreateVerificationBadge       bool
	CalledGetActiveBadgesByUser       bool
	CalledGetBadgeByUserAndType       bool
	CalledUpdateBadge                 bool
	CalledRevokeBadge                 bool
	CalledDeleteExpiredBadges          bool
	CalledGetUserVerificationLevel     bool
	CalledUpdateUserVerificationLevel   bool
	CalledGetUsersByVerificationLevel   bool
	CalledGetAdminUserByEmail         bool
	CalledGetAdminUserByID            bool
	CalledCreateAdminUser             bool
	CalledUpdateAdminUser             bool
	CalledUpdateAdminLastLogin        bool
	CalledGetActiveAdminUsers         bool
}

// NewMockVerificationRepository creates a new mock verification repository
func NewMockVerificationRepository() *MockVerificationRepository {
	return &MockVerificationRepository{
		verifications: make(map[uuid.UUID]*entities.Verification),
		attempts:       make(map[uuid.UUID][]*entities.VerificationAttempt),
		badges:         make(map[uuid.UUID][]*entities.VerificationBadge),
		adminUsers:     make(map[uuid.UUID]*entities.AdminUser),
		users:          make(map[uuid.UUID]*entities.User),
	}
}

// CreateVerification creates a new verification
func (m *MockVerificationRepository) CreateVerification(ctx context.Context, verification *entities.Verification) error {
	m.CalledCreateVerification = true
	m.verifications[verification.ID] = verification
	return nil
}

// GetVerificationByID gets a verification by ID
func (m *MockVerificationRepository) GetVerificationByID(ctx context.Context, id uuid.UUID) (*entities.Verification, error) {
	m.CalledGetVerificationByID = true
	if verification, exists := m.verifications[id]; exists {
		return verification, nil
	}
	return nil, nil
}

// GetVerificationByUserAndType gets a verification by user ID and type
func (m *MockVerificationRepository) GetVerificationByUserAndType(ctx context.Context, userID uuid.UUID, vType entities.VerificationType) (*entities.Verification, error) {
	m.CalledGetVerificationByUserAndType = true
	for _, verification := range m.verifications {
		if verification.UserID == userID && verification.Type == vType {
			return verification, nil
		}
	}
	return nil, nil
}

// GetPendingVerifications gets pending verifications
func (m *MockVerificationRepository) GetPendingVerifications(ctx context.Context, limit, offset int) ([]*entities.Verification, error) {
	m.CalledGetPendingVerifications = true
	var pending []*entities.Verification
	for _, verification := range m.verifications {
		if verification.Status == valueobjects.VerificationStatusPending {
			pending = append(pending, verification)
		}
	}
	return pending, nil
}

// GetVerificationsByUser gets verifications for a user
func (m *MockVerificationRepository) GetVerificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Verification, error) {
	m.CalledGetVerificationsByUser = true
	var userVerifications []*entities.Verification
	for _, verification := range m.verifications {
		if verification.UserID == userID {
			userVerifications = append(userVerifications, verification)
		}
	}
	return userVerifications, nil
}

// UpdateVerification updates a verification
func (m *MockVerificationRepository) UpdateVerification(ctx context.Context, verification *entities.Verification) error {
	m.CalledUpdateVerification = true
	m.verifications[verification.ID] = verification
	return nil
}

// DeleteVerification deletes a verification
func (m *MockVerificationRepository) DeleteVerification(ctx context.Context, id uuid.UUID) error {
	m.CalledDeleteVerification = true
	delete(m.verifications, id)
	return nil
}

// GetVerificationsForReview gets verifications that need review
func (m *MockVerificationRepository) GetVerificationsForReview(ctx context.Context, status valueobjects.VerificationStatus, limit, offset int) ([]*entities.Verification, error) {
	m.CalledGetVerificationsForReview = true
	var forReview []*entities.Verification
	for _, verification := range m.verifications {
		if verification.Status == status {
			forReview = append(forReview, verification)
		}
	}
	return forReview, nil
}

// GetVerificationStats gets verification statistics
func (m *MockVerificationRepository) GetVerificationStats(ctx context.Context) (*repositories.VerificationStats, error) {
	m.CalledGetVerificationStats = true
	stats := &repositories.VerificationStats{
		TotalVerifications:     int64(len(m.verifications)),
		PendingVerifications:  0,
		ApprovedVerifications: 0,
		RejectedVerifications: 0,
		SelfieVerifications:   0,
		DocumentVerifications: 0,
		VerifiedUsers:        0,
		VerificationRate:     0.0,
	}
	
	for _, verification := range m.verifications {
		switch verification.Status {
		case valueobjects.VerificationStatusPending:
			stats.PendingVerifications++
		case valueobjects.VerificationStatusApproved:
			stats.ApprovedVerifications++
		case valueobjects.VerificationStatusRejected:
			stats.RejectedVerifications++
		}
		
		switch verification.Type {
		case entities.VerificationTypeSelfie:
			stats.SelfieVerifications++
		case entities.VerificationTypeDocument:
			stats.DocumentVerifications++
		}
	}
	
	if stats.TotalVerifications > 0 {
		stats.VerificationRate = float64(stats.ApprovedVerifications) / float64(stats.TotalVerifications) * 100
	}
	
	return stats, nil
}

// CreateVerificationAttempt creates a new verification attempt
func (m *MockVerificationRepository) CreateVerificationAttempt(ctx context.Context, attempt *entities.VerificationAttempt) error {
	m.CalledCreateVerificationAttempt = true
	if m.attempts[attempt.UserID] == nil {
		m.attempts[attempt.UserID] = make([]*entities.VerificationAttempt, 0)
	}
	m.attempts[attempt.UserID] = append(m.attempts[attempt.UserID], attempt)
	return nil
}

// GetVerificationAttemptsByUser gets verification attempts for a user
func (m *MockVerificationRepository) GetVerificationAttemptsByUser(ctx context.Context, userID uuid.UUID, vType entities.VerificationType, since time.Time) ([]*entities.VerificationAttempt, error) {
	m.CalledGetVerificationAttemptsByUser = true
	if attempts, exists := m.attempts[userID]; exists {
		var filtered []*entities.VerificationAttempt
		for _, attempt := range attempts {
			if attempt.Type == vType && attempt.CreatedAt.After(since) {
				filtered = append(filtered, attempt)
			}
		}
		return filtered, nil
	}
	return nil, nil
}

// GetVerificationAttemptsByIP gets verification attempts from an IP address
func (m *MockVerificationRepository) GetVerificationAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) ([]*entities.VerificationAttempt, error) {
	m.CalledGetVerificationAttemptsByIP = true
	var ipAttempts []*entities.VerificationAttempt
	for _, attempts := range m.attempts {
		for _, attempt := range attempts {
			if attempt.IPAddress == ipAddress && attempt.CreatedAt.After(since) {
				ipAttempts = append(ipAttempts, attempt)
			}
		}
	}
	return ipAttempts, nil
}

// DeleteExpiredAttempts deletes expired verification attempts
func (m *MockVerificationRepository) DeleteExpiredAttempts(ctx context.Context, olderThan time.Time) error {
	m.CalledDeleteExpiredAttempts = true
	for userID, attempts := range m.attempts {
		var valid []*entities.VerificationAttempt
		for _, attempt := range attempts {
			if attempt.CreatedAt.After(olderThan) {
				valid = append(valid, attempt)
			}
		}
		m.attempts[userID] = valid
	}
	return nil
}

// CreateVerificationBadge creates a new verification badge
func (m *MockVerificationRepository) CreateVerificationBadge(ctx context.Context, badge *entities.VerificationBadge) error {
	m.CalledCreateVerificationBadge = true
	if m.badges[badge.UserID] == nil {
		m.badges[badge.UserID] = make([]*entities.VerificationBadge, 0)
	}
	m.badges[badge.UserID] = append(m.badges[badge.UserID], badge)
	return nil
}

// GetActiveBadgesByUser gets active badges for a user
func (m *MockVerificationRepository) GetActiveBadgesByUser(ctx context.Context, userID uuid.UUID) ([]*entities.VerificationBadge, error) {
	m.CalledGetActiveBadgesByUser = true
	if badges, exists := m.badges[userID]; exists {
		var active []*entities.VerificationBadge
		for _, badge := range badges {
			if !badge.IsRevoked && (badge.ExpiresAt == nil || badge.ExpiresAt.After(time.Now())) {
				active = append(active, badge)
			}
		}
		return active, nil
	}
	return nil, nil
}

// GetBadgeByUserAndType gets a badge by user and type
func (m *MockVerificationRepository) GetBadgeByUserAndType(ctx context.Context, userID uuid.UUID, badgeType string) (*entities.VerificationBadge, error) {
	m.CalledGetBadgeByUserAndType = true
	if badges, exists := m.badges[userID]; exists {
		for _, badge := range badges {
			if badge.BadgeType == badgeType {
				return badge, nil
			}
		}
	}
	return nil, nil
}

// UpdateBadge updates a verification badge
func (m *MockVerificationRepository) UpdateBadge(ctx context.Context, badge *entities.VerificationBadge) error {
	m.CalledUpdateBadge = true
	if badges, exists := m.badges[badge.UserID]; exists {
		for i, b := range badges {
			if b.ID == badge.ID {
				badges[i] = badge
				break
			}
		}
	}
	return nil
}

// RevokeBadge revokes a verification badge
func (m *MockVerificationRepository) RevokeBadge(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	m.CalledRevokeBadge = true
	for userID, badges := range m.badges {
		for _, badge := range badges {
			if badge.ID == id {
				badge.IsRevoked = true
				badge.RevokedBy = &revokedBy
				now := time.Now()
				badge.RevokedAt = &now
				m.badges[userID] = badges
				return nil
			}
		}
	}
	return nil
}

// DeleteExpiredBadges deletes expired badges
func (m *MockVerificationRepository) DeleteExpiredBadges(ctx context.Context) error {
	m.CalledDeleteExpiredBadges = true
	for userID, badges := range m.badges {
		var valid []*entities.VerificationBadge
		for _, badge := range badges {
			if badge.ExpiresAt == nil || badge.ExpiresAt.After(time.Now()) {
				valid = append(valid, badge)
			}
		}
		m.badges[userID] = valid
	}
	return nil
}

// GetUserVerificationLevel gets a user's verification level
func (m *MockVerificationRepository) GetUserVerificationLevel(ctx context.Context, userID uuid.UUID) (entities.VerificationLevel, error) {
	m.CalledGetUserVerificationLevel = true
	if user, exists := m.users[userID]; exists {
		return user.VerificationLevel, nil
	}
	return entities.VerificationLevelNone, nil
}

// UpdateUserVerificationLevel updates a user's verification level
func (m *MockVerificationRepository) UpdateUserVerificationLevel(ctx context.Context, userID uuid.UUID, level entities.VerificationLevel) error {
	m.CalledUpdateUserVerificationLevel = true
	if user, exists := m.users[userID]; exists {
		user.VerificationLevel = level
		m.users[userID] = user
	}
	return nil
}

// GetUsersByVerificationLevel gets users by verification level
func (m *MockVerificationRepository) GetUsersByVerificationLevel(ctx context.Context, level entities.VerificationLevel, limit, offset int) ([]*uuid.UUID, error) {
	m.CalledGetUsersByVerificationLevel = true
	var userIDs []*uuid.UUID
	for userID, user := range m.users {
		if user.VerificationLevel == level {
			userIDs = append(userIDs, &userID)
		}
	}
	return userIDs, nil
}

// GetAdminUserByEmail gets an admin user by email
func (m *MockVerificationRepository) GetAdminUserByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	m.CalledGetAdminUserByEmail = true
	for _, admin := range m.adminUsers {
		if admin.Email == email {
			return admin, nil
		}
	}
	return nil, nil
}

// GetAdminUserByID gets an admin user by ID
func (m *MockVerificationRepository) GetAdminUserByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	m.CalledGetAdminUserByID = true
	if admin, exists := m.adminUsers[id]; exists {
		return admin, nil
	}
	return nil, nil
}

// CreateAdminUser creates a new admin user
func (m *MockVerificationRepository) CreateAdminUser(ctx context.Context, admin *entities.AdminUser) error {
	m.CalledCreateAdminUser = true
	m.adminUsers[admin.ID] = admin
	return nil
}

// UpdateAdminUser updates an admin user
func (m *MockVerificationRepository) UpdateAdminUser(ctx context.Context, admin *entities.AdminUser) error {
	m.CalledUpdateAdminUser = true
	m.adminUsers[admin.ID] = admin
	return nil
}

// UpdateAdminLastLogin updates admin user's last login time
func (m *MockVerificationRepository) UpdateAdminLastLogin(ctx context.Context, id uuid.UUID) error {
	m.CalledUpdateAdminLastLogin = true
	if admin, exists := m.adminUsers[id]; exists {
		now := time.Now()
		admin.LastLogin = &now
		m.adminUsers[id] = admin
	}
	return nil
}

// GetActiveAdminUsers gets all active admin users
func (m *MockVerificationRepository) GetActiveAdminUsers(ctx context.Context) ([]*entities.AdminUser, error) {
	m.CalledGetActiveAdminUsers = true
	var active []*entities.AdminUser
	for _, admin := range m.adminUsers {
		if admin.IsActive {
			active = append(active, admin)
		}
	}
	return active, nil
}

// Reset resets the mock repository state
func (m *MockVerificationRepository) Reset() {
	m.verifications = make(map[uuid.UUID]*entities.Verification)
	m.attempts = make(map[uuid.UUID][]*entities.VerificationAttempt)
	m.badges = make(map[uuid.UUID][]*entities.VerificationBadge)
	m.adminUsers = make(map[uuid.UUID]*entities.AdminUser)
	m.users = make(map[uuid.UUID]*entities.User)
	
	m.CalledCreateVerification = false
	m.CalledGetVerificationByID = false
	m.CalledGetVerificationByUserAndType = false
	m.CalledGetPendingVerifications = false
	m.CalledGetVerificationsByUser = false
	m.CalledUpdateVerification = false
	m.CalledDeleteVerification = false
	m.CalledGetVerificationsForReview = false
	m.CalledGetVerificationStats = false
	m.CalledCreateVerificationAttempt = false
	m.CalledGetVerificationAttemptsByUser = false
	m.CalledGetVerificationAttemptsByIP = false
	m.CalledDeleteExpiredAttempts = false
	m.CalledCreateVerificationBadge = false
	m.CalledGetActiveBadgesByUser = false
	m.CalledGetBadgeByUserAndType = false
	m.CalledUpdateBadge = false
	m.CalledRevokeBadge = false
	m.CalledDeleteExpiredBadges = false
	m.CalledGetUserVerificationLevel = false
	m.CalledUpdateUserVerificationLevel = false
	m.CalledGetUsersByVerificationLevel = false
	m.CalledGetAdminUserByEmail = false
	m.CalledGetAdminUserByID = false
	m.CalledCreateAdminUser = false
	m.CalledUpdateAdminUser = false
	m.CalledUpdateAdminLastLogin = false
	m.CalledGetActiveAdminUsers = false
}

// AddUser adds a user to the mock repository
func (m *MockVerificationRepository) AddUser(user *entities.User) {
	m.users[user.ID] = user
}

// AddVerification adds a verification to the mock repository
func (m *MockVerificationRepository) AddVerification(verification *entities.Verification) {
	m.verifications[verification.ID] = verification
}

// AddAdminUser adds an admin user to the mock repository
func (m *MockVerificationRepository) AddAdminUser(admin *entities.AdminUser) {
	m.adminUsers[admin.ID] = admin
}