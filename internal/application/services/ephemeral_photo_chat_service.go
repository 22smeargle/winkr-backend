package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EphemeralPhotoChatService defines the interface for integrating ephemeral photos with chat
type EphemeralPhotoChatService interface {
	// Message integration
	SendEphemeralPhotoMessage(ctx context.Context, conversationID uuid.UUID, senderID uuid.UUID, photoID uuid.UUID, message string) (*ChatMessageResult, error)
	GetEphemeralPhotoMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error)
	
	// Real-time notifications
	NotifyPhotoViewed(ctx context.Context, photoID uuid.UUID, viewerID *uuid.UUID) error
	NotifyPhotoExpired(ctx context.Context, photoID uuid.UUID) error
	NotifyPhotoDeleted(ctx context.Context, photoID uuid.UUID) error
	
	// Chat integration
	GetConversationEphemeralPhotos(ctx context.Context, conversationID uuid.UUID) ([]*entities.EphemeralPhoto, error)
	MarkPhotoAsViewedInChat(ctx context.Context, messageID uuid.UUID, photoID uuid.UUID) error
	
	// Analytics
	TrackPhotoEngagement(ctx context.Context, photoID uuid.UUID, conversationID uuid.UUID, engagementType string) error
	GetPhotoEngagementStats(ctx context.Context, photoID uuid.UUID) (*EngagementStats, error)
}

// ChatMessageResult represents the result of sending a chat message
type ChatMessageResult struct {
	MessageID   uuid.UUID `json:"message_id"`
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	DeliveryTime int64     `json:"delivery_time_ms"`
}

// EngagementStats represents engagement statistics for a photo
type EngagementStats struct {
	TotalViews      int64 `json:"total_views"`
	UniqueViewers    int64 `json:"unique_viewers"`
	AverageViewTime  int64 `json:"average_view_time_seconds"`
	Shares          int64 `json:"shares"`
	Reactions       int64 `json:"reactions"`
	Comments        int64 `json:"comments"`
	EngagementRate  float64 `json:"engagement_rate"`
	FirstViewed     time.Time `json:"first_viewed"`
	LastViewed      time.Time `json:"last_viewed"`
}

// EphemeralPhotoChatServiceImpl implements EphemeralPhotoChatService
type EphemeralPhotoChatServiceImpl struct {
	messageService    MessageService
	ephemeralService EphemeralPhotoService
	cacheService     EphemeralPhotoCacheService
	websocketService  WebSocketService
}

// NewEphemeralPhotoChatService creates a new ephemeral photo chat service
func NewEphemeralPhotoChatService(
	messageService MessageService,
	ephemeralService EphemeralPhotoService,
	cacheService EphemeralPhotoCacheService,
	websocketService WebSocketService,
) EphemeralPhotoChatService {
	return &EphemeralPhotoChatServiceImpl{
		messageService:    messageService,
		ephemeralService: ephemeralService,
		cacheService:     cacheService,
		websocketService:  websocketService,
	}
}

// SendEphemeralPhotoMessage sends a message with an ephemeral photo
func (s *EphemeralPhotoChatServiceImpl) SendEphemeralPhotoMessage(ctx context.Context, conversationID uuid.UUID, senderID uuid.UUID, photoID uuid.UUID, message string) (*ChatMessageResult, error) {
	startTime := time.Now()
	
	// Get ephemeral photo details
	photo, err := s.ephemeralService.GetEphemeralPhoto(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for chat message", err)
		return nil, fmt.Errorf("failed to get ephemeral photo for chat message: %w", err)
	}
	
	// Verify photo belongs to sender
	if photo.UserID != senderID {
		return nil, fmt.Errorf("user does not own this photo")
	}
	
	// Create message content with photo reference
	messageContent := s.createPhotoMessageContent(photo, message)
	
	// Send message through message service
	messageID, err := s.messageService.SendMessage(ctx, &SendMessageRequest{
		ConversationID: conversationID,
		SenderID:      senderID,
		Content:        messageContent,
		MessageType:    "ephemeral_photo",
	})
	if err != nil {
		logger.Error("Failed to send ephemeral photo message", err)
		return nil, fmt.Errorf("failed to send ephemeral photo message: %w", err)
	}
	
	deliveryTime := time.Since(startTime).Milliseconds()
	
	result := &ChatMessageResult{
		MessageID:    messageID,
		Success:       true,
		Message:       "Ephemeral photo message sent successfully",
		Timestamp:     time.Now(),
		DeliveryTime:  deliveryTime,
	}
	
	logger.Info("Ephemeral photo message sent successfully", map[string]interface{}{
		"message_id":     messageID,
		"conversation_id": conversationID,
		"photo_id":       photoID,
		"sender_id":      senderID,
		"delivery_time":  deliveryTime,
	})
	
	return result, nil
}

// GetEphemeralPhotoMessage gets a message containing an ephemeral photo
func (s *EphemeralPhotoChatServiceImpl) GetEphemeralPhotoMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error) {
	// Get message from message service
	message, err := s.messageService.GetMessage(ctx, messageID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo message", err)
		return nil, fmt.Errorf("failed to get ephemeral photo message: %w", err)
	}
	
	// Verify message type
	if message.MessageType != "ephemeral_photo" {
		return nil, fmt.Errorf("message is not an ephemeral photo message")
	}
	
	return message, nil
}

// NotifyPhotoViewed sends a notification when a photo is viewed
func (s *EphemeralPhotoChatServiceImpl) NotifyPhotoViewed(ctx context.Context, photoID uuid.UUID, viewerID *uuid.UUID) error {
	// Get photo details
	photo, err := s.ephemeralService.GetEphemeralPhoto(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for view notification", err)
		return fmt.Errorf("failed to get ephemeral photo for view notification: %w", err)
	}
	
	// Create notification message
	notificationMessage := s.createViewNotificationMessage(photo, viewerID)
	
	// Send notification through WebSocket
	if s.websocketService != nil {
		notification := map[string]interface{}{
			"type":       "ephemeral_photo_viewed",
			"photo_id":   photoID,
			"viewer_id":  viewerID,
			"owner_id":   photo.UserID,
			"viewed_at":  time.Now(),
		}
		
		if err := s.websocketService.BroadcastToUser(photo.UserID, notification); err != nil {
			logger.Error("Failed to send photo viewed notification", err)
			return fmt.Errorf("failed to send photo viewed notification: %w", err)
		}
	}
	
	logger.Info("Photo viewed notification sent", map[string]interface{}{
		"photo_id":  photoID,
		"viewer_id": viewerID,
		"owner_id":  photo.UserID,
	})
	
	return nil
}

// NotifyPhotoExpired sends a notification when a photo expires
func (s *EphemeralPhotoChatServiceImpl) NotifyPhotoExpired(ctx context.Context, photoID uuid.UUID) error {
	// Get photo details
	photo, err := s.ephemeralService.GetEphemeralPhoto(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for expiration notification", err)
		return fmt.Errorf("failed to get ephemeral photo for expiration notification: %w", err)
	}
	
	// Create notification message
	notificationMessage := s.createExpirationNotificationMessage(photo)
	
	// Send notification through WebSocket
	if s.websocketService != nil {
		notification := map[string]interface{}{
			"type":       "ephemeral_photo_expired",
			"photo_id":   photoID,
			"owner_id":   photo.UserID,
			"expired_at":  time.Now(),
		}
		
		if err := s.websocketService.BroadcastToUser(photo.UserID, notification); err != nil {
			logger.Error("Failed to send photo expired notification", err)
			return fmt.Errorf("failed to send photo expired notification: %w", err)
		}
	}
	
	logger.Info("Photo expired notification sent", map[string]interface{}{
		"photo_id": photoID,
		"owner_id": photo.UserID,
	})
	
	return nil
}

// NotifyPhotoDeleted sends a notification when a photo is deleted
func (s *EphemeralPhotoChatServiceImpl) NotifyPhotoDeleted(ctx context.Context, photoID uuid.UUID) error {
	// Get photo details
	photo, err := s.ephemeralService.GetEphemeralPhoto(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for deletion notification", err)
		return fmt.Errorf("failed to get ephemeral photo for deletion notification: %w", err)
	}
	
	// Create notification message
	notificationMessage := s.createDeletionNotificationMessage(photo)
	
	// Send notification through WebSocket
	if s.websocketService != nil {
		notification := map[string]interface{}{
			"type":       "ephemeral_photo_deleted",
			"photo_id":   photoID,
			"owner_id":   photo.UserID,
			"deleted_at":  time.Now(),
		}
		
		if err := s.websocketService.BroadcastToUser(photo.UserID, notification); err != nil {
			logger.Error("Failed to send photo deleted notification", err)
			return fmt.Errorf("failed to send photo deleted notification: %w", err)
		}
	}
	
	logger.Info("Photo deleted notification sent", map[string]interface{}{
		"photo_id": photoID,
		"owner_id": photo.UserID,
	})
	
	return nil
}

// GetConversationEphemeralPhotos gets all ephemeral photos in a conversation
func (s *EphemeralPhotoChatServiceImpl) GetConversationEphemeralPhotos(ctx context.Context, conversationID uuid.UUID) ([]*entities.EphemeralPhoto, error) {
	// This would typically query for all ephemeral photos shared in a conversation
	// For now, return empty slice
	photos := make([]*entities.EphemeralPhoto, 0)
	
	logger.Debug("Retrieved conversation ephemeral photos", map[string]interface{}{
		"conversation_id": conversationID,
		"photo_count":    len(photos),
	})
	
	return photos, nil
}

// MarkPhotoAsViewedInChat marks a photo as viewed within chat context
func (s *EphemeralPhotoChatServiceImpl) MarkPhotoAsViewedInChat(ctx context.Context, messageID uuid.UUID, photoID uuid.UUID) error {
	// Get message
	message, err := s.GetEphemeralPhotoMessage(ctx, messageID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo message for view marking", err)
		return fmt.Errorf("failed to get ephemeral photo message for view marking: %w", err)
	}
	
	// Mark photo as viewed
	if err := s.ephemeralService.TrackPhotoView(ctx, photoID, message.SenderID, &message.SenderID, 30); err != nil {
		logger.Error("Failed to mark photo as viewed in chat", err)
		return fmt.Errorf("failed to mark photo as viewed in chat: %w", err)
	}
	
	// Invalidate cache
	if err := s.cacheService.InvalidateEphemeralPhoto(ctx, photoID); err != nil {
		logger.Error("Failed to invalidate photo cache", err)
		return fmt.Errorf("failed to invalidate photo cache: %w", err)
	}
	
	logger.Info("Photo marked as viewed in chat", map[string]interface{}{
		"photo_id":  photoID,
		"message_id": messageID,
	})
	
	return nil
}

// TrackPhotoEngagement tracks engagement with a photo
func (s *EphemeralPhotoChatServiceImpl) TrackPhotoEngagement(ctx context.Context, photoID uuid.UUID, conversationID uuid.UUID, engagementType string) error {
	// This would typically track different types of engagement
	// For now, just log the engagement
	logger.Info("Photo engagement tracked", map[string]interface{}{
		"photo_id":       photoID,
		"conversation_id": conversationID,
		"engagement_type": engagementType,
		"timestamp":       time.Now(),
	})
	
	return nil
}

// GetPhotoEngagementStats gets engagement statistics for a photo
func (s *EphemeralPhotoChatServiceImpl) GetPhotoEngagementStats(ctx context.Context, photoID uuid.UUID) (*EngagementStats, error) {
	// This would typically aggregate engagement data
	// For now, return mock stats
	stats := &EngagementStats{
		TotalViews:      0,
		UniqueViewers:    0,
		AverageViewTime:  0,
		Shares:          0,
		Reactions:       0,
		Comments:        0,
		EngagementRate:  0.0,
		FirstViewed:     time.Now(),
		LastViewed:      time.Now(),
	}
	
	logger.Debug("Retrieved photo engagement stats", map[string]interface{}{
		"photo_id": photoID,
	})
	
	return stats, nil
}

// Helper methods

// createPhotoMessageContent creates message content for ephemeral photo
func (s *EphemeralPhotoChatServiceImpl) createPhotoMessageContent(photo *entities.EphemeralPhoto, message string) string {
	return fmt.Sprintf(`{
		"type": "ephemeral_photo",
		"photo_id": "%s",
		"access_key": "%s",
		"thumbnail_url": "%s",
		"expires_at": "%s",
		"message": "%s"
	}`, photo.ID.String(), photo.AccessKey, photo.ThumbnailURL, photo.ExpiresAt.Format(time.RFC3339), message)
}

// createViewNotificationMessage creates a view notification message
func (s *EphemeralPhotoChatServiceImpl) createViewNotificationMessage(photo *entities.EphemeralPhoto, viewerID *uuid.UUID) string {
	viewerStr := "anonymous"
	if viewerID != nil {
		viewerStr = viewerID.String()
	}
	
	return fmt.Sprintf("Your ephemeral photo was viewed by %s at %s", viewerStr, time.Now().Format(time.RFC3339))
}

// createExpirationNotificationMessage creates an expiration notification message
func (s *EphemeralPhotoChatServiceImpl) createExpirationNotificationMessage(photo *entities.EphemeralPhoto) string {
	return fmt.Sprintf("Your ephemeral photo expired at %s", time.Now().Format(time.RFC3339))
}

// createDeletionNotificationMessage creates a deletion notification message
func (s *EphemeralPhotoChatServiceImpl) createDeletionNotificationMessage(photo *entities.EphemeralPhoto) string {
	return fmt.Sprintf("Your ephemeral photo was deleted at %s", time.Now().Format(time.RFC3339))
}