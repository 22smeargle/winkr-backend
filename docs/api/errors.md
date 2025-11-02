# API Error Reference

This comprehensive guide documents all possible errors that can occur when using the Winkr API, including error codes, messages, causes, and recommended solutions.

## Table of Contents

- [Error Response Format](#error-response-format)
- [Error Categories](#error-categories)
- [Authentication Errors](#authentication-errors)
- [Authorization Errors](#authorization-errors)
- [Validation Errors](#validation-errors)
- [Not Found Errors](#not-found-errors)
- [Conflict Errors](#conflict-errors)
- [Rate Limiting Errors](#rate-limiting-errors)
- [Payment Errors](#payment-errors)
- [File Upload Errors](#file-upload-errors)
- [Server Errors](#server-errors)
- [WebSocket Errors](#websocket-errors)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Error Handling Best Practices](#error-handling-best-practices)

## Error Response Format

All API errors follow a consistent JSON format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional error details",
      "reason": "Specific reason for the error"
    },
    "request_id": "req_123456789",
    "timestamp": "2025-12-01T10:30:00Z"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Always `false` for error responses |
| `error.code` | string | Machine-readable error code |
| `error.message` | string | Human-readable error description |
| `error.details` | object | Additional error context (optional) |
| `error.request_id` | string | Unique request identifier for debugging |
| `error.timestamp` | string | ISO 8601 timestamp of the error |

### HTTP Status Codes

| Status Code | Category | Description |
|-------------|----------|-------------|
| 400 | Client Error | Bad request, validation errors |
| 401 | Authentication | Invalid or missing authentication |
| 403 | Authorization | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 422 | Validation | Unprocessable entity |
| 429 | Rate Limiting | Too many requests |
| 500 | Server Error | Internal server error |
| 502 | Server Error | Bad gateway |
| 503 | Server Error | Service unavailable |
| 504 | Server Error | Gateway timeout |

## Error Categories

### 1. Authentication Errors (401)

These errors occur when authentication fails or tokens are invalid/expired.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `TOKEN_MISSING` | Authentication token is required | No token provided in request | Include valid JWT token in `Authorization` header |
| `TOKEN_INVALID` | Invalid authentication token | Token is malformed or tampered | Obtain new token from authentication endpoint |
| `TOKEN_EXPIRED` | Authentication token has expired | Token has passed its expiration time | Refresh token using refresh endpoint |
| `TOKEN_REVOKED` | Authentication token has been revoked | Token was revoked by user or admin | Obtain new token from authentication endpoint |
| `REFRESH_TOKEN_EXPIRED` | Refresh token has expired | Refresh token is no longer valid | Require user to re-authenticate |
| `REFRESH_TOKEN_INVALID` | Invalid refresh token | Refresh token is malformed or invalid | Require user to re-authenticate |
| `CREDENTIALS_INVALID` | Invalid email or password | Incorrect login credentials | Verify credentials and try again |
| `ACCOUNT_LOCKED` | Account is temporarily locked | Too many failed login attempts | Wait for lockout period or contact support |
| `ACCOUNT_SUSPENDED` | Account has been suspended | Account suspended by admin | Contact support for account status |
| `EMAIL_NOT_VERIFIED` | Email address not verified | User hasn't verified email address | Complete email verification process |

### 2. Authorization Errors (403)

These errors occur when the authenticated user doesn't have permission to perform the requested action.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `INSUFFICIENT_PERMISSIONS` | Insufficient permissions to access resource | User lacks required role/permission | Request appropriate permissions from admin |
| `RESOURCE_ACCESS_DENIED` | Access to resource denied | User doesn't have access to specific resource | Verify resource ownership or permissions |
| `PREMIUM_FEATURE_REQUIRED` | Premium subscription required | Feature requires premium subscription | Upgrade to premium subscription |
| `ADMIN_ACCESS_REQUIRED` | Admin access required | Endpoint requires admin privileges | Use admin account or contact admin |
| `SELF_ACTION_REQUIRED` | Action must be performed by user themselves | Operation can only be done by resource owner | User must perform action themselves |
| `VERIFICATION_REQUIRED` | Account verification required | User needs to complete verification | Complete account verification process |
| `AGE_RESTRICTION` | Age restriction applies | User doesn't meet age requirements | Verify user meets minimum age requirements |

### 3. Validation Errors (400/422)

These errors occur when request data is invalid or malformed.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `INVALID_REQUEST_FORMAT` | Invalid request format | JSON is malformed or invalid | Fix JSON syntax and structure |
| `MISSING_REQUIRED_FIELD` | Required field is missing | Required field not provided in request | Include all required fields |
| `INVALID_FIELD_VALUE` | Invalid field value | Field value doesn't meet validation rules | Provide valid field value |
| `FIELD_TOO_LONG` | Field value exceeds maximum length | Text field exceeds character limit | Shorten field value within limits |
| `FIELD_TOO_SHORT` | Field value below minimum length | Text field below minimum characters | Provide longer field value |
| `INVALID_EMAIL_FORMAT` | Invalid email format | Email address is malformed | Provide valid email address |
| `INVALID_PASSWORD_FORMAT` | Password doesn't meet requirements | Password doesn't meet complexity rules | Use password that meets requirements |
| `INVALID_DATE_FORMAT` | Invalid date format | Date not in expected format | Use ISO 8601 date format |
| `INVALID_COORDINATES` | Invalid geographic coordinates | Lat/lng values out of valid range | Provide valid coordinates (-90 to 90, -180 to 180) |
| `INVALID_PHONE_FORMAT` | Invalid phone number format | Phone number doesn't match expected format | Use valid international phone format |
| `INVALID_URL_FORMAT` | Invalid URL format | URL is malformed | Provide valid URL format |
| `INVALID_UUID_FORMAT` | Invalid UUID format | UUID doesn't match expected pattern | Provide valid UUID string |
| `INVALID_ENUM_VALUE` | Invalid enum value | Value not in allowed enum values | Use one of the allowed enum values |

### 4. Not Found Errors (404)

These errors occur when requested resources don't exist.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `USER_NOT_FOUND` | User not found | User ID doesn't exist | Verify user ID is correct |
| `CONVERSATION_NOT_FOUND` | Conversation not found | Conversation ID doesn't exist | Verify conversation ID is correct |
| `MESSAGE_NOT_FOUND` | Message not found | Message ID doesn't exist | Verify message ID is correct |
| `PHOTO_NOT_FOUND` | Photo not found | Photo ID doesn't exist | Verify photo ID is correct |
| `MATCH_NOT_FOUND` | Match not found | Match ID doesn't exist | Verify match ID is correct |
| `SUBSCRIPTION_NOT_FOUND` | Subscription not found | User has no active subscription | Check subscription status |
| `VERIFICATION_NOT_FOUND` | Verification not found | Verification record doesn't exist | Check verification status |
| `REPORT_NOT_FOUND` | Report not found | Report ID doesn't exist | Verify report ID is correct |
| `PAYMENT_NOT_FOUND` | Payment not found | Payment ID doesn't exist | Verify payment ID is correct |
| `WEBHOOK_NOT_FOUND` | Webhook not found | Webhook ID doesn't exist | Verify webhook ID is correct |

### 5. Conflict Errors (409)

These errors occur when there's a conflict with the current state of the resource.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `EMAIL_ALREADY_EXISTS` | Email address already exists | Another user has this email | Use different email address |
| `USERNAME_ALREADY_EXISTS` | Username already exists | Another user has this username | Choose different username |
| `PHONE_ALREADY_EXISTS` | Phone number already exists | Another user has this phone | Use different phone number |
| `MATCH_ALREADY_EXISTS` | Match already exists | Users are already matched | Check existing match status |
| `CONVERSATION_ALREADY_EXISTS` | Conversation already exists | Conversation already exists between users | Use existing conversation |
| `PHOTO_ALREADY_UPLOADED` | Photo already uploaded | Same photo already uploaded | Use different photo |
| `VERIFICATION_ALREADY_COMPLETED` | Verification already completed | User already verified | Check verification status |
| `SUBSCRIPTION_ALREADY_ACTIVE` | Subscription already active | User already has active subscription | Check subscription status |
| `DUPLICATE_ACTION` | Duplicate action detected | Same action performed recently | Wait before retrying action |
| `RESOURCE_LOCKED` | Resource is currently locked | Resource being modified by another process | Wait and retry later |

### 6. Rate Limiting Errors (429)

These errors occur when request limits are exceeded.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded | Too many requests in time window | Implement exponential backoff |
| `DAILY_LIMIT_EXCEEDED` | Daily request limit exceeded | Daily quota exceeded | Wait for daily reset |
| `MONTHLY_LIMIT_EXCEEDED` | Monthly request limit exceeded | Monthly quota exceeded | Upgrade plan or wait for reset |
| `ENDPOINT_RATE_LIMITED` | Endpoint-specific rate limit exceeded | Too many requests to specific endpoint | Reduce request frequency |
| `IP_RATE_LIMITED` | IP address rate limited | Too many requests from IP address | Reduce request frequency |
| `USER_RATE_LIMITED` | User rate limited | Too many requests from user | Reduce request frequency |
| `UPLOAD_RATE_LIMITED` | Upload rate limited | Too many file uploads | Reduce upload frequency |
| `MESSAGE_RATE_LIMITED` | Message rate limited | Too many messages sent | Reduce message frequency |
| `SWIPE_RATE_LIMITED` | Swipe rate limited | Too many swipes performed | Wait before swiping again |

### 7. Payment Errors (402/422)

These errors occur during payment processing.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `PAYMENT_REQUIRED` | Payment required | Feature requires payment | Complete payment process |
| `PAYMENT_METHOD_INVALID` | Invalid payment method | Payment method is invalid or expired | Update payment method |
| `PAYMENT_DECLINED` | Payment declined | Payment declined by processor | Use different payment method |
| `INSUFFICIENT_FUNDS` | Insufficient funds | Not enough funds for payment | Add funds or use different method |
| `SUBSCRIPTION_EXPIRED` | Subscription has expired | Subscription period ended | Renew subscription |
| `SUBSCRIPTION_CANCELLED` | Subscription has been cancelled | Subscription was cancelled | Reactivate subscription |
| `PAYMENT_PROCESSING_ERROR` | Payment processing error | Error during payment processing | Retry payment or contact support |
| `REFUND_FAILED` | Refund processing failed | Error processing refund | Contact support |
| `INVOICE_NOT_FOUND` | Invoice not found | Invoice ID doesn't exist | Verify invoice ID |
| `BILLING_ADDRESS_INVALID` | Invalid billing address | Billing address doesn't match | Update billing address |

### 8. File Upload Errors (400/413/422)

These errors occur during file uploads.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `FILE_TOO_LARGE` | File size exceeds limit | File larger than allowed size | Compress or use smaller file |
| `FILE_TYPE_NOT_SUPPORTED` | File type not supported | Unsupported file format | Use supported file type |
| `FILE_CORRUPTED` | File appears to be corrupted | File is damaged or invalid | Use valid file |
| `UPLOAD_FAILED` | File upload failed | Error during upload process | Retry upload |
| `STORAGE_QUOTA_EXCEEDED` | Storage quota exceeded | User has exceeded storage limit | Delete files or upgrade plan |
| `IMAGE_PROCESSING_FAILED` | Image processing failed | Error processing image | Use valid image format |
| `VIRUS_DETECTED` | Virus detected in file | Uploaded file contains virus | Use clean file |
| `DUPLICATE_FILE` | Duplicate file detected | Same file already uploaded | Use different file |
| `UPLOAD_TIMEOUT` | Upload timeout | Upload took too long | Check connection and retry |
| `INVALID_IMAGE_DIMENSIONS` | Invalid image dimensions | Image dimensions outside allowed range | Resize image to valid dimensions |

### 9. Server Errors (500/502/503/504)

These errors occur when the server encounters an unexpected error.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `INTERNAL_SERVER_ERROR` | Internal server error | Unexpected server error | Retry request or contact support |
| `DATABASE_ERROR` | Database error occurred | Database operation failed | Retry request or contact support |
| `EXTERNAL_SERVICE_ERROR` | External service error | Third-party service unavailable | Retry later or contact support |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable | Service under maintenance | Wait and retry later |
| `GATEWAY_TIMEOUT` | Gateway timeout | Request took too long to process | Retry with shorter timeout |
| `BAD_GATEWAY` | Bad gateway | Invalid response from upstream service | Retry or contact support |
| `MEMORY_LIMIT_EXCEEDED` | Memory limit exceeded | Server ran out of memory | Retry with smaller request |
| `DISK_SPACE_EXCEEDED` | Disk space exceeded | Server ran out of disk space | Contact support |
| `CONFIGURATION_ERROR` | Configuration error | Server misconfiguration | Contact support |
| `DEPENDENCY_FAILURE` | Service dependency failure | Required service unavailable | Retry later or contact support |

### 10. WebSocket Errors

These errors occur during WebSocket connections.

| Error Code | Message | Cause | Solution |
|------------|---------|-------|----------|
| `WEBSOCKET_CONNECTION_FAILED` | WebSocket connection failed | Unable to establish WebSocket connection | Check network and retry |
| `WEBSOCKET_AUTHENTICATION_FAILED` | WebSocket authentication failed | Invalid or expired token | Refresh token and reconnect |
| `WEBSOCKET_RATE_LIMITED` | WebSocket rate limited | Too many WebSocket messages | Reduce message frequency |
| `WEBSOCKET_MESSAGE_TOO_LARGE` | WebSocket message too large | Message exceeds size limit | Reduce message size |
| `WEBSOCKET_INVALID_FORMAT` | Invalid WebSocket message format | Message doesn't match expected format | Use correct message format |
| `WEBSOCKET_ROOM_NOT_FOUND` | WebSocket room not found | Room doesn't exist or user not member | Verify room ID and membership |
| `WEBSOCKET_PERMISSION_DENIED` | WebSocket permission denied | No permission for room/action | Check permissions |
| `WEBSOCKET_CONNECTION_LOST` | WebSocket connection lost | Network interruption | Implement reconnection logic |
| `WEBSOCKET_SERVER_ERROR` | WebSocket server error | Server-side WebSocket error | Reconnect or contact support |

## Troubleshooting Guide

### Common Error Scenarios

#### 1. Authentication Issues

**Problem**: Receiving 401 errors despite providing valid credentials

**Symptoms**:
- `TOKEN_EXPIRED` errors
- `TOKEN_INVALID` errors
- Frequent re-authentication required

**Solutions**:
1. Check token expiration time
2. Implement automatic token refresh
3. Verify token storage is secure
4. Check system clock synchronization

```javascript
// Example: Token refresh implementation
async function refreshToken() {
  try {
    const response = await fetch('/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        refresh_token: getRefreshToken()
      })
    });
    
    if (response.ok) {
      const data = await response.json();
      setAccessToken(data.access_token);
      return data.access_token;
    } else {
      throw new Error('Token refresh failed');
    }
  } catch (error) {
    // Handle refresh failure
    logout();
    throw error;
  }
}
```

#### 2. Rate Limiting Issues

**Problem**: Hitting rate limits frequently

**Symptoms**:
- `RATE_LIMIT_EXCEEDED` errors
- `429` status codes
- Requests being throttled

**Solutions**:
1. Implement exponential backoff
2. Cache responses when possible
3. Optimize request frequency
4. Use batch operations

```javascript
// Example: Exponential backoff implementation
async function makeRequestWithBackoff(url, options = {}) {
  let attempt = 0;
  const maxAttempts = 5;
  const baseDelay = 1000;
  
  while (attempt < maxAttempts) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const retryAfter = parseInt(response.headers.get('Retry-After') || '60');
        const delay = Math.max(retryAfter * 1000, baseDelay * Math.pow(2, attempt));
        
        await new Promise(resolve => setTimeout(resolve, delay));
        attempt++;
        continue;
      }
      
      return response;
    } catch (error) {
      if (attempt === maxAttempts - 1) throw error;
      
      const delay = baseDelay * Math.pow(2, attempt);
      await new Promise(resolve => setTimeout(resolve, delay));
      attempt++;
    }
  }
}
```

#### 3. Validation Errors

**Problem**: Requests failing validation

**Symptoms**:
- `INVALID_FIELD_VALUE` errors
- `MISSING_REQUIRED_FIELD` errors
- `422` status codes

**Solutions**:
1. Validate data client-side before sending
2. Check API documentation for field requirements
3. Use proper data formats
4. Handle edge cases

```javascript
// Example: Client-side validation
function validateUserProfile(profile) {
  const errors = [];
  
  if (!profile.email || !isValidEmail(profile.email)) {
    errors.push('Valid email is required');
  }
  
  if (!profile.username || profile.username.length < 3) {
    errors.push('Username must be at least 3 characters');
  }
  
  if (profile.age && (profile.age < 18 || profile.age > 100)) {
    errors.push('Age must be between 18 and 100');
  }
  
  if (errors.length > 0) {
    throw new Error(errors.join(', '));
  }
  
  return true;
}

function isValidEmail(email) {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
```

#### 4. File Upload Issues

**Problem**: File uploads failing

**Symptoms**:
- `FILE_TOO_LARGE` errors
- `FILE_TYPE_NOT_SUPPORTED` errors
- `UPLOAD_FAILED` errors

**Solutions**:
1. Check file size limits
2. Verify supported file types
3. Implement progress tracking
4. Handle network interruptions

```javascript
// Example: File upload with validation
async function uploadFile(file) {
  // Validate file size (10MB limit)
  const maxSize = 10 * 1024 * 1024;
  if (file.size > maxSize) {
    throw new Error('File size exceeds 10MB limit');
  }
  
  // Validate file type
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
  if (!allowedTypes.includes(file.type)) {
    throw new Error('File type not supported');
  }
  
  const formData = new FormData();
  formData.append('file', file);
  
  try {
    const response = await fetch('/upload', {
      method: 'POST',
      body: formData
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error.message);
    }
    
    return await response.json();
  } catch (error) {
    console.error('Upload failed:', error);
    throw error;
  }
}
```

#### 5. WebSocket Connection Issues

**Problem**: WebSocket connections dropping frequently

**Symptoms**:
- `WEBSOCKET_CONNECTION_LOST` errors
- Frequent disconnections
- Messages not being delivered

**Solutions**:
1. Implement reconnection logic
2. Handle network interruptions
3. Monitor connection health
4. Use heartbeat messages

```javascript
// Example: Robust WebSocket implementation
class RobustWebSocket {
  constructor(url, options = {}) {
    this.url = url;
    this.options = {
      reconnect: true,
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      ...options
    };
    
    this.ws = null;
    this.reconnectAttempts = 0;
    this.heartbeatTimer = null;
    
    this.connect();
  }
  
  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.startHeartbeat();
    };
    
    this.ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      this.stopHeartbeat();
      
      if (this.options.reconnect && event.code !== 1000) {
        this.attemptReconnect();
      }
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }
  
  startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'heartbeat' }));
      }
    }, this.options.heartbeatInterval);
  }
  
  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
  
  attemptReconnect() {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }
    
    this.reconnectAttempts++;
    console.log(`Attempting reconnection (${this.reconnectAttempts}/${this.options.maxReconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, this.options.reconnectInterval);
  }
}
```

## Error Handling Best Practices

### 1. Client-Side Error Handling

```javascript
// Comprehensive error handling wrapper
class APIError extends Error {
  constructor(response, data) {
    super(data.error.message);
    this.name = 'APIError';
    this.code = data.error.code;
    this.status = response.status;
    this.requestId = data.error.request_id;
    this.details = data.error.details;
  }
}

async function apiRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAccessToken()}`,
        ...options.headers
      },
      ...options
    });
    
    const data = await response.json();
    
    if (!response.ok) {
      throw new APIError(response, data);
    }
    
    return data;
  } catch (error) {
    if (error instanceof APIError) {
      handleAPIError(error);
    } else {
      handleNetworkError(error);
    }
    throw error;
  }
}

function handleAPIError(error) {
  console.error('API Error:', error);
  
  switch (error.code) {
    case 'TOKEN_EXPIRED':
      refreshToken().then(() => {
        // Retry original request
      });
      break;
      
    case 'RATE_LIMIT_EXCEEDED':
      showRateLimitError(error);
      break;
      
    case 'VALIDATION_ERROR':
      showValidationErrors(error.details);
      break;
      
    default:
      showGenericError(error.message);
  }
}

function showRateLimitError(error) {
  const retryAfter = error.details?.retry_after || 60;
  showNotification(`Rate limit exceeded. Please try again in ${retryAfter} seconds.`, 'error');
}

function showValidationErrors(details) {
  const messages = Object.values(details).flat();
  showNotification(messages.join(', '), 'error');
}

function showGenericError(message) {
  showNotification(message, 'error');
}
```

### 2. Server-Side Error Handling

```go
// Go example: Structured error handling
type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
    Success   bool      `json:"success"`
    Error     APIError  `json:"error"`
    RequestID string    `json:"request_id"`
    Timestamp time.Time `json:"timestamp"`
}

func NewAPIError(code, message string, details map[string]interface{}) *APIError {
    return &APIError{
        Code:    code,
        Message: message,
        Details: details,
    }
}

func (e *APIError) Error() string {
    return e.Message
}

func WriteError(w http.ResponseWriter, err *APIError, statusCode int) {
    requestID := middleware.GetRequestID(r.Context())
    
    response := ErrorResponse{
        Success:   false,
        Error:     *err,
        RequestID: requestID,
        Timestamp: time.Now(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// Usage examples
func HandleUserProfile(w http.ResponseWriter, r *http.Request) {
    userID := mux.Vars(r)["id"]
    
    user, err := userService.GetUser(userID)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            apiErr := NewAPIError("USER_NOT_FOUND", "User not found", nil)
            WriteError(w, apiErr, http.StatusNotFound)
            return
        }
        
        apiErr := NewAPIError("INTERNAL_SERVER_ERROR", "Failed to get user", nil)
        WriteError(w, apiErr, http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    user,
    })
}
```

### 3. Error Monitoring and Logging

```javascript
// Client-side error monitoring
class ErrorMonitor {
  constructor() {
    this.errors = [];
    this.maxErrors = 100;
  }
  
  captureError(error, context = {}) {
    const errorData = {
      message: error.message,
      stack: error.stack,
      code: error.code,
      status: error.status,
      requestId: error.requestId,
      timestamp: new Date().toISOString(),
      context: context,
      userAgent: navigator.userAgent,
      url: window.location.href
    };
    
    this.errors.push(errorData);
    
    // Keep only recent errors
    if (this.errors.length > this.maxErrors) {
      this.errors.shift();
    }
    
    // Send to monitoring service
    this.sendToMonitoring(errorData);
  }
  
  sendToMonitoring(errorData) {
    // Send to error tracking service
    fetch('/api/errors', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(errorData)
    }).catch(err => {
      console.error('Failed to send error to monitoring:', err);
    });
  }
  
  getErrors() {
    return this.errors;
  }
  
  clearErrors() {
    this.errors = [];
  }
}

// Global error handler
const errorMonitor = new ErrorMonitor();

window.addEventListener('error', (event) => {
  errorMonitor.captureError(event.error, {
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno
  });
});

window.addEventListener('unhandledrejection', (event) => {
  errorMonitor.captureError(event.reason, {
    type: 'unhandled_promise_rejection'
  });
});
```

### 4. Error Recovery Strategies

```javascript
// Error recovery patterns
class ErrorRecovery {
  static async withRetry(fn, maxAttempts = 3, delay = 1000) {
    let lastError;
    
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await fn();
      } catch (error) {
        lastError = error;
        
        if (attempt === maxAttempts) {
          throw error;
        }
        
        // Don't retry on certain errors
        if (this.isNonRetryableError(error)) {
          throw error;
        }
        
        // Exponential backoff
        const backoffDelay = delay * Math.pow(2, attempt - 1);
        await new Promise(resolve => setTimeout(resolve, backoffDelay));
      }
    }
    
    throw lastError;
  }
  
  static isNonRetryableError(error) {
    const nonRetryableCodes = [
      'VALIDATION_ERROR',
      'AUTHENTICATION_ERROR',
      'AUTHORIZATION_ERROR',
      'NOT_FOUND_ERROR'
    ];
    
    return nonRetryableCodes.includes(error.code);
  }
  
  static async withFallback(primaryFn, fallbackFn) {
    try {
      return await primaryFn();
    } catch (error) {
      console.warn('Primary function failed, using fallback:', error);
      return await fallbackFn();
    }
  }
  
  static async withCircuitBreaker(fn, threshold = 5, timeout = 60000) {
    const state = {
      failures: 0,
      lastFailure: 0,
      state: 'CLOSED' // CLOSED, OPEN, HALF_OPEN
    };
    
    return async (...args) => {
      if (state.state === 'OPEN') {
        if (Date.now() - state.lastFailure > timeout) {
          state.state = 'HALF_OPEN';
        } else {
          throw new Error('Circuit breaker is OPEN');
        }
      }
      
      try {
        const result = await fn(...args);
        
        if (state.state === 'HALF_OPEN') {
          state.state = 'CLOSED';
          state.failures = 0;
        }
        
        return result;
      } catch (error) {
        state.failures++;
        state.lastFailure = Date.now();
        
        if (state.failures >= threshold) {
          state.state = 'OPEN';
        }
        
        throw error;
      }
    };
  }
}

// Usage examples
try {
  const data = await ErrorRecovery.withRetry(
    () => apiRequest('/api/data'),
    3,
    1000
  );
} catch (error) {
  console.error('All retry attempts failed:', error);
}

const cachedData = await ErrorRecovery.withFallback(
  () => apiRequest('/api/fresh-data'),
  () => getCachedData()
);
```

This comprehensive error reference provides detailed information about all possible API errors, their causes, solutions, and best practices for handling them effectively in your applications.