# Web Client Integration Guide

This guide provides comprehensive instructions for integrating the Winkr API into web applications.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Authentication](#authentication)
- [Core Features](#core-features)
- [Real-time Features](#real-time-features)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Prerequisites

### Requirements

- **Modern Browser**: Chrome 80+, Firefox 75+, Safari 13+, Edge 80+
- **JavaScript**: ES6+ support required
- **HTTPS**: All API calls must use HTTPS
- **CORS**: Server must allow cross-origin requests

### Development Tools

- **API Key**: Contact api@winkr.com to request access
- **Text Editor**: VS Code, WebStorm, or similar
- **Browser DevTools**: For debugging and testing
- **Node.js**: For local development and testing

## Setup

### 1. Project Initialization

Create a new web project or integrate into existing project:

```bash
# Using npm
npm init -y
npm install axios

# Using yarn
yarn init -y
yarn add axios
```

### 2. API Client Setup

Create a reusable API client:

```javascript
// src/api/winkr.js
import axios from 'axios';

class WinkrAPI {
  constructor(baseURL = 'https://api.winkr.com/v1') {
    this.baseURL = baseURL;
    this.token = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
    
    // Configure axios defaults
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    // Request interceptor for authentication
    this.client.interceptors.request.use(
      (config) => {
        if (this.token) {
          config.headers.Authorization = `Bearer ${this.token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );
    
    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          await this.handleTokenRefresh();
        }
        return Promise.reject(error);
      }
    );
  }
  
  async handleTokenRefresh() {
    try {
      const response = await this.client.post('/auth/refresh', {
        refresh_token: this.refreshToken
      });
      
      const { access_token, refresh_token } = response.data.data.tokens;
      this.token = access_token;
      this.refreshToken = refresh_token;
      
      localStorage.setItem('access_token', access_token);
      localStorage.setItem('refresh_token', refresh_token);
      
      // Retry original request
      error.config.headers.Authorization = `Bearer ${access_token}`;
      return this.client.request(error.config);
    } catch (refreshError) {
      // Refresh failed, redirect to login
      this.logout();
      window.location.href = '/login';
    }
  }
  
  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    this.token = null;
    this.refreshToken = null;
  }
  
  // Authentication methods
  async register(userData) {
    const response = await this.client.post('/auth/register', userData);
    return response.data;
  }
  
  async login(credentials) {
    const response = await this.client.post('/auth/login', credentials);
    const { access_token, refresh_token } = response.data.data.tokens;
    
    this.token = access_token;
    this.refreshToken = refresh_token;
    
    localStorage.setItem('access_token', access_token);
    localStorage.setItem('refresh_token', refresh_token);
    
    return response.data;
  }
  
  // Profile methods
  async getProfile() {
    const response = await this.client.get('/profile/me');
    return response.data;
  }
  
  async updateProfile(profileData) {
    const response = await this.client.put('/profile/me', profileData);
    return response.data;
  }
  
  // Discovery methods
  async getDiscoveryUsers(params = {}) {
    const response = await this.client.get('/discovery/users', { params });
    return response.data;
  }
  
  async swipe(userId, action) {
    const response = await this.client.post('/discovery/swipe', {
      user_id: userId,
      action
    });
    return response.data;
  }
  
  // Chat methods
  async getConversations(params = {}) {
    const response = await this.client.get('/chat/conversations', { params });
    return response.data;
  }
  
  async sendMessage(conversationId, content) {
    const response = await this.client.post(`/chat/conversations/${conversationId}/messages`, {
      content,
      type: 'text'
    });
    return response.data;
  }
  
  // Photo methods
  async uploadPhoto(file) {
    const formData = new FormData();
    formData.append('photo', file);
    
    const response = await this.client.post('/photos', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
    return response.data;
  }
}

export default new WinkrAPI();
```

### 3. Environment Configuration

Create environment configuration:

```javascript
// src/config/environment.js
const environments = {
  development: {
    apiURL: 'http://localhost:8080/v1',
    wsURL: 'ws://localhost:8080/ws',
    debug: true
  },
  staging: {
    apiURL: 'https://staging-api.winkr.com/v1',
    wsURL: 'wss://staging-api.winkr.com/ws',
    debug: true
  },
  production: {
    apiURL: 'https://api.winkr.com/v1',
    wsURL: 'wss://api.winkr.com/ws',
    debug: false
  }
};

const currentEnv = environments[process.env.NODE_ENV] || environments.development;

export default currentEnv;
```

## Authentication

### 1. Registration Form

Create a user registration form:

```html
<!-- src/components/RegisterForm.html -->
<form id="registerForm">
  <div class="form-group">
    <label for="email">Email</label>
    <input type="email" id="email" name="email" required>
  </div>
  
  <div class="form-group">
    <label for="password">Password</label>
    <input type="password" id="password" name="password" required>
  </div>
  
  <div class="form-group">
    <label for="username">Username</label>
    <input type="text" id="username" name="username" required>
  </div>
  
  <div class="form-group">
    <label for="date_of_birth">Date of Birth</label>
    <input type="date" id="date_of_birth" name="date_of_birth" required>
  </div>
  
  <div class="form-group">
    <label for="gender">Gender</label>
    <select id="gender" name="gender" required>
      <option value="">Select Gender</option>
      <option value="male">Male</option>
      <option value="female">Female</option>
      <option value="other">Other</option>
    </select>
  </div>
  
  <button type="submit">Register</button>
</form>
```

```javascript
// src/components/RegisterForm.js
import WinkrAPI from '../api/winkr.js';

class RegisterForm {
  constructor() {
    this.form = document.getElementById('registerForm');
    this.init();
  }
  
  init() {
    this.form.addEventListener('submit', this.handleSubmit.bind(this));
  }
  
  async handleSubmit(event) {
    event.preventDefault();
    
    const formData = new FormData(this.form);
    const userData = Object.fromEntries(formData.entries());
    
    try {
      this.showLoading(true);
      const response = await WinkrAPI.register(userData);
      
      if (response.success) {
        this.showSuccess('Registration successful!');
        this.redirectToDashboard();
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Registration failed. Please try again.');
      console.error('Registration error:', error);
    } finally {
      this.showLoading(false);
    }
  }
  
  showLoading(show) {
    const button = this.form.querySelector('button[type="submit"]');
    button.disabled = show;
    button.textContent = show ? 'Registering...' : 'Register';
  }
  
  showError(message) {
    // Show error message to user
    const errorElement = document.getElementById('error-message');
    if (errorElement) {
      errorElement.textContent = message;
      errorElement.style.display = 'block';
    }
  }
  
  showSuccess(message) {
    // Show success message to user
    const successElement = document.getElementById('success-message');
    if (successElement) {
      successElement.textContent = message;
      successElement.style.display = 'block';
    }
  }
  
  redirectToDashboard() {
    setTimeout(() => {
      window.location.href = '/dashboard';
    }, 2000);
  }
}

// Initialize form when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  new RegisterForm();
});
```

### 2. Login Form

Create a login form:

```javascript
// src/components/LoginForm.js
import WinkrAPI from '../api/winkr.js';

class LoginForm {
  constructor() {
    this.form = document.getElementById('loginForm');
    this.init();
  }
  
  init() {
    this.form.addEventListener('submit', this.handleSubmit.bind(this));
  }
  
  async handleSubmit(event) {
    event.preventDefault();
    
    const formData = new FormData(this.form);
    const credentials = Object.fromEntries(formData.entries());
    
    try {
      this.showLoading(true);
      const response = await WinkrAPI.login(credentials);
      
      if (response.success) {
        this.showSuccess('Login successful!');
        this.redirectToDashboard();
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Login failed. Please try again.');
      console.error('Login error:', error);
    } finally {
      this.showLoading(false);
    }
  }
  
  // ... similar methods as RegisterForm
}

document.addEventListener('DOMContentLoaded', () => {
  new LoginForm();
});
```

## Core Features

### 1. User Profile Management

Create profile management component:

```javascript
// src/components/ProfileManager.js
import WinkrAPI from '../api/winkr.js';

class ProfileManager {
  constructor() {
    this.profileForm = document.getElementById('profileForm');
    this.avatarInput = document.getElementById('avatar-input');
    this.init();
  }
  
  init() {
    this.loadProfile();
    this.profileForm.addEventListener('submit', this.handleProfileUpdate.bind(this));
    this.avatarInput.addEventListener('change', this.handleAvatarUpload.bind(this));
  }
  
  async loadProfile() {
    try {
      const response = await WinkrAPI.getProfile();
      if (response.success) {
        this.populateForm(response.data);
      }
    } catch (error) {
      console.error('Failed to load profile:', error);
    }
  }
  
  populateForm(profile) {
    document.getElementById('first_name').value = profile.first_name || '';
    document.getElementById('last_name').value = profile.last_name || '';
    document.getElementById('bio').value = profile.bio || '';
    
    if (profile.avatar_url) {
      document.getElementById('avatar-preview').src = profile.avatar_url;
    }
  }
  
  async handleProfileUpdate(event) {
    event.preventDefault();
    
    const formData = new FormData(this.profileForm);
    const profileData = Object.fromEntries(formData.entries());
    
    try {
      this.showLoading(true);
      const response = await WinkrAPI.updateProfile(profileData);
      
      if (response.success) {
        this.showSuccess('Profile updated successfully!');
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Failed to update profile. Please try again.');
      console.error('Profile update error:', error);
    } finally {
      this.showLoading(false);
    }
  }
  
  async handleAvatarUpload(event) {
    const file = event.target.files[0];
    if (!file) return;
    
    // Validate file
    if (!this.validateImageFile(file)) {
      return;
    }
    
    try {
      this.showLoading(true);
      const response = await WinkrAPI.uploadPhoto(file);
      
      if (response.success) {
        this.showSuccess('Avatar uploaded successfully!');
        document.getElementById('avatar-preview').src = response.data.url;
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Failed to upload avatar. Please try again.');
      console.error('Avatar upload error:', error);
    } finally {
      this.showLoading(false);
    }
  }
  
  validateImageFile(file) {
    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp'];
    const maxSize = 5 * 1024 * 1024; // 5MB
    
    if (!allowedTypes.includes(file.type)) {
      this.showError('Please select a JPEG, PNG, or WebP image.');
      return false;
    }
    
    if (file.size > maxSize) {
      this.showError('Image size must be less than 5MB.');
      return false;
    }
    
    return true;
  }
}

document.addEventListener('DOMContentLoaded', () => {
  new ProfileManager();
});
```

### 2. Discovery and Matching

Create discovery component:

```javascript
// src/components/Discovery.js
import WinkrAPI from '../api/winkr.js';

class Discovery {
  constructor() {
    this.container = document.getElementById('discovery-container');
    this.currentUserId = null;
    this.currentIndex = 0;
    this.users = [];
    this.init();
  }
  
  init() {
    this.loadUsers();
    this.setupEventListeners();
  }
  
  async loadUsers() {
    try {
      const response = await WinkrAPI.getDiscoveryUsers({
        limit: 20,
        offset: 0
      });
      
      if (response.success) {
        this.users = response.data.data.users;
        this.showCurrentUser();
      }
    } catch (error) {
      console.error('Failed to load users:', error);
    }
  }
  
  showCurrentUser() {
    if (this.currentIndex >= this.users.length) {
      this.showNoMoreUsers();
      return;
    }
    
    const user = this.users[this.currentIndex];
    this.renderUserCard(user);
  }
  
  renderUserCard(user) {
    this.container.innerHTML = `
      <div class="user-card">
        <div class="user-avatar">
          <img src="${user.photos?.[0]?.url || '/default-avatar.png'}" alt="${user.username}">
        </div>
        <div class="user-info">
          <h3>${user.first_name} ${user.last_name}, ${user.age}</h3>
          <p>${user.bio || 'No bio available'}</p>
          <div class="user-location">
            📍 ${user.location?.city || 'Unknown'}
          </div>
        </div>
        <div class="user-actions">
          <button class="btn-pass" onclick="discovery.handleSwipe('${user.id}', 'pass')">
            ✕ Pass
          </button>
          <button class="btn-like" onclick="discovery.handleSwipe('${user.id}', 'like')">
            ❤ Like
          </button>
        </div>
      </div>
    `;
  }
  
  async handleSwipe(userId, action) {
    try {
      const response = await WinkrAPI.swipe(userId, action);
      
      if (response.success) {
        if (response.data.data.is_new_match) {
          this.showMatchNotification(response.data.data.match);
        }
        this.nextUser();
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Failed to process swipe. Please try again.');
      console.error('Swipe error:', error);
    }
  }
  
  nextUser() {
    this.currentIndex++;
    this.showCurrentUser();
  }
  
  showMatchNotification(match) {
    const notification = document.createElement('div');
    notification.className = 'match-notification';
    notification.innerHTML = `
      <div class="match-content">
        <h3>It's a Match! 🎉</h3>
        <p>You matched with ${match.matched_user.first_name}</p>
        <img src="${match.matched_user.photos?.[0]?.url || '/default-avatar.png'}" alt="${match.matched_user.username}">
        <button onclick="discovery.startConversation('${match.matched_user.id}')">
          Send Message
        </button>
      </div>
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
      notification.remove();
    }, 5000);
  }
  
  async startConversation(userId) {
    try {
      const response = await WinkrAPI.sendMessage(userId, 'Hi! We matched! 👋');
      
      if (response.success) {
        window.location.href = `/chat/${userId}`;
      }
    } catch (error) {
      this.showError('Failed to start conversation. Please try again.');
      console.error('Conversation start error:', error);
    }
  }
  
  showNoMoreUsers() {
    this.container.innerHTML = `
      <div class="no-more-users">
        <h3>No more users</h3>
        <p>Check back later for more potential matches!</p>
        <button onclick="discovery.loadUsers()">Refresh</button>
      </div>
    `;
  }
  
  setupEventListeners() {
    // Keyboard navigation
    document.addEventListener('keydown', (event) => {
      if (event.key === 'ArrowLeft') {
        this.handleSwipe(this.users[this.currentIndex]?.id, 'pass');
      } else if (event.key === 'ArrowRight') {
        this.handleSwipe(this.users[this.currentIndex]?.id, 'like');
      }
    });
  }
}

// Make discovery globally accessible
window.discovery = new Discovery();
```

## Real-time Features

### 1. WebSocket Integration

Create WebSocket manager for real-time messaging:

```javascript
// src/services/WebSocketManager.js
import config from '../config/environment.js';

class WebSocketManager {
  constructor() {
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectInterval = 5000;
    this.heartbeatInterval = 30000;
    this.heartbeatTimer = null;
    this.eventHandlers = {};
  }
  
  connect(token) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }
    
    const wsURL = `${config.wsURL}?token=${token}`;
    this.ws = new WebSocket(wsURL);
    
    this.ws.onopen = this.handleOpen.bind(this);
    this.ws.onmessage = this.handleMessage.bind(this);
    this.ws.onclose = this.handleClose.bind(this);
    this.ws.onerror = this.handleError.bind(this);
  }
  
  handleOpen() {
    console.log('WebSocket connected');
    this.reconnectAttempts = 0;
    this.startHeartbeat();
    
    // Join conversation rooms
    this.joinConversations();
  }
  
  handleMessage(event) {
    try {
      const data = JSON.parse(event.data);
      const { event: eventType, data: eventData } = data;
      
      if (this.eventHandlers[eventType]) {
        this.eventHandlers[eventType](eventData);
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }
  
  handleClose(event) {
    console.log('WebSocket disconnected:', event);
    this.stopHeartbeat();
    
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.scheduleReconnect();
    }
  }
  
  handleError(error) {
    console.error('WebSocket error:', error);
  }
  
  scheduleReconnect() {
    this.reconnectAttempts++;
    setTimeout(() => {
      this.connect(this.getToken());
    }, this.reconnectInterval * this.reconnectAttempts);
  }
  
  startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ event: 'ping' }));
      }
    }, this.heartbeatInterval);
  }
  
  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
  
  send(event, data) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ event, data }));
    }
  }
  
  on(eventType, handler) {
    if (!this.eventHandlers[eventType]) {
      this.eventHandlers[eventType] = [];
    }
    this.eventHandlers[eventType].push(handler);
  }
  
  off(eventType, handler) {
    if (this.eventHandlers[eventType]) {
      const index = this.eventHandlers[eventType].indexOf(handler);
      if (index > -1) {
        this.eventHandlers[eventType].splice(index, 1);
      }
    }
  }
  
  getToken() {
    return localStorage.getItem('access_token');
  }
  
  async joinConversations() {
    try {
      const response = await WinkrAPI.getConversations();
      if (response.success) {
        response.data.data.conversations.forEach(conversation => {
          this.send('conversation:join', {
            conversation_id: conversation.id
          });
        });
      }
    } catch (error) {
      console.error('Failed to join conversations:', error);
    }
  }
  
  disconnect() {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export default new WebSocketManager();
```

### 2. Chat Component

Create real-time chat component:

```javascript
// src/components/Chat.js
import WinkrAPI from '../api/winkr.js';
import WebSocketManager from '../services/WebSocketManager.js';

class Chat {
  constructor(conversationId) {
    this.conversationId = conversationId;
    this.container = document.getElementById('chat-container');
    this.messagesContainer = document.getElementById('messages-container');
    this.messageInput = document.getElementById('message-input');
    this.sendButton = document.getElementById('send-button');
    this.typingIndicator = document.getElementById('typing-indicator');
    this.init();
  }
  
  init() {
    this.loadMessages();
    this.setupEventListeners();
    this.setupWebSocketListeners();
  }
  
  async loadMessages() {
    try {
      const response = await WinkrAPI.getMessages(this.conversationId);
      if (response.success) {
        this.renderMessages(response.data.data.messages);
      }
    } catch (error) {
      console.error('Failed to load messages:', error);
    }
  }
  
  renderMessages(messages) {
    this.messagesContainer.innerHTML = '';
    messages.forEach(message => {
      this.renderMessage(message);
    });
    this.scrollToBottom();
  }
  
  renderMessage(message) {
    const messageElement = document.createElement('div');
    messageElement.className = `message ${message.sender_id === this.getCurrentUserId() ? 'sent' : 'received'}`;
    messageElement.innerHTML = `
      <div class="message-content">${this.escapeHtml(message.content)}</div>
      <div class="message-time">${this.formatTime(message.created_at)}</div>
    `;
    
    this.messagesContainer.appendChild(messageElement);
  }
  
  setupEventListeners() {
    this.sendButton.addEventListener('click', this.handleSendMessage.bind(this));
    this.messageInput.addEventListener('keypress', (event) => {
      if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        this.handleSendMessage();
      }
    });
    
    // Typing indicators
    this.messageInput.addEventListener('input', this.handleTyping.bind(this));
  }
  
  setupWebSocketListeners() {
    WebSocketManager.on('message:new', this.handleNewMessage.bind(this));
    WebSocketManager.on('message:viewed', this.handleMessageViewed.bind(this));
    WebSocketManager.on('typing:indicator', this.handleTypingIndicator.bind(this));
  }
  
  async handleSendMessage() {
    const content = this.messageInput.value.trim();
    if (!content) return;
    
    try {
      const response = await WinkrAPI.sendMessage(this.conversationId, content);
      
      if (response.success) {
        this.renderMessage(response.data.data);
        this.messageInput.value = '';
        this.scrollToBottom();
      } else {
        this.showError(response.error.message);
      }
    } catch (error) {
      this.showError('Failed to send message. Please try again.');
      console.error('Send message error:', error);
    }
  }
  
  handleNewMessage(data) {
    if (data.message.conversation_id === this.conversationId) {
      this.renderMessage(data.message);
      this.scrollToBottom();
      
      // Mark as read
      WebSocketManager.send('message:read', {
        conversation_id: this.conversationId,
        message_id: data.message.id
      });
    }
  }
  
  handleMessageViewed(data) {
    if (data.conversation_id === this.conversationId) {
      const messageElement = document.querySelector(`[data-message-id="${data.message_id}"]`);
      if (messageElement) {
        messageElement.classList.add('viewed');
      }
    }
  }
  
  handleTypingIndicator(data) {
    if (data.conversation_id === this.conversationId) {
      if (data.is_typing) {
        this.typingIndicator.textContent = `${data.username} is typing...`;
        this.typingIndicator.style.display = 'block';
      } else {
        this.typingIndicator.style.display = 'none';
      }
    }
  }
  
  handleTyping() {
    WebSocketManager.send('typing:start', {
      conversation_id: this.conversationId
    });
    
    clearTimeout(this.typingTimeout);
    this.typingTimeout = setTimeout(() => {
      WebSocketManager.send('typing:stop', {
        conversation_id: this.conversationId
      });
    }, 3000);
  }
  
  scrollToBottom() {
    this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
  }
  
  getCurrentUserId() {
    return localStorage.getItem('user_id');
  }
  
  formatTime(timestamp) {
    return new Date(timestamp).toLocaleTimeString();
  }
  
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
  
  showError(message) {
    // Show error message to user
    const errorElement = document.getElementById('chat-error');
    if (errorElement) {
      errorElement.textContent = message;
      errorElement.style.display = 'block';
      setTimeout(() => {
        errorElement.style.display = 'none';
      }, 3000);
    }
  }
}

// Initialize chat when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  const conversationId = window.location.pathname.split('/').pop();
  if (conversationId) {
    new Chat(conversationId);
  }
});
```

## Error Handling

### 1. Global Error Handler

Create global error handling:

```javascript
// src/utils/ErrorHandler.js
class ErrorHandler {
  static handle(error, context = {}) {
    console.error('Application Error:', { error, context });
    
    // Show user-friendly message
    const userMessage = this.getUserMessage(error);
    this.showUserMessage(userMessage);
    
    // Report to error tracking service
    this.reportError(error, context);
  }
  
  static getUserMessage(error) {
    const errorMessages = {
      'TOKEN_EXPIRED': 'Your session has expired. Please log in again.',
      'TOO_MANY_REQUESTS': 'You\'ve made too many requests. Please wait a moment.',
      'VALIDATION_ERROR': 'Please check your input and try again.',
      'NETWORK_ERROR': 'Network connection error. Please check your internet connection.',
      'DEFAULT': 'An unexpected error occurred. Please try again.'
    };
    
    return errorMessages[error.code] || errorMessages['DEFAULT'];
  }
  
  static showUserMessage(message) {
    const notification = document.createElement('div');
    notification.className = 'error-notification';
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
      notification.remove();
    }, 5000);
  }
  
  static reportError(error, context) {
    // Send error to tracking service
    fetch('/api/errors', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        error: {
          code: error.code,
          message: error.message,
          stack: error.stack
        },
        context: {
          userAgent: navigator.userAgent,
          url: window.location.href,
          timestamp: new Date().toISOString(),
          ...context
        }
      })
    }).catch(console.error);
  }
}

// Global error handler
window.addEventListener('error', (event) => {
  ErrorHandler.handle(event.error, { type: 'javascript' });
});

// Unhandled promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
  ErrorHandler.handle(event.reason, { type: 'promise' });
});
```

## Best Practices

### 1. Security

- **HTTPS Only**: Always use HTTPS for API calls
- **Token Storage**: Store tokens securely (httpOnly cookies recommended)
- **Input Validation**: Validate all user inputs
- **XSS Prevention**: Escape HTML content before rendering
- **CSRF Protection**: Use CSRF tokens for state-changing requests

### 2. Performance

- **Lazy Loading**: Load data only when needed
- **Image Optimization**: Compress and optimize images
- **Caching**: Implement appropriate caching strategies
- **Bundle Optimization**: Minimize and bundle JavaScript files
- **CDN Usage**: Use CDN for static assets

### 3. User Experience

- **Loading States**: Show loading indicators during API calls
- **Error Messages**: Provide clear, actionable error messages
- **Offline Support**: Handle offline scenarios gracefully
- **Responsive Design**: Ensure mobile-friendly interface
- **Accessibility**: Follow WCAG guidelines for accessibility

### 4. Code Organization

- **Modular Structure**: Organize code into logical modules
- **Component-Based**: Use component-based architecture
- **State Management**: Implement consistent state management
- **Error Boundaries**: Use error boundaries for error isolation
- **Testing**: Write comprehensive tests for all components

## Examples

### 1. Complete Application Structure

```
src/
├── api/
│   └── winkr.js              # API client
├── components/
│   ├── RegisterForm.js         # Registration component
│   ├── LoginForm.js           # Login component
│   ├── ProfileManager.js       # Profile management
│   ├── Discovery.js           # Discovery/matching
│   └── Chat.js               # Chat component
├── services/
│   └── WebSocketManager.js    # WebSocket management
├── utils/
│   ├── ErrorHandler.js        # Error handling
│   └── helpers.js            # Utility functions
├── config/
│   └── environment.js        # Environment configuration
└── styles/
    ├── main.css              # Main styles
    └── components.css        # Component styles
```

### 2. HTML Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Winkr - Dating App</title>
    <link rel="stylesheet" href="/styles/main.css">
    <link rel="stylesheet" href="/styles/components.css">
</head>
<body>
    <div id="app">
        <!-- Navigation -->
        <nav class="navbar">
            <div class="nav-brand">Winkr</div>
            <div class="nav-menu">
                <a href="/discovery">Discover</a>
                <a href="/matches">Matches</a>
                <a href="/chat">Messages</a>
                <a href="/profile">Profile</a>
                <a href="/logout" id="logout-btn">Logout</a>
            </div>
        </nav>
        
        <!-- Main Content -->
        <main class="main-content">
            <!-- Content will be loaded here -->
        </main>
        
        <!-- Error Notifications -->
        <div id="error-notification" class="notification error" style="display: none;"></div>
        <div id="success-notification" class="notification success" style="display: none;"></div>
    </div>
    
    <!-- Scripts -->
    <script src="/js/app.js"></script>
</body>
</html>
```

### 3. CSS Styles

```css
/* src/styles/components.css */
.user-card {
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    margin: 20px auto;
    max-width: 400px;
}

.user-avatar img {
    width: 100%;
    height: 300px;
    object-fit: cover;
}

.user-info {
    padding: 20px;
}

.user-actions {
    display: flex;
    justify-content: space-between;
    padding: 20px;
}

.btn-like, .btn-pass {
    flex: 1;
    padding: 12px 24px;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-like {
    background: #ff6b6b;
    color: white;
    margin-right: 10px;
}

.btn-pass {
    background: #6c757d;
    color: white;
}

.btn-like:hover {
    background: #ff5252;
    transform: scale(1.05);
}

.btn-pass:hover {
    background: #5a6268;
    transform: scale(1.05);
}

.message {
    margin: 10px 0;
    padding: 12px 16px;
    border-radius: 18px;
    max-width: 70%;
    word-wrap: break-word;
}

.message.sent {
    background: #007bff;
    color: white;
    margin-left: auto;
    text-align: right;
}

.message.received {
    background: #f1f3f4;
    color: #333;
    margin-right: auto;
}

.notification {
    position: fixed;
    top: 20px;
    right: 20px;
    padding: 16px 24px;
    border-radius: 8px;
    color: white;
    font-weight: 500;
    z-index: 1000;
    animation: slideIn 0.3s ease-out;
}

.notification.error {
    background: #dc3545;
}

.notification.success {
    background: #28a745;
}

@keyframes slideIn {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}
```

This comprehensive web integration guide provides everything needed to successfully integrate the Winkr API into web applications, from basic setup to advanced real-time features.