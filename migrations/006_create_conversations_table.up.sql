-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraint
ALTER TABLE conversations ADD CONSTRAINT fk_conversations_match_id 
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE;

-- Create unique index for match_id
CREATE UNIQUE INDEX idx_conversations_match_id ON conversations(match_id);