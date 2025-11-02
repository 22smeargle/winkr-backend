package cache

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// KeyGenerator generates cache keys with consistent naming
type KeyGenerator struct {
	prefix string
	salt   string
}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator(prefix, salt string) *KeyGenerator {
	return &KeyGenerator{
		prefix: prefix,
		salt:   salt,
	}
}

// GenerateKey creates a cache key with the given parts
func (kg *KeyGenerator) GenerateKey(parts ...string) string {
	if len(parts) == 0 {
		return kg.prefix
	}
	
	key := kg.prefix + strings.Join(parts, ":")
	
	// Add salt if provided
	if kg.salt != "" {
		hasher := fnv.New64a()
		hasher.Write([]byte(key + kg.salt))
		return fmt.Sprintf("%s:%x", kg.prefix, hasher.Sum64())
	}
	
	return key
}

// GenerateHashKey creates a hash-based cache key
func (kg *KeyGenerator) GenerateHashKey(parts ...string) string {
	key := strings.Join(parts, ":")
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return fmt.Sprintf("%s:hash:%x", kg.prefix, hasher.Sum64())
}

// GenerateMD5Key creates an MD5-based cache key
func (kg *KeyGenerator) GenerateMD5Key(parts ...string) string {
	key := strings.Join(parts, ":")
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%s:md5:%x", kg.prefix, hash)
}

// GenerateUUIDKey creates a UUID-based cache key
func (kg *KeyGenerator) GenerateUUIDKey(parts ...string) string {
	id := uuid.New().String()
	if len(parts) > 0 {
		return fmt.Sprintf("%s:%s:%s", kg.prefix, strings.Join(parts, ":"), id)
	}
	return fmt.Sprintf("%s:uuid:%s", kg.prefix, id)
}

// GenerateTimeKey creates a time-based cache key
func (kg *KeyGenerator) GenerateTimeKey(parts ...string) string {
	timestamp := time.Now().Unix()
	key := strings.Join(parts, ":")
	return fmt.Sprintf("%s:%s:%d", kg.prefix, key, timestamp)
}

// GenerateVersionedKey creates a versioned cache key
func (kg *KeyGenerator) GenerateVersionedKey(version string, parts ...string) string {
	key := strings.Join(parts, ":")
	return fmt.Sprintf("%s:v%s:%s", kg.prefix, version, key)
}

// Serialization helpers

// Serialize serializes data to JSON
func Serialize(data interface{}) ([]byte, error) {
	// In a real implementation, you might want to use more efficient serialization
	// like msgpack or protocol buffers for high-performance scenarios
	return json.Marshal(data)
}

// Deserialize deserializes data from JSON
func Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// SerializeToString serializes data to JSON string
func SerializeToString(data interface{}) (string, error) {
	bytes, err := Serialize(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// DeserializeFromString deserializes data from JSON string
func DeserializeFromString(data string, v interface{}) error {
	return Deserialize([]byte(data), v)
}

// Cache invalidation strategies

// InvalidationStrategy defines how cache should be invalidated
type InvalidationStrategy int

const (
	InvalidationStrategyImmediate InvalidationStrategy = iota // Remove immediately
	InvalidationStrategyTTL                                 // Let TTL expire naturally
	InvalidationStrategyVersioned                          // Use versioning
	InvalidationStrategyTagged                             // Use tags for group invalidation
)

// InvalidationPlan defines how to invalidate cache entries
type InvalidationPlan struct {
	Strategy InvalidationStrategy
	Patterns []string
	Tags     []string
	Version  string
	Delay    time.Duration
}

// NewInvalidationPlan creates a new invalidation plan
func NewInvalidationPlan(strategy InvalidationStrategy) *InvalidationPlan {
	return &InvalidationPlan{
		Strategy: strategy,
		Patterns: make([]string, 0),
		Tags:     make([]string, 0),
	}
}

// AddPattern adds a pattern to the invalidation plan
func (ip *InvalidationPlan) AddPattern(pattern string) *InvalidationPlan {
	ip.Patterns = append(ip.Patterns, pattern)
	return ip
}

// AddTag adds a tag to the invalidation plan
func (ip *InvalidationPlan) AddTag(tag string) *InvalidationPlan {
	ip.Tags = append(ip.Tags, tag)
	return ip
}

// SetVersion sets the version for the invalidation plan
func (ip *InvalidationPlan) SetVersion(version string) *InvalidationPlan {
	ip.Version = version
	return ip
}

// SetDelay sets the delay for the invalidation plan
func (ip *InvalidationPlan) SetDelay(delay time.Duration) *InvalidationPlan {
	ip.Delay = delay
	return ip
}

// Cache warming utilities

// WarmupConfig defines cache warming parameters
type WarmupConfig struct {
	Concurrency int           // Number of concurrent warmup workers
	BatchSize   int           // Number of items to warm up per batch
	Delay       time.Duration // Delay between batches
	Timeout     time.Duration // Timeout for warmup operations
}

// NewWarmupConfig creates a new warmup configuration
func NewWarmupConfig() *WarmupConfig {
	return &WarmupConfig{
		Concurrency: 5,
		BatchSize:   100,
		Delay:       100 * time.Millisecond,
		Timeout:     30 * time.Second,
	}
}

// WarmupItem represents an item to be warmed up
type WarmupItem struct {
	Key       string
	Value     interface{}
	TTL       time.Duration
	Priority  int // Higher priority = warmed up first
}

// WarmupQueue manages cache warming operations
type WarmupQueue struct {
	items    []WarmupItem
	config   *WarmupConfig
	keyGen   *KeyGenerator
}

// NewWarmupQueue creates a new warmup queue
func NewWarmupQueue(config *WarmupConfig, keyGen *KeyGenerator) *WarmupQueue {
	return &WarmupQueue{
		items:  make([]WarmupItem, 0),
		config: config,
		keyGen: keyGen,
	}
}

// AddItem adds an item to the warmup queue
func (wq *WarmupQueue) AddItem(key string, value interface{}, ttl time.Duration, priority int) {
	item := WarmupItem{
		Key:      key,
		Value:    value,
		TTL:      ttl,
		Priority: priority,
	}
	
	wq.items = append(wq.items, item)
}

// SortByPriority sorts items by priority (higher first)
func (wq *WarmupQueue) SortByPriority() {
	// Simple bubble sort for demonstration
	// In production, use a more efficient sorting algorithm
	n := len(wq.items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if wq.items[j].Priority < wq.items[j+1].Priority {
				wq.items[j], wq.items[j+1] = wq.items[j+1], wq.items[j]
			}
		}
	}
}

// GetBatch returns the next batch of items to warm up
func (wq *WarmupQueue) GetBatch() []WarmupItem {
	if len(wq.items) == 0 {
		return nil
	}
	
	batchSize := wq.config.BatchSize
	if len(wq.items) < batchSize {
		batchSize = len(wq.items)
	}
	
	batch := wq.items[:batchSize]
	wq.items = wq.items[batchSize:]
	
	return batch
}

// IsEmpty returns true if the queue is empty
func (wq *WarmupQueue) IsEmpty() bool {
	return len(wq.items) == 0
}

// Size returns the number of items in the queue
func (wq *WarmupQueue) Size() int {
	return len(wq.items)
}

// Utility functions

// GenerateCacheKey generates a standardized cache key
func GenerateCacheKey(prefix string, parts ...string) string {
	keyGen := NewKeyGenerator(prefix, "")
	return keyGen.GenerateKey(parts...)
}

// GenerateUserCacheKey generates a user-specific cache key
func GenerateUserCacheKey(userID string, dataType string) string {
	return GenerateCacheKey("user", userID, dataType)
}

// GenerateSessionCacheKey generates a session-specific cache key
func GenerateSessionCacheKey(sessionID string) string {
	return GenerateCacheKey("session", sessionID)
}

// GenerateAPIResponseCacheKey generates an API response cache key
func GenerateAPIResponseCacheKey(endpoint string, params map[string]string) string {
	keyParts := []string{"api", endpoint}
	
	// Add sorted parameters for consistent key generation
	if params != nil {
		for k, v := range params {
			keyParts = append(keyParts, k+":"+v)
		}
	}
	
	return GenerateCacheKey("", keyParts...)
}

// GenerateRateLimitCacheKey generates a rate limit cache key
func GenerateRateLimitCacheKey(keyType, identifier, endpoint string) string {
	return GenerateCacheKey("rate_limit", keyType, endpoint, identifier)
}

// GenerateBlacklistCacheKey generates a blacklist cache key
func GenerateBlacklistCacheKey(token string) string {
	return GenerateCacheKey("blacklist", token)
}

// GeneratePubSubChannel generates a Pub/Sub channel name
func GeneratePubSubChannel(channelType string, parts ...string) string {
	keyParts := append([]string{channelType}, parts...)
	return GenerateCacheKey("pubsub", keyParts...)
}

// ParseTTL parses a TTL string into time.Duration
func ParseTTL(ttlStr string) (time.Duration, error) {
	if ttlStr == "" {
		return 0, nil
	}
	
	// Try to parse as duration first
	if duration, err := time.ParseDuration(ttlStr); err == nil {
		return duration, nil
	}
	
	// Try to parse as seconds
	if seconds, err := strconv.Atoi(ttlStr); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}
	
	// Try to parse as minutes
	if minutes, err := strconv.Atoi(ttlStr); err == nil {
		return time.Duration(minutes) * time.Minute, nil
	}
	
	return 0, fmt.Errorf("invalid TTL format: %s", ttlStr)
}

// FormatTTL formats a time.Duration into a TTL string
func FormatTTL(ttl time.Duration) string {
	return ttl.String()
}

// IsValidCacheKey checks if a cache key is valid
func IsValidCacheKey(key string) bool {
	if key == "" {
		return false
	}
	
	// Check for invalid characters
	invalidChars := []string{" ", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(key, char) {
			return false
		}
	}
	
	// Check key length
	if len(key) > 250 {
		return false
	}
	
	return true
}

// ExtractKeyParts extracts parts from a cache key
func ExtractKeyParts(key string) []string {
	return strings.Split(key, ":")
}

// GetKeyType returns the type of cache key
func GetKeyType(key string) string {
	parts := ExtractKeyParts(key)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}