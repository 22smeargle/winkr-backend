package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/application/usecases/payment"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/logger"
	"github.com/22smeargle/winkr-backend/pkg/utils"
	"github.com/22smeargle/winkr-backend/pkg/validator"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	getPlansUseCase              *payment.GetPlansUseCase
	getPlanByIDUseCase            *payment.GetPlanByIDUseCase
	subscribeUseCase              *payment.SubscribeUseCase
	getSubscriptionUseCase         *payment.GetSubscriptionUseCase
	getActiveSubscriptionUseCase   *payment.GetActiveSubscriptionUseCase
	cancelSubscriptionUseCase       *payment.CancelSubscriptionUseCase
	getPaymentMethodsUseCase       *payment.GetPaymentMethodsUseCase
	getDefaultPaymentMethodUseCase *payment.GetDefaultPaymentMethodUseCase
	addPaymentMethodUseCase        *payment.AddPaymentMethodUseCase
	deletePaymentMethodUseCase     *payment.DeletePaymentMethodUseCase
	processWebhookUseCase          *payment.ProcessWebhookUseCase
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(
	getPlansUseCase *payment.GetPlansUseCase,
	getPlanByIDUseCase *payment.GetPlanByIDUseCase,
	subscribeUseCase *payment.SubscribeUseCase,
	getSubscriptionUseCase *payment.GetSubscriptionUseCase,
	getActiveSubscriptionUseCase *payment.GetActiveSubscriptionUseCase,
	cancelSubscriptionUseCase *payment.CancelSubscriptionUseCase,
	getPaymentMethodsUseCase *payment.GetPaymentMethodsUseCase,
	getDefaultPaymentMethodUseCase *payment.GetDefaultPaymentMethodUseCase,
	addPaymentMethodUseCase *payment.AddPaymentMethodUseCase,
	deletePaymentMethodUseCase *payment.DeletePaymentMethodUseCase,
	processWebhookUseCase *payment.ProcessWebhookUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		getPlansUseCase:              getPlansUseCase,
		getPlanByIDUseCase:            getPlanByIDUseCase,
		subscribeUseCase:              subscribeUseCase,
		getSubscriptionUseCase:         getSubscriptionUseCase,
		getActiveSubscriptionUseCase:   getActiveSubscriptionUseCase,
		cancelSubscriptionUseCase:       cancelSubscriptionUseCase,
		getPaymentMethodsUseCase:       getPaymentMethodsUseCase,
		getDefaultPaymentMethodUseCase: getDefaultPaymentMethodUseCase,
		addPaymentMethodUseCase:        addPaymentMethodUseCase,
		deletePaymentMethodUseCase:     deletePaymentMethodUseCase,
		processWebhookUseCase:          processWebhookUseCase,
	}
}

// GetPlans handles GET /plans endpoint
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	logger.Info("Getting subscription plans", nil)

	plans, err := h.getPlansUseCase.Execute(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get subscription plans", err, nil)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription plans")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Subscription plans retrieved successfully", gin.H{
		"plans": plans,
	})
}

// GetPlanByID handles GET /plans/:id endpoint
func (h *PaymentHandler) GetPlanByID(c *gin.Context) {
	planID := c.Param("id")
	if planID == "" {
		logger.Error("Plan ID is required", nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Plan ID is required")
		return
	}

	logger.Info("Getting subscription plan by ID", map[string]interface{}{
		"plan_id": planID,
	})

	plan, err := h.getPlanByIDUseCase.Execute(c.Request.Context(), planID)
	if err != nil {
		if err == payment.ErrPlanNotFound {
			logger.Error("Subscription plan not found", nil, map[string]interface{}{
				"plan_id": planID,
			})
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription plan not found")
			return
		}

		logger.Error("Failed to get subscription plan", err, map[string]interface{}{
			"plan_id": planID,
		})
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription plan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Subscription plan retrieved successfully", gin.H{
		"plan": plan,
	})
}

// Subscribe handles POST /subscribe endpoint
func (h *PaymentHandler) Subscribe(c *gin.Context) {
	var req payment.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind subscription request", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		logger.Error("Validation failed", err, map[string]interface{}{
			"request": req,
		})
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("Creating subscription", map[string]interface{}{
		"user_id":    req.UserID,
		"plan_id":     req.PlanID,
	})

	response, err := h.subscribeUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		switch err {
		case payment.ErrPlanNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription plan not found")
		case payment.ErrUserHasActiveSubscription:
			utils.ErrorResponse(c, http.StatusConflict, "User already has an active subscription")
		case payment.ErrInvalidUserID:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		default:
			logger.Error("Failed to create subscription", err, map[string]interface{}{
				"user_id": req.UserID,
				"plan_id": req.PlanID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Subscription created successfully", response)
}

// GetSubscription handles GET /subscription endpoint
func (h *PaymentHandler) GetSubscription(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		logger.Error("Invalid user ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	logger.Info("Getting user subscription", map[string]interface{}{
		"user_id": userID,
	})

	subscription, err := h.getSubscriptionUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err == payment.ErrSubscriptionNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		} else {
			logger.Error("Failed to get subscription", err, map[string]interface{}{
				"user_id": userID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Subscription retrieved successfully", gin.H{
		"subscription": subscription,
	})
}

// GetActiveSubscription handles GET /subscription/active endpoint
func (h *PaymentHandler) GetActiveSubscription(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		logger.Error("Invalid user ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	logger.Info("Getting user active subscription", map[string]interface{}{
		"user_id": userID,
	})

	subscription, err := h.getActiveSubscriptionUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err == payment.ErrUserHasNoActiveSubscription {
			utils.ErrorResponse(c, http.StatusNotFound, "No active subscription found")
		} else {
			logger.Error("Failed to get active subscription", err, map[string]interface{}{
				"user_id": userID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active subscription")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Active subscription retrieved successfully", gin.H{
		"subscription": subscription,
	})
}

// CancelSubscription handles POST /subscription/cancel endpoint
func (h *PaymentHandler) CancelSubscription(c *gin.Context) {
	var req payment.CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind cancel subscription request", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		logger.Error("Validation failed", err, map[string]interface{}{
			"request": req,
		})
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("Canceling subscription", map[string]interface{}{
		"user_id":            req.UserID,
		"cancel_at_period_end": req.CancelAtPeriodEnd,
		"reason":              req.Reason,
	})

	err = h.cancelSubscriptionUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		switch err {
		case payment.ErrSubscriptionNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		case payment.ErrSubscriptionCanceled:
			utils.ErrorResponse(c, http.StatusConflict, "Subscription already canceled")
		case payment.ErrSubscriptionExpired:
			utils.ErrorResponse(c, http.StatusConflict, "Subscription already expired")
		case payment.ErrCannotCancelSubscription:
			utils.ErrorResponse(c, http.StatusBadRequest, "Cannot cancel subscription")
		case payment.ErrInvalidUserID:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		default:
			logger.Error("Failed to cancel subscription", err, map[string]interface{}{
				"user_id": req.UserID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel subscription")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Subscription canceled successfully", nil)
}

// GetPaymentMethods handles GET /payment-methods endpoint
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		logger.Error("Invalid user ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	paymentMethodType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logger.Info("Getting user payment methods", map[string]interface{}{
		"user_id":           userID,
		"payment_method_type": paymentMethodType,
		"limit":             limit,
		"offset":            offset,
	})

	paymentMethods, err := h.getPaymentMethodsUseCase.Execute(c.Request.Context(), userID, paymentMethodType, limit, offset)
	if err != nil {
		logger.Error("Failed to get payment methods", err, map[string]interface{}{
			"user_id": userID,
		})
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get payment methods")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment methods retrieved successfully", gin.H{
		"payment_methods": paymentMethods,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetDefaultPaymentMethod handles GET /payment-methods/default endpoint
func (h *PaymentHandler) GetDefaultPaymentMethod(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		logger.Error("Invalid user ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	logger.Info("Getting user default payment method", map[string]interface{}{
		"user_id": userID,
	})

	paymentMethod, err := h.getDefaultPaymentMethodUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		if err == payment.ErrUserHasNoDefaultPaymentMethod {
			utils.ErrorResponse(c, http.StatusNotFound, "No default payment method found")
		} else {
			logger.Error("Failed to get default payment method", err, map[string]interface{}{
				"user_id": userID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get default payment method")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Default payment method retrieved successfully", gin.H{
		"payment_method": paymentMethod,
	})
}

// AddPaymentMethod handles POST /payment-methods endpoint
func (h *PaymentHandler) AddPaymentMethod(c *gin.Context) {
	var req payment.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind add payment method request", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := validator.ValidateStruct(&req); err != nil {
		logger.Error("Validation failed", err, map[string]interface{}{
			"request": req,
		})
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("Adding payment method", map[string]interface{}{
		"user_id": req.UserID,
		"type":    req.Type,
	})

	paymentMethod, err := h.addPaymentMethodUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		switch err {
		case payment.ErrInvalidUserID:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		case payment.ErrInvalidPaymentMethodType:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method type")
		default:
			logger.Error("Failed to add payment method", err, map[string]interface{}{
				"user_id": req.UserID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add payment method")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Payment method added successfully", gin.H{
		"payment_method": paymentMethod,
	})
}

// DeletePaymentMethod handles DELETE /payment-methods/:id endpoint
func (h *PaymentHandler) DeletePaymentMethod(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		logger.Error("Invalid user ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	paymentMethodID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.Error("Invalid payment method ID", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	logger.Info("Deleting payment method", map[string]interface{}{
		"user_id":          userID,
		"payment_method_id": paymentMethodID,
	})

	err = h.deletePaymentMethodUseCase.Execute(c.Request.Context(), userID, paymentMethodID)
	if err != nil {
		switch err {
		case payment.ErrInvalidUserID:
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		case payment.ErrPaymentMethodNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, "Payment method not found")
		case payment.ErrPaymentAccessDenied:
			utils.ErrorResponse(c, http.StatusForbidden, "Payment method access denied")
		case payment.ErrDefaultPaymentMethodRequired:
			utils.ErrorResponse(c, http.StatusBadRequest, "Cannot delete default payment method")
		default:
			logger.Error("Failed to delete payment method", err, map[string]interface{}{
				"user_id":          userID,
				"payment_method_id": paymentMethodID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete payment method")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment method deleted successfully", nil)
}

// ProcessWebhook handles POST /webhook endpoint
func (h *PaymentHandler) ProcessWebhook(c *gin.Context) {
	// Read request body
	body, err := c.GetRawData()
	if err != nil {
		logger.Error("Failed to read request body", err, nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Get Stripe signature header
	signatureHeader := c.GetHeader("Stripe-Signature")
	if signatureHeader == "" {
		logger.Error("Missing Stripe signature header", nil)
		utils.ErrorResponse(c, http.StatusBadRequest, "Missing Stripe signature header")
		return
	}

	logger.Info("Processing webhook", map[string]interface{}{
		"signature_length": len(signatureHeader),
		"body_length":       len(body),
	})

	err = h.processWebhookUseCase.Execute(c.Request.Context(), []byte(body), signatureHeader)
	if err != nil {
		switch err {
		case payment.ErrWebhookSignatureInvalid:
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid webhook signature")
		case payment.ErrWebhookEventNotSupported:
			utils.ErrorResponse(c, http.StatusBadRequest, "Webhook event not supported")
		case payment.ErrWebhookAlreadyProcessed:
			utils.ErrorResponse(c, http.StatusConflict, "Webhook event already processed")
		default:
			logger.Error("Failed to process webhook", err, nil)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process webhook")
		}
		return
	}

	// Return 200 OK for successful webhook processing
	c.Status(http.StatusOK)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}