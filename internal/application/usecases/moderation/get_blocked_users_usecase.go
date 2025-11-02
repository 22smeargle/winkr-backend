package moderation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/moderation"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetBlockedUsersRequest represents a request to get blocked users
type GetBlockedUsersRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Limit   int       `json:"limit,omitempty"`
	Offset  int       `json:"offset,omitempty"`
}

// GetBlockedUsersResponse represents the response with blocked users
type GetBlockedUsersResponse struct {
	UserID      uuid.UUID              `json:"user_id"`
	BlockedUsers []*moderation.BlockedUser `json:"blocked_users"`
	TotalCount   int64                  `json:"total_count"`
	HasMore      bool                    `json:"has_more"`
}

// GetBlockedUsersUseCase handles getting blocked users functionality
type GetBlockedUsersUseCase struct {
	blockUserUseCase *moderation.BlockUserUseCase
}

// NewGetBlockedUsersUseCase creates a new GetBlockedUsersUseCase
func NewGetBlockedUsersUseCase(blockUserUseCase *moderation.BlockUserUseCase) *GetBlockedUsersUseCase {
	return &GetBlockedUsersUseCase{
		blockUserUseCase: blockUserUseCase,
	}
}

// Execute executes the get blocked users use case
func (uc *GetBlockedUsersUseCase) Execute(ctx context.Context, req GetBlockedUsersRequest) (*GetBlockedUsersResponse, error) {
	logger.Info("Executing GetBlockedUsers use case", "user_id", req.UserID, "limit", req.Limit, "offset", req.Offset)
	
	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}
	
	// Set default offset if not provided
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	
	// Get blocked users from block use case
	blockedUsers, err := uc.blockUserUseCase.GetBlockedUsers(ctx, req.UserID, limit, offset)
	if err != nil {
		logger.Error("Failed to get blocked users", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get blocked users: %w", err)
	}
	
	// Get total count for pagination
	totalCount, err := uc.blockUserUseCase.GetBlockCount(ctx, req.UserID)
	if err != nil {
		logger.Error("Failed to get block count", err, "user_id", req.UserID)
		return nil, fmt.Errorf("failed to get block count: %w", err)
	}
	
	// Determine if there are more results
	hasMore := (offset + len(blockedUsers)) < int(totalCount)
	
	response := &GetBlockedUsersResponse{
		UserID:      req.UserID,
		BlockedUsers: blockedUsers,
		TotalCount:   totalCount,
		HasMore:      hasMore,
	}
	
	logger.Info("GetBlockedUsers use case executed successfully", "user_id", req.UserID, "blocked_count", len(blockedUsers))
	return response, nil
}

// GetUsersBlocking gets all users who are blocking a specific user
func (uc *GetBlockedUsersUseCase) GetUsersBlocking(ctx context.Context, userID uuid.UUID, limit, offset int) (*GetBlockedUsersResponse, error) {
	logger.Info("Getting users blocking user", "user_id", userID, "limit", limit, "offset", offset)
	
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20 // Default limit
	}
	
	// Set default offset if not provided
	if offset < 0 {
		offset = 0
	}
	
	// Get users blocking from block use case
	blockingUsers, err := uc.blockUserUseCase.GetUsersBlocking(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get users blocking", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get users blocking: %w", err)
	}
	
	// For this response, we'll convert blocking users to blocked users format
	// for consistency in the response structure
	blockedUsers := make([]*moderation.BlockedUser, len(blockingUsers))
	for i, blockingUser := range blockingUsers {
		blockedUsers[i] = &moderation.BlockedUser{
			ID:        blockingUser.ID,
			BlockerID: blockingUser.BlockerID,
			BlockedID: blockingUser.BlockedID,
			Reason:    blockingUser.Reason,
			CreatedAt: blockingUser.CreatedAt,
			UpdatedAt: blockingUser.UpdatedAt,
		}
	}
	
	// Get total count for pagination
	totalCount, err := uc.blockUserUseCase.GetBlockCount(ctx, userID)
	if err != nil {
		logger.Error("Failed to get block count", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get block count: %w", err)
	}
	
	// Determine if there are more results
	hasMore := (offset + len(blockedUsers)) < int(totalCount)
	
	response := &GetBlockedUsersResponse{
		UserID:      userID,
		BlockedUsers: blockedUsers,
		TotalCount:   totalCount,
		HasMore:      hasMore,
	}
	
	logger.Info("GetUsersBlocking executed successfully", "user_id", userID, "blocking_count", len(blockingUsers))
	return response, nil
}

// IsBlocked checks if a user is blocked by another user
func (uc *GetBlockedUsersUseCase) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	logger.Info("Checking block status", "blocker_id", blockerID, "blocked_id", blockedID)
	
	isBlocked, err := uc.blockUserUseCase.GetBlockStatus(ctx, blockerID, blockedID)
	if err != nil {
		logger.Error("Failed to check block status", err, "blocker_id", blockerID, "blocked_id", blockedID)
		return false, fmt.Errorf("failed to check block status: %w", err)
	}
	
	return isBlocked, nil
}

// IsMutualBlock checks if two users have blocked each other
func (uc *GetBlockedUsersUseCase) IsMutualBlock(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	logger.Info("Checking mutual block", "user1", userID1, "user2", userID2)
	
	isMutual, err := uc.blockUserUseCase.IsMutualBlock(ctx, userID1, userID2)
	if err != nil {
		logger.Error("Failed to check mutual block", err, "user1", userID1, "user2", userID2)
		return false, fmt.Errorf("failed to check mutual block: %w", err)
	}
	
	return isMutual, nil
}