# Winkr API Documentation

Welcome to the comprehensive API documentation for the Winkr dating application backend. This documentation provides everything you need to integrate with our platform, from basic authentication to advanced features like real-time messaging and AI-powered verification.

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [API Versioning](#api-versioning)
- [Security](#security)
- [Integration Guides](#integration-guides)
- [Testing](#testing)
- [Performance](#performance)
- [Deployment](#deployment)

## Overview

The Winkr API is a RESTful API that provides access to all features of the Winkr dating platform. It includes:

- **User Authentication & Management** - Secure user registration, login, and profile management
- **Matching & Discovery** - Advanced matching algorithms with customizable preferences
- **Real-time Chat** - WebSocket-based messaging with typing indicators and read receipts
- **Photo Management** - Secure photo upload, storage, and management
- **Verification System** - AI-powered selfie and document verification
- **Payment Processing** - Stripe integration for subscriptions and payments
- **Moderation** - Content moderation and user reporting
- **Admin Panel** - Comprehensive admin tools for platform management

### Key Features

- **RESTful Design** - Clean, intuitive API endpoints following REST principles
- **JWT Authentication** - Secure token-based authentication with refresh tokens
- **Real-time Communication** - WebSocket support for instant messaging
- **AI-Powered Verification** - Automated user verification using AWS Rekognition
- **Comprehensive Rate Limiting** - Protection against abuse and API overuse
- **Detailed Logging** - Full audit trail for all operations
- **Scalable Architecture** - Designed for high availability and performance

## Getting Started

### Prerequisites

- A valid API key (contact api@winkr.com to request access)
- Basic understanding of REST APIs and JSON
- For real-time features: WebSocket client library

### Base URL

```
Production: https://api.winkr.com/v1
Staging: https://staging-api.winkr.com/v1
Development: http://localhost:8080/v1
```

### Making Your First Request

1. **Register a new user:**

```bash
curl -X POST "https://api.winkr.com/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "username": "johndoe",
    "date_of_birth": "1990-01-01",
    "gender": "male"
  }'
```

2. **Login to get access tokens:**

```bash
curl -X POST "https://api.winkr.com/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

3. **Make an authenticated request:**

```bash
curl -X GET "https://api.winkr.com/v1/profile/me" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Authentication

The Winkr API uses JWT (JSON Web Tokens) for authentication. All protected endpoints require a valid access token in the `Authorization` header.

### Authentication Flow

1. **Register** or **Login** to receive access and refresh tokens
2. **Include** the access token in the `Authorization` header for API requests
3. **Refresh** the access token using the refresh token before it expires
4. **Logout** to invalidate tokens and end the session

### Token Types

- **Access Token**: Short-lived (1 hour) token for API requests
- **Refresh Token**: Long-lived (30 days) token for obtaining new access tokens

### Example Authentication Header

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

For detailed authentication flows and examples, see [Authentication Documentation](auth_flows.md).

## API Endpoints

The Winkr API is organized into the following modules:

### Core Modules

| Module | Description | Documentation |
|--------|-------------|---------------|
| [Authentication](auth.yaml) | User registration, login, and session management | [Auth Flows](auth_flows.md) |
| [Profile](profile.yaml) | User profile management and settings | - |
| [Photos](photo.yaml) | Photo upload, management, and storage | - |
| [Discovery](discovery.yaml) | User discovery and matching algorithm | - |
| [Chat](chat.yaml) | Real-time messaging and conversations | [WebSocket Events](websocket_events.md) |
| [Verification](verification.yaml) | User verification with AI analysis | [Verification Flows](verification_flows.md) |

### Advanced Features

| Module | Description | Documentation |
|--------|-------------|---------------|
| [Ephemeral Photos](ephemeral_photo.yaml) | Self-destructing photo messages | - |
| [Moderation](moderation.yaml) | Content moderation and user reporting | - |
| [Payments](payment.yaml) | Subscription management and payments | [Payment Flows](payment_flows.md) |
| [Admin](admin.yaml) | Administrative tools and analytics | - |

### Quick Reference

#### Authentication Endpoints
- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - User logout
- `POST /auth/password-reset` - Request password reset

#### Profile Endpoints
- `GET /profile/me` - Get current user profile
- `PUT /profile/me` - Update user profile
- `GET /profile/{id}` - Get user profile by ID
- `POST /profile/preferences` - Update matching preferences

#### Discovery Endpoints
- `GET /discovery/users` - Get potential matches
- `POST /discovery/swipe` - Swipe on a user
- `GET /discovery/matches` - Get user matches
- `POST /discovery/boost` - Boost profile visibility

#### Chat Endpoints
- `GET /chat/conversations` - Get user conversations
- `POST /chat/conversations` - Start new conversation
- `GET /chat/conversations/{id}/messages` - Get conversation messages
- `POST /chat/conversations/{id}/messages` - Send message

## Rate Limiting

The Winkr API implements comprehensive rate limiting to ensure fair usage and prevent abuse.

### Rate Limit Tiers

| Endpoint Type | Limit | Duration |
|---------------|-------|----------|
| Authentication | 10 requests | per minute |
| Profile Updates | 20 requests | per minute |
| Photo Upload | 5 requests | per minute |
| Discovery/Swipes | 100 requests | per minute |
| Messaging | 30 messages | per minute |
| Admin Endpoints | 100 requests | per minute |

### Rate Limit Headers

All API responses include rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Rate Limit Response

When rate limits are exceeded:

```json
{
  "success": false,
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Rate limit exceeded",
    "details": {
      "retry_after": 60
    }
  }
}
```

## Error Handling

The Winkr API uses standard HTTP status codes and provides detailed error information in the response body.

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Access denied |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

For a complete list of error codes and troubleshooting, see [Error Reference](errors.md).

## API Versioning

The Winkr API uses URL-based versioning to ensure backward compatibility.

### Current Version

The current stable version is `v1`, included in all endpoint URLs:

```
https://api.winkr.com/v1/auth/login
```

### Versioning Strategy

- **Major versions** (`v1`, `v2`, etc.) indicate breaking changes
- **Minor updates** within a version are backward compatible
- **Deprecated endpoints** are announced 3 months before removal
- **Migration guides** are provided for version upgrades

### Version Headers

You can specify the API version using the `Accept` header:

```
Accept: application/vnd.winkr.v1+json
```

## Security

The Winkr API implements multiple layers of security to protect user data and prevent abuse.

### Security Features

- **HTTPS Only** - All API calls require HTTPS encryption
- **JWT Authentication** - Secure token-based authentication
- **Rate Limiting** - Protection against abuse and DDoS attacks
- **Input Validation** - Comprehensive input sanitization and validation
- **Content Filtering** - AI-powered content moderation
- **Audit Logging** - Complete audit trail for all operations
- **Data Encryption** - Sensitive data encrypted at rest and in transit

### Security Best Practices

1. **Never** expose API keys or tokens in client-side code
2. **Always** use HTTPS for API calls
3. **Implement** proper token storage and refresh mechanisms
4. **Validate** all user input before sending to the API
5. **Handle** errors gracefully without exposing sensitive information
6. **Monitor** for suspicious activity and implement proper logging

For detailed security guidelines, see [Security Documentation](security.md).

## Integration Guides

We provide comprehensive integration guides for different platforms:

- [Web Client Integration Guide](guides/web-integration.md)
- [Mobile App Integration Guide](guides/mobile-integration.md)
- [Third-Party Integration Guide](guides/third-party-integration.md)
- [WebSocket Integration Guide](guides/websocket-integration.md)

## Testing

### Test Environment

Use our staging environment for testing and development:

```
Base URL: https://staging-api.winkr.com/v1
```

### Testing Tools

- **Postman Collection** - Ready-to-use API collection
- **OpenAPI Specification** - Complete API specification
- **Test Data** - Sample data for testing
- **Mock Services** - Mock responses for development

For detailed testing instructions, see [API Testing Guide](testing.md).

## Performance

### Performance Expectations

- **API Response Time**: < 200ms (95th percentile)
- **WebSocket Latency**: < 50ms
- **Photo Upload**: < 5 seconds for 5MB files
- **AI Verification**: < 30 seconds

### Optimization Tips

1. **Use pagination** for large data sets
2. **Implement caching** for frequently accessed data
3. **Optimize image sizes** before upload
4. **Use WebSocket** for real-time features
5. **Batch operations** where possible

For performance guidelines and best practices, see [Performance Documentation](performance.md).

## Deployment

### Requirements

- **Node.js 18+** or **Go 1.19+** for server deployment
- **PostgreSQL 14+** for database
- **Redis 6+** for caching
- **AWS S3** for file storage
- **Stripe** for payment processing

### Environment Configuration

```yaml
# Database
DB_HOST: localhost
DB_PORT: 5432
DB_NAME: winkr
DB_USER: winkr_user
DB_PASSWORD: secure_password

# Redis
REDIS_HOST: localhost
REDIS_PORT: 6379
REDIS_PASSWORD: redis_password

# AWS S3
AWS_ACCESS_KEY_ID: your_access_key
AWS_SECRET_ACCESS_KEY: your_secret_key
AWS_REGION: us-east-1
S3_BUCKET: winkr-storage

# Stripe
STRIPE_SECRET_KEY: sk_test_...
STRIPE_WEBHOOK_SECRET: whsec_...

# JWT
JWT_SECRET: your_jwt_secret
JWT_EXPIRY: 1h
REFRESH_TOKEN_EXPIRY: 720h
```

For deployment instructions and scaling guidelines, see [Deployment Documentation](deployment.md).

## Support

### Getting Help

- **Documentation**: https://docs.winkr.com
- **API Status**: https://status.winkr.com
- **Support Email**: api@winkr.com
- **Developer Community**: https://community.winkr.com

### Reporting Issues

If you encounter any issues with the API:

1. Check the [API Status Page](https://status.winkr.com)
2. Review the [Error Reference](errors.md)
3. Search existing issues in our support system
4. Contact support with detailed information about the issue

### Feedback

We welcome feedback on our API and documentation. Please send your suggestions to api@winkr.com.

---

© 2025 Winkr. All rights reserved.