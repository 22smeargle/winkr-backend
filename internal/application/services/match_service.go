package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// MatchService handles match operations
type MatchService struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	swipeRepo    repositories.SwipeRepository
	cacheService CacheService
}

// NewMatchService creates a new MatchService
func NewMatchService(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	swipeRepo repositories.SwipeRepository,
	cacheService CacheService,
) *MatchService {
	return &MatchService{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		swipeRepo:    swipeRepo,
		cacheService: cacheService,
	}
}

// CreateMatch creates a new match
func (s *MatchService) CreateMatch(ctx context.Context, match *entities.Match) error {
	// Check if match already exists
	exists, err := s.matchRepo.MatchExists(ctx, match.User1ID, match.User2ID)
	if err != nil {
		return fmt.Errorf("failed to check match existence: %w", err)
	}

	if exists {
		return fmt.Errorf("match already exists")
	}

	// Create match
	err = s.matchRepo.CreateMatch(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to create match: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateMatchCaches(ctx, match.User1ID, match.User2ID)

	return nil
}

// CheckForMatch checks if two users have mutually liked each other
func (s *MatchService) CheckForMatch(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, *entities.Match, error) {
	// Check if match already exists
	existingMatch, err := s.matchRepo.GetMatchByUsers(ctx, user1ID, user2ID)
	if err != nil && err != repositories.ErrNotFound {
		return false, nil, fmt.Errorf("failed to check existing match: %w", err)
	}

	if existingMatch != nil {
		return existingMatch.IsActive, existingMatch, nil
	}

	// Check if user1 liked user2
	user1LikedUser2, err := s.swipeRepo.GetSwipeDirection(ctx, user1ID, user2ID)
	if err != nil && err != repositories.ErrNotFound {
		return false, nil, fmt.Errorf("failed to check user1 swipe: %w", err)
	}

	// Check if user2 liked user1
	user2LikedUser1, err := s.swipeRepo.GetSwipeDirection(ctx, user2ID, user1ID)
	if err != nil && err != repositories.ErrNotFound {
		return false, nil, fmt.Errorf("failed to check user2 swipe: %w", err)
	}

	// It's a match if both users liked each other
	isMatch := user1LikedUser2 && user2LikedUser1

	if isMatch {
		// Create new match
		newMatch := &entities.Match{
			User1ID: user1ID,
			User2ID: user2ID,
			IsActive: true,
		}

		err = s.matchRepo.CreateMatch(ctx, newMatch)
		if err != nil {
			return false, nil, fmt.Errorf("failed to create match: %w", err)
		}

		// Invalidate caches
		s.invalidateMatchCaches(ctx, user1ID, user2ID)

		return true, newMatch, nil
	}

	return false, nil, nil
}

// GetMatch gets a match by ID
func (s *MatchService) GetMatch(ctx context.Context, matchID uuid.UUID) (*entities.Match, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("match:%s", matchID.String())
	if cached, err := s.cacheService.GetMatch(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get from database
	match, err := s.matchRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	// Cache result
	s.cacheService.SetMatch(ctx, cacheKey, match, 10*time.Minute)

	return match, nil
}

// GetUserMatches gets matches for a user with pagination
func (s *MatchService) GetUserMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, int64, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("user_matches:%s:%d:%d", userID.String(), limit, offset)
	if cached, err := s.cacheService.GetUserMatches(ctx, cacheKey); err == nil && cached != nil {
		return cached.Matches, cached.Total, nil
	}

	// Get from database
	matches, err := s.matchRepo.GetUserMatches(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user matches: %w", err)
	}

	// Get total count
	total, err := s.matchRepo.GetMatchCount(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get match count: %w", err)
	}

	// Cache result
	s.cacheService.SetUserMatches(ctx, cacheKey, &UserMatchesCache{
		Matches: matches,
		Total:    total,
	}, 5*time.Minute)

	return matches, total, nil
}

// DeactivateMatch deactivates a match
func (s *MatchService) DeactivateMatch(ctx context.Context, matchID uuid.UUID) error {
	// Get match
	match, err := s.matchRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Deactivate
	match.Deactivate()

	// Update
	err = s.matchRepo.UpdateMatch(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	// Invalidate caches
	s.invalidateMatchCaches(ctx, match.User1ID, match.User2ID)

	return nil
}

// Unmatch removes a match between two users
func (s *MatchService) Unmatch(ctx context.Context, user1ID, user2ID uuid.UUID) error {
	// Get match
	match, err := s.matchRepo.GetMatchByUsers(ctx, user1ID, user2ID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Deactivate instead of deleting to preserve history
	match.Deactivate()

	// Update
	err = s.matchRepo.UpdateMatch(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	// Invalidate caches
	s.invalidateMatchCaches(ctx, user1ID, user2ID)

	return nil
}

// GetMatchQuality calculates quality score for a match
func (s *MatchService) GetMatchQuality(ctx context.Context, match *entities.Match) (*MatchQuality, error) {
	// Get users
	user1, err := s.userRepo.GetByID(ctx, match.User1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user1: %w", err)
	}

	user2, err := s.userRepo.GetByID(ctx, match.User2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user2: %w", err)
	}

	// Calculate quality factors
	distance := s.calculateDistance(user1, user2)
	ageDiff := s.calculateAgeDifference(user1, user2)
	verificationCompatibility := s.calculateVerificationCompatibility(user1, user2)
	activityScore := s.calculateActivityScore(user1, user2)

	// Calculate overall quality score (0-100)
	qualityScore := s.calculateQualityScore(distance, ageDiff, verificationCompatibility, activityScore)

	return &MatchQuality{
		MatchID:               match.ID,
		QualityScore:          qualityScore,
		Distance:              distance,
		AgeDifference:          ageDiff,
		VerificationCompatibility: verificationCompatibility,
		ActivityScore:          activityScore,
		CalculatedAt:          time.Now(),
	}, nil
}

// calculateDistance calculates distance between two users
func (s *MatchService) calculateDistance(user1, user2 *entities.User) float64 {
	if !user1.HasLocation() || !user2.HasLocation() {
		return 0
	}

	lat1, lng1, _ := user1.GetLocation()
	lat2, lng2, _ := user2.GetLocation()

	// Use matching algorithm service's distance calculation
	matchingService := NewMatchingAlgorithmService(s.userRepo, nil, s.matchRepo, s.cacheService)
	return matchingService.calculateDistance(user1, user2)
}

// calculateAgeDifference calculates age difference between two users
func (s *MatchService) calculateAgeDifference(user1, user2 *entities.User) int {
	age1 := user1.GetAge()
	age2 := user2.GetAge()
	
	if age1 > age2 {
		return age1 - age2
	}
	return age2 - age1
}

// calculateVerificationCompatibility calculates verification compatibility
func (s *MatchService) calculateVerificationCompatibility(user1, user2 *entities.User) float64 {
	// Higher compatibility if both users have similar verification levels
	level1 := int(user1.VerificationLevel)
	level2 := int(user2.VerificationLevel)

	diff := level1 - level2
	if diff < 0 {
		diff = -diff
	}

	// Perfect match (same level) = 100
	// Each level difference = -25 points
	compatibility := 100 - (diff * 25)
	if compatibility < 0 {
		compatibility = 0
	}

	return compatibility
}

// calculateActivityScore calculates activity compatibility
func (s *MatchService) calculateActivityScore(user1, user2 *entities.User) float64 {
	score1 := s.calculateUserActivityScore(user1)
	score2 := s.calculateUserActivityScore(user2)

	// Average of both users' activity scores
	return (score1 + score2) / 2
}

// calculateUserActivityScore calculates activity score for a single user
func (s *MatchService) calculateUserActivityScore(user *entities.User) float64 {
	if user.LastActive == nil {
		return 0
	}

	hoursSinceActive := time.Since(*user.LastActive).Hours()

	// More recent activity = higher score
	// 0-24h = 100, 24-72h = 50, 72h+ = 0
	if hoursSinceActive <= 24 {
		return 100
	}
	if hoursSinceActive <= 72 {
		return 50
	}
	return 0
}

// calculateQualityScore calculates overall match quality score
func (s *MatchService) calculateQualityScore(distance, ageDiff, verificationCompatibility, activityScore float64) float64 {
	// Weighted calculation
	// Distance: 30% (closer is better)
	// Age difference: 20% (smaller difference is better)
	// Verification compatibility: 25% (higher is better)
	// Activity score: 25% (higher is better)

	distanceScore := s.calculateDistanceQualityScore(distance)
	ageScore := s.calculateAgeQualityScore(ageDiff)

	totalScore := (distanceScore * 0.30) + (ageScore * 0.20) + 
		(verificationCompatibility * 0.25) + (activityScore * 0.25)

	return totalScore
}

// calculateDistanceQualityScore converts distance to quality score
func (s *MatchService) calculateDistanceQualityScore(distance float64) float64 {
	// 0-10km = 100, 10-50km = 50, 50km+ = 0
	if distance <= 10 {
		return 100
	}
	if distance <= 50 {
		return 50
	}
	return 0
}

// calculateAgeQualityScore converts age difference to quality score
func (s *MatchService) calculateAgeQualityScore(ageDiff int) float64 {
	// 0-2 years = 100, 2-5 years = 75, 5-10 years = 50, 10+ years = 0
	if ageDiff <= 2 {
		return 100
	}
	if ageDiff <= 5 {
		return 75
	}
	if ageDiff <= 10 {
		return 50
	}
	return 0
}

// invalidateMatchCaches invalidates caches related to matches
func (s *MatchService) invalidateMatchCaches(ctx context.Context, user1ID, user2ID uuid.UUID) {
	// Invalidate match caches for both users
	for _, userID := range []uuid.UUID{user1ID, user2ID} {
		// User matches cache
		matchesCacheKey := fmt.Sprintf("user_matches:%s:", userID.String())
		s.cacheService.DeletePattern(ctx, matchesCacheKey+"*")

		// Discovery cache
		s.cacheService.InvalidateUserDiscoveryCache(ctx, userID)
	}

	// Invalidate specific match cache
	matchCacheKey := fmt.Sprintf("match:%s", user1ID.String()+"_"+user2ID.String())
	s.cacheService.Delete(ctx, matchCacheKey)
}

// MatchQuality represents quality analysis of a match
type MatchQuality struct {
	MatchID               uuid.UUID `json:"match_id"`
	QualityScore          float64   `json:"quality_score"`
	Distance              float64   `json:"distance"`
	AgeDifference          int       `json:"age_difference"`
	VerificationCompatibility float64   `json:"verification_compatibility"`
	ActivityScore          float64   `json:"activity_score"`
	CalculatedAt          time.Time `json:"calculated_at"`
}

// UserMatchesCache represents cached user matches
type UserMatchesCache struct {
	Matches []*entities.Match `json:"matches"`
	Total    int64            `json:"total"`
}