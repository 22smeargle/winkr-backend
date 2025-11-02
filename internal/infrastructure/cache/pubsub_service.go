package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// PubSubService handles Redis Pub/Sub operations
type PubSubService struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewPubSubService creates a new Pub/Sub service
func NewPubSubService(redisClient *redis.RedisClient) *PubSubService {
	return &PubSubService{
		redisClient: redisClient,
		prefix:      "pubsub:",
	}
}

// MessageType represents different types of real-time messages
type MessageType string

const (
	MessageTypeChat        MessageType = "chat"
	MessageTypeNotification MessageType = "notification"
	MessageTypeOnlineStatus MessageType = "online_status"
	MessageTypeMatch       MessageType = "match"
	MessageTypeTyping     MessageType = "typing"
)

// Message represents a real-time message
type Message struct {
	Type      MessageType      `json:"type"`
	Channel   string           `json:"channel"`
	Data      interface{}      `json:"data"`
	Timestamp time.Time        `json:"timestamp"`
	SenderID  string           `json:"sender_id,omitempty"`
	RecipientID string          `json:"recipient_id,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ConversationID string    `json:"conversation_id"`
	SenderID      string    `json:"sender_id"`
	Content       string    `json:"content"`
	MessageType   string    `json:"message_type"`
	Timestamp     time.Time `json:"timestamp"`
}

// NotificationMessage represents a notification
type NotificationMessage struct {
	UserID    string `json:"user_id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Read       bool   `json:"read"`
}

// OnlineStatusMessage represents online status update
type OnlineStatusMessage struct {
	UserID    string    `json:"user_id"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen"`
	Location   string    `json:"location,omitempty"`
}

// MatchMessage represents a match notification
type MatchMessage struct {
	UserID    string    `json:"user_id"`
	MatchID    string    `json:"match_id"`
	MatchedWith string    `json:"matched_with"`
	Timestamp  time.Time `json:"timestamp"`
}

// TypingMessage represents typing indicator
type TypingMessage struct {
	ConversationID string    `json:"conversation_id"`
	SenderID      string    `json:"sender_id"`
	IsTyping      bool      `json:"is_typing"`
	Timestamp     time.Time `json:"timestamp"`
}

// PublishChatMessage publishes a chat message
func (ps *PubSubService) PublishChatMessage(ctx context.Context, conversationID, senderID, content, messageType string) error {
	chatMsg := ChatMessage{
		ConversationID: conversationID,
		SenderID:      senderID,
		Content:       content,
		MessageType:   messageType,
		Timestamp:     time.Now(),
	}

	channel := ps.getChatChannel(conversationID)
	return ps.publishMessage(ctx, MessageTypeChat, channel, chatMsg)
}

// PublishNotification publishes a notification
func (ps *PubSubService) PublishNotification(ctx context.Context, userID, notificationType, title, message string, data interface{}) error {
	notification := NotificationMessage{
		UserID:   userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		Read:      false,
	}

	channel := ps.getNotificationChannel(userID)
	return ps.publishMessage(ctx, MessageTypeNotification, channel, notification)
}

// PublishOnlineStatus publishes online status update
func (ps *PubSubService) PublishOnlineStatus(ctx context.Context, userID string, isOnline bool, location string) error {
	onlineStatus := OnlineStatusMessage{
		UserID:  userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
		Location: location,
	}

	channel := ps.getOnlineStatusChannel()
	return ps.publishMessage(ctx, MessageTypeOnlineStatus, channel, onlineStatus)
}

// PublishMatch publishes a match notification
func (ps *PubSubService) PublishMatch(ctx context.Context, userID, matchID, matchedWith string) error {
	matchMsg := MatchMessage{
		UserID:     userID,
		MatchID:    matchID,
		MatchedWith: matchedWith,
		Timestamp:   time.Now(),
	}

	channel := ps.getMatchChannel(userID)
	return ps.publishMessage(ctx, MessageTypeMatch, channel, matchMsg)
}

// PublishTyping publishes typing indicator
func (ps *PubSubService) PublishTyping(ctx context.Context, conversationID, senderID string, isTyping bool) error {
	typingMsg := TypingMessage{
		ConversationID: conversationID,
		SenderID:      senderID,
		IsTyping:      isTyping,
		Timestamp:     time.Now(),
	}

	channel := ps.getTypingChannel(conversationID)
	return ps.publishMessage(ctx, MessageTypeTyping, channel, typingMsg)
}

// SubscribeToChat subscribes to chat messages
func (ps *PubSubService) SubscribeToChat(ctx context.Context, conversationID string) (<-chan Message, error) {
	channel := ps.getChatChannel(conversationID)
	return ps.subscribeToChannel(ctx, channel)
}

// SubscribeToNotifications subscribes to user notifications
func (ps *PubSubService) SubscribeToNotifications(ctx context.Context, userID string) (<-chan Message, error) {
	channel := ps.getNotificationChannel(userID)
	return ps.subscribeToChannel(ctx, channel)
}

// SubscribeToOnlineStatus subscribes to online status updates
func (ps *PubSubService) SubscribeToOnlineStatus(ctx context.Context) (<-chan Message, error) {
	channel := ps.getOnlineStatusChannel()
	return ps.subscribeToChannel(ctx, channel)
}

// SubscribeToMatches subscribes to match notifications
func (ps *PubSubService) SubscribeToMatches(ctx context.Context, userID string) (<-chan Message, error) {
	channel := ps.getMatchChannel(userID)
	return ps.subscribeToChannel(ctx, channel)
}

// SubscribeToTyping subscribes to typing indicators
func (ps *PubSubService) SubscribeToTyping(ctx context.Context, conversationID string) (<-chan Message, error) {
	channel := ps.getTypingChannel(conversationID)
	return ps.subscribeToChannel(ctx, channel)
}

// BroadcastToUsers broadcasts a message to multiple users
func (ps *PubSubService) BroadcastToUsers(ctx context.Context, userIDs []string, message Message) error {
	for _, userID := range userIDs {
		switch message.Type {
		case MessageTypeNotification:
			channel := ps.getNotificationChannel(userID)
			err := ps.publishMessageToChannel(ctx, channel, message)
			if err != nil {
				logger.Error("Failed to broadcast notification to user", err, "user_id", userID)
			}
		case MessageTypeChat:
			// For chat messages, you might want to target specific conversations
			logger.Debug("Broadcasting chat message to user", "user_id", userID)
		default:
			logger.Warn("Unknown message type for broadcasting", "type", message.Type)
		}
	}
	return nil
}

// GetActiveSubscriptions returns information about active subscriptions
func (ps *PubSubService) GetActiveSubscriptions(ctx context.Context) (map[string]interface{}, error) {
	// This would typically track active subscriptions
	// For simplicity, we'll return basic info
	
	info := map[string]interface{}{
		"prefix": ps.prefix,
		"channels": map[string]string{
			"chat_pattern":        ps.getChatChannelPattern(),
			"notification_pattern":  ps.getNotificationChannelPattern(),
			"online_status":       ps.getOnlineStatusChannel(),
			"match_pattern":       ps.getMatchChannelPattern(),
			"typing_pattern":       ps.getTypingChannelPattern(),
		},
	}
	
	return info, nil
}

// publishMessage publishes a message to a specific channel
func (ps *PubSubService) publishMessage(ctx context.Context, msgType MessageType, channel string, data interface{}) error {
	message := Message{
		Type:      msgType,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message for publishing", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = ps.redisClient.Publish(ctx, channel, string(messageData))
	if err != nil {
		logger.Error("Failed to publish message", err, "channel", channel, "type", msgType)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Debug("Message published", "channel", channel, "type", msgType)
	return nil
}

// publishMessageToChannel publishes a message to a specific channel
func (ps *PubSubService) publishMessageToChannel(ctx context.Context, channel string, message Message) error {
	messageData, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message for channel publishing", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = ps.redisClient.Publish(ctx, channel, string(messageData))
	if err != nil {
		logger.Error("Failed to publish message to channel", err, "channel", channel)
		return fmt.Errorf("failed to publish message to channel: %w", err)
	}

	return nil
}

// subscribeToChannel subscribes to a channel and returns a message channel
func (ps *PubSubService) subscribeToChannel(ctx context.Context, channel string) (<-chan Message, error) {
	pubsub := ps.redisClient.Subscribe(ctx, channel)
	if pubsub == nil {
		return nil, fmt.Errorf("failed to subscribe to channel: %s", channel)
	}

	msgChan := make(chan Message, 100)
	
	// Start goroutine to handle messages
	go func() {
		defer close(msgChan)
		defer pubsub.Close()
		
		for {
			select {
			case <-ctx.Done():
				logger.Debug("Subscription context cancelled", "channel", channel)
				return
			case msg := <-pubsub.Channel():
				if msg.Payload == "" {
					continue // Skip empty messages
				}
				
				var message Message
				err := json.Unmarshal([]byte(msg.Payload), &message)
				if err != nil {
					logger.Error("Failed to unmarshal received message", err)
					continue
				}
				
				msgChan <- message
			}
		}
	}()

	return msgChan, nil
}

// Helper methods for channel generation

func (ps *PubSubService) getChatChannel(conversationID string) string {
	return fmt.Sprintf("%schat:%s", ps.prefix, conversationID)
}

func (ps *PubSubService) getNotificationChannel(userID string) string {
	return fmt.Sprintf("%snotifications:%s", ps.prefix, userID)
}

func (ps *PubSubService) getOnlineStatusChannel() string {
	return fmt.Sprintf("%sonline_status", ps.prefix)
}

func (ps *PubSubService) getMatchChannel(userID string) string {
	return fmt.Sprintf("%smatches:%s", ps.prefix, userID)
}

func (ps *PubSubService) getTypingChannel(conversationID string) string {
	return fmt.Sprintf("%styping:%s", ps.prefix, conversationID)
}

// Pattern methods for wildcard subscriptions

func (ps *PubSubService) getChatChannelPattern() string {
	return fmt.Sprintf("%schat:*", ps.prefix)
}

func (ps *PubSubService) getNotificationChannelPattern() string {
	return fmt.Sprintf("%snotifications:*", ps.prefix)
}

func (ps *PubSubService) getMatchChannelPattern() string {
	return fmt.Sprintf("%smatches:*", ps.prefix)
}

func (ps *PubSubService) getTypingChannelPattern() string {
	return fmt.Sprintf("%styping:*", ps.prefix)
}