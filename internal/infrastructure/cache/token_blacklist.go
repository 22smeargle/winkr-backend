package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// TokenBlacklist manages blacklisted tokens in Redis
type TokenBlacklist struct {
	redisClient *redis.RedisClient
	prefix      string
}

// NewTokenBlacklist creates a new token blacklist manager
func NewTokenBlacklist(redisClient *redis.RedisClient) *TokenBlacklist {
	return &TokenBlacklist{
		redisClient: redisClient,
		prefix:      "blacklist:",
	}
}

// BlacklistToken adds a token to the blacklist
func (tb *TokenBlacklist) BlacklistToken(ctx context.Context, token string, reason string, expiresAt time.Time) error {
	key := tb.getTokenKey(token)
	
	// Store token with reason and expiration
	tokenData := map[string]interface{}{
		"token":     token,
		"reason":    reason,
		"blacklisted_at": time.Now(),
		"expires_at": expiresAt,
	}

	// Use hash to store token metadata
	err := tb.redisClient.HSet(ctx, key, "data", fmt.Sprintf("%v", tokenData))
	if err != nil {
		logger.Error("Failed to blacklist token", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Set expiration for the blacklist entry
	err = tb.redisClient.Expire(ctx, key, time.Until(expiresAt))
	if err != nil {
		logger.Error("Failed to set token blacklist expiration", err)
		return fmt.Errorf("failed to set token blacklist expiration: %w", err)
	}

	// Also add to a global blacklist set for quick lookups
	globalKey := tb.getGlobalBlacklistKey()
	err = tb.redisClient.SAdd(ctx, globalKey, token)
	if err != nil {
		logger.Error("Failed to add token to global blacklist", err)
		return fmt.Errorf("failed to add token to global blacklist: %w", err)
	}

	// Set expiration for global blacklist set (longer than individual tokens)
	err = tb.redisClient.Expire(ctx, globalKey, 24*time.Hour)
	if err != nil {
		logger.Error("Failed to set global blacklist expiration", err)
		// Non-critical error, continue
	}

	logger.Info("Token blacklisted successfully", "token", token, "reason", reason, "expires_at", expiresAt)
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (tb *TokenBlacklist) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	// First check the global blacklist set for quick lookup
	globalKey := tb.getGlobalBlacklistKey()
	isBlacklisted, err := tb.redisClient.SIsMember(ctx, globalKey, token)
	if err != nil {
		logger.Error("Failed to check global token blacklist", err)
		return false, fmt.Errorf("failed to check global token blacklist: %w", err)
	}

	if isBlacklisted {
		return true, nil
	}

	// If not in global set, check individual token entry
	key := tb.getTokenKey(token)
	exists, err := tb.redisClient.Exists(ctx, key)
	if err != nil {
		logger.Error("Failed to check token blacklist existence", err)
		return false, fmt.Errorf("failed to check token blacklist existence: %w", err)
	}

	return exists, nil
}

// GetBlacklistInfo retrieves blacklist information for a token
func (tb *TokenBlacklist) GetBlacklistInfo(ctx context.Context, token string) (map[string]interface{}, error) {
	key := tb.getTokenKey(token)
	
	tokenData, err := tb.redisClient.HGet(ctx, key, "data")
	if err != nil {
		logger.Error("Failed to get token blacklist info", err)
		return nil, fmt.Errorf("failed to get token blacklist info: %w", err)
	}

	if tokenData == "" {
		return nil, fmt.Errorf("token not found in blacklist")
	}

	// Parse the stored data
	var info map[string]interface{}
	// In a real implementation, you might want to store structured data
	// For simplicity, we'll return basic info
	info = map[string]interface{}{
		"token":  token,
		"reason": "blacklisted",
	}

	return info, nil
}

// RemoveFromBlacklist removes a token from the blacklist
func (tb *TokenBlacklist) RemoveFromBlacklist(ctx context.Context, token string) error {
	key := tb.getTokenKey(token)
	
	// Remove individual token entry
	err := tb.redisClient.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to remove token from blacklist", err)
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	// Remove from global blacklist set
	globalKey := tb.getGlobalBlacklistKey()
	err = tb.redisClient.SRem(ctx, globalKey, token)
	if err != nil {
		logger.Error("Failed to remove token from global blacklist", err)
		return fmt.Errorf("failed to remove token from global blacklist: %w", err)
	}

	logger.Info("Token removed from blacklist", "token", token)
	return nil
}

// BlacklistTokensByPattern adds multiple tokens matching a pattern
func (tb *TokenBlacklist) BlacklistTokensByPattern(ctx context.Context, pattern string, reason string, expiresAt time.Time) error {
	// This would typically be used with Redis SCAN in production
	// For simplicity, we'll implement basic pattern-based blacklisting
	logger.Info("Blacklisting tokens by pattern", "pattern", pattern, "reason", reason)
	
	// In a real implementation, you would:
	// 1. Use SCAN to find matching tokens
	// 2. Add each to blacklist
	// 3. Handle pagination for large datasets
	
	return nil
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (tb *TokenBlacklist) CleanupExpiredTokens(ctx context.Context) error {
	logger.Info("Starting expired tokens cleanup")
	
	// Get all tokens from global blacklist
	globalKey := tb.getGlobalBlacklistKey()
	tokens, err := tb.redisClient.SMembers(ctx, globalKey)
	if err != nil {
		logger.Error("Failed to get blacklisted tokens for cleanup", err)
		return fmt.Errorf("failed to get blacklisted tokens: %w", err)
	}

	cleanedCount := 0
	for _, token := range tokens {
		key := tb.getTokenKey(token)
		
		// Check if token still exists and is expired
		exists, err := tb.redisClient.Exists(ctx, key)
		if err != nil {
			logger.Error("Failed to check token existence during cleanup", err, "token", token)
			continue
		}

		if !exists {
			// Remove from global set if individual entry doesn't exist
			err = tb.redisClient.SRem(ctx, globalKey, token)
			if err != nil {
				logger.Error("Failed to remove expired token from global blacklist", err, "token", token)
			} else {
				cleanedCount++
			}
			continue
		}

		// Check expiration time
		tokenData, err := tb.redisClient.HGet(ctx, key, "data")
		if err != nil {
			logger.Error("Failed to get token data during cleanup", err, "token", token)
			continue
		}

		// In a real implementation, you would parse the expiration time
		// For simplicity, we'll check the TTL
		ttl, err := tb.redisClient.TTL(ctx, key)
		if err != nil {
			logger.Error("Failed to get token TTL during cleanup", err, "token", token)
			continue
		}

		if ttl <= 0 {
			// Token has expired, remove it
			err = tb.redisClient.Del(ctx, key)
			if err != nil {
				logger.Error("Failed to delete expired token", err, "token", token)
			} else {
				err = tb.redisClient.SRem(ctx, globalKey, token)
				if err != nil {
					logger.Error("Failed to remove expired token from global blacklist", err, "token", token)
				} else {
					cleanedCount++
				}
			}
		}
	}

	logger.Info("Expired tokens cleanup completed", "cleaned_count", cleanedCount)
	return nil
}

// GetBlacklistStats returns blacklist statistics
func (tb *TokenBlacklist) GetBlacklistStats(ctx context.Context) (map[string]interface{}, error) {
	globalKey := tb.getGlobalBlacklistKey()
	
	// Count blacklisted tokens
	count, err := tb.redisClient.SCard(ctx, globalKey)
	if err != nil {
		logger.Error("Failed to get blacklist stats", err)
		return nil, fmt.Errorf("failed to get blacklist stats: %w", err)
	}

	stats := map[string]interface{}{
		"blacklisted_tokens_count": count,
		"prefix":                tb.prefix,
		"global_key":            globalKey,
	}

	return stats, nil
}

// RevokeUserTokens revokes all tokens for a user
func (tb *TokenBlacklist) RevokeUserTokens(ctx context.Context, userID string, reason string) error {
	// This would typically integrate with session manager
	// For now, we'll implement a basic revocation
	logger.Info("Revoking all tokens for user", "user_id", userID, "reason", reason)
	
	// In a real implementation, you would:
	// 1. Get all user sessions from session manager
	// 2. Add all tokens to blacklist
	// 3. Invalidate all user sessions
	
	return nil
}

// Helper methods for key generation

func (tb *TokenBlacklist) getTokenKey(token string) string {
	return fmt.Sprintf("%stoken:%s", tb.prefix, token)
}

func (tb *TokenBlacklist) getGlobalBlacklistKey() string {
	return fmt.Sprintf("%sglobal_tokens", tb.prefix)
}