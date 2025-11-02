package profile

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// UpdateLocationUseCase handles updating user location
type UpdateLocationUseCase struct {
	userRepo     repositories.UserRepository
	cacheService ProfileCacheService
	geoService   GeoValidationService
}

// NewUpdateLocationUseCase creates a new UpdateLocationUseCase instance
func NewUpdateLocationUseCase(
	userRepo repositories.UserRepository,
	cacheService ProfileCacheService,
	geoService GeoValidationService,
) *UpdateLocationUseCase {
	return &UpdateLocationUseCase{
		userRepo:     userRepo,
		cacheService: cacheService,
		geoService:   geoService,
	}
}

// UpdateLocationRequest represents update location request
type UpdateLocationRequest struct {
	UserID    uuid.UUID  `json:"user_id"`
	Latitude  float64    `json:"latitude"`
	Longitude float64    `json:"longitude"`
	City      *string    `json:"city"`
	Country   *string    `json:"country"`
}

// UpdateLocationResponse represents update location response
type UpdateLocationResponse struct {
	Message string `json:"message"`
}

// Execute handles the update location use case
func (uc *UpdateLocationUseCase) Execute(ctx context.Context, req *UpdateLocationRequest) (*UpdateLocationResponse, error) {
	// Validate coordinates
	if err := uc.geoService.ValidateCoordinates(req.Latitude, req.Longitude); err != nil {
		return nil, errors.WrapError(err, "Invalid coordinates")
	}

	// Get current user
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user is active and not banned
	if !user.IsActive || user.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Validate location change frequency
	if err := uc.geoService.ValidateLocationUpdateFrequency(ctx, req.UserID); err != nil {
		return nil, err
	}

	// Reverse geocode to get city and country if not provided
	if req.City == nil || req.Country == nil {
		city, country, err := uc.geoService.ReverseGeocode(req.Latitude, req.Longitude)
		if err == nil {
			if req.City == nil {
				req.City = &city
			}
			if req.Country == nil {
				req.Country = &country
			}
		}
		// Don't fail if geocoding fails, just use provided values
	}

	// Update user location
	user.LocationLat = &req.Latitude
	user.LocationLng = &req.Longitude
	user.LocationCity = req.City
	user.LocationCountry = req.Country

	// Update user in database
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errors.WrapError(err, "Failed to update user location")
	}

	// Invalidate cache
	if err := uc.cacheService.DeleteProfile(ctx, req.UserID); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	// Update location-based caches
	if err := uc.cacheService.UpdateUserLocation(ctx, req.UserID, req.Latitude, req.Longitude); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return &UpdateLocationResponse{
		Message: "Location updated successfully",
	}, nil
}

// CalculateDistance calculates the distance between two points in kilometers
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371

	// Convert latitude and longitude from degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Calculate the change in coordinates
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Apply the Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Calculate the distance
	distance := earthRadiusKm * c

	return distance
}