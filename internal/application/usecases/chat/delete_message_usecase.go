package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// DeleteMessageRequest represents a request to delete a message
type DeleteMessageRequest struct {
	MessageID uuid.UUID `json:"message_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
}

// DeleteMessageResponse represents the response after deleting a message
type DeleteMessageResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DeleteMessageUseCase handles deleting a message
type DeleteMessageUseCase struct {
	messageRepo repositories.MessageRepository
}

// NewDeleteMessageUseCase creates a new delete message use case
func NewDeleteMessageUseCase(messageRepo repositories.MessageRepository) *DeleteMessageUseCase {
	return &DeleteMessageUseCase{
		messageRepo: messageRepo,
	}
}

// Execute deletes a message after validation
func (uc *DeleteMessageUseCase) Execute(ctx context.Context, req *DeleteMessageRequest) (*DeleteMessageResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return &DeleteMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Get message to verify ownership and permissions
	message, err := uc.messageRepo.GetByID(ctx, req.MessageID)
	if err != nil {
		logger.Error("Failed to get message", err)
		return &DeleteMessageResponse{
			Success: false,
			Error:   "Message not found",
		}, nil
	}

	// Check if user is the sender
	if message.SenderID != req.UserID {
		return &DeleteMessageResponse{
			Success: false,
			Error:   "User can only delete their own messages",
		}, nil
	}

	// Check if message can be deleted
	if !message.CanBeDeleted() {
		return &DeleteMessageResponse{
			Success: false,
			Error:   "Message cannot be deleted",
		}, nil
	}

	// Check if message is too old to delete (24 hours)
	if time.Since(message.CreatedAt) > 24*time.Hour {
		return &DeleteMessageResponse{
			Success: false,
			Error:   "Message can only be deleted within 24 hours of sending",
		}, nil
	}

	// Soft delete message
	if err := uc.messageRepo.SoftDeleteMessage(ctx, req.MessageID); err != nil {
		logger.Error("Failed to delete message", err)
		return &DeleteMessageResponse{
			Success: false,
			Error:   "Failed to delete message",
		}, nil
	}

	logger.Info("Message deleted successfully", 
		"message_id", req.MessageID,
		"user_id", req.UserID,
		"deleted_at", time.Now(),
	)

	return &DeleteMessageResponse{
		Success: true,
	}, nil
}

// Validate validates the request
func (req *DeleteMessageRequest) Validate() error {
	if req.MessageID == uuid.Nil {
		return fmt.Errorf("message_id is required")
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	return nil
}