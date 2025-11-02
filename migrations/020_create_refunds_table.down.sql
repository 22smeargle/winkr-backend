-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop indexes
DROP INDEX IF EXISTS idx_refunds_payment_id;
DROP INDEX IF EXISTS idx_refunds_stripe_refund_id;
DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_reason;
DROP INDEX IF EXISTS idx_refunds_created_at;
DROP INDEX IF EXISTS idx_refunds_amount;
DROP INDEX IF EXISTS idx_refunds_stripe_charge_id;

-- Drop foreign key constraint
ALTER TABLE refunds DROP CONSTRAINT IF EXISTS fk_refunds_payment_id;

-- Drop table
DROP TABLE IF EXISTS refunds;