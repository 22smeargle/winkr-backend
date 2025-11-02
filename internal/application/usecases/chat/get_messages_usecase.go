package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetMessagesRequest represents a request to get conversation messages
type GetMessagesRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	Limit         int       `json:"limit" validate:"min=1,max=100"`
	Offset        int       `json:"offset" validate:"min=0"`
}

// GetMessagesResponse represents the response with conversation messages
type GetMessagesResponse struct {
	Messages   []*entities.Message `json:"messages"`
	Total      int64            `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	HasMore    bool              `json:"has_more"`
}

// GetMessagesUseCase retrieves messages from a conversation
type GetMessagesUseCase struct {
	messageRepo repositories.MessageRepository
}

// NewGetMessagesUseCase creates a new get messages use case
func NewGetMessagesUseCase(messageRepo repositories.MessageRepository) *GetMessagesUseCase {
	return &GetMessagesUseCase{
		messageRepo: messageRepo,
	}
}

// Execute retrieves messages from a conversation with pagination
func (uc *GetMessagesUseCase) Execute(ctx context.Context, req *GetMessagesRequest) (*GetMessagesResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if user can access conversation
	canAccess, err := uc.messageRepo.UserCanAccessConversation(ctx, req.UserID, req.ConversationID)
	if err != nil {
		logger.Error("Failed to check conversation access", err)
		return nil, fmt.Errorf("failed to check conversation access: %w", err)
	}

	if !canAccess {
		return nil, fmt.Errorf("user cannot access conversation")
	}

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 50
	}

	// Get messages
	messages, err := uc.messageRepo.GetMessages(ctx, req.ConversationID, req.Limit, req.Offset)
	if err != nil {
		logger.Error("Failed to get messages", err)
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Get total count
	total, err := uc.messageRepo.GetConversationMessageCount(ctx, req.ConversationID)
	if err != nil {
		logger.Error("Failed to get message count", err)
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}

	// Mark messages as read for this user
	if len(messages) > 0 {
		if err := uc.messageRepo.MarkConversationAsRead(ctx, req.ConversationID, req.UserID); err != nil {
			logger.Error("Failed to mark conversation as read", err)
			// Don't fail the request, just log the error
		}
	}

	response := &GetMessagesResponse{
		Messages: messages,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
		HasMore:  int64(req.Offset+req.Limit) < total,
	}

	logger.Info("Retrieved conversation messages", 
		"conversation_id", req.ConversationID,
		"user_id", req.UserID,
		"count", len(messages),
		"total", total,
	)

	return response, nil
}

// Validate validates the request
func (req *GetMessagesRequest) Validate() error {
	if req.ConversationID == uuid.Nil {
		return fmt.Errorf("conversation_id is required")
	}
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