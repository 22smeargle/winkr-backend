package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClientForSession is a mock for the Redis client used in session tests
type MockRedisClientForSession struct {
	mock.Mock
}

func (m *MockRedisClientForSession) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForSession) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForSession) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForSession) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClientForSession) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForSession) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringStringMapCmd)
}

func (m *MockRedisClientForSession) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClientForSession) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := m.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := m.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (m *MockRedisClientForSession) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForSession) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForSession) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	args := m.Called(ctx, patterns)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForSession) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, script, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForSession) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, sha, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForSession) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := m.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClientForSession) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// SessionManagerTestSuite is the test suite for session manager
type SessionManagerTestSuite struct {
	suite.Suite
	sessionManager *SessionManager
	mockClient     *MockRedisClientForSession
}

func (suite *SessionManagerTestSuite) SetupTest() {
	suite.mockClient = new(MockRedisClientForSession)
	suite.sessionManager = &SessionManager{
		redisClient: suite.mockClient,
		config: &SessionConfig{
			SessionTTL:      24 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			CleanupInterval: 1 * time.Hour,
			MaxSessionsPerUser: 5,
		},
	}
}

func (suite *SessionManagerTestSuite) TestCreateSession() {
	ctx := context.Background()
	userID := "user123"
	accessToken := "access_token"
	refreshToken := "refresh_token"
	ip := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	
	deviceInfo := DeviceInfo{
		DeviceID:   "device123",
		DeviceType: "mobile",
		OS:         "iOS",
		Browser:    "Safari",
		Location:   "New York",
	}
	
	// Mock successful session creation
	sessionData, _ := json.Marshal(Session{
		ID:           "session123",
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IPAddress:    ip,
		UserAgent:    userAgent,
		DeviceInfo:   deviceInfo,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	})
	
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), sessionData, 24*time.Hour).Return(statusCmd)
	
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("ZAdd", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*redis.Z")).Return(intCmd)
	
	intCmd2 := redis.NewIntCmd(ctx)
	intCmd2.SetVal(1)
	suite.mockClient.On("ZAdd", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("*redis.Z")).Return(intCmd2)
	
	session, err := suite.sessionManager.CreateSession(ctx, userID, accessToken, refreshToken, ip, userAgent, deviceInfo)
	
	suite.NoError(err)
	suite.Equal(userID, session.UserID)
	suite.Equal(accessToken, session.AccessToken)
	suite.Equal(refreshToken, session.RefreshToken)
	suite.Equal(ip, session.IPAddress)
	suite.Equal(userAgent, session.UserAgent)
	suite.Equal(deviceInfo.DeviceID, session.DeviceInfo.DeviceID)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestGetSession() {
	ctx := context.Background()
	sessionID := "session123"
	
	expectedSession := Session{
		ID:           sessionID,
		UserID:       "user123",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		DeviceInfo: DeviceInfo{
			DeviceID:   "device123",
			DeviceType: "mobile",
			OS:         "iOS",
			Browser:    "Safari",
			Location:   "New York",
		},
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	
	sessionData, _ := json.Marshal(expectedSession)
	
	// Mock successful session retrieval
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(sessionData))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	session, err := suite.sessionManager.GetSession(ctx, sessionID)
	
	suite.NoError(err)
	suite.Equal(expectedSession.ID, session.ID)
	suite.Equal(expectedSession.UserID, session.UserID)
	suite.Equal(expectedSession.AccessToken, session.AccessToken)
	suite.Equal(expectedSession.RefreshToken, session.RefreshToken)
	suite.Equal(expectedSession.IPAddress, session.IPAddress)
	suite.Equal(expectedSession.UserAgent, session.UserAgent)
	suite.Equal(expectedSession.DeviceInfo.DeviceID, session.DeviceInfo.DeviceID)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestGetSessionNotFound() {
	ctx := context.Background()
	sessionID := "nonexistent_session"
	
	// Mock session not found
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetErr(redis.Nil)
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	session, err := suite.sessionManager.GetSession(ctx, sessionID)
	
	suite.Error(err)
	suite.Nil(session)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestUpdateSessionActivity() {
	ctx := context.Background()
	sessionID := "session123"
	
	// Mock successful session update
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(statusCmd)
	
	err := suite.sessionManager.UpdateSessionActivity(ctx, sessionID)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestDeleteSession() {
	ctx := context.Background()
	sessionID := "session123"
	userID := "user123"
	
	// Mock successful session deletion
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string")).Return(intCmd)
	
	intCmd2 := redis.NewIntCmd(ctx)
	intCmd2.SetVal(1)
	suite.mockClient.On("ZRem", ctx, mock.AnythingOfType("string"), sessionID).Return(intCmd2)
	
	intCmd3 := redis.NewIntCmd(ctx)
	intCmd3.SetVal(1)
	suite.mockClient.On("ZRem", ctx, mock.AnythingOfType("string"), sessionID).Return(intCmd3)
	
	err := suite.sessionManager.DeleteSession(ctx, sessionID)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestDeleteUserSessions() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock getting user sessions
	stringSliceCmd := redis.NewStringSliceCmd(ctx)
	stringSliceCmd.SetVal([]string{"session1", "session2", "session3"})
	suite.mockClient.On("ZRange", ctx, mock.AnythingOfType("string"), int64(0), int64(-1)).Return(stringSliceCmd)
	
	// Mock deleting sessions
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(3)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(intCmd)
	
	// Mock removing from user sessions set
	intCmd2 := redis.NewIntCmd(ctx)
	intCmd2.SetVal(3)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string")).Return(intCmd2)
	
	err := suite.sessionManager.DeleteUserSessions(ctx, userID)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestIsUserOnline() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock user is online
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("ZCard", ctx, mock.AnythingOfType("string")).Return(intCmd)
	
	online, err := suite.sessionManager.IsUserOnline(ctx, userID)
	
	suite.NoError(err)
	suite.True(online)
	
	// Mock user is offline
	intCmd2 := redis.NewIntCmd(ctx)
	intCmd2.SetVal(0)
	suite.mockClient.On("ZCard", ctx, mock.AnythingOfType("string")).Return(intCmd2)
	
	online, err = suite.sessionManager.IsUserOnline(ctx, userID)
	
	suite.NoError(err)
	suite.False(online)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestGetOnlineUsers() {
	ctx := context.Background()
	
	// Mock getting online users
	stringSliceCmd := redis.NewStringSliceCmd(ctx)
	stringSliceCmd.SetVal([]string{"user1", "user2", "user3"})
	suite.mockClient.On("ZRange", ctx, mock.AnythingOfType("string"), int64(0), int64(-1)).Return(stringSliceCmd)
	
	onlineUsers, err := suite.sessionManager.GetOnlineUsers(ctx)
	
	suite.NoError(err)
	suite.Equal(3, len(onlineUsers))
	suite.Contains(onlineUsers, "user1")
	suite.Contains(onlineUsers, "user2")
	suite.Contains(onlineUsers, "user3")
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestGetUserSessions() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock getting user sessions
	stringSliceCmd := redis.NewStringSliceCmd(ctx)
	stringSliceCmd.SetVal([]string{"session1", "session2", "session3"})
	suite.mockClient.On("ZRange", ctx, mock.AnythingOfType("string"), int64(0), int64(-1)).Return(stringSliceCmd)
	
	// Mock getting session data
	session1 := Session{ID: "session1", UserID: userID}
	session2 := Session{ID: "session2", UserID: userID}
	session3 := Session{ID: "session3", UserID: userID}
	
	session1Data, _ := json.Marshal(session1)
	session2Data, _ := json.Marshal(session2)
	session3Data, _ := json.Marshal(session3)
	
	stringCmd1 := redis.NewStringCmd(ctx)
	stringCmd1.SetVal(string(session1Data))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd1)
	
	stringCmd2 := redis.NewStringCmd(ctx)
	stringCmd2.SetVal(string(session2Data))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd2)
	
	stringCmd3 := redis.NewStringCmd(ctx)
	stringCmd3.SetVal(string(session3Data))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd3)
	
	sessions, err := suite.sessionManager.GetUserSessions(ctx, userID)
	
	suite.NoError(err)
	suite.Equal(3, len(sessions))
	suite.Equal("session1", sessions[0].ID)
	suite.Equal("session2", sessions[1].ID)
	suite.Equal("session3", sessions[2].ID)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *SessionManagerTestSuite) TestCleanupExpiredSessions() {
	ctx := context.Background()
	
	// Mock getting all session keys
	stringSliceCmd := redis.NewStringSliceCmd(ctx)
	stringSliceCmd.SetVal([]string{"session1", "session2", "session3"})
	suite.mockClient.On("ZRange", ctx, mock.AnythingOfType("string"), int64(0), int64(-1)).Return(stringSliceCmd)
	
	// Mock checking session expiration
	stringCmd1 := redis.NewStringCmd(ctx)
	stringCmd1.SetVal(`{"expires_at":"2020-01-01T00:00:00Z"}`) // Expired
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd1)
	
	stringCmd2 := redis.NewStringCmd(ctx)
	stringCmd2.SetVal(`{"expires_at":"2030-01-01T00:00:00Z"}`) // Not expired
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd2)
	
	stringCmd3 := redis.NewStringCmd(ctx)
	stringCmd3.SetVal(`{"expires_at":"2020-01-01T00:00:00Z"}`) // Expired
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd3)
	
	// Mock deleting expired sessions
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(2)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(intCmd)
	
	// Mock removing from user sessions set
	intCmd2 := redis.NewIntCmd(ctx)
	intCmd2.SetVal(2)
	suite.mockClient.On("ZRem", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(intCmd2)
	
	err := suite.sessionManager.CleanupExpiredSessions(ctx)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func TestSessionManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SessionManagerTestSuite))
}

// TestDeviceInfo tests the DeviceInfo struct
func TestDeviceInfo(t *testing.T) {
	deviceInfo := DeviceInfo{
		DeviceID:   "device123",
		DeviceType: "mobile",
		OS:         "iOS",
		Browser:    "Safari",
		Location:   "New York",
	}
	
	assert.Equal(t, "device123", deviceInfo.DeviceID)
	assert.Equal(t, "mobile", deviceInfo.DeviceType)
	assert.Equal(t, "iOS", deviceInfo.OS)
	assert.Equal(t, "Safari", deviceInfo.Browser)
	assert.Equal(t, "New York", deviceInfo.Location)
}

// TestSessionConfig tests the SessionConfig struct
func TestSessionConfig(t *testing.T) {
	config := &SessionConfig{
		SessionTTL:      24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
		MaxSessionsPerUser: 5,
	}
	
	assert.Equal(t, 24*time.Hour, config.SessionTTL)
	assert.Equal(t, 7*24*time.Hour, config.RefreshTokenTTL)
	assert.Equal(t, 1*time.Hour, config.CleanupInterval)
	assert.Equal(t, 5, config.MaxSessionsPerUser)
}