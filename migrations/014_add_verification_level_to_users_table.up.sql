-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE users ADD COLUMN verification_level INTEGER DEFAULT 0 CHECK (verification_level IN (0, 1, 2));

-- Create index for verification level
CREATE INDEX idx_users_verification_level ON users(verification_level);