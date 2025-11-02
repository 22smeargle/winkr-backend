package integration

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*entities.User, error) {
	args := m.Called(ctx, limit, offset, filters)
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*entities.User), args.Error(1)
}

// MockPhotoRepository is a mock implementation of PhotoRepository
type MockPhotoRepository struct {
	mock.Mock
}

func (m *MockPhotoRepository) Create(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Photo, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) Update(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPhotoRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) GetPrimaryByUserID(ctx context.Context, userID uuid.UUID) (*entities.Photo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) SetPrimary(ctx context.Context, userID, photoID uuid.UUID) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	args := m.Called(ctx, conversationID, userID)
	return args.Error(0)
}

func (m *MockMessageRepository) GetUnreadCount(ctx context.Context, conversationID, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, conversationID, userID)
	return args.Int(0), args.Error(1)
}

// MockMatchRepository is a mock implementation of MatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Match, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) Update(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Match, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetMatch(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Match, error) {
	args := m.Called(ctx, userID1, userID2)
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Match), args.Error(1)
}

// MockReportRepository is a mock implementation of ReportRepository
type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Report), args.Error(1)
}

func (m *MockReportRepository) Update(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReportRepository) GetByReporterID(ctx context.Context, reporterID uuid.UUID) ([]*entities.Report, error) {
	args := m.Called(ctx, reporterID)
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportedID(ctx context.Context, reportedID uuid.UUID) ([]*entities.Report, error) {
	args := m.Called(ctx, reportedID)
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Report), args.Error(1)
}

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Payment, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.Payment, error) {
	args := m.Called(ctx, subscriptionID)
	return args.Get(0).([]*entities.Payment), args.Error(1)
}

// MockSubscriptionRepository is a mock implementation of SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

// MockVerificationRepository is a mock implementation of VerificationRepository
type MockVerificationRepository struct {
	mock.Mock
}

func (m *MockVerificationRepository) Create(ctx context.Context, verification *entities.Verification) error {
	args := m.Called(ctx, verification)
	return args.Error(0)
}

func (m *MockVerificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Verification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Verification), args.Error(1)
}

func (m *MockVerificationRepository) Update(ctx context.Context, verification *entities.Verification) error {
	args := m.Called(ctx, verification)
	return args.Error(0)
}

func (m *MockVerificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVerificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Verification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Verification), args.Error(1)
}

func (m *MockVerificationRepository) GetPending(ctx context.Context, limit, offset int) ([]*entities.Verification, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entities.Verification), args.Error(1)
}

func (m *MockVerificationRepository) GetByType(ctx context.Context, verificationType string, limit, offset int) ([]*entities.Verification, error) {
	args := m.Called(ctx, verificationType, limit, offset)
	return args.Get(0).([]*entities.Verification), args.Error(1)
}

// Helper functions to create mock entities for testing

func createMockUser(id uuid.UUID, email string) *entities.User {
	return &entities.User{
		ID:        id,
		Email:     email,
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createMockPhoto(id, userID uuid.UUID) *entities.Photo {
	return &entities.Photo{
		ID:        id,
		UserID:    userID,
		URL:       "https://example.com/photo.jpg",
		IsPrimary: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createMockMessage(id, conversationID, senderID uuid.UUID) *entities.Message {
	return &entities.Message{
		ID:            id,
		ConversationID: conversationID,
		SenderID:      senderID,
		Content:       "Test message",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func createMockMatch(id, userID1, userID2 uuid.UUID) *entities.Match {
	return &entities.Match{
		ID:        id,
		UserID1:   userID1,
		UserID2:   userID2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createMockReport(id, reporterID, reportedID uuid.UUID) *entities.Report {
	return &entities.Report{
		ID:          id,
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Reason:      "inappropriate_content",
		Description:  "Test report",
		Status:       "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createMockPayment(id, userID, subscriptionID uuid.UUID) *entities.Payment {
	return &entities.Payment{
		ID:            id,
		UserID:        userID,
		SubscriptionID: subscriptionID,
		Amount:        19.99,
		Currency:      "USD",
		Status:        "completed",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func createMockSubscription(id, userID uuid.UUID) *entities.Subscription {
	return &entities.Subscription{
		ID:        id,
		UserID:    userID,
		Plan:      "premium",
		Status:    "active",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createMockVerification(id, userID uuid.UUID, verificationType string) *entities.Verification {
	return &entities.Verification{
		ID:               id,
		UserID:           userID,
		Type:             verificationType,
		Status:           "pending",
		VerificationData:  map[string]interface{}{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}
