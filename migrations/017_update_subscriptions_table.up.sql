-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Add new columns to subscriptions table
ALTER TABLE subscriptions 
ADD COLUMN stripe_customer_id VARCHAR(255),
ADD COLUMN stripe_price_id VARCHAR(255),
ADD COLUMN quantity INTEGER DEFAULT 1,
ADD COLUMN trial_start TIMESTAMP WITH TIME ZONE,
ADD COLUMN trial_end TIMESTAMP WITH TIME ZONE,
ADD COLUMN canceled_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN ended_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN created_by VARCHAR(50) DEFAULT 'user',
ADD COLUMN metadata JSONB DEFAULT '{}';

-- Update plan_type check constraint to include free tier
ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_plan_type_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_plan_type_check 
    CHECK (plan_type IN ('free', 'basic', 'premium', 'platinum'));

-- Update status check constraint to include more statuses
ALTER TABLE subscriptions DROP CONSTRAINT subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check 
    CHECK (status IN ('trialing', 'active', 'past_due', 'canceled', 'unpaid', 'incomplete', 'incomplete_expired', 'paused'));

-- Add new indexes for performance
CREATE INDEX idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);
CREATE INDEX idx_subscriptions_stripe_price_id ON subscriptions(stripe_price_id);
CREATE INDEX idx_subscriptions_trial_end ON subscriptions(trial_end);
CREATE INDEX idx_subscriptions_canceled_at ON subscriptions(canceled_at);

-- Add comment for documentation
COMMENT ON TABLE subscriptions IS 'User subscription information with Stripe integration';
COMMENT ON COLUMN subscriptions.stripe_customer_id IS 'Stripe customer ID';
COMMENT ON COLUMN subscriptions.stripe_subscription_id IS 'Stripe subscription ID';
COMMENT ON COLUMN subscriptions.stripe_price_id IS 'Stripe price ID';
COMMENT ON COLUMN subscriptions.plan_type IS 'Subscription plan type: free, basic, premium, platinum';
COMMENT ON COLUMN subscriptions.status IS 'Subscription status: trialing, active, past_due, canceled, unpaid, incomplete, incomplete_expired, paused';
COMMENT ON COLUMN subscriptions.quantity IS 'Quantity of the subscription';
COMMENT ON COLUMN subscriptions.trial_start IS 'Trial period start time';
COMMENT ON COLUMN subscriptions.trial_end IS 'Trial period end time';
COMMENT ON COLUMN subscriptions.canceled_at IS 'When the subscription was canceled';
COMMENT ON COLUMN subscriptions.ended_at IS 'When the subscription ended';
COMMENT ON COLUMN subscriptions.created_by IS 'Who created the subscription: user, admin, system';
COMMENT ON COLUMN subscriptions.metadata IS 'Additional metadata in JSON format';