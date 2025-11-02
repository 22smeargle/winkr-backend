package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/22smeargle/winkr-backend/internal/application/usecases/moderation"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/response"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// ModerationHandler handles moderation-related HTTP endpoints
type ModerationHandler struct {
	reportContentUseCase   *moderation.ReportContentUseCase
	blockUserUseCase      *moderation.BlockUserUseCase
	getBlockedUsersUseCase *moderation.GetBlockedUsersUseCase
	validator             validator.Validator
}

// NewModerationHandler creates a new moderation handler
func NewModerationHandler(
	reportContentUseCase *moderation.ReportContentUseCase,
	blockUserUseCase *moderation.BlockUserUseCase,
	getBlockedUsersUseCase *moderation.GetBlockedUsersUseCase,
	validator validator.Validator,
) *ModerationHandler {
	return &ModerationHandler{
		reportContentUseCase:   reportContentUseCase,
		blockUserUseCase:      blockUserUseCase,
		getBlockedUsersUseCase: getBlockedUsersUseCase,
		validator:             validator,
	}
}

// ReportContent handles POST /report endpoint
func (h *ModerationHandler) ReportContent(c *gin.Context) {
	logger.Info("ReportContent request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	var req moderation.ReportContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	req.ReporterID = userID
	
	// Execute use case
	result, err := h.reportContentUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute ReportContent use case", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to submit report", err)
		return
	}
	
	response.Success(c, http.StatusCreated, "Report submitted successfully", result)
}

// BlockUser handles POST /block/:id endpoint
func (h *ModerationHandler) BlockUser(c *gin.Context) {
	logger.Info("BlockUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get blocked user ID from URL parameter
	blockedIDStr := c.Param("id")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		logger.Error("Invalid blocked user ID", err, "blocked_id", blockedIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	var req moderation.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	req.BlockerID = userID
	req.BlockedID = blockedID
	
	// Execute use case
	result, err := h.blockUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute BlockUser use case", err, "blocker_id", userID, "blocked_id", blockedID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to block user", err)
		return
	}
	
	response.Success(c, http.StatusCreated, "User blocked successfully", result)
}

// UnblockUser handles DELETE /block/:id endpoint
func (h *ModerationHandler) UnblockUser(c *gin.Context) {
	logger.Info("UnblockUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get blocked user ID from URL parameter
	blockedIDStr := c.Param("id")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		logger.Error("Invalid blocked user ID", err, "blocked_id", blockedIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Execute use case
	err = h.blockUserUseCase.UnblockUser(c.Request.Context(), userID, blockedID)
	if err != nil {
		logger.Error("Failed to execute UnblockUser use case", err, "blocker_id", userID, "blocked_id", blockedID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to unblock user", err)
		return
	}
	
	response.Success(c, http.StatusOK, "User unblocked successfully", gin.H{
		"message": "User unblocked successfully",
	})
}

// GetBlockedUsers handles GET /me/blocked endpoint
func (h *ModerationHandler) GetBlockedUsers(c *gin.Context) {
	logger.Info("GetBlockedUsers request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	req := moderation.GetBlockedUsersRequest{
		UserID: userID,
		Limit:   limit,
		Offset:  offset,
	}
	
	// Execute use case
	result, err := h.getBlockedUsersUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetBlockedUsers use case", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get blocked users", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Blocked users retrieved successfully", result)
}

// GetUsersBlocking handles GET /me/blocking endpoint
func (h *ModerationHandler) GetUsersBlocking(c *gin.Context) {
	logger.Info("GetUsersBlocking request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	// Execute use case
	result, err := h.getBlockedUsersUseCase.GetUsersBlocking(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetUsersBlocking use case", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get users blocking", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Users blocking retrieved successfully", result)
}

// GetMyReports handles GET /me/reports endpoint
func (h *ModerationHandler) GetMyReports(c *gin.Context) {
	logger.Info("GetMyReports request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	
	// Execute use case
	reports, err := h.reportContentUseCase.GetUserReports(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetUserReports use case", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get user reports", err)
		return
	}
	
	response.Success(c, http.StatusOK, "User reports retrieved successfully", gin.H{
		"reports": reports,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetReportStatus handles GET /reports/:id endpoint
func (h *ModerationHandler) GetReportStatus(c *gin.Context) {
	logger.Info("GetReportStatus request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get report ID from URL parameter
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logger.Error("Invalid report ID", err, "report_id", reportIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Execute use case
	report, err := h.reportContentUseCase.GetReportStatus(c.Request.Context(), reportID, userID)
	if err != nil {
		logger.Error("Failed to execute GetReportStatus use case", err, "report_id", reportID, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get report status", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Report status retrieved successfully", report)
}

// CancelReport handles DELETE /reports/:id endpoint
func (h *ModerationHandler) CancelReport(c *gin.Context) {
	logger.Info("CancelReport request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get report ID from URL parameter
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logger.Error("Invalid report ID", err, "report_id", reportIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Execute use case
	err = h.reportContentUseCase.CancelReport(c.Request.Context(), reportID, userID)
	if err != nil {
		logger.Error("Failed to execute CancelReport use case", err, "report_id", reportID, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to cancel report", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Report cancelled successfully", gin.H{
		"message": "Report cancelled successfully",
	})
}

// CheckBlockStatus handles GET /block/:id/status endpoint
func (h *ModerationHandler) CheckBlockStatus(c *gin.Context) {
	logger.Info("CheckBlockStatus request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get blocked user ID from URL parameter
	blockedIDStr := c.Param("id")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		logger.Error("Invalid blocked user ID", err, "blocked_id", blockedIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Execute use case
	isBlocked, err := h.getBlockedUsersUseCase.IsBlocked(c.Request.Context(), userID, blockedID)
	if err != nil {
		logger.Error("Failed to execute IsBlocked use case", err, "blocker_id", userID, "blocked_id", blockedID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to check block status", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Block status retrieved successfully", gin.H{
		"is_blocked": isBlocked,
	})
}

// CheckMutualBlock handles GET /block/:id/mutual endpoint
func (h *ModerationHandler) CheckMutualBlock(c *gin.Context) {
	logger.Info("CheckMutualBlock request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get other user ID from URL parameter
	otherUserIDStr := c.Param("id")
	otherUserID, err := uuid.Parse(otherUserIDStr)
	if err != nil {
		logger.Error("Invalid other user ID", err, "other_user_id", otherUserIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	// Get user ID from context (from auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		logger.Error("Invalid user ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid user ID", err)
		return
	}
	
	// Execute use case
	isMutual, err := h.getBlockedUsersUseCase.IsMutualBlock(c.Request.Context(), userID, otherUserID)
	if err != nil {
		logger.Error("Failed to execute IsMutualBlock use case", err, "user1_id", userID, "user2_id", otherUserID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to check mutual block", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Mutual block status retrieved successfully", gin.H{
		"is_mutual_block": isMutual,
	})
}