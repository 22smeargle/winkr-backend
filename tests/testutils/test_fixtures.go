package testutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// FixtureLoader handles loading test data fixtures
type FixtureLoader struct {
	db        *sqlx.DB
	fixtures  map[string]interface{}
	factory   *MockFactory
	dataDir   string
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader(db *sqlx.DB, dataDir string) *FixtureLoader {
	return &FixtureLoader{
		db:       db,
		fixtures: make(map[string]interface{}),
		factory:  NewMockFactory(),
		dataDir:  dataDir,
	}
}

// LoadFixtures loads all fixtures from the data directory
func (fl *FixtureLoader) LoadFixtures(ctx context.Context) error {
	if err := fl.loadUserFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load user fixtures: %w", err)
	}

	if err := fl.loadProfileFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load profile fixtures: %w", err)
	}

	if err := fl.loadPhotoFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load photo fixtures: %w", err)
	}

	if err := fl.loadMatchFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load match fixtures: %w", err)
	}

	if err := fl.loadMessageFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load message fixtures: %w", err)
	}

	if err := fl.loadSubscriptionFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load subscription fixtures: %w", err)
	}

	if err := fl.loadVerificationFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load verification fixtures: %w", err)
	}

	if err := fl.loadReportFixtures(ctx); err != nil {
		return fmt.Errorf("failed to load report fixtures: %w", err)
	}

	return nil
}

// loadUserFixtures loads user test data
func (fl *FixtureLoader) loadUserFixtures(ctx context.Context) error {
	users := []*entities.User{
		fl.factory.CreateUser(func(u *entities.User) {
			u.Email = "john.doe@example.com"
			u.FirstName = "John"
			u.LastName = "Doe"
			u.IsVerified = true
		}),
		fl.factory.CreateUser(func(u *entities.User) {
			u.Email = "jane.smith@example.com"
			u.FirstName = "Jane"
			u.LastName = "Smith"
			u.IsVerified = true
		}),
		fl.factory.CreateUser(func(u *entities.User) {
			u.Email = "unverified@example.com"
			u.FirstName = "Unverified"
			u.LastName = "User"
			u.IsVerified = false
		}),
		fl.factory.CreateUser(func(u *entities.User) {
			u.Email = "banned@example.com"
			u.FirstName = "Banned"
			u.LastName = "User"
			u.IsBanned = true
		}),
		fl.factory.CreateUser(func(u *entities.User) {
			u.Email = "premium@example.com"
			u.FirstName = "Premium"
			u.LastName = "User"
			u.IsVerified = true
		}),
	}

	for _, user := range users {
		query := `
			INSERT INTO users (id, email, first_name, last_name, password, is_verified, is_banned, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			user.ID, user.Email, user.FirstName, user.LastName,
			user.Password, user.IsVerified, user.IsBanned,
			user.CreatedAt, user.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %w", user.Email, err)
		}
		fl.fixtures["user_"+user.Email] = user
	}

	return nil
}

// loadProfileFixtures loads profile test data
func (fl *FixtureLoader) loadProfileFixtures(ctx context.Context) error {
	profiles := []*entities.Profile{
		fl.factory.CreateProfile(func(p *entities.Profile) {
			p.UserID = fl.getFixture("user_john.doe@example.com").(*entities.User).ID
			p.Bio = "Software engineer who loves hiking and photography"
			p.Age = 28
			p.Gender = "male"
			p.InterestedIn = []string{"female"}
			p.Location = "San Francisco, CA"
		}),
		fl.factory.CreateProfile(func(p *entities.Profile) {
			p.UserID = fl.getFixture("user_jane.smith@example.com").(*entities.User).ID
			p.Bio = "Artist and yoga enthusiast"
			p.Age = 26
			p.Gender = "female"
			p.InterestedIn = []string{"male"}
			p.Location = "New York, NY"
		}),
		fl.factory.CreateProfile(func(p *entities.Profile) {
			p.UserID = fl.getFixture("user_unverified@example.com").(*entities.User).ID
			p.Bio = "Looking for new friends"
			p.Age = 24
			p.Gender = "female"
			p.InterestedIn = []string{"male", "female"}
			p.Location = "Los Angeles, CA"
		}),
		fl.factory.CreateProfile(func(p *entities.Profile) {
			p.UserID = fl.getFixture("user_premium@example.com").(*entities.User).ID
			p.Bio = "Travel enthusiast and foodie"
			p.Age = 32
			p.Gender = "male"
			p.InterestedIn = []string{"female"}
			p.Location = "Chicago, IL"
		}),
	}

	for _, profile := range profiles {
		interestedInJSON, _ := json.Marshal(profile.InterestedIn)
		query := `
			INSERT INTO profiles (id, user_id, bio, age, gender, interested_in, location, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			profile.ID, profile.UserID, profile.Bio, profile.Age,
			profile.Gender, interestedInJSON, profile.Location,
			profile.CreatedAt, profile.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert profile for user %s: %w", profile.UserID, err)
		}
		fl.fixtures["profile_"+profile.UserID.String()] = profile
	}

	return nil
}

// loadPhotoFixtures loads photo test data
func (fl *FixtureLoader) loadPhotoFixtures(ctx context.Context) error {
	photos := []*entities.Photo{
		fl.factory.CreatePhoto(func(p *entities.Photo) {
			p.UserID = fl.getFixture("user_john.doe@example.com").(*entities.User).ID
			p.URL = "https://example.com/photos/john1.jpg"
			p.IsPrimary = true
			p.IsApproved = true
		}),
		fl.factory.CreatePhoto(func(p *entities.Photo) {
			p.UserID = fl.getFixture("user_john.doe@example.com").(*entities.User).ID
			p.URL = "https://example.com/photos/john2.jpg"
			p.IsPrimary = false
			p.IsApproved = true
		}),
		fl.factory.CreatePhoto(func(p *entities.Photo) {
			p.UserID = fl.getFixture("user_jane.smith@example.com").(*entities.User).ID
			p.URL = "https://example.com/photos/jane1.jpg"
			p.IsPrimary = true
			p.IsApproved = true
		}),
		fl.factory.CreatePhoto(func(p *entities.Photo) {
			p.UserID = fl.getFixture("user_jane.smith@example.com").(*entities.User).ID
			p.URL = "https://example.com/photos/jane2.jpg"
			p.IsPrimary = false
			p.IsApproved = false
		}),
	}

	for _, photo := range photos {
		query := `
			INSERT INTO photos (id, user_id, url, is_primary, is_approved, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			photo.ID, photo.UserID, photo.URL, photo.IsPrimary,
			photo.IsApproved, photo.CreatedAt, photo.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert photo %s: %w", photo.URL, err)
		}
		fl.fixtures["photo_"+photo.URL] = photo
	}

	return nil
}

// loadMatchFixtures loads match test data
func (fl *FixtureLoader) loadMatchFixtures(ctx context.Context) error {
	john := fl.getFixture("user_john.doe@example.com").(*entities.User)
	jane := fl.getFixture("user_jane.smith@example.com").(*entities.User)

	matches := []*entities.Match{
		fl.factory.CreateMatch(func(m *entities.Match) {
			m.UserID = john.ID
			m.MatchedUserID = jane.ID
			m.Status = "matched"
		}),
		fl.factory.CreateMatch(func(m *entities.Match) {
			m.UserID = jane.ID
			m.MatchedUserID = john.ID
			m.Status = "matched"
		}),
	}

	for _, match := range matches {
		query := `
			INSERT INTO matches (id, user_id, matched_user_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			match.ID, match.UserID, match.MatchedUserID,
			match.Status, match.CreatedAt, match.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert match %s: %w", match.ID, err)
		}
		fl.fixtures["match_"+match.UserID.String()+"_"+match.MatchedUserID.String()] = match
	}

	return nil
}

// loadMessageFixtures loads message test data
func (fl *FixtureLoader) loadMessageFixtures(ctx context.Context) error {
	john := fl.getFixture("user_john.doe@example.com").(*entities.User)
	jane := fl.getFixture("user_jane.smith@example.com").(*entities.User)

	messages := []*entities.Message{
		fl.factory.CreateMessage(func(m *entities.Message) {
			m.SenderID = john.ID
			m.ReceiverID = jane.ID
			m.Content = "Hey! How are you doing?"
			m.IsRead = true
		}),
		fl.factory.CreateMessage(func(m *entities.Message) {
			m.SenderID = jane.ID
			m.ReceiverID = john.ID
			m.Content = "I'm doing great! Just finished a yoga class. How about you?"
			m.IsRead = true
		}),
		fl.factory.CreateMessage(func(m *entities.Message) {
			m.SenderID = john.ID
			m.ReceiverID = jane.ID
			m.Content = "That sounds relaxing! I've been working on a new project."
			m.IsRead = false
		}),
	}

	for _, message := range messages {
		query := `
			INSERT INTO messages (id, sender_id, receiver_id, content, is_read, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			message.ID, message.SenderID, message.ReceiverID,
			message.Content, message.IsRead, message.CreatedAt, message.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert message %s: %w", message.ID, err)
		}
		fl.fixtures["message_"+message.ID.String()] = message
	}

	return nil
}

// loadSubscriptionFixtures loads subscription test data
func (fl *FixtureLoader) loadSubscriptionFixtures(ctx context.Context) error {
	premiumUser := fl.getFixture("user_premium@example.com").(*entities.User)

	subscriptions := []*entities.Subscription{
		fl.factory.CreateSubscription(func(s *entities.Subscription) {
			s.UserID = premiumUser.ID
			s.PlanType = "premium"
			s.Status = "active"
			s.StartDate = time.Now().AddDate(0, -1, 0)
			s.EndDate = time.Now().AddDate(0, 1, 0)
		}),
	}

	for _, subscription := range subscriptions {
		query := `
			INSERT INTO subscriptions (id, user_id, plan_type, status, start_date, end_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			subscription.ID, subscription.UserID, subscription.PlanType,
			subscription.Status, subscription.StartDate, subscription.EndDate,
			subscription.CreatedAt, subscription.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert subscription %s: %w", subscription.ID, err)
		}
		fl.fixtures["subscription_"+subscription.UserID.String()] = subscription
	}

	return nil
}

// loadVerificationFixtures loads verification test data
func (fl *FixtureLoader) loadVerificationFixtures(ctx context.Context) error {
	unverifiedUser := fl.getFixture("user_unverified@example.com").(*entities.User)

	verifications := []*entities.Verification{
		fl.factory.CreateVerification(func(v *entities.Verification) {
			v.UserID = unverifiedUser.ID
			v.Type = "document"
			v.Status = "pending"
		}),
		fl.factory.CreateVerification(func(v *entities.Verification) {
			v.UserID = unverifiedUser.ID
			v.Type = "selfie"
			v.Status = "approved"
		}),
	}

	for _, verification := range verifications {
		query := `
			INSERT INTO verifications (id, user_id, type, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			verification.ID, verification.UserID, verification.Type,
			verification.Status, verification.CreatedAt, verification.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert verification %s: %w", verification.ID, err)
		}
		fl.fixtures["verification_"+verification.UserID.String()+"_"+verification.Type] = verification
	}

	return nil
}

// loadReportFixtures loads report test data
func (fl *FixtureLoader) loadReportFixtures(ctx context.Context) error {
	john := fl.getFixture("user_john.doe@example.com").(*entities.User)
	bannedUser := fl.getFixture("user_banned@example.com").(*entities.User)

	reports := []*entities.Report{
		fl.factory.CreateReport(func(r *entities.Report) {
			r.ReporterID = john.ID
			r.ReportedUserID = bannedUser.ID
			r.Reason = "inappropriate_content"
			r.Description = "User sent inappropriate messages"
			r.Status = "resolved"
		}),
	}

	for _, report := range reports {
		query := `
			INSERT INTO reports (id, reporter_id, reported_user_id, reason, description, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`
		_, err := fl.db.ExecContext(ctx, query,
			report.ID, report.ReporterID, report.ReportedUserID,
			report.Reason, report.Description, report.Status,
			report.CreatedAt, report.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert report %s: %w", report.ID, err)
		}
		fl.fixtures["report_"+report.ID.String()] = report
	}

	return nil
}

// getFixture retrieves a fixture by key
func (fl *FixtureLoader) getFixture(key string) interface{} {
	return fl.fixtures[key]
}

// GetFixture retrieves a fixture by key with type assertion
func (fl *FixtureLoader) GetFixture(key string) interface{} {
	return fl.getFixture(key)
}

// GetUserFixture gets a user fixture by email
func (fl *FixtureLoader) GetUserFixture(email string) *entities.User {
	if user, ok := fl.getFixture("user_"+email).(*entities.User); ok {
		return user
	}
	return nil
}

// GetProfileFixture gets a profile fixture by user ID
func (fl *FixtureLoader) GetProfileFixture(userID uuid.UUID) *entities.Profile {
	if profile, ok := fl.getFixture("profile_"+userID.String()).(*entities.Profile); ok {
		return profile
	}
	return nil
}

// GetPhotoFixture gets a photo fixture by URL
func (fl *FixtureLoader) GetPhotoFixture(url string) *entities.Photo {
	if photo, ok := fl.getFixture("photo_"+url).(*entities.Photo); ok {
		return photo
	}
	return nil
}

// GetMatchFixture gets a match fixture by user IDs
func (fl *FixtureLoader) GetMatchFixture(userID, matchedUserID uuid.UUID) *entities.Match {
	key := fmt.Sprintf("match_%s_%s", userID.String(), matchedUserID.String())
	if match, ok := fl.getFixture(key).(*entities.Match); ok {
		return match
	}
	return nil
}

// GetMessageFixture gets a message fixture by ID
func (fl *FixtureLoader) GetMessageFixture(messageID uuid.UUID) *entities.Message {
	if message, ok := fl.getFixture("message_"+messageID.String()).(*entities.Message); ok {
		return message
	}
	return nil
}

// GetSubscriptionFixture gets a subscription fixture by user ID
func (fl *FixtureLoader) GetSubscriptionFixture(userID uuid.UUID) *entities.Subscription {
	if subscription, ok := fl.getFixture("subscription_"+userID.String()).(*entities.Subscription); ok {
		return subscription
	}
	return nil
}

// GetVerificationFixture gets a verification fixture by user ID and type
func (fl *FixtureLoader) GetVerificationFixture(userID uuid.UUID, verificationType string) *entities.Verification {
	key := fmt.Sprintf("verification_%s_%s", userID.String(), verificationType)
	if verification, ok := fl.getFixture(key).(*entities.Verification); ok {
		return verification
	}
	return nil
}

// GetReportFixture gets a report fixture by ID
func (fl *FixtureLoader) GetReportFixture(reportID uuid.UUID) *entities.Report {
	if report, ok := fl.getFixture("report_"+reportID.String()).(*entities.Report); ok {
		return report
	}
	return nil
}

// LoadFixturesFromJSON loads fixtures from JSON files
func (fl *FixtureLoader) LoadFixturesFromJSON(ctx context.Context, filename string) error {
	filePath := filepath.Join(fl.dataDir, filename)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture file %s: %w", filePath, err)
	}

	var fixtures map[string]interface{}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return fmt.Errorf("failed to unmarshal fixture data: %w", err)
	}

	for key, value := range fixtures {
		fl.fixtures[key] = value
	}

	return nil
}

// SaveFixturesToJSON saves current fixtures to JSON file
func (fl *FixtureLoader) SaveFixturesToJSON(filename string) error {
	filePath := filepath.Join(fl.dataDir, filename)
	
	data, err := json.MarshalIndent(fl.fixtures, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal fixture data: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write fixture file %s: %w", filePath, err)
	}

	return nil
}

// ClearFixtures clears all fixtures from memory
func (fl *FixtureLoader) ClearFixtures() {
	fl.fixtures = make(map[string]interface{})
}

// CleanupDatabase cleans up all test data from the database
func (fl *FixtureLoader) CleanupDatabase(ctx context.Context) error {
	tables := []string{
		"reports", "verifications", "subscriptions", "messages",
		"matches", "photos", "profiles", "users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := fl.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}

	return nil
}

// ResetDatabase resets the database and reloads fixtures
func (fl *FixtureLoader) ResetDatabase(ctx context.Context) error {
	if err := fl.CleanupDatabase(ctx); err != nil {
		return fmt.Errorf("failed to cleanup database: %w", err)
	}

	fl.ClearFixtures()

	if err := fl.LoadFixtures(ctx); err != nil {
		return fmt.Errorf("failed to reload fixtures: %w", err)
	}

	return nil
}

// CreateTestDataDirectory creates the test data directory if it doesn't exist
func CreateTestDataDirectory(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create test data directory: %w", err)
	}
	return nil
}

// TestDataManager manages test data operations
type TestDataManager struct {
	fixtureLoader *FixtureLoader
	db            *sqlx.DB
	config        *TestConfig
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager(db *sqlx.DB, config *TestConfig) *TestDataManager {
	dataDir := filepath.Join("tests", "data")
	CreateTestDataDirectory(dataDir)
	
	return &TestDataManager{
		fixtureLoader: NewFixtureLoader(db, dataDir),
		db:            db,
		config:        config,
	}
}

// SetupTestData sets up all test data
func (tdm *TestDataManager) SetupTestData(ctx context.Context) error {
	return tdm.fixtureLoader.LoadFixtures(ctx)
}

// CleanupTestData cleans up all test data
func (tdm *TestDataManager) CleanupTestData(ctx context.Context) error {
	return tdm.fixtureLoader.CleanupDatabase(ctx)
}

// ResetTestData resets all test data
func (tdm *TestDataManager) ResetTestData(ctx context.Context) error {
	return tdm.fixtureLoader.ResetDatabase(ctx)
}

// GetFixtureLoader returns the fixture loader
func (tdm *TestDataManager) GetFixtureLoader() *FixtureLoader {
	return tdm.fixtureLoader
}

// GetDB returns the database connection
func (tdm *TestDataManager) GetDB() *sqlx.DB {
	return tdm.db
}

// GetConfig returns the test configuration
func (tdm *TestDataManager) GetConfig() *TestConfig {
	return tdm.config
}

// WithTransaction executes a function within a database transaction
func (tdm *TestDataManager) WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := tdm.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// ExecuteInIsolation executes a function in complete isolation
func (tdm *TestDataManager) ExecuteInIsolation(ctx context.Context, fn func(context.Context, *TestDataManager) error) error {
	// Create a temporary database for this test
	tempDBName := fmt.Sprintf("winkr_test_%d", time.Now().UnixNano())
	
	// Create temporary database
	if err := tdm.createTempDatabase(ctx, tempDBName); err != nil {
		return fmt.Errorf("failed to create temporary database: %w", err)
	}
	defer tdm.dropTempDatabase(ctx, tempDBName)
	
	// Connect to temporary database
	tempDB, err := tdm.connectToDatabase(ctx, tempDBName)
	if err != nil {
		return fmt.Errorf("failed to connect to temporary database: %w", err)
	}
	defer tempDB.Close()
	
	// Create temporary data manager
	tempTDM := NewTestDataManager(tempDB, tdm.config)
	
	// Setup test data
	if err := tempTDM.SetupTestData(ctx); err != nil {
		return fmt.Errorf("failed to setup test data: %w", err)
	}
	
	// Execute the function
	return fn(ctx, tempTDM)
}

// createTempDatabase creates a temporary database
func (tdm *TestDataManager) createTempDatabase(ctx context.Context, dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := tdm.db.ExecContext(ctx, query)
	return err
}

// dropTempDatabase drops a temporary database
func (tdm *TestDataManager) dropTempDatabase(ctx context.Context, dbName string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := tdm.db.ExecContext(ctx, query)
	return err
}

// connectToDatabase connects to a specific database
func (tdm *TestDataManager) connectToDatabase(ctx context.Context, dbName string) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		tdm.config.Database.Host, tdm.config.Database.Port,
		tdm.config.Database.User, tdm.config.Database.Password,
		dbName, tdm.config.Database.SSLMode)
	
	return sqlx.ConnectContext(ctx, "postgres", connStr)
}