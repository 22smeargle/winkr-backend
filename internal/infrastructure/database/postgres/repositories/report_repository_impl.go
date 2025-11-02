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

// ReportRepositoryImpl implements ReportRepository interface using GORM
type ReportRepositoryImpl struct {
	db *gorm.DB
}

// NewReportRepository creates a new ReportRepository instance
func NewReportRepository(db *gorm.DB) repositories.ReportRepository {
	return &ReportRepositoryImpl{db: db}
}

// Report methods

// Create creates a new report
func (r *ReportRepositoryImpl) Create(ctx context.Context, report *entities.Report) error {
	modelReport := r.domainToModelReport(report)
	if err := r.db.WithContext(ctx).Create(modelReport).Error; err != nil {
		logger.Error("Failed to create report", err)
		return fmt.Errorf("failed to create report: %w", err)
	}

	logger.Info("Report created successfully", map[string]interface{}{
		"report_id": report.ID,
		"reporter_id": report.ReporterID,
		"reported_id": report.ReportedID,
		"reason": report.Reason,
	})
	return nil
}

// GetByID retrieves a report by ID
func (r *ReportRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error) {
	var report models.Report
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		logger.Error("Failed to get report by ID", err)
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	// Convert to domain entity
	domainReport := r.modelToDomainReport(&report)
	return domainReport, nil
}

// Update updates a report
func (r *ReportRepositoryImpl) Update(ctx context.Context, report *entities.Report) error {
	modelReport := r.domainToModelReport(report)
	if err := r.db.WithContext(ctx).Save(modelReport).Error; err != nil {
		logger.Error("Failed to update report", err)
		return fmt.Errorf("failed to update report: %w", err)
	}

	logger.Info("Report updated successfully", map[string]interface{}{
		"report_id": report.ID,
	})
	return nil
}

// Delete soft deletes a report
func (r *ReportRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Report{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete report", err)
		return fmt.Errorf("failed to delete report: %w", err)
	}

	logger.Info("Report deleted successfully", map[string]interface{}{
		"report_id": id,
	})
	return nil
}

// GetReportsByReporter retrieves reports by reporter
func (r *ReportRepositoryImpl) GetReportsByReporter(ctx context.Context, reporterID uuid.UUID, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("reporter_id = ?", reporterID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get reports by reporter", err)
		return nil, fmt.Errorf("failed to get reports by reporter: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// GetReportsByReported retrieves reports by reported user
func (r *ReportRepositoryImpl) GetReportsByReported(ctx context.Context, reportedID uuid.UUID, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("reported_id = ?", reportedID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get reports by reported", err)
		return nil, fmt.Errorf("failed to get reports by reported: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// GetReportsByStatus retrieves reports by status
func (r *ReportRepositoryImpl) GetReportsByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get reports by status", err)
		return nil, fmt.Errorf("failed to get reports by status: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// GetReportsByReason retrieves reports by reason
func (r *ReportRepositoryImpl) GetReportsByReason(ctx context.Context, reason string, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("reason = ?", reason).Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get reports by reason", err)
		return nil, fmt.Errorf("failed to get reports by reason: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// GetPendingReports retrieves pending reports
func (r *ReportRepositoryImpl) GetPendingReports(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get pending reports", err)
		return nil, fmt.Errorf("failed to get pending reports: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// GetResolvedReports retrieves resolved reports
func (r *ReportRepositoryImpl) GetResolvedReports(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("status = ?", "resolved").Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get resolved reports", err)
		return nil, fmt.Errorf("failed to get resolved reports: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// UpdateReportStatus updates report status
func (r *ReportRepositoryImpl) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status string, adminID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if adminID != nil {
		updates["admin_id"] = *adminID
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("id = ?", reportID).Updates(updates).Error; err != nil {
		logger.Error("Failed to update report status", err)
		return fmt.Errorf("failed to update report status: %w", err)
	}

	logger.Info("Report status updated", map[string]interface{}{
		"report_id": reportID,
		"status": status,
	})
	return nil
}

// ResolveReport resolves a report with admin notes
func (r *ReportRepositoryImpl) ResolveReport(ctx context.Context, reportID uuid.UUID, adminID uuid.UUID, notes string) error {
	updates := map[string]interface{}{
		"status":    "resolved",
		"admin_id":  adminID,
		"notes":     notes,
		"resolved_at": time.Now(),
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("id = ?", reportID).Updates(updates).Error; err != nil {
		logger.Error("Failed to resolve report", err)
		return fmt.Errorf("failed to resolve report: %w", err)
	}

	logger.Info("Report resolved", map[string]interface{}{
		"report_id": reportID,
		"admin_id": adminID,
	})
	return nil
}

// GetReportCount retrieves report count
func (r *ReportRepositoryImpl) GetReportCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Count(&count).Error; err != nil {
		logger.Error("Failed to get report count", err)
		return 0, fmt.Errorf("failed to get report count: %w", err)
	}

	return count, nil
}

// GetReportCountByStatus retrieves report count by status
func (r *ReportRepositoryImpl) GetReportCountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("status = ?", status).Count(&count).Error; err != nil {
		logger.Error("Failed to get report count by status", err)
		return 0, fmt.Errorf("failed to get report count by status: %w", err)
	}

	return count, nil
}

// GetReportCountByReason retrieves report count by reason
func (r *ReportRepositoryImpl) GetReportCountByReason(ctx context.Context, reason string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("reason = ?", reason).Count(&count).Error; err != nil {
		logger.Error("Failed to get report count by reason", err)
		return 0, fmt.Errorf("failed to get report count by reason: %w", err)
	}

	return count, nil
}

// GetReportsCreatedInRange retrieves reports created in date range
func (r *ReportRepositoryImpl) GetReportsCreatedInRange(ctx context.Context, startDate, endDate interface{}) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&reports).Error; err != nil {
		logger.Error("Failed to get reports created in range", err)
		return nil, fmt.Errorf("failed to get reports created in range: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// ExistsReport checks if report exists
func (r *ReportRepositoryImpl) ExistsReport(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("id = ?", id).Count(&count).Error; err != nil {
		logger.Error("Failed to check report existence", err)
		return false, fmt.Errorf("failed to check report existence: %w", err)
	}

	return count > 0, nil
}

// UserCanViewReport checks if user can view a report
func (r *ReportRepositoryImpl) UserCanViewReport(ctx context.Context, userID, reportID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Report{}).Where("id = ? AND (reporter_id = ? OR reported_id = ?)", reportID, userID, userID).Count(&count).Error; err != nil {
		logger.Error("Failed to check report view permission", err)
		return false, fmt.Errorf("failed to check report view permission: %w", err)
	}

	return count > 0, nil
}

// GetAllReports retrieves all reports with pagination
func (r *ReportRepositoryImpl) GetAllReports(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	var reports []models.Report
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		logger.Error("Failed to get all reports", err)
		return nil, fmt.Errorf("failed to get all reports: %w", err)
	}

	// Convert to domain entities
	domainReports := make([]*entities.Report, len(reports))
	for i, report := range reports {
		domainReports[i] = r.modelToDomainReport(&report)
	}

	return domainReports, nil
}

// AdminUser methods

// CreateAdminUser creates a new admin user
func (r *ReportRepositoryImpl) CreateAdminUser(ctx context.Context, adminUser *entities.AdminUser) error {
	modelAdminUser := r.domainToModelAdminUser(adminUser)
	if err := r.db.WithContext(ctx).Create(modelAdminUser).Error; err != nil {
		logger.Error("Failed to create admin user", err)
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logger.Info("Admin user created successfully", map[string]interface{}{
		"admin_id": adminUser.ID,
		"user_id": adminUser.UserID,
		"role": adminUser.Role,
	})
	return nil
}

// GetAdminUserByID retrieves an admin user by ID
func (r *ReportRepositoryImpl) GetAdminUserByID(ctx context.Context, id uuid.UUID) (*entities.AdminUser, error) {
	var adminUser models.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin user not found")
		}
		logger.Error("Failed to get admin user by ID", err)
		return nil, fmt.Errorf("failed to get admin user by ID: %w", err)
	}

	// Convert to domain entity
	domainAdminUser := r.modelToDomainAdminUser(&adminUser)
	return domainAdminUser, nil
}

// GetAdminUserByUserID retrieves an admin user by user ID
func (r *ReportRepositoryImpl) GetAdminUserByUserID(ctx context.Context, userID uuid.UUID) (*entities.AdminUser, error) {
	var adminUser models.AdminUser
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin user not found")
		}
		logger.Error("Failed to get admin user by user ID", err)
		return nil, fmt.Errorf("failed to get admin user by user ID: %w", err)
	}

	// Convert to domain entity
	domainAdminUser := r.modelToDomainAdminUser(&adminUser)
	return domainAdminUser, nil
}

// UpdateAdminUser updates an admin user
func (r *ReportRepositoryImpl) UpdateAdminUser(ctx context.Context, adminUser *entities.AdminUser) error {
	modelAdminUser := r.domainToModelAdminUser(adminUser)
	if err := r.db.WithContext(ctx).Save(modelAdminUser).Error; err != nil {
		logger.Error("Failed to update admin user", err)
		return fmt.Errorf("failed to update admin user: %w", err)
	}

	logger.Info("Admin user updated successfully", map[string]interface{}{
		"admin_id": adminUser.ID,
	})
	return nil
}

// DeleteAdminUser soft deletes an admin user
func (r *ReportRepositoryImpl) DeleteAdminUser(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.AdminUser{}, "id = ?", id).Error; err != nil {
		logger.Error("Failed to delete admin user", err)
		return fmt.Errorf("failed to delete admin user: %w", err)
	}

	logger.Info("Admin user deleted successfully", map[string]interface{}{
		"admin_id": id,
	})
	return nil
}

// GetAdminUsersByRole retrieves admin users by role
func (r *ReportRepositoryImpl) GetAdminUsersByRole(ctx context.Context, role string, limit, offset int) ([]*entities.AdminUser, error) {
	var adminUsers []models.AdminUser
	if err := r.db.WithContext(ctx).Where("role = ?", role).Order("created_at DESC").Limit(limit).Offset(offset).Find(&adminUsers).Error; err != nil {
		logger.Error("Failed to get admin users by role", err)
		return nil, fmt.Errorf("failed to get admin users by role: %w", err)
	}

	// Convert to domain entities
	domainAdminUsers := make([]*entities.AdminUser, len(adminUsers))
	for i, adminUser := range adminUsers {
		domainAdminUsers[i] = r.modelToDomainAdminUser(&adminUser)
	}

	return domainAdminUsers, nil
}

// GetAllAdminUsers retrieves all admin users
func (r *ReportRepositoryImpl) GetAllAdminUsers(ctx context.Context, limit, offset int) ([]*entities.AdminUser, error) {
	var adminUsers []models.AdminUser
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&adminUsers).Error; err != nil {
		logger.Error("Failed to get all admin users", err)
		return nil, fmt.Errorf("failed to get all admin users: %w", err)
	}

	// Convert to domain entities
	domainAdminUsers := make([]*entities.AdminUser, len(adminUsers))
	for i, adminUser := range adminUsers {
		domainAdminUsers[i] = r.modelToDomainAdminUser(&adminUser)
	}

	return domainAdminUsers, nil
}

// ExistsAdminUser checks if admin user exists
func (r *ReportRepositoryImpl) ExistsAdminUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.AdminUser{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to check admin user existence", err)
		return false, fmt.Errorf("failed to check admin user existence: %w", err)
	}

	return count > 0, nil
}

// GetReportStats retrieves report statistics
func (r *ReportRepositoryImpl) GetReportStats(ctx context.Context) (*repositories.ReportStats, error) {
	var stats repositories.ReportStats
	
	// Get total reports count
	r.db.WithContext(ctx).Model(&models.Report{}).Count(&stats.TotalReports)
	
	// Get pending reports count
	r.db.WithContext(ctx).Model(&models.Report{}).Where("status = ?", "pending").Count(&stats.PendingReports)
	
	// Get resolved reports count
	r.db.WithContext(ctx).Model(&models.Report{}).Where("status = ?", "resolved").Count(&stats.ResolvedReports)
	
	// Get rejected reports count
	r.db.WithContext(ctx).Model(&models.Report{}).Where("status = ?", "rejected").Count(&stats.RejectedReports)
	
	// Get reports created today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.WithContext(ctx).Model(&models.Report{}).Where("DATE(created_at) = DATE(?)", today).Count(&stats.ReportsToday)
	
	// Get reports created this week
	weekStart := time.Now().AddDate(-7, 0, 0)
	r.db.WithContext(ctx).Model(&models.Report{}).Where("created_at >= ?", weekStart).Count(&stats.ReportsThisWeek)
	
	// Get reports created this month
	monthStart := time.Now().AddDate(-30, 0, 0)
	r.db.WithContext(ctx).Model(&models.Report{}).Where("created_at >= ?", monthStart).Count(&stats.ReportsThisMonth)
	
	// Get admin users count
	r.db.WithContext(ctx).Model(&models.AdminUser{}).Count(&stats.AdminUsers)
	
	return &stats, nil
}

// Helper methods to convert between domain and model entities

// modelToDomainReport converts model Report to domain Report
func (r *ReportRepositoryImpl) modelToDomainReport(model *models.Report) *entities.Report {
	return &entities.Report{
		ID:          model.ID,
		ReporterID:  model.ReporterID,
		ReportedID:  model.ReportedID,
		Reason:      model.Reason,
		Description: model.Description,
		Status:      model.Status,
		AdminID:     model.AdminID,
		Notes:       model.Notes,
		ResolvedAt:  model.ResolvedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// domainToModelReport converts domain Report to model Report
func (r *ReportRepositoryImpl) domainToModelReport(report *entities.Report) *models.Report {
	return &models.Report{
		ID:          report.ID,
		ReporterID:  report.ReporterID,
		ReportedID:  report.ReportedID,
		Reason:      report.Reason,
		Description: report.Description,
		Status:      report.Status,
		AdminID:     report.AdminID,
		Notes:       report.Notes,
		ResolvedAt:  report.ResolvedAt,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
	}
}

// modelToDomainAdminUser converts model AdminUser to domain AdminUser
func (r *ReportRepositoryImpl) modelToDomainAdminUser(model *models.AdminUser) *entities.AdminUser {
	return &entities.AdminUser{
		ID:        model.ID,
		UserID:    model.UserID,
		Role:      model.Role,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// domainToModelAdminUser converts domain AdminUser to model AdminUser
func (r *ReportRepositoryImpl) domainToModelAdminUser(adminUser *entities.AdminUser) *models.AdminUser {
	return &models.AdminUser{
		ID:        adminUser.ID,
		UserID:    adminUser.UserID,
		Role:      adminUser.Role,
		CreatedAt: adminUser.CreatedAt,
		UpdatedAt: adminUser.UpdatedAt,
	}
}