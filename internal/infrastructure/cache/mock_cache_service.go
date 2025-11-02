package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MockCacheService is a mock implementation of CacheService for testing
type MockCacheService struct {
	data map[string]interface{}
	
	// Method call tracking
	CalledSet    bool
	CalledGet    bool
	CalledDelete  bool
	CalledExists  bool
	CalledClear   bool
	CalledSetTTL  bool
	CalledGetTTL  bool
}

// NewMockCacheService creates a new mock cache service
func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string]interface{}),
	}
}

// Set sets a value in the cache
func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.CalledSet = true
	m.data[key] = value
	return nil
}

// Get gets a value from the cache
func (m *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	m.CalledGet = true
	if value, exists := m.data[key]; exists {
		// Try to unmarshal if it's JSON
		if str, ok := value.(string); ok {
			return json.Unmarshal([]byte(str), dest)
		}
		// Direct assignment for simple types
		return json.Unmarshal([]byte("{}"), dest)
	}
	return nil
}

// Delete deletes a value from the cache
func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	m.CalledDelete = true
	delete(m.data, key)
	return nil
}

// Exists checks if a key exists in the cache
func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.CalledExists = true
	_, exists := m.data[key]
	return exists, nil
}

// Clear clears all values from the cache
func (m *MockCacheService) Clear(ctx context.Context) error {
	m.CalledClear = true
	m.data = make(map[string]interface{})
	return nil
}

// SetWithTTL sets a value with TTL
func (m *MockCacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.CalledSetTTL = true
	return m.Set(ctx, key, value, ttl)
}

// GetWithTTL gets a value with TTL
func (m *MockCacheService) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	m.CalledGetTTL = true
	err := m.Get(ctx, key, dest)
	return time.Hour, err // Mock TTL
}

// Reset resets the mock cache service state
func (m *MockCacheService) Reset() {
	m.data = make(map[string]interface{})
	
	m.CalledSet = false
	m.CalledGet = false
	m.CalledDelete = false
	m.CalledExists = false
	m.CalledClear = false
	m.CalledSetTTL = false
	m.CalledGetTTL = false
}

// SetValue sets a specific value for testing
func (m *MockCacheService) SetValue(key string, value interface{}) {
	m.data[key] = value
}

// GetValue gets a specific value for testing
func (m *MockCacheService) GetValue(key string) (interface{}, bool) {
	value, exists := m.data[key]
	return value, exists
}

// VerifyMethodCalls verifies that specific methods were called
func (m *MockCacheService) VerifyMethodCalls(expectedSet, expectedGet, expectedDelete, expectedExists, expectedClear bool) error {
	if expectedSet && !m.CalledSet {
		return fmt.Errorf("expected Set to be called")
	}
	if !expectedSet && m.CalledSet {
		return fmt.Errorf("expected Set NOT to be called")
	}
	
	if expectedGet && !m.CalledGet {
		return fmt.Errorf("expected Get to be called")
	}
	if !expectedGet && m.CalledGet {
		return fmt.Errorf("expected Get NOT to be called")
	}
	
	if expectedDelete && !m.CalledDelete {
		return fmt.Errorf("expected Delete to be called")
	}
	if !expectedDelete && m.CalledDelete {
		return fmt.Errorf("expected Delete NOT to be called")
	}
	
	if expectedExists && !m.CalledExists {
		return fmt.Errorf("expected Exists to be called")
	}
	if !expectedExists && m.CalledExists {
		return fmt.Errorf("expected Exists NOT to be called")
	}
	
	if expectedClear && !m.CalledClear {
		return fmt.Errorf("expected Clear to be called")
	}
	if !expectedClear && m.CalledClear {
		return fmt.Errorf("expected Clear NOT to be called")
	}
	
	return nil
}