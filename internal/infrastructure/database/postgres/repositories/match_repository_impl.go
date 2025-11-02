package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// MatchRepositoryImpl implements MatchRepository interface using GORM
type MatchRepositoryImpl struct {
	db *gorm.DB
}

// NewMatchRepository creates a new MatchRepository instance
func NewMatchRepository(db *gorm.DB) repositories.MatchRepository {
	return &MatchRepositoryImpl{db: db}
}

// Match methods

// CreateMatch creates a new match
func (r *MatchRepositoryImpl) CreateMatch(ctx context.Context, match *entities.Match) error {
	modelMatch := r.domainToModelMatch(match)
	if err := r.db.WithContext(ctx).Create(modelMatch).Error; err != nil {
		logger.Error("Failed to create match", err)
		return fmt.Errorf("failed to create match: %w", err)
	}

	logger.Info("Match created successfully", map[string]interface{}{
		"match_id": match.ID,
		"user1_id": match.User1ID,
		"user2_id": match.User2ID,
	})
	return nil
}

// GetByID retrieves a match by ID
func (r *MatchRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Match, error) {
	var match models.Match
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&match).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("match not found")
		}
		logger.Error("Failed to get match by ID", err)
		return nil, fmt.Errorf("failed to get match by ID: %w", err)
	}

	// Convert to domain entity
	domainMatch := r.modelToDomainMatch(&match)
	return domainMatch, nil
}

// Update updates a match
func (r *MatchRepositoryImpl) Update(ctx context.Context, match *entities.Match) error {
	modelMatch := r.domainToModelMatch(match)
	if err := r.db.WithContext(ctx).Save(modelMatch).Error; err != nil {
		logger.Error("Failed to update match", err)
		return fmt.Errorf("failed to update match: %w", err)
	}

	logger.Info("Match updated successfully", map[string]interface{}{
		"match_id": match.ID,
	})
	return nil
}

// DeleteMatch soft deletes a match
func (r *MatchRepositoryImpl) DeleteMatch(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Match{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete match", err)
		return fmt.Errorf("failed to delete match: %w", err)
	}

	logger.Info("Match deleted successfully", map[string]interface{}{
		"match_id": id,
	})
	return nil
}

// GetUserMatches retrieves matches for a user
func (r *MatchRepositoryImpl) GetUserMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error) {
	var matches []models.Match
	if err := r.db.WithContext(ctx).Where("user1_id = ? OR user2_id = ?", userID, userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&matches).Error; err != nil {
		logger.Error("Failed to get user matches", err)
		return nil, fmt.Errorf("failed to get user matches: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetMatchByUsers retrieves match between two users
func (r *MatchRepositoryImpl) GetMatchByUsers(ctx context.Context, user1ID, user2ID uuid.UUID) (*entities.Match, error) {
	var match models.Match
	if err := r.db.WithContext(ctx).Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).First(&match).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("match not found")
		}
		logger.Error("Failed to get match by users", err)
		return nil, fmt.Errorf("failed to get match by users: %w", err)
	}

	// Convert to domain entity
	domainMatch := r.modelToDomainMatch(&match)
	return domainMatch, nil
}

// ExistsMatch checks if match exists between two users
func (r *MatchRepositoryImpl) ExistsMatch(ctx context.Context, user1ID, user2ID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).Count(&count).Error; err != nil {
		logger.Error("Failed to check match existence", err)
		return false, fmt.Errorf("failed to check match existence: %w", err)
	}

	return count > 0, nil
}

// GetActiveMatches retrieves active matches for a user
func (r *MatchRepositoryImpl) GetActiveMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Match, error) {
	var matches []models.Match
	if err := r.db.WithContext(ctx).Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).Order("created_at DESC").Limit(limit).Offset(offset).Find(&matches).Error; err != nil {
		logger.Error("Failed to get active matches", err)
		return nil, fmt.Errorf("failed to get active matches: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetMatchCount retrieves match count for a user
func (r *MatchRepositoryImpl) GetMatchCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Match{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&count).Error; err != nil {
		logger.Error("Failed to get match count", err)
		return 0, fmt.Errorf("failed to get match count: %w", err)
	}

	return count, nil
}

// GetActiveMatchCount retrieves active match count for a user
func (r *MatchRepositoryImpl) GetActiveMatchCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).Count(&count).Error; err != nil {
		logger.Error("Failed to get active match count", err)
		return 0, fmt.Errorf("failed to get active match count: %w", err)
	}

	return count, nil
}

// DeactivateMatch deactivates a match
func (r *MatchRepositoryImpl) DeactivateMatch(ctx context.Context, matchID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Match{}).Where("id = ?", matchID).Update("is_active", false).Error; err != nil {
		logger.Error("Failed to deactivate match", err)
		return fmt.Errorf("failed to deactivate match: %w", err)
	}

	logger.Info("Match deactivated", map[string]interface{}{
		"match_id": matchID,
	})
	return nil
}

// ActivateMatch activates a match
func (r *MatchRepositoryImpl) ActivateMatch(ctx context.Context, matchID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Match{}).Where("id = ?", matchID).Update("is_active", true).Error; err != nil {
		logger.Error("Failed to activate match", err)
		return fmt.Errorf("failed to activate match: %w", err)
	}

	logger.Info("Match activated", map[string]interface{}{
		"match_id": matchID,
	})
	return nil
}

// GetMatchesCreatedInRange retrieves matches created in date range
func (r *MatchRepositoryImpl) GetMatchesCreatedInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.Match, error) {
	var matches []models.Match
	if err := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&matches).Error; err != nil {
		logger.Error("Failed to get matches created in range", err)
		return nil, fmt.Errorf("failed to get matches created in range: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// Swipe methods

// CreateSwipe creates a new swipe
func (r *MatchRepositoryImpl) CreateSwipe(ctx context.Context, swipe *entities.Swipe) error {
	modelSwipe := r.domainToModelSwipe(swipe)
	if err := r.db.WithContext(ctx).Create(modelSwipe).Error; err != nil {
		logger.Error("Failed to create swipe", err)
		return fmt.Errorf("failed to create swipe: %w", err)
	}

	logger.Info("Swipe created successfully", map[string]interface{}{
		"swipe_id": swipe.ID,
		"swiper_id": swipe.SwiperID,
		"swiped_id": swipe.SwipedID,
		"is_like": swipe.IsLike,
	})
	return nil
}

// GetSwipeByID retrieves a swipe by ID
func (r *MatchRepositoryImpl) GetSwipeByID(ctx context.Context, id uuid.UUID) (*entities.Swipe, error) {
	var swipe models.Swipe
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&swipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("swipe not found")
		}
		logger.Error("Failed to get swipe by ID", err)
		return nil, fmt.Errorf("failed to get swipe by ID: %w", err)
	}

	// Convert to domain entity
	domainSwipe := r.modelToDomainSwipe(&swipe)
	return domainSwipe, nil
}

// UpdateSwipe updates a swipe
func (r *MatchRepositoryImpl) UpdateSwipe(ctx context.Context, swipe *entities.Swipe) error {
	modelSwipe := r.domainToModelSwipe(swipe)
	if err := r.db.WithContext(ctx).Save(modelSwipe).Error; err != nil {
		logger.Error("Failed to update swipe", err)
		return fmt.Errorf("failed to update swipe: %w", err)
	}

	logger.Info("Swipe updated successfully", map[string]interface{}{
		"swipe_id": swipe.ID,
	})
	return nil
}


// DeleteSwipe soft deletes a swipe
func (r *MatchRepositoryImpl) DeleteSwipe(ctx context.Context, swiperID, swipedID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).Delete(&models.Swipe{}).Error; err != nil {
		logger.Error("Failed to delete swipe", err)
		return fmt.Errorf("failed to delete swipe: %w", err)
	}

	logger.Info("Swipe deleted successfully", map[string]interface{}{
		"swiper_id": swiperID,
		"swiped_id": swipedID,
	})
	return nil
}

// GetUserSwipes retrieves swipes for a user
func (r *MatchRepositoryImpl) GetUserSwipes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get user swipes", err)
		return nil, fmt.Errorf("failed to get user swipes: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// GetSwipeByUsers retrieves swipe from user to target
func (r *MatchRepositoryImpl) GetSwipeByUsers(ctx context.Context, userID, targetID uuid.UUID) (*entities.Swipe, error) {
	var swipe models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND swiped_id = ?", userID, targetID).First(&swipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("swipe not found")
		}
		logger.Error("Failed to get swipe by users", err)
		return nil, fmt.Errorf("failed to get swipe by users: %w", err)
	}

	// Convert to domain entity
	domainSwipe := r.modelToDomainSwipe(&swipe)
	return domainSwipe, nil
}

// ExistsSwipe checks if swipe exists from user to target
func (r *MatchRepositoryImpl) ExistsSwipe(ctx context.Context, userID, targetID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND swiped_id = ?", userID, targetID).Count(&count).Error; err != nil {
		logger.Error("Failed to check swipe existence", err)
		return false, fmt.Errorf("failed to check swipe existence: %w", err)
	}

	return count > 0, nil
}

// GetUserSwipesByDirection retrieves swipes for a user by direction
func (r *MatchRepositoryImpl) GetUserSwipesByDirection(ctx context.Context, userID uuid.UUID, direction string, limit, offset int) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Where("user_id = ? AND direction = ?", userID, direction).Order("created_at DESC").Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get user swipes by direction", err)
		return nil, fmt.Errorf("failed to get user swipes by direction: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// GetSwipeCount retrieves swipe count for a user
func (r *MatchRepositoryImpl) GetSwipeCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to get swipe count", err)
		return 0, fmt.Errorf("failed to get swipe count: %w", err)
	}

	return count, nil
}

// GetSwipeCountByDirection retrieves swipe count for a user by direction
func (r *MatchRepositoryImpl) GetSwipeCountByDirection(ctx context.Context, userID uuid.UUID, direction string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("user_id = ? AND direction = ?", userID, direction).Count(&count).Error; err != nil {
		logger.Error("Failed to get swipe count by direction", err)
		return 0, fmt.Errorf("failed to get swipe count by direction: %w", err)
	}

	return count, nil
}

// GetSwipesCreatedInRange retrieves swipes created in date range
func (r *MatchRepositoryImpl) GetSwipesCreatedInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get swipes created in range", err)
		return nil, fmt.Errorf("failed to get swipes created in range: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// UserPreferences methods

// CreateUserPreferences creates new user preferences
func (r *MatchRepositoryImpl) CreateUserPreferences(ctx context.Context, preferences *entities.UserPreferences) error {
	modelPreferences := r.domainToModelUserPreferences(preferences)
	if err := r.db.WithContext(ctx).Create(modelPreferences).Error; err != nil {
		logger.Error("Failed to create user preferences", err)
		return fmt.Errorf("failed to create user preferences: %w", err)
	}

	logger.Info("User preferences created successfully", map[string]interface{}{
		"user_id": preferences.UserID,
	})
	return nil
}

// GetUserPreferences retrieves user preferences
func (r *MatchRepositoryImpl) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*entities.UserPreferences, error) {
	var preferences models.UserPreferences
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user preferences not found")
		}
		logger.Error("Failed to get user preferences", err)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Convert to domain entity
	domainPreferences := r.modelToDomainUserPreferences(&preferences)
	return domainPreferences, nil
}

// UpdateUserPreferences updates user preferences
func (r *MatchRepositoryImpl) UpdateUserPreferences(ctx context.Context, preferences *entities.UserPreferences) error {
	modelPreferences := r.domainToModelUserPreferences(preferences)
	if err := r.db.WithContext(ctx).Save(modelPreferences).Error; err != nil {
		logger.Error("Failed to update user preferences", err)
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	logger.Info("User preferences updated successfully", map[string]interface{}{
		"user_id": preferences.UserID,
	})
	return nil
}

// DeleteUserPreferences deletes user preferences
func (r *MatchRepositoryImpl) DeleteUserPreferences(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserPreferences{}).Error; err != nil {
		logger.Error("Failed to delete user preferences", err)
		return fmt.Errorf("failed to delete user preferences: %w", err)
	}

	logger.Info("User preferences deleted successfully", map[string]interface{}{
		"user_id": userID,
	})
	return nil
}

// ExistsUserPreferences checks if user preferences exist
func (r *MatchRepositoryImpl) ExistsUserPreferences(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserPreferences{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to check user preferences existence", err)
		return false, fmt.Errorf("failed to check user preferences existence: %w", err)
	}

	return count > 0, nil
}

// GetMatchStats retrieves match statistics
func (r *MatchRepositoryImpl) GetMatchStats(ctx context.Context, userID uuid.UUID) (*repositories.MatchStats, error) {
	var stats repositories.MatchStats
	
	// Get total matches count
	r.db.WithContext(ctx).Model(&models.Match{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&stats.TotalMatches)
	
	// Get active matches count
	r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).Count(&stats.ActiveMatches)
	
	// Get matches created today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? OR user2_id = ?) AND DATE(created_at) = DATE(?)", userID, userID, today).Count(&stats.MatchesToday)
	
	// Get matches created this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? OR user2_id = ?) AND created_at >= ?", userID, userID, weekStart).Count(&stats.MatchesThisWeek)
	
	// Get matches created this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Match{}).Where("(user1_id = ? OR user2_id = ?) AND created_at >= ?", userID, userID, monthStart).Count(&stats.MatchesThisMonth)
	
	return &stats, nil
}

// GetSwipeStats retrieves swipe statistics
func (r *MatchRepositoryImpl) GetSwipeStats(ctx context.Context, userID uuid.UUID) (*repositories.SwipeStats, error) {
	var stats repositories.SwipeStats
	
	// Get total swipes count
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ?", userID).Count(&stats.TotalSwipes)
	
	// Get total likes count
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND is_like = ?", userID, true).Count(&stats.TotalLikes)
	
	// Get total passes count
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND is_like = ?", userID, false).Count(&stats.TotalPasses)
	
	// Get swipes created today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND DATE(created_at) = DATE(?)", userID, today).Count(&stats.SwipesToday)
	
	// Get swipes created this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND created_at >= ?", userID, weekStart).Count(&stats.SwipesThisWeek)
	
	// Get swipes created this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND created_at >= ?", userID, monthStart).Count(&stats.SwipesThisMonth)
	
	// Calculate like rate
	if stats.TotalSwipes > 0 {
		stats.LikeRate = float64(stats.TotalLikes) / float64(stats.TotalSwipes) * 100
	}
	
	return &stats, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainMatch converts model Match to domain Match
func (r *MatchRepositoryImpl) modelToDomainMatch(model *models.Match) *entities.Match {
	return &entities.Match{
		ID:        model.ID,
		User1ID:   model.User1ID,
		User2ID:   model.User2ID,
		IsActive:  model.IsActive,
		MatchedAt: model.CreatedAt,
		CreatedAt:  model.CreatedAt,
	}
}

// domainToModelMatch converts domain Match to model Match
func (r *MatchRepositoryImpl) domainToModelMatch(match *entities.Match) *models.Match {
	return &models.Match{
		ID:        match.ID,
		User1ID:   match.User1ID,
		User2ID:   match.User2ID,
		IsActive:  match.IsActive,
		CreatedAt:  match.MatchedAt,
	}
}

// modelToDomainSwipe converts model Swipe to domain Swipe
func (r *MatchRepositoryImpl) modelToDomainSwipe(model *models.Swipe) *entities.Swipe {
	return &entities.Swipe{
		ID:        model.ID,
		SwiperID: model.SwiperID,
		SwipedID:  model.SwipedID,
		IsLike:    model.IsLike,
		CreatedAt: model.CreatedAt,
	}
}

// domainToModelSwipe converts domain Swipe to model Swipe
func (r *MatchRepositoryImpl) domainToModelSwipe(swipe *entities.Swipe) *models.Swipe {
	return &models.Swipe{
		ID:        swipe.ID,
		SwiperID:  swipe.SwiperID,
		SwipedID:  swipe.SwipedID,
		IsLike:    swipe.IsLike,
		CreatedAt: swipe.CreatedAt,
	}
}

// BatchCreateMatches creates multiple matches in a single transaction
func (r *MatchRepositoryImpl) BatchCreateMatches(ctx context.Context, matches []*entities.Match) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Convert to model entities
	modelMatches := make([]*models.Match, len(matches))
	for i, match := range matches {
		modelMatches[i] = r.domainToModelMatch(match)
	}

	// Batch insert
	if err := tx.CreateInBatches(modelMatches, 100).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to batch create matches", err)
		return fmt.Errorf("failed to batch create matches: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit batch create matches transaction", err)
		return fmt.Errorf("failed to commit batch create matches transaction: %w", err)
	}

	logger.Info("Batch created matches successfully", map[string]interface{}{
		"count": len(matches),
	})
	return nil
}

// BatchCreateSwipes creates multiple swipes in a single transaction
func (r *MatchRepositoryImpl) BatchCreateSwipes(ctx context.Context, swipes []*entities.Swipe) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Convert to model entities
	modelSwipes := make([]*models.Swipe, len(swipes))
	for i, swipe := range swipes {
		modelSwipes[i] = r.domainToModelSwipe(swipe)
	}

	// Batch insert
	if err := tx.CreateInBatches(modelSwipes, 100).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to batch create swipes", err)
		return fmt.Errorf("failed to batch create swipes: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit batch create swipes transaction", err)
		return fmt.Errorf("failed to commit batch create swipes transaction: %w", err)
	}

	logger.Info("Batch created swipes successfully", map[string]interface{}{
		"count": len(swipes),
	})
	return nil
}

// GetSwipe retrieves a swipe between two users
func (r *MatchRepositoryImpl) GetSwipe(ctx context.Context, swiperID, swipedID uuid.UUID) (*entities.Swipe, error) {
	var swipe models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).First(&swipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("swipe not found")
		}
		logger.Error("Failed to get swipe", err)
		return nil, fmt.Errorf("failed to get swipe: %w", err)
	}

	// Convert to domain entity
	domainSwipe := r.modelToDomainSwipe(&swipe)
	return domainSwipe, nil
}


// GetUserLikes retrieves likes for a user
func (r *MatchRepositoryImpl) GetUserLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND is_like = ?", userID, true).Order("created_at DESC").Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get user likes", err)
		return nil, fmt.Errorf("failed to get user likes: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// GetUserPasses retrieves passes for a user
func (r *MatchRepositoryImpl) GetUserPasses(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND is_like = ?", userID, false).Order("created_at DESC").Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get user passes", err)
		return nil, fmt.Errorf("failed to get user passes: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// GetLikeCount retrieves like count for a user
func (r *MatchRepositoryImpl) GetLikeCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND is_like = ?", userID, true).Count(&count).Error; err != nil {
		logger.Error("Failed to get like count", err)
		return 0, fmt.Errorf("failed to get like count: %w", err)
	}

	return count, nil
}

// HasSwiped checks if user has swiped another user
func (r *MatchRepositoryImpl) HasSwiped(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).Count(&count).Error; err != nil {
		logger.Error("Failed to check if user has swiped", err)
		return false, fmt.Errorf("failed to check if user has swiped: %w", err)
	}

	return count > 0, nil
}

// GetSwipeDirection returns the direction of a swipe (true for like, false for pass)
func (r *MatchRepositoryImpl) GetSwipeDirection(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) {
	var swipe models.Swipe
	if err := r.db.WithContext(ctx).Select("is_like").Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).First(&swipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("swipe not found")
		}
		logger.Error("Failed to get swipe direction", err)
		return false, fmt.Errorf("failed to get swipe direction: %w", err)
	}

	return swipe.IsLike, nil
}

// CheckForMatch checks if there's a match between two users
func (r *MatchRepositoryImpl) CheckForMatch(ctx context.Context, swiperID, swipedID uuid.UUID) (*entities.Match, bool, error) {
	// Check if there's a reverse swipe
	var reverseSwipe models.Swipe
	if err := r.db.WithContext(ctx).Where("swiper_id = ? AND swiped_id = ? AND is_like = ?", swipedID, swiperID, true).First(&reverseSwipe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No reverse swipe, no match
			return nil, false, nil
		}
		logger.Error("Failed to check for reverse swipe", err)
		return nil, false, fmt.Errorf("failed to check for reverse swipe: %w", err)
	}

	// Check if match already exists
	exists, err := r.SwipeExists(ctx, swiperID, swipedID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check match existence: %w", err)
	}

	if exists {
		// Match already exists
		var match models.Match
		if err := r.db.WithContext(ctx).Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", swiperID, swipedID, swipedID, swiperID).First(&match).Error; err != nil {
			logger.Error("Failed to get existing match", err)
			return nil, false, fmt.Errorf("failed to get existing match: %w", err)
		}

		domainMatch := r.modelToDomainMatch(&match)
		return domainMatch, true, nil
	}

	// Create new match
	newMatch := &entities.Match{
		User1ID:  swiperID,
		User2ID:  swipedID,
		IsActive:  true,
	}

	if err := r.CreateMatch(ctx, newMatch); err != nil {
		return nil, false, fmt.Errorf("failed to create match: %w", err)
	}

	return newMatch, true, nil
}

// GetMutualLikes retrieves users who have liked the user and whom the user has also liked
func (r *MatchRepositoryImpl) GetMutualLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	query := `
		SELECT u.* FROM users u
		INNER JOIN swipes s1 ON u.id = s1.swiped_id AND s1.swiper_id = ? AND s1.is_like = true
		INNER JOIN swipes s2 ON u.id = s2.swiper_id AND s2.swiped_id = ? AND s2.is_like = true
		ORDER BY s1.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, userID, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get mutual likes", err)
		return nil, fmt.Errorf("failed to get mutual likes: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&user)
	}

	return domainUsers, nil
}

// GetUsersWhoLikedUser retrieves users who have liked the user
func (r *MatchRepositoryImpl) GetUsersWhoLikedUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	query := `
		SELECT u.* FROM users u
		INNER JOIN swipes s ON u.id = s.swiper_id AND s.swiped_id = ? AND s.is_like = true
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get users who liked user", err)
		return nil, fmt.Errorf("failed to get users who liked user: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&user)
	}

	return domainUsers, nil
}

// GetPotentialMatches retrieves potential matches for a user
func (r *MatchRepositoryImpl) GetPotentialMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	query := `
		SELECT u.* FROM users u
		WHERE u.id != ?
		AND u.id NOT IN (
			SELECT swiped_id FROM swipes WHERE swiper_id = ?
		)
		AND u.id NOT IN (
			SELECT user2_id FROM matches WHERE user1_id = ?
		)
		AND u.id NOT IN (
			SELECT user1_id FROM matches WHERE user2_id = ?
		)
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, userID, userID, userID, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get potential matches", err)
		return nil, fmt.Errorf("failed to get potential matches: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&user)
	}

	return domainUsers, nil
}

// GetPotentialMatchesExcluding retrieves potential matches excluding specific users
func (r *MatchRepositoryImpl) GetPotentialMatchesExcluding(ctx context.Context, userID uuid.UUID, excludeUserIDs []uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	query := `
		SELECT u.* FROM users u
		WHERE u.id != ?
		AND u.id NOT IN (?)
		AND u.id NOT IN (
			SELECT swiped_id FROM swipes WHERE swiper_id = ?
		)
		AND u.id NOT IN (
			SELECT user2_id FROM matches WHERE user1_id = ?
		)
		AND u.id NOT IN (
			SELECT user1_id FROM matches WHERE user2_id = ?
		)
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, excludeUserIDs, userID, userID, userID, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get potential matches excluding", err)
		return nil, fmt.Errorf("failed to get potential matches excluding: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&user)
	}

	return domainUsers, nil
}

// GetPotentialMatchesByPreferences retrieves potential matches based on user preferences
func (r *MatchRepositoryImpl) GetPotentialMatchesByPreferences(ctx context.Context, userID uuid.UUID, preferences *entities.UserPreferences, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	query := `
		SELECT u.* FROM users u
		WHERE u.id != ?
		AND u.age BETWEEN ? AND ?
		AND u.id NOT IN (
			SELECT swiped_id FROM swipes WHERE swiper_id = ?
		)
		AND u.id NOT IN (
			SELECT user2_id FROM matches WHERE user1_id = ?
		)
		AND u.id NOT IN (
			SELECT user1_id FROM matches WHERE user2_id = ?
		)
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, preferences.AgeMin, preferences.AgeMax, userID, userID, userID, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get potential matches by preferences", err)
		return nil, fmt.Errorf("failed to get potential matches by preferences: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&user)
	}

	return domainUsers, nil
}

// SwipeExists checks if a swipe exists between two users
func (r *MatchRepositoryImpl) SwipeExists(ctx context.Context, swiperID, swipedID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ? AND swiped_id = ?", swiperID, swipedID).Count(&count).Error; err != nil {
		logger.Error("Failed to check swipe existence", err)
		return false, fmt.Errorf("failed to check swipe existence: %w", err)
	}

	return count > 0, nil
}



// GetAllMatches retrieves all matches with pagination
func (r *MatchRepositoryImpl) GetAllMatches(ctx context.Context, limit, offset int) ([]*entities.Match, error) {
	var matches []models.Match
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&matches).Error; err != nil {
		logger.Error("Failed to get all matches", err)
		return nil, fmt.Errorf("failed to get all matches: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetAllSwipes retrieves all swipes with pagination
func (r *MatchRepositoryImpl) GetAllSwipes(ctx context.Context, limit, offset int) ([]*entities.Swipe, error) {
	var swipes []models.Swipe
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		logger.Error("Failed to get all swipes", err)
		return nil, fmt.Errorf("failed to get all swipes: %w", err)
	}

	// Convert to domain entities
	domainSwipes := make([]*entities.Swipe, len(swipes))
	for i, swipe := range swipes {
		domainSwipes[i] = r.modelToDomainSwipe(&swipe)
	}

	return domainSwipes, nil
}

// GetMatchAnalytics retrieves global match analytics
func (r *MatchRepositoryImpl) GetMatchAnalytics(ctx context.Context, startDate, endDate interface{}) (*repositories.MatchAnalytics, error) {
	var analytics repositories.MatchAnalytics
	
	// Get total matches
	r.db.WithContext(ctx).Model(&models.Match{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&analytics.TotalMatches)
	
	// Get total swipes
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&analytics.TotalSwipes)
	
	// Calculate match rate
	if analytics.TotalSwipes > 0 {
		analytics.MatchRate = float64(analytics.TotalMatches) / float64(analytics.TotalSwipes) * 100
	}
	
	// Get average swipes per user (simplified calculation)
	var userCount int64
	r.db.WithContext(ctx).Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		analytics.AverageSwipesPerUser = float64(analytics.TotalSwipes) / float64(userCount)
		analytics.AverageMatchesPerUser = float64(analytics.TotalMatches) / float64(userCount)
	}
	
	return &analytics, nil
}

// GetRecentMatches retrieves recent matches for a user
func (r *MatchRepositoryImpl) GetRecentMatches(ctx context.Context, userID uuid.UUID, days int, limit int) ([]*entities.Match, error) {
	var matches []models.Match
	since := time.Now().AddDate(-days, 0, 0)
	
	if err := r.db.WithContext(ctx).Where("(user1_id = ? OR user2_id = ?) AND created_at >= ?", userID, userID, since).Order("created_at DESC").Limit(limit).Find(&matches).Error; err != nil {
		logger.Error("Failed to get recent matches", err)
		return nil, fmt.Errorf("failed to get recent matches: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetUnreadMatches retrieves unread matches for a user
func (r *MatchRepositoryImpl) GetUnreadMatches(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.Match, error) {
	var matches []models.Match
	// This is a simplified implementation - in a real app, you'd track read status
	if err := r.db.WithContext(ctx).Where("(user1_id = ? OR user2_id = ?) AND is_active = ?", userID, userID, true).Order("created_at DESC").Limit(limit).Find(&matches).Error; err != nil {
		logger.Error("Failed to get unread matches", err)
		return nil, fmt.Errorf("failed to get unread matches: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetMatchesWithoutConversation retrieves matches without conversation
func (r *MatchRepositoryImpl) GetMatchesWithoutConversation(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.Match, error) {
	var matches []models.Match
	query := `
		SELECT m.* FROM matches m
		WHERE (m.user1_id = ? OR m.user2_id = ?)
		AND m.is_active = true
		AND NOT EXISTS (
			SELECT 1 FROM conversations c WHERE c.match_id = m.id
		)
		ORDER BY m.created_at DESC
		LIMIT ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, userID, limit).Scan(&matches).Error; err != nil {
		logger.Error("Failed to get matches without conversation", err)
		return nil, fmt.Errorf("failed to get matches without conversation: %w", err)
	}

	// Convert to domain entities
	domainMatches := make([]*entities.Match, len(matches))
	for i, match := range matches {
		domainMatches[i] = r.modelToDomainMatch(&match)
	}

	return domainMatches, nil
}

// GetSwipeHistory retrieves swipe history with user details
func (r *MatchRepositoryImpl) GetSwipeHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repositories.SwipeWithUser, error) {
	var results []struct {
		models.Swipe
		models.User `gorm:"foreignKey:SwipedID"`
	}
	
	query := `
		SELECT s.*, u.* FROM swipes s
		INNER JOIN users u ON s.swiped_id = u.id
		WHERE s.swiper_id = ?
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, limit, offset).Scan(&results).Error; err != nil {
		logger.Error("Failed to get swipe history", err)
		return nil, fmt.Errorf("failed to get swipe history: %w", err)
	}

	// Convert to domain entities
	swipeHistory := make([]*repositories.SwipeWithUser, len(results))
	for i, result := range results {
		swipeHistory[i] = &repositories.SwipeWithUser{
			Swipe:      r.modelToDomainSwipe(&result.Swipe),
			SwipedUser: r.modelToDomainUser(&result.User),
		}
	}

	return swipeHistory, nil
}

// GetUserMatchesWithDetails retrieves matches with additional details
func (r *MatchRepositoryImpl) GetUserMatchesWithDetails(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repositories.MatchWithDetails, error) {
	var results []struct {
		models.Match
		OtherUser models.User `gorm:"foreignKey:User2ID"`
		LastMessage *models.Message `gorm:"foreignKey:MatchID"`
	}
	
	query := `
		SELECT
			m.*,
			CASE
				WHEN m.user1_id = ? THEN u2.id
				ELSE u1.id
			END as other_user_id,
			CASE
				WHEN m.user1_id = ? THEN u2.*
				ELSE u1.*
			END as other_user,
			msg.*
		FROM matches m
		LEFT JOIN users u1 ON m.user1_id = u1.id
		LEFT JOIN users u2 ON m.user2_id = u2.id
		LEFT JOIN conversations c ON m.id = c.match_id
		LEFT JOIN messages msg ON c.id = msg.conversation_id AND msg.id = (
			SELECT MAX(id) FROM messages WHERE conversation_id = c.id
		)
		WHERE (m.user1_id = ? OR m.user2_id = ?) AND m.is_active = true
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`
	
	if err := r.db.WithContext(ctx).Raw(query, userID, userID, userID, userID, limit, offset).Scan(&results).Error; err != nil {
		logger.Error("Failed to get matches with details", err)
		return nil, fmt.Errorf("failed to get matches with details: %w", err)
	}

	// Convert to domain entities
	matchesWithDetails := make([]*repositories.MatchWithDetails, len(results))
	for i, result := range results {
		matchWithDetails := &repositories.MatchWithDetails{
			Match:       r.modelToDomainMatch(&result.Match),
			OtherUser:   r.modelToDomainUser(&result.OtherUser),
			UnreadCount: 0, // Simplified - would need proper tracking
			HasConversation: result.LastMessage != nil,
		}
		
		if result.LastMessage != nil {
			matchWithDetails.LastMessage = r.modelToDomainMessage(result.LastMessage)
		}
		
		matchesWithDetails[i] = matchWithDetails
	}

	return matchesWithDetails, nil
}

// Helper methods for user conversion
func (r *MatchRepositoryImpl) modelToDomainUser(model *models.User) *entities.User {
	return &entities.User{
		ID:              model.ID,
		Email:           model.Email,
		PasswordHash:    model.PasswordHash,
		FirstName:       model.FirstName,
		LastName:        model.LastName,
		DateOfBirth:     model.DateOfBirth,
		Gender:          model.Gender,
		InterestedIn:    model.InterestedIn,
		Bio:             model.Bio,
		LocationLat:     model.LocationLat,
		LocationLng:     model.LocationLng,
		LocationCity:    model.LocationCity,
		LocationCountry: model.LocationCountry,
		IsVerified:      model.IsVerified,
		IsPremium:       model.IsPremium,
		IsActive:        model.IsActive,
		IsBanned:        model.IsBanned,
		LastActive:      model.LastActive,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

func (r *MatchRepositoryImpl) modelToDomainMessage(model *models.Message) *entities.Message {
	return &entities.Message{
		ID:             model.ID,
		ConversationID: model.ConversationID,
		SenderID:       model.SenderID,
		Content:        model.Content,
		MessageType:    model.MessageType,
		IsRead:         model.IsRead,
		CreatedAt:      model.CreatedAt,
	}
}

// modelToDomainUserPreferences converts model UserPreferences to domain UserPreferences
func (r *MatchRepositoryImpl) modelToDomainUserPreferences(model *models.UserPreferences) *entities.UserPreferences {
	return &entities.UserPreferences{
		UserID:           model.UserID,
		AgeMin:           model.AgeMin,
		AgeMax:           model.AgeMax,
		MaxDistance:      model.MaxDistance,
		ShowMe:           model.ShowMe,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}

// domainToModelUserPreferences converts domain UserPreferences to model UserPreferences
func (r *MatchRepositoryImpl) domainToModelUserPreferences(preferences *entities.UserPreferences) *models.UserPreferences {
	return &models.UserPreferences{
		UserID:      preferences.UserID,
		AgeMin:      preferences.AgeMin,
		AgeMax:      preferences.AgeMax,
		MaxDistance: preferences.MaxDistance,
		ShowMe:      preferences.ShowMe,
		CreatedAt:   preferences.CreatedAt,
		UpdatedAt:   preferences.UpdatedAt,
	}
}