# WebSocket Integration Guide

This guide provides comprehensive instructions for integrating with Winkr's WebSocket API for real-time communication features.

## Table of Contents

- [Overview](#overview)
- [Connection Management](#connection-management)
- [Authentication](#authentication)
- [Events](#events)
- [Message Types](#message-types)
- [Real-time Features](#real-time-features)
- [Error Handling](#error-handling)
- [Reconnection Strategy](#reconnection-strategy)
- [Scaling Considerations](#scaling-considerations)
- [Examples](#examples)

## Overview

The Winkr WebSocket API enables real-time communication between clients and the server, supporting features like:

- **Live Messaging**: Instant message delivery and read receipts
- **Typing Indicators**: Real-time typing status updates
- **Online Status**: User presence and availability
- **Match Notifications**: Real-time match updates
- **Ephemeral Photos**: Time-limited photo sharing
- **Live Notifications**: Push notifications via WebSocket

### WebSocket Endpoint

```
wss://api.winkr.com/v1/ws
```

### Connection URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token` | string | Yes | JWT access token for authentication |
| `version` | string | No | API version (default: 1.0) |
| `client_id` | string | No | Client identifier for tracking |

## Connection Management

### 1. Basic Connection

```javascript
// JavaScript WebSocket connection
const token = 'your-jwt-access-token';
const ws = new WebSocket(`wss://api.winkr.com/v1/ws?token=${token}`);

ws.onopen = function(event) {
  console.log('WebSocket connected');
  
  // Send initial connection message
  ws.send(JSON.stringify({
    type: 'connection:established',
    data: {
      client_version: '1.0.0',
      platform: 'web'
    }
  }));
};

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  handleMessage(message);
};

ws.onerror = function(error) {
  console.error('WebSocket error:', error);
};

ws.onclose = function(event) {
  console.log('WebSocket closed:', event.code, event.reason);
  
  // Implement reconnection logic
  if (event.code !== 1000) {
    setTimeout(connect, 5000);
  }
};
```

### 2. Connection with Authentication

```javascript
class WinkrWebSocket {
  constructor(token, options = {}) {
    this.token = token;
    this.options = {
      version: '1.0',
      client_id: this.generateClientId(),
      platform: 'web',
      reconnect: true,
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      ...options
    };
    
    this.ws = null;
    this.reconnectAttempts = 0;
    this.isConnecting = false;
    this.eventHandlers = new Map();
    
    this.connect();
  }
  
  connect() {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }
    
    this.isConnecting = true;
    
    const url = `wss://api.winkr.com/v1/ws?token=${this.token}&version=${this.options.version}&client_id=${this.options.client_id}`;
    
    this.ws = new WebSocket(url);
    
    this.ws.onopen = (event) => {
      console.log('WebSocket connected');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      
      // Send connection info
      this.send('connection:established', {
        client_version: this.options.version,
        platform: this.options.platform
      });
      
      this.emit('connected', event);
    };
    
    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.emit('error', error);
    };
    
    this.ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      this.isConnecting = false;
      this.emit('disconnected', event);
      
      // Attempt reconnection if enabled
      if (this.options.reconnect && event.code !== 1000) {
        this.attemptReconnect();
      }
    };
  }
  
  disconnect() {
    this.options.reconnect = false;
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }
  
  send(type, data = {}) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return false;
    }
    
    const message = {
      type: type,
      data: data,
      timestamp: new Date().toISOString(),
      message_id: this.generateMessageId()
    };
    
    this.ws.send(JSON.stringify(message));
    return true;
  }
  
  on(event, handler) {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event).push(handler);
  }
  
  off(event, handler) {
    if (this.eventHandlers.has(event)) {
      const handlers = this.eventHandlers.get(event);
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }
  
  emit(event, data) {
    if (this.eventHandlers.has(event)) {
      this.eventHandlers.get(event).forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error('Event handler error:', error);
        }
      });
    }
  }
  
  handleMessage(message) {
    console.log('Received message:', message);
    
    // Emit specific event type
    this.emit(message.type, message.data);
    
    // Emit generic message event
    this.emit('message', message);
  }
  
  attemptReconnect() {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('reconnect_failed');
      return;
    }
    
    this.reconnectAttempts++;
    console.log(`Attempting reconnection (${this.reconnectAttempts}/${this.options.maxReconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, this.options.reconnectInterval);
  }
  
  generateClientId() {
    return 'client_' + Math.random().toString(36).substr(2, 9);
  }
  
  generateMessageId() {
    return 'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }
}

// Usage example
const winkrWS = new WinkrWebSocket('your-jwt-token', {
  platform: 'web',
  reconnect: true
});

winkrWS.on('connected', () => {
  console.log('Connected to Winkr WebSocket');
});

winkrWS.on('message:new', (data) => {
  console.log('New message:', data);
  // Handle new message
});

winkrWS.on('typing:indicator', (data) => {
  console.log('Typing indicator:', data);
  // Handle typing indicator
});
```

## Authentication

### 1. Token-based Authentication

WebSocket connections require a valid JWT access token passed as a query parameter.

```javascript
// Get token from your authentication system
const token = await getAccessToken();

// Connect with token
const ws = new WebSocket(`wss://api.winkr.com/v1/ws?token=${token}`);
```

### 2. Token Refresh

When the token expires, the server will send an authentication error. Implement token refresh logic:

```javascript
class WinkrWebSocketWithAuth extends WinkrWebSocket {
  constructor(token, refreshToken, options = {}) {
    super(token, options);
    this.refreshToken = refreshToken;
    this.tokenExpiry = this.getTokenExpiry(token);
  }
  
  handleMessage(message) {
    super.handleMessage(message);
    
    // Handle authentication errors
    if (message.type === 'error' && message.data.code === 'TOKEN_EXPIRED') {
      this.refreshTokenAndReconnect();
    }
  }
  
  async refreshTokenAndReconnect() {
    try {
      const newToken = await this.refreshAccessToken();
      this.token = newToken;
      this.tokenExpiry = this.getTokenExpiry(newToken);
      
      // Reconnect with new token
      this.disconnect();
      this.connect();
    } catch (error) {
      console.error('Token refresh failed:', error);
      this.emit('auth_failed', error);
    }
  }
  
  async refreshAccessToken() {
    const response = await fetch('https://api.winkr.com/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        refresh_token: this.refreshToken
      })
    });
    
    if (!response.ok) {
      throw new Error('Token refresh failed');
    }
    
    const data = await response.json();
    return data.access_token;
  }
  
  getTokenExpiry(token) {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000; // Convert to milliseconds
    } catch (error) {
      return Date.now() + 3600000; // Default to 1 hour
    }
  }
}
```

## Events

### 1. Connection Events

| Event | Description | Data |
|-------|-------------|------|
| `connection:established` | Initial connection established | Connection info |
| `connection:authenticated` | Authentication successful | User info |
| `connection:error` | Connection error | Error details |
| `connection:closed` | Connection closed | Close reason |

### 2. Message Events

| Event | Description | Data |
|-------|-------------|------|
| `message:new` | New message received | Message object |
| `message:sent` | Message sent successfully | Message object |
| `message:viewed` | Message viewed by recipient | Message ID |
| `message:deleted` | Message deleted | Message ID |

### 3. Chat Events

| Event | Description | Data |
|-------|-------------|------|
| `conversation:joined` | Joined conversation | Conversation info |
| `conversation:left` | Left conversation | Conversation ID |
| `typing:start` | User started typing | User and conversation info |
| `typing:stop` | User stopped typing | User and conversation info |

### 4. Match Events

| Event | Description | Data |
|-------|-------------|------|
| `match:new` | New match created | Match object |
| `match:updated` | Match status updated | Match object |
| `match:deleted` | Match deleted | Match ID |

### 5. Profile Events

| Event | Description | Data |
|-------|-------------|------|
| `profile:viewed` | Profile viewed by another user | Viewer info |
| `profile:updated` | User profile updated | Profile changes |
| `profile:online` | User came online | User info |
| `profile:offline` | User went offline | User info |

## Message Types

### 1. Client Messages

#### Join Conversation

```javascript
winkrWS.send('conversation:join', {
  conversation_id: 'conv_123456789'
});
```

#### Leave Conversation

```javascript
winkrWS.send('conversation:leave', {
  conversation_id: 'conv_123456789'
});
```

#### Send Message

```javascript
winkrWS.send('message:send', {
  conversation_id: 'conv_123456789',
  content: 'Hello, world!',
  type: 'text',
  metadata: {
    reply_to: 'msg_987654321'
  }
});
```

#### Mark Message as Viewed

```javascript
winkrWS.send('message:viewed', {
  message_id: 'msg_987654321',
  conversation_id: 'conv_123456789'
});
```

#### Start Typing

```javascript
winkrWS.send('typing:start', {
  conversation_id: 'conv_123456789'
});
```

#### Stop Typing

```javascript
winkrWS.send('typing:stop', {
  conversation_id: 'conv_123456789'
});
```

#### Update Online Status

```javascript
winkrWS.send('status:update', {
  status: 'online', // online, away, busy, offline
  message: 'Working from home'
});
```

### 2. Server Messages

#### New Message

```json
{
  "type": "message:new",
  "data": {
    "id": "msg_987654321",
    "conversation_id": "conv_123456789",
    "sender_id": "user_456789",
    "content": "Hello, world!",
    "type": "text",
    "created_at": "2025-12-01T10:30:00Z",
    "metadata": {}
  },
  "timestamp": "2025-12-01T10:30:00Z",
  "message_id": "server_msg_123456789"
}
```

#### Typing Indicator

```json
{
  "type": "typing:indicator",
  "data": {
    "user_id": "user_456789",
    "conversation_id": "conv_123456789",
    "is_typing": true
  },
  "timestamp": "2025-12-01T10:30:00Z",
  "message_id": "server_msg_123456790"
}
```

#### Online Status Update

```json
{
  "type": "status:update",
  "data": {
    "user_id": "user_456789",
    "status": "online",
    "message": "Available",
    "last_seen": "2025-12-01T10:25:00Z"
  },
  "timestamp": "2025-12-01T10:30:00Z",
  "message_id": "server_msg_123456791"
}
```

## Real-time Features

### 1. Live Chat Implementation

```javascript
class LiveChat extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, options);
    this.currentConversation = null;
    this.typingTimeout = null;
    
    this.setupEventHandlers();
  }
  
  setupEventHandlers() {
    this.on('message:new', (data) => {
      this.handleNewMessage(data);
    });
    
    this.on('typing:indicator', (data) => {
      this.handleTypingIndicator(data);
    });
    
    this.on('conversation:joined', (data) => {
      this.handleConversationJoined(data);
    });
  }
  
  joinConversation(conversationId) {
    this.currentConversation = conversationId;
    this.send('conversation:join', {
      conversation_id: conversationId
    });
  }
  
  leaveConversation() {
    if (this.currentConversation) {
      this.send('conversation:leave', {
        conversation_id: this.currentConversation
      });
      this.currentConversation = null;
    }
  }
  
  sendMessage(content, type = 'text', metadata = {}) {
    if (!this.currentConversation) {
      throw new Error('No active conversation');
    }
    
    return this.send('message:send', {
      conversation_id: this.currentConversation,
      content: content,
      type: type,
      metadata: metadata
    });
  }
  
  startTyping() {
    if (!this.currentConversation) return;
    
    this.send('typing:start', {
      conversation_id: this.currentConversation
    });
    
    // Auto-stop typing after 3 seconds
    clearTimeout(this.typingTimeout);
    this.typingTimeout = setTimeout(() => {
      this.stopTyping();
    }, 3000);
  }
  
  stopTyping() {
    if (!this.currentConversation) return;
    
    clearTimeout(this.typingTimeout);
    this.send('typing:stop', {
      conversation_id: this.currentConversation
    });
  }
  
  handleNewMessage(data) {
    console.log('New message received:', data);
    
    // Update UI
    this.addMessageToUI(data);
    
    // Mark as viewed if in current conversation
    if (data.conversation_id === this.currentConversation) {
      this.markMessageAsViewed(data.id);
    }
  }
  
  handleTypingIndicator(data) {
    console.log('Typing indicator:', data);
    
    // Update UI to show/hide typing indicator
    this.updateTypingIndicator(data);
  }
  
  handleConversationJoined(data) {
    console.log('Joined conversation:', data);
    
    // Load conversation history
    this.loadConversationHistory(data.conversation_id);
  }
  
  markMessageAsViewed(messageId) {
    this.send('message:viewed', {
      message_id: messageId,
      conversation_id: this.currentConversation
    });
  }
  
  addMessageToUI(message) {
    // Implementation depends on your UI framework
    // This is a placeholder for UI update logic
    const messageElement = document.createElement('div');
    messageElement.className = 'message';
    messageElement.textContent = message.content;
    document.getElementById('chat-messages').appendChild(messageElement);
  }
  
  updateTypingIndicator(data) {
    const indicator = document.getElementById('typing-indicator');
    if (data.is_typing) {
      indicator.textContent = `${data.user_id} is typing...`;
      indicator.style.display = 'block';
    } else {
      indicator.style.display = 'none';
    }
  }
  
  async loadConversationHistory(conversationId) {
    try {
      const response = await fetch(
        `https://api.winkr.com/v1/chat/conversations/${conversationId}/messages`,
        {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        }
      );
      
      const data = await response.json();
      
      // Display messages in UI
      data.messages.forEach(message => {
        this.addMessageToUI(message);
      });
      
    } catch (error) {
      console.error('Failed to load conversation history:', error);
    }
  }
}

// Usage example
const chat = new LiveChat('your-jwt-token');

// Join a conversation
chat.joinConversation('conv_123456789');

// Send a message
chat.sendMessage('Hello, world!');

// Handle typing in input field
document.getElementById('message-input').addEventListener('input', () => {
  chat.startTyping();
});

document.getElementById('send-button').addEventListener('click', () => {
  const input = document.getElementById('message-input');
  chat.sendMessage(input.value);
  input.value = '';
  chat.stopTyping();
});
```

### 2. Real-time Match Notifications

```javascript
class MatchNotifications extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, options);
    this.setupEventHandlers();
  }
  
  setupEventHandlers() {
    this.on('match:new', (data) => {
      this.handleNewMatch(data);
    });
    
    this.on('match:updated', (data) => {
      this.handleMatchUpdate(data);
    });
  }
  
  handleNewMatch(matchData) {
    console.log('New match:', matchData);
    
    // Show notification
    this.showNotification('New Match!', `You matched with ${matchData.user.name}`);
    
    // Update UI
    this.addMatchToList(matchData);
    
    // Play sound
    this.playNotificationSound();
  }
  
  handleMatchUpdate(matchData) {
    console.log('Match updated:', matchData);
    
    // Update match in UI
    this.updateMatchInList(matchData);
  }
  
  showNotification(title, body) {
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(title, {
        body: body,
        icon: '/assets/icons/match.png',
        badge: '/assets/icons/badge.png'
      });
    }
  }
  
  addMatchToList(matchData) {
    const matchesList = document.getElementById('matches-list');
    const matchElement = document.createElement('div');
    matchElement.className = 'match-item';
    matchElement.innerHTML = `
      <img src="${matchData.user.avatar_url}" alt="${matchData.user.name}">
      <div class="match-info">
        <h3>${matchData.user.name}</h3>
        <p>Matched ${this.formatTime(matchData.created_at)}</p>
      </div>
    `;
    matchesList.appendChild(matchElement);
  }
  
  updateMatchInList(matchData) {
    const matchElement = document.querySelector(`[data-match-id="${matchData.id}"]`);
    if (matchElement) {
      // Update match status or other properties
      matchElement.setAttribute('data-status', matchData.status);
    }
  }
  
  playNotificationSound() {
    const audio = new Audio('/sounds/match-notification.mp3');
    audio.play().catch(error => {
      console.error('Failed to play notification sound:', error);
    });
  }
  
  formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return 'just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)} minutes ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} hours ago`;
    return `${Math.floor(diff / 86400000)} days ago`;
  }
}

// Request notification permission
if ('Notification' in window && Notification.permission === 'default') {
  Notification.requestPermission();
}

// Usage example
const matchNotifications = new MatchNotifications('your-jwt-token');
```

### 3. Ephemeral Photos Integration

```javascript
class EphemeralPhotos extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, options);
    this.setupEventHandlers();
  }
  
  setupEventHandlers() {
    this.on('ephemeral_photo:new', (data) => {
      this.handleNewEphemeralPhoto(data);
    });
    
    this.on('ephemeral_photo:viewed', (data) => {
      this.handlePhotoViewed(data);
    });
    
    this.on('ephemeral_photo:expired', (data) => {
      this.handlePhotoExpired(data);
    });
  }
  
  sendEphemeralPhoto(conversationId, photoUrl, expiresIn = 10) {
    return this.send('ephemeral_photo:send', {
      conversation_id: conversationId,
      photo_url: photoUrl,
      expires_in: expiresIn
    });
  }
  
  viewEphemeralPhoto(photoId) {
    return this.send('ephemeral_photo:view', {
      photo_id: photoId
    });
  }
  
  handleNewEphemeralPhoto(data) {
    console.log('New ephemeral photo:', data);
    
    // Show photo with countdown timer
    this.displayEphemeralPhoto(data);
    
    // Start expiration timer
    this.startExpirationTimer(data);
  }
  
  handlePhotoViewed(data) {
    console.log('Photo viewed:', data);
    
    // Mark photo as viewed in UI
    this.markPhotoAsViewed(data.photo_id);
  }
  
  handlePhotoExpired(data) {
    console.log('Photo expired:', data);
    
    // Remove photo from UI
    this.removePhotoFromUI(data.photo_id);
    
    // Show expiration notification
    this.showExpirationNotification(data);
  }
  
  displayEphemeralPhoto(photoData) {
    const container = document.getElementById('ephemeral-photos');
    const photoElement = document.createElement('div');
    photoElement.className = 'ephemeral-photo';
    photoElement.setAttribute('data-photo-id', photoData.id);
    
    photoElement.innerHTML = `
      <img src="${photoData.photo_url}" alt="Ephemeral photo">
      <div class="photo-timer">
        <div class="timer-bar"></div>
        <span class="timer-text">${photoData.expires_in}s</span>
      </div>
    `;
    
    container.appendChild(photoElement);
    
    // Auto-view the photo
    this.viewEphemeralPhoto(photoData.id);
  }
  
  startExpirationTimer(photoData) {
    let timeLeft = photoData.expires_in;
    const photoElement = document.querySelector(`[data-photo-id="${photoData.id}"]`);
    const timerText = photoElement.querySelector('.timer-text');
    const timerBar = photoElement.querySelector('.timer-bar');
    
    const interval = setInterval(() => {
      timeLeft--;
      timerText.textContent = `${timeLeft}s`;
      timerBar.style.width = `${(timeLeft / photoData.expires_in) * 100}%`;
      
      if (timeLeft <= 0) {
        clearInterval(interval);
      }
    }, 1000);
  }
  
  markPhotoAsViewed(photoId) {
    const photoElement = document.querySelector(`[data-photo-id="${photoId}"]`);
    if (photoElement) {
      photoElement.classList.add('viewed');
    }
  }
  
  removePhotoFromUI(photoId) {
    const photoElement = document.querySelector(`[data-photo-id="${photoId}"]`);
    if (photoElement) {
      photoElement.remove();
    }
  }
  
  showExpirationNotification(data) {
    this.showNotification('Photo Expired', 'The ephemeral photo has expired');
  }
  
  showNotification(title, body) {
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(title, {
        body: body,
        icon: '/assets/icons/photo.png'
      });
    }
  }
}

// Usage example
const ephemeralPhotos = new EphemeralPhotos('your-jwt-token');

// Send an ephemeral photo
ephemeralPhotos.sendEphemeralPhoto('conv_123456789', 'https://example.com/photo.jpg', 10);
```

## Error Handling

### 1. Error Types

| Error Code | Description | Action |
|-----------|-------------|--------|
| `TOKEN_EXPIRED` | JWT token has expired | Refresh token and reconnect |
| `TOKEN_INVALID` | JWT token is invalid | Re-authenticate user |
| `RATE_LIMITED` | Too many messages sent | Implement backoff strategy |
| `CONVERSATION_NOT_FOUND` | Conversation doesn't exist | Verify conversation ID |
| `PERMISSION_DENIED` | No permission for action | Check user permissions |
| `MESSAGE_TOO_LARGE` | Message exceeds size limit | Reduce message size |

### 2. Error Handling Implementation

```javascript
class RobustWebSocket extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, options);
    this.setupErrorHandling();
  }
  
  setupErrorHandling() {
    this.on('error', (error) => {
      this.handleError(error);
    });
    
    this.on('message', (message) => {
      if (message.type === 'error') {
        this.handleServerError(message.data);
      }
    });
  }
  
  handleError(error) {
    console.error('WebSocket error:', error);
    
    // Implement error-specific handling
    if (error.code === 'TOKEN_EXPIRED') {
      this.handleTokenExpired();
    } else if (error.code === 'RATE_LIMITED') {
      this.handleRateLimited();
    }
  }
  
  handleServerError(errorData) {
    console.error('Server error:', errorData);
    
    switch (errorData.code) {
      case 'TOKEN_EXPIRED':
        this.handleTokenExpired();
        break;
      case 'RATE_LIMITED':
        this.handleRateLimited();
        break;
      case 'CONVERSATION_NOT_FOUND':
        this.handleConversationNotFound(errorData);
        break;
      default:
        this.handleGenericError(errorData);
    }
  }
  
  handleTokenExpired() {
    console.log('Token expired, attempting refresh...');
    
    // Implement token refresh logic
    this.refreshToken()
      .then(newToken => {
        this.token = newToken;
        this.reconnect();
      })
      .catch(error => {
        console.error('Token refresh failed:', error);
        this.emit('auth_failed', error);
      });
  }
  
  handleRateLimited() {
    console.log('Rate limited, implementing backoff...');
    
    // Implement exponential backoff
    const backoffTime = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    
    setTimeout(() => {
      this.reconnect();
    }, backoffTime);
  }
  
  handleConversationNotFound(errorData) {
    console.error('Conversation not found:', errorData);
    
    // Notify user and handle gracefully
    this.emit('conversation_error', errorData);
  }
  
  handleGenericError(errorData) {
    console.error('Generic error:', errorData);
    
    // Show user-friendly error message
    this.emit('error', errorData);
  }
  
  async refreshToken() {
    // Implement token refresh logic
    const response = await fetch('https://api.winkr.com/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        refresh_token: this.refreshToken
      })
    });
    
    if (!response.ok) {
      throw new Error('Token refresh failed');
    }
    
    const data = await response.json();
    return data.access_token;
  }
}
```

## Reconnection Strategy

### 1. Exponential Backoff

```javascript
class ReconnectingWebSocket extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, {
      reconnect: true,
      reconnectInterval: 1000,
      maxReconnectInterval: 30000,
      reconnectDecay: 1.5,
      maxReconnectAttempts: 10,
      ...options
    });
    
    this.currentReconnectInterval = this.options.reconnectInterval;
  }
  
  attemptReconnect() {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('reconnect_failed');
      return;
    }
    
    this.reconnectAttempts++;
    console.log(`Attempting reconnection (${this.reconnectAttempts}/${this.options.maxReconnectAttempts}) in ${this.currentReconnectInterval}ms`);
    
    setTimeout(() => {
      this.connect();
      
      // Increase reconnect interval for next attempt
      this.currentReconnectInterval = Math.min(
        this.currentReconnectInterval * this.options.reconnectDecay,
        this.options.maxReconnectInterval
      );
    }, this.currentReconnectInterval);
  }
  
  onopen(event) {
    super.onopen(event);
    
    // Reset reconnect interval on successful connection
    this.currentReconnectInterval = this.options.reconnectInterval;
  }
}
```

### 2. Connection Health Monitoring

```javascript
class HealthMonitoringWebSocket extends WinkrWebSocket {
  constructor(token, options = {}) {
    super(token, options);
    this.heartbeatInterval = null;
    this.heartbeatTimeout = null;
    this.isHealthy = true;
    
    this.setupHealthMonitoring();
  }
  
  setupHealthMonitoring() {
    this.on('connected', () => {
      this.startHeartbeat();
    });
    
    this.on('disconnected', () => {
      this.stopHeartbeat();
    });
    
    this.on('message', () => {
      // Reset heartbeat timeout on any message
      this.resetHeartbeatTimeout();
    });
  }
  
  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.sendHeartbeat();
    }, 30000); // Send heartbeat every 30 seconds
  }
  
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
    
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout);
      this.heartbeatTimeout = null;
    }
  }
  
  sendHeartbeat() {
    this.send('heartbeat', {
      timestamp: new Date().toISOString()
    });
    
    // Set timeout for heartbeat response
    this.heartbeatTimeout = setTimeout(() => {
      console.warn('Heartbeat timeout, connection may be unhealthy');
      this.isHealthy = false;
      this.emit('unhealthy');
    }, 5000); // Wait 5 seconds for response
  }
  
  resetHeartbeatTimeout() {
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout);
      this.heartbeatTimeout = null;
    }
    
    if (!this.isHealthy) {
      this.isHealthy = true;
      this.emit('healthy');
    }
  }
}
```

## Scaling Considerations

### 1. Connection Pooling

```javascript
class ConnectionPool {
  constructor(maxConnections = 5) {
    this.maxConnections = maxConnections;
    this.connections = [];
    this.waitingQueue = [];
  }
  
  async getConnection(token) {
    // Check for available connection
    const availableConnection = this.connections.find(conn => 
      conn.isAvailable() && conn.token === token
    );
    
    if (availableConnection) {
      availableConnection.reserve();
      return availableConnection;
    }
    
    // Create new connection if under limit
    if (this.connections.length < this.maxConnections) {
      const connection = new WinkrWebSocket(token);
      connection.reserve();
      this.connections.push(connection);
      return connection;
    }
    
    // Wait for available connection
    return new Promise((resolve) => {
      this.waitingQueue.push({ token, resolve });
    });
  }
  
  releaseConnection(connection) {
    connection.release();
    
    // Check waiting queue
    if (this.waitingQueue.length > 0) {
      const { token, resolve } = this.waitingQueue.shift();
      
      // Reuse connection if token matches
      if (connection.token === token) {
        connection.reserve();
        resolve(connection);
      } else {
        // Create new connection for different token
        this.getConnection(token).then(resolve);
      }
    }
  }
}
```

### 2. Load Balancing

```javascript
class LoadBalancedWebSocket {
  constructor(endpoints, options = {}) {
    this.endpoints = endpoints;
    this.currentEndpointIndex = 0;
    this.options = options;
    this.connection = null;
  }
  
  async connect(token) {
    const endpoint = this.selectEndpoint();
    
    try {
      this.connection = new WinkrWebSocket(token, {
        ...this.options,
        endpoint: endpoint
      });
      
      await this.waitForConnection();
      return this.connection;
    } catch (error) {
      console.error(`Failed to connect to ${endpoint}:`, error);
      
      // Try next endpoint
      this.currentEndpointIndex = (this.currentEndpointIndex + 1) % this.endpoints.length;
      return this.connect(token);
    }
  }
  
  selectEndpoint() {
    // Simple round-robin selection
    return this.endpoints[this.currentEndpointIndex];
  }
  
  async waitForConnection() {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Connection timeout'));
      }, 10000);
      
      this.connection.on('connected', () => {
        clearTimeout(timeout);
        resolve();
      });
      
      this.connection.on('error', (error) => {
        clearTimeout(timeout);
        reject(error);
      });
    });
  }
}
```

## Examples

### 1. Complete Chat Application

```javascript
// Complete chat application with all features
class WinkrChatApp {
  constructor(config) {
    this.config = config;
    this.ws = null;
    this.currentUser = null;
    this.currentConversation = null;
    this.conversations = new Map();
    this.messages = new Map();
    
    this.init();
  }
  
  async init() {
    try {
      // Authenticate user
      await this.authenticate();
      
      // Initialize WebSocket connection
      this.initWebSocket();
      
      // Load conversations
      await this.loadConversations();
      
      // Setup UI event handlers
      this.setupUIHandlers();
      
    } catch (error) {
      console.error('Failed to initialize app:', error);
      this.showError('Failed to initialize application');
    }
  }
  
  async authenticate() {
    const response = await fetch('https://api.winkr.com/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: this.config.email,
        password: this.config.password
      })
    });
    
    if (!response.ok) {
      throw new Error('Authentication failed');
    }
    
    const data = await response.json();
    this.currentUser = data.user;
    this.token = data.access_token;
  }
  
  initWebSocket() {
    this.ws = new WinkrWebSocket(this.token, {
      reconnect: true,
      maxReconnectAttempts: 10
    });
    
    this.ws.on('connected', () => {
      console.log('Connected to Winkr WebSocket');
      this.updateConnectionStatus('connected');
    });
    
    this.ws.on('disconnected', () => {
      console.log('Disconnected from Winkr WebSocket');
      this.updateConnectionStatus('disconnected');
    });
    
    this.ws.on('message:new', (data) => {
      this.handleNewMessage(data);
    });
    
    this.ws.on('typing:indicator', (data) => {
      this.handleTypingIndicator(data);
    });
    
    this.ws.on('match:new', (data) => {
      this.handleNewMatch(data);
    });
  }
  
  async loadConversations() {
    const response = await fetch('https://api.winkr.com/v1/chat/conversations', {
      headers: {
        'Authorization': `Bearer ${this.token}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Failed to load conversations');
    }
    
    const data = await response.json();
    this.conversations = new Map(
      data.conversations.map(conv => [conv.id, conv])
    );
    
    this.renderConversationsList();
  }
  
  setupUIHandlers() {
    // Conversation selection
    document.getElementById('conversations-list').addEventListener('click', (e) => {
      const conversationElement = e.target.closest('.conversation-item');
      if (conversationElement) {
        const conversationId = conversationElement.getAttribute('data-conversation-id');
        this.selectConversation(conversationId);
      }
    });
    
    // Message sending
    document.getElementById('send-button').addEventListener('click', () => {
      this.sendMessage();
    });
    
    document.getElementById('message-input').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        this.sendMessage();
      }
    });
    
    // Typing indicator
    document.getElementById('message-input').addEventListener('input', () => {
      this.handleTyping();
    });
  }
  
  selectConversation(conversationId) {
    // Leave current conversation
    if (this.currentConversation) {
      this.ws.send('conversation:leave', {
        conversation_id: this.currentConversation
      });
    }
    
    // Join new conversation
    this.currentConversation = conversationId;
    this.ws.send('conversation:join', {
      conversation_id: conversationId
    });
    
    // Load conversation history
    this.loadConversationHistory(conversationId);
    
    // Update UI
    this.updateConversationUI(conversationId);
  }
  
  async loadConversationHistory(conversationId) {
    const response = await fetch(
      `https://api.winkr.com/v1/chat/conversations/${conversationId}/messages`,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      }
    );
    
    if (!response.ok) {
      console.error('Failed to load conversation history');
      return;
    }
    
    const data = await response.json();
    this.messages.set(conversationId, data.messages);
    this.renderMessages(data.messages);
  }
  
  sendMessage() {
    const input = document.getElementById('message-input');
    const content = input.value.trim();
    
    if (!content || !this.currentConversation) {
      return;
    }
    
    this.ws.send('message:send', {
      conversation_id: this.currentConversation,
      content: content,
      type: 'text'
    });
    
    input.value = '';
    this.stopTyping();
  }
  
  handleTyping() {
    if (!this.currentConversation) return;
    
    this.ws.send('typing:start', {
      conversation_id: this.currentConversation
    });
    
    clearTimeout(this.typingTimeout);
    this.typingTimeout = setTimeout(() => {
      this.stopTyping();
    }, 3000);
  }
  
  stopTyping() {
    if (!this.currentConversation) return;
    
    clearTimeout(this.typingTimeout);
    this.ws.send('typing:stop', {
      conversation_id: this.currentConversation
    });
  }
  
  handleNewMessage(data) {
    const messages = this.messages.get(data.conversation_id) || [];
    messages.push(data);
    this.messages.set(data.conversation_id, messages);
    
    if (data.conversation_id === this.currentConversation) {
      this.renderMessages(messages);
      this.markMessageAsViewed(data.id);
    }
    
    this.updateConversationList(data.conversation_id);
  }
  
  handleTypingIndicator(data) {
    if (data.conversation_id === this.currentConversation) {
      this.updateTypingIndicator(data);
    }
  }
  
  handleNewMatch(data) {
    this.showNotification('New Match!', `You matched with ${data.user.name}`);
    this.addMatchToList(data);
  }
  
  markMessageAsViewed(messageId) {
    this.ws.send('message:viewed', {
      message_id: messageId,
      conversation_id: this.currentConversation
    });
  }
  
  // UI rendering methods
  renderConversationsList() {
    const container = document.getElementById('conversations-list');
    container.innerHTML = '';
    
    this.conversations.forEach((conversation, id) => {
      const element = document.createElement('div');
      element.className = 'conversation-item';
      element.setAttribute('data-conversation-id', id);
      element.innerHTML = `
        <img src="${conversation.other_user.avatar_url}" alt="${conversation.other_user.name}">
        <div class="conversation-info">
          <h3>${conversation.other_user.name}</h3>
          <p class="last-message">${conversation.last_message || 'No messages yet'}</p>
        </div>
        <div class="conversation-meta">
          <span class="timestamp">${this.formatTime(conversation.updated_at)}</span>
          ${conversation.unread_count > 0 ? `<span class="unread-count">${conversation.unread_count}</span>` : ''}
        </div>
      `;
      container.appendChild(element);
    });
  }
  
  renderMessages(messages) {
    const container = document.getElementById('messages-container');
    container.innerHTML = '';
    
    messages.forEach(message => {
      const element = document.createElement('div');
      element.className = `message ${message.sender_id === this.currentUser.id ? 'sent' : 'received'}`;
      element.innerHTML = `
        <div class="message-content">${message.content}</div>
        <div class="message-time">${this.formatTime(message.created_at)}</div>
      `;
      container.appendChild(element);
    });
    
    // Scroll to bottom
    container.scrollTop = container.scrollHeight;
  }
  
  updateTypingIndicator(data) {
    const indicator = document.getElementById('typing-indicator');
    if (data.is_typing) {
      indicator.textContent = `${data.user_id} is typing...`;
      indicator.style.display = 'block';
    } else {
      indicator.style.display = 'none';
    }
  }
  
  updateConnectionStatus(status) {
    const statusElement = document.getElementById('connection-status');
    statusElement.textContent = status;
    statusElement.className = `status ${status}`;
  }
  
  showNotification(title, body) {
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(title, {
        body: body,
        icon: '/assets/icons/notification.png'
      });
    }
  }
  
  formatTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) return 'just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  }
  
  showError(message) {
    const errorElement = document.getElementById('error-message');
    errorElement.textContent = message;
    errorElement.style.display = 'block';
    
    setTimeout(() => {
      errorElement.style.display = 'none';
    }, 5000);
  }
}

// Initialize the app
const app = new WinkrChatApp({
  email: 'user@example.com',
  password: 'password123'
});
```

This comprehensive WebSocket integration guide provides everything needed to implement real-time features in applications using the Winkr WebSocket API, from basic connections to advanced features like ephemeral photos and load balancing.