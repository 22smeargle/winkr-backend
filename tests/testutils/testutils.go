package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestSuite provides a complete test environment
type TestSuite struct {
	T               *testing.T
	Config          *TestConfig
	DB              *sqlx.DB
	GORM            *gorm.DB
	Redis           *redis.Client
	Router          *gin.Engine
	HTTPHelper      *HTTPTestHelper
	WebSocketHelper *WebSocketTestHelper
	DatabaseHelper  *DatabaseHelper
	RedisHelper     *RedisHelper
	Assert          *AssertionHelper
	MockFactory     *MockFactory
	MockServices    *MockServiceManager
	DataManager     *TestDataManager
	TestContext     *TestContext
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	testConfig := LoadTestConfig()

	// Setup database
	db, gormDB := setupTestDatabase(t, testConfig)

	// Setup Redis
	redisClient := setupTestRedis(t, testConfig)

	// Setup router
	router := setupTestRouter()

	// Create helpers
	httpHelper := NewHTTPTestHelper(t, router)
	httpHelper.StartServer()

	wsHelper := NewWebSocketTestHelper(t, httpHelper.server.URL)
	dbHelper := NewDatabaseHelper(t, db, testConfig)
	redisHelper := NewRedisHelper(t, redisClient, testConfig)
	assertHelper := NewAssertionHelper(t)
	mockFactory := NewMockFactory()
	mockServices := NewMockServiceManager()
	dataManager := NewTestDataManager(db, testConfig)

	// Create test context
	testContext := NewTestContext(t, router, testConfig, dataManager)

	return &TestSuite{
		T:               t,
		Config:          testConfig,
		DB:              db,
		GORM:            gormDB,
		Redis:           redisClient,
		Router:          router,
		HTTPHelper:      httpHelper,
		WebSocketHelper: wsHelper,
		DatabaseHelper:  dbHelper,
		RedisHelper:     redisHelper,
		Assert:          assertHelper,
		MockFactory:     mockFactory,
		MockServices:    mockServices,
		DataManager:     dataManager,
		TestContext:     testContext,
	}
}

// setupTestDatabase sets up the test database
func setupTestDatabase(t *testing.T, config *TestConfig) (*sqlx.DB, *gorm.DB) {
	// Connect to PostgreSQL
	connStr := config.GetDatabaseConnectionString()
	
	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	// Setup GORM
	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to setup GORM")

	// Run migrations
	err = runMigrations(gormDB)
	require.NoError(t, err, "Failed to run migrations")

	return db, gormDB
}

// setupTestRedis sets up the test Redis connection
func setupTestRedis(t *testing.T, config *TestConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.GetRedisConnectionString(),
		Password:  config.Redis.Password,
		DB:        config.Redis.DB,
	})

	// Test connection
	err := client.Ping(context.Background()).Err()
	require.NoError(t, err, "Failed to connect to test Redis")

	// Flush database
	client.FlushDB(context.Background())

	return client
}

// setupTestRouter sets up the test router
func setupTestRouter() *gin.Engine {
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	
	return router
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB) error {
	// Auto migrate all entities
	return db.AutoMigrate(
		&entities.User{},
		&entities.Profile{},
		&entities.Photo{},
		&entities.Match{},
		&entities.Message{},
		&entities.Subscription{},
		&entities.Verification{},
		&entities.Report{},
		&entities.EphemeralPhoto{},
	)
}

// SetupTestData sets up test data
func (ts *TestSuite) SetupTestData() {
	ctx := context.Background()
	err := ts.DataManager.SetupTestData(ctx)
	require.NoError(ts.T, err, "Failed to setup test data")
}

// CleanupTestData cleans up test data
func (ts *TestSuite) CleanupTestData() {
	ctx := context.Background()
	err := ts.DataManager.CleanupTestData(ctx)
	require.NoError(ts.T, err, "Failed to cleanup test data")
}

// ResetTestData resets test data
func (ts *TestSuite) ResetTestData() {
	ctx := context.Background()
	err := ts.DataManager.ResetTestData(ctx)
	require.NoError(ts.T, err, "Failed to reset test data")
}

// Cleanup cleans up the test suite
func (ts *TestSuite) Cleanup() {
	// Close HTTP server
	if ts.HTTPHelper != nil {
		ts.HTTPHelper.StopServer()
	}

	// Close database connections
	if ts.DB != nil {
		ts.DB.Close()
	}
	if ts.GORM != nil {
		sqlDB, err := ts.GORM.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Close Redis connection
	if ts.Redis != nil {
		ts.Redis.Close()
	}

	// Close WebSocket connections
	if ts.WebSocketHelper != nil {
		ts.WebSocketHelper.CloseAllConnections()
	}
}

// CreateTestUser creates a test user and returns auth token
func (ts *TestSuite) CreateTestUser(email, password, firstName, lastName string) string {
	return ts.TestContext.CreateTestUser(email, password, firstName, lastName)
}

// CreateDefaultTestUser creates a default test user
func (ts *TestSuite) CreateDefaultTestUser() string {
	return ts.TestContext.CreateDefaultTestUser()
}

// WithTransaction executes a function within a database transaction
func (ts *TestSuite) WithTransaction(fn func(*sqlx.Tx) error) {
	ts.DatabaseHelper.WithTransaction(fn)
}

// WithTimeout executes a function with timeout
func (ts *TestSuite) WithTimeout(timeout time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return fn(ctx)
}

// Eventually asserts that a condition eventually becomes true
func (ts *TestSuite) Eventually(condition func() bool, timeout time.Duration, message string) {
	ts.Assert.AssertEventually(condition, timeout, message)
}

// WaitForDatabase waits for database to be ready
func (ts *TestSuite) WaitForDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		err := ts.DB.PingContext(ctx)
		if err == nil {
			break
		}
		
		select {
		case <-ctx.Done():
			require.Fail(ts.T, "Database not ready within timeout")
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}
}

// WaitForRedis waits for Redis to be ready
func (ts *TestSuite) WaitForRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		err := ts.Redis.Ping(ctx).Err()
		if err == nil {
			break
		}
		
		select {
		case <-ctx.Done():
			require.Fail(ts.T, "Redis not ready within timeout")
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}
}

// CreateTempFile creates a temporary file for testing
func (ts *TestSuite) CreateTempFile(content string) string {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(ts.T, err, "Failed to create temp file")
	
	_, err = tmpFile.WriteString(content)
	require.NoError(ts.T, err, "Failed to write to temp file")
	
	err = tmpFile.Close()
	require.NoError(ts.T, err, "Failed to close temp file")
	
	return tmpFile.Name()
}

// CreateTempDir creates a temporary directory for testing
func (ts *TestSuite) CreateTempDir() string {
	tmpDir, err := os.MkdirTemp("", "test_*")
	require.NoError(ts.T, err, "Failed to create temp dir")
	
	return tmpDir
}

// CleanupTempFile removes a temporary file
func (ts *TestSuite) CleanupTempFile(path string) {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		ts.T.Logf("Warning: Failed to remove temp file %s: %v", path, err)
	}
}

// CleanupTempDir removes a temporary directory
func (ts *TestSuite) CleanupTempDir(path string) {
	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		ts.T.Logf("Warning: Failed to remove temp dir %s: %v", path, err)
	}
}

// GetTestDataPath returns the path to test data directory
func (ts *TestSuite) GetTestDataPath() string {
	return filepath.Join("tests", "data")
}

// GetTestFixturesPath returns the path to test fixtures directory
func (ts *TestSuite) GetTestFixturesPath() string {
	return filepath.Join("tests", "fixtures")
}

// LoadTestFile loads a test file
func (ts *TestSuite) LoadTestFile(filename string) []byte {
	path := filepath.Join(ts.GetTestDataPath(), filename)
	data, err := os.ReadFile(path)
	require.NoError(ts.T, err, "Failed to read test file %s", filename)
	return data
}

// SaveTestFile saves data to a test file
func (ts *TestSuite) SaveTestFile(filename string, data []byte) {
	path := filepath.Join(ts.GetTestDataPath(), filename)
	err := os.WriteFile(path, data, 0644)
	require.NoError(ts.T, err, "Failed to write test file %s", filename)
}

// AssertDatabaseState asserts database state
func (ts *TestSuite) AssertDatabaseState(assertions func(*DatabaseHelper)) {
	assertions(ts.DatabaseHelper)
}

// AssertRedisState asserts Redis state
func (ts *TestSuite) AssertRedisState(assertions func(*RedisHelper)) {
	assertions(ts.RedisHelper)
}

// AssertMockCalls asserts mock service calls
func (ts *TestSuite) AssertMockCalls(assertions func(*MockServiceManager)) {
	assertions(ts.MockServices)
}

// SkipTestIf skips the test if condition is true
func (ts *TestSuite) SkipTestIf(condition bool, reason string) {
	if condition {
		ts.T.Skip(reason)
	}
}

// SkipTestIfEnv skips the test if environment variable is set
func (ts *TestSuite) SkipTestIfEnv(envVar, reason string) {
	if os.Getenv(envVar) != "" {
		ts.T.Skipf(reason + " (env: %s)", envVar)
	}
}

// RunTestIf runs the test only if condition is true
func (ts *TestSuite) RunTestIf(condition bool, testName string, testFunc func()) {
	if condition {
		ts.T.Run(testName, func(t *testing.T) {
			testFunc()
		})
	}
}

// RunTestIfEnv runs the test only if environment variable is set
func (ts *TestSuite) RunTestIfEnv(envVar, testName string, testFunc func()) {
	if os.Getenv(envVar) != "" {
		ts.T.Run(testName, func(t *testing.T) {
			testFunc()
		})
	}
}

// BenchmarkTestHelper provides utilities for benchmark testing
type BenchmarkTestHelper struct {
	b *testing.B
}

// NewBenchmarkTestHelper creates a new benchmark test helper
func NewBenchmarkTestHelper(b *testing.B) *BenchmarkTestHelper {
	return &BenchmarkTestHelper{b: b}
}

// ResetTimer resets the benchmark timer
func (bth *BenchmarkTestHelper) ResetTimer() {
	bth.b.ResetTimer()
}

// StopTimer stops the benchmark timer
func (bth *BenchmarkTestHelper) StopTimer() {
	bth.b.StopTimer()
}

// StartTimer starts the benchmark timer
func (bth *BenchmarkTestHelper) StartTimer() {
	bth.b.StartTimer()
}

// ReportAllocs reports memory allocations
func (bth *BenchmarkTestHelper) ReportAllocs() {
	bth.b.ReportAllocs()
}

// ReportMetric reports a custom metric
func (bth *BenchmarkTestHelper) ReportMetric(name string, value float64) {
	bth.b.ReportMetric(value, name)
}

// Parallel runs the benchmark in parallel
func (bth *BenchmarkTestHelper) Parallel(f func()) {
	bth.b.RunParallel(f)
}

// ParallelTestHelper provides utilities for parallel testing
type ParallelTestHelper struct {
	t *testing.T
}

// NewParallelTestHelper creates a new parallel test helper
func NewParallelTestHelper(t *testing.T) *ParallelTestHelper {
	return &ParallelTestHelper{t: t}
}

// Run runs a test in parallel
func (pth *ParallelTestHelper) Run(name string, f func(*testing.T)) {
	pth.t.Run(name, func(t *testing.T) {
		t.Parallel()
		f(t)
	})
}

// RunSequential runs tests sequentially
func (pth *ParallelTestHelper) RunSequential(name string, f func(*testing.T)) {
	pth.t.Run(name, func(t *testing.T) {
		f(t)
	})
}

// IntegrationTestHelper provides utilities for integration testing
type IntegrationTestHelper struct {
	suite *TestSuite
}

// NewIntegrationTestHelper creates a new integration test helper
func NewIntegrationTestHelper(suite *TestSuite) *IntegrationTestHelper {
	return &IntegrationTestHelper{suite: suite}
}

// SetupIntegrationTest sets up an integration test
func (ith *IntegrationTestHelper) SetupIntegrationTest() {
	// Wait for all services to be ready
	ith.suite.WaitForDatabase()
	ith.suite.WaitForRedis()
	
	// Setup test data
	ith.suite.SetupTestData()
}

// CleanupIntegrationTest cleans up an integration test
func (ith *IntegrationTestHelper) CleanupIntegrationTest() {
	// Cleanup test data
	ith.suite.CleanupTestData()
	
	// Flush Redis
	ith.suite.RedisHelper.FlushDB()
}

// E2ETestHelper provides utilities for end-to-end testing
type E2ETestHelper struct {
	suite *TestSuite
}

// NewE2ETestHelper creates a new end-to-end test helper
func NewE2ETestHelper(suite *TestSuite) *E2ETestHelper {
	return &E2ETestHelper{suite: suite}
}

// SetupE2ETest sets up an end-to-end test
func (eth *E2ETestHelper) SetupE2ETest() {
	// Wait for all services to be ready
	eth.suite.WaitForDatabase()
	eth.suite.WaitForRedis()
	
	// Setup mock services with default expectations
	eth.suite.MockServices.SetupDefaultExpectations()
}

// CleanupE2ETest cleans up an end-to-end test
func (eth *E2ETestHelper) CleanupE2ETest() {
	// Cleanup all mock services
	eth.suite.MockServices.ClearAll()
	
	// Cleanup test data
	eth.suite.CleanupTestData()
	
	// Flush Redis
	eth.suite.RedisHelper.FlushDB()
}

// LoadTestHelper provides utilities for load testing
type LoadTestHelper struct {
	suite *TestSuite
}

// NewLoadTestHelper creates a new load test helper
func NewLoadTestHelper(suite *TestSuite) *LoadTestHelper {
	return &LoadTestHelper{suite: suite}
}

// SetupLoadTest sets up a load test
func (lth *LoadTestHelper) SetupLoadTest() {
	// Wait for all services to be ready
	lth.suite.WaitForDatabase()
	lth.suite.WaitForRedis()
	
	// Optimize database for load testing
	lth.optimizeDatabaseForLoad()
}

// optimizeDatabaseForLoad optimizes database for load testing
func (lth *LoadTestHelper) optimizeDatabaseForLoad() {
	// Disable logging for performance
	lth.suite.GORM = lth.suite.GORM.Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// CleanupLoadTest cleans up a load test
func (lth *LoadTestHelper) CleanupLoadTest() {
	// Cleanup test data
	lth.suite.CleanupTestData()
	
	// Flush Redis
	lth.suite.RedisHelper.FlushDB()
}

// SecurityTestHelper provides utilities for security testing
type SecurityTestHelper struct {
	suite *TestSuite
}

// NewSecurityTestHelper creates a new security test helper
func NewSecurityTestHelper(suite *TestSuite) *SecurityTestHelper {
	return &SecurityTestHelper{suite: suite}
}

// SetupSecurityTest sets up a security test
func (sth *SecurityTestHelper) SetupSecurityTest() {
	// Wait for all services to be ready
	sth.suite.WaitForDatabase()
	sth.suite.WaitForRedis()
	
	// Enable all security features
	sth.enableSecurityFeatures()
}

// enableSecurityFeatures enables security features for testing
func (sth *SecurityTestHelper) enableSecurityFeatures() {
	// This would enable rate limiting, CSRF protection, etc.
	// Implementation depends on your security middleware
}

// CleanupSecurityTest cleans up a security test
func (sth *SecurityTestHelper) CleanupSecurityTest() {
	// Cleanup test data
	sth.suite.CleanupTestData()
	
	// Flush Redis
	sth.suite.RedisHelper.FlushDB()
}