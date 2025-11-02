package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"internal/application/dto"
	"internal/application/services"
	"internal/application/usecases/matching"
	"internal/domain/entities"
	"internal/domain/repositories"
	"internal/interfaces/http/handlers"
	"internal/interfaces/http/routes"
)

// Test setup
func setupDiscoveryTestRouter() (*gin.Engine, *MockUserRepository, *MockMatchRepository, *MockSwipeRepository, *services.CacheService, *services.RateLimiter) {
	gin.SetMode(gin.TestMode)
	
	// Create mock repositories
	userRepo := &MockUserRepository{}
	matchRepo := &MockMatchRepository{}
	swipeRepo := &MockSwipeRepository{}
	
	// Create services
	cacheService := &services.CacheService{}
	rateLimiter := &services.RateLimiter{}
	
	// Create use cases
	discoverUsersUC := matching.NewDiscoverUsersUseCase(userRepo, cacheService, rateLimiter)
	likeUserUC := matching.NewLikeUserUseCase(userRepo, matchRepo, swipeRepo, cacheService)
	dislikeUserUC := matching.NewDislikeUserUseCase(userRepo, swipeRepo, cacheService)
	superlikeUserUC := matching.NewSuperLikeUserUseCase(userRepo, matchRepo, swipeRepo, cacheService)
	getMatchesUC := matching.NewGetMatchesUseCase(userRepo, matchRepo, cacheService)
	getDiscoveryStatsUC := matching.NewGetDiscoveryStatsUseCase(userRepo, matchRepo, swipeRepo, cacheService)
	
	// Create handler
	discoveryHandler := handlers.NewDiscoveryHandler(
		discoverUsersUC,
		likeUserUC,
		dislikeUserUC,
		superlikeUserUC,
		getMatchesUC,
		getDiscoveryStatsUC,
	)
	
	// Create router
	router := gin.New()
	routes.RegisterDiscoveryRoutes(router, discoveryHandler, rateLimiter)
	
	return router, userRepo, matchRepo, swipeRepo, cacheService, rateLimiter
}

// Test DiscoverUsers endpoint
func TestDiscoverUsers(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create discoverable users
	discoverableUsers := []*entities.User{
		{
			ID:             uuid.New(),
			Email:          "user1@example.com",
			Name:           "User One",
			Age:            26,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelBasic,
			IsPremium:      false,
			IsActive:       true,
			ProfileCompletion: 85,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Email:          "user2@example.com",
			Name:           "User Two",
			Age:            24,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelVerified,
			IsPremium:      true,
			IsActive:       true,
			ProfileCompletion: 90,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)
	userRepo.On("GetByLocation", mock.Anything, 40.7128, -74.0060, 50, 20, 0).Return(discoverableUsers, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/discover?limit=20&offset=0", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.DiscoverUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Users))
	assert.Equal(t, "User One", response.Users[0].Name)
	assert.Equal(t, "User Two", response.Users[1].Name)
	
	userRepo.AssertExpectations(t)
}

// Test LikeUser endpoint
func TestLikeUser(t *testing.T) {
	router, userRepo, matchRepo, swipeRepo, _, _ := setupDiscoveryTestRouter()
	
	// Create test users
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	targetUser := &entities.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Name:           "Target User",
		Age:            26,
		Gender:         "female",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 85,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	userRepo.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
	swipeRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Swipe")).Return(nil)
	swipeRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	matchRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/like/"+targetUserID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.LikeUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, targetUserID, response.TargetUserID)
	assert.False(t, response.IsMatch) // No match since target user hasn't liked current user
	
	userRepo.AssertExpectations(t)
	swipeRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
}

// Test DislikeUser endpoint
func TestDislikeUser(t *testing.T) {
	router, userRepo, _, swipeRepo, _, _ := setupDiscoveryTestRouter()
	
	// Create test users
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	targetUser := &entities.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Name:           "Target User",
		Age:            26,
		Gender:         "female",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 85,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	userRepo.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
	swipeRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Swipe")).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/dislike/"+targetUserID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.DislikeUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, targetUserID, response.TargetUserID)
	
	userRepo.AssertExpectations(t)
	swipeRepo.AssertExpectations(t)
}

// Test SuperLikeUser endpoint
func TestSuperLikeUser(t *testing.T) {
	router, userRepo, matchRepo, swipeRepo, _, _ := setupDiscoveryTestRouter()
	
	// Create test users
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      true, // Premium user for super like
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	targetUser := &entities.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Name:           "Target User",
		Age:            26,
		Gender:         "female",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 85,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	userRepo.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
	swipeRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Swipe")).Return(nil)
	swipeRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	matchRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/superlike/"+targetUserID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.SuperLikeUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, targetUserID, response.TargetUserID)
	assert.False(t, response.IsMatch) // No match since target user hasn't liked current user
	
	userRepo.AssertExpectations(t)
	swipeRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
}

// Test GetMatches endpoint
func TestGetMatches(t *testing.T) {
	router, userRepo, matchRepo, _, _, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create matches
	matchedUserID1 := uuid.New()
	matchedUserID2 := uuid.New()
	
	matches := []*entities.Match{
		{
			ID:         uuid.New(),
			UserID:     userID,
			MatchedUserID: matchedUserID1,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			MatchedUserID: matchedUserID2,
			CreatedAt:  time.Now(),
		},
	}
	
	matchedUsers := []*entities.User{
		{
			ID:             matchedUserID1,
			Email:          "match1@example.com",
			Name:           "Match One",
			Age:            26,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelBasic,
			IsPremium:      false,
			IsActive:       true,
			ProfileCompletion: 85,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             matchedUserID2,
			Email:          "match2@example.com",
			Name:           "Match Two",
			Age:            24,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelVerified,
			IsPremium:      true,
			IsActive:       true,
			ProfileCompletion: 90,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)
	matchRepo.On("GetByUserID", mock.Anything, userID, 20, 0).Return(matches, nil)
	userRepo.On("GetUsersByIDs", mock.Anything, []uuid.UUID{matchedUserID1, matchedUserID2}).Return(matchedUsers, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/matches?limit=20&offset=0", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.GetMatchesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Matches))
	assert.Equal(t, "Match One", response.Matches[0].Name)
	assert.Equal(t, "Match Two", response.Matches[1].Name)
	
	userRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
}

// Test GetDiscoveryStats endpoint
func TestGetDiscoveryStats(t *testing.T) {
	router, userRepo, matchRepo, swipeRepo, _, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)
	matchRepo.On("CountByUserID", mock.Anything, userID).Return(5, nil)
	swipeRepo.On("CountByUserID", mock.Anything, userID).Return(20, nil)
	swipeRepo.On("CountByUserIDAndType", mock.Anything, userID, "like").Return(15, nil)
	swipeRepo.On("CountByUserIDAndType", mock.Anything, userID, "dislike").Return(4, nil)
	swipeRepo.On("CountByUserIDAndType", mock.Anything, userID, "superlike").Return(1, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/discover/stats", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.GetDiscoveryStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 5, response.TotalMatches)
	assert.Equal(t, 20, response.TotalSwipes)
	assert.Equal(t, 15, response.TotalLikes)
	assert.Equal(t, 4, response.TotalDislikes)
	assert.Equal(t, 1, response.TotalSuperLikes)
	
	userRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
	swipeRepo.AssertExpectations(t)
}

// Test performance with large dataset
func TestDiscoverUsersPerformance(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create large dataset of discoverable users
	discoverableUsers := make([]*entities.User, 1000)
	for i := 0; i < 1000; i++ {
		discoverableUsers[i] = &entities.User{
			ID:             uuid.New(),
			Email:          "user@example.com",
			Name:           "User",
			Age:            25,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelBasic,
			IsPremium:      false,
			IsActive:       true,
			ProfileCompletion: 80,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)
	userRepo.On("GetByLocation", mock.Anything, 40.7128, -74.0060, 50, 20, 0).Return(discoverableUsers[:20], nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/discover?limit=20&offset=0", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// Measure performance
	start := time.Now()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	duration := time.Since(start)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, duration, 100*time.Millisecond, "Discovery should complete within 100ms")
	
	var response dto.DiscoverUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(response.Users))
	
	userRepo.AssertExpectations(t)
}

// Test caching effectiveness
func TestDiscoverUsersCaching(t *testing.T) {
	router, userRepo, _, _, cacheService, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create discoverable users
	discoverableUsers := []*entities.User{
		{
			ID:             uuid.New(),
			Email:          "user1@example.com",
			Name:           "User One",
			Age:            26,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelBasic,
			IsPremium:      false,
			IsActive:       true,
			ProfileCompletion: 85,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
	
	// Setup mock expectations - only called once due to caching
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil).Twice()
	userRepo.On("GetByLocation", mock.Anything, 40.7128, -74.0060, 50, 20, 0).Return(discoverableUsers, nil).Once()
	
	// Create request
	req, _ := http.NewRequest("GET", "/discover?limit=20&offset=0", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// First request - should hit database
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Second request - should hit cache
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusOK, w2.Code)
	
	// Verify responses are the same
	assert.Equal(t, w1.Body.String(), w2.Body.String())
	
	userRepo.AssertExpectations(t)
}

// Test error cases
func TestDiscoverUsersErrorCases(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Test user not found
	userID := uuid.New()
	userRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("user not found"))
	
	req, _ := http.NewRequest("GET", "/discover", nil)
	req.Header.Set("Content-Type", "application/json")
	
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	userRepo.AssertExpectations(t)
}

func TestLikeUserErrorCases(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Test invalid target user ID
	currentUserID := uuid.New()
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	
	// Test with invalid UUID
	req, _ := http.NewRequest("POST", "/like/invalid-uuid", nil)
	req.Header.Set("Content-Type", "application/json")
	
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	userRepo.AssertExpectations(t)
}

func TestSuperLikeUserPremiumRequired(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Create non-premium user
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false, // Non-premium user
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	targetUser := &entities.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Name:           "Target User",
		Age:            26,
		Gender:         "female",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 85,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Setup mock expectations
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	userRepo.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/superlike/"+targetUserID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response - should return error for non-premium user
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	userRepo.AssertExpectations(t)
}

func TestDiscoverUsersWithFilters(t *testing.T) {
	router, userRepo, _, _, _, _ := setupDiscoveryTestRouter()
	
	// Create test user
	userID := uuid.New()
	testUser := &entities.User{
		ID:             userID,
		Email:          "test@example.com",
		Name:           "Test User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create discoverable users
	discoverableUsers := []*entities.User{
		{
			ID:             uuid.New(),
			Email:          "user1@example.com",
			Name:           "User One",
			Age:            26,
			Gender:         "female",
			Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
			VerificationLevel: entities.VerificationLevelBasic,
			IsPremium:      false,
			IsActive:       true,
			ProfileCompletion: 85,
			LastActive:     time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}
	
	// Setup mock expectations with filters
	userRepo.On("GetByID", mock.Anything, userID).Return(testUser, nil)
	userRepo.On("GetByLocation", mock.Anything, 40.7128, -74.0060, 50, 20, 0).Return(discoverableUsers, nil)
	
	// Create request with filters
	req, _ := http.NewRequest("GET", "/discover?min_age=20&max_age=30&gender=female&radius=50&limit=20&offset=0", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", userID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.DiscoverUsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(response.Users))
	assert.Equal(t, "User One", response.Users[0].Name)
	
	userRepo.AssertExpectations(t)
}

func TestMatchDetection(t *testing.T) {
	router, userRepo, matchRepo, swipeRepo, _, _ := setupDiscoveryTestRouter()
	
	// Create test users
	currentUserID := uuid.New()
	targetUserID := uuid.New()
	
	currentUser := &entities.User{
		ID:             currentUserID,
		Email:          "current@example.com",
		Name:           "Current User",
		Age:            25,
		Gender:         "male",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 80,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	targetUser := &entities.User{
		ID:             targetUserID,
		Email:          "target@example.com",
		Name:           "Target User",
		Age:            26,
		Gender:         "female",
		Location:       entities.Location{Latitude: 40.7128, Longitude: -74.0060},
		VerificationLevel: entities.VerificationLevelBasic,
		IsPremium:      false,
		IsActive:       true,
		ProfileCompletion: 85,
		LastActive:     time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create existing swipe from target user to current user
	existingSwipe := &entities.Swipe{
		ID:         uuid.New(),
		UserID:     targetUserID,
		TargetUserID: currentUserID,
		Type:       "like",
		CreatedAt:  time.Now(),
	}
	
	// Setup mock expectations for match detection
	userRepo.On("GetByID", mock.Anything, currentUserID).Return(currentUser, nil)
	userRepo.On("GetByID", mock.Anything, targetUserID).Return(targetUser, nil)
	swipeRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Swipe")).Return(nil)
	swipeRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	swipeRepo.On("GetByUsers", mock.Anything, targetUserID, currentUserID).Return(existingSwipe, nil)
	matchRepo.On("GetByUsers", mock.Anything, currentUserID, targetUserID).Return(nil, nil)
	matchRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Match")).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/like/"+targetUserID.String(), nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Add user context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", currentUserID)
	req = req.WithContext(c.Request.Context())
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response - should indicate a match
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.LikeUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, targetUserID, response.TargetUserID)
	assert.True(t, response.IsMatch) // Should be a match
	
	userRepo.AssertExpectations(t)
	swipeRepo.AssertExpectations(t)
	matchRepo.AssertExpectations(t)
}

// Mock implementations
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByLocation(ctx context.Context, lat, lng float64, radiusKm int, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, lat, lng, radiusKm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]*entities.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0), args.Error(1)
}

// Add other required methods with minimal implementations
func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*entities.User, error) { return nil, nil }
func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error { return nil }
func (m *MockUserRepository) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error { return nil }
func (m *MockUserRepository) UpdatePhone(ctx context.Context, userID uuid.UUID, phone string) error { return nil }
func (m *MockUserRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64, city string) error { return nil }
func (m *MockUserRepository) UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences map[string]interface{}) error { return nil }
func (m *MockUserRepository) UpdateVerificationLevel(ctx context.Context, userID uuid.UUID, level entities.VerificationLevel) error { return nil }
func (m *MockUserRepository) UpdatePremiumStatus(ctx context.Context, userID uuid.UUID, isPremium bool) error { return nil }
func (m *MockUserRepository) UpdateActiveStatus(ctx context.Context, userID uuid.UUID, isActive bool) error { return nil }
func (m *MockUserRepository) UpdateBannedStatus(ctx context.Context, userID uuid.UUID, isBanned bool) error { return nil }
func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByAgeRange(ctx context.Context, minAge, maxAge int, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByGender(ctx context.Context, gender string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByVerificationLevel(ctx context.Context, level entities.VerificationLevel, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetPremiumUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetBannedUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCity(ctx context.Context, city string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCountry(ctx context.Context, country string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByInterests(ctx context.Context, interests []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByRelationshipStatus(ctx context.Context, status string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByLookingFor(ctx context.Context, lookingFor []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDistance(ctx context.Context, lat, lng float64, maxDistance int, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByAge(ctx context.Context, age int, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHeight(ctx context.Context, height int, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByWeight(ctx context.Context, weight int, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByBMI(ctx context.Context, bmi float64, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByBodyType(ctx context.Context, bodyType string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHairColor(ctx context.Context, hairColor string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByEyeColor(ctx context.Context, eyeColor string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySkinTone(ctx context.Context, skinTone string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByFacialHair(ctx context.Context, facialHair string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByTattoos(ctx context.Context, tattoos string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByPiercings(ctx context.Context, piercings string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByScars(ctx context.Context, scars string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByBirthmarks(ctx context.Context, birthmarks string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMoles(ctx context.Context, moles string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByFreckles(ctx context.Context, freckles string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDimples(ctx context.Context, dimples string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCleftChin(ctx context.Context, cleftChin string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByGlasses(ctx context.Context, glasses string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByContacts(ctx context.Context, contacts string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHearingAid(ctx context.Context, hearingAid string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByWheelchair(ctx context.Context, wheelchair string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCane(ctx context.Context, cane string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCrutches(ctx context.Context, crutches string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByWalker(ctx context.Context, walker string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByScooter(ctx context.Context, scooter string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByServiceAnimal(ctx context.Context, serviceAnimal string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByEmotionalSupportAnimal(ctx context.Context, emotionalSupportAnimal string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByTherapyAnimal(ctx context.Context, therapyAnimal string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByPet(ctx context.Context, pet string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByAllergies(ctx context.Context, allergies []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDietaryRestrictions(ctx context.Context, dietaryRestrictions []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMedicalConditions(ctx context.Context, medicalConditions []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMedications(ctx context.Context, medications []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySurgeries(ctx context.Context, surgeries []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByInjuries(ctx context.Context, injuries []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDisabilities(ctx context.Context, disabilities []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMentalHealth(ctx context.Context, mentalHealth []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByTherapy(ctx context.Context, therapy []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMedication(ctx context.Context, medication []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySupportGroup(ctx context.Context, supportGroup []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCounseling(ctx context.Context, counseling []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCoaching(ctx context.Context, coaching []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMentoring(ctx context.Context, mentoring []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySponsorship(ctx context.Context, sponsorship []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByVolunteering(ctx context.Context, volunteering []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDonations(ctx context.Context, donations []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByFundraising(ctx context.Context, fundraising []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByAdvocacy(ctx context.Context, advocacy []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByActivism(ctx context.Context, activism []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByPolitics(ctx context.Context, politics []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByReligion(ctx context.Context, religion []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySpirituality(ctx context.Context, spirituality []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCultural(ctx context.Context, cultural []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByEthnic(ctx context.Context, ethnic []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByLGBTQ(ctx context.Context, lgbtq []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByGenderIdentity(ctx context.Context, genderIdentity []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySexualOrientation(ctx context.Context, sexualOrientation []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByRelationship(ctx context.Context, relationship []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDating(ctx context.Context, dating []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMarriage(ctx context.Context, marriage []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDivorce(ctx context.Context, divorce []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByParenting(ctx context.Context, parenting []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByChildcare(ctx context.Context, childcare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByElderCare(ctx context.Context, elderCare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByPetCare(ctx context.Context, petCare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHomeCare(ctx context.Context, homeCare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHealthCare(ctx context.Context, healthCare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByMentalHealthCare(ctx context.Context, mentalHealthCare []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySubstanceAbuse(ctx context.Context, substanceAbuse []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDomesticViolence(ctx context.Context, domesticViolence []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByHomeless(ctx context.Context, homeless []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDisaster(ctx context.Context, disaster []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByImmigration(ctx context.Context, immigration []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCitizenship(ctx context.Context, citizenship []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByLanguage(ctx context.Context, language []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByDisability(ctx context.Context, disability []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByVeteran(ctx context.Context, veteran []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySenior(ctx context.Context, senior []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByYouth(ctx context.Context, youth []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByFamily(ctx context.Context, family []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCommunity(ctx context.Context, community []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByReligious(ctx context.Context, religious []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersBySpiritual(ctx context.Context, spiritual []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByCultural(ctx context.Context, cultural []string, limit, offset int) ([]*entities.User, error) { return nil, nil }
func (m *MockUserRepository) GetUsersByEthnic(ctx context.Context, ethnic []string, limit, offset int) ([]*entities.User, error) { return nil, nil }

// MockMatchRepository implementation
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) Create(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) GetByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Match, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Match), args.Error(1)
}

func (m *MockMatchRepository) Update(ctx context.Context, match *entities.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// MockSwipeRepository implementation
type MockSwipeRepository struct {
	mock.Mock
}

func (m *MockSwipeRepository) Create(ctx context.Context, swipe *entities.Swipe) error {
	args := m.Called(ctx, swipe)
	return args.Error(0)
}

func (m *MockSwipeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Swipe, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Swipe), args.Error(1)
}

func (m *MockSwipeRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Swipe), args.Error(1)
}

func (m *MockSwipeRepository) GetByUsers(ctx context.Context, userID1, userID2 uuid.UUID) (*entities.Swipe, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Swipe), args.Error(1)
}

func (m *MockSwipeRepository) Update(ctx context.Context, swipe *entities.Swipe) error {
	args := m.Called(ctx, swipe)
	return args.Error(0)
}

func (m *MockSwipeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSwipeRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockSwipeRepository) CountByUserIDAndType(ctx context.Context, userID uuid.UUID, swipeType string) (int, error) {
	args := m.Called(ctx, userID, swipeType)
	return args.Int(0), args.Error(1)
}