-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE swipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    swiper_id UUID NOT NULL,
    swiped_id UUID NOT NULL,
    is_like BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create foreign key constraints
ALTER TABLE swipes ADD CONSTRAINT fk_swipes_swiper_id 
    FOREIGN KEY (swiper_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE swipes ADD CONSTRAINT fk_swipes_swiped_id 
    FOREIGN KEY (swiped_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE UNIQUE INDEX idx_swipes_unique ON swipes(swiper_id, swiped_id);
CREATE INDEX idx_swipes_swiper ON swipes(swiper_id);
CREATE INDEX idx_swipes_swiped ON swipes(swiped_id);