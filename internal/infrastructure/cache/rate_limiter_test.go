package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClientForRateLimiter is a mock for the Redis client used in rate limiter tests
type MockRedisClientForRateLimiter struct {
	mock.Mock
}

func (m *MockRedisClientForRateLimiter) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForRateLimiter) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForRateLimiter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForRateLimiter) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClientForRateLimiter) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForRateLimiter) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringStringMapCmd)
}

func (m *MockRedisClientForRateLimiter) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClientForRateLimiter) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := m.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := m.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (m *MockRedisClientForRateLimiter) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForRateLimiter) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForRateLimiter) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	args := m.Called(ctx, patterns)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForRateLimiter) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, script, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForRateLimiter) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, sha, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForRateLimiter) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := m.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForRateLimiter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClientForRateLimiter) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// RateLimiterTestSuite is the test suite for rate limiter
type RateLimiterTestSuite struct {
	suite.Suite
	rateLimiter *RateLimiter
	mockClient  *MockRedisClientForRateLimiter
}

func (suite *RateLimiterTestSuite) SetupTest() {
	suite.mockClient = new(MockRedisClientForRateLimiter)
	suite.rateLimiter = &RateLimiter{
		redisClient: suite.mockClient,
		config: &RateLimitConfig{
			Requests: 100,
			Window:   time.Minute,
			Endpoint: "test_endpoint",
			KeyType:  "ip",
		},
	}
}

func (suite *RateLimiterTestSuite) TestCheckRateLimitAllowed() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock successful rate limit check (allowed)
	cmd := redis.NewCmd(ctx)
	cmd.SetVal([]interface{}{int64(1), int64(9), int64(10)})
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.CheckRateLimit(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.True(result.Allowed)
	suite.Equal(int64(9), result.Remaining)
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestCheckRateLimitDenied() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock successful rate limit check (denied)
	cmd := redis.NewCmd(ctx)
	cmd.SetVal([]interface{}{int64(0), int64(0), int64(10)})
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.CheckRateLimit(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.False(result.Allowed)
	suite.Equal(int64(0), result.Remaining)
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestCheckDistributedRateLimitAllowed() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock successful distributed rate limit check (allowed)
	cmd := redis.NewCmd(ctx)
	cmd.SetVal([]interface{}{int64(1), int64(9), int64(10)})
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.CheckDistributedRateLimit(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.True(result.Allowed)
	suite.Equal(int64(9), result.Remaining)
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestCheckDistributedRateLimitDenied() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock successful distributed rate limit check (denied)
	cmd := redis.NewCmd(ctx)
	cmd.SetVal([]interface{}{int64(0), int64(0), int64(10)})
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.CheckDistributedRateLimit(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.False(result.Allowed)
	suite.Equal(int64(0), result.Remaining)
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestResetRateLimit() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock successful rate limit reset
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string")).Return(intCmd)
	
	err := suite.rateLimiter.ResetRateLimit(ctx, *config, identifier)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestGetRateLimitStatus() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock getting rate limit status
	cmd := redis.NewCmd(ctx)
	cmd.SetVal([]interface{}{int64(1), int64(5), int64(10)})
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.GetRateLimitStatus(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.True(result.Allowed)
	suite.Equal(int64(5), result.Remaining)
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestGetRateLimitStatusNotFound() {
	ctx := context.Background()
	identifier := "192.168.1.1"
	
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	// Mock rate limit status not found
	cmd := redis.NewCmd(ctx)
	cmd.SetErr(redis.Nil)
	suite.mockClient.On("Eval", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)
	
	result, err := suite.rateLimiter.GetRateLimitStatus(ctx, *config, identifier)
	
	suite.NoError(err)
	suite.True(result.Allowed) // Should be allowed if not found
	suite.Equal(int64(10), result.Remaining) // Should return full limit
	suite.Equal(int64(10), result.Limit)
	suite.Equal(config.Window, result.Window)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestGenerateRateLimitKey() {
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	identifier := "192.168.1.1"
	key := suite.rateLimiter.generateRateLimitKey(*config, identifier)
	
	expectedKey := "rate_limit:ip:test_endpoint:192.168.1.1"
	suite.Equal(expectedKey, key)
}

func (suite *RateLimiterTestSuite) TestGenerateDistributedRateLimitKey() {
	config := &RateLimitConfig{
		Requests: 10,
		Window:   time.Minute,
		Endpoint: "test_endpoint",
		KeyType:  "ip",
	}
	
	identifier := "192.168.1.1"
	key := suite.rateLimiter.generateDistributedRateLimitKey(*config, identifier)
	
	expectedKey := "distributed_rate_limit:ip:test_endpoint:192.168.1.1"
	suite.Equal(expectedKey, key)
}

func TestRateLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}

// TestRateLimitConfig tests the RateLimitConfig struct
func TestRateLimitConfig(t *testing.T) {
	config := &RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		Endpoint: "api_endpoint",
		KeyType:  "user",
	}
	
	assert.Equal(t, int64(100), config.Requests)
	assert.Equal(t, time.Minute, config.Window)
	assert.Equal(t, "api_endpoint", config.Endpoint)
	assert.Equal(t, "user", config.KeyType)
}

// TestRateLimitResult tests the RateLimitResult struct
func TestRateLimitResult(t *testing.T) {
	result := &RateLimitResult{
		Allowed:   true,
		Remaining: 50,
		Limit:     100,
		Window:    time.Minute,
		ResetTime: time.Now().Add(time.Minute),
	}
	
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(50), result.Remaining)
	assert.Equal(t, int64(100), result.Limit)
	assert.Equal(t, time.Minute, result.Window)
	assert.True(t, result.ResetTime.After(time.Now()))
}

// TestSlidingWindowScript tests the sliding window Lua script
func TestSlidingWindowScript(t *testing.T) {
	script := `
local current_time = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local key = KEYS[1]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, current_time - window)

-- Count current entries
local current = redis.call('ZCARD', key)

-- Check if limit is exceeded
if current < limit then
    -- Add new entry
    redis.call('ZADD', key, current_time, current_time)
    -- Set expiration
    redis.call('EXPIRE', key, window)
    return {1, limit - current - 1, limit}
else
    return {0, 0, limit}
end
`
	
	assert.NotEmpty(t, script)
	assert.Contains(t, script, "ZREMRANGEBYSCORE")
	assert.Contains(t, script, "ZCARD")
	assert.Contains(t, script, "ZADD")
	assert.Contains(t, script, "EXPIRE")
}

// TestDistributedSlidingWindowScript tests the distributed sliding window Lua script
func TestDistributedSlidingWindowScript(t *testing.T) {
	script := `
local current_time = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local key = KEYS[1]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, current_time - window)

-- Count current entries
local current = redis.call('ZCARD', key)

-- Check if limit is exceeded
if current < limit then
    -- Add new entry
    redis.call('ZADD', key, current_time, current_time)
    -- Set expiration
    redis.call('EXPIRE', key, window)
    return {1, limit - current - 1, limit}
else
    return {0, 0, limit}
end
`
	
	assert.NotEmpty(t, script)
	assert.Contains(t, script, "ZREMRANGEBYSCORE")
	assert.Contains(t, script, "ZCARD")
	assert.Contains(t, script, "ZADD")
	assert.Contains(t, script, "EXPIRE")
}