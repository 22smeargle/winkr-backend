package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
	"github.com/22smeargle/winkr-backend/pkg/validator"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/auth"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/photo"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/verification"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/chat"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/payment"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/auth"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/postgres/repositories"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/storage"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/websocket"
)

// Server represents HTTP server
type Server struct {
	config   *config.Config
	engine   *gin.Engine
	server   *http.Server
	db       *gorm.DB
	redis    *redis.Client
	jwtUtils *utils.JWTUtils
	middlewareConfig *middleware.MiddlewareConfig
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *Server {
	// Create JWT utilities
	jwtUtils := utils.NewJWTUtils(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Load middleware configuration
	middlewareConfig := middleware.LoadMiddlewareConfig(cfg, redisClient, jwtUtils)

	// Create Gin engine
	engine := gin.New()

	// Add middleware in proper order
	// 1. Security middleware (first line of defense)
	engine.Use(middleware.Security(middlewareConfig.Security))
	
	// 2. CORS middleware
	engine.Use(middleware.CORS(middlewareConfig.CORS))
	
	// 3. Error handling middleware (for panic recovery)
	engine.Use(middleware.ErrorHandler(middlewareConfig.ErrorHandler))
	
	// 4. Request ID middleware
	engine.Use(middleware.RequestID(middlewareConfig.Logging.RequestIDHeader))
	
	// 5. Logging middleware
	engine.Use(middleware.Logging(middlewareConfig.Logging))
	
	// 6. Rate limiting middleware
	engine.Use(middleware.RateLimiter(middlewareConfig.RateLimit))
	
	// 7. Validation middleware
	engine.Use(middleware.Validation(middlewareConfig.Validation))
	
	// 8. Authentication middleware
	engine.Use(middleware.Auth(middlewareConfig.Auth))

	// Create server instance
	server := &Server{
		config:          cfg,
		engine:          engine,
		db:              db,
		redis:           redisClient,
		jwtUtils:        jwtUtils,
		middlewareConfig: middlewareConfig,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.App.Port),
			Handler:      engine,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.Info("Starting HTTP server on port %d", s.config.App.Port)
	
	// Setup all routes
	s.SetupRoutes()
	
	// Add legacy health check routes for backward compatibility
	s.engine.GET("/health", s.healthCheck)
	s.engine.GET("/health/db", s.databaseHealthCheck)

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down HTTP server...")
	
	return s.server.Shutdown(ctx)
}

// GetEngine returns the Gin engine
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// GetJWTUtils returns the JWT utilities instance
func (s *Server) GetJWTUtils() *utils.JWTUtils {
	return s.jwtUtils
}

// GetRedis returns the Redis client
func (s *Server) GetRedis() *redis.Client {
	return s.redis
}

// GetMiddlewareConfig returns the middleware configuration
func (s *Server) GetMiddlewareConfig() *middleware.MiddlewareConfig {
	return s.middlewareConfig
}

// SetupRoutes sets up all application routes
func (s *Server) SetupRoutes() {
	// Initialize repositories
	userRepo := repositories.NewUserRepository(s.db)
	photoRepo := repositories.NewPhotoRepository(s.db)
	verificationRepo := repositories.NewVerificationRepository(s.db)
	messageRepo := repositories.NewMessageRepository(s.db)
	matchRepo := repositories.NewMatchRepository(s.db)
	subscriptionRepo := repositories.NewSubscriptionRepository(s.db)
	paymentRepo := repositories.NewPaymentRepository(s.db)
	paymentMethodRepo := repositories.NewPaymentMethodRepository(s.db)
	refundRepo := repositories.NewRefundRepository(s.db)
	invoiceRepo := repositories.NewInvoiceRepository(s.db)
	webhookEventRepo := repositories.NewWebhookEventRepository(s.db)
	
	// Initialize services
	tokenManager := auth.NewTokenManager(s.jwtUtils)
	cacheService := cache.NewCacheService(s.redis)
	sessionManager := cache.NewSessionManager(s.redis, tokenManager)
	rateLimiter := cache.NewRateLimiter(s.redis)
	verificationService := services.NewVerificationService(cacheService, rateLimiter)
	
	// Initialize chat services
	messageService := services.NewMessageService(&s.config.Chat, cacheService)
	chatSecurityService := services.NewChatSecurityService(&s.config.Chat.Security, cacheService)
	chatCacheService := services.NewChatCacheService(s.redis, &s.config.Chat.Cache)
	
	// Initialize WebSocket connection manager
	connectionManager := websocket.NewConnectionManager(
		chatCacheService,
		messageService,
		chatSecurityService,
		&s.config.Chat.WebSocket,
	)
	
	// Initialize AI service
	aiService := external.NewAIService(&s.config.Verification.AIService)
	
	// Initialize Stripe service
	stripeService := stripe.NewStripeService(&s.config.Stripe)
	
	// Initialize document service
	documentService := services.NewDocumentService(&s.config.Verification.DocumentProcessing)
	
	// Initialize verification workflow service
	verificationWorkflowService := services.NewVerificationWorkflowService(
		verificationRepo,
		userRepo,
		aiService,
		documentService,
		storageService,
		cacheService,
		&s.config.Verification,
	)
	
	// Initialize storage service
	storageService, err := storage.NewS3Storage(&s.config.Storage)
	if err != nil {
		logger.Fatal("Failed to initialize storage service: %v", err)
	}
	
	// Initialize image processing service
	imageProcessor := services.NewImageProcessor(&s.config.Storage)
	
	// Initialize use cases
	registerUseCase := auth.NewRegisterUseCase(userRepo, tokenManager, sessionManager, verificationService)
	loginUseCase := auth.NewLoginUseCase(userRepo, tokenManager, sessionManager, rateLimiter)
	refreshUseCase := auth.NewRefreshTokenUseCase(tokenManager, sessionManager)
	logoutUseCase := auth.NewLogoutUseCase(tokenManager, sessionManager)
	passwordResetUseCase := auth.NewPasswordResetUseCase(userRepo, verificationService, rateLimiter)
	confirmPasswordResetUseCase := auth.NewConfirmPasswordResetUseCase(userRepo, verificationService, rateLimiter)
	emailVerificationUseCase := auth.NewEmailVerificationUseCase(userRepo, verificationService, rateLimiter)
	getProfileUseCase := auth.NewGetProfileUseCase(userRepo)
	getSessionsUseCase := auth.NewGetSessionsUseCase(sessionManager)
	
	// Initialize photo use cases
	uploadPhotoUseCase := photo.NewUploadPhotoUseCase(photoRepo, storageService, imageProcessor)
	deletePhotoUseCase := photo.NewDeletePhotoUseCase(photoRepo, storageService)
	getUploadURLUseCase := photo.NewGetUploadURLUseCase(photoRepo, storageService, s.config.Storage.MaxFileSize, s.config.Storage.AllowedTypes)
	getDownloadURLUseCase := photo.NewGetDownloadURLUseCase(photoRepo, storageService)
	setPrimaryPhotoUseCase := photo.NewSetPrimaryPhotoUseCase(photoRepo)
	markPhotoViewedUseCase := photo.NewMarkPhotoViewedUseCase(photoRepo)
	
	// Initialize verification use cases
	requestSelfieVerificationUseCase := verification.NewRequestSelfieVerificationUseCase(verificationRepo, userRepo, verificationWorkflowService, rateLimiter)
	submitSelfieVerificationUseCase := verification.NewSubmitSelfieVerificationUseCase(verificationRepo, userRepo, verificationWorkflowService, storageService, rateLimiter)
	getVerificationStatusUseCase := verification.NewGetVerificationStatusUseCase(verificationRepo, userRepo)
	requestDocumentVerificationUseCase := verification.NewRequestDocumentVerificationUseCase(verificationRepo, userRepo, verificationWorkflowService, rateLimiter)
	submitDocumentVerificationUseCase := verification.NewSubmitDocumentVerificationUseCase(verificationRepo, userRepo, verificationWorkflowService, storageService, rateLimiter)
	processVerificationResultUseCase := verification.NewProcessVerificationResultUseCase(verificationRepo, userRepo, verificationWorkflowService)
	getPendingVerificationsUseCase := verification.NewGetPendingVerificationsUseCase(verificationRepo)
	
	// Initialize chat use cases
	getConversationsUseCase := chat.NewGetConversationsUseCase(messageRepo, chatCacheService)
	getMessagesUseCase := chat.NewGetMessagesUseCase(messageRepo, chatCacheService)
	sendMessageUseCase := chat.NewSendMessageUseCase(messageRepo, messageService, chatSecurityService, chatCacheService, connectionManager)
	markMessagesReadUseCase := chat.NewMarkMessagesReadUseCase(messageRepo, chatCacheService, connectionManager)
	deleteMessageUseCase := chat.NewDeleteMessageUseCase(messageRepo, chatSecurityService, chatCacheService, connectionManager)
	startConversationUseCase := chat.NewStartConversationUseCase(messageRepo, matchRepo, chatCacheService, connectionManager)
	
	// Initialize payment use cases
	getPlansUseCase := payment.NewGetPlansUseCase(stripeService)
	getPlanByIDUseCase := payment.NewGetPlanByIDUseCase(stripeService)
	subscribeUseCase := payment.NewSubscribeUseCase(subscriptionRepo, paymentRepo, userRepo, stripeService, cacheService)
	getSubscriptionUseCase := payment.NewGetSubscriptionUseCase(subscriptionRepo, userRepo, cacheService)
	getActiveSubscriptionUseCase := payment.NewGetActiveSubscriptionUseCase(subscriptionRepo, userRepo, cacheService)
	cancelSubscriptionUseCase := payment.NewCancelSubscriptionUseCase(subscriptionRepo, stripeService, cacheService)
	updateSubscriptionUseCase := payment.NewUpdateSubscriptionUseCase(subscriptionRepo, stripeService, cacheService)
	addPaymentMethodUseCase := payment.NewAddPaymentMethodUseCase(paymentMethodRepo, userRepo, stripeService, cacheService)
	getPaymentMethodsUseCase := payment.NewGetPaymentMethodsUseCase(paymentMethodRepo, userRepo, cacheService)
	getDefaultPaymentMethodUseCase := payment.NewGetDefaultPaymentMethodUseCase(paymentMethodRepo, userRepo, cacheService)
	deletePaymentMethodUseCase := payment.NewDeletePaymentMethodUseCase(paymentMethodRepo, stripeService, cacheService)
	processWebhookUseCase := payment.NewProcessWebhookUseCase(webhookEventRepo, subscriptionRepo, paymentRepo, paymentMethodRepo, refundRepo, invoiceRepo, userRepo, stripeService, cacheService)
	
	// Initialize subscription service
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, stripeService, cacheService)
	
	// Initialize validators
	authValidator := validator.NewAuthValidator()
	
	// Initialize middleware
	authRateLimiter := middleware.NewAuthRateLimiter(rateLimiter)
	suspiciousDetector := middleware.NewSuspiciousActivityDetector(rateLimiter)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(
		registerUseCase,
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		passwordResetUseCase,
		confirmPasswordResetUseCase,
		emailVerificationUseCase,
		getProfileUseCase,
		getSessionsUseCase,
		s.jwtUtils,
		authValidator,
		authRateLimiter,
		suspiciousDetector,
	)
	
	photoHandler := handlers.NewPhotoHandler(
		uploadPhotoUseCase,
		deletePhotoUseCase,
		getUploadURLUseCase,
		getDownloadURLUseCase,
		setPrimaryPhotoUseCase,
		markPhotoViewedUseCase,
		s.jwtUtils,
	)
	
	// Initialize verification handlers
	verificationHandler := handlers.NewVerificationHandler(
		requestSelfieVerificationUseCase,
		submitSelfieVerificationUseCase,
		getVerificationStatusUseCase,
		requestDocumentVerificationUseCase,
		submitDocumentVerificationUseCase,
		s.jwtUtils,
	)
	
	adminVerificationHandler := handlers.NewAdminVerificationHandler(
		processVerificationResultUseCase,
		getPendingVerificationsUseCase,
		s.jwtUtils,
	)
	
	// Initialize chat handler
	chatHandler := handlers.NewChatHandler(
		getConversationsUseCase,
		getMessagesUseCase,
		sendMessageUseCase,
		markMessagesReadUseCase,
		deleteMessageUseCase,
		startConversationUseCase,
		connectionManager,
		s.jwtUtils,
	)
	
	// Initialize payment handler
	paymentHandler := handlers.NewPaymentHandler(
		getPlansUseCase,
		getPlanByIDUseCase,
		subscribeUseCase,
		getSubscriptionUseCase,
		getActiveSubscriptionUseCase,
		cancelSubscriptionUseCase,
		updateSubscriptionUseCase,
		addPaymentMethodUseCase,
		getPaymentMethodsUseCase,
		getDefaultPaymentMethodUseCase,
		deletePaymentMethodUseCase,
		processWebhookUseCase,
		subscriptionService,
		s.jwtUtils,
	)
	
	// Initialize routes
	authRoutes := routes.NewAuthRoutes(
		authHandler,
		s.middlewareConfig.Security,
		s.middlewareConfig.RateLimit,
		s.middlewareConfig.CSRF,
	)
	
	photoRoutes := routes.NewPhotoRoutes(
		photoHandler,
		rateLimiter,
	)
	
	verificationRoutes := routes.NewVerificationRoutes(
		verificationHandler,
		adminVerificationHandler,
		rateLimiter,
	)
	
	// Initialize chat routes
	chatRoutes := routes.NewChatRoutes(
		chatHandler,
		rateLimiter,
		&s.config.Chat.RateLimit,
	)
	
	// Initialize payment routes
	paymentRoutes := routes.NewPaymentRoutes(
		paymentHandler,
		rateLimiter,
		&s.config.Stripe,
	)
	
	// Create API v1 group
	v1 := s.engine.Group("/api/v1")
	
	// Register auth routes
	authRoutes.RegisterRoutes(v1, s.redis)
	
	// Register photo routes
	photoRoutes.RegisterRoutes(v1, s.redis)
	
	// Register verification routes
	verificationRoutes.RegisterRoutes(v1, s.redis)
	
	// Register chat routes
	chatRoutes.RegisterRoutes(v1, s.redis)
	
	// Register payment routes
	paymentRoutes.RegisterRoutes(v1, s.redis)
	
	// Register WebSocket endpoint
	connectionManager.RegisterWebSocketRoutes(s.engine)
	
	// Register health routes
	healthRoutes := routes.NewHealthRoutes()
	healthRoutes.RegisterRoutes(v1)
	
	logger.Info("Routes registered successfully")
}

// GetServerInfo returns server information
func (s *Server) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"port":     s.config.App.Port,
		"host":     s.config.App.Host,
		"env":      s.config.App.Env,
		"version":  "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
	}
}

// RegisterSwagger registers Swagger documentation
func (s *Server) RegisterSwagger() {
	// TODO: Implement Swagger documentation
	// swaggerInfo := docs.SwaggerInfo
	// s.engine.GET("/swagger/*", gin.WrapH(swaggerFiles.Handler))
	// s.engine.GET("/swagger.json", gin.WrapH(swaggerFiles.Handler))
	
	logger.Info("Swagger documentation registered")
}

// healthCheck handles basic health check
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// databaseHealthCheck handles database health check
func (s *Server) databaseHealthCheck(c *gin.Context) {
	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		logger.Error("Failed to get database instance", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"timestamp": time.Now().UTC(),
			"database":  "unavailable",
			"error":     "Failed to get database instance",
		})
		return
	}

	// Ping database
	if err := sqlDB.Ping(); err != nil {
		logger.Error("Database ping failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"timestamp": time.Now().UTC(),
			"database":  "unavailable",
			"error":     err.Error(),
		})
		return
	}

	// Get database stats
	stats := sqlDB.Stats()
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"database":  "available",
		"stats": gin.H{
			"open_connections":     stats.OpenConnections,
			"in_use":            stats.InUse,
			"idle":              stats.Idle,
			"wait_count":         stats.WaitCount,
			"wait_duration":      stats.WaitDuration.String(),
			"max_idle_closed":    stats.MaxIdleClosed,
			"max_idle_time_closed": fmt.Sprintf("%d", stats.MaxIdleTimeClosed),
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		},
	})
}