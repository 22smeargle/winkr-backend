-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_event_id VARCHAR(255) UNIQUE,
    event_type VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processed', 'failed')),
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    payload JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_webhook_events_stripe_event_id ON webhook_events(stripe_event_id);
CREATE INDEX idx_webhook_events_event_type ON webhook_events(event_type);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created_at ON webhook_events(created_at);
CREATE INDEX idx_webhook_events_next_retry_at ON webhook_events(next_retry_at);
CREATE INDEX idx_webhook_events_retry_count ON webhook_events(retry_count);

-- Add comments for documentation
COMMENT ON TABLE webhook_events IS 'Stripe webhook event processing records';
COMMENT ON COLUMN webhook_events.stripe_event_id IS 'Stripe event ID';
COMMENT ON COLUMN webhook_events.event_type IS 'Stripe event type (e.g., invoice.payment_succeeded)';
COMMENT ON COLUMN webhook_events.status IS 'Processing status: pending, processed, failed';
COMMENT ON COLUMN webhook_events.processed_at IS 'When the event was processed';
COMMENT ON COLUMN webhook_events.error_message IS 'Error message if processing failed';
COMMENT ON COLUMN webhook_events.retry_count IS 'Number of retry attempts';
COMMENT ON COLUMN webhook_events.max_retries IS 'Maximum retry attempts allowed';
COMMENT ON COLUMN webhook_events.next_retry_at IS 'When to retry processing';
COMMENT ON COLUMN webhook_events.payload IS 'Full webhook event payload';
COMMENT ON COLUMN webhook_events.metadata IS 'Additional metadata in JSON format';