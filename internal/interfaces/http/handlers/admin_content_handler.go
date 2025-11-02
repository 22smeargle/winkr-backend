package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/usecases/admin"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/response"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// AdminContentHandler handles admin content moderation HTTP endpoints
type AdminContentHandler struct {
	getReportedPhotosUseCase   *admin.GetReportedPhotosUseCase
	approvePhotoUseCase        *admin.ApprovePhotoUseCase
	rejectPhotoUseCase         *admin.RejectPhotoUseCase
	getReportedMessagesUseCase *admin.GetReportedMessagesUseCase
	deleteMessageUseCase       *admin.DeleteMessageUseCase
	validator                  validator.Validator
}

// NewAdminContentHandler creates a new admin content handler
func NewAdminContentHandler(
	getReportedPhotosUseCase *admin.GetReportedPhotosUseCase,
	approvePhotoUseCase *admin.ApprovePhotoUseCase,
	rejectPhotoUseCase *admin.RejectPhotoUseCase,
	getReportedMessagesUseCase *admin.GetReportedMessagesUseCase,
	deleteMessageUseCase *admin.DeleteMessageUseCase,
	validator validator.Validator,
) *AdminContentHandler {
	return &AdminContentHandler{
		getReportedPhotosUseCase:   getReportedPhotosUseCase,
		approvePhotoUseCase:        approvePhotoUseCase,
		rejectPhotoUseCase:         rejectPhotoUseCase,
		getReportedMessagesUseCase: getReportedMessagesUseCase,
		deleteMessageUseCase:       deleteMessageUseCase,
		validator:                  validator,
	}
}

// GetReportedPhotos handles GET /admin/content/photos endpoint
func (h *AdminContentHandler) GetReportedPhotos(c *gin.Context) {
	logger.Info("GetReportedPhotos request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status") // pending, approved, rejected, all
	priority := c.Query("priority") // high, medium, low, all

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Validate status parameter
	validStatuses := []string{"pending", "approved", "rejected", "all"}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValidStatus = true
			break
		}
	}

	if status != "" && !isValidStatus {
		response.Error(c, http.StatusBadRequest, "Invalid status parameter", nil)
		return
	}

	// Validate priority parameter
	validPriorities := []string{"high", "medium", "low", "all"}
	isValidPriority := false
	for _, validPriority := range validPriorities {
		if priority == validPriority {
			isValidPriority = true
			break
		}
	}

	if priority != "" && !isValidPriority {
		response.Error(c, http.StatusBadRequest, "Invalid priority parameter", nil)
		return
	}

	// Create request
	req := admin.GetReportedPhotosRequest{
		AdminID:  adminID,
		Status:   status,
		Priority: priority,
		Limit:    limit,
		Offset:   offset,
	}

	// Execute use case
	result, err := h.getReportedPhotosUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetReportedPhotos use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get reported photos", err)
		return
	}

	response.Success(c, http.StatusOK, "Reported photos retrieved successfully", gin.H{
		"photos": result.Photos,
		"pagination": gin.H{
			"total":  result.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// ApprovePhoto handles POST /admin/content/photos/:id/approve endpoint
func (h *AdminContentHandler) ApprovePhoto(c *gin.Context) {
	logger.Info("ApprovePhoto request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Get photo ID from URL parameter
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		logger.Error("Invalid photo ID", err, "photo_id", photoIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid photo ID", err)
		return
	}

	var req admin.ApprovePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "photo_id", photoID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.PhotoID = photoID
	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.approvePhotoUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute ApprovePhoto use case", err, "photo_id", photoID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to approve photo", err)
		return
	}

	response.Success(c, http.StatusOK, "Photo approved successfully", result)
}

// RejectPhoto handles POST /admin/content/photos/:id/reject endpoint
func (h *AdminContentHandler) RejectPhoto(c *gin.Context) {
	logger.Info("RejectPhoto request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Get photo ID from URL parameter
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		logger.Error("Invalid photo ID", err, "photo_id", photoIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid photo ID", err)
		return
	}

	var req admin.RejectPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "photo_id", photoID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.PhotoID = photoID
	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.rejectPhotoUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute RejectPhoto use case", err, "photo_id", photoID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to reject photo", err)
		return
	}

	response.Success(c, http.StatusOK, "Photo rejected successfully", result)
}

// GetReportedMessages handles GET /admin/content/messages endpoint
func (h *AdminContentHandler) GetReportedMessages(c *gin.Context) {
	logger.Info("GetReportedMessages request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status") // pending, reviewed, dismissed, all
	priority := c.Query("priority") // high, medium, low, all

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Validate status parameter
	validStatuses := []string{"pending", "reviewed", "dismissed", "all"}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValidStatus = true
			break
		}
	}

	if status != "" && !isValidStatus {
		response.Error(c, http.StatusBadRequest, "Invalid status parameter", nil)
		return
	}

	// Validate priority parameter
	validPriorities := []string{"high", "medium", "low", "all"}
	isValidPriority := false
	for _, validPriority := range validPriorities {
		if priority == validPriority {
			isValidPriority = true
			break
		}
	}

	if priority != "" && !isValidPriority {
		response.Error(c, http.StatusBadRequest, "Invalid priority parameter", nil)
		return
	}

	// Create request
	req := admin.GetReportedMessagesRequest{
		AdminID:  adminID,
		Status:   status,
		Priority: priority,
		Limit:    limit,
		Offset:   offset,
	}

	// Execute use case
	result, err := h.getReportedMessagesUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetReportedMessages use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get reported messages", err)
		return
	}

	response.Success(c, http.StatusOK, "Reported messages retrieved successfully", gin.H{
		"messages": result.Messages,
		"pagination": gin.H{
			"total":  result.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// DeleteMessage handles POST /admin/content/messages/:id/delete endpoint
func (h *AdminContentHandler) DeleteMessage(c *gin.Context) {
	logger.Info("DeleteMessage request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Get message ID from URL parameter
	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		logger.Error("Invalid message ID", err, "message_id", messageIDStr, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid message ID", err)
		return
	}

	var req admin.DeleteMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "message_id", messageID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.MessageID = messageID
	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.deleteMessageUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute DeleteMessage use case", err, "message_id", messageID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to delete message", err)
		return
	}

	response.Success(c, http.StatusOK, "Message deleted successfully", result)
}

// GetContentQueue handles GET /admin/content/queue endpoint
func (h *AdminContentHandler) GetContentQueue(c *gin.Context) {
	logger.Info("GetContentQueue request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	contentType := c.DefaultQuery("type", "all") // photos, messages, all

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Validate content type parameter
	validTypes := []string{"photos", "messages", "all"}
	isValidType := false
	for _, validType := range validTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid type parameter", nil)
		return
	}

	// Create request
	req := admin.GetContentQueueRequest{
		AdminID:     adminID,
		ContentType: contentType,
		Limit:       limit,
		Offset:      offset,
	}

	// Execute use case
	result, err := h.getReportedPhotosUseCase.GetContentQueue(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetContentQueue use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get content queue", err)
		return
	}

	response.Success(c, http.StatusOK, "Content queue retrieved successfully", gin.H{
		"queue": result.Queue,
		"pagination": gin.H{
			"total":  result.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// BulkApprovePhotos handles POST /admin/content/photos/bulk-approve endpoint
func (h *AdminContentHandler) BulkApprovePhotos(c *gin.Context) {
	logger.Info("BulkApprovePhotos request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.BulkApprovePhotosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.approvePhotoUseCase.BulkApprove(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute BulkApprovePhotos use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to bulk approve photos", err)
		return
	}

	response.Success(c, http.StatusOK, "Photos bulk approved successfully", result)
}

// BulkRejectPhotos handles POST /admin/content/photos/bulk-reject endpoint
func (h *AdminContentHandler) BulkRejectPhotos(c *gin.Context) {
	logger.Info("BulkRejectPhotos request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.BulkRejectPhotosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.rejectPhotoUseCase.BulkReject(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute BulkRejectPhotos use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to bulk reject photos", err)
		return
	}

	response.Success(c, http.StatusOK, "Photos bulk rejected successfully", result)
}

// GetContentAnalytics handles GET /admin/content/analytics endpoint
func (h *AdminContentHandler) GetContentAnalytics(c *gin.Context) {
	logger.Info("GetContentAnalytics request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	contentType := c.DefaultQuery("type", "all") // photos, messages, all

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

	// Validate content type parameter
	validTypes := []string{"photos", "messages", "all"}
	isValidType := false
	for _, validType := range validTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid type parameter", nil)
		return
	}

	// Create request
	req := admin.GetContentAnalyticsRequest{
		AdminID:     adminID,
		Period:      period,
		ContentType: contentType,
	}

	// Execute use case
	analytics, err := h.getReportedPhotosUseCase.GetAnalytics(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetContentAnalytics use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get content analytics", err)
		return
	}

	response.Success(c, http.StatusOK, "Content analytics retrieved successfully", analytics)
}