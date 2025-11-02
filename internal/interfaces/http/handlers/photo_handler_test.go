package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/photo"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// MockUploadPhotoUseCase is a mock implementation of the upload photo use case
type MockUploadPhotoUseCase struct {
	mock.Mock
}

func (m *MockUploadPhotoUseCase) Execute(ctx context.Context, userID uuid.UUID, imageData []byte, contentType string, options ...photo.UploadOption) (*entities.Photo, error) {
	args := m.Called(ctx, userID, imageData, contentType, options)
	return args.Get(0).(*entities.Photo), args.Error(1)
}

// MockDeletePhotoUseCase is a mock implementation of the delete photo use case
type MockDeletePhotoUseCase struct {
	mock.Mock
}

func (m *MockDeletePhotoUseCase) Execute(ctx context.Context, userID, photoID uuid.UUID) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

// MockGetUploadURLUseCase is a mock implementation of the get upload URL use case
type MockGetUploadURLUseCase struct {
	mock.Mock
}

func (m *MockGetUploadURLUseCase) Execute(ctx context.Context, userID uuid.UUID, contentType string) (string, error) {
	args := m.Called(ctx, userID, contentType)
	return args.String(0), args.Error(1)
}

// MockGetDownloadURLUseCase is a mock implementation of the get download URL use case
type MockGetDownloadURLUseCase struct {
	mock.Mock
}

func (m *MockGetDownloadURLUseCase) Execute(ctx context.Context, userID, photoID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID, photoID)
	return args.String(0), args.Error(1)
}

// MockSetPrimaryPhotoUseCase is a mock implementation of the set primary photo use case
type MockSetPrimaryPhotoUseCase struct {
	mock.Mock
}

func (m *MockSetPrimaryPhotoUseCase) Execute(ctx context.Context, userID, photoID uuid.UUID) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

// MockMarkPhotoViewedUseCase is a mock implementation of the mark photo viewed use case
type MockMarkPhotoViewedUseCase struct {
	mock.Mock
}

func (m *MockMarkPhotoViewedUseCase) Execute(ctx context.Context, userID, photoID uuid.UUID) error {
	args := m.Called(ctx, userID, photoID)
	return args.Error(0)
}

func setupTestRouter(handler *PhotoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add middleware to set user context
	router.Use(func(c *gin.Context) {
		userID := uuid.New()
		c.Set("user_id", userID)
		c.Set("user_role", "user")
		c.Next()
	})
	
	return router
}

func TestNewPhotoHandler(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUpload, handler.uploadPhotoUseCase)
	assert.Equal(t, mockDelete, handler.deletePhotoUseCase)
	assert.Equal(t, mockGetUploadURL, handler.getUploadURLUseCase)
	assert.Equal(t, mockGetDownloadURL, handler.getDownloadURLUseCase)
	assert.Equal(t, mockSetPrimary, handler.setPrimaryPhotoUseCase)
	assert.Equal(t, mockMarkViewed, handler.markPhotoViewedUseCase)
	assert.Equal(t, jwtUtils, handler.jwtUtils)
}

func TestPhotoHandler_UploadPhoto(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.POST("/me/photos", handler.UploadPhoto)

	tests := []struct {
		name           string
		requestBody    interface{}
		contentType    string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Successful upload",
			requestBody: map[string]interface{}{
				"image_data":    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k=",
				"content_type": "image/jpeg",
			},
			contentType: "application/json",
			setupMocks: func() {
				photo := &entities.Photo{
					ID:           uuid.New(),
					UserID:       uuid.New(),
					StorageKey:   "test-key",
					ThumbnailKey: "test-thumb-key",
					ContentType:  "image/jpeg",
					Width:        800,
					Height:       600,
					FileSize:     1024,
					IsPrimary:    false,
					CreatedAt:    time.Now(),
				}
				mockUpload.On("Execute", mock.Anything, mock.Anything, mock.AnythingOfType("[]uint8"), "image/jpeg", mock.Anything).Return(photo, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			contentType: "application/json",
			setupMocks: func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
		},
		{
			name: "Missing image data",
			requestBody: map[string]interface{}{
				"content_type": "image/jpeg",
			},
			contentType: "application/json",
			setupMocks: func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Image data is required",
		},
		{
			name: "Upload use case error",
			requestBody: map[string]interface{}{
				"image_data":    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k=",
				"content_type": "image/jpeg",
			},
			contentType: "application/json",
			setupMocks: func() {
				mockUpload.On("Execute", mock.Anything, mock.Anything, mock.AnythingOfType("[]uint8"), "image/jpeg", mock.Anything).Return(nil, errors.New("upload failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to upload photo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUpload.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request
			var body bytes.Buffer
			if tt.requestBody != nil {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/me/photos", &body)
			req.Header.Set("Content-Type", tt.contentType)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.NotNil(t, response["data"])
			}

			// Verify mock expectations
			mockUpload.AssertExpectations(t)
		})
	}
}

func TestPhotoHandler_DeletePhoto(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.DELETE("/me/photos/:id", handler.DeletePhoto)

	tests := []struct {
		name           string
		photoID        string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Successful deletion",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockDelete.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid photo ID",
			photoID:        "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid photo ID",
		},
		{
			name:    "Delete use case error",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockDelete.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("photo not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to delete photo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockDelete.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/me/photos/"+tt.photoID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.Equal(t, "success", response["status"])
			}

			// Verify mock expectations
			mockDelete.AssertExpectations(t)
		})
	}
}

func TestPhotoHandler_GetUploadURL(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.GET("/media/request-upload", handler.GetUploadURL)

	tests := []struct {
		name           string
		queryParams    map[string]string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Successful upload URL generation",
			queryParams: map[string]string{
				"content_type": "image/jpeg",
			},
			setupMocks: func() {
				mockGetUploadURL.On("Execute", mock.Anything, mock.Anything, "image/jpeg").Return("https://example.com/upload-url", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Missing content type",
			queryParams: map[string]string{},
			setupMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content type is required",
		},
		{
			name: "Upload URL use case error",
			queryParams: map[string]string{
				"content_type": "image/jpeg",
			},
			setupMocks: func() {
				mockGetUploadURL.On("Execute", mock.Anything, mock.Anything, "image/jpeg").Return("", errors.New("failed to generate URL"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to generate upload URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockGetUploadURL.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/media/request-upload", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.NotNil(t, response["data"])
				assert.Equal(t, "https://example.com/upload-url", response["data"].(map[string]interface{})["upload_url"])
			}

			// Verify mock expectations
			mockGetUploadURL.AssertExpectations(t)
		})
	}
}

func TestPhotoHandler_GetDownloadURL(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.GET("/media/:id/url", handler.GetDownloadURL)

	tests := []struct {
		name           string
		photoID        string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Successful download URL generation",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockGetDownloadURL.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("https://example.com/download-url", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid photo ID",
			photoID:        "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid photo ID",
		},
		{
			name:    "Download URL use case error",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockGetDownloadURL.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("photo not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to generate download URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockGetDownloadURL.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/media/"+tt.photoID+"/url", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.NotNil(t, response["data"])
				assert.Equal(t, "https://example.com/download-url", response["data"].(map[string]interface{})["download_url"])
			}

			// Verify mock expectations
			mockGetDownloadURL.AssertExpectations(t)
		})
	}
}

func TestPhotoHandler_SetPrimaryPhoto(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.PUT("/me/photos/:id/set-primary", handler.SetPrimaryPhoto)

	tests := []struct {
		name           string
		photoID        string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Successful set primary photo",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockSetPrimary.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid photo ID",
			photoID:        "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid photo ID",
		},
		{
			name:    "Set primary use case error",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockSetPrimary.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("photo not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to set primary photo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockSetPrimary.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodPut, "/me/photos/"+tt.photoID+"/set-primary", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.Equal(t, "success", response["status"])
			}

			// Verify mock expectations
			mockSetPrimary.AssertExpectations(t)
		})
	}
}

func TestPhotoHandler_MarkPhotoViewed(t *testing.T) {
	mockUpload := &MockUploadPhotoUseCase{}
	mockDelete := &MockDeletePhotoUseCase{}
	mockGetUploadURL := &MockGetUploadURLUseCase{}
	mockGetDownloadURL := &MockGetDownloadURLUseCase{}
	mockSetPrimary := &MockSetPrimaryPhotoUseCase{}
	mockMarkViewed := &MockMarkPhotoViewedUseCase{}
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24)

	handler := NewPhotoHandler(
		mockUpload,
		mockDelete,
		mockGetUploadURL,
		mockGetDownloadURL,
		mockSetPrimary,
		mockMarkViewed,
		jwtUtils,
	)

	router := setupTestRouter(handler)
	router.POST("/media/:id/view", handler.MarkPhotoViewed)

	tests := []struct {
		name           string
		photoID        string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Successful mark photo viewed",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockMarkViewed.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid photo ID",
			photoID:        "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid photo ID",
		},
		{
			name:    "Mark viewed use case error",
			photoID: uuid.New().String(),
			setupMocks: func() {
				mockMarkViewed.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("photo not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to mark photo as viewed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockMarkViewed.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/media/"+tt.photoID+"/view", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				assert.Equal(t, "success", response["status"])
			}

			// Verify mock expectations
			mockMarkViewed.AssertExpectations(t)
		})
	}
}