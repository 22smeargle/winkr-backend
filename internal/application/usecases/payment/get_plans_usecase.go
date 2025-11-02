package payment

import (
	"context"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetPlansUseCase retrieves available subscription plans
type GetPlansUseCase struct{}

// NewGetPlansUseCase creates a new GetPlansUseCase
func NewGetPlansUseCase() *GetPlansUseCase {
	return &GetPlansUseCase{}
}

// Execute retrieves available subscription plans
func (uc *GetPlansUseCase) Execute(ctx context.Context) ([]entities.SubscriptionPlan, error) {
	logger.Info("Getting available subscription plans", nil)
	
	plans := entities.GetAvailablePlans()
	
	logger.Info("Retrieved subscription plans", map[string]interface{}{
		"count": len(plans),
	})
	
	return plans, nil
}

// GetPlanByIDUseCase retrieves a subscription plan by ID
type GetPlanByIDUseCase struct{}

// NewGetPlanByIDUseCase creates a new GetPlanByIDUseCase
func NewGetPlanByIDUseCase() *GetPlanByIDUseCase {
	return &GetPlanByIDUseCase{}
}

// Execute retrieves a subscription plan by ID
func (uc *GetPlanByIDUseCase) Execute(ctx context.Context, planID string) (*entities.SubscriptionPlan, error) {
	logger.Info("Getting subscription plan by ID", map[string]interface{}{
		"plan_id": planID,
	})
	
	plan, exists := entities.GetPlanByID(planID)
	if !exists {
		logger.Error("Subscription plan not found", nil, map[string]interface{}{
			"plan_id": planID,
		})
		return nil, ErrPlanNotFound
	}
	
	logger.Info("Retrieved subscription plan", map[string]interface{}{
		"plan_id": plan.ID,
		"name":    plan.Name,
	})
	
	return plan, nil
}