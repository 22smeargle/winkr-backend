package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ChatSecurityService provides security features for chat functionality
type ChatSecurityService struct {
	userRepo      repositories.UserRepository
	messageRepo   repositories.MessageRepository
	cache         *cache.CacheService
	rateLimiter   *cache.RateLimiter
	blockedUsers  map[string]bool // User ID -> IsBlocked
	spamKeywords  []string
}

// NewChatSecurityService creates a new chat security service
func NewChatSecurityService(
	userRepo repositories.UserRepository,
	messageRepo repositories.MessageRepository,
	cache *cache.CacheService,
	rateLimiter *cache.RateLimiter,
) *ChatSecurityService {
	return &ChatSecurityService{
		userRepo:     userRepo,
		messageRepo:   messageRepo,
		cache:        cache,
		rateLimiter:   rateLimiter,
		blockedUsers:  make(map[string]bool),
		spamKeywords: []string{
			"click here", "buy now", "limited time", "act fast", "free money",
			"guaranteed", "winner", "congratulations", "claim now",
			"viagra", "cialis", "casino", "lottery", "investment",
			"bitcoin", "cryptocurrency", "make money", "work from home",
		},
	}
}

// SecurityCheckResult represents the result of a security check
type SecurityCheckResult struct {
	IsAllowed     bool     `json:"is_allowed"`
	RiskScore     float64  `json:"risk_score"`
	Violations    []string  `json:"violations"`
	Warnings      []string  `json:"warnings"`
	FilteredContent string    `json:"filtered_content,omitempty"`
}

// MessageSecurity represents security metadata for a message
type MessageSecurity struct {
	MessageID      uuid.UUID `json:"message_id"`
	IsEncrypted    bool      `json:"is_encrypted"`
	EncryptionKey   string    `json:"encryption_key,omitempty"`
	FilteredWords   []string  `json:"filtered_words,omitempty"`
	RiskScore       float64  `json:"risk_score"`
	SecurityLevel   string    `json:"security_level"`
	CreatedAt       time.Time `json:"created_at"`
}

// CheckMessageSecurity performs security checks on a message
func (s *ChatSecurityService) CheckMessageSecurity(ctx context.Context, senderID, receiverID, content string) (*SecurityCheckResult, error) {
	result := &SecurityCheckResult{
		IsAllowed:  true,
		RiskScore:  0.0,
		Violations: []string{},
		Warnings:   []string{},
	}

	// Check if sender is blocked by receiver
	if s.isUserBlocked(ctx, senderID, receiverID) {
		result.IsAllowed = false
		result.Violations = append(result.Violations, "sender is blocked by receiver")
		result.RiskScore += 0.8
	}

	// Check for spam content
	if spamScore, spamWords := s.checkSpamContent(content); spamScore > 0.5 {
		result.RiskScore += spamScore
		result.Violations = append(result.Violations, "potential spam content")
		result.Warnings = append(result.Warnings, fmt.Sprintf("spam keywords detected: %v", spamWords))
	}

	// Check for inappropriate content
	if inapScore, inapWords := s.checkInappropriateContent(content); inapScore > 0.3 {
		result.RiskScore += inapScore
		result.Violations = append(result.Violations, "inappropriate content")
		result.Warnings = append(result.Warnings, fmt.Sprintf("inappropriate words detected: %v", inapWords))
	}

	// Check for personal information sharing
	if piiScore, piiTypes := s.checkPIIContent(content); piiScore > 0.4 {
		result.RiskScore += piiScore
		result.Violations = append(result.Violations, "personal information sharing")
		result.Warnings = append(result.Warnings, fmt.Sprintf("PII detected: %v", piiTypes))
	}

	// Check for external links
	if linkScore := s.checkExternalLinks(content); linkScore > 0.2 {
		result.RiskScore += linkScore
		result.Warnings = append(result.Warnings, "external links detected")
	}

	// Filter content if violations found
	if len(result.Violations) > 0 {
		result.FilteredContent = s.filterContent(content)
	}

	// Determine security level
	result.SecurityLevel = s.determineSecurityLevel(result.RiskScore)

	// Log security check
	logger.Info("Message security check completed", 
		"sender_id", senderID,
		"receiver_id", receiverID,
		"risk_score", result.RiskScore,
		"violations_count", len(result.Violations),
		"security_level", result.SecurityLevel,
	)

	return result, nil
}

// EncryptMessage encrypts sensitive message content
func (s *ChatSecurityService) EncryptMessage(content string) (string, string, error) {
	// Generate encryption key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return "", "", fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt content
	encrypted := gcm.Seal(nonce, nil, []byte(content))
	
	// Encode result
	encryptedContent := base64.StdEncoding.EncodeToString(encrypted)
	keyString := base64.StdEncoding.EncodeToString(key)
	nonceString := base64.StdEncoding.EncodeToString(nonce)

	// Return encrypted content with key and nonce
	return encryptedContent, fmt.Sprintf("%s:%s", keyString, nonceString), nil
}

// DecryptMessage decrypts encrypted message content
func (s *ChatSecurityService) DecryptMessage(encryptedContent, keyData string) (string, error) {
	// Parse key and nonce
	parts := strings.Split(keyData, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid key data format")
	}

	key, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode key: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Decode encrypted content
	encrypted, err := base64.StdEncoding.DecodeString(encryptedContent)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted content: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt content
	decrypted, err := gcm.Open(nil, nonce, encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt content: %w", err)
	}

	return string(decrypted), nil
}

// BlockUser blocks a user from sending messages
func (s *ChatSecurityService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	// Add to blocked users map
	s.blockedUsers[blockedID] = true

	// Cache block relationship
	blockKey := fmt.Sprintf("blocked:%s:%s", blockerID, blockedID)
	if err := s.cache.Set(ctx, blockKey, true, 24*time.Hour); err != nil {
		logger.Error("Failed to cache block relationship", err)
	}

	// Log block action
	logger.Info("User blocked", 
		"blocker_id", blockerID,
		"blocked_id", blockedID,
	)

	return nil
}

// UnblockUser unblocks a user
func (s *ChatSecurityService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	// Remove from blocked users map
	delete(s.blockedUsers, blockedID)

	// Remove from cache
	blockKey := fmt.Sprintf("blocked:%s:%s", blockerID, blockedID)
	if err := s.cache.Delete(ctx, blockKey); err != nil {
		logger.Error("Failed to delete block relationship from cache", err)
	}

	// Log unblock action
	logger.Info("User unblocked", 
		"blocker_id", blockerID,
		"blocked_id", blockedID,
	)

	return nil
}

// IsUserBlocked checks if a user is blocked
func (s *ChatSecurityService) IsUserBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	// Check in-memory map
	if s.blockedUsers[blockedID] {
		return true, nil
	}

	// Check in cache
	blockKey := fmt.Sprintf("blocked:%s:%s", blockerID, blockedID)
	if blocked, err := s.cache.Get(ctx, blockKey); err == nil {
		return blocked.(bool), nil
	}

	return false, nil
}

// ReportMessage reports a message for security review
func (s *ChatSecurityService) ReportMessage(ctx context.Context, reporterID, messageID string, reason string) error {
	// Create security record
	security := &MessageSecurity{
		MessageID:    uuid.MustParse(messageID),
		IsEncrypted:  false,
		SecurityLevel: "reported",
		CreatedAt:    time.Now(),
	}

	// Cache security record
	securityKey := fmt.Sprintf("message_security:%s", messageID)
	if err := s.cache.Set(ctx, securityKey, security, 24*time.Hour); err != nil {
		logger.Error("Failed to cache message security record", err)
	}

	// Log report action
	logger.Info("Message reported for security review", 
		"reporter_id", reporterID,
		"message_id", messageID,
		"reason", reason,
	)

	return nil
}

// GetBlockedUsers returns list of blocked users for a user
func (s *ChatSecurityService) GetBlockedUsers(ctx context.Context, blockerID string) ([]string, error) {
	var blockedUsers []string

	// Get from cache pattern
	pattern := fmt.Sprintf("blocked:%s:*", blockerID)
	keys, err := s.cache.Keys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked users: %w", err)
	}

	// Extract blocked user IDs
	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			blockedUsers = append(blockedUsers, parts[1])
		}
	}

	return blockedUsers, nil
}

// Helper methods

// isUserBlocked checks if user is blocked (internal method)
func (s *ChatSecurityService) isUserBlocked(ctx context.Context, senderID, receiverID string) bool {
	blockKey := fmt.Sprintf("blocked:%s:%s", receiverID, senderID)
	if blocked, err := s.cache.Get(ctx, blockKey); err == nil {
		return blocked.(bool)
	}
	return false
}

// checkSpamContent checks for spam indicators
func (s *ChatSecurityService) checkSpamContent(content string) (float64, []string) {
	content = strings.ToLower(content)
	score := 0.0
	foundWords := []string{}

	for _, keyword := range s.spamKeywords {
		if strings.Contains(content, keyword) {
			score += 0.3
			foundWords = append(foundWords, keyword)
		}
	}

	// Check for excessive capitalization
	if s.hasExcessiveCaps(content) {
		score += 0.2
		foundWords = append(foundWords, "excessive_caps")
	}

	// Check for excessive repetition
	if s.hasExcessiveRepetition(content) {
		score += 0.2
		foundWords = append(foundWords, "excessive_repetition")
	}

	return score, foundWords
}

// checkInappropriateContent checks for inappropriate content
func (s *ChatSecurityService) checkInappropriateContent(content string) (float64, []string) {
	// List of inappropriate words (would be expanded in production)
	inappropriateWords := []string{
		"profanity1", "profanity2", "profanity3", // Add actual profanity words
	}

	content = strings.ToLower(content)
	score := 0.0
	foundWords := []string{}

	for _, word := range inappropriateWords {
		if strings.Contains(content, word) {
			score += 0.4
			foundWords = append(foundWords, word)
		}
	}

	return score, foundWords
}

// checkPIIContent checks for personal information sharing
func (s *ChatSecurityService) checkPIIContent(content string) (float64, []string) {
	// Patterns for PII
	piiPatterns := []struct {
		pattern string
		type    string
	}{
		{`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`, "credit_card"},
		{`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`, "ssn"},
		{`\b\d{10}\b`, "phone_number"},
		{`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`, "email"},
		{`\b\d+\s+[A-Z][a-z]+\s+[A-Z][a-z]+\b`, "address"},
	}

	content = strings.ToLower(content)
	score := 0.0
	foundTypes := []string{}

	for _, pii := range piiPatterns {
		if matched, _ := regexp.MatchString(pii.pattern, content); matched {
			score += 0.5
			foundTypes = append(foundTypes, pii.type)
		}
	}

	return score, foundTypes
}

// checkExternalLinks checks for external links
func (s *ChatSecurityService) checkExternalLinks(content string) float64 {
	// Pattern for URLs
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlPattern.FindAllString(content)

	if len(matches) == 0 {
		return 0.0
	}

	// Check for suspicious domains
	suspiciousDomains := []string{
		"bit.ly", "tinyurl.com", "short.link",
		"suspicious-domain.com", "malware-site.net",
	}

	score := 0.2
	for _, match := range matches {
		for _, domain := range suspiciousDomains {
			if strings.Contains(match, domain) {
				score += 0.3
				break
			}
		}
	}

	return score
}

// filterContent filters inappropriate content
func (s *ChatSecurityService) filterContent(content string) string {
	// Replace inappropriate words with asterisks
	inappropriateWords := []string{
		"profanity1", "profanity2", "profanity3", // Add actual profanity words
	}

	filtered := content
	for _, word := range inappropriateWords {
		filtered = strings.ReplaceAll(filtered, word, strings.Repeat("*", len(word)))
	}

	return filtered
}

// hasExcessiveCaps checks for excessive capitalization
func (s *ChatSecurityService) hasExcessiveCaps(content string) bool {
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

// hasExcessiveRepetition checks for excessive character repetition
func (s *ChatSecurityService) hasExcessiveRepetition(content string) bool {
	// Check for same character repeated more than 5 times
	repeatPattern := regexp.MustCompile(`(.)\1{5,}`)
	return repeatPattern.MatchString(content)
}

// determineSecurityLevel determines security level based on risk score
func (s *ChatSecurityService) determineSecurityLevel(riskScore float64) string {
	if riskScore >= 0.8 {
		return "high"
	} else if riskScore >= 0.5 {
		return "medium"
	} else if riskScore >= 0.2 {
		return "low"
	}
	return "minimal"
}