-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

-- Drop indexes
DROP INDEX IF EXISTS idx_webhook_events_stripe_event_id;
DROP INDEX IF EXISTS idx_webhook_events_event_type;
DROP INDEX IF EXISTS idx_webhook_events_status;
DROP INDEX IF EXISTS idx_webhook_events_created_at;
DROP INDEX IF EXISTS idx_webhook_events_next_retry_at;
DROP INDEX IF EXISTS idx_webhook_events_retry_count;

-- Drop table
DROP TABLE IF EXISTS webhook_events;