package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/internal/domain/valueobjects"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/models"
)

// VerificationRepositoryImpl implements the VerificationRepository interface
type VerificationRepositoryImpl struct {
	db *gorm.DB
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *gorm.DB) repositories.VerificationRepository {
	return &VerificationRepositoryImpl{
		db: db,
	}
}

// CreateVerification creates a new verification
func (r *VerificationRepositoryImpl) CreateVerification(ctx context.Context, verification *entities.Verification) error {
	model := r.entityToModel(verification)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetVerificationByID gets a verification by ID
func (r *VerificationRepositoryImpl) GetVerificationByID(ctx context.Context, id uuid.UUID) (*entities.Verification, error) {
	var model models.Verification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Reviewer").
		First(&model, id).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// GetVerificationByUserAndType gets a verification by user ID and type
func (r *VerificationRepositoryImpl) GetVerificationByUserAndType(ctx context.Context, userID uuid.UUID, vType entities.VerificationType) (*entities.Verification, error) {
	var model models.Verification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Reviewer").
		Where("user_id = ? AND type = ?", userID, vType).
		Order("created_at DESC").
		First(&model).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.modelToEntity(&model), nil
}

// GetPendingVerifications gets pending verifications
func (r *VerificationRepositoryImpl) GetPendingVerifications(ctx context.Context, limit, offset int) ([]*entities.Verification, error) {
	var models []models.Verification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Reviewer").
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.modelsToEntities(models), nil
}

// GetVerificationsByUser gets verifications for a user
func (r *VerificationRepositoryImpl) GetVerificationsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Verification, error) {
	var models []models.Verification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Reviewer").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.modelsToEntities(models), nil
}

// UpdateVerification updates a verification
func (r *VerificationRepositoryImpl) UpdateVerification(ctx context.Context, verification *entities.Verification) error {
	model := r.entityToModel(verification)
	return r.db.WithContext(ctx).Save(model).Error
}

// DeleteVerification deletes a verification
func (r *VerificationRepositoryImpl) DeleteVerification(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Verification{}, id).Error
}

// GetVerificationsForReview gets verifications that need review
func (r *VerificationRepositoryImpl) GetVerificationsForReview(ctx context.Context, status valueobjects.VerificationStatus, limit, offset int) ([]*entities.Verification, error) {
	var models []models.Verification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Reviewer").
		Where("status = ?", status.String()).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.modelsToEntities(models), nil
}

// GetVerificationStats gets verification statistics
func (r *VerificationRepositoryImpl) GetVerificationStats(ctx context.Context) (*repositories.VerificationStats, error) {
	var stats repositories.VerificationStats
	
	// Count total verifications
	r.db.WithContext(ctx).Model(&models.Verification{}).Count(&stats.TotalVerifications)
	
	// Count by status
	r.db.WithContext(ctx).Model(&models.Verification{}).Where("status = ?", "pending").Count(&stats.PendingVerifications)
	r.db.WithContext(ctx).Model(&models.Verification{}).Where("status = ?", "approved").Count(&stats.ApprovedVerifications)
	r.db.WithContext(ctx).Model(&models.Verification{}).Where("status = ?", "rejected").Count(&stats.RejectedVerifications)
	
	// Count by type
	r.db.WithContext(ctx).Model(&models.Verification{}).Where("type = ?", "selfie").Count(&stats.SelfieVerifications)
	r.db.WithContext(ctx).Model(&models.Verification{}).Where("type = ?", "document").Count(&stats.DocumentVerifications)
	
	// Calculate verification rate
	if stats.TotalVerifications > 0 {
		stats.VerificationRate = float64(stats.ApprovedVerifications) / float64(stats.TotalVerifications) * 100
	}
	
	// Count verified users (users with at least one approved verification)
	r.db.WithContext(ctx).Table("users").
		Joins("JOIN verifications ON users.id = verifications.user_id").
		Where("verifications.status = ?", "approved").
		Distinct("users.id").
		Count(&stats.VerifiedUsers)
	
	return &stats, nil
}

// CreateVerificationAttempt creates a new verification attempt
func (r *VerificationRepositoryImpl) CreateVerificationAttempt(ctx context.Context, attempt *entities.VerificationAttempt) error {
	model := r.attemptEntityToModel(attempt)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetVerificationAttemptsByUser gets verification attempts for a user
func (r *VerificationRepositoryImpl) GetVerificationAttemptsByUser(ctx context.Context, userID uuid.UUID, vType entities.VerificationType, since time.Time) ([]*entities.VerificationAttempt, error) {
	var models []models.VerificationAttempt
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ? AND type = ? AND created_at >= ?", userID, vType, since).
		Order("created_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.attemptModelsToEntities(models), nil
}

// GetVerificationAttemptsByIP gets verification attempts from an IP address
func (r *VerificationRepositoryImpl) GetVerificationAttemptsByIP(ctx context.Context, ipAddress string, since time.Time) ([]*entities.VerificationAttempt, error) {
	var models []models.VerificationAttempt
	err := r.db.WithContext(ctx).
		Where("ip_address = ? AND created_at >= ?", ipAddress, since).
		Order("created_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.attemptModelsToEntities(models), nil
}

// DeleteExpiredAttempts deletes expired verification attempts
func (r *VerificationRepositoryImpl) DeleteExpiredAttempts(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&models.VerificationAttempt{}).Error
}

// CreateVerificationBadge creates a new verification badge
func (r *VerificationRepositoryImpl) CreateVerificationBadge(ctx context.Context, badge *entities.VerificationBadge) error {
	model := r.badgeEntityToModel(badge)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetActiveBadgesByUser gets active badges for a user
func (r *VerificationRepositoryImpl) GetActiveBadgesByUser(ctx context.Context, userID uuid.UUID) ([]*entities.VerificationBadge, error) {
	var models []models.VerificationBadge
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Revoker").
		Where("user_id = ? AND is_revoked = ? AND (expires_at IS NULL OR expires_at > ?)", userID, false, time.Now()).
		Order("created_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.badgeModelsToEntities(models), nil
}

// GetBadgeByUserAndType gets a badge by user and type
func (r *VerificationRepositoryImpl) GetBadgeByUserAndType(ctx context.Context, userID uuid.UUID, badgeType string) (*entities.VerificationBadge, error) {
	var model models.VerificationBadge
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Revoker").
		Where("user_id = ? AND badge_type = ?", userID, badgeType).
		Order("created_at DESC").
		First(&model).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.badgeModelToEntity(&model), nil
}

// UpdateBadge updates a verification badge
func (r *VerificationRepositoryImpl) UpdateBadge(ctx context.Context, badge *entities.VerificationBadge) error {
	model := r.badgeEntityToModel(badge)
	return r.db.WithContext(ctx).Save(model).Error
}

// RevokeBadge revokes a verification badge
func (r *VerificationRepositoryImpl) RevokeBadge(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.VerificationBadge{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_by": revokedBy,
			"revoked_at": time.Now(),
		}).Error
}

// DeleteExpiredBadges deletes expired badges
func (r *VerificationRepositoryImpl) DeleteExpiredBadges(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.VerificationBadge{}).Error
}

// GetUserVerificationLevel gets a user's verification level
func (r *VerificationRepositoryImpl) GetUserVerificationLevel(ctx context.Context, userID uuid.UUID) (entities.VerificationLevel, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Select("verification_level").
		First(&user, userID).Error
	
	if err != nil {
		return entities.VerificationLevelNone, err
	}
	
	return entities.VerificationLevel(user.VerificationLevel), nil
}

// UpdateUserVerificationLevel updates a user's verification level
func (r *VerificationRepositoryImpl) UpdateUserVerificationLevel(ctx context.Context, userID uuid.UUID, level entities.VerificationLevel) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("verification_level", int(level)).Error
}

// GetUsersByVerificationLevel gets users by verification level
func (r *VerificationRepositoryImpl) GetUsersByVerificationLevel(ctx context.Context, level entities.VerificationLevel, limit, offset int) ([]*uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("verification_level = ?", int(level)).
		Limit(limit).
		Offset(offset).
		Pluck("id", &userIDs).Error
	
	return userIDs, err
}

// GetAdminUserByEmail gets an admin user by email
func (r *VerificationRepositoryImpl) GetAdminUserByEmail(ctx context.Context, email string) (*entities.AdminUser, error) {
	var model models.AdminUser
	err := r.db.WithContext(ctx).
		First(&model, "email = ?", email).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.adminModelToEntity(&model), nil
}

// GetAdminUserByID gets an admin user by ID
func (r *VerificationRepositoryImpl) GetAdminUserByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	var model models.AdminUser
	err := r.db.WithContext(ctx).
		First(&model, id).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.adminModelToEntity(&model), nil
}

// CreateAdminUser creates a new admin user
func (r *VerificationRepositoryImpl) CreateAdminUser(ctx context.Context, admin *entities.AdminUser) error {
	model := r.adminEntityToModel(admin)
	return r.db.WithContext(ctx).Create(model).Error
}

// UpdateAdminUser updates an admin user
func (r *VerificationRepositoryImpl) UpdateAdminUser(ctx context.Context, admin *entities.AdminUser) error {
	model := r.adminEntityToModel(admin)
	return r.db.WithContext(ctx).Save(model).Error
}

// UpdateAdminLastLogin updates admin user's last login time
func (r *VerificationRepositoryImpl) UpdateAdminLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.AdminUser{}).
		Where("id = ?", id).
		Update("last_login", time.Now()).Error
}

// GetActiveAdminUsers gets all active admin users
func (r *VerificationRepositoryImpl) GetActiveAdminUsers(ctx context.Context) ([]*entities.AdminUser, error) {
	var models []models.AdminUser
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at ASC").
		Find(&models).Error
	
	if err != nil {
		return nil, err
	}
	
	return r.adminModelsToEntities(models), nil
}

// Helper methods for entity/model conversion

func (r *VerificationRepositoryImpl) entityToModel(entity *entities.Verification) *models.Verification {
	if entity == nil {
		return nil
	}
	
	model := &models.Verification{
		ID:               entity.ID,
		UserID:           entity.UserID,
		Type:             string(entity.Type),
		Status:           entity.Status.String(),
		PhotoURL:         entity.PhotoURL,
		PhotoKey:         entity.PhotoKey,
		DocumentType:     entity.DocumentType,
		DocumentData:     entity.DocumentData,
		AIScore:          entity.AIScore,
		AIDetails:        entity.AIDetails,
		RejectionReason:  entity.RejectionReason,
		ReviewedBy:       entity.ReviewedBy,
		ReviewedAt:       entity.ReviewedAt,
		ExpiresAt:        entity.ExpiresAt,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
	}
	
	return model
}

func (r *VerificationRepositoryImpl) modelToEntity(model *models.Verification) *entities.Verification {
	if model == nil {
		return nil
	}
	
	status, _ := valueobjects.NewVerificationStatus(model.Status)
	
	entity := &entities.Verification{
		ID:               model.ID,
		UserID:           model.UserID,
		Type:             entities.VerificationType(model.Type),
		Status:           status,
		PhotoURL:         model.PhotoURL,
		PhotoKey:         model.PhotoKey,
		DocumentType:     model.DocumentType,
		DocumentData:     model.DocumentData,
		AIScore:          model.AIScore,
		AIDetails:        model.AIDetails,
		RejectionReason:  model.RejectionReason,
		ReviewedBy:       model.ReviewedBy,
		ReviewedAt:       model.ReviewedAt,
		ExpiresAt:        model.ExpiresAt,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
	
	if model.User != nil {
		entity.User = r.userModelToEntity(model.User)
	}
	
	if model.Reviewer != nil {
		entity.Reviewer = r.adminModelToEntity(model.Reviewer)
	}
	
	return entity
}

func (r *VerificationRepositoryImpl) modelsToEntities(models []models.Verification) []*entities.Verification {
	entities := make([]*entities.Verification, len(models))
	for i, model := range models {
		entities[i] = r.modelToEntity(&model)
	}
	return entities
}

func (r *VerificationRepositoryImpl) attemptEntityToModel(entity *entities.VerificationAttempt) *models.VerificationAttempt {
	if entity == nil {
		return nil
	}
	
	return &models.VerificationAttempt{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Type:      string(entity.Type),
		IPAddress:  entity.IPAddress,
		UserAgent:  entity.UserAgent,
		Status:    entity.Status,
		Reason:    entity.Reason,
		CreatedAt:  entity.CreatedAt,
	}
}

func (r *VerificationRepositoryImpl) attemptModelToEntity(model *models.VerificationAttempt) *entities.VerificationAttempt {
	if model == nil {
		return nil
	}
	
	entity := &entities.VerificationAttempt{
		ID:        model.ID,
		UserID:    model.UserID,
		Type:      entities.VerificationType(model.Type),
		IPAddress:  model.IPAddress,
		UserAgent:  model.UserAgent,
		Status:    model.Status,
		Reason:    model.Reason,
		CreatedAt:  model.CreatedAt,
	}
	
	if model.User != nil {
		entity.User = r.userModelToEntity(model.User)
	}
	
	return entity
}

func (r *VerificationRepositoryImpl) attemptModelsToEntities(models []models.VerificationAttempt) []*entities.VerificationAttempt {
	entities := make([]*entities.VerificationAttempt, len(models))
	for i, model := range models {
		entities[i] = r.attemptModelToEntity(&model)
	}
	return entities
}

func (r *VerificationRepositoryImpl) badgeEntityToModel(entity *entities.VerificationBadge) *models.VerificationBadge {
	if entity == nil {
		return nil
	}
	
	return &models.VerificationBadge{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Level:     int(entity.Level),
		BadgeType: entity.BadgeType,
		ExpiresAt: entity.ExpiresAt,
		IsRevoked: entity.IsRevoked,
		RevokedAt: entity.RevokedAt,
		RevokedBy: entity.RevokedBy,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func (r *VerificationRepositoryImpl) badgeModelToEntity(model *models.VerificationBadge) *entities.VerificationBadge {
	if model == nil {
		return nil
	}
	
	entity := &entities.VerificationBadge{
		ID:        model.ID,
		UserID:    model.UserID,
		Level:     entities.VerificationLevel(model.Level),
		BadgeType: model.BadgeType,
		ExpiresAt: model.ExpiresAt,
		IsRevoked: model.IsRevoked,
		RevokedAt: model.RevokedAt,
		RevokedBy: model.RevokedBy,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	
	if model.User != nil {
		entity.User = r.userModelToEntity(model.User)
	}
	
	if model.Revoker != nil {
		entity.Revoker = r.adminModelToEntity(model.Revoker)
	}
	
	return entity
}

func (r *VerificationRepositoryImpl) badgeModelsToEntities(models []models.VerificationBadge) []*entities.VerificationBadge {
	entities := make([]*entities.VerificationBadge, len(models))
	for i, model := range models {
		entities[i] = r.badgeModelToEntity(&model)
	}
	return entities
}

func (r *VerificationRepositoryImpl) adminEntityToModel(entity *entities.AdminUser) *models.AdminUser {
	if entity == nil {
		return nil
	}
	
	return &models.AdminUser{
		ID:           entity.ID,
		Email:        entity.Email,
		PasswordHash: entity.PasswordHash,
		FirstName:    entity.FirstName,
		LastName:     entity.LastName,
		Role:         entity.Role,
		IsActive:     entity.IsActive,
		LastLogin:    entity.LastLogin,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}

func (r *VerificationRepositoryImpl) adminModelToEntity(model *models.AdminUser) *entities.AdminUser {
	if model == nil {
		return nil
	}
	
	return &entities.AdminUser{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FirstName:    model.FirstName,
		LastName:     model.LastName,
		Role:         model.Role,
		IsActive:     model.IsActive,
		LastLogin:    model.LastLogin,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func (r *VerificationRepositoryImpl) adminModelsToEntities(models []models.AdminUser) []*entities.AdminUser {
	entities := make([]*entities.AdminUser, len(models))
	for i, model := range models {
		entities[i] = r.adminModelToEntity(&model)
	}
	return entities
}

func (r *VerificationRepositoryImpl) userModelToEntity(model *models.User) *entities.User {
	if model == nil {
		return nil
	}
	
	return &entities.User{
		ID:             model.ID,
		Email:          model.Email,
		PasswordHash:   model.PasswordHash,
		FirstName:      model.FirstName,
		LastName:       model.LastName,
		DateOfBirth:    model.DateOfBirth,
		Gender:         model.Gender,
		InterestedIn:   model.InterestedIn,
		Bio:            model.Bio,
		LocationLat:    model.LocationLat,
		LocationLng:    model.LocationLng,
		LocationCity:   model.LocationCity,
		LocationCountry: model.LocationCountry,
		IsVerified:     model.IsVerified,
		IsPremium:      model.IsPremium,
		IsActive:       model.IsActive,
		IsBanned:       model.IsBanned,
		LastActive:     model.LastActive,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}