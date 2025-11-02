package testutils

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
)

// MockFactory provides factory methods for creating test entities
type MockFactory struct{}

// NewMockFactory creates a new mock factory
func NewMockFactory() *MockFactory {
	return &MockFactory{}
}

// CreateUser creates a mock user with optional overrides
func (f *MockFactory) CreateUser(overrides ...func(*entities.User)) *entities.User {
	userID := uuid.New()
	user := &entities.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "hashedpassword",
		IsVerified: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(user)
	}

	return user
}

// CreateProfile creates a mock profile with optional overrides
func (f *MockFactory) CreateProfile(userID uuid.UUID, overrides ...func(*entities.Profile)) *entities.Profile {
	profile := &entities.Profile{
		UserID:             userID,
		Bio:                "Test bio",
		Age:                25,
		Gender:             valueobjects.GenderMale,
		InterestedIn:       []valueobjects.Gender{valueobjects.GenderFemale},
		RelationshipStatus:  valueobjects.RelationshipStatusSingle,
		Location:           "Test City",
		Photos:             []string{},
		IsVerified:         true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	for _, override := range overrides {
		override(profile)
	}

	return profile
}

// CreateMatch creates a mock match with optional overrides
func (f *MockFactory) CreateMatch(userID1, userID2 uuid.UUID, overrides ...func(*entities.Match)) *entities.Match {
	match := &entities.Match{
		ID:        uuid.New(),
		UserID1:   userID1,
		UserID2:   userID2,
		MatchedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(match)
	}

	return match
}

// CreateMessage creates a mock message with optional overrides
func (f *MockFactory) CreateMessage(matchID, senderID uuid.UUID, overrides ...func(*entities.Message)) *entities.Message {
	message := &entities.Message{
		ID:        uuid.New(),
		MatchID:   matchID,
		SenderID:  senderID,
		Content:   "Test message",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(message)
	}

	return message
}

// CreatePhoto creates a mock photo with optional overrides
func (f *MockFactory) CreatePhoto(userID uuid.UUID, overrides ...func(*entities.Photo)) *entities.Photo {
	photo := &entities.Photo{
		ID:        uuid.New(),
		UserID:    userID,
		URL:       "https://example.com/photo.jpg",
		IsPrimary: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(photo)
	}

	return photo
}

// CreateReport creates a mock report with optional overrides
func (f *MockFactory) CreateReport(reporterID, reportedID uuid.UUID, overrides ...func(*entities.Report)) *entities.Report {
	report := &entities.Report{
		ID:          uuid.New(),
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Reason:      "Inappropriate content",
		Description: "Test report description",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, override := range overrides {
		override(report)
	}

	return report
}

// CreateSubscription creates a mock subscription with optional overrides
func (f *MockFactory) CreateSubscription(userID uuid.UUID, overrides ...func(*entities.Subscription)) *entities.Subscription {
	subscription := &entities.Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		PlanType:  "premium",
		Status:    "active",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(subscription)
	}

	return subscription
}

// CreateEphemeralPhoto creates a mock ephemeral photo with optional overrides
func (f *MockFactory) CreateEphemeralPhoto(userID uuid.UUID, overrides ...func(*entities.EphemeralPhoto)) *entities.EphemeralPhoto {
	ephemeralPhoto := &entities.EphemeralPhoto{
		ID:        uuid.New(),
		UserID:    userID,
		URL:       "https://example.com/ephemeral.jpg",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		ViewCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(ephemeralPhoto)
	}

	return ephemeralPhoto
}

// CreateVerification creates a mock verification with optional overrides
func (f *MockFactory) CreateVerification(userID uuid.UUID, overrides ...func(*entities.Verification)) *entities.Verification {
	verification := &entities.Verification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "photo",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(verification)
	}

	return verification
}

// CreatePayment creates a mock payment with optional overrides
func (f *MockFactory) CreatePayment(userID uuid.UUID, overrides ...func(*entities.Payment)) *entities.Payment {
	payment := &entities.Payment{
		ID:        uuid.New(),
		UserID:    userID,
		Amount:    9.99,
		Currency:  "USD",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(payment)
	}

	return payment
}

// UserWithProfile creates a user with an associated profile
func (f *MockFactory) UserWithProfile(overrides ...func(*entities.User)) (*entities.User, *entities.Profile) {
	user := f.CreateUser(overrides...)
	profile := f.CreateProfile(user.ID)
	return user, profile
}

// MatchWithUsers creates a match with associated users
func (f *MockFactory) MatchWithUsers(overrides ...func(*entities.Match)) (*entities.Match, *entities.User, *entities.User) {
	user1 := f.CreateUser()
	user2 := f.CreateUser()
	match := f.CreateMatch(user1.ID, user2.ID, overrides...)
	return match, user1, user2
}

// MessageWithMatch creates a message with associated match and users
func (f *MockFactory) MessageWithMatch(overrides ...func(*entities.Message)) (*entities.Message, *entities.Match, *entities.User, *entities.User) {
	match, user1, user2 := f.MatchWithUsers()
	message := f.CreateMessage(match.ID, user1.ID, overrides...)
	return message, match, user1, user2
}

// PhotoWithUser creates a photo with an associated user
func (f *MockFactory) PhotoWithUser(overrides ...func(*entities.Photo)) (*entities.Photo, *entities.User) {
	user := f.CreateUser()
	photo := f.CreatePhoto(user.ID, overrides...)
	return photo, user
}

// ReportWithUsers creates a report with associated users
func (f *MockFactory) ReportWithUsers(overrides ...func(*entities.Report)) (*entities.Report, *entities.User, *entities.User) {
	reporter := f.CreateUser()
	reported := f.CreateUser()
	report := f.CreateReport(reporter.ID, reported.ID, overrides...)
	return report, reporter, reported
}

// SubscriptionWithUser creates a subscription with an associated user
func (f *MockFactory) SubscriptionWithUser(overrides ...func(*entities.Subscription)) (*entities.Subscription, *entities.User) {
	user := f.CreateUser()
	subscription := f.CreateSubscription(user.ID, overrides...)
	return subscription, user
}

// EphemeralPhotoWithUser creates an ephemeral photo with an associated user
func (f *MockFactory) EphemeralPhotoWithUser(overrides ...func(*entities.EphemeralPhoto)) (*entities.EphemeralPhoto, *entities.User) {
	user := f.CreateUser()
	ephemeralPhoto := f.CreateEphemeralPhoto(user.ID, overrides...)
	return ephemeralPhoto, user
}

// VerificationWithUser creates a verification with an associated user
func (f *MockFactory) VerificationWithUser(overrides ...func(*entities.Verification)) (*entities.Verification, *entities.User) {
	user := f.CreateUser()
	verification := f.CreateVerification(user.ID, overrides...)
	return verification, user
}

// PaymentWithUser creates a payment with an associated user
func (f *MockFactory) PaymentWithUser(overrides ...func(*entities.Payment)) (*entities.Payment, *entities.User) {
	user := f.CreateUser()
	payment := f.CreatePayment(user.ID, overrides...)
	return payment, user
}

// CreateMultipleUsers creates multiple users with optional overrides
func (f *MockFactory) CreateMultipleUsers(count int, overrides ...func(*entities.User)) []*entities.User {
	users := make([]*entities.User, count)
	for i := 0; i < count; i++ {
		users[i] = f.CreateUser(func(u *entities.User) {
			u.Email = RandomEmail()
			for _, override := range overrides {
				override(u)
			}
		})
	}
	return users
}

// CreateMultiplePhotos creates multiple photos for a user
func (f *MockFactory) CreateMultiplePhotos(userID uuid.UUID, count int, overrides ...func(*entities.Photo)) []*entities.Photo {
	photos := make([]*entities.Photo, count)
	for i := 0; i < count; i++ {
		photos[i] = f.CreatePhoto(userID, func(p *entities.Photo) {
			p.URL = fmt.Sprintf("https://example.com/photo-%d.jpg", i)
			p.IsPrimary = i == 0 // First photo is primary
			for _, override := range overrides {
				override(p)
			}
		})
	}
	return photos
}

// CreateMultipleMessages creates multiple messages in a match
func (f *MockFactory) CreateMultipleMessages(matchID, senderID uuid.UUID, count int, overrides ...func(*entities.Message)) []*entities.Message {
	messages := make([]*entities.Message, count)
	for i := 0; i < count; i++ {
		messages[i] = f.CreateMessage(matchID, senderID, func(m *entities.Message) {
			m.Content = fmt.Sprintf("Test message %d", i+1)
			for _, override := range overrides {
				override(m)
			}
		})
	}
	return messages
}