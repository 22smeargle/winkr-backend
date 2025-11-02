package matching

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// DiscoverUsersUseCase handles user discovery with filtering and pagination
type DiscoverUsersUseCase struct {
	userRepo         repositories.UserRepository
	matchRepo        repositories.MatchRepository
	photoRepo        repositories.PhotoRepository
	matchingService  MatchingAlgorithmService
	swipeService     SwipeService
	cacheService     CacheService
}

// NewDiscoverUsersUseCase creates a new DiscoverUsersUseCase
func NewDiscoverUsersUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	photoRepo repositories.PhotoRepository,
	matchingService MatchingAlgorithmService,
	swipeService SwipeService,
	cacheService CacheService,
) *DiscoverUsersUseCase {
	return &DiscoverUsersUseCase{
		userRepo:        userRepo,
		matchRepo:       matchRepo,
		photoRepo:       photoRepo,
		matchingService: matchingService,
		swipeService:    swipeService,
		cacheService:    cacheService,
	}
}

// DiscoverUsersRequest represents the request to discover users
type DiscoverUsersRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Limit       int       `json:"limit" validate:"min=1,max=100"`
	Offset      int       `json:"offset" validate:"min=0"`
	AgeMin      *int      `json:"age_min,omitempty"`
	AgeMax      *int      `json:"age_max,omitempty"`
	MaxDistance *int      `json:"max_distance,omitempty"` // in kilometers
	Gender      *string   `json:"gender,omitempty"`
	Verified    *bool     `json:"verified,omitempty"`
	HasPhotos   *bool     `json:"has_photos,omitempty"`
}

// DiscoverUsersResponse represents the response from discovering users
type DiscoverUsersResponse struct {
	Users      []*dto.DiscoveryUser `json:"users"`
	Total      int64               `json:"total"`
	HasMore    bool                `json:"has_more"`
	NextCursor string              `json:"next_cursor,omitempty"`
}

// Execute discovers users for the given user with filtering and pagination
func (uc *DiscoverUsersUseCase) Execute(ctx context.Context, req *DiscoverUsersRequest) (*DiscoverUsersResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Get current user
	currentUser, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Get user preferences
	preferences, err := uc.userRepo.GetPreferences(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Apply default values from preferences if not provided in request
	filter := uc.buildDiscoveryFilter(req, preferences, currentUser)

	// Check cache first
	cacheKey := uc.generateCacheKey(req.UserID, filter)
	if cached, err := uc.cacheService.GetDiscoveryUsers(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get already swiped users to exclude them
	swipedUserIDs, err := uc.swipeService.GetSwipedUserIDs(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swiped users: %w", err)
	}

	// Get matched users to exclude them
	matchedUserIDs, err := uc.matchRepo.GetMatchedUserIDs(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matched users: %w", err)
	}

	// Combine excluded user IDs
	excludedUserIDs := append(swipedUserIDs, matchedUserIDs...)

	// Get potential matches using matching algorithm
	potentialUsers, total, err := uc.matchingService.GetPotentialMatches(ctx, currentUser, filter, excludedUserIDs, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get potential matches: %w", err)
	}

	// Convert to DTOs
	discoveryUsers := make([]*dto.DiscoveryUser, 0, len(potentialUsers))
	for _, user := range potentialUsers {
		// Get user photos
		photos, err := uc.photoRepo.GetByUserID(ctx, user.ID)
		if err != nil {
			continue // Skip user if we can't get photos
		}

		// Calculate distance
		distance := uc.calculateDistance(currentUser, user)

		// Create discovery user DTO
		discoveryUser := dto.NewDiscoveryUser(user, photos, distance)
		discoveryUsers = append(discoveryUsers, discoveryUser)
	}

	// Create response
	response := &DiscoverUsersResponse{
		Users:   discoveryUsers,
		Total:   total,
		HasMore: int64(req.Offset+req.Limit) < total,
	}

	// Generate next cursor if there are more results
	if response.HasMore {
		response.NextCursor = fmt.Sprintf("%d", req.Offset+req.Limit)
	}

	// Cache the result
	uc.cacheService.SetDiscoveryUsers(ctx, cacheKey, response, 5*time.Minute)

	return response, nil
}

// buildDiscoveryFilter builds the discovery filter from request and preferences
func (uc *DiscoverUsersUseCase) buildDiscoveryFilter(req *DiscoverUsersRequest, preferences *entities.UserPreferences, currentUser *entities.User) *MatchingFilter {
	filter := &MatchingFilter{
		UserID:        req.UserID,
		AgeMin:        preferences.AgeMin,
		AgeMax:        preferences.AgeMax,
		MaxDistance:   preferences.MaxDistance,
		InterestedIn:  currentUser.InterestedIn,
		ExcludeUserIDs: []uuid.UUID{req.UserID},
	}

	// Override with request parameters if provided
	if req.AgeMin != nil {
		filter.AgeMin = *req.AgeMin
	}
	if req.AgeMax != nil {
		filter.AgeMax = *req.AgeMax
	}
	if req.MaxDistance != nil {
		filter.MaxDistance = *req.MaxDistance
	}
	if req.Gender != nil {
		filter.Gender = *req.Gender
	}
	if req.Verified != nil {
		filter.Verified = *req.Verified
	}
	if req.HasPhotos != nil {
		filter.HasPhotos = *req.HasPhotos
	}

	return filter
}

// calculateDistance calculates the distance between two users in kilometers
func (uc *DiscoverUsersUseCase) calculateDistance(user1, user2 *entities.User) float64 {
	if !user1.HasLocation() || !user2.HasLocation() {
		return 0
	}

	lat1, lng1, _ := user1.GetLocation()
	lat2, lng2, _ := user2.GetLocation()

	// Haversine formula
	const earthRadiusKm = 6371

	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLng/2)*math.Sin(dLng/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// generateCacheKey generates a cache key for discovery results
func (uc *DiscoverUsersUseCase) generateCacheKey(userID uuid.UUID, filter *MatchingFilter) string {
	return fmt.Sprintf("discovery:%s:%d:%d:%d:%s:%t:%t",
		userID.String(),
		filter.AgeMin,
		filter.AgeMax,
		filter.MaxDistance,
		filter.Gender,
		filter.Verified,
		filter.HasPhotos,
	)
}

// Validate validates the request
func (req *DiscoverUsersRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if req.Limit <= 0 {
		req.Limit = 10 // Default limit
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	return nil
}