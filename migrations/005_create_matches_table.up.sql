-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id UUID NOT NULL,
    user2_id UUID NOT NULL,
    matched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE matches ADD CONSTRAINT fk_matches_user1_id 
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE matches ADD CONSTRAINT fk_matches_user2_id 
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE UNIQUE INDEX idx_matches_users ON matches(LEAST(user1_id, user2_id), GREATEST(user1_id, user2_id));
CREATE INDEX idx_matches_user1 ON matches(user1_id);
CREATE INDEX idx_matches_user2 ON matches(user2_id);
CREATE INDEX idx_matches_active ON matches(is_active);