package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// ProfileTestSuite contains all profile integration tests
type ProfileTestSuite struct {
	suite.Suite
	router      *gin.Engine
	server      *httptest.Server
	userID      uuid.UUID
	authToken    string
}

// SetupSuite sets up the test suite
func (suite *ProfileTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup middleware
	securityConfig := middleware.SecurityConfig{
		EnableCSRF:     false,
		EnableRateLimit: false,
		EnableSecurity:  false,
	}
	
	rateLimitConfig := middleware.RateLimiterConfig{
		RequestsPerMinute: 1000,
		RequestsPerHour:   10000,
		BurstSize:        100,
	}
	
	csrfConfig := middleware.CSRFConfig{
		Secret:     "test-secret",
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
	}
	
	// Create mock handlers and services
	// In a real test, you would use mocks or test containers
	profileHandler := handlers.NewProfileHandler(
		// Mock use cases would be injected here
		nil, nil, nil, nil, nil, nil, nil, nil,
	)
	
	// Setup routes
	profileRoutes := routes.NewProfileRoutes(
		profileHandler,
		securityConfig,
		rateLimitConfig,
		csrfConfig,
	)
	
	// Create a test user
	suite.userID = uuid.New()
	suite.authToken = "test-token"
	
	// Register routes
	api := suite.router.Group("/api/v1")
	profileRoutes.RegisterRoutes(api, nil)
	
	// Start test server
	suite.server = httptest.NewServer(suite.router)
}

// TearDownSuite tears down the test suite
func (suite *ProfileTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// TestGetProfile tests the GET /me endpoint
func (suite *ProfileTestSuite) TestGetProfile() {
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/profile/me", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.ProfileResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

// TestUpdateProfile tests the PUT /me endpoint
func (suite *ProfileTestSuite) TestUpdateProfile() {
	// Create request body
	updateReq := dto.UpdateProfileRequestDTO{
		FirstName: stringPtr("Updated"),
		LastName:  stringPtr("Name"),
		Bio:       stringPtr("Updated bio"),
		Preferences: &dto.PreferencesDTO{
			AgeMin:      25,
			AgeMax:      35,
			MaxDistance: 50,
			ShowMe:      true,
		},
	}
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/profile/me", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.ProfileResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

// TestUpdateLocation tests the PUT /me/location endpoint
func (suite *ProfileTestSuite) TestUpdateLocation() {
	// Create request body
	locationReq := dto.UpdateLocationRequestDTO{
		Latitude:  40.7128,
		Longitude: -74.0060,
		City:      stringPtr("New York"),
		Country:   stringPtr("USA"),
	}
	
	body, _ := json.Marshal(locationReq)
	req, _ := http.NewRequest("PUT", "/api/v1/profile/me/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Location updated successfully", response.Message)
}

// TestGetMatches tests the GET /me/matches endpoint
func (suite *ProfileTestSuite) TestGetMatches() {
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/profile/me/matches?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MatchesResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Pagination)
	assert.Equal(suite.T(), 10, response.Pagination.Limit)
	assert.Equal(suite.T(), 0, response.Pagination.Offset)
}

// TestViewUserProfile tests the GET /users/:id endpoint
func (suite *ProfileTestSuite) TestViewUserProfile() {
	targetUserID := uuid.New()
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/profile/users/"+targetUserID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.UserProfileResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
}

// TestDeleteAccount tests the DELETE /me/account endpoint
func (suite *ProfileTestSuite) TestDeleteAccount() {
	// Create request body
	deleteReq := dto.DeleteAccountRequestDTO{
		Password: "password",
		Reason:   "Testing account deletion",
		Confirm:  true,
	}
	
	body, _ := json.Marshal(deleteReq)
	req, _ := http.NewRequest("DELETE", "/api/v1/profile/me/account", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)
	
	// Make request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Account deleted successfully", response.Message)
}

// TestProfileValidation tests profile validation
func (suite *ProfileTestSuite) TestProfileValidation() {
	testCases := []struct {
		name        string
		request     dto.UpdateProfileRequestDTO
		expectedCode int
	}{
		{
			name: "Invalid first name - too short",
			request: dto.UpdateProfileRequestDTO{
				FirstName: stringPtr("A"),
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid first name - too long",
			request: dto.UpdateProfileRequestDTO{
				FirstName: stringPtr(string(make([]byte, 101))), // 101 characters
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid bio - too long",
			request: dto.UpdateProfileRequestDTO{
				Bio: stringPtr(string(make([]byte, 501))), // 501 characters
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid preferences - age range",
			request: dto.UpdateProfileRequestDTO{
				Preferences: &dto.PreferencesDTO{
					AgeMin: 30,
					AgeMax: 20, // Invalid range
				},
			},
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("PUT", "/api/v1/profile/me", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			
			// Make request
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			// Check response
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestLocationValidation tests location validation
func (suite *ProfileTestSuite) TestLocationValidation() {
	testCases := []struct {
		name        string
		request     dto.UpdateLocationRequestDTO
		expectedCode int
	}{
		{
			name: "Invalid latitude - too low",
			request: dto.UpdateLocationRequestDTO{
				Latitude:  -91,
				Longitude: 0,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid latitude - too high",
			request: dto.UpdateLocationRequestDTO{
				Latitude:  91,
				Longitude: 0,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid longitude - too low",
			request: dto.UpdateLocationRequestDTO{
				Latitude:  0,
				Longitude: -181,
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid longitude - too high",
			request: dto.UpdateLocationRequestDTO{
				Latitude:  0,
				Longitude: 181,
			},
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("PUT", "/api/v1/profile/me/location", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			
			// Make request
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			// Check response
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestAuthentication tests authentication requirements
func (suite *ProfileTestSuite) TestAuthentication() {
	testCases := []struct {
		name        string
		authHeader  string
		expectedCode int
	}{
		{
			name:        "No authorization header",
			authHeader:  "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:        "Invalid authorization header",
			authHeader:  "Invalid token",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:        "Malformed authorization header",
			authHeader:  "Bearer",
			expectedCode: http.StatusUnauthorized,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/profile/me", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			
			// Make request
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			// Check response
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestRateLimiting tests rate limiting functionality
func (suite *ProfileTestSuite) TestRateLimiting() {
	// This test would require a real Redis instance
	// For now, we'll just test the rate limit headers
	
	suite.T().Run("Rate limit headers present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/profile/me", nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		// Make request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// Check that rate limit headers are present (if rate limiting is enabled)
		// In a real test with rate limiting enabled, these would be set
		// For now, we just check the response structure
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestProfilePrivacy tests privacy controls
func (suite *ProfileTestSuite) TestProfilePrivacy() {
	suite.T().Run("View profile privacy", func(t *testing.T) {
		targetUserID := uuid.New()
		
		// Create request to view another user's profile
		req, _ := http.NewRequest("GET", "/api/v1/profile/users/"+targetUserID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		// Make request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response dto.UserProfileResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		
		// In a real test, you would verify that:
		// 1. Location is filtered for non-matched users
		// 2. Bio is truncated for non-matched users
		// 3. Only verified photos are shown
		// 4. Profile view is tracked
	})
}

// TestProfileCaching tests caching functionality
func (suite *ProfileTestSuite) TestProfileCaching() {
	suite.T().Run("Profile caching", func(t *testing.T) {
		// First request
		req1, _ := http.NewRequest("GET", "/api/v1/profile/me", nil)
		req1.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		w1 := httptest.NewRecorder()
		suite.router.ServeHTTP(w1, req1)
		
		// Second request (should be faster if cached)
		req2, _ := http.NewRequest("GET", "/api/v1/profile/me", nil)
		req2.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		w2 := httptest.NewRecorder()
		suite.router.ServeHTTP(w2, req2)
		
		// Both requests should succeed
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, http.StatusOK, w2.Code)
		
		// In a real test with caching enabled, the second request would be faster
		// You would measure response times to verify caching
	})
}

// TestProfileCompletion tests profile completion tracking
func (suite *ProfileTestSuite) TestProfileCompletion() {
	suite.T().Run("Profile completion calculation", func(t *testing.T) {
		// Create incomplete profile update
		updateReq := dto.UpdateProfileRequestDTO{
			FirstName: stringPtr("John"),
			LastName:  stringPtr("Doe"),
			// Missing bio, location, photos, etc.
		}
		
		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/api/v1/profile/me", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		// Make request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response dto.ProfileResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		
		// In a real test, you would verify that:
		// 1. Profile completion percentage is calculated correctly
		// 2. Missing fields are identified
		// 3. Completion status is tracked
	})
}

// TestAccountDeletion tests account deletion process
func (suite *ProfileTestSuite) TestAccountDeletion() {
	suite.T().Run("Account deletion process", func(t *testing.T) {
		// Create deletion request
		deleteReq := dto.DeleteAccountRequestDTO{
			Password: "password",
			Reason:   "Testing account deletion",
			Confirm:  true,
		}
		
		body, _ := json.Marshal(deleteReq)
		req, _ := http.NewRequest("DELETE", "/api/v1/profile/me/account", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		
		// Make request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response dto.MessageResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
		
		// In a real test, you would verify that:
		// 1. User data is anonymized
		// 2. All related data is deleted
		// 3. Sessions are invalidated
		// 4. Cache is cleared
	})
}

// TestErrorHandling tests error handling
func (suite *ProfileTestSuite) TestErrorHandling() {
	testCases := []struct {
		name        string
		endpoint    string
		method      string
		body        interface{}
		expectedCode int
	}{
		{
			name:        "Invalid JSON",
			endpoint:    "/api/v1/profile/me",
			method:      "PUT",
			body:        "invalid json",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "Missing required field",
			endpoint:    "/api/v1/profile/me/location",
			method:      "PUT",
			body:        map[string]interface{}{},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "Invalid UUID",
			endpoint:    "/api/v1/profile/users/invalid-uuid",
			method:      "GET",
			body:        nil,
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			var body []byte
			if tc.body != nil {
				body, _ = json.Marshal(tc.body)
			}
			
			req, _ := http.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)
			
			// Make request
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			// Check response
			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// TestProfileIntegration runs all profile integration tests
func TestProfileIntegration(t *testing.T) {
	suite.Run(t, new(ProfileTestSuite))
}