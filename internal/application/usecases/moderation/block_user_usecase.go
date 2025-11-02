
package moderation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// BlockUserRequest represents a request to block a user
type BlockUserRequest struct {
	BlockerID uuid.UUID `json:"blocker_id" validate:"required"`
	BlockedID uuid.UUID `json:"blocked_id" validate:"required"`
	Reason    *string   `json:"reason,omitempty"`
}

// BlockUserResponse represents the response after blocking a user
type BlockUserResponse struct {
	BlockID    uuid.UUID `json:"block_id"`
	BlockerID  uuid.UUID `json:"blocker_id"`
	BlockedID  uuid.UUID `json:"blocked_id"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
}

// BlockedUser represents a blocked user relationship
type BlockedUser struct {
	ID         uuid.UUID  `json:"id"`
	BlockerID  uuid.UUID  `json:"blocker_id"`
	BlockedID  uuid.UUID  `json:"blocked_id"`
	Reason     *string    `json:"reason,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// BlockUserUseCase handles user blocking functionality
type BlockUserUseCase struct {
	userRepo     repositories.UserRepository
	blockRepo    repositories.BlockRepository
	validator    validator.Validator
	cacheService CacheService
}

// BlockRepository defines interface for block operations
type BlockRepository interface {
	Create(ctx context.Context, block *BlockedUser) error
	Delete(ctx context.Context, blockerID, blockedID uuid.UUID) error
	Exists(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID, limit, offset int) ([]*BlockedUser, error)
	GetUsersBlocking(ctx context.Context, blockedID uuid.UUID, limit, offset int) ([]*BlockedUser, error)
	GetBlockCount(ctx context.Context, userID uuid.UUID) (int64, error)
	IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
	IsMutualBlock(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error)
}

// NewBlockUserUseCase creates a new BlockUserUseCase
func NewBlockUserUseCase(
	userRepo repositories.UserRepository,
	blockRepo BlockRepository,
	validator validator.Validator,
	cacheService CacheService,
) *BlockUserUseCase {
	return &BlockUserUseCase{
		userRepo:     userRepo,
		blockRepo:    blockRepo,
		validator:    validator,
		cacheService: cacheService,
	}
}

// Execute executes the block user use case
func (uc *BlockUserUseCase) Execute(ctx context.Context, req BlockUserRequest) (*BlockUserResponse, error) {
	logger.Info("Executing BlockUser use case", "blocker_id", req.BlockerID, "blocked_id", req.BlockedID)
	
	// Validate request
	if err := uc.validator.Struct(req); err != nil {
		logger.Error("Request validation failed", err, "request", req)
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Additional business validation
	if err := uc.validateBusinessRules(ctx, req); err != nil {
		logger.Error("Business validation failed", err, "request", req)
		return nil, fmt.Errorf("business validation failed: %w", err)
	}
	
	// Verify users exist
	if err := uc.verifyUsers(ctx, req); err != nil {
		logger.Error("User verification failed", err, "request", req)
		return nil, fmt.Errorf("user verification failed: %w", err)
	}
	
	// Check if already blocked
	isBlocked, err := uc.blockRepo.IsBlocked(ctx, req.BlockerID, req.BlockedID)
	if err != nil {
		logger.Error("Failed to check block status", err, "blocker_id", req.BlockerID, "blocked_id", req.BlockedID)
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}
	if isBlocked {
		return nil, fmt.Errorf("user is already blocked")
	}
	
	// Create block entity
	block := &BlockedUser{
		ID:        uuid.New(),
		BlockerID: req.BlockerID,
		BlockedID: req.BlockedID,
		Reason:    req.Reason,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save block to database
	if err := uc.blockRepo.Create(ctx, block); err != nil {
		logger.Error("Failed to create block", err, "block_id", block.ID)
		return nil, fmt.Errorf("failed to create block: %w", err)
	}
	
	// Invalidate cache
	uc.invalidateBlockCache(ctx, req.BlockerID)
	uc.invalidateBlockCache(ctx, req.BlockedID)
	
	// Handle mutual blocking
	if err := uc.handleMutualBlocking(ctx, req); err != nil {
		logger.Error("Failed to handle mutual blocking", err, "blocker_id", req.BlockerID, "blocked_id", req.BlockedID)
		// Don't fail the operation, just log the error
	}
	
	// Update user statistics
	if err := uc.updateUserStats(ctx, req); err != nil {
		logger.Error("Failed to update user stats", err, "request", req)
		// Don't fail the operation, just log the error
	}
	
	response := &BlockUserResponse{
		BlockID:   block.ID,
		BlockerID: req.BlockerID,
		BlockedID: req.BlockedID,
		CreatedAt: block.CreatedAt,
		Message:   "User blocked successfully",
	}
	
	logger.Info("BlockUser use case executed successfully", "block_id", block.ID, "blocker_id", req.BlockerID)
	return response, nil
}

// validateBusinessRules validates business rules for blocking
func (uc *BlockUserUseCase) validateBusinessRules(ctx context.Context, req BlockUserRequest) error {
	// Cannot block yourself
	if req.BlockerID == req.BlockedID {
		return fmt.Errorf("cannot block yourself")
	}
	
	// Check if users have an active match (optional business rule)
	if err := uc.checkUserMatch(ctx, req); err != nil {
		return fmt.Errorf("user match check failed: %w", err)
	}
	
	return nil
}

// verifyUsers verifies that both blocker and blocked users exist and are active
func (uc *BlockUserUseCase) verifyUsers(ctx context.Context, req BlockUserRequest) error {
	// Verify blocker exists and is active
	blocker, err := uc.userRepo.GetByID(ctx, req.BlockerID)
	if err != nil {
		return fmt.Errorf("blocker not found: %w", err)
	}
	if !blocker.IsActive || blocker.IsBanned {
		return fmt.Errorf("blocker account is not active")
	}
	
	// Verify blocked user exists
	blocked, err := uc.userRepo.GetByID(ctx, req.BlockedID)
	if err != nil {
		return fmt.Errorf("blocked user not found: %w", err)
	}
	
	// Allow blocking banned users (this is how we protect users)
	
	return nil
}

// checkUserMatch checks if users have an active match (optional business rule)
func (uc *BlockUserUseCase) checkUserMatch(ctx context.Context, req BlockUserRequest) error {
	// This is an optional business rule - some platforms don't allow blocking
	// users you have matched with, others do
	// For this implementation, we'll allow blocking anyone
	
	return nil
}

// handleMutualBlocking handles mutual blocking logic
func (uc *BlockUserUseCase) handleMutualBlocking(ctx context.Context, req BlockUserRequest) error {
	// Check if this is a mutual block
	isMutual, err := uc.blockRepo.IsMutualBlock(ctx, req.BlockerID, req.BlockedID)
	if err != nil {
		return fmt.Errorf("failed to check mutual block: %w", err)
	}
	
	if isMutual {
		// This is a mutual block - both users have blocked each other
		logger.Info("Mutual block detected", "user1", req.BlockerID, "user2", req.BlockedID)
		
		// In a mutual block scenario, we might want to:
		// 1. Hide both users from each other's feeds
		// 2. Prevent any further interaction
		// 3. Notify both users about the mutual block
		
		if err := uc.notifyMutualBlock(ctx, req); err != nil {
			logger.Error("Failed to notify mutual block", err, "user1", req.BlockerID, "user2", req.BlockedID)
			// Don't fail the operation, just log the error
		}
	}
	
	return nil
}

// updateUserStats updates user statistics after blocking
func (uc *BlockUserUseCase) updateUserStats(ctx context.Context, req BlockUserRequest) error {
	// In a real implementation, this would update user statistics
	// For now, we'll just log it
	logger.Info("User stats updated after blocking", "blocker_id", req.BlockerID, "blocked_id", req.BlockedID)
	return nil
}

// notifyMutualBlock notifies users about mutual blocking
func (uc *BlockUserUseCase) notifyMutualBlock(ctx context.Context, req BlockUserRequest) error {
	// In a real implementation, this would send notifications
	// For now, we'll just log it
	logger.Info("Mutual block notification sent", "user1", req.BlockerID, "user2", req.BlockedID)
	return nil
}

// invalidateBlockCache invalidates block cache for a user
func (uc *BlockUserUseCase) invalidateBlockCache(ctx context.Context, userID uuid.UUID) {
	cacheKey := fmt.Sprintf("blocked_users:%s", userID.String())
	uc.cacheService.Delete(ctx, cacheKey)
	
	cacheKey = fmt.Sprintf("blocking_users:%s", userID.String())
	uc.cacheService.Delete(ctx, cacheKey)
}

// GetBlockedUsers gets all users blocked by a user
func (uc *BlockUserUseCase) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID, limit, offset int) ([]*BlockedUser, error) {
	logger.Info("Getting blocked users", "blocker_id", blockerID, "limit", limit, "offset", offset)
	
	// Check cache first
	cacheKey := fmt.Sprintf("blocked_users:%s", blockerID.String())
	var cached []*BlockedUser
	if err := uc.cacheService.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}
	
	blockedUsers, err := uc.blockRepo.GetBlockedUsers(ctx, blockerID, limit, offset)
	if err != nil {
		logger.Error("Failed to get blocked users", err, "blocker_id", blockerID)
		return nil, fmt.Errorf("failed to get blocked users: %w", err)
	}
	
	// Cache result
	uc.cacheService.Set(ctx, cacheKey, blockedUsers, 15*time.Minute)
	
	return blockedUsers, nil
}

// GetUsersBlocking gets all users who are blocking a specific user
func (uc *BlockUserUseCase) GetUsersBlocking(ctx context.Context, blockedID uuid.UUID, limit, offset int) ([]*BlockedUser, error) {
	logger.Info("Getting users blocking user", "blocked_id", blockedID, "limit", limit, "offset", offset)
	
	// Check cache first
	cacheKey := fmt.Sprintf("blocking_users:%s", blockedID.String())
	var cached []*BlockedUser
	if err := uc.cacheService.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}
	
	blockingUsers, err := uc.blockRepo.GetUsersBlocking(ctx, blockedID, limit, offset)
	if err != nil {
		logger.Error("Failed to get users blocking", err, "blocked_id", blockedID)
		return nil, fmt.Errorf("failed to get users blocking: %w", err)
	}
	
	// Cache result
	uc.cacheService.Set(ctx, cacheKey, blockingUsers, 15*time.Minute)
	
	return blockingUsers, nil
}

// UnblockUser unblocks a user
func (uc *BlockUserUseCase) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	logger.Info("Unblocking user", "blocker_id", blockerID, "blocked_id", blockedID)
	
	// Check if block exists
	isBlocked, err := uc.blockRepo.IsBlocked(ctx, blockerID, blockedID)
	if err != nil {
		logger.Error("Failed to check block status", err, "blocker_id", blockerID, "blocked_id", blockedID)
		return fmt.Errorf("failed to check block status: %w", err)
	}
	if !isBlocked {
		return fmt.Errorf("user is not blocked")
	}
	
	// Delete block
	if err := uc.blockRepo.Delete(ctx, blockerID, blockedID); err != nil {
		logger.Error("Failed to delete block", err, "blocker_id", blockerID, "blocked_id", blockedID)
		return fmt.Errorf("failed to delete block: %w", err)
	}
	
	// Invalidate cache
	uc.invalidateBlockCache(ctx, blockerID)
	uc.invalidateBlockCache(ctx, blockedID)
	
	// Update user statistics
	if err := uc.updateUnblockStats(ctx, blockerID, blockedID); err != nil {
		logger.Error("Failed to update unblock stats", err, "blocker_id", blockerID, "blocked_id", blockedID)
		// Don't fail the operation, just log the error
	}
	
	logger.Info("User unblocked successfully", "blocker_id", blockerID, "blocked_id", blockedID)
	return nil
}

// updateUnblockStats updates user statistics after unblocking
func (uc *BlockUserUseCase) updateUnblockStats(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	// In a real implementation, this would update user statistics
	// For now, we'll just log it
	logger.Info("User stats updated after unblocking", "blocker_id", blockerID, "blocked_id", blockedID)
	return nil
}

// GetBlockStatus checks if a user is blocked by another user
func (uc *BlockUserUseCase) GetBlockStatus(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	logger.Info("Getting block status", "blocker_id", blockerID, "blocked_id", blockedID)
	
	isBlocked, err := uc.blockRepo.IsBlocked(ctx, blockerID, blockedID)
	if err != nil {
		logger.Error("Failed to get block status", err, "blocker_id", blockerID, "blocked_id", blockedID)
		return false, fmt.Errorf("failed to get block status: %w", err)
	}
	
	return isBlocked, nil
}

// GetBlockCount gets the number of users blocked by a user
func (uc *BlockUserUseCase) GetBlockCount(ctx context.Context, blockerID uuid.UUID) (int64, error) {
	logger.Info("Getting block count", "blocker_id", blockerID)
	
	count, err := uc.blockRepo.GetBlockCount(ctx, blockerID)
	if err != nil {
		logger.Error("Failed to get block count", err, "blocker_id", blockerID)
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}
	
	return count, nil
}

// IsMutualBlock checks if two users have blocked each other
func (uc *BlockUserUseCase) IsMutualBlock(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	logger.Info("Checking mutual block", "user1", userID1, "user2", userID2)
	
	isMutual, err := uc.blockRepo.IsMutualBlock(ctx, userID1, userID2)
	if err != nil {
		logger.Error("Failed to check mutual block", err, "user1", userID1, "user2", userID2)
		return false, fmt.Errorf("failed to check mutual block: %w", err)
	}
	
	return isMutual, nil
}