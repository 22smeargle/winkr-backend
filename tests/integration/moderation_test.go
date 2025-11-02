package integration

import (
	"bytes"
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

	"github.com/22smeargle/winkr-backend/internal/application/usecases/moderation"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/validator"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ModerationIntegrationTestSuite tests moderation endpoints
type ModerationIntegrationTestSuite struct {
	suite.Suite
	router               *gin.Engine
	moderationHandler    *handlers.ModerationHandler
	adminModerationHandler *handlers.AdminModerationHandler
	redisClient          *redis.RedisClient
	cacheService         *cache.CacheService
	moderationCacheService *services.ModerationCacheService
	aiModerationService  *external.AIModerationService
	contentAnalysisService *services.ContentAnalysisService
	moderationService    *services.ModerationService
}

// SetupSuite sets up test suite
func (suite *ModerationIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	suite.redisClient = redis.NewMockRedisClient()
	suite.cacheService = cache.NewCacheService(suite.redisClient)
	suite.moderationCacheService = services.NewModerationCacheService(suite.cacheService)
	suite.aiModerationService = external.NewMockAIModerationService()
	suite.contentAnalysisService = services.NewContentAnalysisService(
		suite.aiModerationService,
		suite.moderationCacheService,
		nil, // rate limiter
		nil, // profanity filter
		nil, // link analyzer
		nil, // PII detector
		services.ContentAnalysisConfig{},
	)
	suite.moderationService = services.NewModerationService(
		nil, // report repo
		nil, // user repo
		suite.aiModerationService,
		suite.contentAnalysisService,
		suite.moderationCacheService,
		nil, // notification service
		services.ModerationConfig{},
	)

	// Create use cases
	reportContentUseCase := moderation.NewReportContentUseCase(
		nil, // report repo
		suite.moderationService,
		suite.moderationCacheService,
	)
	blockUserUseCase := moderation.NewBlockUserUseCase(
		nil, // user repo
		nil, // block repo
		suite.moderationService,
		suite.moderationCacheService,
	)
	unblockUserUseCase := moderation.NewUnblockUserUseCase(
		nil, // user repo
		nil, // block repo
		suite.moderationService,
		suite.moderationCacheService,
	)
	getBlockedUsersUseCase := moderation.NewGetBlockedUsersUseCase(
		nil, // user repo
		nil, // block repo
		suite.moderationCacheService,
	)
	getMyReportsUseCase := moderation.NewGetMyReportsUseCase(
		nil, // report repo
		suite.moderationCacheService,
	)
	reviewReportUseCase := moderation.NewReviewReportUseCase(
		nil, // report repo
		nil, // user repo
		suite.moderationService,
		suite.moderationCacheService,
	)
	banUserUseCase := moderation.NewBanUserUseCase(
		nil, // user repo
		nil, // ban repo
		nil, // appeal repo
		suite.moderationService,
		suite.moderationCacheService,
		nil, // notification service
	)

	// Create validators
	moderationValidator := validator.NewModerationValidator()

	// Create handlers
	suite.moderationHandler = handlers.NewModerationHandler(
		reportContentUseCase,
		blockUserUseCase,
		unblockUserUseCase,
		getBlockedUsersUseCase,
		getMyReportsUseCase,
		moderationValidator,
	)

	suite.adminModerationHandler = handlers.NewAdminModerationHandler(
		reviewReportUseCase,
		banUserUseCase,
		moderationValidator,
	)

	// Create router
	suite.router = gin.New()
	
	// Add moderation routes
	moderationGroup := suite.router.Group("/api/v1/moderation")
	{
		moderationGroup.POST("/report", suite.moderationHandler.ReportContent)
		moderationGroup.POST("/block/:id", suite.moderationHandler.BlockUser)
		moderationGroup.DELETE("/block/:id", suite.moderationHandler.UnblockUser)
		moderationGroup.GET("/me/blocked", suite.moderationHandler.GetBlockedUsers)
		moderationGroup.GET("/me/reports", suite.moderationHandler.GetMyReports)
		moderationGroup.GET("/reports/:id", suite.moderationHandler.GetReportStatus)
		moderationGroup.DELETE("/reports/:id", suite.moderationHandler.CancelReport)
		moderationGroup.GET("/block/:id/status", suite.moderationHandler.CheckBlockStatus)
		moderationGroup.GET("/block/:id/mutual", suite.moderationHandler.CheckMutualBlock)
	}

	// Add admin moderation routes
	adminGroup := suite.router.Group("/api/v1/admin")
	{
		adminGroup.GET("/reports", suite.adminModerationHandler.GetReports)
		adminGroup.GET("/reports/:id", suite.adminModerationHandler.GetReportDetails)
		adminGroup.POST("/reports/:id/review", suite.adminModerationHandler.ReviewReport)
		adminGroup.POST("/users/:id/ban", suite.adminModerationHandler.BanUser)
		adminGroup.POST("/users/:id/suspend", suite.adminModerationHandler.SuspendUser)
		adminGroup.GET("/users/:id/bans", suite.adminModerationHandler.GetBanHistory)
		adminGroup.GET("/users/:id/appeals", suite.adminModerationHandler.GetAppealHistory)
		adminGroup.POST("/appeals/:id/review", suite.adminModerationHandler.ReviewAppeal)
		adminGroup.GET("/moderation-queue", suite.adminModerationHandler.GetModerationQueue)
		adminGroup.GET("/analytics", suite.adminModerationHandler.GetModerationAnalytics)
	}
}

// TestReportContentFlow tests the complete report content flow
func (suite *ModerationIntegrationTestSuite) TestReportContentFlow() {
	// Create test users
	reporterID := uuid.New()
	reportedUserID := uuid.New()

	// Prepare report request
	req := moderation.ReportContentRequest{
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		Reason:         "inappropriate_behavior",
		Description:     stringPtr("User sent inappropriate messages"),
		ContentType:    "message",
		ContentID:      stringPtr(uuid.New().String()),
		Evidence:       stringPtr("Screenshot of inappropriate content"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestBlockUserFlow tests the complete block user flow
func (suite *ModerationIntegrationTestSuite) TestBlockUserFlow() {
	// Create test users
	blockerID := uuid.New()
	blockedID := uuid.New()

	// Prepare block request
	req := moderation.BlockUserRequest{
		BlockerID: blockerID,
		BlockedID: blockedID,
		Reason:    stringPtr("Harassment"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/moderation/block/%s", blockedID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestUnblockUserFlow tests the complete unblock user flow
func (suite *ModerationIntegrationTestSuite) TestUnblockUserFlow() {
	// First block a user
	suite.TestBlockUserFlow()

	// Create test users
	blockerID := uuid.New()
	blockedID := uuid.New()

	// Prepare unblock request
	httpReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/moderation/block/%s", blockedID.String()), nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestGetBlockedUsersFlow tests the get blocked users flow
func (suite *ModerationIntegrationTestSuite) TestGetBlockedUsersFlow() {
	// First block some users
	suite.TestBlockUserFlow()

	// Create test user
	userID := uuid.New()

	// Prepare get blocked users request
	httpReq := httptest.NewRequest("GET", "/api/v1/moderation/me/blocked?limit=10&offset=0", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestGetMyReportsFlow tests the get my reports flow
func (suite *ModerationIntegrationTestSuite) TestGetMyReportsFlow() {
	// First create some reports
	suite.TestReportContentFlow()

	// Create test user
	userID := uuid.New()

	// Prepare get my reports request
	httpReq := httptest.NewRequest("GET", "/api/v1/moderation/me/reports?limit=10&offset=0", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetReportsFlow tests the admin get reports flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetReportsFlow() {
	// First create some reports
	suite.TestReportContentFlow()

	// Prepare admin get reports request
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/reports?status=pending&limit=10&offset=0", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetReportDetailsFlow tests the admin get report details flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetReportDetailsFlow() {
	// First create a report
	suite.TestReportContentFlow()
	reportID := uuid.New()

	// Prepare admin get report details request
	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/reports/%s", reportID.String()), nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminReviewReportFlow tests the admin review report flow
func (suite *ModerationIntegrationTestSuite) TestAdminReviewReportFlow() {
	// First create a report
	suite.TestReportContentFlow()
	reportID := uuid.New()
	adminID := uuid.New()

	// Prepare admin review report request
	req := moderation.ReviewReportRequest{
		ReportID:    reportID,
		ReviewerID:   adminID,
		Action:       "resolve",
		Reason:       "Content violates community guidelines",
		Notes:        stringPtr("User sent inappropriate messages"),
		TakeAction:   true,
		ActionType:   stringPtr("warning"),
		ActionDuration: stringPtr("7d"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/reports/%s/review", reportID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminBanUserFlow tests the admin ban user flow
func (suite *ModerationIntegrationTestSuite) TestAdminBanUserFlow() {
	// Create test users
	userID := uuid.New()
	adminID := uuid.New()

	// Prepare admin ban user request
	req := moderation.BanUserRequest{
		UserID:        userID,
		BannerID:      adminID,
		Reason:        "Repeated harassment",
		Duration:      stringPtr("30d"),
		Notes:         stringPtr("User has been warned multiple times"),
		Evidence:      stringPtr("Screenshots of harassment"),
		NotifyUser:    true,
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/users/%s/ban", userID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminSuspendUserFlow tests the admin suspend user flow
func (suite *ModerationIntegrationTestSuite) TestAdminSuspendUserFlow() {
	// Create test users
	userID := uuid.New()
	adminID := uuid.New()

	// Prepare admin suspend user request
	req := moderation.BanUserRequest{
		UserID:        userID,
		BannerID:      adminID,
		Reason:        "Temporary violation",
		Notes:         stringPtr("User violated guidelines temporarily"),
		Evidence:      stringPtr("Evidence of violation"),
		NotifyUser:    true,
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/users/%s/suspend", userID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetBanHistoryFlow tests the admin get ban history flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetBanHistoryFlow() {
	// First ban a user
	suite.TestAdminBanUserFlow()
	userID := uuid.New()

	// Prepare admin get ban history request
	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/users/%s/bans?limit=10&offset=0", userID.String()), nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetAppealHistoryFlow tests the admin get appeal history flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetAppealHistoryFlow() {
	// First create an appeal (this would normally be done by user)
	userID := uuid.New()

	// Prepare admin get appeal history request
	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/users/%s/appeals?limit=10&offset=0", userID.String()), nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminReviewAppealFlow tests the admin review appeal flow
func (suite *ModerationIntegrationTestSuite) TestAdminReviewAppealFlow() {
	// First create an appeal
	appealID := uuid.New()
	adminID := uuid.New()

	// Prepare admin review appeal request
	req := struct {
		Approved bool   `json:"approved"`
		Notes    string `json:"notes"`
	}{
		Approved: true,
		Notes:    "Appeal approved after review",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/appeals/%s/review", appealID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetModerationQueueFlow tests the admin get moderation queue flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetModerationQueueFlow() {
	// Prepare admin get moderation queue request
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/moderation-queue?priority=high&limit=10&offset=0", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestAdminGetModerationAnalyticsFlow tests the admin get moderation analytics flow
func (suite *ModerationIntegrationTestSuite) TestAdminGetModerationAnalyticsFlow() {
	// Prepare admin get moderation analytics request
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/analytics?period=7d", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-admin-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.NotNil(suite.T(), response["data"])
}

// TestModerationInputValidation tests input validation for moderation endpoints
func (suite *ModerationIntegrationTestSuite) TestModerationInputValidation() {
	// Test invalid report request
	req := moderation.ReportContentRequest{
		ReporterID:     uuid.New(),
		ReportedUserID: uuid.New(),
		Reason:         "invalid_reason", // Invalid reason
		Description:     stringPtr("Test description"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should return validation error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["error"].(map[string]interface{})["message"].(string), "reason")
}

// TestModerationAuthentication tests authentication for moderation endpoints
func (suite *ModerationIntegrationTestSuite) TestModerationAuthentication() {
	// Test without authentication
	req := moderation.ReportContentRequest{
		ReporterID:     uuid.New(),
		ReportedUserID: uuid.New(),
		Reason:         "inappropriate_behavior",
		Description:     stringPtr("Test description"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should return unauthorized
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestModerationAuthorization tests authorization for admin endpoints
func (suite *ModerationIntegrationTestSuite) TestModerationAuthorization() {
	// Test admin endpoint with user token
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/reports", nil)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-user-token") // User token, not admin

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should return forbidden
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

// TestModerationRateLimiting tests rate limiting for moderation endpoints
func (suite *ModerationIntegrationTestSuite) TestModerationRateLimiting() {
	// Make multiple rapid report requests
	for i := 0; i < 10; i++ {
		req := moderation.ReportContentRequest{
			ReporterID:     uuid.New(),
			ReportedUserID: uuid.New(),
			Reason:         "inappropriate_behavior",
			Description:     stringPtr("Test description"),
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "test-agent")
		httpReq.Header.Set("Authorization", "Bearer mock-access-token")

		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)

		// First few requests should succeed
		if i < 5 {
			assert.Equal(suite.T(), http.StatusCreated, w.Code)
		}
	}

	// Make a request that should be rate limited
	req := moderation.ReportContentRequest{
		ReporterID:     uuid.New(),
		ReportedUserID: uuid.New(),
		Reason:         "inappropriate_behavior",
		Description:     stringPtr("Test description"),
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/moderation/report", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should be rate limited
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestModerationCaching tests caching functionality for moderation endpoints
func (suite *ModerationIntegrationTestSuite) TestModerationCaching() {
	// Create test user
	userID := uuid.New()

	// First request to get blocked users
	httpReq1 := httptest.NewRequest("GET", "/api/v1/moderation/me/blocked?limit=10&offset=0", nil)
	httpReq1.Header.Set("Content-Type", "application/json")
	httpReq1.Header.Set("User-Agent", "test-agent")
	httpReq1.Header.Set("Authorization", "Bearer mock-access-token")

	// Perform first request
	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, httpReq1)

	// Should succeed
	assert.Equal(suite.T(), http.StatusOK, w1.Code)

	// Second request should be faster (from cache)
	httpReq2 := httptest.NewRequest("GET", "/api/v1/moderation/me/blocked?limit=10&offset=0", nil)
	httpReq2.Header.Set("Content-Type", "application/json")
	httpReq2.Header.Set("User-Agent", "test-agent")
	httpReq2.Header.Set("Authorization", "Bearer mock-access-token")

	start := time.Now()
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, httpReq2)
	duration := time.Since(start)

	// Should succeed and be faster
	assert.Equal(suite.T(), http.StatusOK, w2.Code)
	assert.Less(suite.T(), duration, 100*time.Millisecond) // Should be faster due to cache
}

// TestAIModerationIntegration tests AI moderation integration
func (suite *ModerationIntegrationTestSuite) TestAIModerationIntegration() {
	// Test AI moderation service integration
	content := "This is a test message for moderation"
	contentType := "text"
	contentID := uuid.New()

	// Analyze content
	result, err := suite.aiModerationService.AnalyzeText(content)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), contentType, result.ContentType)
	assert.Equal(suite.T(), contentID, result.ContentID)
	assert.GreaterOrEqual(suite.T(), result.Confidence, 0.0)
	assert.LessOrEqual(suite.T(), result.Confidence, 1.0)
}

// TestContentAnalysisIntegration tests content analysis integration
func (suite *ModerationIntegrationTestSuite) TestContentAnalysisIntegration() {
	// Test content analysis service integration
	content := "This is a test message for content analysis"
	userID := uuid.New()

	// Analyze content
	result, err := suite.contentAnalysisService.AnalyzeText(userID, content)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "text", result.ContentType)
	assert.GreaterOrEqual(suite.T(), result.Score, 0.0)
	assert.LessOrEqual(suite.T(), result.Score, 1.0)
}

// TestModerationServiceIntegration tests moderation service integration
func (suite *ModerationIntegrationTestSuite) TestModerationServiceIntegration() {
	// Test moderation service integration
	reportID := uuid.New()
	userID := uuid.New()

	// Process report
	result, err := suite.moderationService.ProcessReport(reportID, userID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), reportID, result.ReportID)
	assert.Equal(suite.T(), userID, result.UserID)
}

// TestModerationCacheIntegration tests moderation cache integration
func (suite *ModerationIntegrationTestSuite) TestModerationCacheIntegration() {
	// Test moderation cache service integration
	userID := uuid.New()
	blockedUsers := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Cache blocked users
	err := suite.moderationCacheService.CacheBlockedUsers(nil, userID, blockedUsers)
	assert.NoError(suite.T(), err)

	// Retrieve blocked users from cache
	cachedUsers, err := suite.moderationCacheService.GetBlockedUsers(nil, userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), blockedUsers, cachedUsers)

	// Invalidate cache
	err = suite.moderationCacheService.InvalidateBlockedUsers(nil, userID)
	assert.NoError(suite.T(), err)

	// Retrieve again should return nil (cache miss)
	cachedUsers, err = suite.moderationCacheService.GetBlockedUsers(nil, userID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), cachedUsers)
}

// TestModerationIntegration runs all moderation integration tests
func TestModerationIntegration(t *testing.T) {
	suite.Run(t, new(ModerationIntegrationTestSuite))
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}