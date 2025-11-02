# Ephemeral Photos Documentation

## Overview

Ephemeral photos are temporary images that automatically expire after being viewed or after a specified time period. This feature enhances privacy and security in the dating application by allowing users to share photos that have limited lifetimes.

## Features

- **Automatic Expiration**: Photos expire after viewing or time-based expiration
- **Secure Access**: Photos are protected by unique access keys
- **View Tracking**: Track when and how many times photos are viewed
- **Chat Integration**: Send ephemeral photos directly in conversations
- **Watermarking**: Optional watermarking for additional security
- **Download Prevention**: Configurable download restrictions
- **Analytics**: Comprehensive view and engagement analytics
- **Background Cleanup**: Automatic cleanup of expired photos

## Architecture

### Core Components

1. **Ephemeral Photo Entity**: Core data model with expiration and view tracking
2. **Service Layer**: Business logic for photo lifecycle management
3. **Storage Service**: Secure temporary storage with automatic cleanup
4. **Cache Service**: Redis-based caching for performance
5. **Background Jobs**: Automated cleanup and maintenance
6. **Chat Integration**: Seamless integration with messaging system
7. **Security Layer**: Access control and protection mechanisms

### Data Flow

```
Upload → Validation → Storage → Cache → Chat Integration → View Tracking → Expiration → Cleanup
```

## API Endpoints

### Core Ephemeral Photo Operations

#### Upload Ephemeral Photo
```
POST /api/v1/ephemeral-photos
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

**Request Parameters:**
- `file` (required): Photo file (max 5MB, JPEG/PNG/WebP)
- `caption` (optional): Photo caption (max 500 chars)
- `duration_seconds` (optional): Custom duration (10-300s, default: 30s)
- `max_views` (optional): Maximum views (1-10, default: 1)

**Response:**
```json
{
  "success": true,
  "message": "Ephemeral photo uploaded successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "access_key": "a1b2c3d4e5f6789012345678901234567890ab",
    "thumbnail_url": "https://cdn.winkr.com/ephemeral/thumbnails/...",
    "expires_at": "2025-12-01T12:00:00Z",
    "max_views": 1,
    "duration_seconds": 30,
    "view_count": 0,
    "is_viewed": false,
    "is_expired": false
  }
}
```

#### View Ephemeral Photo
```
GET /api/v1/ephemeral-photos/{id}/view?access_key={key}
Authorization: Bearer {token}
```

**Parameters:**
- `id` (path): Photo ID
- `access_key` (query): 32-character access key
- `download` (query, optional): Set to true to download (may be restricted)

**Response:**
```json
{
  "success": true,
  "message": "Photo retrieved successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "url": "https://cdn.winkr.com/ephemeral/photos/...",
    "thumbnail_url": "https://cdn.winkr.com/ephemeral/thumbnails/...",
    "expires_at": "2025-12-01T12:00:00Z",
    "view_count": 1,
    "max_views": 1,
    "is_viewed": true,
    "is_expired": false,
    "caption": "Check out this photo!",
    "created_at": "2025-12-01T11:30:00Z"
  }
}
```

#### Delete Ephemeral Photo
```
DELETE /api/v1/ephemeral-photos/{id}
Authorization: Bearer {token}
```

#### Get Photo Status
```
GET /api/v1/ephemeral-photos/{id}/status
Authorization: Bearer {token}
```

#### Manually Expire Photo
```
POST /api/v1/ephemeral-photos/{id}/expire
Authorization: Bearer {token}
```

#### Get User's Photos
```
GET /api/v1/ephemeral-photos/my?status={status}&limit={limit}&offset={offset}
Authorization: Bearer {token}
```

### Chat Integration

#### Send Ephemeral Photo in Chat
```
POST /api/v1/chats/{conversationId}/ephemeral-photos
Authorization: Bearer {token}
```

**Request Body:**
```json
{
  "photo_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Check this out!"
}
```

#### Get Ephemeral Photo Message Details
```
GET /api/v1/chats/{conversationId}/messages/{messageId}/ephemeral-photo
Authorization: Bearer {token}
```

## Security Features

### Access Control
- **Unique Access Keys**: 32-character randomly generated keys
- **One-Time URLs**: URLs expire after first use
- **Time-Based Expiration**: Automatic expiration after specified duration
- **View Limits**: Configurable maximum view counts

### Protection Mechanisms
- **Watermarking**: Optional watermark overlay on photos
- **Download Prevention**: Configurable download restrictions
- **Screenshot Detection**: Client-side detection (implementation dependent)
- **Secure Storage**: Encrypted storage with limited access

### Rate Limiting
- **Upload Limits**: 10 uploads per hour per user
- **View Limits**: 50 views per hour per user
- **API Rate Limits**: Standard API rate limiting applies

## Configuration

### Environment Variables

```bash
# Photo Settings
EPHEMERAL_PHOTO_MAX_FILE_SIZE=5242880
EPHEMERAL_PHOTO_ALLOWED_TYPES=image/jpeg,image/png,image/webp
EPHEMERAL_PHOTO_MAX_PHOTOS_PER_USER=10

# Expiration Settings
EPHEMERAL_PHOTO_DEFAULT_DURATION=30s
EPHEMERAL_PHOTO_MAX_DURATION=300s
EPHEMERAL_PHOTO_VIEW_DURATION=30s

# Security Settings
EPHEMERAL_PHOTO_ACCESS_KEY_LENGTH=32
EPHEMERAL_PHOTO_ENABLE_WATERMARK=true
EPHEMERAL_PHOTO_WATERMARK_TEXT=Ephemeral
EPHEMERAL_PHOTO_PREVENT_DOWNLOAD=true

# Storage Settings
EPHEMERAL_PHOTO_STORAGE_TIER=hot
EPHEMERAL_PHOTO_CLEANUP_INTERVAL=5m
EPHEMERAL_PHOTO_RETENTION_PERIOD=24h

# Caching Settings
EPHEMERAL_PHOTO_CACHE_TTL=1m
EPHEMERAL_PHOTO_VIEW_CACHE_TTL=30s

# Rate Limiting
EPHEMERAL_PHOTO_UPLOAD_RATE_LIMIT=10
EPHEMERAL_PHOTO_VIEW_RATE_LIMIT=50

# Analytics Settings
EPHEMERAL_PHOTO_ENABLE_ANALYTICS=true
EPHEMERAL_PHOTO_ANALYTICS_TTL=168h

# Background Jobs
EPHEMERAL_PHOTO_JOB_INTERVAL=1m
EPHEMERAL_PHOTO_JOB_BATCH_SIZE=100
EPHEMERAL_PHOTO_ENABLE_JOB_RETRY=true
EPHEMERAL_PHOTO_MAX_JOB_RETRIES=3
```

## WebSocket Events

### Ephemeral Photo Events

#### New Ephemeral Photo Message
```json
{
  "type": "ephemeral_photo:new",
  "data": {
    "conversation_id": "550e8400-e29b-41d4-a716-446655440002",
    "photo_id": "550e8400-e29b-41d4-a716-446655440000",
    "access_key": "a1b2c3d4e5f6789012345678901234567890ab",
    "thumbnail_url": "https://cdn.winkr.com/ephemeral/thumbnails/...",
    "expires_at": "2025-12-01T12:00:00Z",
    "message": "Check this out!",
    "sender_id": "550e8400-e29b-41d4-a716-446655440003"
  },
  "timestamp": "2025-12-01T11:30:00Z",
  "sender_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

#### Photo Viewed Event
```json
{
  "type": "ephemeral_photo:viewed",
  "data": {
    "photo_id": "550e8400-e29b-41d4-a716-446655440000",
    "viewer_id": "550e8400-e29b-41d4-a716-446655440004",
    "viewed_at": "2025-12-01T11:35:00Z"
  },
  "timestamp": "2025-12-01T11:35:00Z",
  "sender_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

#### Photo Expired Event
```json
{
  "type": "ephemeral_photo:expired",
  "data": {
    "photo_id": "550e8400-e29b-41d4-a716-446655440000",
    "owner_id": "550e8400-e29b-41d4-a716-446655440003",
    "expired_at": "2025-12-01T12:00:00Z"
  },
  "timestamp": "2025-12-01T12:00:00Z",
  "sender_id": "system"
}
```

## Background Jobs

### Cleanup Jobs

1. **Expired Photo Cleanup**: Removes photos past their expiration time
2. **Viewed Photo Cleanup**: Removes photos that have been viewed
3. **Orphaned File Cleanup**: Removes storage files without database records
4. **Analytics Processing**: Processes view analytics and engagement data
5. **Cache Cleanup**: Clears expired cache entries

### Job Configuration

- **Interval**: Runs every minute
- **Batch Size**: Processes 100 items per batch
- **Retry Logic**: Up to 3 retries on failure
- **Monitoring**: Health checks and error tracking

## Analytics and Monitoring

### Tracked Metrics

- **View Count**: Number of times each photo is viewed
- **View Duration**: How long photos are viewed
- **Engagement Rate**: Views per unique viewers
- **Expiration Rate**: Photos expired vs viewed
- **Upload Success Rate**: Successful uploads vs attempts
- **Storage Usage**: Current and historical storage usage

### Performance Monitoring

- **API Response Times**: Endpoint performance metrics
- **Cache Hit Rates**: Redis cache effectiveness
- **Background Job Health**: Job success/failure rates
- **Storage Performance**: Upload/download speeds
- **Error Rates**: Error tracking and alerting

## Best Practices

### For Developers

1. **Validate Access Keys**: Always validate access keys before granting access
2. **Check Expiration**: Verify photo hasn't expired before processing
3. **Handle Race Conditions**: Use proper locking for view tracking
4. **Cache Strategically**: Cache metadata but not photo content
5. **Log Security Events**: Track access attempts and failures

### For Users

1. **Secure Sharing**: Only share access keys with intended recipients
2. **Monitor Expiration**: Be aware of photo expiration times
3. **Respect Privacy**: Don't screenshot or redistribute photos
4. **Report Issues**: Report any security concerns immediately

## Error Handling

### Common Error Codes

- `EPHEMERAL_PHOTO_NOT_FOUND`: Photo doesn't exist
- `EPHEMERAL_PHOTO_EXPIRED`: Photo has expired
- `EPHEMERAL_PHOTO_ALREADY_VIEWED`: Photo already viewed (if max_views=1)
- `INVALID_ACCESS_KEY`: Access key is invalid or expired
- `RATE_LIMIT_EXCEEDED`: User has exceeded rate limits
- `FILE_TOO_LARGE`: Photo exceeds size limits
- `UNSUPPORTED_FILE_TYPE`: File type not allowed

### Error Response Format

```json
{
  "success": false,
  "error": "EPHEMERAL_PHOTO_EXPIRED",
  "message": "The ephemeral photo has expired",
  "code": "EPHEMERAL_PHOTO_EXPIRED"
}
```

## Testing

### Integration Tests

Comprehensive test suite covering:
- Upload and view flow
- Automatic expiration
- Security controls
- Chat integration
- Background cleanup jobs
- Rate limiting
- Error scenarios

### Test Scenarios

1. **Happy Path**: Successful upload, view, and expiration
2. **Security Tests**: Invalid access keys, unauthorized access
3. **Expiration Tests**: Time-based and view-based expiration
4. **Integration Tests**: Chat message sending and receiving
5. **Performance Tests**: Load testing and stress testing
6. **Edge Cases**: Network failures, concurrent access

## Deployment Considerations

### Storage Requirements

- **Hot Storage**: Fast access for active photos
- **Cold Storage**: Cost-effective for expired photos
- **CDN Integration**: Global content delivery
- **Backup Strategy**: Redundant storage for reliability

### Scaling Considerations

- **Horizontal Scaling**: Multiple service instances
- **Database Sharding**: Distribute photo metadata
- **Cache Clustering**: Redis cluster for high availability
- **Load Balancing**: Distribute API requests
- **Monitoring**: Comprehensive observability

### Security Considerations

- **Encryption**: At-rest and in-transit encryption
- **Access Controls**: Role-based access control
- **Audit Logging**: Comprehensive access logging
- **Compliance**: Data protection regulations
- **Vulnerability Management**: Regular security updates

## Troubleshooting

### Common Issues

1. **Photos Not Expiring**: Check background job status
2. **Access Key Invalid**: Verify key format and expiration
3. **Slow Upload Performance**: Check storage configuration
4. **High Memory Usage**: Review cache configuration
5. **Database Performance**: Check query optimization

### Debugging Tools

- **Application Logs**: Detailed error and access logs
- **Metrics Dashboard**: Real-time performance metrics
- **Health Endpoints**: Service health status
- **Database Queries**: Slow query analysis
- **Cache Analysis**: Redis performance metrics

## Future Enhancements

### Planned Features

1. **Advanced Watermarking**: Custom watermark designs
2. **View Duration Limits**: Time-based view limits
3. **Geographic Restrictions**: Location-based access controls
4. **AI Content Moderation**: Automated content filtering
5. **Enhanced Analytics**: More detailed engagement metrics
6. **Mobile Optimizations**: Better mobile experience

### Technical Improvements

1. **Machine Learning**: Predictive analytics for usage patterns
2. **Edge Computing**: CDN edge processing
3. **Blockchain Integration**: Immutable photo metadata
4. **Advanced Encryption**: Zero-knowledge encryption
5. **Performance Optimization**: Further performance improvements

## Support

For questions, issues, or feature requests related to ephemeral photos:

- **Documentation**: Check this documentation first
- **API Reference**: See OpenAPI specification
- **Issue Tracker**: Report bugs and feature requests
- **Community Forum**: Discuss with other developers
- **Support Team**: Contact for urgent issues

## Changelog

### Version 1.0.0
- Initial release of ephemeral photos feature
- Core upload, view, and expiration functionality
- Chat integration
- Background cleanup jobs
- Comprehensive API documentation
- Security and rate limiting features