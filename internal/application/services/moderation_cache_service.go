package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// ModerationCacheService handles caching operations for moderation data
type ModerationCacheService struct {
	cacheService *cache.CacheService
	prefix       string
}

// NewModerationCacheService creates a new moderation cache service
func NewModerationCacheService(cacheService *cache.CacheService) *ModerationCacheService {
	return &ModerationCacheService{
		cacheService: cacheService,
		prefix:       "moderation:",
	}
}

// Moderation cache TTL constants
const (
	BlockedUsersCacheTTL      = 15 * time.Minute
	UserReputationCacheTTL     = 30 * time.Minute
	ModerationDecisionCacheTTL = 60 * time.Minute
	ReportStatsCacheTTL        = 10 * time.Minute
	ContentAnalysisCacheTTL    = 45 * time.Minute
	ModerationQueueCacheTTL    = 5 * time.Minute
	BanStatusCacheTTL          = 30 * time.Minute
	AppealStatusCacheTTL       = 15 * time.Minute
	ModeratorStatsCacheTTL     = 20 * time.Minute
)

// CacheBlockedUsers caches a user's blocked users list
func (mcs *ModerationCacheService) CacheBlockedUsers(ctx context.Context, userID uuid.UUID, blockedUsers []uuid.UUID) error {
	key := mcs.getBlockedUsersKey(userID)
	
	blockedUsersData, err := json.Marshal(blockedUsers)
	if err != nil {
		logger.Error("Failed to marshal blocked users for caching", err)
		return fmt.Errorf("failed to marshal blocked users: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(blockedUsersData), BlockedUsersCacheTTL)
	if err != nil {
		logger.Error("Failed to cache blocked users", err)
		return fmt.Errorf("failed to cache blocked users: %w", err)
	}

	logger.Debug("Blocked users cached", "user_id", userID, "count", len(blockedUsers))
	return nil
}

// GetBlockedUsers retrieves cached blocked users list
func (mcs *ModerationCacheService) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	key := mcs.getBlockedUsersKey(userID)
	
	blockedUsersData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached blocked users", err)
		return nil, fmt.Errorf("failed to get cached blocked users: %w", err)
	}

	if blockedUsersData == "" {
		return nil, nil // Cache miss
	}

	var blockedUsers []uuid.UUID
	err = json.Unmarshal([]byte(blockedUsersData), &blockedUsers)
	if err != nil {
		logger.Error("Failed to unmarshal cached blocked users", err)
		return nil, fmt.Errorf("failed to unmarshal cached blocked users: %w", err)
	}

	logger.Debug("Blocked users retrieved from cache", "user_id", userID, "count", len(blockedUsers))
	return blockedUsers, nil
}

// InvalidateBlockedUsers removes blocked users from cache
func (mcs *ModerationCacheService) InvalidateBlockedUsers(ctx context.Context, userID uuid.UUID) error {
	key := mcs.getBlockedUsersKey(userID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate blocked users cache", err)
		return fmt.Errorf("failed to invalidate blocked users cache: %w", err)
	}

	logger.Debug("Blocked users cache invalidated", "user_id", userID)
	return nil
}

// CacheUserReputation caches user reputation data
func (mcs *ModerationCacheService) CacheUserReputation(ctx context.Context, userID uuid.UUID, reputation *models.UserReputation) error {
	key := mcs.getUserReputationKey(userID)
	
	reputationData, err := json.Marshal(reputation)
	if err != nil {
		logger.Error("Failed to marshal user reputation for caching", err)
		return fmt.Errorf("failed to marshal user reputation: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(reputationData), UserReputationCacheTTL)
	if err != nil {
		logger.Error("Failed to cache user reputation", err)
		return fmt.Errorf("failed to cache user reputation: %w", err)
	}

	logger.Debug("User reputation cached", "user_id", userID, "score", reputation.Score)
	return nil
}

// GetUserReputation retrieves cached user reputation
func (mcs *ModerationCacheService) GetUserReputation(ctx context.Context, userID uuid.UUID) (*models.UserReputation, error) {
	key := mcs.getUserReputationKey(userID)
	
	reputationData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached user reputation", err)
		return nil, fmt.Errorf("failed to get cached user reputation: %w", err)
	}

	if reputationData == "" {
		return nil, nil // Cache miss
	}

	var reputation models.UserReputation
	err = json.Unmarshal([]byte(reputationData), &reputation)
	if err != nil {
		logger.Error("Failed to unmarshal cached user reputation", err)
		return nil, fmt.Errorf("failed to unmarshal cached user reputation: %w", err)
	}

	logger.Debug("User reputation retrieved from cache", "user_id", userID, "score", reputation.Score)
	return &reputation, nil
}

// InvalidateUserReputation removes user reputation from cache
func (mcs *ModerationCacheService) InvalidateUserReputation(ctx context.Context, userID uuid.UUID) error {
	key := mcs.getUserReputationKey(userID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate user reputation cache", err)
		return fmt.Errorf("failed to invalidate user reputation cache: %w", err)
	}

	logger.Debug("User reputation cache invalidated", "user_id", userID)
	return nil
}

// CacheModerationDecision caches moderation decision for content
func (mcs *ModerationCacheService) CacheModerationDecision(ctx context.Context, contentType string, contentID uuid.UUID, decision string) error {
	key := mcs.getModerationDecisionKey(contentType, contentID)
	
	err := mcs.cacheService.Set(ctx, key, decision, ModerationDecisionCacheTTL)
	if err != nil {
		logger.Error("Failed to cache moderation decision", err)
		return fmt.Errorf("failed to cache moderation decision: %w", err)
	}

	logger.Debug("Moderation decision cached", "content_type", contentType, "content_id", contentID, "decision", decision)
	return nil
}

// GetModerationDecision retrieves cached moderation decision
func (mcs *ModerationCacheService) GetModerationDecision(ctx context.Context, contentType string, contentID uuid.UUID) (string, error) {
	key := mcs.getModerationDecisionKey(contentType, contentID)
	
	decision, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached moderation decision", err)
		return "", fmt.Errorf("failed to get cached moderation decision: %w", err)
	}

	if decision == "" {
		return "", nil // Cache miss
	}

	logger.Debug("Moderation decision retrieved from cache", "content_type", contentType, "content_id", contentID, "decision", decision)
	return decision, nil
}

// InvalidateModerationDecision removes moderation decision from cache
func (mcs *ModerationCacheService) InvalidateModerationDecision(ctx context.Context, contentType string, contentID uuid.UUID) error {
	key := mcs.getModerationDecisionKey(contentType, contentID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate moderation decision cache", err)
		return fmt.Errorf("failed to invalidate moderation decision cache: %w", err)
	}

	logger.Debug("Moderation decision cache invalidated", "content_type", contentType, "content_id", contentID)
	return nil
}

// CacheReportStats caches report statistics
func (mcs *ModerationCacheService) CacheReportStats(ctx context.Context, stats map[string]interface{}) error {
	key := mcs.getReportStatsKey()
	
	statsData, err := json.Marshal(stats)
	if err != nil {
		logger.Error("Failed to marshal report stats for caching", err)
		return fmt.Errorf("failed to marshal report stats: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(statsData), ReportStatsCacheTTL)
	if err != nil {
		logger.Error("Failed to cache report stats", err)
		return fmt.Errorf("failed to cache report stats: %w", err)
	}

	logger.Debug("Report stats cached")
	return nil
}

// GetReportStats retrieves cached report statistics
func (mcs *ModerationCacheService) GetReportStats(ctx context.Context) (map[string]interface{}, error) {
	key := mcs.getReportStatsKey()
	
	statsData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached report stats", err)
		return nil, fmt.Errorf("failed to get cached report stats: %w", err)
	}

	if statsData == "" {
		return nil, nil // Cache miss
	}

	var stats map[string]interface{}
	err = json.Unmarshal([]byte(statsData), &stats)
	if err != nil {
		logger.Error("Failed to unmarshal cached report stats", err)
		return nil, fmt.Errorf("failed to unmarshal cached report stats: %w", err)
	}

	logger.Debug("Report stats retrieved from cache")
	return stats, nil
}

// InvalidateReportStats removes report statistics from cache
func (mcs *ModerationCacheService) InvalidateReportStats(ctx context.Context) error {
	key := mcs.getReportStatsKey()
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate report stats cache", err)
		return fmt.Errorf("failed to invalidate report stats cache: %w", err)
	}

	logger.Debug("Report stats cache invalidated")
	return nil
}

// CacheContentAnalysis caches content analysis result
func (mcs *ModerationCacheService) CacheContentAnalysis(ctx context.Context, contentID uuid.UUID, analysis *models.ContentAnalysis) error {
	key := mcs.getContentAnalysisKey(contentID)
	
	analysisData, err := json.Marshal(analysis)
	if err != nil {
		logger.Error("Failed to marshal content analysis for caching", err)
		return fmt.Errorf("failed to marshal content analysis: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(analysisData), ContentAnalysisCacheTTL)
	if err != nil {
		logger.Error("Failed to cache content analysis", err)
		return fmt.Errorf("failed to cache content analysis: %w", err)
	}

	logger.Debug("Content analysis cached", "content_id", contentID, "result", analysis.Result)
	return nil
}

// GetContentAnalysis retrieves cached content analysis
func (mcs *ModerationCacheService) GetContentAnalysis(ctx context.Context, contentID uuid.UUID) (*models.ContentAnalysis, error) {
	key := mcs.getContentAnalysisKey(contentID)
	
	analysisData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached content analysis", err)
		return nil, fmt.Errorf("failed to get cached content analysis: %w", err)
	}

	if analysisData == "" {
		return nil, nil // Cache miss
	}

	var analysis models.ContentAnalysis
	err = json.Unmarshal([]byte(analysisData), &analysis)
	if err != nil {
		logger.Error("Failed to unmarshal cached content analysis", err)
		return nil, fmt.Errorf("failed to unmarshal cached content analysis: %w", err)
	}

	logger.Debug("Content analysis retrieved from cache", "content_id", contentID, "result", analysis.Result)
	return &analysis, nil
}

// InvalidateContentAnalysis removes content analysis from cache
func (mcs *ModerationCacheService) InvalidateContentAnalysis(ctx context.Context, contentID uuid.UUID) error {
	key := mcs.getContentAnalysisKey(contentID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate content analysis cache", err)
		return fmt.Errorf("failed to invalidate content analysis cache: %w", err)
	}

	logger.Debug("Content analysis cache invalidated", "content_id", contentID)
	return nil
}

// CacheModerationQueue caches moderation queue items
func (mcs *ModerationCacheService) CacheModerationQueue(ctx context.Context, queueItems []*models.ModerationQueue) error {
	key := mcs.getModerationQueueKey()
	
	queueData, err := json.Marshal(queueItems)
	if err != nil {
		logger.Error("Failed to marshal moderation queue for caching", err)
		return fmt.Errorf("failed to marshal moderation queue: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(queueData), ModerationQueueCacheTTL)
	if err != nil {
		logger.Error("Failed to cache moderation queue", err)
		return fmt.Errorf("failed to cache moderation queue: %w", err)
	}

	logger.Debug("Moderation queue cached", "count", len(queueItems))
	return nil
}

// GetModerationQueue retrieves cached moderation queue
func (mcs *ModerationCacheService) GetModerationQueue(ctx context.Context) ([]*models.ModerationQueue, error) {
	key := mcs.getModerationQueueKey()
	
	queueData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached moderation queue", err)
		return nil, fmt.Errorf("failed to get cached moderation queue: %w", err)
	}

	if queueData == "" {
		return nil, nil // Cache miss
	}

	var queueItems []*models.ModerationQueue
	err = json.Unmarshal([]byte(queueData), &queueItems)
	if err != nil {
		logger.Error("Failed to unmarshal cached moderation queue", err)
		return nil, fmt.Errorf("failed to unmarshal cached moderation queue: %w", err)
	}

	logger.Debug("Moderation queue retrieved from cache", "count", len(queueItems))
	return queueItems, nil
}

// InvalidateModerationQueue removes moderation queue from cache
func (mcs *ModerationCacheService) InvalidateModerationQueue(ctx context.Context) error {
	key := mcs.getModerationQueueKey()
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate moderation queue cache", err)
		return fmt.Errorf("failed to invalidate moderation queue cache: %w", err)
	}

	logger.Debug("Moderation queue cache invalidated")
	return nil
}

// CacheBanStatus caches user ban status
func (mcs *ModerationCacheService) CacheBanStatus(ctx context.Context, userID uuid.UUID, isBanned bool, banInfo *models.Ban) error {
	key := mcs.getBanStatusKey(userID)
	
	banStatus := map[string]interface{}{
		"is_banned": isBanned,
		"ban_info":  banInfo,
	}
	
	banStatusData, err := json.Marshal(banStatus)
	if err != nil {
		logger.Error("Failed to marshal ban status for caching", err)
		return fmt.Errorf("failed to marshal ban status: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(banStatusData), BanStatusCacheTTL)
	if err != nil {
		logger.Error("Failed to cache ban status", err)
		return fmt.Errorf("failed to cache ban status: %w", err)
	}

	logger.Debug("Ban status cached", "user_id", userID, "is_banned", isBanned)
	return nil
}

// GetBanStatus retrieves cached ban status
func (mcs *ModerationCacheService) GetBanStatus(ctx context.Context, userID uuid.UUID) (bool, *models.Ban, error) {
	key := mcs.getBanStatusKey(userID)
	
	banStatusData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached ban status", err)
		return false, nil, fmt.Errorf("failed to get cached ban status: %w", err)
	}

	if banStatusData == "" {
		return false, nil // Cache miss
	}

	var banStatus map[string]interface{}
	err = json.Unmarshal([]byte(banStatusData), &banStatus)
	if err != nil {
		logger.Error("Failed to unmarshal cached ban status", err)
		return false, nil, fmt.Errorf("failed to unmarshal cached ban status: %w", err)
	}

	isBanned := banStatus["is_banned"].(bool)
	var banInfo *models.Ban
	
	if banInfoData, exists := banStatus["ban_info"]; exists && banInfoData != nil {
		banInfoBytes, _ := json.Marshal(banInfoData)
		json.Unmarshal(banInfoBytes, &banInfo)
	}

	logger.Debug("Ban status retrieved from cache", "user_id", userID, "is_banned", isBanned)
	return isBanned, banInfo, nil
}

// InvalidateBanStatus removes ban status from cache
func (mcs *ModerationCacheService) InvalidateBanStatus(ctx context.Context, userID uuid.UUID) error {
	key := mcs.getBanStatusKey(userID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate ban status cache", err)
		return fmt.Errorf("failed to invalidate ban status cache: %w", err)
	}

	logger.Debug("Ban status cache invalidated", "user_id", userID)
	return nil
}

// CacheAppealStatus caches user appeal status
func (mcs *ModerationCacheService) CacheAppealStatus(ctx context.Context, appealID uuid.UUID, appeal *models.Appeal) error {
	key := mcs.getAppealStatusKey(appealID)
	
	appealData, err := json.Marshal(appeal)
	if err != nil {
		logger.Error("Failed to marshal appeal status for caching", err)
		return fmt.Errorf("failed to marshal appeal status: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(appealData), AppealStatusCacheTTL)
	if err != nil {
		logger.Error("Failed to cache appeal status", err)
		return fmt.Errorf("failed to cache appeal status: %w", err)
	}

	logger.Debug("Appeal status cached", "appeal_id", appealID, "status", appeal.Status)
	return nil
}

// GetAppealStatus retrieves cached appeal status
func (mcs *ModerationCacheService) GetAppealStatus(ctx context.Context, appealID uuid.UUID) (*models.Appeal, error) {
	key := mcs.getAppealStatusKey(appealID)
	
	appealData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached appeal status", err)
		return nil, fmt.Errorf("failed to get cached appeal status: %w", err)
	}

	if appealData == "" {
		return nil, nil // Cache miss
	}

	var appeal models.Appeal
	err = json.Unmarshal([]byte(appealData), &appeal)
	if err != nil {
		logger.Error("Failed to unmarshal cached appeal status", err)
		return nil, fmt.Errorf("failed to unmarshal cached appeal status: %w", err)
	}

	logger.Debug("Appeal status retrieved from cache", "appeal_id", appealID, "status", appeal.Status)
	return &appeal, nil
}

// InvalidateAppealStatus removes appeal status from cache
func (mcs *ModerationCacheService) InvalidateAppealStatus(ctx context.Context, appealID uuid.UUID) error {
	key := mcs.getAppealStatusKey(appealID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate appeal status cache", err)
		return fmt.Errorf("failed to invalidate appeal status cache: %w", err)
	}

	logger.Debug("Appeal status cache invalidated", "appeal_id", appealID)
	return nil
}

// CacheModeratorStats caches moderator statistics
func (mcs *ModerationCacheService) CacheModeratorStats(ctx context.Context, moderatorID uuid.UUID, stats map[string]interface{}) error {
	key := mcs.getModeratorStatsKey(moderatorID)
	
	statsData, err := json.Marshal(stats)
	if err != nil {
		logger.Error("Failed to marshal moderator stats for caching", err)
		return fmt.Errorf("failed to marshal moderator stats: %w", err)
	}

	err = mcs.cacheService.Set(ctx, key, string(statsData), ModeratorStatsCacheTTL)
	if err != nil {
		logger.Error("Failed to cache moderator stats", err)
		return fmt.Errorf("failed to cache moderator stats: %w", err)
	}

	logger.Debug("Moderator stats cached", "moderator_id", moderatorID)
	return nil
}

// GetModeratorStats retrieves cached moderator statistics
func (mcs *ModerationCacheService) GetModeratorStats(ctx context.Context, moderatorID uuid.UUID) (map[string]interface{}, error) {
	key := mcs.getModeratorStatsKey(moderatorID)
	
	statsData, err := mcs.cacheService.Get(ctx, key)
	if err != nil {
		logger.Error("Failed to get cached moderator stats", err)
		return nil, fmt.Errorf("failed to get cached moderator stats: %w", err)
	}

	if statsData == "" {
		return nil, nil // Cache miss
	}

	var stats map[string]interface{}
	err = json.Unmarshal([]byte(statsData), &stats)
	if err != nil {
		logger.Error("Failed to unmarshal cached moderator stats", err)
		return nil, fmt.Errorf("failed to unmarshal cached moderator stats: %w", err)
	}

	logger.Debug("Moderator stats retrieved from cache", "moderator_id", moderatorID)
	return stats, nil
}

// InvalidateModeratorStats removes moderator statistics from cache
func (mcs *ModerationCacheService) InvalidateModeratorStats(ctx context.Context, moderatorID uuid.UUID) error {
	key := mcs.getModeratorStatsKey(moderatorID)
	
	err := mcs.cacheService.Del(ctx, key)
	if err != nil {
		logger.Error("Failed to invalidate moderator stats cache", err)
		return fmt.Errorf("failed to invalidate moderator stats cache: %w", err)
	}

	logger.Debug("Moderator stats cache invalidated", "moderator_id", moderatorID)
	return nil
}

// InvalidateUserModerationData removes all moderation-related data for a user
func (mcs *ModerationCacheService) InvalidateUserModerationData(ctx context.Context, userID uuid.UUID) error {
	keys := []string{
		mcs.getBlockedUsersKey(userID),
		mcs.getUserReputationKey(userID),
		mcs.getBanStatusKey(userID),
	}

	for _, key := range keys {
		err := mcs.cacheService.Del(ctx, key)
		if err != nil {
			logger.Error("Failed to invalidate user moderation data", err, "key", key)
			return fmt.Errorf("failed to invalidate user moderation data for key %s: %w", key, err)
		}
	}

	logger.Debug("User moderation data invalidated", "user_id", userID)
	return nil
}

// InvalidateAllModerationData removes all moderation-related data from cache
func (mcs *ModerationCacheService) InvalidateAllModerationData(ctx context.Context) error {
	keys := []string{
		mcs.getReportStatsKey(),
		mcs.getModerationQueueKey(),
	}

	for _, key := range keys {
		err := mcs.cacheService.Del(ctx, key)
		if err != nil {
			logger.Error("Failed to invalidate moderation data", err, "key", key)
			return fmt.Errorf("failed to invalidate moderation data for key %s: %w", key, err)
		}
	}

	logger.Debug("All moderation data invalidated")
	return nil
}

// Helper methods for key generation

func (mcs *ModerationCacheService) getBlockedUsersKey(userID uuid.UUID) string {
	return fmt.Sprintf("%sblocked_users:%s", mcs.prefix, userID)
}

func (mcs *ModerationCacheService) getUserReputationKey(userID uuid.UUID) string {
	return fmt.Sprintf("%suser_reputation:%s", mcs.prefix, userID)
}

func (mcs *ModerationCacheService) getModerationDecisionKey(contentType string, contentID uuid.UUID) string {
	return fmt.Sprintf("%smoderation_decision:%s:%s", mcs.prefix, contentType, contentID)
}

func (mcs *ModerationCacheService) getReportStatsKey() string {
	return fmt.Sprintf("%sreport_stats", mcs.prefix)
}

func (mcs *ModerationCacheService) getContentAnalysisKey(contentID uuid.UUID) string {
	return fmt.Sprintf("%scontent_analysis:%s", mcs.prefix, contentID)
}

func (mcs *ModerationCacheService) getModerationQueueKey() string {
	return fmt.Sprintf("%smoderation_queue", mcs.prefix)
}

func (mcs *ModerationCacheService) getBanStatusKey(userID uuid.UUID) string {
	return fmt.Sprintf("%sban_status:%s", mcs.prefix, userID)
}

func (mcs *ModerationCacheService) getAppealStatusKey(appealID uuid.UUID) string {
	return fmt.Sprintf("%sappeal_status:%s", mcs.prefix, appealID)
}

func (mcs *ModerationCacheService) getModeratorStatsKey(moderatorID uuid.UUID) string {
	return fmt.Sprintf("%smoderator_stats:%s", mcs.prefix, moderatorID)
}