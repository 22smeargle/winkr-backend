-- Create ephemeral_photos table
CREATE TABLE ephemeral_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_url VARCHAR(500) NOT NULL,
    file_key VARCHAR(255) NOT NULL UNIQUE,
    thumbnail_url VARCHAR(500) NOT NULL,
    thumbnail_key VARCHAR(255) NOT NULL,
    access_key VARCHAR(255) NOT NULL UNIQUE,
    is_viewed BOOLEAN DEFAULT FALSE,
    is_expired BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    max_views INTEGER DEFAULT 1,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    viewed_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for ephemeral_photos
CREATE INDEX idx_ephemeral_photos_user_id ON ephemeral_photos(user_id);
CREATE INDEX idx_ephemeral_photos_access_key ON ephemeral_photos(access_key);
CREATE INDEX idx_ephemeral_photos_is_viewed ON ephemeral_photos(is_viewed);
CREATE INDEX idx_ephemeral_photos_is_expired ON ephemeral_photos(is_expired);
CREATE INDEX idx_ephemeral_photos_expires_at ON ephemeral_photos(expires_at);
CREATE INDEX idx_ephemeral_photos_is_deleted ON ephemeral_photos(is_deleted);
CREATE INDEX idx_ephemeral_photos_created_at ON ephemeral_photos(created_at);
CREATE INDEX idx_ephemeral_photos_viewed_at ON ephemeral_photos(viewed_at);
CREATE INDEX idx_ephemeral_photos_expired_at ON ephemeral_photos(expired_at);

-- Create ephemeral_photo_views table
CREATE TABLE ephemeral_photo_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    photo_id UUID NOT NULL REFERENCES ephemeral_photos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration INTEGER DEFAULT 0,
    is_expired BOOLEAN DEFAULT FALSE
);

-- Create indexes for ephemeral_photo_views
CREATE INDEX idx_ephemeral_photo_views_photo_id ON ephemeral_photo_views(photo_id);
CREATE INDEX idx_ephemeral_photo_views_user_id ON ephemeral_photo_views(user_id);
CREATE INDEX idx_ephemeral_photo_views_viewer_id ON ephemeral_photo_views(viewer_id);
CREATE INDEX idx_ephemeral_photo_views_viewed_at ON ephemeral_photo_views(viewed_at);
CREATE INDEX idx_ephemeral_photo_views_is_expired ON ephemeral_photo_views(is_expired);

-- Add comments
COMMENT ON TABLE ephemeral_photos IS 'Table for storing ephemeral photos that expire after viewing or time';
COMMENT ON TABLE ephemeral_photo_views IS 'Table for tracking ephemeral photo views for analytics';

COMMENT ON COLUMN ephemeral_photos.access_key IS 'Unique access key for viewing the photo';
COMMENT ON COLUMN ephemeral_photos.max_views IS 'Maximum number of times the photo can be viewed';
COMMENT ON COLUMN ephemeral_photos.expires_at IS 'Expiration time for the photo';
COMMENT ON COLUMN ephemeral_photos.viewed_at IS 'Time when the photo was first viewed';
COMMENT ON COLUMN ephemeral_photos.expired_at IS 'Time when the photo expired';

COMMENT ON COLUMN ephemeral_photo_views.viewer_id IS 'Optional: ID of the user who viewed the photo (if authenticated)';
COMMENT ON COLUMN ephemeral_photo_views.ip_address IS 'IP address of the viewer';
COMMENT ON COLUMN ephemeral_photo_views.user_agent IS 'User agent string of the viewer';
COMMENT ON COLUMN ephemeral_photo_views.duration IS 'Duration in seconds that the photo was viewed';
COMMENT ON COLUMN ephemeral_photo_views.is_expired IS 'Whether this view record is expired';