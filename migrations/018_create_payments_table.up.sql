-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    stripe_payment_intent_id VARCHAR(255) UNIQUE,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'canceled', 'refunded', 'partially_refunded')),
    payment_method_id UUID,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    stripe_charge_id VARCHAR(255),
    failure_reason TEXT,
    failure_code VARCHAR(50),
    receipt_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE payments ADD CONSTRAINT fk_payments_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE payments ADD CONSTRAINT fk_payments_payment_method_id 
    FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_stripe_payment_intent_id ON payments(stripe_payment_intent_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_amount ON payments(amount);
CREATE INDEX idx_payments_stripe_charge_id ON payments(stripe_charge_id);

-- Add comments for documentation
COMMENT ON TABLE payments IS 'Payment records with Stripe integration';
COMMENT ON COLUMN payments.user_id IS 'User who made the payment';
COMMENT ON COLUMN payments.stripe_payment_intent_id IS 'Stripe payment intent ID';
COMMENT ON COLUMN payments.amount IS 'Payment amount in decimal format';
COMMENT ON COLUMN payments.currency IS 'Payment currency (3-letter ISO code)';
COMMENT ON COLUMN payments.status IS 'Payment status: pending, processing, succeeded, failed, canceled, refunded, partially_refunded';
COMMENT ON COLUMN payments.payment_method_id IS 'Payment method used';
COMMENT ON COLUMN payments.description IS 'Payment description';
COMMENT ON COLUMN payments.metadata IS 'Additional metadata in JSON format';
COMMENT ON COLUMN payments.stripe_charge_id IS 'Stripe charge ID';
COMMENT ON COLUMN payments.failure_reason IS 'Reason for payment failure';
COMMENT ON COLUMN payments.failure_code IS 'Stripe failure code';
COMMENT ON COLUMN payments.receipt_url IS 'URL to payment receipt';