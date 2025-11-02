package ephemeral_photo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// ViewEphemeralPhotoRequest represents the request for viewing an ephemeral photo
type ViewEphemeralPhotoRequest struct {
	AccessKey string    `json:"access_key" validate:"required"`
	ViewerID  *uuid.UUID `json:"viewer_id,omitempty"` // Optional: ID of the viewer if authenticated
	IPAddress string    `json:"ip_address" validate:"required"`
	UserAgent string    `json:"user_agent" validate:"required"`
}

// ViewEphemeralPhotoResponse represents the response for viewing an ephemeral photo
type ViewEphemeralPhotoResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	FileURL       string    `json:"file_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	ViewCount     int       `json:"view_count"`
	MaxViews      int       `json:"max_views"`
	IsViewed      bool      `json:"is_viewed"`
	ExpiresAt     string    `json:"expires_at"`
	RemainingTime int       `json:"remaining_time_seconds"` // Time remaining until expiration
	ViewStatus    string    `json:"view_status"`
	// Security fields
	PreventDownload bool   `json:"prevent_download"`
	WatermarkURL   string `json:"watermark_url,omitempty"` // URL to watermarked version
	// Tracking fields
	ViewID        uuid.UUID `json:"view_id"` // ID of the view record for tracking
	ViewStartTime  int64     `json:"view_start_time"` // Unix timestamp for view tracking
}

// ViewEphemeralPhotoUseCase handles viewing ephemeral photos
type ViewEphemeralPhotoUseCase struct {
	ephemeralPhotoService services.EphemeralPhotoService
	validator           validator.Validator
}

// NewViewEphemeralPhotoUseCase creates a new view ephemeral photo use case
func NewViewEphemeralPhotoUseCase(
	ephemeralPhotoService services.EphemeralPhotoService,
	validator validator.Validator,
) *ViewEphemeralPhotoUseCase {
	return &ViewEphemeralPhotoUseCase{
		ephemeralPhotoService: ephemeralPhotoService,
		validator:           validator,
	}
}

// Execute executes the view ephemeral photo use case
func (uc *ViewEphemeralPhotoUseCase) Execute(ctx context.Context, req *ViewEphemeralPhotoRequest) (*ViewEphemeralPhotoResponse, error) {
	// Validate request
	if err := uc.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Extract IP address from request if not provided
	ipAddress := req.IPAddress
	if ipAddress == "" {
		ipAddress = uc.extractIPAddressFromContext(ctx)
	}

	// Extract user agent from request if not provided
	userAgent := req.UserAgent
	if userAgent == "" {
		userAgent = uc.extractUserAgentFromContext(ctx)
	}

	// View the ephemeral photo
	photo, err := uc.ephemeralPhotoService.ViewEphemeralPhoto(
		ctx,
		req.AccessKey,
		req.ViewerID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to view ephemeral photo: %w", err)
	}

	// Calculate remaining time
	remainingTime := int(photo.GetRemainingTime().Seconds())
	if remainingTime < 0 {
		remainingTime = 0
	}

	// Create response
	response := &ViewEphemeralPhotoResponse{
		ID:             photo.ID,
		UserID:         photo.UserID,
		FileURL:        photo.FileURL,
		ThumbnailURL:   photo.ThumbnailURL,
		ViewCount:       photo.ViewCount,
		MaxViews:        photo.MaxViews,
		IsViewed:       photo.IsViewed,
		ExpiresAt:       photo.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		RemainingTime:   remainingTime,
		ViewStatus:      photo.GetViewStatus(),
		PreventDownload: true, // Always prevent download for security
		ViewID:          uuid.New(), // Generate view ID for tracking
		ViewStartTime:    photo.CreatedAt.Unix(),
	}

	// Add watermark URL if photo is viewed (for security)
	if photo.IsViewed {
		response.WatermarkURL = photo.FileURL + "?watermark=true"
	}

	return response, nil
}

// extractIPAddressFromContext extracts IP address from context
func (uc *ViewEphemeralPhotoUseCase) extractIPAddressFromContext(ctx context.Context) string {
	// This would typically extract from HTTP request context
	// For now, return a default value
	return "127.0.0.1"
}

// extractUserAgentFromContext extracts user agent from context
func (uc *ViewEphemeralPhotoUseCase) extractUserAgentFromContext(ctx context.Context) string {
	// This would typically extract from HTTP request context
	// For now, return a default value
	return "Unknown"
}

// GetClientIP extracts the real client IP address from HTTP request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// GetUserAgent extracts user agent from HTTP request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}