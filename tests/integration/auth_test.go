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

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/auth"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/email"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/sms"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/validator"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AuthIntegrationTestSuite tests authentication endpoints
type AuthIntegrationTestSuite struct {
	suite.Suite
	router       *gin.Engine
	authHandler  *handlers.AuthHandler
	redisClient  *redis.RedisClient
	emailService *email.MockEmailService
	smsService   *sms.MockSMSService
}

// SetupSuite sets up the test suite
func (suite *AuthIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	suite.redisClient = redis.NewMockRedisClient()
	suite.emailService = email.NewMockEmailService()
	suite.smsService = sms.NewMockSMSService()

	// Create verification service
	verificationService := services.NewVerificationService(
		suite.redisClient,
		suite.emailService,
		suite.smsService,
	)

	// Create auth validator
	authValidator := validator.NewAuthValidator()

	// Create rate limiter
	rateLimiter := middleware.NewAuthRateLimiter(suite.redisClient)

	// Create suspicious activity detector
	suspiciousDetector := middleware.NewSuspiciousActivityDetector(suite.redisClient)

	// Create JWT utils
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24*7)

	// Create use cases
	registerUseCase := auth.NewRegisterUseCase(nil) // TODO: Add auth service
	loginUseCase := auth.NewLoginUseCase(nil, jwtUtils)
	refreshUseCase := auth.NewRefreshTokenUseCase(nil) // TODO: Add auth service
	logoutUseCase := auth.NewLogoutUseCase(nil, jwtUtils)
	passwordResetUseCase := auth.NewPasswordResetUseCase(nil, verificationService)
	confirmPasswordResetUseCase := auth.NewConfirmPasswordResetUseCase(nil, verificationService)
	emailVerificationUseCase := auth.NewEmailVerificationUseCase(nil, verificationService)
	getProfileUseCase := auth.NewGetProfileUseCase(nil) // TODO: Add auth service
	getSessionsUseCase := auth.NewGetSessionsUseCase(nil) // TODO: Add auth service

	// Create auth handler
	suite.authHandler = handlers.NewAuthHandler(
		registerUseCase,
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		passwordResetUseCase,
		confirmPasswordResetUseCase,
		emailVerificationUseCase,
		getProfileUseCase,
		getSessionsUseCase,
		jwtUtils,
		authValidator,
		rateLimiter,
		suspiciousDetector,
	)

	// Create router
	suite.router = gin.New()
	
	// Add auth routes
	authGroup := suite.router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", suite.authHandler.Register)
		authGroup.POST("/login", suite.authHandler.Login)
		authGroup.POST("/refresh", suite.authHandler.RefreshToken)
		authGroup.POST("/logout", suite.authHandler.Logout)
		authGroup.GET("/profile", suite.authHandler.GetProfile)
		authGroup.POST("/password-reset", suite.authHandler.PasswordReset)
		authGroup.POST("/password-reset/confirm", suite.authHandler.ConfirmPasswordReset)
		authGroup.POST("/verify/send", suite.authHandler.SendEmailVerification)
		authGroup.POST("/verify", suite.authHandler.VerifyEmail)
		authGroup.GET("/sessions", suite.authHandler.GetSessions)
	}
}

// TestRegistrationFlow tests the complete registration flow
func (suite *AuthIntegrationTestSuite) TestRegistrationFlow() {
	// Prepare registration request
	req := dto.RegisterRequestDTO{
		Email:        "test@example.com",
		Password:     "SecurePass123!",
		FirstName:    "John",
		LastName:     "Doe",
		DateOfBirth:  "1990-01-01",
		Gender:       "male",
		InterestedIn: []string{"female"},
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.AuthResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.User.Email, "test@example.com")
	assert.NotEmpty(suite.T(), response.Data.Tokens.AccessToken)
	assert.NotEmpty(suite.T(), response.Data.Tokens.RefreshToken)
}

// TestLoginFlow tests the complete login flow
func (suite *AuthIntegrationTestSuite) TestLoginFlow() {
	// First register a user
	suite.TestRegistrationFlow()

	// Prepare login request
	req := dto.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "SecurePass123!",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.AuthResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.User.Email, "test@example.com")
	assert.NotEmpty(suite.T(), response.Data.Tokens.AccessToken)
	assert.NotEmpty(suite.T(), response.Data.Tokens.RefreshToken)
}

// TestTokenRefreshFlow tests the token refresh flow
func (suite *AuthIntegrationTestSuite) TestTokenRefreshFlow() {
	// First login to get tokens
	suite.TestLoginFlow()

	// Get refresh token from previous login (this would normally be stored)
	refreshToken := "mock-refresh-token"

	// Prepare refresh request
	req := dto.RefreshTokenRequestDTO{
		RefreshToken: refreshToken,
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.TokenResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotEmpty(suite.T(), response.Data.AccessToken)
	assert.NotEmpty(suite.T(), response.Data.RefreshToken)
}

// TestLogoutFlow tests the logout flow
func (suite *AuthIntegrationTestSuite) TestLogoutFlow() {
	// First login to get tokens
	suite.TestLoginFlow()

	// Get access token from previous login (this would normally be stored)
	accessToken := "mock-access-token"

	// Prepare logout request
	req := dto.LogoutRequestDTO{
		LogoutAll: false,
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.LogoutResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestPasswordResetFlow tests the password reset flow
func (suite *AuthIntegrationTestSuite) TestPasswordResetFlow() {
	// Prepare password reset request
	req := dto.ResetPasswordRequestDTO{
		Email: "test@example.com",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/password-reset", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestEmailVerificationFlow tests the email verification flow
func (suite *AuthIntegrationTestSuite) TestEmailVerificationFlow() {
	// First register a user
	suite.TestRegistrationFlow()

	// Get access token from previous login (this would normally be stored)
	accessToken := "mock-access-token"

	// Prepare send verification request
	req := httptest.NewRequest("POST", "/api/v1/auth/verify/send", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)

	// Check that email was sent
	sentEmails := suite.emailService.FindEmailByRecipient("test@example.com")
	assert.Len(suite.T(), sentEmails, 1)
	assert.Contains(suite.T(), sentEmails[0].Body, "verification code")
}

// TestRateLimiting tests rate limiting functionality
func (suite *AuthIntegrationTestSuite) TestRateLimiting() {
	// Make multiple rapid requests
	for i := 0; i < 10; i++ {
		req := dto.LoginRequestDTO{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "test-agent")

		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)

		// First few requests should succeed (rate limit not exceeded)
		if i < 5 {
			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
		}
	}

	// Make a request that should be rate limited
	req := dto.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should be rate limited
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestInputValidation tests input validation
func (suite *AuthIntegrationTestSuite) TestInputValidation() {
	// Test invalid email
	req := dto.RegisterRequestDTO{
		Email:        "invalid-email",
		Password:     "SecurePass123!",
		FirstName:    "John",
		LastName:     "Doe",
		DateOfBirth:  "1990-01-01",
		Gender:       "male",
		InterestedIn: []string{"female"},
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should return validation error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response dto.AuthResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "email")
}

// TestSecurityHeaders tests security headers
func (suite *AuthIntegrationTestSuite) TestSecurityHeaders() {
	// Make a request
	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(suite.T(), "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(suite.T(), "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(suite.T(), "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.NotEmpty(suite.T(), w.Header().Get("Content-Security-Policy"))
}

// TestSuspiciousActivityDetection tests suspicious activity detection
func (suite *AuthIntegrationTestSuite) TestSuspiciousActivityDetection() {
	// Make requests from multiple IPs with same user agent
	userAgent := "test-agent"
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"}

	for _, ip := range ips {
		req := dto.LoginRequestDTO{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", userAgent)
		httpReq.Header.Set("X-Forwarded-For", ip)

		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)

		// First few requests should succeed
		if ip == "192.168.1.1" || ip == "192.168.1.2" || ip == "192.168.1.3" {
			assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
		}
	}

	// Fourth IP should trigger suspicious activity detection
	req := dto.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", userAgent)
	httpReq.Header.Set("X-Forwarded-For", "192.168.1.4")

	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)

	// Should be blocked due to suspicious activity
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestAuthIntegration runs all auth integration tests
func TestAuthIntegration(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}