package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	httpServer "github.com/22smeargle/winkr-backend/internal/interfaces/http"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.App.Env)

	// Initialize database connection
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	logger.Info("Database connection established successfully")

	// Initialize Redis connection
	redisWrapper, err := redis.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis: %v", err)
	}
	defer func() {
		if err := redisWrapper.Close(); err != nil {
			logger.Error("Failed to close Redis connection: %v", err)
		}
	}()

	logger.Info("Redis connection established successfully")

	// Run database migrations
	database := postgres.NewDatabase(db, &cfg.Database)
	if err := database.RunMigrations(cfg.Database.MigrationsPath); err != nil {
		logger.Fatal("Failed to run database migrations: %v", err)
	}

	logger.Info("Database migrations completed successfully")

	// Create HTTP server with database and Redis
	server := httpServer.NewServer(cfg, db, redisWrapper.GetClient())

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server on port %d", cfg.App.Port)
		if err := server.Start(); err != nil {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}