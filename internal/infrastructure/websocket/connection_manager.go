package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	connections map[string]*ClientConnection
	mu          sync.RWMutex
	pubSub      *cache.PubSubService
	sessionMgr  *cache.SessionManager
	chatRooms   map[string]*ChatRoom // Conversation ID -> ChatRoom
	chatMu      sync.RWMutex
	typingUsers map[string]map[string]time.Time // Conversation ID -> User ID -> Last typing time
	typingMu    sync.RWMutex
}

// ClientConnection represents a WebSocket client connection
type ClientConnection struct {
	Conn       *websocket.Conn
	UserID     string
	SessionID  string
	IPAddress  string
	UserAgent   string
	LastPing   time.Time
	IsAlive    bool
	mu          sync.RWMutex
	Channels    map[string]bool // Subscribed channels
	ActiveConversations map[string]bool // Conversation ID -> IsActive
}

// ChatRoom represents a chat room/conversation
type ChatRoom struct {
	ID           string
	Participants map[string]bool // User ID -> IsParticipant
	LastActivity time.Time
	mu           sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	SenderID  string      `json:"sender_id,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID             string      `json:"id"`
	ConversationID string      `json:"conversation_id"`
	SenderID       string      `json:"sender_id"`
	Content        string      `json:"content"`
	MessageType    string      `json:"message_type"`
	IsRead         bool        `json:"is_read"`
	CreatedAt      time.Time   `json:"created_at"`
}

// TypingIndicator represents a typing indicator
type TypingIndicator struct {
	ConversationID string    `json:"conversation_id"`
	UserID        string    `json:"user_id"`
	IsTyping      bool      `json:"is_typing"`
	Timestamp     time.Time `json:"timestamp"`
}

// OnlineStatus represents online status update
type OnlineStatus struct {
	UserID    string    `json:"user_id"`
	IsOnline  bool      `json:"is_online"`
	Timestamp time.Time `json:"timestamp"`
}

// UnreadCount represents unread message count update
type UnreadCount struct {
	ConversationID string `json:"conversation_id"`
	UserID        string `json:"user_id"`
	Count         int    `json:"count"`
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(pubSub *cache.PubSubService, sessionMgr *cache.SessionManager) *ConnectionManager {
	return &ConnectionManager{
		connections:  make(map[string]*ClientConnection),
		pubSub:       pubSub,
		sessionMgr:   sessionMgr,
		chatRooms:    make(map[string]*ChatRoom),
		typingUsers:  make(map[string]map[string]time.Time),
	}
}

// HandleConnection handles a new WebSocket connection
func (cm *ConnectionManager) HandleConnection(c *gin.Context) error {
	// Upgrade HTTP connection to WebSocket
	conn, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade WebSocket connection", err)
		return fmt.Errorf("failed to upgrade WebSocket connection: %w", err)
	}

	// Get user information from context
	userID, exists := c.Get("user_id")
	if !exists {
		conn.Close()
		return fmt.Errorf("user not authenticated")
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		conn.Close()
		return fmt.Errorf("session not found")
	}

	// Create client connection
	clientConn := &ClientConnection{
		Conn:               conn,
		UserID:             userID.(string),
		SessionID:          sessionID.(string),
		IPAddress:          c.ClientIP(),
		UserAgent:          c.GetHeader("User-Agent"),
		LastPing:          time.Now(),
		IsAlive:            true,
		Channels:           make(map[string]bool),
		ActiveConversations: make(map[string]bool),
	}

	// Add connection to manager
	connectionID := cm.generateConnectionID(userID.(string), sessionID.(string))
	cm.mu.Lock()
	cm.connections[connectionID] = clientConn
	cm.mu.Unlock()

	// Start connection handler
	go cm.handleConnection(clientConn, connectionID)

	logger.Info("WebSocket connection established", 
		"user_id", userID,
		"session_id", sessionID,
		"connection_id", connectionID,
		"ip_address", c.ClientIP(),
	)

	return nil
}

// handleConnection manages an active WebSocket connection
func (cm *ConnectionManager) handleConnection(conn *ClientConnection, connectionID string) {
	defer cm.removeConnection(connectionID)

	// Set ping handler for connection health
	conn.SetPingHandler(func(appData string) error {
		conn.mu.Lock()
		defer conn.mu.Unlock()
		
		conn.LastPing = time.Now()
		conn.IsAlive = true
		
		// Send pong response
		return conn.WriteMessage(Message{
			Type: "pong",
			Data: map[string]interface{}{
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
		})
	})

	// Set close handler
	conn.SetCloseHandler(func(code int, text string) error {
		logger.Info("WebSocket connection closed", 
			"connection_id", connectionID,
			"user_id", conn.UserID,
			"code", code,
			"reason", text,
		)
		
		conn.mu.Lock()
		conn.IsAlive = false
		conn.mu.Unlock()
		
		return nil
	})

	// Subscribe to user-specific channels
	cm.subscribeUserToChannels(conn)
	
	// Set user as online
	cm.setUserOnline(conn.UserID, true)

	// Start message reader
	for {
		select {
		case <-time.After(30 * time.Second):
			// Check connection health
			conn.mu.RLock()
			isAlive := conn.IsAlive
			lastPing := conn.LastPing
			conn.mu.RUnlock()
			
			if !isAlive || time.Since(lastPing) > 60*time.Second {
				logger.Info("Closing inactive WebSocket connection", "connection_id", connectionID)
				return
			}
			
			// Send ping to keep connection alive
			err := conn.WriteMessage(Message{
				Type: "ping",
				Data: map[string]interface{}{
					"timestamp": time.Now(),
				},
				Timestamp: time.Now(),
			})
			if err != nil {
				logger.Error("Failed to send ping", err, "connection_id", connectionID)
				return
			}
		}
	}
}

// removeConnection removes a connection from the manager
func (cm *ConnectionManager) removeConnection(connectionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if conn, exists := cm.connections[connectionID]; exists {
		conn.mu.Lock()
		conn.IsAlive = false
		conn.mu.Unlock()
		
		conn.Conn.Close()
		delete(cm.connections, connectionID)
		
		// Set user as offline if no more connections
		if !cm.hasActiveConnections(conn.UserID) {
			cm.setUserOnline(conn.UserID, false)
		}
		
		logger.Info("WebSocket connection removed", "connection_id", connectionID, "user_id", conn.UserID)
	}
}

// BroadcastToUser sends a message to all connections for a user
func (cm *ConnectionManager) BroadcastToUser(userID string, message Message) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	sentCount := 0
	for connectionID, conn := range cm.connections {
		if conn.UserID == userID && conn.IsAlive {
			err := conn.WriteMessage(message)
			if err != nil {
				logger.Error("Failed to send message to connection", err, 
					"connection_id", connectionID,
					"user_id", userID,
				)
			} else {
				sentCount++
			}
		}
	}
	
	logger.Debug("Message broadcast to user", 
		"user_id", userID,
		"connections_count", sentCount,
		"message_type", message.Type,
	)
	
	return nil
}

// BroadcastToChannel sends a message to all subscribers of a channel
func (cm *ConnectionManager) BroadcastToChannel(channel string, message Message) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	sentCount := 0
	for connectionID, conn := range cm.connections {
		if conn.IsAlive && conn.isSubscribedTo(channel) {
			err := conn.WriteMessage(message)
			if err != nil {
				logger.Error("Failed to send channel message to connection", err, 
					"connection_id", connectionID,
					"channel", channel,
				)
			} else {
				sentCount++
			}
		}
	}
	
	logger.Debug("Message broadcast to channel", 
		"channel", channel,
		"connections_count", sentCount,
		"message_type", message.Type,
	)
	
	return nil
}

// BroadcastToAll sends a message to all active connections
func (cm *ConnectionManager) BroadcastToAll(message Message) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	sentCount := 0
	for connectionID, conn := range cm.connections {
		if conn.IsAlive {
			err := conn.WriteMessage(message)
			if err != nil {
				logger.Error("Failed to send broadcast message to connection", err, "connection_id", connectionID)
			} else {
				sentCount++
			}
		}
	}
	
	logger.Debug("Message broadcast to all", 
		"connections_count", sentCount,
		"message_type", message.Type,
	)
	
	return nil
}

// GetConnectionCount returns the number of active connections
func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	return len(cm.connections)
}

// GetUserConnections returns all connections for a user
func (cm *ConnectionManager) GetUserConnections(userID string) []*ClientConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var connections []*ClientConnection
	for _, conn := range cm.connections {
		if conn.UserID == userID && conn.IsAlive {
			connections = append(connections, conn)
		}
	}
	
	return connections
}

// GetConnectionStats returns connection statistics
func (cm *ConnectionManager) GetConnectionStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	userConnections := make(map[string]int)
	totalConnections := 0
	
	for _, conn := range cm.connections {
		if conn.IsAlive {
			totalConnections++
			userConnections[conn.UserID]++
		}
	}
	
	stats := map[string]interface{}{
		"total_connections":    totalConnections,
		"unique_users":         len(userConnections),
		"user_connections":      userConnections,
	}
	
	return stats
}

// CleanupInactiveConnections removes inactive connections
func (cm *ConnectionManager) CleanupInactiveConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	now := time.Now()
	removedCount := 0
	
	for connectionID, conn := range cm.connections {
		conn.mu.Lock()
		isAlive := conn.IsAlive
		lastPing := conn.LastPing
		conn.mu.Unlock()
		
		if !isAlive || now.Sub(lastPing) > 90*time.Second {
			conn.Conn.Close()
			delete(cm.connections, connectionID)
			removedCount++
			
			logger.Info("Removed inactive WebSocket connection", 
				"connection_id", connectionID,
				"user_id", conn.UserID,
				"inactive_duration", now.Sub(lastPing),
			)
		}
	}
	
	if removedCount > 0 {
		logger.Info("Inactive connections cleanup completed", "removed_count", removedCount)
	}
}

// subscribeUserToChannels subscribes a user to their relevant channels
func (cm *ConnectionManager) subscribeUserToChannels(conn *ClientConnection) {
	// Subscribe to user-specific notification channel
	notificationChannel := cache.GeneratePubSubChannel("notifications", conn.UserID)
	go cm.handlePubSubSubscription(conn, notificationChannel)
	
	// Subscribe to online status updates
	onlineChannel := cache.GeneratePubSubChannel("online_status")
	go cm.handlePubSubSubscription(conn, onlineChannel)
	
	// Subscribe to match notifications
	matchChannel := cache.GeneratePubSubChannel("matches", conn.UserID)
	go cm.handlePubSubSubscription(conn, matchChannel)
}

// handlePubSubSubscription handles a Pub/Sub subscription for a connection
func (cm *ConnectionManager) handlePubSubSubscription(conn *ClientConnection, channel string) {
	msgChan, err := cm.pubSub.SubscribeToChannel(context.Background(), channel)
	if err != nil {
		logger.Error("Failed to subscribe to Pub/Sub channel", err, "channel", channel)
		return
	}
	
	defer func() {
		msgChan.Close()
		conn.mu.Lock()
		delete(conn.Channels, channel)
		conn.mu.Unlock()
	}()
	
	conn.mu.Lock()
	conn.Channels[channel] = true
	conn.mu.Unlock()
	
	for {
		select {
		case <-time.After(30 * time.Second):
			// Check subscription health
			conn.mu.RLock()
			isAlive := conn.IsAlive
			conn.mu.RUnlock()
			
			if !isAlive {
				return
			}
			
		case msg := <-msgChan:
			if !conn.IsAlive {
				return
			}
			
			// Forward Pub/Sub message to WebSocket client
			err := conn.WriteMessage(Message{
				Type:      msg.Type,
				Channel:   msg.Channel,
				Data:      msg.Data,
				Timestamp: msg.Timestamp,
			})
			if err != nil {
				logger.Error("Failed to forward Pub/Sub message to WebSocket", err, 
					"channel", channel,
					"message_type", msg.Type,
				)
			}
		}
	}
}

// generateConnectionID generates a unique connection ID
func (cm *ConnectionManager) generateConnectionID(userID, sessionID string) string {
	return fmt.Sprintf("%s:%s:%d", userID, sessionID, time.Now().UnixNano())
}

// WriteMessage writes a message to the WebSocket connection
func (conn *ClientConnection) WriteMessage(message Message) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	
	if !conn.IsAlive {
		return fmt.Errorf("connection is not alive")
	}
	
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	err = conn.Conn.WriteMessage(websocket.TextMessage, string(messageData))
	if err != nil {
		conn.IsAlive = false
		return fmt.Errorf("failed to write message: %w", err)
	}
	
	return nil
}

// isSubscribedTo checks if connection is subscribed to a channel
func (conn *ClientConnection) isSubscribedTo(channel string) bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	
	return conn.Channels[channel]
}

// SetPingHandler sets the ping handler for the connection
func (conn *ClientConnection) SetPingHandler(handler func(string) error) {
	// This would be implemented based on your WebSocket library
	// For gorilla/websocket, ping handling is built-in
	logger.Debug("Ping handler set for connection")
}

// SetCloseHandler sets the close handler for the connection
func (conn *ClientConnection) SetCloseHandler(handler func(int, string) error) {
	// This would be implemented based on your WebSocket library
	// For gorilla/websocket, close handling is built-in
	logger.Debug("Close handler set for connection")
}

// Chat-related methods

// JoinConversation adds a user to a conversation/chat room
func (cm *ConnectionManager) JoinConversation(userID, conversationID string) error {
	cm.chatMu.Lock()
	defer cm.chatMu.Unlock()
	
	// Get or create chat room
	room, exists := cm.chatRooms[conversationID]
	if !exists {
		room = &ChatRoom{
			ID:           conversationID,
			Participants: make(map[string]bool),
			LastActivity: time.Now(),
		}
		cm.chatRooms[conversationID] = room
	}
	
	// Add user to room
	room.mu.Lock()
	room.Participants[userID] = true
	room.LastActivity = time.Now()
	room.mu.Unlock()
	
	// Update user's active conversations
	cm.mu.Lock()
	for _, conn := range cm.connections {
		if conn.UserID == userID {
			conn.mu.Lock()
			conn.ActiveConversations[conversationID] = true
			conn.mu.Unlock()
		}
	}
	cm.mu.Unlock()
	
	logger.Info("User joined conversation",
		"user_id", userID,
		"conversation_id", conversationID,
	)
	
	return nil
}

// LeaveConversation removes a user from a conversation/chat room
func (cm *ConnectionManager) LeaveConversation(userID, conversationID string) error {
	cm.chatMu.Lock()
	defer cm.chatMu.Unlock()
	
	// Remove user from chat room
	if room, exists := cm.chatRooms[conversationID]; exists {
		room.mu.Lock()
		delete(room.Participants, userID)
		room.LastActivity = time.Now()
		room.mu.Unlock()
		
		// Clean up empty rooms
		if len(room.Participants) == 0 {
			delete(cm.chatRooms, conversationID)
			logger.Info("Empty conversation room removed", "conversation_id", conversationID)
		}
	}
	
	// Update user's active conversations
	cm.mu.Lock()
	for _, conn := range cm.connections {
		if conn.UserID == userID {
			conn.mu.Lock()
			delete(conn.ActiveConversations, conversationID)
			conn.mu.Unlock()
		}
	}
	cm.mu.Unlock()
	
	logger.Info("User left conversation",
		"user_id", userID,
		"conversation_id", conversationID,
	)
	
	return nil
}

// BroadcastToConversation sends a message to all participants in a conversation
func (cm *ConnectionManager) BroadcastToConversation(conversationID string, message Message) error {
	cm.chatMu.RLock()
	room, exists := cm.chatRooms[conversationID]
	cm.chatMu.RUnlock()
	
	if !exists {
		return fmt.Errorf("conversation room not found: %s", conversationID)
	}
	
	room.mu.RLock()
	participants := make([]string, 0, len(room.Participants))
	for userID := range room.Participants {
		participants = append(participants, userID)
	}
	room.mu.RUnlock()
	
	// Send to all participants
	sentCount := 0
	cm.mu.RLock()
	for _, conn := range cm.connections {
		if conn.IsAlive && conn.isParticipantInConversation(conversationID) {
			err := conn.WriteMessage(message)
			if err != nil {
				logger.Error("Failed to send message to conversation participant", err,
					"connection_id", conn.UserID,
					"conversation_id", conversationID,
				)
			} else {
				sentCount++
			}
		}
	}
	cm.mu.RUnlock()
	
	logger.Debug("Message broadcast to conversation",
		"conversation_id", conversationID,
		"participants_count", len(participants),
		"sent_count", sentCount,
		"message_type", message.Type,
	)
	
	return nil
}

// SetTyping sets typing indicator for a user in a conversation
func (cm *ConnectionManager) SetTyping(userID, conversationID string, isTyping bool) error {
	cm.typingMu.Lock()
	defer cm.typingMu.Unlock()
	
	// Initialize conversation typing map if needed
	if _, exists := cm.typingUsers[conversationID]; !exists {
		cm.typingUsers[conversationID] = make(map[string]time.Time)
	}
	
	if isTyping {
		cm.typingUsers[conversationID][userID] = time.Now()
	} else {
		delete(cm.typingUsers[conversationID], userID)
	}
	
	// Broadcast typing indicator
	indicator := TypingIndicator{
		ConversationID: conversationID,
		UserID:        userID,
		IsTyping:      isTyping,
		Timestamp:     time.Now(),
	}
	
	message := Message{
		Type:      "typing:indicator",
		Data:      indicator,
		Timestamp: time.Now(),
		SenderID:  userID,
	}
	
	return cm.BroadcastToConversation(conversationID, message)
}

// GetTypingUsers returns list of users currently typing in a conversation
func (cm *ConnectionManager) GetTypingUsers(conversationID string) []string {
	cm.typingMu.RLock()
	defer cm.typingMu.RUnlock()
	
	if conversationTyping, exists := cm.typingUsers[conversationID]; exists {
		// Clean up old typing indicators (older than 5 seconds)
		now := time.Now()
		typingUsers := make([]string, 0)
		
		for userID, lastTyping := range conversationTyping {
			if now.Sub(lastTyping) < 5*time.Second {
				typingUsers = append(typingUsers, userID)
			} else {
				delete(conversationTyping, userID)
			}
		}
		
		return typingUsers
	}
	
	return []string{}
}

// SetUserOnline sets user's online status
func (cm *ConnectionManager) SetUserOnline(userID string, isOnline bool) {
	cm.setUserOnline(userID, isOnline)
}

// setUserOnline sets user's online status and broadcasts update
func (cm *ConnectionManager) setUserOnline(userID string, isOnline bool) {
	status := OnlineStatus{
		UserID:    userID,
		IsOnline:  isOnline,
		Timestamp: time.Now(),
	}
	
	message := Message{
		Type:      "user:status",
		Data:      status,
		Timestamp: time.Now(),
	}
	
	// Broadcast to all connections
	cm.BroadcastToAll(message)
	
	// Update online status in cache
	if cm.sessionMgr != nil {
		if isOnline {
			cm.sessionMgr.SetUserOnline(userID)
		} else {
			cm.sessionMgr.SetUserOffline(userID)
		}
	}
	
	logger.Info("User online status updated",
		"user_id", userID,
		"is_online", isOnline,
	)
}

// UpdateUnreadCount updates and broadcasts unread count for a user
func (cm *ConnectionManager) UpdateUnreadCount(userID, conversationID string, count int) error {
	unreadCount := UnreadCount{
		ConversationID: conversationID,
		UserID:        userID,
		Count:         count,
	}
	
	message := Message{
		Type:      "conversation:unread",
		Data:      unreadCount,
		Timestamp: time.Now(),
	}
	
	return cm.BroadcastToUser(userID, message)
}

// GetConversationParticipants returns all participants in a conversation
func (cm *ConnectionManager) GetConversationParticipants(conversationID string) []string {
	cm.chatMu.RLock()
	defer cm.chatMu.RUnlock()
	
	if room, exists := cm.chatRooms[conversationID]; exists {
		room.mu.RLock()
		defer room.mu.RUnlock()
		
		participants := make([]string, 0, len(room.Participants))
		for userID := range room.Participants {
			participants = append(participants, userID)
		}
		
		return participants
	}
	
	return []string{}
}

// isParticipantInConversation checks if user is participant in conversation
func (conn *ClientConnection) isParticipantInConversation(conversationID string) bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	
	return conn.ActiveConversations[conversationID]
}

// hasActiveConnections checks if user has any active connections
func (cm *ConnectionManager) hasActiveConnections(userID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	for _, conn := range cm.connections {
		if conn.UserID == userID && conn.IsAlive {
			return true
		}
	}
	
	return false
}

// CleanupExpiredTypingIndicators removes expired typing indicators
func (cm *ConnectionManager) CleanupExpiredTypingIndicators() {
	cm.typingMu.Lock()
	defer cm.typingMu.Unlock()
	
	now := time.Now()
	expiredThreshold := 5 * time.Second
	
	for conversationID, conversationTyping := range cm.typingUsers {
		for userID, lastTyping := range conversationTyping {
			if now.Sub(lastTyping) > expiredThreshold {
				delete(conversationTyping, userID)
			}
		}
		
		// Remove empty conversation typing maps
		if len(conversationTyping) == 0 {
			delete(cm.typingUsers, conversationID)
		}
	}
}

// CleanupInactiveChatRooms removes inactive chat rooms
func (cm *ConnectionManager) CleanupInactiveChatRooms() {
	cm.chatMu.Lock()
	defer cm.chatMu.Unlock()
	
	now := time.Now()
	inactiveThreshold := 30 * time.Minute
	
	for conversationID, room := range cm.chatRooms {
		room.mu.RLock()
		lastActivity := room.LastActivity
		participantCount := len(room.Participants)
		room.mu.RUnlock()
		
		if participantCount == 0 || now.Sub(lastActivity) > inactiveThreshold {
			delete(cm.chatRooms, conversationID)
			logger.Info("Inactive conversation room removed",
				"conversation_id", conversationID,
				"inactive_duration", now.Sub(lastActivity),
				"participant_count", participantCount,
			)
		}
	}
}