package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// RedisTokenBlacklist implements TokenBlacklist interface using Redis
type RedisTokenBlacklist struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewRedisTokenBlacklist creates a new Redis-based token blacklist
func NewRedisTokenBlacklist(redisClient *redis.RedisClient, prefix string) *RedisTokenBlacklist {
	if prefix == "" {
		prefix = "token_blacklist"
	}
	return &RedisTokenBlacklist{
		redisClient: redisClient,
		prefix:      prefix,
	}
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisTokenBlacklist) IsBlacklisted(ctx interface{}, jti string) (bool, error) {
	key := r.getKey(jti)
	exists, err := r.redisClient.Exists(ctx.(context.Context), key)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist status: %w", err)
	}
	return exists, nil
}

// BlacklistToken adds a token to the blacklist
func (r *RedisTokenBlacklist) BlacklistToken(ctx interface{}, jti string, expiry time.Time) error {
	key := r.getKey(jti)
	ttl := time.Until(expiry)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}
	
	err := r.redisClient.Set(ctx.(context.Context), key, "1", ttl)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// RemoveFromBlacklist removes a token from the blacklist
func (r *RedisTokenBlacklist) RemoveFromBlacklist(ctx interface{}, jti string) error {
	key := r.getKey(jti)
	err := r.redisClient.Del(ctx.(context.Context), key)
	if err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}
	return nil
}

// getKey returns the Redis key for a token JTI
func (r *RedisTokenBlacklist) getKey(jti string) string {
	return fmt.Sprintf("%s:%s", r.prefix, jti)
}

// SessionManager manages user sessions
type SessionManager struct {
	redisClient *redis.RedisClient
	prefix      string
	jwtUtils    *utils.JWTUtils
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient *redis.RedisClient, prefix string, jwtUtils *utils.JWTUtils) *SessionManager {
	if prefix == "" {
		prefix = "sessions"
	}
	return &SessionManager{
		redisClient: redisClient,
		prefix:      prefix,
		jwtUtils:    jwtUtils,
	}
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	DeviceInfo   *DeviceInfo `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsActive     bool      `json:"is_active"`
}

// DeviceInfo represents device information for sessions
type DeviceInfo struct {
	Fingerprint string `json:"fingerprint"`
	Platform    string `json:"platform"`
	Device      string `json:"device"`
	Browser     string `json:"browser"`
}

// CreateSession creates a new user session
func (sm *SessionManager) CreateSession(ctx context.Context, userID, deviceID string, deviceInfo *utils.DeviceInfo, ipAddress, userAgent string) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		DeviceID:     deviceID,
		DeviceInfo: &DeviceInfo{
			Fingerprint: deviceInfo.Fingerprint,
			Platform:    deviceInfo.Platform,
			Device:      deviceInfo.Device,
			Browser:     deviceInfo.Browser,
		},
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActivity: now,
		CreatedAt:    now,
		ExpiresAt:    now.Add(7 * 24 * time.Hour), // 7 days
		IsActive:     true,
	}

	// Store session in Redis
	key := sm.getSessionKey(sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	err = sm.redisClient.Set(ctx, key, string(sessionData), 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Add session to user's session list
	userSessionsKey := sm.getUserSessionsKey(userID)
	err = sm.redisClient.SAdd(ctx, userSessionsKey, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to add session to user list: %w", err)
	}

	// Set expiry on user sessions list
	sm.redisClient.Expire(ctx, userSessionsKey, 7*24*time.Hour)

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := sm.getSessionKey(sessionID)
	sessionData, err := sm.redisClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity time for a session
func (sm *SessionManager) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()
	
	key := sm.getSessionKey(sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	err = sm.redisClient.Set(ctx, key, string(sessionData), time.Until(session.ExpiresAt))
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	// Get session to get user ID
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Remove session
	key := sm.getSessionKey(sessionID)
	err = sm.redisClient.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to remove session: %w", err)
	}

	// Remove from user's session list
	userSessionsKey := sm.getUserSessionsKey(session.UserID)
	err = sm.redisClient.SRem(ctx, userSessionsKey, sessionID)
	if err != nil {
		return fmt.Errorf("failed to remove session from user list: %w", err)
	}

	return nil
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	userSessionsKey := sm.getUserSessionsKey(userID)
	sessionIDs, err := sm.redisClient.SMembers(ctx, userSessionsKey)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Remove each session
	for _, sessionID := range sessionIDs {
		key := sm.getSessionKey(sessionID)
		sm.redisClient.Del(ctx, key)
	}

	// Clear user sessions list
	err = sm.redisClient.Del(ctx, userSessionsKey)
	if err != nil {
		return fmt.Errorf("failed to clear user sessions list: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	userSessionsKey := sm.getUserSessionsKey(userID)
	sessionIDs, err := sm.redisClient.SMembers(ctx, userSessionsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []*Session
	for _, sessionID := range sessionIDs {
		session, err := sm.GetSession(ctx, sessionID)
		if err != nil {
			// Skip invalid sessions
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	// This would typically be run as a background job
	// For now, we'll implement a simple cleanup based on session keys
	// In production, you might want to use Redis keyspace notifications or a more sophisticated approach
	return nil
}

// getSessionKey returns the Redis key for a session
func (sm *SessionManager) getSessionKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", sm.prefix, sessionID)
}

// getUserSessionsKey returns the Redis key for a user's sessions list
func (sm *SessionManager) getUserSessionsKey(userID string) string {
	return fmt.Sprintf("%s:user:%s", sm.prefix, userID)
}

// TokenManager manages tokens with rotation and blacklist support
type TokenManager struct {
	jwtUtils       *utils.JWTUtils
	blacklist      utils.TokenBlacklist
	sessionManager *SessionManager
}

// NewTokenManager creates a new token manager
func NewTokenManager(jwtUtils *utils.JWTUtils, blacklist utils.TokenBlacklist, sessionManager *SessionManager) *TokenManager {
	return &TokenManager{
		jwtUtils:       jwtUtils,
		blacklist:      blacklist,
		sessionManager: sessionManager,
	}
}

// RotateRefreshToken rotates a refresh token and invalidates the old one
func (tm *TokenManager) RotateRefreshToken(ctx context.Context, oldRefreshToken string, deviceID, sessionID string) (string, error) {
	// Validate old refresh token
	claims, err := tm.jwtUtils.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Blacklist old refresh token
	err = tm.jwtUtils.BlacklistToken(ctx, oldRefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to blacklist old refresh token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := tm.jwtUtils.GenerateRefreshTokenWithDevice(claims.UserID, claims.Email, claims.IsAdmin, deviceID, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	return newRefreshToken, nil
}

// InvalidateUserTokens invalidates all tokens for a user
func (tm *TokenManager) InvalidateUserTokens(ctx context.Context, userID string) error {
	// Invalidate all sessions
	err := tm.sessionManager.InvalidateAllUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

// ValidateTokenWithSession validates a token and checks if the session is active
func (tm *TokenManager) ValidateTokenWithSession(ctx context.Context, tokenString string) (*utils.Claims, error) {
	claims, err := tm.jwtUtils.ValidateTokenWithContext(tokenString, ctx)
	if err != nil {
		return nil, err
	}

	// If token has session ID, check if session is active
	if claims.SessionID != "" {
		session, err := tm.sessionManager.GetSession(ctx, claims.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session: %w", err)
		}
		if !session.IsActive || time.Now().After(session.ExpiresAt) {
			return nil, fmt.Errorf("session is inactive or expired")
		}
	}

	return claims, nil
}