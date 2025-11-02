package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// EphemeralPhotoService defines the interface for ephemeral photo business logic
type EphemeralPhotoService interface {
	// Upload and management
	UploadEphemeralPhoto(ctx context.Context, userID uuid.UUID, fileURL, fileKey, thumbnailURL, thumbnailKey string, maxViews int, expiresAfter time.Duration) (*entities.EphemeralPhoto, error)
	GetEphemeralPhoto(ctx context.Context, photoID uuid.UUID) (*entities.EphemeralPhoto, error)
	GetEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error)
	ViewEphemeralPhoto(ctx context.Context, accessKey string, viewerID *uuid.UUID, ipAddress, userAgent string) (*entities.EphemeralPhoto, error)
	DeleteEphemeralPhoto(ctx context.Context, userID, photoID uuid.UUID) error
	ExpireEphemeralPhoto(ctx context.Context, userID, photoID uuid.UUID) error
	
	// User operations
	GetUserEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error)
	GetUserActiveEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error)
	GetUserEphemeralPhotoStats(ctx context.Context, userID uuid.UUID) (*entities.EphemeralPhotoStats, error)
	
	// System operations
	GetActiveEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetExpiredEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error)
	GetExpiringSoonPhotos(ctx context.Context, within time.Duration, limit int) ([]*entities.EphemeralPhoto, error)
	GetEphemeralPhotoStats(ctx context.Context) (*entities.EphemeralPhotoStats, error)
	
	// Cleanup operations
	CleanupExpiredPhotos(ctx context.Context, olderThan time.Duration) (int, error)
	CleanupViewedPhotos(ctx context.Context, olderThan time.Duration) (int, error)
	
	// Analytics
	TrackPhotoView(ctx context.Context, photoID, userID uuid.UUID, viewerID *uuid.UUID, ipAddress, userAgent string, duration int) error
	GetPhotoViewStats(ctx context.Context, photoID uuid.UUID) (*ViewStats, error)
}

// EphemeralPhotoServiceImpl implements EphemeralPhotoService
type EphemeralPhotoServiceImpl struct {
	ephemeralPhotoRepo repositories.EphemeralPhotoRepository
	viewRepo           repositories.EphemeralPhotoViewRepository
	userRepo           repositories.UserRepository
}

// NewEphemeralPhotoService creates a new ephemeral photo service
func NewEphemeralPhotoService(
	ephemeralPhotoRepo repositories.EphemeralPhotoRepository,
	viewRepo repositories.EphemeralPhotoViewRepository,
	userRepo repositories.UserRepository,
) EphemeralPhotoService {
	return &EphemeralPhotoServiceImpl{
		ephemeralPhotoRepo: ephemeralPhotoRepo,
		viewRepo:           viewRepo,
		userRepo:           userRepo,
	}
}

// UploadEphemeralPhoto uploads a new ephemeral photo
func (s *EphemeralPhotoServiceImpl) UploadEphemeralPhoto(ctx context.Context, userID uuid.UUID, fileURL, fileKey, thumbnailURL, thumbnailKey string, maxViews int, expiresAfter time.Duration) (*entities.EphemeralPhoto, error) {
	// Validate user exists
	userExists, err := s.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to check user existence", err)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return nil, fmt.Errorf("user not found")
	}

	// Generate unique access key
	accessKey, err := s.generateAccessKey()
	if err != nil {
		logger.Error("Failed to generate access key", err)
		return nil, fmt.Errorf("failed to generate access key: %w", err)
	}

	// Set default values
	if maxViews <= 0 {
		maxViews = 1
	}
	if expiresAfter <= 0 {
		expiresAfter = 30 * time.Second // Default 30 seconds
	}

	// Create ephemeral photo
	photo := &entities.EphemeralPhoto{
		UserID:       userID,
		FileURL:      fileURL,
		FileKey:      fileKey,
		ThumbnailURL: thumbnailURL,
		ThumbnailKey: thumbnailKey,
		AccessKey:    accessKey,
		MaxViews:     maxViews,
		ExpiresAt:    time.Now().Add(expiresAfter),
	}

	if err := s.ephemeralPhotoRepo.Create(ctx, photo); err != nil {
		logger.Error("Failed to create ephemeral photo", err)
		return nil, fmt.Errorf("failed to create ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo uploaded successfully", map[string]interface{}{
		"photo_id":    photo.ID,
		"user_id":     userID,
		"access_key":  accessKey,
		"max_views":   maxViews,
		"expires_at":   photo.ExpiresAt,
	})

	return photo, nil
}

// GetEphemeralPhoto retrieves an ephemeral photo by ID
func (s *EphemeralPhotoServiceImpl) GetEphemeralPhoto(ctx context.Context, photoID uuid.UUID) (*entities.EphemeralPhoto, error) {
	photo, err := s.ephemeralPhotoRepo.GetByID(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get ephemeral photo", err)
		return nil, fmt.Errorf("failed to get ephemeral photo: %w", err)
	}

	return photo, nil
}

// GetEphemeralPhotoByAccessKey retrieves an ephemeral photo by access key
func (s *EphemeralPhotoServiceImpl) GetEphemeralPhotoByAccessKey(ctx context.Context, accessKey string) (*entities.EphemeralPhoto, error) {
	photo, err := s.ephemeralPhotoRepo.GetByAccessKey(ctx, accessKey)
	if err != nil {
		logger.Error("Failed to get ephemeral photo by access key", err)
		return nil, fmt.Errorf("failed to get ephemeral photo by access key: %w", err)
	}

	return photo, nil
}

// ViewEphemeralPhoto handles viewing an ephemeral photo
func (s *EphemeralPhotoServiceImpl) ViewEphemeralPhoto(ctx context.Context, accessKey string, viewerID *uuid.UUID, ipAddress, userAgent string) (*entities.EphemeralPhoto, error) {
	// Get photo by access key
	photo, err := s.ephemeralPhotoRepo.GetByAccessKey(ctx, accessKey)
	if err != nil {
		logger.Error("Failed to get ephemeral photo for viewing", err)
		return nil, fmt.Errorf("failed to get ephemeral photo for viewing: %w", err)
	}

	// Check if photo can be viewed
	if !photo.CanBeViewed() {
		status := photo.GetViewStatus()
		logger.Warn("Ephemeral photo cannot be viewed", map[string]interface{}{
			"photo_id": photo.ID,
			"status":   status,
		})
		return nil, fmt.Errorf("photo cannot be viewed: %s", status)
	}

	// Check if expired by time
	if photo.IsExpiredByTime() {
		// Mark as expired if not already marked
		if !photo.IsExpired {
			if err := s.ephemeralPhotoRepo.MarkAsExpired(ctx, photo.ID); err != nil {
				logger.Error("Failed to mark photo as expired", err)
			}
		}
		return nil, fmt.Errorf("photo has expired")
	}

	// Track view
	view := &entities.EphemeralPhotoView{
		PhotoID:   photo.ID,
		UserID:    photo.UserID,
		ViewerID:  viewerID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  0, // Will be updated later
	}

	if err := s.viewRepo.Create(ctx, view); err != nil {
		logger.Error("Failed to track photo view", err)
		// Continue anyway - don't fail the view
	}

	// Mark as viewed if this is the first view
	if !photo.IsViewed {
		if err := s.ephemeralPhotoRepo.MarkAsViewed(ctx, photo.ID); err != nil {
			logger.Error("Failed to mark photo as viewed", err)
		}
	}

	// Increment view count
	if err := s.ephemeralPhotoRepo.IncrementViewCount(ctx, photo.ID); err != nil {
		logger.Error("Failed to increment view count", err)
	}

	// Check if max views reached
	viewCount, err := s.ephemeralPhotoRepo.GetViewCount(ctx, photo.ID)
	if err != nil {
		logger.Error("Failed to get view count", err)
	} else if viewCount >= photo.MaxViews {
		// Mark as expired if max views reached
		if err := s.ephemeralPhotoRepo.MarkAsExpired(ctx, photo.ID); err != nil {
			logger.Error("Failed to mark photo as expired after max views", err)
		}
	}

	// Refresh photo data
	updatedPhoto, err := s.ephemeralPhotoRepo.GetByID(ctx, photo.ID)
	if err != nil {
		logger.Error("Failed to refresh photo data", err)
		return photo, nil // Return original photo
	}

	logger.Info("Ephemeral photo viewed successfully", map[string]interface{}{
		"photo_id":   photo.ID,
		"access_key":  accessKey,
		"viewer_id":   viewerID,
		"ip_address": ipAddress,
		"view_count":  updatedPhoto.ViewCount,
	})

	return updatedPhoto, nil
}

// DeleteEphemeralPhoto deletes an ephemeral photo
func (s *EphemeralPhotoServiceImpl) DeleteEphemeralPhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	// Check if user owns the photo
	userHasPhoto, err := s.ephemeralPhotoRepo.UserHasPhoto(ctx, userID, photoID)
	if err != nil {
		logger.Error("Failed to check photo ownership", err)
		return fmt.Errorf("failed to check photo ownership: %w", err)
	}
	if !userHasPhoto {
		return fmt.Errorf("user does not own this photo")
	}

	// Delete the photo
	if err := s.ephemeralPhotoRepo.Delete(ctx, photoID); err != nil {
		logger.Error("Failed to delete ephemeral photo", err)
		return fmt.Errorf("failed to delete ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo deleted successfully", map[string]interface{}{
		"photo_id": photoID,
		"user_id":  userID,
	})

	return nil
}

// ExpireEphemeralPhoto manually expires an ephemeral photo
func (s *EphemeralPhotoServiceImpl) ExpireEphemeralPhoto(ctx context.Context, userID, photoID uuid.UUID) error {
	// Check if user owns the photo
	userHasPhoto, err := s.ephemeralPhotoRepo.UserHasPhoto(ctx, userID, photoID)
	if err != nil {
		logger.Error("Failed to check photo ownership", err)
		return fmt.Errorf("failed to check photo ownership: %w", err)
	}
	if !userHasPhoto {
		return fmt.Errorf("user does not own this photo")
	}

	// Mark as expired
	if err := s.ephemeralPhotoRepo.MarkAsExpired(ctx, photoID); err != nil {
		logger.Error("Failed to expire ephemeral photo", err)
		return fmt.Errorf("failed to expire ephemeral photo: %w", err)
	}

	logger.Info("Ephemeral photo expired successfully", map[string]interface{}{
		"photo_id": photoID,
		"user_id":  userID,
	})

	return nil
}

// GetUserEphemeralPhotos retrieves user's ephemeral photos
func (s *EphemeralPhotoServiceImpl) GetUserEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error) {
	photos, err := s.ephemeralPhotoRepo.GetUserEphemeralPhotos(ctx, userID, false)
	if err != nil {
		logger.Error("Failed to get user ephemeral photos", err)
		return nil, fmt.Errorf("failed to get user ephemeral photos: %w", err)
	}

	return photos, nil
}

// GetUserActiveEphemeralPhotos retrieves user's active ephemeral photos
func (s *EphemeralPhotoServiceImpl) GetUserActiveEphemeralPhotos(ctx context.Context, userID uuid.UUID) ([]*entities.EphemeralPhoto, error) {
	photos, err := s.ephemeralPhotoRepo.GetUserActiveEphemeralPhotos(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user active ephemeral photos", err)
		return nil, fmt.Errorf("failed to get user active ephemeral photos: %w", err)
	}

	return photos, nil
}

// GetUserEphemeralPhotoStats retrieves user's ephemeral photo statistics
func (s *EphemeralPhotoServiceImpl) GetUserEphemeralPhotoStats(ctx context.Context, userID uuid.UUID) (*entities.EphemeralPhotoStats, error) {
	stats, err := s.ephemeralPhotoRepo.GetUserEphemeralPhotoStats(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user ephemeral photo stats", err)
		return nil, fmt.Errorf("failed to get user ephemeral photo stats: %w", err)
	}

	return stats, nil
}

// GetActiveEphemeralPhotos retrieves active ephemeral photos
func (s *EphemeralPhotoServiceImpl) GetActiveEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	photos, err := s.ephemeralPhotoRepo.GetActiveEphemeralPhotos(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get active ephemeral photos", err)
		return nil, fmt.Errorf("failed to get active ephemeral photos: %w", err)
	}

	return photos, nil
}

// GetExpiredEphemeralPhotos retrieves expired ephemeral photos
func (s *EphemeralPhotoServiceImpl) GetExpiredEphemeralPhotos(ctx context.Context, limit, offset int) ([]*entities.EphemeralPhoto, error) {
	photos, err := s.ephemeralPhotoRepo.GetExpiredEphemeralPhotos(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get expired ephemeral photos", err)
		return nil, fmt.Errorf("failed to get expired ephemeral photos: %w", err)
	}

	return photos, nil
}

// GetExpiringSoonPhotos retrieves photos expiring soon
func (s *EphemeralPhotoServiceImpl) GetExpiringSoonPhotos(ctx context.Context, within time.Duration, limit int) ([]*entities.EphemeralPhoto, error) {
	photos, err := s.ephemeralPhotoRepo.GetExpiringSoonPhotos(ctx, within, limit)
	if err != nil {
		logger.Error("Failed to get expiring soon photos", err)
		return nil, fmt.Errorf("failed to get expiring soon photos: %w", err)
	}

	return photos, nil
}

// GetEphemeralPhotoStats retrieves ephemeral photo statistics
func (s *EphemeralPhotoServiceImpl) GetEphemeralPhotoStats(ctx context.Context) (*entities.EphemeralPhotoStats, error) {
	stats, err := s.ephemeralPhotoRepo.GetEphemeralPhotoStats(ctx)
	if err != nil {
		logger.Error("Failed to get ephemeral photo stats", err)
		return nil, fmt.Errorf("failed to get ephemeral photo stats: %w", err)
	}

	return stats, nil
}

// CleanupExpiredPhotos cleans up expired photos older than specified duration
func (s *EphemeralPhotoServiceImpl) CleanupExpiredPhotos(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)
	photos, err := s.ephemeralPhotoRepo.GetPhotosForCleanup(ctx, cutoffTime, 100)
	if err != nil {
		logger.Error("Failed to get photos for cleanup", err)
		return 0, fmt.Errorf("failed to get photos for cleanup: %w", err)
	}

	if len(photos) == 0 {
		return 0, nil
	}

	photoIDs := make([]uuid.UUID, len(photos))
	for i, photo := range photos {
		photoIDs[i] = photo.ID
	}

	if err := s.ephemeralPhotoRepo.BatchSoftDelete(ctx, photoIDs); err != nil {
		logger.Error("Failed to batch delete expired photos", err)
		return 0, fmt.Errorf("failed to batch delete expired photos: %w", err)
	}

	logger.Info("Expired photos cleaned up successfully", map[string]interface{}{
		"count":      len(photos),
		"older_than": olderThan,
	})

	return len(photos), nil
}

// CleanupViewedPhotos cleans up viewed photos older than specified duration
func (s *EphemeralPhotoServiceImpl) CleanupViewedPhotos(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)
	photos, err := s.ephemeralPhotoRepo.GetPhotosForCleanup(ctx, cutoffTime, 100)
	if err != nil {
		logger.Error("Failed to get viewed photos for cleanup", err)
		return 0, fmt.Errorf("failed to get viewed photos for cleanup: %w", err)
	}

	// Filter only viewed photos
	viewedPhotos := make([]*entities.EphemeralPhoto, 0)
	photoIDs := make([]uuid.UUID, 0)
	for _, photo := range photos {
		if photo.IsViewed {
			viewedPhotos = append(viewedPhotos, photo)
			photoIDs = append(photoIDs, photo.ID)
		}
	}

	if len(viewedPhotos) == 0 {
		return 0, nil
	}

	if err := s.ephemeralPhotoRepo.BatchSoftDelete(ctx, photoIDs); err != nil {
		logger.Error("Failed to batch delete viewed photos", err)
		return 0, fmt.Errorf("failed to batch delete viewed photos: %w", err)
	}

	logger.Info("Viewed photos cleaned up successfully", map[string]interface{}{
		"count":      len(viewedPhotos),
		"older_than": olderThan,
	})

	return len(viewedPhotos), nil
}

// TrackPhotoView tracks a photo view with duration
func (s *EphemeralPhotoServiceImpl) TrackPhotoView(ctx context.Context, photoID, userID uuid.UUID, viewerID *uuid.UUID, ipAddress, userAgent string, duration int) error {
	// Create view record
	view := &entities.EphemeralPhotoView{
		PhotoID:   photoID,
		UserID:    userID,
		ViewerID:  viewerID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Duration:  duration,
	}

	if err := s.viewRepo.Create(ctx, view); err != nil {
		logger.Error("Failed to track photo view", err)
		return fmt.Errorf("failed to track photo view: %w", err)
	}

	logger.Info("Photo view tracked successfully", map[string]interface{}{
		"photo_id":   photoID,
		"user_id":    userID,
		"viewer_id":  viewerID,
		"duration":   duration,
	})

	return nil
}

// GetPhotoViewStats retrieves view statistics for a photo
func (s *EphemeralPhotoServiceImpl) GetPhotoViewStats(ctx context.Context, photoID uuid.UUID) (*ViewStats, error) {
	// Get view count
	viewCount, err := s.viewRepo.GetViewCountByPhoto(ctx, photoID)
	if err != nil {
		logger.Error("Failed to get photo view count", err)
		return nil, fmt.Errorf("failed to get photo view count: %w", err)
	}

	// Get views for detailed stats
	views, err := s.viewRepo.GetViewsByPhoto(ctx, photoID, 100, 0)
	if err != nil {
		logger.Error("Failed to get photo views", err)
		return nil, fmt.Errorf("failed to get photo views: %w", err)
	}

	// Calculate stats
	stats := &ViewStats{
		TotalViews:     int64(viewCount),
		UniqueViewers:   0,
		AverageViewTime: 0,
		ViewsToday:      0,
		ViewsThisWeek:   0,
		ViewsThisMonth:  0,
	}

	if len(views) > 0 {
		// Count unique viewers
		viewerSet := make(map[uuid.UUID]bool)
		totalDuration := 0
		today := time.Now().Truncate(24 * time.Hour)
		weekStart := time.Now().AddDate(-7, 0, 0)
		monthStart := time.Now().AddDate(-30, 0, 0)

		for _, view := range views {
			if view.ViewerID != nil {
				viewerSet[*view.ViewerID] = true
			}
			totalDuration += view.Duration

			// Count by time periods
			if view.ViewedAt.After(today) {
				stats.ViewsToday++
			}
			if view.ViewedAt.After(weekStart) {
				stats.ViewsThisWeek++
			}
			if view.ViewedAt.After(monthStart) {
				stats.ViewsThisMonth++
			}
		}

		stats.UniqueViewers = int64(len(viewerSet))
		if len(views) > 0 {
			stats.AverageViewTime = int64(totalDuration / len(views))
		}
	}

	return stats, nil
}

// generateAccessKey generates a unique access key
func (s *EphemeralPhotoServiceImpl) generateAccessKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ViewStats represents view statistics for a photo
type ViewStats struct {
	TotalViews      int64 `json:"total_views"`
	UniqueViewers    int64 `json:"unique_viewers"`
	AverageViewTime  int64 `json:"average_view_time_seconds"`
	ViewsToday       int64 `json:"views_today"`
	ViewsThisWeek    int64 `json:"views_this_week"`
	ViewsThisMonth   int64 `json:"views_this_month"`
}