package profile

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// GetMatchesUseCase handles getting user matches
type GetMatchesUseCase struct {
	userRepo     repositories.UserRepository
	matchRepo    repositories.MatchRepository
	photoRepo    repositories.PhotoRepository
	cacheService ProfileCacheService
}

// NewGetMatchesUseCase creates a new GetMatchesUseCase instance
func NewGetMatchesUseCase(
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	photoRepo repositories.PhotoRepository,
	cacheService ProfileCacheService,
) *GetMatchesUseCase {
	return &GetMatchesUseCase{
		userRepo:     userRepo,
		matchRepo:    matchRepo,
		photoRepo:    photoRepo,
		cacheService: cacheService,
	}
}

// GetMatchesResponse represents get matches response
type GetMatchesResponse struct {
	Matches []*MatchWithUser `json:"matches"`
	Total   int64            `json:"total"`
}

// MatchWithUser represents a match with user details
type MatchWithUser struct {
	ID            uuid.UUID  `json:"id"`
	User          *User      `json:"user"`
	MatchedAt     string     `json:"matched_at"`
	IsActive      bool       `json:"is_active"`
	LastMessage   *Message   `json:"last_message,omitempty"`
	UnreadCount   int        `json:"unread_count"`
	HasConversation bool       `json:"has_conversation"`
}

// User represents user in match response
type User struct {
	ID       uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	Age      int       `json:"age"`
	Photos   []*Photo  `json:"photos"`
}

// Message represents last message in match
type Message struct {
	ID        uuid.UUID `json:"id"`
	Content   string    `json:"content"`
	SenderID  uuid.UUID `json:"sender_id"`
	CreatedAt string    `json:"created_at"`
}

// Execute handles the get matches use case
func (uc *GetMatchesUseCase) Execute(ctx context.Context, userID uuid.UUID, limit, offset int) (*GetMatchesResponse, error) {
	// Try to get from cache first
	cacheKey := generateMatchesCacheKey(userID, limit, offset)
	cachedMatches, err := uc.cacheService.GetMatches(ctx, cacheKey)
	if err == nil && cachedMatches != nil {
		return cachedMatches, nil
	}

	// Get user to verify they exist and are active
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive || user.IsBanned {
		return nil, errors.ErrUserNotFound
	}

	// Get matches with details from database
	matchesWithDetails, err := uc.matchRepo.GetUserMatchesWithDetails(ctx, userID, limit, offset)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get user matches")
	}

	// Get total count of matches
	total, err := uc.matchRepo.GetMatchCount(ctx, userID)
	if err != nil {
		return nil, errors.WrapError(err, "Failed to get match count")
	}

	// Convert to response format
	response := &GetMatchesResponse{
		Matches: make([]*MatchWithUser, 0, len(matchesWithDetails)),
		Total:   total,
	}

	for _, matchDetail := range matchesWithDetails {
		// Get other user in the match
		otherUserID := matchDetail.Match.GetOtherUserID(userID)
		if otherUserID == uuid.Nil {
			continue // Skip invalid matches
		}

		// Get other user's photos
		photos, err := uc.photoRepo.GetUserPhotos(ctx, otherUserID, true)
		if err != nil {
			// Log error but continue with empty photos
			photos = []*repositories.PhotoEntity{}
		}

		// Build user response
		userResponse := &User{
			ID:        otherUserID,
			FirstName: matchDetail.OtherUser.FirstName,
			Age:       matchDetail.OtherUser.GetAge(),
			Photos:    make([]*Photo, 0, len(photos)),
		}

		// Add photos to user response
		for _, photo := range photos {
			userResponse.Photos = append(userResponse.Photos, &Photo{
				ID:        photo.ID.String(),
				URL:       photo.FileURL,
				IsPrimary: photo.IsPrimary,
				CreatedAt: photo.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Build last message response
		var lastMessage *Message
		if matchDetail.LastMessage != nil {
			lastMessage = &Message{
				ID:        matchDetail.LastMessage.ID,
				Content:   matchDetail.LastMessage.Content,
				SenderID:  matchDetail.LastMessage.SenderID,
				CreatedAt: matchDetail.LastMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		// Build match response
		matchResponse := &MatchWithUser{
			ID:            matchDetail.Match.ID,
			User:          userResponse,
			MatchedAt:     matchDetail.Match.MatchedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsActive:      matchDetail.Match.IsActive,
			LastMessage:   lastMessage,
			UnreadCount:   matchDetail.UnreadCount,
			HasConversation: matchDetail.HasConversation,
		}

		response.Matches = append(response.Matches, matchResponse)
	}

	// Cache the response
	if err := uc.cacheService.SetMatches(ctx, cacheKey, response, 10*time.Minute); err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	return response, nil
}

// generateMatchesCacheKey generates a cache key for matches
func generateMatchesCacheKey(userID uuid.UUID, limit, offset int) string {
	return "matches:" + userID.String() + ":" + string(rune(limit)) + ":" + string(rune(offset))
}