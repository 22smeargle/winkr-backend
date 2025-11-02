package integration

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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/ephemeral_photo"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// EphemeralPhotoIntegrationTestSuite represents the integration test suite for ephemeral photos
type EphemeralPhotoIntegrationTestSuite struct {
	suite.Suite
	router               *gin.Engine
	ephemeralPhotoRepo   *repositories.MockEphemeralPhotoRepository
	ephemeralPhotoViewRepo *repositories.MockEphemeralPhotoViewRepository
	userRepo             *repositories.MockUserRepository
	storageService       *storage.MockStorageService
	cacheService         *cache.MockCacheService
	ephemeralPhotoService services.EphemeralPhotoService
	ephemeralPhotoHandler *handlers.EphemeralPhotoHandler
	chatHandler          *handlers.ChatHandler
	testUserID           uuid.UUID
	testConversationID   uuid.UUID
	jwtUtils            *utils.JWTUtils
}

// SetupSuite sets up the test suite
func (suite *EphemeralPhotoIntegrationTestSuite) SetupSuite() {
	// Initialize mocks
	suite.ephemeralPhotoRepo = &repositories.MockEphemeralPhotoRepository{}
	suite.ephemeralPhotoViewRepo = &repositories.MockEphemeralPhotoViewRepository{}
	suite.userRepo = &repositories.MockUserRepository{}
	suite.storageService = &storage.MockStorageService{}
	suite.cacheService = &cache.MockCacheService{}

	// Create test user ID
	suite.testUserID = uuid.New()
	suite.testConversationID = uuid.New()

	// Initialize JWT utils
	suite.jwtUtils = utils.NewJWTUtils("test-secret", "test-issuer", 15*time.Minute, 24*time.Hour)

	// Create services
	suite.ephemeralPhotoService = services.NewEphemeralPhotoService(
		suite.ephemeralPhotoRepo,
		suite.ephemeralPhotoViewRepo,
		suite.storageService,
		suite.cacheService,
	)

	// Create handlers
	suite.ephemeralPhotoHandler = handlers.NewEphemeralPhotoHandler(
		suite.ephemeralPhotoService,
	)

	// Create use cases for chat integration
	sendEphemeralPhotoMessageUseCase := ephemeral_photo.NewSendEphemeralPhotoMessageUseCase(
		suite.ephemeralPhotoService,
		nil, // chat service - will be mocked
		nil, // message service - will be mocked
	)

	getEphemeralPhotoMessageUseCase := ephemeral_photo.NewGetEphemeralPhotoMessageUseCase(
		nil, // chat service - will be mocked
	)

	suite.chatHandler = handlers.NewChatHandler(
		nil, // get conversations use case
		nil, // get messages use case
		nil, // send message use case
		nil, // mark read use case
		nil, // delete message use case
		nil, // start conversation use case
		sendEphemeralPhotoMessageUseCase,
		getEphemeralPhotoMessageUseCase,
		nil, // connection manager
	)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware
	authMiddleware := middleware.AuthMiddleware(suite.jwtUtils)
	rateLimitMiddleware := middleware.EphemeralPhotoRateLimitMiddleware(nil)

	// Add routes
	ephemeralPhotoRoutes := routes.NewEphemeralPhotoRoutes(suite.ephemeralPhotoHandler)
	ephemeralPhotoRoutes.RegisterRoutes(suite.router, authMiddleware, rateLimitMiddleware)

	chatRoutes := routes.NewChatRoutes(suite.chatHandler)
	chatRoutes.RegisterRoutes(suite.router, authMiddleware, rateLimitMiddleware)
}

// TearDownSuite cleans up the test suite
func (suite *EphemeralPhotoIntegrationTestSuite) TearDownSuite() {
	// Clean up any test files
	os.RemoveAll("./test-uploads")
}

// TestEphemeralPhotoUploadFlow tests the complete ephemeral photo upload flow
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoUploadFlow() {
	// Create test image file
	testImage := suite.createTestImage("test.jpg", 1024*1024) // 1MB

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file
	part, err := writer.CreateFormFile("file", "test.jpg")
	suite.Require().NoError(err)
	_, err = io.Copy(part, testImage)
	suite.Require().NoError(err)

	// Add form fields
	_ = writer.WriteField("caption", "Test ephemeral photo")
	_ = writer.WriteField("duration_seconds", "30")
	_ = writer.WriteField("max_views", "1")
	
	err = writer.Close()
	suite.Require().NoError(err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ephemeral-photos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Mock repository calls
	mockPhoto := &entities.EphemeralPhoto{
		ID:           uuid.New(),
		UserID:       suite.testUserID,
		AccessKey:    "test-access-key-32-chars-long",
		ThumbnailURL:  "https://cdn.example.com/thumbnail.jpg",
		PhotoURL:      "https://cdn.example.com/photo.jpg",
		ExpiresAt:     time.Now().Add(30 * time.Second),
		MaxViews:     1,
		ViewCount:     0,
		IsViewed:      false,
		IsExpired:     false,
		Caption:       "Test ephemeral photo",
		Duration:      30 * time.Second,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	suite.ephemeralPhotoRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.EphemeralPhoto")).Return(mockPhoto, nil)
	suite.storageService.On("UploadEphemeralPhoto", mock.Anything, mock.Anything, mock.Anything).Return("https://cdn.example.com/photo.jpg", "https://cdn.example.com/thumbnail.jpg", nil)
	suite.cacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.True(response["success"].(bool))
	suite.Equal("Ephemeral photo uploaded successfully", response["message"])
	
	data := response["data"].(map[string]interface{})
	suite.NotEmpty(data["id"])
	suite.Equal("test-access-key-32-chars-long", data["access_key"])
	suite.Equal("https://cdn.example.com/thumbnail.jpg", data["thumbnail_url"])
	suite.Equal(int64(30), data["duration_seconds"])
	suite.Equal(1, data["max_views"])
	suite.Equal(0, data["view_count"])
	suite.False(data["is_viewed"])
	suite.False(data["is_expired"])

	// Verify mock calls
	suite.ephemeralPhotoRepo.AssertExpectations(suite.T())
	suite.storageService.AssertExpectations(suite.T())
	suite.cacheService.AssertExpectations(suite.T())
}

// TestEphemeralPhotoViewFlow tests the ephemeral photo viewing flow
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoViewFlow() {
	photoID := uuid.New()
	accessKey := "test-access-key-32-chars-long"

	// Mock photo
	mockPhoto := &entities.EphemeralPhoto{
		ID:           photoID,
		UserID:       suite.testUserID,
		AccessKey:    accessKey,
		ThumbnailURL:  "https://cdn.example.com/thumbnail.jpg",
		PhotoURL:      "https://cdn.example.com/photo.jpg",
		ExpiresAt:     time.Now().Add(30 * time.Second),
		MaxViews:     1,
		ViewCount:     0,
		IsViewed:      false,
		IsExpired:     false,
		Caption:       "Test ephemeral photo",
		Duration:      30 * time.Second,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	suite.ephemeralPhotoRepo.On("GetByID", mock.Anything, photoID).Return(mockPhoto, nil)
	suite.ephemeralPhotoRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.EphemeralPhoto")).Return(nil)
	suite.ephemeralPhotoViewRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.EphemeralPhotoView")).Return(nil)
	suite.cacheService.On("Get", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("cache miss"))
	suite.cacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/ephemeral-photos/%s/view?access_key=%s", photoID, accessKey), nil)
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.True(response["success"].(bool))
	suite.Equal("Photo retrieved successfully", response["message"])
	
	data := response["data"].(map[string]interface{})
	suite.Equal(photoID.String(), data["id"])
	suite.Equal("https://cdn.example.com/photo.jpg", data["url"])
	suite.Equal("https://cdn.example.com/thumbnail.jpg", data["thumbnail_url"])
	suite.Equal(1, data["view_count"])
	suite.Equal(1, data["max_views"])
	suite.True(data["is_viewed"])
	suite.False(data["is_expired"])
	suite.Equal("Test ephemeral photo", data["caption"])

	// Verify mock calls
	suite.ephemeralPhotoRepo.AssertExpectations(suite.T())
	suite.ephemeralPhotoViewRepo.AssertExpectations(suite.T())
	suite.cacheService.AssertExpectations(suite.T())
}

// TestEphemeralPhotoExpiration tests automatic expiration
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoExpiration() {
	photoID := uuid.New()
	accessKey := "test-access-key-32-chars-long"

	// Mock expired photo
	mockPhoto := &entities.EphemeralPhoto{
		ID:           photoID,
		UserID:       suite.testUserID,
		AccessKey:    accessKey,
		ThumbnailURL:  "https://cdn.example.com/thumbnail.jpg",
		PhotoURL:      "https://cdn.example.com/photo.jpg",
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		MaxViews:     1,
		ViewCount:     0,
		IsViewed:      false,
		IsExpired:     true,
		Caption:       "Test ephemeral photo",
		Duration:      30 * time.Second,
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now().Add(-1 * time.Hour),
	}

	suite.ephemeralPhotoRepo.On("GetByID", mock.Anything, photoID).Return(mockPhoto, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/ephemeral-photos/%s/view?access_key=%s", photoID, accessKey), nil)
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response - should return 410 Gone
	suite.Equal(http.StatusGone, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.False(response["success"].(bool))
	suite.Equal("Photo has expired or been viewed", response["message"])

	// Verify mock calls
	suite.ephemeralPhotoRepo.AssertExpectations(suite.T())
}

// TestEphemeralPhotoChatIntegration tests chat integration
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoChatIntegration() {
	photoID := uuid.New()

	// Mock photo
	mockPhoto := &entities.EphemeralPhoto{
		ID:           photoID,
		UserID:       suite.testUserID,
		AccessKey:    "test-access-key-32-chars-long",
		ThumbnailURL:  "https://cdn.example.com/thumbnail.jpg",
		PhotoURL:      "https://cdn.example.com/photo.jpg",
		ExpiresAt:     time.Now().Add(30 * time.Second),
		MaxViews:     1,
		ViewCount:     0,
		IsViewed:      false,
		IsExpired:     false,
		Caption:       "Test ephemeral photo",
		Duration:      30 * time.Second,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	suite.ephemeralPhotoRepo.On("GetByID", mock.Anything, photoID).Return(mockPhoto, nil)

	// Create request body
	requestBody := map[string]interface{}{
		"photo_id": photoID.String(),
		"message":  "Check this out!",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	// Create request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/chats/%s/ephemeral-photos", suite.testConversationID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	suite.Equal(http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.True(response["success"].(bool))
	suite.Equal("Ephemeral photo message sent successfully", response["message"])
	
	data := response["data"].(map[string]interface{})
	suite.Equal(photoID.String(), data["photo_id"])
	suite.Equal("test-access-key-32-chars-long", data["access_key"])
	suite.Equal("https://cdn.example.com/thumbnail.jpg", data["thumbnail_url"])
	suite.Equal("Check this out!", data["message"])

	// Verify mock calls
	suite.ephemeralPhotoRepo.AssertExpectations(suite.T())
}

// TestEphemeralPhotoRateLimiting tests rate limiting
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoRateLimiting() {
	// Create test image file
	testImage := suite.createTestImage("test.jpg", 1024*1024) // 1MB

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file
	part, err := writer.CreateFormFile("file", "test.jpg")
	suite.Require().NoError(err)
	_, err = io.Copy(part, testImage)
	suite.Require().NoError(err)
	
	err = writer.Close()
	suite.Require().NoError(err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ephemeral-photos", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Mock rate limiter to return rate limited
	mockRateLimiter := &cache.MockRateLimiter{}
	mockRateLimiter.On("Allow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

	// Create router with rate limiting middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := middleware.AuthMiddleware(suite.jwtUtils)
	rateLimitMiddleware := middleware.EphemeralPhotoRateLimitMiddleware(mockRateLimiter)
	
	ephemeralPhotoRoutes := routes.NewEphemeralPhotoRoutes(suite.ephemeralPhotoHandler)
	ephemeralPhotoRoutes.RegisterRoutes(router, authMiddleware, rateLimitMiddleware)

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response - should return 429 Too Many Requests
	suite.Equal(http.StatusTooManyRequests, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.False(response["success"].(bool))
	suite.Equal("Rate limit exceeded", response["message"])

	// Verify mock calls
	mockRateLimiter.AssertExpectations(suite.T())
}

// TestEphemeralPhotoSecurityTests tests security features
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoSecurityTests() {
	photoID := uuid.New()

	// Test 1: Invalid access key
	suite.Run("InvalidAccessKey", func() {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/ephemeral-photos/%s/view?access_key=invalid-key", photoID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusForbidden, w.Code)
	})

	// Test 2: Missing access key
	suite.Run("MissingAccessKey", func() {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/ephemeral-photos/%s/view", photoID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusBadRequest, w.Code)
	})

	// Test 3: Unauthorized access (different user)
	suite.Run("UnauthorizedAccess", func() {
		otherUserID := uuid.New()
		mockPhoto := &entities.EphemeralPhoto{
			ID:     photoID,
			UserID: otherUserID, // Different user
			ExpiresAt: time.Now().Add(30 * time.Second),
		}

		suite.ephemeralPhotoRepo.On("GetByID", mock.Anything, photoID).Return(mockPhoto, nil)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/ephemeral-photos/%s", photoID), nil)
		req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusForbidden, w.Code)
	})
}

// TestEphemeralPhotoBackgroundJobs tests background job functionality
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoBackgroundJobs() {
	// Create expired photos
	expiredPhotos := []*entities.EphemeralPhoto{
		{
			ID:        uuid.New(),
			UserID:    suite.testUserID,
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			IsExpired: false,
		},
		{
			ID:        uuid.New(),
			UserID:    suite.testUserID,
			ExpiresAt: time.Now().Add(-2 * time.Hour),
			IsExpired: false,
		},
	}

	// Mock repository calls
	suite.ephemeralPhotoRepo.On("GetExpiredPhotos", mock.Anything, mock.Anything).Return(expiredPhotos, nil)
	suite.ephemeralPhotoRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.EphemeralPhoto")).Return(nil).Twice()
	suite.storageService.On("DeleteEphemeralPhoto", mock.Anything, mock.Anything).Return(nil).Twice()

	// Create background service
	backgroundService := services.NewEphemeralPhotoBackgroundService(
		suite.ephemeralPhotoRepo,
		suite.ephemeralPhotoViewRepo,
		suite.storageService,
		suite.cacheService,
	)

	// Run cleanup job
	ctx := context.Background()
	err := backgroundService.CleanupExpiredPhotos(ctx)
	suite.NoError(err)

	// Verify mock calls
	suite.ephemeralPhotoRepo.AssertExpectations(suite.T())
	suite.storageService.AssertExpectations(suite.T())
}

// TestEphemeralPhotoCaching tests caching functionality
func (suite *EphemeralPhotoIntegrationTestSuite) TestEphemeralPhotoCaching() {
	photoID := uuid.New()
	accessKey := "test-access-key-32-chars-long"

	// Mock photo
	mockPhoto := &entities.EphemeralPhoto{
		ID:           photoID,
		UserID:       suite.testUserID,
		AccessKey:    accessKey,
		ThumbnailURL:  "https://cdn.example.com/thumbnail.jpg",
		PhotoURL:      "https://cdn.example.com/photo.jpg",
		ExpiresAt:     time.Now().Add(30 * time.Second),
		MaxViews:     1,
		ViewCount:     0,
		IsViewed:      false,
		IsExpired:     false,
		Caption:       "Test ephemeral photo",
		Duration:      30 * time.Second,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test cache hit
	suite.cacheService.On("Get", mock.Anything, mock.Anything).Return(mockPhoto, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/ephemeral-photos/%s/status", photoID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.generateTestToken())

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	
	suite.True(response["success"].(bool))
	data := response["data"].(map[string]interface{})
	suite.Equal(photoID.String(), data["id"])

	// Verify cache was checked
	suite.cacheService.AssertExpectations(suite.T())
}

// Helper methods

// generateTestToken generates a test JWT token
func (suite *EphemeralPhotoIntegrationTestSuite) generateTestToken() string {
	token, err := suite.jwtUtils.GenerateAccessToken(suite.testUserID)
	suite.Require().NoError(err)
	return token
}

// createTestImage creates a test image file
func (suite *EphemeralPhotoIntegrationTestSuite) createTestImage(filename string, size int) io.Reader {
	// Create a simple test image
	imageData := make([]byte, size)
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}
	return bytes.NewReader(imageData)
}

// TestEphemeralPhotoIntegrationTestSuite runs the integration test suite
func TestEphemeralPhotoIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EphemeralPhotoIntegrationTestSuite))
}