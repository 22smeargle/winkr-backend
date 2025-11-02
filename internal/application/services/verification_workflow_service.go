package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// VerificationWorkflowService handles verification workflow management
type VerificationWorkflowService struct {
	verificationRepo repositories.VerificationRepository
	userRepo         repositories.UserRepository
	aiService        external.AIService
	documentService  *DocumentService
	storageService   StorageService
}

// NewVerificationWorkflowService creates a new verification workflow service
func NewVerificationWorkflowService(
	verificationRepo repositories.VerificationRepository,
	userRepo repositories.UserRepository,
	aiService external.AIService,
	documentService *DocumentService,
	storageService StorageService,
) *VerificationWorkflowService {
	return &VerificationWorkflowService{
		verificationRepo: verificationRepo,
		userRepo:         userRepo,
		aiService:        aiService,
		documentService:  documentService,
		storageService:   storageService,
	}
}

// VerificationConfig represents verification configuration
type VerificationConfig struct {
	SelfieSimilarityThreshold    float64 `json:"selfie_similarity_threshold"`
	DocumentConfidenceThreshold   float64 `json:"document_confidence_threshold"`
	MaxVerificationAttempts       int     `json:"max_verification_attempts"`
	AttemptCooldownMinutes        int     `json:"attempt_cooldown_minutes"`
	VerificationExpiryDays        int     `json:"verification_expiry_days"`
	ManualReviewThreshold        float64 `json:"manual_review_threshold"`
	NSFWThreshold               float64 `json:"nsfw_threshold"`
	LivenessThreshold           float64 `json:"liveness_threshold"`
}

// DefaultVerificationConfig returns default verification configuration
func DefaultVerificationConfig() *VerificationConfig {
	return &VerificationConfig{
		SelfieSimilarityThreshold:  0.85,  // 85% similarity required
		DocumentConfidenceThreshold: 0.75,   // 75% confidence required
		MaxVerificationAttempts:     5,        // Max 5 attempts per day
		AttemptCooldownMinutes:      30,       // 30 minutes between attempts
		VerificationExpiryDays:       365,      // Verifications expire after 1 year
		ManualReviewThreshold:        0.70,     // Below 70% confidence requires manual review
		NSFWThreshold:               0.80,     // 80% confidence for NSFW detection
		LivenessThreshold:           0.60,     // 60% confidence for liveness
	}
}

// RequestSelfieVerification initiates selfie verification process
func (vws *VerificationWorkflowService) RequestSelfieVerification(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (*entities.Verification, error) {
	logger.Info("Requesting selfie verification", "user_id", userID, "ip", ipAddress)

	// Check if user already has pending selfie verification
	existing, err := vws.verificationRepo.GetVerificationByUserAndType(ctx, userID, entities.VerificationTypeSelfie)
	if err != nil {
		logger.Error("Failed to check existing selfie verification", err, "user_id", userID)
		return nil, fmt.Errorf("failed to check existing verification: %w", err)
	}

	if existing != nil && existing.Status.IsPending() && !existing.IsExpired() {
		logger.Warn("User already has pending selfie verification", "user_id", userID)
		return nil, fmt.Errorf("selfie verification already in progress")
	}

	// Check verification attempt limits
	attempts, err := vws.checkVerificationAttempts(ctx, userID, entities.VerificationTypeSelfie, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check verification attempts: %w", err)
	}

	if len(attempts) >= DefaultVerificationConfig().MaxVerificationAttempts {
		logger.Warn("Too many verification attempts", "user_id", userID, "attempts", len(attempts))
		return nil, fmt.Errorf("too many verification attempts, please try again later")
	}

	// Create verification attempt record
	attempt := &entities.VerificationAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      entities.VerificationTypeSelfie,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:    "success",
		CreatedAt:  time.Now(),
	}

	err = vws.verificationRepo.CreateVerificationAttempt(ctx, attempt)
	if err != nil {
		logger.Error("Failed to create verification attempt", err, "user_id", userID)
		return nil, fmt.Errorf("failed to record verification attempt: %w", err)
	}

	// Create verification record
	verification := &entities.Verification{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     entities.VerificationTypeSelfie,
		Status:   valueobjects.VerificationStatusPending,
		PhotoURL: "", // Will be set when photo is uploaded
		PhotoKey: "", // Will be set when photo is uploaded
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = vws.verificationRepo.CreateVerification(ctx, verification)
	if err != nil {
		logger.Error("Failed to create selfie verification", err, "user_id", userID)
		return nil, fmt.Errorf("failed to create verification: %w", err)
	}

	logger.Info("Selfie verification requested", "user_id", userID, "verification_id", verification.ID)
	return verification, nil
}

// SubmitSelfieVerification processes submitted selfie photo
func (vws *VerificationWorkflowService) SubmitSelfieVerification(ctx context.Context, verificationID uuid.UUID, photoKey, photoURL string) (*entities.Verification, error) {
	logger.Info("Processing selfie verification submission", "verification_id", verificationID, "photo_key", photoKey)

	// Get verification record
	verification, err := vws.verificationRepo.GetVerificationByID(ctx, verificationID)
	if err != nil {
		logger.Error("Failed to get verification", err, "verification_id", verificationID)
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	if verification == nil {
		return nil, fmt.Errorf("verification not found")
	}

	if verification.Type != entities.VerificationTypeSelfie {
		return nil, fmt.Errorf("verification is not for selfie")
	}

	if !verification.Status.IsPending() {
		return nil, fmt.Errorf("verification is not pending")
	}

	// Update verification with photo info
	verification.PhotoKey = photoKey
	verification.PhotoURL = photoURL

	// Process selfie with AI
	err = vws.processSelfieWithAI(ctx, verification)
	if err != nil {
		logger.Error("Failed to process selfie with AI", err, "verification_id", verificationID)
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "AI processing failed"
		verification.RejectionReason = &reason
		vws.verificationRepo.UpdateVerification(ctx, verification)
		return nil, fmt.Errorf("AI processing failed: %w", err)
	}

	// Update verification in database
	err = vws.verificationRepo.UpdateVerification(ctx, verification)
	if err != nil {
		logger.Error("Failed to update verification", err, "verification_id", verificationID)
		return nil, fmt.Errorf("failed to update verification: %w", err)
	}

	logger.Info("Selfie verification processed", "verification_id", verificationID, "status", verification.Status, "ai_score", verification.AIScore)
	return verification, nil
}

// RequestDocumentVerification initiates document verification process
func (vws *VerificationWorkflowService) RequestDocumentVerification(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (*entities.Verification, error) {
	logger.Info("Requesting document verification", "user_id", userID, "ip", ipAddress)

	// Check if user already has pending document verification
	existing, err := vws.verificationRepo.GetVerificationByUserAndType(ctx, userID, entities.VerificationTypeDocument)
	if err != nil {
		logger.Error("Failed to check existing document verification", err, "user_id", userID)
		return nil, fmt.Errorf("failed to check existing verification: %w", err)
	}

	if existing != nil && existing.Status.IsPending() && !existing.IsExpired() {
		logger.Warn("User already has pending document verification", "user_id", userID)
		return nil, fmt.Errorf("document verification already in progress")
	}

	// Check verification attempt limits
	attempts, err := vws.checkVerificationAttempts(ctx, userID, entities.VerificationTypeDocument, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check verification attempts: %w", err)
	}

	if len(attempts) >= DefaultVerificationConfig().MaxVerificationAttempts {
		logger.Warn("Too many verification attempts", "user_id", userID, "attempts", len(attempts))
		return nil, fmt.Errorf("too many verification attempts, please try again later")
	}

	// Create verification attempt record
	attempt := &entities.VerificationAttempt{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      entities.VerificationTypeDocument,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:    "success",
		CreatedAt:  time.Now(),
	}

	err = vws.verificationRepo.CreateVerificationAttempt(ctx, attempt)
	if err != nil {
		logger.Error("Failed to create verification attempt", err, "user_id", userID)
		return nil, fmt.Errorf("failed to record verification attempt: %w", err)
	}

	// Create verification record
	verification := &entities.Verification{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     entities.VerificationTypeDocument,
		Status:   valueobjects.VerificationStatusPending,
		PhotoURL: "", // Will be set when photo is uploaded
		PhotoKey: "", // Will be set when photo is uploaded
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = vws.verificationRepo.CreateVerification(ctx, verification)
	if err != nil {
		logger.Error("Failed to create document verification", err, "user_id", userID)
		return nil, fmt.Errorf("failed to create verification: %w", err)
	}

	logger.Info("Document verification requested", "user_id", userID, "verification_id", verification.ID)
	return verification, nil
}

// SubmitDocumentVerification processes submitted document photo
func (vws *VerificationWorkflowService) SubmitDocumentVerification(ctx context.Context, verificationID uuid.UUID, photoKey, photoURL string) (*entities.Verification, error) {
	logger.Info("Processing document verification submission", "verification_id", verificationID, "photo_key", photoKey)

	// Get verification record
	verification, err := vws.verificationRepo.GetVerificationByID(ctx, verificationID)
	if err != nil {
		logger.Error("Failed to get verification", err, "verification_id", verificationID)
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	if verification == nil {
		return nil, fmt.Errorf("verification not found")
	}

	if verification.Type != entities.VerificationTypeDocument {
		return nil, fmt.Errorf("verification is not for document")
	}

	if !verification.Status.IsPending() {
		return nil, fmt.Errorf("verification is not pending")
	}

	// Update verification with photo info
	verification.PhotoKey = photoKey
	verification.PhotoURL = photoURL

	// Process document with AI and OCR
	err = vws.processDocumentWithAI(ctx, verification)
	if err != nil {
		logger.Error("Failed to process document with AI", err, "verification_id", verificationID)
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "Document processing failed"
		verification.RejectionReason = &reason
		vws.verificationRepo.UpdateVerification(ctx, verification)
		return nil, fmt.Errorf("document processing failed: %w", err)
	}

	// Update verification in database
	err = vws.verificationRepo.UpdateVerification(ctx, verification)
	if err != nil {
		logger.Error("Failed to update verification", err, "verification_id", verificationID)
		return nil, fmt.Errorf("failed to update verification: %w", err)
	}

	logger.Info("Document verification processed", "verification_id", verificationID, "status", verification.Status, "ai_score", verification.AIScore)
	return verification, nil
}

// GetVerificationStatus gets verification status for a user
func (vws *VerificationWorkflowService) GetVerificationStatus(ctx context.Context, userID uuid.UUID, vType entities.VerificationType) (*VerificationStatusResult, error) {
	logger.Info("Getting verification status", "user_id", userID, "type", vType)

	verification, err := vws.verificationRepo.GetVerificationByUserAndType(ctx, userID, vType)
	if err != nil {
		logger.Error("Failed to get verification", err, "user_id", userID, "type", vType)
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	result := &VerificationStatusResult{
		UserID:          userID,
		Type:            vType,
		Status:          valueobjects.VerificationStatusPending,
		HasVerification:  false,
		CanRequest:      true,
		LastVerification:  nil,
	}

	if verification != nil {
		result.Status = verification.Status
		result.HasVerification = true
		result.LastVerification = verification
		result.CanRequest = verification.Status.IsRejected() || verification.IsExpired()
	}

	// Get user's verification level
	level, err := vws.verificationRepo.GetUserVerificationLevel(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user verification level", err, "user_id", userID)
		return result, fmt.Errorf("failed to get verification level: %w", err)
	}

	result.VerificationLevel = level

	logger.Info("Verification status retrieved", "user_id", userID, "type", vType, "status", result.Status, "level", level)
	return result, nil
}

// ProcessVerificationResult processes the result of a verification
func (vws *VerificationWorkflowService) ProcessVerificationResult(ctx context.Context, verificationID uuid.UUID, approved bool, reason string, reviewedBy uuid.UUID) error {
	logger.Info("Processing verification result", "verification_id", verificationID, "approved", approved, "reviewed_by", reviewedBy)

	// Get verification record
	verification, err := vws.verificationRepo.GetVerificationByID(ctx, verificationID)
	if err != nil {
		logger.Error("Failed to get verification", err, "verification_id", verificationID)
		return fmt.Errorf("failed to get verification: %w", err)
	}

	if verification == nil {
		return fmt.Errorf("verification not found")
	}

	// Update verification status
	if approved {
		verification.Approve(reviewedBy)
		err = vws.awardVerificationBadge(ctx, verification.UserID, verification.Type)
		if err != nil {
			logger.Error("Failed to award verification badge", err, "user_id", verification.UserID, "type", verification.Type)
			return fmt.Errorf("failed to award badge: %w", err)
		}
	} else {
		verification.Reject(reason, reviewedBy)
	}

	// Update verification in database
	err = vws.verificationRepo.UpdateVerification(ctx, verification)
	if err != nil {
		logger.Error("Failed to update verification", err, "verification_id", verificationID)
		return fmt.Errorf("failed to update verification: %w", err)
	}

	// Update user verification level
	err = vws.updateUserVerificationLevel(ctx, verification.UserID)
	if err != nil {
		logger.Error("Failed to update user verification level", err, "user_id", verification.UserID)
		return fmt.Errorf("failed to update user verification level: %w", err)
	}

	logger.Info("Verification result processed", "verification_id", verificationID, "approved", approved, "status", verification.Status)
	return nil
}

// GetPendingVerifications gets verifications pending review
func (vws *VerificationWorkflowService) GetPendingVerifications(ctx context.Context, limit, offset int) ([]*entities.Verification, error) {
	logger.Info("Getting pending verifications", "limit", limit, "offset", offset)

	verifications, err := vws.verificationRepo.GetVerificationsForReview(ctx, valueobjects.VerificationStatusPending, limit, offset)
	if err != nil {
		logger.Error("Failed to get pending verifications", err)
		return nil, fmt.Errorf("failed to get pending verifications: %w", err)
	}

	logger.Info("Retrieved pending verifications", "count", len(verifications))
	return verifications, nil
}

// Helper methods

func (vws *VerificationWorkflowService) checkVerificationAttempts(ctx context.Context, userID uuid.UUID, vType entities.VerificationType, ipAddress string) ([]*entities.VerificationAttempt, error) {
	since := time.Now().Add(-time.Duration(DefaultVerificationConfig().AttemptCooldownMinutes) * time.Minute)
	return vws.verificationRepo.GetVerificationAttemptsByUser(ctx, userID, vType, since)
}

func (vws *VerificationWorkflowService) processSelfieWithAI(ctx context.Context, verification *entities.Verification) error {
	logger.Info("Processing selfie with AI", "verification_id", verification.ID, "photo_key", verification.PhotoKey)

	// Analyze face in selfie
	faceAnalysis, err := vws.aiService.AnalyzeFace(ctx, verification.PhotoKey)
	if err != nil {
		return fmt.Errorf("failed to analyze face: %w", err)
	}

	if !faceAnalysis.HasFace {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "No face detected in photo"
		verification.RejectionReason = &reason
		return nil
	}

	// Check for liveness
	livenessResult, err := vws.aiService.DetectLiveness(ctx, verification.PhotoKey)
	if err != nil {
		return fmt.Errorf("failed to detect liveness: %w", err)
	}

	if !livenessResult.IsLive {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "Photo appears to be spoofed or not live"
		verification.RejectionReason = &reason
		return nil
	}

	// Check for inappropriate content
	moderationResult, err := vws.aiService.DetectModerationLabels(ctx, verification.PhotoKey)
	if err != nil {
		return fmt.Errorf("failed to detect inappropriate content: %w", err)
	}

	if !moderationResult.IsAppropriate {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "Inappropriate content detected"
		verification.RejectionReason = &reason
		return nil
	}

	// Compare with user's primary photo if available
	user, err := vws.userRepo.GetByID(ctx, verification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Find user's primary photo
	var primaryPhotoKey string
	for _, photo := range user.Photos {
		if photo.IsPrimary && photo.IsVerified() {
			primaryPhotoKey = photo.FileKey
			break
		}
	}

	aiScore := 0.8 // Default score
	aiDetails := ""

	if primaryPhotoKey != "" {
		// Compare with primary photo
		comparison, err := vws.aiService.CompareFaces(ctx, primaryPhotoKey, verification.PhotoKey)
		if err != nil {
			logger.Warn("Failed to compare faces", err, "primary", primaryPhotoKey, "new", verification.PhotoKey)
		} else {
			aiScore = comparison.Similarity
			aiDetails = comparison.Details
		}
	}

	// Determine final status based on confidence
	config := DefaultVerificationConfig()
	if aiScore >= config.SelfieSimilarityThreshold && livenessResult.Confidence >= config.LivenessThreshold {
		verification.Status = valueobjects.VerificationStatusApproved
	} else if aiScore < config.ManualReviewThreshold {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := fmt.Sprintf("Low confidence score: %.2f%%", aiScore*100)
		verification.RejectionReason = &reason
	} else {
		verification.Status = valueobjects.VerificationStatusPending // Requires manual review
		reason := "Requires manual review"
		verification.RejectionReason = &reason
	}

	// Set AI analysis results
	verification.SetAIScore(aiScore, aiDetails)

	return nil
}

func (vws *VerificationWorkflowService) processDocumentWithAI(ctx context.Context, verification *entities.Verification) error {
	logger.Info("Processing document with AI", "verification_id", verification.ID, "photo_key", verification.PhotoKey)

	// Analyze document with AI
	documentAnalysis, err := vws.documentService.ProcessDocument(ctx, verification.PhotoKey)
	if err != nil {
		return fmt.Errorf("failed to analyze document: %w", err)
	}

	// Check for inappropriate content
	moderationResult, err := vws.aiService.DetectModerationLabels(ctx, verification.PhotoKey)
	if err != nil {
		return fmt.Errorf("failed to detect inappropriate content: %w", err)
	}

	if !moderationResult.IsAppropriate {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := "Inappropriate content detected in document"
		verification.RejectionReason = &reason
		return nil
	}

	// Set document analysis results
	verification.SetDocumentData(documentAnalysis.DocumentType, vws.marshalDocumentFields(documentAnalysis.Fields))

	// Determine final status based on confidence
	config := DefaultVerificationConfig()
	if documentAnalysis.IsValid && documentAnalysis.Confidence >= config.DocumentConfidenceThreshold {
		verification.Status = valueobjects.VerificationStatusApproved
	} else if documentAnalysis.Confidence < config.ManualReviewThreshold {
		verification.Status = valueobjects.VerificationStatusRejected
		reason := fmt.Sprintf("Low confidence score: %.2f%%", documentAnalysis.Confidence*100)
		verification.RejectionReason = &reason
	} else {
		verification.Status = valueobjects.VerificationStatusPending // Requires manual review
		reason := "Requires manual review"
		verification.RejectionReason = &reason
	}

	// Set AI analysis results
	verification.SetAIScore(documentAnalysis.Confidence, documentAnalysis.Details)

	return nil
}

func (vws *VerificationWorkflowService) awardVerificationBadge(ctx context.Context, userID uuid.UUID, vType entities.VerificationType) error {
	logger.Info("Awarding verification badge", "user_id", userID, "type", vType)

	// Determine badge type and level
	var badgeType string
	var level entities.VerificationLevel

	switch vType {
	case entities.VerificationTypeSelfie:
		badgeType = "selfie_verified"
		level = entities.VerificationLevelSelfie
	case entities.VerificationTypeDocument:
		badgeType = "document_verified"
		level = entities.VerificationLevelDocument
	default:
		return fmt.Errorf("unknown verification type: %s", vType)
	}

	// Check if user already has this badge
	existingBadge, err := vws.verificationRepo.GetBadgeByUserAndType(ctx, userID, badgeType)
	if err != nil {
		return fmt.Errorf("failed to check existing badge: %w", err)
	}

	if existingBadge != nil && existingBadge.IsActive() {
		logger.Info("User already has active badge", "user_id", userID, "badge_type", badgeType)
		return nil // Badge already exists
	}

	// Create new badge
	expiresAt := time.Now().AddDate(1, 0, 0) // Badges expire after 1 year
	badge := &entities.VerificationBadge{
		ID:        uuid.New(),
		UserID:    userID,
		Level:     level,
		BadgeType: badgeType,
		ExpiresAt: &expiresAt,
		IsRevoked: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = vws.verificationRepo.CreateVerificationBadge(ctx, badge)
	if err != nil {
		return fmt.Errorf("failed to create badge: %w", err)
	}

	// Update user verification level
	err = vws.updateUserVerificationLevel(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to update user verification level: %w", err)
	}

	logger.Info("Verification badge awarded", "user_id", userID, "badge_type", badgeType, "level", level)
	return nil
}

func (vws *VerificationWorkflowService) updateUserVerificationLevel(ctx context.Context, userID uuid.UUID) error {
	// Get user's active badges
	badges, err := vws.verificationRepo.GetActiveBadgesByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user badges: %w", err)
	}

	// Determine highest verification level
	level := entities.VerificationLevelNone
	for _, badge := range badges {
		if badge.IsActive() && badge.Level > level {
			level = badge.Level
		}
	}

	// Update user verification level
	return vws.verificationRepo.UpdateUserVerificationLevel(ctx, userID, level)
}

func (vws *VerificationWorkflowService) marshalDocumentFields(fields map[string]interface{}) string {
	data, err := json.Marshal(fields)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// VerificationStatusResult represents verification status result
type VerificationStatusResult struct {
	UserID          uuid.UUID                    `json:"user_id"`
	Type            entities.VerificationType      `json:"type"`
	Status          valueobjects.VerificationStatus `json:"status"`
	HasVerification  bool                         `json:"has_verification"`
	CanRequest      bool                         `json:"can_request"`
	LastVerification *entities.Verification           `json:"last_verification,omitempty"`
	VerificationLevel entities.VerificationLevel       `json:"verification_level"`
}

// StorageService defines interface for storage operations
type StorageService interface {
	GetFileURL(ctx context.Context, key string) (string, error)
}