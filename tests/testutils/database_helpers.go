package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// DatabaseHelper provides utilities for database testing
type DatabaseHelper struct {
	t      *testing.T
	db     *sqlx.DB
	config *TestConfig
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(t *testing.T, db *sqlx.DB, config *TestConfig) *DatabaseHelper {
	return &DatabaseHelper{
		t:      t,
		db:     db,
		config: config,
	}
}

// CleanupTable cleans up all data from a table
func (dh *DatabaseHelper) CleanupTable(tableName string) {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	_, err := dh.db.Exec(query)
	require.NoError(dh.t, err, "Failed to cleanup table %s", tableName)
}

// CleanupAllTables cleans up all test tables
func (dh *DatabaseHelper) CleanupAllTables() {
	tables := []string{
		"reports", "verifications", "subscriptions", "messages",
		"matches", "photos", "profiles", "users",
	}

	for _, table := range tables {
		dh.CleanupTable(table)
	}
}

// TruncateTable truncates a table and resets identity
func (dh *DatabaseHelper) TruncateTable(tableName string) {
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	_, err := dh.db.Exec(query)
	require.NoError(dh.t, err, "Failed to truncate table %s", tableName)
}

// TruncateAllTables truncates all test tables
func (dh *DatabaseHelper) TruncateAllTables() {
	tables := []string{
		"reports", "verifications", "subscriptions", "messages",
		"matches", "photos", "profiles", "users",
	}

	for _, table := range tables {
		dh.TruncateTable(table)
	}
}

// CountRows returns the number of rows in a table
func (dh *DatabaseHelper) CountRows(tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := dh.db.Get(&count, query)
	require.NoError(dh.t, err, "Failed to count rows in table %s", tableName)
	return count
}

// TableExists checks if a table exists
func (dh *DatabaseHelper) TableExists(tableName string) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`
	err := dh.db.Get(&exists, query, tableName)
	require.NoError(dh.t, err, "Failed to check if table %s exists", tableName)
	return exists
}

// WaitForTable waits for a table to exist
func (dh *DatabaseHelper) WaitForTable(tableName string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if dh.TableExists(tableName) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.Fail(dh.t, "Table %s did not appear within timeout", tableName)
}

// InsertUser inserts a test user
func (dh *DatabaseHelper) InsertUser(user *entities.User) {
	query := `
		INSERT INTO users (id, email, first_name, last_name, password, is_verified, is_banned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			password = EXCLUDED.password,
			is_verified = EXCLUDED.is_verified,
			is_banned = EXCLUDED.is_banned,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		user.ID, user.Email, user.FirstName, user.LastName,
		user.Password, user.IsVerified, user.IsBanned,
		user.CreatedAt, user.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert user")
}

// InsertProfile inserts a test profile
func (dh *DatabaseHelper) InsertProfile(profile *entities.Profile) {
	query := `
		INSERT INTO profiles (id, user_id, bio, age, gender, interested_in, location, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bio = EXCLUDED.bio,
			age = EXCLUDED.age,
			gender = EXCLUDED.gender,
			interested_in = EXCLUDED.interested_in,
			location = EXCLUDED.location,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		profile.ID, profile.UserID, profile.Bio, profile.Age,
		profile.Gender, profile.InterestedIn, profile.Location,
		profile.CreatedAt, profile.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert profile")
}

// InsertPhoto inserts a test photo
func (dh *DatabaseHelper) InsertPhoto(photo *entities.Photo) {
	query := `
		INSERT INTO photos (id, user_id, url, is_primary, is_approved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			url = EXCLUDED.url,
			is_primary = EXCLUDED.is_primary,
			is_approved = EXCLUDED.is_approved,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		photo.ID, photo.UserID, photo.URL, photo.IsPrimary,
		photo.IsApproved, photo.CreatedAt, photo.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert photo")
}

// InsertMatch inserts a test match
func (dh *DatabaseHelper) InsertMatch(match *entities.Match) {
	query := `
		INSERT INTO matches (id, user_id, matched_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			matched_user_id = EXCLUDED.matched_user_id,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		match.ID, match.UserID, match.MatchedUserID,
		match.Status, match.CreatedAt, match.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert match")
}

// InsertMessage inserts a test message
func (dh *DatabaseHelper) InsertMessage(message *entities.Message) {
	query := `
		INSERT INTO messages (id, sender_id, receiver_id, content, is_read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			sender_id = EXCLUDED.sender_id,
			receiver_id = EXCLUDED.receiver_id,
			content = EXCLUDED.content,
			is_read = EXCLUDED.is_read,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		message.ID, message.SenderID, message.ReceiverID,
		message.Content, message.IsRead, message.CreatedAt, message.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert message")
}

// InsertSubscription inserts a test subscription
func (dh *DatabaseHelper) InsertSubscription(subscription *entities.Subscription) {
	query := `
		INSERT INTO subscriptions (id, user_id, plan_type, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			plan_type = EXCLUDED.plan_type,
			status = EXCLUDED.status,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		subscription.ID, subscription.UserID, subscription.PlanType,
		subscription.Status, subscription.StartDate, subscription.EndDate,
		subscription.CreatedAt, subscription.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert subscription")
}

// InsertVerification inserts a test verification
func (dh *DatabaseHelper) InsertVerification(verification *entities.Verification) {
	query := `
		INSERT INTO verifications (id, user_id, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		verification.ID, verification.UserID, verification.Type,
		verification.Status, verification.CreatedAt, verification.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert verification")
}

// InsertReport inserts a test report
func (dh *DatabaseHelper) InsertReport(report *entities.Report) {
	query := `
		INSERT INTO reports (id, reporter_id, reported_user_id, reason, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			reporter_id = EXCLUDED.reporter_id,
			reported_user_id = EXCLUDED.reported_user_id,
			reason = EXCLUDED.reason,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := dh.db.Exec(query,
		report.ID, report.ReporterID, report.ReportedUserID,
		report.Reason, report.Description, report.Status,
		report.CreatedAt, report.UpdatedAt,
	)
	require.NoError(dh.t, err, "Failed to insert report")
}

// GetUserByID gets a user by ID
func (dh *DatabaseHelper) GetUserByID(id string) *entities.User {
	var user entities.User
	query := `SELECT id, email, first_name, last_name, password, is_verified, is_banned, created_at, updated_at FROM users WHERE id = $1`
	err := dh.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil
	}
	require.NoError(dh.t, err, "Failed to get user by ID")
	return &user
}

// GetUserByEmail gets a user by email
func (dh *DatabaseHelper) GetUserByEmail(email string) *entities.User {
	var user entities.User
	query := `SELECT id, email, first_name, last_name, password, is_verified, is_banned, created_at, updated_at FROM users WHERE email = $1`
	err := dh.db.Get(&user, query, email)
	if err == sql.ErrNoRows {
		return nil
	}
	require.NoError(dh.t, err, "Failed to get user by email")
	return &user
}

// GetProfileByUserID gets a profile by user ID
func (dh *DatabaseHelper) GetProfileByUserID(userID string) *entities.Profile {
	var profile entities.Profile
	query := `SELECT id, user_id, bio, age, gender, interested_in, location, created_at, updated_at FROM profiles WHERE user_id = $1`
	err := dh.db.Get(&profile, query, userID)
	if err == sql.ErrNoRows {
		return nil
	}
	require.NoError(dh.t, err, "Failed to get profile by user ID")
	return &profile
}

// GetPhotosByUserID gets photos by user ID
func (dh *DatabaseHelper) GetPhotosByUserID(userID string) []*entities.Photo {
	var photos []*entities.Photo
	query := `SELECT id, user_id, url, is_primary, is_approved, created_at, updated_at FROM photos WHERE user_id = $1 ORDER BY created_at`
	err := dh.db.Select(&photos, query, userID)
	require.NoError(dh.t, err, "Failed to get photos by user ID")
	return photos
}

// GetMatchesByUserID gets matches by user ID
func (dh *DatabaseHelper) GetMatchesByUserID(userID string) []*entities.Match {
	var matches []*entities.Match
	query := `SELECT id, user_id, matched_user_id, status, created_at, updated_at FROM matches WHERE user_id = $1 ORDER BY created_at`
	err := dh.db.Select(&matches, query, userID)
	require.NoError(dh.t, err, "Failed to get matches by user ID")
	return matches
}

// GetMessagesBySenderID gets messages by sender ID
func (dh *DatabaseHelper) GetMessagesBySenderID(senderID string) []*entities.Message {
	var messages []*entities.Message
	query := `SELECT id, sender_id, receiver_id, content, is_read, created_at, updated_at FROM messages WHERE sender_id = $1 ORDER BY created_at`
	err := dh.db.Select(&messages, query, senderID)
	require.NoError(dh.t, err, "Failed to get messages by sender ID")
	return messages
}

// GetMessagesByReceiverID gets messages by receiver ID
func (dh *DatabaseHelper) GetMessagesByReceiverID(receiverID string) []*entities.Message {
	var messages []*entities.Message
	query := `SELECT id, sender_id, receiver_id, content, is_read, created_at, updated_at FROM messages WHERE receiver_id = $1 ORDER BY created_at`
	err := dh.db.Select(&messages, query, receiverID)
	require.NoError(dh.t, err, "Failed to get messages by receiver ID")
	return messages
}

// GetSubscriptionByUserID gets subscription by user ID
func (dh *DatabaseHelper) GetSubscriptionByUserID(userID string) *entities.Subscription {
	var subscription entities.Subscription
	query := `SELECT id, user_id, plan_type, status, start_date, end_date, created_at, updated_at FROM subscriptions WHERE user_id = $1`
	err := dh.db.Get(&subscription, query, userID)
	if err == sql.ErrNoRows {
		return nil
	}
	require.NoError(dh.t, err, "Failed to get subscription by user ID")
	return &subscription
}

// GetVerificationsByUserID gets verifications by user ID
func (dh *DatabaseHelper) GetVerificationsByUserID(userID string) []*entities.Verification {
	var verifications []*entities.Verification
	query := `SELECT id, user_id, type, status, created_at, updated_at FROM verifications WHERE user_id = $1 ORDER BY created_at`
	err := dh.db.Select(&verifications, query, userID)
	require.NoError(dh.t, err, "Failed to get verifications by user ID")
	return verifications
}

// GetReportsByReporterID gets reports by reporter ID
func (dh *DatabaseHelper) GetReportsByReporterID(reporterID string) []*entities.Report {
	var reports []*entities.Report
	query := `SELECT id, reporter_id, reported_user_id, reason, description, status, created_at, updated_at FROM reports WHERE reporter_id = $1 ORDER BY created_at`
	err := dh.db.Select(&reports, query, reporterID)
	require.NoError(dh.t, err, "Failed to get reports by reporter ID")
	return reports
}

// GetReportsByReportedUserID gets reports by reported user ID
func (dh *DatabaseHelper) GetReportsByReportedUserID(reportedUserID string) []*entities.Report {
	var reports []*entities.Report
	query := `SELECT id, reporter_id, reported_user_id, reason, description, status, created_at, updated_at FROM reports WHERE reported_user_id = $1 ORDER BY created_at`
	err := dh.db.Select(&reports, query, reportedUserID)
	require.NoError(dh.t, err, "Failed to get reports by reported user ID")
	return reports
}

// AssertUserExists asserts that a user exists
func (dh *DatabaseHelper) AssertUserExists(userID string) {
	user := dh.GetUserByID(userID)
	require.NotNil(dh.t, user, "User should exist")
}

// AssertUserNotExists asserts that a user does not exist
func (dh *DatabaseHelper) AssertUserNotExists(userID string) {
	user := dh.GetUserByID(userID)
	require.Nil(dh.t, user, "User should not exist")
}

// AssertProfileExists asserts that a profile exists
func (dh *DatabaseHelper) AssertProfileExists(userID string) {
	profile := dh.GetProfileByUserID(userID)
	require.NotNil(dh.t, profile, "Profile should exist")
}

// AssertProfileNotExists asserts that a profile does not exist
func (dh *DatabaseHelper) AssertProfileNotExists(userID string) {
	profile := dh.GetProfileByUserID(userID)
	require.Nil(dh.t, profile, "Profile should not exist")
}

// AssertPhotoExists asserts that a photo exists
func (dh *DatabaseHelper) AssertPhotoExists(photoID string) {
	var count int
	query := `SELECT COUNT(*) FROM photos WHERE id = $1`
	err := dh.db.Get(&count, query, photoID)
	require.NoError(dh.t, err, "Failed to check photo existence")
	require.Greater(dh.t, count, 0, "Photo should exist")
}

// AssertPhotoNotExists asserts that a photo does not exist
func (dh *DatabaseHelper) AssertPhotoNotExists(photoID string) {
	var count int
	query := `SELECT COUNT(*) FROM photos WHERE id = $1`
	err := dh.db.Get(&count, query, photoID)
	require.NoError(dh.t, err, "Failed to check photo existence")
	require.Equal(dh.t, 0, count, "Photo should not exist")
}

// AssertMatchExists asserts that a match exists
func (dh *DatabaseHelper) AssertMatchExists(userID, matchedUserID string) {
	var count int
	query := `SELECT COUNT(*) FROM matches WHERE user_id = $1 AND matched_user_id = $2`
	err := dh.db.Get(&count, query, userID, matchedUserID)
	require.NoError(dh.t, err, "Failed to check match existence")
	require.Greater(dh.t, count, 0, "Match should exist")
}

// AssertMatchNotExists asserts that a match does not exist
func (dh *DatabaseHelper) AssertMatchNotExists(userID, matchedUserID string) {
	var count int
	query := `SELECT COUNT(*) FROM matches WHERE user_id = $1 AND matched_user_id = $2`
	err := dh.db.Get(&count, query, userID, matchedUserID)
	require.NoError(dh.t, err, "Failed to check match existence")
	require.Equal(dh.t, 0, count, "Match should not exist")
}

// AssertMessageExists asserts that a message exists
func (dh *DatabaseHelper) AssertMessageExists(messageID string) {
	var count int
	query := `SELECT COUNT(*) FROM messages WHERE id = $1`
	err := dh.db.Get(&count, query, messageID)
	require.NoError(dh.t, err, "Failed to check message existence")
	require.Greater(dh.t, count, 0, "Message should exist")
}

// AssertMessageNotExists asserts that a message does not exist
func (dh *DatabaseHelper) AssertMessageNotExists(messageID string) {
	var count int
	query := `SELECT COUNT(*) FROM messages WHERE id = $1`
	err := dh.db.Get(&count, query, messageID)
	require.NoError(dh.t, err, "Failed to check message existence")
	require.Equal(dh.t, 0, count, "Message should not exist")
}

// AssertSubscriptionExists asserts that a subscription exists
func (dh *DatabaseHelper) AssertSubscriptionExists(userID string) {
	subscription := dh.GetSubscriptionByUserID(userID)
	require.NotNil(dh.t, subscription, "Subscription should exist")
}

// AssertSubscriptionNotExists asserts that a subscription does not exist
func (dh *DatabaseHelper) AssertSubscriptionNotExists(userID string) {
	subscription := dh.GetSubscriptionByUserID(userID)
	require.Nil(dh.t, subscription, "Subscription should not exist")
}

// AssertVerificationExists asserts that a verification exists
func (dh *DatabaseHelper) AssertVerificationExists(userID, verificationType string) {
	var count int
	query := `SELECT COUNT(*) FROM verifications WHERE user_id = $1 AND type = $2`
	err := dh.db.Get(&count, query, userID, verificationType)
	require.NoError(dh.t, err, "Failed to check verification existence")
	require.Greater(dh.t, count, 0, "Verification should exist")
}

// AssertVerificationNotExists asserts that a verification does not exist
func (dh *DatabaseHelper) AssertVerificationNotExists(userID, verificationType string) {
	var count int
	query := `SELECT COUNT(*) FROM verifications WHERE user_id = $1 AND type = $2`
	err := dh.db.Get(&count, query, userID, verificationType)
	require.NoError(dh.t, err, "Failed to check verification existence")
	require.Equal(dh.t, 0, count, "Verification should not exist")
}

// AssertReportExists asserts that a report exists
func (dh *DatabaseHelper) AssertReportExists(reportID string) {
	var count int
	query := `SELECT COUNT(*) FROM reports WHERE id = $1`
	err := dh.db.Get(&count, query, reportID)
	require.NoError(dh.t, err, "Failed to check report existence")
	require.Greater(dh.t, count, 0, "Report should exist")
}

// AssertReportNotExists asserts that a report does not exist
func (dh *DatabaseHelper) AssertReportNotExists(reportID string) {
	var count int
	query := `SELECT COUNT(*) FROM reports WHERE id = $1`
	err := dh.db.Get(&count, query, reportID)
	require.NoError(dh.t, err, "Failed to check report existence")
	require.Equal(dh.t, 0, count, "Report should not exist")
}

// AssertTableRowCount asserts that a table has a specific number of rows
func (dh *DatabaseHelper) AssertTableRowCount(tableName string, expectedCount int) {
	actualCount := dh.CountRows(tableName)
	require.Equal(dh.t, expectedCount, actualCount, "Table %s should have %d rows", tableName, expectedCount)
}

// AssertTableNotEmpty asserts that a table is not empty
func (dh *DatabaseHelper) AssertTableNotEmpty(tableName string) {
	count := dh.CountRows(tableName)
	require.Greater(dh.t, count, 0, "Table %s should not be empty", tableName)
}

// AssertTableEmpty asserts that a table is empty
func (dh *DatabaseHelper) AssertTableEmpty(tableName string) {
	count := dh.CountRows(tableName)
	require.Equal(dh.t, 0, count, "Table %s should be empty", tableName)
}

// WaitForData waits for data to appear in a table
func (dh *DatabaseHelper) WaitForData(tableName string, minRows int, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if dh.CountRows(tableName) >= minRows {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.Fail(dh.t, "Table %s did not contain at least %d rows within timeout", tableName, minRows)
}

// ExecuteQuery executes a custom query
func (dh *DatabaseHelper) ExecuteQuery(query string, args ...interface{}) {
	_, err := dh.db.Exec(query, args...)
	require.NoError(dh.t, err, "Failed to execute query: %s", query)
}

// ExecuteQueryRow executes a query and returns a single row
func (dh *DatabaseHelper) ExecuteQueryRow(dest interface{}, query string, args ...interface{}) {
	err := dh.db.Get(dest, query, args...)
	require.NoError(dh.t, err, "Failed to execute query row: %s", query)
}

// ExecuteQueryRows executes a query and returns multiple rows
func (dh *DatabaseHelper) ExecuteQueryRows(dest interface{}, query string, args ...interface{}) {
	err := dh.db.Select(dest, query, args...)
	require.NoError(dh.t, err, "Failed to execute query rows: %s", query)
}

// BeginTransaction begins a database transaction
func (dh *DatabaseHelper) BeginTransaction() *sqlx.Tx {
	tx, err := dh.db.Beginx()
	require.NoError(dh.t, err, "Failed to begin transaction")
	return tx
}

// WithTransaction executes a function within a transaction
func (dh *DatabaseHelper) WithTransaction(fn func(*sqlx.Tx) error) {
	tx := dh.BeginTransaction()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	
	err := fn(tx)
	if err != nil {
		tx.Rollback()
		require.NoError(dh.t, err, "Transaction failed")
	}
	
	err = tx.Commit()
	require.NoError(dh.t, err, "Failed to commit transaction")
}

// GetDB returns the database connection
func (dh *DatabaseHelper) GetDB() *sqlx.DB {
	return dh.db
}

// Close closes the database connection
func (dh *DatabaseHelper) Close() error {
	return dh.db.Close()
}