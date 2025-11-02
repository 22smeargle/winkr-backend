package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MessageRepositoryImpl implements MessageRepository interface using GORM
type MessageRepositoryImpl struct {
	db *gorm.DB
}

// NewMessageRepository creates a new MessageRepository instance
func NewMessageRepository(db *gorm.DB) repositories.MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

// Create creates a new message
func (r *MessageRepositoryImpl) Create(ctx context.Context, message *entities.Message) error {
	modelMessage := r.domainToModelMessage(message)
	if err := r.db.WithContext(ctx).Create(modelMessage).Error; err != nil {
		logger.Error("Failed to create message", err)
		return fmt.Errorf("failed to create message: %w", err)
	}

	logger.Info("Message created successfully", map[string]interface{}{
		"message_id": message.ID,
		"conversation_id": message.ConversationID,
		"sender_id": message.SenderID,
	})
	return nil
}

// GetByID retrieves a message by ID
func (r *MessageRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Message, error) {
	var message models.Message
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found")
		}
		logger.Error("Failed to get message by ID", err)
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	// Convert to domain entity
	domainMessage := r.modelToDomainMessage(&message)
	return domainMessage, nil
}

// Update updates a message
func (r *MessageRepositoryImpl) Update(ctx context.Context, message *entities.Message) error {
	modelMessage := r.domainToModelMessage(message)
	if err := r.db.WithContext(ctx).Save(modelMessage).Error; err != nil {
		logger.Error("Failed to update message", err)
		return fmt.Errorf("failed to update message: %w", err)
	}

	logger.Info("Message updated successfully", map[string]interface{}{
		"message_id": message.ID,
		"conversation_id": message.ConversationID,
	})
	return nil
}

// Delete soft deletes a message
func (r *MessageRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete message", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	logger.Info("Message deleted successfully", map[string]interface{}{
		"message_id": id,
	})
	return nil
}

// GetConversationMessages retrieves messages for a conversation with pagination
func (r *MessageRepositoryImpl) GetConversationMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		logger.Error("Failed to get conversation messages", err)
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// GetConversationMessagesAfter retrieves messages after a specific timestamp
func (r *MessageRepositoryImpl) GetConversationMessagesAfter(ctx context.Context, conversationID uuid.UUID, after time.Time) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ? AND created_at > ?", conversationID, after).Order("created_at ASC").Find(&messages).Error; err != nil {
		logger.Error("Failed to get conversation messages after timestamp", err)
		return nil, fmt.Errorf("failed to get conversation messages after timestamp: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// GetConversationMessagesBefore retrieves messages before a specific timestamp
func (r *MessageRepositoryImpl) GetConversationMessagesBefore(ctx context.Context, conversationID uuid.UUID, before time.Time, limit int) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ? AND created_at < ?", conversationID, before).Order("created_at DESC").Limit(limit).Find(&messages).Error; err != nil {
		logger.Error("Failed to get conversation messages before timestamp", err)
		return nil, fmt.Errorf("failed to get conversation messages before timestamp: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// GetUnreadMessages retrieves unread messages for a user
func (r *MessageRepositoryImpl) GetUnreadMessages(ctx context.Context, userID uuid.UUID) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Joins("JOIN conversations ON messages.conversation_id = conversations.id").
		Where("conversations.user1_id = ? OR conversations.user2_id = ?", userID, userID).
		Where("messages.sender_id != ?", userID).
		Where("messages.is_read = ?", false).
		Order("messages.created_at DESC").
		Find(&messages).Error; err != nil {
		logger.Error("Failed to get unread messages", err)
		return nil, fmt.Errorf("failed to get unread messages: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// MarkAsRead marks a message as read
func (r *MessageRepositoryImpl) MarkAsRead(ctx context.Context, messageID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", messageID).Update("is_read", true).Error; err != nil {
		logger.Error("Failed to mark message as read", err)
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	logger.Info("Message marked as read", map[string]interface{}{
		"message_id": messageID,
	})
	return nil
}

// MarkConversationAsRead marks all messages in a conversation as read for a user
func (r *MessageRepositoryImpl) MarkConversationAsRead(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id != ?", conversationID, userID).
		Update("is_read", true).Error; err != nil {
		logger.Error("Failed to mark conversation as read", err)
		return fmt.Errorf("failed to mark conversation as read: %w", err)
	}

	logger.Info("Conversation marked as read", map[string]interface{}{
		"conversation_id": conversationID,
		"user_id": userID,
	})
	return nil
}

// GetLastMessage retrieves the last message in a conversation
func (r *MessageRepositoryImpl) GetLastMessage(ctx context.Context, conversationID uuid.UUID) (*entities.Message, error) {
	var message models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC").First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no messages found in conversation")
		}
		logger.Error("Failed to get last message", err)
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	// Convert to domain entity
	domainMessage := r.modelToDomainMessage(&message)
	return domainMessage, nil
}

// GetUnreadMessageCount retrieves unread message count for a user
func (r *MessageRepositoryImpl) GetUnreadMessageCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).
		Joins("JOIN conversations ON messages.conversation_id = conversations.id").
		Where("(conversations.user1_id = ? OR conversations.user2_id = ?)", userID, userID).
		Where("messages.sender_id != ?", userID).
		Where("messages.is_read = ?", false).
		Count(&count).Error; err != nil {
		logger.Error("Failed to get unread message count", err)
		return 0, fmt.Errorf("failed to get unread message count: %w", err)
	}

	return count, nil
}

// GetConversationUnreadCount retrieves unread message count for a conversation
func (r *MessageRepositoryImpl) GetConversationUnreadCount(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).
		Where("conversation_id = ?", conversationID).
		Where("sender_id != ?", userID).
		Where("is_read = ?", false).
		Count(&count).Error; err != nil {
		logger.Error("Failed to get conversation unread count", err)
		return 0, fmt.Errorf("failed to get conversation unread count: %w", err)
	}

	return count, nil
}

// DeleteConversationMessages deletes all messages in a conversation
func (r *MessageRepositoryImpl) DeleteConversationMessages(ctx context.Context, conversationID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Delete(&models.Message{}).Error; err != nil {
		logger.Error("Failed to delete conversation messages", err)
		return fmt.Errorf("failed to delete conversation messages: %w", err)
	}

	logger.Info("Conversation messages deleted", map[string]interface{}{
		"conversation_id": conversationID,
	})
	return nil
}

// BatchCreate creates multiple messages
func (r *MessageRepositoryImpl) BatchCreate(ctx context.Context, messages []*entities.Message) error {
	modelMessages := make([]*models.Message, len(messages))
	for i, message := range messages {
		modelMessages[i] = r.domainToModelMessage(message)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(modelMessages, 100).Error; err != nil {
		logger.Error("Failed to batch create messages", err)
		return fmt.Errorf("failed to batch create messages: %w", err)
	}

	logger.Info("Messages batch created successfully", map[string]interface{}{
		"count": len(messages),
	})
	return nil
}

// BatchUpdate updates multiple messages
func (r *MessageRepositoryImpl) BatchUpdate(ctx context.Context, messages []*entities.Message) error {
	modelMessages := make([]*models.Message, len(messages))
	for i, message := range messages {
		modelMessages[i] = r.domainToModelMessage(message)
	}

	if err := r.db.WithContext(ctx).SaveInBatches(modelMessages, 100).Error; err != nil {
		logger.Error("Failed to batch update messages", err)
		return fmt.Errorf("failed to batch update messages: %w", err)
	}

	logger.Info("Messages batch updated successfully", map[string]interface{}{
		"count": len(messages),
	})
	return nil
}

// BatchDelete soft deletes multiple messages
func (r *MessageRepositoryImpl) BatchDelete(ctx context.Context, messageIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", messageIDs).Delete(&models.Message{}).Error; err != nil {
		logger.Error("Failed to batch delete messages", err)
		return fmt.Errorf("failed to batch delete messages: %w", err)
	}

	logger.Info("Messages batch deleted successfully", map[string]interface{}{
		"count": len(messageIDs),
	})
	return nil
}

// ExistsByID checks if message exists by ID
func (r *MessageRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check message existence", err)
		return false, fmt.Errorf("failed to check message existence: %w", err)
	}

	return count > 0, nil
}

// UserCanAccessMessage checks if user can access a message
func (r *MessageRepositoryImpl) UserCanAccessMessage(ctx context.Context, userID, messageID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).
		Joins("JOIN conversations ON messages.conversation_id = conversations.id").
		Where("messages.id = ?", messageID).
		Where("(conversations.user1_id = ? OR conversations.user2_id = ?)", userID, userID).
		Count(&count).Error; err != nil {
		logger.Error("Failed to check message access", err)
		return false, fmt.Errorf("failed to check message access: %w", err)
	}

	return count > 0, nil
}

// GetMessagesByType retrieves messages by type
func (r *MessageRepositoryImpl) GetMessagesByType(ctx context.Context, conversationID uuid.UUID, messageType string, limit, offset int) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ? AND message_type = ?", conversationID, messageType).Order("created_at DESC").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		logger.Error("Failed to get messages by type", err)
		return nil, fmt.Errorf("failed to get messages by type: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// GetMessagesWithAttachments retrieves messages with attachments
func (r *MessageRepositoryImpl) GetMessagesWithAttachments(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ? AND attachment_url IS NOT NULL", conversationID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		logger.Error("Failed to get messages with attachments", err)
		return nil, fmt.Errorf("failed to get messages with attachments: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// SearchMessages searches messages by content
func (r *MessageRepositoryImpl) SearchMessages(ctx context.Context, conversationID uuid.UUID, query string, limit, offset int) ([]*entities.Message, error) {
	var messages []models.Message
	if err := r.db.WithContext(ctx).Where("conversation_id = ? AND content ILIKE ?", conversationID, "%"+query+"%").Order("created_at DESC").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		logger.Error("Failed to search messages", err)
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	// Convert to domain entities
	domainMessages := make([]*entities.Message, len(messages))
	for i, message := range messages {
		domainMessages[i] = r.modelToDomainMessage(&message)
	}

	return domainMessages, nil
}

// GetMessageStats retrieves message statistics
func (r *MessageRepositoryImpl) GetMessageStats(ctx context.Context) (*repositories.MessageStats, error) {
	var stats repositories.MessageStats
	
	// Get total messages count
	r.db.WithContext(ctx).Model(&models.Message{}).Count(&stats.TotalMessages)
	
	// Get sent messages count
	r.db.WithContext(ctx).Model(&models.Message{}).Where("message_type = ?", "sent").Count(&stats.SentMessages)
	
	// Get received messages count
	r.db.WithContext(ctx).Model(&models.Message{}).Where("message_type = ?", "received").Count(&stats.ReceivedMessages)
	
	// Get unread messages count
	r.db.WithContext(ctx).Model(&models.Message{}).Where("is_read = ?", false).Count(&stats.UnreadMessages)
	
	// Get messages sent today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Message{}).Where("DATE(created_at) = DATE(?)", today).Count(&stats.MessagesToday)
	
	// Get messages sent this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Message{}).Where("created_at >= ?", weekStart).Count(&stats.MessagesThisWeek)
	
	// Get messages sent this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Message{}).Where("created_at >= ?", monthStart).Count(&stats.MessagesThisMonth)
	
	return &stats, nil
}

// GetMessagesSentInRange retrieves count of messages sent in date range
func (r *MessageRepositoryImpl) GetMessagesSentInRange(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Message{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&count).Error; err != nil {
		logger.Error("Failed to get messages sent in range", err)
		return 0, fmt.Errorf("failed to get messages sent in range: %w", err)
	}

	return count, nil
}

// Conversation methods

// CreateConversation creates a new conversation
func (r *MessageRepositoryImpl) CreateConversation(ctx context.Context, conversation *entities.Conversation) error {
	modelConversation := r.domainToModelConversation(conversation)
	if err := r.db.WithContext(ctx).Create(modelConversation).Error; err != nil {
		logger.Error("Failed to create conversation", err)
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	logger.Info("Conversation created successfully", map[string]interface{}{
		"conversation_id": conversation.ID,
		"user1_id": conversation.User1ID,
		"user2_id": conversation.User2ID,
	})
	return nil
}

// GetConversationByID retrieves a conversation by ID
func (r *MessageRepositoryImpl) GetConversationByID(ctx context.Context, id uuid.UUID) (*entities.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found")
		}
		logger.Error("Failed to get conversation by ID", err)
		return nil, fmt.Errorf("failed to get conversation by ID: %w", err)
	}

	// Convert to domain entity
	domainConversation := r.modelToDomainConversation(&conversation)
	return domainConversation, nil
}

// UpdateConversation updates a conversation
func (r *MessageRepositoryImpl) UpdateConversation(ctx context.Context, conversation *entities.Conversation) error {
	modelConversation := r.domainToModelConversation(conversation)
	if err := r.db.WithContext(ctx).Save(modelConversation).Error; err != nil {
		logger.Error("Failed to update conversation", err)
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	logger.Info("Conversation updated successfully", map[string]interface{}{
		"conversation_id": conversation.ID,
	})
	return nil
}

// DeleteConversation soft deletes a conversation
func (r *MessageRepositoryImpl) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Conversation{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete conversation", err)
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	logger.Info("Conversation deleted successfully", map[string]interface{}{
		"conversation_id": id,
	})
	return nil
}

// GetUserConversations retrieves conversations for a user
func (r *MessageRepositoryImpl) GetUserConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Conversation, error) {
	var conversations []models.Conversation
	if err := r.db.WithContext(ctx).Where("user1_id = ? OR user2_id = ?", userID, userID).Order("updated_at DESC").Limit(limit).Offset(offset).Find(&conversations).Error; err != nil {
		logger.Error("Failed to get user conversations", err)
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}

	// Convert to domain entities
	domainConversations := make([]*entities.Conversation, len(conversations))
	for i, conversation := range conversations {
		domainConversations[i] = r.modelToDomainConversation(&conversation)
	}

	return domainConversations, nil
}

// GetConversationByUsers retrieves conversation between two users
func (r *MessageRepositoryImpl) GetConversationByUsers(ctx context.Context, user1ID, user2ID uuid.UUID) (*entities.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.WithContext(ctx).Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found")
		}
		logger.Error("Failed to get conversation by users", err)
		return nil, fmt.Errorf("failed to get conversation by users: %w", err)
	}

	// Convert to domain entity
	domainConversation := r.modelToDomainConversation(&conversation)
	return domainConversation, nil
}

// UpdateLastMessage updates the last message in a conversation
func (r *MessageRepositoryImpl) UpdateLastMessage(ctx context.Context, conversationID, messageID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Conversation{}).Where("id = ?", conversationID).Update("last_message_id", messageID).Error; err != nil {
		logger.Error("Failed to update last message", err)
		return fmt.Errorf("failed to update last message: %w", err)
	}

	logger.Info("Last message updated", map[string]interface{}{
		"conversation_id": conversationID,
		"message_id": messageID,
	})
	return nil
}

// GetConversationCount retrieves conversation count for a user
func (r *MessageRepositoryImpl) GetConversationCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Conversation{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&count).Error; err != nil {
		logger.Error("Failed to get conversation count", err)
		return 0, fmt.Errorf("failed to get conversation count: %w", err)
	}

	return count, nil
}

// ExistsConversation checks if conversation exists between two users
func (r *MessageRepositoryImpl) ExistsConversation(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Conversation{}).Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).Count(&count).Error; err != nil {
		logger.Error("Failed to check conversation existence", err)
		return false, fmt.Errorf("failed to check conversation existence: %w", err)
	}

	return count > 0, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainMessage converts model Message to domain Message
func (r *MessageRepositoryImpl) modelToDomainMessage(model *models.Message) *entities.Message {
	return &entities.Message{
		ID:             model.ID,
		ConversationID: model.ConversationID,
		SenderID:       model.SenderID,
		Content:        model.Content,
		MessageType:    model.MessageType,
		AttachmentURL:  model.AttachmentURL,
		IsRead:         model.IsRead,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

// domainToModelMessage converts domain Message to model Message
func (r *MessageRepositoryImpl) domainToModelMessage(message *entities.Message) *models.Message {
	return &models.Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		AttachmentURL:  message.AttachmentURL,
		IsRead:         message.IsRead,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,
	}
}

// modelToDomainConversation converts model Conversation to domain Conversation
func (r *MessageRepositoryImpl) modelToDomainConversation(model *models.Conversation) *entities.Conversation {
	return &entities.Conversation{
		ID:            model.ID,
		User1ID:       model.User1ID,
		User2ID:       model.User2ID,
		MatchID:       model.MatchID,
		LastMessageID: model.LastMessageID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

// domainToModelConversation converts domain Conversation to model Conversation
func (r *MessageRepositoryImpl) domainToModelConversation(conversation *entities.Conversation) *models.Conversation {
	return &models.Conversation{
		ID:            conversation.ID,
		User1ID:       conversation.User1ID,
		User2ID:       conversation.User2ID,
		MatchID:       conversation.MatchID,
		LastMessageID: conversation.LastMessageID,
		CreatedAt:     conversation.CreatedAt,
		UpdatedAt:     conversation.UpdatedAt,
	}
}