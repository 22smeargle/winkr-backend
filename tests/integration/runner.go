package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// TestRunner manages the integration test environment
type TestRunner struct {
	config     *config.Config
	db         *postgres.Connection
	redis      *redis.Connection
	cleanup    []func() error
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	cfg := TestConfig()
	
	return &TestRunner{
		config:  cfg,
		cleanup: []func() error{},
	}
}

// Setup sets up the test environment
func (r *TestRunner) Setup() error {
	// Initialize logger
	if err := logger.Initialize("debug", "json"); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Setup database
	if err := r.setupDatabase(); err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}

	// Setup Redis
	if err := r.setupRedis(); err != nil {
		return fmt.Errorf("failed to setup redis: %w", err)
	}

	// Setup signal handling for graceful shutdown
	r.setupSignalHandling()

	log.Println("Integration test environment setup complete")
	return nil
}

// Teardown cleans up the test environment
func (r *TestRunner) Teardown() error {
	log.Println("Tearing down integration test environment...")

	var errors []error

	// Run cleanup functions in reverse order
	for i := len(r.cleanup) - 1; i >= 0; i-- {
		if err := r.cleanup[i](); err != nil {
			errors = append(errors, err)
		}
	}

	// Close Redis connection
	if r.redis != nil {
		if err := r.redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close redis: %w", err))
		}
	}

	// Close database connection
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if len(errors) > 0 {
		log.Printf("Errors during teardown: %v", errors)
		return fmt.Errorf("teardown failed with %d errors", len(errors))
	}

	log.Println("Integration test environment teardown complete")
	return nil
}

// setupDatabase sets up the test database
func (r *TestRunner) setupDatabase() error {
	db, err := postgres.NewConnection(r.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	r.db = db
	r.cleanup = append(r.cleanup, func() error {
		return db.Close()
	})

	log.Println("Database setup complete")
	return nil
}

// setupRedis sets up the test Redis connection
func (r *TestRunner) setupRedis() error {
	redis, err := redis.NewConnection(r.config.Redis)
	if err != nil {
		return fmt.Errorf("failed to create redis connection: %w", err)
	}

	// Test the connection
	if err := redis.Ping(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	r.redis = redis
	r.cleanup = append(r.cleanup, func() error {
		return redis.Close()
	})

	log.Println("Redis setup complete")
	return nil
}

// setupSignalHandling sets up graceful shutdown handling
func (r *TestRunner) setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown", sig)
		r.Teardown()
		os.Exit(0)
	}()
}

// RunTests runs all integration tests
func (r *TestRunner) RunTests() error {
	// Setup test environment
	if err := r.Setup(); err != nil {
		return fmt.Errorf("failed to setup test environment: %w", err)
	}

	// Ensure teardown runs even if tests panic
	defer r.Teardown()

	// Create test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run tests
	log.Println("Running integration tests...")
	
	// Create testing.T instance
	t := &testing.T{}
	
	// Run the test suite
	result := testing.RunTests(func(pat string, args []string) (bool, error) {
		return true, nil
	}, []string{
		"-test.v",
		"-test.run", "TestEphemeralPhotoIntegrationTestSuite",
		"-test.timeout", "30m",
	})

	if !result {
		return fmt.Errorf("integration tests failed")
	}

	log.Println("All integration tests passed!")
	return nil
}

// RunIntegrationTests runs the integration tests with proper setup/teardown
func RunIntegrationTests() {
	runner := NewTestRunner()
	
	if err := runner.RunTests(); err != nil {
		log.Fatalf("Integration tests failed: %v", err)
		os.Exit(1)
	}
}

// Main function for running integration tests directly
func main() {
	RunIntegrationTests()
}