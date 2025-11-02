package matching

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// SuperLikeUserUseCase handles super liking a user (premium feature)
type SuperLikeUserUseCase struct {
	userRepo        repositories.UserRepository
	matchRepo       repositories.MatchRepository
	subscriptionRepo repositories.SubscriptionRepository
	swipeService    SwipeService
	matchService    MatchService
	cacheService    CacheService
}

// NewSuperLikeUserUseCase creates a new SuperLikeUserUseCase
func NewSuperLikeUserUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	swipeService SwipeService,
	matchService MatchService,
	cacheService CacheService,
) *SuperLikeUserUseCase {
	return &SuperLikeUserUseCase{
		userRepo:        userRepo,
		matchRepo:       matchRepo,
		subscriptionRepo: subscriptionRepo,
		swipeService:    swipeService,
		matchService:    matchService,
		cacheService:    cacheService,
	}
}

// SuperLikeUserRequest represents a request to super like a user
type SuperLikeUserRequest struct {
	SwiperID uuid.UUID `json:"swiper_id" validate:"required"`
	SwipedID uuid.UUID `json:"swiped_id" validate:"required"`
}

// SuperLikeUserResponse represents the response from super liking a user
type SuperLikeUserResponse struct {
	IsMatch bool     `json:"is_match"`
	Match   *dto.Match `json:"match,omitempty"`
}

// Execute super likes a user and checks for mutual match
func (uc *SuperLikeUserUseCase) Execute(ctx context.Context, req *SuperLikeUserRequest) (*SuperLikeUserResponse, error) {
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

	// Check if user has premium subscription (required for super likes)
	hasPremium, err := uc.checkPremiumAccess(ctx, req.SwiperID)
	if err != nil {
		return nil, fmt.Errorf("failed to check premium access: %w", err)
	}

	if !hasPremium {
		return nil, fmt.Errorf("super like requires premium subscription")
	}

	// Check if already swiped
	hasSwiped, err := uc.swipeService.HasSwiped(ctx, req.SwiperID, req.SwipedID)
	if err != nil {
		return nil, fmt.Errorf("failed to check swipe status: %w", err)
	}

	if hasSwiped {
		return nil, fmt.Errorf("user already swiped")
	}

	// Check daily super like limit
	withinLimit, err := uc.swipeService.CheckSuperLikeLimit(ctx, req.SwiperID)
	if err != nil {
		return nil, fmt.Errorf("failed to check super like limit: %w", err)
	}

	if !withinLimit {
		return nil, fmt.Errorf("daily super like limit exceeded")
	}

	// Create super like swipe
	swipe := &entities.Swipe{
		SwiperID: req.SwiperID,
		SwipedID: req.SwipedID,
		IsLike:   true,
	}

	err = uc.swipeService.CreateSuperLike(ctx, swipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create super like: %w", err)
	}

	// Check for mutual match
	isMatch, existingMatch, err := uc.matchService.CheckForMatch(ctx, req.SwiperID, req.SwipedID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for match: %w", err)
	}

	response := &SuperLikeUserResponse{
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

// checkPremiumAccess checks if user has premium subscription
func (uc *SuperLikeUserUseCase) checkPremiumAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	subscription, err := uc.subscriptionRepo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if user has active premium or platinum subscription
	return subscription != nil && (subscription.PlanType == "premium" || subscription.PlanType == "platinum"), nil
}

// invalidateDiscoveryCache invalidates discovery cache for a user
func (uc *SuperLikeUserUseCase) invalidateDiscoveryCache(ctx context.Context, userID uuid.UUID) {
	uc.cacheService.InvalidateUserDiscoveryCache(ctx, userID)
}

// Validate validates the request
func (req *SuperLikeUserRequest) Validate() error {
	if req.SwiperID == uuid.Nil {
		return fmt.Errorf("swiper_id is required")
	}
	if req.SwipedID == uuid.Nil {
		return fmt.Errorf("swiped_id is required")
	}
	if req.SwiperID == req.SwipedID {
		return fmt.Errorf("cannot super like yourself")
	}
	return nil
}