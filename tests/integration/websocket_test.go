package integration

import (
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

	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/websocket"
)

// WebSocketIntegrationTestSuite tests WebSocket functionality
type WebSocketIntegrationTestSuite struct {
	suite.Suite
	router           *gin.Engine
	connectionManager *websocket.ConnectionManager
	chatCacheService  *services.ChatCacheService
}

// SetupSuite sets up test suite
func (suite *WebSocketIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	redisClient := redis.NewMockRedisClient()
	suite.chatCacheService = services.NewChatCacheService(redisClient, nil)
	
	// Create WebSocket connection manager
	suite.connectionManager = websocket.NewConnectionManager(
		suite.chatCacheService,
		nil, // Mock message service
		nil, // Mock security service
		nil, // Use default WebSocket config
	)
	
	// Create router
	suite.router = gin.New()
	
	// Add WebSocket route
	suite.router.GET("/ws", func(c *gin.Context) {
		suite.connectionManager.HandleWebSocket(c)
	})
}

// TestWebSocketConnectionEstablishment tests basic WebSocket connection
func (suite *WebSocketIntegrationTestSuite) TestWebSocketConnectionEstablishment() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Read connection established message
	_, message, err := conn.ReadMessage()
	require.NoError(suite.T(), err)
	
	var response map[string]interface{}
	err = json.Unmarshal(message, &response)
	require.NoError(suite.T(), err)
	
	// Verify connection established event
	assert.Equal(suite.T(), "connection:established", response["event"])
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["user_id"])
	assert.NotEmpty(suite.T(), data["connection_id"])
}

// TestWebSocketAuthentication tests WebSocket authentication
func (suite *WebSocketIntegrationTestSuite) TestWebSocketAuthentication() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Test connection without auth token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		conn.Close()
		suite.T().Error("Expected connection to fail without auth token")
	}
	
	// Test connection with invalid auth token
	header := http.Header{}
	header.Set("Authorization", "Bearer invalid-token")
	
	conn, _, err = websocket.DefaultDialer.Dial(wsURL, header)
	if err == nil {
		conn.Close()
		suite.T().Error("Expected connection to fail with invalid auth token")
	}
}

// TestMultipleConnectionsPerUser tests multiple connections for same user
func (suite *WebSocketIntegrationTestSuite) TestMultipleConnectionsPerUser() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Create multiple connections for same user
	var connections []*websocket.Conn
	for i := 0; i < 3; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			connections = append(connections, conn)
		}
	}
	
	// Clean up connections
	for _, conn := range connections {
		conn.Close()
	}
	
	// Verify that multiple connections are allowed (up to limit)
	assert.GreaterOrEqual(suite.T(), len(connections), 1)
}

// TestMessageBroadcasting tests message broadcasting to multiple users
func (suite *WebSocketIntegrationTestSuite) TestMessageBroadcasting() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect two users
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn1.Close()
	
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn2.Close()
	
	// Wait for connections to be established
	suite.waitForConnectionEstablished(conn1)
	suite.waitForConnectionEstablished(conn2)
	
	// Join both users to same conversation
	conversationID := uuid.New().String()
	
	joinMessage := map[string]interface{}{
		"event": "conversation:join",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	// User 1 joins conversation
	suite.sendWebSocketMessage(conn1, joinMessage)
	suite.waitForEvent(conn1, "conversation:joined")
	
	// User 2 joins conversation
	suite.sendWebSocketMessage(conn2, joinMessage)
	suite.waitForEvent(conn2, "conversation:joined")
	
	// User 1 sends a message
	message := map[string]interface{}{
		"event": "message:send",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
			"content":         "Hello from user 1!",
			"type":            "text",
		},
	}
	
	suite.sendWebSocketMessage(conn1, message)
	
	// User 2 should receive the message
	receivedMessage := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), "message:new", receivedMessage["event"])
	
	data := receivedMessage["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "message")
	assert.Equal(suite.T(), "Hello from user 1!", data["message"].(map[string]interface{})["content"])
}

// TestTypingIndicatorsBroadcast tests typing indicator broadcasting
func (suite *WebSocketIntegrationTestSuite) TestTypingIndicatorsBroadcast() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect two users
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn1.Close()
	
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn2.Close()
	
	// Wait for connections to be established
	suite.waitForConnectionEstablished(conn1)
	suite.waitForConnectionEstablished(conn2)
	
	// Join both users to same conversation
	conversationID := uuid.New().String()
	
	joinMessage := map[string]interface{}{
		"event": "conversation:join",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	suite.sendWebSocketMessage(conn1, joinMessage)
	suite.waitForEvent(conn1, "conversation:joined")
	
	suite.sendWebSocketMessage(conn2, joinMessage)
	suite.waitForEvent(conn2, "conversation:joined")
	
	// User 1 starts typing
	typingStart := map[string]interface{}{
		"event": "typing:start",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	suite.sendWebSocketMessage(conn1, typingStart)
	
	// User 2 should receive typing indicator
	typingIndicator := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), "typing:indicator", typingIndicator["event"])
	
	data := typingIndicator["data"].(map[string]interface{})
	assert.Equal(suite.T(), true, data["is_typing"])
	assert.Equal(suite.T(), conversationID, data["conversation_id"])
	
	// User 1 stops typing
	typingStop := map[string]interface{}{
		"event": "typing:stop",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	suite.sendWebSocketMessage(conn1, typingStop)
	
	// User 2 should receive typing indicator cleared
	typingCleared := suite.waitForMessage(conn2)
	assert.Equal(suite.T(), "typing:indicator", typingCleared["event"])
	
	clearedData := typingCleared["data"].(map[string]interface{})
	assert.Equal(suite.T(), false, clearedData["is_typing"])
	assert.Equal(suite.T(), conversationID, clearedData["conversation_id"])
}

// TestOnlineStatusBroadcast tests online status broadcasting
func (suite *WebSocketIntegrationTestSuite) TestOnlineStatusBroadcast() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect user
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Update user status to online
	statusMessage := map[string]interface{}{
		"event": "user:status",
		"data": map[string]interface{}{
			"status": "online",
		},
	}
	
	suite.sendWebSocketMessage(conn, statusMessage)
	
	// Should receive status updated confirmation
	statusUpdate := suite.waitForMessage(conn)
	assert.Equal(suite.T(), "user:status_updated", statusUpdate["event"])
}

// TestConnectionTimeout tests connection timeout handling
func (suite *WebSocketIntegrationTestSuite) TestConnectionTimeout() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Don't send any messages, wait for timeout
	time.Sleep(35 * time.Second) // Longer than ping interval
	
	// Try to send a message after timeout
	message := map[string]interface{}{
		"event": "ping",
		"data":  map[string]interface{}{},
	}
	
	messageBytes, _ := json.Marshal(message)
	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	
	// Connection should be closed due to timeout
	if err == nil {
		suite.T().Error("Expected connection to be closed due to timeout")
	}
}

// TestInvalidMessageFormat tests handling of invalid message formats
func (suite *WebSocketIntegrationTestSuite) TestInvalidMessageFormat() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Send invalid JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	require.NoError(suite.T(), err)
	
	// Should receive error message
	errorMessage := suite.waitForMessage(conn)
	assert.Equal(suite.T(), "error", errorMessage["event"])
	
	data := errorMessage["data"].(map[string]interface{})
	assert.Equal(suite.T(), "INVALID_MESSAGE_FORMAT", data["code"])
}

// TestUnsupportedEventType tests handling of unsupported event types
func (suite *WebSocketIntegrationTestSuite) TestUnsupportedEventType() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Send unsupported event type
	message := map[string]interface{}{
		"event": "unsupported:event",
		"data":  map[string]interface{}{},
	}
	
	suite.sendWebSocketMessage(conn, message)
	
	// Should receive error message
	errorMessage := suite.waitForMessage(conn)
	assert.Equal(suite.T(), "error", errorMessage["event"])
	
	data := errorMessage["data"].(map[string]interface{})
	assert.Equal(suite.T(), "UNSUPPORTED_EVENT", data["code"])
}

// TestMessageSizeLimit tests message size limits
func (suite *WebSocketIntegrationTestSuite) TestMessageSizeLimit() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Send very large message (exceeding 32KB limit)
	largeMessage := make([]byte, 33*1024) // 33KB
	for i := range largeMessage {
		largeMessage[i] = 'a'
	}
	
	err = conn.WriteMessage(websocket.TextMessage, largeMessage)
	require.NoError(suite.T(), err)
	
	// Should receive error message
	errorMessage := suite.waitForMessage(conn)
	assert.Equal(suite.T(), "error", errorMessage["event"])
	
	data := errorMessage["data"].(map[string]interface{})
	assert.Equal(suite.T(), "MESSAGE_TOO_LARGE", data["code"])
}

// TestConversationLeave tests leaving a conversation
func (suite *WebSocketIntegrationTestSuite) TestConversationLeave() {
	// Create test server
	server := httptest.NewServer(suite.router)
	defer server.Close()
	
	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + server.URL[4:] + "/ws"
	
	// Connect user
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(suite.T(), err)
	defer conn.Close()
	
	// Wait for connection to be established
	suite.waitForConnectionEstablished(conn)
	
	// Join conversation
	conversationID := uuid.New().String()
	
	joinMessage := map[string]interface{}{
		"event": "conversation:join",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	suite.sendWebSocketMessage(conn, joinMessage)
	suite.waitForEvent(conn, "conversation:joined")
	
	// Leave conversation
	leaveMessage := map[string]interface{}{
		"event": "conversation:leave",
		"data": map[string]interface{}{
			"conversation_id": conversationID,
		},
	}
	
	suite.sendWebSocketMessage(conn, leaveMessage)
	
	// Should receive left confirmation
	leftMessage := suite.waitForMessage(conn)
	assert.Equal(suite.T(), "conversation:left", leftMessage["event"])
	
	data := leftMessage["data"].(map[string]interface{})
	assert.Equal(suite.T(), conversationID, data["conversation_id"])
}

// Helper methods

func (suite *WebSocketIntegrationTestSuite) waitForConnectionEstablished(conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		suite.T().Fatalf("Failed to receive connection established message: %v", err)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(message, &response)
	if err != nil {
		suite.T().Fatalf("Failed to parse connection established message: %v", err)
	}
	
	if response["event"] != "connection:established" {
		suite.T().Fatalf("Expected connection:established event, got: %v", response["event"])
	}
}

func (suite *WebSocketIntegrationTestSuite) sendWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		suite.T().Fatalf("Failed to marshal WebSocket message: %v", err)
	}
	
	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		suite.T().Fatalf("Failed to send WebSocket message: %v", err)
	}
}

func (suite *WebSocketIntegrationTestSuite) waitForMessage(conn *websocket.Conn) map[string]interface{} {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		suite.T().Fatalf("Failed to receive WebSocket message: %v", err)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(message, &response)
	if err != nil {
		suite.T().Fatalf("Failed to parse WebSocket message: %v", err)
	}
	
	return response
}

func (suite *WebSocketIntegrationTestSuite) waitForEvent(conn *websocket.Conn, expectedEvent string) {
	for i := 0; i < 10; i++ { // Try 10 times
		response := suite.waitForMessage(conn)
		if response["event"] == expectedEvent {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	suite.T().Fatalf("Expected event %s not received", expectedEvent)
}

// TestWebSocketIntegration runs all WebSocket integration tests
func TestWebSocketIntegration(t *testing.T) {
	suite.Run(t, new(WebSocketIntegrationTestSuite))
}