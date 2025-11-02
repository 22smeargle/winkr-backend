-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    subscription_id UUID,
    stripe_invoice_id VARCHAR(255) UNIQUE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible', 'paid_out_of_band')),
    amount_due DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0,
    amount_remaining DECIMAL(10,2) GENERATED ALWAYS AS (amount_due - amount_paid) STORED,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    subtotal DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) NOT NULL DEFAULT 0,
    total DECIMAL(10,2) NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    stripe_charge_id VARCHAR(255),
    hosted_invoice_url TEXT,
    invoice_pdf TEXT,
    due_date TIMESTAMP WITH TIME ZONE,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    finalized_at TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    voided_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE invoices ADD CONSTRAINT fk_invoices_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE invoices ADD CONSTRAINT fk_invoices_subscription_id 
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_invoices_user_id ON invoices(user_id);
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_stripe_invoice_id ON invoices(stripe_invoice_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);
CREATE INDEX idx_invoices_amount_due ON invoices(amount_due);
CREATE INDEX idx_invoices_stripe_charge_id ON invoices(stripe_charge_id);

-- Add comments for documentation
COMMENT ON TABLE invoices IS 'Invoice records with Stripe integration';
COMMENT ON COLUMN invoices.user_id IS 'User who owns this invoice';
COMMENT ON COLUMN invoices.subscription_id IS 'Associated subscription';
COMMENT ON COLUMN invoices.stripe_invoice_id IS 'Stripe invoice ID';
COMMENT ON COLUMN invoices.status IS 'Invoice status: draft, open, paid, void, uncollectible, paid_out_of_band';
COMMENT ON COLUMN invoices.amount_due IS 'Amount due on invoice';
COMMENT ON COLUMN invoices.amount_paid IS 'Amount already paid';
COMMENT ON COLUMN invoices.amount_remaining IS 'Remaining amount to be paid';
COMMENT ON COLUMN invoices.currency IS 'Invoice currency (3-letter ISO code)';
COMMENT ON COLUMN invoices.subtotal IS 'Invoice subtotal before tax';
COMMENT ON COLUMN invoices.tax IS 'Tax amount';
COMMENT ON COLUMN invoices.total IS 'Total invoice amount';
COMMENT ON COLUMN invoices.description IS 'Invoice description';
COMMENT ON COLUMN invoices.metadata IS 'Additional metadata in JSON format';
COMMENT ON COLUMN invoices.stripe_charge_id IS 'Associated Stripe charge ID';
COMMENT ON COLUMN invoices.hosted_invoice_url IS 'URL to hosted invoice page';
COMMENT ON COLUMN invoices.invoice_pdf IS 'URL to invoice PDF';
COMMENT ON COLUMN invoices.due_date IS 'Invoice due date';
COMMENT ON COLUMN invoices.period_start IS 'Billing period start';
COMMENT ON COLUMN invoices.period_end IS 'Billing period end';
COMMENT ON COLUMN invoices.finalized_at IS 'When invoice was finalized';
COMMENT ON COLUMN invoices.paid_at IS 'When invoice was paid';
COMMENT ON COLUMN invoices.voided_at IS 'When invoice was voided';