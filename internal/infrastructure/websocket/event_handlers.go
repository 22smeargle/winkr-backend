package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EventHandler handles WebSocket events for chat functionality
type EventHandler struct {
	connManager   *ConnectionManager
	messageRepo   repositories.MessageRepository
	userRepo      repositories.UserRepository
	messageService *services.MessageService
	cache         *cache.CacheService
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	connManager *ConnectionManager,
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
	messageService *services.MessageService,
	cache *cache.CacheService,
) *EventHandler {
	return &EventHandler{
		connManager:   connManager,
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		messageService: messageService,
		cache:         cache,
	}
}

// HandleMessage handles incoming WebSocket messages
func (h *EventHandler) HandleMessage(ctx context.Context, conn *ClientConnection, rawMessage []byte) error {
	// Parse message
	var wsMessage Message
	if err := json.Unmarshal(rawMessage, &wsMessage); err != nil {
		logger.Error("Failed to parse WebSocket message", err)
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Handle different message types
	switch wsMessage.Type {
	case "message:new":
		return h.handleNewMessage(ctx, conn, wsMessage)
	case "message:read":
		return h.handleMessageRead(ctx, conn, wsMessage)
	case "message:delete":
		return h.handleMessageDelete(ctx, conn, wsMessage)
	case "ephemeral_photo:new":
		return h.handleEphemeralPhotoNew(ctx, conn, wsMessage)
	case "ephemeral_photo:viewed":
		return h.handleEphemeralPhotoViewed(ctx, conn, wsMessage)
	case "ephemeral_photo:expired":
		return h.handleEphemeralPhotoExpired(ctx, conn, wsMessage)
	case "typing:start":
		return h.handleTypingStart(ctx, conn, wsMessage)
	case "typing:stop":
		return h.handleTypingStop(ctx, conn, wsMessage)
	case "conversation:join":
		return h.handleConversationJoin(ctx, conn, wsMessage)
	case "conversation:leave":
		return h.handleConversationLeave(ctx, conn, wsMessage)
	case "user:status":
		return h.handleUserStatus(ctx, conn, wsMessage)
	case "ping":
		return h.handlePing(ctx, conn, wsMessage)
	default:
		logger.Warn("Unknown message type", "type", wsMessage.Type)
		return fmt.Errorf("unknown message type: %s", wsMessage.Type)
	}
}

// handleNewMessage handles new message events
func (h *EventHandler) handleNewMessage(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract message data
	var messageData struct {
		ConversationID string `json:"conversation_id"`
		Content        string `json:"content"`
		MessageType    string `json:"message_type"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &messageData); err != nil {
		return fmt.Errorf("failed to parse message data: %w", err)
	}

	// Validate message
	validationResult, err := h.messageService.ValidateMessage(ctx, messageData.Content, messageData.MessageType, conn.UserID)
	if err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	if !validationResult.IsValid {
		// Send validation error to sender
		errorMessage := Message{
			Type: "error",
			Data: map[string]interface{}{
				"code":    "validation_failed",
				"message": "Message validation failed",
				"errors":  validationResult.Errors,
			},
			Timestamp: time.Now(),
		}
		return conn.WriteMessage(errorMessage)
	}

	// Create message entity
	message := &entities.Message{
		ID:             uuid.New(),
		ConversationID: uuid.MustParse(messageData.ConversationID),
		SenderID:       uuid.MustParse(conn.UserID),
		Content:        validationResult.Sanitized,
		MessageType:    messageData.MessageType,
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

	processedMessage, err := h.messageService.ProcessMessage(ctx, message, options)
	if err != nil {
		return fmt.Errorf("message processing failed: %w", err)
	}

	// Save message to database
	if err := h.messageRepo.Create(ctx, processedMessage.Message); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// Update conversation activity
	if err := h.updateConversationActivity(ctx, processedMessage.ConversationID); err != nil {
		logger.Error("Failed to update conversation activity", err)
	}

	// Broadcast message to conversation participants
	broadcastMessage := Message{
		Type:      "message:new",
		Data:      processedMessage,
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	if err := h.connManager.BroadcastToConversation(messageData.ConversationID, broadcastMessage); err != nil {
		logger.Error("Failed to broadcast message", err)
	}

	// Update unread counts for recipients
	if err := h.updateUnreadCounts(ctx, processedMessage); err != nil {
		logger.Error("Failed to update unread counts", err)
	}

	// Cache message for quick access
	if err := h.cacheMessage(ctx, processedMessage); err != nil {
		logger.Error("Failed to cache message", err)
	}

	logger.Info("New message processed", 
		"message_id", processedMessage.ID,
		"conversation_id", processedMessage.ConversationID,
		"sender_id", processedMessage.SenderID,
	)

	return nil
}

// handleMessageRead handles message read events
func (h *EventHandler) handleMessageRead(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract message data
	var messageData struct {
		MessageID string `json:"message_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &messageData); err != nil {
		return fmt.Errorf("failed to parse message data: %w", err)
	}

	messageUUID := uuid.MustParse(messageData.MessageID)

	// Check if user can access message
	canAccess, err := h.messageRepo.UserCanAccessMessage(ctx, uuid.MustParse(conn.UserID), messageUUID)
	if err != nil {
		return fmt.Errorf("failed to check message access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("user cannot access message")
	}

	// Mark message as read
	if err := h.messageRepo.MarkAsRead(ctx, messageUUID); err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	// Broadcast read receipt to conversation
	readMessage := Message{
		Type: "message:viewed",
		Data: map[string]interface{}{
			"message_id": messageData.MessageID,
			"user_id":    conn.UserID,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	// Get message to find conversation
	message, err := h.messageRepo.GetByID(ctx, messageUUID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if err := h.connManager.BroadcastToConversation(message.ConversationID.String(), readMessage); err != nil {
		logger.Error("Failed to broadcast read receipt", err)
	}

	// Update unread count for user
	if err := h.updateUserUnreadCount(ctx, uuid.MustParse(conn.UserID), message.ConversationID); err != nil {
		logger.Error("Failed to update user unread count", err)
	}

	logger.Info("Message marked as read", 
		"message_id", messageUUID,
		"user_id", conn.UserID,
	)

	return nil
}

// handleMessageDelete handles message delete events
func (h *EventHandler) handleMessageDelete(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract message data
	var messageData struct {
		MessageID string `json:"message_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &messageData); err != nil {
		return fmt.Errorf("failed to parse message data: %w", err)
	}

	messageUUID := uuid.MustParse(messageData.MessageID)

	// Get message to verify ownership
	message, err := h.messageRepo.GetByID(ctx, messageUUID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Check if user is the sender
	if message.SenderID.String() != conn.UserID {
		return fmt.Errorf("user can only delete their own messages")
	}

	// Check if message can be deleted
	if !message.CanBeDeleted() {
		return fmt.Errorf("message cannot be deleted")
	}

	// Soft delete message
	if err := h.messageRepo.SoftDeleteMessage(ctx, messageUUID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Broadcast delete event to conversation
	deleteMessage := Message{
		Type: "message:deleted",
		Data: map[string]interface{}{
			"message_id": messageData.MessageID,
			"user_id":    conn.UserID,
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	if err := h.connManager.BroadcastToConversation(message.ConversationID.String(), deleteMessage); err != nil {
		logger.Error("Failed to broadcast delete event", err)
	}

	// Remove from cache
	if err := h.cache.Delete(ctx, fmt.Sprintf("message:%s", messageUUID)); err != nil {
		logger.Error("Failed to delete message from cache", err)
	}

	logger.Info("Message deleted", 
		"message_id", messageUUID,
		"user_id", conn.UserID,
	)

	return nil
}

// handleTypingStart handles typing start events
func (h *EventHandler) handleTypingStart(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract typing data
	var typingData struct {
		ConversationID string `json:"conversation_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &typingData); err != nil {
		return fmt.Errorf("failed to parse typing data: %w", err)
	}

	// Set typing indicator
	if err := h.connManager.SetTyping(conn.UserID, typingData.ConversationID, true); err != nil {
		return fmt.Errorf("failed to set typing indicator: %w", err)
	}

	// Cache typing indicator
	typingKey := fmt.Sprintf("typing:%s:%s", typingData.ConversationID, conn.UserID)
	if err := h.cache.Set(ctx, typingKey, true, 5*time.Second); err != nil {
		logger.Error("Failed to cache typing indicator", err)
	}

	return nil
}

// handleTypingStop handles typing stop events
func (h *EventHandler) handleTypingStop(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract typing data
	var typingData struct {
		ConversationID string `json:"conversation_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &typingData); err != nil {
		return fmt.Errorf("failed to parse typing data: %w", err)
	}

	// Clear typing indicator
	if err := h.connManager.SetTyping(conn.UserID, typingData.ConversationID, false); err != nil {
		return fmt.Errorf("failed to clear typing indicator: %w", err)
	}

	// Remove from cache
	typingKey := fmt.Sprintf("typing:%s:%s", typingData.ConversationID, conn.UserID)
	if err := h.cache.Delete(ctx, typingKey); err != nil {
		logger.Error("Failed to delete typing indicator from cache", err)
	}

	return nil
}

// handleConversationJoin handles conversation join events
func (h *EventHandler) handleConversationJoin(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract conversation data
	var conversationData struct {
		ConversationID string `json:"conversation_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &conversationData); err != nil {
		return fmt.Errorf("failed to parse conversation data: %w", err)
	}

	// Check if user can access conversation
	canAccess, err := h.messageRepo.UserCanAccessConversation(ctx, uuid.MustParse(conn.UserID), uuid.MustParse(conversationData.ConversationID))
	if err != nil {
		return fmt.Errorf("failed to check conversation access: %w", err)
	}

	if !canAccess {
		return fmt.Errorf("user cannot access conversation")
	}

	// Join conversation
	if err := h.connManager.JoinConversation(conn.UserID, conversationData.ConversationID); err != nil {
		return fmt.Errorf("failed to join conversation: %w", err)
	}

	// Subscribe to conversation channel
	conversationChannel := cache.GeneratePubSubChannel("conversation", conversationData.ConversationID)
	go h.connManager.handlePubSubSubscription(conn, conversationChannel)

	// Send conversation history (last 50 messages)
	messages, err := h.messageRepo.GetMessages(ctx, uuid.MustParse(conversationData.ConversationID), 50, 0)
	if err != nil {
		logger.Error("Failed to get conversation history", err)
	} else {
		historyMessage := Message{
			Type:      "conversation:history",
			Data:      messages,
			Timestamp: time.Now(),
		}
		conn.WriteMessage(historyMessage)
	}

	logger.Info("User joined conversation", 
		"user_id", conn.UserID,
		"conversation_id", conversationData.ConversationID,
	)

	return nil
}

// handleConversationLeave handles conversation leave events
func (h *EventHandler) handleConversationLeave(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract conversation data
	var conversationData struct {
		ConversationID string `json:"conversation_id"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &conversationData); err != nil {
		return fmt.Errorf("failed to parse conversation data: %w", err)
	}

	// Leave conversation
	if err := h.connManager.LeaveConversation(conn.UserID, conversationData.ConversationID); err != nil {
		return fmt.Errorf("failed to leave conversation: %w", err)
	}

	logger.Info("User left conversation", 
		"user_id", conn.UserID,
		"conversation_id", conversationData.ConversationID,
	)

	return nil
}

// handleUserStatus handles user status events
func (h *EventHandler) handleUserStatus(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract status data
	var statusData struct {
		Status string `json:"status"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &statusData); err != nil {
		return fmt.Errorf("failed to parse status data: %w", err)
	}

	// Update user status in cache
	statusKey := fmt.Sprintf("user_status:%s", conn.UserID)
	if err := h.cache.Set(ctx, statusKey, statusData.Status, 5*time.Minute); err != nil {
		logger.Error("Failed to cache user status", err)
	}

	// Broadcast status update
	statusMessage := Message{
		Type: "user:status_update",
		Data: map[string]interface{}{
			"user_id": conn.UserID,
			"status":  statusData.Status,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	if err := h.connManager.BroadcastToAll(statusMessage); err != nil {
		logger.Error("Failed to broadcast status update", err)
	}

	logger.Info("User status updated", 
		"user_id", conn.UserID,
		"status", statusData.Status,
	)

	return nil
}

// handlePing handles ping events
func (h *EventHandler) handlePing(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Send pong response
	pongMessage := Message{
		Type:      "pong",
		Data:      map[string]interface{}{"timestamp": time.Now()},
		Timestamp: time.Now(),
	}

	return conn.WriteMessage(pongMessage)
}

// Helper methods

// updateConversationActivity updates conversation activity timestamp
func (h *EventHandler) updateConversationActivity(ctx context.Context, conversationID uuid.UUID) error {
	// Get conversation
	conversation, err := h.messageRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Update timestamp
	conversation.UpdatedAt = time.Now()

	// Save conversation
	if err := h.messageRepo.UpdateConversation(ctx, conversation); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// updateUnreadCounts updates unread counts for message recipients
func (h *EventHandler) updateUnreadCounts(ctx context.Context, message *services.ProcessedMessage) error {
	// Get conversation participants
	participants := h.connManager.GetConversationParticipants(message.ConversationID.String())

	// Update unread count for each recipient (except sender)
	for _, participantID := range participants {
		if participantID != message.SenderID.String() {
			// Get current unread count
			unreadKey := fmt.Sprintf("unread:%s:%s", participantID, message.ConversationID.String())
			var currentCount int
			if count, err := h.cache.Get(ctx, unreadKey); err == nil {
				currentCount = count.(int)
			}

			// Increment count
			newCount := currentCount + 1
			if err := h.cache.Set(ctx, unreadKey, newCount, 24*time.Hour); err != nil {
				logger.Error("Failed to update unread count in cache", err)
			}

			// Send unread count update
			if err := h.connManager.UpdateUnreadCount(participantID, message.ConversationID.String(), newCount); err != nil {
				logger.Error("Failed to update unread count", err)
			}
		}
	}

	return nil
}

// updateUserUnreadCount updates unread count for a user in a conversation
func (h *EventHandler) updateUserUnreadCount(ctx context.Context, userID, conversationID uuid.UUID) error {
	// Get unread messages count
	count, err := h.messageRepo.GetConversationUnreadCount(ctx, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to get unread count: %w", err)
	}

	// Update cache
	unreadKey := fmt.Sprintf("unread:%s:%s", userID.String(), conversationID.String())
	if err := h.cache.Set(ctx, unreadKey, int(count), 24*time.Hour); err != nil {
		logger.Error("Failed to update unread count in cache", err)
	}

	// Send unread count update
	if err := h.connManager.UpdateUnreadCount(userID.String(), conversationID.String(), int(count)); err != nil {
		logger.Error("Failed to update unread count", err)
	}

	return nil
}

// handleEphemeralPhotoNew handles new ephemeral photo events
func (h *EventHandler) handleEphemeralPhotoNew(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract ephemeral photo data
	var photoData struct {
		ConversationID string `json:"conversation_id"`
		PhotoID        string `json:"photo_id"`
		AccessKey      string `json:"access_key"`
		ThumbnailURL   string `json:"thumbnail_url"`
		ExpiresAt      string `json:"expires_at"`
		Message        string `json:"message,omitempty"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &photoData); err != nil {
		return fmt.Errorf("failed to parse ephemeral photo data: %w", err)
	}

	// Broadcast ephemeral photo to conversation participants
	broadcastMessage := Message{
		Type: "ephemeral_photo:new",
		Data: map[string]interface{}{
			"conversation_id": photoData.ConversationID,
			"photo_id":        photoData.PhotoID,
			"access_key":      photoData.AccessKey,
			"thumbnail_url":   photoData.ThumbnailURL,
			"expires_at":      photoData.ExpiresAt,
			"message":         photoData.Message,
			"sender_id":       conn.UserID,
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	if err := h.connManager.BroadcastToConversation(photoData.ConversationID, broadcastMessage); err != nil {
		logger.Error("Failed to broadcast ephemeral photo", err)
	}

	logger.Info("Ephemeral photo broadcasted",
		"photo_id", photoData.PhotoID,
		"conversation_id", photoData.ConversationID,
		"sender_id", conn.UserID,
	)

	return nil
}

// handleEphemeralPhotoViewed handles ephemeral photo viewed events
func (h *EventHandler) handleEphemeralPhotoViewed(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract ephemeral photo data
	var photoData struct {
		PhotoID   string `json:"photo_id"`
		ViewerID  string `json:"viewer_id"`
		ViewedAt  string `json:"viewed_at"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &photoData); err != nil {
		return fmt.Errorf("failed to parse ephemeral photo viewed data: %w", err)
	}

	// Broadcast viewed event to conversation participants
	broadcastMessage := Message{
		Type: "ephemeral_photo:viewed",
		Data: map[string]interface{}{
			"photo_id":  photoData.PhotoID,
			"viewer_id": photoData.ViewerID,
			"viewed_at": photoData.ViewedAt,
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	// Get message to find conversation
	message, err := h.messageRepo.GetByID(ctx, uuid.MustParse(photoData.PhotoID))
	if err != nil {
		logger.Error("Failed to get message for ephemeral photo", err)
		return fmt.Errorf("failed to get message: %w", err)
	}

	if err := h.connManager.BroadcastToConversation(message.ConversationID.String(), broadcastMessage); err != nil {
		logger.Error("Failed to broadcast ephemeral photo viewed event", err)
	}

	logger.Info("Ephemeral photo viewed event broadcasted",
		"photo_id", photoData.PhotoID,
		"viewer_id", photoData.ViewerID,
	)

	return nil
}

// handleEphemeralPhotoExpired handles ephemeral photo expired events
func (h *EventHandler) handleEphemeralPhotoExpired(ctx context.Context, conn *ClientConnection, wsMessage Message) error {
	// Extract ephemeral photo data
	var photoData struct {
		PhotoID    string `json:"photo_id"`
		OwnerID    string `json:"owner_id"`
		ExpiredAt  string `json:"expired_at"`
	}
	
	if err := json.Unmarshal(wsMessage.Data.(json.RawMessage), &photoData); err != nil {
		return fmt.Errorf("failed to parse ephemeral photo expired data: %w", err)
	}

	// Broadcast expired event to photo owner
	expiredMessage := Message{
		Type: "ephemeral_photo:expired",
		Data: map[string]interface{}{
			"photo_id":   photoData.PhotoID,
			"owner_id":   photoData.OwnerID,
			"expired_at": photoData.ExpiredAt,
		},
		Timestamp: time.Now(),
		SenderID:  conn.UserID,
	}

	if err := h.connManager.BroadcastToUser(photoData.OwnerID, expiredMessage); err != nil {
		logger.Error("Failed to broadcast ephemeral photo expired event", err)
	}

	logger.Info("Ephemeral photo expired event broadcasted",
		"photo_id", photoData.PhotoID,
		"owner_id", photoData.OwnerID,
	)

	return nil
}

// cacheMessage caches a message for quick access
func (h *EventHandler) cacheMessage(ctx context.Context, message *services.ProcessedMessage) error {
	messageKey := fmt.Sprintf("message:%s", message.ID)
	
	// Cache message for 1 hour
	if err := h.cache.Set(ctx, messageKey, message, time.Hour); err != nil {
		return fmt.Errorf("failed to cache message: %w", err)
	}

	// Cache in conversation list
	conversationKey := fmt.Sprintf("conversation_messages:%s", message.ConversationID)
	if err := h.cache.LPush(ctx, conversationKey, message.ID.String()); err != nil {
		logger.Error("Failed to add message to conversation cache", err)
	}

	// Trim conversation cache to last 100 messages
	if err := h.cache.LTrim(ctx, conversationKey, 0, 99); err != nil {
		logger.Error("Failed to trim conversation cache", err)
	}

	// Set expiration on conversation cache
	if err := h.cache.Expire(ctx, conversationKey, time.Hour); err != nil {
		logger.Error("Failed to set expiration on conversation cache", err)
	}

	return nil
}