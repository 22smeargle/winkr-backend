package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// PaymentRoutes defines payment-related routes
type PaymentRoutes struct {
	paymentHandler *handlers.PaymentHandler
	authMiddleware middleware.AuthMiddleware
	rateLimiter    middleware.RateLimiter
}

// NewPaymentRoutes creates a new PaymentRoutes instance
func NewPaymentRoutes(
	paymentHandler *handlers.PaymentHandler,
	authMiddleware middleware.AuthMiddleware,
	rateLimiter middleware.RateLimiter,
) *PaymentRoutes {
	return &PaymentRoutes{
		paymentHandler: paymentHandler,
		authMiddleware: authMiddleware,
		rateLimiter:    rateLimiter,
	}
}

// RegisterRoutes registers payment routes
func (pr *PaymentRoutes) RegisterRoutes(router *gin.Engine) {
	logger.Info("Registering payment routes", nil)

	// Public routes (no authentication required)
	public := router.Group("/api/v1/payments")
	{
		// Webhook endpoint (no auth, special rate limiting)
		public.POST("/webhook", 
			pr.rateLimiter.WebhookRateLimit(),
			pr.paymentHandler.ProcessWebhook,
		)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1/payments")
	protected.Use(pr.authMiddleware.RequireAuth())
	{
		// Plan routes
		protected.GET("/plans",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetPlans,
		)
		protected.GET("/plans/:id",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetPlanByID,
		)

		// Subscription routes
		protected.GET("/subscription",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetSubscription,
		)
		protected.GET("/subscription/active",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetActiveSubscription,
		)
		protected.POST("/subscribe",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.Subscribe,
		)
		protected.POST("/subscription/cancel",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.CancelSubscription,
		)

		// Payment method routes
		protected.GET("/payment-methods",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetPaymentMethods,
		)
		protected.GET("/payment-methods/default",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.GetDefaultPaymentMethod,
		)
		protected.POST("/payment-methods",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.AddPaymentMethod,
		)
		protected.DELETE("/payment-methods/:id",
			pr.rateLimiter.PaymentRateLimit(),
			pr.paymentHandler.DeletePaymentMethod,
		)
	}

	logger.Info("Payment routes registered successfully", nil)
}