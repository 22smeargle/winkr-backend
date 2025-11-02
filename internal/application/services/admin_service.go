package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/admin"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// AdminService provides high-level admin operations
type AdminService struct {
	// User management use cases
	getUsersUseCase         *admin.GetUsersUseCase
	getUserDetailsUseCase   *admin.GetUserDetailsUseCase
	updateUserUseCase       *admin.UpdateUserUseCase
	deleteUserUseCase       *admin.DeleteUserUseCase
	suspendUserUseCase      *admin.SuspendUserUseCase

	// Analytics use cases
	getPlatformStatsUseCase *admin.GetPlatformStatsUseCase
	getUserStatsUseCase     *admin.GetUserStatsUseCase
	getMatchStatsUseCase    *admin.GetMatchStatsUseCase
	getMessageStatsUseCase  *admin.GetMessageStatsUseCase
	getPaymentStatsUseCase  *admin.GetPaymentStatsUseCase
	getVerificationStatsUseCase *admin.GetVerificationStatsUseCase

	// System management use cases
	getSystemHealthUseCase *admin.GetSystemHealthUseCase
	getSystemLogsUseCase   *admin.GetSystemLogsUseCase
	manageSystemConfigUseCase *admin.ManageSystemConfigUseCase

	// Content management use cases
	getReportedContentUseCase *admin.GetReportedContentUseCase
	moderateContentUseCase    *admin.ModerateContentUseCase

	// Repositories
	userRepo       repositories.UserRepository
	photoRepo      repositories.PhotoRepository
	messageRepo    repositories.MessageRepository
	matchRepo      repositories.MatchRepository
	reportRepo     repositories.ReportRepository
	paymentRepo    repositories.PaymentRepository
	subscriptionRepo repositories.SubscriptionRepository
	verificationRepo repositories.VerificationRepository
}

// NewAdminService creates a new AdminService
func NewAdminService(
	userRepo repositories.UserRepository,
	photoRepo repositories.PhotoRepository,
	messageRepo repositories.MessageRepository,
	matchRepo repositories.MatchRepository,
	reportRepo repositories.ReportRepository,
	paymentRepo repositories.PaymentRepository,
	subscriptionRepo repositories.SubscriptionRepository,
	verificationRepo repositories.VerificationRepository,
) *AdminService {
	return &AdminService{
		// Initialize use cases
		getUsersUseCase:         admin.NewGetUsersUseCase(userRepo),
		getUserDetailsUseCase:   admin.NewGetUserDetailsUseCase(userRepo, photoRepo, matchRepo, reportRepo),
		updateUserUseCase:       admin.NewUpdateUserUseCase(userRepo),
		deleteUserUseCase:       admin.NewDeleteUserUseCase(userRepo, photoRepo, messageRepo, matchRepo, reportRepo),
		suspendUserUseCase:      admin.NewSuspendUserUseCase(userRepo),

		getPlatformStatsUseCase: admin.NewGetPlatformStatsUseCase(userRepo, matchRepo, messageRepo, paymentRepo, subscriptionRepo),
		getUserStatsUseCase:     admin.NewGetUserStatsUseCase(userRepo),
		getMatchStatsUseCase:    admin.NewGetMatchStatsUseCase(matchRepo, userRepo),
		getMessageStatsUseCase:  admin.NewGetMessageStatsUseCase(messageRepo, userRepo),
		getPaymentStatsUseCase:  admin.NewGetPaymentStatsUseCase(paymentRepo, subscriptionRepo),
		getVerificationStatsUseCase: admin.NewGetVerificationStatsUseCase(verificationRepo),

		getSystemHealthUseCase: admin.NewGetSystemHealthUseCase(),
		getSystemLogsUseCase:   admin.NewGetSystemLogsUseCase(),
		manageSystemConfigUseCase: admin.NewManageSystemConfigUseCase(),

		getReportedContentUseCase: admin.NewGetReportedContentUseCase(reportRepo, photoRepo, messageRepo),
		moderateContentUseCase:    admin.NewModerateContentUseCase(reportRepo, photoRepo, messageRepo, userRepo),

		// Store repositories
		userRepo:       userRepo,
		photoRepo:      photoRepo,
		messageRepo:    messageRepo,
		matchRepo:      matchRepo,
		reportRepo:     reportRepo,
		paymentRepo:    paymentRepo,
		subscriptionRepo: subscriptionRepo,
		verificationRepo: verificationRepo,
	}
}

// User Management Methods

// GetUsers retrieves users with filtering and pagination
func (s *AdminService) GetUsers(ctx context.Context, adminID uuid.UUID, req admin.GetUsersRequest) (*admin.UsersResponse, error) {
	logger.Info("AdminService.GetUsers called", "admin_id", adminID)
	return s.getUsersUseCase.Execute(ctx, req)
}

// GetUserDetails retrieves detailed user information
func (s *AdminService) GetUserDetails(ctx context.Context, adminID, userID uuid.UUID) (*admin.UserDetailsResponse, error) {
	logger.Info("AdminService.GetUserDetails called", "admin_id", adminID, "user_id", userID)
	req := admin.GetUserDetailsRequest{
		AdminID: adminID,
		UserID:  userID,
	}
	return s.getUserDetailsUseCase.Execute(ctx, req)
}

// UpdateUser updates user information
func (s *AdminService) UpdateUser(ctx context.Context, adminID, userID uuid.UUID, req admin.UpdateUserRequest) (*admin.UpdateUserResponse, error) {
	logger.Info("AdminService.UpdateUser called", "admin_id", adminID, "user_id", userID)
	req.AdminID = adminID
	req.UserID = userID
	return s.updateUserUseCase.Execute(ctx, req)
}

// DeleteUser deletes a user account
func (s *AdminService) DeleteUser(ctx context.Context, adminID, userID uuid.UUID, hardDelete bool) (*admin.DeleteUserResponse, error) {
	logger.Info("AdminService.DeleteUser called", "admin_id", adminID, "user_id", userID, "hard_delete", hardDelete)
	req := admin.DeleteUserRequest{
		AdminID:    adminID,
		UserID:     userID,
		HardDelete: hardDelete,
	}
	return s.deleteUserUseCase.Execute(ctx, req)
}

// SuspendUser suspends a user account
func (s *AdminService) SuspendUser(ctx context.Context, adminID, userID uuid.UUID, duration int, reason string) (*admin.SuspendUserResponse, error) {
	logger.Info("AdminService.SuspendUser called", "admin_id", adminID, "user_id", userID, "duration", duration)
	req := admin.SuspendUserRequest{
		AdminID:  adminID,
		UserID:   userID,
		Action:   "suspend",
		Duration: duration,
		Reason:   reason,
	}
	return s.suspendUserUseCase.Execute(ctx, req)
}

// UnsuspendUser unsuspends a user account
func (s *AdminService) UnsuspendUser(ctx context.Context, adminID, userID uuid.UUID) (*admin.SuspendUserResponse, error) {
	logger.Info("AdminService.UnsuspendUser called", "admin_id", adminID, "user_id", userID)
	req := admin.SuspendUserRequest{
		AdminID: adminID,
		UserID:  userID,
		Action:  "unsuspend",
	}
	return s.suspendUserUseCase.Execute(ctx, req)
}

// BanUser bans a user account
func (s *AdminService) BanUser(ctx context.Context, adminID, userID uuid.UUID, reason string) (*admin.SuspendUserResponse, error) {
	logger.Info("AdminService.BanUser called", "admin_id", adminID, "user_id", userID)
	req := admin.SuspendUserRequest{
		AdminID: adminID,
		UserID:  userID,
		Action:  "ban",
		Reason:  reason,
	}
	return s.suspendUserUseCase.Execute(ctx, req)
}

// UnbanUser unbans a user account
func (s *AdminService) UnbanUser(ctx context.Context, adminID, userID uuid.UUID) (*admin.SuspendUserResponse, error) {
	logger.Info("AdminService.UnbanUser called", "admin_id", adminID, "user_id", userID)
	req := admin.SuspendUserRequest{
		AdminID: adminID,
		UserID:  userID,
		Action:  "unban",
	}
	return s.suspendUserUseCase.Execute(ctx, req)
}

// Analytics Methods

// GetPlatformStats retrieves platform-wide statistics
func (s *AdminService) GetPlatformStats(ctx context.Context, adminID uuid.UUID, period string) (*admin.PlatformStatsResponse, error) {
	logger.Info("AdminService.GetPlatformStats called", "admin_id", adminID, "period", period)
	req := admin.GetPlatformStatsRequest{
		AdminID: adminID,
		Period:  period,
	}
	return s.getPlatformStatsUseCase.Execute(ctx, req)
}

// GetUserStats retrieves user statistics
func (s *AdminService) GetUserStats(ctx context.Context, adminID uuid.UUID, period, groupBy string) (*admin.UserStatsResponse, error) {
	logger.Info("AdminService.GetUserStats called", "admin_id", adminID, "period", period, "group_by", groupBy)
	req := admin.GetUserStatsRequest{
		AdminID: adminID,
		Period:  period,
		GroupBy: groupBy,
	}
	return s.getUserStatsUseCase.Execute(ctx, req)
}

// GetMatchStats retrieves match statistics
func (s *AdminService) GetMatchStats(ctx context.Context, adminID uuid.UUID, period, groupBy string) (*admin.MatchStatsResponse, error) {
	logger.Info("AdminService.GetMatchStats called", "admin_id", adminID, "period", period, "group_by", groupBy)
	req := admin.GetMatchStatsRequest{
		AdminID: adminID,
		Period:  period,
		GroupBy: groupBy,
	}
	return s.getMatchStatsUseCase.Execute(ctx, req)
}

// GetMessageStats retrieves message statistics
func (s *AdminService) GetMessageStats(ctx context.Context, adminID uuid.UUID, period, groupBy string) (*admin.MessageStatsResponse, error) {
	logger.Info("AdminService.GetMessageStats called", "admin_id", adminID, "period", period, "group_by", groupBy)
	req := admin.GetMessageStatsRequest{
		AdminID: adminID,
		Period:  period,
		GroupBy: groupBy,
	}
	return s.getMessageStatsUseCase.Execute(ctx, req)
}

// GetPaymentStats retrieves payment statistics
func (s *AdminService) GetPaymentStats(ctx context.Context, adminID uuid.UUID, period, groupBy string) (*admin.PaymentStatsResponse, error) {
	logger.Info("AdminService.GetPaymentStats called", "admin_id", adminID, "period", period, "group_by", groupBy)
	req := admin.GetPaymentStatsRequest{
		AdminID: adminID,
		Period:  period,
		GroupBy: groupBy,
	}
	return s.getPaymentStatsUseCase.Execute(ctx, req)
}

// GetVerificationStats retrieves verification statistics
func (s *AdminService) GetVerificationStats(ctx context.Context, adminID uuid.UUID, period, groupBy, verificationType string) (*admin.VerificationStatsResponse, error) {
	logger.Info("AdminService.GetVerificationStats called", "admin_id", adminID, "period", period, "group_by", groupBy, "verification_type", verificationType)
	req := admin.GetVerificationStatsRequest{
		AdminID:          adminID,
		Period:           period,
		GroupBy:          groupBy,
		VerificationType: verificationType,
	}
	return s.getVerificationStatsUseCase.Execute(ctx, req)
}

// System Management Methods

// GetSystemHealth retrieves system health status
func (s *AdminService) GetSystemHealth(ctx context.Context, adminID uuid.UUID) (*admin.SystemHealthResponse, error) {
	logger.Info("AdminService.GetSystemHealth called", "admin_id", adminID)
	req := admin.GetSystemHealthRequest{
		AdminID: adminID,
	}
	return s.getSystemHealthUseCase.Execute(ctx, req)
}

// GetSystemLogs retrieves system logs
func (s *AdminService) GetSystemLogs(ctx context.Context, adminID uuid.UUID, startTime, endTime *time.Time, level, service string, limit, offset int, search string) (*admin.SystemLogsResponse, error) {
	logger.Info("AdminService.GetSystemLogs called", "admin_id", adminID, "level", level, "service", service)
	req := admin.GetSystemLogsRequest{
		AdminID:   adminID,
		StartTime: startTime,
		EndTime:   endTime,
		Level:     level,
		Service:   service,
		Limit:     limit,
		Offset:    offset,
		Search:    search,
	}
	return s.getSystemLogsUseCase.Execute(ctx, req)
}

// GetSystemConfig retrieves system configuration
func (s *AdminService) GetSystemConfig(ctx context.Context, adminID uuid.UUID, section string) (*admin.SystemConfigResponse, error) {
	logger.Info("AdminService.GetSystemConfig called", "admin_id", adminID, "section", section)
	req := admin.GetSystemConfigRequest{
		AdminID: adminID,
		Section: section,
	}
	return s.manageSystemConfigUseCase.Execute(ctx, req)
}

// UpdateSystemConfig updates system configuration
func (s *AdminService) UpdateSystemConfig(ctx context.Context, adminID uuid.UUID, section string, config map[string]interface{}, reason string) (*admin.UpdateSystemConfigResponse, error) {
	logger.Info("AdminService.UpdateSystemConfig called", "admin_id", adminID, "section", section)
	req := admin.UpdateSystemConfigRequest{
		AdminID: adminID,
		Section: section,
		Config:  config,
		Reason:  reason,
	}
	return s.manageSystemConfigUseCase.ExecuteUpdate(ctx, req)
}

// Content Management Methods

// GetReportedContent retrieves reported content
func (s *AdminService) GetReportedContent(ctx context.Context, adminID uuid.UUID, contentType, status string, limit, offset int) (*admin.ReportedContentResponse, error) {
	logger.Info("AdminService.GetReportedContent called", "admin_id", adminID, "content_type", contentType, "status", status)
	req := admin.GetReportedContentRequest{
		AdminID:     adminID,
		ContentType: contentType,
		Status:      status,
		Limit:       limit,
		Offset:      offset,
	}
	return s.getReportedContentUseCase.Execute(ctx, req)
}

// ModerateContent moderates content
func (s *AdminService) ModerateContent(ctx context.Context, adminID uuid.UUID, contentType string, contentID uuid.UUID, action, reason, notes string, warnUser, banUser bool, banDuration int) (*admin.ModerateContentResponse, error) {
	logger.Info("AdminService.ModerateContent called", "admin_id", adminID, "content_type", contentType, "content_id", contentID, "action", action)
	req := admin.ModerateContentRequest{
		AdminID:     adminID,
		ContentType: contentType,
		ContentID:   contentID,
		Action:      action,
		Reason:      reason,
		Notes:       notes,
		WarnUser:    warnUser,
		BanUser:     banUser,
		BanDuration: banDuration,
	}
	return s.moderateContentUseCase.Execute(ctx, req)
}

// Bulk Operations

// BulkUpdateUsers performs bulk updates on multiple users
func (s *AdminService) BulkUpdateUsers(ctx context.Context, adminID uuid.UUID, userIDs []uuid.UUID, updates map[string]interface{}) (*BulkOperationResponse, error) {
	logger.Info("AdminService.BulkUpdateUsers called", "admin_id", adminID, "user_count", len(userIDs))
	
	response := &BulkOperationResponse{
		Total:     len(userIDs),
		Success:   0,
		Failed:    0,
		Errors:    []string{},
		Timestamp: time.Now(),
	}

	for _, userID := range userIDs {
		req := admin.UpdateUserRequest{
			AdminID: adminID,
			UserID:  userID,
			Updates: updates,
		}
		
		_, err := s.updateUserUseCase.Execute(ctx, req)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to update user %s: %s", userID.String(), err.Error()))
		} else {
			response.Success++
		}
	}

	return response, nil
}

// BulkDeleteUsers performs bulk deletion of multiple users
func (s *AdminService) BulkDeleteUsers(ctx context.Context, adminID uuid.UUID, userIDs []uuid.UUID, hardDelete bool) (*BulkOperationResponse, error) {
	logger.Info("AdminService.BulkDeleteUsers called", "admin_id", adminID, "user_count", len(userIDs), "hard_delete", hardDelete)
	
	response := &BulkOperationResponse{
		Total:     len(userIDs),
		Success:   0,
		Failed:    0,
		Errors:    []string{},
		Timestamp: time.Now(),
	}

	for _, userID := range userIDs {
		req := admin.DeleteUserRequest{
			AdminID:    adminID,
			UserID:     userID,
			HardDelete: hardDelete,
		}
		
		_, err := s.deleteUserUseCase.Execute(ctx, req)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to delete user %s: %s", userID.String(), err.Error()))
		} else {
			response.Success++
		}
	}

	return response, nil
}

// BulkModerateContent performs bulk moderation of multiple content items
func (s *AdminService) BulkModerateContent(ctx context.Context, adminID uuid.UUID, contentItems []BulkModerateContentItem) (*BulkOperationResponse, error) {
	logger.Info("AdminService.BulkModerateContent called", "admin_id", adminID, "content_count", len(contentItems))
	
	response := &BulkOperationResponse{
		Total:     len(contentItems),
		Success:   0,
		Failed:    0,
		Errors:    []string{},
		Timestamp: time.Now(),
	}

	for _, item := range contentItems {
		req := admin.ModerateContentRequest{
			AdminID:     adminID,
			ContentType: item.ContentType,
			ContentID:   item.ContentID,
			Action:      item.Action,
			Reason:      item.Reason,
			Notes:       item.Notes,
			WarnUser:    item.WarnUser,
			BanUser:     item.BanUser,
			BanDuration: item.BanDuration,
		}
		
		_, err := s.moderateContentUseCase.Execute(ctx, req)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to moderate content %s: %s", item.ContentID.String(), err.Error()))
		} else {
			response.Success++
		}
	}

	return response, nil
}

// Helper Types

// BulkOperationResponse represents response from bulk operations
type BulkOperationResponse struct {
	Total     int       `json:"total"`
	Success   int       `json:"success"`
	Failed    int       `json:"failed"`
	Errors    []string  `json:"errors"`
	Timestamp time.Time `json:"timestamp"`
}

// BulkModerateContentItem represents an item for bulk content moderation
type BulkModerateContentItem struct {
	ContentType string    `json:"content_type"`
	ContentID   uuid.UUID `json:"content_id"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
	Notes       string    `json:"notes"`
	WarnUser    bool      `json:"warn_user"`
	BanUser     bool      `json:"ban_user"`
	BanDuration int       `json:"ban_duration"`
}