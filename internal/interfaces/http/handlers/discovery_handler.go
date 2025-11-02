package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/matching"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// DiscoveryHandler handles discovery and matching endpoints
type DiscoveryHandler struct {
	discoverUsersUseCase   *matching.DiscoverUsersUseCase
	likeUserUseCase       *matching.LikeUserUseCase
	dislikeUserUseCase    *matching.DislikeUserUseCase
	superLikeUserUseCase   *matching.SuperLikeUserUseCase
	getMatchesUseCase      *matching.GetMatchesUseCase
	getDiscoveryStatsUseCase *matching.GetDiscoveryStatsUseCase
}

// NewDiscoveryHandler creates a new DiscoveryHandler
func NewDiscoveryHandler(
	discoverUsersUseCase *matching.DiscoverUsersUseCase,
	likeUserUseCase *matching.LikeUserUseCase,
	dislikeUserUseCase *matching.DislikeUserUseCase,
	superLikeUserUseCase *matching.SuperLikeUserUseCase,
	getMatchesUseCase *matching.GetMatchesUseCase,
	getDiscoveryStatsUseCase *matching.GetDiscoveryStatsUseCase,
) *DiscoveryHandler {
	return &DiscoveryHandler{
		discoverUsersUseCase:   discoverUsersUseCase,
		likeUserUseCase:       likeUserUseCase,
		dislikeUserUseCase:    dislikeUserUseCase,
		superLikeUserUseCase:   superLikeUserUseCase,
		getMatchesUseCase:      getMatchesUseCase,
		getDiscoveryStatsUseCase: getDiscoveryStatsUseCase,
	}
}

// DiscoverUsers handles GET /discover
// @Summary Discover users for matching
// @Description Get a list of potential matches based on user preferences and location
// @Tags discovery
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Number of results to return" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of results to skip" default(0) minimum(0)
// @Param age_min query int false "Minimum age filter"
// @Param age_max query int false "Maximum age filter"
// @Param max_distance query int false "Maximum distance in kilometers"
// @Param gender query string false "Gender filter"
// @Param verified query bool false "Filter by verification status"
// @Param has_photos query bool false "Filter by users with photos"
// @Success 200 {object} dto.DiscoverUsersResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/discover [get]
func (h *DiscoveryHandler) DiscoverUsers(c *gin.Context) {
	// Get user ID from context (from JWT middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse query parameters
	req := &matching.DiscoverUsersRequest{
		UserID: userID,
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	// Parse age filters
	if ageMinStr := c.Query("age_min"); ageMinStr != "" {
		if ageMin, err := strconv.Atoi(ageMinStr); err == nil {
			req.AgeMin = &ageMin
		}
	}

	if ageMaxStr := c.Query("age_max"); ageMaxStr != "" {
		if ageMax, err := strconv.Atoi(ageMaxStr); err == nil {
			req.AgeMax = &ageMax
		}
	}

	// Parse distance filter
	if maxDistanceStr := c.Query("max_distance"); maxDistanceStr != "" {
		if maxDistance, err := strconv.Atoi(maxDistanceStr); err == nil {
			req.MaxDistance = &maxDistance
		}
	}

	// Parse gender filter
	if genderStr := c.Query("gender"); genderStr != "" {
		req.Gender = &genderStr
	}

	// Parse verified filter
	if verifiedStr := c.Query("verified"); verifiedStr != "" {
		if verified, err := strconv.ParseBool(verifiedStr); err == nil {
			req.Verified = &verified
		}
	}

	// Parse has photos filter
	if hasPhotosStr := c.Query("has_photos"); hasPhotosStr != "" {
		if hasPhotos, err := strconv.ParseBool(hasPhotosStr); err == nil {
			req.HasPhotos = &hasPhotos
		}
	}

	// Execute use case
	response, err := h.discoverUsersUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// LikeUser handles POST /like/:id
// @Summary Like a user
// @Description Like a user and check for mutual match
// @Tags discovery
// @Accept json
// @Produce json
// @Param id path string true "User ID to like"
// @Success 200 {object} matching.LikeUserResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/like/{id} [post]
func (h *DiscoveryHandler) LikeUser(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	swiperID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get swiped user ID from path
	swipedIDStr := c.Param("id")
	swipedID, err := uuid.Parse(swipedIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID to like")
		return
	}

	// Create request
	req := &matching.LikeUserRequest{
		SwiperID: swiperID,
		SwipedID: swipedID,
	}

	// Execute use case
	response, err := h.likeUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user already swiped" {
			utils.ErrorResponse(c, http.StatusConflict, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// DislikeUser handles POST /dislike/:id
// @Summary Dislike a user
// @Description Dislike/skip a user
// @Tags discovery
// @Accept json
// @Produce json
// @Param id path string true "User ID to dislike"
// @Success 200 {object} matching.DislikeUserResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/dislike/{id} [post]
func (h *DiscoveryHandler) DislikeUser(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	swiperID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get swiped user ID from path
	swipedIDStr := c.Param("id")
	swipedID, err := uuid.Parse(swipedIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID to dislike")
		return
	}

	// Create request
	req := &matching.DislikeUserRequest{
		SwiperID: swiperID,
		SwipedID: swipedID,
	}

	// Execute use case
	response, err := h.dislikeUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user already swiped" {
			utils.ErrorResponse(c, http.StatusConflict, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// SuperLikeUser handles POST /superlike/:id
// @Summary Super like a user
// @Description Super like a user (premium feature) and check for mutual match
// @Tags discovery
// @Accept json
// @Produce json
// @Param id path string true "User ID to super like"
// @Success 200 {object} matching.SuperLikeUserResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 402 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/superlike/{id} [post]
func (h *DiscoveryHandler) SuperLikeUser(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	swiperID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get swiped user ID from path
	swipedIDStr := c.Param("id")
	swipedID, err := uuid.Parse(swipedIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID to super like")
		return
	}

	// Create request
	req := &matching.SuperLikeUserRequest{
		SwiperID: swiperID,
		SwipedID: swipedID,
	}

	// Execute use case
	response, err := h.superLikeUserUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user already swiped" {
			utils.ErrorResponse(c, http.StatusConflict, err.Error())
			return
		}
		if err.Error() == "super like requires premium subscription" {
			utils.ErrorResponse(c, http.StatusPaymentRequired, err.Error())
			return
		}
		if err.Error() == "daily super like limit exceeded" {
			utils.ErrorResponse(c, http.StatusTooManyRequests, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetMatches handles GET /matches
// @Summary Get user's matches
// @Description Get a list of user's mutual matches
// @Tags discovery
// @Accept json
// @Produce json
// @Param limit query int false "Number of results to return" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of results to skip" default(0) minimum(0)
// @Param unread_only query bool false "Only return unread matches"
// @Success 200 {object} dto.GetMatchesResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/matches [get]
func (h *DiscoveryHandler) GetMatches(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse query parameters
	req := &matching.GetMatchesRequest{
		UserID: userID,
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	// Parse unread only filter
	if unreadOnlyStr := c.Query("unread_only"); unreadOnlyStr != "" {
		if unreadOnly, err := strconv.ParseBool(unreadOnlyStr); err == nil {
			req.UnreadOnly = unreadOnly
		}
	}

	// Execute use case
	response, err := h.getMatchesUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetDiscoveryStats handles GET /discover/stats
// @Summary Get discovery statistics
// @Description Get user's discovery and matching statistics
// @Tags discovery
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetDiscoveryStatsResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/discover/stats [get]
func (h *DiscoveryHandler) GetDiscoveryStats(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Create request
	req := &matching.GetDiscoveryStatsRequest{
		UserID: userID,
	}

	// Execute use case
	response, err := h.getDiscoveryStatsUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}