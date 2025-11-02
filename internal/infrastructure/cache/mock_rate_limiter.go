package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MockRateLimiter is a mock implementation of RateLimiter for testing
type MockRateLimiter struct {
	limits map[string]*RateLimitInfo
	
	// Method call tracking
	CalledIsAllowed      bool
	CalledGetRemaining    bool
	CalledReset          bool
	CalledGetKey         bool
	CalledSetKey         bool
	CalledDeleteKey      bool
	CalledIncrement      bool
	CalledGetCount       bool
	CalledSetCount       bool
	CalledExpire         bool
}

// RateLimitInfo represents rate limit information for testing
type RateLimitInfo struct {
	Key       string
	Limit     int
	Remaining int
	ResetTime time.Time
	IsAllowed bool
}

// NewMockRateLimiter creates a new mock rate limiter
func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{
		limits: make(map[string]*RateLimitInfo),
	}
}

// IsAllowed checks if an action is allowed based on rate limits
func (m *MockRateLimiter) IsAllowed(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	m.CalledIsAllowed = true
	
	info, exists := m.limits[key]
	if !exists {
		info = &RateLimitInfo{
			Key:       key,
			Limit:     limit,
			Remaining:  limit - 1,
			ResetTime:  time.Now().Add(window),
			IsAllowed:  true,
		}
		m.limits[key] = info
		return true, limit - 1, info.ResetTime, nil
	}
	
	// Simulate rate limiting
	if info.Remaining <= 0 {
		info.IsAllowed = false
		m.limits[key] = info
		return false, 0, info.ResetTime, nil
	}
	
	info.Remaining--
	info.IsAllowed = true
	m.limits[key] = info
	return true, info.Remaining, info.ResetTime, nil
}

// GetRemaining gets the remaining requests for a key
func (m *MockRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	m.CalledGetRemaining = true
	if info, exists := m.limits[key]; exists {
		return info.Remaining, nil
	}
	return 0, nil
}

// Reset resets the rate limit for a key
func (m *MockRateLimiter) Reset(ctx context.Context, key string) error {
	m.CalledReset = true
	delete(m.limits, key)
	return nil
}

// GetKey gets a value from the rate limiter
func (m *MockRateLimiter) GetKey(ctx context.Context, key string) (string, error) {
	m.CalledGetKey = true
	if info, exists := m.limits[key]; exists {
		return fmt.Sprintf("%d:%d:%t", info.Remaining, info.Limit, info.IsAllowed)
	}
	return "", nil
}

// SetKey sets a value in the rate limiter
func (m *MockRateLimiter) SetKey(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.CalledSetKey = true
	// Parse the value to extract rate limit info
	// This is a simplified implementation for testing
	info := &RateLimitInfo{
		Key:      key,
		Limit:    10, // Default limit
		IsAllowed: true,
	}
	m.limits[key] = info
	return nil
}

// DeleteKey deletes a key from the rate limiter
func (m *MockRateLimiter) DeleteKey(ctx context.Context, key string) error {
	m.CalledDeleteKey = true
	delete(m.limits, key)
	return nil
}

// Increment increments a counter
func (m *MockRateLimiter) Increment(ctx context.Context, key string) (int64, error) {
	m.CalledIncrement = true
	if info, exists := m.limits[key]; exists {
		info.Remaining--
		m.limits[key] = info
		return int64(info.Limit - info.Remaining), nil
	}
	return 1, nil
}

// GetCount gets the count for a key
func (m *MockRateLimiter) GetCount(ctx context.Context, key string) (int64, error) {
	m.CalledGetCount = true
	if info, exists := m.limits[key]; exists {
		return int64(info.Limit - info.Remaining), nil
	}
	return 0, nil
}

// SetCount sets the count for a key
func (m *MockRateLimiter) SetCount(ctx context.Context, key string, count int64, ttl time.Duration) error {
	m.CalledSetCount = true
	info := &RateLimitInfo{
		Key:       key,
		Limit:     int(count),
		Remaining:  int(count),
		IsAllowed:  true,
	}
	m.limits[key] = info
	return nil
}

// Expire sets expiration for a key
func (m *MockRateLimiter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.CalledExpire = true
	if info, exists := m.limits[key]; exists {
		info.ResetTime = time.Now().Add(ttl)
		m.limits[key] = info
	}
	return nil
}

// Reset resets the mock rate limiter state
func (m *MockRateLimiter) Reset() {
	m.limits = make(map[string]*RateLimitInfo)
	
	m.CalledIsAllowed = false
	m.CalledGetRemaining = false
	m.CalledReset = false
	m.CalledGetKey = false
	m.CalledSetKey = false
	m.CalledDeleteKey = false
	m.CalledIncrement = false
	m.CalledGetCount = false
	m.CalledSetCount = false
	m.CalledExpire = false
}

// SetRateLimit sets a specific rate limit for testing
func (m *MockRateLimiter) SetRateLimit(key string, limit int, remaining int, isAllowed bool) {
	info := &RateLimitInfo{
		Key:       key,
		Limit:     limit,
		Remaining:  remaining,
		IsAllowed:  isAllowed,
		ResetTime:  time.Now().Add(time.Hour),
	}
	m.limits[key] = info
}

// GetRateLimit gets the rate limit info for a key
func (m *MockRateLimiter) GetRateLimit(key string) (*RateLimitInfo, bool) {
	info, exists := m.limits[key]
	return info, exists
}

// VerifyMethodCalls verifies that specific methods were called
func (m *MockRateLimiter) VerifyMethodCalls(expectedIsAllowed, expectedGetRemaining, expectedReset, expectedGetKey, expectedSetKey, expectedDeleteKey, expectedIncrement, expectedGetCount, expectedSetCount, expectedExpire bool) error {
	if expectedIsAllowed && !m.CalledIsAllowed {
		return fmt.Errorf("expected IsAllowed to be called")
	}
	if !expectedIsAllowed && m.CalledIsAllowed {
		return fmt.Errorf("expected IsAllowed NOT to be called")
	}
	
	if expectedGetRemaining && !m.CalledGetRemaining {
		return fmt.Errorf("expected GetRemaining to be called")
	}
	if !expectedGetRemaining && m.CalledGetRemaining {
		return fmt.Errorf("expected GetRemaining NOT to be called")
	}
	
	if expectedReset && !m.CalledReset {
		return fmt.Errorf("expected Reset to be called")
	}
	if !expectedReset && m.CalledReset {
		return fmt.Errorf("expected Reset NOT to be called")
	}
	
	if expectedGetKey && !m.CalledGetKey {
		return fmt.Errorf("expected GetKey to be called")
	}
	if !expectedGetKey && m.CalledGetKey {
		return fmt.Errorf("expected GetKey NOT to be called")
	}
	
	if expectedSetKey && !m.CalledSetKey {
		return fmt.Errorf("expected SetKey to be called")
	}
	if !expectedSetKey && m.CalledSetKey {
		return fmt.Errorf("expected SetKey NOT to be called")
	}
	
	if expectedDeleteKey && !m.CalledDeleteKey {
		return fmt.Errorf("expected DeleteKey to be called")
	}
	if !expectedDeleteKey && m.CalledDeleteKey {
		return fmt.Errorf("expected DeleteKey NOT to be called")
	}
	
	if expectedIncrement && !m.CalledIncrement {
		return fmt.Errorf("expected Increment to be called")
	}
	if !expectedIncrement && m.CalledIncrement {
		return fmt.Errorf("expected Increment NOT to be called")
	}
	
	if expectedGetCount && !m.CalledGetCount {
		return fmt.Errorf("expected GetCount to be called")
	}
	if !expectedGetCount && m.CalledGetCount {
		return fmt.Errorf("expected GetCount NOT to be called")
	}
	
	if expectedSetCount && !m.CalledSetCount {
		return fmt.Errorf("expected SetCount to be called")
	}
	if !expectedSetCount && m.CalledSetCount {
		return fmt.Errorf("expected SetCount NOT to be called")
	}
	
	if expectedExpire && !m.CalledExpire {
		return fmt.Errorf("expected Expire to be called")
	}
	if !expectedExpire && m.CalledExpire {
		return fmt.Errorf("expected Expire NOT to be called")
	}
	
	return nil
}