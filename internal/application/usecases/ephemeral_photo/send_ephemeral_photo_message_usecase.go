package ephemeral_photo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// SendEphemeralPhotoMessageUseCase handles sending ephemeral photo messages in chat
type SendEphemeralPhotoMessageUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	chatService          services.EphemeralPhotoChatService
	messageService       services.MessageService
}

// NewSendEphemeralPhotoMessageUseCase creates a new use case for sending ephemeral photo messages
func NewSendEphemeralPhotoMessageUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	chatService services.EphemeralPhotoChatService,
	messageService services.MessageService,
) *SendEphemeralPhotoMessageUseCase {
	return &SendEphemeralPhotoMessageUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		chatService:          chatService,
		messageService:       messageService,
	}
}

// SendEphemeralPhotoMessageRequest represents the request to send an ephemeral photo message
type SendEphemeralPhotoMessageRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
	SenderID       uuid.UUID `json:"sender_id" validate:"required"`
	PhotoID        uuid.UUID `json:"photo_id" validate:"required"`
	Message        string    `json:"message" validate:"max=500"`
}

// SendEphemeralPhotoMessageResponse represents the response after sending an ephemeral photo message
type SendEphemeralPhotoMessageResponse struct {
	MessageID     uuid.UUID `json:"message_id"`
	PhotoID       uuid.UUID `json:"photo_id"`
	AccessKey     string    `json:"access_key"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	ExpiresAt     time.Time `json:"expires_at"`
	Success       bool      `json:"success"`
	Message       string    `json:"message"`
	DeliveryTime  int64     `json:"delivery_time_ms"`
}

// Execute sends an ephemeral photo message in a conversation
func (uc *SendEphemeralPhotoMessageUseCase) Execute(ctx context.Context, req *SendEphemeralPhotoMessageRequest) (*SendEphemeralPhotoMessageResponse, error) {
	startTime := time.Now()
	
	// Validate request
	if err := uc.validateRequest(ctx, req); err != nil {
		logger.Error("Invalid request for sending ephemeral photo message", err)
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Get ephemeral photo details
	photo, err := uc.ephemeralPhotoService.GetEphemeralPhoto(ctx, req.PhotoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for message", err)
		return nil, fmt.Errorf("failed to get ephemeral photo: %w", err)
	}
	
	// Verify photo belongs to sender
	if photo.UserID != req.SenderID {
		return nil, fmt.Errorf("user does not own this photo")
	}
	
	// Verify photo is still valid
	if photo.IsExpired() {
		return nil, fmt.Errorf("photo has expired")
	}
	
	if photo.IsViewed {
		return nil, fmt.Errorf("photo has already been viewed")
	}
	
	// Validate message content
	validationResult, err := uc.messageService.ValidateMessage(ctx, req.Message, "ephemeral_photo", req.SenderID.String())
	if err != nil {
		logger.Error("Failed to validate ephemeral photo message", err)
		return nil, fmt.Errorf("failed to validate message: %w", err)
	}
	
	if !validationResult.IsValid {
		return nil, fmt.Errorf("message validation failed: %v", validationResult.Errors)
	}
	
	// Send ephemeral photo message through chat service
	result, err := uc.chatService.SendEphemeralPhotoMessage(ctx, req.ConversationID, req.SenderID, req.PhotoID, validationResult.Sanitized)
	if err != nil {
		logger.Error("Failed to send ephemeral photo message", err)
		return nil, fmt.Errorf("failed to send ephemeral photo message: %w", err)
	}
	
	// Track engagement
	if err := uc.chatService.TrackPhotoEngagement(ctx, req.PhotoID, req.ConversationID, "message_sent"); err != nil {
		logger.Warn("Failed to track photo engagement", err)
	}
	
	deliveryTime := time.Since(startTime).Milliseconds()
	
	response := &SendEphemeralPhotoMessageResponse{
		MessageID:    result.MessageID,
		PhotoID:      req.PhotoID,
		AccessKey:    photo.AccessKey,
		ThumbnailURL: photo.ThumbnailURL,
		ExpiresAt:    photo.ExpiresAt,
		Success:      result.Success,
		Message:      result.Message,
		DeliveryTime: deliveryTime,
	}
	
	logger.Info("Ephemeral photo message sent successfully", map[string]interface{}{
		"message_id":     result.MessageID,
		"conversation_id": req.ConversationID,
		"photo_id":       req.PhotoID,
		"sender_id":      req.SenderID,
		"delivery_time":  deliveryTime,
	})
	
	return response, nil
}

// validateRequest validates the request
func (uc *SendEphemeralPhotoMessageUseCase) validateRequest(ctx context.Context, req *SendEphemeralPhotoMessageRequest) error {
	if req.ConversationID == uuid.Nil {
		return fmt.Errorf("conversation_id is required")
	}
	
	if req.SenderID == uuid.Nil {
		return fmt.Errorf("sender_id is required")
	}
	
	if req.PhotoID == uuid.Nil {
		return fmt.Errorf("photo_id is required")
	}
	
	return nil
}

// GetEphemeralPhotoMessageUseCase handles getting ephemeral photo messages
type GetEphemeralPhotoMessageUseCase struct {
	chatService services.EphemeralPhotoChatService
}

// NewGetEphemeralPhotoMessageUseCase creates a new use case for getting ephemeral photo messages
func NewGetEphemeralPhotoMessageUseCase(chatService services.EphemeralPhotoChatService) *GetEphemeralPhotoMessageUseCase {
	return &GetEphemeralPhotoMessageUseCase{
		chatService: chatService,
	}
}

// GetEphemeralPhotoMessageRequest represents the request to get an ephemeral photo message
type GetEphemeralPhotoMessageRequest struct {
	MessageID uuid.UUID `json:"message_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
}

// GetEphemeralPhotoMessageResponse represents the response after getting an ephemeral photo message
type GetEphemeralPhotoMessageResponse struct {
	Message       *entities.Message `json:"message"`
	PhotoID       uuid.UUID         `json:"photo_id"`
	AccessKey     string            `json:"access_key"`
	ThumbnailURL  string            `json:"thumbnail_url"`
	ExpiresAt     time.Time         `json:"expires_at"`
	IsViewed      bool              `json:"is_viewed"`
	ViewCount     int               `json:"view_count"`
	CanView       bool              `json:"can_view"`
}

// Execute gets an ephemeral photo message
func (uc *GetEphemeralPhotoMessageUseCase) Execute(ctx context.Context, req *GetEphemeralPhotoMessageRequest) (*GetEphemeralPhotoMessageResponse, error) {
	// Validate request
	if req.MessageID == uuid.Nil {
		return nil, fmt.Errorf("message_id is required")
	}
	
	if req.UserID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	
	// Get message
	message, err := uc.chatService.GetEphemeralPhotoMessage(ctx, req.MessageID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo message", err)
		return nil, fmt.Errorf("failed to get ephemeral photo message: %w", err)
	}
	
	// Parse photo information from message content
	photoInfo, err := uc.parsePhotoInfoFromMessage(message.Content)
	if err != nil {
		logger.Error("Failed to parse photo info from message", err)
		return nil, fmt.Errorf("failed to parse photo info: %w", err)
	}
	
	// Check if user can view this message (is participant in conversation)
	canView := uc.canUserViewMessage(ctx, req.UserID, message)
	
	response := &GetEphemeralPhotoMessageResponse{
		Message:      message,
		PhotoID:      photoInfo.PhotoID,
		AccessKey:    photoInfo.AccessKey,
		ThumbnailURL: photoInfo.ThumbnailURL,
		ExpiresAt:    photoInfo.ExpiresAt,
		IsViewed:     photoInfo.IsViewed,
		ViewCount:    photoInfo.ViewCount,
		CanView:      canView,
	}
	
	logger.Debug("Retrieved ephemeral photo message", map[string]interface{}{
		"message_id": req.MessageID,
		"user_id":    req.UserID,
		"can_view":   canView,
	})
	
	return response, nil
}

// PhotoInfo represents parsed photo information from message content
type PhotoInfo struct {
	PhotoID      uuid.UUID `json:"photo_id"`
	AccessKey    string    `json:"access_key"`
	ThumbnailURL string    `json:"thumbnail_url"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsViewed     bool      `json:"is_viewed"`
	ViewCount    int       `json:"view_count"`
}

// parsePhotoInfoFromMessage parses photo information from message content
func (uc *GetEphemeralPhotoMessageUseCase) parsePhotoInfoFromMessage(content string) (*PhotoInfo, error) {
	// This would parse the JSON content from the message
	// For now, return mock data
	photoInfo := &PhotoInfo{
		PhotoID:      uuid.New(),
		AccessKey:    "mock_access_key",
		ThumbnailURL: "https://example.com/thumbnail.jpg",
		ExpiresAt:    time.Now().Add(30 * time.Second),
		IsViewed:     false,
		ViewCount:    0,
	}
	
	return photoInfo, nil
}

// canUserViewMessage checks if user can view the message
func (uc *GetEphemeralPhotoMessageUseCase) canUserViewMessage(ctx context.Context, userID uuid.UUID, message *entities.Message) bool {
	// User can view if they are the sender or recipient
	// This would typically check conversation participants
	return message.SenderID == userID
}