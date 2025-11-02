package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestDatabase holds test database connection and resources
type TestDatabase struct {
	DB     *gorm.DB
	Mock   sqlmock.Sqlmock
	Cleanup func()
}

// SetupTestDatabase creates a test database with mock
func SetupTestDatabase(t *testing.T) *TestDatabase {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate all entities
	err = gormDB.AutoMigrate(
		&entities.User{},
		&entities.Photo{},
		&entities.Match{},
		&entities.Message{},
		&entities.Report{},
		&entities.Subscription{},
		&entities.Verification{},
		&entities.Payment{},
		&entities.EphemeralPhoto{},
	)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			sqlDB.Close()
		}
		mock.ExpectClose()
	}

	return &TestDatabase{
		DB:     gormDB,
		Mock:   mock,
		Cleanup: cleanup,
	}
}

// SetupTestDatabaseWithDocker creates a test database using Docker
func SetupTestDatabaseWithDocker(t *testing.T) *TestDatabase {
	// Load .env file if it exists
	_ = godotenv.Load("../../.env")

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "Could not construct pool")

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	err = pool.Client.Ping()
	require.NoError(t, err, "Could not connect to Docker")

	// pull postgres docker image
	resource, err := pool.RunWith(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=testpassword",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=testdb",
			"listen_addresses = '*'",
		},
	})
	require.NoError(t, err, "Could not start resource")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	var gormDB *gorm.DB
	if err := pool.Retry(func() error {
		var err error
		dsn := fmt.Sprintf("postgres://testuser:testpassword@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// Auto migrate all entities
	err = gormDB.AutoMigrate(
		&entities.User{},
		&entities.Photo{},
		&entities.Match{},
		&entities.Message{},
		&entities.Report{},
		&entities.Subscription{},
		&entities.Verification{},
		&entities.Payment{},
		&entities.EphemeralPhoto{},
	)
	require.NoError(t, err)

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
		sqlDB, err := gormDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return &TestDatabase{
		DB:     gormDB,
		Cleanup: cleanup,
	}
}

// SetupTestDatabaseFromConfig creates a test database from configuration
func SetupTestDatabaseFromConfig(t *testing.T, cfg config.DatabaseConfig) *TestDatabase {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
		cfg.Timezone,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate all entities
	err = gormDB.AutoMigrate(
		&entities.User{},
		&entities.Photo{},
		&entities.Match{},
		&entities.Message{},
		&entities.Report{},
		&entities.Subscription{},
		&entities.Verification{},
		&entities.Payment{},
		&entities.EphemeralPhoto{},
	)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return &TestDatabase{
		DB:     gormDB,
		Cleanup: cleanup,
	}
}

// CleanupTestDatabase cleans up the test database
func CleanupTestDatabase(db *TestDatabase) {
	if db.Cleanup != nil {
		db.Cleanup()
	}
}

// TruncateAllTables truncates all tables in the test database
func TruncateAllTables(t *testing.T, db *gorm.DB) {
	tables := []string{
		"ephemeral_photos",
		"payments",
		"verifications",
		"subscriptions",
		"reports",
		"messages",
		"matches",
		"photos",
		"users",
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error
		require.NoError(t, err, "Failed to truncate table %s", table)
	}
}

// SeedTestData seeds the test database with test data
func SeedTestData(t *testing.T, db *gorm.DB) map[string]interface{} {
	// Create test users
	users := []*entities.User{
		{
			ID:        "user1",
			Email:     "user1@example.com",
			Password:  "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjdIrEjQ2Q8J2K2vJ2K2vJ2K2vJ2K2",
			FirstName: "John",
			LastName:  "Doe",
			IsActive: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user2",
			Email:     "user2@example.com",
			Password:  "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjdIrEjQ2Q8J2K2vJ2K2vJ2K2vJ2K2",
			FirstName: "Jane",
			LastName:  "Smith",
			IsActive: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, user := range users {
		err := db.Create(user).Error
		require.NoError(t, err)
	}

	// Create test photos
	photos := []*entities.Photo{
		{
			ID:         "photo1",
			UserID:     "user1",
			URL:        "https://example.com/photo1.jpg",
			IsPrimary:  true,
			IsApproved: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "photo2",
			UserID:     "user2",
			URL:        "https://example.com/photo2.jpg",
			IsPrimary:  true,
			IsApproved: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, photo := range photos {
		err := db.Create(photo).Error
		require.NoError(t, err)
	}

	// Create test matches
	matches := []*entities.Match{
		{
			ID:        "match1",
			UserID1:   "user1",
			UserID2:   "user2",
			MatchedAt:  time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, match := range matches {
		err := db.Create(match).Error
		require.NoError(t, err)
	}

	// Create test messages
	messages := []*entities.Message{
		{
			ID:             "message1",
			ConversationID:  "conv1",
			SenderID:       "user1",
			ReceiverID:     "user2",
			Content:        "Hello!",
			MessageType:    entities.MessageTypeText,
			IsRead:         false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, message := range messages {
		err := db.Create(message).Error
		require.NoError(t, err)
	}

	return map[string]interface{}{
		"users":    users,
		"photos":   photos,
		"matches":  matches,
		"messages": messages,
	}
}

// WaitForDatabase waits for the database to be ready
func WaitForDatabase(cfg config.DatabaseConfig, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			"postgres", // Connect to default database first
			cfg.SSLMode,
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			sqlDB, err := db.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					sqlDB.Close()
					return nil
				}
				sqlDB.Close()
			}
		}

		time.Sleep(time.Second * 2)
	}

	return fmt.Errorf("database not ready after %d retries", maxRetries)
}

// CreateTestDatabase creates a test database for integration tests
func CreateTestDatabase(t *testing.T) *gorm.DB {
	// Check if we're running in CI/CD
	if os.Getenv("CI") == "true" {
		return setupCITestDatabase(t)
	}

	// Check if Docker is available
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return setupDockerTestDatabase(t)
	}

	// Fallback to mock database
	testDB := SetupTestDatabase(t)
	return testDB.DB
}

func setupCITestDatabase(t *testing.T) *gorm.DB {
	cfg := config.DatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("TEST_DB_USER", "test_user"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "test_password"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "test_db"),
		SSLMode:  "disable",
		Timezone:  "UTC",
	}

	testDB := SetupTestDatabaseFromConfig(t, cfg)
	return testDB.DB
}

func setupDockerTestDatabase(t *testing.T) *gorm.DB {
	testDB := SetupTestDatabaseWithDocker(t)
	return testDB.DB
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestTransaction wraps operations in a database transaction for testing
func TestTransaction(t *testing.T, db *gorm.DB, fn func(*gorm.DB) error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	err := fn(tx)
	if err != nil {
		tx.Rollback()
		require.NoError(t, err)
	} else {
		err = tx.Commit().Error
		require.NoError(t, err)
	}
}

// AssertTableCount asserts that a table has the expected number of records
func AssertTableCount(t *testing.T, db *gorm.DB, tableName string, expected int) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(expected), count, "Table %s should have %d records", tableName, expected)
}

// AssertRecordExists asserts that a record exists in the database
func AssertRecordExists(t *testing.T, db *gorm.DB, model interface{}, conditions map[string]interface{}) {
	var count int64
	err := db.Model(model).Where(conditions).Count(&count).Error
	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Record should exist")
}

// AssertRecordNotExists asserts that a record does not exist in the database
func AssertRecordNotExists(t *testing.T, db *gorm.DB, model interface{}, conditions map[string]interface{}) {
	var count int64
	err := db.Model(model).Where(conditions).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Record should not exist")
}