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

// AdminSystemHandler handles admin system management HTTP endpoints
type AdminSystemHandler struct {
	getSystemHealthUseCase *admin.GetSystemHealthUseCase
	getSystemMetricsUseCase *admin.GetSystemMetricsUseCase
	getSystemLogsUseCase   *admin.GetSystemLogsUseCase
	maintenanceUseCase     *admin.MaintenanceUseCase
	configUseCase          *admin.ConfigUseCase
	validator              validator.Validator
}

// NewAdminSystemHandler creates a new admin system handler
func NewAdminSystemHandler(
	getSystemHealthUseCase *admin.GetSystemHealthUseCase,
	getSystemMetricsUseCase *admin.GetSystemMetricsUseCase,
	getSystemLogsUseCase *admin.GetSystemLogsUseCase,
	maintenanceUseCase *admin.MaintenanceUseCase,
	configUseCase *admin.ConfigUseCase,
	validator validator.Validator,
) *AdminSystemHandler {
	return &AdminSystemHandler{
		getSystemHealthUseCase: getSystemHealthUseCase,
		getSystemMetricsUseCase: getSystemMetricsUseCase,
		getSystemLogsUseCase:   getSystemLogsUseCase,
		maintenanceUseCase:     maintenanceUseCase,
		configUseCase:          configUseCase,
		validator:              validator,
	}
}

// GetSystemHealth handles GET /admin/system/health endpoint
func (h *AdminSystemHandler) GetSystemHealth(c *gin.Context) {
	logger.Info("GetSystemHealth request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	detailed := c.DefaultQuery("detailed", "false") == "true"

	// Create request
	req := admin.GetSystemHealthRequest{
		AdminID:  adminID,
		Detailed: detailed,
	}

	// Execute use case
	health, err := h.getSystemHealthUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetSystemHealth use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get system health", err)
		return
	}

	response.Success(c, http.StatusOK, "System health retrieved successfully", health)
}

// GetSystemMetrics handles GET /admin/system/metrics endpoint
func (h *AdminSystemHandler) GetSystemMetrics(c *gin.Context) {
	logger.Info("GetSystemMetrics request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	metricType := c.Query("type") // cpu, memory, disk, network, all
	period := c.DefaultQuery("period", "1h") // 1h, 6h, 24h, 7d

	// Validate metric type
	validTypes := []string{"cpu", "memory", "disk", "network", "all"}
	isValidType := false
	for _, validType := range validTypes {
		if metricType == validType {
			isValidType = true
			break
		}
	}

	if metricType != "" && !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid metric type parameter", nil)
		return
	}

	// Validate period
	validPeriods := []string{"1h", "6h", "24h", "7d"}
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

	// Create request
	req := admin.GetSystemMetricsRequest{
		AdminID:    adminID,
		MetricType: metricType,
		Period:     period,
	}

	// Execute use case
	metrics, err := h.getSystemMetricsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetSystemMetrics use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get system metrics", err)
		return
	}

	response.Success(c, http.StatusOK, "System metrics retrieved successfully", metrics)
}

// GetSystemLogs handles GET /admin/system/logs endpoint
func (h *AdminSystemHandler) GetSystemLogs(c *gin.Context) {
	logger.Info("GetSystemLogs request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	level := c.DefaultQuery("level", "info") // debug, info, warn, error
	service := c.Query("service") // api, worker, all
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	isValidLevel := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			isValidLevel = true
			break
		}
	}

	if !isValidLevel {
		response.Error(c, http.StatusBadRequest, "Invalid log level parameter", nil)
		return
	}

	// Validate service
	validServices := []string{"api", "worker", "all"}
	isValidService := false
	for _, validService := range validServices {
		if service == validService {
			isValidService = true
			break
		}
	}

	if service != "" && !isValidService {
		response.Error(c, http.StatusBadRequest, "Invalid service parameter", nil)
		return
	}

	// Validate pagination parameters
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
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
	req := admin.GetSystemLogsRequest{
		AdminID:   adminID,
		Level:      level,
		Service:    service,
		Limit:      limit,
		Offset:     offset,
		StartTime:  startTime,
		EndTime:    endTime,
		Search:     search,
	}

	// Execute use case
	logs, err := h.getSystemLogsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetSystemLogs use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get system logs", err)
		return
	}

	response.Success(c, http.StatusOK, "System logs retrieved successfully", gin.H{
		"logs": logs.Logs,
		"pagination": gin.H{
			"total":  logs.Total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// EnableMaintenance handles POST /admin/system/maintenance endpoint
func (h *AdminSystemHandler) EnableMaintenance(c *gin.Context) {
	logger.Info("EnableMaintenance request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.MaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.AdminID = adminID
	req.Enable = true

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.maintenanceUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute EnableMaintenance use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to enable maintenance mode", err)
		return
	}

	response.Success(c, http.StatusOK, "Maintenance mode enabled successfully", result)
}

// DisableMaintenance handles DELETE /admin/system/maintenance endpoint
func (h *AdminSystemHandler) DisableMaintenance(c *gin.Context) {
	logger.Info("DisableMaintenance request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.MaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	req.AdminID = adminID
	req.Enable = false

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.maintenanceUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute DisableMaintenance use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to disable maintenance mode", err)
		return
	}

	response.Success(c, http.StatusOK, "Maintenance mode disabled successfully", result)
}

// GetSystemConfig handles GET /admin/system/config endpoint
func (h *AdminSystemHandler) GetSystemConfig(c *gin.Context) {
	logger.Info("GetSystemConfig request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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
	category := c.Query("category") // app, database, redis, storage, email, payment, all

	// Validate category
	validCategories := []string{"app", "database", "redis", "storage", "email", "payment", "all"}
	isValidCategory := false
	for _, validCategory := range validCategories {
		if category == validCategory {
			isValidCategory = true
			break
		}
	}

	if category != "" && !isValidCategory {
		response.Error(c, http.StatusBadRequest, "Invalid category parameter", nil)
		return
	}

	// Create request
	req := admin.GetSystemConfigRequest{
		AdminID:  adminID,
		Category: category,
	}

	// Execute use case
	config, err := h.configUseCase.GetConfig(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetSystemConfig use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get system configuration", err)
		return
	}

	response.Success(c, http.StatusOK, "System configuration retrieved successfully", config)
}

// UpdateSystemConfig handles PUT /admin/system/config endpoint
func (h *AdminSystemHandler) UpdateSystemConfig(c *gin.Context) {
	logger.Info("UpdateSystemConfig request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req admin.UpdateSystemConfigRequest
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
	result, err := h.configUseCase.UpdateConfig(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute UpdateSystemConfig use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to update system configuration", err)
		return
	}

	response.Success(c, http.StatusOK, "System configuration updated successfully", result)
}

// RestartService handles POST /admin/system/restart endpoint
func (h *AdminSystemHandler) RestartService(c *gin.Context) {
	logger.Info("RestartService request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req struct {
		Service string `json:"service" validate:"required"`
		Reason  string `json:"reason" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate service
	validServices := []string{"api", "worker", "all"}
	isValidService := false
	for _, validService := range validServices {
		if req.Service == validService {
			isValidService = true
			break
		}
	}

	if !isValidService {
		response.Error(c, http.StatusBadRequest, "Invalid service parameter", nil)
		return
	}

	// Create restart request
	restartReq := admin.RestartServiceRequest{
		AdminID: adminID,
		Service: req.Service,
		Reason:  req.Reason,
	}

	// Validate request
	if err := h.validator.Struct(restartReq); err != nil {
		logger.Error("Request validation failed", err, "request", restartReq, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.maintenanceUseCase.RestartService(c.Request.Context(), restartReq)
	if err != nil {
		logger.Error("Failed to execute RestartService use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to restart service", err)
		return
	}

	response.Success(c, http.StatusOK, "Service restart initiated successfully", result)
}

// GetSystemStatus handles GET /admin/system/status endpoint
func (h *AdminSystemHandler) GetSystemStatus(c *gin.Context) {
	logger.Info("GetSystemStatus request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	// Create request
	req := admin.GetSystemStatusRequest{
		AdminID: adminID,
	}

	// Execute use case
	status, err := h.getSystemHealthUseCase.GetSystemStatus(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to execute GetSystemStatus use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to get system status", err)
		return
	}

	response.Success(c, http.StatusOK, "System status retrieved successfully", status)
}

// ClearCache handles POST /admin/system/cache/clear endpoint
func (h *AdminSystemHandler) ClearCache(c *gin.Context) {
	logger.Info("ClearCache request received", "ip", c.ClientIP(), "user_agent", c.Request.UserAgent())

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

	var req struct {
		CacheType string `json:"cache_type" validate:"required"` // redis, sessions, photos, all
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind request", err, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate cache type
	validTypes := []string{"redis", "sessions", "photos", "all"}
	isValidType := false
	for _, validType := range validTypes {
		if req.CacheType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		response.Error(c, http.StatusBadRequest, "Invalid cache_type parameter", nil)
		return
	}

	// Create clear cache request
	clearReq := admin.ClearCacheRequest{
		AdminID:   adminID,
		CacheType: req.CacheType,
	}

	// Validate request
	if err := h.validator.Struct(clearReq); err != nil {
		logger.Error("Request validation failed", err, "request", clearReq, "ip", c.ClientIP())
		response.Error(c, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Execute use case
	result, err := h.maintenanceUseCase.ClearCache(c.Request.Context(), clearReq)
	if err != nil {
		logger.Error("Failed to execute ClearCache use case", err, "admin_id", adminID, "ip", c.ClientIP())
		response.Error(c, http.StatusInternalServerError, "Failed to clear cache", err)
		return
	}

	response.Success(c, http.StatusOK, "Cache cleared successfully", result)
}