package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/chat"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/websocket"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/validator"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ChatIntegrationTestSuite tests chat endpoints and WebSocket functionality
type ChatIntegrationTestSuite struct {
	suite.Suite
	router              *gin.Engine
	chatHandler         *handlers.ChatHandler
	redisClient         *redis.RedisClient
	connectionManager    *websocket.ConnectionManager
	messageService      *services.MessageService
	chatSecurityService *services.ChatSecurityService
	chatCacheService    *services.ChatCacheService
}

// SetupSuite sets up the test suite
func (suite *ChatIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	suite.redisClient = redis.NewMockRedisClient()
	
	// Create chat services
	suite.messageService = services.NewMessageService(nil, suite.redisClient)
	suite.chatSecurityService = services.NewChatSecurityService(nil, suite.redisClient)
	suite.chatCacheService = services.NewChatCacheService(suite.redisClient, nil)
	
	// Create WebSocket connection manager
	suite.connectionManager = websocket.NewConnectionManager(
		suite.chatCacheService,
		suite.messageService,
		suite.chatSecurityService,
		nil, // Use default WebSocket config
	)
	
	// Create JWT utils
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24*7)
	
	// Create use cases with mock repositories
	getConversationsUseCase := chat.NewGetConversationsUseCase(nil, suite.chatCacheService)
	getMessagesUseCase := chat.NewGetMessagesUseCase(nil, suite.chatCacheService)
	sendMessageUseCase := chat.NewSendMessageUseCase(nil, suite.messageService, suite.chatSecurityService, suite.chatCacheService, suite.connectionManager)
	markMessagesReadUseCase := chat.NewMarkMessagesReadUseCase(nil, suite.chatCacheService, suite.connectionManager)
	deleteMessageUseCase := chat.NewDeleteMessageUseCase(nil, suite.chatSecurityService, suite.chatCacheService, suite.connectionManager)
	startConversationUseCase := chat.NewStartConversationUseCase(nil, nil, suite.chatCacheService, suite.connectionManager)
	
	// Create chat handler
	suite.chatHandler = handlers.NewChatHandler(
		getConversationsUseCase,
		getMessagesUseCase,
		sendMessageUseCase,
		markMessagesReadUseCase,
		deleteMessageUseCase,
		startConversationUseCase,
		suite.connectionManager,
		jwtUtils,
	)
	
	// Create router
	suite.router = gin.New()
	
	// Add chat routes
	chatGroup := suite.router.Group("/api/v1/chats")
	{
		chatGroup.GET("", suite.chatHandler.GetConversations)
		chatGroup.POST("/start", suite.chatHandler.StartConversation)
		chatGroup.GET("/:conversationId/messages", suite.chatHandler.GetMessages)
		chatGroup.POST("/:conversationId/messages", suite.chatHandler.SendMessage)
		chatGroup.POST("/:conversationId/read", suite.chatHandler.MarkMessagesAsRead)
		chatGroup.DELETE("/:conversationId/messages/:messageId", suite.chatHandler.DeleteMessage)
	}
	
	// Add WebSocket route
	suite.router.GET("/ws", func(c *gin.Context) {
		suite.connectionManager.HandleWebSocket(c)
	})
}

// TestGetConversations tests retrieving user conversations
func (suite *ChatIntegrationTestSuite) TestGetConversations() {
	// Create test user and get auth token
	accessToken := suite.createTestUser("testuser1")
	
	// Prepare request
	req := httptest.NewRequest("GET", "/api/v1/chats?page=1&limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.ConversationsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Data.Conversations)
	assert.NotNil(suite.T(), response.Data.Pagination)
}

// TestGetMessages tests retrieving messages from a conversation
func (suite *ChatIntegrationTestSuite) TestGetMessages() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Prepare request
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chats/%s/messages?page=1&limit=20", conversationID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessagesResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Data.Messages)
	assert.NotNil(suite.T(), response.Data.Pagination)
}

// TestSendMessage tests sending a text message
func (suite *ChatIntegrationTestSuite) TestSendMessage() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Prepare message request
	req := dto.SendMessageDTO{
		Content: "Hello, this is a test message!",
		Type:    "text",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Data.Message)
	assert.Equal(suite.T(), response.Data.Message.Content, "Hello, this is a test message!")
	assert.Equal(suite.T(), response.Data.Message.Type, "text")
}

// TestSendPhotoMessage tests sending a photo message
func (suite *ChatIntegrationTestSuite) TestSendPhotoMessage() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Prepare photo message request
	req := dto.SendMessageDTO{
		Content: "https://example.com/test-photo.jpg",
		Type:    "photo",
		Metadata: map[string]interface{}{
			"width":  800,
			"height": 600,
			"size":   1024000,
		},
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.Message.Type, "photo")
	assert.Equal(suite.T(), response.Data.Message.Metadata["width"], 800)
	assert.Equal(suite.T(), response.Data.Message.Metadata["height"], 600)
}

// TestSendLocationMessage tests sending a location message
func (suite *ChatIntegrationTestSuite) TestSendLocationMessage() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Prepare location message request
	req := dto.SendMessageDTO{
		Content: "40.7128,-74.0060",
		Type:    "location",
		Metadata: map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
			"address":   "New York, NY, USA",
		},
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), response.Data.Message.Type, "location")
	assert.Equal(suite.T(), response.Data.Message.Metadata["latitude"], 40.7128)
	assert.Equal(suite.T(), response.Data.Message.Metadata["longitude"], -74.0060)
}

// TestMarkMessagesAsRead tests marking messages as read
func (suite *ChatIntegrationTestSuite) TestMarkMessagesAsRead() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	messageID := uuid.New().String()
	
	// Prepare mark as read request
	req := dto.MarkReadRequestDTO{
		MessageID: messageID,
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/read", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestDeleteMessage tests deleting a message
func (suite *ChatIntegrationTestSuite) TestDeleteMessage() {
	// Create test users and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	messageID := uuid.New().String()
	
	// Prepare delete request
	httpReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/chats/%s/messages/%s", conversationID, messageID), nil)
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestStartConversation tests starting a new conversation
func (suite *ChatIntegrationTestSuite) TestStartConversation() {
	// Create test users
	accessToken := suite.createTestUser("testuser1")
	matchID := uuid.New().String()
	
	// Prepare start conversation request
	req := dto.StartConversationDTO{
		MatchID:       matchID,
		InitialMessage: "Hey! I noticed we matched. How are you?",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/chats/start", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.ConversationResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.NotNil(suite.T(), response.Data.Conversation)
	assert.Equal(suite.T(), response.Data.Conversation.MatchID, matchID)
}

// TestWebSocketConnection tests WebSocket connection and messaging
func (suite *ChatIntegrationTestSuite) TestWebSocketConnection() {
	// Create test user
	accessToken := suite.createTestUser("testuser1")
	
	// Convert HTTP server to WebSocket server URL
	serverURL := "ws://localhost:8080/ws"
	
	// Create WebSocket dialer with auth header
	dialer := websocket.Dialer{}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+accessToken)
	
	// Connect to WebSocket
	conn, _, err := dialer.Dial(serverURL, header)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection established message
	_, message, err := conn.ReadMessage()
	require.NoError(suite.T(), err)
	
	var wsResponse map[string]interface{}
	err = json.Unmarshal(message, &wsResponse)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), wsResponse["event"], "connection:established")
}

// TestWebSocketMessageExchange tests real-time message exchange via WebSocket
func (suite *ChatIntegrationTestSuite) TestWebSocketMessageExchange() {
	// Create two test users
	user1Token := suite.createTestUser("testuser1")
	user2Token := suite.createTestUser("testuser2")
	conversationID := suite.createTestConversation()
	
	// Connect both users via WebSocket
	conn1 := suite.connectWebSocket(user1Token)
	defer conn1.Close()
	
	conn2 := suite.connectWebSocket(user2Token)
	defer conn2.Close()
	
	// Wait for connections to be established
	suite.waitForConnectionEstablished(conn1)
	suite.waitForConnectionEstablished(conn2)
	
	// Join conversation room
	suite.joinConversation(conn1, conversationID)
	suite.joinConversation(conn2, conversationID)
	
	// User 1 sends a message
	message := map[string]interface{}{
		"event": "message:send",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
			"content":         "Hello from user 1!",
			"type":            "text",
		},
	}
	
	messageBytes, _ := json.Marshal(message)
	err := conn1.WriteMessage(websocket.TextMessage, messageBytes)
	require.NoError(suite.T(), err)
	
	// User 2 should receive the message
	receivedMessage := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), receivedMessage["event"], "message:new")
	assert.Contains(suite.T(), receivedMessage["data"], "message")
}

// TestTypingIndicators tests typing indicator functionality
func (suite *ChatIntegrationTestSuite) TestTypingIndicators() {
	// Create two test users
	user1Token := suite.createTestUser("testuser1")
	user2Token := suite.createTestUser("testuser2")
	conversationID := suite.createTestConversation()
	
	// Connect both users via WebSocket
	conn1 := suite.connectWebSocket(user1Token)
	defer conn1.Close()
	
	conn2 := suite.connectWebSocket(user2Token)
	defer conn2.Close()
	
	// Wait for connections to be established
	suite.waitForConnectionEstablished(conn1)
	suite.waitForConnectionEstablished(conn2)
	
	// Join conversation room
	suite.joinConversation(conn1, conversationID)
	suite.joinConversation(conn2, conversationID)
	
	// User 1 starts typing
	typingStart := map[string]interface{}{
		"event": "typing:start",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	typingStartBytes, _ := json.Marshal(typingStart)
	err := conn1.WriteMessage(websocket.TextMessage, typingStartBytes)
	require.NoError(suite.T(), err)
	
	// User 2 should receive typing indicator
	typingIndicator := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), typingIndicator["event"], "typing:indicator")
	
	data := typingIndicator["data"].(map[string]interface{})
	assert.Equal(suite.T(), data["is_typing"], true)
	assert.Equal(suite.T(), data["conversation_id"], conversationID)
	
	// User 1 stops typing
	typingStop := map[string]interface{}{
		"event": "typing:stop",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	typingStopBytes, _ := json.Marshal(typingStop)
	err = conn1.WriteMessage(websocket.TextMessage, typingStopBytes)
	require.NoError(suite.T(), err)
	
	// User 2 should receive typing indicator cleared
	typingCleared := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), typingCleared["event"], "typing:indicator")
	
	clearedData := typingCleared["data"].(map[string]interface{})
	assert.Equal(suite.T(), clearedData["is_typing"], false)
}

// TestMessageValidation tests message content validation
func (suite *ChatIntegrationTestSuite) TestMessageValidation() {
	// Create test user and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Test empty message
	req := dto.SendMessageDTO{
		Content: "",
		Type:    "text",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Should return validation error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response dto.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "content")
}

// TestRateLimiting tests chat rate limiting
func (suite *ChatIntegrationTestSuite) TestRateLimiting() {
	// Create test user and conversation
	accessToken := suite.createTestUser("testuser1")
	conversationID := suite.createTestConversation()
	
	// Send multiple messages rapidly to trigger rate limit
	for i := 0; i < 35; i++ { // Exceeds rate limit of 30 per minute
		req := dto.SendMessageDTO{
			Content: fmt.Sprintf("Message %d", i),
			Type:    "text",
		}
		
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Authorization", "Bearer "+accessToken)
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)
		
		// First 30 messages should succeed
		if i < 30 {
			assert.Equal(suite.T(), http.StatusCreated, w.Code)
		}
	}
	
	// 31st message should be rate limited
	req := dto.SendMessageDTO{
		Content: "This should be rate limited",
		Type:    "text",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Should be rate limited
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestUnauthorizedAccess tests unauthorized access to chat endpoints
func (suite *ChatIntegrationTestSuite) TestUnauthorizedAccess() {
	// Test accessing conversations without auth token
	req := httptest.NewRequest("GET", "/api/v1/chats", nil)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return unauthorized
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestInvalidConversation tests accessing non-existent conversation
func (suite *ChatIntegrationTestSuite) TestInvalidConversation() {
	// Create test user
	accessToken := suite.createTestUser("testuser1")
	invalidConversationID := uuid.New().String()
	
	// Try to access non-existent conversation
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chats/%s/messages", invalidConversationID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return not found
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Helper methods

func (suite *ChatIntegrationTestSuite) createTestUser(username string) string {
	// This would normally create a user in the database
	// For testing purposes, we'll return a mock JWT token
	return "mock-jwt-token-for-" + username
}

func (suite *ChatIntegrationTestSuite) createTestConversation() string {
	// This would normally create a conversation in the database
	// For testing purposes, we'll return a mock conversation ID
	return uuid.New().String()
}

func (suite *ChatIntegrationTestSuite) connectWebSocket(token string) *websocket.Conn {
	dialer := websocket.Dialer{}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	
	conn, _, err := dialer.Dial("ws://localhost:8080/ws", header)
	if err != nil {
		// For testing purposes, we'll create a mock connection
		// In real tests, you'd need to start a test server
		suite.T().Skip("WebSocket test requires running server")
	}
	
	return conn
}

func (suite *ChatIntegrationTestSuite) waitForConnectionEstablished(conn *websocket.Conn) {
	// Wait for connection established message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		suite.T().Fatalf("Failed to receive connection established message: %v", err)
	}
	
	var wsResponse map[string]interface{}
	err = json.Unmarshal(message, &wsResponse)
	if err != nil {
		suite.T().Fatalf("Failed to parse connection established message: %v", err)
	}
	
	if wsResponse["event"] != "connection:established" {
		suite.T().Fatalf("Expected connection:established event, got: %v", wsResponse["event"])
	}
}

func (suite *ChatIntegrationTestSuite) joinConversation(conn *websocket.Conn, conversationID string) {
	joinMessage := map[string]interface{}{
		"event": "conversation:join",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	messageBytes, _ := json.Marshal(joinMessage)
	err := conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		suite.T().Fatalf("Failed to join conversation: %v", err)
	}
}

func (suite *ChatIntegrationTestSuite) waitForMessage(conn *websocket.Conn) map[string]interface{} {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		suite.T().Fatalf("Failed to receive message: %v", err)
	}
	
	var wsResponse map[string]interface{}
	err = json.Unmarshal(message, &wsResponse)
	if err != nil {
		suite.T().Fatalf("Failed to parse message: %v", err)
	}
	
	return wsResponse
}

// TestChatIntegration runs all chat integration tests
func TestChatIntegration(t *testing.T) {
	suite.Run(t, new(ChatIntegrationTestSuite))
}