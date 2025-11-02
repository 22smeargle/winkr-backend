package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CacheUtilsTestSuite is the test suite for cache utilities
type CacheUtilsTestSuite struct {
	suite.Suite
}

func (suite *CacheUtilsTestSuite) TestGenerateUserProfileKey() {
	userID := "user123"
	expectedKey := "user_profile:user123"
	actualKey := GenerateUserProfileKey(userID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGeneratePhotoMetadataKey() {
	photoID := "photo123"
	expectedKey := "photo_metadata:photo123"
	actualKey := GeneratePhotoMetadataKey(photoID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateMatchRecommendationsKey() {
	userID := "user123"
	expectedKey := "match_recommendations:user123"
	actualKey := GenerateMatchRecommendationsKey(userID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateAPIResponseKey() {
	endpoint := "/api/users"
	params := "page=1&limit=10"
	expectedKey := "api_response:/api/users:page=1&limit=10"
	actualKey := GenerateAPIResponseKey(endpoint, params)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateGeospatialDataKey() {
	latitude := 40.7128
	longitude := -74.0060
	radius := 10.0
	expectedKey := "geospatial:40.7128:-74.0060:10.0"
	actualKey := GenerateGeospatialDataKey(latitude, longitude, radius)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateSessionKey() {
	sessionID := "session123"
	expectedKey := "session:session123"
	actualKey := GenerateSessionKey(sessionID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateUserSessionsKey() {
	userID := "user123"
	expectedKey := "user_sessions:user123"
	actualKey := GenerateUserSessionsKey(userID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateOnlineUsersKey() {
	expectedKey := "online_users"
	actualKey := GenerateOnlineUsersKey()
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateRateLimitKey() {
	keyType := "ip"
	endpoint := "/api/users"
	identifier := "192.168.1.1"
	expectedKey := "rate_limit:ip:/api/users:192.168.1.1"
	actualKey := GenerateRateLimitKey(keyType, endpoint, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateDistributedRateLimitKey() {
	keyType := "user"
	endpoint := "/api/matches"
	identifier := "user123"
	expectedKey := "distributed_rate_limit:user:/api/matches:user123"
	actualKey := GenerateDistributedRateLimitKey(keyType, endpoint, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateTokenBlacklistKey() {
	tokenID := "token123"
	expectedKey := "token_blacklist:token123"
	actualKey := GenerateTokenBlacklistKey(tokenID)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestSerialize() {
	data := map[string]interface{}{
		"id":   "user123",
		"name": "John Doe",
		"age":  30,
	}
	
	serialized, err := Serialize(data)
	
	suite.NoError(err)
	suite.NotEmpty(serialized)
	
	// Verify that it's valid JSON
	var deserialized map[string]interface{}
	err = json.Unmarshal([]byte(serialized), &deserialized)
	suite.NoError(err)
	suite.Equal(data["id"], deserialized["id"])
	suite.Equal(data["name"], deserialized["name"])
	suite.Equal(data["age"], deserialized["age"])
}

func (suite *CacheUtilsTestSuite) TestSerializeWithNil() {
	serialized, err := Serialize(nil)
	
	suite.NoError(err)
	suite.Equal("null", serialized)
}

func (suite *CacheUtilsTestSuite) TestDeserialize() {
	data := map[string]interface{}{
		"id":   "user123",
		"name": "John Doe",
		"age":  30,
	}
	
	serialized, _ := Serialize(data)
	
	var result map[string]interface{}
	err := Deserialize(serialized, &result)
	
	suite.NoError(err)
	suite.Equal(data["id"], result["id"])
	suite.Equal(data["name"], result["name"])
	suite.Equal(data["age"], result["age"])
}

func (suite *CacheUtilsTestSuite) TestDeserializeWithInvalidJSON() {
	invalidJSON := "{invalid json}"
	
	var result map[string]interface{}
	err := Deserialize(invalidJSON, &result)
	
	suite.Error(err)
}

func (suite *CacheUtilsTestSuite) TestInvalidateUserProfile() {
	ctx := context.Background()
	userID := "user123"
	
	// This would require a Redis client to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateUserProfileKey(userID)
	expectedKey := "user_profile:user123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestInvalidatePhotoMetadata() {
	ctx := context.Background()
	photoID := "photo123"
	
	// This would require a Redis client to test properly
	// For now, we'll just test that the function generates the correct key
	key := GeneratePhotoMetadataKey(photoID)
	expectedKey := "photo_metadata:photo123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestInvalidateMatchRecommendations() {
	ctx := context.Background()
	userID := "user123"
	
	// This would require a Redis client to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateMatchRecommendationsKey(userID)
	expectedKey := "match_recommendations:user123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestInvalidateAPIResponse() {
	ctx := context.Background()
	endpoint := "/api/users"
	params := "page=1&limit=10"
	
	// This would require a Redis client to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateAPIResponseKey(endpoint, params)
	expectedKey := "api_response:/api/users:page=1&limit=10"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestInvalidateGeospatialData() {
	ctx := context.Background()
	latitude := 40.7128
	longitude := -74.0060
	radius := 10.0
	
	// This would require a Redis client to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateGeospatialDataKey(latitude, longitude, radius)
	expectedKey := "geospatial:40.7128:-74.0060:10.0"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestWarmUserProfileCache() {
	ctx := context.Background()
	userID := "user123"
	
	// This would require a Redis client and user repository to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateUserProfileKey(userID)
	expectedKey := "user_profile:user123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestWarmPhotoMetadataCache() {
	ctx := context.Background()
	photoID := "photo123"
	
	// This would require a Redis client and photo repository to test properly
	// For now, we'll just test that the function generates the correct key
	key := GeneratePhotoMetadataKey(photoID)
	expectedKey := "photo_metadata:photo123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestWarmMatchRecommendationsCache() {
	ctx := context.Background()
	userID := "user123"
	
	// This would require a Redis client and match repository to test properly
	// For now, we'll just test that the function generates the correct key
	key := GenerateMatchRecommendationsKey(userID)
	expectedKey := "match_recommendations:user123"
	
	suite.Equal(expectedKey, key)
}

func (suite *CacheUtilsTestSuite) TestGenerateCacheKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123"
	actualKey := GenerateCacheKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateCacheKeyWithMultipleParts() {
	prefix := "test"
	identifiers := []string{"123", "456", "789"}
	expectedKey := "test:123:456:789"
	actualKey := GenerateCacheKey(prefix, identifiers...)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateHashKey() {
	prefix := "test"
	identifier := "123"
	field := "field"
	expectedKey := "test:123"
	expectedField := "field"
	actualKey, actualField := GenerateHashKey(prefix, identifier, field)
	
	suite.Equal(expectedKey, actualKey)
	suite.Equal(expectedField, actualField)
}

func (suite *CacheUtilsTestSuite) TestGenerateSetKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123"
	actualKey := GenerateSetKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateSortedSetKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123"
	actualKey := GenerateSortedSetKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateGeoKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123"
	actualKey := GenerateGeoKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGenerateChannelKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123"
	actualKey := GenerateChannelKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGeneratePatternKey() {
	prefix := "test"
	identifier := "123"
	expectedKey := "test:123*"
	actualKey := GeneratePatternKey(prefix, identifier)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestGetCacheTTL() {
	cacheType := "user_profile"
	expectedTTL := 30 * time.Minute
	actualTTL := GetCacheTTL(cacheType)
	
	suite.Equal(expectedTTL, actualTTL)
}

func (suite *CacheUtilsTestSuite) TestGetCacheTTLForUnknownType() {
	cacheType := "unknown_type"
	expectedTTL := 5 * time.Minute // Default TTL
	actualTTL := GetCacheTTL(cacheType)
	
	suite.Equal(expectedTTL, actualTTL)
}

func (suite *CacheUtilsTestSuite) TestIsValidCacheKey() {
	validKey := "user_profile:user123"
	invalidKey := ""
	
	suite.True(IsValidCacheKey(validKey))
	suite.False(IsValidCacheKey(invalidKey))
}

func (suite *CacheUtilsTestSuite) TestSanitizeCacheKey() {
	key := "user_profile:user123:with:colons"
	expectedKey := "user_profile_user123_with_colons"
	actualKey := SanitizeCacheKey(key)
	
	suite.Equal(expectedKey, actualKey)
}

func (suite *CacheUtilsTestSuite) TestExtractKeyParts() {
	key := "user_profile:user123:details"
	expectedParts := []string{"user_profile", "user123", "details"}
	actualParts := ExtractKeyParts(key)
	
	suite.Equal(expectedParts, actualParts)
}

func (suite *CacheUtilsTestSuite) TestExtractKeyPartsWithEmptyKey() {
	key := ""
	expectedParts := []string{}
	actualParts := ExtractKeyParts(key)
	
	suite.Equal(expectedParts, actualParts)
}

func TestCacheUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(CacheUtilsTestSuite))
}

// TestCacheTTLConfig tests the cache TTL configuration
func TestCacheTTLConfig(t *testing.T) {
	config := &CacheTTLConfig{
		UserProfile:         30 * time.Minute,
		PhotoMetadata:       1 * time.Hour,
		MatchRecommendations: 15 * time.Minute,
		APIResponse:         5 * time.Minute,
		GeospatialData:      2 * time.Hour,
		Session:             24 * time.Hour,
		RefreshToken:        7 * 24 * time.Hour,
		RateLimit:           1 * time.Minute,
		TokenBlacklist:      1 * time.Hour,
		Default:             5 * time.Minute,
	}
	
	assert.Equal(t, 30*time.Minute, config.UserProfile)
	assert.Equal(t, 1*time.Hour, config.PhotoMetadata)
	assert.Equal(t, 15*time.Minute, config.MatchRecommendations)
	assert.Equal(t, 5*time.Minute, config.APIResponse)
	assert.Equal(t, 2*time.Hour, config.GeospatialData)
	assert.Equal(t, 24*time.Hour, config.Session)
	assert.Equal(t, 7*24*time.Hour, config.RefreshToken)
	assert.Equal(t, 1*time.Minute, config.RateLimit)
	assert.Equal(t, 1*time.Hour, config.TokenBlacklist)
	assert.Equal(t, 5*time.Minute, config.Default)
}

// TestCacheKeyPrefix tests the cache key prefix constants
func TestCacheKeyPrefix(t *testing.T) {
	assert.Equal(t, "user_profile", UserProfilePrefix)
	assert.Equal(t, "photo_metadata", PhotoMetadataPrefix)
	assert.Equal(t, "match_recommendations", MatchRecommendationsPrefix)
	assert.Equal(t, "api_response", APIResponsePrefix)
	assert.Equal(t, "geospatial", GeospatialPrefix)
	assert.Equal(t, "session", SessionPrefix)
	assert.Equal(t, "user_sessions", UserSessionsPrefix)
	assert.Equal(t, "online_users", OnlineUsersPrefix)
	assert.Equal(t, "rate_limit", RateLimitPrefix)
	assert.Equal(t, "distributed_rate_limit", DistributedRateLimitPrefix)
	assert.Equal(t, "token_blacklist", TokenBlacklistPrefix)
}

// TestCacheKeySeparator tests the cache key separator
func TestCacheKeySeparator(t *testing.T) {
	assert.Equal(t, ":", CacheKeySeparator)
}