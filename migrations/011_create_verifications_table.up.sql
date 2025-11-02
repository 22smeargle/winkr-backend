-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('selfie', 'document')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    photo_url VARCHAR(500) NOT NULL,
    photo_key VARCHAR(255) UNIQUE NOT NULL,
    document_type VARCHAR(50),
    document_data TEXT,
    ai_score DECIMAL(3,2),
    ai_details TEXT,
    rejection_reason TEXT,
    reviewed_by UUID,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE verifications ADD CONSTRAINT fk_verifications_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE verifications ADD CONSTRAINT fk_verifications_reviewed_by 
    FOREIGN KEY (reviewed_by) REFERENCES admin_users(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_verifications_user_id ON verifications(user_id);
CREATE INDEX idx_verifications_type ON verifications(type);
CREATE INDEX idx_verifications_status ON verifications(status);
CREATE INDEX idx_verifications_reviewed_by ON verifications(reviewed_by);
CREATE INDEX idx_verifications_expires_at ON verifications(expires_at);
CREATE INDEX idx_verifications_created_at ON verifications(created_at);
CREATE INDEX idx_verifications_user_type ON verifications(user_id, type);
CREATE INDEX idx_verifications_status_reviewed ON verifications(status, reviewed_by);