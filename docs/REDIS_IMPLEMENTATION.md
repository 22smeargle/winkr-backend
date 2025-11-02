# Redis Implementation Guide

This document provides a comprehensive guide to the Redis implementation for the Winkr backend application.

## Overview

The Redis implementation provides:

1. **Session Management**: User session storage with device information and activity tracking
2. **Caching Service**: Multi-layer caching for user profiles, photos, matches, and API responses
3. **Rate Limiting**: Sliding window rate limiting with distributed support
4. **Token Blacklist**: JWT token blacklisting for secure logout
5. **Pub/Sub Service**: Real-time messaging and notifications
6. **Health Monitoring**: Comprehensive health checks and metrics

## Architecture

### Core Components

1. **Redis Connection** (`internal/infrastructure/database/redis/connection.go`)
   - Connection pooling and health checks
   - Cluster support
   - Metrics tracking
   - Automatic reconnection

2. **Session Manager** (`internal/infrastructure/cache/session_manager.go`)
   - Multi-device session support
   - Activity tracking
   - Session expiration
   - Online status management

3. **Cache Service** (`internal/infrastructure/cache/cache_service.go`)
   - User profile caching
   - Photo metadata caching
   - Match recommendations caching
   - API response caching
   - Geospatial data caching

4. **Rate Limiter** (`internal/infrastructure/cache/rate_limiter.go`)
   - Sliding window implementation
   - Distributed rate limiting
   - IP-based and user-based limits
   - Endpoint-specific limits

5. **Token Blacklist** (`internal/infrastructure/cache/token_blacklist.go`)
   - JWT token blacklisting
   - Automatic cleanup
   - Cross-instance synchronization

6. **Pub/Sub Service** (`internal/infrastructure/cache/pubsub_service.go`)
   - Chat message broadcasting
   - Online status tracking
   - Notification system
   - Match notifications

7. **WebSocket Manager** (`internal/infrastructure/websocket/connection_manager.go`)
   - Connection management
   - Message broadcasting
   - Health monitoring

8. **Cache Utilities** (`pkg/cache/utils.go`)
   - Key generation
   - Serialization helpers
   - Cache invalidation
   - Cache warming

## Configuration

### Redis Configuration

```go
type RedisConfig struct {
    Address             string        `mapstructure:"address"`
    Password           string        `mapstructure:"password"`
    DB                 int           `mapstructure:"db"`
    PoolSize           int           `mapstructure:"pool_size"`
    MinIdleConns       int           `mapstructure:"min_idle_conns"`
    MaxRetries         int           `mapstructure:"max_retries"`
    DialTimeout        time.Duration `mapstructure:"dial_timeout"`
    ReadTimeout        time.Duration `mapstructure:"read_timeout"`
    WriteTimeout       time.Duration `mapstructure:"write_timeout"`
    PoolTimeout        time.Duration `mapstructure:"pool_timeout"`
    IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
    IdleCheckFrequency time.Duration `mapstructure:"idle_check_frequency"`
}
```

### Cache Configuration

```go
type CacheConfig struct {
    UserProfileTTL         time.Duration `mapstructure:"user_profile_ttl"`
    PhotoMetadataTTL       time.Duration `mapstructure:"photo_metadata_ttl"`
    MatchRecommendationsTTL time.Duration `mapstructure:"match_recommendations_ttl"`
    APIResponseTTL         time.Duration `mapstructure:"api_response_ttl"`
    GeospatialDataTTL      time.Duration `mapstructure:"geospatial_data_ttl"`
}
```

### Session Configuration

```go
type SessionConfig struct {
    SessionTTL           time.Duration `mapstructure:"session_ttl"`
    RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
    CleanupInterval       time.Duration `mapstructure:"cleanup_interval"`
    MaxSessionsPerUser   int           `mapstructure:"max_sessions_per_user"`
}
```

## Usage Examples

### Session Management

```go
// Create a new session
session, err := sessionManager.CreateSession(
    ctx,
    "user123",
    "access_token",
    "refresh_token",
    "192.168.1.1",
    "Mozilla/5.0...",
    DeviceInfo{
        DeviceID:   "device123",
        DeviceType: "mobile",
        OS:         "iOS",
        Browser:    "Safari",
        Location:   "New York",
    },
)

// Get session
session, err := sessionManager.GetSession(ctx, "session123")

// Update session activity
err := sessionManager.UpdateSessionActivity(ctx, "session123")

// Delete session
err := sessionManager.DeleteSession(ctx, "session123")

// Check if user is online
online, err := sessionManager.IsUserOnline(ctx, "user123")

// Get online users
onlineUsers, err := sessionManager.GetOnlineUsers(ctx)
```

### Caching

```go
// Cache user profile
err := cacheService.CacheUserProfile(ctx, "user123", userProfile)

// Get user profile
var profile UserProfile
found, err := cacheService.GetUserProfile(ctx, "user123", &profile)

// Invalidate user profile
err := cacheService.InvalidateUserProfile(ctx, "user123")

// Cache API response
err := cacheService.CacheAPIResponse(ctx, "api_key", response)

// Get API response
var response APIResponse
found, err := cacheService.GetAPIResponse(ctx, "api_key", &response)
```

### Rate Limiting

```go
// Check rate limit
config := RateLimitConfig{
    Requests: 100,
    Window:   time.Minute,
    Endpoint: "/api/users",
    KeyType:  "ip",
}

result, err := rateLimiter.CheckRateLimit(ctx, config, "192.168.1.1")

if !result.Allowed {
    // Rate limit exceeded
    return errors.New("rate limit exceeded")
}

// Get rate limit status
status, err := rateLimiter.GetRateLimitStatus(ctx, config, "192.168.1.1")
```

### Token Blacklist

```go
// Add token to blacklist
err := tokenBlacklist.AddToken(ctx, "token123", time.Hour)

// Check if token is blacklisted
blacklisted, err := tokenBlacklist.IsTokenBlacklisted(ctx, "token123")

// Remove token from blacklist
err := tokenBlacklist.RemoveToken(ctx, "token123")
```

### Pub/Sub

```go
// Subscribe to chat messages
msgChan := pubSubService.SubscribeToChatMessages(ctx, "user123")

// Publish chat message
err := pubSubService.PublishChatMessage(ctx, "user1", "user2", "Hello!")

// Subscribe to notifications
notifChan := pubSubService.SubscribeToNotifications(ctx, "user123")

// Publish notification
err := pubSubService.PublishNotification(ctx, "user123", "match", "New Match!", data)
```

## Health Checks

The implementation provides comprehensive health checks:

### Endpoints

- `/health` - Overall health status
- `/health/redis` - Redis connection health
- `/health/cache` - Cache service health
- `/health/sessions` - Session manager health
- `/health/rate_limit` - Rate limiter health
- `/health/pubsub` - Pub/Sub service health
- `/health/liveness` - Kubernetes liveness probe
- `/health/readiness` - Kubernetes readiness probe
- `/health/metrics` - Detailed metrics

### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2025-01-01T00:00:00Z",
  "components": {
    "redis": {
      "status": "healthy",
      "ping": "ok",
      "stats": {...},
      "metrics": {...}
    },
    "cache": {
      "status": "healthy",
      "stats": {...}
    },
    "sessions": {
      "status": "healthy",
      "online_users": 150
    },
    "rate_limit": {
      "status": "healthy",
      "standard_test": {...},
      "distributed_test": {...}
    },
    "pubsub": {
      "status": "healthy",
      "stats": {...}
    }
  }
}
```

## Monitoring and Metrics

### Redis Metrics

- Connection pool statistics
- Command execution counts
- Error rates
- Response times

### Cache Metrics

- Hit/miss ratios
- Eviction counts
- Memory usage
- Key counts

### Session Metrics

- Active sessions
- Online users
- Session creation/deletion rates

### Rate Limiting Metrics

- Request counts
- Rejection rates
- Window utilization

## Testing

The implementation includes comprehensive test suites:

- `connection_test.go` - Redis connection tests
- `session_manager_test.go` - Session management tests
- `cache_service_test.go` - Cache service tests
- `rate_limiter_test.go` - Rate limiting tests
- `pubsub_service_test.go` - Pub/Sub service tests
- `utils_test.go` - Cache utilities tests

### Running Tests

```bash
# Run all Redis tests
go test ./internal/infrastructure/database/redis/...
go test ./internal/infrastructure/cache/...
go test ./pkg/cache/...

# Run specific test suite
go test ./internal/infrastructure/cache -run TestSessionManagerTestSuite
```

## Best Practices

### Connection Management

1. Use connection pooling with appropriate pool sizes
2. Set proper timeouts for all operations
3. Implement health checks and automatic reconnection
4. Monitor connection pool metrics

### Caching Strategy

1. Use appropriate TTL values for different data types
2. Implement cache invalidation strategies
3. Use cache warming for frequently accessed data
4. Monitor cache hit/miss ratios

### Session Management

1. Limit sessions per user
2. Implement proper session expiration
3. Track session activity
4. Provide logout from all devices

### Rate Limiting

1. Use sliding window for accurate rate limiting
2. Implement distributed rate limiting for scalability
3. Use different limits for different endpoints
4. Monitor rate limit violations

### Pub/Sub

1. Use proper channel naming conventions
2. Implement subscription management
3. Handle connection failures gracefully
4. Use message acknowledgments where appropriate

## Performance Considerations

### Redis Optimization

1. Use Redis clusters for high availability
2. Optimize data structures for your use case
3. Use pipelining for batch operations
4. Monitor memory usage and eviction policies

### Caching Optimization

1. Cache frequently accessed data
2. Use appropriate data structures
3. Implement cache warming strategies
4. Monitor cache performance

### Session Optimization

1. Use efficient session storage
2. Implement session cleanup
3. Optimize session queries
4. Monitor session performance

## Security Considerations

1. Use Redis authentication
2. Enable TLS for Redis connections
3. Implement proper access controls
4. Use secure session storage
5. Implement token blacklisting
6. Validate all inputs

## Troubleshooting

### Common Issues

1. **Connection Timeouts**: Check network connectivity and Redis server status
2. **Memory Issues**: Monitor Redis memory usage and eviction policies
3. **Performance Issues**: Check slow queries and optimize data structures
4. **Session Issues**: Verify session configuration and cleanup processes

### Debugging

1. Enable Redis logging
2. Monitor health check endpoints
3. Check metrics and performance data
4. Use Redis CLI for debugging

## Deployment

### Docker Configuration

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    environment:
      - REDIS_PASSWORD=your_password
```

### Kubernetes Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
```

## Future Enhancements

1. **Redis Streams**: Implement event streaming
2. **Redis Modules**: Add RedisJSON, RedisSearch
3. **Multi-Region**: Implement cross-region replication
4. **Advanced Analytics**: Implement real-time analytics
5. **Machine Learning**: Add ML-based caching strategies