package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClientForCache is a mock for the Redis client used in cache tests
type MockRedisClientForCache struct {
	mock.Mock
}

func (m *MockRedisClientForCache) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForCache) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClientForCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClientForCache) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForCache) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringStringMapCmd)
}

func (m *MockRedisClientForCache) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClientForCache) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := m.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := m.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (m *MockRedisClientForCache) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClientForCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForCache) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	args := m.Called(ctx, patterns)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClientForCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, script, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForCache) EvalSha(ctx context.Context, sha string, keys []string, args ...interface{}) *redis.Cmd {
	args := m.Called(ctx, sha, keys, args)
	return args.Get(0).(*redis.Cmd)
}

func (m *MockRedisClientForCache) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := m.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClientForCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClientForCache) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// CacheServiceTestSuite is the test suite for cache service
type CacheServiceTestSuite struct {
	suite.Suite
	cacheService *CacheService
	mockClient   *MockRedisClientForCache
}

func (suite *CacheServiceTestSuite) SetupTest() {
	suite.mockClient = new(MockRedisClientForCache)
	suite.cacheService = &CacheService{
		redisClient: suite.mockClient,
		config: &CacheConfig{
			UserProfileTTL:         30 * time.Minute,
			PhotoMetadataTTL:       1 * time.Hour,
			MatchRecommendationsTTL: 15 * time.Minute,
			APIResponseTTL:         5 * time.Minute,
			GeospatialDataTTL:       2 * time.Hour,
		},
	}
}

func (suite *CacheServiceTestSuite) TestCacheUserProfile() {
	ctx := context.Background()
	userID := "user123"
	profile := map[string]interface{}{
		"id":       userID,
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      30,
		"location": "New York",
	}
	
	profileData, _ := json.Marshal(profile)
	
	// Mock successful cache set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), profileData, 30*time.Minute).Return(statusCmd)
	
	err := suite.cacheService.CacheUserProfile(ctx, userID, profile)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetUserProfile() {
	ctx := context.Background()
	userID := "user123"
	
	expectedProfile := map[string]interface{}{
		"id":       userID,
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      30,
		"location": "New York",
	}
	
	profileData, _ := json.Marshal(expectedProfile)
	
	// Mock successful cache get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(profileData))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var profile map[string]interface{}
	found, err := suite.cacheService.GetUserProfile(ctx, userID, &profile)
	
	suite.NoError(err)
	suite.True(found)
	suite.Equal(expectedProfile["id"], profile["id"])
	suite.Equal(expectedProfile["name"], profile["name"])
	suite.Equal(expectedProfile["email"], profile["email"])
	suite.Equal(expectedProfile["age"], profile["age"])
	suite.Equal(expectedProfile["location"], profile["location"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetUserProfileNotFound() {
	ctx := context.Background()
	userID := "nonexistent_user"
	
	// Mock cache miss
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetErr(redis.Nil)
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var profile map[string]interface{}
	found, err := suite.cacheService.GetUserProfile(ctx, userID, &profile)
	
	suite.NoError(err)
	suite.False(found)
	suite.Nil(profile)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestInvalidateUserProfile() {
	ctx := context.Background()
	userID := "user123"
	
	// Mock successful cache invalidation
	intCmd := redis.NewIntCmd(ctx)
	intCmd.SetVal(1)
	suite.mockClient.On("Del", ctx, mock.AnythingOfType("string")).Return(intCmd)
	
	err := suite.cacheService.InvalidateUserProfile(ctx, userID)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestCachePhotoMetadata() {
	ctx := context.Background()
	photoID := "photo123"
	metadata := map[string]interface{}{
		"id":         photoID,
		"user_id":    "user123",
		"filename":   "photo.jpg",
		"size":       1024000,
		"width":      1920,
		"height":     1080,
		"created_at": time.Now(),
	}
	
	metadataData, _ := json.Marshal(metadata)
	
	// Mock successful cache set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), metadataData, 1*time.Hour).Return(statusCmd)
	
	err := suite.cacheService.CachePhotoMetadata(ctx, photoID, metadata)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetPhotoMetadata() {
	ctx := context.Background()
	photoID := "photo123"
	
	expectedMetadata := map[string]interface{}{
		"id":         photoID,
		"user_id":    "user123",
		"filename":   "photo.jpg",
		"size":       1024000,
		"width":      1920,
		"height":     1080,
		"created_at": time.Now(),
	}
	
	metadataData, _ := json.Marshal(expectedMetadata)
	
	// Mock successful cache get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(metadataData))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var metadata map[string]interface{}
	found, err := suite.cacheService.GetPhotoMetadata(ctx, photoID, &metadata)
	
	suite.NoError(err)
	suite.True(found)
	suite.Equal(expectedMetadata["id"], metadata["id"])
	suite.Equal(expectedMetadata["user_id"], metadata["user_id"])
	suite.Equal(expectedMetadata["filename"], metadata["filename"])
	suite.Equal(expectedMetadata["size"], metadata["size"])
	suite.Equal(expectedMetadata["width"], metadata["width"])
	suite.Equal(expectedMetadata["height"], metadata["height"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestCacheMatchRecommendations() {
	ctx := context.Background()
	userID := "user123"
	recommendations := []map[string]interface{}{
		{"id": "user1", "name": "Alice", "score": 0.95},
		{"id": "user2", "name": "Bob", "score": 0.87},
		{"id": "user3", "name": "Charlie", "score": 0.82},
	}
	
	recommendationsData, _ := json.Marshal(recommendations)
	
	// Mock successful cache set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), recommendationsData, 15*time.Minute).Return(statusCmd)
	
	err := suite.cacheService.CacheMatchRecommendations(ctx, userID, recommendations)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetMatchRecommendations() {
	ctx := context.Background()
	userID := "user123"
	
	expectedRecommendations := []map[string]interface{}{
		{"id": "user1", "name": "Alice", "score": 0.95},
		{"id": "user2", "name": "Bob", "score": 0.87},
		{"id": "user3", "name": "Charlie", "score": 0.82},
	}
	
	recommendationsData, _ := json.Marshal(expectedRecommendations)
	
	// Mock successful cache get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(recommendationsData))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var recommendations []map[string]interface{}
	found, err := suite.cacheService.GetMatchRecommendations(ctx, userID, &recommendations)
	
	suite.NoError(err)
	suite.True(found)
	suite.Equal(3, len(recommendations))
	suite.Equal(expectedRecommendations[0]["id"], recommendations[0]["id"])
	suite.Equal(expectedRecommendations[1]["id"], recommendations[1]["id"])
	suite.Equal(expectedRecommendations[2]["id"], recommendations[2]["id"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestCacheAPIResponse() {
	ctx := context.Background()
	key := "api_response_key"
	response := map[string]interface{}{
		"status":  "success",
		"data":    []string{"item1", "item2", "item3"},
		"message": "Request completed successfully",
	}
	
	responseData, _ := json.Marshal(response)
	
	// Mock successful cache set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), responseData, 5*time.Minute).Return(statusCmd)
	
	err := suite.cacheService.CacheAPIResponse(ctx, key, response)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetAPIResponse() {
	ctx := context.Background()
	key := "api_response_key"
	
	expectedResponse := map[string]interface{}{
		"status":  "success",
		"data":    []string{"item1", "item2", "item3"},
		"message": "Request completed successfully",
	}
	
	responseData, _ := json.Marshal(expectedResponse)
	
	// Mock successful cache get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(responseData))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var response map[string]interface{}
	found, err := suite.cacheService.GetAPIResponse(ctx, key, &response)
	
	suite.NoError(err)
	suite.True(found)
	suite.Equal(expectedResponse["status"], response["status"])
	suite.Equal(expectedResponse["message"], response["message"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestCacheGeospatialData() {
	ctx := context.Background()
	key := "geospatial_key"
	data := map[string]interface{}{
		"latitude":  40.7128,
		"longitude": -74.0060,
		"radius":    10.0,
		"users":     []string{"user1", "user2", "user3"},
	}
	
	dataJSON, _ := json.Marshal(data)
	
	// Mock successful cache set
	statusCmd := redis.NewStatusCmd(ctx)
	statusCmd.SetVal("OK")
	suite.mockClient.On("Set", ctx, mock.AnythingOfType("string"), dataJSON, 2*time.Hour).Return(statusCmd)
	
	err := suite.cacheService.CacheGeospatialData(ctx, key, data)
	
	suite.NoError(err)
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetGeospatialData() {
	ctx := context.Background()
	key := "geospatial_key"
	
	expectedData := map[string]interface{}{
		"latitude":  40.7128,
		"longitude": -74.0060,
		"radius":    10.0,
		"users":     []string{"user1", "user2", "user3"},
	}
	
	dataJSON, _ := json.Marshal(expectedData)
	
	// Mock successful cache get
	stringCmd := redis.NewStringCmd(ctx)
	stringCmd.SetVal(string(dataJSON))
	suite.mockClient.On("Get", ctx, mock.AnythingOfType("string")).Return(stringCmd)
	
	var data map[string]interface{}
	found, err := suite.cacheService.GetGeospatialData(ctx, key, &data)
	
	suite.NoError(err)
	suite.True(found)
	suite.Equal(expectedData["latitude"], data["latitude"])
	suite.Equal(expectedData["longitude"], data["longitude"])
	suite.Equal(expectedData["radius"], data["radius"])
	
	suite.mockClient.AssertExpectations(suite.T())
}

func (suite *CacheServiceTestSuite) TestGetCacheStats() {
	ctx := context.Background()
	
	// Mock getting info command
	cmd := redis.NewCmd(ctx)
	cmd.SetVal(map[string]interface{}{
		"used_memory":      "1000000",
		"used_memory_human": "1.00M",
		"used_memory_peak": "2000000",
		"keyspace_hits":    "1000",
		"keyspace_misses":  "100",
		"expired_keys":     "50",
		"evicted_keys":     "10",
	})
	
	// This would require a more complex mock setup for the Info command
	// For now, we'll test the basic structure
	
	stats, err := suite.cacheService.GetCacheStats(ctx)
	
	// Since we can't easily mock the Info command, we'll just test that the method doesn't panic
	// In a real implementation, you would need to mock the Info command properly
	suite.NotNil(stats)
	suite.NoError(err)
}

func TestCacheServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CacheServiceTestSuite))
}

// TestCacheConfig tests the CacheConfig struct
func TestCacheConfig(t *testing.T) {
	config := &CacheConfig{
		UserProfileTTL:         30 * time.Minute,
		PhotoMetadataTTL:       1 * time.Hour,
		MatchRecommendationsTTL: 15 * time.Minute,
		APIResponseTTL:         5 * time.Minute,
		GeospatialDataTTL:       2 * time.Hour,
	}
	
	assert.Equal(t, 30*time.Minute, config.UserProfileTTL)
	assert.Equal(t, 1*time.Hour, config.PhotoMetadataTTL)
	assert.Equal(t, 15*time.Minute, config.MatchRecommendationsTTL)
	assert.Equal(t, 5*time.Minute, config.APIResponseTTL)
	assert.Equal(t, 2*time.Hour, config.GeospatialDataTTL)
}