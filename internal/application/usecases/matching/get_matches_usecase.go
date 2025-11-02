package matching

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// GetMatchesUseCase handles getting user's matches
type GetMatchesUseCase struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	photoRepo    repositories.PhotoRepository
	messageRepo  repositories.MessageRepository
	cacheService CacheService
}

// NewGetMatchesUseCase creates a new GetMatchesUseCase
func NewGetMatchesUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	photoRepo repositories.PhotoRepository,
	messageRepo repositories.MessageRepository,
	cacheService CacheService,
) *GetMatchesUseCase {
	return &GetMatchesUseCase{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		photoRepo:    photoRepo,
		messageRepo:  messageRepo,
		cacheService: cacheService,
	}
}

// GetMatchesRequest represents a request to get user's matches
type GetMatchesRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Limit     int       `json:"limit" validate:"min=1,max=100"`
	Offset    int       `json:"offset" validate:"min=0"`
	UnreadOnly bool      `json:"unread_only,omitempty"`
}

// GetMatchesResponse represents the response from getting matches
type GetMatchesResponse struct {
	Matches   []*dto.MatchWithDetails `json:"matches"`
	Total     int64                 `json:"total"`
	HasMore   bool                  `json:"has_more"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

// Execute gets user's matches with pagination
func (uc *GetMatchesUseCase) Execute(ctx context.Context, req *GetMatchesRequest) (*GetMatchesResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check cache first
	cacheKey := uc.generateCacheKey(req.UserID, req.UnreadOnly, req.Limit, req.Offset)
	if cached, err := uc.cacheService.GetMatches(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Get matches with details
	matchesWithDetails, err := uc.matchRepo.GetUserMatchesWithDetails(ctx, req.UserID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	// Get total count
	total, err := uc.matchRepo.GetMatchCount(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match count: %w", err)
	}

	// Convert to DTOs
	matchDTOs := make([]*dto.MatchWithDetails, 0, len(matchesWithDetails))
	for _, matchWithDetails := range matchesWithDetails {
		// Get photos for the other user
		photos, err := uc.photoRepo.GetByUserID(ctx, matchWithDetails.OtherUser.ID)
		if err != nil {
			continue // Skip match if we can't get photos
		}

		// Create match DTO with details
		matchDTO := dto.NewMatchWithDetails(matchWithDetails, photos)
		matchDTOs = append(matchDTOs, matchDTO)
	}

	// Create response
	response := &GetMatchesResponse{
		Matches: matchDTOs,
		Total:   total,
		HasMore: int64(req.Offset+req.Limit) < total,
	}

	// Generate next cursor if there are more results
	if response.HasMore {
		response.NextCursor = fmt.Sprintf("%d", req.Offset+req.Limit)
	}

	// Cache result
	uc.cacheService.SetMatches(ctx, cacheKey, response, 5*time.Minute)

	return response, nil
}

// generateCacheKey generates a cache key for matches
func (uc *GetMatchesUseCase) generateCacheKey(userID uuid.UUID, unreadOnly bool, limit, offset int) string {
	return fmt.Sprintf("matches:%s:%t:%d:%d", userID.String(), unreadOnly, limit, offset)
}

// Validate validates the request
func (req *GetMatchesRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	return nil
}