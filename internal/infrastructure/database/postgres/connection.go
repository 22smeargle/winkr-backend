package postgres

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// Database represents database connection
type Database struct {
	DB     *gorm.DB
	Config *config.DatabaseConfig
}

// NewConnection creates a new database connection
func NewConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := buildDSN(cfg)
	
	db, err := gorm.Open(postgresDriver.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:              true,
	})
	if err != nil {
		logger.Error("Failed to connect to database", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB", err)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings from config
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

// buildDSN builds data source name for PostgreSQL
func buildDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
	)
}

// buildMigrateDSN builds data source name for migrations
func buildMigrateDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB for closing", err)
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		logger.Error("Failed to close database connection", err)
		return err
	}

	logger.Info("Database connection closed")
	return nil
}

// Ping checks if the database connection is alive
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetDB returns the underlying GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// AutoMigrate runs auto migration for all models
func (d *Database) AutoMigrate(models ...interface{}) error {
	err := d.DB.AutoMigrate(models...)
	if err != nil {
		logger.Error("Failed to run auto migration", err)
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	logger.Info("Database auto migration completed successfully")
	return nil
}

// BeginTransaction starts a new database transaction
func (d *Database) BeginTransaction() *gorm.DB {
	return d.DB.Begin()
}

// Health checks the database health
func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	// Check if we can ping the database
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if there are any idle connections
	stats := sqlDB.Stats()
	if stats.OpenConnections > 0 && stats.Idle == 0 {
		logger.Warn("Database has open connections but no idle connections")
	}

	return nil
}

// GetStats returns database connection statistics
func (d *Database) GetStats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":             stats.InUse,
		"idle":               stats.Idle,
		"wait_count":         stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": fmt.Sprintf("%d", stats.MaxIdleTimeClosed),
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// MigrateStatus represents the status of a migration
type MigrateStatus struct {
	Version     string    `json:"version"`
	Dirty       bool      `json:"dirty"`
	AppliedAt   time.Time `json:"applied_at"`
	Description string    `json:"description"`
}

// RunMigrations runs database migrations using golang-migrate
func (d *Database) RunMigrations(migrationsPath string) error {
	// Get underlying SQL DB for migrations
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Create driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get absolute path to migrations directory
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to run database migrations", err)
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// NewDatabase creates a new Database instance
func NewDatabase(db *gorm.DB, cfg *config.DatabaseConfig) *Database {
	return &Database{
		DB:     db,
		Config: cfg,
	}
}

// GetMigrationVersion returns the current migration version
func (d *Database) GetMigrationVersion() (*MigrateStatus, error) {
	// Get underlying SQL DB for migrations
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Create driver instance
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get absolute path to migrations directory
	absPath, err := filepath.Abs("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return &MigrateStatus{
				Version:     "no migrations",
				Dirty:       false,
				AppliedAt:   time.Now(),
				Description: "No migrations have been applied",
			}, nil
		}
		return nil, fmt.Errorf("failed to get migration version: %w", err)
	}

	return &MigrateStatus{
		Version:     fmt.Sprintf("%d", version),
		Dirty:       dirty,
		AppliedAt:   time.Now(),
		Description: fmt.Sprintf("Migration version %d", version),
	}, nil
}

// CreateMigration creates a new migration file
func (d *Database) CreateMigration(name string) error {
	// This would typically create a new migration file
	// For now, just log that this feature is not implemented
	logger.Info("CreateMigration not implemented yet")
	return nil
}