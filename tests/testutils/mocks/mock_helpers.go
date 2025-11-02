package mocks

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockUserRepository provides a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindNearbyUsers(ctx context.Context, lat, lng float64, radiusKm int, limit int) ([]*entities.User, error) {
	args := m.Called(ctx, lat, lng, radiusKm, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastActive(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockPhotoRepository provides a mock implementation of PhotoRepository
type MockPhotoRepository struct {
	mock.Mock
}

func NewMockPhotoRepository() *MockPhotoRepository {
	return &MockPhotoRepository{}
}

func (m *MockPhotoRepository) Create(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) GetByID(ctx context.Context, id string) (*entities.Photo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Photo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Photo), args.Error(1)
}

func (m *MockPhotoRepository) Update(ctx context.Context, photo *entities.Photo) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockPhotoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPhotoRepository) SetPrimary(ctx context.Context, userID, photoID string) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

func (m *MockPhotoRepository) IncrementViewCount(ctx context.Context, photoID string) error {
	args := m.Called(ctx, photoID)
	return args.Error(0)
}

// MockMatchRepository provides a mock implementation of MatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func NewMockMatchRepository() *MockMatchRepository {
	return &MockMatchRepository{}
}

func (m *MockMatchRepository) Create(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id string) (*entities.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Match, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByUsers(ctx context.Context, userID1, userID2 string) (*entities.Match, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) Update(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockMessageRepository provides a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func NewMockMessageRepository() *MockMessageRepository {
	return &MockMessageRepository{}
}

func (m *MockMessageRepository) Create(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id string) (*entities.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*entities.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) GetUnreadCount(ctx context.Context, userID, conversationID string) (int, error) {
	args := m.Called(ctx, userID, conversationID)
	return args.Int(0), args.Error(1)
}

func (m *MockMessageRepository) MarkAsRead(ctx context.Context, messageIDs []string) error {
	args := m.Called(ctx, messageIDs)
	return args.Error(0)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *entities.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockReportRepository provides a mock implementation of ReportRepository
type MockReportRepository struct {
	mock.Mock
}

func NewMockReportRepository() *MockReportRepository {
	return &MockReportRepository{}
}

func (m *MockReportRepository) Create(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) GetByID(ctx context.Context, id string) (*entities.Report, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReporterID(ctx context.Context, reporterID string) ([]*entities.Report, error) {
	args := m.Called(ctx, reporterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportedID(ctx context.Context, reportedID string) ([]*entities.Report, error) {
	args := m.Called(ctx, reportedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Report), args.Error(1)
}

func (m *MockReportRepository) Update(ctx context.Context, report *entities.Report) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockReportRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockSubscriptionRepository provides a mock implementation of SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{}
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id string) (*entities.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetByUserID(ctx context.Context, userID string) (*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockVerificationRepository provides a mock implementation of VerificationRepository
type MockVerificationRepository struct {
	mock.Mock
}

func NewMockVerificationRepository() *MockVerificationRepository {
	return &MockVerificationRepository{}
}

func (m *MockVerificationRepository) Create(ctx context.Context, verification *entities.Verification) error {
	args := m.Called(ctx, verification)
	return args.Error(0)
}

func (m *MockVerificationRepository) GetByID(ctx context.Context, id string) (*entities.Verification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Verification), args.Error(1)
}

func (m *MockVerificationRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Verification, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Verification), args.Error(1)
}

func (m *MockVerificationRepository) Update(ctx context.Context, verification *entities.Verification) error {
	args := m.Called(ctx, verification)
	return args.Error(0)
}

func (m *MockVerificationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPaymentRepository provides a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func NewMockPaymentRepository() *MockPaymentRepository {
	return &MockPaymentRepository{}
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id string) (*entities.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Payment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *entities.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockEphemeralPhotoRepository provides a mock implementation of EphemeralPhotoRepository
type MockEphemeralPhotoRepository struct {
	mock.Mock
}

func NewMockEphemeralPhotoRepository() *MockEphemeralPhotoRepository {
	return &MockEphemeralPhotoRepository{}
}

func (m *MockEphemeralPhotoRepository) Create(ctx context.Context, photo *entities.EphemeralPhoto) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockEphemeralPhotoRepository) GetByID(ctx context.Context, id string) (*entities.EphemeralPhoto, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EphemeralPhoto), args.Error(1)
}

func (m *MockEphemeralPhotoRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.EphemeralPhoto, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.EphemeralPhoto), args.Error(1)
}

func (m *MockEphemeralPhotoRepository) GetByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error) {
	args := m.Called(ctx, accessKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EphemeralPhoto), args.Error(1)
}

func (m *MockEphemeralPhotoRepository) Update(ctx context.Context, photo *entities.EphemeralPhoto) error {
	args := m.Called(ctx, photo)
	return args.Error(0)
}

func (m *MockEphemeralPhotoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEphemeralPhotoRepository) IncrementViewCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockDriver is a mock SQL driver for testing
type MockDriver struct {
	*sqlmock.Sqlmock
}

func (d *MockDriver) Open(name string) (driver.Conn, error) {
	return d.Sqlmock, nil
}

// SetupTestDB creates a test database with mock
func SetupTestDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mock database: %w", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GORM instance: %w", err)
	}

	return gormDB, mock, nil
}

// RepositoryMocks holds all repository mocks
type RepositoryMocks struct {
	UserRepo         repositories.UserRepository
	PhotoRepo        repositories.PhotoRepository
	MatchRepo        repositories.MatchRepository
	MessageRepo      repositories.MessageRepository
	ReportRepo       repositories.ReportRepository
	SubscriptionRepo repositories.SubscriptionRepository
	VerificationRepo repositories.VerificationRepository
	PaymentRepo      repositories.PaymentRepository
	EphemeralPhotoRepo repositories.EphemeralPhotoRepository
}

// NewRepositoryMocks creates all repository mocks
func NewRepositoryMocks() *RepositoryMocks {
	return &RepositoryMocks{
		UserRepo:         NewMockUserRepository(),
		PhotoRepo:        NewMockPhotoRepository(),
		MatchRepo:        NewMockMatchRepository(),
		MessageRepo:      NewMockMessageRepository(),
		ReportRepo:       NewMockReportRepository(),
		SubscriptionRepo: NewMockSubscriptionRepository(),
		VerificationRepo: NewMockVerificationRepository(),
		PaymentRepo:      NewMockPaymentRepository(),
		EphemeralPhotoRepo: NewMockEphemeralPhotoRepository(),
	}
}