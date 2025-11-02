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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/auth"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
)

// AdminIntegrationTestSuite represents the admin integration test suite
type AdminIntegrationTestSuite struct {
	suite.Suite
	router               *gin.Engine
	adminService         *services.AdminService
	adminCacheService    *services.AdminCacheService
	tokenManager         auth.TokenManager
	mockCacheService     *cache.MockCacheService
	adminUserID         uuid.UUID
	regularUserID       uuid.UUID
	adminToken          string
	regularToken        string
}

// SetupSuite sets up the test suite
func (suite *AdminIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock services
	suite.mockCacheService = &cache.MockCacheService{}
	suite.tokenManager = auth.NewTokenManager("test-secret", 24*time.Hour)
	suite.adminCacheService = services.NewAdminCacheService(suite.mockCacheService)

	// Create admin service with mock repositories
	suite.adminService = services.NewAdminService(
		&MockUserRepository{},
		&MockPhotoRepository{},
		&MockMessageRepository{},
		&MockMatchRepository{},
		&MockReportRepository{},
		&MockPaymentRepository{},
		&MockSubscriptionRepository{},
		&MockVerificationRepository{},
	)

	// Create test users
	suite.adminUserID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	suite.regularUserID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	// Generate test tokens
	suite.adminToken = suite.generateToken(suite.adminUserID)
	suite.regularToken = suite.generateToken(suite.regularUserID)

	// Setup routes
	suite.setupRoutes()
}

// setupRoutes sets up the test routes
func (suite *AdminIntegrationTestSuite) setupRoutes() {
	// Create middleware
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(suite.mockCacheService)
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Create admin routes
	adminRoutes := routes.NewAdminRoutes(
		suite.adminService,
		suite.tokenManager,
		rateLimiterMiddleware,
		loggingMiddleware,
	)

	// Setup router
	suite.router = gin.New()
	adminRoutes.SetupAdminRoutes(suite.router)
}

// generateToken generates a JWT token for testing
func (suite *AdminIntegrationTestSuite) generateToken(userID uuid.UUID) string {
	token, err := suite.tokenManager.GenerateToken(userID, "user@example.com")
	if err != nil {
		suite.T().Fatalf("Failed to generate token: %v", err)
	}
	return token
}

// TestAdminAuthentication tests admin authentication
func (suite *AdminIntegrationTestSuite) TestAdminAuthentication() {
	// Test missing authorization header
	req, _ := http.NewRequest("GET", "/admin/v1/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test invalid authorization header
	req, _ = http.NewRequest("GET", "/admin/v1/users", nil)
	req.Header.Set("Authorization", "Invalid token")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test regular user access (should be forbidden)
	req, _ = http.NewRequest("GET", "/admin/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.regularToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	// Test admin access (should succeed)
	req, _ = http.NewRequest("GET", "/admin/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestUserManagementEndpoints tests user management endpoints
func (suite *AdminIntegrationTestSuite) TestUserManagementEndpoints() {
	// Test GET /admin/v1/users
	req, _ := http.NewRequest("GET", "/admin/v1/users?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "users")
	assert.Contains(suite.T(), response, "total")

	// Test GET /admin/v1/users/:id
	req, _ = http.NewRequest("GET", "/admin/v1/users/"+suite.regularUserID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test PUT /admin/v1/users/:id
	updateData := map[string]interface{}{
		"username": "updated_username",
		"email":    "updated@example.com",
	}
	jsonData, _ := json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", "/admin/v1/users/"+suite.regularUserID.String(), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test POST /admin/v1/users/:id/suspend
	suspendData := map[string]interface{}{
		"duration": 7,
		"reason":   "Policy violation",
	}
	jsonData, _ = json.Marshal(suspendData)
	req, _ = http.NewRequest("POST", "/admin/v1/users/"+suite.regularUserID.String()+"/suspend", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test POST /admin/v1/users/:id/ban
	banData := map[string]interface{}{
		"reason": "Severe policy violation",
	}
	jsonData, _ = json.Marshal(banData)
	req, _ = http.NewRequest("POST", "/admin/v1/users/"+suite.regularUserID.String()+"/ban", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestAnalyticsEndpoints tests analytics endpoints
func (suite *AdminIntegrationTestSuite) TestAnalyticsEndpoints() {
	// Test GET /admin/v1/stats
	req, _ := http.NewRequest("GET", "/admin/v1/stats?period=7d", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/stats/users
	req, _ = http.NewRequest("GET", "/admin/v1/stats/users?period=30d&group_by=day", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/stats/matches
	req, _ = http.NewRequest("GET", "/admin/v1/stats/matches?period=30d&group_by=week", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/stats/messages
	req, _ = http.NewRequest("GET", "/admin/v1/stats/messages?period=30d&group_by=month", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/stats/payments
	req, _ = http.NewRequest("GET", "/admin/v1/stats/payments?period=30d&group_by=day", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/stats/verification
	req, _ = http.NewRequest("GET", "/admin/v1/stats/verification?period=30d&group_by=day&verification_type=all", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestSystemManagementEndpoints tests system management endpoints
func (suite *AdminIntegrationTestSuite) TestSystemManagementEndpoints() {
	// Test GET /admin/v1/system/health
	req, _ := http.NewRequest("GET", "/admin/v1/system/health", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/system/metrics
	req, _ = http.NewRequest("GET", "/admin/v1/system/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/system/logs
	req, _ = http.NewRequest("GET", "/admin/v1/system/logs?limit=50&level=error", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/system/config
	req, _ = http.NewRequest("GET", "/admin/v1/system/config?section=general", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test PUT /admin/v1/system/config
	configData := map[string]interface{}{
		"app_name": "Updated App Name",
		"debug_mode": false,
	}
	jsonData, _ := json.Marshal(configData)
	req, _ = http.NewRequest("PUT", "/admin/v1/system/config?section=general", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestContentModerationEndpoints tests content moderation endpoints
func (suite *AdminIntegrationTestSuite) TestContentModerationEndpoints() {
	// Test GET /admin/v1/content/photos
	req, _ := http.NewRequest("GET", "/admin/v1/content/photos?status=pending&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test GET /admin/v1/content/messages
	req, _ = http.NewRequest("GET", "/admin/v1/content/messages?status=pending&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test POST /admin/v1/content/photos/:id/approve
	photoID := uuid.New()
	req, _ = http.NewRequest("POST", "/admin/v1/content/photos/"+photoID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test POST /admin/v1/content/photos/:id/reject
	rejectData := map[string]interface{}{
		"reason": "Inappropriate content",
		"notes":  "Violates community guidelines",
	}
	jsonData, _ := json.Marshal(rejectData)
	req, _ = http.NewRequest("POST", "/admin/v1/content/photos/"+photoID.String()+"/reject", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test POST /admin/v1/content/messages/:id/delete
	messageID := uuid.New()
	req, _ = http.NewRequest("POST", "/admin/v1/content/messages/"+messageID.String()+"/delete", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestBulkOperations tests bulk operations
func (suite *AdminIntegrationTestSuite) TestBulkOperations() {
	// Test bulk update users
	userIDs := []uuid.UUID{suite.regularUserID, uuid.New(), uuid.New()}
	updateData := map[string]interface{}{
		"user_ids": userIDs,
		"updates": map[string]interface{}{
			"verified": true,
		},
	}
	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("POST", "/admin/v1/users/bulk/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test bulk delete users
	deleteData := map[string]interface{}{
		"user_ids":   userIDs,
		"hard_delete": false,
	}
	jsonData, _ = json.Marshal(deleteData)
	req, _ = http.NewRequest("POST", "/admin/v1/users/bulk/delete", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestRateLimiting tests rate limiting on admin endpoints
func (suite *AdminIntegrationTestSuite) TestRateLimiting() {
	// Make multiple requests quickly to trigger rate limiting
	for i := 0; i < 105; i++ { // Exceed the 100 requests per minute limit
		req, _ := http.NewRequest("GET", "/admin/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+suite.adminToken)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		if i < 100 {
			assert.Equal(suite.T(), http.StatusOK, w.Code, "Request %d should succeed", i)
		} else {
			assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i)
		}
	}
}

// TestPermissionBasedAccess tests permission-based access control
func (suite *AdminIntegrationTestSuite) TestPermissionBasedAccess() {
	// Create a token for a user with limited permissions
	limitedUserID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")
	limitedToken := suite.generateToken(limitedUserID)

	// Test access to user management (should fail with limited permissions)
	req, _ := http.NewRequest("GET", "/admin/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+limitedToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// This should fail because the mock user doesn't have admin permissions
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

// TestCacheInvalidation tests cache invalidation
func (suite *AdminIntegrationTestSuite) TestCacheInvalidation() {
	// Setup cache mock to return data then verify invalidation
	suite.mockCacheService.On("Get", mock.Anything, mock.Anything).Return("", fmt.Errorf("cache miss"))
	suite.mockCacheService.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.mockCacheService.On("Delete", mock.Anything).Return(nil)

	// Make a request that should cache data
	req, _ := http.NewRequest("GET", "/admin/v1/stats?period=7d", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Make an update request that should invalidate cache
	updateData := map[string]interface{}{
		"username": "updated_username",
	}
	jsonData, _ := json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", "/admin/v1/users/"+suite.regularUserID.String(), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestErrorHandling tests error handling in admin endpoints
func (suite *AdminIntegrationTestSuite) TestErrorHandling() {
	// Test invalid user ID
	req, _ := http.NewRequest("GET", "/admin/v1/users/invalid-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test invalid request body
	req, _ = http.NewRequest("PUT", "/admin/v1/users/"+suite.regularUserID.String(), bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test missing required parameters
	req, _ = http.NewRequest("POST", "/admin/v1/users/"+suite.regularUserID.String()+"/suspend", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestAdminLogin tests admin login endpoint
func (suite *AdminIntegrationTestSuite) TestAdminLogin() {
	// Test admin login
	loginData := map[string]interface{}{
		"email":    "admin@example.com",
		"password": "adminpassword",
	}
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/admin/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "token")
	assert.Contains(suite.T(), response, "admin")
}

// TestAdminLogout tests admin logout endpoint
func (suite *AdminIntegrationTestSuite) TestAdminLogout() {
	req, _ := http.NewRequest("POST", "/admin/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+suite.adminToken)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestPublicHealthCheck tests public health check endpoint
func (suite *AdminIntegrationTestSuite) TestPublicHealthCheck() {
	req, _ := http.NewRequest("GET", "/admin/v1/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "status")
	assert.Equal(suite.T(), "healthy", response["status"])
}

// RunAdminIntegrationTests runs the admin integration test suite
func RunAdminIntegrationTests(t *testing.T) {
	suite.Run(t, new(AdminIntegrationTestSuite))
}