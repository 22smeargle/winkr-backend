package chat

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MarkMessagesReadRequest represents a request to mark messages as read
type MarkMessagesReadRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	MessageIDs    []uuid.UUID `json:"message_ids,omitempty"`
}

// MarkMessagesReadResponse represents the response after marking messages as read
type MarkMessagesReadResponse struct {
	Success      bool `json:"success"`
	MarkedCount  int  `json:"marked_count"`
	Error        string `json:"error,omitempty"`
}

// MarkMessagesReadUseCase handles marking messages as read
type MarkMessagesReadUseCase struct {
	messageRepo repositories.MessageRepository
}

// NewMarkMessagesReadUseCase creates a new mark messages read use case
func NewMarkMessagesReadUseCase(messageRepo repositories.MessageRepository) *MarkMessagesReadUseCase {
	return &MarkMessagesReadUseCase{
		messageRepo: messageRepo,
	}
}

// Execute marks messages as read
func (uc *MarkMessagesReadUseCase) Execute(ctx context.Context, req *MarkMessagesReadRequest) (*MarkMessagesReadResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return &MarkMessagesReadResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Check if user can access conversation
	canAccess, err := uc.messageRepo.UserCanAccessConversation(ctx, req.UserID, req.ConversationID)
	if err != nil {
		logger.Error("Failed to check conversation access", err)
		return &MarkMessagesReadResponse{
			Success: false,
			Error:   "Failed to check conversation access",
		}, nil
	}

	if !canAccess {
		return &MarkMessagesReadResponse{
			Success: false,
			Error:   "User cannot access this conversation",
		}, nil
	}

	var markedCount int

	// If specific message IDs are provided, mark only those
	if len(req.MessageIDs) > 0 {
		// Verify user can access all messages
		for _, messageID := range req.MessageIDs {
			canAccess, err := uc.messageRepo.UserCanAccessMessage(ctx, req.UserID, messageID)
			if err != nil {
				logger.Error("Failed to check message access", err)
				continue
			}
			if !canAccess {
				logger.Warn("User cannot access message", "message_id", messageID, "user_id", req.UserID)
				continue
			}

			// Mark individual message as read
			if err := uc.messageRepo.MarkAsRead(ctx, messageID); err != nil {
				logger.Error("Failed to mark message as read", err, "message_id", messageID)
				continue
			}
			markedCount++
		}
	} else {
		// Mark all unread messages in conversation as read
		if err := uc.messageRepo.MarkConversationAsRead(ctx, req.ConversationID, req.UserID); err != nil {
			logger.Error("Failed to mark conversation as read", err)
			return &MarkMessagesReadResponse{
				Success: false,
				Error:   "Failed to mark conversation as read",
			}, nil
		}

		// Get count of unread messages that were marked
		unreadCount, err := uc.messageRepo.GetConversationUnreadCount(ctx, req.ConversationID, req.UserID)
		if err != nil {
			logger.Error("Failed to get unread count", err)
			markedCount = 0
		} else {
			markedCount = int(unreadCount)
		}
	}

	logger.Info("Messages marked as read", 
		"conversation_id", req.ConversationID,
		"user_id", req.UserID,
		"marked_count", markedCount,
	)

	return &MarkMessagesReadResponse{
		Success:     true,
		MarkedCount: markedCount,
	}, nil
}

// Validate validates the request
func (req *MarkMessagesReadRequest) Validate() error {
	if req.ConversationID == uuid.Nil {
		return fmt.Errorf("conversation_id is required")
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	
	// Validate message IDs if provided
	for i, messageID := range req.MessageIDs {
		if messageID == uuid.Nil {
			return fmt.Errorf("message_id at index %d is invalid", i)
		}
	}
	
	return nil
}