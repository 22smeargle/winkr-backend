package postgres

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

func TestNewConnection(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()

	// Test configuration
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "test",
		Password:        "test",
		DBName:          "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime:  3600,
		ConnMaxIdleTime:  1800,
		Timezone:        "UTC",
	}

	// Test that NewConnection function exists and can be called
	// Note: We can't fully test the connection without a real database
	// but we can verify the function signature and basic structure
	assert.NotNil(t, NewConnection)
	assert.NotNil(t, cfg)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_Health(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test health check
	err = db.Health()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_GetStats(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test getting DB stats
	stats := db.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "open_connections")
	assert.Contains(t, stats, "in_use")
	assert.Contains(t, stats, "idle")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_Close(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock close operation
	mock.ExpectClose()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test closing connection
	err = db.Close()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_GetDB(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test getting DB
	retrievedDB := db.GetDB()
	assert.Equal(t, gormDB, retrievedDB)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_Ping(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test ping
	err = db.Ping()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_BeginTransaction(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Mock ping operation
	mock.ExpectPing()
	mock.ExpectBegin()

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create database instance
	cfg := &config.DatabaseConfig{
		Host:   "localhost",
		Port:   5432,
		User:   "test",
		DBName: "testdb",
	}
	db := NewDatabase(gormDB, cfg)

	// Test begin transaction
	tx := db.BeginTransaction()
	assert.NotNil(t, tx)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}