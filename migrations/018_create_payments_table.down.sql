-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_stripe_payment_intent_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_amount;
DROP INDEX IF EXISTS idx_payments_stripe_charge_id;

-- Drop foreign key constraints
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_user_id;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_payment_method_id;

-- Drop table
DROP TABLE IF EXISTS payments;