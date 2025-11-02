package matching

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// LikeUserUseCase handles liking a user
type LikeUserUseCase struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	swipeService SwipeService
	matchService MatchService
	cacheService CacheService
}

// NewLikeUserUseCase creates a new LikeUserUseCase
func NewLikeUserUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	swipeService SwipeService,
	matchService MatchService,
	cacheService CacheService,
) *LikeUserUseCase {
	return &LikeUserUseCase{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		swipeService: swipeService,
		matchService: matchService,
		cacheService: cacheService,
	}
}

// LikeUserRequest represents a request to like a user
type LikeUserRequest struct {
	SwiperID uuid.UUID `json:"swiper_id" validate:"required"`
	SwipedID uuid.UUID `json:"swiped_id" validate:"required"`
}

// LikeUserResponse represents the response from liking a user
type LikeUserResponse struct {
	IsMatch bool     `json:"is_match"`
	Match   *dto.Match `json:"match,omitempty"`
}

// Execute likes a user and checks for mutual match
func (uc *LikeUserUseCase) Execute(ctx context.Context, req *LikeUserRequest) (*LikeUserResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if users exist
	swiper, err := uc.userRepo.GetByID(ctx, req.SwiperID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swiper: %w", err)
	}

	swiped, err := uc.userRepo.GetByID(ctx, req.SwipedID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swiped user: %w", err)
	}

	// Check if already swiped
	hasSwiped, err := uc.swipeService.HasSwiped(ctx, req.SwiperID, req.SwipedID)
	if err != nil {
		return nil, fmt.Errorf("failed to check swipe status: %w", err)
	}

	if hasSwiped {
		return nil, fmt.Errorf("user already swiped")
	}

	// Create like swipe
	swipe := &entities.Swipe{
		SwiperID: req.SwiperID,
		SwipedID: req.SwipedID,
		IsLike:   true,
	}

	err = uc.swipeService.CreateSwipe(ctx, swipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create swipe: %w", err)
	}

	// Check for mutual match
	isMatch, existingMatch, err := uc.matchService.CheckForMatch(ctx, req.SwiperID, req.SwipedID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for match: %w", err)
	}

	response := &LikeUserResponse{
		IsMatch: isMatch,
	}

	// If it's a match, create match and return match details
	if isMatch {
		var match *entities.Match
		if existingMatch != nil {
			match = existingMatch
		} else {
			// Create new match
			match = &entities.Match{
				User1ID: req.SwiperID,
				User2ID: req.SwipedID,
				IsActive: true,
			}

			err = uc.matchService.CreateMatch(ctx, match)
			if err != nil {
				return nil, fmt.Errorf("failed to create match: %w", err)
			}
		}

		// Convert to DTO
		matchDTO := dto.NewMatch(match, swiper, swiped)
		response.Match = matchDTO

		// Invalidate discovery cache for both users
		uc.invalidateDiscoveryCache(ctx, req.SwiperID)
		uc.invalidateDiscoveryCache(ctx, req.SwipedID)
	}

	return response, nil
}

// invalidateDiscoveryCache invalidates discovery cache for a user
func (uc *LikeUserUseCase) invalidateDiscoveryCache(ctx context.Context, userID uuid.UUID) {
	// This would invalidate all discovery cache keys for the user
	// Implementation depends on cache service design
	uc.cacheService.InvalidateUserDiscoveryCache(ctx, userID)
}

// Validate validates the request
func (req *LikeUserRequest) Validate() error {
	if req.SwiperID == uuid.Nil {
		return fmt.Errorf("swiper_id is required")
	}
	if req.SwipedID == uuid.Nil {
		return fmt.Errorf("swiped_id is required")
	}
	if req.SwiperID == req.SwipedID {
		return fmt.Errorf("cannot like yourself")
	}
	return nil
}