package external

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AIModerationService provides AI-powered content moderation using AWS Rekognition
type AIModerationService struct {
	client               *rekognition.Client
	bucket               string
	region               string
	confidenceThreshold   float64
	nsfwThreshold        float64
	violenceThreshold    float64
	adultThreshold       float64
	batchSize           int
	maxRetries          int
	retryDelay          time.Duration
}

// NewAIModerationService creates a new AI moderation service instance
func NewAIModerationService(region, bucket string, confidenceThreshold, nsfwThreshold, violenceThreshold, adultThreshold float64) (*AIModerationService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := rekognition.NewFromConfig(cfg)

	return &AIModerationService{
		client:             client,
		bucket:             bucket,
		region:             region,
		confidenceThreshold: confidenceThreshold,
		nsfwThreshold:      nsfwThreshold,
		violenceThreshold:  violenceThreshold,
		adultThreshold:     adultThreshold,
		batchSize:          10,
		maxRetries:         3,
		retryDelay:         time.Second * 2,
	}, nil
}

// ContentAnalysisRequest represents a request for content analysis
type ContentAnalysisRequest struct {
	ContentID   string            `json:"content_id"`
	ContentType string            `json:"content_type"` // "image", "video", "text"
	ContentURL  string            `json:"content_url"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
}

// ContentAnalysisResult represents the result of content analysis
type ContentAnalysisResult struct {
	ContentID       string                 `json:"content_id"`
	ContentType     string                 `json:"content_type"`
	IsApproved      bool                   `json:"is_approved"`
	Confidence      float64                `json:"confidence"`
	RiskLevel       string                 `json:"risk_level"` // "low", "medium", "high"
	Labels          []ModerationLabel      `json:"labels"`
	NSFWContent     bool                   `json:"nsfw_content"`
	AdultContent    bool                   `json:"adult_content"`
	ViolentContent  bool                   `json:"violent_content"`
	RequiresReview  bool                   `json:"requires_review"`
	ProcessingTime  time.Duration           `json:"processing_time"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// BatchAnalysisRequest represents a batch analysis request
type BatchAnalysisRequest struct {
	Requests []ContentAnalysisRequest `json:"requests"`
	Priority int                    `json:"priority"` // 1=high, 2=medium, 3=low
}

// BatchAnalysisResult represents the result of batch analysis
type BatchAnalysisResult struct {
	BatchID       string                  `json:"batch_id"`
	TotalItems    int                     `json:"total_items"`
	Processed     int                     `json:"processed"`
	Failed        int                     `json:"failed"`
	Results       []ContentAnalysisResult  `json:"results"`
	Errors        []BatchAnalysisError     `json:"errors"`
	ProcessingTime time.Duration           `json:"processing_time"`
	StartedAt     time.Time               `json:"started_at"`
	CompletedAt   time.Time               `json:"completed_at"`
}

// BatchAnalysisError represents an error in batch processing
type BatchAnalysisError struct {
	ContentID string `json:"content_id"`
	Error     string `json:"error"`
}

// ModerationRule represents a custom moderation rule
type ModerationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Conditions  []ModerationCondition   `json:"conditions"`
	Actions     []ModerationAction      `json:"actions"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ModerationCondition represents a condition in a moderation rule
type ModerationCondition struct {
	Type     string      `json:"type"`     // "label", "confidence", "user_age", "user_verification"
	Operator string      `json:"operator"` // "equals", "contains", "greater_than", "less_than"
	Value    interface{} `json:"value"`
}

// ModerationAction represents an action in a moderation rule
type ModerationAction struct {
	Type  string                 `json:"type"`  // "approve", "reject", "flag_for_review", "ban_user"
	Params map[string]interface{} `json:"params"`
}

// AnalyzeContent analyzes a single piece of content
func (s *AIModerationService) AnalyzeContent(ctx context.Context, req ContentAnalysisRequest) (*ContentAnalysisResult, error) {
	logger.Info("Starting content analysis", "content_id", req.ContentID, "content_type", req.ContentType)
	
	startTime := time.Now()
	
	var result *ContentAnalysisResult
	var err error
	
	switch req.ContentType {
	case "image":
		result, err = s.analyzeImage(ctx, req)
	case "video":
		result, err = s.analyzeVideo(ctx, req)
	case "text":
		result, err = s.analyzeText(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", req.ContentType)
	}
	
	if err != nil {
		logger.Error("Content analysis failed", err, "content_id", req.ContentID)
		return nil, fmt.Errorf("content analysis failed: %w", err)
	}
	
	result.ProcessingTime = time.Since(startTime)
	result.AnalyzedAt = time.Now()
	
	logger.Info("Content analysis completed", "content_id", req.ContentID, "approved", result.IsApproved, "confidence", result.Confidence)
	return result, nil
}

// AnalyzeContentBatch analyzes multiple pieces of content in batch
func (s *AIModerationService) AnalyzeContentBatch(ctx context.Context, req BatchAnalysisRequest) (*BatchAnalysisResult, error) {
	logger.Info("Starting batch content analysis", "total_items", len(req.Requests), "priority", req.Priority)
	
	startTime := time.Now()
	batchID := uuid.New().String()
	
	result := &BatchAnalysisResult{
		BatchID:    batchID,
		TotalItems:  len(req.Requests),
		Results:     make([]ContentAnalysisResult, 0),
		Errors:      make([]BatchAnalysisError, 0),
		StartedAt:   time.Now(),
	}
	
	// Process in batches to avoid overwhelming the service
	for i := 0; i < len(req.Requests); i += s.batchSize {
		end := i + s.batchSize
		if end > len(req.Requests) {
			end = len(req.Requests)
		}
		
		batch := req.Requests[i:end]
		
		for _, req := range batch {
			analysisResult, err := s.AnalyzeContent(ctx, req)
			if err != nil {
				result.Errors = append(result.Errors, BatchAnalysisError{
					ContentID: req.ContentID,
					Error:     err.Error(),
				})
				result.Failed++
				continue
			}
			
			result.Results = append(result.Results, *analysisResult)
			result.Processed++
		}
	}
	
	result.CompletedAt = time.Now()
	result.ProcessingTime = time.Since(startTime)
	
	logger.Info("Batch content analysis completed", "batch_id", batchID, "processed", result.Processed, "failed", result.Failed)
	return result, nil
}

// analyzeImage analyzes an image for inappropriate content
func (s *AIModerationService) analyzeImage(ctx context.Context, req ContentAnalysisRequest) (*ContentAnalysisResult, error) {
	input := &rekognition.DetectModerationLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(req.ContentURL),
			},
		},
		MinConfidence: aws.Float64(s.confidenceThreshold),
	}

	var moderationResult *rekognition.DetectModerationLabelsOutput
	var err error
	
	// Retry logic
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		moderationResult, err = s.client.DetectModerationLabels(ctx, input)
		if err == nil {
			break
		}
		
		if attempt < s.maxRetries-1 {
			logger.Warn("Retrying image analysis", "attempt", attempt+1, "content_id", req.ContentID, "error", err)
			time.Sleep(s.retryDelay)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image after %d attempts: %w", s.maxRetries, err)
	}

	result := &ContentAnalysisResult{
		ContentID:      req.ContentID,
		ContentType:    req.ContentType,
		IsApproved:     true,
		Confidence:     100.0,
		RiskLevel:      "low",
		Labels:         make([]ModerationLabel, 0),
		NSFWContent:    false,
		AdultContent:   false,
		ViolentContent: false,
		RequiresReview: false,
		Recommendations: make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}

	if len(moderationResult.ModerationLabels) > 0 {
		minConfidence := 100.0
		hasInappropriateContent := false
		
		for _, label := range moderationResult.ModerationLabels {
			labelName := *label.Name
			confidence := float64(*label.Confidence)
			
			// Add to labels
			result.Labels = append(result.Labels, ModerationLabel{
				Name:       labelName,
				Confidence: confidence,
				ParentName: aws.ToString(label.ParentName, ""),
			})
			
			// Check for inappropriate content
			if s.isInappropriateContent(labelName) {
				hasInappropriateContent = true
				result.NSFWContent = result.NSFWContent || s.isNSFWContent(labelName)
				result.AdultContent = result.AdultContent || s.isAdultContent(labelName)
				result.ViolentContent = result.ViolentContent || s.isViolentContent(labelName)
			}
			
			if confidence < minConfidence {
				minConfidence = confidence
			}
		}
		
		if hasInappropriateContent {
			result.IsApproved = false
			result.Confidence = 100.0 - minConfidence
			result.RiskLevel = s.calculateRiskLevel(result)
			result.RequiresReview = result.Confidence < (100.0 - s.confidenceThreshold)
			
			// Generate recommendations
			result.Recommendations = s.generateRecommendations(result)
		}
	}
	
	// Add metadata
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			result.Metadata[k] = v
		}
	}
	
	return result, nil
}

// analyzeVideo analyzes a video for inappropriate content
func (s *AIModerationService) analyzeVideo(ctx context.Context, req ContentAnalysisRequest) (*ContentAnalysisResult, error) {
	// For video analysis, we'll use StartContentModeration to analyze the video
	// This is an asynchronous operation, but for simplicity, we'll simulate it
	
	input := &rekognition.StartContentModerationInput{
		Video: &types.Video{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(req.ContentURL),
			},
		},
		MinConfidence: aws.Float64(s.confidenceThreshold),
	}

	_, err := s.client.StartContentModeration(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start video moderation: %w", err)
	}
	
	// For now, return a placeholder result
	// In a real implementation, you would need to handle the asynchronous nature
	// and check the results using GetContentModeration
	return &ContentAnalysisResult{
		ContentID:      req.ContentID,
		ContentType:    req.ContentType,
		IsApproved:     true,
		Confidence:     100.0,
		RiskLevel:      "low",
		Labels:         make([]ModerationLabel, 0),
		NSFWContent:    false,
		AdultContent:   false,
		ViolentContent: false,
		RequiresReview: true, // Videos always require manual review
		Recommendations: []string{"Manual review recommended for video content"},
		Metadata:       make(map[string]interface{}),
	}, nil
}

// analyzeText analyzes text for inappropriate content
func (s *AIModerationService) analyzeText(ctx context.Context, req ContentAnalysisRequest) (*ContentAnalysisResult, error) {
	// AWS Rekognition doesn't directly support text moderation
	// In a real implementation, you would use Amazon Comprehend for text analysis
	// For now, we'll implement basic text filtering
	
	text := req.ContentURL // Assuming the text is passed in the URL field for simplicity
	
	result := &ContentAnalysisResult{
		ContentID:      req.ContentID,
		ContentType:    req.ContentType,
		IsApproved:     true,
		Confidence:     100.0,
		RiskLevel:      "low",
		Labels:         make([]ModerationLabel, 0),
		NSFWContent:    false,
		AdultContent:   false,
		ViolentContent: false,
		RequiresReview: false,
		Recommendations: make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}
	
	// Basic profanity and inappropriate content detection
	inappropriateWords := []string{
		"profanity1", "profanity2", "profanity3", // Add actual inappropriate words
		"hate", "violence", "threat",
	}
	
	textLower := strings.ToLower(text)
	hasInappropriateContent := false
	
	for _, word := range inappropriateWords {
		if strings.Contains(textLower, word) {
			hasInappropriateContent = true
			result.Labels = append(result.Labels, ModerationLabel{
				Name:       "Inappropriate Text",
				Confidence: 80.0,
				ParentName: "Text Moderation",
			})
			break
		}
	}
	
	if hasInappropriateContent {
		result.IsApproved = false
		result.Confidence = 80.0
		result.RiskLevel = "medium"
		result.RequiresReview = true
		result.Recommendations = append(result.Recommendations, "Manual review recommended for text content")
	}
	
	return result, nil
}

// calculateRiskLevel calculates the risk level based on content analysis
func (s *AIModerationService) calculateRiskLevel(result *ContentAnalysisResult) string {
	if result.NSFWContent || result.ViolentContent {
		return "high"
	}
	if result.AdultContent {
		return "medium"
	}
	return "low"
}

// generateRecommendations generates recommendations based on content analysis
func (s *AIModerationService) generateRecommendations(result *ContentAnalysisResult) []string {
	recommendations := make([]string, 0)
	
	if result.NSFWContent {
		recommendations = append(recommendations, "Remove NSFW content immediately")
	}
	if result.AdultContent {
		recommendations = append(recommendations, "Age restriction may be required")
	}
	if result.ViolentContent {
		recommendations = append(recommendations, "Violent content violates community guidelines")
	}
	if result.RequiresReview {
		recommendations = append(recommendations, "Manual review recommended")
	}
	
	return recommendations
}

// isInappropriateContent checks if a label represents inappropriate content
func (s *AIModerationService) isInappropriateContent(labelName string) bool {
	inappropriateLabels := []string{
		"Explicit Nudity",
		"Illicit Drugs",
		"Weapons",
		"Violence",
		"Tobacco",
		"Alcohol",
		"Gambling",
		"Hate Symbols",
		"Terrorism",
		"Graphic Violence",
		"Sexual Activity",
		"Adult Toys",
	}

	for _, label := range inappropriateLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

// isNSFWContent checks if a label represents NSFW content
func (s *AIModerationService) isNSFWContent(labelName string) bool {
	nsfwLabels := []string{
		"Explicit Nudity",
		"Graphic Violence",
		"Sexual Activity",
		"Adult Toys",
	}

	for _, label := range nsfwLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

// isAdultContent checks if a label represents adult content
func (s *AIModerationService) isAdultContent(labelName string) bool {
	adultLabels := []string{
		"Alcohol",
		"Tobacco",
		"Gambling",
		"Adult",
		"Partial Nudity",
	}

	for _, label := range adultLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

// isViolentContent checks if a label represents violent content
func (s *AIModerationService) isViolentContent(labelName string) bool {
	violentLabels := []string{
		"Violence",
		"Weapons",
		"Terrorism",
		"Graphic Violence",
		"Physical Violence",
		"Blood",
		"Gore",
	}

	for _, label := range violentLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

// ApplyCustomRules applies custom moderation rules to content analysis results
func (s *AIModerationService) ApplyCustomRules(result *ContentAnalysisResult, rules []ModerationRule) *ContentAnalysisResult {
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		if s.evaluateRule(result, rule) {
			result = s.applyRuleActions(result, rule)
		}
	}
	
	return result
}

// evaluateRule evaluates if a moderation rule applies to the content
func (s *AIModerationService) evaluateRule(result *ContentAnalysisResult, rule ModerationRule) bool {
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(result, condition) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (s *AIModerationService) evaluateCondition(result *ContentAnalysisResult, condition ModerationCondition) bool {
	switch condition.Type {
	case "label":
		return s.evaluateLabelCondition(result, condition)
	case "confidence":
		return s.evaluateConfidenceCondition(result, condition)
	case "risk_level":
		return s.evaluateRiskLevelCondition(result, condition)
	default:
		return false
	}
}

// evaluateLabelCondition evaluates label-based conditions
func (s *AIModerationService) evaluateLabelCondition(result *ContentAnalysisResult, condition ModerationCondition) bool {
	labelName, ok := condition.Value.(string)
	if !ok {
		return false
	}
	
	switch condition.Operator {
	case "contains":
		for _, label := range result.Labels {
			if strings.Contains(strings.ToLower(label.Name), strings.ToLower(labelName)) {
				return true
			}
		}
		return false
	case "equals":
		for _, label := range result.Labels {
			if strings.EqualFold(label.Name, labelName) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// evaluateConfidenceCondition evaluates confidence-based conditions
func (s *AIModerationService) evaluateConfidenceCondition(result *ContentAnalysisResult, condition ModerationCondition) bool {
	threshold, ok := condition.Value.(float64)
	if !ok {
		return false
	}
	
	switch condition.Operator {
	case "greater_than":
		return result.Confidence > threshold
	case "less_than":
		return result.Confidence < threshold
	case "equals":
		return result.Confidence == threshold
	default:
		return false
	}
}

// evaluateRiskLevelCondition evaluates risk level conditions
func (s *AIModerationService) evaluateRiskLevelCondition(result *ContentAnalysisResult, condition ModerationCondition) bool {
	riskLevel, ok := condition.Value.(string)
	if !ok {
		return false
	}
	
	switch condition.Operator {
	case "equals":
		return strings.EqualFold(result.RiskLevel, riskLevel)
	case "contains":
		return strings.Contains(strings.ToLower(result.RiskLevel), strings.ToLower(riskLevel))
	default:
		return false
	}
}

// applyRuleActions applies actions from a moderation rule
func (s *AIModerationService) applyRuleActions(result *ContentAnalysisResult, rule ModerationRule) *ContentAnalysisResult {
	for _, action := range rule.Actions {
		switch action.Type {
		case "approve":
			result.IsApproved = true
		case "reject":
			result.IsApproved = false
		case "flag_for_review":
			result.RequiresReview = true
		case "set_risk_level":
			if riskLevel, ok := action.Params["level"].(string); ok {
				result.RiskLevel = riskLevel
			}
		}
	}
	
	return result
}