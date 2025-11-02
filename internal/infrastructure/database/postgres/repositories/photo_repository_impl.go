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

// PhotoRepositoryImpl implements PhotoRepository interface using GORM
type PhotoRepositoryImpl struct {
	db *gorm.DB
}

// NewPhotoRepository creates a new PhotoRepository instance
func NewPhotoRepository(db *gorm.DB) repositories.PhotoRepository {
	return &PhotoRepositoryImpl{db: db}
}

// Create creates a new photo
func (r *PhotoRepositoryImpl) Create(ctx context.Context, photo *entities.Photo) error {
	modelPhoto := r.domainToModelPhoto(photo)
	if err := r.db.WithContext(ctx).Create(modelPhoto).Error; err != nil {
		logger.Error("Failed to create photo", err)
		return fmt.Errorf("failed to create photo: %w", err)
	}

	logger.Info("Photo created successfully", map[string]interface{}{
		"photo_id": photo.ID,
		"user_id":   photo.UserID,
		"file_key": photo.FileKey,
	})
	return nil
}

// GetByID retrieves a photo by ID
func (r *PhotoRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Photo, error) {
	var photo models.Photo
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("photo not found")
		}
		logger.Error("Failed to get photo by ID", err)
		return nil, fmt.Errorf("failed to get photo by ID: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainPhoto(&photo)
	return domainPhoto, nil
}

// Update updates a photo
func (r *PhotoRepositoryImpl) Update(ctx context.Context, photo *entities.Photo) error {
	modelPhoto := r.domainToModelPhoto(photo)
	if err := r.db.WithContext(ctx).Save(modelPhoto).Error; err != nil {
		logger.Error("Failed to update photo", err)
		return fmt.Errorf("failed to update photo: %w", err)
	}

	logger.Info("Photo updated successfully", map[string]interface{}{
		"photo_id": photo.ID,
		"user_id":   photo.UserID,
	})
	return nil
}

// Delete soft deletes a photo
func (r *PhotoRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Photo{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete photo", err)
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	logger.Info("Photo deleted successfully", map[string]interface{}{
		"photo_id": id,
	})
	return nil
}

// GetUserPhotos retrieves photos for a user
func (r *PhotoRepositoryImpl) GetUserPhotos(ctx context.Context, userID uuid.UUID, includeDeleted bool) ([]*entities.Photo, error) {
	var photos []models.Photo
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}
	
	if err := query.Find(&photos).Error; err != nil {
		logger.Error("Failed to get user photos", err)
		return nil, fmt.Errorf("failed to get user photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetUserPrimaryPhoto retrieves user's primary photo
func (r *PhotoRepositoryImpl) GetUserPrimaryPhoto(ctx context.Context, userID uuid.UUID) (*entities.Photo, error) {
	var photo models.Photo
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_primary = ? AND is_deleted = ?", userID, true, false).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("primary photo not found")
		}
		logger.Error("Failed to get user primary photo", err)
		return nil, fmt.Errorf("failed to get user primary photo: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainPhoto(&photo)
	return domainPhoto, nil
}

// GetUserPhotoCount retrieves count of user's photos
func (r *PhotoRepositoryImpl) GetUserPhotoCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("user_id = ? AND is_deleted = ?", userID, false).Count(&count).Error; err != nil {
		logger.Error("Failed to get user photo count", err)
		return 0, fmt.Errorf("failed to get user photo count: %w", err)
	}

	return int(count), nil
}

// GetPendingVerificationPhotos retrieves photos pending verification
func (r *PhotoRepositoryImpl) GetPendingVerificationPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("verification_status = ? AND is_deleted = ?", "pending", false).Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get pending verification photos", err)
		return nil, fmt.Errorf("failed to get pending verification photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetPhotosByVerificationStatus retrieves photos by verification status
func (r *PhotoRepositoryImpl) GetPhotosByVerificationStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("verification_status = ? AND is_deleted = ?", status, false).Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get photos by verification status", err)
		return nil, fmt.Errorf("failed to get photos by verification status: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// UpdateVerificationStatus updates photo verification status
func (r *PhotoRepositoryImpl) UpdateVerificationStatus(ctx context.Context, photoID uuid.UUID, status string, reason *string) error {
	updates := map[string]interface{}{
		"verification_status": status,
	}
	
	if reason != nil {
		updates["verification_reason"] = *reason
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("id = ?", photoID).Updates(updates).Error; err != nil {
		logger.Error("Failed to update photo verification status", err)
		return fmt.Errorf("failed to update photo verification status: %w", err)
	}

	logger.Info("Photo verification status updated", map[string]interface{}{
		"photo_id": photoID,
		"status":    status,
	})
	return nil
}

// SetPrimaryPhoto sets a photo as primary
func (r *PhotoRepositoryImpl) SetPrimaryPhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Unset current primary photo
	if err := tx.Model(&models.Photo{}).Where("user_id = ? AND is_primary = ?", userID, true).Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to unset current primary photo", err)
		return fmt.Errorf("failed to unset current primary photo: %w", err)
	}

	// Set new primary photo
	if err := tx.Model(&models.Photo{}).Where("id = ?", photoID).Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to set new primary photo", err)
		return fmt.Errorf("failed to set new primary photo: %w", err)
	}

	logger.Info("Primary photo set successfully", map[string]interface{}{
		"user_id":   userID,
		"photo_id": photoID,
	})
	return nil
}

// UnsetPrimaryPhoto unsets primary photo for a user
func (r *PhotoRepositoryImpl) UnsetPrimaryPhoto(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("user_id = ? AND is_primary = ?", userID, true).Update("is_primary", false).Error; err != nil {
		logger.Error("Failed to unset primary photo", err)
		return fmt.Errorf("failed to unset primary photo: %w", err)
	}

	logger.Info("Primary photo unset successfully", map[string]interface{}{
		"user_id": userID,
	})
	return nil
}

// SoftDeletePhoto soft deletes a photo
func (r *PhotoRepositoryImpl) SoftDeletePhoto(ctx context.Context, photoID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("id = ?", photoID).Update("is_deleted", true).Error; err != nil {
		logger.Error("Failed to soft delete photo", err)
		return fmt.Errorf("failed to soft delete photo: %w", err)
	}

	logger.Info("Photo soft deleted successfully", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// RestorePhoto restores a soft-deleted photo
func (r *PhotoRepositoryImpl) RestorePhoto(ctx context.Context, photoID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("id = ?", photoID).Update("is_deleted", false).Error; err != nil {
		logger.Error("Failed to restore photo", err)
		return fmt.Errorf("failed to restore photo: %w", err)
	}

	logger.Info("Photo restored successfully", map[string]interface{}{
		"photo_id": photoID,
	})
	return nil
}

// GetByFileKey retrieves a photo by file key
func (r *PhotoRepositoryImpl) GetByFileKey(ctx context.Context, fileKey string) (*entities.Photo, error) {
	var photo models.Photo
	if err := r.db.WithContext(ctx).Where("file_key = ?", fileKey).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("photo not found")
		}
		logger.Error("Failed to get photo by file key", err)
		return nil, fmt.Errorf("failed to get photo by file key: %w", err)
	}

	// Convert to domain entity
	domainPhoto := r.modelToDomainPhoto(&photo)
	return domainPhoto, nil
}

// UpdateFileURL updates photo file URL
func (r *PhotoRepositoryImpl) UpdateFileURL(ctx context.Context, photoID uuid.UUID, fileURL string) error {
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("id = ?", photoID).Update("file_url", fileURL).Error; err != nil {
		logger.Error("Failed to update photo file URL", err)
		return fmt.Errorf("failed to update photo file URL: %w", err)
	}

	logger.Info("Photo file URL updated", map[string]interface{}{
		"photo_id": photoID,
		"file_url": fileURL,
	})
	return nil
}

// BatchCreate creates multiple photos
func (r *PhotoRepositoryImpl) BatchCreate(ctx context.Context, photos []*entities.Photo) error {
	modelPhotos := make([]*models.Photo, len(photos))
	for i, photo := range photos {
		modelPhotos[i] = r.domainToModelPhoto(photo)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(modelPhotos, 100).Error; err != nil {
		logger.Error("Failed to batch create photos", err)
		return fmt.Errorf("failed to batch create photos: %w", err)
	}

	logger.Info("Photos batch created successfully", map[string]interface{}{
		"count": len(photos),
	})
	return nil
}

// BatchUpdate updates multiple photos
func (r *PhotoRepositoryImpl) BatchUpdate(ctx context.Context, photos []*entities.Photo) error {
	modelPhotos := make([]*models.Photo, len(photos))
	for i, photo := range photos {
		modelPhotos[i] = r.domainToModelPhoto(photo)
	}

	if err := r.db.WithContext(ctx).SaveInBatches(modelPhotos, 100).Error; err != nil {
		logger.Error("Failed to batch update photos", err)
		return fmt.Errorf("failed to batch update photos: %w", err)
	}

	logger.Info("Photos batch updated successfully", map[string]interface{}{
		"count": len(photos),
	})
	return nil
}

// BatchDelete soft deletes multiple photos
func (r *PhotoRepositoryImpl) BatchDelete(ctx context.Context, photoIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", photoIDs).Delete(&models.Photo{}).Error; err != nil {
		logger.Error("Failed to batch delete photos", err)
		return fmt.Errorf("failed to batch delete photos: %w", err)
	}

	logger.Info("Photos batch deleted successfully", map[string]interface{}{
		"count": len(photoIDs),
	})
	return nil
}

// ExistsByID checks if photo exists by ID
func (r *PhotoRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check photo existence", err)
		return false, fmt.Errorf("failed to check photo existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByFileKey checks if photo exists by file key
func (r *PhotoRepositoryImpl) ExistsByFileKey(ctx context.Context, fileKey string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("file_key = ?", fileKey).Count(&count).Error; err != nil {
		logger.Error("Failed to check photo existence by file key", err)
		return false, fmt.Errorf("failed to check photo existence by file key: %w", err)
	}

	return count > 0, nil
}

// UserHasPhoto checks if user owns a photo
func (r *PhotoRepositoryImpl) UserHasPhoto(ctx context.Context, userID, photoID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("user_id = ? AND id = ?", userID, photoID).Count(&count).Error; err != nil {
		logger.Error("Failed to check user photo ownership", err)
		return false, fmt.Errorf("failed to check user photo ownership: %w", err)
	}

	return count > 0, nil
}

// GetAllPhotos retrieves all photos with pagination
func (r *PhotoRepositoryImpl) GetAllPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get all photos", err)
		return nil, fmt.Errorf("failed to get all photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetRejectedPhotos retrieves rejected photos
func (r *PhotoRepositoryImpl) GetRejectedPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("verification_status = ? AND is_deleted = ?", "rejected", false).Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get rejected photos", err)
		return nil, fmt.Errorf("failed to get rejected photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetApprovedPhotos retrieves approved photos
func (r *PhotoRepositoryImpl) GetApprovedPhotos(ctx context.Context, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("verification_status = ? AND is_deleted = ?", "approved", false).Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get approved photos", err)
		return nil, fmt.Errorf("failed to get approved photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetPhotosForVerification retrieves photos for verification queue
func (r *PhotoRepositoryImpl) GetPhotosForVerification(ctx context.Context, limit int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("verification_status = ? AND is_deleted = ?", "pending", false).Limit(limit).Find(&photos).Error; err != nil {
		logger.Error("Failed to get photos for verification", err)
		return nil, fmt.Errorf("failed to get photos for verification: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetUserVerifiedPhotos retrieves user's verified photos
func (r *PhotoRepositoryImpl) GetUserVerifiedPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("user_id = ? AND verification_status = ? AND is_deleted = ?", userID, "approved", false).Find(&photos).Error; err != nil {
		logger.Error("Failed to get user verified photos", err)
		return nil, fmt.Errorf("failed to get user verified photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetUserUnverifiedPhotos retrieves user's unverified photos
func (r *PhotoRepositoryImpl) GetUserUnverifiedPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("user_id = ? AND verification_status IN ? AND is_deleted = ?", userID, []string{"pending", "rejected"}, false).Find(&photos).Error; err != nil {
		logger.Error("Failed to get user unverified photos", err)
		return nil, fmt.Errorf("failed to get user unverified photos: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetPhotosByUser retrieves photos for a user with pagination
func (r *PhotoRepositoryImpl) GetPhotosByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Photo, error) {
	var photos []models.Photo
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_deleted = ?", userID, false).Order("created_at DESC").Limit(limit).Offset(offset).Find(&photos).Error; err != nil {
		logger.Error("Failed to get photos by user", err)
		return nil, fmt.Errorf("failed to get photos by user: %w", err)
	}

	// Convert to domain entities
	domainPhotos := make([]*entities.Photo, len(photos))
	for i, photo := range photos {
		domainPhotos[i] = r.modelToDomainPhoto(&photo)
	}

	return domainPhotos, nil
}

// GetPhotoStats retrieves photo statistics
func (r *PhotoRepositoryImpl) GetPhotoStats(ctx context.Context) (*repositories.PhotoStats, error) {
	var stats repositories.PhotoStats
	
	// Get total photos count
	r.db.WithContext(ctx).Model(&models.Photo{}).Count(&stats.TotalPhotos)
	
	// Get pending photos count
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("verification_status = ? AND is_deleted = ?", "pending", false).Count(&stats.PendingPhotos)
	
	// Get approved photos count
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("verification_status = ? AND is_deleted = ?", "approved", false).Count(&stats.ApprovedPhotos)
	
	// Get rejected photos count
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("verification_status = ? AND is_deleted = ?", "rejected", false).Count(&stats.RejectedPhotos)
	
	// Get photos uploaded today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("DATE(created_at) = DATE(?) AND is_deleted = ?", today, false).Count(&stats.PhotosToday)
	
	// Get photos uploaded this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("created_at >= ? AND is_deleted = ?", weekStart, false).Count(&stats.PhotosThisWeek)
	
	// Get photos uploaded this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("created_at >= ? AND is_deleted = ?", monthStart, false).Count(&stats.PhotosThisMonth)
	
	return &stats, nil
}

// GetPhotosUploadedInRange retrieves count of photos uploaded in date range
func (r *PhotoRepositoryImpl) GetPhotosUploadedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Photo{}).Where("created_at BETWEEN ? AND ? AND is_deleted = ?", startDate, endDate, false).Count(&count).Error; err != nil {
		logger.Error("Failed to get photos uploaded in range", err)
		return 0, fmt.Errorf("failed to get photos uploaded in range: %w", err)
	}

	return count, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainPhoto converts model Photo to domain Photo
func (r *PhotoRepositoryImpl) modelToDomainPhoto(model *models.Photo) *entities.Photo {
	return &entities.Photo{
		ID:                model.ID,
		UserID:            model.UserID,
		FileURL:           model.FileURL,
		FileKey:           model.FileKey,
		IsPrimary:         model.IsPrimary,
		VerificationStatus: model.VerificationStatus,
		VerificationReason: model.VerificationReason,
		IsDeleted:         model.IsDeleted,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// domainToModelPhoto converts domain Photo to model Photo
func (r *PhotoRepositoryImpl) domainToModelPhoto(photo *entities.Photo) *models.Photo {
	return &models.Photo{
		ID:                photo.ID,
		UserID:            photo.UserID,
		FileURL:           photo.FileURL,
		FileKey:           photo.FileKey,
		IsPrimary:         photo.IsPrimary,
		VerificationStatus: photo.VerificationStatus,
		VerificationReason: photo.VerificationReason,
		IsDeleted:         photo.IsDeleted,
		CreatedAt:         photo.CreatedAt,
		UpdatedAt:         photo.UpdatedAt,
	}
}