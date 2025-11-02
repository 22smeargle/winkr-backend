package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// RedisClient represents the Redis connection
type RedisClient struct {
	Client   redis.Cmdable
	isCluster bool
	metrics  *RedisMetrics
}

// RedisMetrics holds Redis connection metrics
type RedisMetrics struct {
	mu                    sync.RWMutex
	ConnectionsCreated    int64
	ConnectionsClosed     int64
	ConnectionErrors      int64
	CommandsExecuted      int64
	CommandErrors         int64
	LastConnectionTime    time.Time
	LastErrorTime         time.Time
	PoolStats             map[string]interface{}
}

// NewRedisClient creates a new Redis connection
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	var client redis.Cmdable
	var isCluster bool

	if cfg.ClusterEnabled && len(cfg.ClusterAddresses) > 0 {
		// Create cluster client
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:            cfg.ClusterAddresses,
			Password:         cfg.Password,
			MaxRetries:       cfg.MaxRetries,
			DialTimeout:      cfg.DialTimeout,
			ReadTimeout:      cfg.ReadTimeout,
			WriteTimeout:     cfg.WriteTimeout,
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     cfg.MinIdleConns,
			PoolTimeout:      cfg.PoolTimeout,
			IdleTimeout:      cfg.IdleTimeout,
			IdleCheckFrequency: cfg.IdleCheckFrequency,
			MaxRedirects:     cfg.MaxRedirects,
			RouteByLatency:    cfg.RouteByLatency,
			RouteRandomly:     cfg.RouteRandomly,
		})
		isCluster = true
		logger.Info("Redis cluster client created", "addresses", cfg.ClusterAddresses)
	} else {
		// Create single client
		client = redis.NewClient(&redis.Options{
			Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:        cfg.Password,
			DB:              cfg.DB,
			MaxRetries:      cfg.MaxRetries,
			DialTimeout:     cfg.DialTimeout,
			ReadTimeout:     cfg.ReadTimeout,
			WriteTimeout:    cfg.WriteTimeout,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			PoolTimeout:     cfg.PoolTimeout,
			IdleTimeout:     cfg.IdleTimeout,
			IdleCheckFrequency: cfg.IdleCheckFrequency,
		})
		isCluster = false
		logger.Info("Redis single client created", "address", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	metrics := &RedisMetrics{
		LastConnectionTime: time.Now(),
		PoolStats:          make(map[string]interface{}),
	}

	logger.Info("Redis connection established successfully")
	return &RedisClient{
		Client:   client,
		isCluster: isCluster,
		metrics:  metrics,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	var err error
	
	if r.isCluster {
		if clusterClient, ok := r.Client.(*redis.ClusterClient); ok {
			err = clusterClient.Close()
		}
	} else {
		if singleClient, ok := r.Client.(*redis.Client); ok {
			err = singleClient.Close()
		}
	}

	if err != nil {
		logger.Error("Failed to close Redis connection", err)
		r.metrics.mu.Lock()
		r.metrics.ConnectionErrors++
		r.metrics.LastErrorTime = time.Now()
		r.metrics.mu.Unlock()
		return err
	}

	r.metrics.mu.Lock()
	r.metrics.ConnectionsClosed++
	r.metrics.mu.Unlock()

	logger.Info("Redis connection closed")
	return nil
}

// Ping checks if the Redis connection is alive
func (r *RedisClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		r.metrics.mu.Lock()
		r.metrics.ConnectionErrors++
		r.metrics.LastErrorTime = time.Now()
		r.metrics.mu.Unlock()
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	r.metrics.mu.Lock()
	r.metrics.CommandsExecuted++
	r.metrics.mu.Unlock()

	return nil
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() redis.Cmdable {
	return r.Client
}

// IsCluster returns true if using Redis cluster
func (r *RedisClient) IsCluster() bool {
	return r.isCluster
}

// Health checks the Redis health
func (r *RedisClient) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if we can ping Redis
	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		r.metrics.mu.Lock()
		r.metrics.ConnectionErrors++
		r.metrics.LastErrorTime = time.Now()
		r.metrics.mu.Unlock()
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// GetStats returns Redis connection statistics
func (r *RedisClient) GetStats() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := r.Client.Info(ctx).Result()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Get pool stats
	var poolStats map[string]interface{}
	if r.isCluster {
		if clusterClient, ok := r.Client.(*redis.ClusterClient); ok {
			poolStats = clusterClient.PoolStats()
		}
	} else {
		if singleClient, ok := r.Client.(*redis.Client); ok {
			stats := singleClient.PoolStats()
			poolStats = map[string]interface{}{
				"hits":         stats.Hits,
				"misses":       stats.Misses,
				"timeouts":     stats.Timeouts,
				"total_conns":  stats.TotalConns,
				"idle_conns":   stats.IdleConns,
				"stale_conns":  stats.StaleConns,
			}
		}
	}

	r.metrics.mu.RLock()
	metrics := map[string]interface{}{
		"info":                info,
		"is_cluster":          r.isCluster,
		"connections_created": r.metrics.ConnectionsCreated,
		"connections_closed":  r.metrics.ConnectionsClosed,
		"connection_errors":   r.metrics.ConnectionErrors,
		"commands_executed":   r.metrics.CommandsExecuted,
		"command_errors":      r.metrics.CommandErrors,
		"last_connection":     r.metrics.LastConnectionTime,
		"last_error":          r.metrics.LastErrorTime,
		"pool_stats":          poolStats,
	}
	r.metrics.mu.RUnlock()

	return metrics
}

// GetMetrics returns Redis metrics
func (r *RedisClient) GetMetrics() *RedisMetrics {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()
	
	// Return a copy to avoid concurrent access issues
	return &RedisMetrics{
		ConnectionsCreated: r.metrics.ConnectionsCreated,
		ConnectionsClosed:  r.metrics.ConnectionsClosed,
		ConnectionErrors:   r.metrics.ConnectionErrors,
		CommandsExecuted:   r.metrics.CommandsExecuted,
		CommandErrors:      r.metrics.CommandErrors,
		LastConnectionTime: r.metrics.LastConnectionTime,
		LastErrorTime:      r.metrics.LastErrorTime,
		PoolStats:          r.metrics.PoolStats,
	}
}

// Set sets a key-value pair with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	r.updateMetrics(err)
	return err
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := r.Client.Get(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// Del deletes a key
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	err := r.Client.Del(ctx, keys...).Err()
	r.updateMetrics(err)
	return err
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	r.updateMetrics(err)
	return result > 0, err
}

// Expire sets expiration for a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.Client.Expire(ctx, key, expiration).Err()
	r.updateMetrics(err)
	return err
}

// TTL returns the time to live for a key
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := r.Client.TTL(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// HSet sets a field in a hash
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	err := r.Client.HSet(ctx, key, field, value).Err()
	r.updateMetrics(err)
	return err
}

// HGet retrieves a field from a hash
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := r.Client.HGet(ctx, key, field).Result()
	r.updateMetrics(err)
	return result, err
}

// HGetAll retrieves all fields and values from a hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := r.Client.HGetAll(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// HDel deletes a field from a hash
func (r *RedisClient) HDel(ctx context.Context, key, field string) error {
	err := r.Client.HDel(ctx, key, field).Err()
	r.updateMetrics(err)
	return err
}

// LPush adds an element to the left side of a list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	err := r.Client.LPush(ctx, key, values...).Err()
	r.updateMetrics(err)
	return err
}

// RPush adds an element to the right side of a list
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	err := r.Client.RPush(ctx, key, values...).Err()
	r.updateMetrics(err)
	return err
}

// LPop removes and returns the leftmost element from a list
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	result, err := r.Client.LPop(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// RPop removes and returns the rightmost element from a list
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	result, err := r.Client.RPop(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// LLen returns the length of a list
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	result, err := r.Client.LLen(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// LRange returns a range of elements from a list
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result, err := r.Client.LRange(ctx, key, start, stop).Result()
	r.updateMetrics(err)
	return result, err
}

// SAdd adds a member to a set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	err := r.Client.SAdd(ctx, key, members...).Err()
	r.updateMetrics(err)
	return err
}

// SRem removes a member from a set
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	err := r.Client.SRem(ctx, key, members...).Err()
	r.updateMetrics(err)
	return err
}

// SMembers returns all members of a set
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	result, err := r.Client.SMembers(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// SIsMember checks if a member is in a set
func (r *RedisClient) SIsMember(ctx context.Context, key, member interface{}) (bool, error) {
	result, err := r.Client.SIsMember(ctx, key, member).Result()
	r.updateMetrics(err)
	return result, err
}

// Incr increments the numeric value of a key by 1
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	result, err := r.Client.Incr(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// IncrBy increments the numeric value of a key by the given amount
func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.Client.IncrBy(ctx, key, value).Result()
	r.updateMetrics(err)
	return result, err
}

// Decr decrements the numeric value of a key by 1
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	result, err := r.Client.Decr(ctx, key).Result()
	r.updateMetrics(err)
	return result, err
}

// DecrBy decrements the numeric value of a key by the given amount
func (r *RedisClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.Client.DecrBy(ctx, key, value).Result()
	r.updateMetrics(err)
	return result, err
}

// FlushDB removes all keys from the current database
func (r *RedisClient) FlushDB(ctx context.Context) error {
	err := r.Client.FlushDB(ctx).Err()
	r.updateMetrics(err)
	return err
}

// DBSize returns the number of keys in the current database
func (r *RedisClient) DBSize(ctx context.Context) (int64, error) {
	result, err := r.Client.DBSize(ctx).Result()
	r.updateMetrics(err)
	return result, err
}

// updateMetrics updates the Redis metrics
func (r *RedisClient) updateMetrics(err error) {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()
	
	r.metrics.CommandsExecuted++
	if err != nil {
		r.metrics.CommandErrors++
		r.metrics.LastErrorTime = time.Now()
	}
}

// Publish publishes a message to a channel
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	err := r.Client.Publish(ctx, channel, message).Err()
	r.updateMetrics(err)
	return err
}

// Subscribe subscribes to channels
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	var pubsub *redis.PubSub
	
	if r.isCluster {
		if clusterClient, ok := r.Client.(*redis.ClusterClient); ok {
			pubsub = clusterClient.Subscribe(ctx, channels...)
		}
	} else {
		if singleClient, ok := r.Client.(*redis.Client); ok {
			pubsub = singleClient.Subscribe(ctx, channels...)
		}
	}
	
	r.updateMetrics(nil)
	return pubsub
}

// PSubscribe subscribes to channels by pattern
func (r *RedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	var pubsub *redis.PubSub
	
	if r.isCluster {
		if clusterClient, ok := r.Client.(*redis.ClusterClient); ok {
			pubsub = clusterClient.PSubscribe(ctx, patterns...)
		}
	} else {
		if singleClient, ok := r.Client.(*redis.Client); ok {
			pubsub = singleClient.PSubscribe(ctx, patterns...)
		}
	}
	
	r.updateMetrics(nil)
	return pubsub
}