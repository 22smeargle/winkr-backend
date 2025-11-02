package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SessionManager handles user sessions in Redis
type SessionManager struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient *redis.RedisClient) *SessionManager {
	return &SessionManager{
		redisClient: redisClient,
		prefix:      "session:",
	}
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Token         string    `json:"token"`
	RefreshToken  string    `json:"refresh_token"`
	DeviceInfo   DeviceInfo `json:"device_info"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
}

// DeviceInfo contains information about the device
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	OS         string `json:"os"`
	Browser    string `json:"browser"`
	Location   string `json:"location"`
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(ctx context.Context, userID, token, refreshToken, ipAddress, userAgent string, deviceInfo DeviceInfo) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		Token:         token,
		RefreshToken:  refreshToken,
		DeviceInfo:   deviceInfo,
		LastActivity: now,
		CreatedAt:    now,
		ExpiresAt:    now.Add(24 * time.Hour), // 24 hours expiry
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	// Store session data as hash
	sessionKey := sm.getSessionKey(sessionID)
	sessionData, err := json.Marshal(session)
	if err != nil {
		logger.Error("Failed to marshal session data", err)
		return nil, fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Set session with expiration
	err = sm.redisClient.HSet(ctx, sessionKey, "data", string(sessionData))
	if err != nil {
		logger.Error("Failed to store session data", err)
		return nil, fmt.Errorf("failed to store session data: %w", err)
	}

	// Set expiration for the session hash
	err = sm.redisClient.Expire(ctx, sessionKey, 24*time.Hour)
	if err != nil {
		logger.Error("Failed to set session expiration", err)
		return nil, fmt.Errorf("failed to set session expiration: %w", err)
	}

	// Add session to user's sessions list
	userSessionsKey := sm.getUserSessionsKey(userID)
	err = sm.redisClient.SAdd(ctx, userSessionsKey, sessionID)
	if err != nil {
		logger.Error("Failed to add session to user sessions", err)
		return nil, fmt.Errorf("failed to add session to user sessions: %w", err)
	}

	// Set expiration for user sessions list
	err = sm.redisClient.Expire(ctx, userSessionsKey, 24*time.Hour)
	if err != nil {
		logger.Error("Failed to set user sessions expiration", err)
		return nil, fmt.Errorf("failed to set user sessions expiration: %w", err)
	}

	// Add user to online users set
	onlineUsersKey := sm.getOnlineUsersKey()
	err = sm.redisClient.SAdd(ctx, onlineUsersKey, userID)
	if err != nil {
		logger.Error("Failed to add user to online users", err)
		// Non-critical error, continue
	}

	logger.Info("Session created successfully", "session_id", sessionID, "user_id", userID)
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := sm.getSessionKey(sessionID)
	
	sessionData, err := sm.redisClient.HGet(ctx, sessionKey, "data")
	if err != nil {
		logger.Error("Failed to get session data", err)
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	if sessionData == "" {
		return nil, fmt.Errorf("session not found")
	}

	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		logger.Error("Failed to unmarshal session data", err)
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("session has expired")
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity time for a session
func (sm *SessionManager) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	sessionKey := sm.getSessionKey(sessionID)
	
	// Get current session data
	sessionData, err := sm.redisClient.HGet(ctx, sessionKey, "data")
	if err != nil {
		logger.Error("Failed to get session data for activity update", err)
		return fmt.Errorf("failed to get session data: %w", err)
	}

	if sessionData == "" {
		return fmt.Errorf("session not found")
	}

	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		logger.Error("Failed to unmarshal session data for activity update", err)
		return fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Update last activity and extend expiration
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(24 * time.Hour) // Extend by 24 hours

	updatedData, err := json.Marshal(session)
	if err != nil {
		logger.Error("Failed to marshal updated session data", err)
		return fmt.Errorf("failed to marshal updated session data: %w", err)
	}

	// Update session data
	err = sm.redisClient.HSet(ctx, sessionKey, "data", string(updatedData))
	if err != nil {
		logger.Error("Failed to update session data", err)
		return fmt.Errorf("failed to update session data: %w", err)
	}

	// Update expiration
	err = sm.redisClient.Expire(ctx, sessionKey, 24*time.Hour)
	if err != nil {
		logger.Error("Failed to update session expiration", err)
		return fmt.Errorf("failed to update session expiration: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err // Session might not exist or be expired
	}

	sessionKey := sm.getSessionKey(sessionID)
	userSessionsKey := sm.getUserSessionsKey(session.UserID)

	// Remove session data
	err = sm.redisClient.Del(ctx, sessionKey)
	if err != nil {
		logger.Error("Failed to delete session data", err)
		return fmt.Errorf("failed to delete session data: %w", err)
	}

	// Remove session from user's sessions list
	err = sm.redisClient.SRem(ctx, userSessionsKey, sessionID)
	if err != nil {
		logger.Error("Failed to remove session from user sessions", err)
		return fmt.Errorf("failed to remove session from user sessions: %w", err)
	}

	// Check if user has any other active sessions
	userSessions, err := sm.GetUserSessions(ctx, session.UserID)
	if err != nil {
		logger.Error("Failed to get user sessions", err)
		// Continue with cleanup
	} else if len(userSessions) == 0 {
		// Remove user from online users if no more sessions
		onlineUsersKey := sm.getOnlineUsersKey()
		err = sm.redisClient.SRem(ctx, onlineUsersKey, session.UserID)
		if err != nil {
			logger.Error("Failed to remove user from online users", err)
			// Non-critical error, continue
		}
	}

	logger.Info("Session deleted successfully", "session_id", sessionID, "user_id", session.UserID)
	return nil
}

// GetUserSessions retrieves all sessions for a user
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	userSessionsKey := sm.getUserSessionsKey(userID)
	
	sessionIDs, err := sm.redisClient.SMembers(ctx, userSessionsKey)
	if err != nil {
		logger.Error("Failed to get user session IDs", err)
		return nil, fmt.Errorf("failed to get user session IDs: %w", err)
	}

	var sessions []*Session
	for _, sessionID := range sessionIDs {
		session, err := sm.GetSession(ctx, sessionID)
		if err != nil {
			// Skip invalid/expired sessions
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteAllUserSessions removes all sessions for a user (logout from all devices)
func (sm *SessionManager) DeleteAllUserSessions(ctx context.Context, userID string) error {
	sessions, err := sm.GetUserSessions(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user sessions for deletion", err)
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, session := range sessions {
		err := sm.DeleteSession(ctx, session.ID)
		if err != nil {
			logger.Error("Failed to delete session during logout all", err, "session_id", session.ID)
			// Continue with other sessions
		}
	}

	// Remove user from online users
	onlineUsersKey := sm.getOnlineUsersKey()
	err = sm.redisClient.SRem(ctx, onlineUsersKey, userID)
	if err != nil {
		logger.Error("Failed to remove user from online users", err)
		// Non-critical error, continue
	}

	logger.Info("All user sessions deleted", "user_id", userID, "sessions_count", len(sessions))
	return nil
}

// IsUserOnline checks if a user is currently online
func (sm *SessionManager) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	onlineUsersKey := sm.getOnlineUsersKey()
	return sm.redisClient.SIsMember(ctx, onlineUsersKey, userID)
}

// GetOnlineUsers retrieves all online users
func (sm *SessionManager) GetOnlineUsers(ctx context.Context) ([]string, error) {
	onlineUsersKey := sm.getOnlineUsersKey()
	return sm.redisClient.SMembers(ctx, onlineUsersKey)
}

// CleanupExpiredSessions removes all expired sessions
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	// This would typically be run as a background job
	// For now, we'll implement a simple cleanup by checking user sessions
	// In a production environment, you might want to use Redis keyspace notifications
	// or a more efficient approach
	
	logger.Info("Starting expired sessions cleanup")
	
	// Get all online users
	onlineUsers, err := sm.GetOnlineUsers(ctx)
	if err != nil {
		logger.Error("Failed to get online users for cleanup", err)
		return fmt.Errorf("failed to get online users: %w", err)
	}

	cleanedCount := 0
	for _, userID := range onlineUsers {
		sessions, err := sm.GetUserSessions(ctx, userID)
		if err != nil {
			logger.Error("Failed to get user sessions during cleanup", err, "user_id", userID)
			continue
		}

		for _, session := range sessions {
			if time.Now().After(session.ExpiresAt) {
				err := sm.DeleteSession(ctx, session.ID)
				if err != nil {
					logger.Error("Failed to delete expired session", err, "session_id", session.ID)
				} else {
					cleanedCount++
				}
			}
		}
	}

	logger.Info("Expired sessions cleanup completed", "cleaned_count", cleanedCount)
	return nil
}

// Helper methods for key generation

func (sm *SessionManager) getSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", sm.prefix, sessionID)
}

func (sm *SessionManager) getUserSessionsKey(userID string) string {
	return fmt.Sprintf("%suser:%s", sm.prefix, userID)
}

func (sm *SessionManager) getOnlineUsersKey() string {
	return fmt.Sprintf("%sonline_users", sm.prefix)
}