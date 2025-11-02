package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetConversationsRequest represents the request to get user conversations
type GetConversationsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Limit  int       `json:"limit" validate:"min=1,max=100"`
	Offset  int       `json:"offset" validate:"min=0"`
}

// GetConversationsResponse represents the response with user conversations
type GetConversationsResponse struct {
	Conversations []*ConversationWithUnread `json:"conversations"`
	Total        int64                    `json:"total"`
	Limit         int                       `json:"limit"`
	Offset        int                       `json:"offset"`
}

// ConversationWithUnread represents a conversation with unread count
type ConversationWithUnread struct {
	*entities.Conversation
	UnreadCount int `json:"unread_count"`
	LastMessage  *entities.Message `json:"last_message,omitempty"`
	OtherUser    *UserInfo         `json:"other_user,omitempty"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	PhotoURL  string    `json:"photo_url,omitempty"`
}

// GetConversationsUseCase retrieves user's conversations with pagination
type GetConversationsUseCase struct {
	messageRepo repositories.MessageRepository
	userRepo    repositories.UserRepository
	matchRepo    repositories.MatchRepository
}

// NewGetConversationsUseCase creates a new get conversations use case
func NewGetConversationsUseCase(
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
) *GetConversationsUseCase {
	return &GetConversationsUseCase{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		matchRepo:    matchRepo,
	}
}

// Execute retrieves user's conversations with pagination
func (uc *GetConversationsUseCase) Execute(ctx context.Context, req *GetConversationsRequest) (*GetConversationsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Get conversations with unread count
	conversationsWithUnread, err := uc.messageRepo.GetUserConversationsWithUnreadCount(ctx, req.UserID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get user conversations", err)
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Get total count
	total, err := uc.messageRepo.GetConversationCount(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get conversation count", err)
		return nil, fmt.Errorf("failed to get conversation count: %w", err)
	}

	// Enrich conversations with additional data
	response := &GetConversationsResponse{
		Conversations: make([]*ConversationWithUnread, len(conversationsWithUnread)),
		Total:        total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	for i, conv := range conversationsWithUnread {
		// Get last message
		lastMessage, err := uc.messageRepo.GetLastMessage(ctx, conv.ID)
		if err != nil {
			logger.Warn("Failed to get last message for conversation", err, "conversation_id", conv.ID)
		}

		// Get other user info
		otherUser, err := uc.getOtherUserInfo(ctx, req.UserID, conv)
		if err != nil {
			logger.Warn("Failed to get other user info for conversation", err, "conversation_id", conv.ID)
		}

		response.Conversations[i] = &ConversationWithUnread{
			Conversation: conv,
			UnreadCount:  conv.UnreadCount,
			LastMessage:  lastMessage,
			OtherUser:    otherUser,
		}
	}

	logger.Info("Retrieved user conversations", 
		"user_id", req.UserID,
		"count", len(response.Conversations),
		"total", total,
	)

	return response, nil
}

// getOtherUserInfo gets information about the other user in a conversation
func (uc *GetConversationsUseCase) getOtherUserInfo(ctx context.Context, userID uuid.UUID, conversation *entities.Conversation) (*UserInfo, error) {
	// Get match to find the other user
	match, err := uc.matchRepo.GetByID(ctx, conversation.MatchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	// Determine which user is the "other" user
	var otherUserID uuid.UUID
	if match.User1ID == userID {
		otherUserID = match.User2ID
	} else {
		otherUserID = match.User1ID
	}

	// Get other user details
	otherUser, err := uc.userRepo.GetByID(ctx, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get other user: %w", err)
	}

	// Get primary photo
	primaryPhotoURL := ""
	if len(otherUser.Photos) > 0 {
		for _, photo := range otherUser.Photos {
			if photo.IsPrimary {
				primaryPhotoURL = photo.FileURL
				break
			}
		}
	}

	return &UserInfo{
		ID:        otherUser.ID,
		FirstName: otherUser.FirstName,
		LastName:  otherUser.LastName,
		PhotoURL:  primaryPhotoURL,
	}, nil
}

// Validate validates the request
func (req *GetConversationsRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if req.Limit < 0 || req.Limit > 100 {
		return fmt.Errorf("limit must be between 0 and 100")
	}
	if req.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}