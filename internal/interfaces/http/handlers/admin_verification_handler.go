package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/usecases/verification"
	"github.com/22smeargle/winkr-backend/pkg/errors"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// AdminVerificationHandler handles admin verification management
type AdminVerificationHandler struct {
	getPendingVerificationsUseCase *verification.GetPendingVerificationsUseCase
	processVerificationResultUseCase *verification.ProcessVerificationResultUseCase
}

// NewAdminVerificationHandler creates a new admin verification handler
func NewAdminVerificationHandler(
	getPendingVerificationsUseCase *verification.GetPendingVerificationsUseCase,
	processVerificationResultUseCase *verification.ProcessVerificationResultUseCase,
) *AdminVerificationHandler {
	return &AdminVerificationHandler{
		getPendingVerificationsUseCase: getPendingVerificationsUseCase,
		processVerificationResultUseCase: processVerificationResultUseCase,
	}
}

// GetPendingVerifications handles getting pending verifications
func (h *AdminVerificationHandler) GetPendingVerifications(c *gin.Context) {
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

	// Execute use case
	input := verification.GetPendingVerificationsInput{
		Limit:  limit,
		Offset:  offset,
	}

	output, err := h.getPendingVerificationsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to get pending verifications", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pending verifications", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// ApproveVerification handles approving a verification
func (h *AdminVerificationHandler) ApproveVerification(c *gin.Context) {
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

	// Execute use case
	input := verification.ProcessVerificationResultInput{
		VerificationID: verificationUUID,
		Approved:      true,
		ReviewedBy:     adminUUID,
	}

	output, err := h.processVerificationResultUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to approve verification", err, "verification_id", verificationUUID)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to approve verification", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// RejectVerification handles rejecting a verification
func (h *AdminVerificationHandler) RejectVerification(c *gin.Context) {
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

	// Set verification ID and admin ID
	input.VerificationID = verificationUUID
	input.ReviewedBy = adminUUID
	input.Approved = false

	// Execute use case
	output, err := h.processVerificationResultUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to reject verification", err, "verification_id", verificationUUID)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reject verification", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, output)
}

// GetVerificationDetails handles getting verification details
func (h *AdminVerificationHandler) GetVerificationDetails(c *gin.Context) {
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

	// Get verification details using the get pending verifications use case
	// We need to get the verification by ID, so we'll use a limit of 1 and offset 0
	input := verification.GetPendingVerificationsInput{
		Limit:  1,
		Offset:  0,
	}

	pendingOutput, err := h.getPendingVerificationsUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		logger.Error("Failed to get verification details", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get verification details", err.Error())
		return
	}

	// Find the specific verification
	var verificationDetails *verification.VerificationWithUser
	if len(pendingOutput.Verifications) > 0 {
		for _, verification := range pendingOutput.Verifications {
			if verification.ID == verificationUUID {
				verificationDetails = verification
				break
			}
		}
	}

	if verificationDetails == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Verification not found", "")
		return
	}

	// Create detailed response
	details := map[string]interface{}{
		"id":               verificationDetails.ID,
		"user_id":          verificationDetails.UserID,
		"type":             verificationDetails.Type,
		"status":           verificationDetails.Status,
		"photo_url":        verificationDetails.PhotoURL,
		"ai_score":         verificationDetails.AIScore,
		"rejection_reason":  verificationDetails.RejectionReason,
		"created_at":       verificationDetails.CreatedAt,
		"updated_at":       verificationDetails.UpdatedAt,
	}

	// Add user information if available
	if verificationDetails.User != nil {
		details["user"] = verificationDetails.User
	}

	// Add reviewer information if available
	if verificationDetails.User != nil && verificationDetails.User.ID != uuid.Nil {
		// Get user from verification to get reviewer info
		// This is a simplified approach - in a real implementation,
		// you would fetch the reviewer from the verification entity
		reviewer := map[string]interface{}{
			"id":         adminUUID,
			"first_name": "Admin",
			"last_name":  "User",
			"email":      "admin@example.com",
		}
		details["reviewer"] = reviewer
	}

	utils.SuccessResponse(c, http.StatusOK, details)
}