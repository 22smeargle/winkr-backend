package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message represents a message entity in conversations in database
type Message struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"sender_id"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	MessageType    string     `gorm:"default:'text';check:message_type IN ('text', 'image', 'gif', 'ephemeral_photo')" json:"message_type"`
	IsRead         bool       `gorm:"default:false;index" json:"is_read"`
	IsDeleted      bool       `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"created_at"`

	// Relationships
	Sender       *User         `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sender,omitempty"`
	Conversation *Conversation `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"conversation,omitempty"`
}

// TableName returns the table name for Message model
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate GORM hook
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
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

// Conversation represents a conversation between matched users in database
type Conversation struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MatchID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"match_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Match    *Match     `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"match,omitempty"`
	Messages []*Message `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// TableName returns the table name for Conversation model
func (Conversation) TableName() string {
	return "conversations"
}

// BeforeCreate GORM hook
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
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