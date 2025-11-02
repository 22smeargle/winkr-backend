package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/usecases/admin"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/response"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// AdminAnalyticsHandler handles admin analytics HTTP endpoints
type AdminAnalyticsHandler struct {
	getPlatformStatsUseCase *admin.GetPlatformStatsUseCase
	getUserStatsUseCase    *admin.GetUserStatsUseCase
	getMatchStatsUseCase   *admin.GetMatchStatsUseCase
	getMessageStatsUseCase  *admin.GetMessageStatsUseCase
	getPaymentStatsUseCase  *admin.GetPaymentStatsUseCase
	getVerificationStatsUseCase *admin.GetVerificationStatsUseCase
	validator              validator.Validator
}

// NewAdminAnalyticsHandler creates a new admin analytics handler
func NewAdminAnalyticsHandler(
	getPlatformStatsUseCase *admin.GetPlatformStatsUseCase,
	getUserStatsUseCase *admin.GetUserStatsUseCase,
	getMatchStatsUseCase *admin.GetMatchStatsUseCase,
	getMessageStatsUseCase *admin.GetMessageStatsUseCase,
	getPaymentStatsUseCase *admin.GetPaymentStatsUseCase,
	getVerificationStatsUseCase *admin.GetVerificationStatsUseCase,
	validator validator.Validator,
) *AdminAnalyticsHandler {
	return &AdminAnalyticsHandler{
		getPlatformStatsUseCase: getPlatformStatsUseCase,
		getUserStatsUseCase:    getUserStatsUseCase,
		getMatchStatsUseCase:   getMatchStatsUseCase,
		getMessageStatsUseCase:  getMessageStatsUseCase,
		getPaymentStatsUseCase:  getPaymentStatsUseCase,
		getVerificationStatsUseCase: getVerificationStatsUseCase,
		validator:              validator,
	}
}

// GetPlatformStats handles GET /admin/stats endpoint
func (h *AdminAnalyticsHandler) GetPlatformStats(c *gin.Context) {
	logger.Info("GetPlatformStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetPlatformStatsRequest{
		AdminID:   adminID,
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Execute use case
	stats, err := h.getPlatformStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetPlatformStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get platform statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Platform statistics retrieved successfully", stats)
}

// GetUserStats handles GET /admin/stats/users endpoint
func (h *AdminAnalyticsHandler) GetUserStats(c *gin.Context) {
	logger.Info("GetUserStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Validate group_by parameter
	validGroupBy := []string{"day", "week", "month"}
	isValidGroupBy := false
	for _, valid := range validGroupBy {
		if groupBy == valid {
			isValidGroupBy = true
			break
		}
	}

	if !isValidGroupBy {
		response.Error(c, http.StatusBadRequest, "Invalid group_by parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetUserStatsRequest{
		AdminID:   adminID,
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   groupBy,
	}

	// Execute use case
	stats, err := h.getUserStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetUserStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get user statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "User statistics retrieved successfully", stats)
}

// GetMatchStats handles GET /admin/stats/matches endpoint
func (h *AdminAnalyticsHandler) GetMatchStats(c *gin.Context) {
	logger.Info("GetMatchStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Validate group_by parameter
	validGroupBy := []string{"day", "week", "month"}
	isValidGroupBy := false
	for _, valid := range validGroupBy {
		if groupBy == valid {
			isValidGroupBy = true
			break
		}
	}

	if !isValidGroupBy {
		response.Error(c, http.StatusBadRequest, "Invalid group_by parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetMatchStatsRequest{
		AdminID:   adminID,
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   groupBy,
	}

	// Execute use case
	stats, err := h.getMatchStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetMatchStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get match statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Match statistics retrieved successfully", stats)
}

// GetMessageStats handles GET /admin/stats/messages endpoint
func (h *AdminAnalyticsHandler) GetMessageStats(c *gin.Context) {
	logger.Info("GetMessageStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Validate group_by parameter
	validGroupBy := []string{"day", "week", "month"}
	isValidGroupBy := false
	for _, valid := range validGroupBy {
		if groupBy == valid {
			isValidGroupBy = true
			break
		}
	}

	if !isValidGroupBy {
		response.Error(c, http.StatusBadRequest, "Invalid group_by parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetMessageStatsRequest{
		AdminID:   adminID,
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   groupBy,
	}

	// Execute use case
	stats, err := h.getMessageStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetMessageStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get message statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Message statistics retrieved successfully", stats)
}

// GetPaymentStats handles GET /admin/stats/payments endpoint
func (h *AdminAnalyticsHandler) GetPaymentStats(c *gin.Context) {
	logger.Info("GetPaymentStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month
	currency := c.DefaultQuery("currency", "USD")

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Validate group_by parameter
	validGroupBy := []string{"day", "week", "month"}
	isValidGroupBy := false
	for _, valid := range validGroupBy {
		if groupBy == valid {
			isValidGroupBy = true
			break
		}
	}

	if !isValidGroupBy {
		response.Error(c, http.StatusBadRequest, "Invalid group_by parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetPaymentStatsRequest{
		AdminID:   adminID,
		Period:    period,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   groupBy,
		Currency:  currency,
	}

	// Execute use case
	stats, err := h.getPaymentStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetPaymentStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get payment statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Payment statistics retrieved successfully", stats)
}

// GetVerificationStats handles GET /admin/stats/verification endpoint
func (h *AdminAnalyticsHandler) GetVerificationStats(c *gin.Context) {
	logger.Info("GetVerificationStats request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day") // day, week, month
	verificationType := c.Query("type") // selfie, document, all

	// Validate period parameter
	validPeriods := []string{"1d", "7d", "30d", "90d", "1y"}
	isValidPeriod := false
	for _, validPeriod := range validPeriods {
		if period == validPeriod {
			isValidPeriod = true
			break
		}
	}

	if !isValidPeriod && startDate == "" && endDate == "" {
		response.Error(c, http.StatusBadRequest, "Invalid period parameter", nil)
		return
	}

	// Validate group_by parameter
	validGroupBy := []string{"day", "week", "month"}
	isValidGroupBy := false
	for _, valid := range validGroupBy {
		if groupBy == valid {
			isValidGroupBy = true
			break
		}
	}

	if !isValidGroupBy {
		response.Error(c, http.StatusBadRequest, "Invalid group_by parameter", nil)
		return
	}

	// Validate verification type parameter
	validTypes := []string{"selfie", "document", "all"}
	isValidType := false
	for _, validType := range validTypes {
		if verificationType == validType {
			isValidType = true
			break
		}
	}

	if verificationType != "" && !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid verification type parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create request
	req := admin.GetVerificationStatsRequest{
		AdminID:          adminID,
		Period:           period,
		StartTime:        startTime,
		EndTime:          endTime,
		GroupBy:          groupBy,
		VerificationType: verificationType,
	}

	// Execute use case
	stats, err := h.getVerificationStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetVerificationStats use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get verification statistics", err)
		return
	}

	response.Success(c, http.StatusOK, "Verification statistics retrieved successfully", stats)
}

// GetDashboardData handles GET /admin/dashboard endpoint
func (h *AdminAnalyticsHandler) GetDashboardData(c *gin.Context) {
	logger.Info("GetDashboardData request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Create request for dashboard data
	req := admin.GetDashboardDataRequest{
		AdminID: adminID,
		Period:  period,
	}

	// Execute use case
	dashboardData, err := h.getPlatformStatsUseCase.GetDashboardData(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetDashboardData use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get dashboard data", err)
		return
	}

	response.Success(c, http.StatusOK, "Dashboard data retrieved successfully", dashboardData)
}

// ExportAnalytics handles GET /admin/analytics/export endpoint
func (h *AdminAnalyticsHandler) ExportAnalytics(c *gin.Context) {
	logger.Info("ExportAnalytics request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	reportType := c.Query("type") // users, matches, messages, payments, verification
	format := c.DefaultQuery("format", "csv") // csv, xlsx, json
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate report type
	validTypes := []string{"users", "matches", "messages", "payments", "verification"}
	isValidType := false
	for _, validType := range validTypes {
		if reportType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid report type parameter", nil)
		return
	}

	// Validate format
	validFormats := []string{"csv", "xlsx", "json"}
	isValidFormat := false
	for _, validFormat := range validFormats {
		if format == validFormat {
			isValidFormat = true
			break
		}
	}

	if !isValidFormat {
		response.Error(c, http.StatusBadRequest, "Invalid format parameter", nil)
		return
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid start_date format", nil)
			return
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		} else {
			response.Error(c, http.StatusBadRequest, "Invalid end_date format", nil)
			return
		}
	}

	// Create export request
	req := admin.ExportAnalyticsRequest{
		AdminID:   adminID,
		Type:      reportType,
		Format:    format,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Execute use case
	exportData, err := h.getPlatformStatsUseCase.ExportAnalytics(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute ExportAnalytics use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to export analytics", err)
		return
	}

	// Set appropriate headers for file download
	c.Header("Content-Disposition", "attachment; filename="+exportData.Filename)
	c.Header("Content-Type", exportData.ContentType)
	c.Data(http.StatusOK, exportData.ContentType, exportData.Data)
}