package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// MockRedisClient implements RedisClient for testing
type MockRedisClient struct {
	data map[string]string
	ttls map[string]time.Duration
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
		ttls: make(map[string]time.Duration),
	}
}

// Set stores a key-value pair with expiry
func (m *MockRedisClient) Set(ctx context.Context, key, value string, expiry time.Duration) error {
	m.data[key] = value
	if expiry > 0 {
		m.ttls[key] = expiry
	}
	return nil
}

// Get retrieves a value by key
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	value, exists := m.data[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

// Del deletes a key
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
		delete(m.ttls, key)
	}
	return nil
}

// Exists checks if a key exists
func (m *MockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.data[key]
	return exists, nil
}

// Incr increments a numeric value
func (m *MockRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	value, exists := m.data[key]
	if !exists {
		m.data[key] = "1"
		return 1, nil
	}
	
	currentValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		currentValue = 0
	}
	
	currentValue++
	m.data[key] = fmt.Sprintf("%d", currentValue)
	return int64(currentValue), nil
}

// HMSet stores a hash
func (m *MockRedisClient) HMSet(ctx context.Context, key string, data map[string]interface{}) error {
	for field, value := range data {
		m.data[fmt.Sprintf("%s:%s", key, field)] = fmt.Sprintf("%v", value)
	}
	return nil
}

// HGetAll retrieves all hash fields
func (m *MockRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result := make(map[string]string)
	
	// Find all fields for this hash
	prefix := key + ":"
	for k, v := range m.data {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			field := k[len(prefix):]
			result[field] = v
		}
	}
	
	return result, nil
}

// Expire sets expiry for a key
func (m *MockRedisClient) Expire(ctx context.Context, key string, expiry time.Duration) error {
	m.ttls[key] = expiry
	return nil
}

// TTL returns time to live for a key
func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, exists := m.ttls[key]
	if !exists {
		return 0, nil
	}
	return ttl, nil
}

// SAdd adds member to a set
func (m *MockRedisClient) SAdd(ctx context.Context, key, member string) error {
	setKey := key + "_set"
	currentSet := m.data[setKey]
	if currentSet == "" {
		m.data[setKey] = member
	} else {
		m.data[setKey] = currentSet + "," + member
	}
	return nil
}

// SMembers returns all members of a set
func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	setKey := key + "_set"
	currentSet := m.data[setKey]
	if currentSet == "" {
		return []string{}, nil
	}
	
	// Split by comma (simple mock implementation)
	members := []string{currentSet}
	return members, nil
}

// Clear clears all mock data
func (m *MockRedisClient) Clear() {
	m.data = make(map[string]string)
	m.ttls = make(map[string]time.Duration)
}