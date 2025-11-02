package testutils

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// IntegrationTestHelper provides utilities for integration testing
type IntegrationTestHelper struct {
	suite        *TestSuite
	testContext  *TestContext
	setupFunc    func() error
	teardownFunc func() error
	cleanupFuncs []func() error
}

// NewIntegrationTestHelper creates a new integration test helper
func NewIntegrationTestHelper(suite *TestSuite) *IntegrationTestHelper {
	return &IntegrationTestHelper{
		suite:        suite,
		testContext:  suite.TestContext,
		cleanupFuncs: make([]func() error, 0),
	}
}

// WithSetup sets up a custom setup function
func (ith *IntegrationTestHelper) WithSetup(setup func() error) *IntegrationTestHelper {
	ith.setupFunc = setup
	return ith
}

// WithTeardown sets up a custom teardown function
func (ith *IntegrationTestHelper) WithTeardown(teardown func() error) *IntegrationTestHelper {
	ith.teardownFunc = teardown
	return ith
}

// AddCleanup adds a cleanup function to be called during teardown
func (ith *IntegrationTestHelper) AddCleanup(cleanup func() error) {
	ith.cleanupFuncs = append(ith.cleanupFuncs, cleanup)
}

// Setup sets up the integration test
func (ith *IntegrationTestHelper) Setup() {
	// Wait for all services to be ready
	ith.suite.WaitForDatabase()
	ith.suite.WaitForRedis()
	
	// Run custom setup if provided
	if ith.setupFunc != nil {
		err := ith.setupFunc()
		require.NoError(ith.suite.T, err, "Integration test setup failed")
	}
	
	// Setup test data
	ith.suite.SetupTestData()
}

// Teardown tears down the integration test
func (ith *IntegrationTestHelper) Teardown() {
	// Run cleanup functions in reverse order
	for i := len(ith.cleanupFuncs) - 1; i >= 0; i-- {
		if err := ith.cleanupFuncs[i](); err != nil {
			ith.suite.T.Logf("Warning: Cleanup function failed: %v", err)
		}
	}
	
	// Cleanup test data
	ith.suite.CleanupTestData()
	
	// Flush Redis
	ith.suite.RedisHelper.FlushDB()
	
	// Run custom teardown if provided
	if ith.teardownFunc != nil {
		err := ith.teardownFunc()
		if err != nil {
			ith.suite.T.Logf("Warning: Integration test teardown failed: %v", err)
		}
	}
}

// RunTest runs an integration test with setup and teardown
func (ith *IntegrationTestHelper) RunTest(testName string, testFunc func()) {
	ith.suite.T.Run(testName, func(t *testing.T) {
		ith.Setup()
		defer ith.Teardown()
		
		testFunc()
	})
}

// E2ETestHelper provides utilities for end-to-end testing
type E2ETestHelper struct {
	suite        *TestSuite
	testContext  *TestContext
	userSessions map[string]*UserSession
	scenarios    []E2EScenario
}

// UserSession represents a user session for E2E testing
type UserSession struct {
	UserID      string
	Email       string
	Password    string
	Token       string
	Headers     map[string]string
	WSConn      *TestConnection
	Profile     map[string]interface{}
}

// E2EScenario represents an end-to-end test scenario
type E2EScenario struct {
	Name        string
	Description string
	Setup       func() error
	Execute     func() error
	Validate    func() error
	Teardown    func() error
}

// NewE2ETestHelper creates a new E2E test helper
func NewE2ETestHelper(suite *TestSuite) *E2ETestHelper {
	return &E2ETestHelper{
		suite:        suite,
		testContext:  suite.TestContext,
		userSessions: make(map[string]*UserSession),
		scenarios:    make([]E2EScenario, 0),
	}
}

// AddScenario adds an E2E test scenario
func (eth *E2ETestHelper) AddScenario(scenario E2EScenario) *E2ETestHelper {
	eth.scenarios = append(eth.scenarios, scenario)
	return eth
}

// CreateUser creates a user session for E2E testing
func (eth *E2ETestHelper) CreateUser(email, password, firstName, lastName string) *UserSession {
	// Register user
	resp := eth.testContext.Auth.RegisterUser(email, password, firstName, lastName)
	require.Equal(eth.suite.T, http.StatusCreated, resp.StatusCode)
	
	// Login and get token
	token := eth.testContext.Auth.GetAuthToken(email, password)
	
	session := &UserSession{
		UserID:   "", // Will be populated after getting user profile
		Email:    email,
		Password:  password,
		Token:     token,
		Headers:   eth.testContext.Auth.GetAuthHeaders(token),
		Profile:   make(map[string]interface{}),
	}
	
	// Get user profile to populate UserID
	profileResp := eth.testContext.API.GetAuthenticated("/profile", token)
	if profileResp.StatusCode == http.StatusOK {
		var profile map[string]interface{}
		eth.testContext.API.ExtractData(profileResp, &profile)
		session.Profile = profile
		if userID, ok := profile["id"].(string); ok {
			session.UserID = userID
		}
	}
	
	// Store session
	sessionKey := fmt.Sprintf("user_%s", email)
	eth.userSessions[sessionKey] = session
	
	return session
}

// CreateDefaultUsers creates default users for E2E testing
func (eth *E2ETestHelper) CreateDefaultUsers() map[string]*UserSession {
	users := make(map[string]*UserSession)
	
	// Create user1
	user1 := eth.CreateUser("user1@example.com", "password123", "User", "One")
	users["user1"] = user1
	
	// Create user2
	user2 := eth.CreateUser("user2@example.com", "password123", "User", "Two")
	users["user2"] = user2
	
	// Create admin user
	admin := eth.CreateUser("admin@example.com", "admin123", "Admin", "User")
	users["admin"] = admin
	
	return users
}

// GetUserSession gets a user session by email
func (eth *E2ETestHelper) GetUserSession(email string) *UserSession {
	sessionKey := fmt.Sprintf("user_%s", email)
	return eth.userSessions[sessionKey]
}

// ConnectWebSocket connects a user to WebSocket
func (eth *E2ETestHelper) ConnectWebSocket(session *UserSession, path string) {
	wsConn := eth.testContext.WebSocket.ConnectWithToken(path, session.Token)
	session.WSConn = wsConn
}

// Setup sets up the E2E test
func (eth *E2ETestHelper) Setup() {
	// Wait for all services to be ready
	eth.suite.WaitForDatabase()
	eth.suite.WaitForRedis()
	
	// Setup mock services with default expectations
	eth.suite.MockServices.SetupDefaultExpectations()
	
	// Setup test data
	eth.suite.SetupTestData()
}

// Teardown tears down the E2E test
func (eth *E2ETestHelper) Teardown() {
	// Close all WebSocket connections
	for _, session := range eth.userSessions {
		if session.WSConn != nil {
			eth.testContext.WebSocket.CloseConnection(session.WSConn)
		}
	}
	
	// Cleanup all mock services
	eth.suite.MockServices.ClearAll()
	
	// Cleanup test data
	eth.suite.CleanupTestData()
	
	// Flush Redis
	eth.suite.RedisHelper.FlushDB()
}

// RunScenario runs a specific E2E scenario
func (eth *E2ETestHelper) RunScenario(scenarioName string) {
	// Find scenario
	var scenario *E2EScenario
	for _, s := range eth.scenarios {
		if s.Name == scenarioName {
			scenario = &s
			break
		}
	}
	
	if scenario == nil {
		eth.suite.T.Fatalf("Scenario %s not found", scenarioName)
	}
	
	eth.suite.T.Run(scenarioName, func(t *testing.T) {
		eth.Setup()
		defer eth.Teardown()
		
		// Run scenario setup
		if scenario.Setup != nil {
			err := scenario.Setup()
			require.NoError(t, err, "Scenario setup failed")
		}
		
		// Run scenario execution
		if scenario.Execute != nil {
			err := scenario.Execute()
			require.NoError(t, err, "Scenario execution failed")
		}
		
		// Run scenario validation
		if scenario.Validate != nil {
			err := scenario.Validate()
			require.NoError(t, err, "Scenario validation failed")
		}
	})
}

// RunAllScenarios runs all E2E scenarios
func (eth *E2ETestHelper) RunAllScenarios() {
	for _, scenario := range eth.scenarios {
		eth.RunScenario(scenario.Name)
	}
}

// WorkflowTestHelper provides utilities for workflow testing
type WorkflowTestHelper struct {
	suite     *TestSuite
	workflows []Workflow
}

// Workflow represents a test workflow
type Workflow struct {
	Name        string
	Description string
	Steps       []WorkflowStep
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	Name        string
	Description string
	Execute     func() error
	Validate    func() error
	Retryable   bool
	MaxRetries  int
}

// NewWorkflowTestHelper creates a new workflow test helper
func NewWorkflowTestHelper(suite *TestSuite) *WorkflowTestHelper {
	return &WorkflowTestHelper{
		suite:     suite,
		workflows: make([]Workflow, 0),
	}
}

// AddWorkflow adds a workflow to test
func (wth *WorkflowTestHelper) AddWorkflow(workflow Workflow) *WorkflowTestHelper {
	wth.workflows = append(wth.workflows, workflow)
	return wth
}

// RunWorkflow runs a specific workflow
func (wth *WorkflowTestHelper) RunWorkflow(workflowName string) {
	// Find workflow
	var workflow *Workflow
	for _, w := range wth.workflows {
		if w.Name == workflowName {
			workflow = &w
			break
		}
	}
	
	if workflow == nil {
		wth.suite.T.Fatalf("Workflow %s not found", workflowName)
	}
	
	wth.suite.T.Run(workflowName, func(t *testing.T) {
		wth.suite.SetupTestData()
		defer wth.suite.CleanupTestData()
		
		// Execute workflow steps
		for i, step := range workflow.Steps {
			t.Run(step.Name, func(t *testing.T) {
				var err error
				
				// Execute step with retry logic
				if step.Retryable && step.MaxRetries > 0 {
					for retry := 0; retry <= step.MaxRetries; retry++ {
						err = step.Execute()
						if err == nil {
							break
						}
						
						if retry < step.MaxRetries {
							t.Logf("Step %s failed (attempt %d/%d): %v", 
								step.Name, retry+1, step.MaxRetries+1, err)
							time.Sleep(time.Second * time.Duration(retry+1))
						}
					}
				} else {
					err = step.Execute()
				}
				
				require.NoError(t, err, "Workflow step %s (step %d) failed", step.Name, i+1)
				
				// Validate step
				if step.Validate != nil {
					err = step.Validate()
					require.NoError(t, err, "Workflow step %s validation failed", step.Name)
				}
			})
		}
	})
}

// RunAllWorkflows runs all workflows
func (wth *WorkflowTestHelper) RunAllWorkflows() {
	for _, workflow := range wth.workflows {
		wth.RunWorkflow(workflow.Name)
	}
}

// APITestHelper provides utilities for API testing
type APITestHelper struct {
	suite     *TestSuite
	endpoints map[string]APIEndpoint
}

// APIEndpoint represents an API endpoint for testing
type APIEndpoint struct {
	Path        string
	Method      string
	Description string
	Headers     map[string]string
	Params      map[string]interface{}
	Body        interface{}
	Expected    APIExpected
}

// APIExpected represents expected API response
type APIExpected struct {
	StatusCode int
	Headers    map[string]string
	Body       interface{}
	Contains   []string
	NotContains []string
}

// NewAPITestHelper creates a new API test helper
func NewAPITestHelper(suite *TestSuite) *APITestHelper {
	return &APITestHelper{
		suite:     suite,
		endpoints: make(map[string]APIEndpoint),
	}
}

// AddEndpoint adds an API endpoint for testing
func (ath *APITestHelper) AddEndpoint(name string, endpoint APIEndpoint) *APITestHelper {
	ath.endpoints[name] = endpoint
	return ath
}

// TestEndpoint tests a specific API endpoint
func (ath *APITestHelper) TestEndpoint(endpointName string, token string) {
	endpoint, exists := ath.endpoints[endpointName]
	if !exists {
		ath.suite.T.Fatalf("Endpoint %s not found", endpointName)
	}
	
	ath.suite.T.Run(endpointName, func(t *testing.T) {
		// Prepare headers
		headers := endpoint.Headers
		if token != "" {
			authHeaders := ath.suite.TestContext.Auth.GetAuthHeaders(token)
			for k, v := range authHeaders {
				headers[k] = v
			}
		}
		
		// Make request
		var resp *http.Response
		switch endpoint.Method {
		case "GET":
			resp = ath.suite.HTTPHelper.Get(endpoint.Path, headers)
		case "POST":
			resp = ath.suite.HTTPHelper.Post(endpoint.Path, endpoint.Body, headers)
		case "PUT":
			resp = ath.suite.HTTPHelper.Put(endpoint.Path, endpoint.Body, headers)
		case "PATCH":
			resp = ath.suite.HTTPHelper.Patch(endpoint.Path, endpoint.Body, headers)
		case "DELETE":
			resp = ath.suite.HTTPHelper.Delete(endpoint.Path, headers)
		default:
			resp = ath.suite.HTTPHelper.Post(endpoint.Path, endpoint.Body, headers)
		}
		
		// Validate status code
		require.Equal(t, endpoint.Expected.StatusCode, resp.StatusCode, 
			"Expected status code %d, got %d", endpoint.Expected.StatusCode, resp.StatusCode)
		
		// Validate headers
		for key, expectedValue := range endpoint.Expected.Headers {
			actualValue := resp.Header.Get(key)
			require.Equal(t, expectedValue, actualValue, 
				"Header %s mismatch", key)
		}
		
		// Validate body content
		if endpoint.Expected.Body != nil {
			ath.suite.API.ExtractData(resp, &endpoint.Expected.Body)
		}
		
		// Validate contains
		body := ath.getResponseBody(resp)
		for _, expectedContent := range endpoint.Expected.Contains {
			require.Contains(t, body, expectedContent, 
				"Response should contain: %s", expectedContent)
		}
		
		// Validate not contains
		for _, unexpectedContent := range endpoint.Expected.NotContains {
			require.NotContains(t, body, unexpectedContent, 
				"Response should not contain: %s", unexpectedContent)
		}
	})
}

// TestAllEndpoints tests all registered endpoints
func (ath *APITestHelper) TestAllEndpoints(token string) {
	for endpointName := range ath.endpoints {
		ath.TestEndpoint(endpointName, token)
	}
}

// TestEndpointWithAuth tests an endpoint with different authentication scenarios
func (ath *APITestHelper) TestEndpointWithAuth(endpointName string) {
	ath.suite.T.Run(endpointName+"_auth", func(t *testing.T) {
		// Test without authentication
		ath.TestEndpoint(endpointName, "")
		
		// Test with valid authentication
		token := ath.suite.CreateDefaultTestUser()
		ath.TestEndpoint(endpointName, token)
		
		// Test with invalid authentication
		ath.TestEndpoint(endpointName, "invalid_token")
	})
}

// getResponseBody extracts response body
func (ath *APITestHelper) getResponseBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	return string(body)
}

// DatabaseTestHelper provides utilities for database testing
type DatabaseTestHelper struct {
	suite    *TestSuite
	scenarios []DatabaseScenario
}

// DatabaseScenario represents a database test scenario
type DatabaseScenario struct {
	Name        string
	Description string
	Setup       func() error
	Execute     func() error
	Validate    func() error
	Teardown    func() error
}

// NewDatabaseTestHelper creates a new database test helper
func NewDatabaseTestHelper(suite *TestSuite) *DatabaseTestHelper {
	return &DatabaseTestHelper{
		suite:    suite,
		scenarios: make([]DatabaseScenario, 0),
	}
}

// AddScenario adds a database scenario
func (dth *DatabaseTestHelper) AddScenario(scenario DatabaseScenario) *DatabaseTestHelper {
	dth.scenarios = append(dth.scenarios, scenario)
	return dth
}

// RunScenario runs a specific database scenario
func (dth *DatabaseTestHelper) RunScenario(scenarioName string) {
	// Find scenario
	var scenario *DatabaseScenario
	for _, s := range dth.scenarios {
		if s.Name == scenarioName {
			scenario = &s
			break
		}
	}
	
	if scenario == nil {
		dth.suite.T.Fatalf("Database scenario %s not found", scenarioName)
	}
	
	dth.suite.T.Run(scenarioName, func(t *testing.T) {
		// Run scenario setup
		if scenario.Setup != nil {
			err := scenario.Setup()
			require.NoError(t, err, "Database scenario setup failed")
		}
		
		// Run scenario execution
		if scenario.Execute != nil {
			err := scenario.Execute()
			require.NoError(t, err, "Database scenario execution failed")
		}
		
		// Run scenario validation
		if scenario.Validate != nil {
			err := scenario.Validate()
			require.NoError(t, err, "Database scenario validation failed")
		}
		
		// Run scenario teardown
		if scenario.Teardown != nil {
			err := scenario.Teardown()
			require.NoError(t, err, "Database scenario teardown failed")
		}
	})
}

// RunAllScenarios runs all database scenarios
func (dth *DatabaseTestHelper) RunAllScenarios() {
	for _, scenario := range dth.scenarios {
		dth.RunScenario(scenario.Name)
	}
}

// RedisTestHelper provides utilities for Redis testing
type RedisTestHelper struct {
	suite    *TestSuite
	scenarios []RedisScenario
}

// RedisScenario represents a Redis test scenario
type RedisScenario struct {
	Name        string
	Description string
	Setup       func() error
	Execute     func() error
	Validate    func() error
	Teardown    func() error
}

// NewRedisTestHelper creates a new Redis test helper
func NewRedisTestHelper(suite *TestSuite) *RedisTestHelper {
	return &RedisTestHelper{
		suite:    suite,
		scenarios: make([]RedisScenario, 0),
	}
}

// AddScenario adds a Redis scenario
func (rth *RedisTestHelper) AddScenario(scenario RedisScenario) *RedisTestHelper {
	rth.scenarios = append(rth.scenarios, scenario)
	return rth
}

// RunScenario runs a specific Redis scenario
func (rth *RedisTestHelper) RunScenario(scenarioName string) {
	// Find scenario
	var scenario *RedisScenario
	for _, s := range rth.scenarios {
		if s.Name == scenarioName {
			scenario = &s
			break
		}
	}
	
	if scenario == nil {
		rth.suite.T.Fatalf("Redis scenario %s not found", scenarioName)
	}
	
	rth.suite.T.Run(scenarioName, func(t *testing.T) {
		// Run scenario setup
		if scenario.Setup != nil {
			err := scenario.Setup()
			require.NoError(t, err, "Redis scenario setup failed")
		}
		
		// Run scenario execution
		if scenario.Execute != nil {
			err := scenario.Execute()
			require.NoError(t, err, "Redis scenario execution failed")
		}
		
		// Run scenario validation
		if scenario.Validate != nil {
			err := scenario.Validate()
			require.NoError(t, err, "Redis scenario validation failed")
		}
		
		// Run scenario teardown
		if scenario.Teardown != nil {
			err := scenario.Teardown()
			require.NoError(t, err, "Redis scenario teardown failed")
		}
	})
}

// RunAllScenarios runs all Redis scenarios
func (rth *RedisTestHelper) RunAllScenarios() {
	for _, scenario := range rth.scenarios {
		rth.RunScenario(scenario.Name)
	}
}

// ConcurrencyTestHelper provides utilities for concurrency testing
type ConcurrencyTestHelper struct {
	suite     *TestSuite
	scenarios []ConcurrencyScenario
}

// ConcurrencyScenario represents a concurrency test scenario
type ConcurrencyScenario struct {
	Name        string
	Description string
	Goroutines  int
	Iterations  int
	Execute     func(workerID int) error
	Validate    func() error
}

// NewConcurrencyTestHelper creates a new concurrency test helper
func NewConcurrencyTestHelper(suite *TestSuite) *ConcurrencyTestHelper {
	return &ConcurrencyTestHelper{
		suite:     suite,
		scenarios: make([]ConcurrencyScenario, 0),
	}
}

// AddScenario adds a concurrency scenario
func (cth *ConcurrencyTestHelper) AddScenario(scenario ConcurrencyScenario) *ConcurrencyTestHelper {
	cth.scenarios = append(cth.scenarios, scenario)
	return cth
}

// RunScenario runs a specific concurrency scenario
func (cth *ConcurrencyTestHelper) RunScenario(scenarioName string) {
	// Find scenario
	var scenario *ConcurrencyScenario
	for _, s := range cth.scenarios {
		if s.Name == scenarioName {
			scenario = &s
			break
		}
	}
	
	if scenario == nil {
		cth.suite.T.Fatalf("Concurrency scenario %s not found", scenarioName)
	}
	
	cth.suite.T.Run(scenarioName, func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, scenario.Goroutines)
		
		// Start goroutines
		for i := 0; i < scenario.Goroutines; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for j := 0; j < scenario.Iterations; j++ {
					if err := scenario.Execute(workerID); err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}
		
		// Wait for all goroutines to complete
		wg.Wait()
		close(errors)
		
		// Check for errors
		for err := range errors {
			require.NoError(t, err, "Concurrency test failed")
		}
		
		// Run validation
		if scenario.Validate != nil {
			err := scenario.Validate()
			require.NoError(t, err, "Concurrency validation failed")
		}
	})
}

// RunAllScenarios runs all concurrency scenarios
func (cth *ConcurrencyTestHelper) RunAllScenarios() {
	for _, scenario := range cth.scenarios {
		cth.RunScenario(scenario.Name)
	}
}