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

// EphemeralPhotoViewRepositoryImpl implements EphemeralPhotoViewRepository interface using GORM
type EphemeralPhotoViewRepositoryImpl struct {
	db *gorm.DB
}

// NewEphemeralPhotoViewRepository creates a new EphemeralPhotoViewRepository instance
func NewEphemeralPhotoViewRepository(db *gorm.DB) repositories.EphemeralPhotoViewRepository {
	return &EphemeralPhotoViewRepositoryImpl{db: db}
}

// Create creates a new ephemeral photo view
func (r *EphemeralPhotoViewRepositoryImpl) Create(ctx context.Context, view *entities.EphemeralPhotoView) error {
	modelView := r.domainToModelEphemeralPhotoView(view)
	if err := r.db.WithContext(ctx).Create(modelView).Error; err != nil {
		logger.Error("Failed to create ephemeral photo view", err)
		return fmt.Errorf("failed to create ephemeral photo view: %w", err)
	}

	logger.Info("Ephemeral photo view created successfully", map[string]interface{}{
		"view_id":  view.ID,
		"photo_id": view.PhotoID,
		"user_id":  view.UserID,
	})
	return nil
}

// GetByID retrieves an ephemeral photo view by ID
func (r *EphemeralPhotoViewRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.EphemeralPhotoView, error) {
	var view models.EphemeralPhotoView
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&view).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ephemeral photo view not found")
		}
		logger.Error("Failed to get ephemeral photo view by ID", err)
		return nil, fmt.Errorf("failed to get ephemeral photo view by ID: %w", err)
	}

	// Convert to domain entity
	domainView := r.modelToDomainEphemeralPhotoView(&view)
	return domainView, nil
}

// Update updates an ephemeral photo view
func (r *EphemeralPhotoViewRepositoryImpl) Update(ctx context.Context, view *entities.EphemeralPhotoView) error {
	modelView := r.domainToModelEphemeralPhotoView(view)
	if err := r.db.WithContext(ctx).Save(modelView).Error; err != nil {
		logger.Error("Failed to update ephemeral photo view", err)
		return fmt.Errorf("failed to update ephemeral photo view: %w", err)
	}

	logger.Info("Ephemeral photo view updated successfully", map[string]interface{}{
		"view_id":  view.ID,
		"photo_id": view.PhotoID,
	})
	return nil
}

// Delete deletes an ephemeral photo view
func (r *EphemeralPhotoViewRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.EphemeralPhotoView{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete ephemeral photo view", err)
		return fmt.Errorf("failed to delete ephemeral photo view: %w", err)
	}

	logger.Info("Ephemeral photo view deleted successfully", map[string]interface{}{
		"view_id": id,
	})
	return nil
}

// GetViewsByPhoto retrieves views for a photo
func (r *EphemeralPhotoViewRepositoryImpl) GetViewsByPhoto(ctx context.Context, photoID uuid.UUID, limit, offset int) ([]*entities.EphemeralPhotoView, error) {
	var views []models.EphemeralPhotoView
	if err := r.db.WithContext(ctx).Where("photo_id = ?", photoID).Order("viewed_at DESC").Limit(limit).Offset(offset).Find(&views).Error; err != nil {
		logger.Error("Failed to get views by photo", err)
		return nil, fmt.Errorf("failed to get views by photo: %w", err)
	}

	// Convert to domain entities
	domainViews := make([]*entities.EphemeralPhotoView, len(views))
	for i, view := range views {
		domainViews[i] = r.modelToDomainEphemeralPhotoView(&view)
	}

	return domainViews, nil
}

// GetViewsByUser retrieves views for a user
func (r *EphemeralPhotoViewRepositoryImpl) GetViewsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.EphemeralPhotoView, error) {
	var views []models.EphemeralPhotoView
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("viewed_at DESC").Limit(limit).Offset(offset).Find(&views).Error; err != nil {
		logger.Error("Failed to get views by user", err)
		return nil, fmt.Errorf("failed to get views by user: %w", err)
	}

	// Convert to domain entities
	domainViews := make([]*entities.EphemeralPhotoView, len(views))
	for i, view := range views {
		domainViews[i] = r.modelToDomainEphemeralPhotoView(&view)
	}

	return domainViews, nil
}

// GetViewCountByPhoto retrieves view count for a photo
func (r *EphemeralPhotoViewRepositoryImpl) GetViewCountByPhoto(ctx context.Context, photoID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("photo_id = ?", photoID).Count(&count).Error; err != nil {
		logger.Error("Failed to get view count by photo", err)
		return 0, fmt.Errorf("failed to get view count by photo: %w", err)
	}

	return int(count), nil
}

// GetViewCountByUser retrieves view count for a user
func (r *EphemeralPhotoViewRepositoryImpl) GetViewCountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to get view count by user", err)
		return 0, fmt.Errorf("failed to get view count by user: %w", err)
	}

	return int(count), nil
}

// GetViewsInRange retrieves views in date range
func (r *EphemeralPhotoViewRepositoryImpl) GetViewsInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.EphemeralPhotoView, error) {
	var views []models.EphemeralPhotoView
	if err := r.db.WithContext(ctx).Where("viewed_at BETWEEN ? AND ?", startDate, endDate).Order("viewed_at DESC").Find(&views).Error; err != nil {
		logger.Error("Failed to get views in range", err)
		return nil, fmt.Errorf("failed to get views in range: %w", err)
	}

	// Convert to domain entities
	domainViews := make([]*entities.EphemeralPhotoView, len(views))
	for i, view := range views {
		domainViews[i] = r.modelToDomainEphemeralPhotoView(&view)
	}

	return domainViews, nil
}

// GetViewStats retrieves view statistics
func (r *EphemeralPhotoViewRepositoryImpl) GetViewStats(ctx context.Context) (*repositories.ViewStats, error) {
	var stats repositories.ViewStats
	
	// Get total views
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Count(&stats.TotalViews)
	
	// Get unique viewers
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Select("COUNT(DISTINCT viewer_id)").Where("viewer_id IS NOT NULL").Scan(&stats.UniqueViewers)
	
	// Get average view time
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Select("COALESCE(AVG(duration), 0)").Scan(&stats.AverageViewTime)
	
	// Get views today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("DATE(viewed_at) = DATE(?)", today).Count(&stats.ViewsToday)
	
	// Get views this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("viewed_at >= ?", weekStart).Count(&stats.ViewsThisWeek)
	
	// Get views this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("viewed_at >= ?", monthStart).Count(&stats.ViewsThisMonth)
	
	return &stats, nil
}

// GetUserViewStats retrieves view statistics for a user
func (r *EphemeralPhotoViewRepositoryImpl) GetUserViewStats(ctx context.Context, userID uuid.UUID) (*repositories.ViewStats, error) {
	var stats repositories.ViewStats
	
	// Get total views
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("user_id = ?", userID).Count(&stats.TotalViews)
	
	// Get unique viewers
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Select("COUNT(DISTINCT viewer_id)").Where("user_id = ? AND viewer_id IS NOT NULL", userID).Scan(&stats.UniqueViewers)
	
	// Get average view time
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Select("COALESCE(AVG(duration), 0)").Where("user_id = ?", userID).Scan(&stats.AverageViewTime)
	
	// Get views today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("user_id = ? AND DATE(viewed_at) = DATE(?)", userID, today).Count(&stats.ViewsToday)
	
	// Get views this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("user_id = ? AND viewed_at >= ?", userID, weekStart).Count(&stats.ViewsThisWeek)
	
	// Get views this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("user_id = ? AND viewed_at >= ?", userID, monthStart).Count(&stats.ViewsThisMonth)
	
	return &stats, nil
}

// DeleteOldViews deletes old views
func (r *EphemeralPhotoViewRepositoryImpl) DeleteOldViews(ctx context.Context, olderThan time.Time) error {
	if err := r.db.WithContext(ctx).Where("viewed_at < ?", olderThan).Delete(&models.EphemeralPhotoView{}).Error; err != nil {
		logger.Error("Failed to delete old views", err)
		return fmt.Errorf("failed to delete old views: %w", err)
	}

	logger.Info("Old views deleted successfully", map[string]interface{}{
		"older_than": olderThan,
	})
	return nil
}

// MarkViewsAsExpired marks views as expired for a photo
func (r *EphemeralPhotoViewRepositoryImpl) MarkViewsAsExpired(ctx context.Context, photoID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhotoView{}).Where("photo_id = ?", photoID).Update("is_expired", true).Error; err != nil {
		logger.Error("Failed to mark views as expired", err)
		return fmt.Errorf("failed to mark views as expired: %w", err)
	}

	logger.Info("Views marked as expired", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// BatchCreate creates multiple ephemeral photo views
func (r *EphemeralPhotoViewRepositoryImpl) BatchCreate(ctx context.Context, views []*entities.EphemeralPhotoView) error {
	modelViews := make([]*models.EphemeralPhotoView, len(views))
	for i, view := range views {
		modelViews[i] = r.domainToModelEphemeralPhotoView(view)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(modelViews, 100).Error; err != nil {
		logger.Error("Failed to batch create ephemeral photo views", err)
		return fmt.Errorf("failed to batch create ephemeral photo views: %w", err)
	}

	logger.Info("Ephemeral photo views batch created successfully", map[string]interface{}{
		"count": len(views),
	})
	return nil
}

// BatchDelete deletes multiple ephemeral photo views
func (r *EphemeralPhotoViewRepositoryImpl) BatchDelete(ctx context.Context, viewIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", viewIDs).Delete(&models.EphemeralPhotoView{}).Error; err != nil {
		logger.Error("Failed to batch delete ephemeral photo views", err)
		return fmt.Errorf("failed to batch delete ephemeral photo views: %w", err)
	}

	logger.Info("Ephemeral photo views batch deleted successfully", map[string]interface{}{
		"count": len(viewIDs),
	})
	return nil
}

// Helper methods to convert between domain and model entities

// modelToDomainEphemeralPhotoView converts model EphemeralPhotoView to domain EphemeralPhotoView
func (r *EphemeralPhotoViewRepositoryImpl) modelToDomainEphemeralPhotoView(model *models.EphemeralPhotoView) *entities.EphemeralPhotoView {
	return &entities.EphemeralPhotoView{
		ID:        model.ID,
		PhotoID:   model.PhotoID,
		UserID:    model.UserID,
		ViewerID:  model.ViewerID,
		IPAddress: model.IPAddress,
		UserAgent: model.UserAgent,
		ViewedAt:  model.ViewedAt,
		Duration:  model.Duration,
		IsExpired: model.IsExpired,
	}
}

// domainToModelEphemeralPhotoView converts domain EphemeralPhotoView to model EphemeralPhotoView
func (r *EphemeralPhotoViewRepositoryImpl) domainToModelEphemeralPhotoView(view *entities.EphemeralPhotoView) *models.EphemeralPhotoView {
	return &models.EphemeralPhotoView{
		ID:        view.ID,
		PhotoID:   view.PhotoID,
		UserID:    view.UserID,
		ViewerID:  view.ViewerID,
		IPAddress: view.IPAddress,
		UserAgent: view.UserAgent,
		ViewedAt:  view.ViewedAt,
		Duration:  view.Duration,
		IsExpired: view.IsExpired,
	}
}