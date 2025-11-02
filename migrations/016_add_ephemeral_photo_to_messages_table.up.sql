-- Add ephemeral_photo to message_type enum
ALTER TYPE message_type_enum ADD VALUE 'ephemeral_photo';

-- Add index for ephemeral photo messages for better performance
CREATE INDEX idx_messages_ephemeral_photo_type ON messages(message_type) WHERE message_type = 'ephemeral_photo';

-- Add index for conversation and message_type to optimize queries for ephemeral photos in conversations
CREATE INDEX idx_messages_conversation_ephemeral_photo ON messages(conversation_id, message_type) WHERE message_type = 'ephemeral_photo';

-- Add index for sender and message_type to optimize queries for user's ephemeral photos
CREATE INDEX idx_messages_sender_ephemeral_photo ON messages(sender_id, message_type) WHERE message_type = 'ephemeral_photo';