# Third-Party Integration Guide

This guide provides comprehensive instructions for integrating the Winkr API into third-party applications, services, and platforms.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Authentication](#authentication)
- [Integration Patterns](#integration-patterns)
- [Webhook Integration](#webhook-integration)
- [OAuth Integration](#oauth-integration)
- [API Client Libraries](#api-client-libraries)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Security Best Practices](#security-best-practices)
- [Examples](#examples)

## Overview

Third-party integration allows external services to connect with the Winkr platform, enabling features like:

- **User Authentication**: Authenticate Winkr users in your application
- **Profile Access**: Access user profile information (with consent)
- **Messaging**: Send and receive messages on behalf of users
- **Matching**: Integrate Winkr's matching algorithm
- **Analytics**: Access user engagement data
- **Webhooks**: Receive real-time notifications about user activities

### Use Cases

- **Social Media Platforms**: Connect Winkr profiles to other social networks
- **Dating Aggregators**: Aggregate profiles from multiple dating platforms
- **Analytics Services**: Analyze user behavior and engagement
- **Marketing Platforms**: Create targeted campaigns based on user data
- **Content Management**: Manage user-generated content across platforms

## Prerequisites

### Requirements

- **API Key**: Contact api@winkr.com to request third-party API access
- **HTTPS**: All API calls must be made over HTTPS
- **Compliance**: Adhere to GDPR, CCPA, and other privacy regulations
- **Terms of Service**: Accept Winkr's third-party integration terms

### Development Setup

- **Server Environment**: Secure server environment for API calls
- **Webhook Endpoint**: Publicly accessible endpoint for webhook events
- **SSL Certificate**: Valid SSL certificate for webhook endpoints
- **Rate Limiting**: Implement client-side rate limiting

## Authentication

### 1. API Key Authentication

Third-party integrations use API key authentication for server-to-server communication.

```bash
curl -X GET "https://api.winkr.com/v1/users/{user_id}" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-API-Version: 1.0"
```

### 2. OAuth 2.0 Flow

For user-authorized access, implement the OAuth 2.0 authorization code flow.

#### Step 1: Authorization Request

```javascript
const authUrl = 'https://api.winkr.com/v1/oauth/authorize?' + new URLSearchParams({
  response_type: 'code',
  client_id: 'YOUR_CLIENT_ID',
  redirect_uri: 'https://your-app.com/callback',
  scope: 'profile messages matches',
  state: generateRandomState()
});

// Redirect user to authUrl
window.location.href = authUrl;
```

#### Step 2: Handle Callback

```javascript
// Handle callback at your redirect URI
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');
const state = urlParams.get('state');

// Verify state matches
if (state !== storedState) {
  throw new Error('Invalid state parameter');
}

// Exchange code for access token
const response = await fetch('https://api.winkr.com/v1/oauth/token', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Basic ${btoa('YOUR_CLIENT_ID:YOUR_CLIENT_SECRET')}`
  },
  body: JSON.stringify({
    grant_type: 'authorization_code',
    code: code,
    redirect_uri: 'https://your-app.com/callback'
  })
});

const { access_token, refresh_token, expires_in } = await response.json();
```

#### Step 3: Use Access Token

```javascript
// Make API requests with user access token
const response = await fetch('https://api.winkr.com/v1/profile/me', {
  headers: {
    'Authorization': `Bearer ${access_token}`,
    'X-API-Version': '1.0'
  }
});

const userProfile = await response.json();
```

### 3. Token Management

```javascript
class WinkrAPIClient {
  constructor(clientId, clientSecret, apiKey) {
    this.clientId = clientId;
    this.clientSecret = clientSecret;
    this.apiKey = apiKey;
    this.accessToken = null;
    this.refreshToken = null;
    this.tokenExpiry = null;
  }

  async getAccessToken() {
    // Check if token is still valid
    if (this.accessToken && this.tokenExpiry > Date.now()) {
      return this.accessToken;
    }

    // Refresh token if available
    if (this.refreshToken) {
      return await this.refreshAccessToken();
    }

    throw new Error('No valid access token available');
  }

  async refreshAccessToken() {
    const response = await fetch('https://api.winkr.com/v1/oauth/token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Basic ${btoa(`${this.clientId}:${this.clientSecret}`)}`
      },
      body: JSON.stringify({
        grant_type: 'refresh_token',
        refresh_token: this.refreshToken
      })
    });

    if (!response.ok) {
      throw new Error('Token refresh failed');
    }

    const data = await response.json();
    this.accessToken = data.access_token;
    this.refreshToken = data.refresh_token;
    this.tokenExpiry = Date.now() + (data.expires_in * 1000);

    return this.accessToken;
  }

  async makeRequest(endpoint, options = {}) {
    const token = await this.getAccessToken();
    
    const response = await fetch(`https://api.winkr.com/v1${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${token}`,
        'X-API-Version': '1.0',
        ...options.headers
      }
    });

    if (response.status === 401) {
      // Try to refresh token and retry
      await this.refreshAccessToken();
      return this.makeRequest(endpoint, options);
    }

    return response;
  }
}
```

## Integration Patterns

### 1. Server-to-Server Integration

For backend services that need to access Winkr data without user interaction.

```python
# Python example using requests
import requests
import jwt
import time

class WinkrServerClient:
    def __init__(self, api_key, client_id, client_secret):
        self.api_key = api_key
        self.client_id = client_id
        self.client_secret = client_secret
        self.base_url = "https://api.winkr.com/v1"
        self.access_token = None
        self.token_expiry = None

    def get_access_token(self):
        if self.access_token and self.token_expiry > time.time():
            return self.access_token

        # Generate JWT for server-to-server auth
        payload = {
            'iss': self.client_id,
            'aud': 'winkr-api',
            'exp': time.time() + 3600,
            'iat': time.time()
        }
        
        jwt_token = jwt.encode(payload, self.client_secret, algorithm='HS256')
        
        response = requests.post(
            f"{self.base_url}/auth/server-token",
            headers={
                'Authorization': f'Bearer {jwt_token}',
                'X-API-Key': self.api_key
            }
        )
        
        if response.status_code == 200:
            data = response.json()
            self.access_token = data['access_token']
            self.token_expiry = time.time() + data['expires_in']
            return self.access_token
        
        raise Exception("Failed to obtain access token")

    def get_user_profile(self, user_id):
        token = self.get_access_token()
        response = requests.get(
            f"{self.base_url}/users/{user_id}",
            headers={
                'Authorization': f'Bearer {token}',
                'X-API-Key': self.api_key
            }
        )
        
        if response.status_code == 200:
            return response.json()
        
        raise Exception(f"API request failed: {response.status_code}")

    def send_message(self, user_id, conversation_id, content):
        token = self.get_access_token()
        response = requests.post(
            f"{self.base_url}/chat/conversations/{conversation_id}/messages",
            headers={
                'Authorization': f'Bearer {token}',
                'X-API-Key': self.api_key,
                'Content-Type': 'application/json'
            },
            json={
                'content': content,
                'type': 'text'
            }
        )
        
        if response.status_code == 201:
            return response.json()
        
        raise Exception(f"Failed to send message: {response.status_code}")
```

### 2. Client-Side Integration

For web applications that interact with Winkr on behalf of users.

```javascript
// JavaScript SDK for client-side integration
class WinkrSDK {
  constructor(config) {
    this.clientId = config.clientId;
    this.redirectUri = config.redirectUri;
    this.scope = config.scope || 'profile messages';
    this.accessToken = null;
    this.refreshToken = null;
  }

  async login() {
    const authUrl = this.buildAuthUrl();
    const popup = window.open(authUrl, 'winkr-auth', 'width=500,height=600');
    
    return new Promise((resolve, reject) => {
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
          reject(new Error('Authentication cancelled'));
        }
      }, 1000);

      const messageHandler = (event) => {
        if (event.origin !== 'https://api.winkr.com') return;
        
        if (event.data.type === 'winkr-auth-success') {
          clearInterval(checkClosed);
          window.removeEventListener('message', messageHandler);
          popup.close();
          
          this.accessToken = event.data.access_token;
          this.refreshToken = event.data.refresh_token;
          resolve(event.data);
        } else if (event.data.type === 'winkr-auth-error') {
          clearInterval(checkClosed);
          window.removeEventListener('message', messageHandler);
          popup.close();
          reject(new Error(event.data.error));
        }
      };

      window.addEventListener('message', messageHandler);
    });
  }

  buildAuthUrl() {
    const params = new URLSearchParams({
      response_type: 'code',
      client_id: this.clientId,
      redirect_uri: this.redirectUri,
      scope: this.scope,
      state: this.generateState()
    });

    return `https://api.winkr.com/v1/oauth/authorize?${params}`;
  }

  generateState() {
    return Math.random().toString(36).substring(2, 15);
  }

  async getProfile() {
    if (!this.accessToken) {
      throw new Error('Not authenticated');
    }

    const response = await fetch('https://api.winkr.com/v1/profile/me', {
      headers: {
        'Authorization': `Bearer ${this.accessToken}`,
        'X-API-Version': '1.0'
      }
    });

    if (!response.ok) {
      throw new Error(`API request failed: ${response.status}`);
    }

    return response.json();
  }

  async sendMessage(conversationId, content) {
    if (!this.accessToken) {
      throw new Error('Not authenticated');
    }

    const response = await fetch(
      `https://api.winkr.com/v1/chat/conversations/${conversationId}/messages`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.accessToken}`,
          'X-API-Version': '1.0',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          content: content,
          type: 'text'
        })
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to send message: ${response.status}`);
    }

    return response.json();
  }

  logout() {
    this.accessToken = null;
    this.refreshToken = null;
  }
}

// Usage example
const winkr = new WinkrSDK({
  clientId: 'your-client-id',
  redirectUri: 'https://your-app.com/callback',
  scope: 'profile messages matches'
});

// Login
try {
  const authResult = await winkr.login();
  console.log('Authenticated:', authResult);
  
  // Get user profile
  const profile = await winkr.getProfile();
  console.log('User profile:', profile);
  
  // Send message
  const message = await winkr.sendMessage('conv-123', 'Hello from third-party app!');
  console.log('Message sent:', message);
  
} catch (error) {
  console.error('Authentication failed:', error);
}
```

## Webhook Integration

Webhooks allow your application to receive real-time notifications about Winkr events.

### 1. Webhook Configuration

```javascript
// Register webhook endpoint
const registerWebhook = async (webhookUrl, events) => {
  const response = await fetch('https://api.winkr.com/v1/webhooks', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      url: webhookUrl,
      events: events,
      secret: 'your-webhook-secret'
    })
  });

  if (!response.ok) {
    throw new Error(`Failed to register webhook: ${response.status}`);
  }

  return response.json();
};

// Register for specific events
registerWebhook('https://your-app.com/webhooks/winkr', [
  'user.created',
  'user.updated',
  'message.sent',
  'match.created',
  'profile.viewed'
]);
```

### 2. Webhook Handler

```javascript
// Express.js webhook handler
const express = require('express');
const crypto = require('crypto');
const app = express();

app.use(express.raw({ type: 'application/json' }));

app.post('/webhooks/winkr', (req, res) => {
  const signature = req.headers['x-winkr-signature'];
  const payload = req.body;
  
  // Verify webhook signature
  const expectedSignature = crypto
    .createHmac('sha256', 'your-webhook-secret')
    .update(payload)
    .digest('hex');
    
  if (signature !== `sha256=${expectedSignature}`) {
    return res.status(401).send('Invalid signature');
  }
  
  try {
    const event = JSON.parse(payload);
    
    switch (event.type) {
      case 'user.created':
        handleUserCreated(event.data);
        break;
      case 'message.sent':
        handleMessageSent(event.data);
        break;
      case 'match.created':
        handleMatchCreated(event.data);
        break;
      default:
        console.log('Unhandled event type:', event.type);
    }
    
    res.status(200).send('OK');
  } catch (error) {
    console.error('Webhook processing error:', error);
    res.status(500).send('Internal server error');
  }
});

const handleUserCreated = (userData) => {
  console.log('New user created:', userData);
  // Process new user registration
  // Update your database
  // Send welcome email
  // etc.
};

const handleMessageSent = (messageData) => {
  console.log('New message sent:', messageData);
  // Process new message
  // Update chat interface
  // Send notification
  // etc.
};

const handleMatchCreated = (matchData) => {
  console.log('New match created:', matchData);
  // Process new match
  // Update matching algorithm
  // Send match notification
  // etc.
};

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});
```

### 3. Webhook Event Types

| Event Type | Description | Data |
|------------|-------------|------|
| `user.created` | New user registered | User object |
| `user.updated` | User profile updated | User object |
| `user.deleted` | User account deleted | User ID |
| `message.sent` | New message sent | Message object |
| `message.viewed` | Message viewed | Message ID |
| `match.created` | New match created | Match object |
| `match.deleted` | Match deleted | Match ID |
| `profile.viewed` | Profile viewed | View event data |
| `photo.uploaded` | New photo uploaded | Photo object |
| `verification.completed` | Verification completed | Verification result |

## API Client Libraries

### 1. Node.js Client

```javascript
// npm install winkr-api-client
const WinkrAPI = require('winkr-api-client');

const client = new WinkrAPI({
  apiKey: 'your-api-key',
  clientId: 'your-client-id',
  clientSecret: 'your-client-secret',
  environment: 'production' // or 'sandbox'
});

// Server-to-server operations
async function example() {
  try {
    // Get user profile
    const user = await client.users.getProfile('user-123');
    console.log('User:', user);
    
    // Send message
    const message = await client.chat.sendMessage('conv-123', 'Hello!');
    console.log('Message:', message);
    
    // Get matches
    const matches = await client.discovery.getMatches('user-123');
    console.log('Matches:', matches);
    
  } catch (error) {
    console.error('API Error:', error);
  }
}
```

### 2. Python Client

```python
# pip install winkr-api-client
from winkr_api import WinkrClient

client = WinkrClient(
    api_key='your-api-key',
    client_id='your-client-id',
    client_secret='your-client-secret',
    environment='production'  # or 'sandbox'
)

# Server-to-server operations
def example():
    try:
        # Get user profile
        user = client.users.get_profile('user-123')
        print(f"User: {user}")
        
        # Send message
        message = client.chat.send_message('conv-123', 'Hello!')
        print(f"Message: {message}")
        
        # Get matches
        matches = client.discovery.get_matches('user-123')
        print(f"Matches: {matches}")
        
    except Exception as error:
        print(f"API Error: {error}")
```

### 3. Ruby Client

```ruby
# gem install winkr-api-client
require 'winkr_api'

client = WinkrAPI::Client.new(
  api_key: 'your-api-key',
  client_id: 'your-client-id',
  client_secret: 'your-client-secret',
  environment: 'production' # or 'sandbox'
)

# Server-to-server operations
begin
  # Get user profile
  user = client.users.get_profile('user-123')
  puts "User: #{user}"
  
  # Send message
  message = client.chat.send_message('conv-123', 'Hello!')
  puts "Message: #{message}"
  
  # Get matches
  matches = client.discovery.get_matches('user-123')
  puts "Matches: #{matches}"
  
rescue => error
  puts "API Error: #{error}"
end
```

## Rate Limiting

Third-party integrations have specific rate limits to ensure fair usage.

### Rate Limit Tiers

| Tier | Requests per Hour | Requests per Minute | Features |
|------|-------------------|---------------------|----------|
| Basic | 1,000 | 100 | Profile access, basic messaging |
| Standard | 5,000 | 500 | Full API access, webhooks |
| Premium | 20,000 | 2,000 | Priority support, advanced features |

### Rate Limit Headers

```javascript
// Check rate limit status
const response = await fetch('https://api.winkr.com/v1/users/123', {
  headers: {
    'Authorization': `Bearer ${API_KEY}`
  }
});

const rateLimit = {
  limit: parseInt(response.headers.get('X-RateLimit-Limit')),
  remaining: parseInt(response.headers.get('X-RateLimit-Remaining')),
  reset: parseInt(response.headers.get('X-RateLimit-Reset'))
};

console.log('Rate limit:', rateLimit);
```

### Rate Limit Handling

```javascript
class RateLimitedClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.rateLimit = {
      limit: 100,
      remaining: 100,
      reset: Date.now() + 3600000
    };
  }

  async makeRequest(url, options = {}) {
    // Check if we're rate limited
    if (this.rateLimit.remaining <= 0) {
      const waitTime = this.rateLimit.reset - Date.now();
      if (waitTime > 0) {
        await this.sleep(waitTime);
      }
    }

    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        ...options.headers
      }
    });

    // Update rate limit info
    this.rateLimit = {
      limit: parseInt(response.headers.get('X-RateLimit-Limit')),
      remaining: parseInt(response.headers.get('X-RateLimit-Remaining')),
      reset: parseInt(response.headers.get('X-RateLimit-Reset')) * 1000
    };

    return response;
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request parameters",
    "details": {
      "field": "user_id",
      "reason": "User not found"
    }
  },
  "request_id": "req_123456789"
}
```

### Error Handling Best Practices

```javascript
class WinkrAPIError extends Error {
  constructor(response, data) {
    super(data.error.message);
    this.name = 'WinkrAPIError';
    this.code = data.error.code;
    this.status = response.status;
    this.requestId = data.request_id;
    this.details = data.error.details;
  }
}

async function handleAPIRequest(requestFunction) {
  try {
    const response = await requestFunction();
    return await response.json();
  } catch (error) {
    if (error.response) {
      const data = await error.response.json();
      throw new WinkrAPIError(error.response, data);
    }
    throw error;
  }
}

// Usage
try {
  const user = await handleAPIRequest(() => 
    fetch('https://api.winkr.com/v1/users/123', {
      headers: { 'Authorization': `Bearer ${API_KEY}` }
    })
  );
  console.log('User:', user);
} catch (error) {
  if (error instanceof WinkrAPIError) {
    console.error(`API Error ${error.code}: ${error.message}`);
    console.error('Request ID:', error.requestId);
    
    // Handle specific error codes
    switch (error.code) {
      case 'RATE_LIMIT_EXCEEDED':
        console.log('Rate limit exceeded, retry later');
        break;
      case 'INVALID_TOKEN':
        console.log('Token invalid, re-authenticate');
        break;
      case 'USER_NOT_FOUND':
        console.log('User not found');
        break;
      default:
        console.log('Unknown error occurred');
    }
  } else {
    console.error('Network error:', error.message);
  }
}
```

## Security Best Practices

### 1. Credential Management

```javascript
// Use environment variables for sensitive data
const config = {
  apiKey: process.env.WINKR_API_KEY,
  clientId: process.env.WINKR_CLIENT_ID,
  clientSecret: process.env.WINKR_CLIENT_SECRET,
  webhookSecret: process.env.WINKR_WEBHOOK_SECRET
};

// Never hardcode credentials in source code
// Use secure key management systems in production
```

### 2. Input Validation

```javascript
// Validate all user inputs before sending to API
function validateUserId(userId) {
  if (!userId || typeof userId !== 'string') {
    throw new Error('Invalid user ID');
  }
  
  if (!/^[a-zA-Z0-9_-]+$/.test(userId)) {
    throw new Error('User ID contains invalid characters');
  }
  
  return userId;
}

function validateMessage(content) {
  if (!content || typeof content !== 'string') {
    throw new Error('Message content is required');
  }
  
  if (content.length > 1000) {
    throw new Error('Message too long');
  }
  
  // Sanitize content to prevent XSS
  return content.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
}
```

### 3. HTTPS Only

```javascript
// Ensure all requests use HTTPS
function makeSecureRequest(url, options = {}) {
  if (!url.startsWith('https://')) {
    throw new Error('Only HTTPS requests are allowed');
  }
  
  return fetch(url, {
    ...options,
    headers: {
      'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
      ...options.headers
    }
  });
}
```

### 4. Webhook Security

```javascript
// Verify webhook signatures
function verifyWebhookSignature(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
    
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(`sha256=${expectedSignature}`)
  );
}

// Use in webhook handler
app.post('/webhooks/winkr', (req, res) => {
  const signature = req.headers['x-winkr-signature'];
  const payload = req.body;
  
  if (!verifyWebhookSignature(payload, signature, webhookSecret)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook...
});
```

## Examples

### 1. Complete Integration Example

```javascript
// Complete third-party integration example
const express = require('express');
const crypto = require('crypto');
const WinkrAPI = require('winkr-api-client');

const app = express();
app.use(express.json());

// Initialize Winkr client
const winkr = new WinkrAPI({
  apiKey: process.env.WINKR_API_KEY,
  clientId: process.env.WINKR_CLIENT_ID,
  clientSecret: process.env.WINKR_CLIENT_SECRET
});

// OAuth callback endpoint
app.get('/auth/callback', async (req, res) => {
  try {
    const { code, state } = req.query;
    
    // Exchange code for access token
    const authResult = await winkr.oauth.exchangeCodeForToken(code);
    
    // Store tokens securely
    // In a real app, store in database with user session
    req.session.winkrTokens = authResult;
    
    res.redirect('/dashboard');
  } catch (error) {
    res.status(500).send('Authentication failed');
  }
});

// Get user profile
app.get('/api/profile', async (req, res) => {
  try {
    const tokens = req.session.winkrTokens;
    if (!tokens) {
      return res.status(401).send('Not authenticated');
    }
    
    const profile = await winkr.users.getProfile(tokens.access_token);
    res.json(profile);
  } catch (error) {
    res.status(500).send('Failed to get profile');
  }
});

// Send message
app.post('/api/messages', async (req, res) => {
  try {
    const { conversationId, content } = req.body;
    const tokens = req.session.winkrTokens;
    
    if (!tokens) {
      return res.status(401).send('Not authenticated');
    }
    
    const message = await winkr.chat.sendMessage(
      conversationId,
      content,
      tokens.access_token
    );
    
    res.json(message);
  } catch (error) {
    res.status(500).send('Failed to send message');
  }
});

// Webhook endpoint
app.post('/webhooks/winkr', (req, res) => {
  const signature = req.headers['x-winkr-signature'];
  const payload = JSON.stringify(req.body);
  
  // Verify signature
  const expectedSignature = crypto
    .createHmac('sha256', process.env.WINKR_WEBHOOK_SECRET)
    .update(payload)
    .digest('hex');
    
  if (signature !== `sha256=${expectedSignature}`) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook events
  const event = req.body;
  
  switch (event.type) {
    case 'message.sent':
      handleNewMessage(event.data);
      break;
    case 'match.created':
      handleNewMatch(event.data);
      break;
    default:
      console.log('Unhandled event:', event.type);
  }
  
  res.status(200).send('OK');
});

function handleNewMessage(messageData) {
  console.log('New message received:', messageData);
  // Update your application state
  // Send notifications
  // Update UI
}

function handleNewMatch(matchData) {
  console.log('New match created:', matchData);
  // Update matching algorithm
  // Send match notifications
  // Update user interface
}

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
```

### 2. Analytics Integration

```javascript
// Analytics service integration
class WinkrAnalytics {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baseURL = 'https://api.winkr.com/v1/analytics';
  }

  async getUserEngagement(userId, startDate, endDate) {
    const response = await fetch(
      `${this.baseURL}/users/${userId}/engagement?start_date=${startDate}&end_date=${endDate}`,
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'X-API-Version': '1.0'
        }
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to get engagement data: ${response.status}`);
    }

    return response.json();
  }

  async getMatchingAnalytics(userId) {
    const response = await fetch(
      `${this.baseURL}/users/${userId}/matching`,
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'X-API-Version': '1.0'
        }
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to get matching analytics: ${response.status}`);
    }

    return response.json();
  }

  async generateReport(reportType, filters = {}) {
    const response = await fetch(
      `${this.baseURL}/reports`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'X-API-Version': '1.0',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          type: reportType,
          filters: filters
        })
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to generate report: ${response.status}`);
    }

    return response.json();
  }
}

// Usage example
const analytics = new WinkrAnalytics(process.env.WINKR_API_KEY);

async function generateUserReport(userId) {
  try {
    const engagement = await analytics.getUserEngagement(
      userId,
      '2025-01-01',
      '2025-12-31'
    );
    
    const matching = await analytics.getMatchingAnalytics(userId);
    
    const report = {
      userId: userId,
      engagement: engagement,
      matching: matching,
      generatedAt: new Date().toISOString()
    };
    
    console.log('User report:', report);
    return report;
  } catch (error) {
    console.error('Failed to generate report:', error);
  }
}
```

This comprehensive third-party integration guide provides everything needed to successfully integrate the Winkr API into external applications, from basic authentication to advanced webhook handling and analytics integration.