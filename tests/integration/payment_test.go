package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/stripe/stripe-go/v76"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/payment"
	"github.com/22smeargle/winkr-backend/internal/application/services"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/cache"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/database/redis"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe"
	"github.com/22smeargle/winkr-backend/internal/infrastructure/external/stripe/mocks"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/handlers"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/middleware"
	"github.com/22smeargle/winkr-backend/internal/interfaces/http/routes"
	"github.com/22smeargle/winkr-backend/pkg/config"
	"github.com/22smeargle/winkr-backend/pkg/utils"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// PaymentIntegrationTestSuite tests payment endpoints
type PaymentIntegrationTestSuite struct {
	suite.Suite
	router           *gin.Engine
	paymentHandler   *handlers.PaymentHandler
	redisClient      *redis.RedisClient
	stripeService     *stripe.StripeService
	mockStripeService *mocks.MockStripeService
	cacheService      *cache.PaymentCacheService
}

// SetupSuite sets up the test suite
func (suite *PaymentIntegrationTestSuite) SetupSuite() {
	// Create test dependencies
	suite.redisClient = redis.NewMockRedisClient()
	suite.mockStripeService = mocks.NewMockStripeService()
	suite.stripeService = stripe.NewStripeService(&config.StripeConfig{
		SecretKey:      "sk_test_mock",
		PublishableKey: "pk_test_mock",
		WebhookSecret:  "whsec_test_mock",
		DefaultCurrency: "usd",
		SuccessURL:      "/payment/success",
		CancelURL:       "/payment/cancel",
		WebhookEndpoint: "/api/v1/payment/webhook",
		EnableRadar:     true,
		FraudLevel:      "normal",
		PaymentRateLimit: 10,
		CacheTTL:        15 * time.Minute,
	})
	
	suite.cacheService = cache.NewPaymentCacheService(suite.redisClient)
	
	// Create payment validator
	paymentValidator := validator.NewPaymentValidator()
	
	// Create rate limiter
	rateLimiter := middleware.NewAuthRateLimiter(suite.redisClient)
	
	// Create JWT utils
	jwtUtils := utils.NewJWTUtils("test-secret", time.Hour, time.Hour*24*7)
	
	// Create use cases
	getPlansUseCase := payment.NewGetPlansUseCase(suite.stripeService)
	getPlanByIDUseCase := payment.NewGetPlanByIDUseCase(suite.stripeService)
	subscribeUseCase := payment.NewSubscribeUseCase(nil, nil, nil, suite.stripeService, suite.cacheService)
	getSubscriptionUseCase := payment.NewGetSubscriptionUseCase(nil, nil, suite.cacheService)
	getActiveSubscriptionUseCase := payment.NewGetActiveSubscriptionUseCase(nil, nil, suite.cacheService)
	cancelSubscriptionUseCase := payment.NewCancelSubscriptionUseCase(nil, suite.stripeService, suite.cacheService)
	updateSubscriptionUseCase := payment.NewUpdateSubscriptionUseCase(nil, suite.stripeService, suite.cacheService)
	addPaymentMethodUseCase := payment.NewAddPaymentMethodUseCase(nil, nil, suite.stripeService, suite.cacheService)
	getPaymentMethodsUseCase := payment.NewGetPaymentMethodsUseCase(nil, nil, suite.cacheService)
	getDefaultPaymentMethodUseCase := payment.NewGetDefaultPaymentMethodUseCase(nil, nil, suite.cacheService)
	deletePaymentMethodUseCase := payment.NewDeletePaymentMethodUseCase(nil, suite.stripeService, suite.cacheService)
	processWebhookUseCase := payment.NewProcessWebhookUseCase(nil, nil, nil, nil, nil, nil, nil, nil, suite.stripeService, suite.cacheService)
	
	// Create subscription service
	subscriptionService := services.NewSubscriptionService(nil, suite.stripeService, suite.cacheService)
	
	// Create payment handler
	suite.paymentHandler = handlers.NewPaymentHandler(
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
		jwtUtils,
	)
	
	// Create router
	suite.router = gin.New()
	
	// Add payment routes
	paymentRoutes := routes.NewPaymentRoutes(
		suite.paymentHandler,
		rateLimiter,
		&config.StripeConfig{},
	)
	
	paymentGroup := suite.router.Group("/api/v1/payment")
	paymentRoutes.RegisterRoutes(paymentGroup, suite.redisClient)
}

// TestGetPlans tests the GET /plans endpoint
func (suite *PaymentIntegrationTestSuite) TestGetPlans() {
	// Setup mock response
	suite.mockStripeService.SetPlansResponse([]*stripe.Price{
		{
			ID:       "price_free",
			Nickname: "Free Plan",
			Amount:   0,
			Currency: "usd",
			Recurring: &stripe.PriceRecurring{
				Interval: stripe.PriceRecurringIntervalMonth,
			},
		},
		{
			ID:       "price_premium",
			Nickname: "Premium Plan",
			Amount:   999, // $9.99
			Currency: "usd",
			Recurring: &stripe.PriceRecurring{
				Interval: stripe.PriceRecurringIntervalMonth,
			},
		},
		{
			ID:       "price_platinum",
			Nickname: "Platinum Plan",
			Amount:   1999, // $19.99
			Currency: "usd",
			Recurring: &stripe.PriceRecurring{
				Interval: stripe.PriceRecurringIntervalMonth,
			},
		},
	}, nil)
	
	// Make request
	req := httptest.NewRequest("GET", "/api/v1/payment/plans", nil)
	req.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.PlansResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Len(suite.T(), response.Data.Plans, 3)
}

// TestSubscribe tests the POST /subscribe endpoint
func (suite *PaymentIntegrationTestSuite) TestSubscribe() {
	// Setup mock responses
	userID := uuid.New()
	
	suite.mockStripeService.SetCreateCustomerResponse(&stripe.Customer{
		ID: "cus_test123",
		Email: "test@example.com",
	}, nil)
	
	suite.mockStripeService.SetCreateSubscriptionResponse(&stripe.Subscription{
		ID: "sub_test123",
		Customer: &stripe.Customer{
			ID: "cus_test123",
		},
		Items: []*stripe.SubscriptionItem{
			{
				Price: &stripe.Price{
					ID: "price_premium",
				},
			},
		},
		Status: "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}, nil)
	
	// Prepare subscription request
	req := dto.SubscribeRequestDTO{
		PriceID: "price_premium",
		PaymentMethodID: "pm_test123",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/payment/subscribe", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer mock-token")
	httpReq.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.SubscriptionResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.Subscription.StripeSubscriptionID, "sub_test123")
	assert.Equal(suite.T(), response.Data.Subscription.Status, "active")
}

// TestGetSubscription tests the GET /me/subscription endpoint
func (suite *PaymentIntegrationTestSuite) TestGetSubscription() {
	// Setup mock response
	userID := uuid.New()
	
	suite.mockStripeService.SetGetSubscriptionResponse(&stripe.Subscription{
		ID: "sub_test123",
		Customer: &stripe.Customer{
			ID: "cus_test123",
		},
		Items: []*stripe.SubscriptionItem{
			{
				Price: &stripe.Price{
					ID: "price_premium",
				},
			},
		},
		Status: "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
	}, nil)
	
	// Make request
	req := httptest.NewRequest("GET", "/api/v1/payment/me/subscription", nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	req.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.SubscriptionResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.Subscription.StripeSubscriptionID, "sub_test123")
}

// TestCancelSubscription tests the POST /cancel endpoint
func (suite *PaymentIntegrationTestSuite) TestCancelSubscription() {
	// Setup mock response
	userID := uuid.New()
	
	suite.mockStripeService.SetCancelSubscriptionResponse(&stripe.Subscription{
		ID: "sub_test123",
		Customer: &stripe.Customer{
			ID: "cus_test123",
		},
		Status: "canceled",
		CancelAtPeriodEnd: true,
	}, nil)
	
	// Prepare cancel request
	req := dto.CancelSubscriptionRequestDTO{
		Immediate: false,
		Reason:    "User requested cancellation",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/payment/cancel", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer mock-token")
	httpReq.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestAddPaymentMethod tests the POST /methods endpoint
func (suite *PaymentIntegrationTestSuite) TestAddPaymentMethod() {
	// Setup mock response
	userID := uuid.New()
	
	suite.mockStripeService.SetCreatePaymentMethodResponse(&stripe.PaymentMethod{
		ID: "pm_test123",
		Type: "card",
		Card: &stripe.PaymentMethodCard{
			Brand:  "visa",
			Last4:  "4242",
			ExpMonth: 12,
			ExpYear:  2025,
		},
	}, nil)
	
	// Prepare payment method request
	req := dto.AddPaymentMethodRequestDTO{
		PaymentMethodID: "pm_stripe_test123",
		IsDefault:      true,
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/payment/methods", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer mock-token")
	httpReq.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response dto.PaymentMethodResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Equal(suite.T(), response.Data.PaymentMethod.StripePaymentMethodID, "pm_test123")
}

// TestGetPaymentMethods tests the GET /methods endpoint
func (suite *PaymentIntegrationTestSuite) TestGetPaymentMethods() {
	// Setup mock response
	userID := uuid.New()
	
	suite.mockStripeService.SetListPaymentMethodsResponse(&stripe.PaymentMethodList{
		Data: []*stripe.PaymentMethod{
			{
				ID:   "pm_test123",
				Type: "card",
				Card: &stripe.PaymentMethodCard{
					Brand:  "visa",
					Last4:  "4242",
					ExpMonth: 12,
					ExpYear:  2025,
				},
			},
			{
				ID:   "pm_test456",
				Type: "card",
				Card: &stripe.PaymentMethodCard{
					Brand:  "mastercard",
					Last4:  "5555",
					ExpMonth: 6,
					ExpYear:  2024,
				},
			},
		},
	}, nil)
	
	// Make request
	req := httptest.NewRequest("GET", "/api/v1/payment/methods", nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	req.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.PaymentMethodsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotNil(suite.T(), response.Data)
	assert.Len(suite.T(), response.Data.PaymentMethods, 2)
}

// TestDeletePaymentMethod tests the DELETE /methods/:id endpoint
func (suite *PaymentIntegrationTestSuite) TestDeletePaymentMethod() {
	// Setup mock response
	userID := uuid.New()
	paymentMethodID := uuid.New()
	
	suite.mockStripeService.SetDetachPaymentMethodResponse(nil)
	
	// Make request
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/payment/methods/%s", paymentMethodID.String()), nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	req.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response dto.MessageResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
	assert.NotEmpty(suite.T(), response.Message)
}

// TestWebhookProcessing tests the POST /webhook endpoint
func (suite *PaymentIntegrationTestSuite) TestWebhookProcessing() {
	// Setup mock response
	userID := uuid.New()
	
	suite.mockStripeService.SetProcessEventResponse(nil)
	
	// Prepare webhook event
	webhookEvent := map[string]interface{}{
		"id":   "evt_test123",
		"type": "invoice.payment_succeeded",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":     "in_test123",
				"amount":  999,
				"currency": "usd",
				"status":  "paid",
			},
		},
	}
	
	reqBody, _ := json.Marshal(webhookEvent)
	httpReq := httptest.NewRequest("POST", "/api/v1/payment/webhook", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Stripe-Signature", "mock-signature")
	httpReq.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Check response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestPaymentSecurity tests security measures for payment endpoints
func (suite *PaymentIntegrationTestSuite) TestPaymentSecurity() {
	// Test without authentication
	req := httptest.NewRequest("GET", "/api/v1/payment/plans", nil)
	req.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return unauthorized
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	// Test rate limiting
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("GET", "/api/v1/payment/plans", nil)
		req.Header.Set("Authorization", "Bearer mock-token")
		req.Header.Set("User-Agent", "test-agent")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		// First 10 requests should succeed
		if i < 10 {
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		} else {
			// Subsequent requests should be rate limited
			assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
		}
	}
}

// TestPaymentInputValidation tests input validation for payment endpoints
func (suite *PaymentIntegrationTestSuite) TestPaymentInputValidation() {
	// Test invalid subscription request
	req := dto.SubscribeRequestDTO{
		PriceID:        "", // Invalid: empty
		PaymentMethodID: "pm_test123",
	}
	
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/payment/subscribe", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer mock-token")
	httpReq.Header.Set("User-Agent", "test-agent")
	
	// Perform request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, httpReq)
	
	// Should return validation error
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response dto.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), response.Success)
	assert.Contains(suite.T(), response.Error.Message, "price_id")
}

// TestPaymentCaching tests caching functionality for payment data
func (suite *PaymentIntegrationTestSuite) TestPaymentCaching() {
	userID := uuid.New().String()
	
	// Test subscription caching
	subscription := &entities.Subscription{
		ID:                 uuid.New(),
		UserID:             uuid.New(),
		StripeSubscriptionID: "sub_test123",
		PlanType:            "premium",
		Status:              "active",
		CurrentPeriodStart:    time.Now(),
		CurrentPeriodEnd:      time.Now().AddDate(0, 1, 0),
	}
	
	err := suite.cacheService.CacheSubscription(context.Background(), userID, subscription)
	require.NoError(suite.T(), err)
	
	cachedSubscription, err := suite.cacheService.GetSubscription(context.Background(), userID)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cachedSubscription)
	assert.Equal(suite.T(), cachedSubscription.StripeSubscriptionID, "sub_test123")
	
	// Test payment methods caching
	paymentMethods := []*entities.PaymentMethod{
		{
			ID:                 uuid.New(),
			UserID:             uuid.New(),
			StripePaymentMethodID: "pm_test123",
			Type:                "card",
			IsDefault:           true,
			CardBrand:            "visa",
			CardLast4:            "4242",
		},
	}
	
	err = suite.cacheService.CachePaymentMethods(context.Background(), userID, paymentMethods)
	require.NoError(suite.T(), err)
	
	cachedPaymentMethods, err := suite.cacheService.GetPaymentMethods(context.Background(), userID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), cachedPaymentMethods, 1)
	assert.Equal(suite.T(), cachedPaymentMethods[0].StripePaymentMethodID, "pm_test123")
	
	// Test cache invalidation
	err = suite.cacheService.InvalidateSubscription(context.Background(), userID)
	require.NoError(suite.T(), err)
	
	_, err = suite.cacheService.GetSubscription(context.Background(), userID)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), nil) // Should be cache miss
}

// TestPaymentIntegration runs all payment integration tests
func TestPaymentIntegration(t *testing.T) {
	suite.Run(t, new(PaymentIntegrationTestSuite))
}