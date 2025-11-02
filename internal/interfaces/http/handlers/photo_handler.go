package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/photo"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// PhotoHandler handles photo-related HTTP requests
type PhotoHandler struct {
	uploadPhotoUseCase        *photo.UploadPhotoUseCase
	deletePhotoUseCase        *photo.DeletePhotoUseCase
	getUploadURLUseCase      *photo.GetUploadURLUseCase
	getDownloadURLUseCase     *photo.GetDownloadURLUseCase
	setPrimaryPhotoUseCase    *photo.SetPrimaryPhotoUseCase
	markPhotoViewedUseCase   *photo.MarkPhotoViewedUseCase
	jwtUtils                 *utils.JWTUtils
}

// NewPhotoHandler creates a new photo handler
func NewPhotoHandler(
	uploadPhotoUseCase *photo.UploadPhotoUseCase,
	deletePhotoUseCase *photo.DeletePhotoUseCase,
	getUploadURLUseCase *photo.GetUploadURLUseCase,
	getDownloadURLUseCase *photo.GetDownloadURLUseCase,
	setPrimaryPhotoUseCase *photo.SetPrimaryPhotoUseCase,
	markPhotoViewedUseCase *photo.MarkPhotoViewedUseCase,
	jwtUtils *utils.JWTUtils,
) *PhotoHandler {
	return &PhotoHandler{
		uploadPhotoUseCase:        uploadPhotoUseCase,
		deletePhotoUseCase:        deletePhotoUseCase,
		getUploadURLUseCase:      getUploadURLUseCase,
		getDownloadURLUseCase:     getDownloadURLUseCase,
		setPrimaryPhotoUseCase:    setPrimaryPhotoUseCase,
		markPhotoViewedUseCase:   markPhotoViewedUseCase,
		jwtUtils:                 jwtUtils,
	}
}

// UploadPhoto handles photo upload
// @Summary Upload a photo
// @Description Upload a new photo for the authenticated user
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param file formData file true "Image file to upload"
// @Param file_name formData string true "Original filename"
// @Param is_primary formData bool false "Set as primary photo"
// @Success 200 {object} utils.SuccessResponse{data=photo.UploadPhotoResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /me/photos [post]
func (h *PhotoHandler) UploadPhoto(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "file_required", "File is required")
		return
	}

	defer file.Close()

	// Get form values
	isPrimaryStr := c.PostForm("is_primary")
	isPrimary, _ := strconv.ParseBool(isPrimaryStr)

	// Create upload request
	req := &photo.UploadPhotoRequest{
		UserID:     userUUID,
		File:        file,
		FileName:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
		IsPrimary:    isPrimary,
	}

	// Execute use case
	result, err := h.uploadPhotoUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Photo upload failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "upload_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "photo_uploaded", result)
}

// DeletePhoto handles photo deletion
// @Summary Delete a photo
// @Description Delete a photo for the authenticated user
// @Tags photos
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param photo_id path string true "Photo ID to delete"
// @Success 200 {object} utils.SuccessResponse{data=photo.DeletePhotoResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /me/photos/{photo_id} [delete]
func (h *PhotoHandler) DeletePhoto(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get photo ID from URL parameter
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_photo_id", "Invalid photo ID")
		return
	}

	// Create delete request
	req := &photo.DeletePhotoRequest{
		UserID:  userUUID,
		PhotoID: photoID,
	}

	// Execute use case
	result, err := h.deletePhotoUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Photo deletion failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "photo_deleted", result)
}

// GetUploadURL handles getting upload URL
// @Summary Get upload URL
// @Description Get a presigned URL for uploading a photo
// @Tags photos
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body photo.GetUploadURLRequest true "Upload URL request"
// @Success 200 {object} utils.SuccessResponse{data=photo.GetUploadURLResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /media/request-upload [get]
func (h *PhotoHandler) GetUploadURL(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Parse request body
	var req photo.GetUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	req.UserID = userUUID

	// Execute use case
	result, err := h.getUploadURLUseCase.Execute(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Get upload URL failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "upload_url_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "upload_url_generated", result)
}

// GetDownloadURL handles getting download URL
// @Summary Get download URL
// @Description Get a presigned URL for downloading a photo
// @Tags photos
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param photo_id path string true "Photo ID to download"
// @Param viewer_id query string false "ID of viewer (for analytics)"
// @Success 200 {object} utils.SuccessResponse{data=photo.GetDownloadURLResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /media/{photo_id}/url [get]
func (h *PhotoHandler) GetDownloadURL(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get photo ID from URL parameter
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_photo_id", "Invalid photo ID")
		return
	}

	// Get optional viewer ID from query parameter
	viewerIDStr := c.Query("viewer_id")
	var viewerID uuid.UUID
	if viewerIDStr != "" {
		viewerID, err = uuid.Parse(viewerIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid_viewer_id", "Invalid viewer ID")
			return
		}
	}

	// Create download request
	req := &photo.GetDownloadURLRequest{
		UserID:   userUUID,
		PhotoID:  photoID,
		ViewerID: viewerID,
	}

	// Execute use case
	result, err := h.getDownloadURLUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Get download URL failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "download_url_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "download_url_generated", result)
}

// MarkPhotoViewed handles marking a photo as viewed
// @Summary Mark photo as viewed
// @Description Mark a photo as viewed by a user (for analytics)
// @Tags photos
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body photo.MarkPhotoViewedRequest true "Mark viewed request"
// @Success 200 {object} utils.SuccessResponse{data=photo.MarkPhotoViewedResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /media/{photo_id}/view [post]
func (h *PhotoHandler) MarkPhotoViewed(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get photo ID from URL parameter
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_photo_id", "Invalid photo ID")
		return
	}

	// Parse request body
	var req photo.MarkPhotoViewedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	req.UserID = userUUID
	req.PhotoID = photoID

	// Execute use case
	result, err := h.markPhotoViewedUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Mark photo viewed failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "mark_viewed_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "photo_marked_viewed", result)
}

// SetPrimaryPhoto handles setting a photo as primary
// @Summary Set primary photo
// @Description Set a photo as the primary photo for the authenticated user
// @Tags photos
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param photo_id path string true "Photo ID to set as primary"
// @Success 200 {object} utils.SuccessResponse{data=photo.SetPrimaryPhotoResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Router /me/photos/{photo_id}/set-primary [put]
func (h *PhotoHandler) SetPrimaryPhoto(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get photo ID from URL parameter
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_photo_id", "Invalid photo ID")
		return
	}

	// Create set primary request
	req := &photo.SetPrimaryPhotoRequest{
		UserID:  userUUID,
		PhotoID: photoID,
	}

	// Execute use case
	result, err := h.setPrimaryPhotoUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		logger.Error("Set primary photo failed", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "set_primary_failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "primary_photo_set", result)
}