package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// VerificationRoutes defines verification routes
type VerificationRoutes struct {
	handler *handlers.VerificationHandler
}

// NewVerificationRoutes creates new verification routes
func NewVerificationRoutes(handler *handlers.VerificationHandler) *VerificationRoutes {
	return &VerificationRoutes{
		handler: handler,
	}
}

// RegisterRoutes registers verification routes
func (r *VerificationRoutes) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	verification := router.Group("/verify")
	verification.Use(authMiddleware) // All verification routes require authentication

	// Selfie verification routes
	verification.POST("/selfie/request", r.handler.RequestSelfieVerification)
	verification.POST("/selfie/submit", r.handler.SubmitSelfieVerification)
	verification.GET("/selfie/status", r.handler.GetVerificationStatus)

	// Document verification routes
	verification.POST("/document/request", r.handler.RequestDocumentVerification)
	verification.POST("/document/submit", r.handler.SubmitDocumentVerification)
	verification.GET("/document/status", r.handler.GetVerificationStatus)

	logger.Info("Verification routes registered")
}

// AdminVerificationRoutes defines admin verification routes
type AdminVerificationRoutes struct {
	handler *handlers.VerificationHandler
}

// NewAdminVerificationRoutes creates new admin verification routes
func NewAdminVerificationRoutes(handler *handlers.VerificationHandler) *AdminVerificationRoutes {
	return &AdminVerificationRoutes{
		handler: handler,
	}
}

// RegisterAdminRoutes registers admin verification routes
func (r *AdminVerificationRoutes) RegisterAdminRoutes(router *gin.RouterGroup, adminAuthMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin/verifications")
	admin.Use(adminAuthMiddleware) // All admin verification routes require admin authentication

	// Admin verification management routes
	admin.GET("", r.handler.GetPendingVerifications)
	admin.POST("/:id/approve", r.handler.ApproveVerification)
	admin.POST("/:id/reject", r.handler.RejectVerification)
	admin.GET("/:id", r.handler.GetVerificationDetails)

	logger.Info("Admin verification routes registered")
}