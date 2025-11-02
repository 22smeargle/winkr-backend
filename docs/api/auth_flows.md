# Authentication API Flows

This document describes the authentication flows and best practices for using the Winkr Authentication API.

## Overview

The Winkr Authentication API provides secure user authentication with support for email and phone number registration, device fingerprinting, session management, and comprehensive security features.

## Authentication Flows

### 1. User Registration Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Email/SMS Service

    Client->>API: POST /auth/register
    API->>API: Validate input data
    API->>Database: Check if user exists
    Database-->>API: User not found
    API->>API: Hash password
    API->>Database: Create user
    Database-->>API: User created
    API->>API: Generate tokens
    API->>Database: Store session
    API->>Email/SMS Service: Send verification code
    Email/SMS Service-->>API: Code sent
    API-->>Client: Registration successful
```

#### Request Example

```bash
curl -X POST "https://api.winkr.com/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "username": "johndoe",
    "date_of_birth": "1990-01-01",
    "gender": "male",
    "device_info": {
      "device_id": "unique-device-id",
      "device_type": "mobile",
      "os": "iOS 15.0",
      "app_version": "1.0.0"
    }
  }'
```

#### Response Example

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "username": "johndoe",
      "verification_status": "pending",
      "created_at": "2025-01-01T00:00:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_in": 3600
    },
    "device": {
      "id": "device-uuid",
      "is_new": true
    }
  }
}
```

### 2. Email/Phone Verification Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Email/SMS Service

    Client->>API: POST /auth/send-verification
    API->>Database: Find user
    Database-->>API: User found
    API->>API: Generate 6-digit code
    API->>Database: Store code with expiration
    API->>Email/SMS Service: Send verification code
    Email/SMS Service-->>API: Code sent
    API-->>Client: Code sent successfully

    Client->>API: POST /auth/verify
    API->>Database: Find user and code
    Database-->>API: Code found
    API->>API: Validate code
    API->>Database: Update verification status
    Database-->>API: Status updated
    API-->>Client: Verification successful
```

#### Send Verification Code

```bash
curl -X POST "https://api.winkr.com/v1/auth/send-verification" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

#### Verify Code

```bash
curl -X POST "https://api.winkr.com/v1/auth/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'
```

### 3. Login Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Security Service

    Client->>API: POST /auth/login
    API->>API: Validate input data
    API->>Database: Find user
    Database-->>API: User found
    API->>Security Service: Check rate limiting
    Security Service-->>API: Rate limit OK
    API->>API: Verify password
    API->>Database: Check device
    Database-->>API: Device info
    API->>API: Generate tokens
    API->>Database: Store session
    API-->>Client: Login successful
```

#### Request Example

```bash
curl -X POST "https://api.winkr.com/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "device_info": {
      "device_id": "unique-device-id",
      "device_type": "mobile",
      "os": "iOS 15.0",
      "app_version": "1.0.0"
    }
  }'
```

### 4. Token Refresh Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Token Service

    Client->>API: POST /auth/refresh
    API->>Token Service: Validate refresh token
    Token Service-->>API: Token valid
    API->>Database: Find session
    Database-->>API: Session found
    API->>API: Generate new tokens
    API->>Database: Update session
    API-->>Client: Tokens refreshed
```

#### Request Example

```bash
curl -X POST "https://api.winkr.com/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### 5. Logout Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Token Service

    Client->>API: POST /auth/logout
    API->>Token Service: Validate access token
    Token Service-->>API: Token valid
    API->>Database: Find session
    Database-->>API: Session found
    API->>Token Service: Blacklist tokens
    Token Service-->>API: Tokens blacklisted
    API->>Database: Remove session
    Database-->>API: Session removed
    API-->>Client: Logout successful
```

#### Request Example

```bash
curl -X POST "https://api.winkr.com/v1/auth/logout" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "device_id": "device-uuid"
  }'
```

### 6. Password Reset Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Database
    participant Email/SMS Service

    Client->>API: POST /auth/password-reset
    API->>Database: Find user
    Database-->>API: User found
    API->>API: Generate reset code
    API->>Database: Store code with expiration
    API->>Email/SMS Service: Send reset code
    Email/SMS Service-->>API: Code sent
    API-->>Client: Reset code sent

    Client->>API: POST /auth/password-reset/confirm
    API->>Database: Find user and code
    Database-->>API: Code found
    API->>API: Validate code
    API->>API: Hash new password
    API->>Database: Update password
    Database-->>API: Password updated
    API-->>Client: Password reset successful
```

#### Request Password Reset

```bash
curl -X POST "https://api.winkr.com/v1/auth/password-reset" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

#### Confirm Password Reset

```bash
curl -X POST "https://api.winkr.com/v1/auth/password-reset/confirm" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456",
    "new_password": "NewSecurePass123!"
  }'
```

## Security Features

### 1. Rate Limiting

All authentication endpoints are protected by rate limiting:

- **Registration**: 5 attempts per hour per IP
- **Login**: 10 attempts per hour per IP
- **Password Reset**: 3 attempts per hour per email/phone
- **Verification Code**: 5 attempts per hour per email/phone
- **Token Refresh**: 20 attempts per hour per token

### 2. Account Lockout

After 5 failed login attempts within 15 minutes, the account is locked for 30 minutes. The lockout duration increases with subsequent failed attempts.

### 3. Device Fingerprinting

Each login attempt includes device information to detect suspicious activities:

- New device detection
- Multiple devices from same IP
- Rapid device switching
- Unusual user agent patterns

### 4. Token Security

- **Access Tokens**: 1 hour expiration
- **Refresh Tokens**: 30 days expiration
- **Token Rotation**: New refresh tokens issued on each refresh
- **Token Blacklisting**: Invalidated tokens are blacklisted

### 5. Password Security

- Minimum 8 characters
- Must contain uppercase, lowercase, number, and special character
- Checked against common password lists
- Hashed using bcrypt with cost factor 12

## Error Handling

### Common Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `VALIDATION_ERROR` | 422 | Input validation failed |
| `UNAUTHORIZED` | 401 | Invalid credentials or token |
| `FORBIDDEN` | 403 | Account locked or suspended |
| `NOT_FOUND` | 404 | User not found |
| `CONFLICT` | 409 | User already exists |
| `TOO_MANY_REQUESTS` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error |

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

## Best Practices

### 1. Client Implementation

- Store tokens securely (Keychain on iOS, Keystore on Android)
- Use HTTPS for all API calls
- Implement automatic token refresh
- Handle token expiration gracefully
- Validate user input before sending to API

### 2. Security Considerations

- Never log or store passwords in plain text
- Implement certificate pinning for mobile apps
- Use device-specific secrets for additional security
- Monitor for suspicious activities
- Implement proper logout on app uninstall

### 3. User Experience

- Provide clear error messages
- Show password strength indicators
- Implement progressive authentication
- Support biometric authentication where available
- Provide multiple verification options

## Testing

### Integration Tests

The API includes comprehensive integration tests covering:

- Registration flow with email verification
- Login flow with device tracking
- Token refresh flow
- Logout flow with session cleanup
- Password reset flow
- Rate limiting and security features

### Test Environment

Use the test environment for development and testing:

```bash
# Test server
https://staging-api.winkr.com/v1

# Test credentials (provided by development team)
```

## Support

For API support and questions:

- Email: api@winkr.com
- Documentation: https://docs.winkr.com
- Status Page: https://status.winkr.com