package matching

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// GetDiscoveryStatsUseCase handles getting discovery statistics
type GetDiscoveryStatsUseCase struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	swipeService SwipeService
	cacheService CacheService
}

// NewGetDiscoveryStatsUseCase creates a new GetDiscoveryStatsUseCase
func NewGetDiscoveryStatsUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	swipeService SwipeService,
	cacheService CacheService,
) *GetDiscoveryStatsUseCase {
	return &GetDiscoveryStatsUseCase{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		swipeService: swipeService,
		cacheService: cacheService,
	}
}

// GetDiscoveryStatsRequest represents a request to get discovery stats
type GetDiscoveryStatsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetDiscoveryStatsResponse represents the response with discovery statistics
type GetDiscoveryStatsResponse struct {
	*dto.DiscoveryStats
}

// Execute gets discovery statistics for a user
func (uc *GetDiscoveryStatsUseCase) Execute(ctx context.Context, req *GetDiscoveryStatsRequest) (*GetDiscoveryStatsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check cache first
	cacheKey := uc.generateCacheKey(req.UserID)
	if cached, err := uc.cacheService.GetDiscoveryStats(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get user stats
	userStats, err := uc.userRepo.GetUserStats(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get match stats
	matchStats, err := uc.matchRepo.GetMatchStats(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match stats: %w", err)
	}

	// Get swipe stats
	swipeStats, err := uc.swipeService.GetSwipeStats(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swipe stats: %w", err)
	}

	// Calculate additional stats
	likeRate := float64(0)
	if swipeStats.TotalSwipes > 0 {
		likeRate = float64(swipeStats.TotalLikes) / float64(swipeStats.TotalSwipes) * 100
	}

	matchRate := float64(0)
	if swipeStats.TotalSwipes > 0 {
		matchRate = float64(matchStats.TotalMatches) / float64(swipeStats.TotalSwipes) * 100
	}

	// Create discovery stats DTO
	discoveryStats := &dto.DiscoveryStats{
		UserID:           req.UserID,
		TotalSwipes:       swipeStats.TotalSwipes,
		TotalLikes:        swipeStats.TotalLikes,
		TotalPasses:       swipeStats.TotalPasses,
		TotalMatches:      matchStats.TotalMatches,
		ActiveMatches:     matchStats.ActiveMatches,
		LikeRate:          likeRate,
		MatchRate:         matchRate,
		SwipesToday:       swipeStats.SwipesToday,
		SwipesThisWeek:   swipeStats.SwipesThisWeek,
		SwipesThisMonth:  swipeStats.SwipesThisMonth,
		MatchesToday:      matchStats.MatchesToday,
		MatchesThisWeek:   matchStats.MatchesThisWeek,
		MatchesThisMonth:  matchStats.MatchesThisMonth,
		ProfileViews:       userStats.ProfileViews,
		PhotosCount:       userStats.PhotosCount,
		LastActiveDays:    userStats.LastActiveDays,
		AccountAgeDays:    userStats.AccountAgeDays,
		GeneratedAt:       time.Now(),
	}

	// Create response
	response := &GetDiscoveryStatsResponse{
		DiscoveryStats: discoveryStats,
	}

	// Cache result for 10 minutes
	uc.cacheService.SetDiscoveryStats(ctx, cacheKey, response, 10*time.Minute)

	return response, nil
}

// generateCacheKey generates a cache key for discovery stats
func (uc *GetDiscoveryStatsUseCase) generateCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf("discovery_stats:%s", userID.String())
}

// Validate validates the request
func (req *GetDiscoveryStatsRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	return nil
}