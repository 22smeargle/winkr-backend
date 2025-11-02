package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	"github.com/22smeargle/winkr-backend/pkg/config"
)

// TestRedis holds test Redis connection and resources
type TestRedis struct {
	Client  *redis.Client
	Cleanup func()
}

// SetupMockRedis creates a mock Redis client for testing
func SetupMockRedis() *TestRedis {
	// This would use a mock Redis implementation
	// For now, we'll use a real Redis client pointed to a test database
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // Use a high DB number for tests
	})

	cleanup := func() {
		client.FlushDB(context.Background())
		client.Close()
	}

	return &TestRedis{
		Client:  client,
		Cleanup: cleanup,
	}
}

// SetupTestRedisWithDocker creates a test Redis instance using Docker
func SetupTestRedisWithDocker(t *testing.T) *TestRedis {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "Could not construct pool")

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	err = pool.Client.Ping()
	require.NoError(t, err, "Could not connect to Docker")

	// pull redis docker image
	resource, err := pool.RunWith(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7-alpine",
	})
	require.NoError(t, err, "Could not start resource")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	var client *redis.Client
	if err := pool.Retry(func() error {
		client = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
			Password: "",
			DB:       0,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return client.Ping(ctx).Err()
	}); err != nil {
		require.NoError(t, err, "Could not connect to Redis")
	}

	cleanup := func() {
		client.FlushDB(context.Background())
		client.Close()
		if err := pool.Purge(resource); err != nil {
			fmt.Printf("Could not purge resource: %s\n", err)
		}
	}

	return &TestRedis{
		Client:  client,
		Cleanup: cleanup,
	}
}

// SetupTestRedisFromConfig creates a test Redis from configuration
func SetupTestRedisFromConfig(t *testing.T, cfg config.RedisConfig) *TestRedis {
	client := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:           cfg.Password,
		DB:                 cfg.DB,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxRetries:         cfg.MaxRetries,
		DialTimeout:        cfg.DialTimeout,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Could not connect to Redis")

	cleanup := func() {
		client.FlushDB(context.Background())
		client.Close()
	}

	return &TestRedis{
		Client:  client,
		Cleanup: cleanup,
	}
}

// SetupTestRedis creates a test Redis instance
func SetupTestRedis(t *testing.T) *TestRedis {
	// Check if we're running in CI/CD
	if os.Getenv("CI") == "true" {
		return setupCITestRedis(t)
	}

	// Check if Docker is available
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return SetupTestRedisWithDocker(t)
	}

	// Fallback to local Redis
	return setupLocalTestRedis(t)
}

func setupCITestRedis(t *testing.T) *TestRedis {
	cfg := config.RedisConfig{
		Host:               getEnvOrDefault("TEST_REDIS_HOST", "localhost"),
		Port:               6379,
		Password:           getEnvOrDefault("TEST_REDIS_PASSWORD", ""),
		DB:                 getEnvOrDefault("TEST_REDIS_DB", 1),
		PoolSize:           10,
		MinIdleConns:       5,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: 1 * time.Minute,
	}

	return SetupTestRedisFromConfig(t, cfg)
}

func setupLocalTestRedis(t *testing.T) *TestRedis {
	cfg := config.RedisConfig{
		Host:               "localhost",
		Port:               6379,
		Password:           "",
		DB:                 15, // Use a high DB number for tests
		PoolSize:           10,
		MinIdleConns:       5,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		PoolTimeout:        4 * time.Second,
		IdleTimeout:        5 * time.Minute,
		IdleCheckFrequency: 1 * time.Minute,
	}

	return SetupTestRedisFromConfig(t, cfg)
}

// CleanupTestRedis cleans up the test Redis
func CleanupTestRedis(redis *TestRedis) {
	if redis.Cleanup != nil {
		redis.Cleanup()
	}
}

// FlushTestRedis flushes all data from the test Redis
func FlushTestRedis(t *testing.T, client *redis.Client) {
	ctx := context.Background()
	err := client.FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush Redis")
}

// WaitForRedis waits for Redis to be ready
func WaitForRedis(cfg config.RedisConfig, maxRetries int) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := client.Ping(ctx).Err()
		cancel()

		if err == nil {
			client.Close()
			return nil
		}

		time.Sleep(time.Second * 2)
	}

	client.Close()
	return fmt.Errorf("Redis not ready after %d retries", maxRetries)
}

// AssertRedisKeyExists asserts that a key exists in Redis
func AssertRedisKeyExists(t *testing.T, client *redis.Client, key string) {
	ctx := context.Background()
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists, "Key %s should exist", key)
}

// AssertRedisKeyNotExists asserts that a key does not exist in Redis
func AssertRedisKeyNotExists(t *testing.T, client *redis.Client, key string) {
	ctx := context.Background()
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists, "Key %s should not exist", key)
}

// AssertRedisValue asserts that a key has the expected value in Redis
func AssertRedisValue(t *testing.T, client *redis.Client, key string, expected string) {
	ctx := context.Background()
	value, err := client.Get(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, expected, value, "Key %s should have value %s", key, expected)
}

// AssertRedisHashField asserts that a hash field has the expected value in Redis
func AssertRedisHashField(t *testing.T, client *redis.Client, key, field string, expected string) {
	ctx := context.Background()
	value, err := client.HGet(ctx, key, field).Result()
	require.NoError(t, err)
	assert.Equal(t, expected, value, "Hash field %s.%s should have value %s", key, field, expected)
}

// AssertRedisListLength asserts that a list has the expected length in Redis
func AssertRedisListLength(t *testing.T, client *redis.Client, key string, expected int64) {
	ctx := context.Background()
	length, err := client.LLen(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, expected, length, "List %s should have length %d", key, expected)
}

// AssertRedisSetMember asserts that a set contains the expected member in Redis
func AssertRedisSetMember(t *testing.T, client *redis.Client, key, member string) {
	ctx := context.Background()
	isMember, err := client.SIsMember(ctx, key, member).Result()
	require.NoError(t, err)
	assert.True(t, isMember, "Set %s should contain member %s", key, member)
}

// AssertRedisSetNotMember asserts that a set does not contain the member in Redis
func AssertRedisSetNotMember(t *testing.T, client *redis.Client, key, member string) {
	ctx := context.Background()
	isMember, err := client.SIsMember(ctx, key, member).Result()
	require.NoError(t, err)
	assert.False(t, isMember, "Set %s should not contain member %s", key, member)
}

// AssertRedisSortedSetScore asserts that a sorted set member has the expected score in Redis
func AssertRedisSortedSetScore(t *testing.T, client *redis.Client, key, member string, expected float64) {
	ctx := context.Background()
	score, err := client.ZScore(ctx, key, member).Result()
	require.NoError(t, err)
	assert.Equal(t, expected, score, "Sorted set %s member %s should have score %f", key, member, expected)
}

// SetTestRedisData sets up test data in Redis
func SetTestRedisData(t *testing.T, client *redis.Client) map[string]interface{} {
	ctx := context.Background()
	
	// Set some test data
	testData := map[string]interface{}{
		"user:123": `{"id":"123","name":"Test User"}`,
		"session:abc": `{"user_id":"123","expires_at":"2024-12-31T23:59:59Z"}`,
		"cache:users": `["123","456","789"]`,
		"rate_limit:login:123": "5",
	}

	for key, value := range testData {
		err := client.Set(ctx, key, value, time.Hour).Err()
		require.NoError(t, err, "Failed to set test data for key %s", key)
	}

	return testData
}

// MockRedisClient provides a mock Redis client for testing
type MockRedisClient struct {
	data map[string]string
	sets map[string]map[string]struct{}
	lists map[string][]string
	sortedSets map[string]map[string]float64
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:        make(map[string]string),
		sets:        make(map[string]map[string]struct{}),
		lists:       make(map[string][]string),
		sortedSets: make(map[string]map[string]float64),
	}
}

// Set implements Redis Set operation
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.data[key] = fmt.Sprintf("%v", value)
	return redis.NewStatusCmd(ctx)
}

// Get implements Redis Get operation
func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if value, exists := m.data[key]; exists {
		return redis.NewStringResult(value, nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

// Del implements Redis Del operation
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	deleted := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			deleted++
		}
		delete(m.sets, key)
		delete(m.lists, key)
		delete(m.sortedSets, key)
	}
	return redis.NewIntCmd(ctx, deleted)
}

// Exists implements Redis Exists operation
func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	count := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			count++
		}
	}
	return redis.NewIntCmd(ctx, count)
}

// HSet implements Redis HSet operation
func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}
	
	for i := 0; i < len(values); i += 2 {
		field := fmt.Sprintf("%v", values[i])
		value := fmt.Sprintf("%v", values[i+1])
		m.data[key+":"+field] = value
	}
	
	return redis.NewIntCmd(ctx, int64(len(values)/2))
}

// HGet implements Redis HGet operation
func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	if value, exists := m.data[key+":"+field]; exists {
		return redis.NewStringResult(value, nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

// SAdd implements Redis SAdd operation
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}
	
	added := int64(0)
	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member)
		if _, exists := m.sets[key][memberStr]; !exists {
			m.sets[key][memberStr] = struct{}{}
			added++
		}
	}
	
	return redis.NewIntCmd(ctx, added)
}

// SMembers implements Redis SMembers operation
func (m *MockRedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	if set, exists := m.sets[key]; exists {
		members := make([]string, 0, len(set))
		for member := range set {
			members = append(members, member)
		}
		return redis.NewStringSliceCmd(ctx, members, nil)
	}
	return redis.NewStringSliceCmd(ctx, []string{}, nil)
}

// ZAdd implements Redis ZAdd operation
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	if m.sortedSets[key] == nil {
		m.sortedSets[key] = make(map[string]float64)
	}
	
	added := int64(0)
	for _, member := range members {
		memberStr := fmt.Sprintf("%v", member.Member)
		if _, exists := m.sortedSets[key][memberStr]; !exists {
			added++
		}
		m.sortedSets[key][memberStr] = member.Score
	}
	
	return redis.NewIntCmd(ctx, added)
}

// ZScore implements Redis ZScore operation
func (m *MockRedisClient) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	if sortedSet, exists := m.sortedSets[key]; exists {
		if score, exists := sortedSet[member]; exists {
			return redis.NewFloatCmd(ctx, score, nil)
		}
	}
	return redis.NewFloatCmd(ctx, 0, redis.Nil)
}

// FlushDB implements Redis FlushDB operation
func (m *MockRedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	m.data = make(map[string]string)
	m.sets = make(map[string]map[string]struct{})
	m.lists = make(map[string][]string)
	m.sortedSets = make(map[string]map[string]float64)
	return redis.NewStatusCmd(ctx)
}

// Ping implements Redis Ping operation
func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx).SetVal("PONG")
}

// Close implements Redis Close operation
func (m *MockRedisClient) Close() error {
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}