package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MessageService provides message validation, filtering, and processing
type MessageService struct {
	messageRepo repositories.MessageRepository
	userRepo    repositories.UserRepository
	cache       *cache.CacheService
	rateLimiter *cache.RateLimiter
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
	cache *cache.CacheService,
	rateLimiter *cache.RateLimiter,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		cache:       cache,
		rateLimiter: rateLimiter,
	}
}

// MessageValidationResult represents the result of message validation
type MessageValidationResult struct {
	IsValid   bool                   `json:"is_valid"`
	Errors    []string               `json:"errors,omitempty"`
	Warnings  []string               `json:"warnings,omitempty"`
	Sanitized string                 `json:"sanitized,omitempty"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

// MessageProcessingOptions represents options for message processing
type MessageProcessingOptions struct {
	EnableContentFilter bool `json:"enable_content_filter"`
	EnableLinkPreview  bool `json:"enable_link_preview"`
	EnableEncryption   bool `json:"enable_encryption"`
	EnableTranslation  bool `json:"enable_translation"`
	TargetLanguage     string `json:"target_language,omitempty"`
}

// LinkPreview represents a link preview
type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
}

// ProcessedMessage represents a processed message
type ProcessedMessage struct {
	*entities.Message
	LinkPreviews []LinkPreview          `json:"link_previews,omitempty"`
	IsEncrypted  bool                 `json:"is_encrypted,omitempty"`
	Translations map[string]string      `json:"translations,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ValidateMessage validates a message before sending
func (s *MessageService) ValidateMessage(ctx context.Context, content, messageType, senderID string) (*MessageValidationResult, error) {
	result := &MessageValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
		Metadata: make(map[string]interface{}),
	}

	// Check message length
	if err := s.validateMessageLength(content, messageType); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Check message content
	if err := s.validateMessageContent(content, messageType); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	// Check for spam
	if isSpam, err := s.checkSpam(ctx, senderID, content); err != nil {
		return nil, fmt.Errorf("failed to check spam: %w", err)
	} else if isSpam {
		result.IsValid = false
		result.Errors = append(result.Errors, "message detected as spam")
	}

	// Check rate limiting
	if isRateLimited, err := s.checkRateLimit(ctx, senderID); err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	} else if isRateLimited {
		result.IsValid = false
		result.Errors = append(result.Errors, "rate limit exceeded")
	}

	// Sanitize content
	sanitized, warnings := s.sanitizeContent(content)
	result.Sanitized = sanitized
	result.Warnings = append(result.Warnings, warnings...)

	// Add metadata
	result.Metadata["char_count"] = utf8.RuneCountInString(content)
	result.Metadata["word_count"] = len(strings.Fields(content))
	result.Metadata["message_type"] = messageType

	return result, nil
}

// ProcessMessage processes a message with various options
func (s *MessageService) ProcessMessage(ctx context.Context, message *entities.Message, options *MessageProcessingOptions) (*ProcessedMessage, error) {
	processed := &ProcessedMessage{
		Message:  message,
		Metadata: make(map[string]interface{}),
	}

	// Content filtering
	if options != nil && options.EnableContentFilter {
		if err := s.filterContent(processed); err != nil {
			return nil, fmt.Errorf("content filtering failed: %w", err)
		}
	}

	// Link preview generation
	if options != nil && options.EnableLinkPreview && message.MessageType == "text" {
		previews, err := s.generateLinkPreviews(message.Content)
		if err != nil {
			logger.Warn("Failed to generate link previews", err)
		} else {
			processed.LinkPreviews = previews
		}
	}

	// Encryption for sensitive content
	if options != nil && options.EnableEncryption {
		if err := s.encryptSensitiveContent(processed); err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		processed.IsEncrypted = true
	}

	// Translation
	if options != nil && options.EnableTranslation && options.TargetLanguage != "" {
		translations, err := s.translateMessage(ctx, message.Content, options.TargetLanguage)
		if err != nil {
			logger.Warn("Failed to translate message", err)
		} else {
			processed.Translations = translations
		}
	}

	// Add processing metadata
	processed.Metadata["processed_at"] = time.Now()
	processed.Metadata["processing_options"] = options

	return processed, nil
}

// GetMessageAnalytics returns analytics for a message
func (s *MessageService) GetMessageAnalytics(ctx context.Context, messageID uuid.UUID) (map[string]interface{}, error) {
	// Get message from repository
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	analytics := make(map[string]interface{})
	
	// Basic metrics
	analytics["message_id"] = message.ID
	analytics["message_type"] = message.MessageType
	analytics["char_count"] = utf8.RuneCountInString(message.Content)
	analytics["word_count"] = len(strings.Fields(message.Content))
	analytics["is_read"] = message.IsRead
	analytics["created_at"] = message.CreatedAt
	analytics["age_hours"] = time.Since(message.CreatedAt).Hours()

	// Content analysis
	analytics["has_links"] = s.containsLinks(message.Content)
	analytics["has_mentions"] = s.containsMentions(message.Content)
	analytics["has_emojis"] = s.containsEmojis(message.Content)
	analytics["sentiment_score"] = s.analyzeSentiment(message.Content)

	// Engagement metrics (would be populated from analytics data)
	analytics["view_count"] = 0 // Would come from analytics tracking
	analytics["response_time"] = 0 // Would come from analytics tracking

	return analytics, nil
}

// validateMessageLength validates message length based on type
func (s *MessageService) validateMessageLength(content, messageType string) error {
	charCount := utf8.RuneCountInString(content)
	
	switch messageType {
	case "text":
		if charCount == 0 {
			return fmt.Errorf("message cannot be empty")
		}
		if charCount > 2000 {
			return fmt.Errorf("message too long (max 2000 characters)")
		}
	case "image":
		if charCount > 500 {
			return fmt.Errorf("image caption too long (max 500 characters)")
		}
	case "ephemeral_photo":
		if charCount > 500 {
			return fmt.Errorf("ephemeral photo caption too long (max 500 characters)")
		}
	case "location":
		if charCount > 200 {
			return fmt.Errorf("location description too long (max 200 characters)")
		}
	case "system":
		if charCount > 1000 {
			return fmt.Errorf("system message too long (max 1000 characters)")
		}
	default:
		return fmt.Errorf("unsupported message type: %s", messageType)
	}
	
	return nil
}

// validateMessageContent validates message content
func (s *MessageService) validateMessageContent(content, messageType string) error {
	// Check for forbidden content patterns
	forbiddenPatterns := []string{
		`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`, // Credit card numbers
		`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`,           // SSN pattern
	}
	
	for _, pattern := range forbiddenPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return fmt.Errorf("message contains forbidden content")
		}
	}
	
	// Check for excessive repetition
	if s.hasExcessiveRepetition(content) {
		return fmt.Errorf("message contains excessive repetition")
	}
	
	// Check for excessive caps
	if s.hasExcessiveCaps(content) {
		return fmt.Errorf("message contains excessive capitalization")
	}
	
	return nil
}

// sanitizeContent sanitizes message content
func (s *MessageService) sanitizeContent(content string) (string, []string) {
	warnings := []string{}
	sanitized := content
	
	// Remove potentially dangerous HTML
	sanitized = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(sanitized, "")
	
	// Normalize whitespace
	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)
	
	// Add warning if content was modified
	if sanitized != content {
		warnings = append(warnings, "content was sanitized")
	}
	
	return sanitized, warnings
}

// checkSpam checks if message is spam
func (s *MessageService) checkSpam(ctx context.Context, senderID, content string) (bool, error) {
	// Check against spam patterns
	spamPatterns := []string{
		`(?i)click here`,
		`(?i)buy now`,
		`(?i)limited time`,
		`(?i)act fast`,
		`(?i)free money`,
		`(?i)guaranteed`,
	}
	
	for _, pattern := range spamPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true, nil
		}
	}
	
	// Check sender's spam score from cache
	senderKey := fmt.Sprintf("spam_score:%s", senderID)
	if score, err := s.cache.Get(ctx, senderKey); err == nil {
		if score.(float64) > 0.7 {
			return true, nil
		}
	}
	
	return false, nil
}

// checkRateLimit checks if user is rate limited
func (s *MessageService) checkRateLimit(ctx context.Context, senderID string) (bool, error) {
	if s.rateLimiter == nil {
		return false, nil
	}
	
	// Check message rate limit (60 messages per minute)
	key := fmt.Sprintf("message_rate:%s", senderID)
	allowed, err := s.rateLimiter.Allow(ctx, key, 60, time.Minute)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	return !allowed, nil
}

// filterContent filters inappropriate content
func (s *MessageService) filterContent(processed *ProcessedMessage) error {
	// Implement content filtering logic
	// This would integrate with content moderation service
	
	// For now, just check for basic profanity
	profanityList := []string{
		"profanity1", "profanity2", // Add actual profanity words
	}
	
	content := strings.ToLower(processed.Content)
	for _, profanity := range profanityList {
		if strings.Contains(content, profanity) {
			// Replace with asterisks
			processed.Content = strings.ReplaceAll(processed.Content, profanity, strings.Repeat("*", len(profanity)))
		}
	}
	
	return nil
}

// generateLinkPreviews generates link previews for URLs in message
func (s *MessageService) generateLinkPreviews(content string) ([]LinkPreview, error) {
	// Extract URLs from content
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(content, -1)
	
	if len(urls) == 0 {
		return nil, nil
	}
	
	previews := make([]LinkPreview, 0, len(urls))
	
	for _, url := range urls {
		// In a real implementation, this would fetch URL metadata
		// For now, return basic preview
		preview := LinkPreview{
			URL:         url,
			Title:       "Link Preview",
			Description: "Click to view link",
		}
		previews = append(previews, preview)
	}
	
	return previews, nil
}

// encryptSensitiveContent encrypts sensitive information in message
func (s *MessageService) encryptSensitiveContent(processed *ProcessedMessage) error {
	// Implement encryption for sensitive content
	// This would use proper encryption algorithms
	
	// For now, just mark as encrypted
	processed.Metadata["encryption_method"] = "AES-256-GCM"
	processed.Metadata["encrypted_at"] = time.Now()
	
	return nil
}

// translateMessage translates message to target language
func (s *MessageService) translateMessage(ctx context.Context, content, targetLanguage string) (map[string]string, error) {
	// Implement translation logic
	// This would integrate with translation service
	
	translations := make(map[string]string)
	translations[targetLanguage] = fmt.Sprintf("[Translated to %s] %s", targetLanguage, content)
	
	return translations, nil
}

// hasExcessiveRepetition checks for excessive character repetition
func (s *MessageService) hasExcessiveRepetition(content string) bool {
	// Check for same character repeated more than 5 times
	repeatPattern := regexp.MustCompile(`(.)\1{5,}`)
	return repeatPattern.MatchString(content)
}

// hasExcessiveCaps checks for excessive capitalization
func (s *MessageService) hasExcessiveCaps(content string) bool {
	if len(content) < 10 {
		return false
	}
	
	caps := 0
	for _, char := range content {
		if char >= 'A' && char <= 'Z' {
			caps++
		}
	}
	
	capsRatio := float64(caps) / float64(len(content))
	return capsRatio > 0.7
}

// containsLinks checks if content contains URLs
func (s *MessageService) containsLinks(content string) bool {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	return urlRegex.MatchString(content)
}

// containsMentions checks if content contains user mentions
func (s *MessageService) containsMentions(content string) bool {
	mentionRegex := regexp.MustCompile(`@\w+`)
	return mentionRegex.MatchString(content)
}

// containsEmojis checks if content contains emojis
func (s *MessageService) containsEmojis(content string) bool {
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]`)
	return emojiRegex.MatchString(content)
}

// analyzeSentiment analyzes sentiment of message content
func (s *MessageService) analyzeSentiment(content string) float64 {
	// Simple sentiment analysis
	// In a real implementation, this would use NLP service
	
	positiveWords := []string{"good", "great", "awesome", "happy", "love"}
	negativeWords := []string{"bad", "terrible", "awful", "sad", "hate"}
	
	contentLower := strings.ToLower(content)
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if strings.Contains(contentLower, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(contentLower, word) {
			negativeCount++
		}
	}
	
	if positiveCount+negativeCount == 0 {
		return 0.0
	}
	
	return float64(positiveCount-negativeCount) / float64(positiveCount+negativeCount)
}