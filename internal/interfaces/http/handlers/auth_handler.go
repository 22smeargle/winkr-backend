package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/auth"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/utils"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	registerUseCase           *auth.RegisterUseCase
	loginUseCase             *auth.LoginUseCase
	refreshUseCase           *auth.RefreshTokenUseCase
	logoutUseCase            *auth.LogoutUseCase
	passwordResetUseCase     *auth.PasswordResetUseCase
	confirmPasswordResetUseCase *auth.ConfirmPasswordResetUseCase
	emailVerificationUseCase *auth.EmailVerificationUseCase
	getProfileUseCase        *auth.GetProfileUseCase
	getSessionsUseCase       *auth.GetSessionsUseCase
	jwtUtils                *utils.JWTUtils
	authValidator            *validator.AuthValidator
	rateLimiter             *middleware.AuthRateLimiter
	suspiciousDetector        *middleware.SuspiciousActivityDetector
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	registerUseCase *auth.RegisterUseCase,
	loginUseCase *auth.LoginUseCase,
	refreshUseCase *auth.RefreshTokenUseCase,
	logoutUseCase *auth.LogoutUseCase,
	passwordResetUseCase *auth.PasswordResetUseCase,
	confirmPasswordResetUseCase *auth.ConfirmPasswordResetUseCase,
	emailVerificationUseCase *auth.EmailVerificationUseCase,
	getProfileUseCase *auth.GetProfileUseCase,
	getSessionsUseCase *auth.GetSessionsUseCase,
	jwtUtils *utils.JWTUtils,
	authValidator *validator.AuthValidator,
	rateLimiter *middleware.AuthRateLimiter,
	suspiciousDetector *middleware.SuspiciousActivityDetector,
) *AuthHandler {
	return &AuthHandler{
		registerUseCase:           registerUseCase,
		loginUseCase:             loginUseCase,
		refreshUseCase:           refreshUseCase,
		logoutUseCase:            logoutUseCase,
		passwordResetUseCase:     passwordResetUseCase,
		confirmPasswordResetUseCase: confirmPasswordResetUseCase,
		emailVerificationUseCase: emailVerificationUseCase,
		getProfileUseCase:        getProfileUseCase,
		getSessionsUseCase:       getSessionsUseCase,
		jwtUtils:                jwtUtils,
		authValidator:            authValidator,
		rateLimiter:             rateLimiter,
		suspiciousDetector:        suspiciousDetector,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email, password, and profile information
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequestDTO true "Registration request"
// @Success 201 {object} dto.AuthResponseDTO
// @Failure 400 {object} dto.AuthResponseDTO
// @Failure 409 {object} dto.AuthResponseDTO
// @Failure 500 {object} dto.AuthResponseDTO
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("register")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.RegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.RegistrationRequest{
		Email:        req.Email,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		InterestedIn: req.InterestedIn,
	}

	if err := h.authValidator.ValidateRegistrationRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Convert to use case request
	useCaseReq := &auth.RegisterRequest{
		Email:        req.Email,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		InterestedIn: req.InterestedIn,
	}

	// Execute use case
	response, err := h.registerUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	authResponse := &dto.AuthResponseDTO{
		Success: true,
		Data: &dto.AuthDataDTO{
			User: &dto.UserDTO{
				ID:        response.User.ID.String(),
				Email:     response.User.Email,
				FirstName:  response.User.FirstName,
				LastName:   response.User.LastName,
				IsVerified: response.User.IsVerified,
				IsPremium:  response.User.IsPremium,
				CreatedAt:  response.User.CreatedAt,
			},
			Tokens: &dto.TokensDTO{
				AccessToken:  response.Tokens.AccessToken,
				RefreshToken: response.Tokens.RefreshToken,
				ExpiresIn:    response.Tokens.ExpiresIn,
			},
		},
	}

	utils.Created(c, authResponse)
}

// Login handles user login
// @Summary Authenticate user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequestDTO true "Login request"
// @Success 200 {object} dto.AuthResponseDTO
// @Failure 400 {object} dto.AuthResponseDTO
// @Failure 401 {object} dto.AuthResponseDTO
// @Failure 500 {object} dto.AuthResponseDTO
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("login")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.LoginRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.authValidator.ValidateLoginRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Parse device info
	deviceInfo := h.jwtUtils.ParseDeviceInfo(userAgent, clientIP)

	// Convert to use case request
	useCaseReq := &auth.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: deviceInfo,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
	}

	// Execute use case
	response, err := h.loginUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	authResponse := &dto.AuthResponseDTO{
		Success: true,
		Data: &dto.AuthDataDTO{
			User: &dto.UserDTO{
				ID:        response.User.ID.String(),
				Email:     response.User.Email,
				FirstName:  response.User.FirstName,
				LastName:   response.User.LastName,
				IsVerified: response.User.IsVerified,
				IsPremium:  response.User.IsPremium,
				CreatedAt:  response.User.CreatedAt,
			},
			Tokens: &dto.TokensDTO{
				AccessToken:  response.Tokens.AccessToken,
				RefreshToken: response.Tokens.RefreshToken,
				ExpiresIn:    response.Tokens.ExpiresIn,
			},
		},
	}

	utils.Success(c, authResponse)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequestDTO true "Refresh token request"
// @Success 200 {object} dto.TokenResponseDTO
// @Failure 400 {object} dto.TokenResponseDTO
// @Failure 401 {object} dto.TokenResponseDTO
// @Failure 500 {object} dto.TokenResponseDTO
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("refresh")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.RefreshTokenRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	if err := h.authValidator.ValidateRefreshTokenRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
		IPAddress:    clientIP,
		UserAgent:    userAgent,
	}

	// Execute use case
	response, err := h.refreshUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	tokenResponse := &dto.TokenResponseDTO{
		Success: true,
		Data: &dto.TokenDataDTO{
			AccessToken:  response.AccessToken,
			RefreshToken: response.RefreshToken,
			ExpiresIn:    response.ExpiresIn,
		},
	}

	utils.Success(c, tokenResponse)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout user and invalidate tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.LogoutRequestDTO false "Logout request"
// @Success 200 {object} dto.LogoutResponseDTO
// @Failure 401 {object} dto.LogoutResponseDTO
// @Failure 500 {object} dto.LogoutResponseDTO
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("logout")(c)
	if c.IsAborted() {
		return
	}

	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.Unauthorized(c, "Authorization header is required")
		return
	}

	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		utils.Unauthorized(c, "Invalid authorization header format")
		return
	}

	// Validate token
	claims, err := h.jwtUtils.ValidateToken(token)
	if err != nil {
		utils.Unauthorized(c, "Invalid token")
		return
	}

	// Parse logout request (optional)
	var logoutReq dto.LogoutRequestDTO
	c.ShouldBindJSON(&logoutReq) // This will be empty if no body is provided

	// Enhanced validation
	validationReq := &validator.LogoutRequest{
		Token:     token,
		LogoutAll: logoutReq.LogoutAll,
	}

	if err := h.authValidator.ValidateLogoutRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Convert user ID string to UUID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID in token")
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.LogoutRequest{
		UserID:    userID,
		Token:     token,
		LogoutAll: logoutReq.LogoutAll,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.logoutUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	logoutResponse := &dto.LogoutResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.SuccessMessage(c, logoutResponse)
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} dto.UserProfileResponseDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 404 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("get-profile")(c)
	if c.IsAborted() {
		return
	}

	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.Unauthorized(c, "Authorization header is required")
		return
	}

	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		utils.Unauthorized(c, "Invalid authorization header format")
		return
	}

	// Validate token
	claims, err := h.jwtUtils.ValidateToken(token)
	if err != nil {
		utils.Unauthorized(c, "Invalid token")
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.GetProfileRequest{
		UserID:    claims.UserID,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.getProfileUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	profileResponse := &dto.UserProfileResponseDTO{
		Success: true,
		Data:    response,
	}

	utils.Success(c, profileResponse)
}

// PasswordReset handles password reset request
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequestDTO true "Password reset request"
// @Success 200 {object} dto.MessageResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/password-reset [post]
func (h *AuthHandler) PasswordReset(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("password-reset")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.ResetPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.PasswordResetRequest{
		Email: req.Email,
	}

	if err := h.authValidator.ValidatePasswordResetRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.PasswordResetRequest{
		Email:     req.Email,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
	}

	// Execute use case
	response, err := h.passwordResetUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	messageResponse := &dto.MessageResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.SuccessMessage(c, messageResponse)
}

// ConfirmPasswordReset handles password reset confirmation
// @Summary Confirm password reset
// @Description Reset password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ConfirmPasswordResetRequestDTO true "Password reset confirmation"
// @Success 200 {object} dto.MessageResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/password-reset/confirm [post]
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("password-reset-confirm")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.ConfirmPasswordResetRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.ConfirmPasswordResetRequest{
		Token:    req.Token,
		Password: req.Password,
	}

	if err := h.authValidator.ValidateConfirmPasswordResetRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.ConfirmPasswordResetRequest{
		Token:     req.Token,
		Password:  req.Password,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.confirmPasswordResetUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	messageResponse := &dto.MessageResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.SuccessMessage(c, messageResponse)
}

// SendEmailVerification handles sending email verification
// @Summary Send email verification
// @Description Send email verification to user
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} dto.MessageResponseDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/verify/send [post]
func (h *AuthHandler) SendEmailVerification(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("send-verification")(c)
	if c.IsAborted() {
		return
	}

	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.Unauthorized(c, "Authorization header is required")
		return
	}

	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		utils.Unauthorized(c, "Invalid authorization header format")
		return
	}

	// Validate token
	claims, err := h.jwtUtils.ValidateToken(token)
	if err != nil {
		utils.Unauthorized(c, "Invalid token")
		return
	}

	// Convert user ID string to UUID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID in token")
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.SendVerificationRequest{
		UserID:    userID,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.emailVerificationUseCase.ExecuteSendVerification(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	messageResponse := &dto.MessageResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.SuccessMessage(c, messageResponse)
}

// VerifyEmail handles email verification
// @Summary Verify email
// @Description Verify email using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequestDTO true "Email verification request"
// @Success 200 {object} dto.VerifyEmailResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/verify [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("verify-email")(c)
	if c.IsAborted() {
		return
	}

	// Check for suspicious activity
	if err := h.suspiciousDetector.CheckSuspiciousActivity(c); err != nil {
		utils.Error(c, err)
		return
	}

	var req dto.VerifyEmailRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Enhanced validation
	validationReq := &validator.VerifyEmailRequest{
		Token: req.Token,
	}

	if err := h.authValidator.ValidateVerifyEmailRequest(validationReq); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.VerifyEmailRequest{
		Token:     req.Token,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.emailVerificationUseCase.ExecuteVerifyEmail(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	verifyResponse := &dto.VerifyEmailResponseDTO{
		Success: true,
		Message: response.Message,
		Success: response.Success,
	}

	utils.Success(c, verifyResponse)
}

// GetSessions handles getting user sessions
// @Summary Get user sessions
// @Description Get all active sessions for the current user
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} dto.SessionsResponseDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("get-sessions")(c)
	if c.IsAborted() {
		return
	}

	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.Unauthorized(c, "Authorization header is required")
		return
	}

	token, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		utils.Unauthorized(c, "Invalid authorization header format")
		return
	}

	// Validate token
	claims, err := h.jwtUtils.ValidateToken(token)
	if err != nil {
		utils.Unauthorized(c, "Invalid token")
		return
	}

	// Convert user ID string to UUID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID in token")
		return
	}

	// Get client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert to use case request
	useCaseReq := &auth.GetSessionsRequest{
		UserID:    userID,
		IPAddress: clientIP,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.getSessionsUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	sessionsResponse := &dto.SessionsResponseDTO{
		Success: response.Success,
		Data:    response.Data,
		Message: response.Message,
	}

	utils.Success(c, sessionsResponse)
}