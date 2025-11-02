package external

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AIService provides AI-powered verification capabilities using AWS Rekognition
type AIService struct {
	client       *rekognition.Client
	bucket       string
	region        string
	confidenceThreshold float64
}

// NewAIService creates a new AI service instance
func NewAIService(region, bucket string, confidenceThreshold float64) (*AIService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := rekognition.NewFromConfig(cfg)

	return &AIService{
		client:       client,
		bucket:       bucket,
		region:        region,
		confidenceThreshold: confidenceThreshold,
	}, nil
}

// FaceAnalysisResult represents the result of face analysis
type FaceAnalysisResult struct {
	HasFace      bool    `json:"has_face"`
	Confidence    float64 `json:"confidence"`
	AgeRange      *AgeRange `json:"age_range,omitempty"`
	Gender        *Gender  `json:"gender,omitempty"`
	Emotions      []string `json:"emotions,omitempty"`
	Quality       *Quality  `json:"quality,omitempty"`
	Pose          *Pose     `json:"pose,omitempty"`
}

// AgeRange represents the estimated age range
type AgeRange struct {
	Low  int `json:"low"`
	High int `json:"high"`
}

// Gender represents the detected gender
type Gender struct {
	Value string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

// Quality represents face quality metrics
type Quality struct {
	Brightness float64 `json:"brightness"`
	Sharpness  float64 `json:"sharpness"`
}

// Pose represents face pose information
type Pose struct {
	Roll  float64 `json:"roll"`
	Yaw   float64 `json:"yaw"`
	Pitch float64 `json:"pitch"`
}

// FaceComparisonResult represents the result of face comparison
type FaceComparisonResult struct {
	Similarity    float64 `json:"similarity"`
	IsMatch       bool     `json:"is_match"`
	Confidence    float64 `json:"confidence"`
	Details       string   `json:"details"`
}

// ModerationResult represents the result of content moderation
type ModerationResult struct {
	IsAppropriate bool                    `json:"is_appropriate"`
	Confidence    float64                  `json:"confidence"`
	Labels        []ModerationLabel         `json:"labels"`
	NSFWContent   bool                     `json:"nsfw_content"`
	AdultContent  bool                     `json:"adult_content"`
	ViolentContent bool                     `json:"violent_content"`
}

// ModerationLabel represents a moderation label
type ModerationLabel struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	ParentName string  `json:"parent_name,omitempty"`
}

// LivenessResult represents the result of liveness detection
type LivenessResult struct {
	IsLive       bool    `json:"is_live"`
	Confidence   float64 `json:"confidence"`
	Challenges   []string `json:"challenges,omitempty"`
	Details      string   `json:"details"`
}

// DocumentAnalysisResult represents the result of document analysis
type DocumentAnalysisResult struct {
	DocumentType string                 `json:"document_type"`
	Confidence   float64                `json:"confidence"`
	Fields       map[string]interface{}   `json:"fields"`
	IsValid      bool                    `json:"is_valid"`
	Details      string                  `json:"details"`
}

// AnalyzeFace analyzes a face in an image
func (s *AIService) AnalyzeFace(ctx context.Context, imageKey string) (*FaceAnalysisResult, error) {
	logger.Info("Analyzing face in image", "image_key", imageKey)

	input := &rekognition.DetectFacesInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(imageKey),
			},
		},
		Attributes: []types.Attribute{
			types.AttributeAgeRange,
			types.AttributeGender,
			types.AttributeEmotions,
			types.AttributeQuality,
			types.AttributePose,
		},
	}

	result, err := s.client.DetectFaces(ctx, input)
	if err != nil {
		logger.Error("Failed to detect faces", err, "image_key", imageKey)
		return nil, fmt.Errorf("failed to detect faces: %w", err)
	}

	if len(result.FaceDetails) == 0 {
		return &FaceAnalysisResult{
			HasFace:   false,
			Confidence: 0,
		}, nil
	}

	face := result.FaceDetails[0]
	analysis := &FaceAnalysisResult{
		HasFace:   true,
		Confidence: float64(*face.Confidence),
	}

	// Extract age range
	if face.AgeRange != nil {
		analysis.AgeRange = &AgeRange{
			Low:  int(*face.AgeRange.Low),
			High: int(*face.AgeRange.High),
		}
	}

	// Extract gender
	if face.Gender != nil {
		analysis.Gender = &Gender{
			Value:      *face.Gender.Value,
			Confidence: float64(*face.Gender.Confidence),
		}
	}

	// Extract emotions
	if len(face.Emotions) > 0 {
		emotions := make([]string, 0, len(face.Emotions))
		for _, emotion := range face.Emotions {
			if *emotion.Confidence > 50 { // Only include emotions with confidence > 50%
				emotions = append(emotions, *emotion.Type)
			}
		}
		analysis.Emotions = emotions
	}

	// Extract quality
	if face.Quality != nil {
		analysis.Quality = &Quality{
			Brightness: float64(*face.Quality.Brightness),
			Sharpness:  float64(*face.Quality.Sharpness),
		}
	}

	// Extract pose
	if face.Pose != nil {
		analysis.Pose = &Pose{
			Roll:  float64(*face.Pose.Roll),
			Yaw:   float64(*face.Pose.Yaw),
			Pitch: float64(*face.Pose.Pitch),
		}
	}

	logger.Info("Face analysis completed", "image_key", imageKey, "confidence", analysis.Confidence)
	return analysis, nil
}

// CompareFaces compares two faces for similarity
func (s *AIService) CompareFaces(ctx context.Context, sourceImageKey, targetImageKey string) (*FaceComparisonResult, error) {
	logger.Info("Comparing faces", "source", sourceImageKey, "target", targetImageKey)

	input := &rekognition.CompareFacesInput{
		SourceImage: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(sourceImageKey),
			},
		},
		TargetImage: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(targetImageKey),
			},
		},
		SimilarityThreshold: aws.Float64(s.confidenceThreshold),
	}

	result, err := s.client.CompareFaces(ctx, input)
	if err != nil {
		logger.Error("Failed to compare faces", err, "source", sourceImageKey, "target", targetImageKey)
		return nil, fmt.Errorf("failed to compare faces: %w", err)
	}

	if len(result.FaceMatches) == 0 {
		return &FaceComparisonResult{
			Similarity: 0,
			IsMatch:    false,
			Confidence: 0,
			Details:    "No faces found for comparison",
		}, nil
	}

	match := result.FaceMatches[0]
	similarity := float64(*match.Similarity)
	isMatch := similarity >= s.confidenceThreshold

	details := fmt.Sprintf("Face similarity: %.2f%%, Threshold: %.2f%%", similarity*100, s.confidenceThreshold*100)
	if isMatch {
		details += " - Match found"
	} else {
		details += " - No match"
	}

	comparison := &FaceComparisonResult{
		Similarity: similarity,
		IsMatch:    isMatch,
		Confidence: similarity,
		Details:    details,
	}

	logger.Info("Face comparison completed", "source", sourceImageKey, "target", targetImageKey, "similarity", similarity, "match", isMatch)
	return comparison, nil
}

// DetectModerationLabels detects inappropriate content in an image
func (s *AIService) DetectModerationLabels(ctx context.Context, imageKey string) (*ModerationResult, error) {
	logger.Info("Detecting moderation labels", "image_key", imageKey)

	input := &rekognition.DetectModerationLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(imageKey),
			},
		},
		MinConfidence: aws.Float64(50.0), // Only detect labels with 50%+ confidence
	}

	result, err := s.client.DetectModerationLabels(ctx, input)
	if err != nil {
		logger.Error("Failed to detect moderation labels", err, "image_key", imageKey)
		return nil, fmt.Errorf("failed to detect moderation labels: %w", err)
	}

	moderation := &ModerationResult{
		IsAppropriate: true,
		Confidence:    100.0,
		Labels:        make([]ModerationLabel, 0),
		NSFWContent:   false,
		AdultContent:  false,
		ViolentContent: false,
	}

	if len(result.ModerationLabels) > 0 {
		labels := make([]ModerationLabel, 0, len(result.ModerationLabels))
		minConfidence := 100.0

		for _, label := range result.ModerationLabels {
			labelName := *label.Name
			confidence := float64(*label.Confidence)

			// Check for inappropriate content
			if s.isInappropriateContent(labelName) {
				moderation.IsAppropriate = false
				moderation.NSFWContent = moderation.NSFWContent || s.isNSFWContent(labelName)
				moderation.AdultContent = moderation.AdultContent || s.isAdultContent(labelName)
				moderation.ViolentContent = moderation.ViolentContent || s.isViolentContent(labelName)
			}

			labels = append(labels, ModerationLabel{
				Name:       labelName,
				Confidence: confidence,
				ParentName: aws.ToString(label.ParentName, ""),
			})

			if confidence < minConfidence {
				minConfidence = confidence
			}
		}

		moderation.Labels = labels
		moderation.Confidence = 100.0 - minConfidence
	}

	logger.Info("Moderation detection completed", "image_key", imageKey, "appropriate", moderation.IsAppropriate, "confidence", moderation.Confidence)
	return moderation, nil
}

// DetectLiveness detects if a face is live (anti-spoofing)
func (s *AIService) DetectLiveness(ctx context.Context, imageKey string) (*LivenessResult, error) {
	logger.Info("Detecting liveness", "image_key", imageKey)

	// First, detect faces to ensure we have a face to analyze
	faceInput := &rekognition.DetectFacesInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(imageKey),
			},
		},
		Attributes: []types.Attribute{
			types.AttributeQuality,
			types.AttributePose,
		},
	}

	faceResult, err := s.client.DetectFaces(ctx, faceInput)
	if err != nil {
		logger.Error("Failed to detect faces for liveness", err, "image_key", imageKey)
		return nil, fmt.Errorf("failed to detect faces for liveness: %w", err)
	}

	if len(faceResult.FaceDetails) == 0 {
		return &LivenessResult{
			IsLive:     false,
			Confidence: 0,
			Details:    "No face detected in image",
		}, nil
	}

	face := faceResult.FaceDetails[0]
	liveness := &LivenessResult{
		IsLive:     true,
		Confidence: float64(*face.Confidence),
		Challenges: []string{},
		Details:    "Face detected with good quality",
	}

	// Basic liveness checks based on face quality and pose
	if face.Quality != nil {
		brightness := float64(*face.Quality.Brightness)
		sharpness := float64(*face.Quality.Sharpness)

		// Check if image quality is too low (potential spoof)
		if brightness < 30 || sharpness < 30 {
			liveness.IsLive = false
			liveness.Confidence = 0.3
			liveness.Details = "Low image quality detected - potential spoof"
			liveness.Challenges = append(liveness.Challenges, "low_quality")
		}
	}

	if face.Pose != nil {
		roll := float64(*face.Pose.Roll)
		yaw := float64(*face.Pose.Yaw)
		pitch := float64(*face.Pose.Pitch)

		// Check for unusual pose angles (potential spoof)
		if roll > 45 || roll < -45 || yaw > 45 || yaw < -45 || pitch > 30 || pitch < -30 {
			liveness.IsLive = false
			liveness.Confidence = 0.4
			liveness.Details = "Unusual face pose detected - potential spoof"
			liveness.Challenges = append(liveness.Challenges, "unusual_pose")
		}
	}

	logger.Info("Liveness detection completed", "image_key", imageKey, "is_live", liveness.IsLive, "confidence", liveness.Confidence)
	return liveness, nil
}

// AnalyzeDocument analyzes a document (ID card, passport, etc.)
func (s *AIService) AnalyzeDocument(ctx context.Context, imageKey string) (*DocumentAnalysisResult, error) {
	logger.Info("Analyzing document", "image_key", imageKey)

	// First detect text in the document
	textInput := &rekognition.DetectTextInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: aws.String(s.bucket),
				Name:   aws.String(imageKey),
			},
		},
	}

	textResult, err := s.client.DetectText(ctx, textInput)
	if err != nil {
		logger.Error("Failed to detect text in document", err, "image_key", imageKey)
		return nil, fmt.Errorf("failed to detect text: %w", err)
	}

	// Analyze detected text to determine document type and extract fields
	documentType := "unknown"
	fields := make(map[string]interface{})
	confidence := 0.0
	isValid := false
	details := "Document analyzed"

	if len(textResult.TextDetections) > 0 {
		allText := ""
		for _, detection := range textResult.TextDetections {
			if detection.DetectedText != nil {
				allText += *detection.DetectedText + " "
			}
		}

		allText = strings.TrimSpace(allText)
		documentType, fields, confidence = s.analyzeDocumentText(allText)
		isValid = confidence > 70.0 // Consider valid if confidence > 70%

		if !isValid {
			details = fmt.Sprintf("Low confidence document analysis: %.2f%%", confidence)
		}
	} else {
		details = "No text detected in document"
		confidence = 0.0
	}

	analysis := &DocumentAnalysisResult{
		DocumentType: documentType,
		Confidence:   confidence,
		Fields:       fields,
		IsValid:      isValid,
		Details:      details,
	}

	logger.Info("Document analysis completed", "image_key", imageKey, "type", documentType, "confidence", confidence, "valid", isValid)
	return analysis, nil
}

// Helper methods

func (s *AIService) isInappropriateContent(labelName string) bool {
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
	}

	for _, label := range inappropriateLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

func (s *AIService) isNSFWContent(labelName string) bool {
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

func (s *AIService) isAdultContent(labelName string) bool {
	adultLabels := []string{
		"Alcohol",
		"Tobacco",
		"Gambling",
		"Adult",
	}

	for _, label := range adultLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

func (s *AIService) isViolentContent(labelName string) bool {
	violentLabels := []string{
		"Violence",
		"Weapons",
		"Terrorism",
		"Graphic Violence",
		"Physical Violence",
	}

	for _, label := range violentLabels {
		if strings.Contains(strings.ToLower(labelName), strings.ToLower(label)) {
			return true
		}
	}
	return false
}

func (s *AIService) analyzeDocumentText(text string) (string, map[string]interface{}, float64) {
	text = strings.ToLower(strings.TrimSpace(text))
	
	// Check for passport patterns
	if s.containsPassportPatterns(text) {
		fields := s.extractPassportFields(text)
		return "passport", fields, 85.0
	}
	
	// Check for ID card patterns
	if s.containsIDCardPatterns(text) {
		fields := s.extractIDCardFields(text)
		return "id_card", fields, 80.0
	}
	
	// Check for driver license patterns
	if s.containsDriverLicensePatterns(text) {
		fields := s.extractDriverLicenseFields(text)
		return "driver_license", fields, 75.0
	}
	
	return "unknown", map[string]interface{}{}, 0.0
}

func (s *AIService) containsPassportPatterns(text string) bool {
	patterns := []string{
		"passport",
		"republic of",
		"passport no",
		"date of birth",
		"place of birth",
		"authority",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func (s *AIService) containsIDCardPatterns(text string) bool {
	patterns := []string{
		"identification",
		"id card",
		"driver license",
		"date of birth",
		"expiry",
		"issue date",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func (s *AIService) containsDriverLicensePatterns(text string) bool {
	patterns := []string{
		"driver license",
		"driving licence",
		"class",
		"endorsements",
		"restrictions",
		"expiry date",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

func (s *AIService) extractPassportFields(text string) map[string]interface{} {
	fields := make(map[string]interface{})
	
	// Extract common passport fields (simplified for demo)
	if strings.Contains(text, "passport no") {
		fields["passport_number"] = "extracted"
	}
	if strings.Contains(text, "date of birth") {
		fields["date_of_birth"] = "extracted"
	}
	if strings.Contains(text, "place of birth") {
		fields["place_of_birth"] = "extracted"
	}
	if strings.Contains(text, "authority") {
		fields["issuing_authority"] = "extracted"
	}
	
	return fields
}

func (s *AIService) extractIDCardFields(text string) map[string]interface{} {
	fields := make(map[string]interface{})
	
	// Extract common ID card fields (simplified for demo)
	if strings.Contains(text, "id no") || strings.Contains(text, "identification") {
		fields["id_number"] = "extracted"
	}
	if strings.Contains(text, "date of birth") {
		fields["date_of_birth"] = "extracted"
	}
	if strings.Contains(text, "expiry") {
		fields["expiry_date"] = "extracted"
	}
	
	return fields
}

func (s *AIService) extractDriverLicenseFields(text string) map[string]interface{} {
	fields := make(map[string]interface{})
	
	// Extract common driver license fields (simplified for demo)
	if strings.Contains(text, "license no") {
		fields["license_number"] = "extracted"
	}
	if strings.Contains(text, "class") {
		fields["license_class"] = "extracted"
	}
	if strings.Contains(text, "expiry") {
		fields["expiry_date"] = "extracted"
	}
	
	return fields
}