package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClient is a mock for the Redis client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringStringMapCmd)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := m.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := m.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	args := m.Called(ctx, patterns)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, script, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClient) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, sha, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClient) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := m.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// RedisConnectionTestSuite is the test suite for Redis connection
type RedisConnectionTestSuite struct {
	suite.Suite
	redisClient *RedisClient
	mockClient  *MockRedisClient
}

func (suite *RedisConnectionTestSuite) SetupTest() {
	suite.mockClient = new(MockRedisClient)
	suite.redisClient = &RedisClient{
		client: suite.mockClient,
		config: &RedisConfig{
			Address:     "localhost:6379",
			Password:    "",
			DB:          0,
			PoolSize:    10,
			MinIdleConns: 5,
			MaxRetries:  3,
			DialTimeout: 5 * time.Second,
			ReadTimeout: 3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolTimeout: 4 * time.Second,
			IdleTimeout: 5 * time.Minute,
			IdleCheckFrequency: 1 * time.Minute,
		},
		metrics: &RedisMetrics{},
	}
}

func (suite *RedisConnectionTestSuite) TestPing() {
	ctx := context.Background()
	
	// Test successful ping
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("PONG")
	suite.mockClient.On("Ping", ctx).Return(statusCmd)
	
	err := suite.redisClient.Ping()
	suite.NoError(err)
	
	// Test failed ping
	statusCmdFail := redis.NewStatusCmd(ctx)
	statusCmdFail.SetErr(redis.Nil)
	suite.mockClient.On("Ping", ctx).Return(statusCmdFail)
	
	err = suite.redisClient.Ping()
	suite.Error(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestGet() {
	ctx := context.Background()
	key := "test_key"
	expectedValue := "test_value"
	
	// Test successful get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(expectedValue)
	suite.mockClient.On("Get", ctx, key).Return(stringCmd)
	
	value, err := suite.redisClient.Get(ctx, key)
	suite.NoError(err)
	suite.Equal(expectedValue, value)
	
	// Test key not found
	stringCmdNotFound := redis.NewStringCmd(ctx)
	stringCmdNotFound.SetErr(redis.Nil)
	suite.mockClient.On("Get", ctx, key).Return(stringCmdNotFound)
	
	value, err = suite.redisClient.Get(ctx, key)
	suite.Error(err)
	suite.Equal("", value)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestSet() {
	ctx := context.Background()
	key := "test_key"
	value := "test_value"
	expiration := time.Hour
	
	// Test successful set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, key, value, expiration).Return(statusCmd)
	
	err := suite.redisClient.Set(ctx, key, value, expiration)
	suite.NoError(err)
	
	// Test failed set
	statusCmdFail := redis.NewStatusCmd(ctx)
	statusCmdFail.SetErr(redis.Nil)
	suite.mockClient.On("Set", ctx, key, value, expiration).Return(statusCmdFail)
	
	err = suite.redisClient.Set(ctx, key, value, expiration)
	suite.Error(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestDel() {
	ctx := context.Background()
	keys := []string{"key1", "key2"}
	
	// Test successful delete
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(2)
	suite.mockClient.On("Del", ctx, keys).Return(intCmd)
	
	count, err := suite.redisClient.Del(ctx, keys...)
	suite.NoError(err)
	suite.Equal(int64(2), count)
	
	// Test failed delete
	intCmdFail := redis.NewIntCmd(ctx)
	intCmdFail.SetErr(redis.Nil)
	suite.mockClient.On("Del", ctx, keys).Return(intCmdFail)
	
	count, err = suite.redisClient.Del(ctx, keys...)
	suite.Error(err)
	suite.Equal(int64(0), count)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestExists() {
	ctx := context.Background()
	keys := []string{"key1", "key2"}
	
	// Test key exists
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Exists", ctx, keys).Return(intCmd)
	
	exists, err := suite.redisClient.Exists(ctx, keys...)
	suite.NoError(err)
	suite.True(exists)
	
	// Test key doesn't exist
	intCmdNotExists := redis.NewIntCmd(ctx)
	intCmdNotExists.SetVal(0)
	suite.mockClient.On("Exists", ctx, keys).Return(intCmdNotExists)
	
	exists, err = suite.redisClient.Exists(ctx, keys...)
	suite.NoError(err)
	suite.False(exists)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestExpire() {
	ctx := context.Background()
	key := "test_key"
	expiration := time.Hour
	
	// Test successful expire
	boolCmd := redis.NewBoolCmd(ctx)
	boolCmd.SetVal(true)
	suite.mockClient.On("Expire", ctx, key, expiration).Return(boolCmd)
	
	success, err := suite.redisClient.Expire(ctx, key, expiration)
	suite.NoError(err)
	suite.True(success)
	
	// Test failed expire
	boolCmdFail := redis.NewBoolCmd(ctx)
	boolCmdFail.SetErr(redis.Nil)
	suite.mockClient.On("Expire", ctx, key, expiration).Return(boolCmdFail)
	
	success, err = suite.redisClient.Expire(ctx, key, expiration)
	suite.Error(err)
	suite.False(success)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestHGet() {
	ctx := context.Background()
	key := "test_hash"
	field := "test_field"
	expectedValue := "test_value"
	
	// Test successful HGet
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(expectedValue)
	suite.mockClient.On("HGet", ctx, key, field).Return(stringCmd)
	
	value, err := suite.redisClient.HGet(ctx, key, field)
	suite.NoError(err)
	suite.Equal(expectedValue, value)
	
	// Test field not found
	stringCmdNotFound := redis.NewStringCmd(ctx)
	stringCmdNotFound.SetErr(redis.Nil)
	suite.mockClient.On("HGet", ctx, key, field).Return(stringCmdNotFound)
	
	value, err = suite.redisClient.HGet(ctx, key, field)
	suite.Error(err)
	suite.Equal("", value)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestHSet() {
	ctx := context.Background()
	key := "test_hash"
	values := []interface{}{"field1", "value1", "field2", "value2"}
	
	// Test successful HSet
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(2)
	suite.mockClient.On("HSet", ctx, key, values).Return(intCmd)
	
	count, err := suite.redisClient.HSet(ctx, key, values...)
	suite.NoError(err)
	suite.Equal(int64(2), count)
	
	// Test failed HSet
	intCmdFail := redis.NewIntCmd(ctx)
	intCmdFail.SetErr(redis.Nil)
	suite.mockClient.On("HSet", ctx, key, values).Return(intCmdFail)
	
	count, err = suite.redisClient.HSet(ctx, key, values...)
	suite.Error(err)
	suite.Equal(int64(0), count)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestGetStats() {
	// Mock pool stats
	poolStats := &redis.PoolStats{
		Hits:     100,
		Misses:   10,
		TotalConns: 10,
		IdleConns: 5,
		StaleConns: 0,
	}
	
	suite.mockClient.On("PoolStats").Return(poolStats)
	
	stats := suite.redisClient.GetStats()
	
	suite.NotNil(stats)
	suite.Equal(float64(100), stats["pool_hits"])
	suite.Equal(float64(10), stats["pool_misses"])
	suite.Equal(float64(10), stats["total_conns"])
	suite.Equal(float64(5), stats["idle_conns"])
	suite.Equal(float64(0), stats["stale_conns"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *RedisConnectionTestSuite) TestGetMetrics() {
	// Set some metrics
	suite.redisClient.metrics.ConnectionsCreated = 5
	suite.redisClient.metrics.ConnectionsClosed = 3
	suite.redisClient.metrics.ConnectionErrors = 1
	suite.redisClient.metrics.CommandsExecuted = 100
	suite.redisClient.metrics.CommandErrors = 2
	
	metrics := suite.redisClient.GetMetrics()
	
	suite.NotNil(metrics)
	suite.Equal(uint64(5), metrics.ConnectionsCreated)
	suite.Equal(uint64(3), metrics.ConnectionsClosed)
	suite.Equal(uint64(1), metrics.ConnectionErrors)
	suite.Equal(uint64(100), metrics.CommandsExecuted)
	suite.Equal(uint64(2), metrics.CommandErrors)
}

func (suite *RedisConnectionTestSuite) TestClose() {
	// Test successful close
	suite.mockClient.On("Close").Return(nil)
	
	err := suite.redisClient.Close()
	suite.NoError(err)
	
	// Test failed close
	suite.mockClient.On("Close").Return(assert.AnError)
	
	err = suite.redisClient.Close()
	suite.Error(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func TestRedisConnectionTestSuite(t *testing.T) {
	suite.Run(t, new(RedisConnectionTestSuite))
}

// TestRedisConfig tests the Redis configuration
func TestRedisConfig(t *testing.T) {
	config := &RedisConfig{
		Address:     "localhost:6379",
		Password:    "password",
		DB:          1,
		PoolSize:    20,
		MinIdleConns: 10,
		MaxRetries:  5,
		DialTimeout: 10 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolTimeout: 8 * time.Second,
		IdleTimeout: 10 * time.Minute,
		IdleCheckFrequency: 2 * time.Minute,
	}
	
	assert.Equal(t, "localhost:6379", config.Address)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, 1, config.DB)
	assert.Equal(t, 20, config.PoolSize)
	assert.Equal(t, 10, config.MinIdleConns)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 10*time.Second, config.DialTimeout)
	assert.Equal(t, 5*time.Second, config.ReadTimeout)
	assert.Equal(t, 5*time.Second, config.WriteTimeout)
	assert.Equal(t, 8*time.Second, config.PoolTimeout)
	assert.Equal(t, 10*time.Minute, config.IdleTimeout)
	assert.Equal(t, 2*time.Minute, config.IdleCheckFrequency)
}

// TestRedisClusterConfig tests the Redis cluster configuration
func TestRedisClusterConfig(t *testing.T) {
	config := &RedisClusterConfig{
		Addresses: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
		Password:  "cluster_password",
		PoolSize:  30,
		MaxRetries: 6,
		DialTimeout: 15 * time.Second,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolTimeout: 12 * time.Second,
		IdleTimeout: 15 * time.Minute,
		IdleCheckFrequency: 3 * time.Minute,
	}
	
	assert.Equal(t, []string{"localhost:7000", "localhost:7001", "localhost:7002"}, config.Addresses)
	assert.Equal(t, "cluster_password", config.Password)
	assert.Equal(t, 30, config.PoolSize)
	assert.Equal(t, 6, config.MaxRetries)
	assert.Equal(t, 15*time.Second, config.DialTimeout)
	assert.Equal(t, 10*time.Second, config.ReadTimeout)
	assert.Equal(t, 10*time.Second, config.WriteTimeout)
	assert.Equal(t, 12*time.Second, config.PoolTimeout)
	assert.Equal(t, 15*time.Minute, config.IdleTimeout)
	assert.Equal(t, 3*time.Minute, config.IdleCheckFrequency)
}