package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/auth"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// User authentication
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest, deviceInfo *utils.DeviceInfo, ipAddress string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, sessionID string) error
	LogoutFromAllDevices(ctx context.Context, userID uuid.UUID) error

	// Token validation
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)

	// Password management
	ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	ConfirmPasswordReset(ctx context.Context, req *ConfirmPasswordResetRequest) error

	// Session management
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*auth.Session, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error

	// Account security
	EnableTwoFactor(ctx context.Context, userID uuid.UUID) (*TwoFactorSetup, error)
	VerifyTwoFactor(ctx context.Context, userID uuid.UUID, code string) error
	DisableTwoFactor(ctx context.Context, userID uuid.UUID, password string) error

	// Account lockout
	IncrementFailedAttempts(ctx context.Context, email string) error
	ResetFailedAttempts(ctx context.Context, email string) error
	IsAccountLocked(ctx context.Context, email string) (bool, error)
}

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo       repositories.UserRepository
	jwtUtils       *utils.JWTUtils
	tokenManager   *auth.TokenManager
	sessionManager *auth.SessionManager
	passwordHash   func(string) (string, error)
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repositories.UserRepository,
	jwtUtils *utils.JWTUtils,
	tokenManager *auth.TokenManager,
	sessionManager *auth.SessionManager,
) AuthService {
	return &AuthServiceImpl{
		userRepo:       userRepo,
		jwtUtils:       jwtUtils,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
		passwordHash:   utils.HashPassword,
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required,password"`
	FirstName    string   `json:"first_name" validate:"required,min=2,max=100"`
	LastName     string   `json:"last_name" validate:"required,min=2,max=100"`
	DateOfBirth  string   `json:"date_of_birth" validate:"required"`
	Gender       string   `json:"gender" validate:"required,oneof=male female other"`
	InterestedIn []string `json:"interested_in" validate:"required,min=1,dive,oneof=male female other"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,password"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmPasswordResetRequest represents password reset confirmation
type ConfirmPasswordResetRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User   *UserInfo   `json:"user"`
	Tokens *TokenPair   `json:"tokens"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenResponse represents token refresh response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	TokenType string `json:"token_type"`
}

// UserInfo represents user information for auth response
type UserInfo struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	IsVerified    bool      `json:"is_verified"`
	IsPremium     bool      `json:"is_premium"`
	CreatedAt     string    `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TwoFactorSetup represents two-factor authentication setup
type TwoFactorSetup struct {
	Secret   string `json:"secret"`
	QRCode   string `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// Register implements user registration
func (s *AuthServiceImpl) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to check user existence")
	}
	if exists {
		return nil, errors.ErrEmailExists
	}

	// Check if account is locked
	locked, err := s.IsAccountLocked(ctx, req.Email)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to check account lock status")
	}
	if locked {
		return nil, errors.ErrAccountLocked
	}

	// Hash password
	hashedPassword, err := s.passwordHash(req.Password)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to hash password")
	}

	// Create user entity
	user := &entities.User{
		Email:        req.Email,
		PasswordHash:  hashedPassword,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Gender:        req.Gender,
		InterestedIn:   req.InterestedIn,
		IsActive:      true,
		IsVerified:     false,
		IsPremium:      false,
	}

	// Parse date of birth
	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return nil, errors.NewValidationError("date_of_birth", "Invalid date format. Use YYYY-MM-DD")
		}
		user.DateOfBirth = dob
	}

	// Create user
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to create user")
	}

	// Reset failed attempts on successful registration
	err = s.ResetFailedAttempts(ctx, req.Email)
	if err != nil {
		// Log error but don't fail registration
	}

	// Generate tokens (without device info for registration)
	accessToken, refreshToken, err := s.jwtUtils.GenerateTokenPair(user.ID.String(), user.Email, false)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to generate tokens")
	}

	// Return response
	return &AuthResponse{
		User: &UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			IsVerified: user.IsVerified,
			IsPremium:  user.IsPremium,
			CreatedAt:  user.CreatedAt.Format(time.RFC3339),
		},
		Tokens: &TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(15 * time.Minute.Seconds()), // 15 minutes
		},
	}, nil
}

// Login implements user login
func (s *AuthServiceImpl) Login(ctx context.Context, req *LoginRequest, deviceInfo *utils.DeviceInfo, ipAddress string) (*AuthResponse, error) {
	// Check if account is locked
	locked, err := s.IsAccountLocked(ctx, req.Email)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to check account lock status")
	}
	if locked {
		return nil, errors.ErrAccountLocked
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Increment failed attempts for non-existent users to prevent enumeration
		s.IncrementFailedAttempts(ctx, req.Email)
		return nil, errors.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.ErrAccountInactive
	}

	// Check if user is banned
	if user.IsBanned {
		return nil, errors.ErrAccountBanned
	}

	// Verify password
	err = utils.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		// Increment failed attempts
		s.IncrementFailedAttempts(ctx, req.Email)
		return nil, errors.ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	err = s.ResetFailedAttempts(ctx, req.Email)
	if err != nil {
		// Log error but don't fail login
	}

	// Update last active
	err = s.userRepo.UpdateLastActive(ctx, user.ID)
	if err != nil {
		// Log error but don't fail login
	}

	// Create session
	session, err := s.sessionManager.CreateSession(ctx, user.ID.String(), deviceInfo.Fingerprint, deviceInfo, ipAddress, "")
	if err != nil {
		return nil, errors.WrapError(err, "Failed to create session")
	}

	// Generate tokens with device and session info
	accessToken, refreshToken, err := s.jwtUtils.GenerateTokenPairWithDevice(
		user.ID.String(),
		user.Email,
		false,
		deviceInfo.Fingerprint,
		session.ID,
	)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to generate tokens")
	}

	// Return response
	return &AuthResponse{
		User: &UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			IsVerified: user.IsVerified,
			IsPremium:  user.IsPremium,
			CreatedAt:  user.CreatedAt.Format(time.RFC3339),
		},
		Tokens: &TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(15 * time.Minute.Seconds()), // 15 minutes
		},
	}, nil
}

// RefreshToken implements token refresh
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token with session checking
	claims, err := s.tokenManager.ValidateTokenWithSession(ctx, refreshToken)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, errors.ErrAccountInactive
	}

	// Rotate refresh token
	newRefreshToken, err := s.tokenManager.RotateRefreshToken(ctx, refreshToken, claims.DeviceID, claims.SessionID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to rotate refresh token")
	}

	// Generate new access token
	accessToken, err := s.jwtUtils.GenerateAccessTokenWithDevice(claims.UserID, claims.Email, claims.IsAdmin, claims.DeviceID, claims.SessionID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to generate access token")
	}

	// Update session activity
	if claims.SessionID != "" {
		s.sessionManager.UpdateSessionActivity(ctx, claims.SessionID)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(15 * time.Minute.Seconds()), // 15 minutes
	}, nil
}

// Logout implements user logout
func (s *AuthServiceImpl) Logout(ctx context.Context, userID uuid.UUID, sessionID string) error {
	// Invalidate specific session
	if sessionID != "" {
		err := s.sessionManager.InvalidateSession(ctx, sessionID)
		if err != nil {
			return errors.WrapError(err, "Failed to invalidate session")
		}
	}

	return nil
}

// LogoutFromAllDevices implements logout from all devices
func (s *AuthServiceImpl) LogoutFromAllDevices(ctx context.Context, userID uuid.UUID) error {
	// Invalidate all user sessions
	err := s.sessionManager.InvalidateAllUserSessions(ctx, userID.String())
	if err != nil {
		return errors.WrapError(err, "Failed to invalidate all sessions")
	}

	// Invalidate all user tokens
	err = s.tokenManager.InvalidateUserTokens(ctx, userID.String())
	if err != nil {
		return errors.WrapError(err, "Failed to invalidate user tokens")
	}

	return nil
}

// ValidateToken implements token validation
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtUtils.ValidateToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		IsAdmin:   claims.IsAdmin,
		TokenType: claims.TokenType,
	}, nil
}

// ValidateAccessToken implements access token validation
func (s *AuthServiceImpl) ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtUtils.ValidateAccessToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		IsAdmin:   claims.IsAdmin,
		TokenType: claims.TokenType,
	}, nil
}

// ValidateRefreshToken implements refresh token validation
func (s *AuthServiceImpl) ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error) {
	claims, err := s.jwtUtils.ValidateRefreshToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		IsAdmin:   claims.IsAdmin,
		TokenType: claims.TokenType,
	}, nil
}

// ChangePassword implements password change
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Verify current password
	err = utils.CheckPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return errors.ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := s.passwordHash(req.NewPassword)
	if err != nil {
		return errors.WrapError(err, "Failed to hash new password")
	}

	// Update password
	user.PasswordHash = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.WrapError(err, "Failed to update password")
	}

	return nil
}

// ResetPassword implements password reset
func (s *AuthServiceImpl) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()
	
	// Store reset token in Redis with 1-hour expiry
	resetKey := fmt.Sprintf("password_reset:%s", resetToken)
	resetData := map[string]interface{}{
		"email": req.Email,
		"type":  "password_reset",
	}
	
	// Use session manager's Redis client to store reset token
	err = s.sessionManager.redisClient.HMSet(ctx, resetKey, resetData)
	if err != nil {
		return errors.WrapError(err, "Failed to store reset token")
	}

	// Set expiry for reset token
	err = s.sessionManager.redisClient.Expire(ctx, resetKey, time.Hour)
	if err != nil {
		// Log error but continue as token is stored
	}

	// TODO: Send reset email with token
	// This would be handled by the verification service
	// emailService.SendPasswordResetEmail(user.Email, resetToken)

	return nil
}

// ConfirmPasswordReset implements password reset confirmation
func (s *AuthServiceImpl) ConfirmPasswordReset(ctx context.Context, req *ConfirmPasswordResetRequest) error {
	// Validate reset token and get user ID
	resetKey := fmt.Sprintf("password_reset:%s", req.Token)
	resetData, err := s.sessionManager.redisClient.HGetAll(ctx, resetKey)
	if err != nil {
		return errors.ErrInvalidToken
	}

	if len(resetData) == 0 {
		return errors.ErrInvalidToken
	}

	email, ok := resetData["email"]
	if !ok {
		return errors.ErrInvalidToken
	}

	// Get user and update password
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Hash new password and update
	hashedPassword, err := s.passwordHash(req.Password)
	if err != nil {
		return errors.WrapError(err, "Failed to hash password")
	}

	user.PasswordHash = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.WrapError(err, "Failed to update password")
	}

	// Invalidate reset token
	err = s.sessionManager.redisClient.Del(ctx, resetKey)
	if err != nil {
		// Log error but don't fail password reset
	}

	return nil
}

// GetActiveSessions implements getting active sessions
func (s *AuthServiceImpl) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*auth.Session, error) {
	return s.sessionManager.GetUserSessions(ctx, userID.String())
}

// InvalidateSession implements session invalidation
func (s *AuthServiceImpl) InvalidateSession(ctx context.Context, sessionID string) error {
	return s.sessionManager.InvalidateSession(ctx, sessionID)
}

// InvalidateAllSessions implements invalidating all user sessions
func (s *AuthServiceImpl) InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionManager.InvalidateAllUserSessions(ctx, userID.String())
}

// EnableTwoFactor implements enabling two-factor authentication
func (s *AuthServiceImpl) EnableTwoFactor(ctx context.Context, userID uuid.UUID) (*TwoFactorSetup, error) {
	// TODO: Implement two-factor authentication setup
	return &TwoFactorSetup{}, nil
}

// VerifyTwoFactor implements two-factor verification
func (s *AuthServiceImpl) VerifyTwoFactor(ctx context.Context, userID uuid.UUID, code string) error {
	// TODO: Implement two-factor verification
	return nil
}

// DisableTwoFactor implements disabling two-factor authentication
func (s *AuthServiceImpl) DisableTwoFactor(ctx context.Context, userID uuid.UUID, password string) error {
	// TODO: Implement two-factor disabling
	return nil
}

// IncrementFailedAttempts increments failed login attempts for an email
func (s *AuthServiceImpl) IncrementFailedAttempts(ctx context.Context, email string) error {
	// Use Redis to track failed attempts
	key := fmt.Sprintf("failed_attempts:%s", email)
	
	// Get current attempts
	attempts, err := s.sessionManager.redisClient.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, start from 0
		attempts = "0"
	}
	
	// Parse attempts
	var currentAttempts int
	if attempts != "" {
		_, err := fmt.Sscanf(attempts, "%d", &currentAttempts)
		if err != nil {
			currentAttempts = 0
		}
	}
	
	// Increment attempts
	currentAttempts++
	
	// Store updated attempts with 15-minute expiry
	err = s.sessionManager.redisClient.Set(ctx, key, fmt.Sprintf("%d", currentAttempts), 15*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to store failed attempts: %w", err)
	}
	
	// If attempts exceed threshold (5), lock account for 30 minutes
	if currentAttempts >= 5 {
		lockKey := fmt.Sprintf("account_locked:%s", email)
		err = s.sessionManager.redisClient.Set(ctx, lockKey, "1", 30*time.Minute)
		if err != nil {
			return fmt.Errorf("failed to lock account: %w", err)
		}
	}
	
	return nil
}

// ResetFailedAttempts resets failed login attempts for an email
func (s *AuthServiceImpl) ResetFailedAttempts(ctx context.Context, email string) error {
	// Remove failed attempts counter
	key := fmt.Sprintf("failed_attempts:%s", email)
	err := s.sessionManager.redisClient.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}
	
	// Remove account lock if exists
	lockKey := fmt.Sprintf("account_locked:%s", email)
	err = s.sessionManager.redisClient.Del(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("failed to remove account lock: %w", err)
	}
	
	return nil
}

// IsAccountLocked checks if an account is locked due to too many failed attempts
func (s *AuthServiceImpl) IsAccountLocked(ctx context.Context, email string) (bool, error) {
	// Check if account is locked
	lockKey := fmt.Sprintf("account_locked:%s", email)
	exists, err := s.sessionManager.redisClient.Exists(ctx, lockKey)
	if err != nil {
		return false, fmt.Errorf("failed to check account lock status: %w", err)
	}
	
	return exists, nil
}