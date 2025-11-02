package entities

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message entity in conversations
type Message struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ConversationID uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID  `json:"sender_id" gorm:"type:uuid;not null;index"`
	Content        string     `json:"content" gorm:"type:text;not null"`
	MessageType    string     `json:"message_type" gorm:"default:'text';check:message_type IN ('text', 'image', 'gif', 'ephemeral_photo')"`
	IsRead         bool       `json:"is_read" gorm:"default:false"`
	IsDeleted      bool       `json:"is_deleted" gorm:"default:false"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Sender       *User         `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignKey:ConversationID"`
}

// TableName returns the table name for Message entity
func (Message) TableName() string {
	return "messages"
}

// IsText returns true if the message is a text message
func (m *Message) IsText() bool {
	return m.MessageType == "text"
}

// IsImage returns true if the message is an image message
func (m *Message) IsImage() bool {
	return m.MessageType == "image"
}

// IsGif returns true if the message is a GIF message
func (m *Message) IsGif() bool {
	return m.MessageType == "gif"
}

// IsEphemeralPhoto returns true if the message is an ephemeral photo message
func (m *Message) IsEphemeralPhoto() bool {
	return m.MessageType == "ephemeral_photo"
}

// MarkAsRead marks the message as read
func (m *Message) MarkAsRead() {
	m.IsRead = true
}

// SoftDelete marks the message as deleted
func (m *Message) SoftDelete() {
	m.IsDeleted = true
}

// Restore restores a soft-deleted message
func (m *Message) Restore() {
	m.IsDeleted = false
}

// CanBeEdited returns true if the message can be edited
func (m *Message) CanBeEdited() bool {
	// Messages can only be edited within 15 minutes of creation
	return m.IsText() && !m.IsDeleted && time.Since(m.CreatedAt) < 15*time.Minute
}

// CanBeDeleted returns true if the message can be deleted
func (m *Message) CanBeDeleted() bool {
	return !m.IsDeleted
}

// Conversation represents a conversation between matched users
type Conversation struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MatchID   uuid.UUID  `json:"match_id" gorm:"type:uuid;not null;uniqueIndex"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Match    *Match     `json:"match,omitempty" gorm:"foreignKey:MatchID"`
	Messages []*Message `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// TableName returns the table name for Conversation entity
func (Conversation) TableName() string {
	return "conversations"
}

// GetLastMessage returns the last message in the conversation
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return c.Messages[len(c.Messages)-1]
}

// GetUnreadCount returns the count of unread messages for a specific user
func (c *Conversation) GetUnreadCount(userID uuid.UUID) int {
	count := 0
	for _, message := range c.Messages {
		if message.SenderID != userID && !message.IsRead && !message.IsDeleted {
			count++
		}
	}
	return count
}

// HasMessages returns true if the conversation has messages
func (c *Conversation) HasMessages() bool {
	return len(c.Messages) > 0
}