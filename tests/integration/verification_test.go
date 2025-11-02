package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
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
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/verification"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// VerificationIntegrationTestSuite tests verification endpoints
type VerificationIntegrationTestSuite struct {
	suite.Suite
	router                    *gin.Engine
	verificationHandler        *handlers.VerificationHandler
	adminVerificationHandler    *handlers.AdminVerificationHandler
	verificationRepo          repositories.VerificationRepository
	userRepo                 repositories.UserRepository
	aiService                *external.MockAIService
	documentService          *services.DocumentService
	storageService           *storage.MockStorageService
	cacheService            *cache.MockCacheService
	rateLimiter             *cache.MockRateLimiter
	jwtUtils                *utils.JWTUtils
	testUserID              uuid.UUID
	testAdminID             uuid.UUID
	accessToken             string
	adminAccessToken         string
}

// SetupSuite sets up test suite
func (suite *VerificationIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	suite.cacheService = cache.NewMockCacheService()
	suite.rateLimiter = cache.NewMockRateLimiter()
	suite.storageService = storage.NewMockStorageService()
	suite.aiService = external.NewMockAIService()
	
	// Create mock repositories
	suite.verificationRepo = repositories.NewMockVerificationRepository()
	suite.userRepo = repositories.NewMockUserRepository()
	
	// Create document service
	documentConfig := &config.DocumentProcessingConfig{
		OCRProvider:           "mock",
		MinConfidence:         0.80,
		SupportedDocumentTypes: []string{"id_card", "passport", "driver_license"},
		ExtractionEnabled:     true,
		ValidationEnabled:     true,
	}
	suite.documentService = services.NewDocumentService(documentConfig)
	
	// Create verification workflow service
	verificationConfig := &config.VerificationConfig{
		AIService: config.AIServiceConfig{
			Provider:            "mock",
			SimilarityThreshold: 0.85,
			Enabled:            true,
		},
		Thresholds: config.VerificationThresholdsConfig{
			SelfieSimilarityThreshold:   0.85,
			DocumentConfidenceThreshold: 0.80,
			LivenessConfidenceThreshold: 0.90,
			NSFWThreshold:             0.70,
			ManualReviewThreshold:      0.75,
		},
		Limits: config.VerificationLimitsConfig{
			MaxAttemptsPerDay:      3,
			MaxAttemptsPerMonth:    10,
			CooldownPeriod:         24 * time.Hour,
			VerificationExpiry:     365 * 24 * time.Hour,
			DocumentExpiry:         730 * 24 * time.Hour,
			MaxFileSize:            10485760,
			MaxSelfieFileSize:      5242880,
			MaxDocumentFileSize:    10485760,
		},
		Security: config.VerificationSecurityConfig{
			EncryptedStorage:      true,
			FraudDetectionEnabled: true,
			IPTrackingEnabled:     true,
			DeviceTrackingEnabled: true,
			RequireRecentPhoto:    true,
			MaxPhotoAge:         24 * time.Hour,
		},
	}
	
	verificationWorkflowService := services.NewVerificationWorkflowService(
		suite.verificationRepo,
		suite.userRepo,
		suite.aiService,
		suite.documentService,
		suite.storageService,
		suite.cacheService,
		verificationConfig,
	)
	
	// Create JWT utils
	suite.jwtUtils = utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24*7)
	
	// Create use cases
	requestSelfieVerificationUseCase := verification.NewRequestSelfieVerificationUseCase(
		suite.verificationRepo, suite.userRepo, verificationWorkflowService, suite.rateLimiter)
	submitSelfieVerificationUseCase := verification.NewSubmitSelfieVerificationUseCase(
		suite.verificationRepo, suite.userRepo, verificationWorkflowService, suite.storageService, suite.rateLimiter)
	getVerificationStatusUseCase := verification.NewGetVerificationStatusUseCase(
		suite.verificationRepo, suite.userRepo)
	requestDocumentVerificationUseCase := verification.NewRequestDocumentVerificationUseCase(
		suite.verificationRepo, suite.userRepo, verificationWorkflowService, suite.rateLimiter)
	submitDocumentVerificationUseCase := verification.NewSubmitDocumentVerificationUseCase(
		suite.verificationRepo, suite.userRepo, verificationWorkflowService, suite.storageService, suite.rateLimiter)
	processVerificationResultUseCase := verification.NewProcessVerificationResultUseCase(
		suite.verificationRepo, suite.userRepo, verificationWorkflowService)
	getPendingVerificationsUseCase := verification.NewGetPendingVerificationsUseCase(
		suite.verificationRepo)
	
	// Create handlers
	suite.verificationHandler = handlers.NewVerificationHandler(
		requestSelfieVerificationUseCase,
		submitSelfieVerificationUseCase,
		getVerificationStatusUseCase,
		requestDocumentVerificationUseCase,
		submitDocumentVerificationUseCase,
		suite.jwtUtils,
	)
	
	suite.adminVerificationHandler = handlers.NewAdminVerificationHandler(
		processVerificationResultUseCase,
		getPendingVerificationsUseCase,
		suite.jwtUtils,
	)
	
	// Create router
	suite.router = gin.New()
	
	// Add verification routes
	verifyGroup := suite.router.Group("/api/v1/verify")
	{
		verifyGroup.POST("/selfie/request", suite.verificationHandler.RequestSelfieVerification)
		verifyGroup.POST("/selfie/submit", suite.verificationHandler.SubmitSelfieVerification)
		verifyGroup.GET("/selfie/status", suite.verificationHandler.GetSelfieVerificationStatus)
		verifyGroup.POST("/document/request", suite.verificationHandler.RequestDocumentVerification)
		verifyGroup.POST("/document/submit", suite.verificationHandler.SubmitDocumentVerification)
		verifyGroup.GET("/document/status", suite.verificationHandler.GetDocumentVerificationStatus)
	}
	
	// Add admin verification routes
	adminGroup := suite.router.Group("/api/v1/admin/verifications")
	{
		adminGroup.GET("", suite.adminVerificationHandler.GetPendingVerifications)
		adminGroup.GET("/:id", suite.adminVerificationHandler.GetVerificationDetails)
		adminGroup.POST("/:id/approve", suite.adminVerificationHandler.ApproveVerification)
		adminGroup.POST("/:id/reject", suite.adminVerificationHandler.RejectVerification)
	}
	
	// Setup test data
	suite.setupTestData()
}

// setupTestData creates test users and tokens
func (suite *VerificationIntegrationTestSuite) setupTestData() {
	// Create test user
	suite.testUserID = uuid.New()
	testUser := &entities.User{
		ID:             suite.testUserID,
		Email:          "test@example.com",
		FirstName:      "John",
		LastName:       "Doe",
		VerificationLevel: entities.VerificationLevelNone,
		IsActive:       true,
	}
	suite.userRepo.Create(suite.ctx(), testUser)
	
	// Create test admin
	suite.testAdminID = uuid.New()
	testAdmin := &entities.AdminUser{
		ID:        suite.testAdminID,
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsActive:  true,
	}
	suite.verificationRepo.CreateAdminUser(suite.ctx(), testAdmin)
	
	// Generate tokens
	suite.accessToken = suite.generateToken(suite.testUserID, "user")
	suite.adminAccessToken = suite.generateToken(suite.testAdminID, "admin")
}

// generateToken generates a JWT token for testing
func (suite *VerificationIntegrationTestSuite) generateToken(userID uuid.UUID, role string) string {
	claims := map[string]interface{}{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	
	token, _ := suite.jwtUtils.GenerateToken(claims)
	return token
}

// ctx returns a test context
func (suite *VerificationIntegrationTestSuite) ctx() context.Context {
	return context.Background()
}

// TestSelfieVerificationFlow tests complete selfie verification flow
func (suite *VerificationIntegrationTestSuite) TestSelfieVerificationFlow() {
	// Step 1: Request selfie verification
	req := dto.RequestSelfieVerificationDTO{}
	reqBody, _ := json.Marshal(req)
	
	httpReq := httptest.NewRequest("POST", "/api/v1/verify/selfie/request", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.SelfieVerificationRequestResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Data.UploadURL)
	assert.NotEmpty(suite.T(), response.Data.VerificationID)
	
	verificationID := response.Data.VerificationID
	uploadURL := response.Data.UploadURL
	
	// Step 2: Submit selfie verification
	// Create a mock file upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file field
	part, _ := writer.CreateFormFile("photo", "selfie.jpg")
	part.Write([]byte("mock-image-data"))
	
	// Add verification_id field
	writer.WriteField("verification_id", verificationID.String())
	writer.Close()
	
	httpReq = httptest.NewRequest("POST", "/api/v1/verify/selfie/submit", body)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var submitResponse dto.SelfieVerificationSubmitResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &submitResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), submitResponse.Success)
	assert.Equal(suite.T(), "pending", submitResponse.Data.Status)
	assert.NotZero(suite.T(), submitResponse.Data.AIScore)
	
	// Step 3: Check verification status
	httpReq = httptest.NewRequest("GET", "/api/v1/verify/selfie/status", nil)
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var statusResponse dto.SelfieVerificationStatusResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), statusResponse.Success)
	assert.Equal(suite.T(), "pending", statusResponse.Data.Status)
	assert.NotZero(suite.T(), statusResponse.Data.AIScore)
}

// TestDocumentVerificationFlow tests complete document verification flow
func (suite *VerificationIntegrationTestSuite) TestDocumentVerificationFlow() {
	// Step 1: Request document verification
	req := dto.RequestDocumentVerificationDTO{
		DocumentType: "passport",
	}
	reqBody, _ := json.Marshal(req)
	
	httpReq := httptest.NewRequest("POST", "/api/v1/verify/document/request", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.DocumentVerificationRequestResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Data.UploadURL)
	assert.NotEmpty(suite.T(), response.Data.VerificationID)
	
	verificationID := response.Data.VerificationID
	
	// Step 2: Submit document verification
	// Create a mock file upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add file field
	part, _ := writer.CreateFormFile("photo", "passport.jpg")
	part.Write([]byte("mock-passport-data"))
	
	// Add verification_id field
	writer.WriteField("verification_id", verificationID.String())
	writer.Close()
	
	httpReq = httptest.NewRequest("POST", "/api/v1/verify/document/submit", body)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	httpReq.Header.Set("User-Agent", "test-agent")
	httpReq.Header.Set("X-Forwarded-For", "192.168.1.1")
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var submitResponse dto.DocumentVerificationSubmitResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &submitResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), submitResponse.Success)
	assert.Equal(suite.T(), "pending", submitResponse.Data.Status)
	assert.NotZero(suite.T(), submitResponse.Data.AIScore)
	
	// Step 3: Check verification status
	httpReq = httptest.NewRequest("GET", "/api/v1/verify/document/status", nil)
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var statusResponse dto.DocumentVerificationStatusResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), statusResponse.Success)
	assert.Equal(suite.T(), "pending", statusResponse.Data.Status)
	assert.NotZero(suite.T(), statusResponse.Data.AIScore)
}

// TestAdminVerificationApproval tests admin approval workflow
func (suite *VerificationIntegrationTestSuite) TestAdminVerificationApproval() {
	// First create a pending verification
	suite.TestSelfieVerificationFlow()
	
	// Get pending verifications as admin
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/verifications", nil)
	httpReq.Header.Set("Authorization", "Bearer "+suite.adminAccessToken)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var pendingResponse dto.PendingVerificationsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &pendingResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), pendingResponse.Success)
	assert.Len(suite.T(), pendingResponse.Data.Verifications, 1)
	
	verificationID := pendingResponse.Data.Verifications[0].ID
	
	// Approve the verification
	approveReq := dto.ApproveVerificationDTO{
		Reason: "User verified successfully",
	}
	reqBody, _ := json.Marshal(approveReq)
	
	httpReq = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s/approve", verificationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.adminAccessToken)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var approveResponse dto.VerificationActionResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &approveResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), approveResponse.Success)
	assert.Equal(suite.T(), "approved", approveResponse.Data.Status)
}

// TestAdminVerificationRejection tests admin rejection workflow
func (suite *VerificationIntegrationTestSuite) TestAdminVerificationRejection() {
	// First create a pending verification
	suite.TestSelfieVerificationFlow()
	
	// Get pending verifications as admin
	httpReq := httptest.NewRequest("GET", "/api/v1/admin/verifications", nil)
	httpReq.Header.Set("Authorization", "Bearer "+suite.adminAccessToken)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var pendingResponse dto.PendingVerificationsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &pendingResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), pendingResponse.Success)
	assert.Len(suite.T(), pendingResponse.Data.Verifications, 1)
	
	verificationID := pendingResponse.Data.Verifications[0].ID
	
	// Reject the verification
	rejectReq := dto.RejectVerificationDTO{
		Reason: "Photo quality is too low",
	}
	reqBody, _ := json.Marshal(rejectReq)
	
	httpReq = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s/reject", verificationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.adminAccessToken)
	
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var rejectResponse dto.VerificationActionResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &rejectResponse)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), rejectResponse.Success)
	assert.Equal(suite.T(), "rejected", rejectResponse.Data.Status)
}

// TestVerificationRateLimiting tests rate limiting functionality
func (suite *VerificationIntegrationTestSuite) TestVerificationRateLimiting() {
	// Make multiple verification requests
	for i := 0; i < 5; i++ {
		req := dto.RequestSelfieVerificationDTO{}
		reqBody, _ := json.Marshal(req)
		
		httpReq := httptest.NewRequest("POST", "/api/v1/verify/selfie/request", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
		httpReq.Header.Set("User-Agent", "test-agent")
		httpReq.Header.Set("X-Forwarded-For", "192.168.1.1")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)
		
		// First few requests should succeed
		if i < 3 {
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		} else {
			// Should be rate limited after 3 attempts
			assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
		}
	}
}

// TestVerificationInputValidation tests input validation
func (suite *VerificationIntegrationTestSuite) TestVerificationInputValidation() {
	// Test invalid document type
	req := dto.RequestDocumentVerificationDTO{
		DocumentType: "invalid_type",
	}
	reqBody, _ := json.Marshal(req)
	
	httpReq := httptest.NewRequest("POST", "/api/v1/verify/document/request", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Should return validation error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response dto.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "document_type")
}

// TestVerificationSecurityHeaders tests security headers
func (suite *VerificationIntegrationTestSuite) TestVerificationSecurityHeaders() {
	// Make a request
	httpReq := httptest.NewRequest("GET", "/api/v1/verify/selfie/status", nil)
	httpReq.Header.Set("Authorization", "Bearer "+suite.accessToken)
	httpReq.Header.Set("User-Agent", "test-agent")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check security headers
	assert.Equal(suite.T(), "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(suite.T(), "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(suite.T(), "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.NotEmpty(suite.T(), w.Header().Get("Content-Security-Policy"))
}

// TestAIIntegration tests AI service integration
func (suite *VerificationIntegrationTestSuite) TestAIIntegration() {
	// Create a verification that will trigger AI processing
	suite.TestSelfieVerificationFlow()
	
	// Check that AI service was called
	assert.True(suite.T(), suite.aiService.CalledAnalyzeFace)
	assert.True(suite.T(), suite.aiService.CalledCompareFaces)
	assert.True(suite.T(), suite.aiService.CalledDetectNSFW)
}

// TestDocumentProcessing tests document processing
func (suite *VerificationIntegrationTestSuite) TestDocumentProcessing() {
	// Create a document verification
	suite.TestDocumentVerificationFlow()
	
	// Check that document service was called
	assert.True(suite.T(), suite.documentService.CalledExtractDocumentData)
	assert.True(suite.T(), suite.documentService.CalledValidateDocument)
}

// TestVerificationWorkflow tests complete verification workflow
func (suite *VerificationIntegrationTestSuite) TestVerificationWorkflow() {
	// Test selfie verification workflow
	suite.TestSelfieVerificationFlow()
	
	// Test document verification workflow
	suite.TestDocumentVerificationFlow()
	
	// Test admin approval workflow
	suite.TestAdminVerificationApproval()
	
	// Test admin rejection workflow
	suite.TestAdminVerificationRejection()
}

// TestVerificationIntegration runs all verification integration tests
func TestVerificationIntegration(t *testing.T) {
	suite.Run(t, new(VerificationIntegrationTestSuite))
}