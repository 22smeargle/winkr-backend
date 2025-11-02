# Test Utilities

This directory contains comprehensive test utilities and helpers for the Winkr backend application. These utilities provide a complete testing framework for unit tests, integration tests, end-to-end tests, load tests, and security tests.

## Overview

The test utilities are organized into several modules:

- **Core Helpers**: Basic testing infrastructure and setup
- **Mock Services**: Mock implementations of external services
- **Database Helpers**: Database testing utilities
- **Redis Helpers**: Redis/Cache testing utilities
- **HTTP Helpers**: HTTP/WebSocket testing utilities
- **Security Helpers**: Security testing utilities
- **Performance Helpers**: Performance and load testing utilities
- **Integration Helpers**: Integration and E2E testing utilities
- **Coverage Helpers**: Code coverage analysis utilities

## Core Helpers

### TestSuite (`testutils.go`)

The main test suite that provides a complete testing environment:

```go
func TestMyFeature(t *testing.T) {
    suite := NewTestSuite(t)
    defer suite.Cleanup()
    
    // Setup test data
    suite.SetupTestData()
    
    // Run your test
    // ...
    
    // Test assertions
    suite.Assert.AssertEqual(expected, actual)
}
```

### MockFactory (`mock_factories.go`)

Factory methods for creating test entities with optional overrides:

```go
factory := NewMockFactory()

// Create a user with default values
user := factory.CreateUser()

// Create a user with custom values
user := factory.CreateUser(func(u *entities.User) {
    u.Email = "test@example.com"
    u.IsVerified = true
})
```

### TestConfiguration (`test_config.go`)

Configuration management for different test environments:

```go
// Load test configuration from environment
config := LoadTestConfig()

// Get configuration for specific test type
config := GetConfigForEnvironment(UnitTestEnv)
```

### TestFixtures (`test_fixtures.go`)

Database fixtures and test data management:

```go
dataManager := NewTestDataManager(db, config)

// Setup test data
err := dataManager.SetupTestData(ctx)

// Get specific fixtures
user := dataManager.GetFixtureLoader().GetUserFixture("test@example.com")
```

### Assertions (`assertions.go`)

Custom assertion methods for comprehensive testing:

```go
assert := NewAssertionHelper(t)

// Entity assertions
assert.AssertUserEqual(expectedUser, actualUser)
assert.AssertProfileEqual(expectedProfile, actualProfile)

// HTTP assertions
assert.AssertHTTPResponse(resp, http.StatusOK, "application/json")
assert.AssertJSONResponse(resp)

// Validation assertions
assert.AssertEmail("test@example.com")
assert.AssertUUID(userID)
```

## Mock Services

### MockServiceManager (`mock_services.go`)

Comprehensive mock implementations for all external services:

```go
mockServices := NewMockServiceManager()

// Setup default expectations
mockServices.SetupDefaultExpectations()

// Access specific services
emailService := mockServices.EmailService
paymentService := mockServices.PaymentService

// Verify calls
emailService.AssertCalled(t, "SendEmail", mock.Anything, "test@example.com", "Subject", "Body")
```

Available mock services:
- **EmailService**: Email sending functionality
- **SMSService**: SMS sending functionality
- **StorageService**: File storage functionality
- **PaymentService**: Payment processing functionality
- **AIService**: AI/ML functionality
- **NotificationService**: Push notifications
- **FileUploadService**: File upload handling
- **GeoLocationService**: Geolocation services

## Database Helpers

### DatabaseHelper (`database_helpers.go`)

Database testing utilities with comprehensive assertions:

```go
dbHelper := NewDatabaseHelper(t, db, config)

// Cleanup operations
dbHelper.CleanupAllTables()
dbHelper.TruncateTable("users")

// Data operations
dbHelper.InsertUser(user)
dbHelper.InsertProfile(profile)

// Assertions
dbHelper.AssertUserExists(userID)
dbHelper.AssertProfileExists(userID)
dbHelper.AssertTableRowCount("users", 5)
```

### TestDatabase (`database_setup.go`)

Database setup and teardown for testing:

```go
testDB := NewTestDatabase(config)

// Setup test database
err := testDB.Setup()

// Get database connection
db := testDB.GetDB()

// Cleanup
err := testDB.Teardown()
```

## Redis Helpers

### RedisHelper (`redis_helpers.go`)

Redis/Cache testing utilities:

```go
redisHelper := NewRedisHelper(t, redisClient, config)

// Basic operations
redisHelper.SetString("key", "value")
redisHelper.SetJSON("key", data)
redisHelper.SetHash("hash", "field", "value")

// Assertions
redisHelper.AssertKeyExists("key")
redisHelper.AssertStringEqual("key", "value")
redisHelper.AssertJSONEqual("key", expectedData)

// Advanced operations
redisHelper.SetRateLimit("rate_limit_key", 100, time.Minute)
redisHelper.AssertRateLimitNotExceeded("rate_limit_key", 100)
```

## HTTP Helpers

### HTTPTestHelper (`http_helpers.go`)

HTTP testing utilities with comprehensive request/response handling:

```go
httpHelper := NewHTTPTestHelper(t, router)
httpHelper.StartServer()
defer httpHelper.StopServer()

// Make requests
resp := httpHelper.Get("/api/users", nil)
resp := httpHelper.Post("/api/users", userData, headers)
resp := httpHelper.PostMultipart("/api/upload", fields, files, headers)

// Assertions
httpHelper.AssertSuccess(resp)
httpHelper.AssertBadRequest(resp)
httpHelper.AssertJSONData(resp, expectedData)
```

### AuthHelper (`http_helpers.go`)

Authentication testing utilities:

```go
authHelper := NewAuthHelper(t, httpHelper, config)

// User operations
resp := authHelper.RegisterUser(email, password, firstName, lastName)
resp := authHelper.LoginUser(email, password)
token := authHelper.GetAuthToken(email, password)

// Authenticated requests
headers := authHelper.GetAuthHeaders(token)
resp := httpHelper.GetWithServer("/api/profile", headers)
```

### WebSocketHelper (`http_helpers.go`)

WebSocket testing utilities:

```go
wsHelper := NewWebSocketTestHelper(t, server)

// Connect
conn := wsHelper.Connect("/ws/chat", headers)
conn := wsHelper.ConnectWithToken("/ws/chat", token)

// Send messages
wsHelper.SendMessage(conn, TestMessage{Type: "chat", Data: "Hello"})

// Receive messages
msg := wsHelper.WaitForMessage(conn, time.Second*5)
wsHelper.AssertMessageReceived(conn, expectedMessage)
```

## Security Helpers

### SecurityTestHelper (`security_helpers.go`)

Comprehensive security testing utilities:

```go
securityHelper := NewSecurityTestHelper(t, httpHelper, authHelper)

// Test various vulnerabilities
securityHelper.TestSQLInjection("/api/users", params)
securityHelper.TestXSS("/api/search", params)
securityHelper.TestPathTraversal("/api/files", "file")
securityHelper.TestAuthenticationBypass("/api/admin")
securityHelper.TestRateLimiting("/api/login", 10)

// Generate security report
report := securityHelper.GenerateReport()
report.PrintReport()
```

## Performance Helpers

### PerformanceTestHelper (`performance_helpers.go`)

Performance and load testing utilities:

```go
perfHelper := NewPerformanceTestHelper(t)

// Start performance test
perfHelper.StartTest()

// Record requests
perfHelper.RecordRequest(duration, success)

// Stop and get metrics
perfHelper.StopTest()
metrics := perfHelper.GetMetrics()

// Assertions
perfHelper.AssertThroughput(1000) // 1000 QPS minimum
perfHelper.AssertErrorRate(5.0)    // 5% error rate maximum
perfHelper.AssertAvgResponseTime(time.Millisecond * 100)
```

### LoadTestRunner (`performance_helpers.go`)

Load testing with configurable concurrency:

```go
loadRunner := NewLoadTestRunner(t, 100, time.Minute*5) // 100 users for 5 minutes

// Run load test
loadRunner.Run(func() (time.Duration, bool) {
    start := time.Now()
    resp := httpHelper.Get("/api/users", nil)
    return time.Since(start), resp.StatusCode == 200
})

// Get metrics
metrics := loadRunner.GetMetrics()
```

### BenchmarkHelper (`performance_helpers.go`)

Benchmark testing utilities:

```go
benchmarkHelper := NewBenchmarkHelper(t, 1000)

// Run benchmark
result := benchmarkHelper.Run(func() time.Duration {
    start := time.Now()
    // Code to benchmark
    return time.Since(start)
})

// Assertions
result.AssertThroughput(t, 1000)
result.AssertAvgTime(t, time.Millisecond*10)
```

## Integration Helpers

### IntegrationTestHelper (`integration_helpers.go`)

Integration testing utilities:

```go
integrationHelper := NewIntegrationTestHelper(suite)

// Run integration test
integrationHelper.RunTest("user_registration", func() {
    // Test implementation
    resp := httpHelper.Post("/api/register", userData, nil)
    assert.Equal(http.StatusCreated, resp.StatusCode)
})
```

### E2ETestHelper (`integration_helpers.go`)

End-to-end testing utilities:

```go
e2eHelper := NewE2ETestHelper(suite)

// Create users
users := e2eHelper.CreateDefaultUsers()

// Add scenario
e2eHelper.AddScenario(E2EScenario{
    Name: "complete_user_flow",
    Setup: func() error { /* setup */ return nil },
    Execute: func() error { /* execute */ return nil },
    Validate: func() error { /* validate */ return nil },
})

// Run scenario
e2eHelper.RunScenario("complete_user_flow")
```

## Coverage Helpers

### CoverageHelper (`coverage_helpers.go`)

Code coverage analysis utilities:

```go
coverageHelper := NewCoverageHelper(t)

// Setup coverage
coverageHelper.SetupCoverage()

// Run tests with coverage
err := coverageHelper.RunTestsWithCoverage("./...", "-v")

// Generate report
report, err := coverageHelper.GenerateCoverageReport()

// Generate HTML report
err = coverageHelper.GenerateHTMLReport(report)

// Assert thresholds
coverageHelper.AssertCoverageThresholds(report)

// Generate badge
badgeFile, err := coverageHelper.GenerateCoverageBadge(report)
```

## Usage Examples

### Complete Test Example

```go
func TestUserRegistrationFlow(t *testing.T) {
    // Setup test suite
    suite := NewTestSuite(t)
    defer suite.Cleanup()
    
    // Setup test data
    suite.SetupTestData()
    
    // Create test user
    user := suite.MockFactory.CreateUser(func(u *entities.User) {
        u.Email = "test@example.com"
        u.FirstName = "Test"
        u.LastName = "User"
    })
    
    // Test registration
    resp := suite.HTTPHelper.Post("/api/auth/register", map[string]interface{}{
        "email":     user.Email,
        "firstName": user.FirstName,
        "lastName":  user.LastName,
        "password":  "password123",
    }, nil)
    
    // Assertions
    suite.Assert.AssertHTTPResponse(resp, http.StatusCreated, "application/json")
    suite.Assert.AssertJSONSuccess(resp, "User registered successfully")
    
    // Verify database state
    suite.DatabaseHelper.AssertUserExists(user.ID.String())
}
```

### Integration Test Example

```go
func TestCompleteUserFlow(t *testing.T) {
    integrationHelper := NewIntegrationTestHelper(suite)
    
    integrationHelper.RunTest("complete_flow", func() {
        // Register user
        user := integrationHelper.CreateUser("test@example.com", "password123", "Test", "User")
        
        // Login
        token := integrationHelper.Auth.GetAuthToken("test@example.com", "password123")
        
        // Create profile
        profileData := map[string]interface{}{
            "bio":     "Software engineer",
            "age":     28,
            "gender":   "male",
            "location": "San Francisco",
        }
        
        resp := integrationHelper.API.PostAuthenticated("/api/profile", profileData, token)
        integrationHelper.API.ExpectCreated(resp)
        
        // Upload photo
        imageData := integrationHelper.CreateTestImage(200, 200)
        files := map[string][]byte{"photo": imageData}
        resp = integrationHelper.API.UploadFile("/api/photos", nil, files, token)
        integrationHelper.API.ExpectCreated(resp)
    })
}
```

### Load Test Example

```go
func TestAPIUnderLoad(t *testing.T) {
    suite := NewTestSuite(t)
    defer suite.Cleanup()
    
    loadRunner := NewLoadTestRunner(t, 50, time.Minute*2)
    
    loadRunner.Run(func() (time.Duration, bool) {
        start := time.Now()
        resp := suite.HTTPHelper.Get("/api/users", nil)
        return time.Since(start), resp.StatusCode == 200
    })
    
    metrics := loadRunner.GetMetrics()
    
    // Assert performance requirements
    loadRunner.AssertThroughput(100)  // 100 QPS minimum
    loadRunner.AssertErrorRate(1.0)    // 1% error rate maximum
    loadRunner.AssertAvgResponseTime(time.Millisecond * 100)
}
```

### Security Test Example

```go
func TestSecurityVulnerabilities(t *testing.T) {
    suite := NewTestSuite(t)
    defer suite.Cleanup()
    
    securityHelper := NewSecurityTestHelper(t, suite.HTTPHelper, suite.AuthHelper)
    
    // Test SQL injection
    securityHelper.TestSQLInjection("/api/users", map[string]string{
        "email": "' OR '1'='1",
    })
    
    // Test XSS
    securityHelper.TestXSS("/api/search", map[string]string{
        "query": "<script>alert('XSS')</script>",
    })
    
    // Test authentication bypass
    securityHelper.TestAuthenticationBypass("/api/admin")
    
    // Generate security report
    report := securityHelper.GenerateReport()
    if len(report.Vulnerabilities) > 0 {
        t.Errorf("Security vulnerabilities found: %d", len(report.Vulnerabilities))
    }
}
```

## Configuration

### Environment Variables

The test utilities can be configured using environment variables:

```bash
# Database configuration
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=test
export TEST_DB_PASSWORD=test
export TEST_DB_NAME=winkr_test

# Redis configuration
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6379
export TEST_REDIS_DB=1

# JWT configuration
export TEST_JWT_SECRET=test-secret-key
export TEST_JWT_EXPIRATION=24h

# External services
export TEST_EMAIL_PROVIDER=mock
export TEST_PAYMENT_PROVIDER=mock
export TEST_AI_PROVIDER=mock
```

### Test Environments

Different test environments can be configured:

```go
// Unit test environment (minimal setup)
config := GetConfigForEnvironment(UnitTestEnv)

// Integration test environment
config := GetConfigForEnvironment(IntegrationTestEnv)

// E2E test environment
config := GetConfigForEnvironment(E2ETestEnv)

// Load test environment
config := GetConfigForEnvironment(LoadTestEnv)

// Security test environment
config := GetConfigForEnvironment(SecurityTestEnv)
```

## Best Practices

### 1. Test Organization

- Use descriptive test names
- Group related tests in subtests
- Use table-driven tests for multiple scenarios
- Keep tests focused and independent

### 2. Test Data Management

- Use factories for creating test data
- Clean up test data after each test
- Use transactions for data isolation
- Avoid sharing state between tests

### 3. Mock Usage

- Setup mock expectations before each test
- Verify mock calls after each test
- Use realistic mock data
- Clear mocks between tests

### 4. Assertions

- Use specific assertion methods
- Include descriptive failure messages
- Assert both positive and negative cases
- Check edge cases and error conditions

### 5. Performance Testing

- Use realistic load patterns
- Monitor system resources
- Test under different conditions
- Establish performance baselines

### 6. Security Testing

- Test common vulnerability patterns
- Use both valid and invalid inputs
- Test authentication and authorization
- Verify input validation

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Ensure test database is running
   - Check connection string configuration
   - Verify database permissions

2. **Redis Connection Errors**
   - Ensure Redis is running
   - Check Redis configuration
   - Verify network connectivity

3. **Mock Service Issues**
   - Setup mock expectations before use
   - Clear mocks between tests
   - Verify mock call arguments

4. **Coverage Issues**
   - Ensure test files are in correct packages
   - Check coverage configuration
   - Verify build tags

### Debug Tips

1. **Enable Verbose Logging**
   ```go
   suite.Config.LogLevel = "debug"
   ```

2. **Inspect Test Data**
   ```go
   fmt.Printf("Test data: %+v\n", testData)
   ```

3. **Check Database State**
   ```go
   suite.DatabaseHelper.AssertDatabaseState(func(db *DatabaseHelper) {
       count := db.CountRows("users")
       fmt.Printf("Users in database: %d\n", count)
   })
   ```

4. **Monitor HTTP Requests**
   ```go
   // Enable request logging
   suite.HTTPHelper.EnableRequestLogging()
   ```

## Contributing

When adding new test utilities:

1. Follow existing code patterns
2. Add comprehensive documentation
3. Include usage examples
4. Add unit tests for utilities
5. Update this README

## Dependencies

The test utilities depend on:

- `github.com/stretchr/testify` - Testing assertions
- `github.com/gorilla/websocket` - WebSocket testing
- `github.com/go-redis/redis/v8` - Redis testing
- `github.com/jmoiron/sqlx` - Database testing
- `gorm.io/gorm` - ORM testing

## License

These test utilities are part of the Winkr backend project and follow the same license terms.