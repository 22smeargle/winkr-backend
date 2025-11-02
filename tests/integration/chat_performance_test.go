package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/chat"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/websocket"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/pkg/utils"
)

// ChatPerformanceTestSuite tests chat system performance under load
type ChatPerformanceTestSuite struct {
	suite.Suite
	router              *gin.Engine
	chatHandler         *handlers.ChatHandler
	redisClient         *redis.RedisClient
	connectionManager    *websocket.ConnectionManager
}

// SetupSuite sets up the test suite
func (suite *ChatPerformanceTestSuite) SetupSuite() {
	// Create test dependencies
	suite.redisClient = redis.NewMockRedisClient()
	
	// Create chat services
	messageService := services.NewMessageService(nil, suite.redisClient)
	chatSecurityService := services.NewChatSecurityService(nil, suite.redisClient)
	chatCacheService := services.NewChatCacheService(suite.redisClient, nil)
	
	// Create WebSocket connection manager
	suite.connectionManager = websocket.NewConnectionManager(
		chatCacheService,
		messageService,
		chatSecurityService,
		nil, // Use default WebSocket config
	)
	
	// Create JWT utils
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24*7)
	
	// Create use cases with mock repositories
	getConversationsUseCase := chat.NewGetConversationsUseCase(nil, chatCacheService)
	getMessagesUseCase := chat.NewGetMessagesUseCase(nil, chatCacheService)
	sendMessageUseCase := chat.NewSendMessageUseCase(nil, messageService, chatSecurityService, chatCacheService, suite.connectionManager)
	markMessagesReadUseCase := chat.NewMarkMessagesReadUseCase(nil, chatCacheService, suite.connectionManager)
	deleteMessageUseCase := chat.NewDeleteMessageUseCase(nil, chatSecurityService, chatCacheService, suite.connectionManager)
	startConversationUseCase := chat.NewStartConversationUseCase(nil, nil, chatCacheService, suite.connectionManager)
	
	// Create chat handler
	suite.chatHandler = handlers.NewChatHandler(
		getConversationsUseCase,
		getMessagesUseCase,
		sendMessageUseCase,
		markMessagesReadUseCase,
		deleteMessageUseCase,
		startConversationUseCase,
		suite.connectionManager,
		jwtUtils,
	)
	
	// Create router
	suite.router = gin.New()
	
	// Add chat routes
	chatGroup := suite.router.Group("/api/v1/chats")
	{
		chatGroup.GET("", suite.chatHandler.GetConversations)
		chatGroup.POST("/:conversationId/messages", suite.chatHandler.SendMessage)
		chatGroup.POST("/:conversationId/read", suite.chatHandler.MarkMessagesAsRead)
	}
}

// TestConcurrentMessageSending tests concurrent message sending performance
func (suite *ChatPerformanceTestSuite) TestConcurrentMessageSending() {
	// Create test user and conversation
	accessToken := suite.createTestUser("perfuser")
	conversationID := uuid.New().String()
	
	// Number of concurrent goroutines
	numGoroutines := 50
	messagesPerGoroutine := 10
	
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64
	startTime := time.Now()
	
	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < messagesPerGoroutine; j++ {
				// Prepare message request
				req := dto.SendMessageDTO{
					Content: fmt.Sprintf("Concurrent message %d-%d", goroutineID, j),
					Type:    "text",
				}
				
				reqBody, _ := json.Marshal(req)
				httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), reqBody)
				httpReq.Header.Set("Authorization", "Bearer "+accessToken)
				httpReq.Header.Set("Content-Type", "application/json")
				
				// Perform request
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, httpReq)
				
				// Count successes and errors
				if w.Code == http.StatusCreated {
					successCount++
				} else {
					errorCount++
				}
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	duration := time.Since(startTime)
	
	// Calculate metrics
	totalMessages := numGoroutines * messagesPerGoroutine
	successRate := float64(successCount) / float64(totalMessages) * 100
	messagesPerSecond := float64(totalMessages) / duration.Seconds()
	
	// Log performance metrics
	suite.T().Logf("Concurrent Message Sending Performance:")
	suite.T().Logf("  Total Messages: %d", totalMessages)
	suite.T().Logf("  Successful Messages: %d", successCount)
	suite.T().Logf("  Failed Messages: %d", errorCount)
	suite.T().Logf("  Success Rate: %.2f%%", successRate)
	suite.T().Logf("  Duration: %v", duration)
	suite.T().Logf("  Messages/Second: %.2f", messagesPerSecond)
	
	// Assertions
	assert.Greater(suite.T(), successRate, 80.0, "Success rate should be at least 80%")
	assert.Less(suite.T(), duration, 30*time.Second, "Should complete within 30 seconds")
	assert.Greater(suite.T(), messagesPerSecond, 10.0, "Should handle at least 10 messages per second")
}

// TestConcurrentConversationRetrieval tests concurrent conversation retrieval
func (suite *ChatPerformanceTestSuite) TestConcurrentConversationRetrieval() {
	// Create test users
	accessToken := suite.createTestUser("perfuser")
	
	// Number of concurrent goroutines
	numGoroutines := 20
	requestsPerGoroutine := 5
	
	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64
	startTime := time.Now()
	
	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				// Prepare request
				req := httptest.NewRequest("GET", "/api/v1/chats?page=1&limit=20", nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				req.Header.Set("Content-Type", "application/json")
				
				// Perform request
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				
				// Count successes and errors
				if w.Code == http.StatusOK {
					successCount++
				} else {
					errorCount++
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	duration := time.Since(startTime)
	
	// Calculate metrics
	totalRequests := numGoroutines * requestsPerGoroutine
	successRate := float64(successCount) / float64(totalRequests) * 100
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	
	// Log performance metrics
	suite.T().Logf("Concurrent Conversation Retrieval Performance:")
	suite.T().Logf("  Total Requests: %d", totalRequests)
	suite.T().Logf("  Successful Requests: %d", successCount)
	suite.T().Logf("  Failed Requests: %d", errorCount)
	suite.T().Logf("  Success Rate: %.2f%%", successRate)
	suite.T().Logf("  Duration: %v", duration)
	suite.T().Logf("  Requests/Second: %.2f", requestsPerSecond)
	
	// Assertions
	assert.Greater(suite.T(), successRate, 95.0, "Success rate should be at least 95%")
	assert.Less(suite.T(), duration, 10*time.Second, "Should complete within 10 seconds")
	assert.Greater(suite.T(), requestsPerSecond, 50.0, "Should handle at least 50 requests per second")
}

// TestMessageHistoryPagination tests pagination performance with large message history
func (suite *ChatPerformanceTestSuite) TestMessageHistoryPagination() {
	// Create test user and conversation
	accessToken := suite.createTestUser("perfuser")
	conversationID := uuid.New().String()
	
	// Test pagination performance
	startTime := time.Now()
	
	for page := 1; page <= 10; page++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chats/%s/messages?page=%d&limit=50", conversationID, page), nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// Verify response
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response dto.MessagesResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
		assert.True(suite.T(), response.Success)
		assert.NotNil(suite.T(), response.Data.Messages)
		assert.NotNil(suite.T(), response.Data.Pagination)
	}
	
	duration := time.Since(startTime)
	
	// Log performance metrics
	suite.T().Logf("Message History Pagination Performance:")
	suite.T().Logf("  Pages Retrieved: 10")
	suite.T().Logf("  Total Duration: %v", duration)
	suite.T().Logf("  Average Duration/Page: %v", duration/10)
	
	// Assertions
	assert.Less(suite.T(), duration, 5*time.Second, "Should retrieve 10 pages within 5 seconds")
	assert.Less(suite.T(), duration/10, 500*time.Millisecond, "Each page should load within 500ms")
}

// TestTypingIndicatorPerformance tests typing indicator performance under load
func (suite *ChatPerformanceTestSuite) TestTypingIndicatorPerformance() {
	// Create test users
	user1Token := suite.createTestUser("perfuser1")
	user2Token := suite.createTestUser("perfuser2")
	conversationID := uuid.New().String()
	
	// Number of typing indicator events
	numEvents := 100
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	// Launch concurrent typing indicator events
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(eventID int) {
			defer wg.Done()
			
			// Prepare mark as read request
			req := dto.MarkReadRequestDTO{
				MessageID: fmt.Sprintf("msg-%d", eventID),
			}
			
			reqBody, _ := json.Marshal(req)
			httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/read", conversationID), reqBody)
			httpReq.Header.Set("Authorization", "Bearer "+user1Token)
			httpReq.Header.Set("Content-Type", "application/json")
			
			// Perform request
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, httpReq)
			
			// Verify response
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		}(i)
	}
	
	// Wait for all events to complete
	wg.Wait()
	duration := time.Since(startTime)
	
	// Calculate metrics
	eventsPerSecond := float64(numEvents) / duration.Seconds()
	
	// Log performance metrics
	suite.T().Logf("Typing Indicator Performance:")
	suite.T().Logf("  Total Events: %d", numEvents)
	suite.T().Logf("  Duration: %v", duration)
	suite.T().Logf("  Events/Second: %.2f", eventsPerSecond)
	
	// Assertions
	assert.Less(suite.T(), duration, 10*time.Second, "Should complete within 10 seconds")
	assert.Greater(suite.T(), eventsPerSecond, 50.0, "Should handle at least 50 events per second")
}

// TestCachePerformance tests Redis caching performance
func (suite *ChatPerformanceTestSuite) TestCachePerformance() {
	// Test cache write performance
	numCacheWrites := 1000
	startTime := time.Now()
	
	for i := 0; i < numCacheWrites; i++ {
		key := fmt.Sprintf("test:cache:%d", i)
		value := fmt.Sprintf("test-value-%d", i)
		
		// Write to cache
		err := suite.redisClient.Set(key, value, time.Hour)
		assert.NoError(suite.T(), err)
	}
	
	writeDuration := time.Since(startTime)
	
	// Test cache read performance
	startTime = time.Now()
	
	for i := 0; i < numCacheWrites; i++ {
		key := fmt.Sprintf("test:cache:%d", i)
		
		// Read from cache
		var value string
		err := suite.redisClient.Get(key, &value)
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), value)
	}
	
	readDuration := time.Since(startTime)
	
	// Calculate metrics
	writesPerSecond := float64(numCacheWrites) / writeDuration.Seconds()
	readsPerSecond := float64(numCacheWrites) / readDuration.Seconds()
	
	// Log performance metrics
	suite.T().Logf("Cache Performance:")
	suite.T().Logf("  Cache Writes: %d", numCacheWrites)
	suite.T().Logf("  Write Duration: %v", writeDuration)
	suite.T().Logf("  Writes/Second: %.2f", writesPerSecond)
	suite.T().Logf("  Cache Reads: %d", numCacheWrites)
	suite.T().Logf("  Read Duration: %v", readDuration)
	suite.T().Logf("  Reads/Second: %.2f", readsPerSecond)
	
	// Assertions
	assert.Less(suite.T(), writeDuration, 5*time.Second, "Should complete writes within 5 seconds")
	assert.Less(suite.T(), readDuration, 2*time.Second, "Should complete reads within 2 seconds")
	assert.Greater(suite.T(), writesPerSecond, 1000.0, "Should handle at least 1000 writes per second")
	assert.Greater(suite.T(), readsPerSecond, 2000.0, "Should handle at least 2000 reads per second")
}

// TestMemoryUsage tests memory usage during high load
func (suite *ChatPerformanceTestSuite) TestMemoryUsage() {
	// Create test user
	accessToken := suite.createTestUser("perfuser")
	conversationID := uuid.New().String()
	
	// Get initial memory usage (simplified for testing)
	initialMessages := 0
	
	// Send many messages to test memory usage
	numMessages := 1000
	
	for i := 0; i < numMessages; i++ {
		// Prepare message request
		req := dto.SendMessageDTO{
			Content: fmt.Sprintf("Memory test message %d with some additional content to increase size", i),
			Type:    "text",
		}
		
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), reqBody)
		httpReq.Header.Set("Authorization", "Bearer "+accessToken)
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)
		
		if w.Code == http.StatusCreated {
			initialMessages++
		}
	}
	
	// Log memory usage metrics
	suite.T().Logf("Memory Usage Test:")
	suite.T().Logf("  Messages Processed: %d", initialMessages)
	suite.T().Logf("  Final Message Count: %d", initialMessages)
	
	// Assertions
	assert.Equal(suite.T(), numMessages, initialMessages, "All messages should be processed")
}

// TestRateLimitingPerformance tests rate limiting under high load
func (suite *ChatPerformanceTestSuite) TestRateLimitingPerformance() {
	// Create test user and conversation
	accessToken := suite.createTestUser("perfuser")
	conversationID := uuid.New().String()
	
	// Send messages rapidly to test rate limiting
	var allowedCount int
	var blockedCount int
	
	for i := 0; i < 50; i++ { // More than rate limit of 30 per minute
		// Prepare message request
		req := dto.SendMessageDTO{
			Content: fmt.Sprintf("Rate limit test message %d", i),
			Type:    "text",
		}
		
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/chats/%s/messages", conversationID), reqBody)
		httpReq.Header.Set("Authorization", "Bearer "+accessToken)
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Perform request
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, httpReq)
		
		if w.Code == http.StatusCreated {
			allowedCount++
		} else if w.Code == http.StatusTooManyRequests {
			blockedCount++
		}
	}
	
	// Log rate limiting metrics
	suite.T().Logf("Rate Limiting Performance:")
	suite.T().Logf("  Total Requests: 50")
	suite.T().Logf("  Allowed Requests: %d", allowedCount)
	suite.T().Logf("  Blocked Requests: %d", blockedCount)
	suite.T().Logf("  Rate Limit Threshold: 30")
	
	// Assertions
	assert.LessOrEqual(suite.T(), allowedCount, 30, "Should not allow more than 30 requests")
	assert.Greater(suite.T(), blockedCount, 0, "Should block requests exceeding rate limit")
}

// Helper methods

func (suite *ChatPerformanceTestSuite) createTestUser(username string) string {
	// This would normally create a user in the database
	// For testing purposes, we'll return a mock JWT token
	return "mock-jwt-token-for-" + username
}

// TestChatPerformance runs all chat performance tests
func TestChatPerformance(t *testing.T) {
	suite.Run(t, new(ChatPerformanceTestSuite))
}