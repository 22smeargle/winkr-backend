package testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// TestDatabase provides test database setup and teardown
type TestDatabase struct {
	DB     *sqlx.DB
	Config DatabaseConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// NewTestDatabase creates a new test database instance
func NewTestDatabase(config DatabaseConfig) *TestDatabase {
	return &TestDatabase{
		Config: config,
	}
}

// Setup sets up the test database
func (td *TestDatabase) Setup() error {
	// Connect to PostgreSQL server (without specifying database)
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s",
		td.Config.Host, td.Config.Port, td.Config.User, td.Config.Password, td.Config.SSLMode)
	
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	td.DB = db
	
	// Create test database if it doesn't exist
	if err := td.createTestDatabase(); err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}
	
	// Close the connection to the server
	db.Close()
	
	// Connect to the test database
	testDBConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		td.Config.Host, td.Config.Port, td.Config.User, td.Config.Password, td.Config.Name, td.Config.SSLMode)
	
	testDB, err := sqlx.Connect("postgres", testDBConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}
	td.DB = testDB
	
	// Run migrations
	if err := td.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	return nil
}

// createTestDatabase creates the test database
func (td *TestDatabase) createTestDatabase() error {
	query := fmt.Sprintf("CREATE DATABASE %s", td.Config.Name)
	_, err := td.DB.Exec(query)
	if err != nil {
		// Check if database already exists
		if !isDatabaseExistsError(err) {
			return err
		}
		log.Printf("Database %s already exists", td.Config.Name)
	}
	return nil
}

// isDatabaseExistsError checks if the error is because database already exists
func isDatabaseExistsError(err error) bool {
	return err != nil && (err.Error() == `pq: database "`+td.Config.Name+`" already exists`)
}

// runMigrations runs database migrations
func (td *TestDatabase) runMigrations() error {
	// This would typically read migration files from a directory
	// For now, we'll create basic tables needed for testing
	
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			password VARCHAR(255) NOT NULL,
			is_verified BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS profiles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			bio TEXT,
			age INTEGER,
			gender VARCHAR(20),
			interested_in TEXT[],
			relationship_status VARCHAR(20),
			location VARCHAR(255),
			photos TEXT[],
			is_verified BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS matches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id1 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			user_id2 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			matched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id1, user_id2)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
			sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS photos (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			url VARCHAR(500) NOT NULL,
			is_primary BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			reported_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			reason VARCHAR(100) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			plan_type VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ephemeral_photos (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			url VARCHAR(500) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			view_count INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS verifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			amount DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
	}
	
	for _, migration := range migrations {
		if _, err := td.DB.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}
	
	return nil
}

// Teardown tears down the test database
func (td *TestDatabase) Teardown() error {
	if td.DB != nil {
		// Close database connection
		if err := td.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}

// Cleanup cleans up all data from the test database
func (td *TestDatabase) Cleanup() error {
	tables := []string{
		"payments",
		"verifications",
		"ephemeral_photos",
		"subscriptions",
		"reports",
		"photos",
		"messages",
		"matches",
		"profiles",
		"users",
	}
	
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := td.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}
	
	return nil
}

// GetConnection returns the database connection
func (td *TestDatabase) GetConnection() *sqlx.DB {
	return td.DB
}

// TestRedis provides test Redis setup and teardown
type TestRedis struct {
	Client *redis.Client
	Config RedisConfig
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewTestRedis creates a new test Redis instance
func NewTestRedis(config RedisConfig) *TestRedis {
	return &TestRedis{
		Config: config,
	}
}

// Setup sets up the test Redis connection
func (tr *TestRedis) Setup() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", tr.Config.Host, tr.Config.Port),
		Password: tr.Config.Password,
		DB:       tr.Config.DB,
	})
	
	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	tr.Client = rdb
	return nil
}

// Teardown tears down the test Redis connection
func (tr *TestRedis) Teardown() error {
	if tr.Client != nil {
		// Close Redis connection
		if err := tr.Client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %w", err)
		}
	}
	return nil
}

// Cleanup cleans up all data from the test Redis database
func (tr *TestRedis) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := tr.Client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush Redis database: %w", err)
	}
	
	return nil
}

// GetClient returns the Redis client
func (tr *TestRedis) GetClient() *redis.Client {
	return tr.Client
}

// TestEnvironment provides a complete test environment with database and Redis
type TestEnvironment struct {
	Database *TestDatabase
	Redis    *TestRedis
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(dbConfig DatabaseConfig, redisConfig RedisConfig) *TestEnvironment {
	return &TestEnvironment{
		Database: NewTestDatabase(dbConfig),
		Redis:    NewTestRedis(redisConfig),
	}
}

// Setup sets up the complete test environment
func (te *TestEnvironment) Setup() error {
	if err := te.Database.Setup(); err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	
	if err := te.Redis.Setup(); err != nil {
		return fmt.Errorf("failed to setup Redis: %w", err)
	}
	
	return nil
}

// Teardown tears down the complete test environment
func (te *TestEnvironment) Teardown() error {
	var errors []error
	
	if err := te.Database.Teardown(); err != nil {
		errors = append(errors, fmt.Errorf("database teardown error: %w", err))
	}
	
	if err := te.Redis.Teardown(); err != nil {
		errors = append(errors, fmt.Errorf("Redis teardown error: %w", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("multiple teardown errors: %v", errors)
	}
	
	return nil
}

// Cleanup cleans up all data from the test environment
func (te *TestEnvironment) Cleanup() error {
	var errors []error
	
	if err := te.Database.Cleanup(); err != nil {
		errors = append(errors, fmt.Errorf("database cleanup error: %w", err))
	}
	
	if err := te.Redis.Cleanup(); err != nil {
		errors = append(errors, fmt.Errorf("Redis cleanup error: %w", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("multiple cleanup errors: %v", errors)
	}
	
	return nil
}

// GetDefaultTestDatabaseConfig returns default test database configuration
func GetDefaultTestDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5432"),
		User:     getEnv("TEST_DB_USER", "test"),
		Password: getEnv("TEST_DB_PASSWORD", "test"),
		Name:     getEnv("TEST_DB_NAME", "winkr_test"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// GetDefaultTestRedisConfig returns default test Redis configuration
func GetDefaultTestRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     getEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getEnv("TEST_REDIS_PORT", "6379"),
		Password: getEnv("TEST_REDIS_PASSWORD", ""),
		DB:       1, // Use database 1 for tests
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestEnvironment is a convenience function to set up a test environment with default configs
func SetupTestEnvironment() (*TestEnvironment, error) {
	env := NewTestEnvironment(GetDefaultTestDatabaseConfig(), GetDefaultTestRedisConfig())
	if err := env.Setup(); err != nil {
		return nil, err
	}
	return env, nil
}

// WithTestEnvironment is a helper function that runs a function with a test environment
func WithTestEnvironment(fn func(*TestEnvironment) error) error {
	env, err := SetupTestEnvironment()
	if err != nil {
		return fmt.Errorf("failed to setup test environment: %w", err)
	}
	defer func() {
		if err := env.Teardown(); err != nil {
			log.Printf("Error tearing down test environment: %v", err)
		}
	}()
	
	if err := fn(env); err != nil {
		return fmt.Errorf("test function error: %w", err)
	}
	
	return nil
}

// WithCleanTestEnvironment is a helper function that runs a function with a clean test environment
func WithCleanTestEnvironment(fn func(*TestEnvironment) error) error {
	return WithTestEnvironment(func(env *TestEnvironment) error {
		// Clean up before running the function
		if err := env.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup test environment: %w", err)
		}
		
		// Run the function
		if err := fn(env); err != nil {
			return err
		}
		
		// Clean up after running the function
		if err := env.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup test environment after function: %w", err)
		}
		
		return nil
	})
}