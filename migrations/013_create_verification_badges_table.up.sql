-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE verification_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    level INTEGER NOT NULL CHECK (level IN (0, 1, 2)),
    badge_type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE verification_badges ADD CONSTRAINT fk_verification_badges_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE verification_badges ADD CONSTRAINT fk_verification_badges_revoked_by 
    FOREIGN KEY (revoked_by) REFERENCES admin_users(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_verification_badges_user_id ON verification_badges(user_id);
CREATE INDEX idx_verification_badges_level ON verification_badges(level);
CREATE INDEX idx_verification_badges_type ON verification_badges(badge_type);
CREATE INDEX idx_verification_badges_is_revoked ON verification_badges(is_revoked);
CREATE INDEX idx_verification_badges_expires_at ON verification_badges(expires_at);
CREATE INDEX idx_verification_badges_user_level ON verification_badges(user_id, level);
CREATE INDEX idx_verification_badges_active ON verification_badges(is_revoked, expires_at);