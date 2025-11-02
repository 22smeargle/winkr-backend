-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop indexes
DROP INDEX IF EXISTS idx_invoices_user_id;
DROP INDEX IF EXISTS idx_invoices_subscription_id;
DROP INDEX IF EXISTS idx_invoices_stripe_invoice_id;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_due_date;
DROP INDEX IF EXISTS idx_invoices_created_at;
DROP INDEX IF EXISTS idx_invoices_amount_due;
DROP INDEX IF EXISTS idx_invoices_stripe_charge_id;

-- Drop foreign key constraints
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS fk_invoices_user_id;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS fk_invoices_subscription_id;

-- Drop table
DROP TABLE IF EXISTS invoices;