# WinKr API - Authentication Guide

## Overview

The WinKr API uses a comprehensive authentication system designed to secure user data while providing a seamless experience. This guide covers all aspects of authentication including JWT tokens, refresh mechanisms, role-based access, and security best practices.

## Table of Contents

1. [Authentication Methods](#authentication-methods)
2. [JWT Token Structure](#jwt-token-structure)
3. [Token Management](#token-management)
4. [Role-Based Access](#role-based-access)
5. [Security Features](#security-features)
6. [Authentication Flows](#authentication-flows)
7. [Error Handling](#error-handling)
8. [Best Practices](#best-practices)
9. [Code Examples](#code-examples)
10. [Troubleshooting](#troubleshooting)

## Authentication Methods

### Bearer Token Authentication

All API endpoints (except public health checks) require a valid JWT token sent in the Authorization header:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### API Key Authentication

For server-to-server integration and webhooks:

```http
X-API-Key: winkr_live_1234567890abcdef
X-API-Secret: your_api_secret_here
```

### Admin Authentication

Admin endpoints use a separate authentication flow with enhanced security:

```http
Authorization: Bearer admin_jwt_token_here
X-Admin-Role: super_admin|admin|moderator
```

## JWT Token Structure

### Access Token

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "2025-01-key-id"
  },
  "payload": {
    "sub": "550e8400-e29b-41d4-a716-446655440000",
    "iss": "https://api.winkr.com",
    "aud": "winkr-web-client",
    "exp": 1672531200,
    "iat": 1672527600,
    "jti": "token_123456789",
    "scope": ["read:profile", "write:profile", "read:messages"],
    "role": "user",
    "permissions": ["profile.read", "profile.write", "messages.read"],
    "session_id": "sess_abc123",
    "device_id": "device_xyz789"
  }
}
```

### Refresh Token

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "550e8400-e29b-41d4-a716-446655440000",
    "iss": "https://api.winkr.com",
    "aud": "winkr-refresh-service",
    "exp": 1675119600,
    "iat": 1672527600,
    "jti": "refresh_123456789",
    "type": "refresh",
    "session_id": "sess_abc123"
  }
}
```

### Token Claims

| Claim | Description | Example |
|-------|-------------|---------|
| `sub` | User ID (subject) | `"550e8400-e29b-41d4-a716-446655440000"` |
| `iss` | Token issuer | `"https://api.winkr.com"` |
| `aud` | Token audience | `"winkr-web-client"` |
| `exp` | Expiration timestamp | `1672531200` |
| `iat` | Issued at timestamp | `1672527600` |
| `jti` | Token ID (unique identifier) | `"token_123456789"` |
| `scope` | Permission scopes | `["read:profile", "write:profile"]` |
| `role` | User role | `"user"` |
| `permissions` | Granular permissions | `["profile.read", "messages.send"]` |

## Token Management

### Token Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Active: Login Successful
    Active --> Expired: Token Expires
    Active --> Revoked: Logout/Revoke
    Expired --> Refresh: Valid Refresh Token
    Expired --> Login: No Valid Refresh Token
    Refresh --> Active: Refresh Successful
    Revoked --> [*]: Token Invalidated
    Login --> Active: Login Successful
```

### Token Storage

#### Client-Side Storage (Web)

```javascript
// Secure token storage
const TokenManager = {
  setTokens(accessToken, refreshToken) {
    // Use httpOnly cookies for security
    document.cookie = `access_token=${accessToken}; path=/; secure; httpOnly; samesite=strict; max-age=3600`;
    document.cookie = `refresh_token=${refreshToken}; path=/; secure; httpOnly; samesite=strict; max-age=2592000`;
    
    // Fallback to sessionStorage for development
    if (process.env.NODE_ENV === 'development') {
      sessionStorage.setItem('access_token', accessToken);
      sessionStorage.setItem('refresh_token', refreshToken);
    }
  },

  getAccessToken() {
    // Try cookie first
    const cookieToken = this.getCookie('access_token');
    if (cookieToken) return cookieToken;
    
    // Fallback to sessionStorage
    return sessionStorage.getItem('access_token');
  },

  getRefreshToken() {
    return this.getCookie('refresh_token') || sessionStorage.getItem('refresh_token');
  },

  clearTokens() {
    document.cookie = 'access_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    document.cookie = 'refresh_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
    sessionStorage.removeItem('access_token');
    sessionStorage.removeItem('refresh_token');
  }
};
```

#### Mobile App Storage

```swift
// iOS - Keychain Storage
import Security

class TokenManager {
  static let accessTokenKey = "winkr_access_token"
  static let refreshTokenKey = "winkr_refresh_token"
  
  static func saveTokens(accessToken: String, refreshToken: String) {
    let accessTokenData = accessToken.data(using: .utf8)!
    let refreshTokenData = refreshToken.data(using: .utf8)!
    
    let keychain = Keychain(service: "winkr")
    keychain.save(accessTokenData, forKey: accessTokenKey)
    keychain.save(refreshTokenData, forKey: refreshTokenKey)
  }
  
  static func getAccessToken() -> String? {
    let keychain = Keychain(service: "winkr")
    return keychain.read(forKey: accessTokenKey)
  }
}
```

```kotlin
// Android - Encrypted SharedPreferences
import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences

class TokenManager(context: Context) {
  private val encryptedPrefs = EncryptedSharedPreferences.create(
    "winkr_tokens",
    MasterKey.Builder(context).setKeyScheme(MasterKey.KeyScheme.AES256_GCM).build()
  )
  
  fun saveTokens(accessToken: String, refreshToken: String) {
    encryptedPrefs.edit()
      .putString("access_token", accessToken)
      .putString("refresh_token", refreshToken)
      .apply()
  }
  
  fun getAccessToken(): String? {
    return encryptedPrefs.getString("access_token", null)
  }
}
```

### Token Refresh

#### Automatic Refresh

```javascript
// Axios interceptor for automatic token refresh
const axios = require('axios');

const apiClient = axios.create({
  baseURL: 'https://api.winkr.com/v1'
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      try {
        // Attempt to refresh token
        const refreshToken = TokenManager.getRefreshToken();
        const response = await axios.post('/auth/refresh', {
          refresh_token: refreshToken
        });
        
        const newToken = response.data.token;
        TokenManager.setTokens(newToken, refreshToken);
        
        // Retry original request with new token
        originalRequest._retry = true;
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed, redirect to login
        TokenManager.clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);
```

#### Manual Refresh

```javascript
const refreshAccessToken = async () => {
  const refreshToken = TokenManager.getRefreshToken();
  
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }
  
  try {
    const response = await fetch('/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        refresh_token: refreshToken
      })
    });
    
    if (!response.ok) {
      throw new Error('Token refresh failed');
    }
    
    const data = await response.json();
    TokenManager.setTokens(data.token, refreshToken);
    
    return data.token;
  } catch (error) {
    TokenManager.clearTokens();
    throw error;
  }
};
```

## Role-Based Access

### User Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| `user` | Standard user access | Profile management, messaging, matching |
| `verified_user` | Verified user access | All user permissions + verification features |
| `premium_user` | Premium subscriber | All user permissions + premium features |
| `moderator` | Content moderator | User permissions + moderation tools |
| `admin` | System administrator | All permissions + administrative tools |
| `super_admin` | Super administrator | Full system access and configuration |

### Permission Scopes

| Scope | Description | Required Role |
|-------|-------------|---------------|
| `read:profile` | Read user profile information | user |
| `write:profile` | Update user profile | user |
| `read:messages` | Read messages | user |
| `write:messages` | Send messages | user |
| `read:matches` | View match information | user |
| `write:matches` | Create/accept matches | user |
| `upload:photos` | Upload photos | user |
| `moderate:content` | Moderate content | moderator |
| `manage:users` | Manage user accounts | admin |
| `view:analytics` | View analytics | admin |
| `manage:system` | System configuration | super_admin |

### Permission Check

```javascript
// Permission checking utility
const hasPermission = (userRole, requiredPermission) => {
  const rolePermissions = {
    user: ['read:profile', 'write:profile', 'read:messages', 'write:messages', 'read:matches', 'upload:photos'],
    verified_user: ['read:profile', 'write:profile', 'read:messages', 'write:messages', 'read:matches', 'upload:photos', 'verify:profile'],
    premium_user: ['read:profile', 'write:profile', 'read:messages', 'write:messages', 'read:matches', 'upload:photos', 'premium:features'],
    moderator: ['read:profile', 'write:profile', 'read:messages', 'write:messages', 'read:matches', 'upload:photos', 'moderate:content'],
    admin: ['read:profile', 'write:profile', 'read:messages', 'write:messages', 'read:matches', 'upload:photos', 'moderate:content', 'manage:users', 'view:analytics'],
    super_admin: ['*'] // All permissions
  };
  
  const permissions = rolePermissions[userRole] || [];
  return permissions.includes('*') || permissions.includes(requiredPermission);
};

// Usage
if (hasPermission(user.role, 'moderate:content')) {
  // Show moderation features
}
```

## Security Features

### Multi-Factor Authentication (MFA)

```javascript
// MFA setup
const setupMFA = async (userId, token) => {
  const response = await fetch('/auth/mfa/setup', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      method: 'totp', // Time-based One-Time Password
      backup_codes: true
    })
  });
  
  return response.json(); // Contains QR code and backup codes
};

// MFA verification
const verifyMFA = async (userId, code, token) => {
  const response = await fetch('/auth/mfa/verify', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      code: code,
      user_id: userId
    })
  });
  
  return response.json();
};
```

### Device Fingerprinting

```javascript
// Device fingerprint for additional security
const generateDeviceFingerprint = () => {
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');
  ctx.textBaseline = 'top';
  ctx.font = '14px Arial';
  ctx.fillText('Device fingerprint', 2, 2);
  
  const fingerprint = {
    userAgent: navigator.userAgent,
    language: navigator.language,
    platform: navigator.platform,
    screen: `${screen.width}x${screen.height}`,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    canvas: canvas.toDataURL(),
    webgl: getWebGLFingerprint(),
    fonts: detectFonts()
  };
  
  return btoa(JSON.stringify(fingerprint));
};

// Include in login request
const loginWithFingerprint = async (email, password) => {
  const deviceFingerprint = generateDeviceFingerprint();
  
  return fetch('/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Device-Fingerprint': deviceFingerprint
    },
    body: JSON.stringify({
      email,
      password,
      device_info: {
        fingerprint: deviceFingerprint,
        trusted: false // Initially false, becomes true after verification
      }
    })
  });
};
```

### IP-Based Restrictions

```javascript
// IP validation for admin access
const validateAdminAccess = async (adminToken, clientIP) => {
  try {
    const response = await fetch('/admin/auth/validate-ip', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${adminToken}`,
        'Content-Type': 'application/json',
        'X-Client-IP': clientIP
      }
    });
    
    const data = await response.json();
    return data.allowed; // true if IP is whitelisted
  } catch (error) {
    console.error('IP validation failed:', error);
    return false;
  }
};
```

## Authentication Flows

### Standard User Login

```mermaid
sequenceDiagram
    participant User as User
    participant Client as Client App
    participant API as WinKr API
    participant Auth as Auth Service
    
    User->>Client: Enter credentials
    Client->>API: POST /auth/login
    API->>Auth: Validate credentials
    Auth-->>API: User data + permissions
    API-->>Client: JWT tokens + user info
    Client->>Client: Store tokens securely
    Client-->>User: Login successful
```

### Social Login Flow

```mermaid
sequenceDiagram
    participant User as User
    participant Client as Client App
    participant Social as Social Provider
    participant API as WinKr API
    
    User->>Client: Click "Login with Google"
    Client->>Social: OAuth flow
    Social-->>Client: Authorization code
    Client->>API: POST /auth/social/google
    API->>Social: Exchange code for user info
    Social-->>API: User profile data
    API-->>Client: JWT tokens + user info
    Client->>Client: Create/update user account
    Client-->>User: Login successful
```

### Password Reset Flow

```mermaid
sequenceDiagram
    participant User as User
    participant Client as Client App
    participant API as WinKr API
    participant Email as Email Service
    
    User->>Client: Request password reset
    Client->>API: POST /auth/forgot-password
    API->>Email: Send reset email
    Email-->>User: Reset email with token
    User->>Client: Enter reset token + new password
    Client->>API: POST /auth/reset-password
    API-->>Client: Password reset successful
    Client-->>User: Password updated
```

### Admin Authentication Flow

```mermaid
sequenceDiagram
    participant Admin as Admin User
    participant Client as Admin Panel
    participant API as WinKr API
    participant MFA as MFA Service
    
    Admin->>Client: Enter admin credentials
    Client->>API: POST /admin/auth/login
    API-->>Client: Require MFA
    Client->>Admin: Enter MFA code
    Client->>API: POST /admin/auth/verify-mfa
    API->>MFA: Verify MFA code
    MFA-->>API: MFA valid
    API-->>Client: Admin JWT + permissions
    Client->>Client: Store admin token
    Client-->>Admin: Admin access granted
```

## Error Handling

### Authentication Error Codes

| Error Code | HTTP Status | Description | Solution |
|------------|---------------|-------------|----------|
| `INVALID_CREDENTIALS` | 401 | Email or password incorrect | Check credentials and retry |
| `ACCOUNT_LOCKED` | 423 | Account temporarily locked | Wait for lockout period |
| `MFA_REQUIRED` | 401 | Multi-factor authentication required | Complete MFA flow |
| `INVALID_MFA_CODE` | 401 | MFA code invalid | Retry with correct code |
| `TOKEN_EXPIRED` | 401 | JWT token has expired | Refresh token |
| `INVALID_TOKEN` | 401 | JWT token is invalid | Re-authenticate |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required permissions | Check user role |
| `DEVICE_NOT_TRUSTED` | 403 | Device not recognized | Complete device verification |
| `IP_BLOCKED` | 403 | IP address blocked | Contact support |
| `SESSION_EXPIRED` | 401 | Session has expired | Re-authenticate |

### Error Handling Implementation

```javascript
// Comprehensive authentication error handler
class AuthErrorHandler {
  static handle(error, context = {}) {
    const { response } = error;
    
    if (!response) {
      // Network error
      this.showNetworkError();
      return;
    }
    
    const { status, data } = response;
    const errorCode = data?.error?.code;
    
    switch (errorCode) {
      case 'INVALID_CREDENTIALS':
        this.showCredentialError();
        break;
        
      case 'ACCOUNT_LOCKED':
        this.showAccountLockedError(data?.error?.details?.unlock_time);
        break;
        
      case 'MFA_REQUIRED':
        this.redirectToMFA();
        break;
        
      case 'TOKEN_EXPIRED':
        this.refreshToken().catch(() => this.redirectToLogin());
        break;
        
      case 'INSUFFICIENT_PERMISSIONS':
        this.showPermissionError();
        break;
        
      case 'DEVICE_NOT_TRUSTED':
        this.showDeviceVerification();
        break;
        
      default:
        this.showGenericError(data?.error?.message);
    }
  }
  
  static showCredentialError() {
    // Show user-friendly error message
    // Offer password reset option
  }
  
  static showAccountLockedError(unlockTime) {
    // Show lockout duration
    // Offer contact support option
  }
  
  static redirectToMFA() {
    // Redirect to MFA verification page
  }
  
  static async refreshToken() {
    try {
      const refreshToken = TokenManager.getRefreshToken();
      const response = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken })
      });
      
      const data = await response.json();
      TokenManager.setTokens(data.token, refreshToken);
      
      // Retry original request
      return true;
    } catch (error) {
      return false;
    }
  }
}
```

## Best Practices

### Security Implementation

1. **Never store tokens in localStorage** for production apps
2. **Use httpOnly cookies** for web applications
3. **Implement secure token storage** on mobile (Keychain/EncryptedSharedPreferences)
4. **Always use HTTPS** for authentication requests
5. **Validate tokens server-side** - never trust client-side claims
6. **Implement proper logout** that invalidates tokens server-side
7. **Use short-lived access tokens** (1 hour or less)
8. **Implement token refresh** automatically and transparently

### User Experience

1. **Provide clear error messages** with actionable guidance
2. **Implement remember me functionality** securely
3. **Offer multiple login methods** (email, social, SSO)
4. **Show password strength indicators** during registration
5. **Implement account lockout** after failed attempts
6. **Provide password reset** functionality
7. **Show session timeout warnings** before expiration
8. **Support multiple devices** with device management

### Development Practices

1. **Never hardcode credentials** in source code
2. **Use environment variables** for API keys
3. **Implement proper logging** without sensitive data
4. **Validate all inputs** before processing
5. **Use parameterized queries** to prevent injection
6. **Implement rate limiting** on authentication endpoints
7. **Monitor authentication events** for security
8. **Keep dependencies updated** for security patches

## Code Examples

### Complete Authentication Service

```javascript
class AuthenticationService {
  constructor() {
    this.baseURL = 'https://api.winkr.com/v1';
    this.tokenManager = new TokenManager();
  }

  async login(email, password, rememberMe = false) {
    try {
      const response = await fetch(`${this.baseURL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Device-Fingerprint': this.generateDeviceFingerprint()
        },
        body: JSON.stringify({
          email,
          password,
          remember_me: rememberMe,
          device_info: {
            user_agent: navigator.userAgent,
            platform: navigator.platform,
            screen: `${screen.width}x${screen.height}`
          }
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new AuthenticationError(error.error?.code, error.error?.message);
      }

      const data = await response.json();
      
      // Store tokens securely
      this.tokenManager.setTokens(data.token, data.refresh_token);
      
      // Set up automatic refresh
      this.setupTokenRefresh();
      
      return {
        success: true,
        user: data.user,
        permissions: data.permissions
      };
    } catch (error) {
      return {
        success: false,
        error: error.message
      };
    }
  }

  async logout() {
    try {
      const token = this.tokenManager.getAccessToken();
      await fetch(`${this.baseURL}/auth/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      // Always clear local tokens
      this.tokenManager.clearTokens();
    }
  }

  setupTokenRefresh() {
    // Set up token refresh 5 minutes before expiration
    const token = this.tokenManager.getAccessToken();
    const payload = this.parseJWT(token);
    
    if (payload && payload.exp) {
      const refreshTime = (payload.exp * 1000) - (5 * 60 * 1000); // 5 minutes before expiry
      const delay = refreshTime - Date.now();
      
      if (delay > 0) {
        setTimeout(() => {
          this.refreshToken();
        }, delay);
      }
    }
  }

  async refreshToken() {
    try {
      const refreshToken = this.tokenManager.getRefreshToken();
      const response = await fetch(`${this.baseURL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          refresh_token: refreshToken
        })
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      this.tokenManager.setTokens(data.token, data.refresh_token);
      
      // Set up next refresh
      this.setupTokenRefresh();
      
      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      // Clear tokens and redirect to login
      this.tokenManager.clearTokens();
      window.location.href = '/login';
      return false;
    }
  }

  generateDeviceFingerprint() {
    // Generate unique device fingerprint
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.textBaseline = 'top';
    ctx.font = '14px Arial';
    ctx.fillText('Device fingerprint', 2, 2);
    
    return btoa(canvas.toDataURL());
  }

  parseJWT(token) {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(atob(base64));
      return JSON.parse(jsonPayload);
    } catch (error) {
      return null;
    }
  }

  hasPermission(permission) {
    const token = this.tokenManager.getAccessToken();
    const payload = this.parseJWT(token);
    return payload?.permissions?.includes(permission) || false;
  }
}

// Custom error class
class AuthenticationError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'AuthenticationError';
    this.code = code;
  }
}
```

## Troubleshooting

### Common Issues

#### Token Not Working

**Symptoms:**
- Getting 401 Unauthorized errors
- Token appears valid but is rejected

**Solutions:**
1. Check token expiration time
2. Verify token format (Bearer + space + token)
3. Ensure token is not URL encoded
4. Check for extra whitespace in header
5. Verify token is sent in correct header

#### Refresh Token Failing

**Symptoms:**
- Unable to refresh access token
- Getting invalid refresh token errors

**Solutions:**
1. Check if refresh token has expired
2. Verify refresh token is stored correctly
3. Ensure refresh token is not being reused
4. Check network connectivity
5. Verify refresh endpoint URL is correct

#### CORS Issues

**Symptoms:**
- Browser blocks API requests
- CORS errors in console

**Solutions:**
1. Ensure API base URL is correct
2. Check that requests use HTTPS
3. Verify preflight OPTIONS requests are handled
4. Check that appropriate CORS headers are returned
5. Ensure credentials are included in requests

#### Rate Limiting

**Symptoms:**
- Getting 429 Too Many Requests errors
- Requests being blocked

**Solutions:**
1. Implement exponential backoff
2. Check rate limit headers in response
3. Reduce request frequency
4. Use caching to reduce requests
5. Implement proper retry logic

### Debug Tools

#### Token Inspector

```javascript
// Browser console utility for debugging tokens
const tokenInspector = {
  inspect() {
    const token = TokenManager.getAccessToken();
    if (!token) {
      console.log('No access token found');
      return;
    }
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      console.log('Token Payload:', payload);
      console.log('Expires:', new Date(payload.exp * 1000));
      console.log('Issued:', new Date(payload.iat * 1000));
      console.log('Permissions:', payload.permissions);
      console.log('Role:', payload.role);
    } catch (error) {
      console.error('Invalid token format:', error);
    }
  },
  
  validate() {
    const token = TokenManager.getAccessToken();
    if (!token) return false;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const now = Math.floor(Date.now() / 1000);
      
      return payload.exp > now;
    } catch (error) {
      return false;
    }
  }
};

// Usage in browser console
// tokenInspector.inspect();
// tokenInspector.validate();
```

### Support Resources

- **Authentication Documentation**: [Complete Guide](./getting-started.md)
- **API Reference**: [All Endpoints](./README.md)
- **Error Codes**: [Error Reference](./errors.md)
- **Support Email**: auth-support@winkr.com
- **Developer Forum**: [Authentication Discussions](https://community.winkr.com/auth)

---

For additional authentication support or questions, contact our development team at api-support@winkr.com.