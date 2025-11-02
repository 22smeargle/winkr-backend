-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop indexes
DROP INDEX IF EXISTS idx_payment_methods_user_id;
DROP INDEX IF EXISTS idx_payment_methods_stripe_payment_method_id;
DROP INDEX IF EXISTS idx_payment_methods_type;
DROP INDEX IF EXISTS idx_payment_methods_is_default;
DROP INDEX IF EXISTS idx_payment_methods_fingerprint;

-- Drop foreign key constraint
ALTER TABLE payment_methods DROP CONSTRAINT IF EXISTS fk_payment_methods_user_id;

-- Drop table
DROP TABLE IF EXISTS payment_methods;