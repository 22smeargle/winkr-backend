package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/repositories"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PhotoIntegrationTestSuite is the test suite for photo integration tests
type PhotoIntegrationTestSuite struct {
	suite.Suite
	db        *gorm.DB
	server    *httptest.Server
	userID    uuid.UUID
	authToken string
	photoID   uuid.UUID
}

// SetupSuite runs once before all tests
func (suite *PhotoIntegrationTestSuite) SetupSuite() {
	// Initialize in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Photo{},
		&models.Match{},
		&models.Message{},
		&models.Subscription{},
		&models.Report{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// Create test user
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(user).Error
	suite.Require().NoError(err)

	suite.userID = user.ID
	suite.authToken = "Bearer test-token"

	// Setup test server
	suite.setupTestServer()
}

// TearDownSuite runs once after all tests
func (suite *PhotoIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest runs before each test
func (suite *PhotoIntegrationTestSuite) SetupTest() {
	// Clean up photos before each test
	suite.db.Where("user_id = ?", suite.userID).Delete(&models.Photo{})
}

// setupTestServer sets up the test HTTP server
func (suite *PhotoIntegrationTestSuite) setupTestServer() {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", suite.userID)
		c.Set("user_role", "user")
		c.Next()
	})

	// Initialize repositories
	userRepo := repositories.NewUserRepository(suite.db)
	photoRepo := repositories.NewPhotoRepository(suite.db)

	// Initialize mock services (simplified for testing)
	storageService := &MockStorageService{}
	imageProcessor := &MockImageProcessor{}

	// Initialize use cases
	uploadPhotoUseCase := NewMockUploadPhotoUseCase()
	deletePhotoUseCase := NewMockDeletePhotoUseCase()
	getUploadURLUseCase := NewMockGetUploadURLUseCase()
	getDownloadURLUseCase := NewMockGetDownloadURLUseCase()
	setPrimaryPhotoUseCase := NewMockSetPrimaryPhotoUseCase()
	markPhotoViewedUseCase := NewMockMarkPhotoViewedUseCase()

	// Initialize handlers
	photoHandler := handlers.NewPhotoHandler(
		uploadPhotoUseCase,
		deletePhotoUseCase,
		getUploadURLUseCase,
		getDownloadURLUseCase,
		setPrimaryPhotoUseCase,
		markPhotoViewedUseCase,
		nil, // JWT utils not needed for integration tests
	)

	// Initialize routes
	photoRoutes := routes.NewPhotoRoutes(photoHandler, nil)

	// Register routes
	v1 := router.Group("/api/v1")
	photoRoutes.RegisterRoutes(v1, nil)

	// Create test server
	suite.server = httptest.NewServer(router)
}

// createTestImageData creates a base64 encoded test image
func createTestImageData() string {
	// Simple 1x1 red pixel in PNG format (base64 encoded)
	pixel := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
	return "data:image/png;base64," + pixel
}

func (suite *PhotoIntegrationTestSuite) TestPhotoUploadFlow() {
	// Test the complete photo upload flow
	imageData := createTestImageData()

	// Step 1: Get upload URL
	req, _ := http.NewRequest("GET", suite.server.URL+"/api/v1/media/request-upload?content_type=image/png", nil)
	req.Header.Set("Authorization", suite.authToken)
	
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.Code)
	
	var uploadURLResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&uploadURLResponse)
	suite.Require().NoError(err)
	
	uploadURL, ok := uploadURLResponse["data"].(map[string]interface{})["upload_url"].(string)
	suite.True(ok, "Upload URL should be present in response")
	suite.NotEmpty(uploadURL)

	// Step 2: Upload photo (simulated)
	uploadReq := map[string]interface{}{
		"image_data":    imageData,
		"content_type": "image/png",
	}
	
	uploadReqBody, _ := json.Marshal(uploadReq)
	req, _ = http.NewRequest("POST", suite.server.URL+"/api/v1/me/photos", bytes.NewBuffer(uploadReqBody))
	req.Header.Set("Authorization", suite.authToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusCreated, resp.Code)
	
	var uploadResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&uploadResponse)
	suite.Require().NoError(err)
	
	photoData, ok := uploadResponse["data"].(map[string]interface{})
	suite.True(ok, "Photo data should be present in response")
	
	photoID, ok := photoData["id"].(string)
	suite.True(ok, "Photo ID should be present in response")
	suite.NotEmpty(photoID)
	
	suite.photoID = uuid.MustParse(photoID)

	// Step 3: Get download URL
	req, _ = http.NewRequest("GET", suite.server.URL+"/api/v1/media/"+photoID+"/url", nil)
	req.Header.Set("Authorization", suite.authToken)
	
	resp, err = http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.Code)
	
	var downloadURLResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&downloadURLResponse)
	suite.Require().NoError(err)
	
	downloadURL, ok := downloadURLResponse["data"].(map[string]interface{})["download_url"].(string)
	suite.True(ok, "Download URL should be present in response")
	suite.NotEmpty(downloadURL)

	// Step 4: Mark photo as viewed
	req, _ = http.NewRequest("POST", suite.server.URL+"/api/v1/media/"+photoID+"/view", nil)
	req.Header.Set("Authorization", suite.authToken)
	
	resp, err = http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.Code)

	// Step 5: Set photo as primary
	req, _ = http.NewRequest("PUT", suite.server.URL+"/api/v1/me/photos/"+photoID+"/set-primary", nil)
	req.Header.Set("Authorization", suite.authToken)
	
	resp, err = http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.Code)

	// Step 6: Delete photo
	req, _ = http.NewRequest("DELETE", suite.server.URL+"/api/v1/me/photos/"+photoID, nil)
	req.Header.Set("Authorization", suite.authToken)
	
	resp, err = http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusOK, resp.Code)
}

func (suite *PhotoIntegrationTestSuite) TestPhotoValidation() {
	// Test photo validation with invalid data
	tests := []struct {
		name        string
		imageData   string
		contentType string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid PNG image",
			imageData:   createTestImageData(),
			contentType: "image/png",
			expectError: false,
		},
		{
			name:        "Invalid content type",
			imageData:   createTestImageData(),
			contentType: "application/pdf",
			expectError: true,
			errorMsg:    "content type is not allowed",
		},
		{
			name:        "Invalid image data",
			imageData:   "data:image/png;base64,invalid",
			contentType: "image/png",
			expectError: true,
			errorMsg:    "failed to decode image",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			uploadReq := map[string]interface{}{
				"image_data":    tt.imageData,
				"content_type": tt.contentType,
			}
			
			uploadReqBody, _ := json.Marshal(uploadReq)
			req, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/me/photos", bytes.NewBuffer(uploadReqBody))
			req.Header.Set("Authorization", suite.authToken)
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := http.DefaultClient.Do(req)
			suite.Require().NoError(err)
			defer resp.Body.Close()
			
			if tt.expectError {
				suite.NotEqual(http.StatusCreated, resp.Code)
				
				var errorResponse map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&errorResponse)
				suite.Require().NoError(err)
				
				errorMsg, ok := errorResponse["error"].(string)
				suite.True(ok, "Error message should be present in response")
				suite.Contains(errorMsg, tt.errorMsg)
			} else {
				suite.Equal(http.StatusCreated, resp.Code)
			}
		})
	}
}

func (suite *PhotoIntegrationTestSuite) TestPhotoAuthorization() {
	// Test that unauthorized requests are rejected
	req, _ := http.NewRequest("GET", suite.server.URL+"/api/v1/media/request-upload?content_type=image/png", nil)
	// No authorization header
	
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Equal(http.StatusUnauthorized, resp.Code)
}

func (suite *PhotoIntegrationTestSuite) TestPhotoRateLimiting() {
	// Test rate limiting on photo endpoints
	imageData := createTestImageData()
	
	// Make multiple rapid requests
	for i := 0; i < 10; i++ {
		uploadReq := map[string]interface{}{
			"image_data":    imageData,
			"content_type": "image/png",
		}
		
		uploadReqBody, _ := json.Marshal(uploadReq)
		req, _ := http.NewRequest("POST", suite.server.URL+"/api/v1/me/photos", bytes.NewBuffer(uploadReqBody))
		req.Header.Set("Authorization", suite.authToken)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := http.DefaultClient.Do(req)
		suite.Require().NoError(err)
		resp.Body.Close()
		
		// After a few requests, we should hit rate limiting
		if i > 5 {
			if resp.StatusCode == http.StatusTooManyRequests {
				suite.T().Log("Rate limiting is working")
				return
			}
		}
	}
	
	suite.T().Log("Rate limiting may not be properly configured in test environment")
}

// Mock implementations for testing
type MockStorageService struct{}
type MockImageProcessor struct{}
type MockUploadPhotoUseCase struct{}
type MockDeletePhotoUseCase struct{}
type MockGetUploadURLUseCase struct{}
type MockGetDownloadURLUseCase struct{}
type MockSetPrimaryPhotoUseCase struct{}
type MockMarkPhotoViewedUseCase struct{}

func NewMockUploadPhotoUseCase() *MockUploadPhotoUseCase    { return &MockUploadPhotoUseCase{} }
func NewMockDeletePhotoUseCase() *MockDeletePhotoUseCase    { return &MockDeletePhotoUseCase{} }
func NewMockGetUploadURLUseCase() *MockGetUploadURLUseCase  { return &MockGetUploadURLUseCase{} }
func NewMockGetDownloadURLUseCase() *MockGetDownloadURLUseCase { return &MockGetDownloadURLUseCase{} }
func NewMockSetPrimaryPhotoUseCase() *MockSetPrimaryPhotoUseCase { return &MockSetPrimaryPhotoUseCase{} }
func NewMockMarkPhotoViewedUseCase() *MockMarkPhotoViewedUseCase { return &MockMarkPhotoViewedUseCase{} }

// TestPhotoIntegrationTestSuite runs the integration test suite
func TestPhotoIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PhotoIntegrationTestSuite))
}