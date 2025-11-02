
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ContentAnalysisService provides real-time content scanning and analysis
type ContentAnalysisService struct {
	aiModerationService *external.AIModerationService
	cacheService       CacheService
	rateLimiter        RateLimiter
	profanityFilter    *ProfanityFilter
	linkAnalyzer       *LinkAnalyzer
	piiDetector        *PIIDetector
	config            ContentAnalysisConfig
}

// ContentAnalysisConfig represents configuration for content analysis
type ContentAnalysisConfig struct {
	EnableRealTimeAnalysis    bool          `json:"enable_real_time_analysis"`
	EnableProfanityFilter     bool          `json:"enable_profanity_filter"`
	EnableLinkAnalysis        bool          `json:"enable_link_analysis"`
	EnablePIIDetection       bool          `json:"enable_pii_detection"`
	MaxTextLength           int           `json:"max_text_length"`
	MaxLinksPerMessage       int           `json:"max_links_per_message"`
	AnalysisTimeout         time.Duration `json:"analysis_timeout"`
	CacheResults           bool          `json:"cache_results"`
	CacheTTL              time.Duration `json:"cache_ttl"`
	BatchSize              int           `json:"batch_size"`
	MaxConcurrentAnalyses  int           `json:"max_concurrent_analyses"`
}

// ContentRequest represents a content analysis request
type ContentRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "text", "image", "video", "profile"
	Content     string                 `json:"content"`
	UserID      string                 `json:"user_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Priority    int                    `json:"priority"` // 1=high, 2=medium, 3=low
	Timestamp   time.Time              `json:"timestamp"`
}

// ContentAnalysisResponse represents the response from content analysis
type ContentAnalysisResponse struct {
	RequestID       string                 `json:"request_id"`
	IsApproved      bool                   `json:"is_approved"`
	Confidence      float64                `json:"confidence"`
	RiskLevel       string                 `json:"risk_level"`
	Violations      []ContentViolation      `json:"violations"`
	RequiresReview  bool                   `json:"requires_review"`
	Recommendations []string               `json:"recommendations"`
	ProcessingTime  time.Duration           `json:"processing_time"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// ContentViolation represents a specific content violation
type ContentViolation struct {
	Type        string  `json:"type"`        // "profanity", "inappropriate_content", "pii", "malicious_link"
	Severity    string  `json:"severity"`    // "low", "medium", "high", "critical"
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Position    *int    `json:"position,omitempty"` // Position in text for text violations
	Context     string  `json:"context,omitempty"`  // Context around the violation
}

// ProfanityFilter handles profanity detection in text
type ProfanityFilter struct {
	blockedWords map[string]bool
	patterns    []string
	enabled      bool
}

// LinkAnalyzer analyzes URLs for safety and malicious content
type LinkAnalyzer struct {
	allowedDomains    []string
	blockedDomains   []string
	safeBrowsingAPI  string
	enabled          bool
}

// PIIDetector detects personally identifiable information
type PIIDetector struct {
	patterns map[string]string // pattern name -> regex pattern
	enabled  bool
}

// NewContentAnalysisService creates a new content analysis service
func NewContentAnalysisService(
	aiModerationService *external.AIModerationService,
	cacheService CacheService,
	rateLimiter RateLimiter,
	config ContentAnalysisConfig,
) *ContentAnalysisService {
	service := &ContentAnalysisService{
		aiModerationService: aiModerationService,
		cacheService:       cacheService,
		rateLimiter:        rateLimiter,
		config:            config,
		profanityFilter:    NewProfanityFilter(config.EnableProfanityFilter),
		linkAnalyzer:       NewLinkAnalyzer(config.EnableLinkAnalysis),
		piiDetector:        NewPIIDetector(config.EnablePIIDetection),
	}

	return service
}

// AnalyzeContent analyzes content in real-time
func (s *ContentAnalysisService) AnalyzeContent(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	logger.Info("Starting content analysis", "request_id", req.ID, "type", req.Type, "user_id", req.UserID)
	
	startTime := time.Now()
	
	// Check cache first if enabled
	if s.config.CacheResults {
		if cached, err := s.getCachedResult(ctx, req); err == nil && cached != nil {
			logger.Info("Content analysis result retrieved from cache", "request_id", req.ID)
			cached.ProcessingTime = time.Since(startTime)
			return cached, nil
		}
	}
	
	// Check rate limits
	if err := s.rateLimiter.CheckRateLimit(ctx, "content_analysis", req.UserID, 100, time.Minute); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}
	
	// Perform analysis based on content type
	var response *ContentAnalysisResponse
	var err error
	
	switch req.Type {
	case "text":
		response, err = s.analyzeTextContent(ctx, req)
	case "image":
		response, err = s.analyzeImageContent(ctx, req)
	case "video":
		response, err = s.analyzeVideoContent(ctx, req)
	case "profile":
		response, err = s.analyzeProfileContent(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", req.Type)
	}
	
	if err != nil {
		logger.Error("Content analysis failed", err, "request_id", req.ID)
		return nil, fmt.Errorf("content analysis failed: %w", err)
	}
	
	response.ProcessingTime = time.Since(startTime)
	response.AnalyzedAt = time.Now()
	
	// Cache result if enabled
	if s.config.CacheResults {
		s.cacheResult(ctx, req, response)
	}
	
	logger.Info("Content analysis completed", "request_id", req.ID, "approved", response.IsApproved, "violations", len(response.Violations))
	return response, nil
}

// AnalyzeContentBatch analyzes multiple content items in batch
func (s *ContentAnalysisService) AnalyzeContentBatch(ctx context.Context, requests []ContentRequest) ([]*ContentAnalysisResponse, error) {
	logger.Info("Starting batch content analysis", "total_items", len(requests))
	
	responses := make([]*ContentAnalysisResponse, 0, len(requests))
	
	// Process in batches to avoid overwhelming the system
	for i := 0; i < len(requests); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(requests) {
			end = len(requests)
		}
		
		batch := requests[i:end]
		
		for _, req := range batch {
			response, err := s.AnalyzeContent(ctx, req)
			if err != nil {
				logger.Error("Batch item analysis failed", err, "request_id", req.ID)
				// Create error response
				errorResponse := &ContentAnalysisResponse{
					RequestID:      req.ID,
					IsApproved:     false,
					Confidence:     0,
					RiskLevel:      "high",
					Violations:     []ContentViolation{{Type: "analysis_error", Severity: "critical", Description: err.Error()}},
					RequiresReview: true,
					AnalyzedAt:     time.Now(),
				}
				responses = append(responses, errorResponse)
				continue
			}
			
			responses = append(responses, response)
		}
	}
	
	return responses, nil
}

// analyzeTextContent analyzes text content
func (s *ContentAnalysisService) analyzeTextContent(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	response := &ContentAnalysisResponse{
		RequestID:       req.ID,
		IsApproved:      true,
		Confidence:      100.0,
		RiskLevel:       "low",
		Violations:      make([]ContentViolation, 0),
		RequiresReview:   false,
		Recommendations:  make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}
	
	// Check text length
	if len(req.Content) > s.config.MaxTextLength {
		response.IsApproved = false
		response.RiskLevel = "medium"
		response.Violations = append(response.Violations, ContentViolation{
			Type:        "content_too_long",
			Severity:    "medium",
			Description: fmt.Sprintf("Text exceeds maximum length of %d characters", s.config.MaxTextLength),
			Confidence:  100.0,
		})
	}
	
	// Profanity filter
	if s.config.EnableProfanityFilter {
		violations := s.profanityFilter.Analyze(req.Content)
		for _, violation := range violations {
			response.Violations = append(response.Violations, ContentViolation{
				Type:        "profanity",
				Severity:    "medium",
				Description: violation.Word,
				Confidence:  violation.Confidence,
				Position:    violation.Position,
				Context:     violation.Context,
			})
		}
	}
	
	// PII detection
	if s.config.EnablePIIDetection {
		violations := s.piiDetector.Analyze(req.Content)
		for _, violation := range violations {
			response.Violations = append(response.Violations, ContentViolation{
				Type:        "pii",
				Severity:    "high",
				Description: fmt.Sprintf("PII detected: %s", violation.Type),
				Confidence:  violation.Confidence,
				Position:    violation.Position,
				Context:     violation.Context,
			})
		}
	}
	
	// Link analysis
	if s.config.EnableLinkAnalysis {
		violations := s.linkAnalyzer.Analyze(req.Content)
		for _, violation := range violations {
			response.Violations = append(response.Violations, ContentViolation{
				Type:        "malicious_link",
				Severity:    violation.Severity,
				Description: violation.Description,
				Confidence:  violation.Confidence,
				Context:     violation.URL,
			})
		}
	}
	
	// Evaluate overall approval
	if len(response.Violations) > 0 {
		response.IsApproved = false
		response.RiskLevel = s.calculateOverallRiskLevel(response.Violations)
		response.RequiresReview = s.requiresManualReview(response.Violations)
		response.Recommendations = s.generateRecommendations(response.Violations)
	}
	
	return response, nil
}

// analyzeImageContent analyzes image content
func (s *ContentAnalysisService) analyzeImageContent(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	// Use AI moderation service for image analysis
	aiReq := external.ContentAnalysisRequest{
		ContentID:   req.ID,
		ContentType:  "image",
		ContentURL:   req.Content,
		Metadata:    req.Metadata,
		UserID:      req.UserID,
	}
	
	aiResult, err := s.aiModerationService.AnalyzeContent(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI image analysis failed: %w", err)
	}
	
	// Convert AI result to our response format
	response := &ContentAnalysisResponse{
		RequestID:       req.ID,
		IsApproved:      aiResult.IsApproved,
		Confidence:      aiResult.Confidence,
		RiskLevel:       aiResult.RiskLevel,
		Violations:      make([]ContentViolation, 0),
		RequiresReview:   aiResult.RequiresReview,
		Recommendations:  aiResult.Recommendations,
		Metadata:        aiResult.Metadata,
	}
	
	// Convert AI labels to violations
	for _, label := range aiResult.Labels {
		violation := ContentViolation{
			Type:        "inappropriate_content",
			Severity:    s.getLabelSeverity(label.Name),
			Description: label.Name,
			Confidence:  label.Confidence,
		}
		response.Violations = append(response.Violations, violation)
	}
	
	return response, nil
}

// analyzeVideoContent analyzes video content
func (s *ContentAnalysisService) analyzeVideoContent(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	// Use AI moderation service for video analysis
	aiReq := external.ContentAnalysisRequest{
		ContentID:   req.ID,
		ContentType:  "video",
		ContentURL:   req.Content,
		Metadata:    req.Metadata,
		UserID:      req.UserID,
	}
	
	aiResult, err := s.aiModerationService.AnalyzeContent(ctx, aiReq)
	if err != nil {
		return nil, fmt.Errorf("AI video analysis failed: %w", err)
	}
	
	// Convert AI result to our response format
	response := &ContentAnalysisResponse{
		RequestID:       req.ID,
		IsApproved:      aiResult.IsApproved,
		Confidence:      aiResult.Confidence,
		RiskLevel:       aiResult.RiskLevel,
		Violations:      make([]ContentViolation, 0),
		RequiresReview:   true, // Videos always require review
		Recommendations:  append(aiResult.Recommendations, "Manual review recommended for video content"),
		Metadata:        aiResult.Metadata,
	}
	
	// Convert AI labels to violations
	for _, label := range aiResult.Labels {
		violation := ContentViolation{
			Type:        "inappropriate_content",
			Severity:    s.getLabelSeverity(label.Name),
			Description: label.Name,
			Confidence:  label.Confidence,
		}
		response.Violations = append(response.Violations, violation)
	}
	
	return response, nil
}

// analyzeProfileContent analyzes user profile content
func (s *ContentAnalysisService) analyzeProfileContent(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	// Profile analysis combines text analysis of bio and image analysis of photos
	response := &ContentAnalysisResponse{
		RequestID:       req.ID,
		IsApproved:      true,
		Confidence:      100.0,
		RiskLevel:       "low",
		Violations:      make([]ContentViolation, 0),
		RequiresReview:   false,
		Recommendations:  make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}
	
	// Analyze bio text if present
	if bio, ok := req.Metadata["bio"].(string); ok && bio != "" {
		textReq := ContentRequest{
			ID:       req.ID + "_bio",
			Type:     "text",
			Content:  bio,
			UserID:   req.UserID,
			Metadata: req.Metadata,
		}
		
		bioResponse, err := s.analyzeTextContent(ctx, textReq)
		if err != nil {
			logger.Error("Profile bio analysis failed", err, "user_id", req.UserID)
		} else {
			// Merge violations
			response.Violations = append(response.Violations, bioResponse.Violations...)
		}
	}
	
	// Analyze profile photos if present
	if photos, ok := req.Metadata["photos"].([]string); ok {
		for i, photoURL := range photos {
			imageReq := ContentRequest{
				ID:       fmt.Sprintf("%s_photo_%d", req.ID, i),
				Type:     "image",
				Content:  photoURL,
				UserID:   req.UserID,
				Metadata: req.Metadata,
			}
			
			imageResponse, err := s.analyzeImageContent(ctx, imageReq)
			if err != nil {
				logger.Error("Profile photo analysis failed", err, "user_id", req.UserID, "photo_index", i)
			} else {
				// Merge violations
				response.Violations = append(response.Violations, imageResponse.Violations...)
			}
		}
	}
	
	// Evaluate overall approval
	if len(response.Violations) > 0 {
		response.IsApproved = false
		response.RiskLevel = s.calculateOverallRiskLevel(response.Violations)
		response.RequiresReview = s.requiresManualReview(response.Violations)
		response.Recommendations = s.generateRecommendations(response.Violations)
	}
	
	return response, nil
}

// calculateOverallRiskLevel calculates the overall risk level based on violations
func (s *ContentAnalysisService) calculateOverallRiskLevel(violations []ContentViolation) string {
	hasCritical := false
	hasHigh := false
	hasMedium := false
	
	for _, violation := range violations {
		switch violation.Severity {
		case "critical":
			hasCritical = true
		case "high":
			hasHigh = true
		case "medium":
			hasMedium = true
		}
	}
	
	if hasCritical {
		return "critical"
	}
	if hasHigh {
		return "high"
	}
	if hasMedium {
		return "medium"
	}
	return "low"
}

// requiresManualReview determines if content requires manual review
func (s *ContentAnalysisService) requiresManualReview(violations []ContentViolation) bool {
	for _, violation := range violations {
		if violation.Severity == "critical" || violation.Severity == "high" {
			return true
		}
		if violation.Type == "malicious_link" || violation.Type == "pii" {
			return true
		}
	}
	return false
}

// generateRecommendations generates recommendations based on violations
func (s *ContentAnalysisService) generateRecommendations(violations []ContentViolation) []string {
	recommendations := make([]string, 0)
	
	hasProfanity := false
	hasPII := false
	hasMaliciousLink := false
	hasInappropriateContent := false
	
	for _, violation := range violations {
		switch violation.Type {
		case "profanity":
			hasProfanity = true
		case "pii":
			hasPII = true
		case "malicious_link":
			hasMaliciousLink = true
		case "inappropriate_content":
			hasInappropriateContent = true
		}
	}
	
	if hasProfanity {
		recommendations = append(recommendations, "Remove or replace inappropriate language")
	}
	if hasPII {
		recommendations = append(recommendations, "Remove personally identifiable information")
	}
	if hasMaliciousLink {
		recommendations = append(recommendations, "Remove or verify suspicious links")
	}
	if hasInappropriateContent {
		recommendations = append(recommendations, "Review content for appropriateness")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Manual review recommended")
	}
	
	return recommendations
}

// getLabelSeverity determines severity based on AI label
func (s *ContentAnalysisService) getLabelSeverity(labelName string) string {
	nsfwLabels := []string{"Explicit Nudity", "Sexual Activity", "Adult Toys"}
	violentLabels := []string{"Violence", "Weapons", "Terrorism", "Graphic Violence"}
	
	labelLower := strings.ToLower(labelName)
	
	for _, label := range nsfwLabels {
		if strings.Contains(labelLower, strings.ToLower(label)) {
			return "high"
		}
	}
	
	for _, label := range violentLabels {
		if strings.Contains(labelLower, strings.ToLower(label)) {
			return "high"
		}
	}
	
	return "medium"
}

// getCachedResult retrieves cached analysis result
func (s *ContentAnalysisService) getCachedResult(ctx context.Context, req ContentRequest) (*ContentAnalysisResponse, error) {
	cacheKey := fmt.Sprintf("content_analysis:%s:%s", req.Type, req.ID)
	
	var cached ContentAnalysisResponse
	err := s.cacheService.Get(ctx, cacheKey, &cached)
	if err != nil {
		return nil, err
	}
	
	return &cached, nil
}

// cacheResult stores analysis result in cache
func (s *ContentAnalysisService) cacheResult(ctx context.Context, req ContentRequest, response *ContentAnalysisResponse) {
	cacheKey := fmt.Sprintf("content_analysis:%s:%s", req.Type, req.ID)
	s.cacheService.Set(ctx, cacheKey, response, s.config.CacheTTL)
}

// ProfanityFilter implementation

// ProfanityViolation represents a profanity violation
type ProfanityViolation struct {
	Word       string  `json:"word"`
	Confidence float64 `json:"confidence"`
	Position   *int    `json:"position,omitempty"`
	Context    string  `json:"context,omitempty"`
}

// NewProfanityFilter creates a new profanity filter
func NewProfanityFilter(enabled bool) *ProfanityFilter {
	blockedWords := map[string]bool{
		"profanity1": true,
		"profanity2": true,
		"profanity3": true,
		// Add actual profanity words here
	}
	
	return &ProfanityFilter{
		blockedWords: blockedWords,
		patterns:    []string{},
		enabled:      enabled,
	}
}

// Analyze analyzes text for profanity
func (pf *ProfanityFilter) Analyze(text string) []ProfanityViolation {
	if !pf.enabled {
		return nil
	}
	
	violations := make([]ProfanityViolation, 0)
	words := strings.Fields(strings.ToLower(text))
	
	for i, word := range words {
		if pf.blockedWords[word] {
			violations = append(violations, ProfanityViolation{
				Word:       word,
				Confidence: 90.0,
				Position:   &i,
				Context:    pf.getContext(text, i),
			})
		}
	}
	
	return violations
}

// getContext gets context around a word
func (pf *ProfanityFilter) getContext(text string, position int) string {
	words := strings.Fields(text)
	start := position - 2
	if start < 0 {
		start = 0
	}
	
	end := position + 3
	if end > len(words) {
		end = len(words)
	}
	
	contextWords := words[start:end]
	return strings.Join(contextWords, " ")
}

// LinkAnalyzer implementation

// LinkViolation represents a link violation
type LinkViolation struct {
	URL         string  `json:"url"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// NewLinkAnalyzer creates a new link analyzer
func NewLinkAnalyzer(enabled bool) *LinkAnalyzer {
	allowedDomains := []string{
		"youtube.com",
		"instagram.com",
		"facebook.com",
		"twitter.com",
		"linkedin.com",
	}
	
	blockedDomains := []string{
		"malicious-site.com",
		"spam-site.net",
		// Add actual blocked domains here
	}
	
	return &LinkAnalyzer{
		allowedDomains:   allowedDomains,
		blockedDomains:  blockedDomains,
		safeBrowsingAPI: "", // Configure with actual API
		enabled:         enabled,
	}
}

// Analyze analyzes text for malicious links
func (la *LinkAnalyzer) Analyze(text string) []LinkViolation {
	if !la.enabled {
		return nil
	}
	
	violations := make([]LinkViolation, 0)
	links := la.extractLinks(text)
	
	for _, link := range links {
		violation := la.analyzeLink(link)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}
	
	return violations
}

// extractLinks extracts URLs from text
func (la *LinkAnalyzer) extractLinks(text string) []string {
	// Simple URL extraction - in production, use proper regex
	links := make([]string, 0)
	words := strings.Fields(text)
	
	for _, word := range words {
		if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
			links = append(links, word)
		}
	}
	
	return links
}

// analyzeLink analyzes a single link
func (la *LinkAnalyzer) analyzeLink(url string) *LinkViolation {
	// Check blocked domains
	for _, domain := range la.blockedDomains {
		if strings.Contains(url, domain) {
			return &LinkViolation{
				URL:         url,
				Severity:    "high",
				Description: "Blocked domain detected",
				Confidence:  95.0,
			}
		}
	}
	
	// Check if it's an allowed domain
	for _, domain := range la.allowedDomains {
		if strings.Contains(url, domain) {
			return nil // Allowed
		}
	}
	
	// Unknown domain - flag for review
	return &LinkViolation{
		URL:         url,
		Severity:    "medium",
		Description: "Unknown domain - manual review recommended",
		Confidence:  60.0,
	}
}

// PIIDetector implementation

// PIIViolation represents a PII violation
type PIIViolation struct {
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Position   *int    `json:"position,omitempty"`
	Context    string  `json:"context,omitempty"`
}

// NewPIIDetector creates a new PII detector
func NewPIIDetector(enabled bool) *PIIDetector {
	patterns := map[string]string{
		"email":    `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		"phone":    `\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`,
		"ssn":     `\b\d{3}-\d{2}-\d{4}\b`,
		"credit_card": `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,
	}
	
	return &PIIDetector{
		patterns: patterns,
		enabled:   enabled,
	}
}

// Analyze analyzes text for PII
func (pd *PIIDetector) Analyze(text string) []PIIViolation {
	if !pd.enabled {
		return nil
	}
	
	violations := make([]PIIViolation, 0)
	
	for patternType, pattern := range pd.patterns {
		// In production, use proper regex matching
		if pd.containsPattern(text, pattern) {
			violations = append(violations, PIIViolation{
				Type:       patternType,
				Confidence: 85.0,
				Context:    "PII pattern detected",
			})
		}
	}
	
	return violations
}

// containsPattern checks if text contains a pattern
func (pd *PIIDetector) containsPattern(text, pattern string) bool {
	// Simplified pattern matching - in production, use regex
	return strings.Contains(text, pattern) || len(text) > 0
}