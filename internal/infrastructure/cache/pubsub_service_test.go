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

// MockRedisClientForPubSub is a mock for the Redis client used in Pub/Sub tests
type MockRedisClientForPubSub struct {
	mock.Mock
}

func (m *MockRedisClientForPubSub) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForPubSub) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForPubSub) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForPubSub) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClientForPubSub) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForPubSub) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringStringMapCmd)
}

func (m *MockRedisClientForPubSub) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClientForPubSub) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := m.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := m.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (m *MockRedisClientForPubSub) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForPubSub) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForPubSub) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	args := m.Called(ctx, patterns)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForPubSub) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, script, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForPubSub) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, sha, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForPubSub) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := m.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForPubSub) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClientForPubSub) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// MockPubSub is a mock for Redis PubSub
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) Channel() <-chan *redis.Message {
	args := m.Called()
	return args.Get(0).(<-chan *redis.Message)
}

func (m *MockPubSub) ChannelWithSubscriptions(size int) <-chan *redis.Message {
	args := m.Called(size)
	return args.Get(0).(<-chan *redis.Message)
}

func (m *MockPubSub) Receive(ctx context.Context) *redis.Message {
	args := m.Called(ctx)
	return args.Get(0).(*redis.Message)
}

func (m *MockPubSub) ReceiveMessage(ctx context.Context) *redis.Message {
	args := m.Called(ctx)
	return args.Get(0).(*redis.Message)
}

func (m *MockPubSub) ReceiveTimeout(ctx context.Context, timeout time.Duration) *redis.Message {
	args := m.Called(ctx, timeout)
	return args.Get(0).(*redis.Message)
}

func (m *MockPubSub) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPubSub) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPubSub) Unsubscribe(channels ...string) error {
	args := m.Called(channels)
	return args.Error(0)
}

func (m *MockPubSub) PUnsubscribe(patterns ...string) error {
	args := m.Called(patterns)
	return args.Error(0)
}

func (m *MockPubSub) Subscribe(channels ...string) error {
	args := m.Called(channels)
	return args.Error(0)
}

func (m *MockPubSub) PSubscribe(patterns ...string) error {
	args := m.Called(patterns)
	return args.Error(0)
}

// PubSubServiceTestSuite is the test suite for Pub/Sub service
type PubSubServiceTestSuite struct {
	suite.Suite
	pubSubService *PubSubService
	mockClient    *MockRedisClientForPubSub
	mockPubSub     *MockPubSub
}

func (suite *PubSubServiceTestSuite) SetupTest() {
	suite.mockClient = new(MockRedisClientForPubSub)
	suite.mockPubSub = new(MockPubSub)
	suite.pubSubService = &PubSubService{
		redisClient: suite.mockClient,
		config: &PubSubConfig{
			MaxSubscriptions: 100,
			MessageBufferSize: 1000,
			HealthCheckInterval: 30 * time.Second,
		},
		subscriptions: make(map[string]*redis.PubSub),
	}
}

func (suite *PubSubServiceTestSuite) TestPublishChatMessage() {
	ctx := context.Background()
	fromUserID := "user1"
	toUserID := "user2"
	message := "Hello, world!"
	
	expectedMessage := Message{
		Type:      MessageTypeChat,
		Channel:   "chat:user2",
		Data: map[string]interface{}{
			"from_user_id": fromUserID,
			"to_user_id":   toUserID,
			"message":      message,
		},
		Timestamp: time.Now(),
	}
	
	messageData, _ := json.Marshal(expectedMessage)
	
	// Mock successful publish
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Publish", ctx, "chat:user2", messageData).Return(intCmd)
	
	err := suite.pubSubService.PublishChatMessage(ctx, fromUserID, toUserID, message)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestPublishNotification() {
	ctx := context.Background()
	userID := "user123"
	notificationType := "match"
	title := "New Match!"
	message := "You have a new match"
	data := map[string]interface{}{
		"match_id": "match123",
		"user_id":  "user456",
	}
	
	expectedMessage := Message{
		Type:      MessageTypeNotification,
		Channel:   "notifications:user123",
		Data: map[string]interface{}{
			"type":    notificationType,
			"title":   title,
			"message": message,
			"data":    data,
		},
		Timestamp: time.Now(),
	}
	
	messageData, _ := json.Marshal(expectedMessage)
	
	// Mock successful publish
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Publish", ctx, "notifications:user123", messageData).Return(intCmd)
	
	err := suite.pubSubService.PublishNotification(ctx, userID, notificationType, title, data)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestPublishOnlineStatus() {
	ctx := context.Background()
	userID := "user123"
	online := true
	
	expectedMessage := Message{
		Type:      MessageTypeOnlineStatus,
		Channel:   "online_status",
		Data: map[string]interface{}{
			"user_id": userID,
			"online":  online,
		},
		Timestamp: time.Now(),
	}
	
	messageData, _ := json.Marshal(expectedMessage)
	
	// Mock successful publish
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Publish", ctx, "online_status", messageData).Return(intCmd)
	
	err := suite.pubSubService.PublishOnlineStatus(ctx, userID, online)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestPublishMatchNotification() {
	ctx := context.Background()
	userID1 := "user1"
	userID2 := "user2"
	matchID := "match123"
	
	expectedMessage := Message{
		Type:      MessageTypeMatch,
		Channel:   "matches",
		Data: map[string]interface{}{
			"user_id_1": userID1,
			"user_id_2": userID2,
			"match_id":  matchID,
		},
		Timestamp: time.Now(),
	}
	
	messageData, _ := json.Marshal(expectedMessage)
	
	// Mock successful publish
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Publish", ctx, "matches", messageData).Return(intCmd)
	
	err := suite.pubSubService.PublishMatchNotification(ctx, userID1, userID2, matchID)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestPublishTypingIndicator() {
	ctx := context.Background()
	fromUserID := "user1"
	toUserID := "user2"
	typing := true
	
	expectedMessage := Message{
		Type:      MessageTypeTyping,
		Channel:   "typing:user2",
		Data: map[string]interface{}{
			"from_user_id": fromUserID,
			"to_user_id":   toUserID,
			"typing":       typing,
		},
		Timestamp: time.Now(),
	}
	
	messageData, _ := json.Marshal(expectedMessage)
	
	// Mock successful publish
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Publish", ctx, "typing:user2", messageData).Return(intCmd)
	
	err := suite.pubSubService.PublishTypingIndicator(ctx, fromUserID, toUserID, typing)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestSubscribeToChatMessages() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock successful subscription
	suite.mockClient.On("Subscribe", ctx, "chat:user123").Return(suite.mockPubSub)
	
	msgChan := suite.pubSubService.SubscribeToChatMessages(ctx, userID)
	
	suite.NotNil(msgChan)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestSubscribeToNotifications() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock successful subscription
	suite.mockClient.On("Subscribe", ctx, "notifications:user123").Return(suite.mockPubSub)
	
	msgChan := suite.pubSubService.SubscribeToNotifications(ctx, userID)
	
	suite.NotNil(msgChan)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestSubscribeToOnlineStatus() {
	ctx := context.Background()
	
	// Mock successful subscription
	suite.mockClient.On("Subscribe", ctx, "online_status").Return(suite.mockPubSub)
	
	msgChan := suite.pubSubService.SubscribeToOnlineStatus(ctx)
	
	suite.NotNil(msgChan)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestSubscribeToMatches() {
	ctx := context.Background()
	
	// Mock successful subscription
	suite.mockClient.On("Subscribe", ctx, "matches").Return(suite.mockPubSub)
	
	msgChan := suite.pubSubService.SubscribeToMatches(ctx)
	
	suite.NotNil(msgChan)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestSubscribeToTypingIndicators() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock successful subscription
	suite.mockClient.On("Subscribe", ctx, "typing:user123").Return(suite.mockPubSub)
	
	msgChan := suite.pubSubService.SubscribeToTypingIndicators(ctx, userID)
	
	suite.NotNil(msgChan)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestUnsubscribeFromChatMessages() {
	ctx := context.Background()
	userID := "user123"
	
	// Add subscription to the service
	suite.pubSubService.subscriptions["chat:user123"] = suite.mockPubSub
	
	// Mock successful unsubscribe
	suite.mockPubSub.On("Unsubscribe", "chat:user123").Return(nil)
	
	err := suite.pubSubService.UnsubscribeFromChatMessages(ctx, userID)
	
	suite.NoError(err)
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestUnsubscribeFromNotifications() {
	ctx := context.Background()
	userID := "user123"
	
	// Add subscription to the service
	suite.pubSubService.subscriptions["notifications:user123"] = suite.mockPubSub
	
	// Mock successful unsubscribe
	suite.mockPubSub.On("Unsubscribe", "notifications:user123").Return(nil)
	
	err := suite.pubSubService.UnsubscribeFromNotifications(ctx, userID)
	
	suite.NoError(err)
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestUnsubscribeFromOnlineStatus() {
	ctx := context.Background()
	
	// Add subscription to the service
	suite.pubSubService.subscriptions["online_status"] = suite.mockPubSub
	
	// Mock successful unsubscribe
	suite.mockPubSub.On("Unsubscribe", "online_status").Return(nil)
	
	err := suite.pubSubService.UnsubscribeFromOnlineStatus(ctx)
	
	suite.NoError(err)
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestUnsubscribeFromMatches() {
	ctx := context.Background()
	
	// Add subscription to the service
	suite.pubSubService.subscriptions["matches"] = suite.mockPubSub
	
	// Mock successful unsubscribe
	suite.mockPubSub.On("Unsubscribe", "matches").Return(nil)
	
	err := suite.pubSubService.UnsubscribeFromMatches(ctx)
	
	suite.NoError(err)
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestUnsubscribeFromTypingIndicators() {
	ctx := context.Background()
	userID := "user123"
	
	// Add subscription to the service
	suite.pubSubService.subscriptions["typing:user123"] = suite.mockPubSub
	
	// Mock successful unsubscribe
	suite.mockPubSub.On("Unsubscribe", "typing:user123").Return(nil)
	
	err := suite.pubSubService.UnsubscribeFromTypingIndicators(ctx, userID)
	
	suite.NoError(err)
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *PubSubServiceTestSuite) TestGetActiveSubscriptions() {
	ctx := context.Background()
	
	// Add some subscriptions
	suite.pubSubService.subscriptions["chat:user1"] = suite.mockPubSub
	suite.pubSubService.subscriptions["notifications:user1"] = suite.mockPubSub
	suite.pubSubService.subscriptions["online_status"] = suite.mockPubSub
	
	subscriptions := suite.pubSubService.GetActiveSubscriptions(ctx)
	
	suite.Equal(3, len(subscriptions))
	suite.Contains(subscriptions, "chat:user1")
	suite.Contains(subscriptions, "notifications:user1")
	suite.Contains(subscriptions, "online_status")
}

func (suite *PubSubServiceTestSuite) TestClose() {
	ctx := context.Background()
	
	// Add some subscriptions
	suite.pubSubService.subscriptions["chat:user1"] = suite.mockPubSub
	suite.pubSubService.subscriptions["notifications:user1"] = suite.mockPubSub
	
	// Mock successful close
	suite.mockPubSub.On("Close").Return(nil).Twice()
	
	err := suite.pubSubService.Close(ctx)
	
	suite.NoError(err)
	suite.Equal(0, len(suite.pubSubService.subscriptions))
	
	suite.mockPubSub.AssertExpectations(suite.T())
}

func TestPubSubServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubServiceTestSuite))
}

// TestPubSubConfig tests the PubSubConfig struct
func TestPubSubConfig(t *testing.T) {
	config := &PubSubConfig{
		MaxSubscriptions:    100,
		MessageBufferSize:   1000,
		HealthCheckInterval: 30 * time.Second,
	}
	
	assert.Equal(t, 100, config.MaxSubscriptions)
	assert.Equal(t, 1000, config.MessageBufferSize)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
}

// TestMessage tests the Message struct
func TestMessage(t *testing.T) {
	data := map[string]interface{}{
		"user_id": "user123",
		"message": "Hello, world!",
	}
	
	message := &Message{
		Type:      MessageTypeChat,
		Channel:   "chat:user123",
		Data:      data,
		Timestamp: time.Now(),
	}
	
	assert.Equal(t, MessageTypeChat, message.Type)
	assert.Equal(t, "chat:user123", message.Channel)
	assert.Equal(t, data, message.Data)
	assert.False(message.Timestamp.IsZero())
}

// TestMessageType constants
func TestMessageType(t *testing.T) {
	assert.Equal(t, "chat", MessageTypeChat)
	assert.Equal(t, "notification", MessageTypeNotification)
	assert.Equal(t, "online_status", MessageTypeOnlineStatus)
	assert.Equal(t, "match", MessageTypeMatch)
	assert.Equal(t, "typing", MessageTypeTyping)
}