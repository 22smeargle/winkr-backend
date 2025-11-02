# WinKr API - Code Examples and Samples

## Overview

This document provides comprehensive code examples and samples for integrating with the WinKr API. It includes practical implementations in various programming languages, common use cases, and ready-to-use code snippets.

## Table of Contents

1. [Quick Start Examples](#quick-start-examples)
2. [Authentication Examples](#authentication-examples)
3. [Profile Management Examples](#profile-management-examples)
4. [Photo Management Examples](#photo-management-examples)
5. [Discovery and Matching Examples](#discovery-and-matching-examples)
6. [Messaging Examples](#messaging-examples)
7. [Ephemeral Photos Examples](#ephemeral-photos-examples)
8. [Payment Examples](#payment-examples)
9. [WebSocket Examples](#websocket-examples)
10. [Error Handling Examples](#error-handling-examples)
11. [Advanced Examples](#advanced-examples)
12. [SDK Examples](#sdk-examples)

## Quick Start Examples

### JavaScript/Node.js

```javascript
// Quick start with Node.js
const axios = require('axios');

class WinKrAPIClient {
  constructor(apiKey, baseURL = 'https://api.winkr.com/v1') {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
    this.client = axios.create({
      baseURL: baseURL,
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async getProfile() {
    try {
      const response = await this.client.get('/me/profile');
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  async updateProfile(profileData) {
    try {
      const response = await this.client.put('/me/profile', profileData);
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  async getDiscoveryUsers(options = {}) {
    try {
      const response = await this.client.get('/discovery/users', { params: options });
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  handleError(error) {
    if (error.response) {
      console.error('API Error:', error.response.data);
      throw new Error(error.response.data.message || 'API request failed');
    } else if (error.request) {
      console.error('Network Error:', error.message);
      throw new Error('Network request failed');
    } else {
      console.error('Error:', error.message);
      throw error;
    }
  }
}

// Usage example
async function main() {
  const client = new WinKrAPIClient('your-api-key-here');
  
  try {
    // Get user profile
    const profile = await client.getProfile();
    console.log('User profile:', profile);
    
    // Update profile
    const updatedProfile = await client.updateProfile({
      bio: 'Software developer who loves hiking',
      interests: ['hiking', 'programming', 'photography']
    });
    console.log('Updated profile:', updatedProfile);
    
    // Get discovery users
    const users = await client.getDiscoveryUsers({ limit: 10, offset: 0 });
    console.log('Discovery users:', users);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

main();
```

### Python

```python
# Quick start with Python
import requests
import json
from typing import Dict, Any, Optional

class WinKrAPIClient:
    def __init__(self, api_key: str, base_url: str = "https://api.winkr.com/v1"):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        })

    def get_profile(self) -> Dict[str, Any]:
        try:
            response = self.session.get(f"{self.base_url}/me/profile")
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            self._handle_error(e)

    def update_profile(self, profile_data: Dict[str, Any]) -> Dict[str, Any]:
        try:
            response = self.session.put(f"{self.base_url}/me/profile", json=profile_data)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            self._handle_error(e)

    def get_discovery_users(self, options: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        try:
            response = self.session.get(f"{self.base_url}/discovery/users", params=options or {})
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            self._handle_error(e)

    def _handle_error(self, error: requests.exceptions.RequestException):
        if hasattr(error, 'response') and error.response is not None:
            print(f"API Error: {error.response.json()}")
            raise Exception(error.response.json().get('message', 'API request failed'))
        else:
            print(f"Network Error: {str(error)}")
            raise Exception('Network request failed')

# Usage example
def main():
    client = WinKrAPIClient('your-api-key-here')
    
    try:
        # Get user profile
        profile = client.get_profile()
        print(f"User profile: {json.dumps(profile, indent=2)}")
        
        # Update profile
        updated_profile = client.update_profile({
            'bio': 'Software developer who loves hiking',
            'interests': ['hiking', 'programming', 'photography']
        })
        print(f"Updated profile: {json.dumps(updated_profile, indent=2)}")
        
        # Get discovery users
        users = client.get_discovery_users({'limit': 10, 'offset': 0})
        print(f"Discovery users: {json.dumps(users, indent=2)}")
    except Exception as e:
        print(f"Error: {str(e)}")

if __name__ == "__main__":
    main()
```

### Go

```go
// Quick start with Go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WinKrAPIClient struct {
	BaseURL string
	APIKey string
	Client *http.Client
}

type Profile struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Bio      string   `json:"bio"`
	Interests []string `json:"interests"`
}

func NewWinKrAPIClient(apiKey, baseURL string) *WinKrAPIClient {
	return &WinKrAPIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

func (c *WinKrAPIClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	url := c.BaseURL + endpoint
	
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	return c.Client.Do(req)
}

func (c *WinKrAPIClient) GetProfile() (*Profile, error) {
	resp, err := c.makeRequest("GET", "/me/profile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	var profile Profile
	err = json.NewDecoder(resp.Body).Decode(&profile)
	if err != nil {
		return nil, err
	}
	
	return &profile, nil
}

func (c *WinKrAPIClient) UpdateProfile(profileData map[string]interface{}) (*Profile, error) {
	resp, err := c.makeRequest("PUT", "/me/profile", profileData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	var profile Profile
	err = json.NewDecoder(resp.Body).Decode(&profile)
	if err != nil {
		return nil, err
	}
	
	return &profile, nil
}

func main() {
	client := NewWinKrAPIClient("your-api-key-here", "https://api.winkr.com/v1")
	
	// Get user profile
	profile, err := client.GetProfile()
	if err != nil {
		fmt.Printf("Error getting profile: %v\n", err)
		return
	}
	
	fmt.Printf("User profile: %+v\n", profile)
	
	// Update profile
	updatedProfile, err := client.UpdateProfile(map[string]interface{}{
		"bio":       "Software developer who loves hiking",
		"interests": []string{"hiking", "programming", "photography"},
	})
	if err != nil {
		fmt.Printf("Error updating profile: %v\n", err)
		return
	}
	
	fmt.Printf("Updated profile: %+v\n", updatedProfile)
}
```

## Authentication Examples

### User Registration

```javascript
// JavaScript - User Registration
async function registerUser(userData) {
  try {
    const response = await fetch('https://api.winkr.com/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: userData.email,
        password: userData.password,
        firstName: userData.firstName,
        lastName: userData.lastName,
        age: userData.age,
        gender: userData.gender
      })
    });

    const data = await response.json();

    if (response.ok) {
      console.log('Registration successful:', data);
      // Store tokens
      localStorage.setItem('winkr_token', data.token);
      localStorage.setItem('winkr_refresh_token', data.refreshToken);
      return data;
    } else {
      console.error('Registration failed:', data);
      throw new Error(data.message || 'Registration failed');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
}

// Usage
registerUser({
  email: 'john.doe@example.com',
  password: 'SecurePassword123!',
  firstName: 'John',
  lastName: 'Doe',
  age: 28,
  gender: 'male'
}).then(user => {
  console.log('User registered:', user);
}).catch(error => {
  console.error('Registration error:', error);
});
```

```python
# Python - User Registration
import requests

def register_user(user_data):
    url = "https://api.winkr.com/v1/auth/register"
    payload = {
        "email": user_data["email"],
        "password": user_data["password"],
        "firstName": user_data["firstName"],
        "lastName": user_data["lastName"],
        "age": user_data["age"],
        "gender": user_data["gender"]
    }
    
    try:
        response = requests.post(url, json=payload)
        response.raise_for_status()
        
        data = response.json()
        print("Registration successful:", data)
        
        # Store tokens
        with open('.winkr_tokens', 'w') as f:
            json.dump({
                'token': data['token'],
                'refresh_token': data['refreshToken']
            }, f)
        
        return data
        
    except requests.exceptions.RequestException as e:
        print("Registration failed:", e)
        raise

# Usage
user_data = {
    "email": "john.doe@example.com",
    "password": "SecurePassword123!",
    "firstName": "John",
    "lastName": "Doe",
    "age": 28,
    "gender": "male"
}

try:
    user = register_user(user_data)
    print("User registered:", user)
except Exception as e:
    print("Registration error:", e)
```

### User Login

```javascript
// JavaScript - User Login with MFA
async function loginUser(email, password, mfaCode = null) {
  try {
    const response = await fetch('https://api.winkr.com/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: email,
        password: password,
        mfaCode: mfaCode
      })
    });

    const data = await response.json();

    if (response.ok) {
      console.log('Login successful:', data);
      
      // Store tokens securely
      localStorage.setItem('winkr_token', data.token);
      localStorage.setItem('winkr_refresh_token', data.refreshToken);
      
      // Set up automatic token refresh
      setupTokenRefresh(data.refreshToken);
      
      return data;
    } else if (response.status === 401 && data.requiresMFA) {
      console.log('MFA required');
      throw new Error('MFA code required');
    } else {
      console.error('Login failed:', data);
      throw new Error(data.message || 'Login failed');
    }
  } catch (error) {
    console.error('Network error:', error);
    throw error;
  }
}

// Token refresh setup
function setupTokenRefresh(refreshToken) {
  // Refresh token 5 minutes before expiry
  setInterval(async () => {
    try {
      const response = await fetch('https://api.winkr.com/v1/auth/refresh', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          refreshToken: refreshToken
        })
      });

      const data = await response.json();
      
      if (response.ok) {
        localStorage.setItem('winkr_token', data.token);
        localStorage.setItem('winkr_refresh_token', data.refreshToken);
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
    }
  }, 4 * 60 * 1000); // 4 minutes
}
```

## Profile Management Examples

### Complete Profile Setup

```javascript
// JavaScript - Complete Profile Setup
class ProfileManager {
  constructor(apiClient) {
    this.apiClient = apiClient;
  }

  async completeProfileSetup(profileData) {
    try {
      // Step 1: Update basic profile information
      const basicProfile = await this.updateBasicProfile(profileData.basic);
      console.log('Basic profile updated:', basicProfile);

      // Step 2: Upload photos
      const photos = await this.uploadPhotos(profileData.photos);
      console.log('Photos uploaded:', photos);

      // Step 3: Set preferences
      const preferences = await this.updatePreferences(profileData.preferences);
      console.log('Preferences updated:', preferences);

      // Step 4: Verify profile completion
      const completionStatus = await this.checkProfileCompletion();
      console.log('Profile completion status:', completionStatus);

      return {
        profile: basicProfile,
        photos: photos,
        preferences: preferences,
        completion: completionStatus
      };
    } catch (error) {
      console.error('Profile setup failed:', error);
      throw error;
    }
  }

  async updateBasicProfile(basicData) {
    return await this.apiClient.put('/me/profile', {
      firstName: basicData.firstName,
      lastName: basicData.lastName,
      age: basicData.age,
      gender: basicData.gender,
      bio: basicData.bio,
      interests: basicData.interests,
      location: basicData.location
    });
  }

  async uploadPhotos(photos) {
    const uploadedPhotos = [];
    
    for (let i = 0; i < photos.length; i++) {
      const photo = photos[i];
      const formData = new FormData();
      
      formData.append('photo', photo.file);
      formData.append('caption', photo.caption);
      formData.append('isPrimary', i === 0); // First photo is primary

      const uploadedPhoto = await this.apiClient.post('/photos', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      
      uploadedPhotos.push(uploadedPhoto);
    }
    
    return uploadedPhotos;
  }

  async updatePreferences(preferences) {
    return await this.apiClient.put('/me/preferences', {
      ageRange: preferences.ageRange,
      maxDistance: preferences.maxDistance,
      interestedIn: preferences.interestedIn,
      discoverySettings: preferences.discoverySettings
    });
  }

  async checkProfileCompletion() {
    const response = await this.apiClient.get('/me/profile/completion');
    return response.data;
  }
}

// Usage example
async function setupCompleteProfile() {
  const apiClient = new WinKrAPIClient('your-api-key-here');
  const profileManager = new ProfileManager(apiClient);

  const profileData = {
    basic: {
      firstName: 'John',
      lastName: 'Doe',
      age: 28,
      gender: 'male',
      bio: 'Software developer who loves hiking and photography',
      interests: ['hiking', 'photography', 'programming', 'travel'],
      location: {
        latitude: 40.7128,
        longitude: -74.0060,
        city: 'New York',
        country: 'USA'
      }
    },
    photos: [
      {
        file: document.getElementById('photo1').files[0],
        caption: 'Hiking in the mountains'
      },
      {
        file: document.getElementById('photo2').files[0],
        caption: 'At the office'
      }
    ],
    preferences: {
      ageRange: { min: 25, max: 35 },
      maxDistance: 50,
      interestedIn: ['female'],
      discoverySettings: {
        showDistance: true,
        showAge: true,
        enableSuperLikes: true
      }
    }
  };

  try {
    const result = await profileManager.completeProfileSetup(profileData);
    console.log('Profile setup completed:', result);
    
    if (result.completion.isComplete) {
      alert('Profile setup completed successfully!');
    } else {
      alert('Profile setup incomplete. Please complete: ' + 
            result.completion.missingFields.join(', '));
    }
  } catch (error) {
    console.error('Profile setup failed:', error);
    alert('Profile setup failed: ' + error.message);
  }
}
```

## Photo Management Examples

### Photo Upload with Progress

```javascript
// JavaScript - Photo Upload with Progress
class PhotoUploader {
  constructor(apiClient) {
    this.apiClient = apiClient;
  }

  async uploadPhoto(file, options = {}) {
    return new Promise((resolve, reject) => {
      const formData = new FormData();
      formData.append('photo', file);
      
      if (options.caption) {
        formData.append('caption', options.caption);
      }
      
      if (options.isPrimary !== undefined) {
        formData.append('isPrimary', options.isPrimary);
      }

      const xhr = new XMLHttpRequest();
      
      // Progress tracking
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const percentComplete = (event.loaded / event.total) * 100;
          options.onProgress && options.onProgress(percentComplete, event.loaded, event.total);
        }
      });

      // Completion
      xhr.addEventListener('load', () => {
        if (xhr.status === 201) {
          const response = JSON.parse(xhr.responseText);
          resolve(response);
        } else {
          const error = JSON.parse(xhr.responseText);
          reject(new Error(error.message || 'Upload failed'));
        }
      });

      // Error handling
      xhr.addEventListener('error', () => {
        reject(new Error('Network error during upload'));
      });

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload cancelled'));
      });

      // Start upload
      xhr.open('POST', `${this.apiClient.baseURL}/photos`);
      xhr.setRequestHeader('Authorization', `Bearer ${this.apiClient.apiKey}`);
      xhr.send(formData);
    });
  }

  async uploadMultiplePhotos(files, options = {}) {
    const uploadPromises = files.map((file, index) => 
      this.uploadPhoto(file, {
        ...options,
        isPrimary: index === 0, // First photo is primary
        onProgress: (percent, loaded, total) => {
          options.onProgress && options.onProgress(index, percent, loaded, total);
        }
      })
    );

    try {
      const results = await Promise.all(uploadPromises);
      return results;
    } catch (error) {
      console.error('Multiple photo upload failed:', error);
      throw error;
    }
  }
}

// Usage example with UI
class PhotoUploadUI {
  constructor() {
    this.uploader = new PhotoUploader(new WinKrAPIClient('your-api-key-here'));
    this.setupEventListeners();
  }

  setupEventListeners() {
    const fileInput = document.getElementById('photo-input');
    const uploadButton = document.getElementById('upload-button');
    const progressBar = document.getElementById('progress-bar');
    const progressText = document.getElementById('progress-text');

    uploadButton.addEventListener('click', () => {
      fileInput.click();
    });

    fileInput.addEventListener('change', (event) => {
      const files = Array.from(event.target.files);
      this.uploadFiles(files, progressBar, progressText);
    });
  }

  async uploadFiles(files, progressBar, progressText) {
    try {
      uploadButton.disabled = true;
      progressBar.style.width = '0%';
      progressText.textContent = 'Starting upload...';

      const results = await this.uploader.uploadMultiplePhotos(files, {
        onProgress: (fileIndex, percent, loaded, total) => {
          const overallPercent = ((fileIndex + percent / 100) / files.length) * 100;
          progressBar.style.width = `${overallPercent}%`;
          progressText.textContent = `Uploading file ${fileIndex + 1}/${files.length}: ${Math.round(percent)}%`;
        }
      });

      progressBar.style.width = '100%';
      progressText.textContent = 'Upload completed!';
      
      console.log('Upload results:', results);
      this.displayUploadedPhotos(results);
      
    } catch (error) {
      console.error('Upload failed:', error);
      progressText.textContent = `Upload failed: ${error.message}`;
    } finally {
      uploadButton.disabled = false;
    }
  }

  displayUploadedPhotos(photos) {
    const gallery = document.getElementById('photo-gallery');
    gallery.innerHTML = '';

    photos.forEach(photo => {
      const photoElement = document.createElement('div');
      photoElement.className = 'photo-item';
      photoElement.innerHTML = `
        <img src="${photo.url}" alt="${photo.caption}" />
        <div class="photo-info">
          <p>${photo.caption}</p>
          <small>Uploaded: ${new Date(photo.createdAt).toLocaleString()}</small>
        </div>
      `;
      gallery.appendChild(photoElement);
    });
  }
}

// Initialize the UI
const photoUploadUI = new PhotoUploadUI();
```

## Discovery and Matching Examples

### Advanced Discovery with Filters

```javascript
// JavaScript - Advanced Discovery
class DiscoveryManager {
  constructor(apiClient) {
    this.apiClient = apiClient;
    this.currentFilters = {};
    this.cache = new Map();
  }

  async discoverUsers(filters = {}, pagination = {}) {
    const cacheKey = this.generateCacheKey(filters, pagination);
    
    // Check cache first
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey);
    }

    try {
      const params = {
        ...filters,
        ...pagination,
        timestamp: Date.now() // Prevent caching
      };

      const response = await this.apiClient.get('/discovery/users', { params });
      const data = response.data;

      // Cache results
      this.cache.set(cacheKey, data);
      
      // Clean old cache entries
      this.cleanCache();

      return data;
    } catch (error) {
      console.error('Discovery failed:', error);
      throw error;
    }
  }

  async swipeUser(userId, action) {
    try {
      const response = await this.apiClient.post('/discovery/swipe', {
        targetUserId: userId,
        action: action // 'like' or 'pass'
      });

      const result = response.data;
      
      if (result.isMatch) {
        this.handleNewMatch(result.match);
      }

      return result;
    } catch (error) {
      console.error('Swipe failed:', error);
      throw error;
    }
  }

  async getSuperLikesRemaining() {
    try {
      const response = await this.apiClient.get('/me/super-likes/remaining');
      return response.data.remaining;
    } catch (error) {
      console.error('Failed to get super likes remaining:', error);
      return 0;
    }
  }

  async useSuperLike(userId) {
    try {
      const response = await this.apiClient.post('/discovery/super-like', {
        targetUserId: userId
      });

      const result = response.data;
      
      if (result.isMatch) {
        this.handleNewMatch(result.match);
      }

      return result;
    } catch (error) {
      console.error('Super like failed:', error);
      throw error;
    }
  }

  handleNewMatch(match) {
    // Show match notification
    this.showMatchNotification(match);
    
    // Play sound
    this.playMatchSound();
    
    // Update UI
    this.updateMatchUI(match);
  }

  showMatchNotification(match) {
    const notification = document.createElement('div');
    notification.className = 'match-notification';
    notification.innerHTML = `
      <div class="match-content">
        <h3>It's a Match! 🎉</h3>
        <p>You matched with ${match.matchedUser.firstName}!</p>
        <img src="${match.matchedUser.photos[0].url}" alt="${match.matchedUser.firstName}" />
        <button onclick="this.parentElement.parentElement.remove()">Close</button>
      </div>
    `;
    
    document.body.appendChild(notification);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
      if (notification.parentElement) {
        notification.remove();
      }
    }, 5000);
  }

  playMatchSound() {
    const audio = new Audio('/sounds/match.mp3');
    audio.play().catch(e => console.log('Could not play match sound:', e));
  }

  updateMatchUI(match) {
    // Update matches list
    const matchesList = document.getElementById('matches-list');
    const matchElement = document.createElement('div');
    matchElement.className = 'match-item';
    matchElement.innerHTML = `
      <img src="${match.matchedUser.photos[0].url}" alt="${match.matchedUser.firstName}" />
      <div class="match-info">
        <h4>${match.matchedUser.firstName}</h4>
        <p>Matched ${new Date(match.createdAt).toLocaleString()}</p>
      </div>
    `;
    matchesList.insertBefore(matchElement, matchesList.firstChild);
  }

  generateCacheKey(filters, pagination) {
    return JSON.stringify({ filters, pagination });
  }

  cleanCache() {
    // Remove cache entries older than 5 minutes
    const now = Date.now();
    for (const [key, value] of this.cache.entries()) {
      if (now - value.timestamp > 5 * 60 * 1000) {
        this.cache.delete(key);
      }
    }
  }
}

// Usage example with advanced filters
class DiscoveryUI {
  constructor() {
    this.discoveryManager = new DiscoveryManager(new WinKrAPIClient('your-api-key-here'));
    this.currentUserIndex = 0;
    this.users = [];
    this.setupEventListeners();
  }

  setupEventListeners() {
    document.getElementById('apply-filters').addEventListener('click', () => {
      this.applyFilters();
    });

    document.getElementById('like-button').addEventListener('click', () => {
      this.likeUser();
    });

    document.getElementById('pass-button').addEventListener('click', () => {
      this.passUser();
    });

    document.getElementById('super-like-button').addEventListener('click', () => {
      this.superLikeUser();
    });
  }

  async applyFilters() {
    const filters = {
      ageRange: {
        min: parseInt(document.getElementById('min-age').value),
        max: parseInt(document.getElementById('max-age').value)
      },
      maxDistance: parseInt(document.getElementById('max-distance').value),
      interestedIn: Array.from(document.querySelectorAll('input[name="interested-in"]:checked'))
        .map(input => input.value),
      hasPhotos: document.getElementById('has-photos').checked,
      onlineOnly: document.getElementById('online-only').checked
    };

    try {
      const result = await this.discoveryManager.discoverUsers(filters, { limit: 50 });
      this.users = result.users;
      this.currentUserIndex = 0;
      this.displayCurrentUser();
    } catch (error) {
      console.error('Failed to apply filters:', error);
      alert('Failed to apply filters: ' + error.message);
    }
  }

  displayCurrentUser() {
    if (this.currentUserIndex >= this.users.length) {
      this.showNoMoreUsers();
      return;
    }

    const user = this.users[this.currentUserIndex];
    const userCard = document.getElementById('user-card');
    
    userCard.innerHTML = `
      <div class="user-profile">
        <div class="photos">
          ${user.photos.map(photo => `
            <img src="${photo.url}" alt="${user.firstName}" />
          `).join('')}
        </div>
        <div class="user-info">
          <h2>${user.firstName}, ${user.age}</h2>
          <p>${user.bio}</p>
          <div class="interests">
            ${user.interests.map(interest => `
              <span class="interest-tag">${interest}</span>
            `).join('')}
          </div>
          <div class="location">
            📍 ${user.location.city}, ${user.location.country}
          </div>
        </div>
      </div>
    `;
  }

  async likeUser() {
    if (this.currentUserIndex >= this.users.length) return;
    
    const user = this.users[this.currentUserIndex];
    
    try {
      await this.discoveryManager.swipeUser(user.id, 'like');
      this.nextUser();
    } catch (error) {
      console.error('Like failed:', error);
      alert('Failed to like user: ' + error.message);
    }
  }

  async passUser() {
    if (this.currentUserIndex >= this.users.length) return;
    
    const user = this.users[this.currentUserIndex];
    
    try {
      await this.discoveryManager.swipeUser(user.id, 'pass');
      this.nextUser();
    } catch (error) {
      console.error('Pass failed:', error);
      alert('Failed to pass user: ' + error.message);
    }
  }

  async superLikeUser() {
    if (this.currentUserIndex >= this.users.length) return;
    
    const user = this.users[this.currentUserIndex];
    
    try {
      const remaining = await this.discoveryManager.getSuperLikesRemaining();
      
      if (remaining <= 0) {
        alert('No super likes remaining! Upgrade to premium for more.');
        return;
      }
      
      await this.discoveryManager.useSuperLike(user.id);
      this.nextUser();
    } catch (error) {
      console.error('Super like failed:', error);
      alert('Failed to super like user: ' + error.message);
    }
  }

  nextUser() {
    this.currentUserIndex++;
    this.displayCurrentUser();
  }

  showNoMoreUsers() {
    const userCard = document.getElementById('user-card');
    userCard.innerHTML = `
      <div class="no-more-users">
        <h2>No more users!</h2>
        <p>Try adjusting your filters or check back later.</p>
        <button onclick="location.reload()">Refresh</button>
      </div>
    `;
  }
}

// Initialize the discovery UI
const discoveryUI = new DiscoveryUI();
```

## Messaging Examples

### Real-time Chat with WebSocket

```javascript
// JavaScript - Real-time Chat
class ChatManager {
  constructor(apiClient) {
    this.apiClient = apiClient;
    this.ws = null;
    this.currentMatchId = null;
    this.messageQueue = [];
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  async connectToMatch(matchId) {
    this.currentMatchId = matchId;
    
    try {
      // Get WebSocket URL
      const response = await this.apiClient.get(`/matches/${matchId}/chat-url`);
      const wsUrl = response.data.wsUrl;
      
      // Connect WebSocket
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        console.log('Connected to chat');
        this.reconnectAttempts = 0;
        
        // Send queued messages
        this.flushMessageQueue();
      };
      
      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        this.handleIncomingMessage(message);
      };
      
      this.ws.onclose = () => {
        console.log('Disconnected from chat');
        this.handleDisconnection();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
      
    } catch (error) {
      console.error('Failed to connect to chat:', error);
      throw error;
    }
  }

  async sendMessage(content, type = 'text') {
    const message = {
      matchId: this.currentMatchId,
      content: content,
      type: type,
      timestamp: new Date().toISOString()
    };

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      // Queue message if not connected
      this.messageQueue.push(message);
    }

    // Also send via HTTP for persistence
    try {
      await this.apiClient.post('/messages', message);
    } catch (error) {
      console.error('Failed to send message via HTTP:', error);
    }
  }

  async sendPhoto(file) {
    const formData = new FormData();
    formData.append('photo', file);
    formData.append('matchId', this.currentMatchId);
    formData.append('type', 'photo');

    try {
      const response = await this.apiClient.post('/messages/photo', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });

      const message = response.data;
      this.displayMessage(message, 'sent');
      
      return message;
    } catch (error) {
      console.error('Failed to send photo:', error);
      throw error;
    }
  }

  handleIncomingMessage(message) {
    this.displayMessage(message, 'received');
    this.playMessageSound();
    this.updateLastMessage(message);
  }

  displayMessage(message, direction) {
    const messagesContainer = document.getElementById('messages-container');
    const messageElement = document.createElement('div');
    messageElement.className = `message ${direction}`;
    
    if (message.type === 'text') {
      messageElement.innerHTML = `
        <div class="message-content">${this.escapeHtml(message.content)}</div>
        <div class="message-time">${this.formatTime(message.timestamp)}</div>
      `;
    } else if (message.type === 'photo') {
      messageElement.innerHTML = `
        <img src="${message.content}" alt="Photo" class="message-photo" />
        <div class="message-time">${this.formatTime(message.timestamp)}</div>
      `;
    } else if (message.type === 'ephemeral_photo') {
      messageElement.innerHTML = `
        <div class="ephemeral-photo-container">
          <img src="${message.content}" alt="Ephemeral photo" class="message-photo" />
          <div class="ephemeral-timer">View expires in ${message.viewDuration}s</div>
        </div>
        <div class="message-time">${this.formatTime(message.timestamp)}</div>
      `;
    }

    messagesContainer.appendChild(messageElement);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
  }

  flushMessageQueue() {
    while (this.messageQueue.length > 0) {
      const message = this.messageQueue.shift();
      this.ws.send(JSON.stringify(message));
    }
  }

  handleDisconnection() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      
      setTimeout(() => {
        this.connectToMatch(this.currentMatchId);
      }, 1000 * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
      this.showConnectionError();
    }
  }

  showConnectionError() {
    const errorElement = document.createElement('div');
    errorElement.className = 'connection-error';
    errorElement.innerHTML = `
      <div class="error-content">
        <h3>Connection Lost</h3>
        <p>Unable to connect to chat. Please check your internet connection.</p>
        <button onclick="location.reload()">Reconnect</button>
      </div>
    `;
    
    document.body.appendChild(errorElement);
  }

  playMessageSound() {
    const audio = new Audio('/sounds/message.mp3');
    audio.play().catch(e => console.log('Could not play message sound:', e));
  }

  updateLastMessage(message) {
    // Update last message in matches list
    const matchElement = document.querySelector(`[data-match-id="${message.matchId}"]`);
    if (matchElement) {
      const lastMessageElement = matchElement.querySelector('.last-message');
      lastMessageElement.textContent = this.truncateMessage(message.content);
    }
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  formatTime(timestamp) {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  }

  truncateMessage(content, maxLength = 50) {
    return content.length > maxLength ? content.substring(0, maxLength) + '...' : content;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Usage example
class ChatUI {
  constructor() {
    this.chatManager = new ChatManager(new WinKrAPIClient('your-api-key-here'));
    this.currentMatchId = null;
    this.setupEventListeners();
  }

  setupEventListeners() {
    document.getElementById('send-button').addEventListener('click', () => {
      this.sendMessage();
    });

    document.getElementById('message-input').addEventListener('keypress', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        this.sendMessage();
      }
    });

    document.getElementById('photo-input').addEventListener('change', (e) => {
      this.sendPhoto(e.target.files[0]);
    });
  }

  async openChat(matchId) {
    this.currentMatchId = matchId;
    
    try {
      await this.chatManager.connectToMatch(matchId);
      
      // Load message history
      await this.loadMessageHistory(matchId);
      
      // Show chat interface
      document.getElementById('chat-container').style.display = 'block';
      document.getElementById('matches-list').style.display = 'none';
      
    } catch (error) {
      console.error('Failed to open chat:', error);
      alert('Failed to open chat: ' + error.message);
    }
  }

  async loadMessageHistory(matchId) {
    try {
      const response = await this.chatManager.apiClient.get(`/matches/${matchId}/messages`, {
        params: { limit: 50 }
      });
      
      const messages = response.data.messages;
      const messagesContainer = document.getElementById('messages-container');
      messagesContainer.innerHTML = '';
      
      messages.forEach(message => {
        const direction = message.senderId === this.currentUserId ? 'sent' : 'received';
        this.chatManager.displayMessage(message, direction);
      });
      
    } catch (error) {
      console.error('Failed to load message history:', error);
    }
  }

  async sendMessage() {
    const input = document.getElementById('message-input');
    const content = input.value.trim();
    
    if (!content) return;
    
    try {
      await this.chatManager.sendMessage(content);
      input.value = '';
    } catch (error) {
      console.error('Failed to send message:', error);
      alert('Failed to send message: ' + error.message);
    }
  }

  async sendPhoto(file) {
    if (!file) return;
    
    try {
      await this.chatManager.sendPhoto(file);
      document.getElementById('photo-input').value = '';
    } catch (error) {
      console.error('Failed to send photo:', error);
      alert('Failed to send photo: ' + error.message);
    }
  }

  closeChat() {
    this.chatManager.disconnect();
    document.getElementById('chat-container').style.display = 'none';
    document.getElementById('matches-list').style.display = 'block';
  }
}

// Initialize the chat UI
const chatUI = new ChatUI();
```

## Error Handling Examples

### Comprehensive Error Handling

```javascript
// JavaScript - Comprehensive Error Handling
class ErrorHandler {
  constructor() {
    this.errorTypes = {
      NETWORK_ERROR: 'network_error',
      API_ERROR: 'api_error',
      VALIDATION_ERROR: 'validation_error',
      AUTHENTICATION_ERROR: 'authentication_error',
      RATE_LIMIT_ERROR: 'rate_limit_error',
      SERVER_ERROR: 'server_error'
    };
  }

  classifyError(error) {
    if (!error.response) {
      return this.errorTypes.NETWORK_ERROR;
    }

    const status = error.response.status;
    const data = error.response.data;

    if (status === 401 || status === 403) {
      return this.errorTypes.AUTHENTICATION_ERROR;
    } else if (status === 400) {
      return this.errorTypes.VALIDATION_ERROR;
    } else if (status === 429) {
      return this.errorTypes.RATE_LIMIT_ERROR;
    } else if (status >= 500) {
      return this.errorTypes.SERVER_ERROR;
    } else {
      return this.errorTypes.API_ERROR;
    }
  }

  handleError(error, context = {}) {
    const errorType = this.classifyError(error);
    const errorInfo = {
      type: errorType,
      message: this.getErrorMessage(error),
      context: context,
      timestamp: new Date().toISOString(),
      originalError: error
    };

    // Log error
    this.logError(errorInfo);

    // Show user-friendly message
    this.showUserMessage(errorInfo);

    // Attempt recovery
    this.attemptRecovery(errorInfo);

    return errorInfo;
  }

  getErrorMessage(error) {
    if (error.response && error.response.data) {
      return error.response.data.message || error.response.data.error;
    } else if (error.message) {
      return error.message;
    } else {
      return 'An unexpected error occurred';
    }
  }

  logError(errorInfo) {
    console.error('Error occurred:', errorInfo);
    
    // Send to error tracking service
    if (typeof gtag !== 'undefined') {
      gtag('event', 'exception', {
        description: errorInfo.message,
        fatal: errorInfo.type === this.errorTypes.SERVER_ERROR
      });
    }
  }

  showUserMessage(errorInfo) {
    const messageElement = document.createElement('div');
    messageElement.className = 'error-message';
    
    let message = '';
    let action = '';

    switch (errorInfo.type) {
      case this.errorTypes.NETWORK_ERROR:
        message = 'Network connection lost. Please check your internet connection.';
        action = '<button onclick="location.reload()">Retry</button>';
        break;
        
      case this.errorTypes.AUTHENTICATION_ERROR:
        message = 'Your session has expired. Please log in again.';
        action = '<button onclick="redirectToLogin()">Login</button>';
        break;
        
      case this.errorTypes.VALIDATION_ERROR:
        message = errorInfo.message;
        action = '<button onclick="this.parentElement.remove()">Close</button>';
        break;
        
      case this.errorTypes.RATE_LIMIT_ERROR:
        message = 'Too many requests. Please wait a moment before trying again.';
        action = '<button onclick="this.parentElement.remove()">Close</button>';
        break;
        
      case this.errorTypes.SERVER_ERROR:
        message = 'Server error occurred. Our team has been notified.';
        action = '<button onclick="location.reload()">Retry</button>';
        break;
        
      default:
        message = 'An error occurred. Please try again.';
        action = '<button onclick="this.parentElement.remove()">Close</button>';
    }

    messageElement.innerHTML = `
      <div class="error-content">
        <p>${message}</p>
        <div class="error-actions">${action}</div>
      </div>
    `;

    document.body.appendChild(messageElement);

    // Auto-remove after 10 seconds for non-critical errors
    if (errorInfo.type !== this.errorTypes.AUTHENTICATION_ERROR) {
      setTimeout(() => {
        if (messageElement.parentElement) {
          messageElement.remove();
        }
      }, 10000);
    }
  }

  attemptRecovery(errorInfo) {
    switch (errorInfo.type) {
      case this.errorTypes.NETWORK_ERROR:
        // Implement exponential backoff retry
        this.exponentialBackoffRetry(errorInfo.context);
        break;
        
      case this.errorTypes.AUTHENTICATION_ERROR:
        // Attempt token refresh
        this.attemptTokenRefresh();
        break;
        
      case this.errorTypes.RATE_LIMIT_ERROR:
        // Implement rate limit handling
        this.handleRateLimit(errorInfo);
        break;
    }
  }

  exponentialBackoffRetry(context) {
    const maxRetries = 3;
    const baseDelay = 1000; // 1 second
    
    const retry = (attempt) => {
      if (attempt >= maxRetries) {
        console.error('Max retry attempts reached');
        return;
      }

      const delay = baseDelay * Math.pow(2, attempt);
      
      setTimeout(() => {
        console.log(`Retry attempt ${attempt + 1}/${maxRetries}`);
        
        // Retry the original request
        if (context.originalRequest) {
          context.originalRequest();
        }
      }, delay);
    };

    retry(0);
  }

  async attemptTokenRefresh() {
    try {
      const refreshToken = localStorage.getItem('winkr_refresh_token');
      
      if (refreshToken) {
        const response = await fetch('https://api.winkr.com/v1/auth/refresh', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ refreshToken })
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('winkr_token', data.token);
          localStorage.setItem('winkr_refresh_token', data.refreshToken);
          
          // Reload current page to retry with new token
          location.reload();
        }
      }
    } catch (error) {
      console.error('Token refresh failed:', error);
      // Redirect to login
      window.location.href = '/login';
    }
  }

  handleRateLimit(errorInfo) {
    const retryAfter = errorInfo.originalError?.response?.headers?.['retry-after'];
    
    if (retryAfter) {
      const retryTime = parseInt(retryAfter) * 1000;
      
      setTimeout(() => {
        console.log('Rate limit expired, retrying request');
        if (errorInfo.context.originalRequest) {
          errorInfo.context.originalRequest();
        }
      }, retryTime);
    }
  }
}

// Usage example with API client
class RobustAPIClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baseURL = 'https://api.winkr.com/v1';
    this.errorHandler = new ErrorHandler();
  }

  async makeRequest(method, endpoint, data = null, context = {}) {
    const originalRequest = () => this.makeRequest(method, endpoint, data, context);
    
    try {
      const config = {
        method: method,
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      };

      if (data) {
        config.body = JSON.stringify(data);
      }

      const response = await fetch(`${this.baseURL}${endpoint}`, config);
      
      if (!response.ok) {
        const error = new Error(`HTTP ${response.status}: ${response.statusText}`);
        error.response = {
          status: response.status,
          data: await response.json().catch(() => ({}))
        };
        throw error;
      }

      return await response.json();
      
    } catch (error) {
      const errorInfo = this.errorHandler.handleError(error, {
        ...context,
        originalRequest: originalRequest
      });
      
      throw errorInfo;
    }
  }

  async getProfile() {
    return await this.makeRequest('GET', '/me/profile', null, {
      operation: 'get_profile'
    });
  }

  async updateProfile(profileData) {
    return await this.makeRequest('PUT', '/me/profile', profileData, {
      operation: 'update_profile'
    });
  }
}

// Usage example
const robustClient = new RobustAPIClient('your-api-key-here');

robustClient.getProfile()
  .then(profile => {
    console.log('Profile loaded:', profile);
  })
  .catch(error => {
    console.log('Error handled:', error);
  });
```

## SDK Examples

### JavaScript SDK

```javascript
// WinKr JavaScript SDK
class WinKrSDK {
  constructor(config) {
    this.apiKey = config.apiKey;
    this.baseURL = config.baseURL || 'https://api.winkr.com/v1';
    this.version = config.version || 'v1';
    this.timeout = config.timeout || 10000;
    
    // Initialize modules
    this.auth = new AuthModule(this);
    this.profile = new ProfileModule(this);
    this.photos = new PhotosModule(this);
    this.discovery = new DiscoveryModule(this);
    this.messaging = new MessagingModule(this);
    this.payments = new PaymentsModule(this);
  }

  async request(method, endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      method,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers
      },
      timeout: this.timeout,
      ...options
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    } catch (error) {
      throw new WinKrError(error.message, method, endpoint);
    }
  }
}

class AuthModule {
  constructor(sdk) {
    this.sdk = sdk;
  }

  async register(userData) {
    return await this.sdk.request('POST', '/auth/register', {
      body: JSON.stringify(userData)
    });
  }

  async login(email, password) {
    return await this.sdk.request('POST', '/auth/login', {
      body: JSON.stringify({ email, password })
    });
  }

  async logout() {
    return await this.sdk.request('POST', '/auth/logout');
  }

  async refreshToken(refreshToken) {
    return await this.sdk.request('POST', '/auth/refresh', {
      body: JSON.stringify({ refreshToken })
    });
  }
}

class ProfileModule {
  constructor(sdk) {
    this.sdk = sdk;
  }

  async getProfile() {
    return await this.sdk.request('GET', '/me/profile');
  }

  async updateProfile(profileData) {
    return await this.sdk.request('PUT', '/me/profile', {
      body: JSON.stringify(profileData)
    });
  }

  async getPreferences() {
    return await this.sdk.request('GET', '/me/preferences');
  }

  async updatePreferences(preferences) {
    return await this.sdk.request('PUT', '/me/preferences', {
      body: JSON.stringify(preferences)
    });
  }
}

class WinKrError extends Error {
  constructor(message, method, endpoint) {
    super(message);
    this.name = 'WinKrError';
    this.method = method;
    this.endpoint = endpoint;
  }
}

// Usage example
const sdk = new WinKrSDK({
  apiKey: 'your-api-key-here',
  baseURL: 'https://api.winkr.com/v1',
  timeout: 15000
});

// Register user
sdk.auth.register({
  email: 'user@example.com',
  password: 'password123',
  firstName: 'John',
  lastName: 'Doe',
  age: 28,
  gender: 'male'
}).then(user => {
  console.log('User registered:', user);
  
  // Get profile
  return sdk.profile.getProfile();
}).then(profile => {
  console.log('User profile:', profile);
}).catch(error => {
  console.error('SDK Error:', error);
});
```

---

For additional examples or support, contact our developer team at dev-support@winkr.com.