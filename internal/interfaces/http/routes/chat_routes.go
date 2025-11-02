package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
)

// ChatRoutes defines chat-related routes
type ChatRoutes struct {
	handler *handlers.ChatHandler
}

// NewChatRoutes creates new chat routes
func NewChatRoutes(handler *handlers.ChatHandler) *ChatRoutes {
	return &ChatRoutes{
		handler: handler,
	}
}

// RegisterRoutes registers chat routes with a router
func (r *ChatRoutes) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc, rateLimitMiddleware gin.HandlerFunc) {
	// Apply authentication and rate limiting to all chat routes
	chatGroup := router.Group("/api/v1/chats")
	chatGroup.Use(authMiddleware)
	chatGroup.Use(rateLimitMiddleware)

	{
		// GET /api/v1/chats - Get user's conversations
		chatGroup.GET("", r.handler.GetConversations)

		// GET /api/v1/chats/:id/messages - Get messages in a conversation
		chatGroup.GET("/:id/messages", r.handler.GetMessages)

		// POST /api/v1/chats/:id/messages - Send message to conversation
		chatGroup.POST("/:id/messages", r.handler.SendMessage)

		// POST /api/v1/chats/:id/read - Mark messages as read
		chatGroup.POST("/:id/read", r.handler.MarkMessagesAsRead)

		// DELETE /api/v1/chats/:id/messages/:messageId - Delete a message
		chatGroup.DELETE("/:id/messages/:messageId", r.handler.DeleteMessage)

		// POST /api/v1/chats/:id/ephemeral-photos - Send ephemeral photo message
		chatGroup.POST("/:id/ephemeral-photos", r.handler.SendEphemeralPhotoMessage)

		// GET /api/v1/chats/:id/messages/:messageId/ephemeral-photo - Get ephemeral photo message
		chatGroup.GET("/:id/messages/:messageId/ephemeral-photo", r.handler.GetEphemeralPhotoMessage)

		// POST /api/v1/chats/start - Start a new conversation
		chatGroup.POST("/start", r.handler.StartConversation)
	}

	// WebSocket endpoint for real-time messaging
	// Apply authentication middleware
	wsGroup := router.Group("/api/v1/ws")
	wsGroup.Use(authMiddleware)
	{
		// WebSocket upgrade endpoint
		wsGroup.GET("", r.handler.WebSocketUpgrade)
	}
}

// RegisterRoutesWithCustomMiddleware registers chat routes with custom middleware
func (r *ChatRoutes) RegisterRoutesWithCustomMiddleware(
	router *gin.Engine,
	authMiddleware gin.HandlerFunc,
	rateLimitMiddleware gin.HandlerFunc,
	customMiddleware ...gin.HandlerFunc,
) {
	// Apply authentication and rate limiting to all chat routes
	chatGroup := router.Group("/api/v1/chats")
	chatGroup.Use(authMiddleware)
	chatGroup.Use(rateLimitMiddleware)
	
	// Apply any additional custom middleware
	for _, middleware := range customMiddleware {
		chatGroup.Use(middleware)
	}

	{
		// GET /api/v1/chats - Get user's conversations
		chatGroup.GET("", r.handler.GetConversations)

		// GET /api/v1/chats/:id/messages - Get messages in a conversation
		chatGroup.GET("/:id/messages", r.handler.GetMessages)

		// POST /api/v1/chats/:id/messages - Send message to conversation
		chatGroup.POST("/:id/messages", r.handler.SendMessage)

		// POST /api/v1/chats/:id/read - Mark messages as read
		chatGroup.POST("/:id/read", r.handler.MarkMessagesAsRead)

		// DELETE /api/v1/chats/:id/messages/:messageId - Delete a message
		chatGroup.DELETE("/:id/messages/:messageId", r.handler.DeleteMessage)

		// POST /api/v1/chats/:id/ephemeral-photos - Send ephemeral photo message
		chatGroup.POST("/:id/ephemeral-photos", r.handler.SendEphemeralPhotoMessage)

		// GET /api/v1/chats/:id/messages/:messageId/ephemeral-photo - Get ephemeral photo message
		chatGroup.GET("/:id/messages/:messageId/ephemeral-photo", r.handler.GetEphemeralPhotoMessage)

		// POST /api/v1/chats/start - Start a new conversation
		chatGroup.POST("/start", r.handler.StartConversation)
	}

	// WebSocket endpoint for real-time messaging
	// Apply authentication middleware
	wsGroup := router.Group("/api/v1/ws")
	wsGroup.Use(authMiddleware)
	
	// Apply any additional custom middleware to WebSocket
	for _, middleware := range customMiddleware {
		wsGroup.Use(middleware)
	}
	
	{
		// WebSocket upgrade endpoint
		wsGroup.GET("", r.handler.WebSocketUpgrade)
	}
}

// GetRouteInfo returns information about chat routes
func (r *ChatRoutes) GetRouteInfo() map[string]interface{} {
	return map[string]interface{}{
		"base_path": "/api/v1/chats",
		"websocket_path": "/api/v1/ws",
		"endpoints": []map[string]interface{}{
			{
				"method": "GET",
				"path":   "",
				"description": "Get user's conversations",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "GET",
				"path":   "/:id/messages",
				"description": "Get messages in a conversation",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "POST",
				"path":   "/:id/messages",
				"description": "Send message to conversation",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "POST",
				"path":   "/:id/read",
				"description": "Mark messages as read",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "DELETE",
				"path":   "/:id/messages/:messageId",
				"description": "Delete a message",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "POST",
				"path":   "/:id/ephemeral-photos",
				"description": "Send ephemeral photo message",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "GET",
				"path":   "/:id/messages/:messageId/ephemeral-photo",
				"description": "Get ephemeral photo message",
				"auth_required": true,
				"rate_limited": true,
			},
			{
				"method": "POST",
				"path":   "/start",
				"description": "Start a new conversation",
				"auth_required": true,
				"rate_limited": true,
			},
		},
		"websocket_endpoints": []map[string]interface{}{
			{
				"method": "GET",
				"path": "/api/v1/ws",
				"description": "WebSocket upgrade for real-time messaging",
				"auth_required": true,
				"rate_limited": false,
			},
		},
	}
}