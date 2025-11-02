package services

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// GeoValidationService handles geospatial validation and operations
type GeoValidationService interface {
	ValidateCoordinates(lat, lng float64) error
	ValidateLocationUpdateFrequency(ctx context.Context, userID uuid.UUID) error
	ReverseGeocode(lat, lng float64) (city, country string, err error)
	CalculateDistance(lat1, lng1, lat2, lng2 float64) float64
	IsWithinRadius(centerLat, centerLng, pointLat, pointLng, radiusKm float64) bool
	GetNearbyUsers(ctx context.Context, userID uuid.UUID, radiusKm int, limit int) ([]*NearbyUser, error)
}

// RedisGeoValidationService implements GeoValidationService
type RedisGeoValidationService struct {
	userRepo    repositories.UserRepository
	cacheService ProfileCacheService
	geoProvider  ExternalGeoProvider
}

// NewRedisGeoValidationService creates a new RedisGeoValidationService instance
func NewRedisGeoValidationService(
	userRepo repositories.UserRepository,
	cacheService ProfileCacheService,
	geoProvider ExternalGeoProvider,
) *RedisGeoValidationService {
	return &RedisGeoValidationService{
		userRepo:    userRepo,
		cacheService: cacheService,
		geoProvider:  geoProvider,
	}
}

// ValidateCoordinates validates latitude and longitude
func (s *RedisGeoValidationService) ValidateCoordinates(lat, lng float64) error {
	// Validate latitude
	if lat < -90 || lat > 90 {
		return errors.ErrInvalidLatitude
	}

	// Validate longitude
	if lng < -180 || lng > 180 {
		return errors.ErrInvalidLongitude
	}

	// Check for invalid coordinates (0,0) which might be default values
	if lat == 0 && lng == 0 {
		return errors.ErrInvalidCoordinates
	}

	return nil
}

// ValidateLocationUpdateFrequency checks if user is updating location too frequently
func (s *RedisGeoValidationService) ValidateLocationUpdateFrequency(ctx context.Context, userID uuid.UUID) error {
	// Check cache for last location update
	cacheKey := "location_update:" + userID.String()
	lastUpdateStr, err := s.cacheService.Get(ctx, cacheKey)
	if err != nil {
		// If cache miss, allow update
		return nil
	}

	if lastUpdateStr != "" {
		// Parse last update time
		lastUpdate, err := time.Parse(time.RFC3339, lastUpdateStr)
		if err != nil {
			// If we can't parse, allow update
			return nil
		}

		// Check if enough time has passed (e.g., 5 minutes)
		if time.Since(lastUpdate) < 5*time.Minute {
			return errors.ErrLocationUpdateTooFrequent
		}
	}

	return nil
}

// ReverseGeocode converts coordinates to city and country
func (s *RedisGeoValidationService) ReverseGeocode(lat, lng float64) (city, country string, err error) {
	// Try to get from cache first
	cacheKey := "reverse_geocode:" + formatCoordinates(lat, lng)
	cached, err := s.cacheService.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		// Parse cached result
		var result GeocodeResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result.City, result.Country, nil
		}
	}

	// Call external geocoding service
	city, country, err = s.geoProvider.ReverseGeocode(lat, lng)
	if err != nil {
		return "", "", err
	}

	// Cache the result
	result := GeocodeResult{
		City:    city,
		Country: country,
	}
	resultJSON, _ := json.Marshal(result)
	s.cacheService.Set(ctx, cacheKey, string(resultJSON), 24*time.Hour)

	return city, country, nil
}

// CalculateDistance calculates the distance between two points in kilometers
func (s *RedisGeoValidationService) CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371

	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Calculate the change in coordinates
	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad

	// Apply the Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Calculate the distance
	distance := earthRadiusKm * c

	return distance
}

// IsWithinRadius checks if a point is within a given radius from a center point
func (s *RedisGeoValidationService) IsWithinRadius(centerLat, centerLng, pointLat, pointLng, radiusKm float64) bool {
	distance := s.CalculateDistance(centerLat, centerLng, pointLat, pointLng)
	return distance <= radiusKm
}

// GetNearbyUsers gets users within a specified radius
func (s *RedisGeoValidationService) GetNearbyUsers(ctx context.Context, userID uuid.UUID, radiusKm int, limit int) ([]*NearbyUser, error) {
	// Get current user's location
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.HasLocation() {
		return nil, errors.ErrUserLocationNotSet
	}

	// Get nearby users from database
	users, err := s.userRepo.GetByLocation(ctx, *user.LocationLat, *user.LocationLng, radiusKm, limit, 0)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get nearby users")
	}

	// Convert to response format
	nearbyUsers := make([]*NearbyUser, 0, len(users))
	for _, nearbyUser := range users {
		// Skip self
		if nearbyUser.ID == userID {
			continue
		}

		// Calculate distance
		distance := s.CalculateDistance(*user.LocationLat, *user.LocationLng, *nearbyUser.LocationLat, *nearbyUser.LocationLng)

		// Get user's primary photo
		photos, err := s.photoRepo.GetUserPhotos(ctx, nearbyUser.ID, true)
		var primaryPhotoURL string
		if err == nil && len(photos) > 0 {
			for _, photo := range photos {
				if photo.IsPrimary {
					primaryPhotoURL = photo.FileURL
					break
				}
			}
		}

		nearbyUser := &NearbyUser{
			ID:          nearbyUser.ID,
			FirstName:    nearbyUser.FirstName,
			Age:         nearbyUser.GetAge(),
			Distance:     distance,
			PrimaryPhoto: primaryPhotoURL,
			LastActive:   formatLastActive(nearbyUser.LastActive),
		}

		nearbyUsers = append(nearbyUsers, nearbyUser)
	}

	return nearbyUsers, nil
}

// UpdateLocationCache updates the location update cache
func (s *RedisGeoValidationService) UpdateLocationCache(ctx context.Context, userID uuid.UUID) error {
	cacheKey := "location_update:" + userID.String()
	now := time.Now().Format(time.RFC3339)
	return s.cacheService.Set(ctx, cacheKey, now, time.Hour)
}

// GeocodeResult represents geocoding result
type GeocodeResult struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// NearbyUser represents a nearby user
type NearbyUser struct {
	ID           uuid.UUID `json:"id"`
	FirstName     string     `json:"first_name"`
	Age          int        `json:"age"`
	Distance     float64    `json:"distance"`
	PrimaryPhoto  string     `json:"primary_photo"`
	LastActive   string     `json:"last_active"`
}

// formatCoordinates formats coordinates for cache key
func formatCoordinates(lat, lng float64) string {
	return fmt.Sprintf("%.6f,%.6f", lat, lng)
}

// formatLastActive formats last active time
func formatLastActive(lastActive *time.Time) string {
	if lastActive == nil {
		return ""
	}
	return lastActive.Format("2006-01-02T15:04:05Z07:00")
}

// ExternalGeoProvider defines interface for external geocoding services
type ExternalGeoProvider interface {
	ReverseGeocode(lat, lng float64) (city, country string, err error)
	Geocode(address string) (lat, lng float64, err error)
}

// repositories defines the repository interfaces needed
var (
	repositories struct {
		userRepo    repositories.UserRepository
		photoRepo   repositories.PhotoRepository
	}
)

// ProfileCacheService defines the cache service interface
var (
	ProfileCacheService interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key string, value string, ttl time.Duration) error
	}
)