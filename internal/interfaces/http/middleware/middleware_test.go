package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

func TestCORS(t *testing.T) {
	// Create Gin router with CORS middleware
	router := gin.New()
	router.Use(CORS(DefaultCORSConfig()))

	// Add test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSOptions(t *testing.T) {
	// Create Gin router with CORS middleware
	router := gin.New()
	router.Use(CORS(DefaultCORSConfig()))

	// Add test route
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create test request
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRequestID(t *testing.T) {
	// Create Gin router with RequestID middleware
	router := gin.New()
	router.Use(RequestID("X-Request-ID"))

	// Add test route
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestSecurityHeaders(t *testing.T) {
	// Create Gin router with Security middleware
	router := gin.New()
	router.Use(SecurityHeaders())

	// Add test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestAuthMiddleware(t *testing.T) {
	// Create JWT utils for testing
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)
	
	// Create test token
	token, err := jwtUtils.GenerateAccessToken("user123", "test@example.com", false)
	assert.NoError(t, err)

	// Create Gin router with Auth middleware
	router := gin.New()
	router.Use(RequireAuth(jwtUtils))

	// Add test route
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "user123", userID)
		c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
	})

	// Test with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test without token
	req2, _ := http.NewRequest("GET", "/protected", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestValidationMiddleware(t *testing.T) {
	// Create Gin router with Validation middleware
	router := gin.New()
	router.Use(Validation(DefaultValidationConfig()))

	// Add test route
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "validated"})
	})

	// Test with valid JSON
	validJSON := `{"name": "test", "email": "test@example.com"}`
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = httptest.NewRecorder().Body
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid JSON
	invalidJSON := `{"name": "test", "email": "invalid-email"}`
	req2, _ := http.NewRequest("POST", "/test", nil)
	req2.Header.Set("Content-Type", "application/json")
	req2.Body = httptest.NewRecorder().Body
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	// Note: This test might need adjustment based on actual validation rules
}

func TestErrorHandlerMiddleware(t *testing.T) {
	// Create Gin router with ErrorHandler middleware
	router := gin.New()
	router.Use(ErrorHandler(DefaultErrorHandlerConfig()))

	// Add test route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Add test route with error
	router.GET("/error", func(c *gin.Context) {
		c.Error(assert.AnError)
	})

	// Test panic recovery
	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Test error handling
	req2, _ := http.NewRequest("GET", "/error", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestRateLimitMiddleware(t *testing.T) {
	// Create mock Redis client
	mockRedis := mock.Mock{}
	redisClient := redis.NewClient(&redis.Options{})

	// Create Gin router with RateLimiter middleware
	router := gin.New()
	
	// Skip Redis dependency for this test
	config := &RateLimitConfig{
		RedisClient:              redisClient,
		DefaultRequestsPerMinute: 5,
		DefaultRequestsPerHour:   100,
		WindowDuration:           time.Minute,
		IncludeHeaders:          true,
		SkipPaths:              []string{},
		KeyPrefix:              "test:",
	}

	router.Use(RateLimiter(config))

	// Add test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create multiple requests to test rate limiting
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if i < 5 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			// After rate limit is exceeded
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			break
		}
	}
}

func TestMiddlewareConfig(t *testing.T) {
	// Test loading middleware configuration
	appConfig := &config.Config{
		App: config.AppConfig{
			Env:  "development",
			Port: 8080,
			Host: "localhost",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret",
			AccessTokenExpiry:  time.Minute,
			RefreshTokenExpiry: time.Hour,
		},
	}

	redisClient := redis.NewClient(&redis.Options{})
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)

	// Load middleware configuration
	middlewareConfig := LoadMiddlewareConfig(appConfig, redisClient, jwtUtils)

	// Check that all middleware configs are loaded
	assert.NotNil(t, middlewareConfig.CORS)
	assert.NotNil(t, middlewareConfig.Logging)
	assert.NotNil(t, middlewareConfig.RateLimit)
	assert.NotNil(t, middlewareConfig.Auth)
	assert.NotNil(t, middlewareConfig.ErrorHandler)
	assert.NotNil(t, middlewareConfig.Validation)
	assert.NotNil(t, middlewareConfig.Security)
}

func TestDevelopmentConfig(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)

	// Test development configuration
	devConfig := DevelopmentConfig(redisClient, jwtUtils)
	assert.NotNil(t, devConfig.CORS)
	assert.NotNil(t, devConfig.Logging)
	assert.True(t, devConfig.Logging.EnableColors)
	assert.True(t, devConfig.Logging.LogRequestBody)
}

func TestProductionConfig(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)

	// Test production configuration
	prodConfig := ProductionConfig(redisClient, jwtUtils)
	assert.NotNil(t, prodConfig.CORS)
	assert.NotNil(t, prodConfig.Logging)
	assert.False(t, prodConfig.Logging.EnableColors)
	assert.False(t, prodConfig.Logging.LogRequestBody)
	assert.True(t, prodConfig.Security.RequireHTTPS)
	assert.True(t, prodConfig.Security.EnableHSTS)
}

func TestTestingConfig(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{})
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)

	// Test testing configuration
	testConfig := TestingConfig(redisClient, jwtUtils)
	assert.NotNil(t, testConfig.CORS)
	assert.NotNil(t, testConfig.Logging)
	assert.False(t, testConfig.Logging.EnableColors)
	assert.False(t, testConfig.Logging.LogRequestBody)
	assert.False(t, testConfig.Security.RequireHTTPS)
}

// BenchmarkMiddleware benchmarks middleware performance
func BenchmarkMiddleware(b *testing.B) {
	// Create Gin router with all middleware
	router := gin.New()
	
	redisClient := redis.NewClient(&redis.Options{})
	jwtUtils := utils.NewJWTUtils("test-secret", time.Minute, time.Hour)
	middlewareConfig := DevelopmentConfig(redisClient, jwtUtils)
	
	router.Use(Security(middlewareConfig.Security))
	router.Use(CORS(middlewareConfig.CORS))
	router.Use(ErrorHandler(middlewareConfig.ErrorHandler))
	router.Use(RequestID(middlewareConfig.Logging.RequestIDHeader))
	router.Use(Logging(middlewareConfig.Logging))
	router.Use(Validation(middlewareConfig.Validation))
	router.Use(Auth(middlewareConfig.Auth))

	// Add test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Reset timer
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}