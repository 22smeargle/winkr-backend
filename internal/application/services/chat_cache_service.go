package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ChatCacheService provides caching functionality for chat system
type ChatCacheService struct {
	cache *cache.CacheService
}

// NewChatCacheService creates a new chat cache service
func NewChatCacheService(cache *cache.CacheService) *ChatCacheService {
	return &ChatCacheService{
		cache: cache,
	}
}

// CacheKey represents different cache key types
type CacheKey string

const (
	// User-related keys
	UserOnlineStatusKey    CacheKey = "user:online:%s"
	UserTypingStatusKey   CacheKey = "user:typing:%s:%s"
	UserUnreadCountKey    CacheKey = "user:unread:%s:%s"
	UserBlockedUsersKey    CacheKey = "user:blocked:%s"
	UserLastSeenKey       CacheKey = "user:last_seen:%s"
	
	// Conversation-related keys
	ConversationMessagesKey CacheKey = "conversation:messages:%s"
	ConversationParticipantsKey CacheKey = "conversation:participants:%s"
	ConversationLastMessageKey CacheKey = "conversation:last_message:%s"
	ConversationTypingUsersKey CacheKey = "conversation:typing_users:%s"
	
	// Message-related keys
	MessageKey             CacheKey = "message:%s"
	MessageAnalyticsKey    CacheKey = "message:analytics:%s"
	MessageSecurityKey    CacheKey = "message:security:%s"
	
	// System keys
	OnlineUsersKey         CacheKey = "system:online_users"
	ActiveConversationsKey CacheKey = "system:active_conversations"
	ChatStatsKey           CacheKey = "system:chat_stats"
)

// CacheTTL represents cache TTL values
const (
	ShortTTL  = 5 * time.Minute   // 5 minutes
	MediumTTL = 30 * time.Minute  // 30 minutes
	LongTTL    = 2 * time.Hour    // 2 hours
	DayTTL     = 24 * time.Hour   // 24 hours
	WeekTTL    = 7 * 24 * time.Hour // 1 week
)

// CacheUserOnlineStatus caches user's online status
func (s *ChatCacheService) CacheUserOnlineStatus(ctx context.Context, userID string, isOnline bool) error {
	key := fmt.Sprintf(string(UserOnlineStatusKey), userID)
	return s.cache.Set(ctx, key, isOnline, MediumTTL)
}

// GetUserOnlineStatus retrieves user's online status from cache
func (s *ChatCacheService) GetUserOnlineStatus(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf(string(UserOnlineStatusKey), userID)
	var isOnline bool
	err := s.cache.Get(ctx, key, &isOnline)
	if err != nil {
		return false, fmt.Errorf("failed to get user online status: %w", err)
	}
	return isOnline, nil
}

// CacheUserTypingStatus caches user's typing status in a conversation
func (s *ChatCacheService) CacheUserTypingStatus(ctx context.Context, userID, conversationID string, isTyping bool) error {
	key := fmt.Sprintf(string(UserTypingStatusKey), userID, conversationID)
	return s.cache.Set(ctx, key, isTyping, ShortTTL)
}

// GetUserTypingStatus retrieves user's typing status from cache
func (s *ChatCacheService) GetUserTypingStatus(ctx context.Context, userID, conversationID string) (bool, error) {
	key := fmt.Sprintf(string(UserTypingStatusKey), userID, conversationID)
	var isTyping bool
	err := s.cache.Get(ctx, key, &isTyping)
	if err != nil {
		return false, fmt.Errorf("failed to get user typing status: %w", err)
	}
	return isTyping, nil
}

// CacheUserUnreadCount caches unread message count for a user in a conversation
func (s *ChatCacheService) CacheUserUnreadCount(ctx context.Context, userID, conversationID string, count int) error {
	key := fmt.Sprintf(string(UserUnreadCountKey), userID, conversationID)
	return s.cache.Set(ctx, key, count, LongTTL)
}

// GetUserUnreadCount retrieves unread message count from cache
func (s *ChatCacheService) GetUserUnreadCount(ctx context.Context, userID, conversationID string) (int, error) {
	key := fmt.Sprintf(string(UserUnreadCountKey), userID, conversationID)
	var count int
	err := s.cache.Get(ctx, key, &count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user unread count: %w", err)
	}
	return count, nil
}

// IncrementUserUnreadCount increments unread message count for a user
func (s *ChatCacheService) IncrementUserUnreadCount(ctx context.Context, userID, conversationID string) error {
	key := fmt.Sprintf(string(UserUnreadCountKey), userID, conversationID)
	
	// Get current count
	currentCount, err := s.GetUserUnreadCount(ctx, userID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get current unread count: %w", err)
	}
	
	// Increment and cache
	newCount := currentCount + 1
	return s.cache.Set(ctx, key, newCount, LongTTL)
}

// CacheConversationMessages caches messages for a conversation
func (s *ChatCacheService) CacheConversationMessages(ctx context.Context, conversationID string, messages []*entities.Message) error {
	key := fmt.Sprintf(string(ConversationMessagesKey), conversationID)
	
	// Serialize messages
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	
	return s.cache.Set(ctx, key, string(data), MediumTTL)
}

// GetConversationMessages retrieves cached messages for a conversation
func (s *ChatCacheService) GetConversationMessages(ctx context.Context, conversationID string) ([]*entities.Message, error) {
	key := fmt.Sprintf(string(ConversationMessagesKey), conversationID)
	
	var data string
	err := s.cache.Get(ctx, key, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	
	// Deserialize messages
	var messages []*entities.Message
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}
	
	return messages, nil
}

// AddMessageToConversation adds a message to conversation cache
func (s *ChatCacheService) AddMessageToConversation(ctx context.Context, conversationID string, message *entities.Message) error {
	key := fmt.Sprintf(string(ConversationMessagesKey), conversationID)
	
	// Get current messages
	messages, err := s.GetConversationMessages(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get current messages: %w", err)
	}
	
	// Add new message
	messages = append(messages, message)
	
	// Keep only last 100 messages
	if len(messages) > 100 {
		messages = messages[len(messages)-100:]
	}
	
	// Cache updated messages
	return s.CacheConversationMessages(ctx, conversationID, messages)
}

// CacheConversationParticipants caches participants for a conversation
func (s *ChatCacheService) CacheConversationParticipants(ctx context.Context, conversationID string, participants []string) error {
	key := fmt.Sprintf(string(ConversationParticipantsKey), conversationID)
	return s.cache.Set(ctx, key, participants, LongTTL)
}

// GetConversationParticipants retrieves cached participants for a conversation
func (s *ChatCacheService) GetConversationParticipants(ctx context.Context, conversationID string) ([]string, error) {
	key := fmt.Sprintf(string(ConversationParticipantsKey), conversationID)
	
	var participants []string
	err := s.cache.Get(ctx, key, &participants)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation participants: %w", err)
	}
	
	return participants, nil
}

// CacheConversationLastMessage caches the last message for a conversation
func (s *ChatCacheService) CacheConversationLastMessage(ctx context.Context, conversationID string, message *entities.Message) error {
	key := fmt.Sprintf(string(ConversationLastMessageKey), conversationID)
	return s.cache.Set(ctx, key, message, MediumTTL)
}

// GetConversationLastMessage retrieves the last message for a conversation
func (s *ChatCacheService) GetConversationLastMessage(ctx context.Context, conversationID string) (*entities.Message, error) {
	key := fmt.Sprintf(string(ConversationLastMessageKey), conversationID)
	
	var message *entities.Message
	err := s.cache.Get(ctx, key, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation last message: %w", err)
	}
	
	return message, nil
}

// CacheConversationTypingUsers caches typing users for a conversation
func (s *ChatCacheService) CacheConversationTypingUsers(ctx context.Context, conversationID string, typingUsers map[string]time.Time) error {
	key := fmt.Sprintf(string(ConversationTypingUsersKey), conversationID)
	return s.cache.Set(ctx, key, typingUsers, ShortTTL)
}

// GetConversationTypingUsers retrieves typing users for a conversation
func (s *ChatCacheService) GetConversationTypingUsers(ctx context.Context, conversationID string) (map[string]time.Time, error) {
	key := fmt.Sprintf(string(ConversationTypingUsersKey), conversationID)
	
	var typingUsers map[string]time.Time
	err := s.cache.Get(ctx, key, &typingUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation typing users: %w", err)
	}
	
	return typingUsers, nil
}

// CacheMessage caches a single message
func (s *ChatCacheService) CacheMessage(ctx context.Context, message *entities.Message) error {
	key := fmt.Sprintf(string(MessageKey), message.ID.String())
	return s.cache.Set(ctx, key, message, MediumTTL)
}

// GetMessage retrieves a cached message
func (s *ChatCacheService) GetMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error) {
	key := fmt.Sprintf(string(MessageKey), messageID.String())
	
	var message *entities.Message
	err := s.cache.Get(ctx, key, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	return message, nil
}

// DeleteMessage removes a message from cache
func (s *ChatCacheService) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	key := fmt.Sprintf(string(MessageKey), messageID.String())
	return s.cache.Delete(ctx, key)
}

// CacheOnlineUsers caches the list of online users
func (s *ChatCacheService) CacheOnlineUsers(ctx context.Context, onlineUsers []string) error {
	return s.cache.Set(ctx, string(OnlineUsersKey), onlineUsers, ShortTTL)
}

// GetOnlineUsers retrieves the list of online users
func (s *ChatCacheService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	var onlineUsers []string
	err := s.cache.Get(ctx, string(OnlineUsersKey), &onlineUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %w", err)
	}
	
	return onlineUsers, nil
}

// AddOnlineUser adds a user to the online users list
func (s *ChatCacheService) AddOnlineUser(ctx context.Context, userID string) error {
	// Get current online users
	onlineUsers, err := s.GetOnlineUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current online users: %w", err)
	}
	
	// Add user if not already present
	for _, onlineUser := range onlineUsers {
		if onlineUser == userID {
			return nil // Already online
		}
	}
	
	onlineUsers = append(onlineUsers, userID)
	
	// Cache updated list
	return s.CacheOnlineUsers(ctx, onlineUsers)
}

// RemoveOnlineUser removes a user from the online users list
func (s *ChatCacheService) RemoveOnlineUser(ctx context.Context, userID string) error {
	// Get current online users
	onlineUsers, err := s.GetOnlineUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current online users: %w", err)
	}
	
	// Remove user if present
	filteredUsers := make([]string, 0, len(onlineUsers))
	for _, onlineUser := range onlineUsers {
		if onlineUser != userID {
			filteredUsers = append(filteredUsers, onlineUser)
		}
	}
	
	// Cache updated list
	return s.CacheOnlineUsers(ctx, filteredUsers)
}

// CacheChatStats caches chat system statistics
func (s *ChatCacheService) CacheChatStats(ctx context.Context, stats map[string]interface{}) error {
	return s.cache.Set(ctx, string(ChatStatsKey), stats, MediumTTL)
}

// GetChatStats retrieves cached chat system statistics
func (s *ChatCacheService) GetChatStats(ctx context.Context) (map[string]interface{}, error) {
	var stats map[string]interface{}
	err := s.cache.Get(ctx, string(ChatStatsKey), &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat stats: %w", err)
	}
	
	return stats, nil
}

// InvalidateConversationCache invalidates all cache entries for a conversation
func (s *ChatCacheService) InvalidateConversationCache(ctx context.Context, conversationID string) error {
	keys := []string{
		fmt.Sprintf(string(ConversationMessagesKey), conversationID),
		fmt.Sprintf(string(ConversationParticipantsKey), conversationID),
		fmt.Sprintf(string(ConversationLastMessageKey), conversationID),
		fmt.Sprintf(string(ConversationTypingUsersKey), conversationID),
	}
	
	for _, key := range keys {
		if err := s.cache.Delete(ctx, key); err != nil {
			logger.Error("Failed to delete cache key", err, "key", key)
		}
	}
	
	return nil
}

// InvalidateUserCache invalidates all cache entries for a user
func (s *ChatCacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	// Get all keys for user
	pattern := fmt.Sprintf("*%s*", userID)
	keys, err := s.cache.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get user cache keys: %w", err)
	}
	
	// Delete all user-related keys
	for _, key := range keys {
		if err := s.cache.Delete(ctx, key); err != nil {
			logger.Error("Failed to delete user cache key", err, "key", key)
		}
	}
	
	return nil
}

// CleanupExpiredEntries removes expired cache entries
func (s *ChatCacheService) CleanupExpiredEntries(ctx context.Context) error {
	// This would typically be handled by Redis TTL automatically
	// But we can implement additional cleanup logic if needed
	logger.Info("Cache cleanup completed")
	return nil
}

// GetCacheStats returns cache statistics
func (s *ChatCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Get cache info from Redis
	info, err := s.cache.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache info: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_keys":        info["total_keys"],
		"expired_keys":      info["expired_keys"],
		"avg_ttl":          info["avg_ttl"],
		"memory_usage":      info["memory_usage"],
		"hit_rate":         info["hit_rate"],
		"miss_rate":        info["miss_rate"],
		"last_cleanup":      time.Now(),
	}
	
	return stats, nil
}