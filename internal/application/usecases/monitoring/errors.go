package monitoring

import "errors"

// Monitoring use case errors
var (
	// General errors
	ErrInvalidHealthData     = errors.New("invalid health data type")
	ErrInvalidMetricType     = errors.New("invalid metric type")
	ErrInvalidTimeRange      = errors.New("invalid time range")
	ErrInvalidAlertRule     = errors.New("invalid alert rule")
	ErrAlertNotFound        = errors.New("alert not found")
	ErrAlertAlreadyExists   = errors.New("alert already exists")
	
	// Service errors
	ErrHealthCheckFailed    = errors.New("health check failed")
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	ErrAlertingFailed      = errors.New("alerting failed")
	ErrStorageFailed       = errors.New("storage operation failed")
	
	// Configuration errors
	ErrMonitoringDisabled   = errors.New("monitoring is disabled")
	ErrInvalidConfiguration = errors.New("invalid monitoring configuration")
	
	// Validation errors
	ErrEmptyMetricName     = errors.New("metric name cannot be empty")
	ErrInvalidMetricValue   = errors.New("invalid metric value")
	ErrInvalidThreshold     = errors.New("invalid threshold value")
	ErrInvalidSeverity      = errors.New("invalid alert severity")
)