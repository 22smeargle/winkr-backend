package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MessageRepository defines interface for message data operations
type MessageRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, message *entities.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error)
	Update(ctx context.Context, message *entities.Message) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Conversation operations
	GetConversation(ctx context.Context, conversationID uuid.UUID) (*entities.Conversation, error)
	GetConversationByMatchID(ctx context.Context, matchID uuid.UUID) (*entities.Conversation, error)
	CreateConversation(ctx context.Context, conversation *entities.Conversation) error
	UpdateConversation(ctx context.Context, conversation *entities.Conversation) error

	// Message operations
	GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*entities.Message, error)
	GetMessagesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Message, error)
	GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entities.Message, error)
	GetRecentMessages(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.Message, error)

	// Message status operations
	MarkAsRead(ctx context.Context, messageID uuid.UUID) error
	MarkConversationAsRead(ctx context.Context, conversationID, userID uuid.UUID) error
	SoftDeleteMessage(ctx context.Context, messageID uuid.UUID) error
	RestoreMessage(ctx context.Context, messageID uuid.UUID) error

	// User conversation operations
	GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Conversation, error)
	GetUserConversationsWithUnreadCount(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ConversationWithUnread, error)
	GetConversationWithMessages(ctx context.Context, conversationID, userID uuid.UUID, limit, offset int) (*entities.Conversation, error)

	// Search operations
	SearchMessages(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*entities.Message, error)
	SearchConversations(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*entities.Conversation, error)

	// Batch operations
	BatchCreate(ctx context.Context, messages []*entities.Message) error
	BatchMarkAsRead(ctx context.Context, messageIDs []uuid.UUID) error
	BatchSoftDelete(ctx context.Context, messageIDs []uuid.UUID) error

	// Existence checks
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	UserCanAccessMessage(ctx context.Context, userID, messageID uuid.UUID) (bool, error)
	UserCanAccessConversation(ctx context.Context, userID, conversationID uuid.UUID) (bool, error)

	// Analytics and statistics
	GetMessageStats(ctx context.Context, userID uuid.UUID) (*MessageStats, error)
	GetConversationStats(ctx context.Context, conversationID uuid.UUID) (*ConversationStats, error)
	GetMessagesSentInRange(ctx context.Context, userID uuid.UUID, startDate, endDate interface{}) (int64, error)

	// Admin operations
	GetAllMessages(ctx context.Context, limit, offset int) ([]*entities.Message, error)
	GetAllConversations(ctx context.Context, limit, offset int) ([]*entities.Conversation, error)
	GetDeletedMessages(ctx context.Context, limit, offset int) ([]*entities.Message, error)

	// Advanced queries
	GetMessagesByType(ctx context.Context, conversationID uuid.UUID, messageType string, limit, offset int) ([]*entities.Message, error)
	GetMessagesBeforeDate(ctx context.Context, conversationID uuid.UUID, beforeDate interface{}, limit int) ([]*entities.Message, error)
	GetLastMessage(ctx context.Context, conversationID uuid.UUID) (*entities.Message, error)
	GetUnreadMessageCount(ctx context.Context, userID uuid.UUID) (int, error)
}

// ConversationWithUnread represents a conversation with unread count
type ConversationWithUnread struct {
	*entities.Conversation
	UnreadCount int `json:"unread_count"`
}

// MessageStats represents message statistics for a user
type MessageStats struct {
	TotalMessages     int64 `json:"total_messages"`
	SentMessages      int64 `json:"sent_messages"`
	ReceivedMessages  int64 `json:"received_messages"`
	UnreadMessages   int64 `json:"unread_messages"`
	DeletedMessages   int64 `json:"deleted_messages"`
	Conversations     int64 `json:"conversations"`
	MessagesToday     int64 `json:"messages_today"`
	MessagesThisWeek  int64 `json:"messages_this_week"`
	MessagesThisMonth int64 `json:"messages_this_month"`
}

// ConversationStats represents conversation statistics
type ConversationStats struct {
	TotalMessages    int64 `json:"total_messages"`
	MessagesByUser1  int64 `json:"messages_by_user1"`
	MessagesByUser2  int64 `json:"messages_by_user2"`
	UnreadMessages   int64 `json:"unread_messages"`
	LastMessageTime  interface{} `json:"last_message_time"`
	ConversationAge  int   `json:"conversation_age_days"`
}