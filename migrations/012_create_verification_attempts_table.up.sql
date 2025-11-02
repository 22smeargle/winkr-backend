-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE verification_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('selfie', 'document')),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraint
ALTER TABLE verification_attempts ADD CONSTRAINT fk_verification_attempts_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_verification_attempts_user_id ON verification_attempts(user_id);
CREATE INDEX idx_verification_attempts_type ON verification_attempts(type);
CREATE INDEX idx_verification_attempts_status ON verification_attempts(status);
CREATE INDEX idx_verification_attempts_ip_address ON verification_attempts(ip_address);
CREATE INDEX idx_verification_attempts_created_at ON verification_attempts(created_at);
CREATE INDEX idx_verification_attempts_user_type ON verification_attempts(user_id, type);
CREATE INDEX idx_verification_attempts_ip_created ON verification_attempts(ip_address, created_at);