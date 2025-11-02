package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestContext holds common test context and utilities
type TestContext struct {
	T          *testing.T
	Context    context.Context
	CancelFunc context.CancelFunc
	StartTime  time.Time
}

// NewTestContext creates a new test context
func NewTestContext(t *testing.T) *TestContext {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &TestContext{
		T:          t,
		Context:    ctx,
		CancelFunc: cancel,
		StartTime:  time.Now(),
	}
}

// Cleanup cleans up test context
func (tc *TestContext) Cleanup() {
	if tc.CancelFunc != nil {
		tc.CancelFunc()
	}
}

// Elapsed returns elapsed time since test start
func (tc *TestContext) Elapsed() time.Duration {
	return time.Since(tc.StartTime)
}

// CreateTestFile creates a temporary test file
func CreateTestFile(t *testing.T, content string, filename string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)
	
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file")
	
	return filePath
}

// CreateTestImageFile creates a temporary test image file
func CreateTestImageFile(t *testing.T, filename string) string {
	imgData := CreateTestImage(t, 200, 200)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)
	
	err := os.WriteFile(filePath, imgData, 0644)
	require.NoError(t, err, "Failed to create test image file")
	
	return filePath
}

// CreateTestImage creates a test image file for testing by downloading a sample image
func CreateTestImage(t *testing.T, width, height int) []byte {
	// Try to download a sample image from the internet
	// Using a small test image from a reliable source
	imgURL := "https://picsum.photos/200/200"
	
	// Create a temporary file to store the downloaded image
	tmpFile, err := os.CreateTemp("", "test_image_*.png")
	if err != nil {
		t.Logf("Warning: Could not create temp file: %v. Using fallback image.", err)
		return createFallbackTestImage()
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Download the image using curl
	cmd := exec.Command("curl", "-s", "-o", tmpFile.Name(), imgURL)
	if err := cmd.Run(); err != nil {
		t.Logf("Warning: Could not download image: %v. Using fallback image.", err)
		return createFallbackTestImage()
	}
	
	// Read the downloaded image
	imgData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Logf("Warning: Could not read downloaded image: %v. Using fallback image.", err)
		return createFallbackTestImage()
	}
	
	// Verify it's a valid image
	if len(imgData) < 10 {
		t.Logf("Warning: Downloaded image is too small. Using fallback image.")
		return createFallbackTestImage()
	}
	
	return imgData
}

// createFallbackTestImage creates a minimal PNG image as fallback
func createFallbackTestImage() []byte {
	// This is a minimal 1x1 transparent PNG
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x06, 0x00, 0x00, 0x00, // bit depth, color type, compression, filter, interlace
		0x4F, 0x52, 0x45, 0x4E, // CRC for IHDR
		0x00, 0x00, 0x00, 0x0A, // IDAT chunk length
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01, // compressed data
		0x0D, 0x0A, 0x2D, 0xB4, // CRC for IDAT
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // IEND CRC
	}
}

// CreateMultipartFormData creates multipart form data for file uploads
func CreateMultipartFormData(t *testing.T, fieldName, filename string, fileData []byte, additionalFields map[string]string) (bytes.Buffer, *multipart.Writer) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	
	// Add file field
	part, err := writer.CreateFormFile(fieldName, filename)
	require.NoError(t, err, "Failed to create form file field")
	
	_, err = part.Write(fileData)
	require.NoError(t, err, "Failed to write file data")
	
	// Add additional fields
	for key, value := range additionalFields {
		err := writer.WriteField(key, value)
		require.NoError(t, err, "Failed to write form field")
	}
	
	err = writer.Close()
	require.NoError(t, err, "Failed to close multipart writer")
	
	return body, writer
}

// CreateTestUser creates a test user entity
func CreateTestUser(t *testing.T) *entities.User {
	userID := uuid.New()
	return &entities.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		FirstName: "Test",
		LastName:  "User",
		Password:  "hashedpassword",
		IsVerified: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestProfile creates a test profile entity
func CreateTestProfile(t *testing.T, userID uuid.UUID) *entities.Profile {
	return &entities.Profile{
		UserID:          userID,
		Bio:             "Test bio",
		Age:             25,
		Gender:          valueobjects.GenderMale,
		InterestedIn:    []valueobjects.Gender{valueobjects.GenderFemale},
		RelationshipStatus: valueobjects.RelationshipStatusSingle,
		Location:        "Test City",
		Photos:          []string{},
		IsVerified:      true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CreateTestMatch creates a test match entity
func CreateTestMatch(t *testing.T, userID1, userID2 uuid.UUID) *entities.Match {
	return &entities.Match{
		ID:         uuid.New(),
		UserID1:    userID1,
		UserID2:    userID2,
		MatchedAt:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestMessage creates a test message entity
func CreateTestMessage(t *testing.T, matchID, senderID uuid.UUID) *entities.Message {
	return &entities.Message{
		ID:        uuid.New(),
		MatchID:   matchID,
		SenderID:  senderID,
		Content:   "Test message",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestPhoto creates a test photo entity
func CreateTestPhoto(t *testing.T, userID uuid.UUID) *entities.Photo {
	return &entities.Photo{
		ID:        uuid.New(),
		UserID:    userID,
		URL:       "https://example.com/photo.jpg",
		IsPrimary: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestReport creates a test report entity
func CreateTestReport(t *testing.T, reporterID, reportedID uuid.UUID) *entities.Report {
	return &entities.Report{
		ID:          uuid.New(),
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		Reason:      "Inappropriate content",
		Description: "Test report description",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestSubscription creates a test subscription entity
func CreateTestSubscription(t *testing.T, userID uuid.UUID) *entities.Subscription {
	return &entities.Subscription{
		ID:         uuid.New(),
		UserID:     userID,
		PlanType:   "premium",
		Status:     "active",
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// SetupTestRouter sets up a Gin router for testing
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// CreateTestRequest creates an HTTP request for testing
func CreateTestRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) *http.Request {
	var reqBody io.Reader
	
	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = bytes.NewBuffer(v)
		case string:
			reqBody = strings.NewReader(v)
		case *bytes.Buffer:
			reqBody = v
		default:
			jsonBody, err := json.Marshal(body)
			require.NoError(t, err, "Failed to marshal request body")
			reqBody = bytes.NewBuffer(jsonBody)
		}
	}
	
	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err, "Failed to create request")
	
	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// Set default content type for JSON bodies
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	
	return req
}

// ExecuteTestRequest executes an HTTP request and returns the response
func ExecuteTestRequest(t *testing.T, router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts that the response is valid JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

// AssertStatusCode asserts the HTTP status code
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
	assert.Equal(t, expectedCode, w.Code)
}

// AssertResponseContains asserts that the response body contains a substring
func AssertResponseContains(t *testing.T, w *httptest.ResponseRecorder, substring string) {
	assert.Contains(t, w.Body.String(), substring)
}

// ParseJSONResponse parses a JSON response into the provided interface
func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), v)
	require.NoError(t, err, "Failed to parse JSON response")
}

// GetTestConfig returns a test configuration
func GetTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			Name:     "test_db",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}
}

// SkipIfShort skips the test if testing.Short() is true
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	timeoutChan := time.After(timeout)
	
	for {
		select {
		case <-timeoutChan:
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// AssertEventually asserts that a condition will eventually be true
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	WaitForCondition(t, condition, timeout, message)
	assert.True(t, condition(), message)
}

// GetTestDBConnectionString returns a test database connection string
func GetTestDBConnectionString() string {
	return "postgres://test:test@localhost:5432/test_db?sslmode=disable"
}

// GetTestRedisConnectionString returns a test Redis connection string
func GetTestRedisConnectionString() string {
	return "redis://localhost:6379/1"
}

// CleanupTestData cleans up test data
func CleanupTestData(t *testing.T) {
	// This would be implemented based on your specific cleanup needs
	// For example, truncating test tables, clearing Redis, etc.
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[uuid.New().ID()[i%len(charset)]]
	}
	return string(b)
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", RandomString(8))
}

// RandomPhone generates a random phone number
func RandomPhone() string {
	return fmt.Sprintf("+1%010d", uuid.New().ID()%10000000000)
}

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	assert.NoError(t, err, msgAndArgs...)
}

// AssertError asserts that an error is not nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	assert.Error(t, err, msgAndArgs...)
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotEqual asserts that two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.NotEqual(t, expected, actual, msgAndArgs...)
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	assert.True(t, condition, msgAndArgs...)
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	assert.False(t, condition, msgAndArgs...)
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	assert.NotNil(t, value, msgAndArgs...)
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	assert.Nil(t, value, msgAndArgs...)
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, s, contains interface{}, msgAndArgs ...interface{}) {
	assert.Contains(t, s, contains, msgAndArgs...)
}

// AssertNotContains asserts that a string does not contain a substring
func AssertNotContains(t *testing.T, s, contains interface{}, msgAndArgs ...interface{}) {
	assert.NotContains(t, s, contains, msgAndArgs...)
}

// AssertLen asserts that a slice/map/channel has the expected length
func AssertLen(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) {
	assert.Len(t, object, length, msgAndArgs...)
}

// AssertEmpty asserts that a slice/map/channel/string is empty
func AssertEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	assert.Empty(t, object, msgAndArgs...)
}

// AssertNotEmpty asserts that a slice/map/channel/string is not empty
func AssertNotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	assert.NotEmpty(t, object, msgAndArgs...)
}

// RequireNoError requires that an error is nil
func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
}

// RequireError requires that an error is not nil
func RequireError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.Error(t, err, msgAndArgs...)
}

// RequireEqual requires that two values are equal
func RequireEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	require.Equal(t, expected, actual, msgAndArgs...)
}

// RequireTrue requires that a condition is true
func RequireTrue(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	require.True(t, condition, msgAndArgs...)
}

// RequireFalse requires that a condition is false
func RequireFalse(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	require.False(t, condition, msgAndArgs...)
}

// RequireNotNil requires that a value is not nil
func RequireNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	require.NotNil(t, value, msgAndArgs...)
}

// RequireNil requires that a value is nil
func RequireNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	require.Nil(t, value, msgAndArgs...)
}

// RequireLen requires that a slice/map/channel has the expected length
func RequireLen(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) {
	require.Len(t, object, length, msgAndArgs...)
}

// RequireEmpty requires that a slice/map/channel/string is empty
func RequireEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	require.Empty(t, object, msgAndArgs...)
}

// RequireNotEmpty requires that a slice/map/channel/string is not empty
func RequireNotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	require.NotEmpty(t, object, msgAndArgs...)
}

// GetCaller returns the caller's file and line number
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// LogTestStart logs the start of a test
func LogTestStart(t *testing.T) {
	t.Logf("=== START %s ===", t.Name())
}

// LogTestEnd logs the end of a test
func LogTestEnd(t *testing.T) {
	t.Logf("=== END %s ===", t.Name())
}

// LogTestStep logs a test step
func LogTestStep(t *testing.T, step string) {
	t.Logf("--- STEP: %s ---", step)
}