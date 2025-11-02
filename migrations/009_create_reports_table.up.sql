-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL,
    reported_user_id UUID NOT NULL,
    reason VARCHAR(100) NOT NULL CHECK (reason IN ('inappropriate_behavior', 'fake_profile', 'spam', 'harassment', 'other')),
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
    reviewed_by UUID,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE reports ADD CONSTRAINT fk_reports_reporter_id 
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE reports ADD CONSTRAINT fk_reports_reported_user_id 
    FOREIGN KEY (reported_user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_reports_reporter ON reports(reporter_id);
CREATE INDEX idx_reports_reported ON reports(reported_user_id);
CREATE INDEX idx_reports_reason ON reports(reason);
CREATE INDEX idx_reports_status ON reports(status);