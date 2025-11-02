package testutils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MockEmailService is a mock implementation of email service
type MockEmailService struct {
	mock.Mock
	SentEmails []EmailRecord
}

// EmailRecord represents a sent email record
type EmailRecord struct {
	To      string
	Subject string
	Body    string
	Time    time.Time
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]EmailRecord, 0),
	}
}

// SendEmail mocks sending an email
func (m *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	
	// Record the email
	m.SentEmails = append(m.SentEmails, EmailRecord{
		To:      to,
		Subject: subject,
		Body:    body,
		Time:    time.Now(),
	})
	
	return args.Error(0)
}

// SendVerificationEmail mocks sending a verification email
func (m *MockEmailService) SendVerificationEmail(ctx context.Context, to, token string) error {
	args := m.Called(ctx, to, token)
	return args.Error(0)
}

// SendPasswordResetEmail mocks sending a password reset email
func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	args := m.Called(ctx, to, token)
	return args.Error(0)
}

// GetLastSentEmail returns the last sent email
func (m *MockEmailService) GetLastSentEmail() *EmailRecord {
	if len(m.SentEmails) == 0 {
		return nil
	}
	return &m.SentEmails[len(m.SentEmails)-1]
}

// GetEmailsTo returns all emails sent to a specific address
func (m *MockEmailService) GetEmailsTo(to string) []EmailRecord {
	var emails []EmailRecord
	for _, email := range m.SentEmails {
		if email.To == to {
			emails = append(emails, email)
		}
	}
	return emails
}

// ClearEmails clears all email records
func (m *MockEmailService) ClearEmails() {
	m.SentEmails = make([]EmailRecord, 0)
}

// MockSMSService is a mock implementation of SMS service
type MockSMSService struct {
	mock.Mock
	SentSMSs []SMSRecord
}

// SMSRecord represents a sent SMS record
type SMSRecord struct {
	To      string
	Message string
	Time    time.Time
}

// NewMockSMSService creates a new mock SMS service
func NewMockSMSService() *MockSMSService {
	return &MockSMSService{
		SentSMSs: make([]SMSRecord, 0),
	}
}

// SendSMS mocks sending an SMS
func (m *MockSMSService) SendSMS(ctx context.Context, to, message string) error {
	args := m.Called(ctx, to, message)
	
	// Record the SMS
	m.SentSMSs = append(m.SentSMSs, SMSRecord{
		To:      to,
		Message: message,
		Time:    time.Now(),
	})
	
	return args.Error(0)
}

// SendVerificationSMS mocks sending a verification SMS
func (m *MockSMSService) SendVerificationSMS(ctx context.Context, to, code string) error {
	args := m.Called(ctx, to, code)
	return args.Error(0)
}

// GetLastSentSMS returns the last sent SMS
func (m *MockSMSService) GetLastSentSMS() *SMSRecord {
	if len(m.SentSMSs) == 0 {
		return nil
	}
	return &m.SentSMSs[len(m.SentSMSs)-1]
}

// GetSMSTo returns all SMS sent to a specific number
func (m *MockSMSService) GetSMSTo(to string) []SMSRecord {
	var smsList []SMSRecord
	for _, sms := range m.SentSMSs {
		if sms.To == to {
			smsList = append(smsList, sms)
		}
	}
	return smsList
}

// ClearSMSs clears all SMS records
func (m *MockSMSService) ClearSMSs() {
	m.SentSMSs = make([]SMSRecord, 0)
}

// MockStorageService is a mock implementation of storage service
type MockStorageService struct {
	mock.Mock
	StoredFiles map[string]FileRecord
}

// FileRecord represents a stored file record
type FileRecord struct {
	Key      string
	Data     []byte
	Metadata map[string]string
	Time     time.Time
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		StoredFiles: make(map[string]FileRecord),
	}
}

// UploadFile mocks uploading a file
func (m *MockStorageService) UploadFile(ctx context.Context, key string, data []byte, metadata map[string]string) (string, error) {
	args := m.Called(ctx, key, data, metadata)
	url := args.String(0)
	
	// Record the file
	m.StoredFiles[key] = FileRecord{
		Key:      key,
		Data:     data,
		Metadata: metadata,
		Time:     time.Now(),
	}
	
	return url, args.Error(1)
}

// UploadFileFromReader mocks uploading a file from reader
func (m *MockStorageService) UploadFileFromReader(ctx context.Context, key string, reader io.Reader, metadata map[string]string) (string, error) {
	args := m.Called(ctx, key, reader, metadata)
	url := args.String(0)
	
	// Read data from reader
	data, _ := io.ReadAll(reader)
	
	// Record the file
	m.StoredFiles[key] = FileRecord{
		Key:      key,
		Data:     data,
		Metadata: metadata,
		Time:     time.Now(),
	}
	
	return url, args.Error(1)
}

// DownloadFile mocks downloading a file
func (m *MockStorageService) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	
	if file, exists := m.StoredFiles[key]; exists {
		return file.Data, nil
	}
	
	return nil, args.Error(0)
}

// DeleteFile mocks deleting a file
func (m *MockStorageService) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	
	delete(m.StoredFiles, key)
	
	return args.Error(0)
}

// GetFileURL mocks getting a file URL
func (m *MockStorageService) GetFileURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// GetFile returns a stored file by key
func (m *MockStorageService) GetFile(key string) *FileRecord {
	if file, exists := m.StoredFiles[key]; exists {
		return &file
	}
	return nil
}

// ListFiles returns all stored files
func (m *MockStorageService) ListFiles() map[string]FileRecord {
	return m.StoredFiles
}

// ClearFiles clears all stored files
func (m *MockStorageService) ClearFiles() {
	m.StoredFiles = make(map[string]FileRecord)
}

// MockPaymentService is a mock implementation of payment service
type MockPaymentService struct {
	mock.Mock
	Customers    map[string]CustomerRecord
	Subscriptions map[string]SubscriptionRecord
	Payments     map[string]PaymentRecord
}

// CustomerRecord represents a customer record
type CustomerRecord struct {
	ID       string
	Email    string
	Metadata map[string]string
	Time     time.Time
}

// SubscriptionRecord represents a subscription record
type SubscriptionRecord struct {
	ID         string
	CustomerID string
	PlanID     string
	Status     string
	Time       time.Time
}

// PaymentRecord represents a payment record
type PaymentRecord struct {
	ID         string
	CustomerID string
	Amount     int64
	Currency   string
	Status     string
	Time       time.Time
}

// NewMockPaymentService creates a new mock payment service
func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{
		Customers:     make(map[string]CustomerRecord),
		Subscriptions: make(map[string]SubscriptionRecord),
		Payments:      make(map[string]PaymentRecord),
	}
}

// CreateCustomer mocks creating a customer
func (m *MockPaymentService) CreateCustomer(ctx context.Context, email string, metadata map[string]string) (string, error) {
	args := m.Called(ctx, email, metadata)
	customerID := args.String(0)
	
	// Record the customer
	m.Customers[customerID] = CustomerRecord{
		ID:       customerID,
		Email:    email,
		Metadata: metadata,
		Time:     time.Now(),
	}
	
	return customerID, args.Error(1)
}

// CreateSubscription mocks creating a subscription
func (m *MockPaymentService) CreateSubscription(ctx context.Context, customerID, planID string) (string, error) {
	args := m.Called(ctx, customerID, planID)
	subscriptionID := args.String(0)
	
	// Record the subscription
	m.Subscriptions[subscriptionID] = SubscriptionRecord{
		ID:         subscriptionID,
		CustomerID: customerID,
		PlanID:     planID,
		Status:     "active",
		Time:       time.Now(),
	}
	
	return subscriptionID, args.Error(1)
}

// CancelSubscription mocks canceling a subscription
func (m *MockPaymentService) CancelSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	
	if sub, exists := m.Subscriptions[subscriptionID]; exists {
		sub.Status = "canceled"
		m.Subscriptions[subscriptionID] = sub
	}
	
	return args.Error(0)
}

// ProcessPayment mocks processing a payment
func (m *MockPaymentService) ProcessPayment(ctx context.Context, customerID string, amount int64, currency string) (string, error) {
	args := m.Called(ctx, customerID, amount, currency)
	paymentID := args.String(0)
	
	// Record the payment
	m.Payments[paymentID] = PaymentRecord{
		ID:         paymentID,
		CustomerID: customerID,
		Amount:     amount,
		Currency:   currency,
		Status:     "completed",
		Time:       time.Now(),
	}
	
	return paymentID, args.Error(1)
}

// GetCustomer returns a customer by ID
func (m *MockPaymentService) GetCustomer(customerID string) *CustomerRecord {
	if customer, exists := m.Customers[customerID]; exists {
		return &customer
	}
	return nil
}

// GetSubscription returns a subscription by ID
func (m *MockPaymentService) GetSubscription(subscriptionID string) *SubscriptionRecord {
	if sub, exists := m.Subscriptions[subscriptionID]; exists {
		return &sub
	}
	return nil
}

// GetPayment returns a payment by ID
func (m *MockPaymentService) GetPayment(paymentID string) *PaymentRecord {
	if payment, exists := m.Payments[paymentID]; exists {
		return &payment
	}
	return nil
}

// ClearAll clears all records
func (m *MockPaymentService) ClearAll() {
	m.Customers = make(map[string]CustomerRecord)
	m.Subscriptions = make(map[string]SubscriptionRecord)
	m.Payments = make(map[string]PaymentRecord)
}

// MockAIService is a mock implementation of AI service
type MockAIService struct {
	mock.Mock
	Analyses []AnalysisRecord
}

// AnalysisRecord represents an analysis record
type AnalysisRecord struct {
	Input    string
	Result   string
	Category string
	Time     time.Time
}

// NewMockAIService creates a new mock AI service
func NewMockAIService() *MockAIService {
	return &MockAIService{
		Analyses: make([]AnalysisRecord, 0),
	}
}

// AnalyzeContent mocks analyzing content
func (m *MockAIService) AnalyzeContent(ctx context.Context, content string) (string, error) {
	args := m.Called(ctx, content)
	result := args.String(0)
	
	// Record the analysis
	m.Analyses = append(m.Analyses, AnalysisRecord{
		Input:    content,
		Result:   result,
		Category: "content",
		Time:     time.Now(),
	})
	
	return result, args.Error(1)
}

// AnalyzeImage mocks analyzing an image
func (m *MockAIService) AnalyzeImage(ctx context.Context, imageData []byte) (string, error) {
	args := m.Called(ctx, imageData)
	result := args.String(0)
	
	// Record the analysis
	m.Analyses = append(m.Analyses, AnalysisRecord{
		Input:    fmt.Sprintf("image_data_%d", len(imageData)),
		Result:   result,
		Category: "image",
		Time:     time.Now(),
	})
	
	return result, args.Error(1)
}

// ModerateContent mocks moderating content
func (m *MockAIService) ModerateContent(ctx context.Context, content string) (bool, string, error) {
	args := m.Called(ctx, content)
	isApproved := args.Bool(0)
	reason := args.String(1)
	
	// Record the analysis
	m.Analyses = append(m.Analyses, AnalysisRecord{
		Input:    content,
		Result:   fmt.Sprintf("approved:%t,reason:%s", isApproved, reason),
		Category: "moderation",
		Time:     time.Now(),
	})
	
	return isApproved, reason, args.Error(2)
}

// GetLastAnalysis returns the last analysis
func (m *MockAIService) GetLastAnalysis() *AnalysisRecord {
	if len(m.Analyses) == 0 {
		return nil
	}
	return &m.Analyses[len(m.Analyses)-1]
}

// GetAnalysesByCategory returns analyses by category
func (m *MockAIService) GetAnalysesByCategory(category string) []AnalysisRecord {
	var analyses []AnalysisRecord
	for _, analysis := range m.Analyses {
		if analysis.Category == category {
			analyses = append(analyses, analysis)
		}
	}
	return analyses
}

// ClearAnalyses clears all analysis records
func (m *MockAIService) ClearAnalyses() {
	m.Analyses = make([]AnalysisRecord, 0)
}

// MockNotificationService is a mock implementation of notification service
type MockNotificationService struct {
	mock.Mock
	Notifications []NotificationRecord
}

// NotificationRecord represents a notification record
type NotificationRecord struct {
	UserID    uuid.UUID
	Type      string
	Title     string
	Message   string
	Data      map[string]interface{}
	Time      time.Time
}

// NewMockNotificationService creates a new mock notification service
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		Notifications: make([]NotificationRecord, 0),
	}
}

// SendNotification mocks sending a notification
func (m *MockNotificationService) SendNotification(ctx context.Context, userID uuid.UUID, notificationType, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, notificationType, title, message, data)
	
	// Record the notification
	m.Notifications = append(m.Notifications, NotificationRecord{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Data:    data,
		Time:    time.Now(),
	})
	
	return args.Error(0)
}

// SendPushNotification mocks sending a push notification
func (m *MockNotificationService) SendPushNotification(ctx context.Context, userID uuid.UUID, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, title, message, data)
	
	// Record the notification
	m.Notifications = append(m.Notifications, NotificationRecord{
		UserID:  userID,
		Type:    "push",
		Title:   title,
		Message: message,
		Data:    data,
		Time:    time.Now(),
	})
	
	return args.Error(0)
}

// SendEmailNotification mocks sending an email notification
func (m *MockNotificationService) SendEmailNotification(ctx context.Context, userID uuid.UUID, email, subject, body string) error {
	args := m.Called(ctx, userID, email, subject, body)
	
	// Record the notification
	m.Notifications = append(m.Notifications, NotificationRecord{
		UserID:  userID,
		Type:    "email",
		Title:   subject,
		Message: body,
		Data:    map[string]interface{}{"email": email},
		Time:    time.Now(),
	})
	
	return args.Error(0)
}

// GetNotificationsByUser returns notifications for a specific user
func (m *MockNotificationService) GetNotificationsByUser(userID uuid.UUID) []NotificationRecord {
	var notifications []NotificationRecord
	for _, notification := range m.Notifications {
		if notification.UserID == userID {
			notifications = append(notifications, notification)
		}
	}
	return notifications
}

// GetNotificationsByType returns notifications by type
func (m *MockNotificationService) GetNotificationsByType(notificationType string) []NotificationRecord {
	var notifications []NotificationRecord
	for _, notification := range m.Notifications {
		if notification.Type == notificationType {
			notifications = append(notifications, notification)
		}
	}
	return notifications
}

// ClearNotifications clears all notification records
func (m *MockNotificationService) ClearNotifications() {
	m.Notifications = make([]NotificationRecord, 0)
}

// MockFileUploadService is a mock implementation of file upload service
type MockFileUploadService struct {
	mock.Mock
	UploadedFiles map[string]UploadRecord
}

// UploadRecord represents an upload record
type UploadRecord struct {
	ID       string
	Filename string
	Size     int64
	MimeType string
	Data     []byte
	Time     time.Time
}

// NewMockFileUploadService creates a new mock file upload service
func NewMockFileUploadService() *MockFileUploadService {
	return &MockFileUploadService{
		UploadedFiles: make(map[string]UploadRecord),
	}
}

// UploadFile mocks uploading a file
func (m *MockFileUploadService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, file, header)
	fileID := args.String(0)
	
	// Read file data
	data, _ := io.ReadAll(file)
	
	// Record the upload
	m.UploadedFiles[fileID] = UploadRecord{
		ID:       fileID,
		Filename: header.Filename,
		Size:     header.Size,
		MimeType: header.Header.Get("Content-Type"),
		Data:     data,
		Time:     time.Now(),
	}
	
	return fileID, args.Error(1)
}

// GetUpload returns an upload record by ID
func (m *MockFileUploadService) GetUpload(fileID string) *UploadRecord {
	if upload, exists := m.UploadedFiles[fileID]; exists {
		return &upload
	}
	return nil
}

// ListUploads returns all upload records
func (m *MockFileUploadService) ListUploads() map[string]UploadRecord {
	return m.UploadedFiles
}

// ClearUploads clears all upload records
func (m *MockFileUploadService) ClearUploads() {
	m.UploadedFiles = make(map[string]UploadRecord)
}

// MockGeoLocationService is a mock implementation of geolocation service
type MockGeoLocationService struct {
	mock.Mock
	Locations map[string]LocationRecord
}

// LocationRecord represents a location record
type LocationRecord struct {
	Address   string
	Latitude  float64
	Longitude float64
	City      string
	Country   string
	Time      time.Time
}

// NewMockGeoLocationService creates a new mock geolocation service
func NewMockGeoLocationService() *MockGeoLocationService {
	return &MockGeoLocationService{
		Locations: make(map[string]LocationRecord),
	}
}

// GeocodeAddress mocks geocoding an address
func (m *MockGeoLocationService) GeocodeAddress(ctx context.Context, address string) (float64, float64, error) {
	args := m.Called(ctx, address)
	lat := args.Get(0).(float64)
	lng := args.Get(1).(float64)
	
	// Record the location
	m.Locations[address] = LocationRecord{
		Address:   address,
		Latitude:  lat,
		Longitude: lng,
		City:      "Test City",
		Country:   "Test Country",
		Time:      time.Now(),
	}
	
	return lat, lng, args.Error(2)
}

// ReverseGeocode mocks reverse geocoding
func (m *MockGeoLocationService) ReverseGeocode(ctx context.Context, lat, lng float64) (string, error) {
	args := m.Called(ctx, lat, lng)
	address := args.String(0)
	
	// Record the location
	key := fmt.Sprintf("%.6f,%.6f", lat, lng)
	m.Locations[key] = LocationRecord{
		Address:   address,
		Latitude:  lat,
		Longitude: lng,
		City:      "Test City",
		Country:   "Test Country",
		Time:      time.Now(),
	}
	
	return address, args.Error(1)
}

// GetLocation returns a location record by address
func (m *MockGeoLocationService) GetLocation(address string) *LocationRecord {
	if location, exists := m.Locations[address]; exists {
		return &location
	}
	return nil
}

// ClearLocations clears all location records
func (m *MockGeoLocationService) ClearLocations() {
	m.Locations = make(map[string]LocationRecord)
}

// MockServiceManager manages all mock services
type MockServiceManager struct {
	EmailService       *MockEmailService
	SMSService         *MockSMSService
	StorageService     *MockStorageService
	PaymentService     *MockPaymentService
	AIService          *MockAIService
	NotificationService *MockNotificationService
	FileUploadService  *MockFileUploadService
	GeoLocationService *MockGeoLocationService
}

// NewMockServiceManager creates a new mock service manager
func NewMockServiceManager() *MockServiceManager {
	return &MockServiceManager{
		EmailService:        NewMockEmailService(),
		SMSService:          NewMockSMSService(),
		StorageService:      NewMockStorageService(),
		PaymentService:      NewMockPaymentService(),
		AIService:           NewMockAIService(),
		NotificationService: NewMockNotificationService(),
		FileUploadService:   NewMockFileUploadService(),
		GeoLocationService: NewMockGeoLocationService(),
	}
}

// ClearAll clears all mock service records
func (m *MockServiceManager) ClearAll() {
	m.EmailService.ClearEmails()
	m.SMSService.ClearSMSs()
	m.StorageService.ClearFiles()
	m.PaymentService.ClearAll()
	m.AIService.ClearAnalyses()
	m.NotificationService.ClearNotifications()
	m.FileUploadService.ClearUploads()
	m.GeoLocationService.ClearLocations()
}

// SetupDefaultExpectations sets up default expectations for all mock services
func (m *MockServiceManager) SetupDefaultExpectations() {
	// Email service
	m.EmailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.EmailService.On("SendVerificationEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.EmailService.On("SendPasswordResetEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// SMS service
	m.SMSService.On("SendSMS", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.SMSService.On("SendVerificationSMS", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Storage service
	m.StorageService.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("http://example.com/file.jpg", nil)
	m.StorageService.On("UploadFileFromReader", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("http://example.com/file.jpg", nil)
	m.StorageService.On("DownloadFile", mock.Anything, mock.Anything).Return([]byte("test data"), nil)
	m.StorageService.On("DeleteFile", mock.Anything, mock.Anything).Return(nil)
	m.StorageService.On("GetFileURL", mock.Anything, mock.Anything).Return("http://example.com/file.jpg", nil)
	
	// Payment service
	m.PaymentService.On("CreateCustomer", mock.Anything, mock.Anything, mock.Anything).Return("cust_123", nil)
	m.PaymentService.On("CreateSubscription", mock.Anything, mock.Anything, mock.Anything).Return("sub_123", nil)
	m.PaymentService.On("CancelSubscription", mock.Anything, mock.Anything).Return(nil)
	m.PaymentService.On("ProcessPayment", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("pay_123", nil)
	
	// AI service
	m.AIService.On("AnalyzeContent", mock.Anything, mock.Anything).Return("approved", nil)
	m.AIService.On("AnalyzeImage", mock.Anything, mock.Anything).Return("approved", nil)
	m.AIService.On("ModerateContent", mock.Anything, mock.Anything).Return(true, "", nil)
	
	// Notification service
	m.NotificationService.On("SendNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.NotificationService.On("SendPushNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.NotificationService.On("SendEmailNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// File upload service
	m.FileUploadService.On("UploadFile", mock.Anything, mock.Anything, mock.Anything).Return("file_123", nil)
	
	// Geo location service
	m.GeoLocationService.On("GeocodeAddress", mock.Anything, mock.Anything).Return(40.7128, -74.0060, nil)
	m.GeoLocationService.On("ReverseGeocode", mock.Anything, mock.Anything, mock.Anything).Return("New York, NY, USA", nil)
}