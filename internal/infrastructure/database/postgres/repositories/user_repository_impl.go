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

// UserRepositoryImpl implements UserRepository interface using GORM
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("Failed to create user", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by ID", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Convert to domain entity
	domainUser := r.modelToDomainUser(&user)
	return domainUser, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by email", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Convert to domain entity
	domainUser := r.modelToDomainUser(&user)
	return domainUser, nil
}

// Update updates a user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entities.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.Error("Failed to update user", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User updated successfully", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})
	return nil
}

// Delete soft deletes a user
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete user", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User deleted successfully", map[string]interface{}{
		"user_id": id,
	})
	return nil
}

// GetByLocation retrieves users within a specified radius from a location
func (r *UserRepositoryImpl) GetByLocation(ctx context.Context, lat, lng float64, radiusKm int, limit, offset int) ([]*entities.User, error) {
	// Using PostGIS for geospatial queries would be more efficient
	// For now, we'll use a simple distance calculation
	query := `
		SELECT * FROM users 
		WHERE location_lat IS NOT NULL 
		  AND location_lng IS NOT NULL 
		  AND is_active = true 
		  AND is_banned = false
		  AND (6371 * acos(cos(radians(location_lat)) * cos(radians(?)) * cos(radians(location_lng)) + 
		               sin(radians(location_lat)) * sin(radians(?))) * 6371 * 1000) <= ?
		ORDER BY last_active DESC
		LIMIT ? OFFSET ?
	`

	var users []models.User
	if err := r.db.WithContext(ctx).Raw(query, lat, lng, float64(radiusKm), limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get users by location", err)
		return nil, fmt.Errorf("failed to get users by location: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, user := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetPotentialMatches retrieves potential matches for a user
func (r *UserRepositoryImpl) GetPotentialMatches(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	// Get user preferences first
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		logger.Error("Failed to get user for potential matches", err)
		return nil, fmt.Errorf("failed to get user for potential matches: %w", err)
	}

	// Get user preferences
	var preferences models.UserPreferences
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences).Error; err != nil {
		logger.Error("Failed to get user preferences", err)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Calculate age range
	now := time.Now()
	minBirthDate := now.AddDate(-preferences.AgeMax, 0, 0)
	maxBirthDate := now.AddDate(-preferences.AgeMin, 0, 0)

	// Query for potential matches
	query := `
		SELECT u.* FROM users u
		LEFT JOIN user_preferences up ON u.id = up.user_id
		WHERE u.id != ? 
		  AND u.is_active = true 
		  AND u.is_banned = false
		  AND u.date_of_birth BETWEEN ? AND ?
		  AND u.location_lat IS NOT NULL 
		  AND u.location_lng IS NOT NULL
		  AND up.show_me = true
		  AND u.gender = ANY(up.interested_in)
		  AND ? = ANY(up.interested_in)
		ORDER BY u.last_active DESC
		LIMIT ? OFFSET ?
	`

	var users []models.User
	if err := r.db.WithContext(ctx).Raw(query, userID, minBirthDate, maxBirthDate, user.Gender, user.Gender, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get potential matches", err)
		return nil, fmt.Errorf("failed to get potential matches: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetUsersByPreferences retrieves users based on preferences
func (r *UserRepositoryImpl) GetUsersByPreferences(ctx context.Context, userID uuid.UUID, preferences *entities.UserPreferences, limit, offset int) ([]*entities.User, error) {
	// Calculate age range
	now := time.Now()
	minBirthDate := now.AddDate(-preferences.AgeMax, 0, 0)
	maxBirthDate := now.AddDate(-preferences.AgeMin, 0, 0)

	// Query for users matching preferences
	query := `
		SELECT u.* FROM users u
		LEFT JOIN user_preferences up ON u.id = up.user_id
		WHERE u.id != ? 
		  AND u.is_active = true 
		  AND u.is_banned = false
		  AND u.date_of_birth BETWEEN ? AND ?
		  AND u.location_lat IS NOT NULL 
		  AND u.location_lng IS NOT NULL
		  AND up.show_me = true
		  AND u.gender = ANY(?)
		  AND ? = ANY(up.interested_in)
		ORDER BY u.last_active DESC
		LIMIT ? OFFSET ?
	`

	var users []models.User
	if err := r.db.WithContext(ctx).Raw(query, userID, minBirthDate, maxBirthDate, preferences.InterestedIn[0], preferences.InterestedIn, limit, offset).Scan(&users).Error; err != nil {
		logger.Error("Failed to get users by preferences", err)
		return nil, fmt.Errorf("failed to get users by preferences: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// UpdateLastActive updates the last active timestamp for a user
func (r *UserRepositoryImpl) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_active", now).Error; err != nil {
		logger.Error("Failed to update last active", err)
		return fmt.Errorf("failed to update last active: %w", err)
	}

	logger.Info("User last active updated", map[string]interface{}{
		"user_id": userID,
		"time":    now,
	})
	return nil
}

// SearchUsers searches users by name or email
func (r *UserRepositoryImpl) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	searchQuery := "%" + query + "%"
	
	var users []models.User
	if err := r.db.WithContext(ctx).Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", searchQuery, searchQuery, searchQuery).Where("is_active = ? AND is_banned = ?", true, false).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to search users", err)
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetPreferences retrieves user preferences
func (r *UserRepositoryImpl) GetPreferences(ctx context.Context, userID uuid.UUID) (*entities.UserPreferences, error) {
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

// CreatePreferences creates user preferences
func (r *UserRepositoryImpl) CreatePreferences(ctx context.Context, preferences *entities.UserPreferences) error {
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

// UpdatePreferences updates user preferences
func (r *UserRepositoryImpl) UpdatePreferences(ctx context.Context, preferences *entities.UserPreferences) error {
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

// GetUserStats retrieves user statistics
func (r *UserRepositoryImpl) GetUserStats(ctx context.Context, userID uuid.UUID) (*repositories.UserStats, error) {
	var stats repositories.UserStats
	
	// Get swipe count
	var swipeCount int64
	r.db.WithContext(ctx).Model(&models.Swipe{}).Where("swiper_id = ?", userID).Count(&swipeCount)
	
	// Get match count
	var matchCount int64
	r.db.WithContext(ctx).Model(&models.Match{}).Where("user1_id = ? OR user2_id = ?", userID, userID).Count(&matchCount)
	
	// Get message count
	var messageCount int64
	r.db.WithContext(ctx).Model(&models.Message{}).Where("sender_id = ?", userID).Count(&messageCount)
	
	// Get photo count
	var photoCount int64
	r.db.WithContext(ctx).Model(&models.Photo{}).Where("user_id = ? AND is_deleted = ?", userID, false).Count(&photoCount)
	
	// Get last active
	var user models.User
	r.db.WithContext(ctx).Where("id = ?", userID).First(&user)
	
	// Calculate stats
	stats.TotalSwipes = swipeCount
	stats.TotalMatches = matchCount
	stats.TotalMessages = messageCount
	stats.PhotosCount = photoCount
	stats.LastActiveDays = int(time.Since(*user.LastActive).Hours() / 24)
	stats.AccountAgeDays = int(time.Since(user.CreatedAt).Hours() / 24)
	
	return &stats, nil
}

// GetActiveUsersCount retrieves count of active users
func (r *UserRepositoryImpl) GetActiveUsersCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("is_active = ? AND is_banned = ?", true, false).Count(&count).Error; err != nil {
		logger.Error("Failed to get active users count", err)
		return 0, fmt.Errorf("failed to get active users count: %w", err)
	}

	return count, nil
}

// GetUsersCreatedInRange retrieves count of users created in date range
func (r *UserRepositoryImpl) GetUsersCreatedInRange(ctx context.Context, startDate, endDate interface{}) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&count).Error; err != nil {
		logger.Error("Failed to get users created in range", err)
		return 0, fmt.Errorf("failed to get users created in range: %w", err)
	}

	return count, nil
}

// GetAllUsers retrieves all users with pagination
func (r *UserRepositoryImpl) GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to get all users", err)
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetBannedUsers retrieves banned users
func (r *UserRepositoryImpl) GetBannedUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("is_banned = ?", true).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to get banned users", err)
		return nil, fmt.Errorf("failed to get banned users: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// BanUser bans a user
func (r *UserRepositoryImpl) BanUser(ctx context.Context, userID uuid.UUID, reason string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_banned", true).Error; err != nil {
		logger.Error("Failed to ban user", err)
		return fmt.Errorf("failed to ban user: %w", err)
	}

	logger.Info("User banned successfully", map[string]interface{}{
		"user_id": userID,
		"reason":  reason,
	})
	return nil
}

// UnbanUser unbans a user
func (r *UserRepositoryImpl) UnbanUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_banned", false).Error; err != nil {
		logger.Error("Failed to unban user", err)
		return fmt.Errorf("failed to unban user: %w", err)
	}

	logger.Info("User unbanned successfully", map[string]interface{}{
		"user_id": userID,
	})
	return nil
}

// VerifyUser verifies a user
func (r *UserRepositoryImpl) VerifyUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_verified", true).Error; err != nil {
		logger.Error("Failed to verify user", err)
		return fmt.Errorf("failed to verify user: %w", err)
	}

	logger.Info("User verified successfully", map[string]interface{}{
		"user_id": userID,
	})
	return nil
}

// SetPremiumStatus sets user's premium status
func (r *UserRepositoryImpl) SetPremiumStatus(ctx context.Context, userID uuid.UUID, isPremium bool) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_premium", isPremium).Error; err != nil {
		logger.Error("Failed to set premium status", err)
		return fmt.Errorf("failed to set premium status: %w", err)
	}

	logger.Info("User premium status updated", map[string]interface{}{
		"user_id":    userID,
		"is_premium": isPremium,
	})
	return nil
}

// BatchCreate creates multiple users
func (r *UserRepositoryImpl) BatchCreate(ctx context.Context, users []*entities.User) error {
	modelUsers := make([]*models.User, len(users))
	for i, user := range users {
		modelUsers[i] = r.domainToModelUser(user)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(modelUsers, 100).Error; err != nil {
		logger.Error("Failed to batch create users", err)
		return fmt.Errorf("failed to batch create users: %w", err)
	}

	logger.Info("Users batch created successfully", map[string]interface{}{
		"count": len(users),
	})
	return nil
}

// BatchUpdate updates multiple users
func (r *UserRepositoryImpl) BatchUpdate(ctx context.Context, users []*entities.User) error {
	modelUsers := make([]*models.User, len(users))
	for i, user := range users {
		modelUsers[i] = r.domainToModelUser(user)
	}

	if err := r.db.WithContext(ctx).SaveInBatches(modelUsers, 100).Error; err != nil {
		logger.Error("Failed to batch update users", err)
		return fmt.Errorf("failed to batch update users: %w", err)
	}

	logger.Info("Users batch updated successfully", map[string]interface{}{
		"count": len(users),
	})
	return nil
}

// BatchDelete soft deletes multiple users
func (r *UserRepositoryImpl) BatchDelete(ctx context.Context, userIDs []uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Delete(&models.User{}).Error; err != nil {
		logger.Error("Failed to batch delete users", err)
		return fmt.Errorf("failed to batch delete users: %w", err)
	}

	logger.Info("Users batch deleted successfully", map[string]interface{}{
		"count": len(userIDs),
	})
	return nil
}

// ExistsByEmail checks if user exists by email
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		logger.Error("Failed to check user existence by email", err)
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}

	return count > 0, nil
}

// ExistsByID checks if user exists by ID
func (r *UserRepositoryImpl) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check user existence by ID", err)
		return false, fmt.Errorf("failed to check user existence by ID: %w", err)
	}

	return count > 0, nil
}

// GetUsersWithPhotos retrieves users who have photos
func (r *UserRepositoryImpl) GetUsersWithPhotos(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("EXISTS (SELECT 1 FROM photos WHERE user_id = users.id AND is_deleted = ?)", false).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to get users with photos", err)
		return nil, fmt.Errorf("failed to get users with photos: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetUsersWithoutPhotos retrieves users who don't have photos
func (r *UserRepositoryImpl) GetUsersWithoutPhotos(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("NOT EXISTS (SELECT 1 FROM photos WHERE user_id = users.id AND is_deleted = ?)", false).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to get users without photos", err)
		return nil, fmt.Errorf("failed to get users without photos: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// GetInactiveUsers retrieves inactive users
func (r *UserRepositoryImpl) GetInactiveUsers(ctx context.Context, days int, limit, offset int) ([]*entities.User, error) {
	cutoffDate := time.Now().AddDate(-days, 0, 0)
	
	var users []models.User
	if err := r.db.WithContext(ctx).Where("last_active < ? OR last_active IS NULL", cutoffDate).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		logger.Error("Failed to get inactive users", err)
		return nil, fmt.Errorf("failed to get inactive users: %w", err)
	}

	// Convert to domain entities
	domainUsers := make([]*entities.User, len(users))
	for i, User := range users {
		domainUsers[i] = r.modelToDomainUser(&User)
	}

	return domainUsers, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainUser converts model User to domain User
func (r *UserRepositoryImpl) modelToDomainUser(model *models.User) *entities.User {
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

// domainToModelUser converts domain User to model User
func (r *UserRepositoryImpl) domainToModelUser(user *entities.User) *models.User {
	return &models.User{
		ID:             user.ID,
		Email:          user.Email,
		PasswordHash:   user.PasswordHash,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		DateOfBirth:    user.DateOfBirth,
		Gender:         user.Gender,
		InterestedIn:   user.InterestedIn,
		Bio:            user.Bio,
		LocationLat:    user.LocationLat,
		LocationLng:    user.LocationLng,
		LocationCity:   user.LocationCity,
		LocationCountry: user.LocationCountry,
		IsVerified:     user.IsVerified,
		IsPremium:      user.IsPremium,
		IsActive:       user.IsActive,
		IsBanned:       user.IsBanned,
		LastActive:     user.LastActive,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

// modelToDomainUserPreferences converts model UserPreferences to domain UserPreferences
func (r *UserRepositoryImpl) modelToDomainUserPreferences(model *models.UserPreferences) *entities.UserPreferences {
	return &entities.UserPreferences{
		ID:          model.ID,
		UserID:      model.UserID,
		AgeMin:      model.AgeMin,
		AgeMax:      model.AgeMax,
		MaxDistance: model.MaxDistance,
		ShowMe:      model.ShowMe,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// domainToModelUserPreferences converts domain UserPreferences to model UserPreferences
func (r *UserRepositoryImpl) domainToModelUserPreferences(preferences *entities.UserPreferences) *models.UserPreferences {
	return &models.UserPreferences{
		ID:          preferences.ID,
		UserID:      preferences.UserID,
		AgeMin:      preferences.AgeMin,
		AgeMax:      preferences.AgeMax,
		MaxDistance: preferences.MaxDistance,
		ShowMe:      preferences.ShowMe,
		CreatedAt:   preferences.CreatedAt,
		UpdatedAt:   preferences.UpdatedAt,
	}
}