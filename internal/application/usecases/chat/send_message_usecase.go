package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	SenderID       uuid.UUID `json:"sender_id" validate:"required"`
	Content        string     `json:"content" validate:"required,max=2000"`
	MessageType    string     `json:"message_type" validate:"required,oneof=text image photo_ephemeral location system gift"`
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Message *services.ProcessedMessage `json:"message"`
	Success bool                     `json:"success"`
	Error   string                    `json:"error,omitempty"`
}

// SendMessageUseCase handles sending a message
type SendMessageUseCase struct {
	messageRepo   repositories.MessageRepository
	userRepo      repositories.UserRepository
	matchRepo     repositories.MatchRepository
	messageService *services.MessageService
}

// NewSendMessageUseCase creates a new send message use case
func NewSendMessageUseCase(
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
	matchRepo repositories.MatchRepository,
	messageService *services.MessageService,
) *SendMessageUseCase {
	return &SendMessageUseCase{
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		matchRepo:     matchRepo,
		messageService: messageService,
	}
}

// Execute sends a message after validation and processing
func (uc *SendMessageUseCase) Execute(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return &SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Check if user can access conversation
	canAccess, err := uc.messageRepo.UserCanAccessConversation(ctx, req.SenderID, req.ConversationID)
	if err != nil {
		logger.Error("Failed to check conversation access", err)
		return &SendMessageResponse{
			Success: false,
			Error:   "Failed to check conversation access",
		}, nil
	}

	if !canAccess {
		return &SendMessageResponse{
			Success: false,
			Error:   "User cannot access this conversation",
		}, nil
	}

	// Validate message content
	validationResult, err := uc.messageService.ValidateMessage(ctx, req.Content, req.MessageType, req.SenderID.String())
	if err != nil {
		logger.Error("Message validation failed", err)
		return &SendMessageResponse{
			Success: false,
			Error:   "Message validation failed",
		}, nil
	}

	if !validationResult.IsValid {
		return &SendMessageResponse{
			Success: false,
			Error:   fmt.Sprintf("Message validation failed: %v", validationResult.Errors),
		}, nil
	}

	// Create message entity
	message := &entities.Message{
		ID:             uuid.New(),
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content:        validationResult.Sanitized,
		MessageType:    req.MessageType,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	// Process message with options
	options := &services.MessageProcessingOptions{
		EnableContentFilter: true,
		EnableLinkPreview:  true,
		EnableEncryption:   false, // Could be enabled based on user settings
		EnableTranslation:  false, // Could be enabled based on user settings
	}

	processedMessage, err := uc.messageService.ProcessMessage(ctx, message, options)
	if err != nil {
		logger.Error("Message processing failed", err)
		return &SendMessageResponse{
			Success: false,
			Error:   "Message processing failed",
		}, nil
	}

	// Save message to database
	if err := uc.messageRepo.Create(ctx, processedMessage.Message); err != nil {
		logger.Error("Failed to save message", err)
		return &SendMessageResponse{
			Success: false,
			Error:   "Failed to save message",
		}, nil
	}

	// Update conversation activity
	if err := uc.updateConversationActivity(ctx, processedMessage.ConversationID); err != nil {
		logger.Error("Failed to update conversation activity", err)
		// Don't fail the request, just log the error
	}

	// Update match activity if this is the first message
	if err := uc.updateMatchActivity(ctx, processedMessage.ConversationID, req.SenderID); err != nil {
		logger.Error("Failed to update match activity", err)
		// Don't fail the request, just log the error
	}

	logger.Info("Message sent successfully", 
		"message_id", processedMessage.ID,
		"conversation_id", processedMessage.ConversationID,
		"sender_id", processedMessage.SenderID,
		"message_type", processedMessage.MessageType,
	)

	return &SendMessageResponse{
		Message: processedMessage,
		Success: true,
	}, nil
}

// updateConversationActivity updates the conversation's last activity
func (uc *SendMessageUseCase) updateConversationActivity(ctx context.Context, conversationID uuid.UUID) error {
	// Get conversation
	conversation, err := uc.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Update timestamp
	conversation.UpdatedAt = time.Now()

	// Save conversation
	if err := uc.messageRepo.UpdateConversation(ctx, conversation); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// updateMatchActivity updates the match's last activity
func (uc *SendMessageUseCase) updateMatchActivity(ctx context.Context, conversationID, senderID uuid.UUID) error {
	// Get conversation to find match
	conversation, err := uc.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Get match
	match, err := uc.matchRepo.GetByID(ctx, conversation.MatchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Update last activity
	now := time.Now()
	if match.User1ID == senderID {
		match.User1LastActivity = &now
	} else {
		match.User2LastActivity = &now
	}

	// Save match
	if err := uc.matchRepo.Update(ctx, match); err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	return nil
}

// Validate validates the request
func (req *SendMessageRequest) Validate() error {
	if req.ConversationID == uuid.Nil {
		return fmt.Errorf("conversation_id is required")
	}
	if req.SenderID == uuid.Nil {
		return fmt.Errorf("sender_id is required")
	}
	if req.Content == "" {
		return fmt.Errorf("content is required")
	}
	if len(req.Content) > 2000 {
		return fmt.Errorf("content too long (max 2000 characters)")
	}
	
	validTypes := []string{"text", "image", "photo_ephemeral", "location", "system", "gift"}
	isValidType := false
	for _, validType := range validTypes {
		if req.MessageType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid message_type: %s", req.MessageType)
	}
	
	return nil
}