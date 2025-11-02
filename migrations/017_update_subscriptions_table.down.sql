-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop new indexes
DROP INDEX IF EXISTS idx_subscriptions_stripe_customer_id;
DROP INDEX IF EXISTS idx_subscriptions_stripe_price_id;
DROP INDEX IF EXISTS idx_subscriptions_trial_end;
DROP INDEX IF EXISTS idx_subscriptions_canceled_at;

-- Drop new columns
ALTER TABLE subscriptions 
DROP COLUMN IF EXISTS stripe_customer_id,
DROP COLUMN IF EXISTS stripe_price_id,
DROP COLUMN IF EXISTS quantity,
DROP COLUMN IF EXISTS trial_start,
DROP COLUMN IF EXISTS trial_end,
DROP COLUMN IF EXISTS canceled_at,
DROP COLUMN IF EXISTS ended_at,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS metadata;

-- Restore original check constraints
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_plan_type_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_plan_type_check 
    CHECK (plan_type IN ('basic', 'premium', 'platinum'));

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check 
    CHECK (status IN ('active', 'canceled', 'past_due', 'unpaid'));