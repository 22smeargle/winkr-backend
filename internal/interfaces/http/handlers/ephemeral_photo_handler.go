package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/ephemeral_photo"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// EphemeralPhotoHandler handles HTTP requests for ephemeral photos
type EphemeralPhotoHandler struct {
	uploadUseCase    *ephemeral_photo.UploadEphemeralPhotoUseCase
	viewUseCase      *ephemeral_photo.ViewEphemeralPhotoUseCase
	deleteUseCase    *ephemeral_photo.DeleteEphemeralPhotoUseCase
	getStatusUseCase  *ephemeral_photo.GetEphemeralPhotoStatusUseCase
	expireUseCase    *ephemeral_photo.ExpireEphemeralPhotoUseCase
	getUserUseCase   *ephemeral_photo.GetUserEphemeralPhotosUseCase
}

// NewEphemeralPhotoHandler creates a new ephemeral photo handler
func NewEphemeralPhotoHandler(
	uploadUseCase *ephemeral_photo.UploadEphemeralPhotoUseCase,
	viewUseCase *ephemeral_photo.ViewEphemeralPhotoUseCase,
	deleteUseCase *ephemeral_photo.DeleteEphemeralPhotoUseCase,
	getStatusUseCase *ephemeral_photo.GetEphemeralPhotoStatusUseCase,
	expireUseCase *ephemeral_photo.ExpireEphemeralPhotoUseCase,
	getUserUseCase *ephemeral_photo.GetUserEphemeralPhotosUseCase,
) *EphemeralPhotoHandler {
	return &EphemeralPhotoHandler{
		uploadUseCase:   uploadUseCase,
		viewUseCase:     viewUseCase,
		deleteUseCase:   deleteUseCase,
		getStatusUseCase: getStatusUseCase,
		expireUseCase:   expireUseCase,
		getUserUseCase:  getUserUseCase,
	}
}

// UploadEphemeralPhoto handles uploading an ephemeral photo
func (h *EphemeralPhotoHandler) UploadEphemeralPhoto(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Parse request body
	var req ephemeral_photo.UploadEphemeralPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set user ID from context
	req.UserID = userUUID

	// Execute use case
	response, err := h.uploadUseCase.Execute(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to upload ephemeral photo", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload ephemeral photo")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, response)
}

// ViewEphemeralPhoto handles viewing an ephemeral photo
func (h *EphemeralPhotoHandler) ViewEphemeralPhoto(c *gin.Context) {
	// Get access key from URL
	accessKey := c.Param("accessKey")
	if accessKey == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Access key is required")
		return
	}

	// Get viewer ID from context (optional)
	var viewerID *uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		if userUUID, ok := userID.(uuid.UUID); ok {
			viewerID = &userUUID
		}
	}

	// Get IP address and user agent
	ipAddress := ephemeral_photo.GetClientIP(c.Request)
	userAgent := ephemeral_photo.GetUserAgent(c.Request)

	// Create request
	req := &ephemeral_photo.ViewEphemeralPhotoRequest{
		AccessKey: accessKey,
		ViewerID:  viewerID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Execute use case
	response, err := h.viewUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to view ephemeral photo", err)
		if err.Error() == "photo cannot be viewed: expired" || err.Error() == "photo cannot be viewed: viewed" {
			utils.ErrorResponse(c, http.StatusGone, "Photo is no longer available")
		} else if err.Error() == "ephemeral photo not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Photo not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to view ephemeral photo")
		}
		return
	}

	// Add security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none'")
	
	// Add cache control headers to prevent caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	utils.SuccessResponse(c, http.StatusOK, response)
}

// DeleteEphemeralPhoto handles deleting an ephemeral photo
func (h *EphemeralPhotoHandler) DeleteEphemeralPhoto(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Get photo ID from URL
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid photo ID")
		return
	}

	// Create request
	req := &ephemeral_photo.DeleteEphemeralPhotoRequest{
		UserID:  userUUID,
		PhotoID: photoID,
	}

	// Execute use case
	response, err := h.deleteUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to delete ephemeral photo", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete ephemeral photo")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetEphemeralPhotoStatus handles getting ephemeral photo status
func (h *EphemeralPhotoHandler) GetEphemeralPhotoStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Get photo ID from URL
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid photo ID")
		return
	}

	// Create request
	req := &ephemeral_photo.GetEphemeralPhotoStatusRequest{
		UserID:  userUUID,
		PhotoID: photoID,
	}

	// Execute use case
	response, err := h.getStatusUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get ephemeral photo status", err)
		if err.Error() == "user does not own this photo" {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied")
		} else if err.Error() == "ephemeral photo not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Photo not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get photo status")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// ExpireEphemeralPhoto handles expiring an ephemeral photo
func (h *EphemeralPhotoHandler) ExpireEphemeralPhoto(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Get photo ID from URL
	photoIDStr := c.Param("id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid photo ID")
		return
	}

	// Create request
	req := &ephemeral_photo.ExpireEphemeralPhotoRequest{
		UserID:  userUUID,
		PhotoID: photoID,
	}

	// Execute use case
	response, err := h.expireUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to expire ephemeral photo", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to expire ephemeral photo")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetUserEphemeralPhotos handles getting user's ephemeral photos
func (h *EphemeralPhotoHandler) GetUserEphemeralPhotos(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Parse query parameters
	activeOnlyStr := c.DefaultQuery("active_only", "false")
	activeOnly := activeOnlyStr == "true"

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Create request
	req := &ephemeral_photo.GetUserEphemeralPhotosRequest{
		UserID:    userUUID,
		ActiveOnly: activeOnly,
		Limit:     limit,
		Offset:     offset,
	}

	// Execute use case
	response, err := h.getUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get user ephemeral photos", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ephemeral photos")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// TrackPhotoView handles tracking photo view duration
func (h *EphemeralPhotoHandler) TrackPhotoView(c *gin.Context) {
	// Get user ID from context (optional)
	var userID uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		if userUUID, ok := uid.(uuid.UUID); ok {
			userID = userUUID
		}
	}

	// Parse request body
	var req struct {
		PhotoID   uuid.UUID `json:"photo_id" binding:"required"`
		Duration  int       `json:"duration" binding:"required"`
		ViewerID  *uuid.UUID `json:"viewer_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get IP address and user agent
	ipAddress := ephemeral_photo.GetClientIP(c.Request)
	userAgent := ephemeral_photo.GetUserAgent(c.Request)

	// This would typically be handled by the service layer
	// For now, just return success
	logger.Info("Photo view tracked", map[string]interface{}{
		"photo_id":   req.PhotoID,
		"user_id":    userID,
		"viewer_id":  req.ViewerID,
		"duration":   req.Duration,
		"ip_address": ipAddress,
		"user_agent":  userAgent,
	})

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"success": true,
		"message": "View tracked successfully",
	})
}

// GetEphemeralPhotoAnalytics handles getting analytics for ephemeral photos
func (h *EphemeralPhotoHandler) GetEphemeralPhotoAnalytics(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Parse query parameters
	period := c.DefaultQuery("period", "week") // day, week, month, all

	// This would typically call analytics service
	// For now, return mock data
	analytics := gin.H{
		"period": period,
		"stats": gin.H{
			"total_photos":     0,
			"active_photos":    0,
			"viewed_photos":    0,
			"expired_photos":   0,
			"total_views":      0,
			"average_view_time": 0,
		},
		"generated_at": time.Now().Format(time.RFC3339),
	}

	utils.SuccessResponse(c, http.StatusOK, analytics)
}