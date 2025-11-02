# WinKr API - Testing Guide

## Overview

This comprehensive testing guide provides everything you need to effectively test the WinKr API, from basic endpoint testing to advanced integration and performance testing. It includes testing strategies, tools, examples, and best practices to ensure your integration works reliably.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Testing Environments](#testing-environments)
3. [Testing Tools](#testing-tools)
4. [Unit Testing](#unit-testing)
5. [Integration Testing](#integration-testing)
6. [End-to-End Testing](#end-to-end-testing)
7. [Performance Testing](#performance-testing)
8. [Security Testing](#security-testing)
9. [Load Testing](#load-testing)
10. [Postman Collections](#postman-collections)
11. [Automated Testing](#automated-testing)
12. [Test Data Management](#test-data-management)
13. [Continuous Integration](#continuous-integration)
14. [Testing Best Practices](#testing-best-practices)

## Testing Strategy

### Testing Pyramid

```mermaid
graph TB
    A[E2E Tests - 10%] --> B[Integration Tests - 20%]
    B --> C[Unit Tests - 70%]
    
    D[Performance Tests] --> A
    E[Security Tests] --> B
    F[Load Tests] --> A
```

### Test Categories

| Category | Purpose | Tools | Frequency |
|----------|---------|-------|-----------|
| **Unit Tests** | Test individual functions/methods | Jest, Mocha, Go test | Every commit |
| **Integration Tests** | Test API endpoints and interactions | Supertest, Postman | Every PR |
| **E2E Tests** | Test complete user workflows | Cypress, Playwright | Daily |
| **Performance Tests** | Test response times and resource usage | Artillery, k6 | Weekly |
| **Security Tests** | Test for vulnerabilities | OWASP ZAP, Burp Suite | Monthly |
| **Load Tests** | Test system under high load | JMeter, Gatling | Before releases |

### Test Coverage Goals

- **Unit Tests**: 90%+ code coverage
- **Integration Tests**: 80%+ endpoint coverage
- **E2E Tests**: 100% critical user path coverage
- **Security Tests**: 100% OWASP Top 10 coverage

## Testing Environments

### Environment Configuration

```javascript
// Environment configuration for testing
const environments = {
  development: {
    name: 'Development',
    baseURL: 'http://localhost:8080/api/v1',
    database: 'winkr_dev',
    redis: 'redis://localhost:6379/0',
    logging: 'debug',
    features: {
      experimental: true,
      debugMode: true
    }
  },
  
  staging: {
    name: 'Staging',
    baseURL: 'https://staging-api.winkr.com/api/v1',
    database: 'winkr_staging',
    redis: 'redis://staging-redis.winkr.com:6379/0',
    logging: 'info',
    features: {
      experimental: false,
      debugMode: false
    }
  },
  
  production: {
    name: 'Production',
    baseURL: 'https://api.winkr.com/api/v1',
    database: 'winkr_prod',
    redis: 'redis://prod-redis.winkr.com:6379/0',
    logging: 'error',
    features: {
      experimental: false,
      debugMode: false
    }
  }
};

// Environment switcher
class EnvironmentManager {
  constructor(environment = 'development') {
    this.currentEnvironment = environments[environment];
    if (!this.currentEnvironment) {
      throw new Error(`Unknown environment: ${environment}`);
    }
  }

  getConfig() {
    return this.currentEnvironment;
  }

  getBaseURL() {
    return this.currentEnvironment.baseURL;
  }

  getDatabaseConfig() {
    return {
      name: this.currentEnvironment.database,
      redis: this.currentEnvironment.redis
    };
  }

  isFeatureEnabled(feature) {
    return this.currentEnvironment.features[feature] || false;
  }

  switchEnvironment(environment) {
    this.currentEnvironment = environments[environment];
    if (!this.currentEnvironment) {
      throw new Error(`Unknown environment: ${environment}`);
    }
  }
}
```

### Test Data Setup

```javascript
// Test data factory for creating test users and data
class TestDataFactory {
  constructor() {
    this.users = new Map();
    this.photos = new Map();
    this.matches = new Map();
    this.messages = new Map();
  }

  // Create test user
  async createTestUser(userData = {}) {
    const defaultUser = {
      email: `test_${Date.now()}@example.com`,
      password: 'TestPassword123!',
      firstName: 'Test',
      lastName: 'User',
      age: 25,
      gender: 'male',
      location: {
        latitude: 40.7128,
        longitude: -74.0060,
        city: 'New York',
        country: 'USA'
      },
      bio: 'Test user for API testing',
      interests: ['testing', 'api', 'development'],
      ...userData
    };

    const response = await fetch(`${this.getBaseURL()}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(defaultUser)
    });

    if (!response.ok) {
      throw new Error(`Failed to create test user: ${response.statusText}`);
    }

    const user = await response.json();
    this.users.set(user.id, user);
    
    return user;
  }

  // Create test photo
  async createTestPhoto(userId, photoData = {}) {
    const defaultPhoto = {
      userId: userId,
      url: 'https://example.com/test-photo.jpg',
      caption: 'Test photo',
      isPrimary: false,
      ...photoData
    };

    const response = await fetch(`${this.getBaseURL()}/photos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken(userId)}`
      },
      body: JSON.stringify(defaultPhoto)
    });

    if (!response.ok) {
      throw new Error(`Failed to create test photo: ${response.statusText}`);
    }

    const photo = await response.json();
    this.photos.set(photo.id, photo);
    
    return photo;
  }

  // Create test match
  async createTestMatch(user1Id, user2Id) {
    const matchData = {
      user1Id: user1Id,
      user2Id: user2Id,
      initiatedBy: user1Id
    };

    const response = await fetch(`${this.getBaseURL()}/matches`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken(user1Id)}`
      },
      body: JSON.stringify(matchData)
    });

    if (!response.ok) {
      throw new Error(`Failed to create test match: ${response.statusText}`);
    }

    const match = await response.json();
    this.matches.set(match.id, match);
    
    return match;
  }

  // Create test message
  async createTestMessage(matchId, senderId, content) {
    const messageData = {
      matchId: matchId,
      content: content,
      type: 'text'
    };

    const response = await fetch(`${this.getBaseURL()}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.getAuthToken(senderId)}`
      },
      body: JSON.stringify(messageData)
    });

    if (!response.ok) {
      throw new Error(`Failed to create test message: ${response.statusText}`);
    }

    const message = await response.json();
    this.messages.set(message.id, message);
    
    return message;
  }

  // Clean up test data
  async cleanup() {
    // Delete all created test data
    for (const [userId, user] of this.users) {
      try {
        await fetch(`${this.getBaseURL()}/users/${userId}`, {
          method: 'DELETE',
          headers: { 'Authorization': `Bearer ${this.getAuthToken(userId)}` }
        });
      } catch (error) {
        console.warn(`Failed to delete user ${userId}:`, error.message);
      }
    }

    // Clear local storage
    this.users.clear();
    this.photos.clear();
    this.matches.clear();
    this.messages.clear();
  }

  getBaseURL() {
    return process.env.API_BASE_URL || 'http://localhost:8080/api/v1';
  }

  getAuthToken(userId) {
    const user = this.users.get(userId);
    return user ? user.token : null;
  }
}
```

## Testing Tools

### HTTP Client Testing

```javascript
// HTTP client wrapper for testing
class TestHTTPClient {
  constructor(baseURL, defaultHeaders = {}) {
    this.baseURL = baseURL;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      'User-Agent': 'WinKr-API-Test/1.0',
      ...defaultHeaders
    };
    this.interceptors = [];
  }

  // Add request interceptor
  addRequestInterceptor(interceptor) {
    this.interceptors.push({ type: 'request', fn: interceptor });
  }

  // Add response interceptor
  addResponseInterceptor(interceptor) {
    this.interceptors.push({ type: 'response', fn: interceptor });
  }

  // Make HTTP request with interceptors
  async request(method, endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    
    let requestOptions = {
      method: method.toUpperCase(),
      headers: { ...this.defaultHeaders, ...options.headers },
      body: options.body
    };

    // Apply request interceptors
    for (const interceptor of this.interceptors.filter(i => i.type === 'request')) {
      requestOptions = await interceptor.fn(requestOptions);
    }

    let response;
    try {
      response = await fetch(url, requestOptions);
    } catch (error) {
      throw new Error(`Network error: ${error.message}`);
    }

    // Apply response interceptors
    for (const interceptor of this.interceptors.filter(i => i.type === 'response')) {
      response = await interceptor.fn(response);
    }

    return response;
  }

  // Convenience methods
  async get(endpoint, options = {}) {
    return this.request('GET', endpoint, options);
  }

  async post(endpoint, data, options = {}) {
    return this.request('POST', endpoint, {
      ...options,
      body: JSON.stringify(data)
    });
  }

  async put(endpoint, data, options = {}) {
    return this.request('PUT', endpoint, {
      ...options,
      body: JSON.stringify(data)
    });
  }

  async delete(endpoint, options = {}) {
    return this.request('DELETE', endpoint, options);
  }

  async patch(endpoint, data, options = {}) {
    return this.request('PATCH', endpoint, {
      ...options,
      body: JSON.stringify(data)
    });
  }
}

// Test assertion helpers
class TestAssertions {
  static assertStatus(response, expectedStatus) {
    if (response.status !== expectedStatus) {
      throw new Error(`Expected status ${expectedStatus}, got ${response.status}`);
    }
  }

  static assertJSON(response) {
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      throw new Error('Expected JSON response');
    }
  }

  static async assertValidJSON(response) {
    this.assertJSON(response);
    try {
      return await response.json();
    } catch (error) {
      throw new Error(`Invalid JSON response: ${error.message}`);
    }
  }

  static assertContains(data, expectedValue) {
    const dataString = JSON.stringify(data);
    if (!dataString.includes(JSON.stringify(expectedValue))) {
      throw new Error(`Expected data to contain ${JSON.stringify(expectedValue)}`);
    }
  }

  static assertSchema(data, schema) {
    const errors = this.validateSchema(data, schema);
    if (errors.length > 0) {
      throw new Error(`Schema validation failed: ${errors.join(', ')}`);
    }
  }

  static validateSchema(data, schema) {
    const errors = [];

    for (const [key, type] of Object.entries(schema)) {
      if (!(key in data)) {
        errors.push(`Missing required field: ${key}`);
        continue;
      }

      const actualType = typeof data[key];
      if (actualType !== type) {
        errors.push(`Field ${key} expected ${type}, got ${actualType}`);
      }
    }

    return errors;
  }

  static assertTimeout(promise, timeoutMs) {
    return Promise.race([
      promise,
      new Promise((_, reject) => 
        setTimeout(() => reject(new Error(`Timeout after ${timeoutMs}ms`)), timeoutMs)
      )
    ]);
  }
}
```

### Test Runner

```javascript
// Test runner framework
class TestRunner {
  constructor() {
    this.tests = [];
    this.beforeEachCallbacks = [];
    this.afterEachCallbacks = [];
    this.results = {
      passed: 0,
      failed: 0,
      skipped: 0,
      total: 0,
      errors: []
    };
  }

  // Define test case
  test(name, testFn, options = {}) {
    this.tests.push({
      name,
      fn: testFn,
      timeout: options.timeout || 5000,
      skip: options.skip || false,
      only: options.only || false
    });
  }

  // Run before each test
  beforeEach(callback) {
    this.beforeEachCallbacks.push(callback);
  }

  // Run after each test
  afterEach(callback) {
    this.afterEachCallbacks.push(callback);
  }

  // Run all tests
  async run() {
    console.log('🧪 Starting test run...\n');

    // Filter tests with 'only' flag
    const testsToRun = this.tests.filter(test => test.only) || this.tests;

    for (const test of testsToRun) {
      if (test.skip) {
        console.log(`⏭️  SKIP: ${test.name}`);
        this.results.skipped++;
        continue;
      }

      try {
        console.log(`🔄 RUNNING: ${test.name}`);
        
        // Run before each callbacks
        for (const callback of this.beforeEachCallbacks) {
          await callback();
        }

        // Run test with timeout
        await TestAssertions.assertTimeout(test.fn(), test.timeout);
        
        console.log(`✅ PASS: ${test.name}`);
        this.results.passed++;
      } catch (error) {
        console.log(`❌ FAIL: ${test.name}`);
        console.log(`   Error: ${error.message}`);
        this.results.failed++;
        this.results.errors.push({
          test: test.name,
          error: error.message,
          stack: error.stack
        });
      } finally {
        // Run after each callbacks
        for (const callback of this.afterEachCallbacks) {
          try {
            await callback();
          } catch (error) {
            console.warn(`AfterEach callback failed: ${error.message}`);
          }
        }
      }
    }

    this.results.total = this.tests.length;
    this.printResults();
  }

  printResults() {
    console.log('\n📊 Test Results:');
    console.log(`   Total: ${this.results.total}`);
    console.log(`   Passed: ${this.results.passed}`);
    console.log(`   Failed: ${this.results.failed}`);
    console.log(`   Skipped: ${this.results.skipped}`);

    if (this.results.errors.length > 0) {
      console.log('\n❌ Errors:');
      this.results.errors.forEach(error => {
        console.log(`   ${error.test}: ${error.error}`);
      });
    }

    const successRate = ((this.results.passed / this.results.total) * 100).toFixed(1);
    console.log(`\n🎯 Success Rate: ${successRate}%`);
  }
}
```

## Unit Testing

### API Endpoint Unit Tests

```javascript
// Unit tests for authentication endpoints
class AuthEndpointTests {
  constructor() {
    this.client = new TestHTTPClient('http://localhost:8080/api/v1');
    this.testData = new TestDataFactory();
  }

  async runAllTests() {
    const runner = new TestRunner();

    // Setup and teardown
    runner.beforeEach(async () => {
      // Clean up any existing test data
      await this.testData.cleanup();
    });

    runner.afterEach(async () => {
      // Clean up after each test
      await this.testData.cleanup();
    });

    // Test user registration
    runner.test('User registration with valid data', async () => {
      const userData = {
        email: 'test@example.com',
        password: 'Password123!',
        firstName: 'Test',
        lastName: 'User',
        age: 25,
        gender: 'male'
      };

      const response = await this.client.post('/auth/register', userData);
      TestAssertions.assertStatus(response, 201);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        id: 'string',
        email: 'string',
        token: 'string'
      });

      TestAssertions.assertContains(data, { email: userData.email });
    });

    runner.test('User registration with invalid email', async () => {
      const userData = {
        email: 'invalid-email',
        password: 'Password123!',
        firstName: 'Test',
        lastName: 'User',
        age: 25,
        gender: 'male'
      };

      const response = await this.client.post('/auth/register', userData);
      TestAssertions.assertStatus(response, 400);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        error: 'string',
        message: 'string'
      });
    });

    runner.test('User registration with weak password', async () => {
      const userData = {
        email: 'test@example.com',
        password: '123',
        firstName: 'Test',
        lastName: 'User',
        age: 25,
        gender: 'male'
      };

      const response = await this.client.post('/auth/register', userData);
      TestAssertions.assertStatus(response, 400);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertContains(data, { error: 'weak_password' });
    });

    // Test user login
    runner.test('User login with valid credentials', async () => {
      // First register a user
      const user = await this.testData.createTestUser();

      const loginData = {
        email: user.email,
        password: 'TestPassword123!'
      };

      const response = await this.client.post('/auth/login', loginData);
      TestAssertions.assertStatus(response, 200);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        token: 'string',
        refreshToken: 'string',
        user: 'object'
      });

      TestAssertions.assertContains(data.user, { email: user.email });
    });

    runner.test('User login with invalid credentials', async () => {
      const loginData = {
        email: 'nonexistent@example.com',
        password: 'wrongpassword'
      };

      const response = await this.client.post('/auth/login', loginData);
      TestAssertions.assertStatus(response, 401);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        error: 'string',
        message: 'string'
      });
    });

    // Test token refresh
    runner.test('Token refresh with valid refresh token', async () => {
      const user = await this.testData.createTestUser();
      
      const refreshData = {
        refreshToken: user.refreshToken
      };

      const response = await this.client.post('/auth/refresh', refreshData);
      TestAssertions.assertStatus(response, 200);
      
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        token: 'string',
        refreshToken: 'string'
      });
    });

    runner.test('Token refresh with invalid refresh token', async () => {
      const refreshData = {
        refreshToken: 'invalid-refresh-token'
      };

      const response = await this.client.post('/auth/refresh', refreshData);
      TestAssertions.assertStatus(response, 401);
    });

    await runner.run();
  }
}

// Unit tests for profile endpoints
class ProfileEndpointTests {
  constructor() {
    this.client = new TestHTTPClient('http://localhost:8080/api/v1');
    this.testData = new TestDataFactory();
  }

  async runAllTests() {
    const runner = new TestRunner();

    runner.beforeEach(async () => {
      await this.testData.cleanup();
    });

    runner.afterEach(async () => {
      await this.testData.cleanup();
    });

    // Test profile retrieval
    runner.test('Get user profile', async () => {
      const user = await this.testData.createTestUser();
      
      const response = await this.client.get('/me/profile', {
        headers: { 'Authorization': `Bearer ${user.token}` }
      });
      
      TestAssertions.assertStatus(response, 200);
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertSchema(data, {
        id: 'string',
        email: 'string',
        firstName: 'string',
        lastName: 'string',
        age: 'number',
        gender: 'string'
      });
    });

    runner.test('Get profile without authentication', async () => {
      const response = await this.client.get('/me/profile');
      TestAssertions.assertStatus(response, 401);
    });

    // Test profile update
    runner.test('Update user profile', async () => {
      const user = await this.testData.createTestUser();
      
      const updateData = {
        firstName: 'Updated',
        lastName: 'Name',
        bio: 'Updated bio',
        interests: ['updated', 'interests']
      };

      const response = await this.client.put('/me/profile', updateData, {
        headers: { 'Authorization': `Bearer ${user.token}` }
      });
      
      TestAssertions.assertStatus(response, 200);
      const data = await TestAssertions.assertValidJSON(response);
      TestAssertions.assertContains(data, {
        firstName: 'Updated',
        lastName: 'Name'
      });
    });

    runner.test('Update profile with invalid data', async () => {
      const user = await this.testData.createTestUser();
      
      const updateData = {
        age: 'invalid-age',
        gender: 'invalid-gender'
      };

      const response = await this.client.put('/me/profile', updateData, {
        headers: { 'Authorization': `Bearer ${user.token}` }
      });
      
      TestAssertions.assertStatus(response, 400);
    });

    await runner.run();
  }
}
```

## Integration Testing

### API Integration Tests

```javascript
// Integration tests for complete user workflows
class UserWorkflowTests {
  constructor() {
    this.client = new TestHTTPClient('http://localhost:8080/api/v1');
    this.testData = new TestDataFactory();
  }

  async runAllTests() {
    const runner = new TestRunner();

    runner.beforeEach(async () => {
      await this.testData.cleanup();
    });

    runner.afterEach(async () => {
      await this.testData.cleanup();
    });

    // Complete user registration and profile setup workflow
    runner.test('Complete user onboarding workflow', async () => {
      // Step 1: Register user
      const userData = {
        email: 'onboarding@example.com',
        password: 'Password123!',
        firstName: 'John',
        lastName: 'Doe',
        age: 28,
        gender: 'male'
      };

      const registerResponse = await this.client.post('/auth/register', userData);
      TestAssertions.assertStatus(registerResponse, 201);
      const registerData = await TestAssertions.assertValidJSON(registerResponse);

      // Step 2: Complete profile
      const profileData = {
        bio: 'Software engineer who loves hiking and photography',
        interests: ['hiking', 'photography', 'technology', 'travel'],
        location: {
          latitude: 37.7749,
          longitude: -122.4194,
          city: 'San Francisco',
          country: 'USA'
        },
        preferences: {
          ageRange: { min: 25, max: 35 },
          maxDistance: 50,
          interestedIn: ['female']
        }
      };

      const profileResponse = await this.client.put('/me/profile', profileData, {
        headers: { 'Authorization': `Bearer ${registerData.token}` }
      });
      TestAssertions.assertStatus(profileResponse, 200);

      // Step 3: Upload photos
      const photoData = {
        url: 'https://example.com/photo1.jpg',
        caption: 'Hiking in the mountains',
        isPrimary: true
      };

      const photoResponse = await this.client.post('/photos', photoData, {
        headers: { 'Authorization': `Bearer ${registerData.token}` }
      });
      TestAssertions.assertStatus(photoResponse, 201);

      // Step 4: Verify profile completion
      const profileCheckResponse = await this.client.get('/me/profile', {
        headers: { 'Authorization': `Bearer ${registerData.token}` }
      });
      TestAssertions.assertStatus(profileCheckResponse, 200);
      
      const profileCheckData = await TestAssertions.assertValidJSON(profileCheckResponse);
      TestAssertions.assertContains(profileCheckData, {
        bio: profileData.bio,
        isProfileComplete: true
      });
    });

    // Match and messaging workflow
    runner.test('Match and messaging workflow', async () => {
      // Create two users
      const user1 = await this.testData.createTestUser({
        email: 'user1@example.com',
        firstName: 'Alice',
        gender: 'female'
      });

      const user2 = await this.testData.createTestUser({
        email: 'user2@example.com',
        firstName: 'Bob',
        gender: 'male'
      });

      // User1 swipes right on User2
      const swipeResponse = await this.client.post('/discovery/swipe', {
        targetUserId: user2.id,
        action: 'like'
      }, {
        headers: { 'Authorization': `Bearer ${user1.token}` }
      });
      TestAssertions.assertStatus(swipeResponse, 200);

      // User2 swipes right on User1 (creates match)
      const swipeResponse2 = await this.client.post('/discovery/swipe', {
        targetUserId: user1.id,
        action: 'like'
      }, {
        headers: { 'Authorization': `Bearer ${user2.token}` }
      });
      TestAssertions.assertStatus(swipeResponse2, 200);

      const swipeData2 = await TestAssertions.assertValidJSON(swipeResponse2);
      TestAssertions.assertContains(swipeData2, { isMatch: true });

      // User1 sends message to User2
      const messageData = {
        matchId: swipeData2.match.id,
        content: 'Hi Bob! Nice to meet you!',
        type: 'text'
      };

      const messageResponse = await this.client.post('/messages', messageData, {
        headers: { 'Authorization': `Bearer ${user1.token}` }
      });
      TestAssertions.assertStatus(messageResponse, 201);

      // User2 retrieves messages
      const messagesResponse = await this.client.get(`/matches/${swipeData2.match.id}/messages`, {
        headers: { 'Authorization': `Bearer ${user2.token}` }
      });
      TestAssertions.assertStatus(messagesResponse, 200);

      const messagesData = await TestAssertions.assertValidJSON(messagesResponse);
      TestAssertions.assertSchema(messagesData, {
        messages: 'array',
        total: 'number'
      });

      if (messagesData.messages.length > 0) {
        TestAssertions.assertContains(messagesData.messages[0], {
          content: messageData.content,
          senderId: user1.id
        });
      }
    });

    // Ephemeral photo workflow
    runner.test('Ephemeral photo workflow', async () => {
      const user1 = await this.testData.createTestUser();
      const user2 = await this.testData.createTestUser({ email: 'user2@example.com' });

      // Create match between users
      const match = await this.testData.createTestMatch(user1.id, user2.id);

      // User1 uploads ephemeral photo
      const ephemeralPhotoData = {
        matchId: match.id,
        url: 'https://example.com/ephemeral.jpg',
        viewDuration: 10, // 10 seconds
        expiresAfter: 24 // 24 hours
      };

      const ephemeralResponse = await this.client.post('/ephemeral-photos', ephemeralPhotoData, {
        headers: { 'Authorization': `Bearer ${user1.token}` }
      });
      TestAssertions.assertStatus(ephemeralResponse, 201);

      const ephemeralData = await TestAssertions.assertValidJSON(ephemeralResponse);
      TestAssertions.assertSchema(ephemeralData, {
        id: 'string',
        url: 'string',
        viewDuration: 'number',
        expiresAfter: 'number'
      });

      // User2 views ephemeral photo
      const viewResponse = await this.client.post(`/ephemeral-photos/${ephemeralData.id}/view`, {}, {
        headers: { 'Authorization': `Bearer ${user2.token}` }
      });
      TestAssertions.assertStatus(viewResponse, 200);

      // Check photo status
      const statusResponse = await this.client.get(`/ephemeral-photos/${ephemeralData.id}/status`, {
        headers: { 'Authorization': `Bearer ${user1.token}` }
      });
      TestAssertions.assertStatus(statusResponse, 200);

      const statusData = await TestAssertions.assertValidJSON(statusResponse);
      TestAssertions.assertContains(statusData, { viewed: true });
    });

    await runner.run();
  }
}
```

## End-to-End Testing

### E2E Test Framework

```javascript
// End-to-end testing with browser automation
class E2ETestFramework {
  constructor() {
    this.browser = null;
    this.page = null;
    this.testData = new TestDataFactory();
  }

  async setup() {
    // Initialize browser (using Playwright in this example)
    const { chromium } = require('playwright');
    this.browser = await chromium.launch({ headless: false });
    this.page = await this.browser.newPage();
    
    // Set up request interception for API monitoring
    await this.page.route('**/api/**', route => {
      const request = route.request();
      console.log(`API Request: ${request.method()} ${request.url()}`);
      route.continue();
    });
  }

  async teardown() {
    if (this.page) {
      await this.page.close();
    }
    if (this.browser) {
      await this.browser.close();
    }
    await this.testData.cleanup();
  }

  async runE2ETests() {
    const runner = new TestRunner();

    runner.beforeEach(async () => {
      await this.testData.cleanup();
      await this.page.goto('http://localhost:3000');
    });

    runner.afterEach(async () => {
      await this.testData.cleanup();
    });

    // Complete user journey test
    runner.test('Complete user journey from registration to messaging', async () => {
      // Step 1: User registration
      await this.page.click('[data-testid="register-button"]');
      await this.page.fill('[data-testid="email-input"]', 'e2e@example.com');
      await this.page.fill('[data-testid="password-input"]', 'Password123!');
      await this.page.fill('[data-testid="firstname-input"]', 'E2E');
      await this.page.fill('[data-testid="lastname-input"]', 'User');
      await this.page.fill('[data-testid="age-input"]', '25');
      await this.page.selectOption('[data-testid="gender-select"]', 'male');
      await this.page.click('[data-testid="submit-register"]');

      // Wait for successful registration
      await this.page.waitForSelector('[data-testid="profile-setup"]');

      // Step 2: Complete profile
      await this.page.fill('[data-testid="bio-input"]', 'E2E test user');
      await this.page.fill('[data-testid="interests-input"]', 'testing,automation');
      await this.page.click('[data-testid="complete-profile"]');

      // Wait for profile completion
      await this.page.waitForSelector('[data-testid="discovery-page"]');

      // Step 3: Upload photo
      await this.page.click('[data-testid="upload-photo-button"]');
      await this.page.setInputFiles('[data-testid="photo-input"]', 'test-assets/test-photo.jpg');
      await this.page.fill('[data-testid="photo-caption"]', 'Test photo');
      await this.page.click('[data-testid="save-photo"]');

      // Step 4: Discovery and matching
      await this.page.waitForSelector('[data-testid="user-card"]');
      await this.page.click('[data-testid="like-button"]');

      // Step 5: Messaging (if match occurs)
      try {
        await this.page.waitForSelector('[data-testid="match-notification"]', { timeout: 5000 });
        await this.page.click('[data-testid="match-notification"]');
        await this.page.waitForSelector('[data-testid="chat-input"]');
        await this.page.fill('[data-testid="chat-input"]', 'Hello from E2E test!');
        await this.page.click('[data-testid="send-message"]');

        // Verify message sent
        await this.page.waitForSelector('[data-testid="message-sent"]');
      } catch (error) {
        console.log('No match occurred, continuing test...');
      }

      // Step 6: Verify user dashboard
      await this.page.click('[data-testid="profile-tab"]');
      await this.page.waitForSelector('[data-testid="profile-info"]');
      
      const profileInfo = await this.page.textContent('[data-testid="profile-info"]');
      if (!profileInfo.includes('E2E test user')) {
        throw new Error('Profile information not displayed correctly');
      }
    });

    // Mobile responsive test
    runner.test('Mobile responsive design test', async () => {
      // Set mobile viewport
      await this.page.setViewportSize({ width: 375, height: 667 });
      
      // Test mobile registration flow
      await this.page.click('[data-testid="register-button"]');
      await this.page.fill('[data-testid="email-input"]', 'mobile@example.com');
      await this.page.fill('[data-testid="password-input"]', 'Password123!');
      await this.page.fill('[data-testid="firstname-input"]', 'Mobile');
      await this.page.fill('[data-testid="lastname-input"]', 'User');
      await this.page.fill('[data-testid="age-input"]', '25');
      await this.page.selectOption('[data-testid="gender-select"]', 'female');
      await this.page.click('[data-testid="submit-register"]');

      // Verify mobile-specific elements
      await this.page.waitForSelector('[data-testid="mobile-navigation"]');
      
      // Test mobile discovery
      await this.page.waitForSelector('[data-testid="mobile-discovery"]');
      await this.page.click('[data-testid="mobile-like-button"]');
      
      // Verify mobile messaging
      try {
        await this.page.waitForSelector('[data-testid="mobile-chat"]', { timeout: 5000 });
        await this.page.click('[data-testid="mobile-chat"]');
        await this.page.waitForSelector('[data-testid="mobile-message-input"]');
      } catch (error) {
        console.log('Mobile chat not available, continuing...');
      }
    });

    // Performance test
    runner.test('Page load performance test', async () => {
      const startTime = Date.now();
      
      await this.page.goto('http://localhost:3000');
      await this.page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      
      // Page should load within 3 seconds
      if (loadTime > 3000) {
        throw new Error(`Page load time ${loadTime}ms exceeds 3000ms threshold`);
      }

      // Test API response times
      const apiResponseTimes = [];
      
      this.page.on('response', response => {
        if (response.url().includes('/api/')) {
          apiResponseTimes.push(response.timing().responseEnd);
        }
      });

      // Navigate through app to trigger API calls
      await this.page.click('[data-testid="login-button"]');
      await this.page.fill('[data-testid="email-input"]', 'perf@example.com');
      await this.page.fill('[data-testid="password-input"]', 'Password123!');
      await this.page.click('[data-testid="submit-login"]');

      // Wait for some API calls to complete
      await this.page.waitForTimeout(2000);

      // Check average API response time
      if (apiResponseTimes.length > 0) {
        const avgResponseTime = apiResponseTimes.reduce((a, b) => a + b, 0) / apiResponseTimes.length;
        
        if (avgResponseTime > 1000) {
          throw new Error(`Average API response time ${avgResponseTime}ms exceeds 1000ms threshold`);
        }
      }
    });

    await runner.run();
  }
}
```

## Performance Testing

### Load Testing with Artillery

```yaml
# artillery-config.yml
config:
  target: 'http://localhost:8080/api/v1'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 120
      arrivalRate: 50
      name: "Ramp up load"
    - duration: 300
      arrivalRate: 100
      name: "Sustained load"
    - duration: 60
      arrivalRate: 200
      name: "Peak load"
  processor: "./test-processor.js"

scenarios:
  - name: "User Registration and Login"
    weight: 30
    flow:
      - post:
          url: "/auth/register"
          json:
            email: "{{ randomEmail() }}"
            password: "Password123!"
            firstName: "{{ randomFirstName() }}"
            lastName: "{{ randomLastName() }}"
            age: "{{ randomNumber(18, 65) }}"
            gender: "{{ randomGender() }}"
          capture:
            - json: "$.token"
              as: "authToken"
            - json: "$.id"
              as: "userId"
      - get:
          url: "/me/profile"
          headers:
            Authorization: "Bearer {{ authToken }}"

  - name: "Discovery and Matching"
    weight: 40
    flow:
      - post:
          url: "/auth/login"
          json:
            email: "testuser{{ randomNumber(1, 1000) }}@example.com"
            password: "Password123!"
          capture:
            - json: "$.token"
              as: "authToken"
      - get:
          url: "/discovery/users"
          headers:
            Authorization: "Bearer {{ authToken }}"
          qs:
            limit: 20
            offset: 0
      - post:
          url: "/discovery/swipe"
          headers:
            Authorization: "Bearer {{ authToken }}"
          json:
            targetUserId: "{{ randomNumber(1, 1000) }}"
            action: "like"

  - name: "Messaging"
    weight: 30
    flow:
      - post:
          url: "/auth/login"
          json:
            email: "testuser{{ randomNumber(1, 1000) }}@example.com"
            password: "Password123!"
          capture:
            - json: "$.token"
              as: "authToken"
      - get:
          url: "/matches"
          headers:
            Authorization: "Bearer {{ authToken }}"
      - post:
          url: "/messages"
          headers:
            Authorization: "Bearer {{ authToken }}"
          json:
            matchId: "{{ randomNumber(1, 1000) }}"
            content: "Hello from load test!"
            type: "text"
```

```javascript
// test-processor.js for Artillery
module.exports = {
  randomEmail: function() {
    const timestamp = Date.now();
    const random = Math.floor(Math.random() * 10000);
    return `test${timestamp}${random}@example.com`;
  },
  
  randomFirstName: function() {
    const names = ['John', 'Jane', 'Mike', 'Sarah', 'David', 'Emily', 'Chris', 'Lisa'];
    return names[Math.floor(Math.random() * names.length)];
  },
  
  randomLastName: function() {
    const names = ['Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis'];
    return names[Math.floor(Math.random() * names.length)];
  },
  
  randomNumber: function(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
  },
  
  randomGender: function() {
    const genders = ['male', 'female'];
    return genders[Math.floor(Math.random() * genders.length)];
  }
};
```

### Performance Monitoring

```javascript
// Performance monitoring and metrics collection
class PerformanceMonitor {
  constructor() {
    this.metrics = {
      responseTimes: [],
      errorRates: new Map(),
      throughput: [],
      resourceUsage: []
    };
    this.startTime = null;
    this.requestCount = 0;
    this.errorCount = 0;
  }

  startMonitoring() {
    this.startTime = Date.now();
    
    // Monitor system resources
    this.startResourceMonitoring();
    
    // Monitor API responses
    this.startResponseMonitoring();
  }

  startResourceMonitoring() {
    const monitorResources = () => {
      const usage = process.memoryUsage();
      const cpuUsage = process.cpuUsage();
      
      this.metrics.resourceUsage.push({
        timestamp: Date.now(),
        memory: {
          rss: usage.rss,
          heapUsed: usage.heapUsed,
          heapTotal: usage.heapTotal,
          external: usage.external
        },
        cpu: {
          user: cpuUsage.user,
          system: cpuUsage.system
        }
      });
    };

    // Monitor every 5 seconds
    this.resourceMonitorInterval = setInterval(monitorResources, 5000);
  }

  startResponseMonitoring() {
    // This would integrate with your HTTP client to monitor responses
    // For demonstration, we'll use a mock implementation
  }

  recordResponse(responseTime, statusCode, endpoint) {
    this.requestCount++;
    
    this.metrics.responseTimes.push({
      timestamp: Date.now(),
      responseTime: responseTime,
      statusCode: statusCode,
      endpoint: endpoint
    });

    if (statusCode >= 400) {
      this.errorCount++;
      const currentErrors = this.metrics.errorRates.get(endpoint) || 0;
      this.metrics.errorRates.set(endpoint, currentErrors + 1);
    }

    // Calculate throughput (requests per second)
    const elapsedSeconds = (Date.now() - this.startTime) / 1000;
    const throughput = this.requestCount / elapsedSeconds;
    
    this.metrics.throughput.push({
      timestamp: Date.now(),
      requestsPerSecond: throughput
    });
  }

  stopMonitoring() {
    if (this.resourceMonitorInterval) {
      clearInterval(this.resourceMonitorInterval);
    }
  }

  generateReport() {
    const report = {
      summary: this.generateSummary(),
      responseTimes: this.analyzeResponseTimes(),
      errorAnalysis: this.analyzeErrors(),
      throughput: this.analyzeThroughput(),
      resourceUsage: this.analyzeResourceUsage(),
      recommendations: this.generateRecommendations()
    };

    return report;
  }

  generateSummary() {
    const duration = (Date.now() - this.startTime) / 1000;
    const avgResponseTime = this.calculateAverageResponseTime();
    const errorRate = (this.errorCount / this.requestCount) * 100;
    const avgThroughput = this.calculateAverageThroughput();

    return {
      duration: duration,
      totalRequests: this.requestCount,
      totalErrors: this.errorCount,
      errorRate: errorRate.toFixed(2) + '%',
      averageResponseTime: avgResponseTime.toFixed(2) + 'ms',
      averageThroughput: avgThroughput.toFixed(2) + ' req/s'
    };
  }

  analyzeResponseTimes() {
    const times = this.metrics.responseTimes.map(r => r.responseTime);
    
    return {
      min: Math.min(...times),
      max: Math.max(...times),
      average: this.calculateAverage(times),
      p50: this.calculatePercentile(times, 50),
      p95: this.calculatePercentile(times, 95),
      p99: this.calculatePercentile(times, 99)
    };
  }

  analyzeErrors() {
    const errorsByEndpoint = {};
    
    for (const [endpoint, count] of this.metrics.errorRates) {
      errorsByEndpoint[endpoint] = count;
    }

    return {
      totalErrors: this.errorCount,
      errorsByEndpoint: errorsByEndpoint,
      errorRate: ((this.errorCount / this.requestCount) * 100).toFixed(2) + '%'
    };
  }

  analyzeThroughput() {
    const throughputs = this.metrics.throughput.map(t => t.requestsPerSecond);
    
    return {
      average: this.calculateAverage(throughputs),
      peak: Math.max(...throughputs),
      minimum: Math.min(...throughputs)
    };
  }

  analyzeResourceUsage() {
    const memoryUsage = this.metrics.resourceUsage.map(r => r.memory.heapUsed);
    const cpuUsage = this.metrics.resourceUsage.map(r => r.cpu.user);
    
    return {
      memory: {
        average: this.calculateAverage(memoryUsage),
        peak: Math.max(...memoryUsage),
        minimum: Math.min(...memoryUsage)
      },
      cpu: {
        average: this.calculateAverage(cpuUsage),
        peak: Math.max(...cpuUsage),
        minimum: Math.min(...cpuUsage)
      }
    };
  }

  generateRecommendations() {
    const recommendations = [];
    const avgResponseTime = this.calculateAverageResponseTime();
    const errorRate = (this.errorCount / this.requestCount) * 100;
    const avgMemory = this.calculateAverage(this.metrics.resourceUsage.map(r => r.memory.heapUsed));

    if (avgResponseTime > 1000) {
      recommendations.push({
        type: 'performance',
        priority: 'high',
        title: 'High Response Times',
        description: `Average response time is ${avgResponseTime.toFixed(2)}ms, consider optimizing database queries or implementing caching`
      });
    }

    if (errorRate > 5) {
      recommendations.push({
        type: 'reliability',
        priority: 'high',
        title: 'High Error Rate',
        description: `Error rate is ${errorRate.toFixed(2)}%, investigate error patterns and improve error handling`
      });
    }

    if (avgMemory > 500 * 1024 * 1024) { // 500MB
      recommendations.push({
        type: 'memory',
        priority: 'medium',
        title: 'High Memory Usage',
        description: 'Average memory usage is high, consider memory optimization or scaling'
      });
    }

    return recommendations;
  }

  calculateAverage(values) {
    if (values.length === 0) return 0;
    return values.reduce((sum, value) => sum + value, 0) / values.length;
  }

  calculateAverageResponseTime() {
    const times = this.metrics.responseTimes.map(r => r.responseTime);
    return this.calculateAverage(times);
  }

  calculateAverageThroughput() {
    const throughputs = this.metrics.throughput.map(t => t.requestsPerSecond);
    return this.calculateAverage(throughputs);
  }

  calculatePercentile(values, percentile) {
    const sorted = values.sort((a, b) => a - b);
    const index = Math.ceil((percentile / 100) * sorted.length) - 1;
    return sorted[index];
  }
}
```

## Postman Collections

### Postman Collection for API Testing

```json
{
  "info": {
    "name": "WinKr API Collection",
    "description": "Comprehensive API testing collection for WinKr backend",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080/api/v1",
      "type": "string"
    },
    {
      "key": "authToken",
      "value": "",
      "type": "string"
    },
    {
      "key": "userId",
      "value": "",
      "type": "string"
    },
    {
      "key": "refreshToken",
      "value": "",
      "type": "string"
    }
  ],
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Register User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"{{$randomEmail}}\",\n  \"password\": \"Password123!\",\n  \"firstName\": \"{{$randomFirstName}}\",\n  \"lastName\": \"{{$randomLastName}}\",\n  \"age\": \"{{$randomInt}}\",\n  \"gender\": \"male\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/auth/register",
              "host": ["{{baseUrl}}"],
              "path": ["auth", "register"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 201\", function () {",
                  "    pm.response.to.have.status(201);",
                  "});",
                  "",
                  "pm.test(\"Response has required fields\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('id');",
                  "    pm.expect(jsonData).to.have.property('email');",
                  "    pm.expect(jsonData).to.have.property('token');",
                  "    pm.expect(jsonData).to.have.property('refreshToken');",
                  "});",
                  "",
                  "// Store tokens for subsequent requests",
                  "const jsonData = pm.response.json();",
                  "pm.collectionVariables.set('authToken', jsonData.token);",
                  "pm.collectionVariables.set('refreshToken', jsonData.refreshToken);",
                  "pm.collectionVariables.set('userId', jsonData.id);"
                ]
              }
            }
          ]
        },
        {
          "name": "Login User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"test@example.com\",\n  \"password\": \"Password123!\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/auth/login",
              "host": ["{{baseUrl}}"],
              "path": ["auth", "login"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Response has required fields\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('token');",
                  "    pm.expect(jsonData).to.have.property('refreshToken');",
                  "    pm.expect(jsonData).to.have.property('user');",
                  "});",
                  "",
                  "// Store tokens for subsequent requests",
                  "const jsonData = pm.response.json();",
                  "pm.collectionVariables.set('authToken', jsonData.token);",
                  "pm.collectionVariables.set('refreshToken', jsonData.refreshToken);"
                ]
              }
            }
          ]
        },
        {
          "name": "Refresh Token",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"refreshToken\": \"{{refreshToken}}\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/auth/refresh",
              "host": ["{{baseUrl}}"],
              "path": ["auth", "refresh"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Response has new tokens\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('token');",
                  "    pm.expect(jsonData).to.have.property('refreshToken');",
                  "});"
                ]
              }
            }
          ]
        }
      ]
    },
    {
      "name": "Profile Management",
      "item": [
        {
          "name": "Get Profile",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{authToken}}"
              }
            ],
            "url": {
              "raw": "{{baseUrl}}/me/profile",
              "host": ["{{baseUrl}}"],
              "path": ["me", "profile"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Profile has required fields\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('id');",
                  "    pm.expect(jsonData).to.have.property('email');",
                  "    pm.expect(jsonData).to.have.property('firstName');",
                  "    pm.expect(jsonData).to.have.property('lastName');",
                  "    pm.expect(jsonData).to.have.property('age');",
                  "    pm.expect(jsonData).to.have.property('gender');",
                  "});"
                ]
              }
            }
          ]
        },
        {
          "name": "Update Profile",
          "request": {
            "method": "PUT",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{authToken}}"
              },
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"bio\": \"Updated bio from Postman\",\n  \"interests\": [\"testing\", \"api\", \"postman\"],\n  \"location\": {\n    \"latitude\": 40.7128,\n    \"longitude\": -74.0060,\n    \"city\": \"New York\",\n    \"country\": \"USA\"\n  }\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/me/profile",
              "host": ["{{baseUrl}}"],
              "path": ["me", "profile"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Profile updated successfully\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('bio');",
                  "    pm.expect(jsonData.bio).to.eql('Updated bio from Postman');",
                  "});"
                ]
              }
            }
          ]
        }
      ]
    },
    {
      "name": "Discovery",
      "item": [
        {
          "name": "Get Discovery Users",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{authToken}}"
              }
            ],
            "url": {
              "raw": "{{baseUrl}}/discovery/users?limit=20&offset=0",
              "host": ["{{baseUrl}}"],
              "path": ["discovery", "users"],
              "query": [
                {
                  "key": "limit",
                  "value": "20"
                },
                {
                  "key": "offset",
                  "value": "0"
                }
              ]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Response has users array\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('users');",
                  "    pm.expect(jsonData.users).to.be.an('array');",
                  "});",
                  "",
                  "pm.test(\"Each user has required fields\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    if (jsonData.users.length > 0) {",
                  "        const user = jsonData.users[0];",
                  "        pm.expect(user).to.have.property('id');",
                  "        pm.expect(user).to.have.property('firstName');",
                  "        pm.expect(user).to.have.property('age');",
                  "        pm.expect(user).to.have.property('photos');",
                  "    }",
                  "});"
                ]
              }
            }
          ]
        },
        {
          "name": "Swipe User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{authToken}}"
              },
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"targetUserId\": \"{{userId}}\",\n  \"action\": \"like\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/discovery/swipe",
              "host": ["{{baseUrl}}"],
              "path": ["discovery", "swipe"]
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "",
                  "pm.test(\"Swipe response has required fields\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData).to.have.property('action');",
                  "    pm.expect(jsonData).to.have.property('isMatch');",
                  "    pm.expect(jsonData).to.have.property('timestamp');",
                  "});"
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Automated Testing

### CI/CD Pipeline Integration

```yaml
# .github/workflows/api-tests.yml
name: API Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: |
        go mod download
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    - name: Run linter
      run: golangci-lint run
    
    - name: Run unit tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: winkr_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Wait for services
      run: |
        sleep 10
        pg_isready -h localhost -p 5432
        redis-cli -h localhost -p 6379 ping
    
    - name: Run migrations
      run: |
        go run cmd/migrate/main.go up
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/winkr_test?sslmode=disable
    
    - name: Start API server
      run: |
        go run cmd/api/main.go &
        sleep 10
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/winkr_test?sslmode=disable
        REDIS_URL: redis://localhost:6379
    
    - name: Run integration tests
      run: |
        go test -v -tags=integration ./tests/integration/...
      env:
        API_BASE_URL: http://localhost:8080/api/v1
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/winkr_test?sslmode=disable
        REDIS_URL: redis://localhost:6379

  e2e-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Install dependencies
      run: |
        cd tests/e2e
        npm install
    
    - name: Install Playwright
      run: |
        cd tests/e2e
        npx playwright install
    
    - name: Run E2E tests
      run: |
        cd tests/e2e
        npx playwright test
      env:
        API_BASE_URL: http://localhost:8080/api/v1
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: playwright-report
        path: tests/e2e/playwright-report/

  performance-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Install Artillery
      run: npm install -g artillery
    
    - name: Run performance tests
      run: |
        artillery run tests/performance/artillery-config.yml
      env:
        API_BASE_URL: http://localhost:8080/api/v1
    
    - name: Upload performance report
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: performance-report
        path: artillery-report.html
```

## Testing Best Practices

### Test Organization

```javascript
// Test organization and structure guidelines
class TestBestPractices {
  constructor() {
    this.guidelines = {
      naming: {
        describe: 'Use clear, descriptive test names',
        examples: [
          'should return 200 when user logs in with valid credentials',
          'should return 400 when email is invalid',
          'should create match when both users swipe right'
        ]
      },
      
      structure: {
        arrange: 'Set up test data and conditions',
        act: 'Execute the action being tested',
        assert: 'Verify the expected outcome',
        cleanup: 'Clean up test data'
      },
      
      isolation: {
        description: 'Tests should be independent and not rely on each other',
        practices: [
          'Create fresh test data for each test',
          'Clean up after each test',
          'Avoid shared state between tests'
        ]
      },
      
      coverage: {
        unit: 'Test individual functions and methods',
        integration: 'Test API endpoints and database interactions',
        e2e: 'Test complete user workflows',
        security: 'Test for vulnerabilities and security issues'
      }
    };
  }

  generateTestTemplate(testName, category) {
    const templates = {
      unit: `
describe('${testName}', () => {
  let testData;
  let mockService;

  beforeEach(() => {
    // Arrange: Set up test data
    testData = createTestData();
    mockService = createMockService();
  });

  afterEach(() => {
    // Cleanup: Clean up test data
    cleanupTestData();
  });

  it('should [expected behavior]', async () => {
    // Act: Execute the action
    const result = await functionUnderTest(testData);
    
    // Assert: Verify the outcome
    expect(result).toBeDefined();
    expect(result.status).toBe(200);
  });
});`,

      integration: `
describe('${testName} - Integration Tests', () => {
  let testClient;
  let testDataFactory;

  beforeAll(async () => {
    testClient = new TestHTTPClient(process.env.API_BASE_URL);
    testDataFactory = new TestDataFactory();
  });

  afterAll(async () => {
    await testDataFactory.cleanup();
  });

  beforeEach(async () => {
    await testDataFactory.cleanup();
  });

  it('should [expected behavior]', async () => {
    // Arrange: Create test data
    const user = await testDataFactory.createTestUser();
    
    // Act: Make API call
    const response = await testClient.get('/me/profile', {
      headers: { Authorization: \`Bearer \${user.token}\` }
    });
    
    // Assert: Verify response
    TestAssertions.assertStatus(response, 200);
    const data = await TestAssertions.assertValidJSON(response);
    TestAssertions.assertSchema(data, {
      id: 'string',
      email: 'string'
    });
  });
});`,

      e2e: `
describe('${testName} - E2E Tests', () => {
  let browser;
  let page;
  let testDataFactory;

  beforeAll(async () => {
    browser = await chromium.launch({ headless: true });
    testDataFactory = new TestDataFactory();
  });

  afterAll(async () => {
    await browser.close();
    await testDataFactory.cleanup();
  });

  beforeEach(async () => {
    page = await browser.newPage();
    await testDataFactory.cleanup();
  });

  afterEach(async () => {
    await page.close();
  });

  it('should complete [user workflow]', async () => {
    // Arrange: Navigate to application
    await page.goto('http://localhost:3000');
    
    // Act: Perform user actions
    await page.click('[data-testid="register-button"]');
    await page.fill('[data-testid="email-input"]', 'e2e@example.com');
    // ... more actions
    
    // Assert: Verify expected outcome
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
  });
});`
    };

    return templates[category] || templates.unit;
  }

  generateChecklist() {
    return {
      'Test Naming': [
        'Use descriptive test names that explain what is being tested',
        'Follow consistent naming convention (should/when/if)',
        'Include expected outcome in test name'
      ],
      
      'Test Structure': [
        'Follow Arrange-Act-Assert pattern',
        'Keep tests short and focused',
        'One assertion per test when possible'
      ],
      
      'Test Data': [
        'Use factory pattern for test data creation',
        'Clean up test data after each test',
        'Use realistic but anonymized test data'
      ],
      
      'Assertions': [
        'Use specific assertions over generic ones',
        'Assert on behavior, not implementation details',
        'Include helpful error messages in assertions'
      ],
      
      'Mocking': [
        'Mock external dependencies',
        'Use consistent mock data',
        'Verify mock interactions when necessary'
      ],
      
      'Coverage': [
        'Aim for high test coverage',
        'Test edge cases and error conditions',
        'Include performance and security tests'
      ]
    };
  }
}
```

---

For additional testing support or questions, contact our development team at dev-support@winkr.com.