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

// EphemeralPhotoRepositoryImpl implements EphemeralPhotoRepository interface using GORM
type EphemeralPhotoRepositoryImpl struct {
	db *gorm.DB
}

// NewEphemeralPhotoRepository creates a new EphemeralPhotoRepository instance
func NewEphemeralPhotoRepository(db *gorm.DB) repositories.EphemeralPhotoRepository {
	return &EphemeralPhotoRepositoryImpl{db: db}
}

// Create creates a new ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) Create(ctx context.Context, photo *entities.EphemeralPhoto) error {
	modelPhoto := r.domainToModelEphemeralPhoto(photo)
	if err := r.db.WithContext(ctx).Create(modelPhoto).Error; err != nil {
		logger.Error("Failed to create ephemeral photo", err)
		return fmt.Errorf("failed to create ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo created successfully", map[string]interface{}{
		"photo_id":   photo.ID,
		"user_id":    photo.UserID,
		"file_key":   photo.FileKey,
		"access_key": photo.AccessKey,
	})
	return nil
}

// GetByID retrieves an ephemeral photo by ID
func (r *EphemeralPhotoRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.EphemeralPhoto, error) {
	var photo models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ephemeral photo not found")
		}
		logger.Error("Failed to get ephemeral photo by ID", err)
		return nil, fmt.Errorf("failed to get ephemeral photo by ID: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainEphemeralPhoto(&photo)
	return domainPhoto, nil
}

// GetByAccessKey retrieves an ephemeral photo by access key
func (r *EphemeralPhotoRepositoryImpl) GetByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error) {
	var photo models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("access_key = ?", accessKey).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ephemeral photo not found")
		}
		logger.Error("Failed to get ephemeral photo by access key", err)
		return nil, fmt.Errorf("failed to get ephemeral photo by access key: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainEphemeralPhoto(&photo)
	return domainPhoto, nil
}

// Update updates an ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) Update(ctx context.Context, photo *entities.EphemeralPhoto) error {
	modelPhoto := r.domainToModelEphemeralPhoto(photo)
	if err := r.db.WithContext(ctx).Save(modelPhoto).Error; err != nil {
		logger.Error("Failed to update ephemeral photo", err)
		return fmt.Errorf("failed to update ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo updated successfully", map[string]interface{}{
		"photo_id": photo.ID,
		"user_id":  photo.UserID,
	})
	return nil
}

// Delete soft deletes an ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.EphemeralPhoto{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete ephemeral photo", err)
		return fmt.Errorf("failed to delete ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo deleted successfully", map[string]interface{}{
		"photo_id": id,
	})
	return nil
}

// GetUserEphemeralPhotos retrieves ephemeral photos for a user
func (r *EphemeralPhotoRepositoryImpl) GetUserEphemeralPhotos(ctx context.Context, userID uuid.UUID, includeDeleted bool) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}
	
	if err := query.Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get user ephemeral photos", err)
		return nil, fmt.Errorf("failed to get user ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetUserActiveEphemeralPhotos retrieves active ephemeral photos for a user
func (r *EphemeralPhotoRepositoryImpl) GetUserActiveEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", 
		userID, false, false, false, time.Now()).Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get user active ephemeral photos", err)
		return nil, fmt.Errorf("failed to get user active ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetUserEphemeralPhotoCount retrieves count of user's ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) GetUserEphemeralPhotoCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND is_deleted = ?", userID, false).Count(&count).Error; err != nil {
		logger.Error("Failed to get user ephemeral photo count", err)
		return 0, fmt.Errorf("failed to get user ephemeral photo count: %w", err)
	}

	return int(count), nil
}

// GetActiveEphemeralPhotos retrieves active ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) GetActiveEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", 
		false, false, false, time.Now()).Limit(limit).Offset(offset).Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get active ephemeral photos", err)
		return nil, fmt.Errorf("failed to get active ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetExpiredEphemeralPhotos retrieves expired ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) GetExpiredEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("is_expired = ? OR expires_at < ?", true, time.Now()).Limit(limit).Offset(offset).Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get expired ephemeral photos", err)
		return nil, fmt.Errorf("failed to get expired ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetViewedEphemeralPhotos retrieves viewed ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) GetViewedEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("is_viewed = ?", true).Limit(limit).Offset(offset).Order("viewed_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get viewed ephemeral photos", err)
		return nil, fmt.Errorf("failed to get viewed ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetExpiringSoonPhotos retrieves photos expiring soon
func (r *EphemeralPhotoRepositoryImpl) GetExpiringSoonPhotos(ctx context.Context, within time.Duration, limit int) ([]*entities.EphemeralPhoto, error) {
	expireBefore := time.Now().Add(within)
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at <= ? AND expires_at > ?", 
		false, false, false, expireBefore, time.Now()).Limit(limit).Order("expires_at ASC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get expiring soon photos", err)
		return nil, fmt.Errorf("failed to get expiring soon photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// MarkAsViewed marks an ephemeral photo as viewed
func (r *EphemeralPhotoRepositoryImpl) MarkAsViewed(ctx context.Context, photoID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Updates(map[string]interface{}{
		"is_viewed": true,
		"viewed_at": &now,
	}).Error; err != nil {
		logger.Error("Failed to mark ephemeral photo as viewed", err)
		return fmt.Errorf("failed to mark ephemeral photo as viewed: %w", err)
	}

	logger.Info("Ephemeral photo marked as viewed", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// IncrementViewCount increments the view count of an ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) IncrementViewCount(ctx context.Context, photoID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Update("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		logger.Error("Failed to increment ephemeral photo view count", err)
		return fmt.Errorf("failed to increment ephemeral photo view count: %w", err)
	}

	logger.Info("Ephemeral photo view count incremented", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// GetViewCount gets the view count of an ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) GetViewCount(ctx context.Context, photoID uuid.UUID) (int, error) {
	var viewCount int
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Select("view_count").Scan(&viewCount).Error; err != nil {
		logger.Error("Failed to get ephemeral photo view count", err)
		return 0, fmt.Errorf("failed to get ephemeral photo view count: %w", err)
	}

	return viewCount, nil
}

// MarkAsExpired marks an ephemeral photo as expired
func (r *EphemeralPhotoRepositoryImpl) MarkAsExpired(ctx context.Context, photoID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Updates(map[string]interface{}{
		"is_expired": true,
		"expired_at": &now,
	}).Error; err != nil {
		logger.Error("Failed to mark ephemeral photo as expired", err)
		return fmt.Errorf("failed to mark ephemeral photo as expired: %w", err)
	}

	logger.Info("Ephemeral photo marked as expired", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// BatchMarkExpired marks multiple ephemeral photos as expired
func (r *EphemeralPhotoRepositoryImpl) BatchMarkExpired(ctx context.Context, photoIDs []uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id IN ?", photoIDs).Updates(map[string]interface{}{
		"is_expired": true,
		"expired_at": &now,
	}).Error; err != nil {
		logger.Error("Failed to batch mark ephemeral photos as expired", err)
		return fmt.Errorf("failed to batch mark ephemeral photos as expired: %w", err)
	}

	logger.Info("Ephemeral photos batch marked as expired", map[string]interface{}{
		"count": len(photoIDs),
	})
	return nil
}

// GetPhotosToExpire retrieves photos that should be expired
func (r *EphemeralPhotoRepositoryImpl) GetPhotosToExpire(ctx context.Context, before time.Time, limit int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("is_expired = ? AND expires_at < ?", false, before).Limit(limit).Find(&photos).Error; err != nil {
		logger.Error("Failed to get photos to expire", err)
		return nil, fmt.Errorf("failed to get photos to expire: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetByFileKey retrieves an ephemeral photo by file key
func (r *EphemeralPhotoRepositoryImpl) GetByFileKey(ctx context.Context, fileKey string) (*entities.EphemeralPhoto, error) {
	var photo models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("file_key = ?", fileKey).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("ephemeral photo not found")
		}
		logger.Error("Failed to get ephemeral photo by file key", err)
		return nil, fmt.Errorf("failed to get ephemeral photo by file key: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainEphemeralPhoto(&photo)
	return domainPhoto, nil
}

// UpdateFileURL updates ephemeral photo file URL
func (r *EphemeralPhotoRepositoryImpl) UpdateFileURL(ctx context.Context, photoID uuid.UUID, fileURL string) error {
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Update("file_url", fileURL).Error; err != nil {
		logger.Error("Failed to update ephemeral photo file URL", err)
		return fmt.Errorf("failed to update ephemeral photo file URL: %w", err)
	}

	logger.Info("Ephemeral photo file URL updated", map[string]interface{}{
		"photo_id": photoID,
		"file_url": fileURL,
	})
	return nil
}

// UpdateThumbnailURL updates ephemeral photo thumbnail URL
func (r *EphemeralPhotoRepositoryImpl) UpdateThumbnailURL(ctx context.Context, photoID uuid.UUID, thumbnailURL string) error {
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", photoID).Update("thumbnail_url", thumbnailURL).Error; err != nil {
		logger.Error("Failed to update ephemeral photo thumbnail URL", err)
		return fmt.Errorf("failed to update ephemeral photo thumbnail URL: %w", err)
	}

	logger.Info("Ephemeral photo thumbnail URL updated", map[string]interface{}{
		"photo_id":       photoID,
		"thumbnail_url": thumbnailURL,
	})
	return nil
}

// BatchCreate creates multiple ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) BatchCreate(ctx context.Context, photos []*entities.EphemeralPhoto) error {
	modelPhotos := make([]*models.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		modelPhotos[i] = r.domainToModelEphemeralPhoto(photo)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(modelPhotos, 100).Error; err != nil {
		logger.Error("Failed to batch create ephemeral photos", err)
		return fmt.Errorf("failed to batch create ephemeral photos: %w", err)
	}

	logger.Info("Ephemeral photos batch created successfully", map[string]interface{}{
		"count": len(photos),
	})
	return nil
}

// BatchUpdate updates multiple ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) BatchUpdate(ctx context.Context, photos []*entities.EphemeralPhoto) error {
	modelPhotos := make([]*models.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		modelPhotos[i] = r.domainToModelEphemeralPhoto(photo)
	}

	if err := r.db.WithContext(ctx).SaveInBatches(modelPhotos, 100).Error; err != nil {
		logger.Error("Failed to batch update ephemeral photos", err)
		return fmt.Errorf("failed to batch update ephemeral photos: %w", err)
	}

	logger.Info("Ephemeral photos batch updated successfully", map[string]interface{}{
		"count": len(photos),
	})
	return nil
}

// BatchDelete soft deletes multiple ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) BatchDelete(ctx context.Context, photoIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", photoIDs).Delete(&models.EphemeralPhoto{}).Error; err != nil {
		logger.Error("Failed to batch delete ephemeral photos", err)
		return fmt.Errorf("failed to batch delete ephemeral photos: %w", err)
	}

	logger.Info("Ephemeral photos batch deleted successfully", map[string]interface{}{
		"count": len(photoIDs),
	})
	return nil
}

// BatchSoftDelete soft deletes multiple ephemeral photos
func (r *EphemeralPhotoRepositoryImpl) BatchSoftDelete(ctx context.Context, photoIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id IN ?", photoIDs).Update("is_deleted", true).Error; err != nil {
		logger.Error("Failed to batch soft delete ephemeral photos", err)
		return fmt.Errorf("failed to batch soft delete ephemeral photos: %w", err)
	}

	logger.Info("Ephemeral photos batch soft deleted successfully", map[string]interface{}{
		"count": len(photoIDs),
	})
	return nil
}

// ExistsByID checks if ephemeral photo exists by ID
func (r *EphemeralPhotoRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check ephemeral photo existence", err)
		return false, fmt.Errorf("failed to check ephemeral photo existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByAccessKey checks if ephemeral photo exists by access key
func (r *EphemeralPhotoRepositoryImpl) ExistsByAccessKey(ctx context.Context, accessKey string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("access_key = ?", accessKey).Count(&count).Error; err != nil {
		logger.Error("Failed to check ephemeral photo existence by access key", err)
		return false, fmt.Errorf("failed to check ephemeral photo existence by access key: %w", err)
	}

	return count > 0, nil
}

// ExistsByFileKey checks if ephemeral photo exists by file key
func (r *EphemeralPhotoRepositoryImpl) ExistsByFileKey(ctx context.Context, fileKey string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("file_key = ?", fileKey).Count(&count).Error; err != nil {
		logger.Error("Failed to check ephemeral photo existence by file key", err)
		return false, fmt.Errorf("failed to check ephemeral photo existence by file key: %w", err)
	}

	return count > 0, nil
}

// UserHasPhoto checks if user owns an ephemeral photo
func (r *EphemeralPhotoRepositoryImpl) UserHasPhoto(ctx context.Context, userID, photoID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND id = ?", userID, photoID).Count(&count).Error; err != nil {
		logger.Error("Failed to check user ephemeral photo ownership", err)
		return false, fmt.Errorf("failed to check user ephemeral photo ownership: %w", err)
	}

	return count > 0, nil
}

// GetAllEphemeralPhotos retrieves all ephemeral photos with pagination
func (r *EphemeralPhotoRepositoryImpl) GetAllEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get all ephemeral photos", err)
		return nil, fmt.Errorf("failed to get all ephemeral photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetEphemeralPhotosByStatus retrieves ephemeral photos by status
func (r *EphemeralPhotoRepositoryImpl) GetEphemeralPhotosByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	query := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC")

	switch status {
	case "active":
		query = query.Where("is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", false, false, false, time.Now())
	case "viewed":
		query = query.Where("is_viewed = ?", true)
	case "expired":
		query = query.Where("is_expired = ? OR expires_at < ?", true, time.Now())
	case "deleted":
		query = query.Where("is_deleted = ?", true)
	default:
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	if err := query.Find(&photos).Error; err != nil {
		logger.Error("Failed to get ephemeral photos by status", err)
		return nil, fmt.Errorf("failed to get ephemeral photos by status: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetEphemeralPhotoStats retrieves ephemeral photo statistics
func (r *EphemeralPhotoRepositoryImpl) GetEphemeralPhotoStats(ctx context.Context) (*entities.EphemeralPhotoStats, error) {
	var stats entities.EphemeralPhotoStats
	
	// Get total photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Count(&stats.TotalPhotos)
	
	// Get active photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", false, false, false, time.Now()).Count(&stats.ActivePhotos)
	
	// Get viewed photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("is_viewed = ?", true).Count(&stats.ViewedPhotos)
	
	// Get expired photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("is_expired = ? OR expires_at < ?", true, time.Now()).Count(&stats.ExpiredPhotos)
	
	// Get deleted photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("is_deleted = ?", true).Count(&stats.DeletedPhotos)
	
	// Get photos uploaded today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("DATE(created_at) = DATE(?) AND is_deleted = ?", today, false).Count(&stats.PhotosToday)
	
	// Get photos uploaded this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("created_at >= ? AND is_deleted = ?", weekStart, false).Count(&stats.PhotosThisWeek)
	
	// Get photos uploaded this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("created_at >= ? AND is_deleted = ?", monthStart, false).Count(&stats.PhotosThisMonth)
	
	// Get total views
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.TotalViews)
	
	// Calculate average view time (simplified - would need view tracking table for accurate data)
	stats.AverageViewTime = 30 // Default 30 seconds
	
	return &stats, nil
}

// GetPhotosUploadedInRange retrieves count of photos uploaded in date range
func (r *EphemeralPhotoRepositoryImpl) GetPhotosUploadedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("created_at BETWEEN ? AND ? AND is_deleted = ?", startDate, endDate, false).Count(&count).Error; err != nil {
		logger.Error("Failed to get photos uploaded in range", err)
		return 0, fmt.Errorf("failed to get photos uploaded in range: %w", err)
	}

	return count, nil
}

// GetPhotosViewedInRange retrieves count of photos viewed in date range
func (r *EphemeralPhotoRepositoryImpl) GetPhotosViewedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("viewed_at BETWEEN ? AND ? AND is_viewed = ?", startDate, endDate, true).Count(&count).Error; err != nil {
		logger.Error("Failed to get photos viewed in range", err)
		return 0, fmt.Errorf("failed to get photos viewed in range: %w", err)
	}

	return count, nil
}

// GetUserEphemeralPhotoStats retrieves user's ephemeral photo statistics
func (r *EphemeralPhotoRepositoryImpl) GetUserEphemeralPhotoStats(ctx context.Context, userID uuid.UUID) (*entities.EphemeralPhotoStats, error) {
	var stats entities.EphemeralPhotoStats
	
	// Get total photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ?", userID).Count(&stats.TotalPhotos)
	
	// Get active photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", userID, false, false, false, time.Now()).Count(&stats.ActivePhotos)
	
	// Get viewed photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND is_viewed = ?", userID, true).Count(&stats.ViewedPhotos)
	
	// Get expired photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND (is_expired = ? OR expires_at < ?)", userID, true, time.Now()).Count(&stats.ExpiredPhotos)
	
	// Get deleted photos count
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND is_deleted = ?", userID, true).Count(&stats.DeletedPhotos)
	
	// Get photos uploaded today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND DATE(created_at) = DATE(?) AND is_deleted = ?", userID, today, false).Count(&stats.PhotosToday)
	
	// Get photos uploaded this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND created_at >= ? AND is_deleted = ?", userID, weekStart, false).Count(&stats.PhotosThisWeek)
	
	// Get photos uploaded this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ? AND created_at >= ? AND is_deleted = ?", userID, monthStart, false).Count(&stats.PhotosThisMonth)
	
	// Get total views
	r.db.WithContext(ctx).Model(&models.EphemeralPhoto{}).Where("user_id = ?", userID).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.TotalViews)
	
	// Calculate average view time
	stats.AverageViewTime = 30 // Default 30 seconds
	
	return &stats, nil
}

// GetPhotosForCleanup retrieves photos for cleanup
func (r *EphemeralPhotoRepositoryImpl) GetPhotosForCleanup(ctx context.Context, olderThan time.Time, limit int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("(is_expired = ? OR is_viewed = ? OR expires_at < ?) AND updated_at < ?", true, true, time.Now(), olderThan).Limit(limit).Find(&photos).Error; err != nil {
		logger.Error("Failed to get photos for cleanup", err)
		return nil, fmt.Errorf("failed to get photos for cleanup: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetActivePhotosByUser retrieves active photos for a user
func (r *EphemeralPhotoRepositoryImpl) GetActivePhotosByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_deleted = ? AND is_expired = ? AND is_viewed = ? AND expires_at > ?", 
		userID, false, false, false, time.Now()).Limit(limit).Order("created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get active photos by user", err)
		return nil, fmt.Errorf("failed to get active photos by user: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetExpiredPhotosByUser retrieves expired photos for a user
func (r *EphemeralPhotoRepositoryImpl) GetExpiredPhotosByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entities.EphemeralPhoto, error) {
	var photos []models.EphemeralPhoto
	if err := r.db.WithContext(ctx).Where("user_id = ? AND (is_expired = ? OR expires_at < ?)", userID, true, time.Now()).Limit(limit).Order("expired_at DESC, created_at DESC").Find(&photos).Error; err != nil {
		logger.Error("Failed to get expired photos by user", err)
		return nil, fmt.Errorf("failed to get expired photos by user: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.EphemeralPhoto, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainEphemeralPhoto(&photo)
	}

	return domainPhotos, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainEphemeralPhoto converts model EphemeralPhoto to domain EphemeralPhoto
func (r *EphemeralPhotoRepositoryImpl) modelToDomainEphemeralPhoto(model *models.EphemeralPhoto) *entities.EphemeralPhoto {
	return &entities.EphemeralPhoto{
		ID:            model.ID,
		UserID:        model.UserID,
		FileURL:       model.FileURL,
		FileKey:       model.FileKey,
		ThumbnailURL:  model.ThumbnailURL,
		ThumbnailKey:  model.ThumbnailKey,
		AccessKey:     model.AccessKey,
		IsViewed:      model.IsViewed,
		IsExpired:     model.IsExpired,
		ViewCount:     model.ViewCount,
		MaxViews:      model.MaxViews,
		ExpiresAt:     model.ExpiresAt,
		ViewedAt:      model.ViewedAt,
		ExpiredAt:     model.ExpiredAt,
		IsDeleted:     model.IsDeleted,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

// domainToModelEphemeralPhoto converts domain EphemeralPhoto to model EphemeralPhoto
func (r *EphemeralPhotoRepositoryImpl) domainToModelEphemeralPhoto(photo *entities.EphemeralPhoto) *models.EphemeralPhoto {
	return &models.EphemeralPhoto{
		ID:            photo.ID,
		UserID:        photo.UserID,
		FileURL:       photo.FileURL,
		FileKey:       photo.FileKey,
		ThumbnailURL:  photo.ThumbnailURL,
		ThumbnailKey:  photo.ThumbnailKey,
		AccessKey:     photo.AccessKey,
		IsViewed:      photo.IsViewed,
		IsExpired:     photo.IsExpired,
		ViewCount:     photo.ViewCount,
		MaxViews:      photo.MaxViews,
		ExpiresAt:     photo.ExpiresAt,
		ViewedAt:      photo.ViewedAt,
		ExpiredAt:     photo.ExpiredAt,
		IsDeleted:     photo.IsDeleted,
		CreatedAt:     photo.CreatedAt,
		UpdatedAt:     photo.UpdatedAt,
	}
}