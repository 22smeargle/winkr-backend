package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// MatchingAlgorithmService handles user matching algorithm
type MatchingAlgorithmService struct {
	userRepo    repositories.UserRepository
	photoRepo    repositories.PhotoRepository
	matchRepo    repositories.MatchRepository
	cacheService CacheService
}

// NewMatchingAlgorithmService creates a new MatchingAlgorithmService
func NewMatchingAlgorithmService(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	matchRepo repositories.MatchRepository,
	cacheService CacheService,
) *MatchingAlgorithmService {
	return &MatchingAlgorithmService{
		userRepo:    userRepo,
		photoRepo:    photoRepo,
		matchRepo:    matchRepo,
		cacheService: cacheService,
	}
}

// MatchingFilter represents filtering criteria for matching
type MatchingFilter struct {
	UserID         uuid.UUID  `json:"user_id"`
	AgeMin         int         `json:"age_min"`
	AgeMax         int         `json:"age_max"`
	MaxDistance     int         `json:"max_distance"`     // in kilometers
	Gender         *string     `json:"gender,omitempty"`
	InterestedIn   []string    `json:"interested_in"`
	Verified       *bool       `json:"verified,omitempty"`
	HasPhotos      *bool       `json:"has_photos,omitempty"`
	ExcludeUserIDs []uuid.UUID `json:"exclude_user_ids"`
}

// UserScore represents a user with their matching score
type UserScore struct {
	User       *entities.User `json:"user"`
	Score      float64        `json:"score"`
	Distance   float64        `json:"distance"`
	Recency    float64        `json:"recency"`
	Completion float64        `json:"completion"`
	Verification float64      `json:"verification"`
	Premium    float64        `json:"premium"`
}

// GetPotentialMatches gets potential matches for a user
func (s *MatchingAlgorithmService) GetPotentialMatches(
	ctx context.Context,
	user *entities.User,
	filter *MatchingFilter,
	excludeUserIDs []uuid.UUID,
	limit, offset int,
) ([]*entities.User, int64, error) {
	// Check cache first
	cacheKey := s.generateCacheKey(user.ID, filter, excludeUserIDs, limit, offset)
	if cached, err := s.cacheService.GetPotentialMatches(ctx, cacheKey); err == nil && cached != nil {
		return cached.Users, cached.Total, nil
	}

	// Get base candidates using location and preferences
	candidates, err := s.getBaseCandidates(ctx, user, filter, excludeUserIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get base candidates: %w", err)
	}

	// Score candidates
	scoredUsers := s.scoreCandidates(ctx, user, candidates)

	// Sort by score (descending)
	sort.Slice(scoredUsers, func(i, j int) bool {
		return scoredUsers[i].Score > scoredUsers[j].Score
	})

	// Apply pagination
	start := offset
	if start > len(scoredUsers) {
		start = len(scoredUsers)
	}

	end := start + limit
	if end > len(scoredUsers) {
		end = len(scoredUsers)
	}

	paginatedUsers := scoredUsers[start:end]
	result := make([]*entities.User, len(paginatedUsers))
	for i, scoredUser := range paginatedUsers {
		result[i] = scoredUser.User
	}

	// Cache result
	s.cacheService.SetPotentialMatches(ctx, cacheKey, &PotentialMatchesCache{
		Users: result,
		Total: int64(len(candidates)),
	}, 5*time.Minute)

	return result, int64(len(candidates)), nil
}

// getBaseCandidates gets base candidates using database queries
func (s *MatchingAlgorithmService) getBaseCandidates(
	ctx context.Context,
	user *entities.User,
	filter *MatchingFilter,
	excludeUserIDs []uuid.UUID,
) ([]*entities.User, error) {
	// Combine exclude user IDs
	allExcludes := append(filter.ExcludeUserIDs, excludeUserIDs...)

	// Get candidates by location if user has location
	if user.HasLocation() {
		lat, lng, _ := user.GetLocation()
		return s.userRepo.GetByLocation(ctx, lat, lng, filter.MaxDistance, 1000, 0) // Get up to 1000 candidates
	}

	// Fallback to preference-based search
	return s.userRepo.GetUsersByPreferences(ctx, user.ID, &entities.UserPreferences{
		AgeMin:      filter.AgeMin,
		AgeMax:      filter.AgeMax,
		MaxDistance:  filter.MaxDistance,
		ShowMe:      true,
	}, 1000, 0)
}

// scoreCandidates scores candidates based on various factors
func (s *MatchingAlgorithmService) scoreCandidates(ctx context.Context, currentUser *entities.User, candidates []*entities.User) []*UserScore {
	scoredUsers := make([]*UserScore, 0, len(candidates))

	for _, candidate := range candidates {
		// Skip if candidate doesn't meet basic criteria
		if !s.meetsBasicCriteria(currentUser, candidate) {
			continue
		}

		score := s.calculateScore(currentUser, candidate)
		scoredUsers = append(scoredUsers, score)
	}

	return scoredUsers
}

// meetsBasicCriteria checks if candidate meets basic matching criteria
func (s *MatchingAlgorithmService) meetsBasicCriteria(currentUser, candidate *entities.User) bool {
	// Skip self
	if currentUser.ID == candidate.ID {
		return false
	}

	// Skip inactive/banned users
	if !candidate.IsActive || candidate.IsBanned {
		return false
	}

	// Check gender preference
	if len(currentUser.InterestedIn) > 0 {
		found := false
		for _, interested := range currentUser.InterestedIn {
			if candidate.Gender == interested {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check candidate's preferences (simplified - would need to fetch preferences)
	// For now, just check age
	candidateAge := candidate.GetAge()
	if candidateAge < 18 || candidateAge > 100 {
		return false
	}

	return true
}

// calculateScore calculates matching score for a candidate
func (s *MatchingAlgorithmService) calculateScore(currentUser, candidate *entities.User) *UserScore {
	distance := s.calculateDistance(currentUser, candidate)
	recency := s.calculateRecencyScore(candidate)
	completion := s.calculateCompletionScore(candidate)
	verification := s.calculateVerificationScore(candidate)
	premium := s.calculatePremiumScore(candidate)

	// Weighted score calculation
	// Distance: 30% (closer is better)
	// Recency: 20% (more recent is better)
	// Completion: 20% (more complete is better)
	// Verification: 15% (more verified is better)
	// Premium: 15% (premium users get boost)

	distanceScore := s.calculateDistanceScore(distance)
	totalScore := (distanceScore * 0.30) + (recency * 0.20) + (completion * 0.20) + (verification * 0.15) + (premium * 0.15)

	return &UserScore{
		User:        candidate,
		Score:       totalScore,
		Distance:    distance,
		Recency:     recency,
		Completion:  completion,
		Verification: verification,
		Premium:     premium,
	}
}

// calculateDistance calculates distance between two users in kilometers
func (s *MatchingAlgorithmService) calculateDistance(user1, user2 *entities.User) float64 {
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

// calculateDistanceScore converts distance to a score (0-100)
func (s *MatchingAlgorithmService) calculateDistanceScore(distance float64) float64 {
	// Closer distance gets higher score
	// 0km = 100, 50km = 50, 100km+ = 0
	if distance <= 0 {
		return 100
	}
	if distance >= 100 {
		return 0
	}
	return 100 - distance
}

// calculateRecencyScore calculates score based on last active time
func (s *MatchingAlgorithmService) calculateRecencyScore(user *entities.User) float64 {
	if user.LastActive == nil {
		return 0
	}

	hoursSinceActive := time.Since(*user.LastActive).Hours()

	// More recent activity gets higher score
	// 0-24h = 100, 24-72h = 50, 72h+ = 0
	if hoursSinceActive <= 24 {
		return 100
	}
	if hoursSinceActive <= 72 {
		return 50
	}
	return 0
}

// calculateCompletionScore calculates score based on profile completion
func (s *MatchingAlgorithmService) calculateCompletionScore(user *entities.User) float64 {
	if user.IsComplete() {
		return 100
	}

	score := 0
	if user.FirstName != "" {
		score += 20
	}
	if user.LastName != "" {
		score += 20
	}
	if !user.DateOfBirth.IsZero() {
		score += 20
	}
	if user.Gender != "" {
		score += 20
	}
	if len(user.InterestedIn) > 0 {
		score += 20
	}

	return score
}

// calculateVerificationScore calculates score based on verification level
func (s *MatchingAlgorithmService) calculateVerificationScore(user *entities.User) float64 {
	switch user.VerificationLevel {
	case entities.VerificationLevelNone:
		return 0
	case entities.VerificationLevelSelfie:
		return 50
	case entities.VerificationLevelDocument:
		return 100
	default:
		return 0
	}
}

// calculatePremiumScore calculates score based on premium status
func (s *MatchingAlgorithmService) calculatePremiumScore(user *entities.User) float64 {
	if user.IsPremium {
		return 100
	}
	return 0
}

// generateCacheKey generates cache key for potential matches
func (s *MatchingAlgorithmService) generateCacheKey(
	userID uuid.UUID,
	filter *MatchingFilter,
	excludeUserIDs []uuid.UUID,
	limit, offset int,
) string {
	return fmt.Sprintf("potential_matches:%s:%d:%d:%d:%s:%t:%t",
		userID.String(),
		filter.AgeMin,
		filter.AgeMax,
		filter.MaxDistance,
		filter.Gender,
		filter.Verified,
		filter.HasPhotos,
	)
}

// PotentialMatchesCache represents cached potential matches
type PotentialMatchesCache struct {
	Users []*entities.User `json:"users"`
	Total int64            `json:"total"`
}