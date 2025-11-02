package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/chat"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/ephemeral_photo"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/websocket"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	getConversationsUseCase *chat.GetConversationsUseCase
	getMessagesUseCase     *chat.GetMessagesUseCase
	sendMessageUseCase    *chat.SendMessageUseCase
	markReadUseCase       *chat.MarkMessagesReadUseCase
	deleteMessageUseCase   *chat.DeleteMessageUseCase
	startConversationUseCase *chat.StartConversationUseCase
	sendEphemeralPhotoMessageUseCase *ephemeral_photo.SendEphemeralPhotoMessageUseCase
	getEphemeralPhotoMessageUseCase *ephemeral_photo.GetEphemeralPhotoMessageUseCase
	connManager           *websocket.ConnectionManager
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	getConversationsUseCase *chat.GetConversationsUseCase,
	getMessagesUseCase *chat.GetMessagesUseCase,
	sendMessageUseCase *chat.SendMessageUseCase,
	markReadUseCase *chat.MarkMessagesReadUseCase,
	deleteMessageUseCase *chat.DeleteMessageUseCase,
	startConversationUseCase *chat.StartConversationUseCase,
	sendEphemeralPhotoMessageUseCase *ephemeral_photo.SendEphemeralPhotoMessageUseCase,
	getEphemeralPhotoMessageUseCase *ephemeral_photo.GetEphemeralPhotoMessageUseCase,
	connManager *websocket.ConnectionManager,
) *ChatHandler {
	return &ChatHandler{
		getConversationsUseCase: getConversationsUseCase,
		getMessagesUseCase:     getMessagesUseCase,
		sendMessageUseCase:    sendMessageUseCase,
		markReadUseCase:       markReadUseCase,
		deleteMessageUseCase:   deleteMessageUseCase,
		startConversationUseCase: startConversationUseCase,
		sendEphemeralPhotoMessageUseCase: sendEphemeralPhotoMessageUseCase,
		getEphemeralPhotoMessageUseCase: getEphemeralPhotoMessageUseCase,
		connManager:           connManager,
	}
}

// GetConversations handles GET /api/v1/chats
func (h *ChatHandler) GetConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Create request
	req := &chat.GetConversationsRequest{
		UserID: userID.(uuid.UUID),
		Limit:  limit,
		Offset:  offset,
	}

	// Execute use case
	response, err := h.getConversationsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get conversations", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get conversations")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetMessages handles GET /api/v1/chats/:id/messages
func (h *ChatHandler) GetMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse conversation ID from URL
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Create request
	req := &chat.GetMessagesRequest{
		ConversationID: conversationID,
		UserID:        userID.(uuid.UUID),
		Limit:         limit,
		Offset:        offset,
	}

	// Execute use case
	response, err := h.getMessagesUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get messages", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// SendMessage handles POST /api/v1/chats/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse conversation ID from URL
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Parse request body
	var reqBody struct {
		Content     string `json:"content" validate:"required"`
		MessageType string `json:"message_type" validate:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create request
	req := &chat.SendMessageRequest{
		ConversationID: conversationID,
		SenderID:       userID.(uuid.UUID),
		Content:        reqBody.Content,
		MessageType:    reqBody.MessageType,
	}

	// Execute use case
	response, err := h.sendMessageUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to send message", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send message")
		return
	}

	if !response.Success {
		utils.ErrorResponse(c, http.StatusBadRequest, response.Error)
		return
	}

	// Broadcast message via WebSocket
	if response.Message != nil {
		wsMessage := websocket.Message{
			Type:      "message:new",
			Data:      response.Message,
			Timestamp: response.Message.CreatedAt,
			SenderID:  response.Message.SenderID.String(),
		}

		if err := h.connManager.BroadcastToConversation(conversationID.String(), wsMessage); err != nil {
			logger.Error("Failed to broadcast message via WebSocket", err)
			// Don't fail the request, just log the error
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, response.Message)
}

// MarkMessagesAsRead handles POST /api/v1/chats/:id/read
func (h *ChatHandler) MarkMessagesAsRead(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse conversation ID from URL
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Parse request body
	var reqBody struct {
		MessageIDs []string `json:"message_ids,omitempty"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert message IDs
	messageIDs := make([]uuid.UUID, 0, len(reqBody.MessageIDs))
	for _, idStr := range reqBody.MessageIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			messageIDs = append(messageIDs, id)
		}
	}

	// Create request
	req := &chat.MarkMessagesReadRequest{
		ConversationID: conversationID,
		UserID:        userID.(uuid.UUID),
		MessageIDs:    messageIDs,
	}

	// Execute use case
	response, err := h.markReadUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to mark messages as read", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to mark messages as read")
		return
	}

	if !response.Success {
		utils.ErrorResponse(c, http.StatusBadRequest, response.Error)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// DeleteMessage handles DELETE /api/v1/chats/:id/messages/:messageId
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse conversation ID from URL
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Parse message ID from URL
	messageIDStr := c.Param("messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid message ID")
		return
	}

	// Create request
	req := &chat.DeleteMessageRequest{
		MessageID: messageID,
		UserID:    userID.(uuid.UUID),
	}

	// Execute use case
	response, err := h.deleteMessageUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to delete message", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	if !response.Success {
		utils.ErrorResponse(c, http.StatusBadRequest, response.Error)
		return
	}

	// Broadcast delete event via WebSocket
	wsMessage := websocket.Message{
		Type:      "message:deleted",
		Data: map[string]interface{}{
			"message_id": messageID.String(),
			"user_id":    userID.(uuid.UUID).String(),
			"timestamp":   time.Now(),
		},
		Timestamp: time.Now(),
		SenderID:  userID.(uuid.UUID).String(),
	}

	if err := h.connManager.BroadcastToConversation(conversationID.String(), wsMessage); err != nil {
		logger.Error("Failed to broadcast delete event via WebSocket", err)
		// Don't fail the request, just log the error
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// StartConversation handles POST /api/v1/chats/start
func (h *ChatHandler) StartConversation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse request body
	var reqBody struct {
		MatchID     string `json:"match_id" validate:"required"`
		FirstMessage string `json:"first_message,omitempty"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse match ID
	matchID, err := uuid.Parse(reqBody.MatchID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Create request
	req := &chat.StartConversationRequest{
		UserID:      userID.(uuid.UUID),
		MatchID:      matchID,
		FirstMessage: reqBody.FirstMessage,
	}

	// Execute use case
	response, err := h.startConversationUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to start conversation", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start conversation")
		return
	}

	if response.Error != "" {
		utils.ErrorResponse(c, http.StatusBadRequest, response.Error)
		return
	}

	// If first message was sent, broadcast it via WebSocket
	if response.Message != nil {
		wsMessage := websocket.Message{
			Type:      "message:new",
			Data:      response.Message,
			Timestamp: response.Message.CreatedAt,
			SenderID:  response.Message.SenderID.String(),
		}

		if err := h.connManager.BroadcastToConversation(response.Conversation.ID.String(), wsMessage); err != nil {
			logger.Error("Failed to broadcast first message via WebSocket", err)
			// Don't fail the request, just log the error
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, response)
}

// SendEphemeralPhotoMessage handles POST /api/v1/chats/:id/ephemeral-photos
func (h *ChatHandler) SendEphemeralPhotoMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse conversation ID from URL
	conversationIDStr := c.Param("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	// Parse request body
	var reqBody struct {
		PhotoID uuid.UUID `json:"photo_id" validate:"required"`
		Message string    `json:"message" validate:"max=500"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create request
	req := &ephemeral_photo.SendEphemeralPhotoMessageRequest{
		ConversationID: conversationID,
		SenderID:       userID.(uuid.UUID),
		PhotoID:        reqBody.PhotoID,
		Message:        reqBody.Message,
	}

	// Execute use case
	response, err := h.sendEphemeralPhotoMessageUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to send ephemeral photo message", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send ephemeral photo message")
		return
	}

	// Broadcast message via WebSocket
	wsMessage := websocket.Message{
		Type: "ephemeral_photo:new",
		Data: map[string]interface{}{
			"message_id":    response.MessageID,
			"photo_id":      response.PhotoID,
			"access_key":    response.AccessKey,
			"thumbnail_url": response.ThumbnailURL,
			"expires_at":    response.ExpiresAt,
			"conversation_id": conversationID,
			"sender_id":     userID.(uuid.UUID),
		},
		Timestamp: time.Now(),
		SenderID:  userID.(uuid.UUID).String(),
	}

	if err := h.connManager.BroadcastToConversation(conversationID.String(), wsMessage); err != nil {
		logger.Error("Failed to broadcast ephemeral photo message via WebSocket", err)
		// Don't fail the request, just log the error
	}

	utils.SuccessResponse(c, http.StatusCreated, response)
}

// GetEphemeralPhotoMessage handles GET /api/v1/chats/:id/messages/:messageId/ephemeral-photo
func (h *ChatHandler) GetEphemeralPhotoMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse message ID from URL
	messageIDStr := c.Param("messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid message ID")
		return
	}

	// Create request
	req := &ephemeral_photo.GetEphemeralPhotoMessageRequest{
		MessageID: messageID,
		UserID:    userID.(uuid.UUID),
	}

	// Execute use case
	response, err := h.getEphemeralPhotoMessageUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get ephemeral photo message", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ephemeral photo message")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// WebSocketUpgrade handles WebSocket upgrade to /api/v1/ws
func (h *ChatHandler) WebSocketUpgrade(c *gin.Context) {
	// Handle WebSocket upgrade using connection manager
	if err := h.connManager.HandleConnection(c); err != nil {
		logger.Error("Failed to handle WebSocket connection", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to establish WebSocket connection")
		return
	}
}