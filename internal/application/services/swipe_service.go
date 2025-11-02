package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// SwipeService handles swipe operations
type SwipeService struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	swipeRepo    repositories.SwipeRepository
	cacheService CacheService
	rateLimiter  RateLimiter
}

// NewSwipeService creates a new SwipeService
func NewSwipeService(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	swipeRepo repositories.SwipeRepository,
	cacheService CacheService,
	rateLimiter RateLimiter,
) *SwipeService {
	return &SwipeService{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		swipeRepo:    swipeRepo,
		cacheService: cacheService,
		rateLimiter:  rateLimiter,
	}
}

// CreateSwipe creates a new swipe
func (s *SwipeService) CreateSwipe(ctx context.Context, swipe *entities.Swipe) error {
	// Check rate limit
	allowed, err := s.rateLimiter.AllowSwipe(ctx, swipe.SwiperID)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	if !allowed {
		return fmt.Errorf("swipe rate limit exceeded")
	}

	// Create swipe
	err = s.swipeRepo.CreateSwipe(ctx, swipe)
	if err != nil {
		return fmt.Errorf("failed to create swipe: %w", err)
	}

	// Update user's last active time
	err = s.userRepo.UpdateLastActive(ctx, swipe.SwiperID)
	if err != nil {
		// Log error but don't fail the operation
		// This is a non-critical operation
	}

	// Invalidate relevant caches
	s.invalidateSwipeCaches(ctx, swipe.SwiperID, swipe.SwipedID)

	return nil
}

// CreateSuperLike creates a super like swipe with additional validation
func (s *SwipeService) CreateSuperLike(ctx context.Context, swipe *entities.Swipe) error {
	// Check rate limit for super likes
	allowed, err := s.rateLimiter.AllowSuperLike(ctx, swipe.SwiperID)
	if err != nil {
		return fmt.Errorf("failed to check super like rate limit: %w", err)
	}

	if !allowed {
		return fmt.Errorf("daily super like limit exceeded")
	}

	// Create the swipe
	return s.CreateSwipe(ctx, swipe)
}

// HasSwiped checks if user has already swiped on another user
func (s *SwipeService) HasSwiped(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) {
	exists, err := s.swipeRepo.SwipeExists(ctx, swiperID, swipedID)
	if err != nil {
		return false, fmt.Errorf("failed to check swipe existence: %w", err)
	}
	return exists, nil
}

// GetSwipeDirection gets the direction of a swipe (like/pass)
func (s *SwipeService) GetSwipeDirection(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) {
	isLike, err := s.swipeRepo.GetSwipeDirection(ctx, swiperID, swipedID)
	if err != nil {
		return false, fmt.Errorf("failed to get swipe direction: %w", err)
	}
	return isLike, nil
}

// GetSwipedUserIDs gets all user IDs that the user has swiped on
func (s *SwipeService) GetSwipedUserIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("swiped_users:%s", userID.String())
	if cached, err := s.cacheService.GetSwipedUsers(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get swipes from database
	swipes, err := s.swipeRepo.GetUserSwipes(ctx, userID, 10000, 0) // Get up to 10k swipes
	if err != nil {
		return nil, fmt.Errorf("failed to get user swipes: %w", err)
	}

	// Extract user IDs
	swipedIDs := make([]uuid.UUID, len(swipes))
	for i, swipe := range swipes {
		swipedIDs[i] = swipe.SwipedID
	}

	// Cache result
	s.cacheService.SetSwipedUsers(ctx, cacheKey, swipedIDs, 10*time.Minute)

	return swipedIDs, nil
}

// GetSwipeStats gets swipe statistics for a user
func (s *SwipeService) GetSwipeStats(ctx context.Context, userID uuid.UUID) (*repositories.SwipeStats, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("swipe_stats:%s", userID.String())
	if cached, err := s.cacheService.GetSwipeStats(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get stats from repository
	stats, err := s.swipeRepo.GetSwipeStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swipe stats: %w", err)
	}

	// Cache result
	s.cacheService.SetSwipeStats(ctx, cacheKey, stats, 5*time.Minute)

	return stats, nil
}

// CheckSuperLikeLimit checks if user is within daily super like limit
func (s *SwipeService) CheckSuperLikeLimit(ctx context.Context, userID uuid.UUID) (bool, error) {
	// This would typically check against a daily counter in Redis or database
	// For now, we'll use the rate limiter
	return s.rateLimiter.AllowSuperLike(ctx, userID)
}

// AnalyzeSwipePattern analyzes swipe patterns for bot detection
func (s *SwipeService) AnalyzeSwipePattern(ctx context.Context, userID uuid.UUID) (*SwipePatternAnalysis, error) {
	// Get recent swipes
	swipes, err := s.swipeRepo.GetUserSwipes(ctx, userID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent swipes: %w", err)
	}

	// Analyze patterns
	analysis := &SwipePatternAnalysis{
		UserID:      userID,
		TotalSwipes:  len(swipes),
		LikeRate:     s.calculateLikeRate(swipes),
		SwipeSpeed:   s.calculateSwipeSpeed(swipes),
		TimePattern:   s.analyzeTimePattern(swipes),
		IsSuspicious: s.isSuspiciousPattern(swipes),
		AnalyzedAt:   time.Now(),
	}

	return analysis, nil
}

// calculateLikeRate calculates the percentage of likes
func (s *SwipeService) calculateLikeRate(swipes []*entities.Swipe) float64 {
	if len(swipes) == 0 {
		return 0
	}

	likeCount := 0
	for _, swipe := range swipes {
		if swipe.IsLike {
			likeCount++
		}
	}

	return float64(likeCount) / float64(len(swipes)) * 100
}

// calculateSwipeSpeed calculates average time between swipes
func (s *SwipeService) calculateSwipeSpeed(swipes []*entities.Swipe) float64 {
	if len(swipes) < 2 {
		return 0
	}

	var totalDuration time.Duration
	for i := 1; i < len(swipes); i++ {
		duration := swipes[i].CreatedAt.Sub(swipes[i-1].CreatedAt)
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(len(swipes)-1)
	return avgDuration.Seconds()
}

// analyzeTimePattern analyzes the time pattern of swipes
func (s *SwipeService) analyzeTimePattern(swipes []*entities.Swipe) string {
	if len(swipes) == 0 {
		return "no_data"
	}

	// Simple analysis: check if swipes are clustered in specific hours
	hourCounts := make(map[int]int)
	for _, swipe := range swipes {
		hour := swipe.CreatedAt.Hour()
		hourCounts[hour]++
	}

	// Find the hour with most swipes
	maxHour := 0
	maxCount := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}

	// If most swipes are in a small time window, it might be suspicious
	if maxCount > len(swipes)/2 {
		return fmt.Sprintf("clustered_%d", maxHour)
	}

	return "distributed"
}

// isSuspiciousPattern checks if swipe pattern looks like a bot
func (s *SwipeService) isSuspiciousPattern(swipes []*entities.Swipe) bool {
	if len(swipes) < 10 {
		return false
	}

	// Check for too fast swiping (less than 1 second per swipe)
	avgSpeed := s.calculateSwipeSpeed(swipes)
	if avgSpeed < 1.0 {
		return true
	}

	// Check for consistent like rate (bots often have very consistent rates)
	likeRate := s.calculateLikeRate(swipes)
	if likeRate > 95 && likeRate < 100 {
		return true
	}

	// Check for perfect time intervals (every X seconds)
	for i := 2; i < len(swipes); i++ {
		duration1 := swipes[i-1].CreatedAt.Sub(swipes[i-2].CreatedAt)
		duration2 := swipes[i].CreatedAt.Sub(swipes[i-1].CreatedAt)
		
		// If durations are very similar (within 100ms), it's suspicious
		diff := duration1 - duration2
		if diff < 100*time.Millisecond && diff > -100*time.Millisecond {
			return true
		}
	}

	return false
}

// invalidateSwipeCaches invalidates caches related to swipes
func (s *SwipeService) invalidateSwipeCaches(ctx context.Context, swiperID, swipedID uuid.UUID) {
	// Invalidate swiped users cache for swiper
	swiperCacheKey := fmt.Sprintf("swiped_users:%s", swiperID.String())
	s.cacheService.Delete(ctx, swiperCacheKey)

	// Invalidate swipe stats for swiper
	statsCacheKey := fmt.Sprintf("swipe_stats:%s", swiperID.String())
	s.cacheService.Delete(ctx, statsCacheKey)

	// Invalidate discovery cache for both users
	s.cacheService.InvalidateUserDiscoveryCache(ctx, swiperID)
	s.cacheService.InvalidateUserDiscoveryCache(ctx, swipedID)
}

// SwipePatternAnalysis represents the result of swipe pattern analysis
type SwipePatternAnalysis struct {
	UserID      uuid.UUID `json:"user_id"`
	TotalSwipes  int       `json:"total_swipes"`
	LikeRate     float64   `json:"like_rate"`
	SwipeSpeed   float64   `json:"swipe_speed_seconds"`
	TimePattern  string    `json:"time_pattern"`
	IsSuspicious bool      `json:"is_suspicious"`
	AnalyzedAt   time.Time `json:"analyzed_at"`
}