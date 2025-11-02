package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetSystemLogsUseCase handles retrieving system logs
type GetSystemLogsUseCase struct {
	// In a real implementation, this would include log repository or service
}

// NewGetSystemLogsUseCase creates a new GetSystemLogsUseCase
func NewGetSystemLogsUseCase() *GetSystemLogsUseCase {
	return &GetSystemLogsUseCase{}
}

// GetSystemLogsRequest represents a request to get system logs
type GetSystemLogsRequest struct {
	AdminID    uuid.UUID  `json:"admin_id" validate:"required"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Level      string     `json:"level" validate:"omitempty,oneof=debug info warn error fatal"`
	Service    string     `json:"service"`
	Limit      int        `json:"limit" validate:"omitempty,min=1,max=1000"`
	Offset     int        `json:"offset" validate:"omitempty,min=0"`
	Search     string     `json:"search"`
}

// SystemLogsResponse represents response from getting system logs
type SystemLogsResponse struct {
	Logs      []SystemLog `json:"logs"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PerPage   int         `json:"per_page"`
	TotalPages int        `json:"total_pages"`
	Timestamp time.Time   `json:"timestamp"`
}

// SystemLog represents a system log entry
type SystemLog struct {
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // debug, info, warn, error, fatal
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	Context   LogContext `json:"context"`
	TraceID   string    `json:"trace_id,omitempty"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
}

// LogContext represents additional context information for a log entry
type LogContext struct {
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   int                    `json:"duration,omitempty"` // milliseconds
	IP         string                 `json:"ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Execute retrieves system logs
func (uc *GetSystemLogsUseCase) Execute(ctx context.Context, req GetSystemLogsRequest) (*SystemLogsResponse, error) {
	logger.Info("GetSystemLogs use case executed", "admin_id", req.AdminID, "level", req.Level, "service", req.Service, "limit", req.Limit)

	// Set default values
	if req.Limit == 0 {
		req.Limit = 100
	}

	// Calculate time range
	startTime, endTime := uc.calculateTimeRange(req.StartTime, req.EndTime)

	// Get logs
	logs, total, err := uc.getLogs(ctx, startTime, endTime, req)
	if err != nil {
		logger.Error("Failed to get logs", err, "admin_id", req.AdminID)
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	// Calculate pagination
	page := req.Offset/req.Limit + 1
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	logger.Info("GetSystemLogs use case completed successfully", "admin_id", req.AdminID, "total", total)
	return &SystemLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PerPage:    req.Limit,
		TotalPages: totalPages,
		Timestamp:  time.Now(),
	}, nil
}

// calculateTimeRange calculates start and end time
func (uc *GetSystemLogsUseCase) calculateTimeRange(startTime, endTime *time.Time) (time.Time, time.Time) {
	now := time.Now()

	// If custom time range is provided, use it
	if startTime != nil && endTime != nil {
		return *startTime, *endTime
	}

	// Default to last 24 hours
	if startTime == nil {
		start := now.Add(-24 * time.Hour)
		return start, now
	}

	return *startTime, now
}

// getLogs retrieves logs from the log storage
func (uc *GetSystemLogsUseCase) getLogs(ctx context.Context, startTime, endTime time.Time, req GetSystemLogsRequest) ([]SystemLog, int64, error) {
	// Mock data - in real implementation, this would query the log storage
	logs := []SystemLog{
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-2 * time.Hour),
			Level:     "info",
			Service:   "API Server",
			Message:   "User login successful",
			Context: LogContext{
				Method:     "POST",
				Path:       "/api/auth/login",
				StatusCode: 200,
				Duration:   120,
				IP:         "192.168.1.100",
				UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			TraceID:   "trace-123456",
			UserID:    uuid.New(),
			RequestID: "req-789012",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-3 * time.Hour),
			Level:     "warn",
			Service:   "Database",
			Message:   "Slow query detected",
			Context: LogContext{
				Method: "GET",
				Path:   "/api/users",
				Extra: map[string]interface{}{
					"query": "SELECT * FROM users WHERE active = true",
					"duration_ms": 1500,
				},
			},
			TraceID: "trace-234567",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-4 * time.Hour),
			Level:     "error",
			Service:   "Payment Gateway",
			Message:   "Payment processing failed",
			Context: LogContext{
				Method:     "POST",
				Path:       "/api/payments/process",
				StatusCode: 500,
				Duration:   3000,
				Extra: map[string]interface{}{
					"payment_id": "pay_123456789",
					"error_code": "card_declined",
				},
			},
			TraceID:   "trace-345678",
			UserID:    uuid.New(),
			RequestID: "req-890123",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-5 * time.Hour),
			Level:     "debug",
			Service:   "Cache Service",
			Message:   "Cache miss for user profile",
			Context: LogContext{
				Method: "GET",
				Path:   "/api/users/12345",
				Extra: map[string]interface{}{
					"cache_key": "user:profile:12345",
					"ttl":       3600,
				},
			},
			TraceID: "trace-456789",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-6 * time.Hour),
			Level:     "info",
			Service:   "Email Service",
			Message:   "Verification email sent",
			Context: LogContext{
				Method: "POST",
				Path:   "/api/verification/send",
				Extra: map[string]interface{}{
					"email": "user@example.com",
					"template": "verification",
				},
			},
			TraceID:   "trace-567890",
			UserID:    uuid.New(),
			RequestID: "req-901234",
		},
	}

	// Filter logs based on request parameters
	filteredLogs := uc.filterLogs(logs, req)

	// Apply pagination
	start := req.Offset
	end := start + req.Limit
	if end > len(filteredLogs) {
		end = len(filteredLogs)
	}

	if start >= len(filteredLogs) {
		return []SystemLog{}, int64(len(filteredLogs)), nil
	}

	return filteredLogs[start:end], int64(len(filteredLogs)), nil
}

// filterLogs filters logs based on request parameters
func (uc *GetSystemLogsUseCase) filterLogs(logs []SystemLog, req GetSystemLogsRequest) []SystemLog {
	var filtered []SystemLog

	for _, log := range logs {
		// Filter by level
		if req.Level != "" && log.Level != req.Level {
			continue
		}

		// Filter by service
		if req.Service != "" && log.Service != req.Service {
			continue
		}

		// Filter by search term
		if req.Search != "" {
			if !uc.containsSearchTerm(log, req.Search) {
				continue
			}
		}

		filtered = append(filtered, log)
	}

	return filtered
}

// containsSearchTerm checks if a log contains the search term
func (uc *GetSystemLogsUseCase) containsSearchTerm(log SystemLog, searchTerm string) bool {
	// Simple search implementation - in real implementation, this would be more sophisticated
	if log.Message == searchTerm {
		return true
	}
	if log.Service == searchTerm {
		return true
	}
	if log.Context.Method == searchTerm {
		return true
	}
	if log.Context.Path == searchTerm {
		return true
	}

	return false
}