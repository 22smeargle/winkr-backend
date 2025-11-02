package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ModerateContentUseCase handles content moderation actions
type ModerateContentUseCase struct {
	reportRepo  repositories.ReportRepository
	photoRepo   repositories.PhotoRepository
	messageRepo repositories.MessageRepository
	userRepo    repositories.UserRepository
}

// NewModerateContentUseCase creates a new ModerateContentUseCase
func NewModerateContentUseCase(
	reportRepo repositories.ReportRepository,
	photoRepo repositories.PhotoRepository,
	messageRepo repositories.MessageRepository,
	userRepo repositories.UserRepository,
) *ModerateContentUseCase {
	return &ModerateContentUseCase{
		reportRepo:  reportRepo,
		photoRepo:   photoRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// ModerateContentRequest represents a request to moderate content
type ModerateContentRequest struct {
	AdminID     uuid.UUID `json:"admin_id" validate:"required"`
	ContentType string    `json:"content_type" validate:"required,oneof=photo message"`
	ContentID   uuid.UUID `json:"content_id" validate:"required"`
	Action      string    `json:"action" validate:"required,oneof=approve reject delete"`
	Reason      string    `json:"reason"`
	Notes       string    `json:"notes"`
	WarnUser    bool      `json:"warn_user"`
	BanUser     bool      `json:"ban_user"`
	BanDuration int       `json:"ban_duration"` // in days, 0 for permanent
}

// ModerateContentResponse represents response from moderating content
type ModerateContentResponse struct {
	ContentID   uuid.UUID `json:"content_id"`
	ContentType string    `json:"content_type"`
	Action      string    `json:"action"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	UserAction  *UserAction `json:"user_action,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// UserAction represents actions taken against a user
type UserAction struct {
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"` // warned, banned, suspended
	Duration  int       `json:"duration"` // in days, 0 for permanent
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// Execute moderates content
func (uc *ModerateContentUseCase) Execute(ctx context.Context, req ModerateContentRequest) (*ModerateContentResponse, error) {
	logger.Info("ModerateContent use case executed", "admin_id", req.AdminID, "content_type", req.ContentType, "content_id", req.ContentID, "action", req.Action)

	// Validate request
	if err := uc.validateRequest(req); err != nil {
		logger.Error("Invalid moderation request", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("invalid moderation request: %w", err)
	}

	// Get content details
	content, err := uc.getContentDetails(ctx, req.ContentType, req.ContentID)
	if err != nil {
		logger.Error("Failed to get content details", err, "admin_id", req.AdminID, "content_id", req.ContentID)
		return nil, fmt.Errorf("failed to get content details: %w", err)
	}

	// Perform moderation action
	var userAction *UserAction
	switch req.Action {
	case "approve":
		err = uc.approveContent(ctx, req.ContentType, req.ContentID, req.AdminID)
	case "reject":
		err = uc.rejectContent(ctx, req.ContentType, req.ContentID, req.AdminID, req.Reason, req.Notes)
	case "delete":
		err = uc.deleteContent(ctx, req.ContentType, req.ContentID, req.AdminID, req.Reason, req.Notes)
	}

	if err != nil {
		logger.Error("Failed to perform moderation action", err, "admin_id", req.AdminID, "action", req.Action)
		return nil, fmt.Errorf("failed to perform moderation action: %w", err)
	}

	// Handle user actions if requested
	if req.WarnUser || req.BanUser {
		userAction, err = uc.handleUserAction(ctx, content.UserID, req)
		if err != nil {
			logger.Error("Failed to handle user action", err, "admin_id", req.AdminID, "user_id", content.UserID)
			return nil, fmt.Errorf("failed to handle user action: %w", err)
		}
	}

	logger.Info("ModerateContent use case completed successfully", "admin_id", req.AdminID, "content_id", req.ContentID, "action", req.Action)
	return &ModerateContentResponse{
		ContentID:   req.ContentID,
		ContentType: req.ContentType,
		Action:      req.Action,
		Status:      "completed",
		Message:     fmt.Sprintf("Content %s successfully", req.Action),
		UserAction:  userAction,
		Timestamp:   time.Now(),
	}, nil
}

// validateRequest validates the moderation request
func (uc *ModerateContentUseCase) validateRequest(req ModerateContentRequest) error {
	if req.Action == "reject" && req.Reason == "" {
		return fmt.Errorf("reason is required for rejection")
	}

	if req.Action == "delete" && req.Reason == "" {
		return fmt.Errorf("reason is required for deletion")
	}

	if req.BanUser && req.BanDuration < 0 {
		return fmt.Errorf("ban duration cannot be negative")
	}

	return nil
}

// ContentDetails represents details of content for moderation
type ContentDetails struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Status string    `json:"status"`
}

// getContentDetails retrieves content details
func (uc *ModerateContentUseCase) getContentDetails(ctx context.Context, contentType string, contentID uuid.UUID) (*ContentDetails, error) {
	// Mock implementation - in real implementation, this would query the appropriate repository
	return &ContentDetails{
		ID:     contentID,
		UserID: uuid.New(), // Mock user ID
		Status: "pending",
	}, nil
}

// approveContent approves content
func (uc *ModerateContentUseCase) approveContent(ctx context.Context, contentType string, contentID uuid.UUID, adminID uuid.UUID) error {
	// Mock implementation - in real implementation, this would update the content status
	logger.Info("Content approved", "content_type", contentType, "content_id", contentID, "admin_id", adminID)
	return nil
}

// rejectContent rejects content
func (uc *ModerateContentUseCase) rejectContent(ctx context.Context, contentType string, contentID uuid.UUID, adminID uuid.UUID, reason, notes string) error {
	// Mock implementation - in real implementation, this would update the content status and add notes
	logger.Info("Content rejected", "content_type", contentType, "content_id", contentID, "admin_id", adminID, "reason", reason, "notes", notes)
	return nil
}

// deleteContent deletes content
func (uc *ModerateContentUseCase) deleteContent(ctx context.Context, contentType string, contentID uuid.UUID, adminID uuid.UUID, reason, notes string) error {
	// Mock implementation - in real implementation, this would delete the content
	logger.Info("Content deleted", "content_type", contentType, "content_id", contentID, "admin_id", adminID, "reason", reason, "notes", notes)
	return nil
}

// handleUserAction handles actions against the user who created the content
func (uc *ModerateContentUseCase) handleUserAction(ctx context.Context, userID uuid.UUID, req ModerateContentRequest) (*UserAction, error) {
	var userAction *UserAction

	if req.WarnUser {
		// Mock implementation - in real implementation, this would send a warning to the user
		logger.Info("User warned", "user_id", userID, "admin_id", req.AdminID, "reason", req.Reason)
		userAction = &UserAction{
			UserID:    userID,
			Action:    "warned",
			Duration:  0,
			Reason:    req.Reason,
			Timestamp: time.Now(),
		}
	}

	if req.BanUser {
		// Mock implementation - in real implementation, this would ban the user
		logger.Info("User banned", "user_id", userID, "admin_id", req.AdminID, "duration", req.BanDuration, "reason", req.Reason)
		userAction = &UserAction{
			UserID:    userID,
			Action:    "banned",
			Duration:  req.BanDuration,
			Reason:    req.Reason,
			Timestamp: time.Now(),
		}
	}

	return userAction, nil
}