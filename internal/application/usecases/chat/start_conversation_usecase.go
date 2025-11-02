package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// StartConversationRequest represents a request to start a conversation
type StartConversationRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	MatchID    uuid.UUID `json:"match_id" validate:"required"`
	FirstMessage string     `json:"first_message,omitempty"`
}

// StartConversationResponse represents the response when starting a conversation
type StartConversationResponse struct {
	Conversation *entities.Conversation `json:"conversation"`
	Exists       bool                `json:"exists"`
	Message      *entities.Message    `json:"message,omitempty"`
	Error        string              `json:"error,omitempty"`
}

// StartConversationUseCase handles starting a new conversation
type StartConversationUseCase struct {
	messageRepo repositories.MessageRepository
	matchRepo    repositories.MatchRepository
}

// NewStartConversationUseCase creates a new start conversation use case
func NewStartConversationUseCase(
	messageRepo repositories.MessageRepository,
	matchRepo repositories.MatchRepository,
) *StartConversationUseCase {
	return &StartConversationUseCase{
		messageRepo: messageRepo,
		matchRepo:    matchRepo,
	}
}

// Execute starts a conversation between matched users
func (uc *StartConversationUseCase) Execute(ctx context.Context, req *StartConversationRequest) (*StartConversationResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return &StartConversationResponse{
			Error: err.Error(),
		}, nil
	}

	// Get match to verify users are matched
	match, err := uc.matchRepo.GetByID(ctx, req.MatchID)
	if err != nil {
		logger.Error("Failed to get match", err)
		return &StartConversationResponse{
			Error: "Match not found",
		}, nil
	}

	// Verify user is part of the match
	if match.User1ID != req.UserID && match.User2ID != req.UserID {
		return &StartConversationResponse{
			Error: "User is not part of this match",
		}, nil
	}

	// Check if conversation already exists
	conversation, err := uc.messageRepo.GetConversationByMatchID(ctx, req.MatchID)
	if err == nil {
		// Conversation already exists
		response := &StartConversationResponse{
			Conversation: conversation,
			Exists:       true,
		}

		// If first message is provided, send it
		if req.FirstMessage != "" {
			message, err := uc.sendFirstMessage(ctx, req, conversation.ID)
			if err != nil {
				logger.Error("Failed to send first message", err)
				response.Error = "Failed to send first message"
			} else {
				response.Message = message
			}
		}

		return response, nil
	}

	// Create new conversation
	conversation = &entities.Conversation{
		ID:        uuid.New(),
		MatchID:   req.MatchID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.messageRepo.CreateConversation(ctx, conversation); err != nil {
		logger.Error("Failed to create conversation", err)
		return &StartConversationResponse{
			Error: "Failed to create conversation",
		}, nil
	}

	response := &StartConversationResponse{
		Conversation: conversation,
		Exists:       false,
	}

	// If first message is provided, send it
	if req.FirstMessage != "" {
		message, err := uc.sendFirstMessage(ctx, req, conversation.ID)
		if err != nil {
			logger.Error("Failed to send first message", err)
			response.Error = "Failed to send first message"
		} else {
			response.Message = message
		}
	}

	logger.Info("Conversation started", 
		"conversation_id", conversation.ID,
		"match_id", req.MatchID,
		"user_id", req.UserID,
		"exists", response.Exists,
	)

	return response, nil
}

// sendFirstMessage sends the first message in a conversation
func (uc *StartConversationUseCase) sendFirstMessage(ctx context.Context, req *StartConversationRequest, conversationID uuid.UUID) (*entities.Message, error) {
	// Create first message
	message := &entities.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       req.UserID,
		Content:        req.FirstMessage,
		MessageType:    "text",
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	// Save message
	if err := uc.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create first message: %w", err)
	}

	// Update conversation activity
	conversation := &entities.Conversation{
		ID:        conversationID,
		UpdatedAt: time.Now(),
	}

	if err := uc.messageRepo.UpdateConversation(ctx, conversation); err != nil {
		logger.Error("Failed to update conversation activity", err)
		// Don't fail the request, just log the error
	}

	logger.Info("First message sent", 
		"message_id", message.ID,
		"conversation_id", conversationID,
		"user_id", req.UserID,
	)

	return message, nil
}

// Validate validates the request
func (req *StartConversationRequest) Validate() error {
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if req.MatchID == uuid.Nil {
		return fmt.Errorf("match_id is required")
	}
	if len(req.FirstMessage) > 2000 {
		return fmt.Errorf("first_message too long (max 2000 characters)")
	}
	return nil
}