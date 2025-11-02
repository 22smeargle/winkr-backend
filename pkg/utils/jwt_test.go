package utils

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTokenBlacklist implements TokenBlacklist interface for testing
type MockTokenBlacklist struct {
	blacklistedTokens map[string]bool
}

func NewMockTokenBlacklist() *MockTokenBlacklist {
	return &MockTokenBlacklist{
		blacklistedTokens: make(map[string]bool),
	}
}

func (m *MockTokenBlacklist) IsBlacklisted(ctx interface{}, jti string) (bool, error) {
	return m.blacklistedTokens[jti], nil
}

func (m *MockTokenBlacklist) BlacklistToken(ctx interface{}, jti string, expiry time.Time) error {
	m.blacklistedTokens[jti] = true
	return nil
}

func (m *MockTokenBlacklist) RemoveFromBlacklist(ctx interface{}, jti string) error {
	delete(m.blacklistedTokens, jti)
	return nil
}

func TestJWTUtils_GenerateAccessToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTUtils_GenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	token, err := jwtUtils.GenerateRefreshToken(userID, email, isAdmin)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTUtils_GenerateTokenPair(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	accessToken, refreshToken, err := jwtUtils.GenerateTokenPair(userID, email, isAdmin)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestJWTUtils_GenerateTokenPairWithDevice(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	sessionID := "session-456"

	accessToken, refreshToken, err := jwtUtils.GenerateTokenPairWithDevice(userID, email, isAdmin, deviceID, sessionID)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestJWTUtils_ValidateAccessToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Validate access token
	claims, err := jwtUtils.ValidateAccessToken(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, isAdmin, claims.IsAdmin)
	assert.Equal(t, "access", claims.TokenType)
}

func TestJWTUtils_ValidateRefreshToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate refresh token
	token, err := jwtUtils.GenerateRefreshToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Validate refresh token
	claims, err := jwtUtils.ValidateRefreshToken(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, isAdmin, claims.IsAdmin)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestJWTUtils_ValidateInvalidToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	// Test invalid token
	_, err := jwtUtils.ValidateToken("invalid.token.here")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestJWTUtils_ValidateExpiredToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 1 * time.Millisecond // Very short expiry
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	_, err = jwtUtils.ValidateToken(token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTUtils_BlacklistToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Get token ID
	jti, err := jwtUtils.GetTokenID(token)
	require.NoError(t, err)

	// Blacklist token
	err = jwtUtils.BlacklistToken(context.Background(), token)
	assert.NoError(t, err)

	// Check if token is blacklisted
	isBlacklisted, err := blacklist.IsBlacklisted(context.Background(), jti)
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
}

func TestJWTUtils_ValidateBlacklistedToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Get token ID and blacklist it
	jti, err := jwtUtils.GetTokenID(token)
	require.NoError(t, err)
	blacklist.BlacklistToken(context.Background(), jti, time.Now().Add(time.Hour))

	// Try to validate blacklisted token
	_, err = jwtUtils.ValidateTokenWithContext(token, context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is blacklisted")
}

func TestJWTUtils_GenerateDeviceFingerprint(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	deviceInfo := &DeviceInfo{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.1",
		Platform:  "Windows",
		Device:    "desktop",
		Browser:   "Chrome",
	}

	fingerprint := jwtUtils.GenerateDeviceFingerprint(deviceInfo)

	assert.NotEmpty(t, fingerprint)
	assert.Equal(t, 32, len(fingerprint)) // MD5 hash length
}

func TestJWTUtils_ParseDeviceInfo(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	ipAddress := "192.168.1.1"

	deviceInfo := jwtUtils.ParseDeviceInfo(userAgent, ipAddress)

	assert.Equal(t, userAgent, deviceInfo.UserAgent)
	assert.Equal(t, ipAddress, deviceInfo.IPAddress)
	assert.Equal(t, "Windows", deviceInfo.Platform)
	assert.Equal(t, "desktop", deviceInfo.Device)
	assert.Equal(t, "Chrome", deviceInfo.Browser)
	assert.NotEmpty(t, deviceInfo.Fingerprint)
}

func TestJWTUtils_ExtractTokenFromHeader(t *testing.T) {
	// Test valid header
	token, err := ExtractTokenFromHeader("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	assert.NoError(t, err)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token)

	// Test empty header
	_, err = ExtractTokenFromHeader("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization header is empty")

	// Test invalid header format
	_, err = ExtractTokenFromHeader("Invalid token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid authorization header format")
}

func TestJWTUtils_GetTokenExpiration(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Get token expiration
	expiration, err := jwtUtils.GetTokenExpiration(token)

	assert.NoError(t, err)
	assert.NotNil(t, expiration)
	assert.True(t, expiration.After(time.Now()))
}

func TestJWTUtils_IsTokenExpired(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 1 * time.Millisecond // Very short expiry
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Check if token is expired (should not be expired yet)
	isExpired := jwtUtils.IsTokenExpired(token)
	assert.False(t, isExpired)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Check if token is expired (should be expired now)
	isExpired = jwtUtils.IsTokenExpired(token)
	assert.True(t, isExpired)
}

func TestJWTUtils_GetTokenRemainingTime(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false

	// Generate access token
	token, err := jwtUtils.GenerateAccessToken(userID, email, isAdmin)
	require.NoError(t, err)

	// Get remaining time
	remaining, err := jwtUtils.GetTokenRemainingTime(token)

	assert.NoError(t, err)
	assert.True(t, remaining > 0)
	assert.True(t, remaining <= 15*time.Minute)
}

func TestJWTUtils_GetDeviceIDFromToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	sessionID := "session-456"

	// Generate access token with device info
	token, err := jwtUtils.GenerateAccessTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	require.NoError(t, err)

	// Get device ID from token
	extractedDeviceID, err := jwtUtils.GetDeviceIDFromToken(token)

	assert.NoError(t, err)
	assert.Equal(t, deviceID, extractedDeviceID)
}

func TestJWTUtils_GetSessionIDFromToken(t *testing.T) {
	secret := "test-secret"
	accessTokenExpiry := 15 * time.Minute
	refreshTokenExpiry := 7 * 24 * time.Hour
	blacklist := NewMockTokenBlacklist()

	jwtUtils := NewJWTUtils(secret, accessTokenExpiry, refreshTokenExpiry, blacklist)

	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	sessionID := "session-456"

	// Generate access token with device info
	token, err := jwtUtils.GenerateAccessTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	require.NoError(t, err)

	// Get session ID from token
	extractedSessionID, err := jwtUtils.GetSessionIDFromToken(token)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, extractedSessionID)
}