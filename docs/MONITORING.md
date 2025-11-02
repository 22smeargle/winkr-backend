# Monitoring System Documentation

This document provides comprehensive documentation for the monitoring system implemented in the Winkr backend application.

## Overview

The monitoring system provides comprehensive health checks, metrics collection, alerting, and observability features for the Winkr backend. It follows clean architecture principles and integrates seamlessly with the existing application infrastructure.

## Architecture

The monitoring system consists of several key components:

1. **Health Check Service** - Monitors the health of various system components
2. **Metrics Collection Service** - Collects and aggregates application and system metrics
3. **Alerting Service** - Provides rule-based alerting with multiple notification channels
4. **Monitoring Storage Service** - Handles storage and retrieval of monitoring data
5. **Background Jobs Service** - Manages periodic monitoring tasks
6. **Monitoring Middleware** - Captures request-level metrics and events
7. **Monitoring Handlers** - Exposes monitoring endpoints via HTTP API

## Configuration

The monitoring system is configured through the `MonitoringConfig` section in the main application configuration:

```go
Monitoring: config.MonitoringConfig{
    HealthCheck: config.HealthCheckConfig{
        Enabled: true,
        Timeout: 5 * time.Second,
        Interval: 30 * time.Second,
    },
    Metrics: config.MetricsConfig{
        Enabled: true,
        CollectionInterval: 60 * time.Second,
        RetentionPeriod: 24 * time.Hour,
    },
    Alerting: config.AlertingConfig{
        Enabled: true,
        CPUUsageThreshold: 80.0,
        MemoryUsageThreshold: 85.0,
        DiskUsageThreshold: 90.0,
        ErrorRateThreshold: 5.0,
        ResponseTimeThreshold: 1000,
    },
    BackgroundJobs: config.BackgroundJobsConfig{
        HealthCheckJobInterval: 30 * time.Second,
        MetricsAggregationInterval: 5 * time.Minute,
        LogCleanupInterval: 1 * time.Hour,
        SystemMonitoringInterval: 2 * time.Minute,
        ExternalServiceCheckInterval: 3 * time.Minute,
    },
    Storage: config.StorageConfig{
        HealthRetentionPeriod: 24 * time.Hour,
        HTTPMetricsRetentionPeriod: 24 * time.Hour,
        DatabaseMetricsRetentionPeriod: 24 * time.Hour,
        CacheMetricsRetentionPeriod: 24 * time.Hour,
        BusinessMetricsRetentionPeriod: 7 * 24 * time.Hour,
        SystemMetricsRetentionPeriod: 24 * time.Hour,
        TimeSeriesRetentionPeriod: 7 * 24 * time.Hour,
        MaxTimeSeriesPoints: 1000,
    },
    Logging: config.LoggingConfig{
        StructuredLogging: true,
        CorrelationIDs: true,
        PerformanceLogging: true,
        SecurityEventLogging: true,
        RetentionPeriod: 30 * 24 * time.Hour,
        MaxFileSize: 100 * 1024 * 1024, // 100MB
        MaxBackups: 10,
    },
}
```

## Health Checks

### Health Check Service

The health check service monitors the following components:

- **Database** - PostgreSQL connectivity and performance
- **Redis** - Cache connectivity and performance
- **Storage** - S3 or other storage service connectivity
- **External Services** - Stripe, email services, SMS services
- **System Resources** - CPU, memory, disk usage

### Health Status Levels

- `Healthy` - All components are functioning normally
- `Degraded` - Some components have issues but the system is still operational
- `Unhealthy` - Critical components are failing

### Health Check Endpoints

#### GET /health

Returns basic health status:

```json
{
  "status": "ok",
  "timestamp": "2025-11-02T16:30:00Z",
  "uptime": "2h30m45s"
}
```

#### GET /health/detailed

Returns detailed health status with component information:

```json
{
  "status": "ok",
  "timestamp": "2025-11-02T16:30:00Z",
  "uptime": "2h30m45s",
  "components": {
    "database": {
      "status": "healthy",
      "response_time": 15,
      "error": null
    },
    "redis": {
      "status": "healthy",
      "response_time": 2,
      "error": null
    },
    "storage": {
      "status": "healthy",
      "response_time": 45,
      "error": null
    }
  },
  "system": {
    "cpu": {
      "status": "healthy",
      "usage": 25.5,
      "threshold": 80.0
    },
    "memory": {
      "status": "healthy",
      "usage": 45.2,
      "threshold": 85.0
    },
    "disk": {
      "status": "healthy",
      "usage": 60.1,
      "threshold": 90.0
    }
  }
}
```

#### GET /health/components/{component}

Returns health status for a specific component:

```json
{
  "component": "database",
  "status": "healthy",
  "response_time": 15,
  "error": null,
  "timestamp": "2025-11-02T16:30:00Z"
}
```

## Metrics

### Metrics Collection

The metrics service collects the following types of metrics:

#### HTTP Metrics

- Total requests
- Requests by method
- Requests by status code
- Response times (min, max, average)
- Error rates

#### Database Metrics

- Total queries
- Queries by type
- Slow queries
- Connection pool status
- Query execution times

#### Cache Metrics

- Total operations
- Hit/miss ratios
- Operations by type
- Cache size

#### Business Metrics

- Total users
- Active users
- User registrations
- Matches created
- Messages sent
- Photo uploads

#### System Metrics

- CPU usage
- Memory usage
- Disk usage
- Goroutine count
- GC cycles

### Metrics Endpoints

#### GET /metrics

Returns metrics in Prometheus format:

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1234
http_requests_total{method="POST",status="201"} 567

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 1000
http_request_duration_seconds_bucket{le="0.5"} 1200
http_request_duration_seconds_bucket{le="1.0"} 1230
http_request_duration_seconds_bucket{le="+Inf"} 1234
```

#### GET /metrics/json

Returns metrics in JSON format:

```json
{
  "http": {
    "total_requests": 1234,
    "requests_by_method": {
      "GET": 800,
      "POST": 300,
      "PUT": 100,
      "DELETE": 34
    },
    "requests_by_status": {
      "200": 1000,
      "201": 100,
      "400": 50,
      "404": 34,
      "500": 50
    },
    "response_times": {
      "min": 10,
      "max": 5000,
      "average": 150
    },
    "error_rate": 0.08
  },
  "database": {
    "total_queries": 5000,
    "queries_by_type": {
      "SELECT": 3000,
      "INSERT": 1000,
      "UPDATE": 800,
      "DELETE": 200
    },
    "slow_queries": 5,
    "connections_active": 10,
    "connections_idle": 5
  },
  "cache": {
    "total_operations": 10000,
    "hits": 8500,
    "misses": 1500,
    "hit_ratio": 0.85,
    "operations_by_type": {
      "get": 7000,
      "set": 2000,
      "del": 1000
    }
  },
  "business": {
    "total_users": 10000,
    "active_users": 2000,
    "registrations_today": 50,
    "matches_created": 500,
    "messages_sent": 2000,
    "photos_uploaded": 100
  },
  "system": {
    "cpu_usage": 25.5,
    "memory_usage": 45.2,
    "disk_usage": 60.1,
    "goroutines": 50,
    "gc_cycles": 100
  }
}
```

## Alerting

### Alert Rules

Alert rules define conditions that trigger alerts:

```go
type AlertRule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Description string        `json:"description"`
    Metric      string        `json:"metric"`
    Operator    string        `json:"operator"`
    Threshold   float64       `json:"threshold"`
    Severity    AlertSeverity `json:"severity"`
    Enabled     bool          `json:"enabled"`
    CreatedAt   time.Time     `json:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at"`
}
```

### Alert Severities

- `Info` - Informational alerts
- `Warning` - Warning conditions
- `Critical` - Critical conditions requiring immediate attention

### Alert Statuses

- `Active` - Alert is currently active
- `Acknowledged` - Alert has been acknowledged
- `Resolved` - Alert condition has been resolved

### Notification Channels

The alerting service supports multiple notification channels:

- **Email** - Send alerts via email
- **Slack** - Send alerts to Slack channels
- **Webhook** - Send alerts to custom webhook endpoints

### Alert Management

#### Create Alert Rule

```bash
POST /api/v1/monitoring/alerts/rules
Content-Type: application/json

{
  "name": "High CPU Usage",
  "description": "Alert when CPU usage exceeds 80%",
  "metric": "cpu_usage",
  "operator": ">",
  "threshold": 80.0,
  "severity": "warning",
  "enabled": true
}
```

#### Get Active Alerts

```bash
GET /api/v1/monitoring/alerts?status=active&limit=10
```

#### Acknowledge Alert

```bash
PUT /api/v1/monitoring/alerts/{alert_id}/acknowledge
Content-Type: application/json

{
  "message": "Investigating the issue"
}
```

## Monitoring Dashboard

### GET /monitoring/status

Returns comprehensive monitoring dashboard data:

```json
{
  "health": {
    "status": "ok",
    "components": {
      "database": "healthy",
      "redis": "healthy",
      "storage": "healthy"
    },
    "system": {
      "cpu": "healthy",
      "memory": "healthy",
      "disk": "healthy"
    }
  },
  "metrics": {
    "http": {
      "total_requests": 1234,
      "error_rate": 0.08,
      "average_response_time": 150
    },
    "business": {
      "active_users": 2000,
      "registrations_today": 50
    }
  },
  "alerts": {
    "active": 2,
    "critical": 0,
    "warning": 2
  },
  "system": {
    "cpu_usage": 25.5,
    "memory_usage": 45.2,
    "disk_usage": 60.1
  },
  "timestamp": "2025-11-02T16:30:00Z"
}
```

## Background Jobs

The monitoring system runs several background jobs:

### Health Check Job

Runs every 30 seconds to check the health of all components and store the results.

### Metrics Aggregation Job

Runs every 5 minutes to aggregate metrics and store them for historical analysis.

### Log Cleanup Job

Runs every hour to clean up old log files based on retention policies.

### System Monitoring Job

Runs every 2 minutes to collect system resource metrics.

### External Service Check Job

Runs every 3 minutes to check the health of external services.

## Monitoring Middleware

The monitoring middleware provides automatic request tracking:

### Request Timing Middleware

Tracks request duration and records metrics:

```go
// Apply to specific routes
router.Use(monitoring.RequestTiming())

// Or apply globally
router.Use(monitoring.RequestTiming())
```

### Error Tracking Middleware

Tracks errors and records error metrics:

```go
router.Use(monitoring.ErrorTracking())
```

### Security Event Logging Middleware

Logs security-related events:

```go
router.Use(monitoring.SecurityEventLogging())
```

### Combined Monitoring Middleware

Applies all monitoring middleware with correlation IDs:

```go
router.Use(monitoring.CombinedMonitoring())
```

## Storage

### Time-Series Data

Monitoring data is stored in Redis using time-series data structures:

- Health status data is stored with 24-hour retention
- Metrics data is stored with configurable retention periods
- Time-series data is stored in sorted sets for efficient querying

### Data Retention

Different types of monitoring data have different retention periods:

- Health status: 24 hours
- HTTP metrics: 24 hours
- Database metrics: 24 hours
- Cache metrics: 24 hours
- Business metrics: 7 days
- System metrics: 24 hours
- Time-series data: 7 days
- Alerts: 30 days

## Setup Guide

### 1. Configuration

Add monitoring configuration to your application config:

```go
Monitoring: config.MonitoringConfig{
    HealthCheck: config.HealthCheckConfig{
        Enabled: true,
        Timeout: 5 * time.Second,
        Interval: 30 * time.Second,
    },
    // ... other configuration
}
```

### 2. Initialize Services

Initialize monitoring services in your application startup:

```go
// Initialize monitoring services
healthService := services.NewHealthCheckService(cfg, cacheService)
metricsService := services.NewMetricsService(cfg, cacheService)
alertingService := services.NewAlertingService(cfg, cacheService)
storageService := services.NewMonitoringStorageService(cfg, cacheService)
jobsService := services.NewMonitoringJobsService(cfg, cacheService, healthService, metricsService, alertingService)

// Start background jobs
ctx := context.Background()
if err := jobsService.Start(ctx); err != nil {
    log.Fatal("Failed to start monitoring jobs:", err)
}
```

### 3. Setup Routes

Add monitoring routes to your router:

```go
// Create monitoring handlers
getHealthUseCase := monitoring.NewGetHealthStatusUseCase(healthService)
getDetailedHealthUseCase := monitoring.NewGetDetailedHealthUseCase(healthService)
getMetricsUseCase := monitoring.NewGetMetricsUseCase(metricsService)
getSystemStatusUseCase := monitoring.NewGetSystemStatusUseCase(healthService, metricsService)
getMonitoringDashboardUseCase := monitoring.NewGetMonitoringDashboardUseCase(healthService, metricsService)

monitoringHandler := handlers.NewMonitoringHandler(
    getHealthUseCase,
    getDetailedHealthUseCase,
    getMetricsUseCase,
    getSystemStatusUseCase,
    getMonitoringDashboardUseCase,
)

// Register monitoring routes
monitoringRoutes.RegisterRoutes(router, monitoringHandler)
```

### 4. Apply Middleware

Apply monitoring middleware to your routes:

```go
// Apply monitoring middleware
router.Use(monitoring.CombinedMonitoring())
```

### 5. Setup Alert Rules

Create default alert rules:

```go
// Create default alert rules
cpuRule := &services.AlertRule{
    ID:          "cpu-usage-high",
    Name:        "High CPU Usage",
    Description: "Alert when CPU usage exceeds 80%",
    Metric:      "cpu_usage",
    Operator:    ">",
    Threshold:   80.0,
    Severity:    services.AlertSeverityWarning,
    Enabled:     true,
}

if err := alertingService.CreateAlertRule(cpuRule); err != nil {
    log.Error("Failed to create CPU alert rule:", err)
}
```

## Best Practices

### 1. Monitoring Configuration

- Enable monitoring in production environments
- Configure appropriate retention periods based on your storage capacity
- Set reasonable thresholds for alerts to avoid alert fatigue

### 2. Alert Management

- Create specific alert rules for critical components
- Use appropriate severity levels for different types of alerts
- Regularly review and update alert rules

### 3. Performance Considerations

- Monitor the performance of the monitoring system itself
- Use efficient data structures for time-series data
- Implement proper cleanup of old monitoring data

### 4. Security

- Secure monitoring endpoints with appropriate authentication
- Use correlation IDs for request tracing
- Log security events for audit purposes

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Check metrics retention periods
   - Monitor time-series data growth
   - Implement proper cleanup jobs

2. **Slow Health Checks**
   - Increase timeout values
   - Check network connectivity to external services
   - Optimize database queries

3. **Missing Metrics**
   - Verify middleware is properly applied
   - Check metrics collection intervals
   - Ensure metrics service is initialized

4. **Alert Not Triggering**
   - Verify alert rule configuration
   - Check metric values against thresholds
   - Ensure alerting service is running

### Debugging

Enable debug logging for monitoring components:

```go
logger.SetLevel("debug")
```

Check monitoring service status:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/monitoring/status
```

## API Reference

### Health Endpoints

- `GET /health` - Basic health status
- `GET /health/detailed` - Detailed health status
- `GET /health/components/{component}` - Component health status

### Metrics Endpoints

- `GET /metrics` - Prometheus format metrics
- `GET /metrics/json` - JSON format metrics

### Monitoring Endpoints

- `GET /monitoring/status` - Monitoring dashboard data
- `GET /system/status` - System resource status

### Alert Management Endpoints

- `POST /api/v1/monitoring/alerts/rules` - Create alert rule
- `GET /api/v1/monitoring/alerts` - Get alerts
- `PUT /api/v1/monitoring/alerts/{id}/acknowledge` - Acknowledge alert
- `PUT /api/v1/monitoring/alerts/{id}/resolve` - Resolve alert

## Integration with External Systems

### Prometheus

The metrics endpoint (`/metrics`) provides data in Prometheus format, allowing integration with Prometheus for advanced monitoring and alerting:

```yaml
scrape_configs:
  - job_name: 'winkr-backend'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Grafana

Use the JSON metrics endpoint (`/metrics/json`) to create custom dashboards in Grafana:

```json
{
  "dashboard": {
    "title": "Winkr Backend Monitoring",
    "panels": [
      {
        "title": "HTTP Requests",
        "type": "graph",
        "targets": [
          {
            "expr": "http_requests_total",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      }
    ]
  }
}
```

### Log Aggregation

Structured logs with correlation IDs can be sent to log aggregation systems like ELK Stack or Splunk for centralized log analysis.

## Conclusion

The monitoring system provides comprehensive observability for the Winkr backend application. It includes health checks, metrics collection, alerting, and storage capabilities that help ensure the reliability and performance of the application.

For more information or support, please refer to the source code or contact the development team.