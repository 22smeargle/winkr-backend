package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// MockRedisClient implements a mock Redis client for testing
type MockRedisClient struct {
	data map[string]string
	sets map[string][]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
		sets: make(map[string][]string),
	}
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.data[key] = value.(string)
	return nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	val, exists := m.data[key]
	if !exists {
		return "", redis.ErrNil
	}
	return val, nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if m.sets[key] == nil {
		m.sets[key] = []string{}
	}
	for _, member := range members {
		m.sets[key] = append(m.sets[key], member.(string))
	}
	return nil
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	if m.sets[key] == nil {
		return nil
	}
	for _, member := range members {
		for i, existing := range m.sets[key] {
			if existing == member.(string) {
				m.sets[key] = append(m.sets[key][:i], m.sets[key][i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if m.sets[key] == nil {
		return []string{}, nil
	}
	return m.sets[key], nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	// Mock implementation - doesn't actually set expiry
	return nil
}

func TestRedisTokenBlacklist_IsBlacklisted(t *testing.T) {
	mockRedis := NewMockRedisClient()
	blacklist := NewRedisTokenBlacklist(mockRedis, "test_prefix")

	ctx := context.Background()
	jti := "test-jti-123"

	// Initially not blacklisted
	isBlacklisted, err := blacklist.IsBlacklisted(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, isBlacklisted)

	// Blacklist the token
	err = blacklist.BlacklistToken(ctx, jti, time.Now().Add(time.Hour))
	assert.NoError(t, err)

	// Now should be blacklisted
	isBlacklisted, err = blacklist.IsBlacklisted(ctx, jti)
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
}

func TestRedisTokenBlacklist_RemoveFromBlacklist(t *testing.T) {
	mockRedis := NewMockRedisClient()
	blacklist := NewRedisTokenBlacklist(mockRedis, "test_prefix")

	ctx := context.Background()
	jti := "test-jti-123"

	// Blacklist the token
	err := blacklist.BlacklistToken(ctx, jti, time.Now().Add(time.Hour))
	require.NoError(t, err)

	// Verify it's blacklisted
	isBlacklisted, err := blacklist.IsBlacklisted(ctx, jti)
	require.NoError(t, err)
	require.True(t, isBlacklisted)

	// Remove from blacklist
	err = blacklist.RemoveFromBlacklist(ctx, jti)
	assert.NoError(t, err)

	// Now should not be blacklisted
	isBlacklisted, err = blacklist.IsBlacklisted(ctx, jti)
	assert.NoError(t, err)
	assert.False(t, isBlacklisted)
}

func TestSessionManager_CreateSession(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	jwtUtils := utils.NewJWTUtilsWithoutBlacklist("test-secret", 15*time.Minute, 7*24*time.Hour)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)

	ctx := context.Background()
	userID := uuid.New().String()
	deviceID := "device-123"
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: "fp-123",
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create session
	session, err := sessionManager.CreateSession(ctx, userID, deviceID, deviceInfo, ipAddress, userAgent)

	assert.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, deviceID, session.DeviceID)
	assert.Equal(t, ipAddress, session.IPAddress)
	assert.Equal(t, userAgent, session.UserAgent)
	assert.True(t, session.IsActive)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestSessionManager_GetSession(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	jwtUtils := utils.NewJWTUtilsWithoutBlacklist("test-secret", 15*time.Minute, 7*24*time.Hour)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)

	ctx := context.Background()
	userID := uuid.New().String()
	deviceID := "device-123"
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: "fp-123",
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create session
	createdSession, err := sessionManager.CreateSession(ctx, userID, deviceID, deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	// Get session
	retrievedSession, err := sessionManager.GetSession(ctx, createdSession.ID)

	assert.NoError(t, err)
	assert.Equal(t, createdSession.ID, retrievedSession.ID)
	assert.Equal(t, createdSession.UserID, retrievedSession.UserID)
	assert.Equal(t, createdSession.DeviceID, retrievedSession.DeviceID)
}

func TestSessionManager_InvalidateSession(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	jwtUtils := utils.NewJWTUtilsWithoutBlacklist("test-secret", 15*time.Minute, 7*24*time.Hour)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)

	ctx := context.Background()
	userID := uuid.New().String()
	deviceID := "device-123"
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: "fp-123",
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create session
	session, err := sessionManager.CreateSession(ctx, userID, deviceID, deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	// Verify session exists
	_, err = sessionManager.GetSession(ctx, session.ID)
	require.NoError(t, err)

	// Invalidate session
	err = sessionManager.InvalidateSession(ctx, session.ID)
	assert.NoError(t, err)

	// Session should no longer exist
	_, err = sessionManager.GetSession(ctx, session.ID)
	assert.Error(t, err)
}

func TestSessionManager_InvalidateAllUserSessions(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	jwtUtils := utils.NewJWTUtilsWithoutBlacklist("test-secret", 15*time.Minute, 7*24*time.Hour)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)

	ctx := context.Background()
	userID := uuid.New().String()
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: "fp-123",
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create multiple sessions for the same user
	session1, err := sessionManager.CreateSession(ctx, userID, "device-1", deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	session2, err := sessionManager.CreateSession(ctx, userID, "device-2", deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	// Verify sessions exist
	_, err = sessionManager.GetSession(ctx, session1.ID)
	require.NoError(t, err)
	_, err = sessionManager.GetSession(ctx, session2.ID)
	require.NoError(t, err)

	// Invalidate all user sessions
	err = sessionManager.InvalidateAllUserSessions(ctx, userID)
	assert.NoError(t, err)

	// All sessions should no longer exist
	_, err = sessionManager.GetSession(ctx, session1.ID)
	assert.Error(t, err)
	_, err = sessionManager.GetSession(ctx, session2.ID)
	assert.Error(t, err)
}

func TestSessionManager_GetUserSessions(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	jwtUtils := utils.NewJWTUtilsWithoutBlacklist("test-secret", 15*time.Minute, 7*24*time.Hour)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)

	ctx := context.Background()
	userID := uuid.New().String()
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: "fp-123",
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create multiple sessions for the same user
	session1, err := sessionManager.CreateSession(ctx, userID, "device-1", deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	session2, err := sessionManager.CreateSession(ctx, userID, "device-2", deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	// Get user sessions
	sessions, err := sessionManager.GetUserSessions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Verify session IDs
	sessionIDs := make([]string, 0, 2)
	for _, session := range sessions {
		sessionIDs = append(sessionIDs, session.ID)
	}
	assert.Contains(t, sessionIDs, session1.ID)
	assert.Contains(t, sessionIDs, session2.ID)
}

func TestTokenManager_RotateRefreshToken(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	blacklist := NewRedisTokenBlacklist(redisClient, "test_blacklist")
	jwtUtils := utils.NewJWTUtils("test-secret", 15*time.Minute, 7*24*time.Hour, blacklist)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)
	tokenManager := NewTokenManager(jwtUtils, blacklist, sessionManager)

	ctx := context.Background()
	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	sessionID := "session-456"

	// Generate original refresh token
	originalRefreshToken, err := jwtUtils.GenerateRefreshTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	require.NoError(t, err)

	// Rotate refresh token
	newRefreshToken, err := tokenManager.RotateRefreshToken(ctx, originalRefreshToken, deviceID, sessionID)

	assert.NoError(t, err)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, originalRefreshToken, newRefreshToken)

	// Original token should be blacklisted
	isBlacklisted, err := blacklist.IsBlacklisted(ctx, "test-blacklist:"+jwtUtils.GetTokenID(originalRefreshToken))
	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
}

func TestTokenManager_ValidateTokenWithSession(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	blacklist := NewRedisTokenBlacklist(redisClient, "test_blacklist")
	jwtUtils := utils.NewJWTUtils("test-secret", 15*time.Minute, 7*24*time.Hour, blacklist)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)
	tokenManager := NewTokenManager(jwtUtils, blacklist, sessionManager)

	ctx := context.Background()
	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	deviceInfo := &utils.DeviceInfo{
		Fingerprint: deviceID,
		Platform:    "Windows",
		Device:      "desktop",
		Browser:     "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Create session
	session, err := sessionManager.CreateSession(ctx, userID, deviceID, deviceInfo, ipAddress, userAgent)
	require.NoError(t, err)

	// Generate token with session
	token, err := jwtUtils.GenerateAccessTokenWithDevice(userID, email, isAdmin, deviceID, session.ID)
	require.NoError(t, err)

	// Validate token with session
	claims, err := tokenManager.ValidateTokenWithSession(ctx, token)

	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, deviceID, claims.DeviceID)
	assert.Equal(t, session.ID, claims.SessionID)
}

func TestTokenManager_ValidateTokenWithInvalidSession(t *testing.T) {
	mockRedis := NewMockRedisClient()
	redisClient := &redis.RedisClient{Client: mockRedis}
	blacklist := NewRedisTokenBlacklist(redisClient, "test_blacklist")
	jwtUtils := utils.NewJWTUtils("test-secret", 15*time.Minute, 7*24*time.Hour, blacklist)
	
	sessionManager := NewSessionManager(redisClient, "test_sessions", jwtUtils)
	tokenManager := NewTokenManager(jwtUtils, blacklist, sessionManager)

	ctx := context.Background()
	userID := uuid.New().String()
	email := "test@example.com"
	isAdmin := false
	deviceID := "device-123"
	sessionID := "invalid-session-456"

	// Generate token with invalid session
	token, err := jwtUtils.GenerateAccessTokenWithDevice(userID, email, isAdmin, deviceID, sessionID)
	require.NoError(t, err)

	// Validate token with invalid session
	_, err = tokenManager.ValidateTokenWithSession(ctx, token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get session")
}