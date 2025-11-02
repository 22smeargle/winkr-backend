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

// AdminUserHandler handles admin user management HTTP endpoints
type AdminUserHandler struct {
	getUsersUseCase        *admin.GetUsersUseCase
	getUserDetailsUseCase *admin.GetUserDetailsUseCase
	updateUserUseCase     *admin.UpdateUserUseCase
	deleteUserUseCase     *admin.DeleteUserUseCase
	suspendUserUseCase    *admin.SuspendUserUseCase
	validator             validator.Validator
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(
	getUsersUseCase *admin.GetUsersUseCase,
	getUserDetailsUseCase *admin.GetUserDetailsUseCase,
	updateUserUseCase *admin.UpdateUserUseCase,
	deleteUserUseCase *admin.DeleteUserUseCase,
	suspendUserUseCase *admin.SuspendUserUseCase,
	validator validator.Validator,
) *AdminUserHandler {
	return &AdminUserHandler{
		getUsersUseCase:        getUsersUseCase,
		getUserDetailsUseCase: getUserDetailsUseCase,
		updateUserUseCase:     updateUserUseCase,
		deleteUserUseCase:     deleteUserUseCase,
		suspendUserUseCase:    suspendUserUseCase,
		validator:             validator,
	}
}

// GetUsers handles GET /admin/users endpoint
func (h *AdminUserHandler) GetUsers(c *gin.Context) {
	logger.Info("GetUsers request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	status := c.Query("status")
	verified := c.Query("verified")
	premium := c.Query("premium")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Create filter options
	filters := admin.UserFilters{
		Status:    status,
		Verified:  verified,
		Premium:   premium,
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Offset:    offset,
	}

	// Execute use case
	result, err := h.getUsersUseCase.Execute(c.Request.Context(), adminID, filters)
	if err != nil {
		logger.Error("Failed to execute GetUsers use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get users", err)
		return
	}

	response.Success(c, http.StatusOK, "Users retrieved successfully", gin.H{
		"users": result.Users,
		"pagination": gin.H{
			"total":  result.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUserDetails handles GET /admin/users/:id endpoint
func (h *AdminUserHandler) GetUserDetails(c *gin.Context) {
	logger.Info("GetUserDetails request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Execute use case
	userDetails, err := h.getUserDetailsUseCase.Execute(c.Request.Context(), adminID, userID)
	if err != nil {
		logger.Error("Failed to execute GetUserDetails use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get user details", err)
		return
	}

	response.Success(c, http.StatusOK, "User details retrieved successfully", userDetails)
}

// UpdateUser handles PUT /admin/users/:id endpoint
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	logger.Info("UpdateUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.UserID = userID
	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.updateUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute UpdateUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", result)
}

// DeleteUser handles DELETE /admin/users/:id endpoint
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	logger.Info("DeleteUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	hardDelete := c.DefaultQuery("hard_delete", "false") == "true"
	reason := c.Query("reason")

	// Create request
	req := admin.DeleteUserRequest{
		UserID:    userID,
		AdminID:    adminID,
		HardDelete: hardDelete,
		Reason:     reason,
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	err = h.deleteUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute DeleteUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", gin.H{
		"message": "User deleted successfully",
	})
}

// SuspendUser handles POST /admin/users/:id/suspend endpoint
func (h *AdminUserHandler) SuspendUser(c *gin.Context) {
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

	var req admin.SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.UserID = userID
	req.AdminID = adminID

	// Set default suspension duration if not provided
	if req.Duration == "" {
		req.Duration = "7d" // Default to 7 days
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.suspendUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute SuspendUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to suspend user", err)
		return
	}

	response.Success(c, http.StatusCreated, "User suspended successfully", result)
}

// UnsuspendUser handles POST /admin/users/:id/unsuspend endpoint
func (h *AdminUserHandler) UnsuspendUser(c *gin.Context) {
	logger.Info("UnsuspendUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req struct {
		Reason string `json:"reason" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Create unsuspend request
	unsuspendReq := admin.UnsuspendUserRequest{
		UserID:  userID,
		AdminID: adminID,
		Reason:  req.Reason,
	}

	// Validate request
	if err := h.validator.Struct(unsuspendReq); err != nil {
		logger.Error("Request validation failed", err, "request", unsuspendReq, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.suspendUserUseCase.Unsuspend(c.Request.Context(), unsuspendReq)
	if err != nil {
		logger.Error("Failed to execute UnsuspendUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to unsuspend user", err)
		return
	}

	response.Success(c, http.StatusOK, "User unsuspended successfully", result)
}

// GetUserActivity handles GET /admin/users/:id/activity endpoint
func (h *AdminUserHandler) GetUserActivity(c *gin.Context) {
	logger.Info("GetUserActivity request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	activityType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Parse dates if provided
	var startTime, endTime *time.Time
	if startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			startTime = &parsed
		}
	}
	if endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			endTime = &parsed
		}
	}

	// Create activity request
	req := admin.GetUserActivityRequest{
		UserID:       userID,
		AdminID:      adminID,
		ActivityType: activityType,
		StartTime:    startTime,
		EndTime:      endTime,
		Limit:        limit,
		Offset:       offset,
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.getUserDetailsUseCase.GetActivity(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetUserActivity use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get user activity", err)
		return
	}

	response.Success(c, http.StatusOK, "User activity retrieved successfully", gin.H{
		"activities": result.Activities,
		"pagination": gin.H{
			"total":  result.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// BanUser handles POST /admin/users/:id/ban endpoint
func (h *AdminUserHandler) BanUser(c *gin.Context) {
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

	var req admin.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.UserID = userID
	req.AdminID = adminID

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.suspendUserUseCase.Ban(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute BanUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to ban user", err)
		return
	}

	response.Success(c, http.StatusCreated, "User banned successfully", result)
}

// UnbanUser handles POST /admin/users/:id/unban endpoint
func (h *AdminUserHandler) UnbanUser(c *gin.Context) {
	logger.Info("UnbanUser request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req struct {
		Reason string `json:"reason" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "user_id", userID, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Create unban request
	unbanReq := admin.UnbanUserRequest{
		UserID:  userID,
		AdminID: adminID,
		Reason:  req.Reason,
	}

	// Validate request
	if err := h.validator.Struct(unbanReq); err != nil {
		logger.Error("Request validation failed", err, "request", unbanReq, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.suspendUserUseCase.Unban(c.Request.Context(), unbanReq)
	if err != nil {
		logger.Error("Failed to execute UnbanUser use case", err, "user_id", userID, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to unban user", err)
		return
	}

	response.Success(c, http.StatusOK, "User unbanned successfully", result)
}