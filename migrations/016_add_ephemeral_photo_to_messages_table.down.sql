-- Drop indexes for ephemeral photo messages
DROP INDEX IF EXISTS idx_messages_ephemeral_photo_type;
DROP INDEX IF EXISTS idx_messages_conversation_ephemeral_photo;
DROP INDEX IF EXISTS idx_messages_sender_ephemeral_photo;

-- Note: We cannot remove the 'ephemeral_photo' value from the enum in PostgreSQL
-- This is a limitation of PostgreSQL's enum type system
-- The value will remain in the enum but won't be used after rollback