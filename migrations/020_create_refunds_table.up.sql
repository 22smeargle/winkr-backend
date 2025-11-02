-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    stripe_refund_id VARCHAR(255) UNIQUE,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed', 'canceled')),
    reason VARCHAR(50) CHECK (reason IN ('duplicate', 'fraudulent', 'requested_by_customer', 'expired_uncaptured_charge')),
    metadata JSONB DEFAULT '{}',
    stripe_charge_id VARCHAR(255),
    receipt_number VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraint
ALTER TABLE refunds ADD CONSTRAINT fk_refunds_payment_id 
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);
CREATE INDEX idx_refunds_stripe_refund_id ON refunds(stripe_refund_id);
CREATE INDEX idx_refunds_status ON refunds(status);
CREATE INDEX idx_refunds_reason ON refunds(reason);
CREATE INDEX idx_refunds_created_at ON refunds(created_at);
CREATE INDEX idx_refunds_amount ON refunds(amount);
CREATE INDEX idx_refunds_stripe_charge_id ON refunds(stripe_charge_id);

-- Add comments for documentation
COMMENT ON TABLE refunds IS 'Refund records with Stripe integration';
COMMENT ON COLUMN refunds.payment_id IS 'Original payment being refunded';
COMMENT ON COLUMN refunds.stripe_refund_id IS 'Stripe refund ID';
COMMENT ON COLUMN refunds.amount IS 'Refund amount in decimal format';
COMMENT ON COLUMN refunds.currency IS 'Refund currency (3-letter ISO code)';
COMMENT ON COLUMN refunds.status IS 'Refund status: pending, succeeded, failed, canceled';
COMMENT ON COLUMN refunds.reason IS 'Refund reason: duplicate, fraudulent, requested_by_customer, expired_uncaptured_charge';
COMMENT ON COLUMN refunds.metadata IS 'Additional metadata in JSON format';
COMMENT ON COLUMN refunds.stripe_charge_id IS 'Associated Stripe charge ID';
COMMENT ON COLUMN refunds.receipt_number IS 'Refund receipt number';
COMMENT ON COLUMN refunds.description IS 'Refund description';