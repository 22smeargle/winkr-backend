package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/usecases/verification"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// VerificationHandler handles verification HTTP requests
type VerificationHandler struct {
	requestSelfieVerificationUseCase     *verification.RequestSelfieVerificationUseCase
	submitSelfieVerificationUseCase      *verification.SubmitSelfieVerificationUseCase
	requestDocumentVerificationUseCase    *verification.RequestDocumentVerificationUseCase
	submitDocumentVerificationUseCase     *verification.SubmitDocumentVerificationUseCase
	getVerificationStatusUseCase          *verification.GetVerificationStatusUseCase
	processVerificationResultUseCase      *verification.ProcessVerificationResultUseCase
	getPendingVerificationsUseCase         *verification.GetPendingVerificationsUseCase
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(
	requestSelfieVerificationUseCase *verification.RequestSelfieVerificationUseCase,
	submitSelfieVerificationUseCase *verification.SubmitSelfieVerificationUseCase,
	requestDocumentVerificationUseCase *verification.RequestDocumentVerificationUseCase,
	submitDocumentVerificationUseCase *verification.SubmitDocumentVerificationUseCase,
	getVerificationStatusUseCase *verification.GetVerificationStatusUseCase,
	processVerificationResultUseCase *verification.ProcessVerificationResultUseCase,
	getPendingVerificationsUseCase *verification.GetPendingVerificationsUseCase,
) *VerificationHandler {
	return &VerificationHandler{
		requestSelfieVerificationUseCase:     requestSelfieVerificationUseCase,
		submitSelfieVerificationUseCase:      submitSelfieVerificationUseCase,
		requestDocumentVerificationUseCase:    requestDocumentVerificationUseCase,
		submitDocumentVerificationUseCase:     submitDocumentVerificationUseCase,
		getVerificationStatusUseCase:          getVerificationStatusUseCase,
		processVerificationResultUseCase:      processVerificationResultUseCase,
		getPendingVerificationsUseCase:         getPendingVerificationsUseCase,
	}
}

// RequestSelfieVerification handles selfie verification request
func (h *VerificationHandler) RequestSelfieVerification(c *gin.Context) {
	var input verification.RequestSelfieVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind selfie verification request", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "")
		return
	}

	// Add IP and User Agent to input
	input.IPAddress = c.ClientIP()
	input.UserAgent = c.GetHeader("User-Agent")

	// Execute use case
	output, err := h.requestSelfieVerificationUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to request selfie verification", err, "user_id", userUUID)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// SubmitSelfieVerification handles selfie verification submission
func (h *VerificationHandler) SubmitSelfieVerification(c *gin.Context) {
	var input verification.SubmitSelfieVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind selfie verification submission", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "")
		return
	}

	// Execute use case
	output, err := h.submitSelfieVerificationUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to submit selfie verification", err, "user_id", userUUID)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// RequestDocumentVerification handles document verification request
func (h *VerificationHandler) RequestDocumentVerification(c *gin.Context) {
	var input verification.RequestDocumentVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind document verification request", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "")
		return
	}

	// Add IP and User Agent to input
	input.IPAddress = c.ClientIP()
	input.UserAgent = c.GetHeader("User-Agent")

	// Execute use case
	output, err := h.requestDocumentVerificationUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to request document verification", err, "user_id", userUUID)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// SubmitDocumentVerification handles document verification submission
func (h *VerificationHandler) SubmitDocumentVerification(c *gin.Context) {
	var input verification.SubmitDocumentVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind document verification submission", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "")
		return
	}

	// Execute use case
	output, err := h.submitDocumentVerificationUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to submit document verification", err, "user_id", userUUID)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// GetVerificationStatus handles getting verification status
func (h *VerificationHandler) GetVerificationStatus(c *gin.Context) {
	var input verification.GetVerificationStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind verification status request", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", "")
		return
	}

	// Execute use case
	output, err := h.getVerificationStatusUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to get verification status", err, "user_id", userUUID)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// GetPendingVerifications handles getting pending verifications (admin only)
func (h *VerificationHandler) GetPendingVerifications(c *gin.Context) {
	// Check if user is admin
	adminID, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Admin not authenticated", "")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid limit parameter", "")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid offset parameter", "")
		return
	}

	input := verification.GetPendingVerificationsInput{
		Limit:  limit,
		Offset:  offset,
	}

	// Execute use case
	output, err := h.getPendingVerificationsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to get pending verifications", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pending verifications", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// ApproveVerification handles approving a verification (admin only)
func (h *VerificationHandler) ApproveVerification(c *gin.Context) {
	verificationID := c.Param("id")
	if verificationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Verification ID is required", "")
		return
	}

	verificationUUID, err := uuid.Parse(verificationID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid verification ID", "")
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Admin not authenticated", "")
		return
	}

	adminUUID, err := uuid.Parse(adminID.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid admin ID", "")
		return
	}

	input := verification.ProcessVerificationResultInput{
		VerificationID: verificationUUID,
		Approved:      true,
		ReviewedBy:     adminUUID,
	}

	// Execute use case
	output, err := h.processVerificationResultUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to approve verification", err, "verification_id", verificationUUID)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to approve verification", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// RejectVerification handles rejecting a verification (admin only)
func (h *VerificationHandler) RejectVerification(c *gin.Context) {
	var input verification.ProcessVerificationResultInput
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Error("Failed to bind verification rejection request", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	verificationID := c.Param("id")
	if verificationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Verification ID is required", "")
		return
	}

	verificationUUID, err := uuid.Parse(verificationID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid verification ID", "")
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Admin not authenticated", "")
		return
	}

	adminUUID, err := uuid.Parse(adminID.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid admin ID", "")
		return
	}

	// Set approved to false and add reason
	input.Approved = false
	if input.Reason == "" {
		input.Reason = "Rejected by admin"
	}

	input.ReviewedBy = adminUUID
	input.VerificationID = verificationUUID

	// Execute use case
	output, err := h.processVerificationResultUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to reject verification", err, "verification_id", verificationUUID)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reject verification", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// GetVerificationDetails handles getting verification details (admin only)
func (h *VerificationHandler) GetVerificationDetails(c *gin.Context) {
	verificationID := c.Param("id")
	if verificationID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Verification ID is required", "")
		return
	}

	verificationUUID, err := uuid.Parse(verificationID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid verification ID", "")
		return
	}

	// Get verification from repository to get full details
	verification, err := h.getPendingVerificationsUseCase.verificationService.GetVerificationStatus(c.Request.Context(), verificationUUID, entities.VerificationTypeSelfie)
	if err != nil {
		logger.Error("Failed to get verification details", err, "verification_id", verificationUUID)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get verification details", err.Error())
		return
	}

	// Create detailed response
	details := map[string]interface{}{
		"id":               verification.ID,
		"user_id":          verification.UserID,
		"type":             string(verification.Type),
		"status":           verification.Status.String(),
		"photo_url":        verification.PhotoURL,
		"ai_score":         verification.AIScore,
		"rejection_reason":  verification.RejectionReason,
		"created_at":       verification.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":       verification.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Add user information if available
	if verification.User != nil {
		details["user"] = map[string]interface{}{
			"id":         verification.User.ID,
			"first_name": verification.User.FirstName,
			"last_name":  verification.User.LastName,
			"email":      verification.User.Email,
		}
	}

	// Add reviewer information if available
	if verification.Reviewer != nil {
		details["reviewer"] = map[string]interface{}{
			"id":         verification.Reviewer.ID,
			"first_name": verification.Reviewer.FirstName,
			"last_name":  verification.Reviewer.LastName,
			"email":      verification.Reviewer.Email,
		}
	}

	utils.SuccessResponse(c, http.StatusOK, details)
}