package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// WebSocketTestHelper provides utilities for WebSocket testing
type WebSocketTestHelper struct {
	t            *testing.T
	serverURL    string
	dialer       *websocket.Dialer
	connections  []*TestConnection
	mu           sync.Mutex
	messageCount int
}

// TestConnection represents a WebSocket test connection
type TestConnection struct {
	conn       *websocket.Conn
	mu         sync.Mutex
	messages   []TestMessage
	isClosed   bool
	closeError error
}

// TestMessage represents a WebSocket test message
type TestMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	ID        string      `json:"id,omitempty"`
}

// NewWebSocketTestHelper creates a new WebSocket test helper
func NewWebSocketTestHelper(t *testing.T, serverURL string) *WebSocketTestHelper {
	return &WebSocketTestHelper{
		t:           t,
		serverURL:    serverURL,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		connections: make([]*TestConnection, 0),
	}
}

// Connect establishes a WebSocket connection
func (wsh *WebSocketTestHelper) Connect(path string, headers http.Header) *TestConnection {
	wsURL := wsh.getWebSocketURL(path)
	
	conn, _, err := wsh.dialer.Dial(wsURL, headers)
	require.NoError(wsh.t, err, "Failed to connect to WebSocket")
	
	testConn := &TestConnection{
		conn:     conn,
		messages: make([]TestMessage, 0),
		isClosed: false,
	}
	
	wsh.mu.Lock()
	wsh.connections = append(wsh.connections, testConn)
	wsh.mu.Unlock()
	
	// Start message listener
	go wsh.listenForMessages(testConn)
	
	return testConn
}

// ConnectWithToken establishes a WebSocket connection with authentication token
func (wsh *WebSocketTestHelper) ConnectWithToken(path, token string) *TestConnection {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)
	
	return wsh.Connect(path, headers)
}

// ConnectWithCookies establishes a WebSocket connection with cookies
func (wsh *WebSocketTestHelper) ConnectWithCookies(path string, cookies []*http.Cookie) *TestConnection {
	headers := http.Header{}
	
	for _, cookie := range cookies {
		headers.Add("Cookie", cookie.String())
	}
	
	return wsh.Connect(path, headers)
}

// getWebSocketURL converts HTTP URL to WebSocket URL
func (wsh *WebSocketTestHelper) getWebSocketURL(path string) string {
	u, err := url.Parse(wsh.serverURL)
	require.NoError(wsh.t, err, "Failed to parse server URL")
	
	// Convert scheme
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	
	// Update path
	u.Path = path
	
	return u.String()
}

// listenForMessages listens for messages on a connection
func (wsh *WebSocketTestHelper) listenForMessages(testConn *TestConnection) {
	for {
		var message TestMessage
		err := testConn.conn.ReadJSON(&message)
		if err != nil {
			testConn.mu.Lock()
			testConn.isClosed = true
			testConn.closeError = err
			testConn.mu.Unlock()
			break
		}
		
		testConn.mu.Lock()
		testConn.messages = append(testConn.messages, message)
		testConn.mu.Unlock()
		
		wsh.mu.Lock()
		wsh.messageCount++
		wsh.mu.Unlock()
	}
}

// SendMessage sends a message through a WebSocket connection
func (wsh *WebSocketTestHelper) SendMessage(testConn *TestConnection, message TestMessage) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.False(wsh.t, testConn.isClosed, "Connection is closed")
	
	err := testConn.conn.WriteJSON(message)
	require.NoError(wsh.t, err, "Failed to send message")
}

// SendRawMessage sends a raw message through a WebSocket connection
func (wsh *WebSocketTestHelper) SendRawMessage(testConn *TestConnection, message interface{}) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.False(wsh.t, testConn.isClosed, "Connection is closed")
	
	err := testConn.conn.WriteJSON(message)
	require.NoError(wsh.t, err, "Failed to send raw message")
}

// SendTextMessage sends a text message through a WebSocket connection
func (wsh *WebSocketTestHelper) SendTextMessage(testConn *TestConnection, message string) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.False(wsh.t, testConn.isClosed, "Connection is closed")
	
	err := testConn.conn.WriteMessage(websocket.TextMessage, []byte(message))
	require.NoError(wsh.t, err, "Failed to send text message")
}

// SendBinaryMessage sends a binary message through a WebSocket connection
func (wsh *WebSocketTestHelper) SendBinaryMessage(testConn *TestConnection, data []byte) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.False(wsh.t, testConn.isClosed, "Connection is closed")
	
	err := testConn.conn.WriteMessage(websocket.BinaryMessage, data)
	require.NoError(wsh.t, err, "Failed to send binary message")
}

// GetMessages returns all messages received by a connection
func (wsh *WebSocketTestHelper) GetMessages(testConn *TestConnection) []TestMessage {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	// Return a copy to avoid race conditions
	messages := make([]TestMessage, len(testConn.messages))
	copy(messages, testConn.messages)
	
	return messages
}

// GetLastMessage returns the last message received by a connection
func (wsh *WebSocketTestHelper) GetLastMessage(testConn *TestConnection) *TestMessage {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	if len(testConn.messages) == 0 {
		return nil
	}
	
	return &testConn.messages[len(testConn.messages)-1]
}

// GetMessagesByType returns messages of a specific type
func (wsh *WebSocketTestHelper) GetMessagesByType(testConn *TestConnection, messageType string) []TestMessage {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	var messages []TestMessage
	for _, msg := range testConn.messages {
		if msg.Type == messageType {
			messages = append(messages, msg)
		}
	}
	
	return messages
}

// GetMessageCount returns the number of messages received by a connection
func (wsh *WebSocketTestHelper) GetMessageCount(testConn *TestConnection) int {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	return len(testConn.messages)
}

// GetTotalMessageCount returns the total number of messages across all connections
func (wsh *WebSocketTestHelper) GetTotalMessageCount() int {
	wsh.mu.Lock()
	defer wsh.mu.Unlock()
	
	return wsh.messageCount
}

// WaitForMessage waits for a message to be received
func (wsh *WebSocketTestHelper) WaitForMessage(testConn *TestConnection, timeout time.Duration) *TestMessage {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if msg := wsh.GetLastMessage(testConn); msg != nil {
			return msg
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	require.Fail(wsh.t, "No message received within timeout")
	return nil
}

// WaitForMessageOfType waits for a message of a specific type
func (wsh *WebSocketTestHelper) WaitForMessageOfType(testConn *TestConnection, messageType string, timeout time.Duration) *TestMessage {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		messages := wsh.GetMessagesByType(testConn, messageType)
		if len(messages) > 0 {
			return &messages[len(messages)-1]
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	require.Fail(wsh.t, fmt.Sprintf("No message of type %s received within timeout", messageType))
	return nil
}

// WaitForMessageCount waits for a specific number of messages
func (wsh *WebSocketTestHelper) WaitForMessageCount(testConn *TestConnection, count int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if wsh.GetMessageCount(testConn) >= count {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	require.Fail(wsh.t, fmt.Sprintf("Did not receive %d messages within timeout", count))
}

// AssertMessageReceived asserts that a message was received
func (wsh *WebSocketTestHelper) AssertMessageReceived(testConn *TestConnection, expectedMessage TestMessage) {
	messages := wsh.GetMessages(testConn)
	require.NotEmpty(wsh.t, messages, "No messages received")
	
	// Check if any message matches the expected message
	found := false
	for _, msg := range messages {
		if wsh.messagesEqual(msg, expectedMessage) {
			found = true
			break
		}
	}
	
	require.True(wsh.t, found, "Expected message not received")
}

// AssertMessageTypeReceived asserts that a message of a specific type was received
func (wsh *WebSocketTestHelper) AssertMessageTypeReceived(testConn *TestConnection, messageType string) {
	messages := wsh.GetMessagesByType(testConn, messageType)
	require.NotEmpty(wsh.t, messages, "No message of type %s received", messageType)
}

// AssertMessageCount asserts that a specific number of messages were received
func (wsh *WebSocketTestHelper) AssertMessageCount(testConn *TestConnection, expectedCount int) {
	actualCount := wsh.GetMessageCount(testConn)
	require.Equal(wsh.t, expectedCount, actualCount, "Message count mismatch")
}

// AssertNoMessageReceived asserts that no messages were received
func (wsh *WebSocketTestHelper) AssertNoMessageReceived(testConn *TestConnection) {
	messages := wsh.GetMessages(testConn)
	require.Empty(wsh.t, messages, "No messages should have been received")
}

// AssertConnectionClosed asserts that a connection is closed
func (wsh *WebSocketTestHelper) AssertConnectionClosed(testConn *TestConnection) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.True(wsh.t, testConn.isClosed, "Connection should be closed")
}

// AssertConnectionOpen asserts that a connection is open
func (wsh *WebSocketTestHelper) AssertConnectionOpen(testConn *TestConnection) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	require.False(wsh.t, testConn.isClosed, "Connection should be open")
}

// CloseConnection closes a WebSocket connection
func (wsh *WebSocketTestHelper) CloseConnection(testConn *TestConnection) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	if !testConn.isClosed {
		err := testConn.conn.Close()
		testConn.isClosed = true
		testConn.closeError = err
	}
}

// CloseAllConnections closes all WebSocket connections
func (wsh *WebSocketTestHelper) CloseAllConnections() {
	wsh.mu.Lock()
	defer wsh.mu.Unlock()
	
	for _, conn := range wsh.connections {
		wsh.CloseConnection(conn)
	}
}

// GetConnectionCount returns the number of active connections
func (wsh *WebSocketTestHelper) GetConnectionCount() int {
	wsh.mu.Lock()
	defer wsh.mu.Unlock()
	
	count := 0
	for _, conn := range wsh.connections {
		conn.mu.Lock()
		if !conn.isClosed {
			count++
		}
		conn.mu.Unlock()
	}
	
	return count
}

// ClearMessages clears all messages for a connection
func (wsh *WebSocketTestHelper) ClearMessages(testConn *TestConnection) {
	testConn.mu.Lock()
	defer testConn.mu.Unlock()
	
	testConn.messages = make([]TestMessage, 0)
}

// ClearAllMessages clears all messages for all connections
func (wsh *WebSocketTestHelper) ClearAllMessages() {
	wsh.mu.Lock()
	defer wsh.mu.Unlock()
	
	for _, conn := range wsh.connections {
		wsh.ClearMessages(conn)
	}
	
	wsh.messageCount = 0
}

// messagesEqual checks if two messages are equal
func (wsh *WebSocketTestHelper) messagesEqual(msg1, msg2 TestMessage) bool {
	if msg1.Type != msg2.Type {
		return false
	}
	
	if msg1.ID != msg2.ID {
		return false
	}
	
	// Compare data as JSON
	data1, err1 := json.Marshal(msg1.Data)
	data2, err2 := json.Marshal(msg2.Data)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return string(data1) == string(data2)
}

// WebSocketRoomHelper provides utilities for testing WebSocket rooms
type WebSocketRoomHelper struct {
	wsh *WebSocketTestHelper
	t   *testing.T
}

// NewWebSocketRoomHelper creates a new WebSocket room helper
func NewWebSocketRoomHelper(t *testing.T, wsh *WebSocketTestHelper) *WebSocketRoomHelper {
	return &WebSocketRoomHelper{
		wsh: wsh,
		t:   t,
	}
}

// JoinRoom joins a WebSocket room
func (wrh *WebSocketRoomHelper) JoinRoom(roomID, userID, token string) *TestConnection {
	path := fmt.Sprintf("/ws/room/%s", roomID)
	return wrh.wsh.ConnectWithToken(path, token)
}

// LeaveRoom leaves a WebSocket room
func (wrh *WebSocketRoomHelper) LeaveRoom(testConn *TestConnection) {
	wrh.wsh.CloseConnection(testConn)
}

// SendRoomMessage sends a message to a room
func (wrh *WebSocketRoomHelper) SendRoomMessage(testConn *TestConnection, message string) {
	roomMessage := TestMessage{
		Type: "room_message",
		Data: map[string]interface{}{
			"message": message,
		},
	}
	wrh.wsh.SendMessage(testConn, roomMessage)
}

// AssertRoomMessageReceived asserts that a room message was received
func (wrh *WebSocketRoomHelper) AssertRoomMessageReceived(testConn *TestConnection, expectedMessage string) {
	messages := wrh.wsh.GetMessagesByType(testConn, "room_message")
	require.NotEmpty(wrh.t, messages, "No room messages received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if msg, ok := data["message"].(string); ok && msg == expectedMessage {
				found = true
				break
			}
		}
	}
	
	require.True(wrh.t, found, "Expected room message not received")
}

// AssertUserJoinedRoom asserts that a user joined room notification was received
func (wrh *WebSocketRoomHelper) AssertUserJoinedRoom(testConn *TestConnection, expectedUserID string) {
	messages := wrh.wsh.GetMessagesByType(testConn, "user_joined")
	require.NotEmpty(wrh.t, messages, "No user joined messages received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if userID, ok := data["user_id"].(string); ok && userID == expectedUserID {
				found = true
				break
			}
		}
	}
	
	require.True(wrh.t, found, "Expected user joined notification not received")
}

// AssertUserLeftRoom asserts that a user left room notification was received
func (wrh *WebSocketRoomHelper) AssertUserLeftRoom(testConn *TestConnection, expectedUserID string) {
	messages := wrh.wsh.GetMessagesByType(testConn, "user_left")
	require.NotEmpty(wrh.t, messages, "No user left messages received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if userID, ok := data["user_id"].(string); ok && userID == expectedUserID {
				found = true
				break
			}
		}
	}
	
	require.True(wrh.t, found, "Expected user left notification not received")
}

// WebSocketChatHelper provides utilities for testing WebSocket chat
type WebSocketChatHelper struct {
	wsh *WebSocketTestHelper
	t   *testing.T
}

// NewWebSocketChatHelper creates a new WebSocket chat helper
func NewWebSocketChatHelper(t *testing.T, wsh *WebSocketTestHelper) *WebSocketChatHelper {
	return &WebSocketChatHelper{
		wsh: wsh,
		t:   t,
	}
}

// ConnectToChat connects to a WebSocket chat
func (wch *WebSocketChatHelper) ConnectToChat(matchID, userID, token string) *TestConnection {
	path := fmt.Sprintf("/ws/chat/%s", matchID)
	return wch.wsh.ConnectWithToken(path, token)
}

// SendChatMessage sends a chat message
func (wch *WebSocketChatHelper) SendChatMessage(testConn *TestConnection, message string) {
	chatMessage := TestMessage{
		Type: "chat_message",
		Data: map[string]interface{}{
			"message": message,
		},
	}
	wch.wsh.SendMessage(testConn, chatMessage)
}

// SendTypingIndicator sends a typing indicator
func (wch *WebSocketChatHelper) SendTypingIndicator(testConn *TestConnection, isTyping bool) {
	typingMessage := TestMessage{
		Type: "typing_indicator",
		Data: map[string]interface{}{
			"is_typing": isTyping,
		},
	}
	wch.wsh.SendMessage(testConn, typingMessage)
}

// AssertChatMessageReceived asserts that a chat message was received
func (wch *WebSocketChatHelper) AssertChatMessageReceived(testConn *TestConnection, expectedMessage string, expectedSenderID string) {
	messages := wch.wsh.GetMessagesByType(testConn, "chat_message")
	require.NotEmpty(wch.t, messages, "No chat messages received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			message, msgOk := data["message"].(string)
			senderID, senderOk := data["sender_id"].(string)
			
			if msgOk && senderOk && message == expectedMessage && senderID == expectedSenderID {
				found = true
				break
			}
		}
	}
	
	require.True(wch.t, found, "Expected chat message not received")
}

// AssertTypingIndicatorReceived asserts that a typing indicator was received
func (wch *WebSocketChatHelper) AssertTypingIndicatorReceived(testConn *TestConnection, expectedSenderID string, expectedIsTyping bool) {
	messages := wch.wsh.GetMessagesByType(testConn, "typing_indicator")
	require.NotEmpty(wch.t, messages, "No typing indicators received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			senderID, senderOk := data["sender_id"].(string)
			isTyping, typingOk := data["is_typing"].(bool)
			
			if senderOk && typingOk && senderID == expectedSenderID && isTyping == expectedIsTyping {
				found = true
				break
			}
		}
	}
	
	require.True(wch.t, found, "Expected typing indicator not received")
}

// AssertMessageRead asserts that a message read notification was received
func (wch *WebSocketChatHelper) AssertMessageRead(testConn *TestConnection, expectedMessageID string) {
	messages := wch.wsh.GetMessagesByType(testConn, "message_read")
	require.NotEmpty(wch.t, messages, "No message read notifications received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			messageID, ok := data["message_id"].(string)
			
			if ok && messageID == expectedMessageID {
				found = true
				break
			}
		}
	}
	
	require.True(wch.t, found, "Expected message read notification not received")
}

// WebSocketNotificationHelper provides utilities for testing WebSocket notifications
type WebSocketNotificationHelper struct {
	wsh *WebSocketTestHelper
	t   *testing.T
}

// NewWebSocketNotificationHelper creates a new WebSocket notification helper
func NewWebSocketNotificationHelper(t *testing.T, wsh *WebSocketTestHelper) *WebSocketNotificationHelper {
	return &WebSocketNotificationHelper{
		wsh: wsh,
		t:   t,
	}
}

// ConnectToNotifications connects to WebSocket notifications
func (wnh *WebSocketNotificationHelper) ConnectToNotifications(userID, token string) *TestConnection {
	path := fmt.Sprintf("/ws/notifications/%s", userID)
	return wnh.wsh.ConnectWithToken(path, token)
}

// AssertNotificationReceived asserts that a notification was received
func (wnh *WebSocketNotificationHelper) AssertNotificationReceived(testConn *TestConnection, expectedType string, expectedData map[string]interface{}) {
	messages := wnh.wsh.GetMessagesByType(testConn, "notification")
	require.NotEmpty(wnh.t, messages, "No notifications received")
	
	found := false
	for _, msg := range messages {
		if data, ok := msg.Data.(map[string]interface{}); ok {
			notificationType, typeOk := data["type"].(string)
			
			if typeOk && notificationType == expectedType {
				// Check if all expected data fields match
				dataMatch := true
				for key, expectedValue := range expectedData {
					if actualValue, exists := data[key]; !exists || actualValue != expectedValue {
						dataMatch = false
						break
					}
				}
				
				if dataMatch {
					found = true
					break
				}
			}
		}
	}
	
	require.True(wnh.t, found, "Expected notification not received")
}

// AssertMatchNotificationReceived asserts that a match notification was received
func (wnh *WebSocketNotificationHelper) AssertMatchNotificationReceived(testConn *TestConnection, expectedMatchID string) {
	expectedData := map[string]interface{}{
		"match_id": expectedMatchID,
	}
	wnh.AssertNotificationReceived(testConn, "new_match", expectedData)
}

// AssertMessageNotificationReceived asserts that a message notification was received
func (wnh *WebSocketNotificationHelper) AssertMessageNotificationReceived(testConn *TestConnection, expectedMessageID, expectedSenderID string) {
	expectedData := map[string]interface{}{
		"message_id": expectedMessageID,
		"sender_id":  expectedSenderID,
	}
	wnh.AssertNotificationReceived(testConn, "new_message", expectedData)
}

// AssertProfileViewNotificationReceived asserts that a profile view notification was received
func (wnh *WebSocketNotificationHelper) AssertProfileViewNotificationReceived(testConn *TestConnection, expectedViewerID string) {
	expectedData := map[string]interface{}{
		"viewer_id": expectedViewerID,
	}
	wnh.AssertNotificationReceived(testConn, "profile_view", expectedData)
}