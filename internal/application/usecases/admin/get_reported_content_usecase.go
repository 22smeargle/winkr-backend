package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetReportedContentUseCase handles retrieving reported content
type GetReportedContentUseCase struct {
	reportRepo repositories.ReportRepository
	photoRepo  repositories.PhotoRepository
	messageRepo repositories.MessageRepository
}

// NewGetReportedContentUseCase creates a new GetReportedContentUseCase
func NewGetReportedContentUseCase(
	reportRepo repositories.ReportRepository,
	photoRepo repositories.PhotoRepository,
	messageRepo repositories.MessageRepository,
) *GetReportedContentUseCase {
	return &GetReportedContentUseCase{
		reportRepo:  reportRepo,
		photoRepo:   photoRepo,
		messageRepo: messageRepo,
	}
}

// GetReportedContentRequest represents a request to get reported content
type GetReportedContentRequest struct {
	AdminID     uuid.UUID `json:"admin_id" validate:"required"`
	ContentType string    `json:"content_type" validate:"required,oneof=photo message"`
	Status      string    `json:"status" validate:"omitempty,oneof=pending reviewed resolved dismissed"`
	Limit       int       `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset      int       `json:"offset" validate:"omitempty,min=0"`
	SortBy      string    `json:"sort_by" validate:"omitempty,oneof=created_at updated_at report_count"`
	SortOrder   string    `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// ReportedContentResponse represents response from getting reported content
type ReportedContentResponse struct {
	Content    []ReportedContentItem `json:"content"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	TotalPages int                   `json:"total_pages"`
	Timestamp  time.Time             `json:"timestamp"`
}

// ReportedContentItem represents a reported content item
type ReportedContentItem struct {
	ID           uuid.UUID            `json:"id"`
	ContentType  string               `json:"content_type"`
	ContentID    uuid.UUID            `json:"content_id"`
	Status       string               `json:"status"`
	ReportCount  int                  `json:"report_count"`
	Reports      []ReportSummary      `json:"reports"`
	Content      ContentDetails       `json:"content"`
	Reporter     UserSummary          `json:"reporter"`
	ReportedUser UserSummary          `json:"reported_user"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	ReviewedBy   *uuid.UUID           `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time           `json:"reviewed_at,omitempty"`
	Resolution   *string              `json:"resolution,omitempty"`
}

// ReportSummary represents a summary of a report
type ReportSummary struct {
	ID        uuid.UUID `json:"id"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ContentDetails represents details of the reported content
type ContentDetails struct {
	URL         string     `json:"url,omitempty"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	Message     string     `json:"message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UserSummary represents a summary of a user
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar,omitempty"`
}

// Execute retrieves reported content
func (uc *GetReportedContentUseCase) Execute(ctx context.Context, req GetReportedContentRequest) (*ReportedContentResponse, error) {
	logger.Info("GetReportedContent use case executed", "admin_id", req.AdminID, "content_type", req.ContentType, "status", req.Status)

	// Set default values
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Get reported content
	content, total, err := uc.getReportedContent(ctx, req)
	if err != nil {
		logger.Error("Failed to get reported content", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get reported content: %w", err)
	}

	// Calculate pagination
	page := req.Offset/req.Limit + 1
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	logger.Info("GetReportedContent use case completed successfully", "admin_id", req.AdminID, "total", total)
	return &ReportedContentResponse{
		Content:    content,
		Total:      total,
		Page:       page,
		PerPage:    req.Limit,
		TotalPages: totalPages,
		Timestamp:  time.Now(),
	}, nil
}

// getReportedContent retrieves reported content from the database
func (uc *GetReportedContentUseCase) getReportedContent(ctx context.Context, req GetReportedContentRequest) ([]ReportedContentItem, int64, error) {
	// Mock data - in real implementation, this would query the database
	content := []ReportedContentItem{
		{
			ID:          uuid.New(),
			ContentType: "photo",
			ContentID:   uuid.New(),
			Status:      "pending",
			ReportCount: 3,
			Reports: []ReportSummary{
				{
					ID:        uuid.New(),
					Reason:    "inappropriate_content",
					Message:   "This photo contains inappropriate content",
					Status:    "pending",
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
				{
					ID:        uuid.New(),
					Reason:    "spam",
					Message:   "This appears to be spam content",
					Status:    "pending",
					CreatedAt: time.Now().Add(-4 * time.Hour),
				},
				{
					ID:        uuid.New(),
					Reason:    "fake_profile",
					Message:   "This appears to be a fake profile photo",
					Status:    "pending",
					CreatedAt: time.Now().Add(-6 * time.Hour),
				},
			},
			Content: ContentDetails{
				URL:       "https://example.com/photos/12345.jpg",
				Thumbnail: "https://example.com/photos/12345_thumb.jpg",
				Metadata: map[string]interface{}{
					"width":  800,
					"height": 600,
					"size":   245760,
				},
			},
			Reporter: UserSummary{
				ID:       uuid.New(),
				Username: "reporter1",
				Name:     "Reporter One",
				Avatar:   "https://example.com/avatars/reporter1.jpg",
			},
			ReportedUser: UserSummary{
				ID:       uuid.New(),
				Username: "reported_user1",
				Name:     "Reported User One",
				Avatar:   "https://example.com/avatars/reported_user1.jpg",
			},
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          uuid.New(),
			ContentType: "message",
			ContentID:   uuid.New(),
			Status:      "reviewed",
			ReportCount: 2,
			Reports: []ReportSummary{
				{
					ID:        uuid.New(),
					Reason:    "harassment",
					Message:   "This message contains harassing content",
					Status:    "resolved",
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				{
					ID:        uuid.New(),
					Reason:    "harassment",
					Message:   "This message contains harassing content",
					Status:    "resolved",
					CreatedAt: time.Now().Add(-3 * time.Hour),
				},
			},
			Content: ContentDetails{
				Message: "This is an inappropriate message that has been reported for harassment",
				Metadata: map[string]interface{}{
					"chat_id": uuid.New().String(),
					"length":  67,
				},
			},
			Reporter: UserSummary{
				ID:       uuid.New(),
				Username: "reporter2",
				Name:     "Reporter Two",
				Avatar:   "https://example.com/avatars/reporter2.jpg",
			},
			ReportedUser: UserSummary{
				ID:       uuid.New(),
				Username: "reported_user2",
				Name:     "Reported User Two",
				Avatar:   "https://example.com/avatars/reported_user2.jpg",
			},
			CreatedAt:  time.Now().Add(-3 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
			ReviewedBy: &uuid.UUID{},
			ReviewedAt: &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			Resolution: &[]string{"Content removed"}[0],
		},
		{
			ID:          uuid.New(),
			ContentType: "photo",
			ContentID:   uuid.New(),
			Status:      "resolved",
			ReportCount: 1,
			Reports: []ReportSummary{
				{
					ID:        uuid.New(),
					Reason:    "copyright",
					Message:   "This photo appears to be copyrighted material",
					Status:    "resolved",
					CreatedAt: time.Now().Add(-24 * time.Hour),
				},
			},
			Content: ContentDetails{
				URL:       "https://example.com/photos/67890.jpg",
				Thumbnail: "https://example.com/photos/67890_thumb.jpg",
				Metadata: map[string]interface{}{
					"width":  1024,
					"height": 768,
					"size":   512000,
				},
			},
			Reporter: UserSummary{
				ID:       uuid.New(),
				Username: "reporter3",
				Name:     "Reporter Three",
				Avatar:   "https://example.com/avatars/reporter3.jpg",
			},
			ReportedUser: UserSummary{
				ID:       uuid.New(),
				Username: "reported_user3",
				Name:     "Reported User Three",
				Avatar:   "https://example.com/avatars/reported_user3.jpg",
			},
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now().Add(-12 * time.Hour),
			ReviewedBy: &uuid.UUID{},
			ReviewedAt: &[]time.Time{time.Now().Add(-12 * time.Hour)}[0],
			Resolution: &[]string{"Content removed and user warned"}[0],
		},
	}

	// Filter content based on request parameters
	filteredContent := uc.filterContent(content, req)

	// Apply pagination
	start := req.Offset
	end := start + req.Limit
	if end > len(filteredContent) {
		end = len(filteredContent)
	}

	if start >= len(filteredContent) {
		return []ReportedContentItem{}, int64(len(filteredContent)), nil
	}

	return filteredContent[start:end], int64(len(filteredContent)), nil
}

// filterContent filters content based on request parameters
func (uc *GetReportedContentUseCase) filterContent(content []ReportedContentItem, req GetReportedContentRequest) []ReportedContentItem {
	var filtered []ReportedContentItem

	for _, item := range content {
		// Filter by content type
		if item.ContentType != req.ContentType {
			continue
		}

		// Filter by status
		if req.Status != "" && item.Status != req.Status {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}