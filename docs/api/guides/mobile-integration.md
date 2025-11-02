# Mobile App Integration Guide

This guide provides comprehensive instructions for integrating the Winkr API into mobile applications (iOS and Android).

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Authentication](#authentication)
- [Core Features](#core-features)
- [Real-time Features](#real-time-features)
- [Push Notifications](#push-notifications)
- [Offline Support](#offline-support)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Prerequisites

### Requirements

- **Mobile OS**: iOS 12+ or Android 8+ (API level 26+)
- **Network**: HTTPS required for all API calls
- **Storage**: Local storage for tokens and user data
- **Permissions**: Camera, storage, location (as needed)

### Development Tools

- **iOS**: Xcode 12+, Swift 5+, CocoaPods or Swift Package Manager
- **Android**: Android Studio 4.0+, Kotlin 1.4+ or Java 8+, Gradle 7.0+
- **API Key**: Contact api@winkr.com to request access
- **Testing**: Physical device or emulator with network access

## Setup

### 1. iOS Project Setup

#### CocoaPods Integration

```ruby
# Podfile
platform :ios, '12.0'
use_frameworks!

target 'WinkrApp' do
  pod 'Alamofire', '~> 5.0'
  pod 'Socket.IO-Client-Swift', '~> 15.0'
  pod 'Kingfisher', '~> 7.0'
  pod 'SwiftKeychainWrapper', '~> 4.0'
end
```

#### Swift Package Manager

```swift
// Package.swift
import PackageDescription

let package = Package(
    name: "WinkrApp",
    platforms: [
        .iOS(.v12)
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.0.0"),
        .package(url: "https://github.com/socketio/socket.io-client-swift", from: "15.0.0"),
        .package(url: "https://github.com/onevcat/Kingfisher", from: "7.0.0")
    ]
)
```

### 2. Android Project Setup

#### Gradle Dependencies

```gradle
// app/build.gradle
dependencies {
    implementation 'com.squareup.retrofit2:retrofit:2.9.0'
    implementation 'com.squareup.retrofit2:converter-gson:2.9.0'
    implementation 'com.squareup.okhttp3:logging-interceptor:4.9.0'
    implementation 'io.socket:socket.io-client:2.0.1'
    implementation 'com.github.bumptech.glide:glide:4.12.0'
    implementation 'androidx.security:security-crypto:1.1.0-alpha03'
}
```

#### Permissions

```xml
<!-- AndroidManifest.xml -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
```

## Authentication

### 1. iOS Authentication

#### API Client

```swift
// Network/APIClient.swift
import Alamofire
import Foundation
import SwiftKeychainWrapper

class APIClient {
    static let shared = APIClient()
    
    private let baseURL: String
    private let session: Session
    
    private init() {
        self.baseURL = Bundle.main.infoDictionary?["API_BASE_URL"] as? String ?? "https://api.winkr.com/v1"
        
        let configuration = URLSessionConfiguration.default
        configuration.timeoutIntervalForRequest = 30
        self.session = Session(configuration: configuration)
    }
    
    private var accessToken: String? {
        get {
            return KeychainWrapper.standard.string(forKey: "access_token")
        }
        set {
            if let token = newValue {
                KeychainWrapper.standard.set(token, forKey: "access_token")
            } else {
                KeychainWrapper.standard.removeObject(forKey: "access_token")
            }
        }
    }
    
    private var refreshToken: String? {
        get {
            return KeychainWrapper.standard.string(forKey: "refresh_token")
        }
        set {
            if let token = newValue {
                KeychainWrapper.standard.set(token, forKey: "refresh_token")
            } else {
                KeychainWrapper.standard.removeObject(forKey: "refresh_token")
            }
        }
    }
    
    private func createRequest<T: Codable>(
        url: String,
        method: HTTPMethod = .get,
        parameters: T? = nil,
        headers: HTTPHeaders? = nil
    ) -> DataRequest {
        let fullURL = baseURL + url
        
        var requestHeaders = HTTPHeaders()
        requestHeaders["Content-Type"] = "application/json"
        
        if let token = accessToken {
            requestHeaders["Authorization"] = "Bearer \(token)"
        }
        
        if let headers = headers {
            for (key, value) in headers {
                requestHeaders[key] = value
            }
        }
        
        return session.request(
            fullURL,
            method: method,
            parameters: parameters,
            encoding: JSONEncoding.default,
            headers: requestHeaders
        )
    }
    
    private func handleRefreshToken(completion: @escaping (Bool) -> Void) {
        guard let refreshToken = refreshToken else {
            completion(false)
            return
        }
        
        createRequest(
            url: "/auth/refresh",
            method: .post,
            parameters: ["refresh_token": refreshToken]
        ).responseJSON { response in
            switch response.result {
            case .success(let data):
                if let json = data as? [String: Any],
                   let tokens = json["data"] as? [String: Any],
                   let accessToken = tokens["access_token"] as? String,
                   let refreshToken = tokens["refresh_token"] as? String {
                    
                    self.accessToken = accessToken
                    self.refreshToken = refreshToken
                    completion(true)
                } else {
                    completion(false)
                }
                
            case .failure(let error):
                print("Token refresh failed: \(error)")
                completion(false)
            }
        }
    }
    
    @discardableResult
    private func makeRequest<T: Codable>(
        url: String,
        method: HTTPMethod = .get,
        parameters: T? = nil,
        headers: HTTPHeaders? = nil,
        completion: @escaping (Result<Data, APIError>) -> Void
    ) -> DataRequest {
        
        let request = createRequest(url: url, method: method, parameters: parameters, headers: headers)
        
        return request.validate().responseData { response in
            switch response.result {
            case .success(let data):
                completion(.success(data))
                
            case .failure(let error):
                if let response = response.response {
                    switch response.statusCode {
                    case 401:
                        // Token expired, try to refresh
                        self.handleRefreshToken { success in
                            if success {
                                // Retry original request
                                let retryRequest = self.createRequest(
                                    url: url,
                                    method: method,
                                    parameters: parameters,
                                    headers: headers
                                )
                                retryRequest.validate().responseData { retryResponse in
                                    switch retryResponse.result {
                                    case .success(let data):
                                        completion(.success(data))
                                    case .failure(let error):
                                        completion(.failure(.authenticationError(error.localizedDescription)))
                                    }
                                }
                            } else {
                                completion(.failure(.authenticationError("Token refresh failed")))
                            }
                        }
                        
                    default:
                        completion(.failure(.apiError(error.localizedDescription)))
                    }
                } else {
                    completion(.failure(.networkError(error.localizedDescription)))
                }
            }
        }
    }
}

// Models/APIError.swift
enum APIError: Error, LocalizedError {
    case networkError(String)
    case authenticationError(String)
    case apiError(String)
    case validationError(String)
    
    var errorDescription: String? {
        switch self {
        case .networkError(let message),
             .authenticationError(let message),
             .apiError(let message),
             .validationError(let message):
            return message
        }
    }
}
```

#### Authentication Service

```swift
// Services/AuthService.swift
import Foundation

class AuthService {
    static let shared = AuthService()
    
    private init() {}
    
    struct RegisterRequest: Codable {
        let email: String
        let password: String
        let username: String
        let date_of_birth: String
        let gender: String
    }
    
    struct LoginRequest: Codable {
        let email: String
        let password: String
        let device_info: DeviceInfo
    }
    
    struct DeviceInfo: Codable {
        let device_id: String
        let device_type: String
        let os: String
        let app_version: String
    }
    
    func register(
        email: String,
        password: String,
        username: String,
        dateOfBirth: Date,
        gender: String,
        completion: @escaping (Result<User, APIError>) -> Void
    ) {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        
        let request = RegisterRequest(
            email: email,
            password: password,
            username: username,
            date_of_birth: formatter.string(from: dateOfBirth),
            gender: gender
        )
        
        APIClient.shared.makeRequest(
            url: "/auth/register",
            method: .post,
            parameters: request
        ) { result in
            switch result {
            case .success(let data):
                do {
                    let response = try JSONDecoder().decode(APIResponse<User>.self, from: data)
                    if response.success {
                        completion(.success(response.data))
                    } else {
                        completion(.failure(.apiError(response.error?.message ?? "Registration failed")))
                    }
                } catch {
                    completion(.failure(.networkError("Failed to parse response")))
                }
                
            case .failure(let error):
                completion(.failure(error))
            }
        }
    }
    
    func login(
        email: String,
        password: String,
        completion: @escaping (Result<User, APIError>) -> Void
    ) {
        let deviceInfo = DeviceInfo(
            device_id: UIDevice.current.identifierForVendor ?? UUID().uuidString,
            device_type: "mobile",
            os: UIDevice.current.systemVersion,
            app_version: Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "1.0.0"
        )
        
        let request = LoginRequest(
            email: email,
            password: password,
            device_info: deviceInfo
        )
        
        APIClient.shared.makeRequest(
            url: "/auth/login",
            method: .post,
            parameters: request
        ) { result in
            switch result {
            case .success(let data):
                do {
                    let response = try JSONDecoder().decode(APIResponse<AuthResponse>.self, from: data)
                    if response.success {
                        // Store tokens
                        APIClient.shared.accessToken = response.data.tokens.access_token
                        APIClient.shared.refreshToken = response.data.tokens.refresh_token
                        
                        completion(.success(response.data.user))
                    } else {
                        completion(.failure(.apiError(response.error?.message ?? "Login failed")))
                    }
                } catch {
                    completion(.failure(.networkError("Failed to parse response")))
                }
                
            case .failure(let error):
                completion(.failure(error))
            }
        }
    }
    
    func logout(completion: @escaping (Bool) -> Void) {
        guard let refreshToken = APIClient.shared.refreshToken else {
            completion(false)
            return
        }
        
        APIClient.shared.makeRequest(
            url: "/auth/logout",
            method: .post,
            parameters: ["refresh_token": refreshToken]
        ) { result in
            // Clear tokens regardless of API response
            APIClient.shared.accessToken = nil
            APIClient.shared.refreshToken = nil
            
            switch result {
            case .success:
                completion(true)
            case .failure:
                completion(false)
            }
        }
    }
}

// Models/APIResponse.swift
struct APIResponse<T: Codable>: Codable {
    let success: Bool
    let data: T?
    let error: APIErrorResponse?
}

struct AuthResponse: Codable {
    let user: User
    let tokens: TokenResponse
}

struct TokenResponse: Codable {
    let access_token: String
    let refresh_token: String
    let expires_in: Int
}

struct APIErrorResponse: Codable {
    let code: String
    let message: String
    let details: [String: Any]?
}
```

### 2. Android Authentication

#### API Client

```kotlin
// Network/APIClient.kt
import okhttp3.*
import retrofit2.*
import retrofit2.converter.gson.GsonConverterFactory
import android.content.Context
import androidx.security.crypto.EncryptedSharedPreferences

class APIClient private constructor(context: Context) {
    
    private val baseURL = "https://api.winkr.com/v1"
    private val sharedPreferences = EncryptedSharedPreferences.create(
        "winkr_prefs",
        MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
    )
    
    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(HttpLoggingInterceptor().apply {
            level = if (BuildConfig.DEBUG) {
                HttpLoggingInterceptor.Level.BODY
            } else {
                HttpLoggingInterceptor.Level.NONE
            }
        })
        .addInterceptor(AuthInterceptor())
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .writeTimeout(30, TimeUnit.SECONDS)
        .build()
    
    private val retrofit = Retrofit.Builder()
        .baseUrl(baseURL)
        .client(okHttpClient)
        .addConverterFactory(GsonConverterFactory.create())
        .build()
    
    val apiService: WinkrApiService = retrofit.create(WinkrApiService::class.java)
    
    companion object {
        @Volatile
        private var INSTANCE: APIClient? = null
        
        fun getInstance(context: Context): APIClient {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: APIClient(context).also { INSTANCE = it }
            }
        }
    }
    
    private inner class AuthInterceptor : Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            val originalRequest = chain.request()
            val requestBuilder = originalRequest.newBuilder()
            
            // Add auth token if available
            getAccessToken()?.let { token ->
                requestBuilder.addHeader("Authorization", "Bearer $token")
            }
            
            val request = requestBuilder.build()
            val response = chain.proceed(request)
            
            // Handle 401 responses
            if (response.code == 401) {
                handleTokenRefresh()
            }
            
            return response
        }
    }
    
    private fun getAccessToken(): String? {
        return sharedPreferences.getString("access_token", null)
    }
    
    private fun setAccessToken(token: String?) {
        sharedPreferences.edit()
            .putString("access_token", token)
            .apply()
    }
    
    private fun getRefreshToken(): String? {
        return sharedPreferences.getString("refresh_token", null)
    }
    
    private fun setRefreshToken(token: String?) {
        sharedPreferences.edit()
            .putString("refresh_token", token)
            .apply()
    }
    
    private fun handleTokenRefresh() {
        getRefreshToken()?.let { refreshToken ->
            val call = apiService.refreshToken(RefreshTokenRequest(refreshToken))
            call.enqueue(object : Callback<AuthResponse> {
                override fun onResponse(call: Call<AuthResponse>, response: Response<AuthResponse>) {
                    if (response.isSuccessful) {
                        response.body()?.let { authResponse ->
                            setAccessToken(authResponse.tokens.accessToken)
                            setRefreshToken(authResponse.tokens.refreshToken)
                        }
                    }
                }
                
                override fun onFailure(call: Call<AuthResponse>, t: Throwable) {
                    Log.e("APIClient", "Token refresh failed", t)
                }
            })
        }
    }
}

// Network/WinkrApiService.kt
import retrofit2.*
import retrofit2.http.*

interface WinkrApiService {
    
    @POST("/auth/register")
    fun register(@Body request: RegisterRequest): Call<APIResponse<User>>
    
    @POST("/auth/login")
    fun login(@Body request: LoginRequest): Call<APIResponse<AuthResponse>>
    
    @POST("/auth/refresh")
    fun refreshToken(@Body request: RefreshTokenRequest): Call<AuthResponse>
    
    @POST("/auth/logout")
    fun logout(@Body request: LogoutRequest): Call<APIResponse<Any>>
    
    @GET("/profile/me")
    fun getProfile(): Call<APIResponse<User>>
    
    @PUT("/profile/me")
    fun updateProfile(@Body request: UpdateProfileRequest): Call<APIResponse<User>>
    
    @GET("/discovery/users")
    fun getDiscoveryUsers(@QueryMap params: Map<String, String>): Call<APIResponse<DiscoveryResponse>>
    
    @POST("/discovery/swipe")
    fun swipe(@Body request: SwipeRequest): Call<APIResponse<SwipeResponse>>
    
    @GET("/chat/conversations")
    fun getConversations(@QueryMap params: Map<String, String>): Call<APIResponse<ConversationListResponse>>
    
    @POST("/chat/conversations/{id}/messages")
    fun sendMessage(@Path("id") conversationId: String, @Body request: MessageRequest): Call<APIResponse<Message>>
}

// Data Models
data class RegisterRequest(
    val email: String,
    val password: String,
    val username: String,
    val date_of_birth: String,
    val gender: String
)

data class LoginRequest(
    val email: String,
    val password: String,
    val device_info: DeviceInfo
)

data class DeviceInfo(
    val device_id: String,
    val device_type: String,
    val os: String,
    val app_version: String
)

data class RefreshTokenRequest(
    val refresh_token: String
)

data class AuthResponse(
    val success: Boolean,
    val data: AuthData,
    val error: ErrorResponse?
)

data class AuthData(
    val user: User,
    val tokens: TokenData
)

data class TokenData(
    val access_token: String,
    val refresh_token: String,
    val expires_in: Int
)

data class ErrorResponse(
    val code: String,
    val message: String,
    val details: Map<String, Any>?
)
```

#### Authentication Service

```kotlin
// Services/AuthService.kt
import android.content.Context

class AuthService private constructor(context: Context) {
    
    private val apiClient = APIClient.getInstance(context)
    
    companion object {
        @Volatile
        private var INSTANCE: AuthService? = null
        
        fun getInstance(context: Context): AuthService {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: AuthService(context).also { INSTANCE = it }
            }
        }
    }
    
    fun register(
        email: String,
        password: String,
        username: String,
        dateOfBirth: Date,
        gender: String,
        callback: AuthCallback<User>
    ) {
        val formatter = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault())
        val request = RegisterRequest(
            email = email,
            password = password,
            username = username,
            date_of_birth = formatter.format(dateOfBirth),
            gender = gender
        )
        
        apiClient.apiService.register(request).enqueue(object : Callback<APIResponse<User>> {
            override fun onResponse(call: Call<APIResponse<User>>, response: Response<APIResponse<User>>) {
                if (response.isSuccessful) {
                    response.body()?.let { apiResponse ->
                        if (apiResponse.success) {
                            apiResponse.data?.let { user ->
                                callback.onSuccess(user)
                            } ?: callback.onError("Registration failed")
                        } ?: callback.onError("Invalid response")
                    }
                } else {
                    callback.onError("Registration failed")
                }
            }
            
            override fun onFailure(call: Call<APIResponse<User>>, t: Throwable) {
                callback.onError(t.message ?: "Network error")
            }
        })
    }
    
    fun login(
        email: String,
        password: String,
        callback: AuthCallback<User>
    ) {
        val deviceInfo = DeviceInfo(
            device_id = getDeviceId(),
            device_type = "mobile",
            os = Build.VERSION.RELEASE,
            app_version = BuildConfig.VERSION_NAME
        )
        
        val request = LoginRequest(
            email = email,
            password = password,
            device_info = deviceInfo
        )
        
        apiClient.apiService.login(request).enqueue(object : Callback<APIResponse<AuthResponse>> {
            override fun onResponse(call: Call<APIResponse<AuthResponse>>, response: Response<APIResponse<AuthResponse>>) {
                if (response.isSuccessful) {
                    response.body()?.let { apiResponse ->
                        if (apiResponse.success) {
                            apiResponse.data?.let { authData ->
                                // Store tokens
                                storeTokens(authData.tokens)
                                callback.onSuccess(authData.user)
                            } ?: callback.onError("Login failed")
                        } ?: callback.onError("Invalid response")
                    }
                } else {
                    callback.onError("Login failed")
                }
            }
            
            override fun onFailure(call: Call<APIResponse<AuthResponse>>, t: Throwable) {
                callback.onError(t.message ?: "Network error")
            }
        })
    }
    
    fun logout(callback: AuthCallback<Any>) {
        val refreshToken = getRefreshToken()
        if (refreshToken == null) {
            callback.onError("No refresh token available")
            return
        }
        
        val request = LogoutRequest(refreshToken)
        
        apiClient.apiService.logout(request).enqueue(object : Callback<APIResponse<Any>> {
            override fun onResponse(call: Call<APIResponse<Any>>, response: Response<APIResponse<Any>>) {
                // Clear tokens regardless of API response
                clearTokens()
                
                if (response.isSuccessful) {
                    callback.onSuccess(Any())
                } else {
                    callback.onError("Logout failed")
                }
            }
            
            override fun onFailure(call: Call<APIResponse<Any>>, t: Throwable) {
                clearTokens()
                callback.onError(t.message ?: "Network error")
            }
        })
    }
    
    private fun storeTokens(tokens: TokenData) {
        val sharedPreferences = apiClient.sharedPreferences
        sharedPreferences.edit()
            .putString("access_token", tokens.accessToken)
            .putString("refresh_token", tokens.refreshToken)
            .apply()
    }
    
    private fun clearTokens() {
        val sharedPreferences = apiClient.sharedPreferences
        sharedPreferences.edit()
            .remove("access_token")
            .remove("refresh_token")
            .apply()
    }
    
    private fun getDeviceId(): String {
        return Settings.Secure.getString(
            apiClient.context.contentResolver,
            Settings.Secure.ANDROID_ID
        ) ?: UUID.randomUUID().toString()
    }
}

interface AuthCallback<T> {
    fun onSuccess(data: T)
    fun onError(error: String)
}
```

## Core Features

### 1. Profile Management

#### iOS Profile Service

```swift
// Services/ProfileService.swift
import Foundation
import UIKit
import Kingfisher

class ProfileService {
    static let shared = ProfileService()
    
    private init() {}
    
    func updateProfile(
        firstName: String?,
        lastName: String?,
        bio: String?,
        avatar: UIImage?,
        completion: @escaping (Result<User, APIError>) -> Void
    ) {
        var parameters: [String: Any] = [:]
        
        if let firstName = firstName {
            parameters["first_name"] = firstName
        }
        
        if let lastName = lastName {
            parameters["last_name"] = lastName
        }
        
        if let bio = bio {
            parameters["bio"] = bio
        }
        
        // Handle avatar upload separately if provided
        if let avatar = avatar {
            uploadAvatar(avatar) { result in
                switch result {
                case .success(let avatarUrl):
                    parameters["avatar_url"] = avatarUrl
                    self.updateProfileData(parameters: parameters, completion: completion)
                    
                case .failure(let error):
                    completion(.failure(error))
                }
            }
        } else {
            updateProfileData(parameters: parameters, completion: completion)
        }
    }
    
    private func updateProfileData(
        parameters: [String: Any],
        completion: @escaping (Result<User, APIError>) -> Void
    ) {
        APIClient.shared.makeRequest(
            url: "/profile/me",
            method: .put,
            parameters: parameters
        ) { result in
            switch result {
            case .success(let data):
                do {
                    let response = try JSONDecoder().decode(APIResponse<User>.self, from: data)
                    if response.success {
                        completion(.success(response.data))
                    } else {
                        completion(.failure(.apiError(response.error?.message ?? "Profile update failed")))
                    }
                } catch {
                    completion(.failure(.networkError("Failed to parse response")))
                }
                
            case .failure(let error):
                completion(.failure(error))
            }
        }
    }
    
    private func uploadAvatar(
        _ image: UIImage,
        completion: @escaping (Result<String, APIError>) -> Void
    ) {
        guard let imageData = image.jpegData(compressionQuality: 0.8) else {
            completion(.failure(.validationError("Failed to process image")))
            return
        }
        
        // Create upload request
        let url = URL(string: APIClient.shared.baseURL + "/photos/upload-url")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let uploadRequest = [
            "filename": "avatar.jpg",
            "content_type": "image/jpeg"
        ]
        
        do {
            request.httpBody = try JSONSerialization.data(withJSONObject: uploadRequest)
        } catch {
            completion(.failure(.networkError("Failed to create request")))
            return
        }
        
        URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(.networkError(error.localizedDescription)))
                return
            }
            
            guard let data = data else {
                completion(.failure(.networkError("No data received")))
                return
            }
            
            do {
                let uploadResponse = try JSONDecoder().decode(UploadResponse.self, from: data)
                if uploadResponse.success {
                    self.uploadImageToS3(
                        imageData: imageData,
                        uploadURL: uploadResponse.data.uploadUrl,
                        completion: completion
                    )
                } else {
                    completion(.failure(.apiError(uploadResponse.error?.message ?? "Upload failed")))
                }
            } catch {
                completion(.failure(.networkError("Failed to parse response")))
            }
        }.resume()
    }
    
    private func uploadImageToS3(
        imageData: Data,
        uploadURL: String,
        completion: @escaping (Result<String, APIError>) -> Void
    ) {
        guard let url = URL(string: uploadURL) else {
            completion(.failure(.networkError("Invalid upload URL")))
            return
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"
        request.setValue("image/jpeg", forHTTPHeaderField: "Content-Type")
        request.httpBody = imageData
        
        URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                completion(.failure(.networkError(error.localizedDescription)))
                return
            }
            
            if let httpResponse = response as? HTTPURLResponse,
               httpResponse.statusCode == 200 {
                // Extract photo URL from upload URL
                let photoURL = uploadURL.components(separatedBy: "?").first ?? uploadURL
                completion(.success(photoURL))
            } else {
                completion(.failure(.networkError("Upload failed")))
            }
        }.resume()
    }
}

struct UploadResponse: Codable {
    let success: Bool
    let data: UploadData
    let error: APIErrorResponse?
}

struct UploadData: Codable {
    let uploadUrl: String
    let photoId: String
}
```

#### Android Profile Service

```kotlin
// Services/ProfileService.kt
import android.content.Context
import android.graphics.Bitmap
import android.net.Uri
import com.bumptech.glide.Glide
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.FileOutputStream

class ProfileService private constructor(context: Context) {
    
    private val apiClient = APIClient.getInstance(context)
    private val context = context
    
    companion object {
        @Volatile
        private var INSTANCE: ProfileService? = null
        
        fun getInstance(context: Context): ProfileService {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: ProfileService(context).also { INSTANCE = it }
            }
        }
    }
    
    fun updateProfile(
        firstName: String?,
        lastName: String?,
        bio: String?,
        avatarUri: Uri?,
        callback: ProfileCallback
    ) {
        val parameters = mutableMapOf<String, Any>()
        
        firstName?.let { parameters["first_name"] = it }
        lastName?.let { parameters["last_name"] = it }
        bio?.let { parameters["bio"] = it }
        
        if (avatarUri != null) {
            uploadAvatar(avatarUri!!) { result ->
                when (result) {
                    is ProfileResult.Success -> {
                        parameters["avatar_url"] = result.data
                        updateProfileData(parameters, callback)
                    }
                    is ProfileResult.Error -> {
                        callback.onError(result.message)
                    }
                }
            }
        } else {
            updateProfileData(parameters, callback)
        }
    }
    
    private fun updateProfileData(
        parameters: Map<String, Any>,
        callback: ProfileCallback
    ) {
        apiClient.apiService.updateProfile(parameters).enqueue(object : Callback<APIResponse<User>> {
            override fun onResponse(call: Call<APIResponse<User>>, response: Response<APIResponse<User>>) {
                if (response.isSuccessful) {
                    response.body()?.let { apiResponse ->
                        if (apiResponse.success) {
                            apiResponse.data?.let { user ->
                                callback.onSuccess(user)
                            } ?: callback.onError("Profile update failed")
                        } ?: callback.onError("Invalid response")
                    }
                } else {
                    callback.onError("Profile update failed")
                }
            }
            
            override fun onFailure(call: Call<APIResponse<User>>, t: Throwable) {
                callback.onError(t.message ?: "Network error")
            }
        })
    }
    
    private fun uploadAvatar(
        avatarUri: Uri,
        callback: ProfileResultCallback
    ) {
        try {
            // Get bitmap from URI
            val bitmap = Glide.with(context)
                .asBitmap()
                .load(avatarUri)
                .submit()
                .get()
            
            // Compress bitmap
            val outputStream = ByteArrayOutputStream()
            bitmap.compress(Bitmap.CompressFormat.JPEG, 80, outputStream)
            val imageData = outputStream.toByteArray()
            
            // Get upload URL
            val uploadRequest = mapOf(
                "filename" to "avatar.jpg",
                "content_type" to "image/jpeg"
            )
            
            apiClient.apiService.getUploadUrl(uploadRequest).enqueue(object : Callback<APIResponse<UploadData>> {
                override fun onResponse(call: Call<APIResponse<UploadData>>, response: Response<APIResponse<UploadData>>) {
                    if (response.isSuccessful) {
                        response.body()?.let { apiResponse ->
                            if (apiResponse.success) {
                                apiResponse.data?.let { uploadData ->
                                    uploadImageToS3(imageData, uploadData.uploadUrl, callback)
                                } ?: callback.onError("Invalid upload response")
                            }
                        } ?: callback.onError("Invalid response")
                    } else {
                        callback.onError("Failed to get upload URL")
                    }
                }
                
                override fun onFailure(call: Call<APIResponse<UploadData>>, t: Throwable) {
                    callback.onError(t.message ?: "Network error")
                }
            })
        } catch (e: Exception) {
            callback.onError("Failed to process image: ${e.message}")
        }
    }
    
    private fun uploadImageToS3(
        imageData: ByteArray,
        uploadURL: String,
        callback: ProfileResultCallback
    ) {
        try {
            val url = URL(uploadURL)
            val connection = url.openConnection() as HttpURLConnection
            connection.requestMethod = "PUT"
            connection.setRequestProperty("Content-Type", "image/jpeg")
            connection.doOutput = true
            
            connection.outputStream.use { output ->
                output.write(imageData)
            }
            
            if (connection.responseCode == 200) {
                // Extract photo URL from upload URL
                val photoURL = uploadURL.split("?").first()
                callback.onSuccess(photoURL)
            } else {
                callback.onError("Upload failed with code: ${connection.responseCode}")
            }
        } catch (e: Exception) {
            callback.onError("Upload failed: ${e.message}")
        }
    }
}

sealed class ProfileResult {
    data class Success(val data: String) : ProfileResult()
    data class Error(val message: String) : ProfileResult()
}

interface ProfileCallback {
    fun onSuccess(user: User)
    fun onError(error: String)
}

interface ProfileResultCallback {
    fun onSuccess(data: String)
    fun onError(message: String)
}
```

## Real-time Features

### 1. iOS WebSocket Integration

```swift
// Services/WebSocketService.swift
import Foundation
import SocketIO

class WebSocketService: NSObject {
    static let shared = WebSocketService()
    
    private var manager: SocketManager?
    private var isConnected = false
    
    private override init() {
        super.init()
    }
    
    func connect(token: String) {
        guard let url = URL(string: "wss://api.winkr.com/ws") else {
            return
        }
        
        manager = SocketManager(socketURL: url, config: [
            .log: true,
            .compress: true,
            .connectParams: ["token": token]
        ])
        
        manager?.defaultSocket.connect()
        
        setupEventHandlers()
    }
    
    private func setupEventHandlers() {
        manager?.defaultSocket.on(clientEvent: .connect) { [weak self] data, ack in
            self?.isConnected = true
            print("WebSocket connected")
        }
        
        manager?.defaultSocket.on(clientEvent: .disconnect) { [weak self] data, ack in
            self?.isConnected = false
            print("WebSocket disconnected")
        }
        
        manager?.defaultSocket.on("message:new") { [weak self] data, ack in
            self?.handleNewMessage(data)
        }
        
        manager?.defaultSocket.on("message:viewed") { [weak self] data, ack in
            self?.handleMessageViewed(data)
        }
        
        manager?.defaultSocket.on("typing:indicator") { [weak self] data, ack in
            self?.handleTypingIndicator(data)
        }
    }
    
    func joinConversation(conversationId: String) {
        manager?.defaultSocket.emit("conversation:join", ["conversation_id": conversationId])
    }
    
    func leaveConversation(conversationId: String) {
        manager?.defaultSocket.emit("conversation:leave", ["conversation_id": conversationId])
    }
    
    func sendMessage(conversationId: String, content: String) {
        manager?.defaultSocket.emit("message:send", [
            "conversation_id": conversationId,
            "content": content,
            "type": "text"
        ])
    }
    
    func startTyping(conversationId: String) {
        manager?.defaultSocket.emit("typing:start", ["conversation_id": conversationId])
    }
    
    func stopTyping(conversationId: String) {
        manager?.defaultSocket.emit("typing:stop", ["conversation_id": conversationId])
    }
    
    private func handleNewMessage(_ data: [String: Any]) {
        NotificationCenter.default.post(
            name: .newMessage,
            object: nil,
            userInfo: data
        )
    }
    
    private func handleMessageViewed(_ data: [String: Any]) {
        NotificationCenter.default.post(
            name: .messageViewed,
            object: nil,
            userInfo: data
        )
    }
    
    private func handleTypingIndicator(_ data: [String: Any]) {
        NotificationCenter.default.post(
            name: .typingIndicator,
            object: nil,
            userInfo: data
        )
    }
    
    func disconnect() {
        manager?.defaultSocket.disconnect()
        isConnected = false
    }
}

// Extensions/NotificationNames.swift
extension Notification.Name {
    static let newMessage = Notification.Name("NewMessage")
    static let messageViewed = Notification.Name("MessageViewed")
    static let typingIndicator = Notification.Name("TypingIndicator")
}
```

### 2. Android WebSocket Integration

```kotlin
// Services/WebSocketService.kt
import io.socket.client.IO
import io.socket.client.Socket
import org.json.JSONObject

class WebSocketService private constructor() {
    
    private var socket: Socket? = null
    private var isConnected = false
    
    companion object {
        @Volatile
        private var INSTANCE: WebSocketService? = null
        
        fun getInstance(): WebSocketService {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: WebSocketService().also { INSTANCE = it }
            }
        }
    }
    
    fun connect(token: String) {
        try {
            val opts = IO.Options()
            opts.path = "/ws"
            
            socket = IO.socket("wss://api.winkr.com", opts)
            
            setupEventHandlers()
            socket?.connect()
            
        } catch (e: Exception) {
            Log.e("WebSocketService", "Failed to connect: ${e.message}")
        }
    }
    
    private fun setupEventHandlers() {
        socket?.on(Socket.EVENT_CONNECT) { args ->
            isConnected = true
            Log.d("WebSocketService", "Connected to WebSocket")
        }
        
        socket?.on(Socket.EVENT_DISCONNECT) { args ->
            isConnected = false
            Log.d("WebSocketService", "Disconnected from WebSocket")
        }
        
        socket?.on("message:new") { args ->
            handleNewMessage(args)
        }
        
        socket?.on("message:viewed") { args ->
            handleMessageViewed(args)
        }
        
        socket?.on("typing:indicator") { args ->
            handleTypingIndicator(args)
        }
    }
    
    fun joinConversation(conversationId: String) {
        val data = JSONObject().apply {
            put("conversation_id", conversationId)
        }
        socket?.emit("conversation:join", data)
    }
    
    fun leaveConversation(conversationId: String) {
        val data = JSONObject().apply {
            put("conversation_id", conversationId)
        }
        socket?.emit("conversation:leave", data)
    }
    
    fun sendMessage(conversationId: String, content: String) {
        val data = JSONObject().apply {
            put("conversation_id", conversationId)
            put("content", content)
            put("type", "text")
        }
        socket?.emit("message:send", data)
    }
    
    fun startTyping(conversationId: String) {
        val data = JSONObject().apply {
            put("conversation_id", conversationId)
        }
        socket?.emit("typing:start", data)
    }
    
    fun stopTyping(conversationId: String) {
        val data = JSONObject().apply {
            put("conversation_id", conversationId)
        }
        socket?.emit("typing:stop", data)
    }
    
    private fun handleNewMessage(args: Array<Any>) {
        if (args.isNotEmpty()) {
            val data = args[0] as? JSONObject
            data?.let { 
                // Broadcast to app components
                EventBus.getDefault().post(NewMessageEvent(it))
            }
        }
    }
    
    private fun handleMessageViewed(args: Array<Any>) {
        if (args.isNotEmpty()) {
            val data = args[0] as? JSONObject
            data?.let {
                EventBus.getDefault().post(MessageViewedEvent(it))
            }
        }
    }
    
    private fun handleTypingIndicator(args: Array<Any>) {
        if (args.isNotEmpty()) {
            val data = args[0] as? JSONObject
            data?.let {
                EventBus.getDefault().post(TypingIndicatorEvent(it))
            }
        }
    }
    
    fun disconnect() {
        socket?.disconnect()
        isConnected = false
    }
}

// EventBus Events
data class NewMessageEvent(val data: JSONObject)
data class MessageViewedEvent(val data: JSONObject)
data class TypingIndicatorEvent(val data: JSONObject)
```

## Push Notifications

### 1. iOS Push Notifications

```swift
// Services/PushNotificationService.swift
import UserNotifications
import UIKit

class PushNotificationService: NSObject {
    static let shared = PushNotificationService()
    
    private override init() {
        super.init()
    }
    
    func registerForPushNotifications() {
        // Request permission
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .badge, .sound]) { granted, error in
            if granted {
                self.registerDeviceToken()
            } else {
                print("Push notification permission denied")
            }
        }
    }
    
    private func registerDeviceToken() {
        UIApplication.shared.registerForRemoteNotifications()
    }
    
    func sendDeviceTokenToServer(token: String) {
        guard let userId = getCurrentUserId() else { return }
        
        let parameters = [
            "user_id": userId,
            "device_token": token,
            "platform": "ios"
        ]
        
        APIClient.shared.makeRequest(
            url: "/devices/register",
            method: .post,
            parameters: parameters
        ) { result in
            switch result {
            case .success:
                print("Device token registered successfully")
            case .failure(let error):
                print("Failed to register device token: \(error)")
            }
        }
    }
    
    func handlePushNotification(userInfo: [AnyHashable: Any]) {
        guard let aps = userInfo["aps"] as? [String: Any],
              let alert = aps["alert"] as? [String: Any],
              let body = alert["body"] as? String else {
            return
        }
        
        // Handle different notification types
        if let type = userInfo["type"] as? String {
            switch type {
            case "new_message":
                handleNewMessageNotification(userInfo)
            case "new_match":
                handleNewMatchNotification(userInfo)
            case "profile_view":
                handleProfileViewNotification(userInfo)
            default:
                break
            }
        }
        
        // Show local notification if app is in background
        if UIApplication.shared.applicationState == .background {
            showLocalNotification(title: "Winkr", body: body, userInfo: userInfo)
        }
    }
    
    private func handleNewMessageNotification(_ userInfo: [AnyHashable: Any]) {
        // Navigate to chat
        if let conversationId = userInfo["conversation_id"] as? String {
            NotificationCenter.default.post(
                name: .newMessageNotification,
                object: nil,
                userInfo: userInfo
            )
        }
    }
    
    private func handleNewMatchNotification(_ userInfo: [AnyHashable: Any]) {
        // Navigate to match screen
        NotificationCenter.default.post(
            name: .newMatchNotification,
            object: nil,
            userInfo: userInfo
        )
    }
    
    private func handleProfileViewNotification(_ userInfo: [AnyHashable: Any]) {
        // Show profile view alert
        NotificationCenter.default.post(
            name: .profileViewNotification,
            object: nil,
            userInfo: userInfo
        )
    }
    
    private func showLocalNotification(title: String, body: String, userInfo: [AnyHashable: Any]) {
        let content = UNMutableNotificationContent()
        content.title = title
        content.body = body
        content.sound = UNNotificationSound.default
        content.userInfo = userInfo
        
        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)
        
        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                print("Error showing notification: \(error)")
            }
        }
    }
}

// AppDelegate.swift
import UIKit

class AppDelegate: UIResponder, UIApplicationDelegate {
    
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        
        // Register for push notifications
        PushNotificationService.shared.registerForPushNotifications()
        
        return true
    }
    
    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        let token = deviceToken.map { String(format: "%02x", $0) }.joined()
        PushNotificationService.shared.sendDeviceTokenToServer(token: token)
    }
    
    func application(_ application: UIApplication, didReceiveRemoteNotification userInfo: [AnyHashable: Any], fetchCompletionHandler completionHandler: @escaping (UIBackgroundFetchResult) -> Void) {
        PushNotificationService.shared.handlePushNotification(userInfo: userInfo)
        completionHandler(.newData)
    }
}
```

### 2. Android Push Notifications

```kotlin
// Services/PushNotificationService.kt
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage

class PushNotificationService : FirebaseMessagingService() {
    
    override fun onNewToken(token: String) {
        super.onNewToken(token)
        sendDeviceTokenToServer(token)
    }
    
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        super.onMessageReceived(remoteMessage)
        
        val data = remoteMessage.data
        val type = data["type"]
        
        when (type) {
            "new_message" -> handleNewMessageNotification(data)
            "new_match" -> handleNewMatchNotification(data)
            "profile_view" -> handleProfileViewNotification(data)
            else -> handleGenericNotification(data)
        }
    }
    
    private fun sendDeviceTokenToServer(token: String) {
        val userId = getCurrentUserId() ?: return
        
        val parameters = mapOf(
            "user_id" to userId,
            "device_token" to token,
            "platform" to "android"
        )
        
        APIClient.getInstance(this).apiService.registerDevice(parameters).enqueue(object : Callback<APIResponse<Any>> {
            override fun onResponse(call: Call<APIResponse<Any>>, response: Response<APIResponse<Any>>) {
                if (response.isSuccessful) {
                    Log.d("PushNotificationService", "Device token registered successfully")
                } else {
                    Log.e("PushNotificationService", "Failed to register device token")
                }
            }
            
            override fun onFailure(call: Call<APIResponse<Any>>, t: Throwable) {
                Log.e("PushNotificationService", "Failed to register device token: ${t.message}")
            }
        })
    }
    
    private fun handleNewMessageNotification(data: Map<String, String>) {
        val conversationId = data["conversation_id"]
        val message = data["message"]
        
        // Create notification
        val notification = createNotification(
            title = "New Message",
            body = message,
            data = data
        )
        
        // Send to chat activity
        val intent = Intent(this, ChatActivity::class.java).apply {
            putExtra("conversation_id", conversationId)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        notification.contentIntent = pendingIntent
        showNotification(notification)
        
        // Broadcast to app
        EventBus.getDefault().post(NewMessageNotificationEvent(data))
    }
    
    private fun handleNewMatchNotification(data: Map<String, String>) {
        val matchId = data["match_id"]
        val message = data["message"] ?: "You have a new match!"
        
        val notification = createNotification(
            title = "New Match! 🎉",
            body = message,
            data = data
        )
        
        val intent = Intent(this, MatchActivity::class.java).apply {
            putExtra("match_id", matchId)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        val pendingIntent = PendingIntent.getActivity(
            this, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        notification.contentIntent = pendingIntent
        showNotification(notification)
        
        EventBus.getDefault().post(NewMatchNotificationEvent(data))
    }
    
    private fun createNotification(
        title: String,
        body: String,
        data: Map<String, String>
    ): NotificationCompat.Builder {
        val channelId = "winkr_notifications"
        
        return NotificationCompat.Builder(this, channelId)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(title)
            .setContentText(body)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .setDefaults(NotificationCompat.DEFAULT_ALL)
    }
    
    private fun showNotification(notification: NotificationCompat.Builder) {
        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManagerCompat
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                "winkr_notifications",
                "Winkr Notifications",
                NotificationManager.IMPORTANCE_HIGH
            )
            notificationManager.createNotificationChannel(channel)
        }
        
        notificationManager.notify(System.currentTimeMillis().toInt(), notification.build())
    }
}

// Notification Events
data class NewMessageNotificationEvent(val data: Map<String, String>)
data class NewMatchNotificationEvent(val data: Map<String, String>)
data class ProfileViewNotificationEvent(val data: Map<String, String>)
```

## Offline Support

### 1. iOS Offline Support

```swift
// Services/OfflineService.swift
import Foundation
import CoreData

class OfflineService {
    static let shared = OfflineService()
    
    private init() {}
    
    private lazy var persistentContainer: NSPersistentContainer = {
        let container = NSPersistentContainer(name: "WinkrDataModel")
        container.loadPersistentStores { _, error in
            if let error = error {
                fatalError("Core Data error: \(error)")
            }
        }
        return container
    }()
    
    private var context: NSManagedObjectContext {
        return persistentContainer.viewContext
    }
    
    func saveMessageForOffline(_ message: Message) {
        let offlineMessage = OfflineMessage(context: context)
        offlineMessage.id = message.id
        offlineMessage.conversationId = message.conversationId
        offlineMessage.content = message.content
        offlineMessage.type = message.type
        offlineMessage.senderId = message.senderId
        offlineMessage.createdAt = Date()
        offlineMessage.isPending = true
        
        saveContext()
    }
    
    func getPendingMessages() -> [OfflineMessage] {
        let request: NSFetchRequest<OfflineMessage> = OfflineMessage.fetchRequest()
        request.predicate = NSPredicate(format: "isPending == YES")
        
        do {
            return try context.fetch(request)
        } catch {
            print("Failed to fetch pending messages: \(error)")
            return []
        }
    }
    
    func syncPendingMessages() {
        let pendingMessages = getPendingMessages()
        
        for message in pendingMessages {
            // Try to send message
            APIClient.shared.makeRequest(
                url: "/chat/conversations/\(message.conversationId)/messages",
                method: .post,
                parameters: [
                    "content": message.content ?? "",
                    "type": message.type ?? "text"
                ]
            ) { result in
                switch result {
                case .success:
                    // Mark as sent
                    message.isPending = false
                    message.sentAt = Date()
                    
                case .failure:
                    // Keep as pending, will retry later
                    break
                }
            }
        }
        
        saveContext()
    }
    
    private func saveContext() {
        do {
            try context.save()
        } catch {
            print("Failed to save context: \(error)")
        }
    }
}

// Core Data Entity
@objc(OfflineMessage)
public class OfflineMessage: NSManagedObject {
    @NSManaged public var id: String?
    @NSManaged public var conversationId: String?
    @NSManaged public var content: String?
    @NSManaged public var type: String?
    @NSManaged public var senderId: String?
    @NSManaged public var createdAt: Date?
    @NSManaged public var sentAt: Date?
    @NSManaged public var isPending: Bool = true
}
```

### 2. Android Offline Support

```kotlin
// Services/OfflineService.kt
import android.content.Context
import androidx.room.Room
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class OfflineService private constructor(context: Context) {
    
    private val database = Room.databaseBuilder(
        context.applicationContext,
        WinkrDatabase::class.java,
        "winkr_database"
    ).build()
    
    private val messageDao = database.messageDao()
    private val coroutineScope = CoroutineScope(Dispatchers.IO)
    
    companion object {
        @Volatile
        private var INSTANCE: OfflineService? = null
        
        fun getInstance(context: Context): OfflineService {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: OfflineService(context).also { INSTANCE = it }
            }
        }
    }
    
    fun saveMessageForOffline(message: Message) {
        val offlineMessage = OfflineMessage(
            id = message.id,
            conversationId = message.conversationId,
            content = message.content,
            type = message.type,
            senderId = message.senderId,
            createdAt = System.currentTimeMillis(),
            isPending = true
        )
        
        coroutineScope.launch {
            messageDao.insert(offlineMessage)
        }
    }
    
    fun getPendingMessages(): List<OfflineMessage> {
        return try {
            messageDao.getPendingMessages()
        } catch (e: Exception) {
            Log.e("OfflineService", "Failed to get pending messages: ${e.message}")
            emptyList()
        }
    }
    
    fun syncPendingMessages() {
        val pendingMessages = getPendingMessages()
        
        coroutineScope.launch {
            for (message in pendingMessages) {
                try {
                    // Try to send message
                    val parameters = mapOf(
                        "content" to message.content,
                        "type" to message.type
                    )
                    
                    val response = APIClient.getInstance(applicationContext)
                        .apiService.sendMessage(message.conversationId, parameters)
                        .execute()
                    
                    if (response.isSuccessful) {
                        // Mark as sent
                        message.isPending = false
                        message.sentAt = System.currentTimeMillis()
                        messageDao.update(message)
                    }
                } catch (e: Exception) {
                    Log.e("OfflineService", "Failed to sync message: ${e.message}")
                }
            }
        }
    }
}

// Room Database
@Database(
    entities = [OfflineMessage::class],
    version = 1,
    exportSchema = false
)
abstract class WinkrDatabase : RoomDatabase() {
    abstract fun messageDao(): MessageDao
}

// DAO
@Dao
interface MessageDao {
    @Insert
    suspend fun insert(message: OfflineMessage)
    
    @Query("SELECT * FROM offline_message WHERE isPending = 1")
    fun getPendingMessages(): List<OfflineMessage>
    
    @Update
    suspend fun update(message: OfflineMessage)
}

// Entity
@Entity(tableName = "offline_message")
data class OfflineMessage(
    @PrimaryKey val id: String,
    val conversationId: String,
    val content: String,
    val type: String,
    val senderId: String,
    val createdAt: Long,
    var isPending: Boolean = true,
    var sentAt: Long? = null
)
```

## Best Practices

### 1. Security

- **Token Storage**: Use secure storage (Keychain/EncryptedSharedPreferences)
- **Certificate Pinning**: Implement SSL certificate pinning
- **Root/Jailbreak Detection**: Detect and handle compromised devices
- **App Transport Security**: Enforce HTTPS for all network calls
- **Input Validation**: Validate all user inputs on client side

### 2. Performance

- **Image Optimization**: Compress and cache images appropriately
- **Lazy Loading**: Load data only when needed
- **Background Processing**: Use background threads for network operations
- **Memory Management**: Implement proper memory management for images
- **Battery Optimization**: Minimize battery drain

### 3. User Experience

- **Loading States**: Show loading indicators during operations
- **Error Handling**: Provide clear, actionable error messages
- **Offline Support**: Handle offline scenarios gracefully
- **Push Notifications**: Implement timely push notifications
- **Responsive Design**: Ensure good UX on different screen sizes

### 4. Code Quality

- **Architecture**: Use clean architecture (MVVM/MVP)
- **Dependency Injection**: Use dependency injection frameworks
- **Testing**: Write comprehensive unit and integration tests
- **Code Reviews**: Implement code review process
- **Documentation**: Maintain up-to-date code documentation

## Examples

### 1. Complete iOS App Structure

```
WinkrApp/
├── App/
│   ├── AppDelegate.swift
│   ├── SceneDelegate.swift
│   └── Info.plist
├── Core/
│   ├── Network/
│   │   ├── APIClient.swift
│   │   └── WinkrApiService.swift
│   ├── Services/
│   │   ├── AuthService.swift
│   │   ├── ProfileService.swift
│   │   ├── WebSocketService.swift
│   │   └── PushNotificationService.swift
│   ├── Models/
│   │   ├── User.swift
│   │   ├── Message.swift
│   │   └── APIResponse.swift
│   └── Utils/
│       ├── Extensions.swift
│       └── Helpers.swift
├── Features/
│   ├── Authentication/
│   │   ├── LoginViewController.swift
│   │   └── RegisterViewController.swift
│   ├── Profile/
│   │   ├── ProfileViewController.swift
│   │   └── EditProfileViewController.swift
│   ├── Discovery/
│   │   └── DiscoveryViewController.swift
│   └── Chat/
│       ├── ChatViewController.swift
│       └── ConversationViewController.swift
├── Resources/
│   ├── Assets.xcassets
│   ├── Localizable.strings
│   └── Base.lproj/
└── Tests/
    ├── UnitTests/
    └── UITests/
```

### 2. Complete Android App Structure

```
WinkrApp/
├── app/
│   ├── src/
│   │   ├── main/
│   │   │   ├── java/com/winkr/
│   │   │   │   ├── data/
│   │   │   │   │   ├── database/
│   │   │   │   │   │   ├── WinkrDatabase.kt
│   │   │   │   │   ├── entities/
│   │   │   │   │   └── dao/
│   │   │   │   ├── network/
│   │   │   │   │   ├── APIClient.kt
│   │   │   │   │   ├── WinkrApiService.kt
│   │   │   │   └── models/
│   │   │   │   ├── services/
│   │   │   │   │   ├── AuthService.kt
│   │   │   │   │   ├── ProfileService.kt
│   │   │   │   │   ├── WebSocketService.kt
│   │   │   │   │   └── PushNotificationService.kt
│   │   │   │   └── utils/
│   │   │   ├── ui/
│   │   │   │   ├── auth/
│   │   │   │   │   ├── LoginActivity.kt
│   │   │   │   │   └── RegisterActivity.kt
│   │   │   │   ├── profile/
│   │   │   │   │   ├── ProfileActivity.kt
│   │   │   │   │   └── EditProfileActivity.kt
│   │   │   │   ├── discovery/
│   │   │   │   │   └── DiscoveryActivity.kt
│   │   │   │   └── chat/
│   │   │   │       ├── ChatActivity.kt
│   │   │   │       └── ConversationActivity.kt
│   │   │   └── res/
│   │   │       ├── layout/
│   │   │       ├── values/
│   │   │       └── drawable/
│   │   └── build.gradle
├── gradle/
└── tests/
    ├── unit/
    └── integration/
```

This comprehensive mobile integration guide provides everything needed to successfully integrate the Winkr API into iOS and Android applications, from basic setup to advanced features like real-time messaging and push notifications.