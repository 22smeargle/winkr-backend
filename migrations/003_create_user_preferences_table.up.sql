-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    age_min INTEGER DEFAULT 18,
    age_max INTEGER DEFAULT 100,
    max_distance INTEGER DEFAULT 50,
    show_me BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraint
ALTER TABLE user_preferences ADD CONSTRAINT fk_user_preferences_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create unique index for user_id
CREATE UNIQUE INDEX idx_user_preferences_user_id ON user_preferences(user_id);