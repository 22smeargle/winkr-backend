package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
)

// UserFactory creates test user entities
type UserFactory struct{}

// NewUserFactory creates a new user factory
func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

// CreateUser creates a test user with default values
func (f *UserFactory) CreateUser() *entities.User {
	now := time.Now()
	userID := uuid.New()
	
	return &entities.User{
		ID:        userID,
		Email:     "test-" + userID.String() + "@example.com",
		Password:  "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjdIrEjQ2Q8J2K2vJ2K2vJ2K2vJ2K2", // hashed "password123"
		FirstName: "Test",
		LastName:  "User",
		DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Gender:    valueobjects.Male,
		InterestedIn: []valueobjects.Gender{valueobjects.Female},
		Bio:       "Test user bio",
		Location: &entities.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			City:      "New York",
			Country:   "USA",
		},
		IsVerified:     false,
		IsActive:       true,
		IsPremium:      false,
		ProfileVisible: true,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastActiveAt:   now,
	}
}

// CreateVerifiedUser creates a verified test user
func (f *UserFactory) CreateVerifiedUser() *entities.User {
	user := f.CreateUser()
	user.IsVerified = true
	return user
}

// CreatePremiumUser creates a premium test user
func (f *UserFactory) CreatePremiumUser() *entities.User {
	user := f.CreateUser()
	user.IsPremium = true
	return user
}

// CreateInactiveUser creates an inactive test user
func (f *UserFactory) CreateInactiveUser() *entities.User {
	user := f.CreateUser()
	user.IsActive = false
	return user
}

// CreateCustomUser creates a test user with custom values
func (f *UserFactory) CreateCustomUser(opts ...UserOption) *entities.User {
	user := f.CreateUser()
	
	for _, opt := range opts {
		opt(user)
	}
	
	return user
}

// UserOption defines a function type for customizing user creation
type UserOption func(*entities.User)

// WithEmail sets the user email
func WithEmail(email string) UserOption {
	return func(u *entities.User) {
		u.Email = email
	}
}

// WithPassword sets the user password
func WithPassword(password string) UserOption {
	return func(u *entities.User) {
		u.Password = password
	}
}

// WithName sets the user first and last name
func WithName(firstName, lastName string) UserOption {
	return func(u *entities.User) {
		u.FirstName = firstName
		u.LastName = lastName
	}
}

// WithGender sets the user gender
func WithGender(gender valueobjects.Gender) UserOption {
	return func(u *entities.User) {
		u.Gender = gender
	}
}

// WithInterestedIn sets the user's interested in genders
func WithInterestedIn(interestedIn []valueobjects.Gender) UserOption {
	return func(u *entities.User) {
		u.InterestedIn = interestedIn
	}
}

// WithLocation sets the user location
func WithLocation(lat, lng float64, city, country string) UserOption {
	return func(u *entities.User) {
		u.Location = &entities.Location{
			Latitude:  lat,
			Longitude: lng,
			City:      city,
			Country:   country,
		}
	}
}

// WithBio sets the user bio
func WithBio(bio string) UserOption {
	return func(u *entities.User) {
		u.Bio = bio
	}
}

// WithVerified sets the verification status
func WithVerified(verified bool) UserOption {
	return func(u *entities.User) {
		u.IsVerified = verified
	}
}

// WithActive sets the active status
func WithActive(active bool) UserOption {
	return func(u *entities.User) {
		u.IsActive = active
	}
}

// WithPremium sets the premium status
func WithPremium(premium bool) UserOption {
	return func(u *entities.User) {
		u.IsPremium = premium
	}
}

// WithProfileVisible sets the profile visibility
func WithProfileVisible(visible bool) UserOption {
	return func(u *entities.User) {
		u.ProfileVisible = visible
	}
}

// WithCreatedAt sets the creation time
func WithCreatedAt(createdAt time.Time) UserOption {
	return func(u *entities.User) {
		u.CreatedAt = createdAt
	}
}

// WithUpdatedAt sets the update time
func WithUpdatedAt(updatedAt time.Time) UserOption {
	return func(u *entities.User) {
		u.UpdatedAt = updatedAt
	}
}

// WithLastActiveAt sets the last active time
func WithLastActiveAt(lastActiveAt time.Time) UserOption {
	return func(u *entities.User) {
		u.LastActiveAt = lastActiveAt
	}
}

// CreateMultipleUsers creates multiple test users
func (f *UserFactory) CreateMultipleUsers(count int) []*entities.User {
	users := make([]*entities.User, count)
	for i := 0; i < count; i++ {
		users[i] = f.CreateUser()
	}
	return users
}

// CreateMultipleCustomUsers creates multiple test users with custom options
func (f *UserFactory) CreateMultipleCustomUsers(count int, opts ...UserOption) []*entities.User {
	users := make([]*entities.User, count)
	for i := 0; i < count; i++ {
		users[i] = f.CreateCustomUser(opts...)
	}
	return users
}