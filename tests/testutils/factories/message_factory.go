package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
)

// MessageFactory creates test message entities
type MessageFactory struct{}

// NewMessageFactory creates a new message factory
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{}
}

// CreateMessage creates a test message with default values
func (f *MessageFactory) CreateMessage() *entities.Message {
	now := time.Now()
	messageID := uuid.New()
	senderID := uuid.New()
	receiverID := uuid.New()
	conversationID := uuid.New()
	
	return &entities.Message{
		ID:             messageID,
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        "Hello, this is a test message!",
		MessageType:    entities.MessageTypeText,
		IsRead:         false,
		IsDeleted:      false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// CreateTextMessage creates a text message
func (f *MessageFactory) CreateTextMessage(content string) *entities.Message {
	message := f.CreateMessage()
	message.Content = content
	message.MessageType = entities.MessageTypeText
	return message
}

// CreatePhotoMessage creates a photo message
func (f *MessageFactory) CreatePhotoMessage(photoURL string) *entities.Message {
	message := f.CreateMessage()
	message.Content = photoURL
	message.MessageType = entities.MessageTypePhoto
	return message
}

// CreateEphemeralPhotoMessage creates an ephemeral photo message
func (f *MessageFactory) CreateEphemeralPhotoMessage(photoURL string, duration int) *entities.Message {
	message := f.CreateMessage()
	message.Content = photoURL
	message.MessageType = entities.MessageTypeEphemeralPhoto
	message.EphemeralPhotoDuration = duration
	return message
}

// CreateLocationMessage creates a location message
func (f *MessageFactory) CreateLocationMessage(lat, lng float64, address string) *entities.Message {
	message := f.CreateMessage()
	message.Content = address
	message.MessageType = entities.MessageTypeLocation
	message.Location = &entities.Location{
		Latitude:  lat,
		Longitude: lng,
		Address:   address,
	}
	return message
}

// CreateSystemMessage creates a system message
func (f *MessageFactory) CreateSystemMessage(content string) *entities.Message {
	message := f.CreateMessage()
	message.Content = content
	message.MessageType = entities.MessageTypeSystem
	return message
}

// CreateReadMessage creates a read message
func (f *MessageFactory) CreateReadMessage() *entities.Message {
	message := f.CreateMessage()
	message.IsRead = true
	return message
}

// CreateDeletedMessage creates a deleted message
func (f *MessageFactory) CreateDeletedMessage() *entities.Message {
	message := f.CreateMessage()
	message.IsDeleted = true
	return message
}

// CreateCustomMessage creates a test message with custom values
func (f *MessageFactory) CreateCustomMessage(opts ...MessageOption) *entities.Message {
	message := f.CreateMessage()
	
	for _, opt := range opts {
		opt(message)
	}
	
	return message
}

// MessageOption defines a function type for customizing message creation
type MessageOption func(*entities.Message)

// WithMessageID sets the message ID
func WithMessageID(id uuid.UUID) MessageOption {
	return func(m *entities.Message) {
		m.ID = id
	}
}

// WithConversationID sets the conversation ID
func WithConversationID(conversationID uuid.UUID) MessageOption {
	return func(m *entities.Message) {
		m.ConversationID = conversationID
	}
}

// WithSenderID sets the sender ID
func WithSenderID(senderID uuid.UUID) MessageOption {
	return func(m *entities.Message) {
		m.SenderID = senderID
	}
}

// WithReceiverID sets the receiver ID
func WithReceiverID(receiverID uuid.UUID) MessageOption {
	return func(m *entities.Message) {
		m.ReceiverID = receiverID
	}
}

// WithContent sets the message content
func WithContent(content string) MessageOption {
	return func(m *entities.Message) {
		m.Content = content
	}
}

// WithMessageType sets the message type
func WithMessageType(messageType entities.MessageType) MessageOption {
	return func(m *entities.Message) {
		m.MessageType = messageType
	}
}

// WithRead sets the read status
func WithRead(isRead bool) MessageOption {
	return func(m *entities.Message) {
		m.IsRead = isRead
	}
}

// WithDeleted sets the deleted status
func WithDeleted(isDeleted bool) MessageOption {
	return func(m *entities.Message) {
		m.IsDeleted = isDeleted
	}
}

// WithMessageLocation sets the location
func WithMessageLocation(lat, lng float64, address string) MessageOption {
	return func(m *entities.Message) {
		m.Location = &entities.Location{
			Latitude:  lat,
			Longitude: lng,
			Address:   address,
		}
	}
}

// WithEphemeralPhotoDuration sets the ephemeral photo duration
func WithEphemeralPhotoDuration(duration int) MessageOption {
	return func(m *entities.Message) {
		m.EphemeralPhotoDuration = duration
	}
}

// WithMessageCreatedAt sets the creation time
func WithMessageCreatedAt(createdAt time.Time) MessageOption {
	return func(m *entities.Message) {
		m.CreatedAt = createdAt
	}
}

// WithMessageUpdatedAt sets the update time
func WithMessageUpdatedAt(updatedAt time.Time) MessageOption {
	return func(m *entities.Message) {
		m.UpdatedAt = updatedAt
	}
}

// CreateMultipleMessages creates multiple test messages
func (f *MessageFactory) CreateMultipleMessages(count int) []*entities.Message {
	messages := make([]*entities.Message, count)
	for i := 0; i < count; i++ {
		messages[i] = f.CreateMessage()
	}
	return messages
}

// CreateMultipleCustomMessages creates multiple test messages with custom options
func (f *MessageFactory) CreateMultipleCustomMessages(count int, opts ...MessageOption) []*entities.Message {
	messages := make([]*entities.Message, count)
	for i := 0; i < count; i++ {
		messages[i] = f.CreateCustomMessage(opts...)
	}
	return messages
}

// CreateConversation creates a conversation between two users
func (f *MessageFactory) CreateConversation(userID1, userID2 uuid.UUID, messageCount int) []*entities.Message {
	conversationID := uuid.New()
	messages := make([]*entities.Message, messageCount)
	
	for i := 0; i < messageCount; i++ {
		senderID := userID1
		receiverID := userID2
		if i%2 == 1 {
			senderID, receiverID = receiverID, senderID
		}
		
		message := f.CreateCustomMessage(
			WithConversationID(conversationID),
			WithSenderID(senderID),
			WithReceiverID(receiverID),
			WithContent("Message " + string(rune(i+1))),
		)
		messages[i] = message
	}
	
	return messages
}