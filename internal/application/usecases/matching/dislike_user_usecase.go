package matching

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// DislikeUserUseCase handles disliking/skipping a user
type DislikeUserUseCase struct {
	userRepo     repositories.UserRepository
	swipeService SwipeService
	cacheService CacheService
}

// NewDislikeUserUseCase creates a new DislikeUserUseCase
func NewDislikeUserUseCase(
	userRepo repositories.UserRepository,
	swipeService SwipeService,
	cacheService CacheService,
) *DislikeUserUseCase {
	return &DislikeUserUseCase{
		userRepo:     userRepo,
		swipeService: swipeService,
		cacheService: cacheService,
	}
}

// DislikeUserRequest represents a request to dislike a user
type DislikeUserRequest struct {
	SwiperID uuid.UUID `json:"swiper_id" validate:"required"`
	SwipedID uuid.UUID `json:"swiped_id" validate:"required"`
}

// DislikeUserResponse represents the response from disliking a user
type DislikeUserResponse struct {
	Success bool `json:"success"`
}

// Execute dislikes a user
func (uc *DislikeUserUseCase) Execute(ctx context.Context, req *DislikeUserRequest) (*DislikeUserResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if users exist
	_, err := uc.userRepo.GetByID(ctx, req.SwiperID)
	if err != nil {
		return nil, fmt.Errorf("failed to get swiper: %w", err)
	}

	_, err = uc.userRepo.GetByID(ctx, req.SwipedID)
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

	// Create dislike swipe
	swipe := &entities.Swipe{
		SwiperID: req.SwiperID,
		SwipedID: req.SwipedID,
		IsLike:   false,
	}

	err = uc.swipeService.CreateSwipe(ctx, swipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create swipe: %w", err)
	}

	// Invalidate discovery cache for swiper
	uc.invalidateDiscoveryCache(ctx, req.SwiperID)

	return &DislikeUserResponse{
		Success: true,
	}, nil
}

// invalidateDiscoveryCache invalidates discovery cache for a user
func (uc *DislikeUserUseCase) invalidateDiscoveryCache(ctx context.Context, userID uuid.UUID) {
	uc.cacheService.InvalidateUserDiscoveryCache(ctx, userID)
}

// Validate validates the request
func (req *DislikeUserRequest) Validate() error {
	if req.SwiperID == uuid.Nil {
		return fmt.Errorf("swiper_id is required")
	}
	if req.SwipedID == uuid.Nil {
		return fmt.Errorf("swiped_id is required")
	}
	if req.SwiperID == req.SwipedID {
		return fmt.Errorf("cannot dislike yourself")
	}
	return nil
}