package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/profile"
	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/utils"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// ProfileHandler handles profile HTTP requests
type ProfileHandler struct {
	getProfileUseCase       *profile.GetProfileUseCase
	updateProfileUseCase    *profile.UpdateProfileUseCase
	viewUserProfileUseCase  *profile.ViewUserProfileUseCase
	updateLocationUseCase   *profile.UpdateLocationUseCase
	getMatchesUseCase      *profile.GetMatchesUseCase
	deleteAccountUseCase   *profile.DeleteAccountUseCase
	profileValidator        *validator.ProfileValidator
	rateLimiter           *middleware.ProfileRateLimiter
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(
	getProfileUseCase *profile.GetProfileUseCase,
	updateProfileUseCase *profile.UpdateProfileUseCase,
	viewUserProfileUseCase *profile.ViewUserProfileUseCase,
	updateLocationUseCase *profile.UpdateLocationUseCase,
	getMatchesUseCase *profile.GetMatchesUseCase,
	deleteAccountUseCase *profile.DeleteAccountUseCase,
	profileValidator *validator.ProfileValidator,
	rateLimiter *middleware.ProfileRateLimiter,
) *ProfileHandler {
	return &ProfileHandler{
		getProfileUseCase:      getProfileUseCase,
		updateProfileUseCase:   updateProfileUseCase,
		viewUserProfileUseCase:  viewUserProfileUseCase,
		updateLocationUseCase:   updateLocationUseCase,
		getMatchesUseCase:      getMatchesUseCase,
		deleteAccountUseCase:   deleteAccountUseCase,
		profileValidator:        profileValidator,
		rateLimiter:           rateLimiter,
	}
}

// GetProfile handles GET /me endpoint - retrieve own profile
// @Summary Get current user profile
// @Description Get the current user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} dto.ProfileResponseDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 404 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/me [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("get-profile")(c)
	if c.IsAborted() {
		return
	}

	// Extract user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	// Execute use case
	response, err := h.getProfileUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	profileResponse := &dto.ProfileResponseDTO{
		Success: true,
		Data:    response,
	}

	utils.Success(c, http.StatusOK, profileResponse)
}

// UpdateProfile handles PUT /me endpoint - update own profile
// @Summary Update current user profile
// @Description Update the current user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.UpdateProfileRequestDTO true "Profile update request"
// @Success 200 {object} dto.ProfileResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/me [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("update-profile")(c)
	if c.IsAborted() {
		return
	}

	// Extract user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	var req dto.UpdateProfileRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate request
	if err := h.profileValidator.ValidateUpdateProfileRequest(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Convert to use case request
	useCaseReq := &profile.UpdateProfileRequest{
		UserID:      userID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Bio:          req.Bio,
		InterestedIn: req.InterestedIn,
		Preferences:   req.Preferences,
	}

	// Execute use case
	response, err := h.updateProfileUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	profileResponse := &dto.ProfileResponseDTO{
		Success: true,
		Data:    response,
	}

	utils.Success(c, http.StatusOK, profileResponse)
}

// ViewUserProfile handles GET /users/:id endpoint - view other user's profile
// @Summary View user profile
// @Description View another user's profile with privacy controls
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserProfileResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 404 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/users/{id} [get]
func (h *ProfileHandler) ViewUserProfile(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("view-profile")(c)
	if c.IsAborted() {
		return
	}

	// Extract viewer ID from context
	viewerIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	viewerID, err := uuid.Parse(viewerIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	// Extract target user ID from path
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		utils.BadRequest(c, "Invalid user ID format")
		return
	}

	// Execute use case
	response, err := h.viewUserProfileUseCase.Execute(c.Request.Context(), viewerID, targetUserID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	profileResponse := &dto.UserProfileResponseDTO{
		Success: true,
		Data:    response,
	}

	utils.Success(c, http.StatusOK, profileResponse)
}

// UpdateLocation handles PUT /me/location endpoint - update location
// @Summary Update user location
// @Description Update the current user's location
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.UpdateLocationRequestDTO true "Location update request"
// @Success 200 {object} dto.MessageResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/me/location [put]
func (h *ProfileHandler) UpdateLocation(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("update-location")(c)
	if c.IsAborted() {
		return
	}

	// Extract user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	var req dto.UpdateLocationRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithDetails(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate request
	if err := h.profileValidator.ValidateUpdateLocationRequest(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Convert to use case request
	useCaseReq := &profile.UpdateLocationRequest{
		UserID:    userID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		City:      req.City,
		Country:   req.Country,
	}

	// Execute use case
	response, err := h.updateLocationUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	messageResponse := &dto.MessageResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.Success(c, http.StatusOK, messageResponse)
}

// GetMatches handles GET /me/matches endpoint - list matches
// @Summary Get user matches
// @Description Get the current user's matches with pagination
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param limit query int false "Number of matches to return" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} dto.MatchesResponseDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/me/matches [get]
func (h *ProfileHandler) GetMatches(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("get-matches")(c)
	if c.IsAborted() {
		return
	}

	// Extract user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Execute use case
	response, err := h.getMatchesUseCase.Execute(c.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	matchesResponse := &dto.MatchesResponseDTO{
		Success: true,
		Data:    response.Matches,
		Pagination: &dto.PaginationDTO{
			Total:  response.Total,
			Limit:  limit,
			Offset: offset,
		},
	}

	utils.Success(c, http.StatusOK, matchesResponse)
}

// DeleteAccount handles DELETE /me/account endpoint - delete account
// @Summary Delete user account
// @Description Delete the current user's account and all associated data
// @Tags profile
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.DeleteAccountRequestDTO false "Account deletion request"
// @Success 200 {object} dto.MessageResponseDTO
// @Failure 400 {object} dto.ErrorDTO
// @Failure 401 {object} dto.ErrorDTO
// @Failure 500 {object} dto.ErrorDTO
// @Router /api/v1/profile/me/account [delete]
func (h *ProfileHandler) DeleteAccount(c *gin.Context) {
	// Apply rate limiting
	h.rateLimiter.RateLimit("delete-account")(c)
	if c.IsAborted() {
		return
	}

	// Extract user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.Unauthorized(c, "Invalid user ID")
		return
	}

	// Parse optional request body
	var req dto.DeleteAccountRequestDTO
	c.ShouldBindJSON(&req)

	// Validate request
	if err := h.profileValidator.ValidateDeleteAccountRequest(&req); err != nil {
		utils.ValidationError(c, err.Error())
		return
	}

	// Convert to use case request
	useCaseReq := &profile.DeleteAccountRequest{
		UserID:     userID,
		Password:    req.Password,
		Reason:      req.Reason,
		Confirm:     req.Confirm,
	}

	// Execute use case
	response, err := h.deleteAccountUseCase.Execute(c.Request.Context(), useCaseReq)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// Convert to DTO
	messageResponse := &dto.MessageResponseDTO{
		Success: true,
		Message: response.Message,
	}

	utils.Success(c, http.StatusOK, messageResponse)
}