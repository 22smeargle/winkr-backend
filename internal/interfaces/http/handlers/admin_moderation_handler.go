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

// AdminModerationHandler handles admin moderation-related HTTP endpoints
type AdminModerationHandler struct {
	reviewReportUseCase *moderation.ReviewReportUseCase
	banUserUseCase    *moderation.BanUserUseCase
	validator           validator.Validator
}

// NewAdminModerationHandler creates a new admin moderation handler
func NewAdminModerationHandler(
	reviewReportUseCase *moderation.ReviewReportUseCase,
	banUserUseCase *moderation.BanUserUseCase,
	validator validator.Validator,
) *AdminModerationHandler {
	return &AdminModerationHandler{
		reviewReportUseCase: reviewReportUseCase,
		banUserUseCase:    banUserUseCase,
		validator:           validator,
	}
}

// GetReports handles GET /admin/reports endpoint
func (h *AdminModerationHandler) GetReports(c *gin.Context) {
	logger.Info("GetReports request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Parse query parameters
	status := c.Query("status")
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
	reports, err := h.reviewReportUseCase.GetReportsByStatus(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetReportsByStatus use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get reports", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Reports retrieved successfully", gin.H{
		"reports": reports,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetReportDetails handles GET /admin/reports/:id endpoint
func (h *AdminModerationHandler) GetReportDetails(c *gin.Context) {
	logger.Info("GetReportDetails request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get report ID from URL parameter
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logger.Error("Invalid report ID", err, "report_id", reportIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}
	
	// Execute use case
	report, err := h.reviewReportUseCase.GetReportDetails(c.Request.Context(), reportID, adminID)
	if err != nil {
		logger.Error("Failed to execute GetReportDetails use case", err, "report_id", reportID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get report details", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Report details retrieved successfully", report)
}

// ReviewReport handles POST /admin/reports/:id/review endpoint
func (h *AdminModerationHandler) ReviewReport(c *gin.Context) {
	logger.Info("ReviewReport request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get report ID from URL parameter
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logger.Error("Invalid report ID", err, "report_id", reportIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid report ID", err)
		return
	}
	
	var req moderation.ReviewReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "report_id", reportID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	req.ReportID = reportID
	req.ReviewerID = adminID
	
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}
	
	// Execute use case
	result, err := h.reviewReportUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute ReviewReport use case", err, "report_id", reportID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to review report", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Report reviewed successfully", result)
}

// BanUser handles POST /admin/users/:id/ban endpoint
func (h *AdminModerationHandler) BanUser(c *gin.Context) {
	logger.Info("BanUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", err, "user_id", userIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	var req moderation.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	req.UserID = userID
	req.BannerID = adminID
	
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}
	
	// Execute use case
	result, err := h.banUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute BanUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to ban user", err)
		return
	}
	
	response.Success(c, http.StatusCreated, "User banned successfully", result)
}

// SuspendUser handles POST /admin/users/:id/suspend endpoint
func (h *AdminModerationHandler) SuspendUser(c *gin.Context) {
	logger.Info("SuspendUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", err, "user_id", userIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
		return
	}
	
	var req moderation.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	req.UserID = userID
	req.BannerID = adminID
	
	// Set action type to suspend
	actionType := "suspend"
	req.ActionType = &actionType
	
	// Set default duration for suspension (7 days)
	duration := "7d"
	req.Duration = &duration
	
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}
	
	// Execute use case
	result, err := h.banUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute SuspendUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to suspend user", err)
		return
	}
	
	response.Success(c, http.StatusCreated, "User suspended successfully", result)
}

// GetModerationQueue handles GET /admin/moderation-queue endpoint
func (h *AdminModerationHandler) GetModerationQueue(c *gin.Context) {
	logger.Info("GetModerationQueue request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Parse query parameters
	priority := c.Query("priority")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	
	// Execute use case
	queueItems, err := h.reviewReportUseCase.GetPendingAppeals(c.Request.Context(), adminID, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetPendingAppeals use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get moderation queue", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Moderation queue retrieved successfully", gin.H{
		"queue_items": queueItems,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetBanHistory handles GET /admin/users/:id/bans endpoint
func (h *AdminModerationHandler) GetBanHistory(c *gin.Context) {
	logger.Info("GetBanHistory request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", err, "user_id", userIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
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
	banHistory, err := h.banUserUseCase.GetUserBanHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetUserBanHistory use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get ban history", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Ban history retrieved successfully", gin.H{
		"ban_history": banHistory,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetAppealHistory handles GET /admin/users/:id/appeals endpoint
func (h *AdminModerationHandler) GetAppealHistory(c *gin.Context) {
	logger.Info("GetAppealHistory request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", err, "user_id", userIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid user ID", err)
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
	appealHistory, err := h.banUserUseCase.GetUserAppealHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to execute GetUserAppealHistory use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get appeal history", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Appeal history retrieved successfully", gin.H{
		"appeal_history": appealHistory,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// ReviewAppeal handles POST /admin/appeals/:id/review endpoint
func (h *AdminModerationHandler) ReviewAppeal(c *gin.Context) {
	logger.Info("ReviewAppeal request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Get appeal ID from URL parameter
	appealIDStr := c.Param("id")
	appealID, err := uuid.Parse(appealIDStr)
	if err != nil {
		logger.Error("Invalid appeal ID", err, "appeal_id", appealIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid appeal ID", err)
		return
	}
	
	var req struct {
		Approved bool   `json:"approved" validate:"required"`
		Notes    string `json:"notes" validate:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "appeal_id", appealID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}
	
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}
	
	// Execute use case
	err = h.banUserUseCase.ReviewAppeal(c.Request.Context(), appealID, adminID, req.Approved, req.Notes)
	if err != nil {
		logger.Error("Failed to execute ReviewAppeal use case", err, "appeal_id", appealID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to review appeal", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Appeal reviewed successfully", gin.H{
		"message": "Appeal reviewed successfully",
	})
}

// GetModerationAnalytics handles GET /admin/analytics endpoint
func (h *AdminModerationHandler) GetModerationAnalytics(c *gin.Context) {
	logger.Info("GetModerationAnalytics request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())
	
	// Get admin ID from context (from auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		logger.Error("Admin ID not found in context", nil, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Admin authentication required", nil)
		return
	}
	
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		logger.Error("Invalid admin ID in context", err, "ip", c.ClientIP())
		response.Error(c, http.StatusUnauthorized, "Invalid admin ID", err)
		return
	}
	
	// Parse query parameters
	period := c.DefaultQuery("period", "7d") // Default to last 7 days
	
	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}
	
	if !isValidPeriod {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}
	
	// Execute use case
	analytics, err := h.reviewReportUseCase.GetModerationAnalytics(c.Request.Context(), adminID, period)
	if err != nil {
		logger.Error("Failed to execute GetModerationAnalytics use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get moderation analytics", err)
		return
	}
	
	response.Success(c, http.StatusOK, "Moderation analytics retrieved successfully", analytics)
}