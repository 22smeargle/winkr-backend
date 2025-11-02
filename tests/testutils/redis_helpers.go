package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

// RedisHelper provides utilities for Redis testing
type RedisHelper struct {
	t      *testing.T
	client *redis.Client
	config *TestConfig
}

// NewRedisHelper creates a new Redis helper
func NewRedisHelper(t *testing.T, client *redis.Client, config *TestConfig) *RedisHelper {
	return &RedisHelper{
		t:      t,
		client: client,
		config: config,
	}
}

// FlushDB flushes the current Redis database
func (rh *RedisHelper) FlushDB() {
	err := rh.client.FlushDB(context.Background()).Err()
	require.NoError(rh.t, err, "Failed to flush Redis database")
}

// FlushAll flushes all Redis databases
func (rh *RedisHelper) FlushAll() {
	err := rh.client.FlushAll(context.Background()).Err()
	require.NoError(rh.t, err, "Failed to flush all Redis databases")
}

// KeyExists checks if a key exists
func (rh *RedisHelper) KeyExists(key string) bool {
	count, err := rh.client.Exists(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to check if key exists")
	return count > 0
}

// AssertKeyExists asserts that a key exists
func (rh *RedisHelper) AssertKeyExists(key string) {
	require.True(rh.t, rh.KeyExists(key), "Key %s should exist", key)
}

// AssertKeyNotExists asserts that a key does not exist
func (rh *RedisHelper) AssertKeyNotExists(key string) {
	require.False(rh.t, rh.KeyExists(key), "Key %s should not exist", key)
}

// SetString sets a string value
func (rh *RedisHelper) SetString(key, value string) {
	err := rh.client.Set(context.Background(), key, value, 0).Err()
	require.NoError(rh.t, err, "Failed to set string value")
}

// SetStringWithTTL sets a string value with TTL
func (rh *RedisHelper) SetStringWithTTL(key, value string, ttl time.Duration) {
	err := rh.client.Set(context.Background(), key, value, ttl).Err()
	require.NoError(rh.t, err, "Failed to set string value with TTL")
}

// GetString gets a string value
func (rh *RedisHelper) GetString(key string) string {
	value, err := rh.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return ""
	}
	require.NoError(rh.t, err, "Failed to get string value")
	return value
}

// AssertStringEqual asserts that a string key has the expected value
func (rh *RedisHelper) AssertStringEqual(key, expectedValue string) {
	actualValue := rh.GetString(key)
	require.Equal(rh.t, expectedValue, actualValue, "String value for key %s should match", key)
}

// SetJSON sets a JSON value
func (rh *RedisHelper) SetJSON(key string, value interface{}) {
	jsonData, err := json.Marshal(value)
	require.NoError(rh.t, err, "Failed to marshal JSON value")
	
	err = rh.client.Set(context.Background(), key, jsonData, 0).Err()
	require.NoError(rh.t, err, "Failed to set JSON value")
}

// SetJSONWithTTL sets a JSON value with TTL
func (rh *RedisHelper) SetJSONWithTTL(key string, value interface{}, ttl time.Duration) {
	jsonData, err := json.Marshal(value)
	require.NoError(rh.t, err, "Failed to marshal JSON value")
	
	err = rh.client.Set(context.Background(), key, jsonData, ttl).Err()
	require.NoError(rh.t, err, "Failed to set JSON value with TTL")
}

// GetJSON gets a JSON value
func (rh *RedisHelper) GetJSON(key string, dest interface{}) {
	value, err := rh.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return
	}
	require.NoError(rh.t, err, "Failed to get JSON value")
	
	err = json.Unmarshal([]byte(value), dest)
	require.NoError(rh.t, err, "Failed to unmarshal JSON value")
}

// AssertJSONEqual asserts that a JSON key has the expected value
func (rh *RedisHelper) AssertJSONEqual(key string, expectedValue interface{}) {
	var actualValue interface{}
	rh.GetJSON(key, &actualValue)
	require.Equal(rh.t, expectedValue, actualValue, "JSON value for key %s should match", key)
}

// SetHash sets a hash field
func (rh *RedisHelper) SetHash(key, field string, value interface{}) {
	err := rh.client.HSet(context.Background(), key, field, value).Err()
	require.NoError(rh.t, err, "Failed to set hash field")
}

// GetHash gets a hash field
func (rh *RedisHelper) GetHash(key, field string) string {
	value, err := rh.client.HGet(context.Background(), key, field).Result()
	if err == redis.Nil {
		return ""
	}
	require.NoError(rh.t, err, "Failed to get hash field")
	return value
}

// GetAllHash gets all hash fields
func (rh *RedisHelper) GetAllHash(key string) map[string]string {
	values, err := rh.client.HGetAll(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get all hash fields")
	return values
}

// AssertHashFieldEqual asserts that a hash field has the expected value
func (rh *RedisHelper) AssertHashFieldEqual(key, field, expectedValue string) {
	actualValue := rh.GetHash(key, field)
	require.Equal(rh.t, expectedValue, actualValue, "Hash field %s.%s should match", key, field)
}

// AssertHashExists asserts that a hash exists
func (rh *RedisHelper) AssertHashExists(key string) {
	require.True(rh.t, rh.KeyExists(key), "Hash %s should exist", key)
}

// AssertHashNotExists asserts that a hash does not exist
func (rh *RedisHelper) AssertHashNotExists(key string) {
	require.False(rh.t, rh.KeyExists(key), "Hash %s should not exist", key)
}

// SetList sets a list value
func (rh *RedisHelper) SetList(key string, values []interface{}) {
	err := rh.client.Del(context.Background(), key).Err()
	require.NoError(rh.t, err, "Failed to delete existing list")
	
	if len(values) > 0 {
		err = rh.client.LPush(context.Background(), key, values...).Err()
		require.NoError(rh.t, err, "Failed to set list value")
	}
}

// GetList gets a list value
func (rh *RedisHelper) GetList(key string) []string {
	values, err := rh.client.LRange(context.Background(), key, 0, -1).Result()
	require.NoError(rh.t, err, "Failed to get list value")
	return values
}

// AssertListLength asserts that a list has the expected length
func (rh *RedisHelper) AssertListLength(key string, expectedLength int) {
	length, err := rh.client.LLen(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get list length")
	require.Equal(rh.t, int64(expectedLength), length, "List %s should have length %d", key, expectedLength)
}

// AssertListContains asserts that a list contains a specific value
func (rh *RedisHelper) AssertListContains(key, expectedValue string) {
	values := rh.GetList(key)
	require.Contains(rh.t, values, expectedValue, "List %s should contain value %s", key, expectedValue)
}

// SetSet sets a set value
func (rh *RedisHelper) SetSet(key string, values []interface{}) {
	err := rh.client.Del(context.Background(), key).Err()
	require.NoError(rh.t, err, "Failed to delete existing set")
	
	if len(values) > 0 {
		err = rh.client.SAdd(context.Background(), key, values...).Err()
		require.NoError(rh.t, err, "Failed to set set value")
	}
}

// GetSet gets a set value
func (rh *RedisHelper) GetSet(key string) []string {
	values, err := rh.client.SMembers(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get set value")
	return values
}

// AssertSetSize asserts that a set has the expected size
func (rh *RedisHelper) AssertSetSize(key string, expectedSize int) {
	size, err := rh.client.SCard(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get set size")
	require.Equal(rh.t, int64(expectedSize), size, "Set %s should have size %d", key, expectedSize)
}

// AssertSetContains asserts that a set contains a specific value
func (rh *RedisHelper) AssertSetContains(key, expectedValue string) {
	values := rh.GetSet(key)
	require.Contains(rh.t, values, expectedValue, "Set %s should contain value %s", key, expectedValue)
}

// SetZSet sets a sorted set value
func (rh *RedisHelper) SetZSet(key string, members map[string]float64) {
	err := rh.client.Del(context.Background(), key).Err()
	require.NoError(rh.t, err, "Failed to delete existing sorted set")
	
	for member, score := range members {
		err = rh.client.ZAdd(context.Background(), key, &redis.Z{
			Score:  score,
			Member: member,
		}).Err()
		require.NoError(rh.t, err, "Failed to add member to sorted set")
	}
}

// GetZSet gets a sorted set value
func (rh *RedisHelper) GetZSet(key string) []redis.Z {
	values, err := rh.client.ZRangeWithScores(context.Background(), key, 0, -1).Result()
	require.NoError(rh.t, err, "Failed to get sorted set value")
	return values
}

// AssertZSetSize asserts that a sorted set has the expected size
func (rh *RedisHelper) AssertZSetSize(key string, expectedSize int) {
	size, err := rh.client.ZCard(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get sorted set size")
	require.Equal(rh.t, int64(expectedSize), size, "Sorted set %s should have size %d", key, expectedSize)
}

// AssertZSetContains asserts that a sorted set contains a specific member
func (rh *RedisHelper) AssertZSetContains(key, member string, expectedScore float64) {
	score, err := rh.client.ZScore(context.Background(), key, member).Result()
	require.NoError(rh.t, err, "Failed to get member score from sorted set")
	require.Equal(rh.t, expectedScore, score, "Member %s in sorted set %s should have score %f", member, key, expectedScore)
}

// GetTTL gets the TTL of a key
func (rh *RedisHelper) GetTTL(key string) time.Duration {
	ttl, err := rh.client.TTL(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to get TTL")
	return ttl
}

// AssertTTL asserts that a key has the expected TTL (within tolerance)
func (rh *RedisHelper) AssertTTL(key string, expectedTTL time.Duration, tolerance time.Duration) {
	actualTTL := rh.GetTTL(key)
	require.InDelta(rh.t, expectedTTL.Seconds(), actualTTL.Seconds(), tolerance.Seconds(), "TTL for key %s should be within tolerance", key)
}

// Expire sets expiration for a key
func (rh *RedisHelper) Expire(key string, ttl time.Duration) {
	err := rh.client.Expire(context.Background(), key, ttl).Err()
	require.NoError(rh.t, err, "Failed to set expiration")
}

// Persist removes expiration from a key
func (rh *RedisHelper) Persist(key string) {
	err := rh.client.Persist(context.Background(), key).Err()
	require.NoError(rh.t, err, "Failed to remove expiration")
}

// DeleteKey deletes a key
func (rh *RedisHelper) DeleteKey(key string) {
	err := rh.client.Del(context.Background(), key).Err()
	require.NoError(rh.t, err, "Failed to delete key")
}

// DeleteKeys deletes multiple keys
func (rh *RedisHelper) DeleteKeys(keys ...string) {
	err := rh.client.Del(context.Background(), keys...).Err()
	require.NoError(rh.t, err, "Failed to delete keys")
}

// GetKeys gets keys matching a pattern
func (rh *RedisHelper) GetKeys(pattern string) []string {
	keys, err := rh.client.Keys(context.Background(), pattern).Result()
	require.NoError(rh.t, err, "Failed to get keys")
	return keys
}

// AssertKeyCount asserts that the number of keys matching a pattern is expected
func (rh *RedisHelper) AssertKeyCount(pattern string, expectedCount int) {
	keys := rh.GetKeys(pattern)
	require.Len(rh.t, keys, expectedCount, "Number of keys matching pattern %s should be %d", pattern, expectedCount)
}

// Increment increments a numeric value
func (rh *RedisHelper) Increment(key string) int64 {
	value, err := rh.client.Incr(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to increment value")
	return value
}

// IncrementBy increments a numeric value by a specific amount
func (rh *RedisHelper) IncrementBy(key string, value int64) int64 {
	result, err := rh.client.IncrBy(context.Background(), key, value).Result()
	require.NoError(rh.t, err, "Failed to increment value by amount")
	return result
}

// Decrement decrements a numeric value
func (rh *RedisHelper) Decrement(key string) int64 {
	value, err := rh.client.Decr(context.Background(), key).Result()
	require.NoError(rh.t, err, "Failed to decrement value")
	return value
}

// DecrementBy decrements a numeric value by a specific amount
func (rh *RedisHelper) DecrementBy(key string, value int64) int64 {
	result, err := rh.client.DecrBy(context.Background(), key, value).Result()
	require.NoError(rh.t, err, "Failed to decrement value by amount")
	return result
}

// AssertNumericValue asserts that a key has the expected numeric value
func (rh *RedisHelper) AssertNumericValue(key string, expectedValue int64) {
	value, err := rh.client.Get(context.Background(), key).Int64()
	if err == redis.Nil {
		require.Fail(rh.t, "Key %s does not exist", key)
		return
	}
	require.NoError(rh.t, err, "Failed to get numeric value")
	require.Equal(rh.t, expectedValue, value, "Numeric value for key %s should match", key)
}

// Publish publishes a message to a channel
func (rh *RedisHelper) Publish(channel, message string) {
	err := rh.client.Publish(context.Background(), channel, message).Err()
	require.NoError(rh.t, err, "Failed to publish message")
}

// Subscribe subscribes to a channel
func (rh *RedisHelper) Subscribe(channel string) *redis.PubSub {
	pubsub := rh.client.Subscribe(context.Background(), channel)
	require.NoError(rh.t, pubsub.Err(), "Failed to subscribe to channel")
	return pubsub
}

// WaitForMessage waits for a message on a channel
func (rh *RedisHelper) WaitForMessage(pubsub *redis.PubSub, timeout time.Duration) *redis.Message {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	msg, err := pubsub.ReceiveMessage(ctx)
	require.NoError(rh.t, err, "Failed to receive message")
	return msg
}

// AssertMessageReceived asserts that a message was received on a channel
func (rh *RedisHelper) AssertMessageReceived(pubsub *redis.PubSub, expectedChannel, expectedMessage string, timeout time.Duration) {
	msg := rh.WaitForMessage(pubsub, timeout)
	require.Equal(rh.t, expectedChannel, msg.Channel, "Message channel should match")
	require.Equal(rh.t, expectedMessage, msg.Payload, "Message payload should match")
}

// SetRateLimit sets a rate limit
func (rh *RedisHelper) SetRateLimit(key string, limit int, window time.Duration) {
	pipe := rh.client.Pipeline()
	pipe.Incr(context.Background(), key)
	pipe.Expire(context.Background(), key, window)
	
	_, err := pipe.Exec(context.Background())
	require.NoError(rh.t, err, "Failed to set rate limit")
}

// AssertRateLimitExceeded asserts that a rate limit is exceeded
func (rh *RedisHelper) AssertRateLimitExceeded(key string, limit int) {
	count, err := rh.client.Get(context.Background(), key).Int()
	if err == redis.Nil {
		count = 0
	} else {
		require.NoError(rh.t, err, "Failed to get rate limit count")
	}
	require.Greater(rh.t, count, limit, "Rate limit should be exceeded")
}

// AssertRateLimitNotExceeded asserts that a rate limit is not exceeded
func (rh *RedisHelper) AssertRateLimitNotExceeded(key string, limit int) {
	count, err := rh.client.Get(context.Background(), key).Int()
	if err == redis.Nil {
		count = 0
	} else {
		require.NoError(rh.t, err, "Failed to get rate limit count")
	}
	require.LessOrEqual(rh.t, count, limit, "Rate limit should not be exceeded")
}

// SetCache sets a cache value
func (rh *RedisHelper) SetCache(key string, value interface{}, ttl time.Duration) {
	rh.SetJSONWithTTL(key, value, ttl)
}

// GetCache gets a cache value
func (rh *RedisHelper) GetCache(key string, dest interface{}) bool {
	err := rh.client.Get(context.Background(), key).Scan(dest)
	if err == redis.Nil {
		return false
	}
	require.NoError(rh.t, err, "Failed to get cache value")
	return true
}

// AssertCacheHit asserts that a cache hit occurs
func (rh *RedisHelper) AssertCacheHit(key string, expectedValue interface{}) {
	var actualValue interface{}
	hit := rh.GetCache(key, &actualValue)
	require.True(rh.t, hit, "Cache hit should occur for key %s", key)
	require.Equal(rh.t, expectedValue, actualValue, "Cached value should match expected")
}

// AssertCacheMiss asserts that a cache miss occurs
func (rh *RedisHelper) AssertCacheMiss(key string) {
	var value interface{}
	hit := rh.GetCache(key, &value)
	require.False(rh.t, hit, "Cache miss should occur for key %s", key)
}

// InvalidateCache invalidates cache entries matching a pattern
func (rh *RedisHelper) InvalidateCache(pattern string) {
	keys := rh.GetKeys(pattern)
	if len(keys) > 0 {
		rh.DeleteKeys(keys...)
	}
}

// WaitForKey waits for a key to appear
func (rh *RedisHelper) WaitForKey(key string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if rh.KeyExists(key) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Fail(rh.t, "Key %s did not appear within timeout", key)
}

// WaitForKeyToDisappear waits for a key to disappear
func (rh *RedisHelper) WaitForKeyToDisappear(key string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if !rh.KeyExists(key) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Fail(rh.t, "Key %s did not disappear within timeout", key)
}

// GetInfo gets Redis info
func (rh *RedisHelper) GetInfo() map[string]string {
	info, err := rh.client.Info(context.Background()).Result()
	require.NoError(rh.t, err, "Failed to get Redis info")
	
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}
	return result
}

// Ping tests Redis connection
func (rh *RedisHelper) Ping() {
	err := rh.client.Ping(context.Background()).Err()
	require.NoError(rh.t, err, "Redis ping failed")
}

// GetClient returns the Redis client
func (rh *RedisHelper) GetClient() *redis.Client {
	return rh.client
}

// Close closes the Redis connection
func (rh *RedisHelper) Close() error {
	return rh.client.Close()
}